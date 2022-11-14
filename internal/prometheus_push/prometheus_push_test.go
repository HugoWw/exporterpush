package prometheus_push

import (
	"context"
	"github.com/exporterpush/global"
	promclient "github.com/exporterpush/pkg/prometheus-remote-client"
	"testing"
	"time"
)

func TestRemoteWrite(t *testing.T) {

	metrics := []global.MetricPoint{
		{Metric: "test_metric_1",
			LabelMap: map[string]string{"env": "testing", "op": "test1"},
			Time:     time.Now().Add(-1 * time.Minute).Unix(),
			Value:    1},
		{Metric: "test_metric_1",
			LabelMap: map[string]string{"env": "testing", "op": "test1"},
			Time:     time.Now().Add(-2 * time.Minute).Unix(),
			Value:    2},
		{Metric: "test_metric_2",
			LabelMap: map[string]string{"env": "testing", "op": "test2"},
			Time:     time.Now().Unix(),
			Value:    3},
		{Metric: "test_metric_3",
			LabelMap: map[string]string{"env": "testing", "op": "test3"},
			Time:     time.Now().Unix(),
			Value:    4},
	}

	cfg := promclient.NewConfig(promclient.WriteURLOption("http://127.0.0.1:9090/api/v1/write"))
	remoteWriteClient, err := promclient.NewClient(cfg)
	if err != nil {
		global.LogObj.Errorf("new prometheus remote write client error:%v", err)
		return
	}

	_, writeErr := remoteWriteClient.WriteMetricPointList(context.Background(), metrics, promclient.WriteOptions{})
	if writeErr != nil {
		global.LogObj.Errorf("remote write to prometheus server error:%v", writeErr.Error())
		return
	}

	// or use promclient.TSList struct for remote write
	tsList := promclient.TSList{
		{
			Labels: []promclient.Label{
				{
					Name:  "__name__",
					Value: "foo_bar",
				},
				{
					Name:  "biz",
					Value: "baz",
				},
			},
			Datapoint: promclient.Datapoint{
				Timestamp: time.Now(),
				Value:     1415.92,
			},
		},
	}

	remoteWriteClient.WriteTimeSeries(context.Background(), tsList, promclient.WriteOptions{})

}
