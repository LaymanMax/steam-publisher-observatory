package metrics

import (
	"math"
	"testing"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
)

func int64Ptr(v int64) *int64 {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func intPtr(v int) *int {
	return &v
}

func TestCalculatePortfolioDependency(t *testing.T) {
	// Synthetic fixture where top 5 games have deliberate ownership midpoints:
	// Game 1: 500,000 (50%)
	// Game 2: 100,000 (10%)
	// Game 3: 50,000  (5%)
	// Game 4: 40,000  (4%)
	// Game 5: 30,000  (3%)
	// Subtotal top 5 = 720,000 (72.0%!)
	// Remaining 10 games: 28,000 each = 280,000 (28.0%)
	// Total = 1,000,000.
	games := []GameOwnershipInput{
		{AppID: "1", Name: "Blockbuster 1", OwnersLower: int64Ptr(400000), OwnersUpper: int64Ptr(600000), OwnersMidpoint: int64Ptr(500000)},
		{AppID: "2", Name: "Hit 2", OwnersLower: int64Ptr(80000), OwnersUpper: int64Ptr(120000), OwnersMidpoint: int64Ptr(100000)},
		{AppID: "3", Name: "Hit 3", OwnersLower: int64Ptr(40000), OwnersUpper: int64Ptr(60000), OwnersMidpoint: int64Ptr(50000)},
		{AppID: "4", Name: "Hit 4", OwnersLower: int64Ptr(30000), OwnersUpper: int64Ptr(50000), OwnersMidpoint: int64Ptr(40000)},
		{AppID: "5", Name: "Hit 5", OwnersLower: int64Ptr(20000), OwnersUpper: int64Ptr(40000), OwnersMidpoint: int64Ptr(30000)},
	}
	for i := 6; i <= 15; i++ {
		games = append(games, GameOwnershipInput{
			AppID:          string(rune('0' + i)),
			Name:           "Tail Game",
			OwnersLower:    int64Ptr(20000),
			OwnersUpper:    int64Ptr(36000),
			OwnersMidpoint: int64Ptr(28000),
		})
	}

	report := CalculatePortfolioDependency("pub_1", "Test Publisher", games)

	if report.CatalogOwnersMidpoint != 1000000 {
		t.Fatalf("expected total midpoint 1,000,000, got %d", report.CatalogOwnersMidpoint)
	}

	// Top 5 share MUST be exactly 72.0%
	if math.Abs(report.Top5Share-72.0) > 0.01 {
		t.Errorf("Top5Share: expected 72.0%%, got %.2f%%", report.Top5Share)
	}
	if math.Abs(report.RemainingShare-28.0) > 0.01 {
		t.Errorf("RemainingShare: expected 28.0%%, got %.2f%%", report.RemainingShare)
	}
	if math.Abs(report.Top1Share-50.0) > 0.01 {
		t.Errorf("Top1Share: expected 50.0%%, got %.2f%%", report.Top1Share)
	}
	if math.Abs(report.Top3Share-65.0) > 0.01 {
		t.Errorf("Top3Share: expected 65.0%%, got %.2f%%", report.Top3Share)
	}

	// Check HHI: 0.5^2 + 0.1^2 + 0.05^2 + 0.04^2 + 0.03^2 + 10*(0.028^2)
	// = 0.25 + 0.01 + 0.0025 + 0.0016 + 0.0009 + 10*(0.000784) = 0.265 + 0.00784 = 0.27284
	// Scaled HHI ~ 2728.4
	if report.ScaledHHI < 2700 || report.ScaledHHI > 2750 {
		t.Errorf("ScaledHHI expected ~2728, got %.1f", report.ScaledHHI)
	}
	if report.EffectiveTitleCount < 3.5 || report.EffectiveTitleCount > 4.0 {
		t.Errorf("EffectiveTitleCount expected ~3.66, got %.2f", report.EffectiveTitleCount)
	}
	if report.ConcentrationLevel != "High" {
		t.Errorf("expected ConcentrationLevel High, got %s", report.ConcentrationLevel)
	}
}

func TestCalculatePortfolioDependency_ZeroDenominator(t *testing.T) {
	games := []GameOwnershipInput{
		{AppID: "1", Name: "Zero Game", OwnersMidpoint: int64Ptr(0)},
		{AppID: "2", Name: "Missing Game", OwnersMidpoint: nil},
	}
	report := CalculatePortfolioDependency("pub_2", "Zero Publisher", games)

	if report.Top5Share != 0.0 {
		t.Errorf("expected Top5Share 0.0 for zero denominator, got %f", report.Top5Share)
	}
	if report.ConcentrationLevel != "Undefined" {
		t.Errorf("expected ConcentrationLevel Undefined, got %s", report.ConcentrationLevel)
	}
}

func TestCalculateGenreMetrics_WeightedSum(t *testing.T) {
	// Game 1 has 1 genre: Action (weight 1.0)
	// Game 2 has 2 genres: Action, RPG (weight 0.5 each)
	// Game 3 has 3 genres: RPG, Strategy, Indie (weight 0.333 each)
	// Total games with genres = 3
	// Action weight = 1.0 + 0.5 = 1.5 (50.0%)
	// RPG weight = 0.5 + 0.3333 = 0.8333 (27.78%)
	// Strategy = 0.3333 (11.11%)
	// Indie = 0.3333 (11.11%)
	// Total weighted share sum MUST be exactly 100.0%!
	games := []GameGenreInput{
		{AppID: "1", Genres: []string{"Action"}, ReleaseYear: intPtr(2022), Midpoint: 10000},
		{AppID: "2", Genres: []string{"Action", "RPG"}, ReleaseYear: intPtr(2023), Midpoint: 20000},
		{AppID: "3", Genres: []string{"RPG", "Strategy", "Indie"}, ReleaseYear: intPtr(2024), Midpoint: 30000},
		{AppID: "4", Genres: nil, ReleaseYear: intPtr(2024), Midpoint: 0}, // Missing genre
	}

	report := CalculateGenreMetrics("pub_3", 4, games)

	if report.GamesWithGenres != 3 {
		t.Errorf("expected 3 games with genres, got %d", report.GamesWithGenres)
	}
	if math.Abs(report.GenreCoverage-75.0) > 0.01 {
		t.Errorf("expected GenreCoverage 75.0%%, got %.2f%%", report.GenreCoverage)
	}
	if report.DominantGenre != "Action" {
		t.Errorf("expected DominantGenre Action, got %s", report.DominantGenre)
	}

	var sumShares float64
	for _, g := range report.Genres {
		sumShares += g.WeightedShare
	}
	if math.Abs(sumShares-100.0) > 0.001 {
		t.Errorf("sum of weighted shares MUST equal 100.0%%, got %f", sumShares)
	}
}

func TestCalculatePricingMetrics(t *testing.T) {
	tiers := config.DefaultPriceTiers
	games := []GamePriceInput{
		{AppID: "1", Price: float64Ptr(0.0)},   // Free
		{AppID: "2", Price: float64Ptr(4.99)},  // Very Low
		{AppID: "3", Price: float64Ptr(9.99)},  // Low
		{AppID: "4", Price: float64Ptr(19.99)}, // Mid-range
		{AppID: "5", Price: float64Ptr(29.99)}, // Premium
		{AppID: "6", Price: float64Ptr(59.99)}, // High
		{AppID: "7", Price: nil},               // Missing
	}

	report := CalculatePricingMetrics("pub_4", games, tiers)

	if report.FreeCount != 1 {
		t.Errorf("expected 1 free title, got %d", report.FreeCount)
	}
	if report.PaidCount != 5 {
		t.Errorf("expected 5 paid titles, got %d", report.PaidCount)
	}
	if report.MissingCount != 1 {
		t.Errorf("expected 1 missing price title, got %d", report.MissingCount)
	}
	if report.MinPrice != 0.0 || report.MaxPrice != 59.99 {
		t.Errorf("Min/Max price error: min=%.2f, max=%.2f", report.MinPrice, report.MaxPrice)
	}
}

func TestCalculateReviewMetrics(t *testing.T) {
	games := []GameEngagementInput{
		{AppID: "1", Positive: 80, Negative: 20, Recommendations: 50, PeakCCU: 100, MedPlaytimeForever: 120},
		{AppID: "2", Positive: 10, Negative: 90, Recommendations: 5, PeakCCU: 20, MedPlaytimeForever: 30},
	}

	report := CalculateReviewMetrics(games)

	if report.TotalPositive != 90 || report.TotalNegative != 110 || report.TotalReviews != 200 {
		t.Errorf("reviews mismatch: pos=%d, neg=%d, tot=%d", report.TotalPositive, report.TotalNegative, report.TotalReviews)
	}
	// 90 / 200 = 45%
	if math.Abs(report.PositiveShare-45.0) > 0.01 {
		t.Errorf("PositiveShare expected 45.0%%, got %.2f%%", report.PositiveShare)
	}
	if report.MaxPeakCCU != 100 {
		t.Errorf("MaxPeakCCU expected 100, got %d", report.MaxPeakCCU)
	}
}
