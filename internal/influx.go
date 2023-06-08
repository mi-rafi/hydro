package internal

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/rs/zerolog/log"
	"time"
)

type HydroponicRepo interface {
	GetLastData(ctx context.Context, start, end time.Time) ([]SensorData, error)
}

type HydroponicInfluxRepo struct {
	cli    influxdb2.Client
	bucket string
	org    string
}

type InfluxConfig struct {
	InfluxDBURL          string
	InfluxDBToken        string
	InfluxDBOrganization string
	InfluxDBBucket       string
}

func NewHydroponicRepo(ctx context.Context, cfg *InfluxConfig) (*HydroponicInfluxRepo, func(), error) {
	influxClient := influxdb2.NewClient(cfg.InfluxDBURL, cfg.InfluxDBToken)
	s, err := influxClient.Health(ctx)
	if err != nil {
		return nil, nil, err
	}
	log.Info().Str("health status", string(s.Status)).Msg("healthcheck influxdb")
	h := &HydroponicInfluxRepo{influxClient, cfg.InfluxDBBucket, cfg.InfluxDBOrganization}
	return h, h.Close, nil
}

func (h *HydroponicInfluxRepo) GetLastData(ctx context.Context, start, end time.Time) ([]SensorData, error) {
	query := fmt.Sprintf(`
		from(bucket:"%s")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r._measurement == "sensors")
	`, h.bucket, start.Format(time.RFC3339), end.Format(time.RFC3339))

	queryAPI := h.cli.QueryAPI(h.org)
	result, err := queryAPI.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func(result *api.QueryTableResult) {
		err := result.Close()
		if err != nil {
			log.Err(err)
		}
	}(result)

	if result.Err() != nil {
		return nil, result.Err()
	}

	resultPoints := make([]SensorData, 0)

	for result.Next() {
		if result.TableChanged() {
			log.Debug().Msgf("table: %s", result.TableMetadata().String())
		}

		s := SensorData{Timestamp: result.Record().Time()}

		switch field := result.Record().Field(); field {
		case "ph":
			s.PH = result.Record().Value().(float64)
		case "light":
			s.Light = result.Record().Value().(float64)
		case "soil":
			s.SoilMoisture = result.Record().Value().(float64)
		case "lvl":
			s.MinWaterLevel = result.Record().Value().(bool)
		default:
			log.Warn().Msgf("unrecognized field %s.", field)
		}

		resultPoints = append(resultPoints, s)
	}

	return resultPoints, nil

}

func (h *HydroponicInfluxRepo) Close() {
	h.cli.Close()
}
