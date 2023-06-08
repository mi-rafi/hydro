package internal

import (
	"encoding/json"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type HydroponicClient interface {
	SendUpPh() error
	SendDownPh() error
	SendAddSoil() error
	SendAddWater() error
	SendChangeLight() error
	GetLightState() *LightState
}

type Command uint8

const (
	PhUpCommand Command = iota
	PhDownCommand
	LightChangeCommand
	SoilCommand
	AddWaterCommand
)

const mqttCommandTopic = "hydroponic/command"
const mqttLightTopic = "hydroponic/light"
const mqttErrorTopic = "hydroponic/error"

type MqttHydroponicClient struct {
	cli        mqtt.Client
	lightState bool
}

type MqttConfig struct {
	MqttBroker string
}

type MqttError struct {
	Err string `json:"err"`
}

type LightState struct {
	IsUp bool `json:"isUp"`
}

func NewMqttHydroponicClient(config *MqttConfig) (*MqttHydroponicClient, func(), error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MqttBroker)
	opts.SetClientID("hydro_mqtt_client")
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Info().Msg("mqtt broker connected")
	})
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		log.Info().Msgf("Received message: %s from topic: %s", msg.Payload(), msg.Topic())
	})
	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return nil, nil, errors.Wrap(token.Error(), "can not connect to mqtt")
	}
	m := &MqttHydroponicClient{mqttClient, false}
	mqttClient.Subscribe(mqttLightTopic, 1, m.receiveLightState)
	mqttClient.Subscribe(mqttErrorTopic, 1, m.receiveError(mqttErrorTopic))
	return m, m.Close, nil
}

func (m *MqttHydroponicClient) receiveLightState(_ mqtt.Client, message mqtt.Message) {
	defer message.Ack()
	var ls LightState
	err := json.Unmarshal(message.Payload(), &ls)
	if err != nil {
		log.Error().
			Str("payload", string(message.Payload())).
			Uint16("messageId", message.MessageID()).
			Msg("can not unmarshall light state")
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
			log.Error().
				Str("topic", topic).
				Str("payload", string(message.Payload())).
				Uint16("messageId", message.MessageID()).
				Msg("can not unmarshall error")
			return
		}
		log.Error().
			Str("topic", topic).
			Str("payload", string(message.Payload())).
			Str("error", e.Err).
			Uint16("messageId", message.MessageID()).
			Msg("receive error")
	}
}

func (m *MqttHydroponicClient) Close() {
	m.cli.Disconnect(250)
}

type Marshaller[T any] interface {
	Marshall() ([]byte, error)
}

func (m *MqttHydroponicClient) SendUpPh() error {
	return sendCommand(m, PhUpCommand)
}

func (m *MqttHydroponicClient) SendDownPh() error {
	return sendCommand(m, PhDownCommand)
}

func (m *MqttHydroponicClient) SendAddSoil() error {
	return sendCommand(m, SoilCommand)
}

func (m *MqttHydroponicClient) SendAddWater() error {
	return sendCommand(m, AddWaterCommand)
}

func (m *MqttHydroponicClient) SendChangeLight() error {
	return sendCommand(m, LightChangeCommand)
}

func (m *MqttHydroponicClient) GetLightState() *LightState {
	return &LightState{m.lightState}
}

func (c Command) Marshall() ([]byte, error) {
	cmd := struct {
		Cmd Command `json:"command"`
	}{
		c,
	}
	return json.Marshal(cmd)
}

func sendCommand[E Marshaller[any]](m *MqttHydroponicClient, command E) error {
	b, err := command.Marshall()
	if err != nil {
		return err
	}
	p := m.cli.Publish(mqttCommandTopic, 1, false, b)
	go handleTopicError(mqttCommandTopic, p)
	return nil
}

func handleTopicError(topic string, t mqtt.Token) {
	t.Done()
	if err := t.Error(); err != nil {
		log.Error().Err(err).Str("topic", topic).Msg("can not send data to topic")
	} else {
		log.Debug().Msg("message sent")
	}
}
