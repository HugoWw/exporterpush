package pushgateway_push

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"testing"
)

func TestPushgateway(t *testing.T) {
	metric1 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "db_backup_last_completion_timestamp_seconds",
		Help: "The timestamp of the last successful completion of a DB backup.",
	})
	metric1.SetToCurrentTime() // set可以设置任意值（float64）

	metric2 := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "testname2",
		Help:        "testhelp2",
		ConstLabels: prometheus.Labels{"foo": "bar", "dings": "bums"},
	})

	reg := prometheus.NewRegistry()
	reg.MustRegister(metric1)
	reg.MustRegister(metric2)

	if err := push.New("http://127.0.0.1:9091", "db_backup"). // push.New("pushgateway地址", "job名称")
									Collector(metric1).                                          // Collector(completionTime) 给指标赋值
									Grouping("db", "customers").Grouping("instance", "1.1.1.1"). // 给指标添加标签，可以添加多个
									Push(); err != nil {
		fmt.Println("Could not push to Pushgateway:", err)
	}

	// or use Gatherer push metric info
	//if err := push.New("http://127.0.0.1:9091", "testjob").
	//	Gatherer(reg).
	//	Push(); err != nil {
	//	fmt.Println("Could not push to Pushgateway:", err)
	//}
}
