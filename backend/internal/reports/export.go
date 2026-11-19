package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"time"

	"github.com/laymanmax/steam-publisher-observatory/backend/internal/analytics"
	"github.com/laymanmax/steam-publisher-observatory/backend/internal/metrics"
)

// PublisherFullReport is the comprehensive analytical document for export in JSON.
type PublisherFullReport struct {
	Metadata struct {
		ReportGeneratedAt time.Time `json:"report_generated_at"`
		Product           string    `json:"product"`
		Version           string    `json:"version"`
		Disclaimer        string    `json:"disclaimer"`
	} `json:"metadata"`
	Publisher   analytics.PublisherSummary         `json:"publisher"`
	Dependency  *metrics.PortfolioDependencyReport `json:"dependency_report"`
	Genres      *metrics.GenreAnalysisReport       `json:"genre_analysis"`
	Pricing     *metrics.PricingReport             `json:"pricing_analysis"`
	Developers  *analytics.DeveloperAnalysisReport `json:"developer_relationships"`
	Timeline    *analytics.TimelineReport          `json:"timeline"`
	DataQuality *analytics.PublisherQualityReport  `json:"data_quality"`
	Definitions map[string]string                  `json:"metric_definitions"`
}

// GeneratePublisherJSONReport builds the full analytical report in JSON format.
func GeneratePublisherJSONReport(ctx context.Context, svc *analytics.Service, publisherID string) ([]byte, error) {
	summary, err := svc.GetPublisherSummary(ctx, publisherID)
	if err != nil {
		return nil, err
	}

	dep, _ := svc.GetPublisherDependency(ctx, publisherID)
	genres, _ := svc.GetPublisherGenres(ctx, publisherID)
	pricing, _ := svc.GetPublisherPrices(ctx, publisherID)
	devs, _ := svc.GetPublisherDevelopers(ctx, publisherID)
	timeline, _ := svc.GetPublisherTimeline(ctx, publisherID)
	quality, _ := svc.GetPublisherQuality(ctx, publisherID)

	rep := PublisherFullReport{
		Publisher:   *summary,
		Dependency:  dep,
		Genres:      genres,
		Pricing:     pricing,
		Developers:  devs,
		Timeline:    timeline,
		DataQuality: quality,
	}

	rep.Metadata.ReportGeneratedAt = time.Now()
	rep.Metadata.Product = "Steam Publisher Observatory"
	rep.Metadata.Version = "1.0.0"
	rep.Metadata.Disclaimer = "Analytical estimates are derived from public dataset ranges and credit metadata. They do not constitute actual revenues, sales figures, profit, or legal publishing contracts."

	rep.Definitions = map[string]string{
		"catalog_owners_midpoint": "Sum of per-game midpoint estimates (lower + upper)/2 for all games with valid range estimates. Not a unique player count.",
		"top5_share":              "Percentage of catalog ownership midpoint concentrated in the top 5 largest games.",
		"hhi":                     "Herfindahl-Hirschman Index scaled between 0 and 10,000, measuring ownership concentration across catalog titles.",
		"effective_title_count":   "Reciprocal of unscaled HHI (1 / unscaled_HHI), representing the equivalent number of equally-sized titles in the portfolio.",
		"ownership_coverage":      "Percentage of games in the publisher's catalog that have parseable ownership estimates.",
		"developer_relationships": "Observed game credits associating studios with published titles, not verified legal publishing agreements.",
	}

	return json.MarshalIndent(rep, "", "  ")
}

// ExportPublishersCSV serializes publisher directory records to CSV format.
func ExportPublishersCSV(publishers []analytics.PublisherSummary) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"Publisher ID", "Publisher Name", "Catalog Size", "Developer Count", "Genre Count",
		"Earliest Year", "Latest Year", "Ownership Coverage (%)",
		"Catalog Owners Lower", "Catalog Owners Midpoint", "Catalog Owners Upper",
		"Top Title Name", "Top Title Midpoint", "Top 1 Share (%)", "Top 3 Share (%)",
		"Top 5 Share (%)", "Top 10 Share (%)", "HHI Concentration", "Effective Title Count",
		"Free Title Share (%)", "Median Price (USD)", "Avg Price (USD)",
		"Total Reviews", "Positive Review Share (%)", "Median Playtime (min)",
		"Avg Peak CCU", "Data Completeness Score (%)",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, p := range publishers {
		eyStr := ""
		if p.EarliestYear != nil {
			eyStr = fmt.Sprintf("%d", *p.EarliestYear)
		}
		lyStr := ""
		if p.LatestYear != nil {
			lyStr = fmt.Sprintf("%d", *p.LatestYear)
		}

		row := []string{
			p.PublisherID,
			p.Name,
			fmt.Sprintf("%d", p.CatalogSize),
			fmt.Sprintf("%d", p.DeveloperCount),
			fmt.Sprintf("%d", p.GenreCount),
			eyStr,
			lyStr,
			fmt.Sprintf("%.1f", p.OwnershipCoverage),
			fmt.Sprintf("%d", p.CatalogOwnersLower),
			fmt.Sprintf("%d", p.CatalogOwnersMidpoint),
			fmt.Sprintf("%d", p.CatalogOwnersUpper),
			p.TopTitleName,
			fmt.Sprintf("%d", p.TopTitleMidpoint),
			fmt.Sprintf("%.1f", p.Top1Share),
			fmt.Sprintf("%.1f", p.Top3Share),
			fmt.Sprintf("%.1f", p.Top5Share),
			fmt.Sprintf("%.1f", p.Top10Share),
			fmt.Sprintf("%.1f", p.HHI),
			fmt.Sprintf("%.2f", p.EffectiveTitleCount),
			fmt.Sprintf("%.1f", p.FreeTitleShare),
			fmt.Sprintf("%.2f", p.MedianPrice),
			fmt.Sprintf("%.2f", p.AvgPrice),
			fmt.Sprintf("%d", p.TotalReviews),
			fmt.Sprintf("%.1f", p.AvgPositiveShare),
			fmt.Sprintf("%d", p.MedianPlaytime),
			fmt.Sprintf("%.1f", p.AvgPeakCCU),
			fmt.Sprintf("%.1f", p.AvgQualityScore),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// ExportCatalogCSV serializes games in a publisher's catalog to CSV format.
func ExportCatalogCSV(games []analytics.GameCatalogItem) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	headers := []string{
		"AppID", "Game Name", "Developers", "Genres", "Release Date", "Release Year",
		"Price (USD)", "Price Tier", "Estimated Owners Raw",
		"Owners Lower", "Owners Midpoint", "Owners Upper",
		"Positive Reviews", "Negative Reviews", "Total Reviews", "Positive Share (%)",
		"Peak CCU", "Median Playtime Forever (min)", "Data Completeness Score (%)",
		"Is Playtest", "Is Sparse",
	}
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, g := range games {
		priceStr := ""
		if g.Price != nil {
			priceStr = fmt.Sprintf("%.2f", *g.Price)
		}
		yearStr := ""
		if g.ReleaseYear != nil {
			yearStr = fmt.Sprintf("%d", *g.ReleaseYear)
		}
		lowerStr := ""
		if g.OwnersLower != nil {
			lowerStr = fmt.Sprintf("%d", *g.OwnersLower)
		}
		midStr := ""
		if g.OwnersMidpoint != nil {
			midStr = fmt.Sprintf("%d", *g.OwnersMidpoint)
		}
		upperStr := ""
		if g.OwnersUpper != nil {
			upperStr = fmt.Sprintf("%d", *g.OwnersUpper)
		}
		posShareStr := ""
		if g.PositiveShare != nil {
			posShareStr = fmt.Sprintf("%.1f", *g.PositiveShare)
		}

		row := []string{
			g.AppID,
			g.Name,
			fmt.Sprintf("%v", g.Developers),
			fmt.Sprintf("%v", g.Genres),
			g.ReleaseDate,
			yearStr,
			priceStr,
			g.PriceTier,
			g.OwnersRaw,
			lowerStr,
			midStr,
			upperStr,
			fmt.Sprintf("%d", g.Positive),
			fmt.Sprintf("%d", g.Negative),
			fmt.Sprintf("%d", g.TotalReviews),
			posShareStr,
			fmt.Sprintf("%d", g.PeakCCU),
			fmt.Sprintf("%d", g.MedianPlaytime),
			fmt.Sprintf("%.1f", g.QualityScore),
			fmt.Sprintf("%v", g.IsPlaytest),
			fmt.Sprintf("%v", g.IsSparse),
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}
