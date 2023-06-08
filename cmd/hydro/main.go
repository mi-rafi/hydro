package main

import (
	"context"
	"fmt"
	"github.com/kara/hydro/internal"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"

	"github.com/xlab/closer"
)

func main() {

	defer closer.Close()

	ctx, cancelCtx := context.WithCancel(context.Background())

	closer.Bind(cancelCtx)

	cfg, err := load()

	if err != nil {
		log.Err(err).Msg("error while loading config")
		return
	}
	if err = initLogger(cfg); err != nil {
		log.Err(err).Msg("error while configure logger")
		return
	}
	log.Info().Msg(` 
	  _  _         _         
     | || |_  _ __| |_ _ ___ 
     | __ | || / _' | '_/ _ \
     |_||_|\_, \__,_|_| \___/ v0.0.1-beta
            |__/
`)
	log.Debug().Msg("logger initialized")

	log.Debug().Msg("starting di container")

	log.Debug().Msg("starting db")
	var connCLoser func()
	a, connCLoser, err := initWebApp(ctx, cfg)
	if err != nil {
		log.Err(err).Msg("can not init api")
		return
	}
	closer.Bind(connCLoser)

	log.Debug().Msg("starting web application")
	if err = a.Run(); err != nil {
		log.Err(err).Msg("error while starting api app")
	}
}

func initMqttConfig(c *config) *internal.MqttConfig {
	return &internal.MqttConfig{
		MqttBroker: c.MqttBroker,
	}
}

func initTimeConfig(c *config) *internal.FileTimeLoaderConfig {
	return &internal.FileTimeLoaderConfig{
		FileName: c.StoreTimeFile,
	}
}

func initDbConfig(c *config) *internal.InfluxConfig {
	return &internal.InfluxConfig{
		InfluxDBURL:          c.InfluxDBURL,
		InfluxDBToken:        c.InfluxDBToken,
		InfluxDBOrganization: c.InfluxDBOrg,
		InfluxDBBucket:       c.InfluxDBBucket,
	}
}

func initWebAppCfg(c *config) (internal.AppConfig, error) {
	return internal.AppConfig{Timeout: c.Timeout, NetInterface: c.Listen}, nil
}

func initLogger(c *config) error {
	log.Debug().Msg("initialize logger")
	logLvl, err := zerolog.ParseLevel(strings.ToLower(c.LogLevel))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(logLvl)
	switch c.LogFmt {
	case "console":
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	case "json":
	default:
		return fmt.Errorf("unknown output format %s", c.LogFmt)
	}
	return nil
}
