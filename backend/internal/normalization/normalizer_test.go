package normalization

import (
	"testing"
)

func TestCleanName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"whitespace trim", "  Valve  ", "Valve"},
		{"repeated spaces", "Valve   Corporation", "Valve Corporation"},
		{"tabs and newlines", "Valve\t\nSoftware", "Valve Software"},
		{"unicode normalization NFKC", "Ｆｉｒｅｆｌｙ　Ｓｔｕｄｉｏｓ", "Firefly Studios"},
		{"invisible zero width space", "Valve\u200b\u200cSoftware", "ValveSoftware"},
		{"empty string", "", ""},
		{"spaces only", "    ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CleanName(tt.input)
			if got != tt.expected {
				t.Errorf("CleanName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestCanonicalKey(t *testing.T) {
	tests := []struct {
		input1   string
		input2   string
		shouldEq bool
	}{
		{"Example Studios", "example studios", true},
		{" Example Studios ", "EXAMPLE STUDIOS", true},
		{"Square Enix", "square enix", true},
		{"Valve", "Valve Corporation", false}, // Must NOT auto-merge different corporate names!
		{"Valve", "Valve Software", false},
	}

	for _, tt := range tests {
		k1 := CanonicalKey(tt.input1)
		k2 := CanonicalKey(tt.input2)
		if (k1 == k2) != tt.shouldEq {
			t.Errorf("CanonicalKey(%q) vs CanonicalKey(%q): got %v, want %v (k1=%q, k2=%q)",
				tt.input1, tt.input2, k1 == k2, tt.shouldEq, k1, k2)
		}
	}
}

func TestGenerateEntityID(t *testing.T) {
	id1 := GenerateEntityID("publisher", "Paradox Interactive")
	id2 := GenerateEntityID("publisher", " paradox interactive ")
	id3 := GenerateEntityID("developer", "Paradox Interactive")

	if id1 == "" {
		t.Fatalf("expected non-empty entity ID")
	}
	if id1 != id2 {
		t.Errorf("expected id1 == id2 for case/whitespace difference, got %s vs %s", id1, id2)
	}
	if id1 == id3 {
		t.Errorf("expected publisher and developer to have different IDs even with same name")
	}
}

func TestIsPotentialAlias(t *testing.T) {
	isAlias, score := IsPotentialAlias("Valve", "Valve Corporation")
	if !isAlias || score < 0.3 {
		t.Errorf("expected Valve and Valve Corporation to be flagged as potential alias, got %v (score %f)", isAlias, score)
	}

	isAlias2, _ := IsPotentialAlias("Ubisoft", "Electronic Arts")
	if isAlias2 {
		t.Errorf("expected Ubisoft and Electronic Arts to NOT be potential aliases")
	}
}
