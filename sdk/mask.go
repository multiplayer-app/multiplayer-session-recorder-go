package sdk

import (
	"encoding/json"
	"reflect"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"go.opentelemetry.io/otel/trace"
)

const MAX_DEPTH = 8

var SensitiveFields = []string{
	"password",
	"pass",
	"passwd",
	"pwd",
	"token",
	"access_token",
	"accessToken",
	"refresh_token",
	"refreshToken",
	"secret",
	"api_key",
	"apiKey",
	"authorization",
	"auth_token",
	"authToken",
	"jwt",
	"session_id",
	"sessionId",
	"sessionToken",
	"client_secret",
	"clientSecret",
	"private_key",
	"privateKey",
	"public_key",
	"publicKey",
	"key",
	"encryption_key",
	"encryptionKey",
	"credit_card",
	"creditCard",
	"card_number",
	"cardNumber",
	"cvv",
	"cvc",
	"ssn",
	"sin",
	"pin",
	"security_code",
	"securityCode",
	"bank_account",
	"bankAccount",
	"iban",
	"swift",
	"bic",
	"routing_number",
	"routingNumber",
	"license_key",
	"licenseKey",
	"otp",
	"mfa_code",
	"mfaCode",
	"phone_number",
	"phoneNumber",
	"email",
	"address",
	"dob",
	"tax_id",
	"taxId",
	"passport_number",
	"passportNumber",
	"driver_license",
	"driverLicense",
	"set-cookie",
	"cookie",
	"authorization",
	"proxyAuthorization",
}

var SensitiveHeaders = []string{
	"set-cookie",
	"cookie",
	"authorization",
	"proxyAuthorization",
}

func maskAll(value interface{}, depth int) interface{} {
	if depth > MAX_DEPTH {
		return nil
	}

	v := reflect.ValueOf(value)
	
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = maskAll(v.Index(i).Interface(), depth+1)
		}
		return result
		
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			result[keyStr] = maskAll(v.MapIndex(key).Interface(), depth+1)
		}
		return result
		
	case reflect.String:
		return constants.MASK_PLACEHOLDER
		
	default:
		return value
	}
}

func maskSelected(value interface{}, keysToMask map[string]bool) interface{} {
	v := reflect.ValueOf(value)
	
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = maskSelected(v.Index(i).Interface(), keysToMask)
		}
		return result
		
	case reflect.Map:
		result := make(map[string]interface{})
		for _, key := range v.MapKeys() {
			keyStr := key.String()
			if keysToMask[keyStr] {
				result[keyStr] = constants.MASK_PLACEHOLDER
			} else {
				result[keyStr] = maskSelected(v.MapIndex(key).Interface(), keysToMask)
			}
		}
		return result
		
	default:
		return value
	}
}

type MaskFunc func(value interface{}, span trace.Span) interface{}

// NewMaskFunc creates a new masking function with the specified keys to mask
func NewMaskFunc(keysToMask ...string) MaskFunc {
	return func(value interface{}, span trace.Span) interface{} {
		var payloadData interface{}
		if str, ok := value.(string); ok {
			if err := json.Unmarshal([]byte(str), &payloadData); err != nil {
				payloadData = value
			}
		} else {
			payloadData = value
		}
		
		var maskedData interface{}
		if len(keysToMask) > 0 {
			keySet := make(map[string]bool)
			for _, key := range keysToMask {
				keySet[key] = true
			}
			maskedData = maskSelected(payloadData, keySet)
		} else {
			maskedData = maskAll(payloadData, 0)
		}
		
		if _, ok := value.(string); ok {
			if jsonBytes, err := json.Marshal(maskedData); err == nil {
				return string(jsonBytes)
			}
		}
		
		return maskedData
	}
}

var DefaultMask = NewMaskFunc(SensitiveFields...)

func Mask(value interface{}, span trace.Span) interface{} {
	return DefaultMask(value, span)
}

func MaskWithFields(value interface{}, span trace.Span, fields []string) interface{} {
	maskFunc := NewMaskFunc(fields...)
	return maskFunc(value, span)
}

func MaskHeaders(headers interface{}, span trace.Span) interface{} {
	maskFunc := NewMaskFunc(SensitiveHeaders...)
	return maskFunc(headers, span)
}
