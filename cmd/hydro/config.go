package main

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/caarlos0/env"
)

type config struct {
	Listen        string        `env:"LISTEN" envDefault:"localhost:9000"`
	Timeout       time.Duration `env:"TIMEOUT" envDefault:"10ms"`
	LogLevel      string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFmt        string        `env:"LOG_FMT" envDefault:"console"`
	StoreTimeFile string        `env:"ST_FILE" envDefault:"./tmp/time"`

	MqttBroker     string `env:"MQTT_BROKER" envDefault:"mqtt://localhost:1883"`
	InfluxDBURL    string `env:"INFLUX_URL" envDefault:"http://localhost:8086"`
	InfluxDBToken  string `env:"INFLUX_TOKEN"`
	InfluxDBOrg    string `env:"INFLUX_ORG"  envDefault:"kara"`
	InfluxDBBucket string `env:"INFLUX_BUCKET"  envDefault:"hydroponic"`
}

func load() (*config, error) {
	log.Debug().Msg("loading configuration")
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
