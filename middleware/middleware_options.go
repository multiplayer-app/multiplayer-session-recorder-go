package middleware

import (
	"strings"
	
	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/trace"
)

type MaskBodyFunc func(body interface{}, span trace.Span) interface{}

type MaskHeadersFunc func(headers interface{}, span trace.Span) interface{}

type MiddlewareOptions struct {
	MaxPayloadSizeBytes   *int
	UncompressPayload     *bool
	
	CaptureHeaders        *bool
	CaptureBody          *bool
	
	IsMaskBodyEnabled     *bool
	IsMaskHeadersEnabled  *bool
	
	MaskBody             MaskBodyFunc
	MaskHeaders          MaskHeadersFunc
	
	MaskBodyFieldsList   []string
	MaskHeadersList      []string
	
	HeadersToInclude     []string
	HeadersToExclude     []string
}

type Option func(*MiddlewareOptions)

func boolPtr(b bool) *bool { return &b }

func intPtr(i int) *int { return &i }

func NewMiddlewareOptions(options ...Option) MiddlewareOptions {
	middleware := &MiddlewareOptions{
		MaxPayloadSizeBytes:  intPtr(constants.MULTIPLAYER_MAX_HTTP_REQUEST_RESPONSE_SIZE),
		UncompressPayload:    boolPtr(false),
		
		CaptureHeaders:       boolPtr(true),
		CaptureBody:         boolPtr(true),
		
		IsMaskBodyEnabled:    boolPtr(true),
		IsMaskHeadersEnabled: boolPtr(true),
		
		MaskBodyFieldsList:   []string{},
		MaskHeadersList:      []string{},
		
		HeadersToInclude:     []string{},
		HeadersToExclude:     []string{},
	}

	for _, opt := range options {
		opt(middleware)
	}

	return *middleware
}

func WithMaxPayloadSizeBytes(size int) Option {
	return func(c *MiddlewareOptions) {
		c.MaxPayloadSizeBytes = intPtr(size)
	}
}

func WithUncompressPayload(uncompress bool) Option {
	return func(c *MiddlewareOptions) {
		c.UncompressPayload = boolPtr(uncompress)
	}
}

func WithCaptureHeaders(capture bool) Option {
	return func(c *MiddlewareOptions) {
		c.CaptureHeaders = boolPtr(capture)
	}
}

func WithCaptureBody(capture bool) Option {
	return func(c *MiddlewareOptions) {
		c.CaptureBody = boolPtr(capture)
	}
}

func WithMaskBodyEnabled(enabled bool) Option {
	return func(c *MiddlewareOptions) {
		c.IsMaskBodyEnabled = boolPtr(enabled)
	}
}

func WithMaskHeadersEnabled(enabled bool) Option {
	return func(c *MiddlewareOptions) {
		c.IsMaskHeadersEnabled = boolPtr(enabled)
	}
}

func WithMaskBodyFunc(maskFunc MaskBodyFunc) Option {
	return func(c *MiddlewareOptions) {
		c.MaskBody = maskFunc
	}
}

func WithMaskHeadersFunc(maskFunc MaskHeadersFunc) Option {
	return func(c *MiddlewareOptions) {
		c.MaskHeaders = maskFunc
	}
}

func WithMaskBodyFieldsList(fields []string) Option {
	return func(c *MiddlewareOptions) {
		c.MaskBodyFieldsList = fields
	}
}

func WithMaskHeadersList(headers []string) Option {
	return func(c *MiddlewareOptions) {
		normalizedHeaders := make([]string, len(headers))
		for i, header := range headers {
			normalizedHeaders[i] = strings.ToLower(header)
		}
		c.MaskHeadersList = normalizedHeaders
	}
}

func WithHeadersToInclude(headers []string) Option {
	return func(c *MiddlewareOptions) {
		normalizedHeaders := make([]string, len(headers))
		for i, header := range headers {
			normalizedHeaders[i] = strings.ToLower(header)
		}
		c.HeadersToInclude = normalizedHeaders
	}
}

func WithHeadersToExclude(headers []string) Option {
	return func(c *MiddlewareOptions) {
		normalizedHeaders := make([]string, len(headers))
		for i, header := range headers {
			normalizedHeaders[i] = strings.ToLower(header)
		}
		c.HeadersToExclude = normalizedHeaders
	}
}
