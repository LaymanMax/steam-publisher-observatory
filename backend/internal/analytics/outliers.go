package analytics

import (
	"fmt"
	"math"
	"sort"
)

// OutlierItem represents an identified outlier title in a portfolio.
type OutlierItem struct {
	AppID        string   `json:"appid"`
	Name         string   `json:"name"`
	MetricName   string   `json:"metric_name"`
	MetricValue  float64  `json:"metric_value"`
	OutlierType  string   `json:"outlier_type"` // "high" or "low"
	Q1           float64  `json:"q1"`
	Median       float64  `json:"median"`
	Q3           float64  `json:"q3"`
	IQR          float64  `json:"iqr"`
	UpperFence   float64  `json:"upper_fence"`
	LowerFence   float64  `json:"lower_fence"`
	RankInMetric int      `json:"rank_in_metric"`
	TotalCount   int      `json:"total_count"`
	Genres       []string `json:"genres"`
	Developers   []string `json:"developers"`
}

// OutlierAnalysisReport contains outlier findings for a metric across a catalog.
type OutlierAnalysisReport struct {
	MetricName  string        `json:"metric_name"`
	SampleCount int           `json:"sample_count"`
	Q1          float64       `json:"q1"`
	Median      float64       `json:"median"`
	Q3          float64       `json:"q3"`
	IQR         float64       `json:"iqr"`
	UpperFence  float64       `json:"upper_fence"`
	LowerFence  float64       `json:"lower_fence"`
	Outliers    []OutlierItem `json:"outliers"`
	Methodology string        `json:"methodology"`
	HasOutliers bool          `json:"has_outliers"`
	Explanation string        `json:"explanation"`
}

// DataPoint encapsulates a game record value for outlier evaluation.
type DataPoint struct {
	AppID      string
	Name       string
	Value      float64
	IsValid    bool
	Genres     []string
	Developers []string
}

// DetectOutliersIQR implements standard Tukey IQR fence detection on valid numeric observations.
func DetectOutliersIQR(metricName string, points []DataPoint) *OutlierAnalysisReport {
	report := &OutlierAnalysisReport{
		MetricName:  metricName,
		Methodology: "Interquartile Range (IQR): Outliers lie above Q3 + 1.5*IQR or below Q1 - 1.5*IQR. Calculated solely on valid non-missing observations.",
	}

	var valid []DataPoint
	for _, p := range points {
		if p.IsValid && !math.IsNaN(p.Value) && !math.IsInf(p.Value, 0) {
			valid = append(valid, p)
		}
	}

	report.SampleCount = len(valid)

	// Need at least 4 observations to have meaningful quartiles
	if len(valid) < 4 {
		report.Explanation = fmt.Sprintf("Insufficient sample size (%d valid titles). Outlier detection requires at least 4 valid observations.", len(valid))
		return report
	}

	// Sort ascending
	sort.Slice(valid, func(i, j int) bool {
		return valid[i].Value < valid[j].Value
	})

	q1 := percentile(valid, 0.25)
	med := percentile(valid, 0.50)
	q3 := percentile(valid, 0.75)
	iqr := q3 - q1

	// If IQR is 0 (e.g. constant values or identical 0s), no IQR outliers can be meaningfully flagged
	if iqr == 0 {
		report.Q1 = q1
		report.Median = med
		report.Q3 = q3
		report.IQR = 0
		report.Explanation = fmt.Sprintf("Zero interquartile range (IQR=0). The central 50%% of titles have identical values (%.2f).", med)
		return report
	}

	upperFence := q3 + 1.5*iqr
	lowerFence := q1 - 1.5*iqr
	if lowerFence < 0 {
		lowerFence = 0 // negative boundaries not applicable for ownership/reviews/ccu
	}

	report.Q1 = q1
	report.Median = med
	report.Q3 = q3
	report.IQR = iqr
	report.UpperFence = upperFence
	report.LowerFence = lowerFence

	var outliers []OutlierItem
	n := len(valid)

	for i, p := range valid {
		rankDesc := n - i // 1 = highest value
		if p.Value > upperFence {
			outliers = append(outliers, OutlierItem{
				AppID:        p.AppID,
				Name:         p.Name,
				MetricName:   metricName,
				MetricValue:  p.Value,
				OutlierType:  "high",
				Q1:           q1,
				Median:       med,
				Q3:           q3,
				IQR:          iqr,
				UpperFence:   upperFence,
				LowerFence:   lowerFence,
				RankInMetric: rankDesc,
				TotalCount:   n,
				Genres:       p.Genres,
				Developers:   p.Developers,
			})
		} else if p.Value < lowerFence && p.Value > 0 {
			outliers = append(outliers, OutlierItem{
				AppID:        p.AppID,
				Name:         p.Name,
				MetricName:   metricName,
				MetricValue:  p.Value,
				OutlierType:  "low",
				Q1:           q1,
				Median:       med,
				Q3:           q3,
				IQR:          iqr,
				UpperFence:   upperFence,
				LowerFence:   lowerFence,
				RankInMetric: rankDesc,
				TotalCount:   n,
				Genres:       p.Genres,
				Developers:   p.Developers,
			})
		}
	}

	// Sort outliers by magnitude descending
	sort.Slice(outliers, func(i, j int) bool {
		return outliers[i].MetricValue > outliers[j].MetricValue
	})

	report.Outliers = outliers
	report.HasOutliers = len(outliers) > 0

	if len(outliers) > 0 {
		report.Explanation = fmt.Sprintf(
			"Detected %d statistical outliers (upper fence: %.2f, IQR: %.2f). The top outlier is %s with value %.2f.",
			len(outliers), upperFence, iqr, outliers[0].Name, outliers[0].MetricValue,
		)
	} else {
		report.Explanation = fmt.Sprintf(
			"No statistical outliers detected. All %d evaluated titles fall within the normal IQR range (%.2f to %.2f).",
			n, lowerFence, upperFence,
		)
	}

	return report
}

func percentile(sortedData []DataPoint, p float64) float64 {
	n := len(sortedData)
	if n == 0 {
		return 0
	}
	if n == 1 {
		return sortedData[0].Value
	}

	pos := p * float64(n-1)
	low := int(math.Floor(pos))
	high := int(math.Ceil(pos))

	if low == high {
		return sortedData[low].Value
	}

	weight := pos - float64(low)
	return sortedData[low].Value*(1.0-weight) + sortedData[high].Value*weight
}
