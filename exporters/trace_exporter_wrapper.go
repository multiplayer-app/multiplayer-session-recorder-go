package exporters

import (
	"context"
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
)

type SessionRecorderTraceExporterWrapper struct {
	exporter trace.SpanExporter
}

func NewSessionRecorderTraceExporterWrapper(exporter trace.SpanExporter) *SessionRecorderTraceExporterWrapper {
	return &SessionRecorderTraceExporterWrapper{
		exporter: exporter,
	}
}

type filteredSpan struct {
	trace.ReadOnlySpan
	filteredAttributes []attribute.KeyValue
}

func (fs *filteredSpan) Attributes() []attribute.KeyValue {
	if fs.filteredAttributes == nil {
		originalAttrs := fs.ReadOnlySpan.Attributes()
		fs.filteredAttributes = make([]attribute.KeyValue, 0, len(originalAttrs))
		
		for _, attr := range originalAttrs {
			if !strings.HasPrefix(string(attr.Key), constants.MULTIPLAYER_ATTRIBUTE_PREFIX) {
				fs.filteredAttributes = append(fs.filteredAttributes, attr)
			}
		}
	}
	return fs.filteredAttributes
}

func (w *SessionRecorderTraceExporterWrapper) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	filteredSpans := make([]trace.ReadOnlySpan, len(spans))
	
	for i, span := range spans {
		filteredSpans[i] = &filteredSpan{
			ReadOnlySpan: span,
		}
	}
	
	return w.exporter.ExportSpans(ctx, filteredSpans)
}

func (w *SessionRecorderTraceExporterWrapper) Shutdown(ctx context.Context) error {
	return w.exporter.Shutdown(ctx)
}
