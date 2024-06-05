package clct

import (
	"github.com/prutonis/acquisitor/internal/adc"
)

const (
	KEY_M_CURRENT = "m_current"
	KEY_M_VOLTAGE = "m_voltage"
)

var ad adc.AdcOps

type adcCollector string

func (ac *adcCollector) Init() {
	var ads adc.AdsOps = adc.NewAds(&config.Hardware.Adc)
	ad = adc.NewAdc(ads)
	telemetryData.Init(KEY_M_CURRENT, "mA", TYPE_FLOAT, true)
	telemetryData.Init(KEY_M_VOLTAGE, "V", TYPE_FLOAT, true)
}

func (ac *adcCollector) Collect() {
	// get Current consumption
	val, _ := ad.GetConverted(0)
	telemetryData.AddValue(KEY_M_CURRENT, val.Value)
	// get voltage
	val, _ = ad.GetConverted(2)
	telemetryData.AddValue(KEY_M_VOLTAGE, val.Value)
}
