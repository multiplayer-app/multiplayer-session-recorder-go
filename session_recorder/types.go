package session_recorder

// Session represents a debug session
type Session struct {
	ID                 string                 `json:"_id,omitempty"`
	ShortID            string                 `json:"shortId,omitempty"`
	Name               string                 `json:"name,omitempty"`
	ResourceAttributes map[string]interface{} `json:"resourceAttributes,omitempty"`
	SessionAttributes  map[string]interface{} `json:"sessionAttributes,omitempty"`
	Tags               map[string]string      `json:"tags,omitempty"`
}
