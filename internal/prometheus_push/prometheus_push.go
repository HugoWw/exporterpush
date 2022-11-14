package prometheus_push

import (
	"context"
	"github.com/exporterpush/global"
	"github.com/exporterpush/pkg/prom2json"
	promclient "github.com/exporterpush/pkg/prometheus-remote-client"
	"github.com/exporterpush/pkg/util"
	"time"
)

func PrometheusPush() {
	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	intervalTime := time.Duration(global.GlobalSetting.ScrapeInterval) * time.Second
	ticker := time.NewTicker(intervalTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go push()
		}
	}

}

func push() {
	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	prometheusSerAdd := global.PrometheusSetting.StaticConfigs[0].Destination[0]

	cfg := promclient.NewConfig(promclient.WriteURLOption(prometheusSerAdd))
	remoteWriteClient, err := promclient.NewClient(cfg)
	if err != nil {
		global.LogObj.Errorf("new prometheus remote write client error:%v", err)
		return
	}

	addLabel := global.PrometheusSetting.StaticConfigs[0].Labels
	metricPointList := prom2json.GetProm2MetricPointList(global.GlobalSetting.ScrapeTargetTypes.NodeExporter, addLabel)

	_, writeErr := remoteWriteClient.WriteMetricPointList(context.Background(), metricPointList, promclient.WriteOptions{})
	if writeErr != nil {
		global.LogObj.Errorf("remote write to prometheus server %v error:%v", prometheusSerAdd, writeErr.Error())
		return
	}

	global.LogObj.Infof("remote write to %v success", prometheusSerAdd)
}
