package ptd

import (
	"encoding/json"
	"testing"
	"time"
)

func TestEnvelope_Validate(t *testing.T) {
	tests := []struct {
		name     string
		envelope Envelope[Tournament]
		wantErr  error
	}{
		{
			name: "valid envelope",
			envelope: Envelope[Tournament]{
				ID:   "ptd:tournament:01ABC123",
				Type: TypeTournament,
				Spec: Tournament{Name: "Test Tournament"},
				Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
			},
			wantErr: nil,
		},
		{
			name: "missing ID",
			envelope: Envelope[Tournament]{
				Type: TypeTournament,
				Spec: Tournament{Name: "Test Tournament"},
				Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
			},
			wantErr: ErrInvalidID,
		},
		{
			name: "missing type",
			envelope: Envelope[Tournament]{
				ID:   "ptd:tournament:01ABC123",
				Spec: Tournament{Name: "Test Tournament"},
				Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
			},
			wantErr: ErrInvalidType,
		},
		{
			name: "missing schema",
			envelope: Envelope[Tournament]{
				ID:   "ptd:tournament:01ABC123",
				Type: TypeTournament,
				Spec: Tournament{Name: "Test Tournament"},
				Meta: Meta{},
			},
			wantErr: ErrMissingSchema,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.envelope.Validate()
			if err != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEnvelope_CanonicalJSON(t *testing.T) {
	now := time.Now()
	tournament := Tournament{
		Name:      "Test Tournament",
		StartDate: now,
		EndDate:   now.Add(2 * 24 * time.Hour),
		Status:    "draft",
	}

	envelope := Envelope[Tournament]{
		ID:   "ptd:tournament:test123",
		Type: TypeTournament,
		Spec: tournament,
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: now,
			UpdatedAt: now,
			Source:    "test",
			Signature: &Signature{
				Algorithm:   "ed25519",
				PublicKeyID: "test-key",
				Signature:   "test-signature",
				SignedAt:    now,
				SignedBy:    "test-signer",
			},
		},
	}

	// Get canonical JSON
	canonical, err := envelope.CanonicalJSON()
	if err != nil {
		t.Fatalf("CanonicalJSON() error = %v", err)
	}

	// Unmarshal to check signature was removed
	var result map[string]interface{}
	if err := json.Unmarshal(canonical, &result); err != nil {
		t.Fatalf("Failed to unmarshal canonical JSON: %v", err)
	}

	// Check that meta exists but signature is nil
	meta, ok := result["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("Meta field not found in canonical JSON")
	}

	if _, exists := meta["signature"]; exists {
		if meta["signature"] != nil {
			t.Error("Signature should be nil in canonical JSON")
		}
	}
}

func TestMeta_Extensions(t *testing.T) {
	meta := Meta{
		Schema:    "ptd.v1.tournament@1.0.0",
		Version:   1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Source:    "test",
		Tags:      []string{"official", "rated"},
		Extensions: map[string]interface{}{
			"vendor":       "icc",
			"custom_field": "custom_value",
			"priority":     1,
		},
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal meta: %v", err)
	}

	var decoded Meta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal meta: %v", err)
	}

	// Check extensions
	if decoded.Extensions["vendor"] != "icc" {
		t.Errorf("Expected vendor='icc', got %v", decoded.Extensions["vendor"])
	}

	// Check priority is correctly decoded as float64 (JSON number behavior)
	priority, ok := decoded.Extensions["priority"].(float64)
	if !ok || priority != 1 {
		t.Errorf("Expected priority=1, got %v", decoded.Extensions["priority"])
	}
}

func TestProvenance(t *testing.T) {
	now := time.Now()
	importTime := now.Add(-1 * time.Hour)

	provenance := Provenance{
		OriginalSource: "tournament-software",
		ImportedFrom:   "ptd:package:abc123",
		ImportedAt:     &importTime,
		Transformations: []Transform{
			{
				Type:        "normalize_dates",
				Description: "Converted all dates to UTC",
				AppliedAt:   now,
				AppliedBy:   "ptd-importer",
			},
			{
				Type:        "validate_players",
				Description: "Validated player IDs against registry",
				AppliedAt:   now.Add(1 * time.Minute),
				AppliedBy:   "ptd-validator",
			},
		},
	}

	meta := Meta{
		Schema:     "ptd.v1.tournament@1.0.0",
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
		Source:     "import",
		Provenance: &provenance,
	}

	// Test JSON marshaling/unmarshaling
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal meta with provenance: %v", err)
	}

	var decoded Meta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal meta: %v", err)
	}

	if decoded.Provenance == nil {
		t.Fatal("Provenance should not be nil")
	}

	if decoded.Provenance.OriginalSource != "tournament-software" {
		t.Errorf("Expected original source 'tournament-software', got %s", decoded.Provenance.OriginalSource)
	}

	if len(decoded.Provenance.Transformations) != 2 {
		t.Errorf("Expected 2 transformations, got %d", len(decoded.Provenance.Transformations))
	}
}

func TestSignature(t *testing.T) {
	now := time.Now()
	sig := Signature{
		Algorithm:   "ed25519",
		PublicKeyID: "suparena-master-2025",
		Signature:   "dGVzdCBzaWduYXR1cmU=", // base64 "test signature"
		SignedAt:    now,
		SignedBy:    "suparena-signer",
	}

	data, err := json.Marshal(sig)
	if err != nil {
		t.Fatalf("Failed to marshal signature: %v", err)
	}

	var decoded Signature
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal signature: %v", err)
	}

	if decoded.Algorithm != "ed25519" {
		t.Errorf("Expected algorithm 'ed25519', got %s", decoded.Algorithm)
	}

	if decoded.PublicKeyID != "suparena-master-2025" {
		t.Errorf("Expected public key ID 'suparena-master-2025', got %s", decoded.PublicKeyID)
	}

	if decoded.SignedBy != "suparena-signer" {
		t.Errorf("Expected signed by 'suparena-signer', got %s", decoded.SignedBy)
	}
}
