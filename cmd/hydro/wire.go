//go:build wireinject
// +build wireinject

package main

import (
	"HydroponicSystem/internal"
	"context"
	"github.com/google/wire"
)

var (
	clientSetter = wire.NewSet(
		initMqttConfig,
		wire.Bind(
			new(internal.HydroponicClient),
			new(*internal.MqttHydroponicClient),
		),
		internal.NewMqttHydroponicClient,
	)

	dbSetter = wire.NewSet(
		initDbConfig,
		wire.Bind(
			new(internal.HydroponicRepo),
			new(*internal.HydroponicInfluxRepo),
		),
		internal.NewHydroponicRepo,
	)

	timeSetter = wire.NewSet(
		initTimeConfig,
		wire.Bind(
			new(internal.TimeLoader),
			new(*internal.FileTimeLoader),
		),
	)
)

func initWebApp(ctx context.Context, c *config) (*internal.API, func(), error) {
	wire.Build(initWebAppCfg, timeSetter, clientSetter, dbSetter, internal.NewApp)
	return nil, nil, nil
}
