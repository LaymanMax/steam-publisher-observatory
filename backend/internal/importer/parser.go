package importer

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
)

// OwnershipEstimate holds parsed lower bound, upper bound, and midpoint.
type OwnershipEstimate struct {
	Lower    int64
	Upper    int64
	Midpoint int64
	IsValid  bool
}

// Regex to match two numbers separated by dash/hyphen with optional commas and spaces.
var (
	// Matches dash characters: -, –, —, −
	dashReplacer = strings.NewReplacer(
		"–", "-",
		"—", "-",
		"−", "-",
	)
	cleanNumRegex = regexp.MustCompile(`[^\d-]`)
)

// ParseEstimatedOwners parses ranges like "0 - 20000", "20,000 - 50,000", "0 - 0", etc.
func ParseEstimatedOwners(raw string) (*OwnershipEstimate, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, "null") || strings.EqualFold(trimmed, "none") {
		return nil, fmt.Errorf("empty or null ownership estimate")
	}

	// Normalize dashes
	normalized := dashReplacer.Replace(trimmed)

	// Split by "-"
	parts := strings.Split(normalized, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid ownership format %q: expected 2 bounds separated by dash", raw)
	}

	// Clean commas and spaces
	cleanLower := cleanNumRegex.ReplaceAllString(parts[0], "")
	cleanUpper := cleanNumRegex.ReplaceAllString(parts[1], "")

	if cleanLower == "" || cleanUpper == "" {
		return nil, fmt.Errorf("invalid ownership numbers in %q", raw)
	}

	lower, err := strconv.ParseInt(cleanLower, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lower bound %q: %w", cleanLower, err)
	}

	upper, err := strconv.ParseInt(cleanUpper, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse upper bound %q: %w", cleanUpper, err)
	}

	if lower < 0 || upper < 0 {
		return nil, fmt.Errorf("negative bounds not permitted: %d - %d", lower, upper)
	}

	// If reversed (e.g. 50000 - 20000), swap or reject. Swap gracefully.
	if lower > upper {
		lower, upper = upper, lower
	}

	midpoint := (lower + upper) / 2

	return &OwnershipEstimate{
		Lower:    lower,
		Upper:    upper,
		Midpoint: midpoint,
		IsValid:  true,
	}, nil
}

// ParsedReleaseDate holds parsed release date info.
type ParsedReleaseDate struct {
	Formatted string // YYYY-MM-DD or YYYY-MM
	Year      *int
	IsValid   bool
}

var dateFormats = []string{
	"Jan 2, 2006",
	"2 Jan, 2006",
	"Jan 2006",
	"2006-01-02",
	"2006/01/02",
	"2006-01",
	"January 2, 2006",
	"2 January, 2006",
	"January 2006",
	"02 Jan 2006",
	"2006",
}

// ParseReleaseDate attempts to parse various Steam release date formats.
func ParseReleaseDate(raw string) ParsedReleaseDate {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || strings.EqualFold(trimmed, "null") || strings.EqualFold(trimmed, "none") {
		return ParsedReleaseDate{IsValid: false}
	}

	// Try standard date formats
	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, trimmed); err == nil {
			year := t.Year()
			// Basic sanity check on year (Steam games range roughly 1980 to 2035)
			if year >= 1980 && year <= 2035 {
				formatted := t.Format("2006-01-02")
				return ParsedReleaseDate{
					Formatted: formatted,
					Year:      &year,
					IsValid:   true,
				}
			}
		}
	}

	// Fallback regex to extract a 4-digit year between 1980 and 2035
	yearRegex := regexp.MustCompile(`\b(19[89]\d|20[0-3]\d)\b`)
	if match := yearRegex.FindString(trimmed); match != "" {
		if y, err := strconv.Atoi(match); err == nil {
			return ParsedReleaseDate{
				Formatted: fmt.Sprintf("%04d", y),
				Year:      &y,
				IsValid:   true,
			}
		}
	}

	return ParsedReleaseDate{IsValid: false}
}

// ClassifyPriceTier returns the pricing tier name for a given price.
func ClassifyPriceTier(price *float64, tiers []config.PriceTier) string {
	if price == nil {
		return "Missing"
	}
	val := *price
	if val < 0 {
		return "Missing"
	}

	for _, tier := range tiers {
		if tier.Name == "Free" && val == 0.0 {
			return "Free"
		}
		if val >= tier.Min && val <= tier.Max {
			return tier.Name
		}
	}

	return "Other"
}

// AssessDataQuality computes completeness score and flags for a game record.
func AssessDataQuality(
	name string,
	publishers []string,
	developers []string,
	genres []string,
	hasValidOwners bool,
	hasValidDate bool,
	price *float64,
	positive, negative, reviews int64,
) (score float64, isPlaytest bool, isSparse bool, flags []string) {
	// Flags tracking
	if strings.TrimSpace(name) == "" {
		flags = append(flags, "missing_name")
	} else {
		score += 10.0
	}

	lowerName := strings.ToLower(name)
	if strings.Contains(lowerName, "playtest") ||
		strings.Contains(lowerName, "beta test") ||
		strings.Contains(lowerName, "demo / playtest") {
		isPlaytest = true
		flags = append(flags, "suspected_playtest")
	}

	if len(publishers) == 0 {
		flags = append(flags, "missing_publisher")
	} else {
		score += 20.0
	}

	if len(developers) == 0 {
		flags = append(flags, "missing_developer")
	} else {
		score += 15.0
	}

	if len(genres) == 0 {
		flags = append(flags, "missing_genre")
	} else {
		score += 15.0
	}

	if !hasValidOwners {
		flags = append(flags, "missing_owners")
	} else {
		score += 15.0
	}

	if !hasValidDate {
		flags = append(flags, "missing_release_date")
	} else {
		score += 10.0
	}

	if price == nil {
		flags = append(flags, "missing_price")
	} else {
		score += 5.0
	}

	if positive+negative > 0 || reviews > 0 {
		score += 10.0
	} else {
		flags = append(flags, "no_reviews")
	}

	// Sparse check: lacks publishers AND developers AND has no reviews
	if len(publishers) == 0 && len(developers) == 0 && (positive+negative == 0) {
		isSparse = true
		flags = append(flags, "sparse_metadata")
	}

	return score, isPlaytest, isSparse, flags
}
