package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/genesis410/fogger/internal/analyzer"
)

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
		noColor, _ := cmd.Flags().GetBool("no-color")
		timeout, _ := cmd.Flags().GetInt("timeout")
		profile, _ := cmd.Flags().GetString("profile")
		save, _ := cmd.Flags().GetBool("save")

		if noColor {
			color.NoColor = true
		}

		// Set timeout
		clientTimeout := time.Duration(timeout) * time.Second

		fmt.Printf("Scanning domain: %s\n", color.GreenString(domain))
		
		// Perform the analysis
		result := analyzer.AnalyzeDomain(domain, clientTimeout, profile)
		
		if jsonOutput {
			analyzer.OutputJSON(result)
		} else if csvOutput {
			analyzer.OutputCSV(result)
		} else {
			analyzer.OutputTable(result)
		}

		if save {
			// Save to local DB
			analyzer.SaveToDB(result)
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

func init() {
	rootCmd.AddCommand(scanCmd)

	// Add flags for the scan command
	scanCmd.Flags().Bool("json", false, "Output JSON only")
	scanCmd.Flags().Bool("csv", false, "Output CSV")
	scanCmd.Flags().Bool("no-color", false, "Disable ANSI coloring")
	scanCmd.Flags().Int("timeout", 10, "Network timeout (default: 10)")
	scanCmd.Flags().String("profile", "standard", "Scoring profile (default: standard)")
	scanCmd.Flags().Bool("save", false, "Persist result to local DB")
}