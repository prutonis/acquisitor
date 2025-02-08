//go:build windows

package clct

import (
	"github.com/prutonis/acquisitor/pkg/logger"
)

type gpioCollector struct{}

func (gc *gpioCollector) Init() {
	logger.Warningf("GPIO unavailable on Windows!")
}

func (gc *gpioCollector) Collect() {
	logger.Warningf("Collecting gpio data (unavailable on windows)")
}

func (gc *gpioCollector) Name() string {
	return "gpio"
}

func (gc *gpioCollector) ReadPins() map[string]int {
	var pinMap = make(map[string]int)
	logger.Warningf("ReadPins unavailable on windows")
	return pinMap
}

func (gc *gpioCollector) SetPins(pins map[string]interface{}) {
	logger.Warningf("SetPins unavailable on windows")
}
