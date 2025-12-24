package analyzer

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// Exporter handles data export functionality
type Exporter struct{}

// NewExporter creates a new exporter instance
func NewExporter() *Exporter {
	return &Exporter{}
}

// ExportJSON exports analysis results to JSON format
func (e *Exporter) ExportJSON(results []*models.AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	return encoder.Encode(results)
}

// ExportCSV exports analysis results to CSV format
func (e *Exporter) ExportCSV(results []*models.AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{
		"domain", "jli_score", "jli_level", "cdn_provider", 
		"first_seen", "last_seen", "cluster_id", "total_signals",
		"ux_signals", "payment_signals", "infra_signals", "dns_signals", "cdn_signals",
	}
	
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %v", err)
	}

	// Write data rows
	for _, result := range results {
		clusterID := ""
		if result.Domain.ClusterID != nil {
			clusterID = *result.Domain.ClusterID
		}
		
		row := []string{
			result.Domain.Domain,
			fmt.Sprintf("%.3f", result.JLIScore),
			result.JLILevel,
			result.Domain.CDNProvider,
			result.Domain.FirstSeen.Format(time.RFC3339),
			result.Domain.LastSeen.Format(time.RFC3339),
			clusterID,
			fmt.Sprintf("%d", len(result.Domain.Signals)),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "UX")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "PAYMENT")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "INFRA")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "DNS")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "CDN")),
		}
		
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %v", err)
		}
	}

	return nil
}

// ExportToDatabase exports results to a database (placeholder)
func (e *Exporter) ExportToDatabase(results []*models.AnalysisResult, dbPath string) error {
	// In a real implementation, this would connect to a database
	// and export the results to structured tables
	return fmt.Errorf("database export not implemented yet")
}

// countSignalsByCategory counts signals in a specific category
func countSignalsByCategory(signals []models.Signal, category string) int {
	count := 0
	for _, signal := range signals {
		if signal.Category == category {
			count++
		}
	}
	return count
}

// FilterResultsByTime filters results by a time range
func (e *Exporter) FilterResultsByTime(results []*models.AnalysisResult, startTime, endTime time.Time) []*models.AnalysisResult {
	var filteredResults []*models.AnalysisResult
	
	for _, result := range results {
		if result.Domain.LastSeen.After(startTime) && result.Domain.LastSeen.Before(endTime) {
			filteredResults = append(filteredResults, result)
		}
	}
	
	return filteredResults
}

// FilterResultsByJLIScore filters results by JLI score range
func (e *Exporter) FilterResultsByJLIScore(results []*models.AnalysisResult, minScore, maxScore float64) []*models.AnalysisResult {
	var filteredResults []*models.AnalysisResult
	
	for _, result := range results {
		if result.JLIScore >= minScore && result.JLIScore <= maxScore {
			filteredResults = append(filteredResults, result)
		}
	}
	
	return filteredResults
}

// FilterResultsByCDNProvider filters results by CDN provider
func (e *Exporter) FilterResultsByCDNProvider(results []*models.AnalysisResult, provider string) []*models.AnalysisResult {
	var filteredResults []*models.AnalysisResult
	
	for _, result := range results {
		if result.Domain.CDNProvider == provider {
			filteredResults = append(filteredResults, result)
		}
	}
	
	return filteredResults
}

// ExportSummary exports a summary of results
func (e *Exporter) ExportSummary(results []*models.AnalysisResult, filename string) error {
	summary := e.GenerateSummary(results)
	
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create summary file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	return encoder.Encode(summary)
}

// GenerateSummary generates a summary of analysis results
func (e *Exporter) GenerateSummary(results []*models.AnalysisResult) map[string]interface{} {
	summary := make(map[string]interface{})
	
	if len(results) == 0 {
		summary["total_domains"] = 0
		return summary
	}
	
	totalDomains := len(results)
	highRisk := 0
	mediumRisk := 0
	lowRisk := 0
	
	// Count risk levels
	for _, result := range results {
		switch result.JLILevel {
		case "HIGH":
			highRisk++
		case "MEDIUM":
			mediumRisk++
		case "LOW":
			lowRisk++
		}
	}
	
	// Calculate average JLI score
	totalScore := 0.0
	for _, result := range results {
		totalScore += result.JLIScore
	}
	avgScore := totalScore / float64(totalDomains)
	
	// Count CDN providers
	cdnCount := make(map[string]int)
	for _, result := range results {
		cdnCount[result.Domain.CDNProvider]++
	}
	
	// Count signal categories
	categoryCount := make(map[string]int)
	for _, result := range results {
		for _, signal := range result.Domain.Signals {
			categoryCount[signal.Category]++
		}
	}
	
	summary["total_domains"] = totalDomains
	summary["high_risk_domains"] = highRisk
	summary["medium_risk_domains"] = mediumRisk
	summary["low_risk_domains"] = lowRisk
	summary["high_risk_percentage"] = float64(highRisk) / float64(totalDomains) * 100
	summary["average_jli_score"] = avgScore
	summary["cdn_distribution"] = cdnCount
	summary["signal_category_distribution"] = categoryCount
	
	return summary
}

// ParseTimeRange parses a time range string (e.g., "30d", "7d", "1h")
func ParseTimeRange(timeStr string) (time.Time, time.Time, error) {
	// Parse the time range string
	var duration time.Duration
	var err error
	
	if len(timeStr) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time range format")
	}
	
	numStr := timeStr[:len(timeStr)-1]
	unit := timeStr[len(timeStr)-1:]
	
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid number in time range: %v", err)
	}
	
	switch unit {
	case "d":
		duration = time.Duration(num) * 24 * time.Hour
	case "h":
		duration = time.Duration(num) * time.Hour
	case "m":
		duration = time.Duration(num) * time.Minute
	case "s":
		duration = time.Duration(num) * time.Second
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time unit: %s", unit)
	}
	
	endTime := time.Now()
	startTime := endTime.Add(-duration)
	
	return startTime, endTime, nil
}