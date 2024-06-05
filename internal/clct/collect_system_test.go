package clct

import (
	"testing"
)

func TestSysCollector_Init(t *testing.T) {
	var sysCol sysCollector = sysCollector("system")

	sysCol.Init()
	sysCol.Collect()

	if len(telemetryData.Data) != 4 {
		t.Errorf("Expected 4, got %v", len(telemetryData.Data))
	}

	if telemetryData.Data[KEY_M_CPU].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data[KEY_M_CPU].Type)
	}

	if telemetryData.Data[KEY_M_MEM].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data[KEY_M_MEM].Type)
	}

	if telemetryData.Data[KEY_M_DISK].isMedianCalculable() {
		t.Errorf("Expected false, got true")
	}

	if telemetryData.GetFloatValue(KEY_M_DISK) == 0.0 {
		t.Errorf("Expected non-zero, got 0.0")
	}
	if telemetryData.GetFloatValue("non-existing key") != 0.0 {
		t.Errorf("Expected 0.0, got %v", telemetryData.GetFloatValue("non-existing key"))
	}
}
