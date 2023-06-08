package internal

import "time"

type SensorData struct {
	Light         float64   `json:"light"`
	SoilMoisture  float64   `json:"soilMoisture"`
	PH            float64   `json:"pH"`
	MinWaterLevel bool      `json:"minWaterLevel"`
	Timestamp     time.Time `json:"ts"`
}

type LightState struct {
	IsUp bool `json:"isUp"`
}

type PhState struct {
	IsUp bool `json:"isUp"`
}

type MqttError struct {
	Err string `json:"err"`
}
