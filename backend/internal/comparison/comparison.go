package comparison

import (
	"context"
	"fmt"
	"math"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/metrics"
)

// PublisherComparisonResult holds the multi-publisher comparison data.
type PublisherComparisonResult struct {
	Publishers   []PublisherComparisonProfile `json:"publishers"`
	MetricRanges map[string]MetricRange       `json:"metric_ranges"`
	Warnings     []string                     `json:"warnings"`
	Timestamp    string                       `json:"timestamp"`
}

// PublisherComparisonProfile holds metrics for one compared publisher.
type PublisherComparisonProfile struct {
	PublisherID           string                  `json:"publisher_id"`
	Name                  string                  `json:"name"`
	CatalogSize           int64                   `json:"catalog_size"`
	DeveloperCount        int64                   `json:"developer_count"`
	GenreCount            int64                   `json:"genre_count"`
	FreeTitleShare        float64                 `json:"free_title_share"`
	MedianPrice           float64                 `json:"median_price"`
	OwnershipCoverage     float64                 `json:"ownership_coverage"`
	CatalogOwnersMidpoint int64                   `json:"catalog_owners_midpoint"`
	Top5Share             float64                 `json:"top5_share"`
	Top10Share            float64                 `json:"top10_share"`
	HHI                   float64                 `json:"hhi"`
	EffectiveTitleCount   float64                 `json:"effective_title_count"`
	TotalReviews          int64                   `json:"total_reviews"`
	PositiveShare         float64                 `json:"positive_share"`
	MedianPlaytime        int64                   `json:"median_playtime"`
	AvgPeakCCU            float64                 `json:"avg_peak_ccu"`
	MaxPeakCCU            int64                   `json:"max_peak_ccu"`
	Genres                []metrics.GenreItem     `json:"genres"`
	PricingTiers          []metrics.PriceTierItem `json:"pricing_tiers"`
	TopTitles             []struct {
		Name     string  `json:"name"`
		Midpoint int64   `json:"midpoint"`
		Share    float64 `json:"share"`
	} `json:"top_titles"`
}

// MetricRange provides min/max bounds for consistent visualization scaling across compared entities.
type MetricRange struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// Compare computes multi-publisher comparative metrics for up to 4 publishers.
func Compare(ctx context.Context, svc *analytics.Service, publisherIDs []string) (*PublisherComparisonResult, error) {
	if len(publisherIDs) == 0 {
		return nil, fmt.Errorf("at least one publisher ID must be specified for comparison")
	}
	if len(publisherIDs) > 4 {
		publisherIDs = publisherIDs[:4]
	}

	result := &PublisherComparisonResult{
		MetricRanges: make(map[string]MetricRange),
	}

	var coverages []float64

	for _, pid := range publisherIDs {
		sum, err := svc.GetPublisherSummary(ctx, pid)
		if err != nil {
			continue
		}

		profile := PublisherComparisonProfile{
			PublisherID:           sum.PublisherID,
			Name:                  sum.Name,
			CatalogSize:           sum.CatalogSize,
			DeveloperCount:        sum.DeveloperCount,
			GenreCount:            sum.GenreCount,
			FreeTitleShare:        sum.FreeTitleShare,
			MedianPrice:           sum.MedianPrice,
			OwnershipCoverage:     sum.OwnershipCoverage,
			CatalogOwnersMidpoint: sum.CatalogOwnersMidpoint,
			Top5Share:             sum.Top5Share,
			Top10Share:            sum.Top10Share,
			HHI:                   sum.HHI,
			EffectiveTitleCount:   sum.EffectiveTitleCount,
			TotalReviews:          sum.TotalReviews,
			PositiveShare:         sum.AvgPositiveShare,
			MedianPlaytime:        sum.MedianPlaytime,
			AvgPeakCCU:            sum.AvgPeakCCU,
			MaxPeakCCU:            sum.MaxPeakCCU,
		}

		coverages = append(coverages, sum.OwnershipCoverage)

		// Fetch genres
		if gRep, err := svc.GetPublisherGenres(ctx, pid); err == nil {
			profile.Genres = gRep.Genres
		}

		// Fetch prices
		if pRep, err := svc.GetPublisherPrices(ctx, pid); err == nil {
			profile.PricingTiers = pRep.Tiers
		}

		// Fetch top titles from dependency
		if dep, err := svc.GetPublisherDependency(ctx, pid); err == nil {
			topN := 5
			if len(dep.RankedTitles) < topN {
				topN = len(dep.RankedTitles)
			}
			for i := 0; i < topN; i++ {
				t := dep.RankedTitles[i]
				profile.TopTitles = append(profile.TopTitles, struct {
					Name     string  `json:"name"`
					Midpoint int64   `json:"midpoint"`
					Share    float64 `json:"share"`
				}{
					Name:     t.Name,
					Midpoint: t.OwnersMidpoint,
					Share:    t.ShareOfMidpoint,
				})
			}
		}

		result.Publishers = append(result.Publishers, profile)
	}

	if len(result.Publishers) == 0 {
		return nil, fmt.Errorf("none of the specified publishers were found")
	}

	// Compute metric ranges for chart scale consistency
	updateRange := func(name string, val float64) {
		r, ok := result.MetricRanges[name]
		if !ok {
			result.MetricRanges[name] = MetricRange{Min: val, Max: val}
			return
		}
		if val < r.Min {
			r.Min = val
		}
		if val > r.Max {
			r.Max = val
		}
		result.MetricRanges[name] = r
	}

	for _, p := range result.Publishers {
		updateRange("catalog_size", float64(p.CatalogSize))
		updateRange("owners_midpoint", float64(p.CatalogOwnersMidpoint))
		updateRange("top5_share", p.Top5Share)
		updateRange("hhi", p.HHI)
		updateRange("median_price", p.MedianPrice)
		updateRange("total_reviews", float64(p.TotalReviews))
		updateRange("developer_count", float64(p.DeveloperCount))
	}

	// Check warnings: differing ownership coverage
	if len(coverages) > 1 {
		minCov, maxCov := coverages[0], coverages[0]
		for _, c := range coverages {
			if c < minCov {
				minCov = c
			}
			if c > maxCov {
				maxCov = c
			}
		}
		if math.Abs(maxCov-minCov) > 20.0 {
			result.Warnings = append(result.Warnings, fmt.Sprintf(
				"Ownership coverage varies significantly between compared publishers (%.1f%% vs %.1f%%). Midpoint comparisons may be biased toward portfolios with higher metadata completeness.",
				minCov, maxCov,
			))
		}
	}

	return result, nil
}
