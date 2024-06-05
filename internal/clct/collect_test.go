package clct

import (
	"testing"
)

func TestAddValue(t *testing.T) {
	telemetryData.Init("test", "%", TYPE_FLOAT, true)
	telemetryData.AddValue("test", 1)
	if telemetryData.Data["test"].Value != float64(1) {
		t.Errorf("Expected 1, got %v", telemetryData.Data["test"].Value)
	}
	if telemetryData.Data["test"].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data["test"].Type)
	}
	if telemetryData.Data["test"].Count != 1 {
		t.Errorf("Expected 1, got %v", telemetryData.Data["test"].Count)
	}
	telemetryData.AddValue("test", 2)
	if telemetryData.Data["test"].Value != float64(2) {
		t.Errorf("Expected 2, got %v", telemetryData.Data["test"].Value)
	}
	if telemetryData.Data["test"].Type != TYPE_FLOAT {
		t.Errorf("Expected TYPE_FLOAT, got %v", telemetryData.Data["test"].Type)
	}
	if telemetryData.Data["test"].Count != 2 {
		t.Errorf("Expected 2, got %v", telemetryData.Data["test"].Count)
	}
	telemetryData.AddValue("test", "test1")
	if telemetryData.Data["test"].Value != "test1" {
		t.Errorf("Expected test1, got %v", telemetryData.Data["test"].Value)
	}
	if telemetryData.Data["test"].Type != TYPE_STRING {
		t.Errorf("Expected TYPE_STRING, got %v", telemetryData.Data["test"].Type)
	}
	if telemetryData.Data["test"].Count != 2 {
		t.Errorf("Expected 2, got %v", telemetryData.Data["test"].Count)
	}
}

func TestGetValue(t *testing.T) {
	telemetryData.Init("test", "%", TYPE_FLOAT, true)
	telemetryData.AddValue("test", 1)
	if telemetryData.GetValue("test") != float64(1) {
		t.Errorf("Expected 1, got %v", telemetryData.GetValue("test"))
	}
	telemetryData.AddValue("test", 2)
	if telemetryData.GetValue("test") != float64(2) {
		t.Errorf("Expected 2, got %v", telemetryData.GetValue("test"))
	}
	telemetryData.AddValue("test", "test1")
	if telemetryData.GetValue("test") != "test1" {
		t.Errorf("Expected test1, got %v", telemetryData.GetValue("test"))
	}
}

func TestIsMedianCalculable(t *testing.T) {
	tc := &TelemetryChannel{Median: true}
	if !tc.isMedianCalculable() {
		t.Errorf("Expected true, got false")
	}
	tc.Type = TYPE_FLOAT
	if !tc.isMedianCalculable() {
		t.Errorf("Expected true, got false")
	}
	tc.Type = TYPE_STRING
	if tc.isMedianCalculable() {
		t.Errorf("Expected false, got true")
	}
}

func TestGetStringValue(t *testing.T) {
	telemetryData.AddValue("test", 1)
	if telemetryData.GetStringValue("test") != "1" {
		t.Errorf("Expected 1, got %v", telemetryData.GetStringValue("test"))
	}
	telemetryData.AddValue("test", 2)
	if telemetryData.GetStringValue("test") != "2" {
		t.Errorf("Expected 2, got %v", telemetryData.GetStringValue("test"))
	}
	telemetryData.AddValue("test", "test1")
	if telemetryData.GetStringValue("test") != "test1" {
		t.Errorf("Expected test1, got %v", telemetryData.GetStringValue("test"))
	}
}

func TestGetBoolValue(t *testing.T) {
	telemetryData.AddValue("test", true)
	if !telemetryData.GetBoolValue("test") {
		t.Errorf("Expected true, got false")
	}
	telemetryData.AddValue("test", 1)
	if !telemetryData.GetBoolValue("test") {
		t.Errorf("Expected true, got false")
	}
	telemetryData.AddValue("test", 0)
	if telemetryData.GetBoolValue("test") {
		t.Errorf("Expected false, got true")
	}
	telemetryData.AddValue("test", "test1")
	if !telemetryData.GetBoolValue("test") {
		t.Errorf("Expected true, got false")
	}
}

func TestGetFloatValue(t *testing.T) {
	telemetryData.AddValue("test", 1)
	if telemetryData.GetFloatValue("test") != 1 {
		t.Errorf("Expected 1, got %v", telemetryData.GetFloatValue("test"))
	}
	telemetryData.AddValue("test", 2)
	if telemetryData.GetFloatValue("test") != 2 {
		t.Errorf("Expected 2, got %v", telemetryData.GetFloatValue("test"))
	}
	telemetryData.AddValue("test", "test1")
	if telemetryData.GetFloatValue("test") != 0 {
		t.Errorf("Expected 0, got %v", telemetryData.GetFloatValue("test"))
	}
}

func TestGetMedianValue(t *testing.T) {
	telemetryData.Init("test", "%", TYPE_FLOAT, true)
	telemetryData.AddValue("test", 1)
	if telemetryData.GetMedianValue("test") != 1 {
		t.Errorf("Expected 1, got %v", telemetryData.GetMedianValue("test"))
	}
	telemetryData.AddValue("test", 2)
	if telemetryData.GetMedianValue("test") != 1.5 {
		t.Errorf("Expected 1.5, got %v", telemetryData.GetMedianValue("test"))
	}
	telemetryData.AddValue("test", 3)
	if telemetryData.GetMedianValue("test") != 2 {
		t.Errorf("Expected 2, got %v", telemetryData.GetMedianValue("test"))
	}
	telemetryData.AddValue("test", 4)
	if telemetryData.GetMedianValue("test") != 2.5 {
		t.Errorf("Expected 2.5, got %v", telemetryData.GetMedianValue("test"))
	}
	telemetryData.AddValue("test", 5)
	if telemetryData.GetMedianValue("test") != 3 {
		t.Errorf("Expected 3, got %v", telemetryData.GetMedianValue("test"))
	}
	telemetryData.AddValue("test", "test1")
	if telemetryData.GetMedianValue("test") != 0 {
		t.Errorf("Expected 0, got %v", telemetryData.GetMedianValue("test"))
	}
}
