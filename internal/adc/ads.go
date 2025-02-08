package adc

import (
	"log"
	"time"

	cfg "github.com/prutonis/acquisitor/internal/cfg"
	"github.com/prutonis/acquisitor/pkg/logger"
	ads "github.com/prutonis/go-ads1115"
	i2c "github.com/prutonis/go-i2c"
)

type ConVal struct {
	Channel int
	Value   float32
	Unit    string
}

type AdsOps interface {
	SetConfig(adcInput *cfg.AdcInput) error
	SetCurrentChannel(channel int) error
	ReadValue() (int16, error)
	ReadConfig() (uint16, error)
	Close() error
}

type Ads struct {
	Cfg    *cfg.Adc
	Sensor *ads.ADS
	I2C    *i2c.I2C
}

func NewAds(adcCfg *cfg.Adc) *Ads {
	logger.Infof("Using ADC: %v\n", adcCfg)
	i2c, err := i2c.NewI2C(uint8(adcCfg.I2cAddr), adcCfg.I2cBus)
	if err != nil {
		logger.Fatalf("Cannot initiate I2C", err)
	}

	sensor, err := ads.NewADS(ads.ADS1115, i2c) // signature=0x58
	if err != nil {
		logger.Fatalf("Cannot initiate ADS1115", err)
	}

	err = sensor.SetConversionMode(ads.MODE_SINGLE_SHOT)
	if err != nil {
		log.Fatal(err)
	}
	logger.Infof("  Configured for single shot mode")

	err = sensor.SetDataRate(ads.RATE_8)
	if err != nil {
		logger.Fatalf("Cannot set data rate", err)
	}
	logger.Infof("  Configured for 128 Samples per Second") // is working for single shot mode?

	err = sensor.SetComparatorMode(ads.COMP_MODE_TRADITIONAL)
	if err != nil {
		logger.Fatalf("Cannot set comparator mode", err)
	}
	logger.Infof("  Configured for traditional comparator mode")

	err = sensor.SetComparatorPolarity(ads.COMP_POL_ACTIVE_LOW)
	if err != nil {
		logger.Fatalf("Cannot set comparator polarity", err)
	}
	logger.Infof("  Configured comparator active low")

	err = sensor.SetComparatorLatch(ads.COMP_LAT_OFF)
	if err != nil {
		logger.Fatalf("Cannot set comparator latch", err)
	}
	logger.Infof("  Configured comparator latch off")

	err = sensor.SetComparatorQueue(ads.COMP_QUE_DISABLE)
	if err != nil {
		logger.Fatalf("Cannot set comparator queue", err)
	}
	err = sensor.SetPgaMode(ads.PGA_2_048)
	if err != nil {
		logger.Fatalf("Cannot set PGA mode", err)
	}
	err = sensor.SetMuxMode(ads.MUX_SINGLE_0)
	if err != nil {
		logger.Fatalf("Cannot set MUX mode", err)
	}

	logger.Infof("ADC initialized", sensor)
	sensor.WriteConfig()
	return &Ads{Cfg: adcCfg, Sensor: sensor, I2C: i2c}
}

func (ads *Ads) SetConfig(config *cfg.AdcInput) error {
	//logger.Infof("Setting ADC config: %v\n", config)
	err := ads.Sensor.SetPgaMode(config.Gain)
	if err != nil {
		return err
	}
	err = ads.Sensor.SetMuxMode(4 + uint16(config.Channel))
	if err != nil {
		return err
	}
	ads.Sensor.WriteConfig()
	return err
}

func (ads *Ads) SetCurrentChannel(channel int) error {
	return ads.SetConfig(&ads.Cfg.Inputs[channel])
}

func (ads *Ads) ReadValue() (int16, error) {
	err := ads.Sensor.StartConversion()
	if err != nil {
		return 0, err
	}
	//status, err := ads.Sensor.ReadStatus()
	if err != nil {
		return 0, err
	}
	for i := 0; i < 10; i++ {
		//status, err = ads.Sensor.ReadStatus()
		if err != nil {
			return 0, err
		}
		// delay for conversion to complete
		milis := ads.Cfg.ConvDelay
		if milis == 0 {
			milis = 10
		}
		time.Sleep(time.Millisecond * time.Duration(milis))
	}

	if err != nil {
		return 0, err
	}

	val, err := ads.Sensor.ReadConversion()
	return val, err
}

func (ads *Ads) ReadConfig() (uint16, error) {
	return ads.Sensor.ReadConfig()
}

func (ads *Ads) Close() error {
	return ads.I2C.Close()
}
