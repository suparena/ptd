package ptd

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// IDGenerator generates unique identifiers for PTD entities
type IDGenerator struct {
	entropy *ulid.MonotonicEntropy
	mu      sync.Mutex
}

// NewIDGenerator creates a new ID generator
func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		entropy: ulid.Monotonic(rand.Reader, 0),
	}
}

// GenerateID generates a new PTD ID for the given entity type
func (g *IDGenerator) GenerateID(entityType string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), g.entropy)
	return fmt.Sprintf("ptd:%s:%s", entityType, strings.ToLower(id.String()))
}

// GenerateULID generates a raw ULID
func (g *IDGenerator) GenerateULID() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), g.entropy)
	return id.String()
}

// ParseID parses a PTD ID and returns its components
func ParseID(id string) (prefix string, entityType string, identifier string, err error) {
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("%w: expected format ptd:type:identifier", ErrInvalidID)
	}

	if parts[0] != "ptd" {
		return "", "", "", fmt.Errorf("%w: ID must start with 'ptd:'", ErrInvalidID)
	}

	return parts[0], parts[1], parts[2], nil
}

// ValidateID checks if an ID is valid PTD format
func ValidateID(id string) bool {
	_, _, _, err := ParseID(id)
	return err == nil
}

// IsULID checks if a string is a valid ULID
func IsULID(s string) bool {
	_, err := ulid.ParseStrict(s)
	return err == nil
}

// Global ID generator instance
var defaultGenerator = NewIDGenerator()

// GenerateID generates a new PTD ID using the default generator
func GenerateID(entityType string) string {
	return defaultGenerator.GenerateID(entityType)
}

// GenerateULID generates a raw ULID using the default generator
func GenerateULID() string {
	return defaultGenerator.GenerateULID()
}

// Standard entity type constants
const (
	TypeTournament = "tournament"
	TypeEvent      = "event"
	TypeMatch      = "match"
	TypeEntry      = "entry"
	TypePlayer     = "player"
	TypeRound      = "round"
	TypeBracket    = "bracket"
	TypeVenue      = "venue"
	TypeOrganizer  = "organizer"
	TypeOfficial   = "official"
)
