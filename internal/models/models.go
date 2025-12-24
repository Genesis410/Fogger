package models

import (
	"time"
)

// Domain represents a domain entity with analysis results
type Domain struct {
	Domain      string    `json:"domain"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	CDNProvider string    `json:"cdn_provider"`
	JLIScore    float64   `json:"jli_score"`
	JLILevel    string    `json:"jli_level"`
	ClusterID   *string   `json:"cluster_id"`
	Signals     []Signal  `json:"signals"`
}

// Signal represents an atomic signal found during analysis
type Signal struct {
	SignalID    string     `json:"signal_id"`
	Category    string     `json:"category"`
	Description string     `json:"description"`
	Confidence  float64    `json:"confidence"`
	Evidence    []Evidence `json:"evidence"`
}

// Evidence represents human-auditable evidence for a signal
type Evidence struct {
	Type      string    `json:"type"`
	Reference string    `json:"reference"`
	Timestamp time.Time `json:"timestamp"`
}

// ScoringProfile represents a configuration profile for scoring
type ScoringProfile struct {
	Name       string             `json:"name"`
	Weights    map[string]float64 `json:"weights"`
	Thresholds ThresholdConfig    `json:"thresholds"`
}

// ThresholdConfig holds classification thresholds
type ThresholdConfig struct {
	High   float64 `json:"high"`
	Medium float64 `json:"medium"`
}

// AnalysisResult holds the complete analysis result
type AnalysisResult struct {
	Domain        Domain            `json:"domain"`
	JLIScore      float64           `json:"jli_score"`
	JLILevel      string            `json:"jli_level"`
	CategoryBreakdown map[string]CategoryBreakdown `json:"category_breakdown"`
	ProfileUsed   string            `json:"profile_used"`
}

// CategoryBreakdown holds the breakdown of scores by category
type CategoryBreakdown struct {
	Score        float64 `json:"score"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
}