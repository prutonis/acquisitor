package clct

import (
	"fmt"
	"time"

	cfg "github.com/prutonis/acquisitor/internal/cfg"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var config = &cfg.AcqConfig

var mqttClient mqtt.Client

func Init() {
	resolveCollectors(telemetryCollector)
	telemetryCollector.Init()
	createMqttClient()
}

func resolveCollectors(tc *TelemetryCollector) {
	var adcCol adcCollector = adcCollector("adc")
	var sysCol sysCollector = sysCollector("system")
	tc.Collectors = append(tc.Collectors, &adcCol, &sysCol)
}

func Collect() {
	Init()
	defer ad.Close()
	defer mqttClient.Disconnect(250)
	collectTicker := time.NewTicker(time.Second)
	sendTicker := time.NewTicker(time.Minute)

	for {
		select {
		case <-collectTicker.C:
			collectTelemetry()
		case <-sendTicker.C:
			sendTelemetry()
		}
	}
}

func collectTelemetry() {
	telemetryCollector.Collect()
}

func sendTelemetry() {
	fmt.Println("Sending telemetry")
	var payload = createPayload()
	fmt.Println(payload)

	token := mqttClient.Publish("v1/devices/me/telemetry", 0, false, payload)
	token.Wait()
}

func createPayload() string {
	var payload string = "{"
	for _, c := range telemetryData.Data {
		if c.isMedianCalculable() {
			payload += fmt.Sprintf(`"%s": %f,`, c.Name, telemetryData.GetMedianValue(c.Name))

		}
		c.Reset()
	}
	payload += "}"
	return payload
}

func createMqttClient() {
	var brokerConnectionStr = fmt.Sprintf("tcp://%s:%d", config.Telemetry.Server.Host, config.Telemetry.Server.Port)
	opts := mqtt.NewClientOptions().AddBroker(brokerConnectionStr)
	opts.SetClientID("acquisitor")
	opts.SetUsername(config.Telemetry.Server.User)
	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}
