// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"context"
	"github.com/google/wire"
	"github.com/kara/hydro/internal"
)

// Injectors from wire.go:

func initWebApp(ctx context.Context, c *config) (*internal.API, func(), error) {
	appConfig, err := initWebAppCfg(c)
	if err != nil {
		return nil, nil, err
	}
	mqttConfig := initMqttConfig(c)
	mqttHydroponicClient, cleanup, err := internal.NewMqttHydroponicClient(mqttConfig)
	if err != nil {
		return nil, nil, err
	}
	influxConfig := initDbConfig(c)
	hydroponicInfluxRepo, cleanup2, err := internal.NewHydroponicRepo(ctx, influxConfig)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	fileTimeLoaderConfig := initTimeConfig(c)
	fileTimeLoader, cleanup3, err := internal.NewFileTimeLoader(fileTimeLoaderConfig)
	if err != nil {
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	api, err := internal.NewApp(ctx, appConfig, mqttHydroponicClient, hydroponicInfluxRepo, fileTimeLoader)
	if err != nil {
		cleanup3()
		cleanup2()
		cleanup()
		return nil, nil, err
	}
	return api, func() {
		cleanup3()
		cleanup2()
		cleanup()
	}, nil
}

// wire.go:

var (
	clientSetter = wire.NewSet(
		initMqttConfig, wire.Bind(
			new(internal.HydroponicClient),
			new(*internal.MqttHydroponicClient),
		), internal.NewMqttHydroponicClient,
	)

	dbSetter = wire.NewSet(
		initDbConfig, wire.Bind(
			new(internal.HydroponicRepo),
			new(*internal.HydroponicInfluxRepo),
		), internal.NewHydroponicRepo,
	)

	timeSetter = wire.NewSet(
		initTimeConfig, wire.Bind(
			new(internal.TimeLoader),
			new(*internal.FileTimeLoader),
		), internal.NewFileTimeLoader,
	)
)
