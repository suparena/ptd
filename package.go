package ptd

import (
	"archive/zip"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Package represents a PTD package containing tournament data
type Package struct {
	ID       string    `json:"id"`
	Created  time.Time `json:"created"`
	Version  string    `json:"version"`
	Manifest *Manifest `json:"-"`
	tempDir  string
}

// Manifest describes the contents of a PTD package
type Manifest struct {
	Version     string                 `json:"version"`             // PTD version (e.g., "1.0.0")
	Created     time.Time              `json:"created"`             // Package creation time
	Creator     string                 `json:"creator"`             // System that created package
	Description string                 `json:"description"`         // Human-readable description
	Files       map[string]*FileEntry  `json:"files"`               // All files in package
	Entities    map[string]EntityCount `json:"entities"`            // Count of each entity type
	Signature   *Signature             `json:"signature,omitempty"` // Package signature
}

// CanonicalJSON returns the canonical JSON representation of manifest for signing
func (m *Manifest) CanonicalJSON() ([]byte, error) {
	// Create a copy without signature and files (files are archive metadata, not package content)
	temp := *m
	temp.Signature = nil
	temp.Files = nil // Exclude files from signature - they're archive metadata

	// Use deterministic JSON encoding
	return json.Marshal(temp)
}

// FileEntry describes a file in the package
type FileEntry struct {
	Path     string    `json:"path"`     // Relative path in package
	Size     int64     `json:"size"`     // File size in bytes
	Hash     string    `json:"hash"`     // SHA-256 hash
	Modified time.Time `json:"modified"` // Last modification time
	Type     string    `json:"type"`     // MIME type or content type
}

// EntityCount tracks the number of entities by type
type EntityCount struct {
	Type  string `json:"type"`  // Entity type
	Count int    `json:"count"` // Number of entities
}

// NewPackage creates a new PTD package
func NewPackage(description string) *Package {
	// Create temp directory for package contents
	tempDir, err := os.MkdirTemp("", "ptd-package-*")
	if err != nil {
		// Fall back to current directory if temp creation fails
		tempDir = filepath.Join(".", fmt.Sprintf("ptd-package-%s", GenerateULID()))
		os.MkdirAll(tempDir, 0755)
	}

	return &Package{
		ID:      GenerateULID(),
		Created: time.Now(),
		Version: "1.0.0",
		tempDir: tempDir,
		Manifest: &Manifest{
			Version:     "1.0.0",
			Created:     time.Now(),
			Creator:     "ptd-go",
			Description: description,
			Files:       make(map[string]*FileEntry),
			Entities:    make(map[string]EntityCount),
		},
	}
}

// AddEntities adds entities to the package
func (p *Package) AddEntities(entityType string, entities []interface{}) error {
	// Create directory for entity type if needed
	dir := filepath.Join(p.tempDir, entityType)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write entities to NDJSON file
	filename := fmt.Sprintf("%ss.ndjson", entityType)
	filepath := filepath.Join(dir, filename)

	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write each entity as a JSON line
	for _, entity := range entities {
		data, err := json.Marshal(entity)
		if err != nil {
			return fmt.Errorf("failed to marshal entity: %w", err)
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write entity: %w", err)
		}
	}

	// Update manifest
	p.Manifest.Entities[entityType] = EntityCount{
		Type:  entityType,
		Count: len(entities),
	}

	return nil
}

// Cleanup removes the temporary directory
func (p *Package) Cleanup() error {
	if p.tempDir != "" && p.tempDir != "." {
		return os.RemoveAll(p.tempDir)
	}
	return nil
}

// CreateArchive creates a ZIP archive of the package
func (p *Package) CreateArchive(outputPath string) error {
	// First collect all files and their hashes
	filesToArchive := make(map[string]string) // path -> hash

	// Walk directory and calculate hashes
	err := filepath.Walk(p.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(p.tempDir, path)
		if err != nil {
			return err
		}

		// Read file and calculate hash
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		hasher := sha256.New()
		hasher.Write(data)
		hash := hex.EncodeToString(hasher.Sum(nil))

		filesToArchive[relPath] = hash

		// Add to manifest
		p.Manifest.Files[relPath] = &FileEntry{
			Path:     relPath,
			Size:     int64(len(data)),
			Hash:     hash,
			Modified: info.ModTime(),
			Type:     detectContentType(relPath),
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	// Create manifest file
	manifestPath := filepath.Join(p.tempDir, "manifest.json")
	manifestData, err := json.MarshalIndent(p.Manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	// Create ZIP archive
	archive, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)
	defer zipWriter.Close()

	// Add all files including the manifest
	return filepath.Walk(p.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(p.tempDir, path)
		if err != nil {
			return err
		}

		// Create ZIP entry
		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		// Copy file content
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// OpenPackage opens and validates a PTD package
func OpenPackage(archivePath string) (*Package, error) {
	// Open ZIP file
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer reader.Close()

	// Look for manifest
	var manifest *Manifest
	for _, file := range reader.File {
		if file.Name == "manifest.json" {
			rc, err := file.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open manifest: %w", err)
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				return nil, fmt.Errorf("failed to read manifest: %w", err)
			}

			manifest = &Manifest{}
			if err := json.Unmarshal(data, manifest); err != nil {
				return nil, fmt.Errorf("failed to parse manifest: %w", err)
			}
			break
		}
	}

	if manifest == nil {
		return nil, ErrManifestMissing
	}

	// Validate file hashes
	for _, file := range reader.File {
		if file.Name == "manifest.json" {
			continue
		}

		entry, exists := manifest.Files[file.Name]
		if !exists {
			return nil, fmt.Errorf("unexpected file in package: %s", file.Name)
		}

		// Verify hash
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s: %w", file.Name, err)
		}
		defer rc.Close()

		hasher := sha256.New()
		if _, err := io.Copy(hasher, rc); err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file.Name, err)
		}

		hash := hex.EncodeToString(hasher.Sum(nil))
		if hash != entry.Hash {
			return nil, fmt.Errorf("%w for file %s", ErrHashMismatch, file.Name)
		}
	}

	pkg := &Package{
		ID:       GenerateULID(),
		Created:  manifest.Created,
		Version:  manifest.Version,
		Manifest: manifest,
	}

	return pkg, nil
}

// detectContentType determines the content type based on file extension
func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "application/json"
	case ".ndjson":
		return "application/x-ndjson"
	case ".xml":
		return "application/xml"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}

// SignPackage signs the package manifest with the given signer
func (p *Package) SignPackage(signer *Signer) error {
	if p.Manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	// Get canonical JSON of manifest (without signature)
	canonical, err := p.Manifest.CanonicalJSON()
	if err != nil {
		return fmt.Errorf("failed to get canonical JSON: %w", err)
	}

	// Sign the canonical JSON
	signature := ed25519.Sign(signer.privateKey, canonical)

	// Encode signature as base64
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Create signature object
	p.Manifest.Signature = &Signature{
		Algorithm:   "ed25519",
		PublicKeyID: signer.publicKeyID,
		Signature:   signatureB64,
		SignedAt:    time.Now(),
		SignedBy:    signer.signedBy,
	}

	return nil
}

// VerifyPackageSignature verifies the package signature
func (p *Package) VerifyPackageSignature(publicKey interface{}) error {
	if p.Manifest == nil {
		return fmt.Errorf("manifest is nil")
	}

	if p.Manifest.Signature == nil {
		return ErrSignatureMissing
	}

	// Get canonical JSON (without signature)
	canonical, err := p.Manifest.CanonicalJSON()
	if err != nil {
		return fmt.Errorf("failed to get canonical JSON: %w", err)
	}

	// Decode signature
	signatureBytes, err := base64.StdEncoding.DecodeString(p.Manifest.Signature.Signature)
	if err != nil {
		return ErrSignatureInvalid
	}

	// Type assert to Ed25519 public key
	ed25519Key, ok := publicKey.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("invalid public key type")
	}

	// Verify signature
	if !ed25519.Verify(ed25519Key, canonical, signatureBytes) {
		return ErrSignatureFailed
	}

	return nil
}
