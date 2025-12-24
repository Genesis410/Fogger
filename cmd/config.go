package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/genesis410/fogger/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management for fogger",
	Long: `Manage configuration for fogger including
scoring profiles and thresholds.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		fmt.Printf("Current Configuration:\n")
		fmt.Printf("Gambling UI Weight: %.2f\n", cfg.Scoring.GamblingUI)
		fmt.Printf("Payment Signal Weight: %.2f\n", cfg.Scoring.PaymentSignal)
		fmt.Printf("Infrastructure Correlation Weight: %.2f\n", cfg.Scoring.InfraCorrelation)
		fmt.Printf("Domain Churn Weight: %.2f\n", cfg.Scoring.DomainChurn)
		fmt.Printf("CDN Pattern Weight: %.2f\n", cfg.Scoring.CDNPattern)
		fmt.Printf("High Threshold: %.2f\n", cfg.Threshold.High)
		fmt.Printf("Medium Threshold: %.2f\n", cfg.Threshold.Medium)
	},
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate current configuration",
	Long:  `Check if the current configuration is valid.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		
		totalWeight := cfg.Scoring.GamblingUI +
			cfg.Scoring.PaymentSignal +
			cfg.Scoring.InfraCorrelation +
			cfg.Scoring.DomainChurn +
			cfg.Scoring.CDNPattern

		if totalWeight == 1.0 {
			fmt.Println("Configuration is valid")
		} else {
			fmt.Printf("Configuration warning: weights sum to %.2f, not 1.0\n", totalWeight)
		}
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	rootCmd.AddCommand(configCmd)
}