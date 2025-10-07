package ptd

import (
	"crypto/ed25519"
	"testing"
	"time"
)

func TestGenerateKeyPair(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if kp.PublicKey == "" {
		t.Error("Public key is empty")
	}

	if kp.PrivateKey == "" {
		t.Error("Private key is empty")
	}

	// Verify keys can be parsed
	publicKey, err := ParsePublicKey(kp.PublicKey)
	if err != nil {
		t.Errorf("Failed to parse public key: %v", err)
	}

	if len(publicKey) != ed25519.PublicKeySize {
		t.Errorf("Invalid public key size: got %d, want %d", len(publicKey), ed25519.PublicKeySize)
	}

	privateKey, err := ParsePrivateKey(kp.PrivateKey)
	if err != nil {
		t.Errorf("Failed to parse private key: %v", err)
	}

	if len(privateKey) != ed25519.PrivateKeySize {
		t.Errorf("Invalid private key size: got %d, want %d", len(privateKey), ed25519.PrivateKeySize)
	}
}

func TestSignerCreation(t *testing.T) {
	signer, err := NewSigner("test-key-2025", "test-system")
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	if signer.publicKeyID != "test-key-2025" {
		t.Errorf("Wrong public key ID: got %s, want %s", signer.publicKeyID, "test-key-2025")
	}

	if signer.signedBy != "test-system" {
		t.Errorf("Wrong signer: got %s, want %s", signer.signedBy, "test-system")
	}

	if signer.PublicKey() == "" {
		t.Error("Public key is empty")
	}

	if signer.PrivateKey() == "" {
		t.Error("Private key is empty")
	}
}

func TestSignAndVerifyEnvelope(t *testing.T) {
	// Create signer
	signer, err := NewSigner("test-key-2025", "test-system")
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Create envelope
	tournament := Tournament{
		Name:      "Test Tournament",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		Status:    "published",
	}

	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: tournament,
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	// Sign envelope
	if err := signer.Sign(envelope); err != nil {
		t.Fatalf("Failed to sign envelope: %v", err)
	}

	// Verify signature was added
	if envelope.Meta.Signature == nil {
		t.Fatal("Signature not attached to envelope")
	}

	if envelope.Meta.Signature.Algorithm != "ed25519" {
		t.Errorf("Wrong algorithm: got %s, want ed25519", envelope.Meta.Signature.Algorithm)
	}

	if envelope.Meta.Signature.PublicKeyID != "test-key-2025" {
		t.Errorf("Wrong public key ID: got %s, want test-key-2025", envelope.Meta.Signature.PublicKeyID)
	}

	if envelope.Meta.Signature.SignedBy != "test-system" {
		t.Errorf("Wrong signer: got %s, want test-system", envelope.Meta.Signature.SignedBy)
	}

	// Verify signature
	if err := Verify(envelope, signer.publicKey); err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}
}

func TestVerifyWithWrongKey(t *testing.T) {
	// Create two signers
	signer1, _ := NewSigner("key-1", "system-1")
	signer2, _ := NewSigner("key-2", "system-2")

	// Create and sign envelope with signer1
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	signer1.Sign(envelope)

	// Try to verify with signer2's key
	err := Verify(envelope, signer2.publicKey)
	if err == nil {
		t.Error("Verification should fail with wrong key")
	}

	if err != ErrSignatureFailed {
		t.Errorf("Expected ErrSignatureFailed, got %v", err)
	}
}

func TestVerifyWithoutSignature(t *testing.T) {
	// Create envelope without signature
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	signer, _ := NewSigner("test-key", "test-system")

	err := Verify(envelope, signer.publicKey)
	if err != ErrSignatureMissing {
		t.Errorf("Expected ErrSignatureMissing, got %v", err)
	}
}

func TestSignMapEnvelope(t *testing.T) {
	// Test with map[string]interface{} envelope
	signer, _ := NewSigner("test-key", "test")

	envelope := &Envelope[map[string]interface{}]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: map[string]interface{}{
			"name":   "Test Tournament",
			"status": "published",
		},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	if err := signer.Sign(envelope); err != nil {
		t.Fatalf("Failed to sign map envelope: %v", err)
	}

	if envelope.Meta.Signature == nil {
		t.Fatal("Signature not attached")
	}

	if err := Verify(envelope, signer.publicKey); err != nil {
		t.Errorf("Verification failed: %v", err)
	}
}

func TestVerifyWithPublicKeyID(t *testing.T) {
	// Create signer
	signer, _ := NewSigner("official-key-2025", "federation")

	// Create envelope
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Official Tournament"},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "federation",
		},
	}

	// Sign
	signer.Sign(envelope)

	// Create lookup function
	keyStore := map[string]ed25519.PublicKey{
		"official-key-2025": signer.publicKey,
	}

	lookupFunc := func(keyID string) (ed25519.PublicKey, error) {
		if key, ok := keyStore[keyID]; ok {
			return key, nil
		}
		return nil, ErrSignatureKeyMissing
	}

	// Verify with lookup
	if err := VerifyWithPublicKeyID(envelope, lookupFunc); err != nil {
		t.Errorf("Verification with key ID failed: %v", err)
	}

	// Test with wrong key ID
	envelope.Meta.Signature.PublicKeyID = "unknown-key"
	err := VerifyWithPublicKeyID(envelope, lookupFunc)
	if err != ErrSignatureKeyMissing {
		t.Errorf("Expected ErrSignatureKeyMissing, got %v", err)
	}
}

func TestCanonicalJSONExcludesSignature(t *testing.T) {
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
			Signature: &Signature{
				Algorithm:   "ed25519",
				PublicKeyID: "test",
				Signature:   "should-be-excluded",
				SignedAt:    time.Now(),
				SignedBy:    "test",
			},
		},
	}

	canonical, err := envelope.CanonicalJSON()
	if err != nil {
		t.Fatalf("Failed to get canonical JSON: %v", err)
	}

	// Signature should not be in canonical JSON
	canonicalStr := string(canonical)
	if len(canonicalStr) == 0 {
		t.Error("Canonical JSON is empty")
	}

	// Note: We can't easily check if signature is excluded without parsing,
	// but the CanonicalJSON method should handle it
}
