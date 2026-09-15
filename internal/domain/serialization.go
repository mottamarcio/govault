package domain

import (
	"encoding/json"
	"fmt"
)

// SerializePayload converts a typed payload struct into canonical JSON bytes.
func SerializePayload(payload any) ([]byte, error) {
	if payload == nil {
		return nil, ErrNilPayload
	}

	switch p := payload.(type) {
	case []byte:
		return p, nil
	case LoginPayload, *LoginPayload, NotePayload, *NotePayload, APIKeyPayload, *APIKeyPayload, CustomPayload, *CustomPayload:
		return json.Marshal(p)
	default:
		return nil, fmt.Errorf("domain: unsupported payload type %T", payload)
	}
}

// DeserializePayload converts raw bytes into the corresponding typed payload struct for the given RecordType.
func DeserializePayload(recordType RecordType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, ErrNilPayload
	}

	switch recordType {
	case RecordTypeLogin:
		var p LoginPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("domain: failed to unmarshal login payload: %w", err)
		}
		return p, nil

	case RecordTypeNote:
		var p NotePayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("domain: failed to unmarshal note payload: %w", err)
		}
		return p, nil

	case RecordTypeAPIKey:
		var p APIKeyPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("domain: failed to unmarshal api key payload: %w", err)
		}
		return p, nil

	case RecordTypeCustom:
		var p CustomPayload
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("domain: failed to unmarshal custom payload: %w", err)
		}
		return p, nil

	default:
		return nil, ErrInvalidRecordType
	}
}
