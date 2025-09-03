package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"github.com/multiplayer-app/multiplayer-otlp-go/sdk"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func WithRequestData(h http.Handler, options MiddlewareOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		traceId := span.SpanContext().TraceID().String()
		if !sdk.IsMultiplayerTrace(traceId) {
			h.ServeHTTP(w, r)
			return
		}

		if options.CaptureHeaders != nil && *options.CaptureHeaders {
			headers := processHeaders(r.Header, options, span)
			span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_REQUEST_HEADERS, headers))
		}

		if options.CaptureBody != nil && *options.CaptureBody && r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				h.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			
			if len(bodyBytes) != 0 {
				processedBody := processBody(bodyBytes, options, span)
				maxSize := constants.MULTIPLAYER_MAX_HTTP_REQUEST_RESPONSE_SIZE
				if options.MaxPayloadSizeBytes != nil {
					maxSize = *options.MaxPayloadSizeBytes
				}
				truncatedBody := sdk.TruncateIfNeeded(processedBody, maxSize)
				span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_REQUEST_BODY, truncatedBody))
			}
		}
		h.ServeHTTP(w, r)
	})
}

func WithResponseData(next http.Handler, options MiddlewareOptions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		traceId := span.SpanContext().TraceID().String()
		if !sdk.IsMultiplayerTrace(traceId) {
			next.ServeHTTP(w, r)
			return
		}

		isDebugTrace := sdk.IsDebugTrace(traceId)
		if isDebugTrace {
			w.Header().Set("X-Trace-Id", traceId)
		}

		rww := NewResponseWriterWrapper(w)
		defer func() {
			if options.CaptureHeaders != nil && *options.CaptureHeaders {
				headers := processHeaders(w.Header(), options, span)
				span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_RESPONSE_HEADERS, headers))
			}
			
			if options.CaptureBody != nil && *options.CaptureBody {
				bodyBytes := rww.GetBody()
				if len(bodyBytes) != 0 {
					processedBody := processBody(bodyBytes, options, span)
					maxSize := constants.MULTIPLAYER_MAX_HTTP_REQUEST_RESPONSE_SIZE
					if options.MaxPayloadSizeBytes != nil {
						maxSize = *options.MaxPayloadSizeBytes
					}
					truncatedBody := sdk.TruncateIfNeeded(processedBody, maxSize)
					span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_RESPONSE_BODY, truncatedBody))
				}
			}
		}()
		next.ServeHTTP(rww, r)
	})
}

// processHeaders processes headers according to the middleware options
func processHeaders(headers http.Header, options MiddlewareOptions, span trace.Span) string {
	filteredHeaders := filterHeaders(headers, options)
	
	if options.MaskHeaders != nil {
		masked := options.MaskHeaders(filteredHeaders, span)
		return convertToString(masked)
	}
	
	if options.IsMaskHeadersEnabled != nil && *options.IsMaskHeadersEnabled {
		maskList := options.MaskHeadersList
		if len(maskList) == 0 {
			// Use default sensitive headers from SDK
			maskList = []string{"authorization", "cookie", "x-api-key", "bearer"}
		}
		return maskHeadersWithList(filteredHeaders, maskList)
	}
	
	return convertToString(filteredHeaders)
}

// processBody processes body according to the middleware options
func processBody(bodyBytes []byte, options MiddlewareOptions, span trace.Span) string {
	body := string(bodyBytes)
	
	if options.MaskBody != nil {
		masked := options.MaskBody(body, span)
		return convertToString(masked)
	}
	
	if options.IsMaskBodyEnabled != nil && *options.IsMaskBodyEnabled {
		if sdk.IsDebugTrace(span.SpanContext().TraceID().String()) {
			return body
		}
	}
	
	return body
}

func filterHeaders(headers http.Header, options MiddlewareOptions) http.Header {
	filtered := make(http.Header)
	
	for name, values := range headers {
		lowerName := strings.ToLower(name)
		
		if len(options.HeadersToExclude) > 0 {
			excluded := false
			for _, exclude := range options.HeadersToExclude {
				if lowerName == exclude {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		
		if len(options.HeadersToInclude) > 0 {
			included := false
			for _, include := range options.HeadersToInclude {
				if lowerName == include {
					included = true
					break
				}
			}
			if !included {
				continue
			}
		}
		
		filtered[name] = values
	}
	
	return filtered
}

func maskHeadersWithList(headers http.Header, maskList []string) string {
	masked := make(map[string]interface{})
	
	for name, values := range headers {
		lowerName := strings.ToLower(name)
		shouldMask := false
		
		for _, maskHeader := range maskList {
			if lowerName == maskHeader {
				shouldMask = true
				break
			}
		}
		
		if shouldMask {
			masked[name] = constants.MASK_PLACEHOLDER
		} else {
			masked[name] = values
		}
	}
	
	return convertToString(masked)
}

func convertToString(data interface{}) string {
	if data == nil {
		return ""
	}
	
	switch v := data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		if bytes, err := json.Marshal(data); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", data)
	}
}
