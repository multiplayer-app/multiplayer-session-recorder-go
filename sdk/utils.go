package sdk

import (
	"strings"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
)



func TruncateIfNeeded(data string, maxPayloadSize int) string {
	if len(data) > maxPayloadSize {
		return data[:maxPayloadSize] + "...[TRUNCATED]"
	}
	return data
}

func IsDebugTrace(traceId string) bool {
	return strings.HasPrefix(traceId, constants.MULTIPLAYER_TRACE_DEBUG_PREFIX)
}


func IsMultiplayerTrace(traceId string) bool {
	return IsDebugTrace(traceId)
}
