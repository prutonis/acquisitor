package clct

import (
	"fmt"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	KEY_LOAD   = "load"
	KEY_MEM    = "mem"
	KEY_UPTIME = "uptime"
	KEY_DISK   = "disk"
)

type sysCollector string

func (ac *sysCollector) Init() {
	var sysCol = config.Telemetry.ResolveCollector("system")
	if sysCol == nil {
		return
	}
	for _, key := range sysCol.Keys {
		telemetryData.Init(key.Name, key.Unit, key.Type, key.Median)
	}
}

func (ac *sysCollector) Collect() {
	fmt.Println("Collecting system data")
	if telemetryData.ResolveChannel(KEY_LOAD) != nil {
		percentages, err := cpu.Percent(0, false)
		if err == nil {
			telemetryData.AddValue(KEY_LOAD, percentages[0])
		}
	}

	if telemetryData.ResolveChannel(KEY_MEM) != nil {
		vm, err := mem.VirtualMemory()
		if err == nil {
			telemetryData.AddValue(KEY_MEM, vm.UsedPercent)
		}
	}
	if telemetryData.ResolveChannel(KEY_DISK) != nil {
		dusage, err := disk.Usage("/")
		if err == nil {
			telemetryData.AddValue(KEY_DISK, dusage.UsedPercent)
		}
	}
	if telemetryData.ResolveChannel(KEY_UPTIME) != nil {
		u, err := host.Uptime()
		if err == nil {
			telemetryData.AddValue(KEY_UPTIME, u)
		}
	}
}

func (sc *sysCollector) Name() string {
	return "system"
}
