package domain

// RecordType represents the classification of a secret record.
type RecordType string

const (
	RecordTypeLogin  RecordType = "login"
	RecordTypeNote   RecordType = "note"
	RecordTypeAPIKey RecordType = "api_key"
	RecordTypeCustom RecordType = "custom"
)

// IsValid checks whether the RecordType is recognized.
func (rt RecordType) IsValid() bool {
	switch rt {
	case RecordTypeLogin, RecordTypeNote, RecordTypeAPIKey, RecordTypeCustom:
		return true
	default:
		return false
	}
}

// String returns the string representation.
func (rt RecordType) String() string {
	return string(rt)
}
