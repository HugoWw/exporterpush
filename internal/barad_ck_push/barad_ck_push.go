package barad_ck_push

import (
	"encoding/json"
	"fmt"
	"github.com/exporterpush/global"
	"github.com/exporterpush/internal/model"
	"github.com/exporterpush/internal/node_calc"
	"github.com/exporterpush/pkg/prom2json"
	"github.com/exporterpush/pkg/util"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/net"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func BaradCKPush() {

	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	intervalTime := time.Duration(global.GlobalSetting.ScrapeInterval) * time.Second
	ticker := time.NewTicker(intervalTime)
	defer ticker.Stop()

	oldDiskIOInfo, err := node_calc.GetDiskRWAndIO(global.GlobalSetting.Disk)
	if err != nil {
		global.LogObj.Errorf("init get node disk read and write info error: %v", err)
	}
	oldNetInfo, err := node_calc.GetNetWorkInfo(global.GlobalSetting.NetInterface)
	if err != nil {
		global.LogObj.Errorf("init get node network info error: %v", err)
	}

	oldFamilyMetric := prom2json.GetProm2JsonMapStruct(global.GlobalSetting.ScrapeTargetTypes.ClickhouseExporter)

	for {
		select {
		case <-ticker.C:
			requestInfo := BaradCKCalc(&oldDiskIOInfo, oldNetInfo, &oldFamilyMetric)
			if len(requestInfo.Batch) <= 0 {
				global.LogObj.Error("init barad request struct batch is nil")
				continue
			}
			reqBodyBty, _ := json.Marshal(requestInfo)
			fmt.Println(string(reqBodyBty))

			// request barad
			reqBody := strings.NewReader(string(reqBodyBty))
			req, err := http.NewRequest("POST", global.BaradSetting.StaticConfigs[0].Destination[0], reqBody)
			if err != nil {
				global.LogObj.Errorf("init barad request error:%v", err)
				continue
			}
			req.Header.Add("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(req)
			if err != nil {
				global.LogObj.Errorf("request barad error:%v", err)
				continue
			}
			respBodyStr, _ := ioutil.ReadAll(response.Body)
			if response != nil && response.StatusCode != 200 {
				global.LogObj.Errorf("push monitor info failed response code:%v, info:%v", response.StatusCode, string(respBodyStr))
				response.Body.Close()
				continue
			}
			global.LogObj.Infof("response from barad info:%v", string(respBodyStr))
			global.LogObj.Infof("post monitor info:%v", respBodyStr)

			response.Body.Close()
		}
	}

}

func BaradCKCalc(old_disck_io_info *map[string]disk.IOCountersStat, old_net_info *net.IOCountersStat,
	old_family_metric *map[string]*prom2json.Family) model.BaradCk {

	defer util.CatchException(func(e interface{}) {
		global.LogObj.Panic(e)
	})

	newDiskIOInfo, err := node_calc.GetDiskRWAndIO(global.GlobalSetting.Disk)
	if err != nil {
		global.LogObj.Errorf("get node disk read and write info error: %v", err)
	}
	newNetInfo, err := node_calc.GetNetWorkInfo(global.GlobalSetting.NetInterface)
	if err != nil {
		global.LogObj.Errorf("get node network info error: %v", err)
	}
	newFamilyMetric := prom2json.GetProm2JsonMapStruct(global.GlobalSetting.ScrapeTargetTypes.ClickhouseExporter)

	perSecInfo, err := node_calc.GetPerSecondMetric()
	if err != nil {
		global.LogObj.Errorf("get node base info error: %v", err)
	}

	barad := model.BaradCk{
		Timestamp: int(time.Now().Unix()),
		Namespace: global.BaradSetting.Namespace,
		Dimension: model.Dimensions{
			AppId:      global.BaradSetting.AppId,
			InstanceId: global.BaradSetting.InstanceId,
			NodeId:     global.BaradSetting.NodeId,
			ProjectId:  global.BaradSetting.ProjectId,
		},
	}

	metricList := []model.Batchs{}
	timeScrapeInterval_float := float64(global.GlobalSetting.ScrapeInterval)
	scrapeIntervalTime_unit64 := uint64(global.GlobalSetting.ScrapeInterval)

	/*
		node base monitor info(cpu、memory、disk、network)
	*/

	// cpu usage percent
	cpu_use_rate := model.Batchs{
		Unit:  "%",
		Name:  "cpu_use_rate",
		Value: perSecInfo.CpuUsagePercent,
	}

	// memory usage info
	real_mem_use := model.Batchs{
		Unit:  "MBytes",
		Name:  "real_mem_use",
		Value: float64(perSecInfo.MemoryUseState.Used / 1024 / 1024),
	}

	mem_use_rate := model.Batchs{
		Unit:  "%",
		Name:  "mem_use_rate",
		Value: perSecInfo.MemoryUsedPercent,
	}
	metricList = append(metricList, cpu_use_rate, real_mem_use, mem_use_rate)

	// disk storage usage info
	disk_use_size := model.Batchs{
		Unit:  "MBytes",
		Name:  "disk_use_size",
		Value: float64(perSecInfo.DiskUseStat.Used / 1024 / 1024),
	}

	disk_use_rate := model.Batchs{
		Unit:  "%",
		Name:  "disk_use_rate",
		Value: util.Decimal(perSecInfo.DiskUseStat.UsedPercent, 2),
	}

	inode_use_rate := model.Batchs{
		Unit:  "%",
		Name:  "inode_use_rate",
		Value: util.Decimal(perSecInfo.DiskUseStat.InodesUsedPercent, 2),
	}
	metricList = append(metricList, disk_use_size, disk_use_rate, inode_use_rate)

	// disk read write bytes and read write io
	disk_read_bytes := newDiskIOInfo[global.GlobalSetting.Disk].ReadBytes - (*old_disck_io_info)[global.GlobalSetting.Disk].ReadBytes
	disk_read_bytes_tmp := float64(disk_read_bytes / scrapeIntervalTime_unit64)
	io_read_bytes := model.Batchs{
		Unit:  "Bytes",
		Name:  "io_read_bytes",
		Value: util.Decimal(disk_read_bytes_tmp, 2),
	}

	disk_write_bytes := newDiskIOInfo[global.GlobalSetting.Disk].WriteBytes - (*old_disck_io_info)[global.GlobalSetting.Disk].WriteBytes
	disk_write_bytes_tmp := float64(disk_write_bytes / scrapeIntervalTime_unit64)
	io_write_bytes := model.Batchs{
		Unit:  "Bytes",
		Name:  "io_write_bytes",
		Value: util.Decimal(disk_write_bytes_tmp, 2),
	}

	read_iops := newDiskIOInfo[global.GlobalSetting.Disk].ReadCount - (*old_disck_io_info)[global.GlobalSetting.Disk].ReadCount
	read_iops_tmp := float64(read_iops / scrapeIntervalTime_unit64)
	disk_read_iops := model.Batchs{
		Unit:  "count",
		Name:  "disk_read_iops",
		Value: util.Decimal(read_iops_tmp, 2),
	}

	write_iops := newDiskIOInfo[global.GlobalSetting.Disk].WriteCount - (*old_disck_io_info)[global.GlobalSetting.Disk].WriteCount
	write_iops_tmp := float64(write_iops / scrapeIntervalTime_unit64)
	disk_write_iops := model.Batchs{
		Unit:  "count",
		Name:  "disk_write_iops",
		Value: util.Decimal(write_iops_tmp, 2),
	}
	metricList = append(metricList, io_read_bytes, io_write_bytes, disk_read_iops, disk_write_iops)
	(*old_disck_io_info)[global.GlobalSetting.Disk] = newDiskIOInfo[global.GlobalSetting.Disk] // new data migration

	// network send and receive bytes
	net_send_bytes := newNetInfo.BytesSent - old_net_info.BytesSent
	net_send_bytes_tmp := float64(net_send_bytes / scrapeIntervalTime_unit64)
	network_send_bytes := model.Batchs{
		Unit:  "Bytes",
		Name:  "network_send_bytes",
		Value: util.Decimal(net_send_bytes_tmp, 2),
	}

	net_receive_bytes := newNetInfo.BytesRecv - old_net_info.BytesRecv
	net_receive_bytes_tmp := float64(net_receive_bytes / scrapeIntervalTime_unit64)
	network_receive_bytes := model.Batchs{
		Unit:  "Bytes",
		Name:  "network_receive_bytes",
		Value: util.Decimal(net_receive_bytes_tmp, 2),
	}
	metricList = append(metricList, network_send_bytes, network_receive_bytes)
	*old_net_info = *newNetInfo // new data migration

	/*
		clickhouse exporter metrics
	*/

	// clickhouse tcp connection
	if clickhouseTcpCon, ok := newFamilyMetric[global.ClickhouseTcpConnection]; ok {
		tcp_connect := model.Batchs{
			Unit:  "count",
			Name:  "tcp_connection",
			Value: clickHouseCalc(clickhouseTcpCon),
		}
		metricList = append(metricList, tcp_connect)
	}

	// clickhouse http connection
	if clickhouseHttpCon, ok := newFamilyMetric[global.ClickhouseHttpConnection]; ok {
		http_connection := model.Batchs{
			Unit:  "count",
			Name:  "http_connection",
			Value: clickHouseCalc(clickhouseHttpCon),
		}
		metricList = append(metricList, http_connection)
	}

	// clickhouse mysql connection
	if clickhouseMysqlCon, ok := newFamilyMetric[global.ClickhouseMysqlConnection]; ok {
		mysql_connection := model.Batchs{
			Unit:  "count",
			Name:  "mysql_connection",
			Value: clickHouseCalc(clickhouseMysqlCon),
		}
		metricList = append(metricList, mysql_connection)
	}

	// clickhouse internal server connection
	if clickhouseIntSvcCon, ok := newFamilyMetric[global.ClickhouseInterServerConnection]; ok {
		intSvc_connection := model.Batchs{
			Unit:  "count",
			Name:  "interserver_connection",
			Value: clickHouseCalc(clickhouseIntSvcCon),
		}
		metricList = append(metricList, intSvc_connection)
	}

	// clickhouse postgre sql connection
	if clickhousePgCon, ok := newFamilyMetric[global.ClickhousePostgreSQLConnection]; ok {
		postgresql_connection := model.Batchs{
			Unit:  "count",
			Name:  "postgresql_connection",
			Value: clickHouseCalc(clickhousePgCon),
		}
		metricList = append(metricList, postgresql_connection)
	}

	// clickhouse failed query count avg seconds
	if clickhouseFailQueryCount, ok := newFamilyMetric[global.ClickhouseFailedQueryCount]; ok {
		faileQueryCount_valueInt := clickHouseCalc(clickhouseFailQueryCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseFailedQueryCount])
		failed_query_count := model.Batchs{
			Unit:  "count",
			Name:  "failed_query_count",
			Value: util.Decimal(faileQueryCount_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, failed_query_count)
	}

	//clickhous query count avg seconds
	var queryCount_valueInt float64
	if clickhouseQueryCount, ok := newFamilyMetric[global.ClickhouseQueryCount]; ok {
		queryCount_valueInt = clickHouseCalc(clickhouseQueryCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseQueryCount])
		query_count := model.Batchs{
			Unit:  "count",
			Name:  "query_count",
			Value: util.Decimal(queryCount_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, query_count)
	}

	// clickhouse delay insert count avg seconds
	if clickhouseDelayedInsertsCount, ok := newFamilyMetric[global.ClickhouseDelayedInsertsCount]; ok {
		delayedInsertsCount_valueInt := clickHouseCalc(clickhouseDelayedInsertsCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseDelayedInsertsCount])
		delayed_inserts_count := model.Batchs{
			Unit:  "count",
			Name:  "delayed_inserts_count",
			Value: util.Decimal(delayedInsertsCount_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, delayed_inserts_count)
	}

	// clickhouse merge count avg seconds
	if clickhouseMergeCount, ok := newFamilyMetric[global.ClickhouseMergeCount]; ok {
		mergeCount_valueInt := clickHouseCalc(clickhouseMergeCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseMergeCount])
		merge_count := model.Batchs{
			Unit:  "count",
			Name:  "merge_count",
			Value: util.Decimal(mergeCount_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, merge_count)
	}

	// clickhouse mutations count avg seconds
	if replicatedPartMutationsCount, ok := newFamilyMetric[global.ClickhouseReplicatedPartMutationsCount]; ok {
		replicatedPartMutationsCount_valueInt := clickHouseCalc(replicatedPartMutationsCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseReplicatedPartMutationsCount])
		replicated_part_mutations_count := model.Batchs{
			Unit:  "count",
			Name:  "replicated_part_mutations_count",
			Value: replicatedPartMutationsCount_valueInt / timeScrapeInterval_float,
		}
		metricList = append(metricList, replicated_part_mutations_count)
	}

	// clickhouse inset rows count avg seconds
	if insertRowCount, ok := newFamilyMetric[global.ClickhouseInsertedRowsCount]; ok {
		insertRows_valueInt := clickHouseCalc(insertRowCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseInsertedRowsCount])
		inserted_rows_count := model.Batchs{
			Unit:  "count",
			Name:  "inserted_rows_count",
			Value: util.Decimal(insertRows_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, inserted_rows_count)
	}

	// clickhouse insert bytes avg seconds
	if insertBytCount, ok := newFamilyMetric[global.ClickhouseInsertedBytesSize]; ok {
		insertByt_valueInt := clickHouseCalc(insertBytCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseInsertedBytesSize])
		inserted_size_bytes := model.Batchs{
			Unit:  "count",
			Name:  "inserted_size_bytes",
			Value: util.Decimal(insertByt_valueInt/timeScrapeInterval_float, 2),
		}
		metricList = append(metricList, inserted_size_bytes)
	}

	// clickhouse qps avg seconds
	if queryCount_valueInt != 0 {
		qps_count := model.Batchs{
			Unit:  "count",
			Name:  "qps_count",
			Value: queryCount_valueInt / timeScrapeInterval_float,
		}
		metricList = append(metricList, qps_count)
	}

	// clickhouse tps avg seconds
	if insertQueryCount, ok := newFamilyMetric[global.ClickhouseTPS_InsertQueryCount]; ok {
		tps_valueInt := clickHouseCalc(insertQueryCount) - clickHouseCalc((*old_family_metric)[global.ClickhouseTPS_InsertQueryCount])
		tps_count := model.Batchs{
			Unit:  "count",
			Name:  "tps_count",
			Value: tps_valueInt / timeScrapeInterval_float,
		}
		metricList = append(metricList, tps_count)
	}

	// clickhouse data active part
	if dataPartGauge, ok := newFamilyMetric[global.ClickhouseDataPartsGauge]; ok {
		data_part_count := model.Batchs{
			Unit:  "count",
			Name:  "data_part_count",
			Value: clickHouseCalc(dataPartGauge),
		}
		metricList = append(metricList, data_part_count)
	}

	for k, v := range newFamilyMetric { // new data migration
		(*old_family_metric)[k] = v
	}

	barad.Batch = metricList
	return barad
}

func clickHouseCalc(metric *prom2json.Family) float64 {
	gaugeOrCounter := prom2json.Metric{}
	if metric.Type == "GAUGE" || metric.Type == "COUNTER" {
		gaugeOrCounter = metric.Metrics[0].(prom2json.Metric)
	}

	valueFloat, _ := strconv.ParseFloat(gaugeOrCounter.Value, 64)

	return util.Decimal(valueFloat, 2)
}
