package ptd_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/suparena/ptd"
)

func TestCreateTournament(t *testing.T) {
	// Create a tournament
	tournament := ptd.Tournament{
		Name:        "Summer Championship 2025",
		Description: "Annual summer table tennis championship",
		StartDate:   time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC),
		EndDate:     time.Date(2025, 7, 3, 18, 0, 0, 0, time.UTC),
		Status:      "published",
		Format:      "round_robin",
		Venue: &ptd.Venue{
			Name:    "Sports Complex",
			City:    "San Francisco",
			Country: "USA",
		},
	}

	// Wrap in envelope
	envelope := ptd.Envelope[ptd.Tournament]{
		ID:   ptd.GenerateID(ptd.TypeTournament),
		Type: ptd.TypeTournament,
		Spec: tournament,
		Meta: ptd.Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "icc:test",
		},
	}

	// Validate
	if err := envelope.Validate(); err != nil {
		t.Fatalf("Failed to validate envelope: %v", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal envelope: %v", err)
	}

	fmt.Printf("Tournament PTD:\n%s\n", string(data))
}

func TestGenerateID(t *testing.T) {
	// Test ID generation
	id := ptd.GenerateID(ptd.TypeTournament)

	// Parse and validate
	prefix, entityType, identifier, err := ptd.ParseID(id)
	if err != nil {
		t.Fatalf("Failed to parse ID: %v", err)
	}

	if prefix != "ptd" {
		t.Errorf("Expected prefix 'ptd', got '%s'", prefix)
	}

	if entityType != ptd.TypeTournament {
		t.Errorf("Expected type '%s', got '%s'", ptd.TypeTournament, entityType)
	}

	if !ptd.IsULID(identifier) {
		t.Errorf("Identifier is not a valid ULID: %s", identifier)
	}

	fmt.Printf("Generated ID: %s\n", id)
	fmt.Printf("  Prefix: %s\n", prefix)
	fmt.Printf("  Type: %s\n", entityType)
	fmt.Printf("  Identifier: %s\n", identifier)
}

func TestCanonicalJSON(t *testing.T) {
	// Create an event
	event := ptd.Event{
		TournamentID: ptd.GenerateID(ptd.TypeTournament),
		Name:         "Men's Singles",
		EventCode:    "MS",
		EventType:    "singles",
		Gender:       "male",
		Format:       "single_elimination",
		MaxEntries:   64,
		StartDate:    time.Date(2025, 7, 1, 9, 0, 0, 0, time.UTC),
		EndDate:      time.Date(2025, 7, 2, 18, 0, 0, 0, time.UTC),
		Status:       "registration_open",
	}

	envelope := ptd.Envelope[ptd.Event]{
		ID:   ptd.GenerateID(ptd.TypeEvent),
		Type: ptd.TypeEvent,
		Spec: event,
		Meta: ptd.Meta{
			Schema:    "ptd.v1.event@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "icc:test",
			Tags:      []string{"official", "rated"},
		},
	}

	// Get canonical JSON (for signing)
	canonical, err := envelope.CanonicalJSON()
	if err != nil {
		t.Fatalf("Failed to get canonical JSON: %v", err)
	}

	fmt.Printf("Canonical JSON length: %d bytes\n", len(canonical))

	// Verify it's valid JSON
	var check map[string]interface{}
	if err := json.Unmarshal(canonical, &check); err != nil {
		t.Fatalf("Canonical JSON is not valid: %v", err)
	}
}
