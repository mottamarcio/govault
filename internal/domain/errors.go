package domain

import "errors"

var (
	// ErrInvalidRecordType indicates an unrecognized or empty record type.
	ErrInvalidRecordType = errors.New("domain: invalid record type")

	// ErrEmptyTitle indicates a missing or whitespace-only record title.
	ErrEmptyTitle = errors.New("domain: record title cannot be empty")

	// ErrTitleTooLong indicates the record title exceeds max length.
	ErrTitleTooLong = errors.New("domain: record title exceeds maximum length (255 characters)")

	// ErrInvalidUUID indicates an invalid or malformed UUID identifier.
	ErrInvalidUUID = errors.New("domain: invalid UUID identifier")

	// ErrEmptyKey indicates an empty field key.
	ErrEmptyKey = errors.New("domain: field key cannot be empty")

	// ErrDuplicateFieldKey indicates duplicate custom field keys in a single payload.
	ErrDuplicateFieldKey = errors.New("domain: duplicate custom field key")

	// ErrEmptyTagName indicates a missing or whitespace-only tag name.
	ErrEmptyTagName = errors.New("domain: tag name cannot be empty")

	// ErrTagNameTooLong indicates a tag name exceeds max length.
	ErrTagNameTooLong = errors.New("domain: tag name exceeds maximum length (64 characters)")

	// ErrNilPayload indicates a nil payload was provided.
	ErrNilPayload = errors.New("domain: payload cannot be nil")
)
