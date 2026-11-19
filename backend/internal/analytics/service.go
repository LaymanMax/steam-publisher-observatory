package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/database"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/metrics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/normalization"
)

// Service coordinates data retrieval and analytics computations.
type Service struct {
	db     *database.DB
	config *config.Config
}

// NewService creates a new Analytics Service.
func NewService(db *database.DB, cfg *config.Config) *Service {
	return &Service{
		db:     db,
		config: cfg,
	}
}

// OverviewMetrics contains top-level observatory figures.
type OverviewMetrics struct {
	TotalGames             int64              `json:"total_games"`
	TotalPublishers        int64              `json:"total_publishers"`
	TotalDevelopers        int64              `json:"total_developers"`
	TotalGenres            int64              `json:"total_genres"`
	OwnershipCoverage      float64            `json:"ownership_coverage"`
	DatasetFreshness       *time.Time         `json:"dataset_freshness"`
	Source                 string             `json:"source"`
	LargestPublishers      []PublisherSummary `json:"largest_publishers"`
	ConcentratedPublishers []PublisherSummary `json:"concentrated_publishers"`
	DiversifiedPublishers  []PublisherSummary `json:"diversified_publishers"`
	Insights               []string           `json:"insights"`
}

// PublisherSummary represents row in publisher directory or summary table.
type PublisherSummary struct {
	PublisherID           string  `json:"publisher_id"`
	Name                  string  `json:"name"`
	CatalogSize           int64   `json:"catalog_size"`
	DeveloperCount        int64   `json:"developer_count"`
	GenreCount            int64   `json:"genre_count"`
	EarliestYear          *int    `json:"earliest_year"`
	LatestYear            *int    `json:"latest_year"`
	ValidOwnersCount      int64   `json:"valid_owners_count"`
	OwnershipCoverage     float64 `json:"ownership_coverage"`
	CatalogOwnersLower    int64   `json:"catalog_owners_lower"`
	CatalogOwnersMidpoint int64   `json:"catalog_owners_midpoint"`
	CatalogOwnersUpper    int64   `json:"catalog_owners_upper"`
	TopTitleName          string  `json:"top_title_name"`
	TopTitleMidpoint      int64   `json:"top_title_midpoint"`
	Top1Share             float64 `json:"top1_share"`
	Top3Share             float64 `json:"top3_share"`
	Top5Share             float64 `json:"top5_share"`
	Top10Share            float64 `json:"top10_share"`
	RemainingShare        float64 `json:"remaining_share"`
	HHI                   float64 `json:"hhi"`
	EffectiveTitleCount   float64 `json:"effective_title_count"`
	FreeTitleCount        int64   `json:"free_title_count"`
	FreeTitleShare        float64 `json:"free_title_share"`
	MedianPrice           float64 `json:"median_price"`
	AvgPrice              float64 `json:"avg_price"`
	TotalPositiveReviews  int64   `json:"total_positive_reviews"`
	TotalNegativeReviews  int64   `json:"total_negative_reviews"`
	TotalReviews          int64   `json:"total_reviews"`
	AvgPositiveShare      float64 `json:"avg_positive_share"`
	MedianPlaytime        int64   `json:"median_playtime"`
	AvgPeakCCU            float64 `json:"avg_peak_ccu"`
	MaxPeakCCU            int64   `json:"max_peak_ccu"`
	DominantGenre         string  `json:"dominant_genre"`
	GenreConcentration    float64 `json:"genre_concentration"`
	AvgQualityScore       float64 `json:"avg_quality_score"`
}

// PublisherFilter defines query parameters for publisher directory.
type PublisherFilter struct {
	Query     string `json:"query"`
	MinGames  int    `json:"min_games"`
	Genre     string `json:"genre"`
	SortBy    string `json:"sort_by"`
	SortOrder string `json:"sort_order"` // "asc" or "desc"
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

// GetOverview computes observatory dashboard highlights.
func (s *Service) GetOverview(ctx context.Context) (*OverviewMetrics, error) {
	metricsRes := &OverviewMetrics{}

	// Basic counts
	var latestImportTime sql.NullTime
	var source sql.NullString
	_ = s.db.QueryRowContext(ctx, "SELECT MAX(imported_at), MAX(source) FROM dataset_imports WHERE status = 'completed'").
		Scan(&latestImportTime, &source)
	if latestImportTime.Valid {
		metricsRes.DatasetFreshness = &latestImportTime.Time
		metricsRes.Source = source.String
	}

	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM games").Scan(&metricsRes.TotalGames)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM publishers").Scan(&metricsRes.TotalPublishers)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM developers").Scan(&metricsRes.TotalDevelopers)
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT genre) FROM game_genres").Scan(&metricsRes.TotalGenres)

	var validOwnersCount int64
	_ = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM games WHERE has_valid_owners = TRUE").Scan(&validOwnersCount)
	if metricsRes.TotalGames > 0 {
		metricsRes.OwnershipCoverage = (float64(validOwnersCount) / float64(metricsRes.TotalGames)) * 100.0
	}

	if metricsRes.TotalPublishers == 0 {
		return metricsRes, nil
	}

	// Largest publishers
	largest, err := s.queryPublisherSummaries(ctx, "ORDER BY catalog_size DESC LIMIT 5")
	if err == nil {
		metricsRes.LargestPublishers = largest
	}

	// Concentrated publishers (min 5 games)
	concentrated, err := s.queryPublisherSummaries(ctx, "WHERE catalog_size >= 5 AND catalog_owners_midpoint > 0 ORDER BY top5_share DESC LIMIT 5")
	if err == nil {
		metricsRes.ConcentratedPublishers = concentrated
	}

	// Diversified publishers (min 10 games)
	diversified, err := s.queryPublisherSummaries(ctx, "WHERE catalog_size >= 10 AND catalog_owners_midpoint > 0 ORDER BY hhi ASC LIMIT 5")
	if err == nil {
		metricsRes.DiversifiedPublishers = diversified
	}

	// Dynamic insights
	if len(metricsRes.ConcentratedPublishers) > 0 {
		topConc := metricsRes.ConcentratedPublishers[0]
		metricsRes.Insights = append(metricsRes.Insights, fmt.Sprintf(
			"%s exhibits high portfolio concentration: 5 titles account for %.1f%% of its catalog ownership midpoint.",
			topConc.Name, topConc.Top5Share,
		))
	}
	if len(metricsRes.DiversifiedPublishers) > 0 {
		topDiv := metricsRes.DiversifiedPublishers[0]
		metricsRes.Insights = append(metricsRes.Insights, fmt.Sprintf(
			"%s maintains an expansive, diversified portfolio with an effective title count of %.1f across %d games.",
			topDiv.Name, topDiv.EffectiveTitleCount, topDiv.CatalogSize,
		))
	}

	return metricsRes, nil
}

// ListPublishers returns paginated, sorted, and filtered publisher summaries.
func (s *Service) ListPublishers(ctx context.Context, filter PublisherFilter) ([]PublisherSummary, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 25
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	var whereClauses []string
	var args []interface{}

	if strings.TrimSpace(filter.Query) != "" {
		cleaned := strings.ToLower(normalization.CleanName(filter.Query))
		whereClauses = append(whereClauses, "LOWER(name) LIKE ?")
		args = append(args, "%"+cleaned+"%")
	}

	if filter.MinGames > 0 {
		whereClauses = append(whereClauses, "catalog_size >= ?")
		args = append(args, filter.MinGames)
	}

	if filter.Genre != "" {
		whereClauses = append(whereClauses, `
			publisher_id IN (
				SELECT DISTINCT gp.publisher_id 
				FROM game_publishers gp 
				JOIN game_genres gg ON gp.appid = gg.appid 
				WHERE LOWER(gg.genre) = LOWER(?)
			)
		`)
		args = append(args, filter.Genre)
	}

	whereSQL := ""
	if len(whereClauses) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClauses, " AND ")
	}

	// Count total matching
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM publisher_summaries %s", whereSQL)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count publishers query failed: %w", err)
	}

	// Order by
	sortCol := "catalog_size"
	switch filter.SortBy {
	case "catalog_size":
		sortCol = "catalog_size"
	case "catalog_owners_midpoint":
		sortCol = "catalog_owners_midpoint"
	case "hhi":
		sortCol = "hhi"
	case "top5_share":
		sortCol = "top5_share"
	case "developer_count":
		sortCol = "developer_count"
	case "genre_count":
		sortCol = "genre_count"
	case "total_reviews":
		sortCol = "total_reviews"
	case "avg_positive_share":
		sortCol = "avg_positive_share"
	case "latest_year":
		sortCol = "latest_year"
	case "earliest_year":
		sortCol = "earliest_year"
	case "avg_quality_score":
		sortCol = "avg_quality_score"
	}

	orderDir := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		orderDir = "ASC"
	}

	querySQL := fmt.Sprintf(`
		SELECT 
			publisher_id, name, catalog_size, developer_count, genre_count,
			earliest_year, latest_year, valid_owners_count, ownership_coverage,
			catalog_owners_lower, catalog_owners_midpoint, catalog_owners_upper,
			top_title_name, top_title_midpoint, top1_share, top3_share, top5_share,
			top10_share, remaining_share, hhi, effective_title_count,
			free_title_count, free_title_share, median_price, avg_price,
			total_positive_reviews, total_negative_reviews, total_reviews,
			avg_positive_share, median_playtime, avg_peak_ccu, max_peak_ccu,
			dominant_genre, genre_concentration, avg_quality_score
		FROM publisher_summaries
		%s
		ORDER BY %s %s, name ASC
		LIMIT ? OFFSET ?
	`, whereSQL, sortCol, orderDir)

	queryArgs := append(args, filter.Limit, filter.Offset)
	rows, err := s.db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("list publishers query failed: %w", err)
	}
	defer rows.Close()

	var results []PublisherSummary
	for rows.Next() {
		var s SummaryScan
		if err := s.Scan(rows); err == nil {
			results = append(results, s.ToSummary())
		}
	}

	return results, total, nil
}

// GetPublisherSummary retrieves the executive summary for a specific publisher.
func (s *Service) GetPublisherSummary(ctx context.Context, publisherID string) (*PublisherSummary, error) {
	querySQL := `
		SELECT 
			publisher_id, name, catalog_size, developer_count, genre_count,
			earliest_year, latest_year, valid_owners_count, ownership_coverage,
			catalog_owners_lower, catalog_owners_midpoint, catalog_owners_upper,
			top_title_name, top_title_midpoint, top1_share, top3_share, top5_share,
			top10_share, remaining_share, hhi, effective_title_count,
			free_title_count, free_title_share, median_price, avg_price,
			total_positive_reviews, total_negative_reviews, total_reviews,
			avg_positive_share, median_playtime, avg_peak_ccu, max_peak_ccu,
			dominant_genre, genre_concentration, avg_quality_score
		FROM publisher_summaries
		WHERE publisher_id = ?
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var scan SummaryScan
		if err := scan.Scan(rows); err != nil {
			return nil, err
		}
		summary := scan.ToSummary()
		return &summary, nil
	}

	return nil, fmt.Errorf("publisher with ID %s not found", publisherID)
}

// GameCatalogItem models a game in the publisher catalog explorer.
type GameCatalogItem struct {
	AppID            string   `json:"appid"`
	Name             string   `json:"name"`
	Developers       []string `json:"developers"`
	Genres           []string `json:"genres"`
	ReleaseDate      string   `json:"release_date"`
	ReleaseYear      *int     `json:"release_year"`
	Price            *float64 `json:"price"`
	PriceTier        string   `json:"price_tier"`
	OwnersRaw        string   `json:"owners_raw"`
	OwnersLower      *int64   `json:"owners_lower"`
	OwnersUpper      *int64   `json:"owners_upper"`
	OwnersMidpoint   *int64   `json:"owners_midpoint"`
	Positive         int64    `json:"positive"`
	Negative         int64    `json:"negative"`
	TotalReviews     int64    `json:"total_reviews"`
	PositiveShare    *float64 `json:"positive_share"`
	PeakCCU          int64    `json:"peak_ccu"`
	MedianPlaytime   int64    `json:"median_playtime"`
	HeaderImage      string   `json:"header_image"`
	ShortDescription string   `json:"short_description"`
	QualityScore     float64  `json:"quality_score"`
	IsPlaytest       bool     `json:"is_playtest"`
	IsSparse         bool     `json:"is_sparse"`
}

// CatalogFilter models query params for publisher catalog table.
type CatalogFilter struct {
	Query           string `json:"query"`
	Genre           string `json:"genre"`
	DeveloperID     string `json:"developer_id"`
	PriceTier       string `json:"price_tier"`
	MinYear         *int   `json:"min_year"`
	MaxYear         *int   `json:"max_year"`
	MinReviews      *int64 `json:"min_reviews"`
	ExcludePlaytest bool   `json:"exclude_playtest"`
	SortBy          string `json:"sort_by"`
	SortOrder       string `json:"sort_order"`
	Limit           int    `json:"limit"`
	Offset          int    `json:"offset"`
}

// GetPublisherCatalog retrieves filtered, sorted, paginated game catalog for a publisher.
func (s *Service) GetPublisherCatalog(ctx context.Context, publisherID string, filter CatalogFilter) ([]GameCatalogItem, int64, error) {
	if filter.Limit <= 0 {
		filter.Limit = 25
	}
	if filter.Limit > 200 {
		filter.Limit = 200
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	whereClauses := []string{"gp.publisher_id = ?"}
	args := []interface{}{publisherID}

	if filter.ExcludePlaytest {
		whereClauses = append(whereClauses, "g.is_playtest = FALSE")
	}

	if strings.TrimSpace(filter.Query) != "" {
		whereClauses = append(whereClauses, "LOWER(g.name) LIKE ?")
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filter.Query))+"%")
	}

	if filter.Genre != "" {
		whereClauses = append(whereClauses, `
			g.appid IN (
				SELECT appid FROM game_genres WHERE LOWER(genre) = LOWER(?)
			)
		`)
		args = append(args, filter.Genre)
	}

	if filter.DeveloperID != "" {
		whereClauses = append(whereClauses, `
			g.appid IN (
				SELECT appid FROM game_developers WHERE developer_id = ?
			)
		`)
		args = append(args, filter.DeveloperID)
	}

	if filter.PriceTier != "" {
		whereClauses = append(whereClauses, "g.price_tier = ?")
		args = append(args, filter.PriceTier)
	}

	if filter.MinYear != nil {
		whereClauses = append(whereClauses, "g.release_year >= ?")
		args = append(args, *filter.MinYear)
	}
	if filter.MaxYear != nil {
		whereClauses = append(whereClauses, "g.release_year <= ?")
		args = append(args, *filter.MaxYear)
	}
	if filter.MinReviews != nil {
		whereClauses = append(whereClauses, "g.total_reviews >= ?")
		args = append(args, *filter.MinReviews)
	}

	whereSQL := "WHERE " + strings.Join(whereClauses, " AND ")

	// Total count
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT g.appid) 
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		%s
	`, whereSQL)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Sort mapping
	sortCol := "g.owners_midpoint"
	switch filter.SortBy {
	case "name":
		sortCol = "g.name"
	case "release_date":
		sortCol = "g.release_year"
	case "price":
		sortCol = "g.price"
	case "owners_midpoint":
		sortCol = "g.owners_midpoint"
	case "total_reviews":
		sortCol = "g.total_reviews"
	case "positive_share":
		sortCol = "g.positive_share"
	case "peak_ccu":
		sortCol = "g.peak_ccu"
	case "median_playtime":
		sortCol = "g.median_playtime_forever"
	case "quality_score":
		sortCol = "g.quality_score"
	}

	orderDir := "DESC"
	if strings.EqualFold(filter.SortOrder, "asc") {
		orderDir = "ASC"
	}

	selectSQL := fmt.Sprintf(`
		SELECT 
			g.appid, g.name, g.release_date, g.release_year, g.price, g.price_tier,
			g.estimated_owners_raw, g.owners_lower, g.owners_upper, g.owners_midpoint,
			g.positive, g.negative, g.total_reviews, g.positive_share,
			g.peak_ccu, g.median_playtime_forever, g.header_image, g.short_description,
			g.quality_score, g.is_playtest, g.is_sparse
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		%s
		ORDER BY %s %s NULLS LAST, g.name ASC
		LIMIT ? OFFSET ?
	`, whereSQL, sortCol, orderDir)

	queryArgs := append(args, filter.Limit, filter.Offset)
	rows, err := s.db.QueryContext(ctx, selectSQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var games []GameCatalogItem
	var appIDs []string

	for rows.Next() {
		var (
			item                                     GameCatalogItem
			relDate, ownersRaw, headerImg, shortDesc sql.NullString
			relYear                                  sql.NullInt32
			price, posShare                          sql.NullFloat64
			ownersLower, ownersUpper, ownersMid      sql.NullInt64
		)

		err := rows.Scan(
			&item.AppID, &item.Name, &relDate, &relYear, &price, &item.PriceTier,
			&ownersRaw, &ownersLower, &ownersUpper, &ownersMid,
			&item.Positive, &item.Negative, &item.TotalReviews, &posShare,
			&item.PeakCCU, &item.MedianPlaytime, &headerImg, &shortDesc,
			&item.QualityScore, &item.IsPlaytest, &item.IsSparse,
		)
		if err != nil {
			continue
		}

		if relDate.Valid {
			item.ReleaseDate = relDate.String
		}
		if relYear.Valid {
			y := int(relYear.Int32)
			item.ReleaseYear = &y
		}
		if price.Valid {
			p := price.Float64
			item.Price = &p
		}
		if ownersRaw.Valid {
			item.OwnersRaw = ownersRaw.String
		}
		if ownersLower.Valid {
			l := ownersLower.Int64
			item.OwnersLower = &l
		}
		if ownersUpper.Valid {
			u := ownersUpper.Int64
			item.OwnersUpper = &u
		}
		if ownersMid.Valid {
			m := ownersMid.Int64
			item.OwnersMidpoint = &m
		}
		if posShare.Valid {
			ps := posShare.Float64
			item.PositiveShare = &ps
		}
		if headerImg.Valid {
			item.HeaderImage = headerImg.String
		}
		if shortDesc.Valid {
			item.ShortDescription = shortDesc.String
		}

		games = append(games, item)
		appIDs = append(appIDs, item.AppID)
	}

	// Fetch genres and developers for these games
	if len(appIDs) > 0 {
		s.attachGenresAndDevelopers(ctx, games)
	}

	return games, total, nil
}

func (s *Service) attachGenresAndDevelopers(ctx context.Context, games []GameCatalogItem) {
	appMap := make(map[string]*GameCatalogItem)
	var placeholders []string
	var args []interface{}
	for i := range games {
		appMap[games[i].AppID] = &games[i]
		placeholders = append(placeholders, "?")
		args = append(args, games[i].AppID)
	}
	inSQL := strings.Join(placeholders, ",")

	// Genres
	gRows, err := s.db.QueryContext(ctx, fmt.Sprintf("SELECT appid, genre FROM game_genres WHERE appid IN (%s)", inSQL), args...)
	if err == nil {
		defer gRows.Close()
		for gRows.Next() {
			var appid, genre string
			if gRows.Scan(&appid, &genre) == nil {
				if g, ok := appMap[appid]; ok {
					g.Genres = append(g.Genres, genre)
				}
			}
		}
	}

	// Developers
	dRows, err := s.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT gd.appid, d.name 
		FROM game_developers gd 
		JOIN developers d ON gd.developer_id = d.id 
		WHERE gd.appid IN (%s)
	`, inSQL), args...)
	if err == nil {
		defer dRows.Close()
		for dRows.Next() {
			var appid, devName string
			if dRows.Scan(&appid, &devName) == nil {
				if g, ok := appMap[appid]; ok {
					g.Developers = append(g.Developers, devName)
				}
			}
		}
	}
}

// GetPublisherDependency calculates portfolio dependency for a publisher.
func (s *Service) GetPublisherDependency(ctx context.Context, publisherID string) (*metrics.PortfolioDependencyReport, error) {
	summary, err := s.GetPublisherSummary(ctx, publisherID)
	if err != nil {
		return nil, err
	}

	querySQL := `
		SELECT g.appid, g.name, g.owners_lower, g.owners_upper, g.owners_midpoint 
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		WHERE gp.publisher_id = ?
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inputs []metrics.GameOwnershipInput
	for rows.Next() {
		var appid, name string
		var lower, upper, mid sql.NullInt64
		if err := rows.Scan(&appid, &name, &lower, &upper, &mid); err == nil {
			var lPtr, uPtr, mPtr *int64
			if lower.Valid {
				l := lower.Int64
				lPtr = &l
			}
			if upper.Valid {
				u := upper.Int64
				uPtr = &u
			}
			if mid.Valid {
				m := mid.Int64
				mPtr = &m
			}
			inputs = append(inputs, metrics.GameOwnershipInput{
				AppID:          appid,
				Name:           name,
				OwnersLower:    lPtr,
				OwnersUpper:    uPtr,
				OwnersMidpoint: mPtr,
			})
		}
	}

	report := metrics.CalculatePortfolioDependency(publisherID, summary.Name, inputs)
	return report, nil
}

// GetPublisherGenres calculates genre distribution for a publisher.
func (s *Service) GetPublisherGenres(ctx context.Context, publisherID string) (*metrics.GenreAnalysisReport, error) {
	summary, err := s.GetPublisherSummary(ctx, publisherID)
	if err != nil {
		return nil, err
	}

	querySQL := `
		SELECT g.appid, g.release_year, COALESCE(g.owners_midpoint, 0), gg.genre
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		LEFT JOIN game_genres gg ON g.appid = gg.appid
		WHERE gp.publisher_id = ?
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gameGenresMap := make(map[string]*metrics.GameGenreInput)
	for rows.Next() {
		var appid string
		var relYear sql.NullInt32
		var midpoint int64
		var genre sql.NullString

		if err := rows.Scan(&appid, &relYear, &midpoint, &genre); err == nil {
			entry, ok := gameGenresMap[appid]
			if !ok {
				var yr *int
				if relYear.Valid {
					y := int(relYear.Int32)
					yr = &y
				}
				entry = &metrics.GameGenreInput{
					AppID:       appid,
					ReleaseYear: yr,
					Midpoint:    midpoint,
				}
				gameGenresMap[appid] = entry
			}
			if genre.Valid && strings.TrimSpace(genre.String) != "" {
				entry.Genres = append(entry.Genres, strings.TrimSpace(genre.String))
			}
		}
	}

	var inputs []metrics.GameGenreInput
	for _, v := range gameGenresMap {
		inputs = append(inputs, *v)
	}

	report := metrics.CalculateGenreMetrics(publisherID, int(summary.CatalogSize), inputs)
	return report, nil
}

// GetPublisherPrices calculates price tier composition for a publisher.
func (s *Service) GetPublisherPrices(ctx context.Context, publisherID string) (*metrics.PricingReport, error) {
	querySQL := `
		SELECT g.appid, g.price 
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		WHERE gp.publisher_id = ?
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var inputs []metrics.GamePriceInput
	for rows.Next() {
		var appid string
		var price sql.NullFloat64
		if err := rows.Scan(&appid, &price); err == nil {
			var pPtr *float64
			if price.Valid {
				p := price.Float64
				pPtr = &p
			}
			inputs = append(inputs, metrics.GamePriceInput{
				AppID: appid,
				Price: pPtr,
			})
		}
	}

	report := metrics.CalculatePricingMetrics(publisherID, inputs, s.config.PriceTiers)
	return report, nil
}

// DeveloperRelationshipItem models a studio that collaborates with the publisher.
type DeveloperRelationshipItem struct {
	DeveloperID      string   `json:"developer_id"`
	DeveloperName    string   `json:"developer_name"`
	AssociatedGames  int      `json:"associated_games"`
	ShareOfCatalog   float64  `json:"share_of_catalog"` // % of publisher's titles
	OwnersMidpoint   int64    `json:"owners_midpoint"`
	Genres           []string `json:"genres"`
	SharedPublishers int      `json:"shared_publishers"` // other publishers this developer worked with
	Games            []struct {
		AppID       string `json:"appid"`
		Name        string `json:"name"`
		ReleaseYear *int   `json:"release_year"`
		Midpoint    int64  `json:"midpoint"`
	} `json:"games"`
}

// DeveloperAnalysisReport encapsulates developer relationships for a publisher.
type DeveloperAnalysisReport struct {
	PublisherID            string                      `json:"publisher_id"`
	TotalDevelopers        int                         `json:"total_developers"`
	SingleTitleDevelopers  int                         `json:"single_title_developers"`
	MultiTitleDevelopers   int                         `json:"multi_title_developers"`
	DeveloperConcentration float64                     `json:"developer_concentration"` // HHI of developer shares
	TopDeveloperName       string                      `json:"top_developer_name"`
	TopDeveloperShare      float64                     `json:"top_developer_share"`
	Developers             []DeveloperRelationshipItem `json:"developers"`
	Notice                 string                      `json:"notice"`
}

// GetPublisherDevelopers analyzes developer collaboration patterns.
func (s *Service) GetPublisherDevelopers(ctx context.Context, publisherID string) (*DeveloperAnalysisReport, error) {
	summary, err := s.GetPublisherSummary(ctx, publisherID)
	if err != nil {
		return nil, err
	}

	querySQL := `
		SELECT 
			d.id, d.name, g.appid, g.name, g.release_year, COALESCE(g.owners_midpoint, 0)
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		JOIN game_developers gd ON g.appid = gd.appid
		JOIN developers d ON gd.developer_id = d.id
		WHERE gp.publisher_id = ?
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devMap := make(map[string]*DeveloperRelationshipItem)

	for rows.Next() {
		var devID, devName, appid, gameName string
		var relYear sql.NullInt32
		var midpoint int64

		if err := rows.Scan(&devID, &devName, &appid, &gameName, &relYear, &midpoint); err == nil {
			item, ok := devMap[devID]
			if !ok {
				item = &DeveloperRelationshipItem{
					DeveloperID:   devID,
					DeveloperName: devName,
				}
				devMap[devID] = item
			}
			item.AssociatedGames++
			item.OwnersMidpoint += midpoint

			var yr *int
			if relYear.Valid {
				y := int(relYear.Int32)
				yr = &y
			}
			item.Games = append(item.Games, struct {
				AppID       string `json:"appid"`
				Name        string `json:"name"`
				ReleaseYear *int   `json:"release_year"`
				Midpoint    int64  `json:"midpoint"`
			}{
				AppID:       appid,
				Name:        gameName,
				ReleaseYear: yr,
				Midpoint:    midpoint,
			})
		}
	}

	report := &DeveloperAnalysisReport{
		PublisherID:     publisherID,
		TotalDevelopers: len(devMap),
		Notice:          "Observed game credits from catalog metadata; does not represent verified contractual publishing arrangements.",
	}

	totalCatalog := float64(summary.CatalogSize)
	var list []DeveloperRelationshipItem
	var hhiSum float64

	for _, item := range devMap {
		if item.AssociatedGames == 1 {
			report.SingleTitleDevelopers++
		} else {
			report.MultiTitleDevelopers++
		}

		if totalCatalog > 0 {
			item.ShareOfCatalog = (float64(item.AssociatedGames) / totalCatalog) * 100.0
			shareFrac := float64(item.AssociatedGames) / totalCatalog
			hhiSum += shareFrac * shareFrac
		}

		// Count other publishers for this developer
		var otherPubsCount int
		_ = s.db.QueryRowContext(ctx, `
			SELECT COUNT(DISTINCT gp2.publisher_id) 
			FROM game_developers gd2 
			JOIN game_publishers gp2 ON gd2.appid = gp2.appid 
			WHERE gd2.developer_id = ? AND gp2.publisher_id != ?
		`, item.DeveloperID, publisherID).Scan(&otherPubsCount)
		item.SharedPublishers = otherPubsCount

		list = append(list, *item)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].AssociatedGames == list[j].AssociatedGames {
			return list[i].OwnersMidpoint > list[j].OwnersMidpoint
		}
		return list[i].AssociatedGames > list[j].AssociatedGames
	})

	report.Developers = list
	report.DeveloperConcentration = math.Min(10000.0, hhiSum*10000.0)

	if len(list) > 0 {
		report.TopDeveloperName = list[0].DeveloperName
		report.TopDeveloperShare = list[0].ShareOfCatalog
	}

	return report, nil
}

// TimelineYearEntry models annual catalog activity.
type TimelineYearEntry struct {
	Year             int     `json:"year"`
	GameCount        int     `json:"game_count"`
	OwnersMidpoint   int64   `json:"owners_midpoint"`
	FreeCount        int     `json:"free_count"`
	PaidCount        int     `json:"paid_count"`
	DeveloperCount   int     `json:"developer_count"`
	TotalReviews     int64   `json:"total_reviews"`
	AvgPositiveShare float64 `json:"avg_positive_share"`
	TopTitle         string  `json:"top_title"`
}

// TimelineReport models the historical evolution of a publisher's catalog.
type TimelineReport struct {
	PublisherID      string              `json:"publisher_id"`
	EarliestYear     *int                `json:"earliest_year"`
	LatestYear       *int                `json:"latest_year"`
	UnknownDateGames int                 `json:"unknown_date_games"`
	Years            []TimelineYearEntry `json:"years"`
}

// GetPublisherTimeline retrieves annual breakdown of releases and metrics.
func (s *Service) GetPublisherTimeline(ctx context.Context, publisherID string) (*TimelineReport, error) {
	querySQL := `
		SELECT 
			g.release_year,
			COUNT(DISTINCT g.appid) AS game_count,
			COALESCE(SUM(g.owners_midpoint), 0) AS owners_midpoint,
			COUNT(CASE WHEN g.price = 0.0 THEN 1 END) AS free_count,
			COUNT(CASE WHEN g.price > 0.0 THEN 1 END) AS paid_count,
			COUNT(DISTINCT gd.developer_id) AS dev_count,
			COALESCE(SUM(g.total_reviews), 0) AS total_reviews,
			COALESCE(SUM(g.positive), 0) AS pos_reviews
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		LEFT JOIN game_developers gd ON g.appid = gd.appid
		WHERE gp.publisher_id = ?
		GROUP BY g.release_year
		ORDER BY g.release_year ASC
	`
	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	report := &TimelineReport{PublisherID: publisherID}
	for rows.Next() {
		var relYear sql.NullInt32
		var gameCount, freeCount, paidCount, devCount int
		var midpoint, totalReviews, posReviews int64

		if err := rows.Scan(&relYear, &gameCount, &midpoint, &freeCount, &paidCount, &devCount, &totalReviews, &posReviews); err == nil {
			if !relYear.Valid {
				report.UnknownDateGames += gameCount
				continue
			}
			y := int(relYear.Int32)
			if report.EarliestYear == nil || y < *report.EarliestYear {
				report.EarliestYear = &y
			}
			if report.LatestYear == nil || y > *report.LatestYear {
				report.LatestYear = &y
			}

			posShare := 0.0
			if totalReviews > 0 {
				posShare = (float64(posReviews) / float64(totalReviews)) * 100.0
			}

			report.Years = append(report.Years, TimelineYearEntry{
				Year:             y,
				GameCount:        gameCount,
				OwnersMidpoint:   midpoint,
				FreeCount:        freeCount,
				PaidCount:        paidCount,
				DeveloperCount:   devCount,
				TotalReviews:     totalReviews,
				AvgPositiveShare: posShare,
			})
		}
	}

	return report, nil
}

// GetPublisherOutliers runs IQR outlier detection on a chosen portfolio metric.
func (s *Service) GetPublisherOutliers(ctx context.Context, publisherID, metricName string) (*OutlierAnalysisReport, error) {
	colName := "owners_midpoint"
	switch metricName {
	case "owners_midpoint":
		colName = "owners_midpoint"
	case "positive_reviews":
		colName = "positive"
	case "total_reviews":
		colName = "total_reviews"
	case "peak_ccu":
		colName = "peak_ccu"
	case "recommendations":
		colName = "recommendations"
	case "median_playtime":
		colName = "median_playtime_forever"
	case "price":
		colName = "price"
	default:
		metricName = "owners_midpoint"
		colName = "owners_midpoint"
	}

	querySQL := fmt.Sprintf(`
		SELECT g.appid, g.name, CAST(g.%s AS DOUBLE) 
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		WHERE gp.publisher_id = ? AND g.%s IS NOT NULL
	`, colName, colName)

	rows, err := s.db.QueryContext(ctx, querySQL, publisherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []DataPoint
	for rows.Next() {
		var appid, name string
		var val sql.NullFloat64
		if err := rows.Scan(&appid, &name, &val); err == nil && val.Valid {
			points = append(points, DataPoint{
				AppID:   appid,
				Name:    name,
				Value:   val.Float64,
				IsValid: true,
			})
		}
	}

	report := DetectOutliersIQR(metricName, points)
	return report, nil
}

// PublisherQualityReport models missing and incomplete data for a publisher.
type PublisherQualityReport struct {
	PublisherID           string  `json:"publisher_id"`
	TotalGames            int64   `json:"total_games"`
	CompletenessScore     float64 `json:"completeness_score"` // 0 to 100
	MissingOwnersCount    int64   `json:"missing_owners_count"`
	MissingPriceCount     int64   `json:"missing_price_count"`
	MissingDateCount      int64   `json:"missing_date_count"`
	MissingGenreCount     int64   `json:"missing_genre_count"`
	MissingDeveloperCount int64   `json:"missing_developer_count"`
	PlaytestCount         int64   `json:"playtest_count"`
	SparseCount           int64   `json:"sparse_count"`
}

// GetPublisherQuality returns data completeness and missing record diagnostics for a publisher.
func (s *Service) GetPublisherQuality(ctx context.Context, publisherID string) (*PublisherQualityReport, error) {
	querySQL := `
		SELECT 
			COUNT(*),
			COALESCE(AVG(g.quality_score), 0.0),
			COUNT(CASE WHEN g.has_valid_owners = FALSE THEN 1 END),
			COUNT(CASE WHEN g.price IS NULL THEN 1 END),
			COUNT(CASE WHEN g.has_valid_date = FALSE THEN 1 END),
			COUNT(CASE WHEN g.appid NOT IN (SELECT appid FROM game_genres) THEN 1 END),
			COUNT(CASE WHEN g.appid NOT IN (SELECT appid FROM game_developers) THEN 1 END),
			COUNT(CASE WHEN g.is_playtest = TRUE THEN 1 END),
			COUNT(CASE WHEN g.is_sparse = TRUE THEN 1 END)
		FROM game_publishers gp
		JOIN games g ON gp.appid = g.appid
		WHERE gp.publisher_id = ?
	`
	rep := &PublisherQualityReport{PublisherID: publisherID}
	err := s.db.QueryRowContext(ctx, querySQL, publisherID).Scan(
		&rep.TotalGames, &rep.CompletenessScore,
		&rep.MissingOwnersCount, &rep.MissingPriceCount,
		&rep.MissingDateCount, &rep.MissingGenreCount,
		&rep.MissingDeveloperCount, &rep.PlaytestCount,
		&rep.SparseCount,
	)
	return rep, err
}

// GlobalDataQualityDiagnostics models system-wide dataset ingestion diagnostics.
type GlobalDataQualityDiagnostics struct {
	TotalRows              int64   `json:"total_rows"`
	UniqueAppIDs           int64   `json:"unique_appids"`
	DuplicateAppIDs        int64   `json:"duplicate_appids"`
	GamesWithPublishers    int64   `json:"games_with_publishers"`
	GamesWithoutPublishers int64   `json:"games_without_publishers"`
	GamesWithDevelopers    int64   `json:"games_with_developers"`
	GamesWithoutDevelopers int64   `json:"games_without_developers"`
	MissingGenres          int64   `json:"missing_genres"`
	MissingReleaseDates    int64   `json:"missing_release_dates"`
	MissingPrices          int64   `json:"missing_prices"`
	MissingOwners          int64   `json:"missing_owners"`
	SuspectedPlaytests     int64   `json:"suspected_playtests"`
	SparseRecords          int64   `json:"sparse_records"`
	AverageCompleteness    float64 `json:"average_completeness"`
}

// GetDataQualityDiagnostics queries overall dataset health metrics.
func (s *Service) GetDataQualityDiagnostics(ctx context.Context) (*GlobalDataQualityDiagnostics, error) {
	querySQL := `
		SELECT 
			COUNT(*),
			COUNT(DISTINCT appid),
			COUNT(CASE WHEN appid IN (SELECT appid FROM game_publishers) THEN 1 END),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_publishers) THEN 1 END),
			COUNT(CASE WHEN appid IN (SELECT appid FROM game_developers) THEN 1 END),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_developers) THEN 1 END),
			COUNT(CASE WHEN appid NOT IN (SELECT appid FROM game_genres) THEN 1 END),
			COUNT(CASE WHEN has_valid_date = FALSE THEN 1 END),
			COUNT(CASE WHEN price IS NULL THEN 1 END),
			COUNT(CASE WHEN has_valid_owners = FALSE THEN 1 END),
			COUNT(CASE WHEN is_playtest = TRUE THEN 1 END),
			COUNT(CASE WHEN is_sparse = TRUE THEN 1 END),
			COALESCE(AVG(quality_score), 0.0)
		FROM games
	`
	diag := &GlobalDataQualityDiagnostics{}
	err := s.db.QueryRowContext(ctx, querySQL).Scan(
		&diag.TotalRows, &diag.UniqueAppIDs,
		&diag.GamesWithPublishers, &diag.GamesWithoutPublishers,
		&diag.GamesWithDevelopers, &diag.GamesWithoutDevelopers,
		&diag.MissingGenres, &diag.MissingReleaseDates,
		&diag.MissingPrices, &diag.MissingOwners,
		&diag.SuspectedPlaytests, &diag.SparseRecords,
		&diag.AverageCompleteness,
	)
	diag.DuplicateAppIDs = diag.TotalRows - diag.UniqueAppIDs
	return diag, err
}

// ProblematicRecord models a game flagged with data quality issues.
type ProblematicRecord struct {
	AppID        string  `json:"appid"`
	Name         string  `json:"name"`
	Issue        string  `json:"issue"`
	QualityScore float64 `json:"quality_score"`
	Flags        string  `json:"flags"`
}

// ListProblematicRecords retrieves games with specific data quality issues.
func (s *Service) ListProblematicRecords(ctx context.Context, issue string, limit, offset int) ([]ProblematicRecord, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	whereClause := "1=1"
	switch issue {
	case "playtest":
		whereClause = "is_playtest = TRUE"
	case "sparse":
		whereClause = "is_sparse = TRUE"
	case "missing_owners":
		whereClause = "has_valid_owners = FALSE"
	case "missing_price":
		whereClause = "price IS NULL"
	case "missing_date":
		whereClause = "has_valid_date = FALSE"
	case "missing_publisher":
		whereClause = "appid NOT IN (SELECT appid FROM game_publishers)"
	case "missing_developer":
		whereClause = "appid NOT IN (SELECT appid FROM game_developers)"
	default:
		whereClause = "quality_score < 70.0"
	}

	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM games WHERE %s", whereClause)
	var total int64
	if err := s.db.QueryRowContext(ctx, countSQL).Scan(&total); err != nil {
		return nil, 0, err
	}

	selectSQL := fmt.Sprintf(`
		SELECT appid, name, quality_score, quality_flags 
		FROM games 
		WHERE %s 
		ORDER BY quality_score ASC, appid ASC 
		LIMIT ? OFFSET ?
	`, whereClause)

	rows, err := s.db.QueryContext(ctx, selectSQL, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []ProblematicRecord
	for rows.Next() {
		var appid, name, flags string
		var score float64
		if err := rows.Scan(&appid, &name, &score, &flags); err == nil {
			records = append(records, ProblematicRecord{
				AppID:        appid,
				Name:         name,
				Issue:        issue,
				QualityScore: score,
				Flags:        flags,
			})
		}
	}

	return records, total, nil
}

// EntityAliasRecord models a stored manual alias.
type EntityAliasRecord struct {
	ID            string    `json:"id"`
	EntityType    string    `json:"entity_type"`
	SourceName    string    `json:"source_name"`
	CanonicalID   string    `json:"canonical_id"`
	CanonicalName string    `json:"canonical_name"`
	CreatedAt     time.Time `json:"created_at"`
	Method        string    `json:"method"`
}

// PotentialAliasPair models two entities that might be duplicates.
type PotentialAliasPair struct {
	EntityA    PublisherSummary `json:"entity_a"`
	EntityB    PublisherSummary `json:"entity_b"`
	Similarity float64          `json:"similarity"`
	Reason     string           `json:"reason"`
}

// ListPotentialAliases detects pairs of entities that might be duplicates for manual review.
func (s *Service) ListPotentialAliases(ctx context.Context, limit int) ([]PotentialAliasPair, error) {
	if limit <= 0 {
		limit = 20
	}

	// Fetch top publishers by catalog size
	pubs, err := s.queryPublisherSummaries(ctx, "ORDER BY catalog_size DESC LIMIT 200")
	if err != nil {
		return nil, err
	}

	var pairs []PotentialAliasPair
	for i := 0; i < len(pubs); i++ {
		for j := i + 1; j < len(pubs); j++ {
			p1 := pubs[i]
			p2 := pubs[j]
			isAlias, score := normalization.IsPotentialAlias(p1.Name, p2.Name)
			if isAlias {
				pairs = append(pairs, PotentialAliasPair{
					EntityA:    p1,
					EntityB:    p2,
					Similarity: score,
					Reason:     fmt.Sprintf("Name substring or token overlap (%s vs %s)", p1.Name, p2.Name),
				})
				if len(pairs) >= limit {
					return pairs, nil
				}
			}
		}
	}

	return pairs, nil
}

// CreateAlias merges a source entity into a canonical entity and triggers recalculation.
func (s *Service) CreateAlias(ctx context.Context, entityType, sourceName, canonicalID string) error {
	s.db.LockForWrite()
	defer s.db.UnlockForWrite()

	aliasID := uuid.New().String()

	// Get canonical name
	var canonicalName string
	tableName := "publishers"
	if entityType == "developer" {
		tableName = "developers"
	}
	err := s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT name FROM %s WHERE id = ?", tableName), canonicalID).Scan(&canonicalName)
	if err != nil {
		return fmt.Errorf("canonical %s with ID %s not found: %w", entityType, canonicalID, err)
	}

	// Insert alias record
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO entity_aliases (id, entity_type, source_name, canonical_id, canonical_name, created_at, method)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, 'manual_merge')
	`, aliasID, entityType, sourceName, canonicalID, canonicalName)
	if err != nil {
		return fmt.Errorf("failed to save alias record: %w", err)
	}

	// Remap junction tables
	sourceKey := normalization.CanonicalKey(sourceName)
	var sourceID string
	_ = s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT id FROM %s WHERE normalized_key = ?", tableName), sourceKey).Scan(&sourceID)

	if sourceID != "" && sourceID != canonicalID {
		relTable := "game_publishers"
		colName := "publisher_id"
		if entityType == "developer" {
			relTable = "game_developers"
			colName = "developer_id"
		}
		_, _ = s.db.ExecContext(ctx, fmt.Sprintf("UPDATE OR IGNORE %s SET %s = ? WHERE %s = ?", relTable, colName, colName), canonicalID, sourceID)
		_, _ = s.db.ExecContext(ctx, fmt.Sprintf("UPDATE %s SET alias_of_id = ? WHERE id = ?", tableName), canonicalID, sourceID)
	}

	// Materialize publisher summaries to reflect updated mapping
	imp := s.importerInstance()
	return imp.MaterializePublisherSummaries(ctx)
}

// RemoveAlias removes a manual merge decision and restores original relationship.
func (s *Service) RemoveAlias(ctx context.Context, aliasID string) error {
	s.db.LockForWrite()
	defer s.db.UnlockForWrite()

	_, err := s.db.ExecContext(ctx, "DELETE FROM entity_aliases WHERE id = ?", aliasID)
	if err != nil {
		return err
	}

	imp := s.importerInstance()
	return imp.MaterializePublisherSummaries(ctx)
}

// ListAliases returns all active manual aliases.
func (s *Service) ListAliases(ctx context.Context) ([]EntityAliasRecord, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, entity_type, source_name, canonical_id, canonical_name, created_at, method FROM entity_aliases ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []EntityAliasRecord
	for rows.Next() {
		var a EntityAliasRecord
		if err := rows.Scan(&a.ID, &a.EntityType, &a.SourceName, &a.CanonicalID, &a.CanonicalName, &a.CreatedAt, &a.Method); err == nil {
			list = append(list, a)
		}
	}
	return list, nil
}

func (s *Service) importerInstance() *importSummarizer {
	return &importSummarizer{db: s.db}
}

type importSummarizer struct {
	db *database.DB
}

func (is *importSummarizer) MaterializePublisherSummaries(ctx context.Context) error {
	// Re-executes the materialization logic
	query := `
		INSERT OR REPLACE INTO publisher_summaries
		WITH pub_games AS (
			SELECT 
				gp.publisher_id,
				p.name AS publisher_name,
				g.appid,
				g.name AS game_name,
				g.release_year,
				g.price,
				g.has_valid_owners,
				g.owners_lower,
				g.owners_upper,
				g.owners_midpoint,
				g.positive,
				g.negative,
				g.total_reviews,
				g.peak_ccu,
				g.median_playtime_forever,
				g.quality_score
			FROM game_publishers gp
			JOIN publishers p ON gp.publisher_id = p.id
			JOIN games g ON gp.appid = g.appid
		),
		pub_ranked AS (
			SELECT 
				*,
				ROW_NUMBER() OVER(PARTITION BY publisher_id ORDER BY COALESCE(owners_midpoint, 0) DESC, total_reviews DESC) AS rank
			FROM pub_games
		),
		pub_devs AS (
			SELECT 
				gp.publisher_id,
				COUNT(DISTINCT gd.developer_id) AS dev_count
			FROM game_publishers gp
			JOIN game_developers gd ON gp.appid = gd.appid
			GROUP BY gp.publisher_id
		),
		pub_genres AS (
			SELECT 
				gp.publisher_id,
				COUNT(DISTINCT gg.genre) AS genre_count
			FROM game_publishers gp
			JOIN game_genres gg ON gp.appid = gg.appid
			GROUP BY gp.publisher_id
		),
		pub_top_shares AS (
			SELECT 
				publisher_id,
				MAX(CASE WHEN rank = 1 THEN game_name END) AS top_title_name,
				MAX(CASE WHEN rank = 1 THEN owners_midpoint END) AS top_title_midpoint,
				SUM(CASE WHEN rank <= 1 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top1_midpoint,
				SUM(CASE WHEN rank <= 3 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top3_midpoint,
				SUM(CASE WHEN rank <= 5 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top5_midpoint,
				SUM(CASE WHEN rank <= 10 THEN COALESCE(owners_midpoint, 0) ELSE 0 END) AS top10_midpoint
			FROM pub_ranked
			GROUP BY publisher_id
		),
		pub_hhi_prep AS (
			SELECT 
				pr.publisher_id,
				SUM(
					CASE 
						WHEN tot.total_midpoint > 0 AND pr.owners_midpoint IS NOT NULL 
						THEN POWER(CAST(pr.owners_midpoint AS DOUBLE) / CAST(tot.total_midpoint AS DOUBLE), 2)
						ELSE 0.0 
					END
				) AS unscaled_hhi
			FROM pub_ranked pr
			JOIN (
				SELECT publisher_id, SUM(COALESCE(owners_midpoint, 0)) AS total_midpoint 
				FROM pub_games 
				GROUP BY publisher_id
			) tot ON pr.publisher_id = tot.publisher_id
			GROUP BY pr.publisher_id
		)
		SELECT 
			pg.publisher_id,
			MAX(pg.publisher_name),
			COUNT(DISTINCT pg.appid) AS catalog_size,
			COALESCE(MAX(pd.dev_count), 0) AS developer_count,
			COALESCE(MAX(pgn.genre_count), 0) AS genre_count,
			MIN(pg.release_year) AS earliest_year,
			MAX(pg.release_year) AS latest_year,
			COUNT(CASE WHEN pg.has_valid_owners THEN 1 END) AS valid_owners_count,
			CAST(COUNT(CASE WHEN pg.has_valid_owners THEN 1 END) AS DOUBLE) / CAST(COUNT(DISTINCT pg.appid) AS DOUBLE) * 100.0 AS ownership_coverage,
			COALESCE(SUM(pg.owners_lower), 0) AS catalog_owners_lower,
			COALESCE(SUM(pg.owners_midpoint), 0) AS catalog_owners_midpoint,
			COALESCE(SUM(pg.owners_upper), 0) AS catalog_owners_upper,
			MAX(pts.top_title_name) AS top_title_name,
			COALESCE(MAX(pts.top_title_midpoint), 0) AS top_title_midpoint,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top1_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top1_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top3_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top3_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top5_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top5_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN CAST(MAX(pts.top10_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE) * 100.0 ELSE 0.0 END AS top10_share,
			CASE WHEN SUM(COALESCE(pg.owners_midpoint, 0)) > 0 THEN (1.0 - (CAST(MAX(pts.top5_midpoint) AS DOUBLE) / CAST(SUM(pg.owners_midpoint) AS DOUBLE))) * 100.0 ELSE 0.0 END AS remaining_share,
			COALESCE(MAX(ph.unscaled_hhi), 0.0) * 10000.0 AS hhi,
			CASE WHEN COALESCE(MAX(ph.unscaled_hhi), 0.0) > 0.00001 THEN 1.0 / MAX(ph.unscaled_hhi) ELSE 0.0 END AS effective_title_count,
			COUNT(CASE WHEN pg.price = 0.0 THEN 1 END) AS free_title_count,
			CAST(COUNT(CASE WHEN pg.price = 0.0 THEN 1 END) AS DOUBLE) / CAST(COUNT(DISTINCT pg.appid) AS DOUBLE) * 100.0 AS free_title_share,
			COALESCE(MEDIAN(pg.price), 0.0) AS median_price,
			COALESCE(AVG(pg.price), 0.0) AS avg_price,
			COALESCE(SUM(pg.positive), 0) AS total_positive_reviews,
			COALESCE(SUM(pg.negative), 0) AS total_negative_reviews,
			COALESCE(SUM(pg.total_reviews), 0) AS total_reviews,
			CASE WHEN SUM(pg.total_reviews) > 0 THEN CAST(SUM(pg.positive) AS DOUBLE) / CAST(SUM(pg.total_reviews) AS DOUBLE) * 100.0 ELSE 0.0 END AS avg_positive_share,
			COALESCE(MEDIAN(pg.median_playtime_forever), 0) AS median_playtime,
			COALESCE(AVG(pg.peak_ccu), 0.0) AS avg_peak_ccu,
			COALESCE(MAX(pg.peak_ccu), 0) AS max_peak_ccu,
			'' AS dominant_genre,
			0.0 AS genre_concentration,
			COALESCE(AVG(pg.quality_score), 0.0) AS avg_quality_score,
			CURRENT_TIMESTAMP AS updated_at
		FROM pub_games pg
		LEFT JOIN pub_devs pd ON pg.publisher_id = pd.publisher_id
		LEFT JOIN pub_genres pgn ON pg.publisher_id = pgn.publisher_id
		LEFT JOIN pub_top_shares pts ON pg.publisher_id = pts.publisher_id
		LEFT JOIN pub_hhi_prep ph ON pg.publisher_id = ph.publisher_id
		GROUP BY pg.publisher_id
	`
	_, err := is.db.ExecContext(ctx, query)
	return err
}

func (s *Service) queryPublisherSummaries(ctx context.Context, clause string) ([]PublisherSummary, error) {
	querySQL := fmt.Sprintf(`
		SELECT 
			publisher_id, name, catalog_size, developer_count, genre_count,
			earliest_year, latest_year, valid_owners_count, ownership_coverage,
			catalog_owners_lower, catalog_owners_midpoint, catalog_owners_upper,
			top_title_name, top_title_midpoint, top1_share, top3_share, top5_share,
			top10_share, remaining_share, hhi, effective_title_count,
			free_title_count, free_title_share, median_price, avg_price,
			total_positive_reviews, total_negative_reviews, total_reviews,
			avg_positive_share, median_playtime, avg_peak_ccu, max_peak_ccu,
			dominant_genre, genre_concentration, avg_quality_score
		FROM publisher_summaries
		%s
	`, clause)

	rows, err := s.db.QueryContext(ctx, querySQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PublisherSummary
	for rows.Next() {
		var scan SummaryScan
		if err := scan.Scan(rows); err == nil {
			list = append(list, scan.ToSummary())
		}
	}
	return list, nil
}

type SummaryScan struct {
	PublisherID           string
	Name                  string
	CatalogSize           int64
	DeveloperCount        int64
	GenreCount            int64
	EarliestYear          sql.NullInt32
	LatestYear            sql.NullInt32
	ValidOwnersCount      int64
	OwnershipCoverage     float64
	CatalogOwnersLower    int64
	CatalogOwnersMidpoint int64
	CatalogOwnersUpper    int64
	TopTitleName          sql.NullString
	TopTitleMidpoint      sql.NullInt64
	Top1Share             float64
	Top3Share             float64
	Top5Share             float64
	Top10Share            float64
	RemainingShare        float64
	HHI                   float64
	EffectiveTitleCount   float64
	FreeTitleCount        int64
	FreeTitleShare        float64
	MedianPrice           float64
	AvgPrice              float64
	TotalPositiveReviews  int64
	TotalNegativeReviews  int64
	TotalReviews          int64
	AvgPositiveShare      float64
	MedianPlaytime        int64
	AvgPeakCCU            float64
	MaxPeakCCU            int64
	DominantGenre         sql.NullString
	GenreConcentration    float64
	AvgQualityScore       float64
}

func (s *SummaryScan) Scan(rows *sql.Rows) error {
	return rows.Scan(
		&s.PublisherID, &s.Name, &s.CatalogSize, &s.DeveloperCount, &s.GenreCount,
		&s.EarliestYear, &s.LatestYear, &s.ValidOwnersCount, &s.OwnershipCoverage,
		&s.CatalogOwnersLower, &s.CatalogOwnersMidpoint, &s.CatalogOwnersUpper,
		&s.TopTitleName, &s.TopTitleMidpoint, &s.Top1Share, &s.Top3Share, &s.Top5Share,
		&s.Top10Share, &s.RemainingShare, &s.HHI, &s.EffectiveTitleCount,
		&s.FreeTitleCount, &s.FreeTitleShare, &s.MedianPrice, &s.AvgPrice,
		&s.TotalPositiveReviews, &s.TotalNegativeReviews, &s.TotalReviews,
		&s.AvgPositiveShare, &s.MedianPlaytime, &s.AvgPeakCCU, &s.MaxPeakCCU,
		&s.DominantGenre, &s.GenreConcentration, &s.AvgQualityScore,
	)
}

func (s *SummaryScan) ToSummary() PublisherSummary {
	var ey, ly *int
	if s.EarliestYear.Valid {
		y := int(s.EarliestYear.Int32)
		ey = &y
	}
	if s.LatestYear.Valid {
		y := int(s.LatestYear.Int32)
		ly = &y
	}

	topTitle := ""
	if s.TopTitleName.Valid {
		topTitle = s.TopTitleName.String
	}
	topMid := int64(0)
	if s.TopTitleMidpoint.Valid {
		topMid = s.TopTitleMidpoint.Int64
	}
	domGenre := ""
	if s.DominantGenre.Valid {
		domGenre = s.DominantGenre.String
	}

	return PublisherSummary{
		PublisherID:           s.PublisherID,
		Name:                  s.Name,
		CatalogSize:           s.CatalogSize,
		DeveloperCount:        s.DeveloperCount,
		GenreCount:            s.GenreCount,
		EarliestYear:          ey,
		LatestYear:            ly,
		ValidOwnersCount:      s.ValidOwnersCount,
		OwnershipCoverage:     s.OwnershipCoverage,
		CatalogOwnersLower:    s.CatalogOwnersLower,
		CatalogOwnersMidpoint: s.CatalogOwnersMidpoint,
		CatalogOwnersUpper:    s.CatalogOwnersUpper,
		TopTitleName:          topTitle,
		TopTitleMidpoint:      topMid,
		Top1Share:             s.Top1Share,
		Top3Share:             s.Top3Share,
		Top5Share:             s.Top5Share,
		Top10Share:            s.Top10Share,
		RemainingShare:        s.RemainingShare,
		HHI:                   s.HHI,
		EffectiveTitleCount:   s.EffectiveTitleCount,
		FreeTitleCount:        s.FreeTitleCount,
		FreeTitleShare:        s.FreeTitleShare,
		MedianPrice:           s.MedianPrice,
		AvgPrice:              s.AvgPrice,
		TotalPositiveReviews:  s.TotalPositiveReviews,
		TotalNegativeReviews:  s.TotalNegativeReviews,
		TotalReviews:          s.TotalReviews,
		AvgPositiveShare:      s.AvgPositiveShare,
		MedianPlaytime:        s.MedianPlaytime,
		AvgPeakCCU:            s.AvgPeakCCU,
		MaxPeakCCU:            s.MaxPeakCCU,
		DominantGenre:         domGenre,
		GenreConcentration:    s.GenreConcentration,
		AvgQualityScore:       s.AvgQualityScore,
	}
}

// DeveloperItem represents a developer entity summary.
type DeveloperItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	GameCount int64  `json:"game_count"`
}

// ListDevelopers queries developers with optional name filter and pagination.
func (s *Service) ListDevelopers(ctx context.Context, query string, limit, offset int) ([]DeveloperItem, int64, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	whereClause := "WHERE alias_of_id IS NULL"
	var args []interface{}
	if query != "" {
		whereClause += " AND LOWER(name) LIKE ?"
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(query))+"%")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM developers %s", whereClause)
	var total int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, name, game_count
		FROM developers
		%s
		ORDER BY game_count DESC, name ASC
		LIMIT ? OFFSET ?
	`, whereClause)
	dataArgs := append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var developers []DeveloperItem
	for rows.Next() {
		var d DeveloperItem
		if err := rows.Scan(&d.ID, &d.Name, &d.GameCount); err != nil {
			return nil, 0, err
		}
		developers = append(developers, d)
	}

	return developers, total, nil
}
