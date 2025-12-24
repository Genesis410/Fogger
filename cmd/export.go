package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/genesis410/fogger/internal/analyzer"
	"github.com/genesis410/fogger/internal/models"
	"github.com/spf13/cobra"
)

// ExportData handles exporting analysis results
func ExportData(results []*models.AnalysisResult, format string, filename string) error {
	switch format {
	case "json":
		return exportJSON(results, filename)
	case "csv":
		return exportCSV(results, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportJSON exports results in JSON format
func exportJSON(results []*models.AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	return encoder.Encode(results)
}

// exportCSV exports results in CSV format
func exportCSV(results []*models.AnalysisResult, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
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
		return err
	}

	// Write data rows
	for _, result := range results {
		row := []string{
			result.Domain.Domain,
			fmt.Sprintf("%.3f", result.JLIScore),
			result.JLILevel,
			result.Domain.CDNProvider,
			result.Domain.FirstSeen.Format(time.RFC3339),
			result.Domain.LastSeen.Format(time.RFC3339),
			fmt.Sprintf("%v", result.Domain.ClusterID),
			fmt.Sprintf("%d", len(result.Domain.Signals)),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "UX")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "PAYMENT")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "INFRA")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "DNS")),
			fmt.Sprintf("%d", countSignalsByCategory(result.Domain.Signals, "CDN")),
		}
		
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
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

// MonitorDomain continuously monitors a domain
func MonitorDomain(domain string, interval time.Duration, duration time.Duration) {
	endTime := time.Now().Add(duration)
	
	fmt.Printf("Monitoring %s every %v for %v\n", domain, interval, duration)
	
	for time.Now().Before(endTime) {
		fmt.Printf("Scanning %s at %s...\n", domain, time.Now().Format(time.RFC3339))
		
		// Perform analysis
		result := analyzer.AnalyzeDomain(domain, 10*time.Second, "standard")
		
		// Display result
		fmt.Printf("JLI Score: %.3f, Level: %s\n", result.JLIScore, result.JLILevel)
		
		// Wait for next scan
		time.Sleep(interval)
	}
	
	fmt.Println("Monitoring completed")
}

// Add monitoring command to the CLI
var monitorCmd = &cobra.Command{
	Use:   "monitor <domain>",
	Short: "Continuously monitor a domain for changes",
	Long: `Monitor continuously checks a domain at specified intervals
to detect changes in its gambling indicators.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		interval, _ := cmd.Flags().GetDuration("interval")
		duration, _ := cmd.Flags().GetDuration("duration")
		
		MonitorDomain(domain, interval, duration)
	},
}

func init() {
	monitorCmd.Flags().Duration("interval", 5*time.Minute, "Monitoring interval")
	monitorCmd.Flags().Duration("duration", 1*time.Hour, "Total monitoring duration")
	rootCmd.AddCommand(monitorCmd)
}