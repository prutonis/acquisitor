//go:build windows

package clct

import (
	"github.com/prutonis/acquisitor/pkg/logger"
)

type gpioCollector struct{}

func (gc *gpioCollector) Init() {
	logger.Log.Warn("GPIO unavailable on Windows!")
}

func (gc *gpioCollector) Collect() {
	logger.Log.Warn("Collecting gpio data (unavailable on windows)")
}

func (gc *gpioCollector) Name() string {
	return "gpio"
}

func (gc *gpioCollector) ReadPins() map[string]int {
	var pinMap = make(map[string]int)
	logger.Log.Warn("ReadPins unavailable on windows")
	return pinMap
}

func (gc *gpioCollector) SetPins(pins map[string]interface{}) {
	logger.Log.Warn("SetPins unavailable on windows")
}
