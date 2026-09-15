package domain

import "strings"

// LoginPayload holds secret data for a user login credential.
type LoginPayload struct {
	Username     string  `json:"username"`
	Password     string  `json:"password"`
	URI          string  `json:"uri,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

// NotePayload holds secret data for a secure note.
type NotePayload struct {
	Content      string  `json:"content"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

// APIKeyPayload holds secret data for an API key or developer secret.
type APIKeyPayload struct {
	Service      string  `json:"service"`
	Key          string  `json:"key"`
	Secret       string  `json:"secret"`
	Endpoint     string  `json:"endpoint,omitempty"`
	Notes        string  `json:"notes,omitempty"`
	CustomFields []Field `json:"custom_fields,omitempty"`
}

// CustomPayload holds arbitrary secret key-value pairs and notes.
type CustomPayload struct {
	Notes  string  `json:"notes,omitempty"`
	Fields []Field `json:"fields,omitempty"`
}

// validateUniqueFieldKeys helper to check for duplicate keys in field slices.
func validateUniqueFieldKeys(fields []Field) error {
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		if err := f.Validate(); err != nil {
			return err
		}
		keyLower := strings.ToLower(strings.TrimSpace(f.Key))
		if _, exists := seen[keyLower]; exists {
			return ErrDuplicateFieldKey
		}
		seen[keyLower] = struct{}{}
	}
	return nil
}
