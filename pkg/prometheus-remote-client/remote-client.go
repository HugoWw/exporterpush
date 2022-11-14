package prometheus_remote_client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/exporterpush/global"
	"github.com/prometheus/common/model"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/prompb"
)

const (
	// DefaultRemoteWrite is the default Prom remote write endpoint in prometheus server.
	DefaultRemoteWrite = "http://127.0.0.1:9091/api/v1/write"

	defaulHTTPClientTimeout = 30 * time.Second
	defaultUserAgent        = "promremote-go/1.0.0"

	LABEL_NAME = "__name__"
)

// DefaultConfig represents the default configuration used to construct a client.
var DefaultConfig = Config{
	WriteURL:          DefaultRemoteWrite,
	HTTPClientTimeout: defaulHTTPClientTimeout,
	UserAgent:         defaultUserAgent,
}

/*
TSList is the metric type converted to prompb.WriteRequest
*/

// Label is a metric label.
type Label struct {
	Name  string
	Value string
}

// A Datapoint is a single data value reported at a given time.
type Datapoint struct {
	Timestamp time.Time
	Value     float64
}

// TimeSeries are made of labels and a datapoint.
type TimeSeries struct {
	Labels    []Label
	Datapoint Datapoint
}

// TSList is a slice of TimeSeries.
type TSList []TimeSeries

/*
MetricPoint is second type to converted metric data to prompb.WriteRequest
*/

var MetricNameRE = regexp.MustCompile(`^[a-zA-Z_:][a-zA-Z0-9_:]*$`)

// MetricPointList is a slice of MetricPoint.
type MetricPointList []global.MetricPoint

type sample struct {
	labels labels.Labels
	t      int64
	v      float64
}

// Client is used to write timeseries data to a Prom remote write endpoint
// such as the one in m3coordinator.
type Client interface {
	// WriteProto writes the Prom proto WriteRequest to the specified endpoint.
	WriteProto(ctx context.Context, req *prompb.WriteRequest, opts WriteOptions) (WriteResult, WriteError)

	// WriteTimeSeries converts the []TimeSeries to Protobuf then writes it to the specified endpoint.
	WriteTimeSeries(ctx context.Context, ts TSList, opts WriteOptions) (WriteResult, WriteError)

	// WriteMetricPointList converts the []MetricPoint to Protobuf then writes it to the specified endpoint.
	WriteMetricPointList(ctx context.Context, metricPointList MetricPointList, opts WriteOptions) (WriteResult, WriteError)
}

// WriteOptions specifies additional write options.
type WriteOptions struct {
	// Headers to append or override the outgoing headers.
	Headers map[string]string
}

// WriteResult returns the successful HTTP status code.
type WriteResult struct {
	StatusCode int
}

// WriteError is an error that can also return the HTTP status code
// if the response is what caused an error.
type WriteError interface {
	error
	StatusCode() int
}

// Config defines the configuration used to construct a client.
type Config struct {
	// WriteURL is the URL which the client uses to write to m3coordinator.
	WriteURL string

	//HTTPClientTimeout is the timeout that is set for the client.
	HTTPClientTimeout time.Duration

	// If not nil, http client is used instead of constructing one.
	HTTPClient *http.Client

	// UserAgent is the `User-Agent` header in the request.
	UserAgent string
}

// ConfigOption defines a config option that can be used when constructing a client.
type ConfigOption func(*Config)

// NewConfig creates a new Config struct based on options passed to the function.
func NewConfig(opts ...ConfigOption) Config {
	cfg := DefaultConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

func (c Config) validate() error {
	if c.HTTPClientTimeout <= 0 {
		return fmt.Errorf("http client timeout should be greater than 0: %d", c.HTTPClientTimeout)
	}

	if c.WriteURL == "" {
		return errors.New("remote write URL should not be blank")
	}

	if c.UserAgent == "" {
		return errors.New("User-Agent should not be blank")
	}

	return nil
}

// WriteURLOption sets the URL which the client uses to write to m3coordinator.
func WriteURLOption(writeURL string) ConfigOption {
	return func(c *Config) {
		c.WriteURL = writeURL
	}
}

// HTTPClientTimeoutOption sets the timeout that is set for the client.
func HTTPClientTimeoutOption(httpClientTimeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.HTTPClientTimeout = httpClientTimeout
	}
}

// HTTPClientOption sets the HTTP client that is set for the client.
func HTTPClientOption(httpClient *http.Client) ConfigOption {
	return func(c *Config) {
		c.HTTPClient = httpClient
	}
}

// UserAgent sets the `User-Agent` header in the request.
func UserAgent(userAgent string) ConfigOption {
	return func(c *Config) {
		c.UserAgent = userAgent
	}
}

type client struct {
	writeURL   string
	httpClient *http.Client
	userAgent  string
}

// NewClient creates a new remote write coordinator client.
func NewClient(c Config) (Client, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: c.HTTPClientTimeout,
	}

	if c.HTTPClient != nil {
		httpClient = c.HTTPClient
	}

	return &client{
		writeURL:   c.WriteURL,
		httpClient: httpClient,
	}, nil
}

func (c *client) WriteMetricPointList(ctx context.Context, metricPointList MetricPointList, opts WriteOptions) (WriteResult, WriteError) {

	var result WriteResult

	err, promWR := metricPointList.convertMetricPointToWriteRequest()
	if err != nil {
		return result, writeError{err: fmt.Errorf("convert metricPoint data to writeRequest error: %v", err)}
	}

	return c.WriteProto(ctx, promWR, opts)
}

func (c *client) WriteTimeSeries(ctx context.Context, seriesList TSList, opts WriteOptions) (WriteResult, WriteError) {
	return c.WriteProto(ctx, seriesList.toPromWriteRequest(), opts)
}

func (c *client) WriteProto(ctx context.Context, promWR *prompb.WriteRequest, opts WriteOptions) (WriteResult, WriteError) {
	var result WriteResult
	data, err := proto.Marshal(promWR)
	if err != nil {
		return result, writeError{err: fmt.Errorf("unable to marshal protobuf: %v", err)}
	}

	encoded := snappy.Encode(nil, data)

	body := bytes.NewReader(encoded)
	req, err := http.NewRequest("POST", c.writeURL, body)
	if err != nil {
		return result, writeError{err: err}
	}

	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")
	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return result, writeError{err: err}
	}

	result.StatusCode = resp.StatusCode

	defer resp.Body.Close()

	if result.StatusCode/100 != 2 {
		writeErr := writeError{
			err:  fmt.Errorf("expected HTTP 200 status code: actual=%d", resp.StatusCode),
			code: result.StatusCode,
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			writeErr.err = fmt.Errorf("%v, body_read_error=%s", writeErr.err, err)
			return result, writeErr
		}

		writeErr.err = fmt.Errorf("%v, body=%s", writeErr.err, body)
		return result, writeErr
	}

	return result, nil
}

// toPromWriteRequest converts a list of timeseries to a Prometheus proto write request.
func (t TSList) toPromWriteRequest() *prompb.WriteRequest {
	promTS := make([]prompb.TimeSeries, len(t))

	for i, ts := range t {
		labels := make([]prompb.Label, len(ts.Labels))
		for j, label := range ts.Labels {
			labels[j] = prompb.Label{Name: label.Name, Value: label.Value}
		}

		sample := []prompb.Sample{prompb.Sample{
			// Timestamp is int milliseconds for remote write.
			Timestamp: ts.Datapoint.Timestamp.UnixNano() / int64(time.Millisecond),
			Value:     ts.Datapoint.Value,
		}}
		promTS[i] = prompb.TimeSeries{Labels: labels, Samples: sample}
	}

	return &prompb.WriteRequest{
		Timeseries: promTS,
	}
}

type writeError struct {
	err  error
	code int
}

func (e writeError) Error() string {
	return e.err.Error()
}

// StatusCode returns the HTTP status code of the error if error
// was caused by the response, otherwise it will be just zero.
func (e writeError) StatusCode() int {
	return e.code
}

// convertToWriteRequest convert []MetricPoint data to WriteRequest
func (items MetricPointList) convertMetricPointToWriteRequest() (err error, writeReq *prompb.WriteRequest) {
	if len(items) == 0 {
		return
	}
	ts := make([]prompb.TimeSeries, len(items))
	for i := range items {
		ts[i], err = convertPromTimeSeries(items[i])
		if err != nil {
			return err, nil
		}
	}

	return nil, &prompb.WriteRequest{Timeseries: ts}
}

func convertPromTimeSeries(item global.MetricPoint) (prompb.TimeSeries, error) {
	pt := prompb.TimeSeries{}
	pt.Samples = []prompb.Sample{{}}
	s := sample{}
	s.t = item.Time
	s.v = item.Value
	// name
	if !MetricNameRE.MatchString(item.Metric) {
		return pt, errors.New("invalid metrics name")
	}
	nameLs := labels.Label{
		Name:  LABEL_NAME,
		Value: item.Metric,
	}
	s.labels = append(s.labels, nameLs)
	for k, v := range item.LabelMap {
		if model.LabelNameRE.MatchString(k) {
			ls := labels.Label{
				Name:  k,
				Value: v,
			}
			s.labels = append(s.labels, ls)
		}
	}

	pt.Labels = labelsToLabelsProto(s.labels, pt.Labels)
	// 时间赋值问题,使用毫秒时间戳
	tsMs := time.Unix(s.t, 0).UnixNano() / 1e6
	pt.Samples[0].Timestamp = tsMs
	pt.Samples[0].Value = s.v
	return pt, nil
}

func labelsToLabelsProto(labels labels.Labels, buf []prompb.Label) []prompb.Label {
	result := buf[:0]
	if cap(buf) < len(labels) {
		result = make([]prompb.Label, 0, len(labels))
	}
	for _, l := range labels {
		result = append(result, prompb.Label{
			Name:  l.Name,
			Value: l.Value,
		})
	}
	return result
}
