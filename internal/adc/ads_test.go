package adc

import (
	"fmt"
	"testing"

	cfg "github.com/prutonis/acquisitor/internal/cfg"
)

func TestAds(t *testing.T) {
	testCfg := cfg.Adc{
		Name: "Fake ADC",
		Type: "Fake",
		Inputs: []cfg.AdcInput{
			{
				Name:    "Fake Input 0",
				Channel: 0,
				Gain:    1,
			},
			{
				Name:    "Fake Input 1",
				Channel: 1,
				Gain:    1,
			},
			{
				Name:    "Fake Input 2",
				Channel: 2,
				Gain:    1,
			},
			{
				Name:    "Fake Input 3",
				Channel: 3,
				Gain:    1,
			},
		},
	}
	ads := newFakeAds(&testCfg)
	ads.SetConfig(&testCfg.Inputs[0])
	val, err := ads.ReadValue()
	if err != nil {
		t.Errorf("Error reading value: %v", err)
	}
	if val != 2022 {
		t.Errorf("Expected 2022, got %v", val)
	}
	ads.SetConfig(&testCfg.Inputs[1])
	val, _ = ads.ReadValue()
	if val != -3 {
		t.Errorf("Expected -3, got %v", val)
	}
	ads.SetConfig(&testCfg.Inputs[2])
	val, _ = ads.ReadValue()
	if val != 25290 {
		t.Errorf("Expected 25290, got %v", val)
	}
	ads.SetConfig(&testCfg.Inputs[3])
	val, _ = ads.ReadValue()
	if val != 581 {
		t.Errorf("Expected 581, got %v", val)
	}
}

func newFakeAds(adcCfg *cfg.Adc) *FakeAds {
	fmt.Printf("Using Fake ADC: %v\n", adcCfg)
	return &FakeAds{ct: 0, values: [24]int16{
		2049, -1, 25229, 521,
		2022, -2, 25239, 523,
		2078, -3, 25529, 525,
		2021, -1, 25290, 532,
		2002, -4, 25223, 581,
		2041, -1, 25264, 511}}
}

type FakeAds struct {
	values  [24]int16 // fake values
	current *cfg.AdcInput
	ct      int
}

// implement AdcOps
func (adc *FakeAds) SetConfig(config *cfg.AdcInput) error {
	adc.current = config
	return nil
}

func (adc *FakeAds) ReadValue() (int16, error) {
	adc.ct++
	adc.ct = adc.ct % 6
	return adc.values[(uint16(adc.ct*4))+adc.current.Channel], nil
}

func (adc *FakeAds) ReadConfig() (uint16, error) {
	return adc.current.Channel, nil
}

func (adc *FakeAds) Close() error {
	return nil
}
