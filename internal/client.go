package internal

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

type HydroponicClient interface {
	SendUpPh()
	SendDownPh()
	SendAddSoil()
	SendAddWater()
	SendChangeLight()
	GetLightState() bool
}

const mqttPhUpTopic = "hydroponic/phUp"
const mqttPhDownTopic = "hydroponic/phDown"
const mqttLightTopic = "hydroponic/light"
const mqttWaterTopic = "hydroponic/water"
const mqttSoilTopic = "hydroponic/soil"
const mqttErrorTopic = "hydroponic/error"

type MqttHydroponicClient struct {
	cli        mqtt.Client
	lightState bool
}

type MqttConfig struct {
	MqttBroker string
}

func NewMqttHydroponicClient(config MqttConfig) (*MqttHydroponicClient, func(), error) {
	opts := mqtt.NewClientOptions().AddBroker(config.MqttBroker)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Info().Msg("mqtt broker connected")
	})
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, nil, token.Error()
	}
	m := &MqttHydroponicClient{mqttClient, false}
	mqttClient.Subscribe(mqttLightTopic, 1, m.receiveLightState)
	mqttClient.Subscribe(mqttSoilTopic, 1, m.receiveError(mqttSoilTopic))
	mqttClient.Subscribe(mqttWaterTopic, 1, m.receiveError(mqttWaterTopic))
	mqttClient.Subscribe(mqttPhUpTopic, 1, m.receiveError(mqttPhUpTopic))
	mqttClient.Subscribe(mqttPhDownTopic, 1, m.receiveError(mqttPhDownTopic))
	mqttClient.Subscribe(mqttErrorTopic, 1, m.receiveError(mqttErrorTopic))
	return m, m.Close, nil
}

func (m *MqttHydroponicClient) receiveLightState(_ mqtt.Client, message mqtt.Message) {
	defer message.Ack()
	var ls LightState
	err := json.Unmarshal(message.Payload(), &ls)
	if err != nil {
		log.Error().Msg("can not unmarshall light state")
		return
	}
	m.lightState = ls.IsUp
}

func (m *MqttHydroponicClient) receiveError(topic string) func(_ mqtt.Client, message mqtt.Message) {
	return func(_ mqtt.Client, message mqtt.Message) {
		defer message.Ack()
		var e MqttError
		err := json.Unmarshal(message.Payload(), &e)
		if err != nil {
			log.Error().Msgf("can not unmarshall error from topic %s", topic)
			return
		}
		log.Error().Msgf("receive error %s from topic %s", e.Err, topic)
	}
}

func (m *MqttHydroponicClient) Close() {
	m.cli.Disconnect(10)
}

func (m *MqttHydroponicClient) SendUpPh() {
	p := m.cli.Publish(mqttPhUpTopic, 1, false, nil)
	go handleTopicError(mqttPhUpTopic, p)
}

func (m *MqttHydroponicClient) SendDownPh() {
	p := m.cli.Publish(mqttPhDownTopic, 1, false, nil)
	go handleTopicError(mqttPhDownTopic, p)
}

func (m *MqttHydroponicClient) SendAddSoil() {
	p := m.cli.Publish(mqttSoilTopic, 1, false, nil)
	go handleTopicError(mqttSoilTopic, p)
}

func (m *MqttHydroponicClient) SendAddWater() {
	p := m.cli.Publish(mqttWaterTopic, 1, false, nil)
	go handleTopicError(mqttWaterTopic, p)
}

func (m *MqttHydroponicClient) SendChangeLight() {
	p := m.cli.Publish(mqttLightTopic, 1, false, nil)
	go handleTopicError(mqttLightTopic, p)
}

func (m *MqttHydroponicClient) GetLightState() *LightState {
	return &LightState{m.lightState}
}

func handleTopicError(topic string, t mqtt.Token) {
	t.Done()
	if err := t.Error(); err != nil {
		log.Error().Err(err).Msgf("can not p data to topic %s", topic)
	}
}
