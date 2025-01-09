package clct

import (
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/prutonis/acquisitor/internal/cfg"
	"github.com/prutonis/acquisitor/pkg/logger"

	"github.com/spf13/cast"
)

const (
	TYPE_FLOAT int = iota
	TYPE_STRING
	TYPE_BOOL
)

type TelemetryChannel struct {
	Value  interface{}
	Type   int // TYPE_FLOAT, TYPE_STRING, TYPE_BOOL
	Name   string
	Unit   string
	Vsum   interface{}
	Count  int
	Median bool
}

type TelemetryData struct {
	Data map[string]*TelemetryChannel
}

type ITelemetryCollector interface {
	Name() string
	Init()
	Collect()
}

type Collector struct {
	Collectors []ITelemetryCollector
}

var collector *Collector = &Collector{Collectors: make([]ITelemetryCollector, 0)}
var telemetryData *TelemetryData = &TelemetryData{Data: make(map[string]*TelemetryChannel)}

func (tc *Collector) Init() {
	for _, c := range tc.Collectors {
		fmt.Println("Initializing collector: ", c.Name())
		c.Init()
	}
}

func (tc *Collector) Collect() {
	for _, c := range tc.Collectors {
		c.Collect()
	}
}

func (td *TelemetryData) Init(name string, unit string, ttype int, median bool) {
	tc := &TelemetryChannel{Name: name, Unit: unit, Type: ttype, Count: 0, Median: median}
	td.Data[name] = tc
}

func (t *TelemetryData) ResolveChannel(name string) *TelemetryChannel {
	tc, ok := t.Data[name]
	if !ok {
		return nil
	}
	return tc
}

func (t *TelemetryData) Convert(rawVal int16, key cfg.CollectorKey) float64 {
	if key.Function != "" {
		// Define the CEL environment
		env, err := cel.NewEnv(
			cel.Declarations(
				decls.NewVar("raw", decls.Int),
			),
		)
		if err != nil {
			logger.Info("test my logger")
			logger.Fatalf("Failed to create CEL environment: %v", err)
		}

		expression := key.Function
		// Parse and check the expression
		ast, issues := env.Compile(expression)
		if issues != nil && issues.Err() != nil {
			logger.Fatalf("Failed to compile expression: %v", issues.Err())
		}

		// Create a program from the AST
		program, err := env.Program(ast)
		if err != nil {
			logger.Fatalf("Failed to create CEL program: %v", err)
		}

		// Define input data (activation)
		input := map[string]interface{}{
			"raw": rawVal, // Provide a value for the variable `raw`
		}

		// Evaluate the program
		result, _, err := program.Eval(input)
		if err != nil {
			logger.Fatalf("Failed to evaluate expression: %v", err)
		}
		// Retrieve the float64 result
		fmt.Printf("Result of evaluation: %v", result.Value())
		return result.Value().(float64)
	}
	return float64(rawVal) * float64(key.Factor)
}

func (t *TelemetryData) AddRawValue(name string, rawValue int16, cfg cfg.CollectorKey) {
	var converted = t.Convert(rawValue, cfg)
	t.AddValue(name, converted)
}

func (t *TelemetryData) AddValue(name string, value interface{}) {
	var tc, ok = t.Data[name]
	if !ok {
		tc = &TelemetryChannel{}
		t.Data[name] = tc
	}
	tc.Type = checkType(value)
	tc.Name = name
	if tc.isMedianCalculable() {
		tc.Count++
		var fv float64
		if tc.Vsum == nil {
			fv = 0.0
		} else {
			fv = tc.Vsum.(float64)
		}
		switch uv := value.(type) {
		case int:
			fv += float64(uv)
			tc.Value = float64(uv)
		case float32, float64:
			fv += float64(uv.(float64))
			tc.Value = uv
		}
		tc.Vsum = fv
	} else {
		tc.Value = value
	}
}

func (t *TelemetryData) GetValue(name string) interface{} {
	tc, ok := t.Data[name]
	if !ok {
		return nil
	}
	return tc.Value
}

func (t *TelemetryData) GetStringValue(name string) string {
	tc, ok := t.Data[name]
	if !ok {
		return ""
	}
	if tc.Type == TYPE_FLOAT {
		return fmt.Sprintf("%v", tc.Value)

	}
	if tc.Type == TYPE_BOOL {
		return fmt.Sprintf("%v", tc.Value)
	}
	return tc.Value.(string)
}

func (t *TelemetryData) GetBoolValue(name string) bool {
	tc, ok := t.Data[name]
	if !ok {
		return false
	}
	if tc.Type == TYPE_FLOAT {
		return cast.ToFloat64(tc.Value) != 0.0
	}
	if tc.Type == TYPE_STRING {
		return tc.Value.(string) != ""
	}
	return tc.Value.(bool)
}

func (t *TelemetryData) GetFloatValue(name string) float64 {
	tc, ok := t.Data[name]
	if !ok {
		return 0.0
	}
	if tc.Type == TYPE_STRING {
		return 0.0
	}
	if tc.Type == TYPE_BOOL {
		return 0.0
	}
	return cast.ToFloat64(tc.Value)
}

func (t *TelemetryData) GetMedianValue(name string) float64 {
	tc, ok := t.Data[name]
	if !ok {
		return 0
	}
	if tc.isMedianCalculable() && tc.Count > 0 {
		return tc.Vsum.(float64) / float64(tc.Count)
	} else {
		return 0
	}
}

func (tc *TelemetryChannel) GetMedianValue() float64 {
	if tc.isMedianCalculable() && tc.Count > 0 {
		return tc.Vsum.(float64) / float64(tc.Count)
	} else {
		return 0
	}
}

func (tc *TelemetryChannel) Reset() {
	tc.Count = 0
	tc.Vsum = 0.0
}

func (tc *TelemetryChannel) isMedianCalculable() bool {
	return tc.Median && tc.Type == TYPE_FLOAT
}

func checkType(value interface{}) int {
	switch value.(type) {
	case int, int32, int64, uint32, uint64:
		return TYPE_FLOAT
	case float32, float64:
		return TYPE_FLOAT
	case bool:
		return TYPE_BOOL
	default:
		return TYPE_STRING
	}
}

func (t *TelemetryData) PutValue(name string, value interface{}) {
	var tc TelemetryChannel
	if _, ok := t.Data[name]; !ok {
		tc = TelemetryChannel{}
		t.Data[name] = &tc
	}
	tc.Value = value
	tc.Type = checkType(value)
}

func (t *TelemetryData) GetMedian(name string) float64 {
	tc := t.Data[name]
	if tc.isMedianCalculable() {
		return tc.Vsum.(float64) / float64(tc.Count)
	}
	return 0.0
}
