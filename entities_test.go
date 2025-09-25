package ptd

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTournament_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tournament := Tournament{
		Name:        "Summer Championship 2025",
		Description: "Annual summer championship",
		StartDate:   now,
		EndDate:     now.Add(48 * time.Hour),
		TimeZone:    "America/New_York",
		Status:      "published",
		Format:      "round_robin",
		Website:     "https://example.com",
		Venue: &Venue{
			Name:     "Sports Complex",
			Address:  "123 Main St",
			City:     "New York",
			State:    "NY",
			Country:  "USA",
			PostCode: "10001",
			Courts:   []string{"Court 1", "Court 2", "Court 3"},
			Capacity: 500,
		},
		Organizer: &Organizer{
			Name:    "Table Tennis Federation",
			Type:    "federation",
			Website: "https://ttf.example.com",
		},
		Rules: &Rules{
			ScoringSystem: "best_of_5",
			GamePoints:    11,
			TiebreakAt:    10,
			ServiceChange: 2,
			TimeLimit:     "15min",
		},
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(tournament, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal tournament: %v", err)
	}

	// Unmarshal back
	var decoded Tournament
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal tournament: %v", err)
	}

	// Check key fields
	if decoded.Name != tournament.Name {
		t.Errorf("Name mismatch: got %s, want %s", decoded.Name, tournament.Name)
	}

	if !decoded.StartDate.Equal(tournament.StartDate) {
		t.Errorf("StartDate mismatch: got %v, want %v", decoded.StartDate, tournament.StartDate)
	}

	if decoded.Venue == nil {
		t.Error("Venue should not be nil")
	} else if decoded.Venue.Name != tournament.Venue.Name {
		t.Errorf("Venue name mismatch: got %s, want %s", decoded.Venue.Name, tournament.Venue.Name)
	}
}

func TestEvent_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	event := Event{
		TournamentID: GenerateID(TypeTournament),
		Name:         "Men's Singles",
		EventCode:    "MS",
		EventType:    "singles",
		Gender:       "male",
		AgeGroup: &AgeGroup{
			Name:       "Under 19",
			Code:       "U19",
			MinAge:     0,
			MaxAge:     19,
			CutoffDate: now.Add(-19 * 365 * 24 * time.Hour),
		},
		Format:     "single_elimination",
		MaxEntries: 64,
		EntryFee: &Money{
			Amount:   50.00,
			Currency: "USD",
		},
		StartDate: now,
		EndDate:   now.Add(24 * time.Hour),
		Status:    "registration_open",
	}

	// Test marshaling
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	// Verify fields
	if decoded.EventCode != event.EventCode {
		t.Errorf("EventCode mismatch: got %s, want %s", decoded.EventCode, event.EventCode)
	}

	if decoded.AgeGroup == nil {
		t.Error("AgeGroup should not be nil")
	} else {
		if decoded.AgeGroup.Code != event.AgeGroup.Code {
			t.Errorf("AgeGroup code mismatch: got %s, want %s", decoded.AgeGroup.Code, event.AgeGroup.Code)
		}
	}

	if decoded.EntryFee == nil {
		t.Error("EntryFee should not be nil")
	} else {
		if decoded.EntryFee.Amount != event.EntryFee.Amount {
			t.Errorf("EntryFee amount mismatch: got %f, want %f", decoded.EntryFee.Amount, event.EntryFee.Amount)
		}
	}
}

func TestMatch_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	scheduledAt := now.Add(2 * time.Hour)

	match := Match{
		EventID:     GenerateID(TypeEvent),
		RoundID:     GenerateID(TypeRound),
		BracketID:   GenerateID(TypeBracket),
		MatchNumber: "M001",
		ScheduledAt: &scheduledAt,
		Court:       "Court 1",
		Status:      "completed",
		HomeEntry: &EntryRef{
			EntryID:     GenerateID(TypeEntry),
			DisplayName: "John Doe",
			Seed:        intPtr(1),
		},
		AwayEntry: &EntryRef{
			EntryID:     GenerateID(TypeEntry),
			DisplayName: "Jane Smith",
			Seed:        intPtr(4),
		},
		Winner: "ptd:entry:winner123",
		Score: &Score{
			Sets: []SetScore{
				{SetNumber: 1, HomeScore: 11, AwayScore: 6},
				{SetNumber: 2, HomeScore: 11, AwayScore: 8},
				{SetNumber: 3, HomeScore: 9, AwayScore: 11},
				{SetNumber: 4, HomeScore: 11, AwayScore: 7},
			},
			Final: "3-1",
			Duration: &Duration{
				Minutes: 45,
				Seconds: 30,
			},
		},
		Officials: []Official{
			{Name: "Referee One", Role: "referee"},
			{Name: "Umpire One", Role: "umpire"},
		},
		StreamingURL: "https://stream.example.com/match/M001",
		Notes:        "Great match!",
	}

	// Marshal and unmarshal
	data, err := json.MarshalIndent(match, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal match: %v", err)
	}

	var decoded Match
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal match: %v", err)
	}

	// Verify complex fields
	if decoded.Score == nil {
		t.Fatal("Score should not be nil")
	}

	if len(decoded.Score.Sets) != 4 {
		t.Errorf("Expected 4 sets, got %d", len(decoded.Score.Sets))
	}

	if decoded.Score.Final != "3-1" {
		t.Errorf("Expected final score '3-1', got %s", decoded.Score.Final)
	}

	if decoded.HomeEntry.Seed == nil || *decoded.HomeEntry.Seed != 1 {
		t.Error("HomeEntry seed should be 1")
	}

	if len(decoded.Officials) != 2 {
		t.Errorf("Expected 2 officials, got %d", len(decoded.Officials))
	}
}

func TestEntry_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	entry := Entry{
		EventID:   GenerateID(TypeEvent),
		EntryType: "doubles",
		Status:    "confirmed",
		Seed:      intPtr(2),
		Players: []Player{
			{
				FirstName:   "John",
				LastName:    "Doe",
				DisplayName: "John Doe",
				Country:     "USA",
				Club:        "Table Tennis Club",
				Rating: &Rating{
					Value:     2100,
					System:    "ITTF",
					UpdatedAt: now,
				},
				BirthDate: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Gender:    "male",
				Email:     "john@example.com",
				PlayerID:  "ITTF123456",
			},
			{
				FirstName:   "Mike",
				LastName:    "Smith",
				DisplayName: "Mike Smith",
				Country:     "USA",
				Club:        "Table Tennis Club",
				Rating: &Rating{
					Value:     2050,
					System:    "ITTF",
					UpdatedAt: now,
				},
				BirthDate: time.Date(2001, 6, 15, 0, 0, 0, 0, time.UTC),
				Gender:    "male",
				Email:     "mike@example.com",
				PlayerID:  "ITTF789012",
			},
		},
		Registration: &Registration{
			RegisteredAt: now.Add(-7 * 24 * time.Hour),
			ConfirmedAt:  &now,
			PaidAt:       &now,
			Notes:        "Paid via credit card",
		},
	}

	// Test JSON roundtrip
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal entry: %v", err)
	}

	var decoded Entry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal entry: %v", err)
	}

	// Verify
	if len(decoded.Players) != 2 {
		t.Errorf("Expected 2 players, got %d", len(decoded.Players))
	}

	if decoded.Seed == nil || *decoded.Seed != 2 {
		t.Error("Seed should be 2")
	}

	if decoded.Registration == nil {
		t.Fatal("Registration should not be nil")
	}

	if decoded.Registration.PaidAt == nil {
		t.Error("PaidAt should not be nil")
	}
}

func TestScore_SpecialCases(t *testing.T) {
	tests := []struct {
		name  string
		score Score
	}{
		{
			name: "retirement",
			score: Score{
				Sets: []SetScore{
					{SetNumber: 1, HomeScore: 11, AwayScore: 6},
					{SetNumber: 2, HomeScore: 5, AwayScore: 3},
				},
				Final:      "RET",
				Retirement: true,
			},
		},
		{
			name: "walkover",
			score: Score{
				Final:    "W/O",
				Walkover: true,
			},
		},
		{
			name: "disqualification",
			score: Score{
				Sets: []SetScore{
					{SetNumber: 1, HomeScore: 11, AwayScore: 9},
				},
				Final:      "DQ",
				Disqualify: true,
			},
		},
		{
			name: "tiebreak",
			score: Score{
				Sets: []SetScore{
					{SetNumber: 1, HomeScore: 11, AwayScore: 9},
					{SetNumber: 2, HomeScore: 14, AwayScore: 12, Tiebreak: true},
				},
				Final: "2-0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.score)
			if err != nil {
				t.Fatalf("Failed to marshal score: %v", err)
			}

			var decoded Score
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("Failed to unmarshal score: %v", err)
			}

			if decoded.Final != tt.score.Final {
				t.Errorf("Final score mismatch: got %s, want %s", decoded.Final, tt.score.Final)
			}

			if decoded.Retirement != tt.score.Retirement {
				t.Errorf("Retirement flag mismatch: got %v, want %v", decoded.Retirement, tt.score.Retirement)
			}

			if decoded.Walkover != tt.score.Walkover {
				t.Errorf("Walkover flag mismatch: got %v, want %v", decoded.Walkover, tt.score.Walkover)
			}

			if decoded.Disqualify != tt.score.Disqualify {
				t.Errorf("Disqualify flag mismatch: got %v, want %v", decoded.Disqualify, tt.score.Disqualify)
			}
		})
	}
}

func TestVenue_Validation(t *testing.T) {
	venue := Venue{
		Name:     "Olympic Stadium",
		Address:  "1 Olympic Way",
		City:     "Los Angeles",
		State:    "CA",
		Country:  "USA",
		PostCode: "90001",
		Courts:   []string{"Center Court", "Court 1", "Court 2", "Court 3", "Court 4"},
		Capacity: 5000,
	}

	// Test JSON marshaling
	data, err := json.Marshal(venue)
	if err != nil {
		t.Fatalf("Failed to marshal venue: %v", err)
	}

	var decoded Venue
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal venue: %v", err)
	}

	if len(decoded.Courts) != 5 {
		t.Errorf("Expected 5 courts, got %d", len(decoded.Courts))
	}

	if decoded.Capacity != 5000 {
		t.Errorf("Expected capacity 5000, got %d", decoded.Capacity)
	}
}

func TestMoney_Currencies(t *testing.T) {
	currencies := []struct {
		amount   float64
		currency string
	}{
		{50.00, "USD"},
		{45.00, "EUR"},
		{40.00, "GBP"},
		{5000.00, "JPY"},
		{100.00, "CAD"},
		{75.00, "AUD"},
	}

	for _, c := range currencies {
		money := Money{
			Amount:   c.amount,
			Currency: c.currency,
		}

		data, err := json.Marshal(money)
		if err != nil {
			t.Errorf("Failed to marshal money %v: %v", c, err)
			continue
		}

		var decoded Money
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Errorf("Failed to unmarshal money: %v", err)
			continue
		}

		if decoded.Amount != c.amount {
			t.Errorf("Amount mismatch for %s: got %f, want %f", c.currency, decoded.Amount, c.amount)
		}

		if decoded.Currency != c.currency {
			t.Errorf("Currency mismatch: got %s, want %s", decoded.Currency, c.currency)
		}
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}