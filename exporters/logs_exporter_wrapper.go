package exporters

import (
	"context"
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type LogsExporter interface {
	Export(ctx context.Context, records []sdklog.Record) error
	Shutdown(ctx context.Context) error
	ForceFlush(ctx context.Context) error
}

type SessionRecorderLogsExporterWrapper struct {
	exporter LogsExporter
}

func NewSessionRecorderLogsExporterWrapper(exporter LogsExporter) *SessionRecorderLogsExporterWrapper {
	return &SessionRecorderLogsExporterWrapper{
		exporter: exporter,
	}
}

type filteredLogRecord struct {
	sdklog.Record
	filteredAttributes []log.KeyValue
}

func (flr *filteredLogRecord) WalkAttributes(fn func(log.KeyValue) bool) {
	if flr.filteredAttributes == nil {
		var allAttrs []log.KeyValue
		flr.Record.WalkAttributes(func(kv log.KeyValue) bool {
			allAttrs = append(allAttrs, kv)
			return true
		})
		
		flr.filteredAttributes = make([]log.KeyValue, 0, len(allAttrs))
		for _, attr := range allAttrs {
			if !strings.HasPrefix(string(attr.Key), constants.MULTIPLAYER_ATTRIBUTE_PREFIX) {
				flr.filteredAttributes = append(flr.filteredAttributes, attr)
			}
		}
	}
	
	for _, attr := range flr.filteredAttributes {
		if !fn(attr) {
			break
		}
	}
}

func (flr *filteredLogRecord) AttributesLen() int {
	if flr.filteredAttributes == nil {
		count := 0
		flr.Record.WalkAttributes(func(kv log.KeyValue) bool {
			if !strings.HasPrefix(string(kv.Key), constants.MULTIPLAYER_ATTRIBUTE_PREFIX) {
				count++
			}
			return true
		})
		return count
	}
	return len(flr.filteredAttributes)
}

func (w *SessionRecorderLogsExporterWrapper) Export(ctx context.Context, records []sdklog.Record) error {
	return w.exporter.Export(ctx, records)
}

func (w *SessionRecorderLogsExporterWrapper) Shutdown(ctx context.Context) error {
	return w.exporter.Shutdown(ctx)
}

func (w *SessionRecorderLogsExporterWrapper) ForceFlush(ctx context.Context) error {
	return w.exporter.ForceFlush(ctx)
}
