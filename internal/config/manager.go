package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ConfigManager handles configuration management
type ConfigManager struct {
	config *Config
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		config: Get(), // Use the existing config
	}
}

// LoadConfig loads configuration from file
func (cm *ConfigManager) LoadConfig(configPath string) error {
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Set up default config locations
		viper.SetConfigName(".fogger")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Unmarshal config
	if err := viper.Unmarshal(&cm.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return nil
}

// SaveConfig saves configuration to file
func (cm *ConfigManager) SaveConfig(configPath string) error {
	if configPath == "" {
		// Default to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		configPath = filepath.Join(home, ".fogger.yaml")
	}

	// Set viper values
	viper.Set("scoring.gambling_ui", cm.config.Scoring.GamblingUI)
	viper.Set("scoring.payment_signal", cm.config.Scoring.PaymentSignal)
	viper.Set("scoring.infra_correlation", cm.config.Scoring.InfraCorrelation)
	viper.Set("scoring.domain_churn", cm.config.Scoring.DomainChurn)
	viper.Set("scoring.cdn_pattern", cm.config.Scoring.CDNPattern)
	
	viper.Set("thresholds.high", cm.config.Threshold.High)
	viper.Set("thresholds.medium", cm.config.Threshold.Medium)

	// Write to file
	if err := viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.config
}

// SetScoringConfig updates the scoring configuration
func (cm *ConfigManager) SetScoringConfig(scoring ScoringConfig) {
	cm.config.Scoring = scoring
}

// SetThresholdConfig updates the threshold configuration
func (cm *ConfigManager) SetThresholdConfig(threshold ThresholdConfig) {
	cm.config.Threshold = threshold
}

// ValidateConfig validates the current configuration
func (cm *ConfigManager) ValidateConfig() error {
	// Validate weights sum to 1.0
	totalWeight := cm.config.Scoring.GamblingUI +
		cm.config.Scoring.PaymentSignal +
		cm.config.Scoring.InfraCorrelation +
		cm.config.Scoring.DomainChurn +
		cm.config.Scoring.CDNPattern

	if totalWeight != 1.0 {
		return fmt.Errorf("scoring weights sum to %f, not 1.0", totalWeight)
	}

	// Validate thresholds
	if cm.config.Threshold.High < cm.config.Threshold.Medium {
		return fmt.Errorf("high threshold (%f) must be >= medium threshold (%f)", 
			cm.config.Threshold.High, cm.config.Threshold.Medium)
	}

	if cm.config.Threshold.High > 1.0 || cm.config.Threshold.Medium > 1.0 {
		return fmt.Errorf("thresholds must be between 0 and 1")
	}

	if cm.config.Threshold.High < 0.0 || cm.config.Threshold.Medium < 0.0 {
		return fmt.Errorf("thresholds must be between 0 and 1")
	}

	return nil
}

// GetDefaultConfig returns the default configuration
func (cm *ConfigManager) GetDefaultConfig() Config {
	return Config{
		Scoring: ScoringConfig{
			GamblingUI:       0.30,
			PaymentSignal:    0.25,
			InfraCorrelation: 0.20,
			DomainChurn:      0.15,
			CDNPattern:       0.10,
		},
		Threshold: ThresholdConfig{
			High:   0.75,
			Medium: 0.50,
		},
	}
}

// ResetToDefault resets configuration to default values
func (cm *ConfigManager) ResetToDefault() {
	defaultConfig := cm.GetDefaultConfig()
	cm.config.Scoring = defaultConfig.Scoring
	cm.config.Threshold = defaultConfig.Threshold
}

// CreateProfile creates a new scoring profile
func (cm *ConfigManager) CreateProfile(name string, scoring ScoringConfig, threshold ThresholdConfig) error {
	// In a real implementation, this would save profiles to a separate file or database
	// For now, we'll just validate the profile
	profile := Config{
		Scoring:   scoring,
		Threshold: threshold,
	}
	
	return cm.validateProfile(profile)
}

// validateProfile validates a scoring profile
func (cm *ConfigManager) validateProfile(profile Config) error {
	// Validate weights sum to 1.0
	totalWeight := profile.Scoring.GamblingUI +
		profile.Scoring.PaymentSignal +
		profile.Scoring.InfraCorrelation +
		profile.Scoring.DomainChurn +
		profile.Scoring.CDNPattern

	if totalWeight != 1.0 {
		return fmt.Errorf("profile scoring weights sum to %f, not 1.0", totalWeight)
	}

	// Validate thresholds
	if profile.Threshold.High < profile.Threshold.Medium {
		return fmt.Errorf("profile high threshold (%f) must be >= medium threshold (%f)", 
			profile.Threshold.High, profile.Threshold.Medium)
	}

	if profile.Threshold.High > 1.0 || profile.Threshold.Medium > 1.0 ||
		profile.Threshold.High < 0.0 || profile.Threshold.Medium < 0.0 {
		return fmt.Errorf("profile thresholds must be between 0 and 1")
	}

	return nil
}

// ApplyProfile applies a scoring profile
func (cm *ConfigManager) ApplyProfile(scoring ScoringConfig, threshold ThresholdConfig) {
	cm.config.Scoring = scoring
	cm.config.Threshold = threshold
}

// GetAvailableProfiles returns a list of available profiles
func (cm *ConfigManager) GetAvailableProfiles() []string {
	// In a real implementation, this would read from a profiles directory
	// For now, return built-in profiles
	return []string{"standard", "intensive", "conservative", "aggressive"}
}

// GetProfile returns a specific profile configuration
func (cm *ConfigManager) GetProfile(name string) (*Config, error) {
	switch name {
	case "standard":
		return &Config{
			Scoring: ScoringConfig{
				GamblingUI:       0.30,
				PaymentSignal:    0.25,
				InfraCorrelation: 0.20,
				DomainChurn:      0.15,
				CDNPattern:       0.10,
			},
			Threshold: ThresholdConfig{
				High:   0.75,
				Medium: 0.50,
			},
		}, nil
	case "intensive":
		return &Config{
			Scoring: ScoringConfig{
				GamblingUI:       0.35,
				PaymentSignal:    0.30,
				InfraCorrelation: 0.20,
				DomainChurn:      0.10,
				CDNPattern:       0.05,
			},
			Threshold: ThresholdConfig{
				High:   0.60,
				Medium: 0.30,
			},
		}, nil
	case "conservative":
		return &Config{
			Scoring: ScoringConfig{
				GamblingUI:       0.20,
				PaymentSignal:    0.20,
				InfraCorrelation: 0.25,
				DomainChurn:      0.25,
				CDNPattern:       0.10,
			},
			Threshold: ThresholdConfig{
				High:   0.85,
				Medium: 0.65,
			},
		}, nil
	case "aggressive":
		return &Config{
			Scoring: ScoringConfig{
				GamblingUI:       0.40,
				PaymentSignal:    0.30,
				InfraCorrelation: 0.15,
				DomainChurn:      0.10,
				CDNPattern:       0.05,
			},
			Threshold: ThresholdConfig{
				High:   0.50,
				Medium: 0.25,
			},
		}, nil
	default:
		return nil, fmt.Errorf("profile '%s' not found", name)
	}
}

// UpdateConfigValue updates a specific configuration value
func (cm *ConfigManager) UpdateConfigValue(key string, value interface{}) error {
	switch key {
	case "scoring.gambling_ui":
		if v, ok := value.(float64); ok {
			cm.config.Scoring.GamblingUI = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "scoring.payment_signal":
		if v, ok := value.(float64); ok {
			cm.config.Scoring.PaymentSignal = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "scoring.infra_correlation":
		if v, ok := value.(float64); ok {
			cm.config.Scoring.InfraCorrelation = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "scoring.domain_churn":
		if v, ok := value.(float64); ok {
			cm.config.Scoring.DomainChurn = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "scoring.cdn_pattern":
		if v, ok := value.(float64); ok {
			cm.config.Scoring.CDNPattern = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "thresholds.high":
		if v, ok := value.(float64); ok {
			cm.config.Threshold.High = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	case "thresholds.medium":
		if v, ok := value.(float64); ok {
			cm.config.Threshold.Medium = v
		} else {
			return fmt.Errorf("invalid value type for %s", key)
		}
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// Validate the updated config
	return cm.ValidateConfig()
}

// GetConfigValue returns a specific configuration value
func (cm *ConfigManager) GetConfigValue(key string) (interface{}, error) {
	switch key {
	case "scoring.gambling_ui":
		return cm.config.Scoring.GamblingUI, nil
	case "scoring.payment_signal":
		return cm.config.Scoring.PaymentSignal, nil
	case "scoring.infra_correlation":
		return cm.config.Scoring.InfraCorrelation, nil
	case "scoring.domain_churn":
		return cm.config.Scoring.DomainChurn, nil
	case "scoring.cdn_pattern":
		return cm.config.Scoring.CDNPattern, nil
	case "thresholds.high":
		return cm.config.Threshold.High, nil
	case "thresholds.medium":
		return cm.config.Threshold.Medium, nil
	default:
		return nil, fmt.Errorf("unknown configuration key: %s", key)
	}
}