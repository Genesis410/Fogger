package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export data for integration with other systems",
	Long: `Export allows integration with SIEM, payment systems,
or regulator pipelines.`,
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		since, _ := cmd.Flags().GetString("since")
		domain, _ := cmd.Flags().GetString("domain")
		cluster, _ := cmd.Flags().GetString("cluster")
		output, _ := cmd.Flags().GetString("output")

		fmt.Printf("Export format: %s\n", format)
		fmt.Printf("Since: %s\n", since)
		fmt.Printf("Domain: %s\n", domain)
		fmt.Printf("Cluster: %s\n", cluster)
		fmt.Printf("Output: %s\n", output)
		
		// In a real implementation, this would export data
		fmt.Println("Export functionality would be implemented here...")
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Add flags for the export command
	exportCmd.Flags().String("format", "json", "Export format (json, csv)")
	exportCmd.Flags().String("since", "30d", "Time period to export (e.g., 30d)")
	exportCmd.Flags().String("domain", "", "Specific domain to export")
	exportCmd.Flags().String("cluster", "", "Specific cluster to export")
	exportCmd.Flags().String("output", "", "Output file path")
}