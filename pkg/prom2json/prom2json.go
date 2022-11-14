package prom2json

import (
	"crypto/tls"
	"fmt"
	"github.com/exporterpush/global"
	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"time"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

// Family mirrors the MetricFamily proto message.
type Family struct {
	//Time    time.Time
	Name    string        `json:"name"`
	Help    string        `json:"help"`
	Type    string        `json:"type"`
	Metrics []interface{} `json:"metrics,omitempty"` // Either metric or summary.
}

// Metric is for all "single value" metrics, i.e. Counter, Gauge, and Untyped.
type Metric struct {
	Labels      map[string]string `json:"labels,omitempty"`
	TimestampMs string            `json:"timestamp_ms,omitempty"`
	Value       string            `json:"value"`
}

// Summary mirrors the Summary proto message.
type Summary struct {
	Labels      map[string]string `json:"labels,omitempty"`
	TimestampMs string            `json:"timestamp_ms,omitempty"`
	Quantiles   map[string]string `json:"quantiles,omitempty"`
	Count       string            `json:"count"`
	Sum         string            `json:"sum"`
}

// Histogram mirrors the Histogram proto message.
type Histogram struct {
	Labels      map[string]string `json:"labels,omitempty"`
	TimestampMs string            `json:"timestamp_ms,omitempty"`
	Buckets     map[string]string `json:"buckets,omitempty"`
	Count       string            `json:"count"`
	Sum         string            `json:"sum"`
}

// NewFamily consumes a MetricFamily and transforms it to the local Family type.
func NewFamily(dtoMF *dto.MetricFamily) (string, *Family) {
	mf := &Family{
		//Time:    time.Now(),
		Name:    dtoMF.GetName(),
		Help:    dtoMF.GetHelp(),
		Type:    dtoMF.GetType().String(),
		Metrics: make([]interface{}, len(dtoMF.Metric)),
	}
	for i, m := range dtoMF.Metric {
		switch dtoMF.GetType() {
		case dto.MetricType_SUMMARY:
			mf.Metrics[i] = Summary{
				Labels:      makeLabels(m),
				TimestampMs: makeTimestamp(m),
				Quantiles:   makeQuantiles(m),
				Count:       fmt.Sprint(m.GetSummary().GetSampleCount()),
				Sum:         fmt.Sprint(m.GetSummary().GetSampleSum()),
			}
		case dto.MetricType_HISTOGRAM:
			mf.Metrics[i] = Histogram{
				Labels:      makeLabels(m),
				TimestampMs: makeTimestamp(m),
				Buckets:     makeBuckets(m),
				Count:       fmt.Sprint(m.GetHistogram().GetSampleCount()),
				Sum:         fmt.Sprint(m.GetHistogram().GetSampleSum()),
			}
		default:
			mf.Metrics[i] = Metric{
				Labels:      makeLabels(m),
				TimestampMs: makeTimestamp(m),
				Value:       fmt.Sprint(getValue(m)),
			}
		}
	}
	return mf.Name, mf
}

// NewMetricPointList consumes a MetricFamily and transforms it to the local []MetricPoint type.
func NewMetricPointList(dtoMF *dto.MetricFamily, addLabel map[string]string) []global.MetricPoint {
	now := time.Now().Unix()
	tsList := []global.MetricPoint{}

	for _, m := range dtoMF.Metric {
		switch dtoMF.GetType() {
		case dto.MetricType_SUMMARY:
			if len(makeQuantiles(m)) > 0 {
				for k, v := range makeQuantiles(m) {
					float64V, _ := strconv.ParseFloat(v, 64)

					mp := global.MetricPoint{
						Metric:   dtoMF.GetName(),
						LabelMap: makeLabels(m),
						Time:     now,
						Value:    float64V,
					}
					mp.LabelMap["quantile"] = k

					if len(addLabel) > 0 {
						for addKey, addValue := range addLabel {
							mp.LabelMap[addKey] = addValue
						}
					}

					tsList = append(tsList, mp)
				}
			}

		case dto.MetricType_HISTOGRAM:
			//todo
		default:
			mp := global.MetricPoint{
				Metric:   dtoMF.GetName(),
				LabelMap: makeLabels(m),
				Time:     now,
				Value:    getValue(m),
			}

			if len(addLabel) > 0 {
				for addKey, addValue := range addLabel {
					mp.LabelMap[addKey] = addValue
				}
			}

			tsList = append(tsList, mp)
		}
	}

	return tsList
}

func getValue(m *dto.Metric) float64 {
	switch {
	case m.Gauge != nil:
		return m.GetGauge().GetValue()
	case m.Counter != nil:
		return m.GetCounter().GetValue()
	case m.Untyped != nil:
		return m.GetUntyped().GetValue()
	default:
		return 0.
	}
}

func makeLabels(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, lp := range m.Label {
		result[lp.GetName()] = lp.GetValue()
	}
	return result
}

func makeTimestamp(m *dto.Metric) string {
	if m.TimestampMs == nil {
		return ""
	}
	return fmt.Sprint(m.GetTimestampMs())
}

func makeQuantiles(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, q := range m.GetSummary().Quantile {
		result[fmt.Sprint(q.GetQuantile())] = fmt.Sprint(q.GetValue())
	}
	return result
}

func makeBuckets(m *dto.Metric) map[string]string {
	result := map[string]string{}
	for _, b := range m.GetHistogram().Bucket {
		result[fmt.Sprint(b.GetUpperBound())] = fmt.Sprint(b.GetCumulativeCount())
	}
	return result
}

// FetchMetricFamilies retrieves metrics from the provided URL, decodes them
// into MetricFamily proto messages, and sends them to the provided channel. It
// returns after all MetricFamilies have been sent. The provided transport
// may be nil (in which case the default Transport is used).
func FetchMetricFamilies(url string, ch chan<- *dto.MetricFamily, transport http.RoundTripper) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		close(ch)
		return fmt.Errorf("creating GET request for URL %q failed: %v", url, err)
	}
	req.Header.Add("Accept", acceptHeader)
	client := http.Client{Transport: transport}
	resp, err := client.Do(req)
	if err != nil {
		close(ch)
		return fmt.Errorf("executing GET request for URL %q failed: %v", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		close(ch)
		return fmt.Errorf("GET request for URL %q returned HTTP status %s", url, resp.Status)
	}
	return ParseResponse(resp, ch)
}

// ParseResponse consumes an http.Response and pushes it to the MetricFamily
// channel. It returns when all MetricFamilies are parsed and put on the
// channel.
func ParseResponse(resp *http.Response, ch chan<- *dto.MetricFamily) error {
	mediatype, params, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err == nil && mediatype == "application/vnd.google.protobuf" &&
		params["encoding"] == "delimited" &&
		params["proto"] == "io.prometheus.client.MetricFamily" {
		defer close(ch)
		for {
			mf := &dto.MetricFamily{}
			if _, err = pbutil.ReadDelimited(resp.Body, mf); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("reading metric family protocol buffer failed: %v", err)
			}
			ch <- mf
		}
	} else {
		if err := ParseReader(resp.Body, ch); err != nil {
			return err
		}
	}
	return nil
}

// ParseReader consumes an io.Reader and pushes it to the MetricFamily
// channel. It returns when all MetricFamilies are parsed and put on the
// channel.
func ParseReader(in io.Reader, ch chan<- *dto.MetricFamily) error {
	defer close(ch)
	// We could do further content-type checks here, but the
	// fallback for now will anyway be the text format
	// version 0.0.4, so just go for it and see if it works.
	var parser expfmt.TextParser
	metricFamilies, err := parser.TextToMetricFamilies(in)
	if err != nil {
		return fmt.Errorf("reading text format failed: %v", err)
	}
	for _, mf := range metricFamilies {
		ch <- mf
	}
	return nil
}

// AddLabel allows to add key/value labels to an already existing Family.
func (f *Family) AddLabel(key, val string) {
	for i, item := range f.Metrics {
		switch m := item.(type) {
		case Metric:
			m.Labels[key] = val
			f.Metrics[i] = m
		}
	}
}

func makeTransport(certificate string, key string, skipServerCertCheck bool) (*http.Transport, error) {
	// Start with the DefaultTransport for sane defaults.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	// Conservatively disable HTTP keep-alives as this program will only
	// ever need a single HTTP request.
	transport.DisableKeepAlives = true
	// Timeout early if the server doesn't even return the headers.
	transport.ResponseHeaderTimeout = time.Minute
	tlsConfig := &tls.Config{InsecureSkipVerify: skipServerCertCheck}
	if certificate != "" && key != "" {
		cert, err := tls.LoadX509KeyPair(certificate, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}

// GetProm2JsonMapStruct get exporter info and parsing into Family struct return map data
func GetProm2JsonMapStruct(exporter_url string) map[string]*Family {
	transport, err := makeTransport("", "", false)
	if err != nil {
		global.LogObj.Error(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mfChan := make(chan *dto.MetricFamily, 1024)

	go func() {
		err := FetchMetricFamilies(exporter_url, mfChan, transport)
		if err != nil {
			global.LogObj.Error(err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	result := map[string]*Family{}
	for mf := range mfChan {
		metricName, metricObj := NewFamily(mf)
		result[metricName] = metricObj
	}

	//jsonText, err := json.Marshal(result)
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "error marshaling JSON:", err)
	//	//os.Exit(1)
	//}
	//fmt.Printf("exporter json:%v\n", string(jsonText))

	return result
}

// GetProm2JsonStruct get exporter info and parsing into Family struct return slice data
func GetProm2JsonStruct(exporter_url string) []*Family {
	transport, err := makeTransport("", "", false)
	if err != nil {
		global.LogObj.Error(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mfChan := make(chan *dto.MetricFamily, 1024)

	go func() {
		err := FetchMetricFamilies(exporter_url, mfChan, transport)
		if err != nil {
			global.LogObj.Error(err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	result := []*Family{}
	for mf := range mfChan {
		_, metricObj := NewFamily(mf)
		result = append(result, metricObj)
	}

	//jsonText, err := json.Marshal(result)
	//if err != nil {
	//	fmt.Fprintln(os.Stderr, "error marshaling JSON:", err)
	//	//os.Exit(1)
	//}
	//fmt.Printf("exporter json:%v\n", string(jsonText))

	return result
}

// GetProm2MetricPointList get exporter info and parsing into MetricPoint struct return slice data
func GetProm2MetricPointList(exporter_url string, adddLabel map[string]string) []global.MetricPoint {
	transport, err := makeTransport("", "", false)
	if err != nil {
		global.LogObj.Error(err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	mfChan := make(chan *dto.MetricFamily, 1024)

	go func() {
		err := FetchMetricFamilies(exporter_url, mfChan, transport)
		if err != nil {
			global.LogObj.Error(err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	result := []global.MetricPoint{}
	for mf := range mfChan {
		metricObj := NewMetricPointList(mf, adddLabel)
		if len(metricObj) > 0 {
			for _, point := range metricObj {
				result = append(result, point)
			}
		}
	}

	return result
}

type TransFormGather struct {
	exporter_url string
}

func NewTransFormGather(url string) *TransFormGather {
	return &TransFormGather{exporter_url: url}
}

// Gather get metric info from exporter returned MetricFamily protobufs
func (t TransFormGather) Gather() ([]*dto.MetricFamily, error) {
	transport, err := makeTransport("", "", false)
	if err != nil {
		global.LogObj.Error(err)
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	mfChan := make(chan *dto.MetricFamily, 1024)

	go func() {
		err := FetchMetricFamilies(t.exporter_url, mfChan, transport)
		if err != nil {
			global.LogObj.Error(err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	result := []*dto.MetricFamily{}
	for mf := range mfChan {
		result = append(result, mf)
	}

	return result, nil
}
