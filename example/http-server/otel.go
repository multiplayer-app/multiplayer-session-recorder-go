package main

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/multiplayer-app/multiplayer-otlp-go/exporters"
	multiplayer "github.com/multiplayer-app/multiplayer-otlp-go/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func setupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	return
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

var globalIdGenerator *multiplayer.SessionRecorderIdGenerator

func newTraceProvider() (*trace.TracerProvider, error) {
	multiplayerOtlpKey := os.Getenv("MULTIPLAYER_OTLP_KEY")
	if multiplayerOtlpKey == "" {
		return nil, errors.New("MULTIPLAYER_OTLP_KEY environment variable is required")
	}

	// Create trace exporter using the new API
	traceExporter, err := exporters.NewSessionRecorderHttpTraceExporter(
		multiplayerOtlpKey,
		getEnv("OTLP_TRACES_ENDPOINT", "https://api.multiplayer.app/v1/traces"),
	)
	if err != nil {
		return nil, err
	}

	// Create ID generator
	globalIdGenerator = multiplayer.NewSessionRecorderIdGenerator()

	// Create sampler
	sampler := multiplayer.NewSampler(trace.TraceIDRatioBased(0.1)) // 10% sampling

	traceProvider := trace.NewTracerProvider(
		trace.WithIDGenerator(globalIdGenerator),
		trace.WithSampler(sampler),
		trace.WithBatcher(traceExporter,
			trace.WithBatchTimeout(time.Second)),
	)
	return traceProvider, nil
}

// getIdGenerator returns the global ID generator for session recorder integration
func getIdGenerator() *multiplayer.SessionRecorderIdGenerator {
	return globalIdGenerator
}
