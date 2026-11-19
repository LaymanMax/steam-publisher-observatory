package tests

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/api"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/config"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/database"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/importer"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/jobs"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/normalization"
)

func setupTestEnvironment(t *testing.T) (*database.DB, *analytics.Service, *api.Server) {
	// In-memory DuckDB for isolated, fast integration testing
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	cfg := &config.Config{
		Port:         8080,
		DataDir:      t.TempDir(),
		DBPath:       ":memory:",
		CORSAllowAll: true,
		PriceTiers:   config.DefaultPriceTiers,
	}

	imp := importer.NewImporter(db, cfg)
	jobMgr := jobs.NewManager(imp)
	svc := analytics.NewService(db, cfg)
	server := api.NewServer(cfg, svc, imp, jobMgr, nil)

	return db, svc, server
}

func seedTestFixtures(t *testing.T, db *database.DB) {
	ctx := context.Background()

	// 1. Insert Publishers
	// Pub 1: Highly Concentrated ("Titans Publishing")
	// Pub 2: Diversified Portfolio ("Indie Collective")
	// Pub 3: Single Game Publisher ("Mono Publisher")
	pubQueries := []string{
		`INSERT INTO publishers (id, name, normalized_key, game_count) VALUES ('pub_conc', 'Titans Publishing', 'titans publishing', 15)`,
		`INSERT INTO publishers (id, name, normalized_key, game_count) VALUES ('pub_div', 'Indie Collective', 'indie collective', 10)`,
		`INSERT INTO publishers (id, name, normalized_key, game_count) VALUES ('pub_mono', 'Mono Publisher', 'mono publisher', 1)`,
		`INSERT INTO publishers (id, name, normalized_key, game_count) VALUES ('pub_shared', 'Shared Partner Publishing', 'shared partner publishing', 2)`,
	}
	for _, q := range pubQueries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("failed to insert test publisher: %v", err)
		}
	}

	// 2. Insert Developers
	// Dev 1: "Firaxis Studio"
	// Dev 2: "Shared Dev Studio" (collaborates with Titans AND Shared Partner)
	devQueries := []string{
		`INSERT INTO developers (id, name, normalized_key, game_count) VALUES ('dev_1', 'Titan Dev', 'titan dev', 10)`,
		`INSERT INTO developers (id, name, normalized_key, game_count) VALUES ('dev_shared', 'Shared Dev Studio', 'shared dev studio', 5)`,
		`INSERT INTO developers (id, name, normalized_key, game_count) VALUES ('dev_indie', 'Indie Devs', 'indie devs', 10)`,
	}
	for _, q := range devQueries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("failed to insert test developer: %v", err)
		}
	}

	// 3. Insert Games with designed ownership midpoints for Titans Publishing:
	// Top 5 midpoints: 500,000 + 100,000 + 50,000 + 40,000 + 30,000 = 720,000 (72.0%!)
	// Remaining 10 games: 28,000 each = 280,000 (28.0%)
	// Total = 1,000,000.
	gameInsertQueries := []string{
		// Titans top 5
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('101', 'Titan Hit 1', '2024-01-01', 2024, TRUE, 59.99, 'High-priced', 400000, 600000, 500000, TRUE, 10000, 1000, 11000, 90.9, 100.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('102', 'Titan Hit 2', '2023-01-01', 2023, TRUE, 29.99, 'Premium', 80000, 120000, 100000, TRUE, 5000, 500, 5500, 90.9, 100.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('103', 'Titan Hit 3', '2022-01-01', 2022, TRUE, 19.99, 'Mid-range', 40000, 60000, 50000, TRUE, 2000, 200, 2200, 90.9, 100.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('104', 'Titan Hit 4', '2021-01-01', 2021, TRUE, 14.99, 'Mid-range', 30000, 50000, 40000, TRUE, 1500, 150, 1650, 90.9, 100.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('105', 'Titan Hit 5', '2020-01-01', 2020, TRUE, 9.99, 'Low', 20000, 40000, 30000, TRUE, 1000, 100, 1100, 90.9, 100.0)`,

		// Titans remaining 10 titles (midpoint 28,000 each)
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('106', 'Titan Tail 6', '2020-01-01', 2020, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('107', 'Titan Tail 7', '2019-01-01', 2019, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('108', 'Titan Tail 8', '2019-01-01', 2019, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('109', 'Titan Tail 9', '2018-01-01', 2018, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('110', 'Titan Tail 10', '2018-01-01', 2018, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('111', 'Titan Tail 11', '2017-01-01', 2017, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('112', 'Titan Tail 12', '2017-01-01', 2017, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('113', 'Titan Tail 13', '2016-01-01', 2016, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('114', 'Titan Tail 14', '2016-01-01', 2016, TRUE, 4.99, 'Very Low', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('115', 'Titan Tail 15', '2015-01-01', 2015, TRUE, 0.0, 'Free', 20000, 36000, 28000, TRUE, 100, 10, 110, 90.9, 90.0)`,

		// Shared game between Titans and Shared Partner
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, owners_lower, owners_upper, owners_midpoint, has_valid_owners, positive, negative, total_reviews, positive_share, quality_score)
		 VALUES ('201', 'Co-Published Adventure', '2023-05-01', 2023, TRUE, 19.99, 'Mid-range', 50000, 100000, 75000, TRUE, 2000, 200, 2200, 90.9, 100.0)`,

		// Playtest and sparse games
		`INSERT INTO games (appid, name, is_playtest, is_sparse, quality_score, quality_flags)
		 VALUES ('301', 'Black Dragon Mage Playtest', TRUE, TRUE, 15.0, 'suspected_playtest,missing_publisher,missing_developer,missing_owners')`,

		// Missing ownership game
		`INSERT INTO games (appid, name, release_date, release_year, has_valid_date, price, price_tier, has_valid_owners, quality_score)
		 VALUES ('401', 'Unknown Horizon', '2024-06-01', 2024, TRUE, 9.99, 'Low', FALSE, 60.0)`,
	}
	for _, q := range gameInsertQueries {
		if _, err := db.ExecContext(ctx, q); err != nil {
			t.Fatalf("failed to insert test game: %v", err)
		}
	}

	// 4. Map game_publishers
	for i := 101; i <= 115; i++ {
		db.ExecContext(ctx, "INSERT INTO game_publishers VALUES (?, 'pub_conc')", string(rune('0'+i/100))+string(rune('0'+(i%100)/10))+string(rune('0'+i%10)))
	}
	// Shared game mapped to BOTH publishers!
	db.ExecContext(ctx, "INSERT INTO game_publishers VALUES ('201', 'pub_conc')")
	db.ExecContext(ctx, "INSERT INTO game_publishers VALUES ('201', 'pub_shared')")
	db.ExecContext(ctx, "INSERT INTO game_publishers VALUES ('401', 'pub_mono')")

	// 5. Map game_developers
	for i := 101; i <= 115; i++ {
		db.ExecContext(ctx, "INSERT INTO game_developers VALUES (?, 'dev_1')", string(rune('0'+i/100))+string(rune('0'+(i%100)/10))+string(rune('0'+i%10)))
	}
	db.ExecContext(ctx, "INSERT INTO game_developers VALUES ('201', 'dev_shared')")

	// 6. Map genres
	db.ExecContext(ctx, "INSERT INTO game_genres VALUES ('101', 'Action')")
	db.ExecContext(ctx, "INSERT INTO game_genres VALUES ('101', 'RPG')")
	db.ExecContext(ctx, "INSERT INTO game_genres VALUES ('102', 'Strategy')")

	// Materialize summary tables
	imp := importer.NewImporter(db, &config.Config{})
	if err := imp.MaterializePublisherSummaries(ctx); err != nil {
		t.Fatalf("failed to materialize publisher summaries: %v", err)
	}
}

func TestIntegration_DependencyReportExactCalculation(t *testing.T) {
	db, svc, _ := setupTestEnvironment(t)
	seedTestFixtures(t, db)

	dep, err := svc.GetPublisherDependency(context.Background(), "pub_conc")
	if err != nil {
		t.Fatalf("failed to get dependency report: %v", err)
	}

	// 101..115 (1,000,000) + 201 (75,000) = 1,075,000
	// Top 5 titles: 101 (500k), 102 (100k), 201 (75k), 103 (50k), 104 (40k)
	// Top 5 sum = 765,000
	// Share = 765,000 / 1,075,000 = 71.16%
	if dep.CatalogOwnersMidpoint != 1075000 {
		t.Errorf("CatalogOwnersMidpoint: expected 1,075,000, got %d", dep.CatalogOwnersMidpoint)
	}
	if math.Abs(dep.Top5Share-71.16) > 0.1 {
		t.Errorf("Top5Share: expected ~71.16%%, got %.2f%%", dep.Top5Share)
	}
	if dep.ConcentrationLevel != "High" {
		t.Errorf("ConcentrationLevel: expected High, got %s", dep.ConcentrationLevel)
	}
	if dep.Top1Share < 45.0 || dep.Top1Share > 47.0 {
		t.Errorf("Top1Share: expected ~46.5%%, got %.2f%%", dep.Top1Share)
	}
}

func TestIntegration_SharedGameMultiPublishers(t *testing.T) {
	db, svc, _ := setupTestEnvironment(t)
	seedTestFixtures(t, db)

	// Game 201 must appear in BOTH 'pub_conc' and 'pub_shared' catalog!
	cat1, total1, err := svc.GetPublisherCatalog(context.Background(), "pub_conc", analytics.CatalogFilter{})
	if err != nil || total1 == 0 {
		t.Fatalf("failed to get pub_conc catalog")
	}

	cat2, total2, err := svc.GetPublisherCatalog(context.Background(), "pub_shared", analytics.CatalogFilter{})
	if err != nil || total2 == 0 {
		t.Fatalf("failed to get pub_shared catalog")
	}

	foundIn1, foundIn2 := false, false
	for _, g := range cat1 {
		if g.AppID == "201" {
			foundIn1 = true
			break
		}
	}
	for _, g := range cat2 {
		if g.AppID == "201" {
			foundIn2 = true
			break
		}
	}

	if !foundIn1 || !foundIn2 {
		t.Errorf("expected game 201 to be credited to BOTH publishers: foundIn1=%v, foundIn2=%v", foundIn1, foundIn2)
	}
}

func TestIntegration_APIEndpoints(t *testing.T) {
	db, _, server := setupTestEnvironment(t)
	seedTestFixtures(t, db)

	handler := server.Router()

	endpoints := []string{
		"/api/health",
		"/api/overview",
		"/api/dataset/status",
		"/api/dataset/quality",
		"/api/publishers",
		"/api/publishers/pub_conc",
		"/api/publishers/pub_conc/summary",
		"/api/publishers/pub_conc/catalog",
		"/api/publishers/pub_conc/dependency",
		"/api/publishers/pub_conc/genres",
		"/api/publishers/pub_conc/prices",
		"/api/publishers/pub_conc/developers",
		"/api/publishers/pub_conc/timeline",
		"/api/publishers/pub_conc/outliers",
		"/api/publishers/pub_conc/quality",
		"/api/publishers/pub_conc/report",
		"/api/compare?publishers=pub_conc,pub_shared",
		"/api/entities/publishers",
		"/api/entities/developers",
		"/api/entities/aliases",
		"/api/entities/potential-aliases",
		"/api/exports/publishers.csv",
		"/api/exports/catalog.csv?publisher_id=pub_conc",
		"/api/exports/report.json?publisher_id=pub_conc",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(http.MethodGet, ep, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET %s returned HTTP %d, body: %s", ep, rec.Code, rec.Body.String())
		}
	}
}

func TestIntegration_EntityAliasingRecalculation(t *testing.T) {
	db, svc, _ := setupTestEnvironment(t)
	seedTestFixtures(t, db)

	// Clean display name check
	cleaned := normalization.CleanName("  Valve   Corporation  ")
	if cleaned != "Valve Corporation" {
		t.Errorf("CleanName failed: %s", cleaned)
	}

	// Create alias: Merge 'pub_shared' into 'pub_conc'
	err := svc.CreateAlias(context.Background(), "publisher", "Shared Partner Publishing", "pub_conc")
	if err != nil {
		t.Fatalf("CreateAlias failed: %v", err)
	}

	aliases, err := svc.ListAliases(context.Background())
	if err != nil || len(aliases) == 0 {
		t.Fatalf("expected at least 1 alias recorded")
	}

	// Verify alias removal
	err = svc.RemoveAlias(context.Background(), aliases[0].ID)
	if err != nil {
		t.Fatalf("RemoveAlias failed: %v", err)
	}
}

func TestIntegration_JSONReportStructure(t *testing.T) {
	db, _, server := setupTestEnvironment(t)
	seedTestFixtures(t, db)

	handler := server.Router()
	req := httptest.NewRequest(http.MethodGet, "/api/publishers/pub_conc/report", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("failed to get report: %d", rec.Code)
	}

	var rep map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &rep); err != nil {
		t.Fatalf("failed to parse JSON report: %v", err)
	}

	if rep["dependency_report"] == nil {
		t.Errorf("expected dependency_report in JSON export")
	}
	if rep["metric_definitions"] == nil {
		t.Errorf("expected metric_definitions in JSON export")
	}
}
