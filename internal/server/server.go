package server

import (
	"github.com/exporterpush/global"
	"github.com/exporterpush/internal/barad_ck_push"
	"github.com/exporterpush/internal/prometheus_push"
	"github.com/exporterpush/internal/pushgateway_push"
)

var services []func()

func registry(f func(), name string) {
	services = append(services, f)
	global.LogObj.Infof("registry service handler:%v", name)
}

func Run() {

	if services != nil {
		for _, f := range services {
			go f()
		}
	} else {
		global.LogObj.Warnf("no service to run，services Handler is nil")
	}

}

func init() {
	if global.BaradSetting.IsUse {
		registry(barad_ck_push.BaradCKPush, "BaradClickhousePush")
	}

	if global.PushgatewaySetting.IsUse {
		registry(pushgateway_push.PushGatewayPush, "PushGatewayPush")
	}

	if global.PrometheusSetting.IsUse {
		registry(prometheus_push.PrometheusPush, "PrometheusPush")
	}
}
