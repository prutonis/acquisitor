package clct

import (
	"math/rand"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/prutonis/acquisitor/internal/adc"
	"github.com/prutonis/acquisitor/internal/cfg"
	"github.com/prutonis/acquisitor/pkg/logger"
)

const (
	COLLECTOR_NAME_ADC = "adc"
)

var ad adc.AdcOps = nil

type adcCollector struct {
	Keys []adcKey
}

type adcKey struct {
	collectorKey  cfg.CollectorKey
	adcInput      cfg.AdcInput
	transformFunc cel.Program
}

func (ac *adcCollector) Init() {
	if config.Hardware.Adc.Enabled {
		// if adc name contains 'fake' then use fake adc
		ac.initCollector()
		if strings.Contains(config.Hardware.Adc.Name, "fake") {
			// The string contains the substring "fake"
			logger.Infof("Using fake ADC")
			ad = FakeAdc("fake")
			return
		}
		var ads adc.AdsOps = adc.NewAds(&config.Hardware.Adc)
		ad = adc.NewAdc(ads)
	} else {
		logger.Infof("ADC disabled")
	}
}

func (ac *adcCollector) Collect() {
	logger.Debugf("Collecting adc data")
	for _, key := range ac.Keys {
		rawVal, err := ad.ReadValue(int(key.adcInput.Channel))
		if err == nil {
			telemetryData.AddRawValue(key.collectorKey.Name, rawVal, key.collectorKey, key.transformFunc)
		}
	}
}

func (ac *adcCollector) Name() string {
	return COLLECTOR_NAME_ADC
}

func (ac *adcCollector) initCollector() {
	ac.Keys = make([]adcKey, 0)
	var adcCol = config.Telemetry.ResolveCollector(ac.Name())
	if adcCol == nil {
		return
	}
	// create a map from an array of inputs
	var adcInputs = make(map[string]cfg.AdcInput)
	for _, input := range config.Hardware.Adc.Inputs {
		adcInputs[input.Name] = input
	}
	for _, key := range adcCol.Keys {
		var adcInput, ok = adcInputs[key.Source]
		if ok {
			ac.Keys = append(ac.Keys, adcKey{key, adcInput, loadTransformFunction(key.Function)})
			telemetryData.Init(key.Name, key.Unit, key.Type, key.Median)
		}
	}
}

func loadTransformFunction(function string) cel.Program {
	if function == "" {
		function = "raw"
	}
	// Define the CEL environment
	env, err := cel.NewEnv(
		cel.Declarations(
			decls.NewVar("raw", decls.Double),
		),
	)
	if err != nil {
		logger.Info("test my logger")
		logger.Fatalf("Failed to create CEL environment: %v", err)
	}

	// Parse and check the expression
	ast, issues := env.Compile(function)
	if issues != nil && issues.Err() != nil {
		logger.Fatalf("Failed to compile expression: %v", issues.Err())
	}

	// Create a program from the AST
	prog, err := env.Program(ast)
	if err != nil {
		logger.Fatalf("Failed to create CEL program: %v", err)
	}

	return prog
}

type FakeAdc string

func (f FakeAdc) Init() error {
	return nil
}

func (f FakeAdc) ReadValue(channel int) (int16, error) {
	min := 100 // Define minimum value
	max := 400 // Define maximum value
	randomValue := rand.Intn(max-min+1) + min
	return int16(randomValue), nil

}

func (f FakeAdc) Close() error {
	return nil
}
