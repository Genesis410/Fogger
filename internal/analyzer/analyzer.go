package analyzer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/genesis410/fogger/internal/config"
	"github.com/genesis410/fogger/internal/models"
	"github.com/genesis410/fogger/internal/scanner"
)

// AnalyzeDomain performs a complete analysis of a domain
func AnalyzeDomain(domain string, timeout time.Duration, profile string) *models.AnalysisResult {
	// Get configuration
	cfg := config.Get()

	// Perform scanning
	scanResult := scanner.ScanDomain(domain, timeout)

	// Perform behavioral analysis
	behavioralAnalyzer := NewBehavioralAnalyzer()
	behavioralSignals := behavioralAnalyzer.AnalyzeContent(scanResult.Body)

	// Also analyze DOM structure
	domSignals := behavioralAnalyzer.AnalyzeDOMStructure(scanResult.Body)

	// Combine all signals
	allSignals := append(scanResult.Signals, behavioralSignals...)
	allSignals = append(allSignals, domSignals...)

	// Calculate JLI score
	categoryScores := calculateCategoryScoresWithSignals(allSignals)
	jliScore := calculateEnhancedJLIScore(categoryScores, cfg.Scoring, allSignals)
	jliLevel := classifyJLILevel(jliScore, cfg.Threshold)

	// Create domain model
	domainModel := models.Domain{
		Domain:      domain,
		FirstSeen:   time.Now(),
		LastSeen:    time.Now(),
		CDNProvider: scanResult.CDNProvider,
		JLIScore:    jliScore,
		JLILevel:    jliLevel,
		Signals:     allSignals,
	}

	// Create category breakdown
	categoryBreakdown := make(map[string]models.CategoryBreakdown)
	for category, score := range categoryScores {
		var weight float64
		switch category {
		case "UX":
			weight = cfg.Scoring.GamblingUI
		case "PAYMENT":
			weight = cfg.Scoring.PaymentSignal
		case "INFRA":
			weight = cfg.Scoring.InfraCorrelation
		case "DNS":
			weight = cfg.Scoring.DomainChurn
		case "CDN":
			weight = cfg.Scoring.CDNPattern
		}
		categoryBreakdown[category] = models.CategoryBreakdown{
			Score:        score,
			Weight:       weight,
			Contribution: score * weight,
		}
	}

	// Create analysis result
	result := &models.AnalysisResult{
		Domain:            domainModel,
		JLIScore:          jliScore,
		JLILevel:          jliLevel,
		CategoryBreakdown: categoryBreakdown,
		ProfileUsed:       profile,
	}

	return result
}

// calculateCategoryScores calculates scores for each category
func calculateCategoryScores(scanResult *scanner.ScanResult) map[string]float64 {
	categoryScores := make(map[string]float64)

	for _, signal := range scanResult.Signals {
		currentScore, exists := categoryScores[signal.Category]
		if !exists {
			currentScore = 0.0
		}
		// Use max score for category (not sum to prevent spamming)
		if signal.Confidence > currentScore {
			categoryScores[signal.Category] = signal.Confidence
		}
	}

	return categoryScores
}

// calculateCategoryScoresWithSignals calculates scores for each category from a slice of signals
func calculateCategoryScoresWithSignals(signals []models.Signal) map[string]float64 {
	categoryScores := make(map[string]float64)

	for _, signal := range signals {
		currentScore, exists := categoryScores[signal.Category]
		if !exists {
			currentScore = 0.0
		}
		// Use max score for category (not sum to prevent spamming)
		if signal.Confidence > currentScore {
			categoryScores[signal.Category] = signal.Confidence
		}
	}

	return categoryScores
}

// calculateJLIScore calculates the Judol Likelihood Index score
func calculateJLIScore(categoryScores map[string]float64, weights config.ScoringConfig) float64 {
	// Calculate weighted sum
	jliRaw := 0.0
	jliRaw += categoryScores["UX"] * weights.GamblingUI
	jliRaw += categoryScores["PAYMENT"] * weights.PaymentSignal
	jliRaw += categoryScores["INFRA"] * weights.InfraCorrelation
	jliRaw += categoryScores["DNS"] * weights.DomainChurn
	jliRaw += categoryScores["CDN"] * weights.CDNPattern

	// Apply confidence damping
	confidenceFactor := calculateConfidenceFactor(categoryScores)
	jliScore := jliRaw * confidenceFactor

	// Ensure score is between 0 and 1
	if jliScore > 1.0 {
		jliScore = 1.0
	}
	if jliScore < 0.0 {
		jliScore = 0.0
	}

	return jliScore
}

// Enhanced JLI calculation with additional factors
func calculateEnhancedJLIScore(categoryScores map[string]float64, weights config.ScoringConfig, signals []models.Signal) float64 {
	// Start with basic calculation
	jliBase := calculateJLIScore(categoryScores, weights)

	// Apply additional factors based on signal patterns
	signalFactor := calculateSignalFactor(signals)

	// Apply temporal factors if available
	temporalFactor := calculateTemporalFactor()

	// Combine factors
	enhancedScore := jliBase * signalFactor * temporalFactor

	// Ensure score is between 0 and 1
	if enhancedScore > 1.0 {
		enhancedScore = 1.0
	}
	if enhancedScore < 0.0 {
		enhancedScore = 0.0
	}

	return enhancedScore
}

// calculateSignalFactor adjusts score based on signal patterns
func calculateSignalFactor(signals []models.Signal) float64 {
	// Count high-confidence signals
	highConfidenceCount := 0
	totalCount := len(signals)

	for _, signal := range signals {
		if signal.Confidence >= 0.8 {
			highConfidenceCount++
		}
	}

	// If most signals are high confidence, boost score
	if totalCount > 0 {
		highConfRatio := float64(highConfidenceCount) / float64(totalCount)
		if highConfRatio >= 0.7 { // 70% or more high confidence
			return 1.2 // Boost for consistent high confidence
		}
	}

	return 1.0
}

// calculateTemporalFactor adjusts score based on time factors
func calculateTemporalFactor() float64 {
	// In a real implementation, this would consider:
	// - How recently the domain was registered
	// - How long similar patterns have been observed
	// - Time-based trends in behavior

	// For now, return neutral factor
	return 1.0
}

// calculateConfidenceFactor calculates a factor based on number of categories with signals
func calculateConfidenceFactor(categoryScores map[string]float64) float64 {
	count := 0
	for _, score := range categoryScores {
		if score > 0.0 {
			count++
		}
	}

	// More categories with signals = higher confidence
	// But cap it to prevent overconfidence
	if count >= 3 {
		return 1.0
	}
	return float64(count) * 0.33
}

// classifyJLILevel classifies the JLI score into LOW, MEDIUM, or HIGH
func classifyJLILevel(jliScore float64, thresholds config.ThresholdConfig) string {
	if jliScore >= thresholds.High {
		return "HIGH"
	} else if jliScore >= thresholds.Medium {
		return "MEDIUM"
	}
	return "LOW"
}

// OutputJSON outputs the result in JSON format
func OutputJSON(r *models.AnalysisResult) {
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// OutputCSV outputs the result in CSV format
func OutputCSV(r *models.AnalysisResult) {
	fmt.Println("domain,jli_score,jli_level,profile_used")
	fmt.Printf("%s,%.3f,%s,%s\n", r.Domain.Domain, r.JLIScore, r.JLILevel, r.ProfileUsed)
}

// OutputTable outputs the result in a formatted table
func OutputTable(r *models.AnalysisResult) {
	// Domain Summary Table
	t := table.NewWriter()
	t.SetOutputMirror(color.Output)
	t.AppendHeader(table.Row{"Domain", "JLI Score", "JLI Level", "CDN Provider"})
	t.AppendRow([]interface{}{r.Domain.Domain, fmt.Sprintf("%.3f", r.JLIScore), r.Domain.JLILevel, r.Domain.CDNProvider})
	t.SetStyle(table.StyleLight)
	t.Render()

	fmt.Println()

	// Category Breakdown Table
	t2 := table.NewWriter()
	t2.SetOutputMirror(color.Output)
	t2.AppendHeader(table.Row{"Category", "Score", "Weight", "Contribution"})

	totalContribution := 0.0
	for category, breakdown := range r.CategoryBreakdown {
		t2.AppendRow([]interface{}{
			category,
			fmt.Sprintf("%.3f", breakdown.Score),
			fmt.Sprintf("%.3f", breakdown.Weight),
			fmt.Sprintf("%.3f", breakdown.Contribution),
		})
		totalContribution += breakdown.Contribution
	}

	// Add total row
	t2.AppendSeparator()
	t2.AppendRow([]interface{}{"TOTAL", "", "", fmt.Sprintf("%.3f", totalContribution)})
	t2.SetStyle(table.StyleLight)
	t2.Render()

	fmt.Println()

	// Print JLI Level with appropriate color
	levelColor := color.FgWhite
	switch r.JLILevel {
	case "HIGH":
		levelColor = color.FgRed
	case "MEDIUM":
		levelColor = color.FgYellow
	case "LOW":
		levelColor = color.FgGreen
	}
	coloredLevel := color.New(levelColor).Sprint(r.JLILevel)
	fmt.Printf("Judol Likelihood Level: %s\n", coloredLevel)
}

// SaveToDB saves the result to local database
func SaveToDB(r *models.AnalysisResult) {
	// In a real implementation, this would save to a local SQLite database
	fmt.Println("Saving to local database... (not implemented in this example)")
}