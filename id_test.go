package ptd

import (
	"strings"
	"testing"
	"time"
)

func TestIDGenerator_GenerateID(t *testing.T) {
	gen := NewIDGenerator()

	// Test different entity types
	entityTypes := []string{
		TypeTournament,
		TypeEvent,
		TypeMatch,
		TypeEntry,
		TypePlayer,
	}

	for _, entityType := range entityTypes {
		t.Run(entityType, func(t *testing.T) {
			id := gen.GenerateID(entityType)

			// Check format
			if !strings.HasPrefix(id, "ptd:") {
				t.Errorf("ID should start with 'ptd:', got %s", id)
			}

			parts := strings.Split(id, ":")
			if len(parts) != 3 {
				t.Errorf("ID should have 3 parts, got %d: %s", len(parts), id)
			}

			if parts[1] != entityType {
				t.Errorf("Expected entity type %s, got %s", entityType, parts[1])
			}

			// Check ULID validity
			if !IsULID(parts[2]) {
				t.Errorf("Third part should be valid ULID: %s", parts[2])
			}
		})
	}
}

func TestIDGenerator_Uniqueness(t *testing.T) {
	gen := NewIDGenerator()
	ids := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		id := gen.GenerateID(TypeTournament)
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != count {
		t.Errorf("Expected %d unique IDs, got %d", count, len(ids))
	}
}

func TestIDGenerator_Monotonic(t *testing.T) {
	gen := NewIDGenerator()

	// Generate IDs in quick succession
	id1 := gen.GenerateULID()
	id2 := gen.GenerateULID()
	id3 := gen.GenerateULID()

	// They should be lexicographically sortable
	if id1 >= id2 {
		t.Errorf("ID1 should be less than ID2: %s >= %s", id1, id2)
	}
	if id2 >= id3 {
		t.Errorf("ID2 should be less than ID3: %s >= %s", id2, id3)
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		wantPrefix string
		wantType   string
		wantID     string
		wantErr    bool
	}{
		{
			name:       "valid tournament ID",
			id:         "ptd:tournament:01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantPrefix: "ptd",
			wantType:   "tournament",
			wantID:     "01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantErr:    false,
		},
		{
			name:       "valid event ID",
			id:         "ptd:event:01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantPrefix: "ptd",
			wantType:   "event",
			wantID:     "01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantErr:    false,
		},
		{
			name:    "missing prefix",
			id:      "tournament:01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantErr: true,
		},
		{
			name:    "wrong prefix",
			id:      "xyz:tournament:01h5kxb5d4kcnqf7qb6zx4nfkp",
			wantErr: true,
		},
		{
			name:    "too few parts",
			id:      "ptd:tournament",
			wantErr: true,
		},
		{
			name:    "too many parts",
			id:      "ptd:tournament:01h5kxb5d4kcnqf7qb6zx4nfkp:extra",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, entityType, identifier, err := ParseID(tt.id)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseID() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseID() unexpected error: %v", err)
				return
			}

			if prefix != tt.wantPrefix {
				t.Errorf("ParseID() prefix = %v, want %v", prefix, tt.wantPrefix)
			}
			if entityType != tt.wantType {
				t.Errorf("ParseID() entityType = %v, want %v", entityType, tt.wantType)
			}
			if identifier != tt.wantID {
				t.Errorf("ParseID() identifier = %v, want %v", identifier, tt.wantID)
			}
		})
	}
}

func TestValidateID(t *testing.T) {
	tests := []struct {
		name  string
		id    string
		valid bool
	}{
		{"valid tournament", "ptd:tournament:01h5kxb5d4kcnqf7qb6zx4nfkp", true},
		{"valid event", "ptd:event:01h5kxb5d4kcnqf7qb6zx4nfkp", true},
		{"missing prefix", "tournament:01h5kxb5d4kcnqf7qb6zx4nfkp", false},
		{"wrong format", "ptd-tournament-123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateID(tt.id); got != tt.valid {
				t.Errorf("ValidateID() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestIsULID(t *testing.T) {
	gen := NewIDGenerator()

	tests := []struct {
		name  string
		s     string
		valid bool
	}{
		{"valid ULID", gen.GenerateULID(), true},
		{"valid ULID lowercase", strings.ToLower(gen.GenerateULID()), true},
		{"invalid - too short", "01h5kxb5d4", false},
		{"invalid - too long", "01h5kxb5d4kcnqf7qb6zx4nfkp123", false},
		{"invalid - bad chars", "01h5kxb5d4kcnqf7qb6zx4nfk!", false},
		{"empty", "", false},
		{"not ULID", "not-a-ulid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsULID(tt.s); got != tt.valid {
				t.Errorf("IsULID(%s) = %v, want %v", tt.s, got, tt.valid)
			}
		})
	}
}

func TestGlobalGenerators(t *testing.T) {
	// Test global GenerateID
	id := GenerateID(TypeMatch)
	if !ValidateID(id) {
		t.Errorf("Global GenerateID produced invalid ID: %s", id)
	}

	// Test global GenerateULID
	ulid := GenerateULID()
	if !IsULID(ulid) {
		t.Errorf("Global GenerateULID produced invalid ULID: %s", ulid)
	}
}

func BenchmarkGenerateID(b *testing.B) {
	gen := NewIDGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateID(TypeTournament)
	}
}

func BenchmarkGenerateULID(b *testing.B) {
	gen := NewIDGenerator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = gen.GenerateULID()
	}
}

func BenchmarkParseID(b *testing.B) {
	id := "ptd:tournament:01h5kxb5d4kcnqf7qb6zx4nfkp"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _, _, _ = ParseID(id)
	}
}

func TestIDGeneratorConcurrency(t *testing.T) {
	gen := NewIDGenerator()
	done := make(chan string, 100)

	// Generate IDs concurrently
	for i := 0; i < 100; i++ {
		go func() {
			id := gen.GenerateID(TypeTournament)
			done <- id
		}()
	}

	// Collect all IDs
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := <-done
		if ids[id] {
			t.Errorf("Duplicate ID in concurrent generation: %s", id)
		}
		ids[id] = true
	}

	if len(ids) != 100 {
		t.Errorf("Expected 100 unique IDs, got %d", len(ids))
	}
}

func TestULIDTimeExtraction(t *testing.T) {
	gen := NewIDGenerator()

	// Generate ID and wait a bit
	before := time.Now()
	time.Sleep(10 * time.Millisecond)

	ulid1 := gen.GenerateULID()

	time.Sleep(10 * time.Millisecond)
	after := time.Now()

	// ULIDs encode timestamp, so we can verify timing
	// (Note: This is more of an integration test with the ULID library)
	ulid2 := gen.GenerateULID()

	// The second ULID should be greater (later timestamp)
	if ulid1 >= ulid2 {
		t.Errorf("Later ULID should be greater: %s >= %s", ulid1, ulid2)
	}

	// Basic sanity check on timing
	elapsed := after.Sub(before)
	if elapsed < 20*time.Millisecond {
		t.Errorf("Test timing issue: elapsed %v", elapsed)
	}
}