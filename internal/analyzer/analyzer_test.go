package analyzer

import (
	"testing"
)

// TestAnalyzerInitialization tests that the analyzer package initializes correctly
func TestAnalyzerInitialization(t *testing.T) {
	// Just test that we can call a function without error
	behavioralAnalyzer := NewBehavioralAnalyzer()
	if behavioralAnalyzer == nil {
		t.Error("Expected to create behavioral analyzer successfully")
	}
	
	if len(behavioralAnalyzer.GamblingKeywords) == 0 {
		t.Error("Expected behavioral analyzer to have gambling keywords")
	}
	
	if len(behavioralAnalyzer.PaymentKeywords) == 0 {
		t.Error("Expected behavioral analyzer to have payment keywords")
	}
	
	if len(behavioralAnalyzer.RegexPatterns) == 0 {
		t.Error("Expected behavioral analyzer to have compiled regex patterns")
	}
	
	t.Log("Analyzer initialization test passed")
}