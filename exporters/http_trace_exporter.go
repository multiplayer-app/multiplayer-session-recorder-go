package exporters

import (
	"context"
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"
)

type SessionRecorderHttpTraceExporter struct {
	exporter *otlptrace.Exporter
	debugPrefixes []string
}

func NewSessionRecorderHttpTraceExporter(apiKey string, endpoint ...string) (*SessionRecorderHttpTraceExporter, error) {
	headers := map[string]string{
		"Authorization": apiKey,
	}

	endpointURL := constants.MULTIPLAYER_OTEL_DEFAULT_TRACES_EXPORTER_HTTP_URL
	if len(endpoint) > 0 && endpoint[0] != "" {
		endpointURL = endpoint[0]
	}

	client := otlptracehttp.NewClient(
		otlptracehttp.WithEndpointURL(endpointURL),
		otlptracehttp.WithHeaders(headers),
	)

	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, err
	}

	return &SessionRecorderHttpTraceExporter{
		exporter: exporter,
		debugPrefixes: []string{
			constants.MULTIPLAYER_TRACE_CONTINUOUS_DEBUG_PREFIX,
			constants.MULTIPLAYER_TRACE_DEBUG_PREFIX,
		},
	}, nil
}

// ExportSpans exports spans that have trace IDs starting with debug prefixes
func (e *SessionRecorderHttpTraceExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	var filteredSpans []trace.ReadOnlySpan

	for _, span := range spans {
		traceID := span.SpanContext().TraceID().String()
		shouldExport := false

		for _, prefix := range e.debugPrefixes {
			if strings.HasPrefix(traceID, prefix) {
				shouldExport = true
				break
			}
		}

		if shouldExport {
			filteredSpans = append(filteredSpans, span)
		}
	}

	if len(filteredSpans) > 0 {
		return e.exporter.ExportSpans(ctx, filteredSpans)
	}

	return nil
}

func (e *SessionRecorderHttpTraceExporter) Shutdown(ctx context.Context) error {
	return e.exporter.Shutdown(ctx)
}
