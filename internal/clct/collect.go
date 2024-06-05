package clct

import (
	"encoding/json"
	"fmt"
	"time"

	cfg "github.com/prutonis/acquisitor/internal/cfg"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var config = &cfg.AcqConfig

var mqttClient mqtt.Client

func Init() {
	resolveCollectors(collector)
	collector.Init()
	createMqttClient()
}

var adcCol adcCollector = adcCollector{}
var sysCol sysCollector = sysCollector("system")

func resolveCollectors(tc *Collector) {
	tc.Collectors = append(tc.Collectors, &adcCol, &sysCol)
}

func Collect() {
	Init()
	if ad != nil {
		defer ad.Close()
	}
	defer mqttClient.Disconnect(250)

	sendTicker := time.NewTicker(time.Second * time.Duration(config.Telemetry.Pusher.Interval))
	adcColConf := config.Telemetry.ResolveCollector("adc")
	sysColConf := config.Telemetry.ResolveCollector("system")
	var adcColTicker *time.Ticker = createCollectorTicker(adcColConf)
	var sysColTicker *time.Ticker = createCollectorTicker(sysColConf)

	for {
		select {
		case <-adcColTicker.C:
			adcCol.Collect()
		case <-sysColTicker.C:
			sysCol.Collect()
		case <-sendTicker.C:
			sendTelemetry()
		}
	}
}

func createCollectorTicker(collectorCfg *cfg.Collector) *time.Ticker {
	if collectorCfg != nil {
		return time.NewTicker(time.Second * time.Duration(collectorCfg.Interval))
	} else {
		return time.NewTicker(time.Hour * time.Duration(1000000))
	}
}

func sendTelemetry() {
	fmt.Println("Sending telemetry")
	var payload, err = createPayload()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(payload)
	//go func() {
	token := mqttClient.Publish(config.Telemetry.Server.Topic, 0, false, payload)
	token.Wait()
	//}()
}

func createPayload() (string, error) {
	payload := make(map[string]interface{})

	for _, cc := range config.Telemetry.Pusher.Keys {
		c := telemetryData.Data[cc.Source]
		if c != nil {
			if c.isMedianCalculable() {
				payload[c.Name] = telemetryData.GetMedianValue(c.Name)
			} else {
				switch c.Type {
				case TYPE_FLOAT:
					payload[c.Name] = c.Value
				case TYPE_STRING:
					payload[c.Name] = c.Value
				}
			}
		}
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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
