package clct

import (
	"testing"

	"github.com/prutonis/acquisitor/internal/cfg"
)

func TestTelemetryData_Convert(t *testing.T) {
	type args struct {
		rawVal int16
		key    cfg.CollectorKey
	}
	tests := []struct {
		name string
		tr   *TelemetryData
		args args
		want float64
	}{
		{
			name: "Test with empty key.Function",
			tr:   &TelemetryData{},
			args: args{
				rawVal: 1000,
				key:    cfg.CollectorKey{Factor: 2.5},
			},
			want: 2500.0, // rawVal * key.Factor
		},
		{
			name: "Test with valid CEL expression",
			tr:   &TelemetryData{},
			args: args{
				rawVal: 7000,
				key:    cfg.CollectorKey{Function: "(((double(raw) * 1.25) / 1000.0) - 4.0) * (6.0/16.0)"},
			},
			want: 1.78125,
		},
		{
			name: "Test with negative rawVal",
			tr:   &TelemetryData{},
			args: args{
				rawVal: -1000, // Test with negative raw value
				key:    cfg.CollectorKey{Factor: 2.5},
			},
			want: -2500.0, // rawVal * key.Factor
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tr.Convert(tt.args.rawVal, tt.args.key); got != tt.want {
				t.Errorf("TelemetryData.Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}
