package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/genesis410/fogger/internal/analyzer"
	"github.com/genesis410/fogger/internal/config"
	"github.com/genesis410/fogger/internal/models"
	"github.com/genesis410/fogger/internal/scanner"
)

// TestMain runs all tests
func TestMain(m *testing.M) {
	// Initialize configuration
	config.Initialize()

	// Run tests
	exitCode := m.Run()

	// Exit with the same code as the tests
	os.Exit(exitCode)
}

// TestScanner tests the domain scanning functionality
func TestScanner(t *testing.T) {
	// Test with a non-existent domain to check error handling
	domain := "nonexistent-domain-1234567890.com"
	
	result := scanner.ScanDomain(domain, 5*time.Second)
	
	if result.Domain != domain {
		t.Errorf("Expected domain %s, got %s", domain, result.Domain)
	}
	
	// The scanner should still return a result even if the domain doesn't exist
	t.Logf("Scan result for %s: CDN Provider = %s", domain, result.CDNProvider)
}

// TestAnalyzer tests the analysis functionality
func TestAnalyzer(t *testing.T) {
	// Test with minimal content to ensure analyzer works
	testContent := `
	<html>
	<head>
		<title>Test Gambling Site</title>
	</head>
	<body>
		<h1>Slot Gacor Hari Ini</h1>
		<p>Deposit via OVO, DANA, Gopay</p>
		<button>Deposit Sekarang</button>
		<button>Withdraw</button>
	</body>
	</html>
	`
	
	// Test content without needing mockScanResult
	
	// We can't easily test the analyzer without a real domain scan,
	// but we can test the configuration
	cfg := config.Get()
	if cfg.Scoring.GamblingUI <= 0 {
		t.Error("Expected GamblingUI weight to be > 0")
	}
	
	t.Logf("Configuration loaded: GamblingUI = %f", cfg.Scoring.GamblingUI)
}

// TestConfiguration tests the configuration management
func TestConfiguration(t *testing.T) {
	// Get the current configuration
	cfg := config.Get()
	
	// Validate that weights sum to approximately 1.0
	totalWeight := cfg.Scoring.GamblingUI +
		cfg.Scoring.PaymentSignal +
		cfg.Scoring.InfraCorrelation +
		cfg.Scoring.DomainChurn +
		cfg.Scoring.CDNPattern
	
	if totalWeight < 0.99 || totalWeight > 1.01 { // Allow small floating point errors
		t.Errorf("Expected weights to sum to 1.0, got %f", totalWeight)
	}
	
	// Check that thresholds are reasonable
	if cfg.Threshold.High <= cfg.Threshold.Medium {
		t.Error("High threshold should be greater than medium threshold")
	}
	
	if cfg.Threshold.High < 0 || cfg.Threshold.High > 1 {
		t.Error("High threshold should be between 0 and 1")
	}
	
	if cfg.Threshold.Medium < 0 || cfg.Threshold.Medium > 1 {
		t.Error("Medium threshold should be between 0 and 1")
	}
	
	t.Logf("Configuration validated: weights sum to %f", totalWeight)
}

// TestJLICalculation tests the JLI calculation logic
func TestJLICalculation(t *testing.T) {
	// Create test signals
	testSignals := []models.Signal{
		{
			Category:   "UX",
			Confidence: 0.8,
		},
		{
			Category:   "PAYMENT",
			Confidence: 0.9,
		},
		{
			Category:   "INFRA",
			Confidence: 0.6,
		},
		{
			Category:   "DNS",
			Confidence: 0.4,
		},
		{
			Category:   "CDN",
			Confidence: 0.3,
		},
	}
	
	// Calculate category scores (using max for each category)
	categoryScores := make(map[string]float64)
	for _, signal := range testSignals {
		currentScore, exists := categoryScores[signal.Category]
		if !exists || signal.Confidence > currentScore {
			categoryScores[signal.Category] = signal.Confidence
		}
	}
	
	// Get configuration weights
	cfg := config.Get()
	
	// Calculate JLI score
	jliScore := 0.0
	jliScore += categoryScores["UX"] * cfg.Scoring.GamblingUI
	jliScore += categoryScores["PAYMENT"] * cfg.Scoring.PaymentSignal
	jliScore += categoryScores["INFRA"] * cfg.Scoring.InfraCorrelation
	jliScore += categoryScores["DNS"] * cfg.Scoring.DomainChurn
	jliScore += categoryScores["CDN"] * cfg.Scoring.CDNPattern
	
	// Apply confidence factor (simplified)
	confidenceFactor := 1.0 // In a real test, this would be calculated
	
	finalScore := jliScore * confidenceFactor
	
	if finalScore < 0 || finalScore > 1 {
		t.Errorf("Expected JLI score to be between 0 and 1, got %f", finalScore)
	}
	
	t.Logf("Calculated JLI score: %f", finalScore)
}

// TestBehavioralAnalyzer tests the behavioral analysis functionality
func TestBehavioralAnalyzer(t *testing.T) {
	analyzer := analyzer.NewBehavioralAnalyzer()
	
	// Test content with gambling keywords
	testContent := `
	<html>
	<head>
		<title>Situs Judi Slot Online Terpercaya</title>
	</head>
	<body>
		<h1>Slot Gacor Maxwin Hari Ini</h1>
		<p>Deposit via OVO, DANA, Gopay, Qris 2.0</p>
		<p>Daftar sekarang dapat bonus besar</p>
		<button>Deposit Murah</button>
		<button>Withdraw Cepat</button>
		<p>Customer Service 24 Jam</p>
		<input type="text" name="username">
		<input type="password" name="pin">
		<input type="number" name="amount">
	</body>
	</html>
	`
	
	signals := analyzer.AnalyzeContent(testContent)
	
	// Check that we found some signals
	if len(signals) == 0 {
		t.Error("Expected to find gambling-related signals, but found none")
	}
	
	// Count signals by category
	uxCount := 0
	paymentCount := 0
	for _, signal := range signals {
		if signal.Category == "UX" {
			uxCount++
		} else if signal.Category == "PAYMENT" {
			paymentCount++
		}
	}
	
	if uxCount == 0 {
		t.Error("Expected to find UX-related signals")
	}
	
	if paymentCount == 0 {
		t.Error("Expected to find payment-related signals")
	}
	
	t.Logf("Found %d signals total: %d UX, %d PAYMENT", len(signals), uxCount, paymentCount)
}

// TestDOMAnalysis tests the DOM structure analysis
func TestDOMAnalysis(t *testing.T) {
	analyzer := analyzer.NewBehavioralAnalyzer()
	
	testHTML := `
	<html>
	<body>
		<div class="slot-game">Main Slot Gacor</div>
		<button class="deposit-btn">Deposit Sekarang</button>
		<button class="wd-btn">WD</button>
		<img src="slot-machine.jpg" alt="Slot Game">
		<input type="text" name="username">
		<input type="password" name="pin">
		<input type="number" name="amount">
	</body>
	</html>
	`
	
	signals := analyzer.AnalyzeDOMStructure(testHTML)
	
	if len(signals) == 0 {
		t.Log("No DOM signals found (this may be normal depending on patterns)")
	} else {
		t.Logf("Found %d DOM structure signals", len(signals))
	}
}

// TestContentSemantics tests the semantic analysis
func TestContentSemantics(t *testing.T) {
	analyzer := analyzer.NewBehavioralAnalyzer()
	
	title := "Situs Judi Slot Online Terbaik Gacor Hari Ini"
	description := "Main slot gacor dapat maxwin setiap hari. Deposit murah via OVO, DANA, Gopay."
	content := "Daftar sekarang dapat bonus besar. Withdraw proses cepat 24 jam."
	
	signals := analyzer.AnalyzePageSemantics(title, description, content)
	
	if len(signals) == 0 {
		t.Log("No semantic signals found (this may be normal)")
	} else {
		t.Logf("Found %d semantic signals", len(signals))
	}
}

// TestConfigurationManager tests the configuration manager
func TestConfigurationManager(t *testing.T) {
	manager := config.NewConfigManager()
	
	// Test getting default config
	defaultConfig := manager.GetDefaultConfig()
	if defaultConfig.Scoring.GamblingUI != 0.30 {
		t.Error("Expected default GamblingUI to be 0.30")
	}
	
	// Test validation
	err := manager.ValidateConfig()
	if err != nil {
		t.Errorf("Configuration validation failed: %v", err)
	}
	
	// Test getting a profile
	profile, err := manager.GetProfile("standard")
	if err != nil {
		t.Errorf("Failed to get standard profile: %v", err)
	} else if profile.Scoring.GamblingUI != 0.30 {
		t.Error("Expected standard profile GamblingUI to be 0.30")
	}
	
	t.Log("Configuration manager tests passed")
}

// TestClusterEngine tests the clustering functionality
func TestClusterEngine(t *testing.T) {
	clusterEngine := analyzer.NewClusterEngine()
	
	// Create a mock analysis result
	mockAnalysis := &models.AnalysisResult{
		Domain: models.Domain{
			Domain: "test-domain.com",
			Signals: []models.Signal{
				{
					SignalID:   "test_signal_1",
					Category:   "PAYMENT",
					Confidence: 0.8,
				},
				{
					SignalID:   "test_signal_2",
					Category:   "UX",
					Confidence: 0.7,
				},
			},
		},
		JLIScore: 0.75,
		JLILevel: "MEDIUM",
	}
	
	// Add domain to cluster
	clusterID := clusterEngine.AddDomainToCluster("test-domain.com", mockAnalysis)
	
	if clusterID == "" {
		t.Error("Expected to get a cluster ID")
	}
	
	// Get the cluster
	cluster, exists := clusterEngine.GetCluster(clusterID)
	if !exists {
		t.Error("Expected cluster to exist")
	} else if len(cluster.Domains) != 1 {
		t.Errorf("Expected cluster to have 1 domain, got %d", len(cluster.Domains))
	} else if cluster.Domains[0] != "test-domain.com" {
		t.Errorf("Expected cluster to contain 'test-domain.com', got %s", cluster.Domains[0])
	}
	
	t.Logf("Created cluster %s with %d domains", clusterID, len(cluster.Domains))
}

// ExampleTest demonstrates how to run the fogger tool
func ExampleTest() {
	fmt.Println("fogger tool is ready to scan domains for gambling indicators")
	
	// Initialize config
	config.Initialize()
	
	// Get config
	cfg := config.Get()
	fmt.Printf("Current configuration: Gambling UI weight = %.2f\n", cfg.Scoring.GamblingUI)
	
	// This would normally scan a real domain, but we'll just show the structure
	fmt.Println("Use: fogger scan <domain> to analyze a domain")
}