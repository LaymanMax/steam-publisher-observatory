package metrics

import (
	"sort"
)

// ReviewStats holds aggregated review and engagement metrics.
type ReviewStats struct {
	TotalGames            int     `json:"total_games"`
	GamesWithReviews      int     `json:"games_with_reviews"`
	ReviewCoverage        float64 `json:"review_coverage"` // percentage 0 to 100
	TotalPositive         int64   `json:"total_positive"`
	TotalNegative         int64   `json:"total_negative"`
	TotalReviews          int64   `json:"total_reviews"`
	PositiveShare         float64 `json:"positive_share"` // percentage 0 to 100
	AvgReviewsPerGame     float64 `json:"avg_reviews_per_game"`
	MedianReviewsPerGame  float64 `json:"median_reviews_per_game"`
	TotalRecommendations  int64   `json:"total_recommendations"`
	MaxPeakCCU            int64   `json:"max_peak_ccu"`
	AvgPeakCCU            float64 `json:"avg_peak_ccu"`
	MedianPlaytimeForever int64   `json:"median_playtime_forever"`
	AvgPlaytimeForever    float64 `json:"avg_playtime_forever"`
	MedianPlaytime2Weeks  int64   `json:"median_playtime_2weeks"`
	AvgPlaytime2Weeks     float64 `json:"avg_playtime_2weeks"`
}

// GameEngagementInput contains review and engagement attributes for a game.
type GameEngagementInput struct {
	AppID              string
	Name               string
	Positive           int64
	Negative           int64
	Recommendations    int64
	PeakCCU            int64
	MedPlaytimeForever int64
	AvgPlaytimeForever int64
	MedPlaytime2Weeks  int64
	AvgPlaytime2Weeks  int64
}

// CalculateReviewMetrics aggregates reviews and engagement across a publisher catalog.
func CalculateReviewMetrics(games []GameEngagementInput) *ReviewStats {
	stats := &ReviewStats{
		TotalGames: len(games),
	}

	if len(games) == 0 {
		return stats
	}

	var (
		reviewCounts []int64
		playtimes    []int64
		playtimes2w  []int64
		ccuList      []int64
		sumPlaytime  int64
		sumPlay2w    int64
		sumCCU       int64
	)

	for _, g := range games {
		tot := g.Positive + g.Negative
		if tot > 0 {
			stats.GamesWithReviews++
			reviewCounts = append(reviewCounts, tot)
		}

		stats.TotalPositive += g.Positive
		stats.TotalNegative += g.Negative
		stats.TotalReviews += tot
		stats.TotalRecommendations += g.Recommendations

		if g.PeakCCU > stats.MaxPeakCCU {
			stats.MaxPeakCCU = g.PeakCCU
		}
		ccuList = append(ccuList, g.PeakCCU)
		sumCCU += g.PeakCCU

		if g.MedPlaytimeForever > 0 {
			playtimes = append(playtimes, g.MedPlaytimeForever)
		}
		sumPlaytime += g.AvgPlaytimeForever

		if g.MedPlaytime2Weeks > 0 {
			playtimes2w = append(playtimes2w, g.MedPlaytime2Weeks)
		}
		sumPlay2w += g.AvgPlaytime2Weeks
	}

	stats.ReviewCoverage = (float64(stats.GamesWithReviews) / float64(len(games))) * 100.0

	if stats.TotalReviews > 0 {
		stats.PositiveShare = (float64(stats.TotalPositive) / float64(stats.TotalReviews)) * 100.0
		stats.AvgReviewsPerGame = float64(stats.TotalReviews) / float64(len(games))
	}

	if len(reviewCounts) > 0 {
		sort.Slice(reviewCounts, func(i, j int) bool { return reviewCounts[i] < reviewCounts[j] })
		stats.MedianReviewsPerGame = float64(reviewCounts[len(reviewCounts)/2])
	}

	if len(games) > 0 {
		stats.AvgPeakCCU = float64(sumCCU) / float64(len(games))
		stats.AvgPlaytimeForever = float64(sumPlaytime) / float64(len(games))
		stats.AvgPlaytime2Weeks = float64(sumPlay2w) / float64(len(games))
	}

	if len(playtimes) > 0 {
		sort.Slice(playtimes, func(i, j int) bool { return playtimes[i] < playtimes[j] })
		stats.MedianPlaytimeForever = playtimes[len(playtimes)/2]
	}

	if len(playtimes2w) > 0 {
		sort.Slice(playtimes2w, func(i, j int) bool { return playtimes2w[i] < playtimes2w[j] })
		stats.MedianPlaytime2Weeks = playtimes2w[len(playtimes2w)/2]
	}

	return stats
}
