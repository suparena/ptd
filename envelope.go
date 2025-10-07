// Package ptd provides the Portable Tournament Data format for universal tournament data exchange
package ptd

import (
	"encoding/json"
	"time"
)

// Envelope is the universal wrapper for all PTD entities
type Envelope[T any] struct {
	ID   string `json:"id"`   // Format: ptd:ulid:{ULID} or ptd:{type}:{identifier}
	Type string `json:"type"` // Entity type: tournament, event, match, etc.
	Spec T      `json:"spec"` // The actual entity data
	Meta Meta   `json:"meta"` // Metadata about this entity
}

// Meta contains metadata about the entity
type Meta struct {
	Schema    string    `json:"schema"`     // Schema version (e.g., "ptd.v1.tournament@1.0.0")
	Version   int       `json:"version"`    // Entity version number for optimistic locking
	CreatedAt time.Time `json:"created_at"` // When this entity was created
	UpdatedAt time.Time `json:"updated_at"` // When this entity was last updated
	Source    string    `json:"source"`     // Source system (e.g., "icc:prod-us-west")

	// Optional metadata fields
	Tags       []string               `json:"tags,omitempty"`       // User-defined tags
	Extensions map[string]interface{} `json:"extensions,omitempty"` // Vendor-specific extensions
	Signature  *Signature             `json:"signature,omitempty"`  // Digital signature
	Provenance *Provenance            `json:"provenance,omitempty"` // Data lineage
}

// Signature contains digital signature information
type Signature struct {
	Algorithm   string    `json:"algorithm"`     // Signature algorithm (e.g., "ed25519")
	PublicKeyID string    `json:"public_key_id"` // ID of the signing key
	Signature   string    `json:"signature"`     // Base64-encoded signature
	SignedAt    time.Time `json:"signed_at"`     // When the signature was created
	SignedBy    string    `json:"signed_by"`     // Identity of signer
}

// Provenance tracks the origin and history of the data
type Provenance struct {
	OriginalSource  string      `json:"original_source"`         // Original data source
	ImportedFrom    string      `json:"imported_from,omitempty"` // If imported from another PTD
	ImportedAt      *time.Time  `json:"imported_at,omitempty"`
	Transformations []Transform `json:"transformations,omitempty"` // Data transformations applied
}

// Transform represents a data transformation
type Transform struct {
	Type        string    `json:"type"`        // Type of transformation
	Description string    `json:"description"` // Human-readable description
	AppliedAt   time.Time `json:"applied_at"`  // When transformation was applied
	AppliedBy   string    `json:"applied_by"`  // System or user that applied it
}

// CanonicalJSON returns the canonical JSON representation for signing
func (e *Envelope[T]) CanonicalJSON() ([]byte, error) {
	// Create a copy without signature for canonical representation
	temp := *e
	if temp.Meta.Signature != nil {
		metaCopy := temp.Meta
		metaCopy.Signature = nil
		temp.Meta = metaCopy
	}

	// Use deterministic JSON encoding
	return json.Marshal(temp)
}

// Validate checks if the envelope is valid
func (e *Envelope[T]) Validate() error {
	if e.ID == "" {
		return ErrInvalidID
	}
	if e.Type == "" {
		return ErrInvalidType
	}
	if e.Meta.Schema == "" {
		return ErrMissingSchema
	}
	return nil
}
