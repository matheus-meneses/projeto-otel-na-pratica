package telemetry

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func InitMeter() (metric.MeterProvider, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(mp)
	return mp, nil
}
