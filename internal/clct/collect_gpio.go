package clct

import (
	"fmt"

	"github.com/prutonis/acquisitor/pkg/logger"
	"github.com/warthog618/go-gpiocdev"
)

const (
	KEY_CAM1 = "cam1"
	KEY_CAM2 = "cam2"
	KEY_CAM3 = "cam3"
	GPIOCHIP = "gpiochip0"
)

type gpioCollector struct {
	Pins map[string]gpioPin
}

type gpioPin struct {
	name string
	line *gpiocdev.Line
}

func (gc *gpioCollector) Init() {
	if !config.Hardware.Gpio.Enabled {
		logger.Log.Warn("GPIO pins not configured!")
		return
	}
	var gpioCol = config.Telemetry.ResolveCollector(gc.Name())
	if gpioCol == nil {
		logger.Log.Warn("GPIO collector not configured")
		return
	}
	gc.Pins = map[string]gpioPin{}
	for _, p := range config.Hardware.Gpio.Pins {
		logger.Log.Infof("Configuring GPIO pin %s[%d]=%d", p.Name, p.Pin, p.Default)
		line, _ := gpiocdev.RequestLine(GPIOCHIP, p.Pin, gpiocdev.AsOutput(p.Default))
		gc.Pins[p.Name] = gpioPin{
			name: p.Name,
			line: line,
		}
	}

	for _, key := range gpioCol.Keys {
		telemetryData.Init(key.Name, key.Unit, key.Type, key.Median)
	}
}

func (gc *gpioCollector) Collect() {
	fmt.Println("Collecting gpio data")
	var gpioCol = config.Telemetry.ResolveCollector(gc.Name())
	for _, key := range gpioCol.Keys {
		gp, exists := gc.Pins[key.Source]
		if exists {
			lineState, err := gp.line.Value()
			if err == nil {
				telemetryData.AddValue(key.Name, lineState)
			}
		}
	}
}

func (gc *gpioCollector) Name() string {
	return "gpio"
}