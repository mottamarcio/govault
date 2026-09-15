package domain_test

import (
	"reflect"
	"testing"

	"github.com/mottamarcio/govault/internal/domain"
)

func TestPayloadSerialization(t *testing.T) {
	tests := []struct {
		name       string
		recordType domain.RecordType
		payload    any
	}{
		{
			name:       "LoginPayload round-trip with Unicode and custom fields",
			recordType: domain.RecordTypeLogin,
			payload: domain.LoginPayload{
				Username: "usuário@domínio.com",
				Password: "s3cr3t!P@ssw0rd🔐",
				URI:      "https://example.com/login",
				Notes:    "Linha 1\nLinha 2 com emojis 🚀",
				CustomFields: []domain.Field{
					{Key: "PIN", Value: "1234", Masked: true},
					{Key: "Security Question", Value: "Resposta secreta", Masked: false},
				},
			},
		},
		{
			name:       "NotePayload round-trip with multiline content",
			recordType: domain.RecordTypeNote,
			payload: domain.NotePayload{
				Content: "# Recovery Keys\n\n- key1: abcd\n- key2: efgh\n\n🎉 Done!",
				CustomFields: []domain.Field{
					{Key: "Category", Value: "Infrastructure", Masked: false},
				},
			},
		},
		{
			name:       "APIKeyPayload round-trip",
			recordType: domain.RecordTypeAPIKey,
			payload: domain.APIKeyPayload{
				Service:  "AWS Production",
				Key:      "AKIAIOSFODNN7EXAMPLE",
				Secret:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
				Endpoint: "https://iam.amazonaws.com",
				Notes:    "Master deployment key",
				CustomFields: []domain.Field{
					{Key: "Region", Value: "us-east-1", Masked: false},
				},
			},
		},
		{
			name:       "CustomPayload round-trip",
			recordType: domain.RecordTypeCustom,
			payload: domain.CustomPayload{
				Notes: "Server credentials",
				Fields: []domain.Field{
					{Key: "Host", Value: "192.168.1.100", Masked: false},
					{Key: "SSH Key", Value: "-----BEGIN RSA PRIVATE KEY-----...", Masked: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := domain.SerializePayload(tt.payload)
			if err != nil {
				t.Fatalf("SerializePayload() error = %v", err)
			}
			if len(data) == 0 {
				t.Fatal("SerializePayload() returned empty bytes")
			}

			deserialized, err := domain.DeserializePayload(tt.recordType, data)
			if err != nil {
				t.Fatalf("DeserializePayload() error = %v", err)
			}

			if !reflect.DeepEqual(tt.payload, deserialized) {
				t.Errorf("Payload mismatch after round-trip:\ngot  %+v\nwant %+v", deserialized, tt.payload)
			}
		})
	}
}

func TestSerializeNilPayload(t *testing.T) {
	_, err := domain.SerializePayload(nil)
	if err != domain.ErrNilPayload {
		t.Fatalf("expected ErrNilPayload, got %v", err)
	}
}

func TestDeserializeEmptyData(t *testing.T) {
	_, err := domain.DeserializePayload(domain.RecordTypeLogin, []byte{})
	if err != domain.ErrNilPayload {
		t.Fatalf("expected ErrNilPayload, got %v", err)
	}
}

func TestDeserializeInvalidType(t *testing.T) {
	_, err := domain.DeserializePayload(domain.RecordType("invalid_type"), []byte(`{"key":"val"}`))
	if err != domain.ErrInvalidRecordType {
		t.Fatalf("expected ErrInvalidRecordType, got %v", err)
	}
}
