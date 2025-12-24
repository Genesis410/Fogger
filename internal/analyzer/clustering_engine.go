package analyzer

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// ClusterEngine handles domain clustering and attribution
type ClusterEngine struct {
	Clusters map[string]*Cluster
}

// Cluster represents a group of related domains
type Cluster struct {
	ID              string            `json:"cluster_id"`
	Confidence      float64           `json:"confidence"`
	Domains         []string          `json:"domains"`
	SharedSignals   []string          `json:"shared_signals"`
	FirstSeen       time.Time         `json:"first_seen"`
	LastSeen        time.Time         `json:"last_seen"`
	SharedResources map[string]string `json:"shared_resources"` // IPs, wallets, etc.
}

// NewClusterEngine creates a new clustering engine
func NewClusterEngine() *ClusterEngine {
	return &ClusterEngine{
		Clusters: make(map[string]*Cluster),
	}
}

// AddDomainToCluster adds a domain to an appropriate cluster based on similarities
func (ce *ClusterEngine) AddDomainToCluster(domain string, analysis *models.AnalysisResult) string {
	// Calculate similarity with existing clusters
	bestClusterID := ce.findBestCluster(analysis)
	
	if bestClusterID != "" {
		// Add domain to existing cluster
		cluster := ce.Clusters[bestClusterID]
		cluster.Domains = append(cluster.Domains, domain)
		cluster.LastSeen = time.Now()
		
		// Update shared resources if needed
		ce.updateSharedResources(cluster, analysis)
		
		return bestClusterID
	}
	
	// Create new cluster
	clusterID := ce.generateClusterID(domain, analysis)
	newCluster := &Cluster{
		ID:              clusterID,
		Confidence:      1.0, // New cluster has high confidence initially
		Domains:         []string{domain},
		SharedSignals:   ce.extractSharedSignals(analysis),
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		SharedResources: ce.extractSharedResources(analysis),
	}
	
	ce.Clusters[clusterID] = newCluster
	return clusterID
}

// findBestCluster finds the most similar cluster for a domain
func (ce *ClusterEngine) findBestCluster(analysis *models.AnalysisResult) string {
	if len(ce.Clusters) == 0 {
		return ""
	}
	
	var bestClusterID string
	bestScore := 0.0
	
	for clusterID, cluster := range ce.Clusters {
		score := ce.calculateClusterSimilarity(cluster, analysis)
		if score > bestScore && score >= 0.5 { // Threshold for clustering
			bestScore = score
			bestClusterID = clusterID
		}
	}
	
	return bestClusterID
}

// calculateClusterSimilarity calculates similarity between a cluster and an analysis result
func (ce *ClusterEngine) calculateClusterSimilarity(cluster *Cluster, analysis *models.AnalysisResult) float64 {
	score := 0.0
	
	// Check for shared signals
	sharedSignalCount := 0
	analysisSignals := ce.extractSignalCategories(analysis)
	
	for _, clusterSignal := range cluster.SharedSignals {
		for _, analysisSignal := range analysisSignals {
			if clusterSignal == analysisSignal {
				sharedSignalCount++
				break
			}
		}
	}
	
	if len(cluster.SharedSignals) > 0 {
		score += float64(sharedSignalCount) / float64(len(cluster.SharedSignals)) * 0.4
	}
	
	// Check for shared resources
	sharedResourceCount := 0
	analysisResources := ce.extractSharedResources(analysis)
	
	for resType, resValue := range analysisResources {
		if clusterRes, exists := cluster.SharedResources[resType]; exists {
			if clusterRes == resValue {
				sharedResourceCount++
			}
		}
	}
	
	if len(analysisResources) > 0 {
		score += float64(sharedResourceCount) / float64(len(analysisResources)) * 0.6
	}
	
	return score
}

// extractSignalCategories extracts signal categories from analysis
func (ce *ClusterEngine) extractSignalCategories(analysis *models.AnalysisResult) []string {
	signalMap := make(map[string]bool)
	
	for _, signal := range analysis.Domain.Signals {
		signalMap[signal.Category] = true
	}
	
	var categories []string
	for category := range signalMap {
		categories = append(categories, category)
	}
	
	return categories
}

// extractSharedSignals extracts signals that are likely to be shared across domains
func (ce *ClusterEngine) extractSharedSignals(analysis *models.AnalysisResult) []string {
	var sharedSignals []string
	
	// Look for signals that are likely to be consistent across related domains
	for _, signal := range analysis.Domain.Signals {
		// Focus on infrastructure and payment signals that might be shared
		if signal.Category == "INFRA" || signal.Category == "PAYMENT" {
			sharedSignals = append(sharedSignals, signal.SignalID)
		}
	}
	
	return sharedSignals
}

// extractSharedResources extracts resources that might be shared across domains
func (ce *ClusterEngine) extractSharedResources(analysis *models.AnalysisResult) map[string]string {
	resources := make(map[string]string)
	
	for _, signal := range analysis.Domain.Signals {
		// Look for IP addresses, wallets, or other shared infrastructure
		if signal.Category == "INFRA" && strings.Contains(signal.Description, "origin IP") {
			// Extract IP from description
			ip := ce.extractIPFromDescription(signal.Description)
			if ip != "" {
				resources["ip"] = ip
			}
		} else if signal.Category == "PAYMENT" && strings.Contains(signal.Description, "cryptocurrency address") {
			// Extract wallet address
			wallet := ce.extractWalletFromDescription(signal.Description)
			if wallet != "" {
				resources["wallet"] = wallet
			}
		}
	}
	
	return resources
}

// extractIPFromDescription extracts IP address from signal description
func (ce *ClusterEngine) extractIPFromDescription(desc string) string {
	// Simple extraction - in real implementation, use regex
	parts := strings.Fields(desc)
	for _, part := range parts {
		if net.ParseIP(part) != nil {
			return part
		}
	}

	return ""
}

// extractWalletFromDescription extracts wallet address from signal description
func (ce *ClusterEngine) extractWalletFromDescription(desc string) string {
	// Look for wallet address in description
	if colonIndex := strings.Index(desc, ":"); colonIndex != -1 {
		parts := strings.Split(desc[colonIndex+1:], " ")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			// Simple check for common wallet patterns
			if len(trimmed) > 20 && len(trimmed) < 50 { // Typical wallet length
				return trimmed
			}
		}
	}
	
	return ""
}

// generateClusterID generates a unique ID for a cluster
func (ce *ClusterEngine) generateClusterID(domain string, analysis *models.AnalysisResult) string {
	// Create a hash based on domain and key signals
	data := domain
	
	// Add important signals to the hash
	for _, signal := range analysis.Domain.Signals {
		if signal.Category == "PAYMENT" || signal.Category == "INFRA" {
			data += signal.SignalID
		}
	}
	
	// Create MD5 hash
	hasher := md5.New()
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))[:12] // Use first 12 chars
}

// updateSharedResources updates the shared resources of a cluster
func (ce *ClusterEngine) updateSharedResources(cluster *Cluster, analysis *models.AnalysisResult) {
	newResources := ce.extractSharedResources(analysis)
	
	for resType, resValue := range newResources {
		if _, exists := cluster.SharedResources[resType]; !exists {
			cluster.SharedResources[resType] = resValue
		}
	}
}

// GetCluster retrieves a cluster by ID
func (ce *ClusterEngine) GetCluster(clusterID string) (*Cluster, bool) {
	cluster, exists := ce.Clusters[clusterID]
	return cluster, exists
}

// GetAllClusters returns all clusters
func (ce *ClusterEngine) GetAllClusters() []*Cluster {
	var clusters []*Cluster
	for _, cluster := range ce.Clusters {
		clusters = append(clusters, cluster)
	}
	
	// Sort by last seen (most recent first)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].LastSeen.After(clusters[j].LastSeen)
	})
	
	return clusters
}

// GetClusterForDomain finds the cluster for a specific domain
func (ce *ClusterEngine) GetClusterForDomain(domain string) (*Cluster, bool) {
	for _, cluster := range ce.Clusters {
		for _, clusterDomain := range cluster.Domains {
			if clusterDomain == domain {
				return cluster, true
			}
		}
	}
	
	return nil, false
}

// GetClustersByConfidence returns clusters sorted by confidence
func (ce *ClusterEngine) GetClustersByConfidence() []*Cluster {
	clusters := ce.GetAllClusters()
	
	// Sort by confidence (highest first)
	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Confidence > clusters[j].Confidence
	})
	
	return clusters
}

// FindClustersByResource finds clusters that share a specific resource
func (ce *ClusterEngine) FindClustersByResource(resourceType, resourceValue string) []*Cluster {
	var matchingClusters []*Cluster
	
	for _, cluster := range ce.Clusters {
		if res, exists := cluster.SharedResources[resourceType]; exists && res == resourceValue {
			matchingClusters = append(matchingClusters, cluster)
		}
	}
	
	return matchingClusters
}

// UpdateClusterConfidence recalculates the confidence of a cluster
func (ce *ClusterEngine) UpdateClusterConfidence(clusterID string) {
	cluster, exists := ce.Clusters[clusterID]
	if !exists {
		return
	}
	
	// Calculate confidence based on various factors:
	// - Number of domains in cluster
	// - Number of shared signals
	// - Number of shared resources
	// - Recency of activity
	
	domainCount := len(cluster.Domains)
	signalCount := len(cluster.SharedSignals)
	resourceCount := len(cluster.SharedResources)
	
	// Base confidence on domain count (more domains = higher confidence)
	confidence := float64(domainCount) * 0.3
	
	// Add confidence for shared signals
	confidence += float64(signalCount) * 0.2
	
	// Add confidence for shared resources
	confidence += float64(resourceCount) * 0.3
	
	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	
	cluster.Confidence = confidence
}

// MergeClusters merges two clusters together
func (ce *ClusterEngine) MergeClusters(clusterID1, clusterID2 string) error {
	cluster1, exists1 := ce.Clusters[clusterID1]
	cluster2, exists2 := ce.Clusters[clusterID2]
	
	if !exists1 || !exists2 {
		return fmt.Errorf("one or both clusters do not exist")
	}
	
	// Merge domains
	for _, domain := range cluster2.Domains {
		// Check if domain already exists in cluster1
		found := false
		for _, existingDomain := range cluster1.Domains {
			if existingDomain == domain {
				found = true
				break
			}
		}
		if !found {
			cluster1.Domains = append(cluster1.Domains, domain)
		}
	}
	
	// Merge shared signals
	for _, signal := range cluster2.SharedSignals {
		found := false
		for _, existingSignal := range cluster1.SharedSignals {
			if existingSignal == signal {
				found = true
				break
			}
		}
		if !found {
			cluster1.SharedSignals = append(cluster1.SharedSignals, signal)
		}
	}
	
	// Merge shared resources
	for resType, resValue := range cluster2.SharedResources {
		cluster1.SharedResources[resType] = resValue
	}
	
	// Update confidence and timestamps
	cluster1.Confidence = (cluster1.Confidence + cluster2.Confidence) / 2
	if cluster2.LastSeen.After(cluster1.LastSeen) {
		cluster1.LastSeen = cluster2.LastSeen
	}
	
	// Remove the second cluster
	delete(ce.Clusters, clusterID2)
	
	return nil
}

// GetClusterStatistics returns statistics about clustering
func (ce *ClusterEngine) GetClusterStatistics() map[string]interface{} {
	stats := make(map[string]interface{})
	
	totalClusters := len(ce.Clusters)
	totalDomains := 0
	highConfidenceClusters := 0
	
	for _, cluster := range ce.Clusters {
		totalDomains += len(cluster.Domains)
		if cluster.Confidence >= 0.7 {
			highConfidenceClusters++
		}
	}
	
	stats["total_clusters"] = totalClusters
	stats["total_domains"] = totalDomains
	stats["high_confidence_clusters"] = highConfidenceClusters
	stats["avg_domains_per_cluster"] = 0.0
	if totalClusters > 0 {
		stats["avg_domains_per_cluster"] = float64(totalDomains) / float64(totalClusters)
	}
	
	return stats
}