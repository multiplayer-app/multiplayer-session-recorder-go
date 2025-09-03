package exporters

import (
	"context"
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/sdk/log"
)

type SessionRecorderGrpcLogsExporter struct {
	exporter      *otlploggrpc.Exporter
	debugPrefixes []string
}

func NewSessionRecorderGrpcLogsExporter(apiKey string, endpoint ...string) (*SessionRecorderGrpcLogsExporter, error) {
	headers := map[string]string{
		"Authorization": apiKey,
	}

	endpointURL := constants.MULTIPLAYER_OTEL_DEFAULT_LOGS_EXPORTER_GRPC_URL
	if len(endpoint) > 0 && endpoint[0] != "" {
		endpointURL = endpoint[0]
	}

	exporter, err := otlploggrpc.New(context.Background(),
		otlploggrpc.WithEndpointURL(endpointURL),
		otlploggrpc.WithHeaders(headers),
	)
	if err != nil {
		return nil, err
	}

	return &SessionRecorderGrpcLogsExporter{
		exporter: exporter,
		debugPrefixes: []string{
			constants.MULTIPLAYER_TRACE_CONTINUOUS_DEBUG_PREFIX,
			constants.MULTIPLAYER_TRACE_DEBUG_PREFIX,
		},
	}, nil
}

func (e *SessionRecorderGrpcLogsExporter) Export(ctx context.Context, records []log.Record) error {
	var filteredRecords []log.Record

	for _, record := range records {
		traceID := record.TraceID().String()
		shouldExport := false

		if traceID != "00000000000000000000000000000000" {
			for _, prefix := range e.debugPrefixes {
				if strings.HasPrefix(traceID, prefix) {
					shouldExport = true
					break
				}
			}
		}

		if shouldExport {
			filteredRecords = append(filteredRecords, record)
		}
	}

	if len(filteredRecords) > 0 {
		return e.exporter.Export(ctx, filteredRecords)
	}

	return nil
}

func (e *SessionRecorderGrpcLogsExporter) Shutdown(ctx context.Context) error {
	return e.exporter.Shutdown(ctx)
}

func (e *SessionRecorderGrpcLogsExporter) ForceFlush(ctx context.Context) error {
	return e.exporter.ForceFlush(ctx)
}
