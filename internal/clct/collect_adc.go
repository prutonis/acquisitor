package clct

import (
	"fmt"

	"github.com/prutonis/acquisitor/internal/adc"
	"github.com/prutonis/acquisitor/internal/cfg"
)

const (
	COLLECTOR_NAME_ADC = "adc"
)

var ad adc.AdcOps = nil

type adcCollector struct {
	Keys []adcKey
}

type adcKey struct {
	collectorKey cfg.CollectorKey
	adcInput     cfg.AdcInput
}

func (ac *adcCollector) Init() {
	if config.Hardware.Adc.Enabled {
		var ads adc.AdsOps = adc.NewAds(&config.Hardware.Adc)
		ad = adc.NewAdc(ads)
		ac.initCollector()
	} else {
		fmt.Println("ADC disabled")
	}
}

func (ac *adcCollector) Collect() {
	for _, key := range ac.Keys {
		rawVal, err := ad.ReadValue(int(key.adcInput.Channel))
		if err == nil {
			telemetryData.AddRawValue(key.collectorKey.Name, rawVal, key.collectorKey)
		}
	}
}

func (ac *adcCollector) Name() string {
	return COLLECTOR_NAME_ADC
}

func (ac *adcCollector) initCollector() {
	ac.Keys = make([]adcKey, 0)
	var adcCol = config.Telemetry.ResolveCollector(COLLECTOR_NAME_ADC)
	if adcCol == nil {
		return
	}
	// create a map from an array of inputs
	var adcInputs = make(map[string]cfg.AdcInput)
	for _, input := range config.Hardware.Adc.Inputs {
		adcInputs[input.Name] = input
	}
	for _, key := range adcCol.Keys {
		var adcInput, ok = adcInputs[key.Name]
		if ok {
			ac.Keys = append(ac.Keys, adcKey{key, adcInput})
			telemetryData.Init(key.Name, key.Unit, key.Type, key.Median)
		}
	}
}
