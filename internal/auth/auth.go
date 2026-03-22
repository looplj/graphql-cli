package auth

import (
	"encoding/base64"
	"fmt"
)

// Credential represents stored auth credentials for an endpoint.
type Credential struct {
	Type  string `json:"type"` // "token", "basic", "header"
	Token string `json:"token,omitempty"`
	User  string `json:"user,omitempty"`
	Pass  string `json:"pass,omitempty"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// AuthHeaders returns HTTP headers for this credential.
func (c *Credential) AuthHeaders() map[string]string {
	switch c.Type {
	case "token":
		return map[string]string{"Authorization": "Bearer " + c.Token}
	case "basic":
		encoded := base64.StdEncoding.EncodeToString([]byte(c.User + ":" + c.Pass))
		return map[string]string{"Authorization": "Basic " + encoded}
	case "header":
		return map[string]string{c.Key: c.Value}
	default:
		return nil
	}
}

// String returns a display-safe summary (no secrets).
func (c *Credential) String() string {
	switch c.Type {
	case "token":
		return fmt.Sprintf("token (%s****)", maskToken(c.Token))
	case "basic":
		return fmt.Sprintf("basic (user: %s)", c.User)
	case "header":
		return fmt.Sprintf("header (%s)", c.Key)
	default:
		return "unknown"
	}
}

func maskToken(t string) string {
	if len(t) <= 4 {
		return "****"
	}

	return t[:4]
}
