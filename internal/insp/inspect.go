package insp

import (
	"fmt"
	"log"
	"time"

	ads "github.com/prutonis/acquisitor/internal/adc"
	cfg "github.com/prutonis/acquisitor/internal/cfg"
)

var config = &cfg.AcqConfig
var adc ads.AdsOps

func Init() {
	log.Println("Inspect init called")
	fmt.Println(config)
	adc = ads.NewAds(&config.Hardware.Adc)
}

func Inspect() {
	Init()
	defer adc.Close()
	var v int16
	for {
		<-time.After(time.Second)
		adc.SetConfig(&config.Hardware.Adc.Inputs[0])
		<-time.After(300 * time.Millisecond)
		v, _ = adc.ReadValue()
		fmt.Println("read from sensor 0:", v)
		adc.SetConfig(&config.Hardware.Adc.Inputs[1])
		<-time.After(300 * time.Millisecond)
		v, _ = adc.ReadValue()
		fmt.Println("read from sensor 1:", v)
		adc.SetConfig(&config.Hardware.Adc.Inputs[2])
		<-time.After(300 * time.Millisecond)
		v, _ = adc.ReadValue()
		fmt.Println("read from sensor 2:", v)
		adc.SetConfig(&config.Hardware.Adc.Inputs[3])
		<-time.After(300 * time.Millisecond)
		v, _ = adc.ReadValue()
		fmt.Println("read from sensor 3:", v)

	}
}
