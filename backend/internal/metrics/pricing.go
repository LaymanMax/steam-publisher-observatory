package metrics

import (
	"sort"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
)

// PriceTierItem represents one pricing bucket.
type PriceTierItem struct {
	TierName  string  `json:"tier_name"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	GameCount int     `json:"game_count"`
	Share     float64 `json:"share"` // percentage 0 to 100
}

// PricingReport holds complete portfolio pricing analytics.
type PricingReport struct {
	PublisherID  string          `json:"publisher_id"`
	TotalGames   int             `json:"total_games"`
	FreeCount    int             `json:"free_count"`
	FreeShare    float64         `json:"free_share"`
	PaidCount    int             `json:"paid_count"`
	PaidShare    float64         `json:"paid_share"`
	MissingCount int             `json:"missing_count"`
	MissingShare float64         `json:"missing_share"`
	MinPrice     float64         `json:"min_price"`
	MaxPrice     float64         `json:"max_price"`
	AvgPrice     float64         `json:"avg_price"`
	MedianPrice  float64         `json:"median_price"`
	Tiers        []PriceTierItem `json:"tiers"`
	Currency     string          `json:"currency"`
}

// GamePriceInput represents a game's price for pricing analysis.
type GamePriceInput struct {
	AppID string
	Price *float64
}

// CalculatePricingMetrics computes price distribution, tiers, and central tendencies.
func CalculatePricingMetrics(publisherID string, games []GamePriceInput, tiers []config.PriceTier) *PricingReport {
	report := &PricingReport{
		PublisherID: publisherID,
		TotalGames:  len(games),
		Currency:    "USD",
	}

	if len(games) == 0 {
		return report
	}

	tierCounts := make(map[string]int)
	var validPrices []float64

	for _, g := range games {
		if g.Price == nil {
			report.MissingCount++
			tierCounts["Missing"]++
			continue
		}

		p := *g.Price
		if p < 0 {
			report.MissingCount++
			tierCounts["Missing"]++
			continue
		}

		validPrices = append(validPrices, p)
		if p == 0.0 {
			report.FreeCount++
			tierCounts["Free"]++
		} else {
			report.PaidCount++
			// Determine tier
			matched := false
			for _, t := range tiers {
				if t.Name != "Free" && p >= t.Min && p <= t.Max {
					tierCounts[t.Name]++
					matched = true
					break
				}
			}
			if !matched {
				tierCounts["Other"]++
			}
		}
	}

	total := float64(len(games))
	report.FreeShare = (float64(report.FreeCount) / total) * 100.0
	report.PaidShare = (float64(report.PaidCount) / total) * 100.0
	report.MissingShare = (float64(report.MissingCount) / total) * 100.0

	if len(validPrices) > 0 {
		sort.Float64s(validPrices)
		report.MinPrice = validPrices[0]
		report.MaxPrice = validPrices[len(validPrices)-1]

		var sum float64
		for _, p := range validPrices {
			sum += p
		}
		report.AvgPrice = sum / float64(len(validPrices))

		// Median
		n := len(validPrices)
		if n%2 == 1 {
			report.MedianPrice = validPrices[n/2]
		} else {
			report.MedianPrice = (validPrices[n/2-1] + validPrices[n/2]) / 2.0
		}
	}

	// Populate tier items in standard order
	for _, t := range tiers {
		c := tierCounts[t.Name]
		report.Tiers = append(report.Tiers, PriceTierItem{
			TierName:  t.Name,
			MinPrice:  t.Min,
			MaxPrice:  t.Max,
			GameCount: c,
			Share:     (float64(c) / total) * 100.0,
		})
	}
	if report.MissingCount > 0 {
		report.Tiers = append(report.Tiers, PriceTierItem{
			TierName:  "Missing",
			MinPrice:  -1,
			MaxPrice:  -1,
			GameCount: report.MissingCount,
			Share:     report.MissingShare,
		})
	}

	return report
}
