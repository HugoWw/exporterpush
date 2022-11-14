package pushgateway_push

import (
	"github.com/exporterpush/global"
	"github.com/exporterpush/pkg/prom2json"
	"github.com/exporterpush/pkg/util"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"time"
)

func PushGatewayPush() {
	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	intervalTime := time.Duration(global.GlobalSetting.ScrapeInterval) * time.Second
	ticker := time.NewTicker(intervalTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			gather := prom2json.NewTransFormGather(global.GlobalSetting.ScrapeTargetTypes.NodeExporter)

			if len(global.PushgatewaySetting.StaticConfigs[0].Destination) >= 1 {
				for key, destPushGateway := range global.PushgatewaySetting.StaticConfigs[0].Destination {
					go pushInfo(key, destPushGateway, gather)
				}
			} else {
				global.LogObj.Error("There is no push target when use PushGatewayPush")
			}

		}
	}

}

func pushInfo(numb int, dest string, g prometheus.Gatherer) {
	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	var job_name string

	if global.PushgatewaySetting.JobnName != "" {
		job_name = global.PushgatewaySetting.JobnName
	} else {
		job_name = "exporter_push"
	}

	push := push.New(dest, job_name)

	if len(global.PushgatewaySetting.StaticConfigs[0].Labels) > 0 {
		for labelKey, labelValue := range global.PushgatewaySetting.StaticConfigs[0].Labels {
			push.Grouping(labelKey, labelValue)
		}
	}

	if err := push.Gatherer(g).Push(); err != nil {
		global.LogObj.Errorf("PushGatewayPush goroutine %v Could not push to PushGateway %v,error:%v", numb, dest, err)
	} else {
		global.LogObj.Infof("PushGatewayPush goroutine %v push monitor info to PushGateway %v success !",
			numb, dest)
	}
}
