package analyzer

import (
	"fmt"
	"sync"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// Monitor tracks changes to domains over time
type Monitor struct {
	domains    map[string]*DomainMonitor
	mu         sync.RWMutex
	exporter   *Exporter
}

// DomainMonitor holds monitoring state for a single domain
type DomainMonitor struct {
	Domain     string
	LastResult *models.AnalysisResult
	Changes    []ChangeRecord
	Interval   time.Duration
	Active     bool
	StopChan   chan bool
}

// ChangeRecord records a change in domain analysis
type ChangeRecord struct {
	Timestamp time.Time
	OldScore  float64
	NewScore  float64
	Reason    string
	Signals   []models.Signal
}

// NewMonitor creates a new monitoring instance
func NewMonitor() *Monitor {
	return &Monitor{
		domains:  make(map[string]*DomainMonitor),
		exporter: NewExporter(),
	}
}

// AddDomain adds a domain to monitoring
func (m *Monitor) AddDomain(domain string, interval time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if _, exists := m.domains[domain]; exists {
		return fmt.Errorf("domain %s is already being monitored", domain)
	}
	
	monitor := &DomainMonitor{
		Domain:   domain,
		Changes:  make([]ChangeRecord, 0),
		Interval: interval,
		Active:   true,
		StopChan: make(chan bool, 1),
	}
	
	m.domains[domain] = monitor
	
	// Perform initial scan
	result := AnalyzeDomain(domain, 10*time.Second, "standard")
	monitor.LastResult = result
	
	go m.runMonitor(monitor)
	
	return nil
}

// RemoveDomain removes a domain from monitoring
func (m *Monitor) RemoveDomain(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	monitor, exists := m.domains[domain]
	if !exists {
		return fmt.Errorf("domain %s is not being monitored", domain)
	}
	
	monitor.Active = false
	monitor.StopChan <- true
	
	delete(m.domains, domain)
	
	return nil
}

// runMonitor runs the monitoring loop for a domain
func (m *Monitor) runMonitor(monitor *DomainMonitor) {
	ticker := time.NewTicker(monitor.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if !monitor.Active {
				return
			}
			
			// Perform analysis
			result := AnalyzeDomain(monitor.Domain, 10*time.Second, "standard")
			
			// Check for changes
			if monitor.LastResult != nil {
				changes := m.detectChanges(monitor.LastResult, result)
				if len(changes) > 0 {
					changeRecord := ChangeRecord{
						Timestamp: time.Now(),
						OldScore:  monitor.LastResult.JLIScore,
						NewScore:  result.JLIScore,
						Reason:    fmt.Sprintf("%d changes detected", len(changes)),
						Signals:   changes,
					}
					monitor.Changes = append(monitor.Changes, changeRecord)
					
					// Log change
					fmt.Printf("Change detected for %s: JLI changed from %.3f to %.3f\n", 
						monitor.Domain, monitor.LastResult.JLIScore, result.JLIScore)
				}
			}
			
			// Update last result
			monitor.LastResult = result
			
		case <-monitor.StopChan:
			return
		}
	}
}

// detectChanges detects changes between two analysis results
func (m *Monitor) detectChanges(oldResult, newResult *models.AnalysisResult) []models.Signal {
	var changes []models.Signal
	
	// Compare signals
	oldSignals := make(map[string]models.Signal)
	for _, signal := range oldResult.Domain.Signals {
		oldSignals[signal.SignalID] = signal
	}
	
	for _, newSignal := range newResult.Domain.Signals {
		if _, exists := oldSignals[newSignal.SignalID]; !exists {
			// New signal detected
			changes = append(changes, newSignal)
		}
	}
	
	// Check for significant score changes
	scoreDiff := newResult.JLIScore - oldResult.JLIScore
	if scoreDiff > 0.1 || scoreDiff < -0.1 { // 10% threshold
		changes = append(changes, models.Signal{
			SignalID:    "jli_score_change",
			Category:    "MONITOR",
			Description: fmt.Sprintf("JLI score changed from %.3f to %.3f", oldResult.JLIScore, newResult.JLIScore),
			Confidence:  1.0,
		})
	}
	
	return changes
}

// GetChanges returns changes for a domain
func (m *Monitor) GetChanges(domain string) ([]ChangeRecord, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	monitor, exists := m.domains[domain]
	if !exists {
		return nil, fmt.Errorf("domain %s is not being monitored", domain)
	}
	
	return monitor.Changes, nil
}

// GetAllMonitoredDomains returns all monitored domains
func (m *Monitor) GetAllMonitoredDomains() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	domains := make([]string, 0, len(m.domains))
	for domain := range m.domains {
		domains = append(domains, domain)
	}
	
	return domains
}

// GetDomainStatus returns the status of a monitored domain
func (m *Monitor) GetDomainStatus(domain string) (*DomainMonitor, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	monitor, exists := m.domains[domain]
	if !exists {
		return nil, fmt.Errorf("domain %s is not being monitored", domain)
	}
	
	return monitor, nil
}

// ExportChanges exports change records to file
func (m *Monitor) ExportChanges(domain, format, filename string) error {
	changes, err := m.GetChanges(domain)
	if err != nil {
		return err
	}
	
	// Convert changes to results for export
	var results []*models.AnalysisResult
	for _, change := range changes {
		// Create a dummy result for export purposes
		result := &models.AnalysisResult{
			Domain: models.Domain{
				Domain: domain,
				Signals: change.Signals,
			},
			JLIScore: change.NewScore,
		}
		results = append(results, result)
	}
	
	switch format {
	case "json":
		return m.exporter.ExportJSON(results, filename)
	case "csv":
		return m.exporter.ExportCSV(results, filename)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

// PauseMonitoring pauses monitoring for a domain
func (m *Monitor) PauseMonitoring(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	monitor, exists := m.domains[domain]
	if !exists {
		return fmt.Errorf("domain %s is not being monitored", domain)
	}
	
	monitor.Active = false
	return nil
}

// ResumeMonitoring resumes monitoring for a domain
func (m *Monitor) ResumeMonitoring(domain string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	monitor, exists := m.domains[domain]
	if !exists {
		return fmt.Errorf("domain %s is not being monitored", domain)
	}
	
	monitor.Active = true
	return nil
}

// StopAll stops all monitoring
func (m *Monitor) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, monitor := range m.domains {
		monitor.Active = false
		monitor.StopChan <- true
	}
	
	// Clear the map
	m.domains = make(map[string]*DomainMonitor)
}