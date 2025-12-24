package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/genesis410/fogger/internal/analyzer"
	"github.com/genesis410/fogger/internal/models"
)

// OutputJSON outputs the result in JSON format with enhanced structure
func OutputJSON(r *models.AnalysisResult) {
	// Create enhanced output structure
	output := map[string]interface{}{
		"scan_metadata": map[string]interface{}{
			"domain":        r.Domain.Domain,
			"timestamp":     time.Now().Format(time.RFC3339),
			"scan_duration": "N/A", // Would be added in real implementation
		},
		"risk_assessment": map[string]interface{}{
			"jli_score":   r.JLIScore,
			"risk_level":  r.JLILevel,
			"confidence":  calculateOverallConfidence(r),
		},
		"technical_details": map[string]interface{}{
			"cdn_provider":    r.Domain.CDNProvider,
			"ip_address":      "N/A", // Would be added in real implementation
			"origin_ip_guess": "N/A", // Would be added in real implementation
			"ssl_info":        map[string]interface{}{},
		},
		"detection_evidence": r.Domain.Signals,
		"category_breakdown": r.CategoryBreakdown,
	}

	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// OutputCSV outputs the result in CSV format with enhanced structure
func OutputCSV(r *models.AnalysisResult) {
	fmt.Println("domain,jli_score,risk_level,cdn_provider,scan_timestamp,total_signals,ux_signals,payment_signals,infra_signals,dns_signals,cdn_signals,evidence_count")

	uxCount := countSignalsByCategory(r.Domain.Signals, "UX")
	paymentCount := countSignalsByCategory(r.Domain.Signals, "PAYMENT")
	infraCount := countSignalsByCategory(r.Domain.Signals, "INFRA")
	dnsCount := countSignalsByCategory(r.Domain.Signals, "DNS")
	cdnCount := countSignalsByCategory(r.Domain.Signals, "CDN")

	fmt.Printf("%s,%.3f,%s,%s,%s,%d,%d,%d,%d,%d,%d,%d\n",
		r.Domain.Domain,
		r.JLIScore,
		r.JLILevel,
		r.Domain.CDNProvider,
		time.Now().Format(time.RFC3339),
		len(r.Domain.Signals),
		uxCount,
		paymentCount,
		infraCount,
		dnsCount,
		cdnCount,
		len(r.Domain.Signals),
	)
}

// OutputTable outputs the result in a rich formatted table
func OutputTable(r *models.AnalysisResult) {
	// Domain Summary Table
	summaryTable := table.NewWriter()
	summaryTable.SetOutputMirror(color.Output)
	summaryTable.AppendHeader(table.Row{"Domain", "JLI Score", "Risk Level", "CDN Provider", "Scan Time"})
	summaryTable.AppendRow([]interface{}{
		r.Domain.Domain,
		fmt.Sprintf("%.3f", r.JLIScore),
		r.JLILevel,
		r.Domain.CDNProvider,
		time.Now().Format("2006-01-02 15:04:05"),
	})
	summaryTable.SetStyle(table.StyleLight)
	summaryTable.Render()

	fmt.Println()

	// Category Breakdown Table
	breakdownTable := table.NewWriter()
	breakdownTable.SetOutputMirror(color.Output)
	breakdownTable.AppendHeader(table.Row{"Category", "Score", "Weight", "Contribution"})

	totalContribution := 0.0
	for category, breakdown := range r.CategoryBreakdown {
		breakdownTable.AppendRow([]interface{}{
			category,
			fmt.Sprintf("%.3f", breakdown.Score),
			fmt.Sprintf("%.3f", breakdown.Weight),
			fmt.Sprintf("%.3f", breakdown.Contribution),
		})
		totalContribution += breakdown.Contribution
	}

	// Add total row
	breakdownTable.AppendSeparator()
	breakdownTable.AppendRow([]interface{}{"TOTAL", "", "", fmt.Sprintf("%.3f", totalContribution)})
	breakdownTable.SetStyle(table.StyleLight)
	breakdownTable.Render()

	fmt.Println()

	// Evidence Summary
	if len(r.Domain.Signals) > 0 {
		evidenceTable := table.NewWriter()
		evidenceTable.SetOutputMirror(color.Output)
		evidenceTable.AppendHeader(table.Row{"#", "Category", "Description", "Confidence"})

		for i, signal := range r.Domain.Signals {
			if i < 10 { // Show first 10 signals to avoid cluttering
				evidenceTable.AppendRow([]interface{}{
					i + 1,
					signal.Category,
					truncateString(signal.Description, 50),
					fmt.Sprintf("%.2f", signal.Confidence),
				})
			}
		}

		if len(r.Domain.Signals) > 10 {
			evidenceTable.AppendRow([]interface{}{
				fmt.Sprintf("+%d more", len(r.Domain.Signals)-10),
				"",
				"Additional evidence...",
				"",
			})
		}

		evidenceTable.SetStyle(table.StyleLight)
		evidenceTable.Render()
		fmt.Printf("\nTotal evidence found: %d\n", len(r.Domain.Signals))
	}

	fmt.Println()

	// Risk Level with appropriate color
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

// OutputDetailedReport creates a comprehensive report with all details
func OutputDetailedReport(r *models.AnalysisResult) {
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│                        DETAILED SCAN REPORT                     │")
	fmt.Println("├─────────────────────────────────────────────────────────────────┤")

	// Summary section
	fmt.Printf("│ Domain: %-55s │\n", r.Domain.Domain)
	fmt.Printf("│ Scan Time: %-51s │\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("│ Risk Level: %-50s │\n", r.JLILevel)
	fmt.Printf("│ JLI Score: %-51s │\n", fmt.Sprintf("%.3f", r.JLIScore))
	fmt.Printf("│ CDN Provider: %-48s │\n", r.Domain.CDNProvider)
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")

	fmt.Println()

	// Detailed breakdown
	fmt.Println("CATEGORIZATION BREAKDOWN:")
	OutputTable(r) // Reuse the table function for consistency

	fmt.Println()

	// Evidence details
	fmt.Println("EVIDENCE DETAILS:")
	for i, signal := range r.Domain.Signals {
		if i < 15 { // Limit to first 15 for readability
			fmt.Printf("  %d. [%s] %s (Confidence: %.2f)\n",
				i+1, signal.Category, signal.Description, signal.Confidence)
		}
	}

	if len(r.Domain.Signals) > 15 {
		fmt.Printf("  ... and %d more evidence items\n", len(r.Domain.Signals)-15)
	}

	fmt.Println()

	// Confidence summary
	confidence := calculateOverallConfidence(r)
	fmt.Printf("OVERALL CONFIDENCE: %.2f\n", confidence)

	// Risk assessment
	riskAssessment := getRiskAssessment(r.JLIScore, r.JLILevel)
	fmt.Printf("RISK ASSESSMENT: %s\n", riskAssessment)

	// Recommendations
	recommendations := getRecommendations(r.JLILevel, r.Domain.Signals)
	fmt.Println("RECOMMENDATIONS:")
	for _, rec := range recommendations {
		fmt.Printf("  • %s\n", rec)
	}
}

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan <domain>",
	Short: "Analyze a domain for gambling indicators",
	Long: `Scan analyzes a domain and produces a Judol Likelihood Index (JLI)
along with evidence of gambling-related activities.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]

		// Validate domain format
		if !isValidDomain(domain) {
			fmt.Printf("Invalid domain format: %s\n", domain)
			os.Exit(1)
		}

		jsonOutput, _ := cmd.Flags().GetBool("json")
		csvOutput, _ := cmd.Flags().GetBool("csv")
		detailedOutput, _ := cmd.Flags().GetBool("detailed")
		batchMode, _ := cmd.Flags().GetBool("batch")
		noColor, _ := cmd.Flags().GetBool("no-color")
		timeout, _ := cmd.Flags().GetInt("timeout")
		profile, _ := cmd.Flags().GetString("profile")
		save, _ := cmd.Flags().GetBool("save")

		if noColor {
			color.NoColor = true
		}

		// Set timeout
		clientTimeout := time.Duration(timeout) * time.Second

		if !batchMode {
			fmt.Printf("Scanning domain: %s\n", color.GreenString(domain))
		}

		// Perform the analysis
		result := analyzer.AnalyzeDomain(domain, clientTimeout, profile)

		if jsonOutput {
			OutputJSON(result)
		} else if csvOutput {
			OutputCSV(result)
		} else if detailedOutput {
			OutputDetailedReport(result)
		} else {
			OutputTable(result)
		}

		if save {
			// Save to local DB
			SaveToDB(result)
		}
	},
}

func isValidDomain(domain string) bool {
	// Simple domain validation - in a real implementation, use proper validation
	domain = strings.TrimSpace(domain)
	if len(domain) < 1 || len(domain) > 253 {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
	}

	return true
}

// Helper functions
func countSignalsByCategory(signals []models.Signal, category string) int {
	count := 0
	for _, signal := range signals {
		if signal.Category == category {
			count++
		}
	}
	return count
}

func truncateString(str string, num int) string {
	if len(str) > num {
		return str[0:num] + "..."
	}
	return str
}

func calculateOverallConfidence(r *models.AnalysisResult) float64 {
	// Calculate based on number of high-confidence signals
	highConfidenceCount := 0
	for _, signal := range r.Domain.Signals {
		if signal.Confidence > 0.8 {
			highConfidenceCount++
		}
	}

	if len(r.Domain.Signals) == 0 {
		return 0.0
	}

	return float64(highConfidenceCount) / float64(len(r.Domain.Signals))
}

func getRiskAssessment(score float64, level string) string {
	switch level {
	case "HIGH":
		return "High probability of gambling-related activity. Immediate action recommended."
	case "MEDIUM":
		return "Moderate probability of gambling-related activity. Investigation suggested."
	case "LOW":
		return "Low probability of gambling-related activity. Monitor for changes."
	default:
		return "Unknown risk level."
	}
}

func getRecommendations(level string, signals []models.Signal) []string {
	var recommendations []string

	switch level {
	case "HIGH":
		recommendations = append(recommendations,
			"Block domain access at network level",
			"Investigate associated domains and infrastructure",
			"Report to appropriate authorities")
	case "MEDIUM":
		recommendations = append(recommendations,
			"Monitor domain for changes",
			"Review associated infrastructure",
			"Consider further investigation")
	case "LOW":
		recommendations = append(recommendations,
			"Continue monitoring",
			"No immediate action required")
	}

	// Add specific recommendations based on signals
	hasPayment := false
	hasGamblingUX := false
	for _, signal := range signals {
		if signal.Category == "PAYMENT" {
			hasPayment = true
		}
		if signal.Category == "UX" {
			hasGamblingUX = true
		}
	}

	if hasPayment {
		recommendations = append(recommendations,
			"Investigate payment methods used on this domain")
	}

	if hasGamblingUX {
		recommendations = append(recommendations,
			"Review user interface elements for gambling indicators")
	}

	return recommendations
}

// SaveToDB saves the result to local database
func SaveToDB(r *models.AnalysisResult) {
	// In a real implementation, this would save to a local SQLite database
	fmt.Println("Saving to local database... (not implemented in this example)")
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Add flags for the scan command
	scanCmd.Flags().Bool("json", false, "Output JSON only")
	scanCmd.Flags().Bool("csv", false, "Output CSV")
	scanCmd.Flags().Bool("detailed", false, "Output detailed report")
	scanCmd.Flags().Bool("batch", false, "Batch mode (no extra output)")
	scanCmd.Flags().Bool("no-color", false, "Disable ANSI coloring")
	scanCmd.Flags().Int("timeout", 10, "Network timeout (default: 10)")
	scanCmd.Flags().String("profile", "standard", "Scoring profile (default: standard)")
	scanCmd.Flags().Bool("save", false, "Persist result to local DB")
}