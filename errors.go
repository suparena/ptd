package ptd

import "errors"

// Common PTD errors
var (
	// Envelope errors
	ErrInvalidID     = errors.New("ptd: invalid or missing ID")
	ErrInvalidType   = errors.New("ptd: invalid or missing entity type")
	ErrMissingSchema = errors.New("ptd: missing schema version")
	ErrInvalidSchema = errors.New("ptd: invalid schema version")

	// Validation errors
	ErrValidation    = errors.New("ptd: validation failed")
	ErrInvalidFormat = errors.New("ptd: invalid format")
	ErrMissingField  = errors.New("ptd: required field missing")

	// Signature errors
	ErrSignatureFailed     = errors.New("ptd: signature verification failed")
	ErrSignatureInvalid    = errors.New("ptd: invalid signature")
	ErrSignatureMissing    = errors.New("ptd: signature required but missing")
	ErrSignatureKeyMissing = errors.New("ptd: signing key not found")

	// Package errors
	ErrInvalidPackage  = errors.New("ptd: invalid package format")
	ErrManifestMissing = errors.New("ptd: manifest.json not found")
	ErrManifestInvalid = errors.New("ptd: invalid manifest")
	ErrHashMismatch    = errors.New("ptd: file hash mismatch")

	// Import/Export errors
	ErrImportFailed       = errors.New("ptd: import failed")
	ErrExportFailed       = errors.New("ptd: export failed")
	ErrUnsupportedVersion = errors.New("ptd: unsupported PTD version")
	ErrDuplicateEntity    = errors.New("ptd: duplicate entity detected")
)
