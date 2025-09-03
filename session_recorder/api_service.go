package session_recorder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/multiplayer-app/multiplayer-otlp-go/constants"
)

// APIServiceConfig holds the configuration for the API service
type APIServiceConfig struct {
	APIKey              string
	APIBaseURL          string
	ContinuousRecording bool
}

type Tag struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value"`
}

type StartSessionRequest struct {
	Name               string                 `json:"name,omitempty"`
	ResourceAttributes map[string]interface{} `json:"resourceAttributes,omitempty"`
	SessionAttributes  map[string]interface{} `json:"sessionAttributes,omitempty"`
	Tags               []Tag                  `json:"tags,omitempty"`
}

type StopSessionRequest struct {
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
}

type CheckRemoteSessionResponse struct {
	State string `json:"state"` // "START" or "STOP"
}

type APIService struct {
	config APIServiceConfig
	client *http.Client
}

func NewAPIService() *APIService {
	return &APIService{
		config: APIServiceConfig{
			APIBaseURL: constants.MULTIPLAYER_BASE_API_URL,
		},
		client: &http.Client{},
	}
}

func (a *APIService) Init(config APIServiceConfig) {
	if config.APIBaseURL == "" {
		config.APIBaseURL = constants.MULTIPLAYER_BASE_API_URL
	}
	
	a.config = config
}

func (a *APIService) UpdateConfigs(config APIServiceConfig) {
	if config.APIBaseURL == "" {
		config.APIBaseURL = constants.MULTIPLAYER_BASE_API_URL
	}
	
	if config.APIKey != "" {
		a.config.APIKey = config.APIKey
	}
	if config.APIBaseURL != "" {
		a.config.APIBaseURL = config.APIBaseURL
	}
	a.config.ContinuousRecording = config.ContinuousRecording
}

func (a *APIService) GetAPIBaseURL() string {
	if a.config.APIBaseURL != "" {
		return a.config.APIBaseURL
	}
	return constants.MULTIPLAYER_BASE_API_URL
}

func (a *APIService) StartSession(requestBody Session) (*Session, error) {
	req := StartSessionRequest{
		Name:               requestBody.Name,
		ResourceAttributes: requestBody.ResourceAttributes,
		SessionAttributes:  requestBody.SessionAttributes,
		Tags:               convertToTags(requestBody.Tags),
	}
	
	var response Session
	err := a.makeRequest("/debug-sessions/start", "POST", req, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

func (a *APIService) StopSession(sessionID string, requestBody Session) error {
	req := StopSessionRequest{
		SessionAttributes: requestBody.SessionAttributes,
	}
	
	return a.makeRequest(fmt.Sprintf("/debug-sessions/%s/stop", sessionID), "PATCH", req, nil)
}

func (a *APIService) CancelSession(sessionID string) error {
	return a.makeRequest(fmt.Sprintf("/debug-sessions/%s/cancel", sessionID), "DELETE", nil, nil)
}

func (a *APIService) StartContinuousSession(requestBody Session) (*Session, error) {
	req := StartSessionRequest{
		Name:               requestBody.Name,
		ResourceAttributes: requestBody.ResourceAttributes,
		SessionAttributes:  requestBody.SessionAttributes,
		Tags:               convertToTags(requestBody.Tags),
	}
	
	var response Session
	err := a.makeRequest("/continuous-debug-sessions/start", "POST", req, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

func (a *APIService) SaveContinuousSession(sessionID string, requestBody Session) error {
	req := StartSessionRequest{
		Name:               requestBody.Name,
		ResourceAttributes: requestBody.ResourceAttributes,
		SessionAttributes:  requestBody.SessionAttributes,
		Tags:               convertToTags(requestBody.Tags),
	}
	
	return a.makeRequest(fmt.Sprintf("/continuous-debug-sessions/%s/save", sessionID), "POST", req, nil)
}

func (a *APIService) StopContinuousSession(sessionID string) error {
	return a.makeRequest(fmt.Sprintf("/continuous-debug-sessions/%s/cancel", sessionID), "DELETE", nil, nil)
}

func (a *APIService) CheckRemoteSession(requestBody Session) (*CheckRemoteSessionResponse, error) {
	req := StartSessionRequest{
		Name:               requestBody.Name,
		ResourceAttributes: requestBody.ResourceAttributes,
		SessionAttributes:  requestBody.SessionAttributes,
		Tags:               convertToTags(requestBody.Tags),
	}
	
	var response CheckRemoteSessionResponse
	err := a.makeRequest("/remote-debug-session/check", "POST", req, &response)
	if err != nil {
		return nil, err
	}
	
	return &response, nil
}

func (a *APIService) makeRequest(path, method string, body interface{}, response interface{}) error {
	url := fmt.Sprintf("%s/v0/radar%s", a.GetAPIBaseURL(), path)
	
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}
	
	req, err := http.NewRequestWithContext(context.Background(), method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	if a.config.APIKey != "" {
		req.Header.Set("X-Api-Key", a.config.APIKey)
	}
	
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("network response was not ok: %s, body: %s", resp.Status, string(bodyBytes))
	}
	
	if resp.StatusCode == 204 {
		return nil
	}
	
	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	
	return nil
}

func convertToTags(tags map[string]string) []Tag {
	if tags == nil {
		return nil
	}
	
	result := make([]Tag, 0, len(tags))
	for key, value := range tags {
		result = append(result, Tag{
			Key:   key,
			Value: value,
		})
	}
	
	return result
}
