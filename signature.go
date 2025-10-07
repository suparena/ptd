package ptd

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// Signer provides digital signature functionality for PTD entities
type Signer struct {
	privateKey  ed25519.PrivateKey
	publicKey   ed25519.PublicKey
	publicKeyID string
	signedBy    string
}

// NewSigner creates a new signer with generated Ed25519 keys
func NewSigner(publicKeyID, signedBy string) (*Signer, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return &Signer{
		privateKey:  privateKey,
		publicKey:   publicKey,
		publicKeyID: publicKeyID,
		signedBy:    signedBy,
	}, nil
}

// NewSignerFromKeys creates a signer from existing Ed25519 keys
func NewSignerFromKeys(privateKey ed25519.PrivateKey, publicKeyID, signedBy string) *Signer {
	return &Signer{
		privateKey:  privateKey,
		publicKey:   privateKey.Public().(ed25519.PublicKey),
		publicKeyID: publicKeyID,
		signedBy:    signedBy,
	}
}

// PublicKey returns the base64-encoded public key
func (s *Signer) PublicKey() string {
	return base64.StdEncoding.EncodeToString(s.publicKey)
}

// PrivateKey returns the base64-encoded private key
func (s *Signer) PrivateKey() string {
	return base64.StdEncoding.EncodeToString(s.privateKey)
}

// Sign signs an envelope and attaches the signature to its metadata
func (s *Signer) Sign(envelope interface{}) error {
	// Type assertion to access CanonicalJSON method
	type canonicalJSONer interface {
		CanonicalJSON() ([]byte, error)
	}

	canonicalizer, ok := envelope.(canonicalJSONer)
	if !ok {
		return fmt.Errorf("envelope does not support canonical JSON")
	}

	// Get canonical JSON (without signature)
	canonical, err := canonicalizer.CanonicalJSON()
	if err != nil {
		return fmt.Errorf("failed to get canonical JSON: %w", err)
	}

	// Sign the canonical JSON
	signature := ed25519.Sign(s.privateKey, canonical)

	// Encode signature as base64
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Create signature object
	sig := &Signature{
		Algorithm:   "ed25519",
		PublicKeyID: s.publicKeyID,
		Signature:   signatureB64,
		SignedAt:    time.Now(),
		SignedBy:    s.signedBy,
	}

	// Attach to envelope (type-specific)
	return attachSignature(envelope, sig)
}

// attachSignature attaches a signature to an envelope's metadata
func attachSignature(envelope interface{}, sig *Signature) error {
	// Use reflection to set Meta.Signature
	switch e := envelope.(type) {
	case *Envelope[map[string]interface{}]:
		e.Meta.Signature = sig
	case *Envelope[Tournament]:
		e.Meta.Signature = sig
	case *Envelope[Event]:
		e.Meta.Signature = sig
	case *Envelope[Match]:
		e.Meta.Signature = sig
	case *Envelope[Entry]:
		e.Meta.Signature = sig
	default:
		// Fallback: try to set via JSON marshaling/unmarshaling
		data, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal envelope: %w", err)
		}

		var temp map[string]interface{}
		if err := json.Unmarshal(data, &temp); err != nil {
			return fmt.Errorf("failed to unmarshal envelope: %w", err)
		}

		if meta, ok := temp["meta"].(map[string]interface{}); ok {
			meta["signature"] = sig

			// Re-marshal and unmarshal back to original type
			data, err := json.Marshal(temp)
			if err != nil {
				return fmt.Errorf("failed to re-marshal: %w", err)
			}

			if err := json.Unmarshal(data, envelope); err != nil {
				return fmt.Errorf("failed to unmarshal back: %w", err)
			}
		}
	}

	return nil
}

// Verify verifies the signature of an envelope
func Verify(envelope interface{}, publicKey ed25519.PublicKey) error {
	// Type assertion to access CanonicalJSON method
	type canonicalJSONer interface {
		CanonicalJSON() ([]byte, error)
	}

	canonicalizer, ok := envelope.(canonicalJSONer)
	if !ok {
		return ErrSignatureFailed
	}

	// Extract signature from envelope
	sig, err := extractSignature(envelope)
	if err != nil {
		return err
	}

	if sig == nil {
		return ErrSignatureMissing
	}

	// Get canonical JSON (without signature)
	canonical, err := canonicalizer.CanonicalJSON()
	if err != nil {
		return fmt.Errorf("failed to get canonical JSON: %w", err)
	}

	// Decode signature
	signatureBytes, err := base64.StdEncoding.DecodeString(sig.Signature)
	if err != nil {
		return ErrSignatureInvalid
	}

	// Verify signature
	if !ed25519.Verify(publicKey, canonical, signatureBytes) {
		return ErrSignatureFailed
	}

	return nil
}

// VerifyWithPublicKeyID verifies using a public key ID lookup function
func VerifyWithPublicKeyID(envelope interface{}, lookupFunc func(string) (ed25519.PublicKey, error)) error {
	// Extract signature
	sig, err := extractSignature(envelope)
	if err != nil {
		return err
	}

	if sig == nil {
		return ErrSignatureMissing
	}

	// Lookup public key
	publicKey, err := lookupFunc(sig.PublicKeyID)
	if err != nil {
		return ErrSignatureKeyMissing
	}

	// Verify
	return Verify(envelope, publicKey)
}

// extractSignature extracts the signature from an envelope
func extractSignature(envelope interface{}) (*Signature, error) {
	switch e := envelope.(type) {
	case *Envelope[map[string]interface{}]:
		return e.Meta.Signature, nil
	case *Envelope[Tournament]:
		return e.Meta.Signature, nil
	case *Envelope[Event]:
		return e.Meta.Signature, nil
	case *Envelope[Match]:
		return e.Meta.Signature, nil
	case *Envelope[Entry]:
		return e.Meta.Signature, nil
	default:
		// Fallback: try to extract via JSON
		data, err := json.Marshal(envelope)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal envelope: %w", err)
		}

		var temp map[string]interface{}
		if err := json.Unmarshal(data, &temp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal envelope: %w", err)
		}

		if meta, ok := temp["meta"].(map[string]interface{}); ok {
			if sigData, ok := meta["signature"]; ok {
				sigJSON, err := json.Marshal(sigData)
				if err != nil {
					return nil, err
				}

				var sig Signature
				if err := json.Unmarshal(sigJSON, &sig); err != nil {
					return nil, err
				}

				return &sig, nil
			}
		}
	}

	return nil, nil
}

// KeyPair represents an Ed25519 key pair
type KeyPair struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// GenerateKeyPair generates a new Ed25519 key pair
func GenerateKeyPair() (*KeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keys: %w", err)
	}

	return &KeyPair{
		PublicKey:  base64.StdEncoding.EncodeToString(publicKey),
		PrivateKey: base64.StdEncoding.EncodeToString(privateKey),
	}, nil
}

// ParsePublicKey parses a base64-encoded Ed25519 public key
func ParsePublicKey(publicKeyB64 string) (ed25519.PublicKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	if len(keyBytes) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key size: got %d, want %d", len(keyBytes), ed25519.PublicKeySize)
	}

	return ed25519.PublicKey(keyBytes), nil
}

// ParsePrivateKey parses a base64-encoded Ed25519 private key
func ParsePrivateKey(privateKeyB64 string) (ed25519.PrivateKey, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(privateKeyB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	if len(keyBytes) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key size: got %d, want %d", len(keyBytes), ed25519.PrivateKeySize)
	}

	return ed25519.PrivateKey(keyBytes), nil
}
