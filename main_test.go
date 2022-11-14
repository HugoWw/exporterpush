package main

import (
	"fmt"
	_ "github.com/exporterpush/config"
	"github.com/exporterpush/global"
	"github.com/exporterpush/internal/model"
	"github.com/exporterpush/pkg/prom2json"
	"strconv"
	"testing"
	"time"
)

func clickHouseCalc(metric *prom2json.Family) float64 {

	gaugeOrCounter := prom2json.Metric{}
	if metric.Type == "GAUGE" || metric.Type == "COUNTER" {
		gaugeOrCounter = metric.Metrics[0].(prom2json.Metric)
	}
	valueInt, _ := strconv.ParseFloat(gaugeOrCounter.Value, 64)

	return valueInt
}

func TestFunc(t *testing.T) {
	metric := prom2json.GetProm2JsonMapStruct("http://127.0.0.1:9363/metrics")
	//clickhouseTcpCon := metric[global.ClickhouseTcpConnection]
	//tcp_valueMap, ok := clickhouseTcpCon.Metrics[0].(map[string]string)
	//if ok {
	//	fmt.Printf("interface{} to map ok:%v\n", tcp_valueMap)
	//}
	//tcp_valueInt, _ := strconv.Atoi(tcp_valueMap["value"])
	//tcp_connect := model.Batchs{
	//	Unit:  "count",
	//	Name:  "tcp_connection",
	//	Value: tcp_valueInt,
	//}

	clickhouseTcpCon := metric[global.ClickhouseTcpConnection]

	tcp_connect := model.Batchs{
		Unit:  "count",
		Name:  "tcp_connection",
		Value: clickHouseCalc(clickhouseTcpCon),
	}

	fmt.Printf("tcp conn:%+v\n", tcp_connect)

}

func TestMainFunc(t *testing.T) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	//Oldmp := map[string]Test{}
	//Oldmp["old"] = Test{name: "oldname"}
	//fmt.Printf("oldmap info:%+v\n", Oldmp)

	//oldtest := &Test{name: "old test"}

	Oldmp := map[string]*Test{}
	Oldmp["old"] = &Test{name: "oldname"}
	fmt.Printf("oldmap info:%+v\n", Oldmp["old"])

	count := 0

	for {
		select {
		case <-ticker.C:
			//fmt.Printf("befor newmap oldmap info:%+v\n", Oldmp)
			//NewMap(&Oldmp, "new-"+strconv.Itoa(count))
			//fmt.Printf("after newmap oldmap info:%+v\n", Oldmp)

			//fmt.Printf("befor NewTestStrc oldtest info:%+v\n", oldtest)
			//NewTestStrc(oldtest, "new-"+strconv.Itoa(count))
			//fmt.Printf("after NewTestStrc oldtest info:%+v\n", oldtest)

			fmt.Printf("befor NewTestStrcMap oldtest info:%+v\n", Oldmp["old"])
			NewTestStrcMap(&Oldmp, "new-"+strconv.Itoa(count))
			fmt.Printf("after NewTestStrcMap oldtest info:%+v\n", Oldmp["old"])

			count++
			if count == 2 {
				return
			}
		}
	}

}

type Test struct {
	name string
}

func NewMap(oldmap *map[string]Test, names string) {

	newmap := map[string]Test{}
	newmap["new"] = Test{name: names}

	//oldmap = &newmap
	(*oldmap)["old"] = newmap["new"]
}

func NewTestStrc(oldtest *Test, names string) {
	newtest := &Test{name: names}

	*oldtest = *newtest
}

func NewTestStrcMap(oldmap *map[string]*Test, names string) {
	newmap := map[string]*Test{}
	newmap["new"] = &Test{name: names}

	//oldmap = &newmap
	(*oldmap)["old"] = newmap["new"]
}
