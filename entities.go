package ptd

import (
	"time"
)

// Tournament represents a tournament entity
type Tournament struct {
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     time.Time  `json:"end_date"`
	TimeZone    string     `json:"time_zone,omitempty"`
	Status      string     `json:"status"` // draft, published, in_progress, completed
	Venue       *Venue     `json:"venue,omitempty"`
	Organizer   *Organizer `json:"organizer,omitempty"`
	Format      string     `json:"format,omitempty"` // single_elimination, round_robin, etc.
	Rules       *Rules     `json:"rules,omitempty"`
	Website     string     `json:"website,omitempty"`
	ContactInfo *Contact   `json:"contact_info,omitempty"`
}

// Event represents an event within a tournament
type Event struct {
	TournamentID string    `json:"tournament_id"`
	Name         string    `json:"name"`
	EventCode    string    `json:"event_code"` // e.g., "MS", "WD", "XD"
	EventType    string    `json:"event_type"` // singles, doubles, team
	Gender       string    `json:"gender,omitempty"` // male, female, mixed
	AgeGroup     *AgeGroup `json:"age_group,omitempty"`
	Format       string    `json:"format,omitempty"` // Can override tournament format
	MaxEntries   int       `json:"max_entries,omitempty"`
	EntryFee     *Money    `json:"entry_fee,omitempty"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Status       string    `json:"status"`
}

// Match represents a match in a tournament
type Match struct {
	EventID      string       `json:"event_id"`
	RoundID      string       `json:"round_id,omitempty"`
	BracketID    string       `json:"bracket_id,omitempty"`
	MatchNumber  string       `json:"match_number"`
	ScheduledAt  *time.Time   `json:"scheduled_at,omitempty"`
	Court        string       `json:"court,omitempty"`
	Status       string       `json:"status"` // scheduled, in_progress, completed, cancelled
	HomeEntry    *EntryRef    `json:"home_entry,omitempty"`
	AwayEntry    *EntryRef    `json:"away_entry,omitempty"`
	Winner       string       `json:"winner,omitempty"` // entry_id of winner
	Score        *Score       `json:"score,omitempty"`
	Officials    []Official   `json:"officials,omitempty"`
	StreamingURL string       `json:"streaming_url,omitempty"`
	Notes        string       `json:"notes,omitempty"`
}

// Entry represents a participant entry in an event
type Entry struct {
	EventID      string        `json:"event_id"`
	EntryType    string        `json:"entry_type"` // individual, doubles, team
	Status       string        `json:"status"`     // registered, confirmed, withdrawn
	Seed         *int          `json:"seed,omitempty"`
	Players      []Player      `json:"players"`
	Team         *Team         `json:"team,omitempty"`
	Registration *Registration `json:"registration,omitempty"`
}

// Player represents an individual player
type Player struct {
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name,omitempty"`
	Country     string    `json:"country,omitempty"`
	Club        string    `json:"club,omitempty"`
	Rating      *Rating   `json:"rating,omitempty"`
	BirthDate   time.Time `json:"birth_date,omitempty"`
	Gender      string    `json:"gender,omitempty"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	PlayerID    string    `json:"player_id,omitempty"` // External ID (e.g., ITTF ID)
}

// Score represents match score
type Score struct {
	Sets       []SetScore `json:"sets"`
	Final      string     `json:"final"` // e.g., "3-1"
	Duration   *Duration  `json:"duration,omitempty"`
	Retirement bool       `json:"retirement,omitempty"`
	Walkover   bool       `json:"walkover,omitempty"`
	Disqualify bool       `json:"disqualify,omitempty"`
}

// SetScore represents score for a single set/game
type SetScore struct {
	SetNumber  int    `json:"set_number"`
	HomeScore  int    `json:"home_score"`
	AwayScore  int    `json:"away_score"`
	Tiebreak   bool   `json:"tiebreak,omitempty"`
	Duration   string `json:"duration,omitempty"`
}

// Supporting types

// Venue represents a competition venue
type Venue struct {
	Name     string   `json:"name"`
	Address  string   `json:"address,omitempty"`
	City     string   `json:"city,omitempty"`
	State    string   `json:"state,omitempty"`
	Country  string   `json:"country,omitempty"`
	PostCode string   `json:"post_code,omitempty"`
	Courts   []string `json:"courts,omitempty"`
	Capacity int      `json:"capacity,omitempty"`
}

// Organizer represents tournament organizer
type Organizer struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"` // federation, club, company
	Contact *Contact `json:"contact,omitempty"`
	Website string   `json:"website,omitempty"`
	Logo    string   `json:"logo,omitempty"`
}

// AgeGroup represents age category
type AgeGroup struct {
	Name     string `json:"name"`     // e.g., "Under 19"
	Code     string `json:"code"`     // e.g., "U19"
	MinAge   int    `json:"min_age,omitempty"`
	MaxAge   int    `json:"max_age,omitempty"`
	CutoffDate time.Time `json:"cutoff_date,omitempty"`
}

// Rules represents tournament rules
type Rules struct {
	ScoringSystem string `json:"scoring_system"` // e.g., "best_of_5"
	GamePoints    int    `json:"game_points,omitempty"`
	TiebreakAt    int    `json:"tiebreak_at,omitempty"`
	ServiceChange int    `json:"service_change,omitempty"`
	TimeLimit     string `json:"time_limit,omitempty"`
	CustomRules   string `json:"custom_rules,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	Phone string `json:"phone,omitempty"`
	Role  string `json:"role,omitempty"`
}

// Money represents monetary amount
type Money struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"` // ISO 4217 code
}

// Rating represents player rating
type Rating struct {
	Value    int       `json:"value"`
	System   string    `json:"system"` // e.g., "ITTF", "USATT", "ELO"
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// Team represents a team entry
type Team struct {
	Name    string   `json:"name"`
	Code    string   `json:"code,omitempty"`
	Country string   `json:"country,omitempty"`
	Club    string   `json:"club,omitempty"`
	Players []string `json:"players"` // List of player IDs
}

// Registration represents entry registration details
type Registration struct {
	RegisteredAt time.Time `json:"registered_at"`
	ConfirmedAt  *time.Time `json:"confirmed_at,omitempty"`
	PaidAt       *time.Time `json:"paid_at,omitempty"`
	CheckedInAt  *time.Time `json:"checked_in_at,omitempty"`
	WithdrawnAt  *time.Time `json:"withdrawn_at,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// EntryRef is a reference to an entry
type EntryRef struct {
	EntryID     string `json:"entry_id"`
	DisplayName string `json:"display_name"`
	Seed        *int   `json:"seed,omitempty"`
}

// Official represents match official
type Official struct {
	Name string `json:"name"`
	Role string `json:"role"` // referee, umpire, line_judge
}

// Duration represents time duration
type Duration struct {
	Minutes int `json:"minutes"`
	Seconds int `json:"seconds,omitempty"`
}