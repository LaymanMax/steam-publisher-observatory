package metrics

import (
	"fmt"
	"math"
	"sort"
)

// GameOwnershipContribution represents an individual game's contribution to catalog ownership.
type GameOwnershipContribution struct {
	AppID           string  `json:"appid"`
	Name            string  `json:"name"`
	OwnersLower     int64   `json:"owners_lower"`
	OwnersUpper     int64   `json:"owners_upper"`
	OwnersMidpoint  int64   `json:"owners_midpoint"`
	ShareOfMidpoint float64 `json:"share_of_midpoint"` // percentage 0 to 100
	CumulativeShare float64 `json:"cumulative_share"`  // percentage 0 to 100
	Rank            int     `json:"rank"`
}

// PortfolioDependencyReport contains full portfolio dependency analytics.
type PortfolioDependencyReport struct {
	PublisherID            string                      `json:"publisher_id"`
	PublisherName          string                      `json:"publisher_name"`
	CatalogSize            int                         `json:"catalog_size"`
	ValidOwnersCount       int                         `json:"valid_owners_count"`
	OwnershipCoverage      float64                     `json:"ownership_coverage"` // percentage 0 to 100
	CatalogOwnersLower     int64                       `json:"catalog_owners_lower"`
	CatalogOwnersMidpoint  int64                       `json:"catalog_owners_midpoint"`
	CatalogOwnersUpper     int64                       `json:"catalog_owners_upper"`
	Top1Share              float64                     `json:"top1_share"`
	Top3Share              float64                     `json:"top3_share"`
	Top5Share              float64                     `json:"top5_share"`
	Top10Share             float64                     `json:"top10_share"`
	RemainingShare         float64                     `json:"remaining_share"`
	ScaledHHI              float64                     `json:"scaled_hhi"`            // 0 to 10,000
	UnscaledHHI            float64                     `json:"unscaled_hhi"`          // 0.0 to 1.0
	EffectiveTitleCount    float64                     `json:"effective_title_count"` // 1 / unscaled_HHI
	RankedTitles           []GameOwnershipContribution `json:"ranked_titles"`
	HeadlineMetric         string                      `json:"headline_metric"`
	Interpretation         string                      `json:"interpretation"`
	ConcentrationLevel     string                      `json:"concentration_level"` // "High", "Moderate", "Diversified"
	MethodologyLimitations string                      `json:"methodology_limitations"`
}

// GameOwnershipInput represents a game record for dependency calculation.
type GameOwnershipInput struct {
	AppID          string
	Name           string
	OwnersLower    *int64
	OwnersUpper    *int64
	OwnersMidpoint *int64
}

// CalculatePortfolioDependency computes top title shares, HHI, and effective title count.
func CalculatePortfolioDependency(publisherID, publisherName string, games []GameOwnershipInput) *PortfolioDependencyReport {
	report := &PortfolioDependencyReport{
		PublisherID:            publisherID,
		PublisherName:          publisherName,
		CatalogSize:            len(games),
		MethodologyLimitations: "Estimates are aggregated midpoints from public range data and do not represent unique players, sales, profit, or revenue dependency.",
	}

	if len(games) == 0 {
		report.HeadlineMetric = "No games recorded for this publisher."
		report.Interpretation = "The catalog has 0 titles in the dataset."
		report.ConcentrationLevel = "N/A"
		return report
	}

	// Filter valid ownership records
	var validGames []GameOwnershipContribution
	var sumLower, sumMid, sumUpper int64

	for _, g := range games {
		if g.OwnersMidpoint != nil && *g.OwnersMidpoint >= 0 {
			lower := int64(0)
			if g.OwnersLower != nil {
				lower = *g.OwnersLower
			}
			upper := int64(0)
			if g.OwnersUpper != nil {
				upper = *g.OwnersUpper
			}
			mid := *g.OwnersMidpoint

			sumLower += lower
			sumMid += mid
			sumUpper += upper

			validGames = append(validGames, GameOwnershipContribution{
				AppID:          g.AppID,
				Name:           g.Name,
				OwnersLower:    lower,
				OwnersUpper:    upper,
				OwnersMidpoint: mid,
			})
		}
	}

	report.ValidOwnersCount = len(validGames)
	report.OwnershipCoverage = (float64(len(validGames)) / float64(len(games))) * 100.0
	report.CatalogOwnersLower = sumLower
	report.CatalogOwnersMidpoint = sumMid
	report.CatalogOwnersUpper = sumUpper

	if len(validGames) == 0 || sumMid == 0 {
		report.HeadlineMetric = "No ownership estimate data available for catalog calculation."
		report.Interpretation = "Cannot compute concentration because no games in the portfolio have a non-zero ownership estimate."
		report.ConcentrationLevel = "Undefined"
		return report
	}

	// Sort descending by midpoint
	sort.Slice(validGames, func(i, j int) bool {
		if validGames[i].OwnersMidpoint == validGames[j].OwnersMidpoint {
			return validGames[i].Name < validGames[j].Name
		}
		return validGames[i].OwnersMidpoint > validGames[j].OwnersMidpoint
	})

	// Compute shares and cumulative shares
	var cumMid int64
	var unscaledHHI float64

	for i := range validGames {
		validGames[i].Rank = i + 1
		shareFraction := float64(validGames[i].OwnersMidpoint) / float64(sumMid)
		validGames[i].ShareOfMidpoint = shareFraction * 100.0

		cumMid += validGames[i].OwnersMidpoint
		validGames[i].CumulativeShare = (float64(cumMid) / float64(sumMid)) * 100.0

		unscaledHHI += shareFraction * shareFraction
	}

	report.RankedTitles = validGames
	report.UnscaledHHI = unscaledHHI
	report.ScaledHHI = math.Min(10000.0, unscaledHHI*10000.0)
	if unscaledHHI > 0.000001 {
		report.EffectiveTitleCount = 1.0 / unscaledHHI
	}

	// Compute top N shares
	topNShare := func(n int) float64 {
		count := min(n, len(validGames))
		var subSum int64
		for i := 0; i < count; i++ {
			subSum += validGames[i].OwnersMidpoint
		}
		return (float64(subSum) / float64(sumMid)) * 100.0
	}

	report.Top1Share = topNShare(1)
	report.Top3Share = topNShare(3)
	report.Top5Share = topNShare(5)
	report.Top10Share = topNShare(10)
	report.RemainingShare = math.Max(0.0, 100.0-report.Top5Share)

	// Headline & Interpretation
	top5Count := min(5, len(validGames))
	report.HeadlineMetric = fmt.Sprintf(
		"Top %d titles account for %.1f%% of estimated catalog ownership midpoint.",
		top5Count, report.Top5Share,
	)

	if report.ScaledHHI >= 2500.0 || report.Top5Share >= 70.0 {
		report.ConcentrationLevel = "High"
		report.Interpretation = fmt.Sprintf(
			"%d titles account for %.1f%% of the catalog's estimated ownership midpoint. This indicates that portfolio ownership estimates are heavily concentrated in a small subset of the publisher's games (Effective Title Count: %.1f).",
			top5Count, report.Top5Share, report.EffectiveTitleCount,
		)
	} else if report.ScaledHHI >= 1500.0 || report.Top5Share >= 45.0 {
		report.ConcentrationLevel = "Moderate"
		report.Interpretation = fmt.Sprintf(
			"The top %d titles account for %.1f%% of estimated ownership midpoint, reflecting moderate catalog dependency with an effective title count of %.1f.",
			top5Count, report.Top5Share, report.EffectiveTitleCount,
		)
	} else {
		report.ConcentrationLevel = "Diversified"
		report.Interpretation = fmt.Sprintf(
			"Ownership estimates are broadly distributed across the catalog. The top %d titles contribute %.1f%% of estimated ownership midpoint (Effective Title Count: %.1f).",
			top5Count, report.Top5Share, report.EffectiveTitleCount,
		)
	}

	return report
}
