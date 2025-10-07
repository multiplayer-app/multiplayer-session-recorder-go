package session_recorder

import (
	"errors"
	"fmt"
	"time"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
	"github.com/multiplayer-app/multiplayer-otlp-go/types"
)

type SessionState string

const (
	SessionStateStarted SessionState = "STARTED"
	SessionStateStopped SessionState = "STOPPED"
	SessionStatePaused  SessionState = "PAUSED"
)

type SessionRecorderConfig struct {
	APIKey                        string
	TraceIDGenerator              TraceIDGenerator
	ResourceAttributes            map[string]interface{}
	GenerateSessionShortIDLocally interface{}
	APIBaseURL                    string
}

type TraceIDGenerator interface {
	SetSessionId(sessionShortId string, sessionType types.SessionType)
}

type SessionRecorder struct {
	isInitialized           bool
	shortSessionID          string
	traceIDGenerator        TraceIDGenerator
	sessionType             types.SessionType
	sessionState            SessionState
	apiService              *APIService
	sessionShortIDGenerator func() string
	resourceAttributes      map[string]interface{}
}

func NewSessionRecorder() *SessionRecorder {
	return &SessionRecorder{
		isInitialized:           false,
		shortSessionID:          "",
		sessionType:             types.SESSION_TYPE_MANUAL,
		sessionState:            SessionStateStopped,
		apiService:              NewAPIService(),
		sessionShortIDGenerator: defaultSessionShortIDGenerator,
		resourceAttributes:      make(map[string]interface{}),
	}
}

func (sr *SessionRecorder) Init(config SessionRecorderConfig) error {
	if config.APIKey == "" {
		return errors.New("api key not provided")
	}

	if config.TraceIDGenerator == nil {
		return errors.New("incompatible trace id generator")
	}

	sr.resourceAttributes = config.ResourceAttributes
	if sr.resourceAttributes == nil {
		sr.resourceAttributes = make(map[string]interface{})
	}
	if _, exists := sr.resourceAttributes[constants.ATTR_MULTIPLAYER_SESSION_RECORDER_VERSION]; !exists {
		sr.resourceAttributes[constants.ATTR_MULTIPLAYER_SESSION_RECORDER_VERSION] = getSessionRecorderVersion()
	}

	switch v := config.GenerateSessionShortIDLocally.(type) {
	case func() string:
		sr.sessionShortIDGenerator = v
	case bool:
		if v {
			sr.sessionShortIDGenerator = defaultSessionShortIDGenerator
		}
	}

	sr.traceIDGenerator = config.TraceIDGenerator

	apiConfig := APIServiceConfig{
		APIKey:     config.APIKey,
		APIBaseURL: config.APIBaseURL,
	}
	sr.apiService.Init(apiConfig)

	sr.isInitialized = true
	return nil
}

func (sr *SessionRecorder) Start(sessionType types.SessionType, sessionPayload *Session) error {
	if !sr.isInitialized {
		return errors.New("configuration not initialized. Call Init() before performing any actions")
	}

	if sessionPayload != nil && sessionPayload.ShortID != "" && len(sessionPayload.ShortID) != constants.MULTIPLAYER_TRACE_DEBUG_SESSION_SHORT_ID_LENGTH {
		return errors.New("invalid short session id")
	}

	if sr.sessionState != SessionStateStopped {
		return errors.New("session should be ended before starting new one")
	}

	if sessionPayload == nil {
		sessionPayload = &Session{}
	}

	sr.sessionType = sessionType

	if sessionPayload.Name == "" {
		sessionPayload.Name = fmt.Sprintf("Session on %s", getFormattedDate(time.Now()))
	}

	if sessionPayload.ResourceAttributes == nil {
		sessionPayload.ResourceAttributes = make(map[string]interface{})
	}
	for k, v := range sr.resourceAttributes {
		sessionPayload.ResourceAttributes[k] = v
	}

	var session *Session
	var err error

	if sr.sessionType == types.SESSION_TYPE_CONTINUOUS {
		session, err = sr.apiService.StartContinuousSession(*sessionPayload)
	} else {
		session, err = sr.apiService.StartSession(*sessionPayload)
	}

	if err != nil {
		return err
	}

	if session == nil || session.ShortID == "" {
		return errors.New("failed to start session")
	}

	sr.shortSessionID = session.ShortID
	sr.traceIDGenerator.SetSessionId(sr.shortSessionID, sr.sessionType)
	sr.sessionState = SessionStateStarted

	return nil
}

func (sr *SessionRecorder) Save(sessionData *Session) error {
	if !sr.isInitialized {
		return errors.New("configuration not initialized. Call Init() before performing any actions")
	}

	if sr.sessionState == SessionStateStopped || sr.shortSessionID == "" {
		return errors.New("session should be active or paused")
	}

	if sr.sessionType != types.SESSION_TYPE_CONTINUOUS {
		return errors.New("invalid session type")
	}

	if sessionData == nil {
		sessionData = &Session{}
	}

	if sessionData.Name == "" {
		sessionData.Name = fmt.Sprintf("Session on %s", getFormattedDate(time.Now()))
	}

	return sr.apiService.SaveContinuousSession(sr.shortSessionID, *sessionData)
}

func (sr *SessionRecorder) Stop(sessionData *Session) error {
	defer func() {
		sr.traceIDGenerator.SetSessionId("", types.SESSION_TYPE_MANUAL)
		sr.shortSessionID = ""
		sr.sessionState = SessionStateStopped
	}()

	if !sr.isInitialized {
		return errors.New("configuration not initialized. Call Init() before performing any actions")
	}

	if sr.sessionState == SessionStateStopped || sr.shortSessionID == "" {
		return errors.New("session should be active or paused")
	}

	if sr.sessionType != types.SESSION_TYPE_MANUAL {
		return errors.New("invalid session type")
	}

	if sessionData == nil {
		sessionData = &Session{}
	}

	return sr.apiService.StopSession(sr.shortSessionID, *sessionData)
}

func (sr *SessionRecorder) Cancel() error {
	defer func() {
		sr.traceIDGenerator.SetSessionId("", types.SESSION_TYPE_MANUAL)
		sr.shortSessionID = ""
		sr.sessionState = SessionStateStopped
	}()

	if !sr.isInitialized {
		return errors.New("configuration not initialized. Call Init() before performing any actions")
	}

	if sr.sessionState == SessionStateStopped || sr.shortSessionID == "" {
		return errors.New("session should be active or paused")
	}

	if sr.sessionType == types.SESSION_TYPE_CONTINUOUS {
		return sr.apiService.StopContinuousSession(sr.shortSessionID)
	} else if sr.sessionType == types.SESSION_TYPE_MANUAL {
		return sr.apiService.CancelSession(sr.shortSessionID)
	}

	return nil
}

func (sr *SessionRecorder) CheckRemoteContinuousSession(sessionPayload *Session) error {
	if !sr.isInitialized {
		return errors.New("configuration not initialized. Call Init() before performing any actions")
	}

	if sessionPayload == nil {
		sessionPayload = &Session{}
	}

	if sessionPayload.ResourceAttributes == nil {
		sessionPayload.ResourceAttributes = make(map[string]interface{})
	}
	for k, v := range sr.resourceAttributes {
		sessionPayload.ResourceAttributes[k] = v
	}

	response, err := sr.apiService.CheckRemoteSession(*sessionPayload)
	if err != nil {
		return err
	}

	if response.State == "START" && sr.sessionState != SessionStateStarted {
		return sr.Start(types.SESSION_TYPE_CONTINUOUS, sessionPayload)
	} else if response.State == "STOP" && sr.sessionState != SessionStateStopped {
		return sr.Stop(nil)
	}

	return nil
}

func defaultSessionShortIDGenerator() string {
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, constants.MULTIPLAYER_TRACE_DEBUG_SESSION_SHORT_ID_LENGTH)
	for i := range result {
		result[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(result)
}

func getFormattedDate(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func getSessionRecorderVersion() string {
	return "1.0.0"
}
