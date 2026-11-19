package metrics

import (
	"math"
	"sort"
)

// GenreItem represents a genre and its metrics within a catalog.
type GenreItem struct {
	Genre             string  `json:"genre"`
	GameCount         int     `json:"game_count"`         // Raw count of games tagged with this genre
	WeightedWeight    float64 `json:"weighted_weight"`    // Sum of 1/K weights
	WeightedShare     float64 `json:"weighted_share"`     // percentage 0 to 100
	OwnershipMidpoint int64   `json:"ownership_midpoint"` // Sum of midpoints for games in this genre
}

// GenreYearTrend represents genre distribution in a specific release year.
type GenreYearTrend struct {
	Year        int                `json:"year"`
	TotalGames  int                `json:"total_games"`
	GenreCounts map[string]int     `json:"genre_counts"`
	GenreShares map[string]float64 `json:"genre_shares"`
}

// GenreAnalysisReport holds full genre portfolio analytics.
type GenreAnalysisReport struct {
	PublisherID        string           `json:"publisher_id"`
	TotalCatalogGames  int              `json:"total_catalog_games"`
	GamesWithGenres    int              `json:"games_with_genres"`
	GenreCoverage      float64          `json:"genre_coverage"` // percentage 0 to 100
	DistinctGenreCount int              `json:"distinct_genre_count"`
	DominantGenre      string           `json:"dominant_genre"`
	DominantGenreShare float64          `json:"dominant_genre_share"`
	GenreConcentration float64          `json:"genre_concentration"` // HHI of weighted genre shares (0 to 10,000)
	Genres             []GenreItem      `json:"genres"`
	Timeline           []GenreYearTrend `json:"timeline,omitempty"`
}

// GameGenreInput represents a game and its assigned genres.
type GameGenreInput struct {
	AppID       string
	Genres      []string
	ReleaseYear *int
	Midpoint    int64
}

// CalculateGenreMetrics computes raw titles, weighted 100% composition, and HHI concentration.
func CalculateGenreMetrics(publisherID string, totalGames int, games []GameGenreInput) *GenreAnalysisReport {
	report := &GenreAnalysisReport{
		PublisherID:       publisherID,
		TotalCatalogGames: totalGames,
	}

	if len(games) == 0 {
		return report
	}

	type genreAccum struct {
		count          int
		weight         float64
		ownersMidpoint int64
	}

	accumMap := make(map[string]*genreAccum)
	gamesWithGenres := 0
	timelineMap := make(map[int]map[string]int)

	for _, g := range games {
		validGenres := make([]string, 0, len(g.Genres))
		for _, rawG := range g.Genres {
			if rawG != "" {
				validGenres = append(validGenres, rawG)
			}
		}

		if len(validGenres) > 0 {
			gamesWithGenres++
			weight := 1.0 / float64(len(validGenres))

			for _, genre := range validGenres {
				acc, ok := accumMap[genre]
				if !ok {
					acc = &genreAccum{}
					accumMap[genre] = acc
				}
				acc.count++
				acc.weight += weight
				acc.ownersMidpoint += g.Midpoint

				if g.ReleaseYear != nil {
					yr := *g.ReleaseYear
					if _, yOk := timelineMap[yr]; !yOk {
						timelineMap[yr] = make(map[string]int)
					}
					timelineMap[yr][genre]++
				}
			}
		}
	}

	report.GamesWithGenres = gamesWithGenres
	if totalGames > 0 {
		report.GenreCoverage = (float64(gamesWithGenres) / float64(totalGames)) * 100.0
	}
	report.DistinctGenreCount = len(accumMap)

	if gamesWithGenres == 0 {
		return report
	}

	totalWeight := float64(gamesWithGenres)
	var list []GenreItem
	var hhiSum float64

	for gName, acc := range accumMap {
		share := (acc.weight / totalWeight) * 100.0
		shareFraction := acc.weight / totalWeight
		hhiSum += shareFraction * shareFraction

		list = append(list, GenreItem{
			Genre:             gName,
			GameCount:         acc.count,
			WeightedWeight:    acc.weight,
			WeightedShare:     share,
			OwnershipMidpoint: acc.ownersMidpoint,
		})
	}

	// Sort descending by weighted share
	sort.Slice(list, func(i, j int) bool {
		if list[i].WeightedShare == list[j].WeightedShare {
			return list[i].GameCount > list[j].GameCount
		}
		return list[i].WeightedShare > list[j].WeightedShare
	})

	report.Genres = list
	report.GenreConcentration = math.Min(10000.0, hhiSum*10000.0)

	if len(list) > 0 {
		report.DominantGenre = list[0].Genre
		report.DominantGenreShare = list[0].WeightedShare
	}

	// Build timeline
	var years []int
	for yr := range timelineMap {
		years = append(years, yr)
	}
	sort.Ints(years)

	for _, yr := range years {
		counts := timelineMap[yr]
		yearTotal := 0
		for _, c := range counts {
			yearTotal += c
		}
		shares := make(map[string]float64)
		for g, c := range counts {
			if yearTotal > 0 {
				shares[g] = (float64(c) / float64(yearTotal)) * 100.0
			}
		}
		report.Timeline = append(report.Timeline, GenreYearTrend{
			Year:        yr,
			TotalGames:  yearTotal,
			GenreCounts: counts,
			GenreShares: shares,
		})
	}

	return report
}
