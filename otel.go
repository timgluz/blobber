package main

import (
	"context"
	"errors"
	"regexp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

type shutdownFunc func(context.Context) error

func setupOpenTelemetry(ctx context.Context) (shutdownFunc, error) {
	var shutdownFuncs []func(context.Context) error
	var err error

	shutdown := func(ctx context.Context) error {
		var err error
		for _, shutdownFunc := range shutdownFuncs {
			if shutdownErr := shutdownFunc(ctx); shutdownErr != nil {
				err = shutdownErr
			}
		}
		return err
	}

	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	res, err := newResource()
	if err != nil {
		handleErr(err)
		return nil, err
	}

	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		handleErr(err)
		return nil, err
	}

	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)

	otel.SetMeterProvider(meterProvider)
	return shutdown, nil
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(appName),
			semconv.ServiceVersion(appVersion),
		),
	)
}

func newMeterProvider(ctx context.Context,
	res *resource.Resource,
) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	// custom view to drop http.server.request.size and http.server.response.size metrics
	re := regexp.MustCompile(`http\.sercer\.(request|response)\.size`)
	var dropMetricView metric.View = func(i metric.Instrument) (metric.Stream, bool) {
		s := metric.Stream{
			Name:        i.Name,
			Description: i.Description,
			Unit:        i.Unit,
			Aggregation: metric.AggregationDrop{},
		}

		if re.MatchString(i.Name) {
			return s, true
		}
		return s, false
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(metricExporter),
		),
		metric.WithView(dropMetricView),
	)

	return meterProvider, nil

}
