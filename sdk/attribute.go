package sdk

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AttributeOptions defines options for attribute setting functions
type AttributeOptions struct {
	Mask bool
}

// DefaultAttributeOptions returns default options with masking enabled
func DefaultAttributeOptions() AttributeOptions {
	return AttributeOptions{Mask: true}
}

// SetAttribute sets an attribute on the current active span
func SetAttribute(ctx context.Context, key string, value attribute.Value) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(attribute.KeyValue{
		Key:   attribute.Key(key),
		Value: value,
	})
}

// SetHttpRequestBody sets the HTTP request body attribute on the current active span
func SetHttpRequestBody(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_REQUEST_BODY, value))
}

func SetHttpRequestHeaders(ctx context.Context, headers interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := MaskHeaders(headers, span)
		value = convertToString(masked)
	} else {
		value = convertToString(headers)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_REQUEST_HEADERS, value))
}

func SetHttpResponseBody(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_RESPONSE_BODY, value))
}

func SetHttpResponseHeaders(ctx context.Context, headers interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := MaskHeaders(headers, span)
		value = convertToString(masked)
	} else {
		value = convertToString(headers)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_HTTP_RESPONSE_HEADERS, value))
}

func SetMessageBody(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_MESSAGING_MESSAGE_BODY, value))
}

func SetRpcRequestMessage(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_RPC_REQUEST_MESSAGE, value))
}

func SetRpcResponseMessage(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_RPC_RESPONSE_MESSAGE, value))
}

func SetGrpcRequestMessage(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_GRPC_REQUEST_MESSAGE, value))
}

func SetGrpcResponseMessage(ctx context.Context, body interface{}, options ...AttributeOptions) {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	opts := DefaultAttributeOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	var value string
	if opts.Mask {
		masked := Mask(body, span)
		value = convertToString(masked)
	} else {
		value = convertToString(body)
	}

	span.SetAttributes(attribute.String(constants.ATTR_MULTIPLAYER_GRPC_RESPONSE_MESSAGE, value))
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
	case fmt.Stringer:
		return v.String()
	default:
		if bytes, err := json.Marshal(data); err == nil {
			return string(bytes)
		}
		return fmt.Sprintf("%v", data)
	}
}
