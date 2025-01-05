package clct

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	cfg "github.com/prutonis/acquisitor/internal/cfg"
	"github.com/spf13/cast"

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
	var payload, err = createPayload()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Sending telemetry: ", payload)
	go func() {
		token := mqttClient.Publish(config.Telemetry.Server.Topic, 0, false, payload)
		if token.WaitTimeout(5 * time.Second) {
			fmt.Println("Telemetry sent ", payload)
		} else {
			fmt.Println("Timeout on telemetry sending")
		}
	}()
}

func createPayload() (string, error) {
	payload := make(map[string]interface{})
	precision := config.Telemetry.Pusher.Precision

	for _, cc := range config.Telemetry.Pusher.Keys {
		c := telemetryData.Data[cc.Source]
		if c != nil {
			if c.isMedianCalculable() {
				payload[c.Name] = roundUp(c.GetMedianValue(), precision)
			} else {
				switch c.Type {
				case TYPE_FLOAT:
					payload[cc.Name] = roundUp(c.Value, precision)
				case TYPE_STRING:
					payload[cc.Name] = c.Value
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

func roundUp(value interface{}, precision int) float64 {
	m := math.Pow10(precision)
	fv := cast.ToFloat64(value)
	return math.Round(fv*m) / m
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
