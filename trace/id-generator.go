package multiplayer

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"sync"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"github.com/multiplayer-app/multiplayer-otlp-go/types"
	otelTrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type SessionRecorderIdGenerator struct {
	sessionShortId  string
	sessionType     types.SessionType
	generateLongId  func() string
	generateShortId func() string
	randSource      *rand.Rand
	mutex           sync.Mutex
}

var _ otelTrace.IDGenerator = &SessionRecorderIdGenerator{}

func NewSessionRecorderIdGenerator() *SessionRecorderIdGenerator {
	var rngSeed int64
	_ = binary.Read(crand.Reader, binary.LittleEndian, &rngSeed)
	randSource := rand.New(rand.NewSource(rngSeed))

	return &SessionRecorderIdGenerator{
		sessionShortId:  "",
		sessionType:     types.SESSION_TYPE_MANUAL,
		generateLongId:  getIdGenerator(16, randSource), // 16 bytes = 32 hex chars
		generateShortId: getIdGenerator(8, randSource),  // 8 bytes = 16 hex chars
		randSource:      randSource,
	}
}

func getIdGenerator(byteLength int, randSource *rand.Rand) func() string {
	return func() string {
		bytes := make([]byte, byteLength)
		randSource.Read(bytes)
		return hex.EncodeToString(bytes)
	}
}

func (gen *SessionRecorderIdGenerator) generateTraceId() string {
	traceId := gen.generateLongId()

	if gen.sessionShortId != "" {
		var sessionTypePrefix string
		switch gen.sessionType {
		case types.SESSION_TYPE_CONTINUOUS:
			sessionTypePrefix = constants.MULTIPLAYER_TRACE_CONTINUOUS_DEBUG_PREFIX
		default:
			sessionTypePrefix = constants.MULTIPLAYER_TRACE_DEBUG_PREFIX
		}

		prefix := sessionTypePrefix + gen.sessionShortId

		if len(prefix) < len(traceId) {
			sessionTraceId := prefix + traceId[len(prefix):]
			return sessionTraceId
		}
	}

	return traceId
}

func (gen *SessionRecorderIdGenerator) generateSpanId() string {
	return gen.generateShortId()
}

func (gen *SessionRecorderIdGenerator) SetSessionId(sessionShortId string, sessionType types.SessionType) {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	gen.sessionShortId = sessionShortId
	gen.sessionType = sessionType
}

func (gen *SessionRecorderIdGenerator) NewIDs(ctx context.Context) (trace.TraceID, trace.SpanID) {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	var tid trace.TraceID
	var sid trace.SpanID

	if gen.sessionShortId != "" {
		traceIdHex := gen.generateTraceId()
		traceIdBytes, err := hex.DecodeString(traceIdHex)
		if err == nil && len(traceIdBytes) == 16 {
			copy(tid[:], traceIdBytes)
		}

		if !tid.IsValid() {
			for {
				binary.NativeEndian.PutUint64(tid[:8], gen.randSource.Uint64())
				binary.NativeEndian.PutUint64(tid[8:], gen.randSource.Uint64())
				if tid.IsValid() {
					break
				}
			}
		}
	} else {
		for {
			binary.NativeEndian.PutUint64(tid[:8], gen.randSource.Uint64())
			binary.NativeEndian.PutUint64(tid[8:], gen.randSource.Uint64())
			if tid.IsValid() {
				break
			}
		}
	}

	for {
		binary.NativeEndian.PutUint64(sid[:], gen.randSource.Uint64())
		if sid.IsValid() {
			break
		}
	}

	return tid, sid
}

func (gen *SessionRecorderIdGenerator) NewSpanID(ctx context.Context, traceID trace.TraceID) trace.SpanID {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()

	var sid trace.SpanID
	for {
		binary.NativeEndian.PutUint64(sid[:], gen.randSource.Uint64())
		if sid.IsValid() {
			break
		}
	}

	return sid
}
