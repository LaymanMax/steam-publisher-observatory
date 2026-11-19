package importer

import (
	"testing"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
)

func TestParseEstimatedOwners(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLower int64
		wantUpper int64
		wantMid   int64
		wantErr   bool
	}{
		{
			name:      "standard range",
			input:     "0 - 20000",
			wantLower: 0,
			wantUpper: 20000,
			wantMid:   10000,
		},
		{
			name:      "comma-separated numbers",
			input:     "20,000 - 50,000",
			wantLower: 20000,
			wantUpper: 50000,
			wantMid:   35000,
		},
		{
			name:      "en-dash character",
			input:     "50,000 – 100,000",
			wantLower: 50000,
			wantUpper: 100000,
			wantMid:   75000,
		},
		{
			name:      "em-dash character",
			input:     "100,000 — 200,000",
			wantLower: 100000,
			wantUpper: 200000,
			wantMid:   150000,
		},
		{
			name:      "unicode minus character",
			input:     "200,000 − 500,000",
			wantLower: 200000,
			wantUpper: 500000,
			wantMid:   350000,
		},
		{
			name:      "zero to zero",
			input:     "0 - 0",
			wantLower: 0,
			wantUpper: 0,
			wantMid:   0,
		},
		{
			name:      "extra surrounding whitespace",
			input:     "   1,000,000  -   2,000,000   ",
			wantLower: 1000000,
			wantUpper: 2000000,
			wantMid:   1500000,
		},
		{
			name:      "reversed range gracefully swapped",
			input:     "50,000 - 20,000",
			wantLower: 20000,
			wantUpper: 50000,
			wantMid:   35000,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "null string",
			input:   "null",
			wantErr: true,
		},
		{
			name:    "invalid format text",
			input:   "unknown owners",
			wantErr: true,
		},
		{
			name:    "single number without dash",
			input:   "50000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEstimatedOwners(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseEstimatedOwners(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if err == nil {
				if got.Lower != tt.wantLower {
					t.Errorf("Lower: got %d, want %d", got.Lower, tt.wantLower)
				}
				if got.Upper != tt.wantUpper {
					t.Errorf("Upper: got %d, want %d", got.Upper, tt.wantUpper)
				}
				if got.Midpoint != tt.wantMid {
					t.Errorf("Midpoint: got %d, want %d", got.Midpoint, tt.wantMid)
				}
			}
		})
	}
}

func TestParseReleaseDate(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantValid bool
	}{
		{"US style standard", "Oct 23, 2025", 2025, true},
		{"Day first", "23 Oct, 2025", 2025, true},
		{"ISO date", "2024-11-05", 2024, true},
		{"Month year", "Dec 2023", 2023, true},
		{"Bare year", "2022", 2022, true},
		{"Empty string", "", 0, false},
		{"Invalid textual garbage", "To be announced", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReleaseDate(tt.input)
			if got.IsValid != tt.wantValid {
				t.Errorf("ParseReleaseDate(%q) IsValid = %v, want %v", tt.input, got.IsValid, tt.wantValid)
			}
			if tt.wantValid && (got.Year == nil || *got.Year != tt.wantYear) {
				t.Errorf("ParseReleaseDate(%q) Year = %v, want %d", tt.input, got.Year, tt.wantYear)
			}
		})
	}
}

func TestClassifyPriceTier(t *testing.T) {
	tiers := config.DefaultPriceTiers

	tests := []struct {
		price    *float64
		expected string
	}{
		{nil, "Missing"},
		{floatPtr(0.0), "Free"},
		{floatPtr(1.99), "Very Low"},
		{floatPtr(7.99), "Low"},
		{floatPtr(14.99), "Mid-range"},
		{floatPtr(29.99), "Premium"},
		{floatPtr(59.99), "High-priced"},
	}

	for _, tt := range tests {
		got := ClassifyPriceTier(tt.price, tiers)
		if got != tt.expected {
			t.Errorf("ClassifyPriceTier(%v) = %q, want %q", tt.price, got, tt.expected)
		}
	}
}

func TestAssessDataQuality(t *testing.T) {
	// Full complete game
	p := 19.99
	score, isPlaytest, isSparse, flags := AssessDataQuality(
		"Civilization VII",
		[]string{"2K"},
		[]string{"Firaxis Games"},
		[]string{"Strategy"},
		true,
		true,
		&p,
		1000, 200, 1200,
	)

	if score != 100.0 {
		t.Errorf("expected score 100.0, got %f", score)
	}
	if isPlaytest {
		t.Errorf("expected isPlaytest = false")
	}
	if isSparse {
		t.Errorf("expected isSparse = false")
	}
	if len(flags) != 0 {
		t.Errorf("expected 0 flags, got %v", flags)
	}

	// Playtest / Sparse record (like Black Dragon Mage Playtest)
	score2, isPlaytest2, isSparse2, flags2 := AssessDataQuality(
		"Black Dragon Mage Playtest",
		nil,
		nil,
		nil,
		false,
		false,
		nil,
		0, 0, 0,
	)

	if !isPlaytest2 {
		t.Errorf("expected isPlaytest = true for 'Black Dragon Mage Playtest'")
	}
	if !isSparse2 {
		t.Errorf("expected isSparse = true for empty record")
	}
	if score2 > 20.0 {
		t.Errorf("expected low completeness score for sparse record, got %f", score2)
	}
	if len(flags2) < 5 {
		t.Errorf("expected multiple quality flags, got %v", flags2)
	}
}

func floatPtr(v float64) *float64 {
	return &v
}
