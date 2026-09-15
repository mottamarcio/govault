package service

import "errors"

var (
	// ErrVaultLocked indicates an operation required an unlocked vault session.
	ErrVaultLocked = errors.New("service: vault is locked")

	// ErrAlreadyUnlocked indicates the vault is already unlocked.
	ErrAlreadyUnlocked = errors.New("service: vault is already unlocked")

	// ErrInvalidPassword indicates master password authentication failed.
	ErrInvalidPassword = errors.New("service: invalid master password")

	// ErrVaultNotInitialized indicates the vault database does not contain initialized metadata.
	ErrVaultNotInitialized = errors.New("service: vault is not initialized")

	// ErrVaultAlreadyExists indicates the vault has already been initialized.
	ErrVaultAlreadyExists = errors.New("service: vault is already initialized")

	// ErrEmptyPassword indicates a blank master password was supplied.
	ErrEmptyPassword = errors.New("service: master password cannot be empty")
)
