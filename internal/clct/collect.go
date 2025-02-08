package clct

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	cfg "github.com/prutonis/acquisitor/internal/cfg"
	"github.com/prutonis/acquisitor/pkg/logger"
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
var gpioCol gpioCollector = gpioCollector{}

func resolveCollectors(tc *Collector) {
	tc.Collectors = append(tc.Collectors, &adcCol, &sysCol, &gpioCol)
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
	gpioColConf := config.Telemetry.ResolveCollector("gpio")
	var adcColTicker *time.Ticker = createCollectorTicker(adcColConf)
	var sysColTicker *time.Ticker = createCollectorTicker(sysColConf)
	var gpioColTicker *time.Ticker = createCollectorTicker(gpioColConf)
	for {
		select {
		case <-adcColTicker.C:
			adcCol.Collect()
		case <-sysColTicker.C:
			sysCol.Collect()
		case <-gpioColTicker.C:
			gpioCol.Collect()
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
	payload := createTelemetryJson()
	logger.Info("Sending telemetry: ", payload)
	go func() {
		token := mqttClient.Publish(config.Telemetry.Server.PublishTopic, 0, false, payload)
		if token.WaitTimeout(5 * time.Second) {
			logger.Debug("Telemetry sent ", payload)
		} else {
			logger.Info("Timeout on telemetry sending")
		}
	}()
}

func createTelemetryPayload() map[string]interface{} {
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
	return payload
}

func createTelemetryJson() string {
	payload := createTelemetryPayload()
	return serialize(payload)
}

func serialize(payload map[string]interface{}) string {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(bytes)
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
	st := config.Telemetry.Server.SubscribeTopic + "+"
	mqttClient.Subscribe(st, 1, messageSubHandler)
}

type RpcPayload struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

var messageSubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	logger.Infof("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())

	reqId, found := strings.CutPrefix(msg.Topic(), config.Telemetry.Server.SubscribeTopic)
	if found {
		var payload RpcPayload
		err := json.Unmarshal(msg.Payload(), &payload)
		if err != nil {
			logger.Errorf("Error parsing JSON payload: %v\n", err)
			return
		}

		// Access parsed data
		logger.Infof("Parsed Payload: %+v\n", payload)
		logger.Infof("Method: %s, params: %+v\n", payload.Method, payload.Params)

		switch v := payload.Params.(type) {
		case string:
			logger.Infof("Params Value (string): %s\n", v)
		case float64: // JSON numbers are parsed as float64 in Go
			logger.Infof("Params Value (number): %f\n", v)
		case map[string]interface{}: // Nested objects
			logger.Infof("Params Value (object): %v\n", v)
		case []interface{}: // Array of values
			logger.Infof("Params Value (array): %v\n", v)
		default:
			logger.Infof("Params Value (unknown type): %v\n", v)
		}

		ExecuteRpc(reqId, Command(payload.Method), payload.Params)
	}

}

type Command string

const (
	GetStatus    Command = "getStatus"
	GetTelemetry Command = "getTelemetry"
	GetPins      Command = "getPins"
	SetPins      Command = "setPins"
	SetPin       Command = "setPin"
	Help         Command = "Help"
)

type RpcResponse map[string]interface{}

func ExecuteRpc(requestId string, cmd Command, params interface{}) {
	resp := make(RpcResponse)
	switch cmd {
	case GetStatus:
		logger.Info("Status is OK.")
		resp["status"] = "ok"
	case GetTelemetry:
		logger.Info("Get telemetry.")
		resp = createTelemetryPayload()
	case GetPins:
		logger.Info("Get pins.")
		resp["pins"] = gpioCol.ReadPins()
	case SetPins:
		if pinMap, ok := params.(map[string]interface{}); ok {
			gpioCol.SetPins(pinMap)
			resp["setPins"] = "ok"
		} else {
			resp["setPins"] = "failed"
		}
		gpioCol.Collect()
		sendTelemetry()
	case SetPin:
		if pinMap, ok := params.(map[string]interface{}); ok {
			gpioCol.SetPins(pinMap)
			resp["setPin"] = "ok"
		} else {
			resp["setPin"] = "failed"
		}
		gpioCol.Collect()
		sendTelemetry()
	case Help:
		resp["cmds"] = [...]string{string(Help), string(GetStatus), string(GetTelemetry), string(GetPins), string(SetPins), string(SetPin)}
	default:
		logger.Warningf("Unknown cmd %s", cmd)
	}

	token := mqttClient.Publish(config.Telemetry.Server.ResponseTopic+requestId, 0, false, serialize(resp))
	if token.WaitTimeout(5 * time.Second) {
		logger.Debug("RPC Response sent %+v\n", resp)
	} else {
		logger.Warningf("Timeout on response sending")
	}

}
