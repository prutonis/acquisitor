package clct

import (
	"testing"

	cfg "github.com/prutonis/acquisitor/internal/cfg"
)

func init() {
	cfg.AcqConfig = cfg.Config{
		Telemetry: cfg.Telemetry{
			Collectors: []cfg.Collector{
				{
					Name:     "system",
					Enabled:  true,
					Interval: 1,
					Keys: []cfg.CollectorKey{
						{
							Name:   KEY_LOAD,
							Unit:   "%",
							Type:   TYPE_FLOAT,
							Median: false,
							Source: "",
							Factor: 1.0,
						},
						{
							Name:   KEY_MEM,
							Unit:   "%",
							Type:   TYPE_FLOAT,
							Median: false,
							Source: "",
							Factor: 1.0,
						},
						{
							Name:   KEY_DISK,
							Unit:   "%",
							Type:   TYPE_FLOAT,
							Median: false,
							Source: "",
							Factor: 1.0,
						},
						{
							Name:   KEY_UPTIME,
							Unit:   "%",
							Type:   TYPE_FLOAT,
							Median: false,
							Source: "",
							Factor: 1.0,
						},
					},
				},
			},
		},
	}
}
func TestSysCollector_Init(t *testing.T) {
	var sysCol sysCollector = sysCollector("system")

	sysCol.Init()
	sysCol.Collect()

	if len(telemetryData.Data) != 4 {
		t.Errorf("Expected 4, got %v", len(telemetryData.Data))
	}

	if telemetryData.Data[KEY_LOAD].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data[KEY_LOAD].Type)
	}

	if telemetryData.Data[KEY_MEM].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data[KEY_MEM].Type)
	}

	if telemetryData.Data[KEY_DISK].isMedianCalculable() {
		t.Errorf("Expected false, got true")
	}

	if telemetryData.GetFloatValue(KEY_DISK) == 0.0 {
		t.Errorf("Expected non-zero, got 0.0")
	}
	if telemetryData.GetFloatValue("non-existing key") != 0.0 {
		t.Errorf("Expected 0.0, got %v", telemetryData.GetFloatValue("non-existing key"))
	}
}
