package normalization

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

var (
	// Regex for multiple whitespace sequences (including tabs, newlines)
	whitespaceRegex = regexp.MustCompile(`\s+`)
	// Regex for invisible characters (zero-width spaces, BOM, control characters excluding regular whitespace)
	invisibleCharsRegex = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F\x{200B}-\x{200D}\x{FEFF}]`)
)

// CleanName cleans and normalizes a display name while preserving human casing.
func CleanName(name string) string {
	// 1. Unicode NFKC normalization
	cleaned := norm.NFKC.String(name)

	// 2. Normalize whitespace first: tabs, newlines, multiple spaces -> single space
	cleaned = whitespaceRegex.ReplaceAllString(cleaned, " ")

	// 3. Remove invisible and non-whitespace control characters
	cleaned = invisibleCharsRegex.ReplaceAllString(cleaned, "")

	// 4. Trim leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	return cleaned
}

// CanonicalKey generates a deterministic normalized matching key for an entity name.
// Used for exact normalized matching (case-insensitive, punctuation/whitespace normalized).
func CanonicalKey(name string) string {
	cleaned := CleanName(name)
	if cleaned == "" {
		return ""
	}

	// Lowercase for case-insensitivity
	return strings.ToLower(cleaned)
}

// GenerateEntityID creates a deterministic ID for an entity based on its canonical key.
func GenerateEntityID(entityType, name string) string {
	key := CanonicalKey(name)
	if key == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(entityType + ":" + key))
	return hex.EncodeToString(hash[:8]) // 16-hex char deterministic ID
}

// IsPotentialAlias checks if two distinct canonical names might be candidates for manual aliasing.
// Returns true and similarity score (0.0 to 1.0) if they are close.
func IsPotentialAlias(nameA, nameB string) (bool, float64) {
	keyA := CanonicalKey(nameA)
	keyB := CanonicalKey(nameB)
	if keyA == "" || keyB == "" || keyA == keyB {
		return false, 0
	}

	// Tokenize words
	wordsA := strings.FieldsFunc(keyA, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	wordsB := strings.FieldsFunc(keyB, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(wordsA) == 0 || len(wordsB) == 0 {
		return false, 0
	}

	setA := make(map[string]struct{}, len(wordsA))
	for _, w := range wordsA {
		setA[w] = struct{}{}
	}
	intersection := 0
	for _, w := range wordsB {
		if _, ok := setA[w]; ok {
			intersection++
		}
	}
	union := len(setA)
	for _, w := range wordsB {
		if _, ok := setA[w]; !ok {
			union++
		}
	}

	jaccard := float64(intersection) / float64(union)

	// If one is a complete subset of the other's words (e.g. "Valve" in "Valve Corporation")
	if intersection == len(wordsA) || intersection == len(wordsB) {
		score := float64(intersection) / float64(max(len(wordsA), len(wordsB)))
		return true, score
	}

	if jaccard >= 0.5 {
		return true, jaccard
	}

	return false, jaccard
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
