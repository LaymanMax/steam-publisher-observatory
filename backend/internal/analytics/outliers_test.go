package analytics

import (
	"testing"
)

func TestDetectOutliersIQR(t *testing.T) {
	// 8 games with normal distribution around 10k-20k, and 1 massive blockbuster at 1,000,000
	points := []DataPoint{
		{AppID: "1", Name: "Small 1", Value: 10000, IsValid: true},
		{AppID: "2", Name: "Small 2", Value: 12000, IsValid: true},
		{AppID: "3", Name: "Small 3", Value: 14000, IsValid: true},
		{AppID: "4", Name: "Small 4", Value: 15000, IsValid: true},
		{AppID: "5", Name: "Small 5", Value: 16000, IsValid: true},
		{AppID: "6", Name: "Small 6", Value: 18000, IsValid: true},
		{AppID: "7", Name: "Small 7", Value: 20000, IsValid: true},
		{AppID: "8", Name: "Blockbuster", Value: 1000000, IsValid: true}, // Massive outlier
		{AppID: "9", Name: "Invalid Game", Value: 0, IsValid: false},     // Must be excluded!
	}

	report := DetectOutliersIQR("owners_midpoint", points)

	if report.SampleCount != 8 {
		t.Fatalf("expected 8 valid samples, got %d", report.SampleCount)
	}
	if !report.HasOutliers {
		t.Fatalf("expected outliers to be detected")
	}
	if len(report.Outliers) != 1 {
		t.Fatalf("expected exactly 1 outlier, got %d", len(report.Outliers))
	}
	if report.Outliers[0].AppID != "8" || report.Outliers[0].Name != "Blockbuster" {
		t.Errorf("expected Blockbuster to be flagged, got %v", report.Outliers[0])
	}
}

func TestDetectOutliersIQR_SmallSample(t *testing.T) {
	points := []DataPoint{
		{AppID: "1", Name: "Game 1", Value: 100, IsValid: true},
		{AppID: "2", Name: "Game 2", Value: 200, IsValid: true},
	}
	report := DetectOutliersIQR("price", points)
	if report.HasOutliers {
		t.Errorf("small sample (<4) must not produce outliers")
	}
}

func TestDetectOutliersIQR_ConstantValues(t *testing.T) {
	points := []DataPoint{
		{AppID: "1", Name: "Game 1", Value: 500, IsValid: true},
		{AppID: "2", Name: "Game 2", Value: 500, IsValid: true},
		{AppID: "3", Name: "Game 3", Value: 500, IsValid: true},
		{AppID: "4", Name: "Game 4", Value: 500, IsValid: true},
	}
	report := DetectOutliersIQR("price", points)
	if report.HasOutliers {
		t.Errorf("constant values (IQR=0) must not produce outliers")
	}
}
