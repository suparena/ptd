package ptd

import (
	"os"
	"testing"
	"time"
)

func TestPackageSignature(t *testing.T) {
	// Create a signer
	signer, err := NewSigner("federation-2025", "ITTF Federation")
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Create a package
	pkg := NewPackage("Official Tournament Package")
	defer pkg.Cleanup()

	// Add some entities
	tournament := map[string]interface{}{
		"name":       "World Championship 2025",
		"start_date": time.Now().Format(time.RFC3339),
		"status":     "published",
	}

	envelope := Envelope[map[string]interface{}]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: tournament,
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "ittf:official",
		},
	}

	entities := []interface{}{envelope}
	if err := pkg.AddEntities("tournament", entities); err != nil {
		t.Fatalf("Failed to add entities: %v", err)
	}

	// Sign the package
	if err := pkg.SignPackage(signer); err != nil {
		t.Fatalf("Failed to sign package: %v", err)
	}

	// Verify signature was added
	if pkg.Manifest.Signature == nil {
		t.Fatal("Signature not attached to manifest")
	}

	if pkg.Manifest.Signature.Algorithm != "ed25519" {
		t.Errorf("Wrong algorithm: got %s, want ed25519", pkg.Manifest.Signature.Algorithm)
	}

	if pkg.Manifest.Signature.PublicKeyID != "federation-2025" {
		t.Errorf("Wrong public key ID: got %s, want federation-2025", pkg.Manifest.Signature.PublicKeyID)
	}

	// Verify signature
	if err := pkg.VerifyPackageSignature(signer.publicKey); err != nil {
		t.Errorf("Signature verification failed: %v", err)
	}

	t.Logf("✓ Package signed and verified successfully")
	t.Logf("  Algorithm: %s", pkg.Manifest.Signature.Algorithm)
	t.Logf("  Public Key ID: %s", pkg.Manifest.Signature.PublicKeyID)
	t.Logf("  Signed By: %s", pkg.Manifest.Signature.SignedBy)
	t.Logf("  Signed At: %s", pkg.Manifest.Signature.SignedAt.Format(time.RFC3339))
}

func TestPackageSignatureWithArchive(t *testing.T) {
	// Create a signer
	signer, err := NewSigner("test-key-2025", "test-system")
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}

	// Create a package
	pkg := NewPackage("Signed Tournament Package")
	defer pkg.Cleanup()

	// Add entities
	tournament := map[string]interface{}{
		"name":   "Test Tournament",
		"status": "published",
	}

	envelope := Envelope[map[string]interface{}]{
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

	pkg.AddEntities("tournament", []interface{}{envelope})

	// Sign package BEFORE creating archive
	if err := pkg.SignPackage(signer); err != nil {
		t.Fatalf("Failed to sign package: %v", err)
	}

	// Create archive
	outputPath := "/tmp/signed-test.ptd"
	if err := pkg.CreateArchive(outputPath); err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}
	defer os.Remove(outputPath)

	// Open the archived package
	openedPkg, err := OpenPackage(outputPath)
	if err != nil {
		t.Fatalf("Failed to open package: %v", err)
	}

	// Verify signature persisted in archive
	if openedPkg.Manifest.Signature == nil {
		t.Fatal("Signature not found in archived package")
	}

	// Verify signature
	if err := openedPkg.VerifyPackageSignature(signer.publicKey); err != nil {
		t.Errorf("Signature verification failed for archived package: %v", err)
	}

	t.Logf("✓ Signed package archived and verified successfully")
}

func TestPackageSignatureVerificationFailure(t *testing.T) {
	// Create two signers
	signer1, _ := NewSigner("key-1", "system-1")
	signer2, _ := NewSigner("key-2", "system-2")

	// Create and sign package with signer1
	pkg := NewPackage("Test Package")
	defer pkg.Cleanup()

	pkg.AddEntities("tournament", []interface{}{
		Envelope[map[string]interface{}]{
			ID:   GenerateID(TypeTournament),
			Type: TypeTournament,
			Spec: map[string]interface{}{"name": "Test"},
			Meta: Meta{
				Schema:    "ptd.v1.tournament@1.0.0",
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Source:    "test",
			},
		},
	})

	pkg.SignPackage(signer1)

	// Try to verify with signer2's key
	err := pkg.VerifyPackageSignature(signer2.publicKey)
	if err == nil {
		t.Error("Verification should fail with wrong key")
	}

	if err != ErrSignatureFailed {
		t.Errorf("Expected ErrSignatureFailed, got %v", err)
	}

	t.Logf("✓ Signature verification correctly failed with wrong key")
}

func TestPackageWithoutSignature(t *testing.T) {
	// Create package without signing
	pkg := NewPackage("Unsigned Package")
	defer pkg.Cleanup()

	pkg.AddEntities("tournament", []interface{}{
		Envelope[map[string]interface{}]{
			ID:   GenerateID(TypeTournament),
			Type: TypeTournament,
			Spec: map[string]interface{}{"name": "Test"},
			Meta: Meta{
				Schema:    "ptd.v1.tournament@1.0.0",
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Source:    "test",
			},
		},
	})

	signer, _ := NewSigner("test-key", "test")

	// Try to verify unsigned package
	err := pkg.VerifyPackageSignature(signer.publicKey)
	if err != ErrSignatureMissing {
		t.Errorf("Expected ErrSignatureMissing, got %v", err)
	}

	t.Logf("✓ Correctly detected missing signature")
}
