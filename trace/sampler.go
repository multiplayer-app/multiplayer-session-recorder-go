package multiplayer

import (
	"github.com/multiplayer-app/multiplayer-otlp-go/sdk"
	"go.opentelemetry.io/otel/sdk/trace"
	trace_ "go.opentelemetry.io/otel/trace"
)

type traceIDBasedSampler struct {
	sampler     trace.Sampler
	description string
}

var _ trace.Sampler = &traceIDBasedSampler{}

func NewSampler(baseSampler trace.Sampler) trace.Sampler {
	return &traceIDBasedSampler{
		sampler:     baseSampler,
		description: "SessionRecorderTraceIDBasedSampler_" + baseSampler.Description(),
	}
}

func (ts traceIDBasedSampler) ShouldSample(p trace.SamplingParameters) trace.SamplingResult {
	if sdk.IsMultiplayerTrace(p.TraceID.String()) {
		return trace.SamplingResult{
			Decision:   trace.RecordAndSample,
			Tracestate: trace_.SpanContextFromContext(p.ParentContext).TraceState(),
		}
	}

	return ts.sampler.ShouldSample(p)
}

func (ts traceIDBasedSampler) Description() string {
	return ts.description
}
