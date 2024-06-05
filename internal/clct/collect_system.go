package clct

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

const (
	KEY_M_CPU    = "m_cpu"
	KEY_M_MEM    = "m_mem"
	KEY_M_UPTIME = "m_uptime"
	KEY_M_DISK   = "m_disk"
)

type sysCollector string

func (ac *sysCollector) Init() {
	telemetryData.Init(KEY_M_CPU, "%", TYPE_FLOAT, false)
	telemetryData.Init(KEY_M_MEM, "%", TYPE_FLOAT, false)
	telemetryData.Init(KEY_M_DISK, "%", TYPE_FLOAT, false)
	telemetryData.Init(KEY_M_UPTIME, "s", TYPE_STRING, false)
}

func (ac *sysCollector) Collect() {
	percentages, err := cpu.Percent(0, false)
	if err == nil {
		telemetryData.AddValue(KEY_M_CPU, percentages[0])
	}
	vm, err := mem.VirtualMemory()
	if err == nil {
		telemetryData.AddValue(KEY_M_MEM, vm.UsedPercent)
	}
	dusage, err := disk.Usage("/")
	if err == nil {
		telemetryData.AddValue(KEY_M_DISK, dusage.UsedPercent)
	}
	uptime, err := host.Uptime()
	if err == nil {
		telemetryData.AddValue(KEY_M_UPTIME, uptime)
	}
}
