package node_calc

import (
	"github.com/exporterpush/pkg/util"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
	"time"
)

type NodePerSecondMetric struct {
	DiskUseStat       *disk.UsageStat
	CpuUsagePercent   float64
	MemoryUseState    *mem.VirtualMemoryStat
	MemoryUsedPercent float64
}

// GetDiskUseInfo return "/data" partition usage info(used,free,total,usedPercent,inodesUsedPercent...)
// if there is no "/ data" return "/"
func GetDiskUseInfo() (*disk.UsageStat, error) {
	diskPath := "/"
	deviceDiskPartitons, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	for _, partition := range deviceDiskPartitons {
		if partition.Mountpoint == "/data" {
			diskPath = partition.Mountpoint
		}
	}

	partitionUsageInfo, err := disk.Usage(diskPath)
	if err != nil {
		return nil, err
	}

	return partitionUsageInfo, nil
}

func GetCpuUsage() (float64, error) {
	cpuUsePercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}

	return util.Decimal(cpuUsePercent[0], 2), nil
}

func GetPerSecondMetric() (*NodePerSecondMetric, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	disUseInfo, err := GetDiskUseInfo()
	if err != nil {
		return nil, err
	}

	cpuUsage, err := GetCpuUsage()
	if err != nil {
		return nil, err
	}

	memUsePercent := 100 - ((float64(memInfo.Available) / float64(memInfo.Total)) * 100)

	nodeInfo := &NodePerSecondMetric{
		DiskUseStat:       disUseInfo,
		CpuUsagePercent:   cpuUsage,
		MemoryUseState:    memInfo,
		MemoryUsedPercent: util.Decimal(memUsePercent, 2),
	}

	return nodeInfo, nil
}

func GetMemUseInfo() (*mem.VirtualMemoryStat, error) {
	memInfo, err := mem.VirtualMemory()

	if err != nil {
		return nil, err
	}

	return memInfo, nil
}

func GetDiskRWAndIO(names string) (map[string]disk.IOCountersStat, error) {

	disk, err := disk.IOCounters(names)
	if err != nil {
		return nil, err
	}

	return disk, nil
}

func GetNetWorkInfo(eth string) (*net.IOCountersStat, error) {
	netInfoList, err := net.IOCounters(true)
	if err != nil {
		return nil, err
	}

	var netInfo net.IOCountersStat

	for _, net := range netInfoList {
		if net.Name == eth {
			netInfo = net
		}
	}

	return &netInfo, nil
}
