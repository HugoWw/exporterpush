package global

import (
	"github.com/exporterpush/pkg/logger"
	"github.com/exporterpush/pkg/setting"
)

var (
	GlobalSetting      *setting.GlobalS
	BaradSetting       *setting.BaradS
	PrometheusSetting  *setting.PrometheusS
	PushgatewaySetting *setting.PushgatewayS
	LogObj             *logger.Logger
)

const (
	/*
		node_exporter
	*/

	NodeDisKReadsCompeted    = "node_disk_reads_completed"
	NodeDiskWritesComplete   = "node_disk_writes_completed"
	NodeDiskBytesWritten     = "node_disk_bytes_written"
	NodeDiskBytesRead        = "node_disk_bytes_read"
	NodeFileSystemSize       = "node_filesystem_size"
	NodeFileSystemFree       = "node_filesystem_free"
	NodeFileSystemAvail      = "node_filesystem_avail"
	NodeFileSystemFiles      = "node_filesystem_files"
	NodeFileSystemFilesFree  = "node_filesystem_files_free"
	NodeCpu                  = "node_cpu"
	NodeMemoryMemFree        = "node_memory_MemFree"
	NodeMemoryBuffers        = "node_memory_Buffers"
	NodeMemoryCached         = "node_memory_Cached"
	NodeMemoryMemTotal       = "node_memory_MemTotal"
	NodeMemorySlab           = "node_memory_Slab"
	NodeNetworkTransmitBytes = "node_network_transmit_bytes"
	NodeNetworkReceiveBytes  = "node_network_receive_bytes"

	/*
		clickhouse_exporter
	*/

	ClickhouseTcpConnection                = "ClickHouseMetrics_TCPConnection"
	ClickhouseHttpConnection               = "ClickHouseMetrics_HTTPConnection"
	ClickhouseMysqlConnection              = "ClickHouseMetrics_MySQLConnection"
	ClickhouseInterServerConnection        = "ClickHouseMetrics_InterserverConnection"
	ClickhousePostgreSQLConnection         = "ClickHouseMetrics_PostgreSQLConnection"
	ClickhouseFailedQueryCount             = "ClickHouseProfileEvents_FailedQuery"
	ClickhouseQueryCount                   = "ClickHouseProfileEvents_Query"
	ClickhouseDelayedInsertsCount          = "ClickHouseProfileEvents_DelayedInserts"
	ClickhouseMergeCount                   = "ClickHouseProfileEvents_Merge"
	ClickhouseReplicatedPartMutationsCount = "ClickHouseProfileEvents_ReplicatedPartMutations"
	ClickhouseInsertedRowsCount            = "ClickHouseProfileEvents_InsertedRows"
	ClickhouseInsertedBytesSize            = "ClickHouseProfileEvents_InsertedBytes"
	ClickhouseTPS_InsertQueryCount         = "ClickHouseProfileEvents_InsertQuery"
	ClickhouseQPS_QueryCount               = "ClickHouseProfileEvents_Query"
	ClickhouseDataPartsGauge               = "ClickHouseMetrics_PartsCommitted"
	//ClickhouseDataPartsGauge               = "ClickHouseMetrics_PartsActive"
)
