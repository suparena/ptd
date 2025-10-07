package main

import (
	"fmt"
	"log"
	"time"

	"github.com/suparena/ptd"
)

func main() {
	// Example 1: Generate a key pair
	fmt.Println("=== Example 1: Generate Key Pair ===")
	keyPair, err := ptd.GenerateKeyPair()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Public Key:  %s...\n", keyPair.PublicKey[:32])
	fmt.Printf("Private Key: %s...\n", keyPair.PrivateKey[:32])
	fmt.Println()

	// Example 2: Sign an envelope
	fmt.Println("=== Example 2: Sign an Envelope ===")

	// Create a signer
	signer, err := ptd.NewSigner("ittf-official-2025", "ITTF Federation")
	if err != nil {
		log.Fatal(err)
	}

	// Create a tournament envelope
	tournament := ptd.Tournament{
		Name:      "World Championship 2025",
		StartDate: time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 7, 10, 18, 0, 0, 0, time.UTC),
		Status:    "published",
		Venue: &ptd.Venue{
			Name:    "Tokyo Metropolitan Gymnasium",
			City:    "Tokyo",
			Country: "Japan",
		},
	}

	envelope := &ptd.Envelope[ptd.Tournament]{
		ID:   ptd.GenerateID(ptd.TypeTournament),
		Type: ptd.TypeTournament,
		Spec: tournament,
		Meta: ptd.Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "ittf:official",
		},
	}

	// Sign the envelope
	if err := signer.Sign(envelope); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Envelope signed successfully\n")
	fmt.Printf("  Algorithm:     %s\n", envelope.Meta.Signature.Algorithm)
	fmt.Printf("  Public Key ID: %s\n", envelope.Meta.Signature.PublicKeyID)
	fmt.Printf("  Signed By:     %s\n", envelope.Meta.Signature.SignedBy)
	fmt.Printf("  Signed At:     %s\n", envelope.Meta.Signature.SignedAt.Format(time.RFC3339))
	fmt.Println()

	// Example 3: Verify envelope signature
	fmt.Println("=== Example 3: Verify Envelope Signature ===")

	// Parse public key from base64
	publicKey, err := ptd.ParsePublicKey(signer.PublicKey())
	if err != nil {
		log.Fatal(err)
	}

	if err := ptd.Verify(envelope, publicKey); err != nil {
		log.Fatal(err)
	}
	fmt.Println("✓ Envelope signature verified successfully")
	fmt.Println()

	// Example 4: Sign a package
	fmt.Println("=== Example 4: Sign a Package ===")

	// Create package
	pkg := ptd.NewPackage("Official ITTF Tournament Package")
	defer pkg.Cleanup()

	// Add tournament
	pkg.AddEntities("tournament", []interface{}{envelope})

	// Sign the package
	if err := pkg.SignPackage(signer); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Package signed successfully\n")
	fmt.Printf("  Public Key ID: %s\n", pkg.Manifest.Signature.PublicKeyID)
	fmt.Printf("  Signed By:     %s\n", pkg.Manifest.Signature.SignedBy)
	fmt.Println()

	// Create archive
	outputPath := "/tmp/signed-tournament.ptd"
	if err := pkg.CreateArchive(outputPath); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created signed archive: %s\n", outputPath)
	fmt.Println()

	// Example 5: Verify package signature
	fmt.Println("=== Example 5: Verify Package Signature ===")

	// Open the package
	openedPkg, err := ptd.OpenPackage(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	// Verify signature (parse public key again for this example)
	publicKey2, err := ptd.ParsePublicKey(signer.PublicKey())
	if err != nil {
		log.Fatal(err)
	}

	if err := openedPkg.VerifyPackageSignature(publicKey2); err != nil {
		log.Fatal(err)
	}

	fmt.Println("✓ Package signature verified successfully")
	fmt.Printf("  Package contains %d entity types\n", len(openedPkg.Manifest.Entities))
	for entityType, count := range openedPkg.Manifest.Entities {
		fmt.Printf("    - %s: %d\n", entityType, count.Count)
	}
	fmt.Println()

	// Example 6: Key lookup function
	fmt.Println("=== Example 6: Verify with Key Lookup ===")

	// Simulate a key store
	keyStore := map[string]string{
		"ittf-official-2025": signer.PublicKey(),
	}

	lookupFunc := func(keyID string) (string, error) {
		if key, ok := keyStore[keyID]; ok {
			return key, nil
		}
		return "", ptd.ErrSignatureKeyMissing
	}

	// Verify with lookup
	keyID := envelope.Meta.Signature.PublicKeyID
	publicKeyB64, err := lookupFunc(keyID)
	if err != nil {
		log.Fatal(err)
	}

	publicKey3, err := ptd.ParsePublicKey(publicKeyB64)
	if err != nil {
		log.Fatal(err)
	}

	if err := ptd.Verify(envelope, publicKey3); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("✓ Verified using key lookup for: %s\n", keyID)
	fmt.Println()

	fmt.Println("=== All Examples Completed Successfully ===")
}
