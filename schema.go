package ptd

import (
	"fmt"
	"reflect"
	"strings"
)

// SchemaValidator validates PTD entities against their schemas
type SchemaValidator struct {
	strictMode bool
}

// NewSchemaValidator creates a new schema validator
func NewSchemaValidator(strict bool) *SchemaValidator {
	return &SchemaValidator{
		strictMode: strict,
	}
}

// ValidateEntity validates an entity's spec against its schema
func (v *SchemaValidator) ValidateEntity(entityType string, spec interface{}) error {
	switch entityType {
	case TypeTournament:
		return v.validateTournament(spec)
	case TypeEvent:
		return v.validateEvent(spec)
	case TypeMatch:
		return v.validateMatch(spec)
	case TypeEntry:
		return v.validateEntry(spec)
	case TypePlayer:
		return v.validatePlayer(spec)
	default:
		// Unknown entity type - allow in non-strict mode
		if v.strictMode {
			return fmt.Errorf("%w: unknown entity type: %s", ErrValidation, entityType)
		}
		return nil
	}
}

// ValidateEnvelope validates an entire envelope (structure + spec)
func (v *SchemaValidator) ValidateEnvelope(envelope interface{}) error {
	// Use reflection to extract fields
	val := reflect.ValueOf(envelope)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Extract ID field
	idField := val.FieldByName("ID")
	if !idField.IsValid() || idField.String() == "" {
		return fmt.Errorf("%w: missing ID field", ErrValidation)
	}

	// Validate ID format
	if !ValidateID(idField.String()) {
		return fmt.Errorf("%w: invalid ID format: %s", ErrValidation, idField.String())
	}

	// Extract Type field
	typeField := val.FieldByName("Type")
	if !typeField.IsValid() || typeField.String() == "" {
		return fmt.Errorf("%w: missing Type field", ErrValidation)
	}

	// Extract Meta field
	metaField := val.FieldByName("Meta")
	if !metaField.IsValid() {
		return fmt.Errorf("%w: missing Meta field", ErrValidation)
	}

	// Validate Meta.Schema
	schemaField := metaField.FieldByName("Schema")
	if !schemaField.IsValid() || schemaField.String() == "" {
		return fmt.Errorf("%w: missing Meta.Schema", ErrValidation)
	}

	// Validate schema format
	if err := validateSchemaVersion(schemaField.String()); err != nil {
		return err
	}

	// Extract and validate Spec
	specField := val.FieldByName("Spec")
	if !specField.IsValid() {
		return fmt.Errorf("%w: missing Spec field", ErrValidation)
	}

	// Validate spec content
	return v.ValidateEntity(typeField.String(), specField.Interface())
}

// validateTournament validates a Tournament spec
func (v *SchemaValidator) validateTournament(spec interface{}) error {
	tournament, ok := spec.(Tournament)
	if !ok {
		// Try map[string]interface{} for generic envelopes
		return v.validateTournamentMap(spec)
	}

	// Required fields
	if tournament.Name == "" {
		return fmt.Errorf("%w: tournament.name is required", ErrMissingField)
	}

	// Validate status
	validStatuses := []string{"draft", "published", "in_progress", "completed", "cancelled"}
	if tournament.Status != "" && !contains(validStatuses, tournament.Status) {
		return fmt.Errorf("%w: invalid tournament.status: %s", ErrValidation, tournament.Status)
	}

	// Validate dates
	if !tournament.StartDate.IsZero() && !tournament.EndDate.IsZero() {
		if tournament.EndDate.Before(tournament.StartDate) {
			return fmt.Errorf("%w: tournament.end_date must be after start_date", ErrValidation)
		}
	}

	return nil
}

// validateTournamentMap validates a tournament from map[string]interface{}
func (v *SchemaValidator) validateTournamentMap(spec interface{}) error {
	m, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%w: tournament spec must be object", ErrInvalidFormat)
	}

	// Required: name
	name, ok := m["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("%w: tournament.name is required", ErrMissingField)
	}

	// Validate status if present
	if status, ok := m["status"].(string); ok {
		validStatuses := []string{"draft", "published", "in_progress", "completed", "cancelled"}
		if !contains(validStatuses, status) {
			return fmt.Errorf("%w: invalid tournament.status: %s", ErrValidation, status)
		}
	}

	return nil
}

// validateEvent validates an Event spec
func (v *SchemaValidator) validateEvent(spec interface{}) error {
	event, ok := spec.(Event)
	if !ok {
		return v.validateEventMap(spec)
	}

	// Required fields
	if event.TournamentID == "" {
		return fmt.Errorf("%w: event.tournament_id is required", ErrMissingField)
	}

	if event.Name == "" {
		return fmt.Errorf("%w: event.name is required", ErrMissingField)
	}

	// Validate tournament_id format
	if !ValidateID(event.TournamentID) {
		return fmt.Errorf("%w: invalid event.tournament_id format", ErrValidation)
	}

	// Validate event type
	validTypes := []string{"singles", "doubles", "team", "mixed"}
	if event.EventType != "" && !contains(validTypes, event.EventType) {
		return fmt.Errorf("%w: invalid event.event_type: %s", ErrValidation, event.EventType)
	}

	// Validate gender
	validGenders := []string{"male", "female", "mixed"}
	if event.Gender != "" && !contains(validGenders, event.Gender) {
		return fmt.Errorf("%w: invalid event.gender: %s", ErrValidation, event.Gender)
	}

	return nil
}

// validateEventMap validates an event from map[string]interface{}
func (v *SchemaValidator) validateEventMap(spec interface{}) error {
	m, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%w: event spec must be object", ErrInvalidFormat)
	}

	// Required: tournament_id
	tournamentID, ok := m["tournament_id"].(string)
	if !ok || tournamentID == "" {
		return fmt.Errorf("%w: event.tournament_id is required", ErrMissingField)
	}

	// Required: name
	name, ok := m["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("%w: event.name is required", ErrMissingField)
	}

	return nil
}

// validateMatch validates a Match spec
func (v *SchemaValidator) validateMatch(spec interface{}) error {
	match, ok := spec.(Match)
	if !ok {
		return v.validateMatchMap(spec)
	}

	// Required fields
	if match.EventID == "" {
		return fmt.Errorf("%w: match.event_id is required", ErrMissingField)
	}

	if match.MatchNumber == "" {
		return fmt.Errorf("%w: match.match_number is required", ErrMissingField)
	}

	// Validate status
	validStatuses := []string{"scheduled", "in_progress", "completed", "cancelled"}
	if match.Status != "" && !contains(validStatuses, match.Status) {
		return fmt.Errorf("%w: invalid match.status: %s", ErrValidation, match.Status)
	}

	// Validate winner if present
	if match.Winner != "" && !ValidateID(match.Winner) {
		return fmt.Errorf("%w: invalid match.winner format", ErrValidation)
	}

	return nil
}

// validateMatchMap validates a match from map[string]interface{}
func (v *SchemaValidator) validateMatchMap(spec interface{}) error {
	m, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%w: match spec must be object", ErrInvalidFormat)
	}

	// Required: event_id
	eventID, ok := m["event_id"].(string)
	if !ok || eventID == "" {
		return fmt.Errorf("%w: match.event_id is required", ErrMissingField)
	}

	return nil
}

// validateEntry validates an Entry spec
func (v *SchemaValidator) validateEntry(spec interface{}) error {
	entry, ok := spec.(Entry)
	if !ok {
		return v.validateEntryMap(spec)
	}

	// Required fields
	if entry.EventID == "" {
		return fmt.Errorf("%w: entry.event_id is required", ErrMissingField)
	}

	// Validate entry type
	validTypes := []string{"individual", "doubles", "team"}
	if entry.EntryType != "" && !contains(validTypes, entry.EntryType) {
		return fmt.Errorf("%w: invalid entry.entry_type: %s", ErrValidation, entry.EntryType)
	}

	// Validate status
	validStatuses := []string{"registered", "confirmed", "withdrawn", "cancelled"}
	if entry.Status != "" && !contains(validStatuses, entry.Status) {
		return fmt.Errorf("%w: invalid entry.status: %s", ErrValidation, entry.Status)
	}

	// Validate players based on entry type
	if len(entry.Players) == 0 && entry.Team == nil {
		return fmt.Errorf("%w: entry must have players or team", ErrValidation)
	}

	return nil
}

// validateEntryMap validates an entry from map[string]interface{}
func (v *SchemaValidator) validateEntryMap(spec interface{}) error {
	m, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%w: entry spec must be object", ErrInvalidFormat)
	}

	// Required: event_id
	eventID, ok := m["event_id"].(string)
	if !ok || eventID == "" {
		return fmt.Errorf("%w: entry.event_id is required", ErrMissingField)
	}

	return nil
}

// validatePlayer validates a Player spec
func (v *SchemaValidator) validatePlayer(spec interface{}) error {
	player, ok := spec.(Player)
	if !ok {
		return v.validatePlayerMap(spec)
	}

	// Required fields
	if player.FirstName == "" && player.LastName == "" && player.DisplayName == "" {
		return fmt.Errorf("%w: player must have at least one name field", ErrMissingField)
	}

	return nil
}

// validatePlayerMap validates a player from map[string]interface{}
func (v *SchemaValidator) validatePlayerMap(spec interface{}) error {
	m, ok := spec.(map[string]interface{})
	if !ok {
		return fmt.Errorf("%w: player spec must be object", ErrInvalidFormat)
	}

	// At least one name field required
	firstName, _ := m["first_name"].(string)
	lastName, _ := m["last_name"].(string)
	displayName, _ := m["display_name"].(string)

	if firstName == "" && lastName == "" && displayName == "" {
		return fmt.Errorf("%w: player must have at least one name field", ErrMissingField)
	}

	return nil
}

// validateSchemaVersion validates schema version format
func validateSchemaVersion(schema string) error {
	// Expected format: ptd.v1.tournament@1.0.0
	parts := strings.Split(schema, "@")
	if len(parts) != 2 {
		return fmt.Errorf("%w: schema must be in format 'ptd.v1.type@version'", ErrInvalidSchema)
	}

	schemaPart := parts[0]
	versionPart := parts[1]

	// Validate schema part
	if !strings.HasPrefix(schemaPart, "ptd.v") {
		return fmt.Errorf("%w: schema must start with 'ptd.v'", ErrInvalidSchema)
	}

	// Validate version part (simple semver check)
	versionParts := strings.Split(versionPart, ".")
	if len(versionParts) != 3 {
		return fmt.Errorf("%w: version must be semantic (major.minor.patch)", ErrInvalidSchema)
	}

	return nil
}

// contains checks if a string slice contains a value
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// ValidateEnvelopeQuick is a convenience function for quick envelope validation
func ValidateEnvelopeQuick(envelope interface{}) error {
	validator := NewSchemaValidator(false)
	return validator.ValidateEnvelope(envelope)
}

// ValidateEnvelopeStrict is a convenience function for strict envelope validation
func ValidateEnvelopeStrict(envelope interface{}) error {
	validator := NewSchemaValidator(true)
	return validator.ValidateEnvelope(envelope)
}
