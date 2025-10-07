package ptd

import (
	"testing"
	"time"
)

func TestValidateTournament(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid tournament
	tournament := Tournament{
		Name:      "Test Tournament",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(24 * time.Hour),
		Status:    "published",
	}

	if err := validator.validateTournament(tournament); err != nil {
		t.Errorf("Valid tournament failed validation: %v", err)
	}

	// Invalid: missing name
	invalid := Tournament{
		Status: "published",
	}

	if err := validator.validateTournament(invalid); err == nil {
		t.Error("Tournament with missing name should fail validation")
	}

	// Invalid: bad status
	badStatus := Tournament{
		Name:   "Test",
		Status: "invalid_status",
	}

	if err := validator.validateTournament(badStatus); err == nil {
		t.Error("Tournament with invalid status should fail validation")
	}

	// Invalid: end before start
	badDates := Tournament{
		Name:      "Test",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(-24 * time.Hour),
	}

	if err := validator.validateTournament(badDates); err == nil {
		t.Error("Tournament with end_date before start_date should fail validation")
	}
}

func TestValidateEvent(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid event
	event := Event{
		TournamentID: GenerateID(TypeTournament),
		Name:         "Men's Singles",
		EventType:    "singles",
		Gender:       "male",
		Status:       "published",
	}

	if err := validator.validateEvent(event); err != nil {
		t.Errorf("Valid event failed validation: %v", err)
	}

	// Invalid: missing tournament_id
	invalid := Event{
		Name: "Test Event",
	}

	if err := validator.validateEvent(invalid); err == nil {
		t.Error("Event with missing tournament_id should fail validation")
	}

	// Invalid: missing name
	invalid2 := Event{
		TournamentID: GenerateID(TypeTournament),
	}

	if err := validator.validateEvent(invalid2); err == nil {
		t.Error("Event with missing name should fail validation")
	}

	// Invalid: bad event type
	badType := Event{
		TournamentID: GenerateID(TypeTournament),
		Name:         "Test",
		EventType:    "invalid",
	}

	if err := validator.validateEvent(badType); err == nil {
		t.Error("Event with invalid type should fail validation")
	}
}

func TestValidateMatch(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid match
	match := Match{
		EventID:     GenerateID(TypeEvent),
		MatchNumber: "M1",
		Status:      "scheduled",
	}

	if err := validator.validateMatch(match); err != nil {
		t.Errorf("Valid match failed validation: %v", err)
	}

	// Invalid: missing event_id
	invalid := Match{
		MatchNumber: "M1",
	}

	if err := validator.validateMatch(invalid); err == nil {
		t.Error("Match with missing event_id should fail validation")
	}

	// Invalid: missing match_number
	invalid2 := Match{
		EventID: GenerateID(TypeEvent),
	}

	if err := validator.validateMatch(invalid2); err == nil {
		t.Error("Match with missing match_number should fail validation")
	}
}

func TestValidateEntry(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid entry
	entry := Entry{
		EventID:   GenerateID(TypeEvent),
		EntryType: "individual",
		Status:    "registered",
		Players: []Player{
			{FirstName: "John", LastName: "Doe"},
		},
	}

	if err := validator.validateEntry(entry); err != nil {
		t.Errorf("Valid entry failed validation: %v", err)
	}

	// Invalid: missing event_id
	invalid := Entry{
		Players: []Player{{FirstName: "John"}},
	}

	if err := validator.validateEntry(invalid); err == nil {
		t.Error("Entry with missing event_id should fail validation")
	}

	// Invalid: no players or team
	invalid2 := Entry{
		EventID: GenerateID(TypeEvent),
	}

	if err := validator.validateEntry(invalid2); err == nil {
		t.Error("Entry with no players should fail validation")
	}
}

func TestValidatePlayer(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid player
	player := Player{
		FirstName: "John",
		LastName:  "Doe",
	}

	if err := validator.validatePlayer(player); err != nil {
		t.Errorf("Valid player failed validation: %v", err)
	}

	// Invalid: no name fields
	invalid := Player{}

	if err := validator.validatePlayer(invalid); err == nil {
		t.Error("Player with no name should fail validation")
	}

	// Valid: only display name
	displayOnly := Player{
		DisplayName: "John Doe",
	}

	if err := validator.validatePlayer(displayOnly); err != nil {
		t.Error("Player with only display_name should be valid")
	}
}

func TestValidateEnvelope(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid envelope
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{
			Name:   "Test Tournament",
			Status: "published",
		},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			Version:   1,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	if err := validator.ValidateEnvelope(envelope); err != nil {
		t.Errorf("Valid envelope failed validation: %v", err)
	}

	// Invalid: missing ID
	invalid := &Envelope[Tournament]{
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{Schema: "ptd.v1.tournament@1.0.0"},
	}

	if err := validator.ValidateEnvelope(invalid); err == nil {
		t.Error("Envelope with missing ID should fail validation")
	}

	// Invalid: bad schema format
	badSchema := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{Schema: "invalid"},
	}

	if err := validator.ValidateEnvelope(badSchema); err == nil {
		t.Error("Envelope with invalid schema should fail validation")
	}
}

func TestValidateMapEnvelope(t *testing.T) {
	validator := NewSchemaValidator(false)

	// Valid map envelope
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

	if err := validator.ValidateEnvelope(envelope); err != nil {
		t.Errorf("Valid map envelope failed validation: %v", err)
	}

	// Invalid spec
	invalid := &Envelope[map[string]interface{}]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: map[string]interface{}{
			"status": "published",
			// Missing name
		},
		Meta: Meta{
			Schema: "ptd.v1.tournament@1.0.0",
		},
	}

	if err := validator.ValidateEnvelope(invalid); err == nil {
		t.Error("Map envelope with invalid spec should fail validation")
	}
}

func TestValidateSchemaVersion(t *testing.T) {
	tests := []struct {
		schema string
		valid  bool
	}{
		{"ptd.v1.tournament@1.0.0", true},
		{"ptd.v1.event@2.3.1", true},
		{"ptd.v2.match@1.0.0", true},
		{"invalid", false},
		{"ptd.tournament@1.0.0", false},  // missing v1
		{"ptd.v1.tournament", false},     // missing version
		{"ptd.v1.tournament@1.0", false}, // incomplete version
	}

	for _, tt := range tests {
		err := validateSchemaVersion(tt.schema)
		if tt.valid && err != nil {
			t.Errorf("Schema %s should be valid, got error: %v", tt.schema, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("Schema %s should be invalid", tt.schema)
		}
	}
}

func TestStrictMode(t *testing.T) {
	// Strict validator
	strict := NewSchemaValidator(true)
	lenient := NewSchemaValidator(false)

	envelope := &Envelope[map[string]interface{}]{
		ID:   GenerateID("custom_type"),
		Type: "custom_type",
		Spec: map[string]interface{}{"data": "test"},
		Meta: Meta{Schema: "ptd.v1.custom_type@1.0.0"},
	}

	// Lenient should allow unknown types
	if err := lenient.ValidateEnvelope(envelope); err != nil {
		t.Error("Lenient validator should allow unknown entity types")
	}

	// Strict should reject unknown types
	if err := strict.ValidateEnvelope(envelope); err == nil {
		t.Error("Strict validator should reject unknown entity types")
	}
}

func TestQuickValidation(t *testing.T) {
	// Test convenience functions
	envelope := &Envelope[Tournament]{
		ID:   GenerateID(TypeTournament),
		Type: TypeTournament,
		Spec: Tournament{Name: "Test"},
		Meta: Meta{
			Schema:    "ptd.v1.tournament@1.0.0",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Source:    "test",
		},
	}

	if err := ValidateEnvelopeQuick(envelope); err != nil {
		t.Errorf("Quick validation failed: %v", err)
	}

	if err := ValidateEnvelopeStrict(envelope); err != nil {
		t.Errorf("Strict validation failed: %v", err)
	}
}

func TestValidateTournamentWithVenue(t *testing.T) {
	validator := NewSchemaValidator(false)

	tournament := Tournament{
		Name: "Championship 2025",
		Venue: &Venue{
			Name:    "Sports Complex",
			City:    "Tokyo",
			Country: "Japan",
		},
		Status: "published",
	}

	if err := validator.validateTournament(tournament); err != nil {
		t.Errorf("Tournament with venue failed validation: %v", err)
	}
}

func TestValidateEventWithAgeGroup(t *testing.T) {
	validator := NewSchemaValidator(false)

	event := Event{
		TournamentID: GenerateID(TypeTournament),
		Name:         "U19 Singles",
		EventType:    "singles",
		AgeGroup: &AgeGroup{
			Name:   "Under 19",
			Code:   "U19",
			MaxAge: 19,
		},
	}

	if err := validator.validateEvent(event); err != nil {
		t.Errorf("Event with age group failed validation: %v", err)
	}
}
