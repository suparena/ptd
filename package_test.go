package ptd

import (
	"archive/zip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewPackage(t *testing.T) {
	pkg := NewPackage("Test package")

	if pkg.ID == "" {
		t.Error("Package ID should not be empty")
	}

	if pkg.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", pkg.Version)
	}

	if pkg.Manifest == nil {
		t.Fatal("Manifest should not be nil")
	}

	if pkg.Manifest.Description != "Test package" {
		t.Errorf("Expected description 'Test package', got %s", pkg.Manifest.Description)
	}

	if pkg.Manifest.Files == nil {
		t.Error("Files map should be initialized")
	}

	if pkg.Manifest.Entities == nil {
		t.Error("Entities map should be initialized")
	}
}

func TestPackage_AddEntities(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ptd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pkg := NewPackage("Test package")
	pkg.tempDir = tmpDir

	// Create test tournaments
	tournaments := []interface{}{
		Envelope[Tournament]{
			ID:   GenerateID(TypeTournament),
			Type: TypeTournament,
			Spec: Tournament{
				Name:      "Tournament 1",
				StartDate: time.Now(),
				EndDate:   time.Now().Add(24 * time.Hour),
				Status:    "draft",
			},
			Meta: Meta{
				Schema:    "ptd.v1.tournament@1.0.0",
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Source:    "test",
			},
		},
		Envelope[Tournament]{
			ID:   GenerateID(TypeTournament),
			Type: TypeTournament,
			Spec: Tournament{
				Name:      "Tournament 2",
				StartDate: time.Now().Add(7 * 24 * time.Hour),
				EndDate:   time.Now().Add(9 * 24 * time.Hour),
				Status:    "published",
			},
			Meta: Meta{
				Schema:    "ptd.v1.tournament@1.0.0",
				Version:   1,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Source:    "test",
			},
		},
	}

	// Add entities
	if err := pkg.AddEntities("tournament", tournaments); err != nil {
		t.Fatalf("Failed to add entities: %v", err)
	}

	// Check manifest
	if count, exists := pkg.Manifest.Entities["tournament"]; !exists {
		t.Error("Tournament count not found in manifest")
	} else if count.Count != 2 {
		t.Errorf("Expected 2 tournaments, got %d", count.Count)
	}

	// Check file was created
	filePath := filepath.Join(tmpDir, "tournament", "tournaments.ndjson")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("NDJSON file was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read NDJSON file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in NDJSON, got %d", len(lines))
	}

	// Verify each line is valid JSON
	for i, line := range lines {
		var envelope map[string]interface{}
		if err := json.Unmarshal([]byte(line), &envelope); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}

func TestPackage_CreateArchive(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "ptd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pkg := NewPackage("Test archive")
	pkg.tempDir = tmpDir

	// Add some test data
	tournaments := []interface{}{
		Envelope[Tournament]{
			ID:   GenerateID(TypeTournament),
			Type: TypeTournament,
			Spec: Tournament{Name: "Test Tournament"},
			Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
		},
	}

	if err := pkg.AddEntities("tournament", tournaments); err != nil {
		t.Fatalf("Failed to add entities: %v", err)
	}

	// Create archive
	archivePath := filepath.Join(tmpDir, "test.ptd")
	if err := pkg.CreateArchive(archivePath); err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Verify archive exists
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		t.Fatal("Archive was not created")
	}

	// Open and verify archive
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer reader.Close()

	// Check for manifest
	manifestFound := false
	tournamentFileFound := false

	for _, file := range reader.File {
		switch file.Name {
		case "manifest.json":
			manifestFound = true

			// Read and verify manifest
			rc, err := file.Open()
			if err != nil {
				t.Errorf("Failed to open manifest: %v", err)
				continue
			}
			defer rc.Close()

			data, err := io.ReadAll(rc)
			if err != nil {
				t.Errorf("Failed to read manifest: %v", err)
				continue
			}

			var manifest Manifest
			if err := json.Unmarshal(data, &manifest); err != nil {
				t.Errorf("Failed to unmarshal manifest: %v", err)
			}

		case "tournament/tournaments.ndjson":
			tournamentFileFound = true
		}
	}

	if !manifestFound {
		t.Error("Manifest not found in archive")
	}

	if !tournamentFileFound {
		t.Error("Tournament file not found in archive")
	}
}

func TestOpenPackage(t *testing.T) {
	// Create a test package
	tmpDir, err := os.MkdirTemp("", "ptd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a separate work directory for package content
	workDir, err := os.MkdirTemp("", "ptd-work-*")
	if err != nil {
		t.Fatalf("Failed to create work dir: %v", err)
	}
	defer os.RemoveAll(workDir)

	// Create package with known content
	pkg := NewPackage("Test package for opening")
	pkg.tempDir = workDir

	events := []interface{}{
		Envelope[Event]{
			ID:   GenerateID(TypeEvent),
			Type: TypeEvent,
			Spec: Event{
				Name:      "Men's Singles",
				EventCode: "MS",
			},
			Meta: Meta{Schema: "ptd.v1.event@1.0.0"},
		},
	}

	if err := pkg.AddEntities("event", events); err != nil {
		t.Fatalf("Failed to add entities: %v", err)
	}

	// Archive to separate directory
	archivePath := filepath.Join(tmpDir, "test-open.ptd")
	if err := pkg.CreateArchive(archivePath); err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Now open the package
	openedPkg, err := OpenPackage(archivePath)
	if err != nil {
		t.Fatalf("Failed to open package: %v", err)
	}

	if openedPkg.Manifest == nil {
		t.Fatal("Opened package should have manifest")
	}

	if openedPkg.Manifest.Description != "Test package for opening" {
		t.Errorf("Description mismatch: got %s", openedPkg.Manifest.Description)
	}

	// Verify entity count
	if count, exists := openedPkg.Manifest.Entities["event"]; !exists {
		t.Error("Event count not found in manifest")
	} else if count.Count != 1 {
		t.Errorf("Expected 1 event, got %d", count.Count)
	}
}

func TestOpenPackage_InvalidHash(t *testing.T) {
	// Create a corrupted package
	tmpDir, err := os.MkdirTemp("", "ptd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a simple ZIP with manifest
	archivePath := filepath.Join(tmpDir, "corrupt.ptd")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	zipWriter := zip.NewWriter(archive)

	// Add manifest with wrong hash
	manifest := Manifest{
		Version:     "1.0.0",
		Created:     time.Now(),
		Creator:     "test",
		Description: "Corrupt package",
		Files: map[string]*FileEntry{
			"data.txt": {
				Path: "data.txt",
				Size: 4,
				Hash: "wronghash123", // Intentionally wrong
			},
		},
		Entities: make(map[string]EntityCount),
	}

	manifestData, _ := json.Marshal(manifest)
	manifestWriter, _ := zipWriter.Create("manifest.json")
	manifestWriter.Write(manifestData)

	// Add data file
	dataWriter, _ := zipWriter.Create("data.txt")
	dataWriter.Write([]byte("test"))

	zipWriter.Close()
	archive.Close()

	// Try to open - should fail due to hash mismatch
	_, err = OpenPackage(archivePath)
	if err == nil {
		t.Error("Expected error for hash mismatch, got nil")
	}
}

func TestOpenPackage_MissingManifest(t *testing.T) {
	// Create ZIP without manifest
	tmpDir, err := os.MkdirTemp("", "ptd-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, "no-manifest.ptd")
	archive, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	zipWriter := zip.NewWriter(archive)

	// Add some file but no manifest
	writer, _ := zipWriter.Create("data.txt")
	writer.Write([]byte("test data"))

	zipWriter.Close()
	archive.Close()

	// Try to open - should fail
	_, err = OpenPackage(archivePath)
	if err != ErrManifestMissing {
		t.Errorf("Expected ErrManifestMissing, got %v", err)
	}
}

func TestDetectContentType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"data.json", "application/json"},
		{"tournaments.ndjson", "application/x-ndjson"},
		{"config.xml", "application/xml"},
		{"results.csv", "text/csv"},
		{"unknown.xyz", "application/octet-stream"},
		{"README", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			contentType := detectContentType(tt.path)
			if contentType != tt.expected {
				t.Errorf("detectContentType(%s) = %s, want %s", tt.path, contentType, tt.expected)
			}
		})
	}
}

func BenchmarkPackageCreation(b *testing.B) {
	tmpDir, _ := os.MkdirTemp("", "ptd-bench-*")
	defer os.RemoveAll(tmpDir)

	for i := 0; i < b.N; i++ {
		pkg := NewPackage("Benchmark package")
		pkg.tempDir = tmpDir

		// Add some entities
		tournaments := make([]interface{}, 10)
		for j := 0; j < 10; j++ {
			tournaments[j] = Envelope[Tournament]{
				ID:   GenerateID(TypeTournament),
				Type: TypeTournament,
				Spec: Tournament{Name: "Tournament"},
				Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
			}
		}

		pkg.AddEntities("tournament", tournaments)

		archivePath := filepath.Join(tmpDir, "bench.ptd")
		pkg.CreateArchive(archivePath)

		// Clean up for next iteration
		os.Remove(archivePath)
	}
}