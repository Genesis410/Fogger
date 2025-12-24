package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/spf13/viper"
)

// ScoringConfig holds the weights for different signal categories
type ScoringConfig struct {
	GamblingUI       float64 `mapstructure:"gambling_ui"`
	PaymentSignal    float64 `mapstructure:"payment_signal"`
	InfraCorrelation float64 `mapstructure:"infra_correlation"`
	DomainChurn      float64 `mapstructure:"domain_churn"`
	CDNPattern       float64 `mapstructure:"cdn_pattern"`
}

// ThresholdConfig holds the thresholds for classification
type ThresholdConfig struct {
	High   float64 `mapstructure:"high"`
	Medium float64 `mapstructure:"medium"`
}

// Config holds the complete configuration
type Config struct {
	Scoring   ScoringConfig   `mapstructure:"scoring"`
	Threshold ThresholdConfig `mapstructure:"thresholds"`
}

var (
	config     *Config
	configOnce sync.Once
)

// Initialize loads the configuration
func Initialize() {
	configOnce.Do(func() {
		viper.SetDefault("scoring.gambling_ui", 0.30)
		viper.SetDefault("scoring.payment_signal", 0.25)
		viper.SetDefault("scoring.infra_correlation", 0.20)
		viper.SetDefault("scoring.domain_churn", 0.15)
		viper.SetDefault("scoring.cdn_pattern", 0.10)

		viper.SetDefault("thresholds.high", 0.75)
		viper.SetDefault("thresholds.medium", 0.50)

		// Read in configuration from file
		viper.SetConfigName(".fogger")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath(".")

		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("Config file not found, using defaults: %v\n", err)
		}

		config = &Config{}
		if err := viper.Unmarshal(config); err != nil {
			log.Fatalf("Failed to unmarshal config: %v", err)
		}

		// Validate weights sum to 1.0
		totalWeight := config.Scoring.GamblingUI +
			config.Scoring.PaymentSignal +
			config.Scoring.InfraCorrelation +
			config.Scoring.DomainChurn +
			config.Scoring.CDNPattern

		if totalWeight != 1.0 {
			log.Printf("Warning: Scoring weights sum to %f, not 1.0", totalWeight)
		}
	})
}

// Get returns the current configuration
func Get() *Config {
	if config == nil {
		Initialize()
	}
	return config
}