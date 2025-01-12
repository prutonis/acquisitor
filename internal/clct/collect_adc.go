package clct

import (
	"fmt"
	"math/rand"
	"strings"

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
		// if adc name contains 'fake' then use fake adc
		ac.initCollector()
		if strings.Contains(config.Hardware.Adc.Name, "fake") {
			// The string contains the substring "fake"
			fmt.Println("Using fake ADC")
			ad = FakeAdc("fake")
			return
		}
		var ads adc.AdsOps = adc.NewAds(&config.Hardware.Adc)
		ad = adc.NewAdc(ads)
	} else {
		fmt.Println("ADC disabled")
	}
}

func (ac *adcCollector) Collect() {
	fmt.Println("Collecting adc data")
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
			ac.Keys = append(ac.Keys, adcKey{key, adcInput})
			telemetryData.Init(key.Name, key.Unit, key.Type, key.Median)
		}
	}
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
