package detector

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// OriginIPDetector detects potential origin IPs behind CDNs
type OriginIPDetector struct {
	Client *http.Client
}

// NewOriginIPDetector creates a new instance of OriginIPDetector
func NewOriginIPDetector() *OriginIPDetector {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	return &OriginIPDetector{
		Client: client,
	}
}

// DetectOriginIPs attempts to find origin IPs for a domain
func (d *OriginIPDetector) DetectOriginIPs(domain string) ([]string, []models.Evidence, error) {
	var originIPs []string
	var evidence []models.Evidence

	// Method 1: Check subdomains that might not be behind CDN
	subdomainIPs, subEvidence := d.checkSubdomains(domain)
	originIPs = append(originIPs, subdomainIPs...)
	evidence = append(evidence, subEvidence...)

	// Method 2: Check historical DNS records (simplified)
	historicalIPs, histEvidence := d.checkHistoricalDNS(domain)
	originIPs = append(originIPs, historicalIPs...)
	evidence = append(evidence, histEvidence...)

	// Method 3: Check for mail servers (MX records) which might be on same infra
	mxIPs, mxEvidence := d.checkMXRecords(domain)
	originIPs = append(originIPs, mxIPs...)
	evidence = append(evidence, mxEvidence...)

	// Method 4: Check for other DNS records that might reveal origin
	otherIPs, otherEvidence := d.checkOtherDNSRecords(domain)
	originIPs = append(originIPs, otherIPs...)
	evidence = append(evidence, otherEvidence...)

	// Remove duplicates
	uniqueIPs := removeDuplicates(append(append(append(subdomainIPs, historicalIPs...), mxIPs...), otherIPs...))

	return uniqueIPs, evidence, nil
}

// checkSubdomains checks common subdomains that might not be CDN-protected
func (d *OriginIPDetector) checkSubdomains(domain string) ([]string, []models.Evidence) {
	var ips []string
	var evidence []models.Evidence
	
	subdomains := []string{
		"mail", "webmail", "autodiscover", "autoconfig", 
		"cpanel", "whm", "ftp", "smtp", "pop", "imap",
		"ns1", "ns2", "ns3", "ns4", "dns1", "dns2",
		"dev", "staging", "test", "admin", "api",
		"shop", "blog", "m", "mobile", "api", "cdn",
		"img", "images", "static", "media", "video",
	}

	for _, subdomain := range subdomains {
		fullDomain := fmt.Sprintf("%s.%s", subdomain, domain)
		
		// Resolve the subdomain to IP
		ip, err := net.ResolveIPAddr("ip4", fullDomain)
		if err != nil {
			continue
		}
		
		// Check if this subdomain is behind the same CDN
		isBehindCDN := d.isBehindCDN(fullDomain)
		
		// If not behind CDN, this might be the origin IP
		if !isBehindCDN {
			ips = append(ips, ip.String())
			evidence = append(evidence, models.Evidence{
				Type:      "dns",
				Reference: fmt.Sprintf("Subdomain %s resolves to IP %s (not behind CDN)", fullDomain, ip.String()),
				Timestamp: time.Now(),
			})
		}
	}
	
	return ips, evidence
}

// checkHistoricalDNS checks for historical DNS records (simulated)
func (d *OriginIPDetector) checkHistoricalDNS(domain string) ([]string, []models.Evidence) {
	var ips []string
	var evidence []models.Evidence

	// In a real implementation, this would query passive DNS services like:
	// - CIRCL Passive DNS
	// - DNSDB
	// - RiskIQ
	// - SecurityTrails
	// - etc.
	
	// For this example, we'll simulate checking historical records
	// This is a simplified approach
	
	// Resolve current domain
	currentIPs, err := net.LookupIP(domain)
	if err != nil {
		return ips, evidence
	}

	// In real implementation, we'd compare these with historical records
	// to find when the domain was not behind CDN
	
	for _, ip := range currentIPs {
		if ip.To4() != nil { // IPv4 only
			ips = append(ips, ip.String())
			evidence = append(evidence, models.Evidence{
				Type:      "dns",
				Reference: fmt.Sprintf("Current DNS record for %s points to IP %s", domain, ip.String()),
				Timestamp: time.Now(),
			})
		}
	}

	return ips, evidence
}

// checkMXRecords checks mail server records which might be on same infrastructure
func (d *OriginIPDetector) checkMXRecords(domain string) ([]string, []models.Evidence) {
	var ips []string
	var evidence []models.Evidence

	mxRecords, err := net.LookupMX(domain)
	if err != nil {
		return ips, evidence
	}

	for _, mx := range mxRecords {
		// Resolve the MX host to IP
		mxIPs, err := net.LookupIP(mx.Host)
		if err != nil {
			continue
		}

		for _, ip := range mxIPs {
			if ip.To4() != nil { // IPv4 only
				ips = append(ips, ip.String())
				evidence = append(evidence, models.Evidence{
					Type:      "dns",
					Reference: fmt.Sprintf("MX record %s for %s resolves to IP %s", mx.Host, domain, ip.String()),
					Timestamp: time.Now(),
				})
			}
		}
	}

	return ips, evidence
}

// checkOtherDNSRecords checks other DNS records that might reveal origin
func (d *OriginIPDetector) checkOtherDNSRecords(domain string) ([]string, []models.Evidence) {
	var ips []string
	var evidence []models.Evidence

	// Check for SRV records
	// Check for TXT records that might contain IP addresses
	// Check for A records of related services
	
	// SRV records
	serviceNames := []string{
		"_sip._tcp", "_sip._tls", "_sips._tcp", 
		"_xmpp-client._tcp", "_xmpp-server._tcp",
		"_ftp._tcp", "_ssh._tcp",
	}

	for _, service := range serviceNames {
		serviceDomain := fmt.Sprintf("%s.%s", service, domain)
		_, addrs, err := net.LookupSRV("", "", serviceDomain)
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Resolve the target to IP
			ip, err := net.ResolveIPAddr("ip4", strings.TrimSuffix(addr.Target, "."))
			if err != nil {
				continue
			}
			
			ips = append(ips, ip.String())
			evidence = append(evidence, models.Evidence{
				Type:      "dns",
				Reference: fmt.Sprintf("SRV record %s points to %s which resolves to IP %s", serviceDomain, addr.Target, ip.String()),
				Timestamp: time.Now(),
			})
		}
	}

	// Check for TXT records that might contain IP addresses
	txtRecords, err := net.LookupTXT(domain)
	if err == nil {
		ipRegex := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
		for _, txt := range txtRecords {
			matches := ipRegex.FindAllString(txt, -1)
			for _, ip := range matches {
				// Validate that it's a real IP
				if net.ParseIP(ip) != nil {
					ips = append(ips, ip)
					evidence = append(evidence, models.Evidence{
						Type:      "dns",
						Reference: fmt.Sprintf("TXT record contains potential IP: %s", ip),
						Timestamp: time.Now(),
					})
				}
			}
		}
	}

	return ips, evidence
}

// isBehindCDN checks if a domain is behind a CDN
func (d *OriginIPDetector) isBehindCDN(domain string) bool {
	// Make a request to the domain
	url := fmt.Sprintf("http://%s", domain)
	resp, err := d.Client.Get(url)
	if err != nil {
		// Try HTTPS
		url = fmt.Sprintf("https://%s", domain)
		resp, err = d.Client.Get(url)
		if err != nil {
			return true // Assume CDN if we can't connect
		}
	}
	defer resp.Body.Close()

	// Check for CDN-specific headers
	return d.checkCDNHeaders(resp.Header) || d.checkCDNCertificates(resp.TLS)
}

// checkCDNHeaders checks response headers for CDN indicators
func (d *OriginIPDetector) checkCDNHeaders(headers http.Header) bool {
	// Cloudflare headers
	if headers.Get("server") == "cloudflare" ||
		headers.Get("cf-ray") != "" ||
		headers.Get("cf-request-id") != "" {
		return true
	}

	// CloudFront headers
	if strings.Contains(headers.Get("x-cache"), "cloudfront") ||
		headers.Get("x-amz-cf-pop") != "" {
		return true
	}

	// Akamai headers
	if headers.Get("x-akamai-transformed") != "" ||
		headers.Get("server") == "AkamaiGHost" {
		return true
	}

	// Other common CDN headers
	if strings.Contains(headers.Get("via"), "cloudflare") ||
		strings.Contains(headers.Get("via"), "amazon") ||
		strings.Contains(headers.Get("via"), "akamai") {
		return true
	}

	return false
}

// checkCDNCertificates checks if the TLS certificate indicates CDN usage
func (d *OriginIPDetector) checkCDNCertificates(connState *tls.ConnectionState) bool {
	if connState == nil {
		return false
	}

	for _, cert := range connState.PeerCertificates {
		// Check if certificate contains CDN-related strings
		if strings.Contains(strings.ToLower(cert.Subject.CommonName), "cloudflaressl") ||
			strings.Contains(strings.ToLower(cert.Subject.CommonName), "cloudflare") {
			return true
		}

		for _, name := range cert.DNSNames {
			if strings.Contains(strings.ToLower(name), "cloudflaressl") ||
				strings.Contains(strings.ToLower(name), "cloudflare") {
				return true
			}
		}
	}

	return false
}

// removeDuplicates removes duplicate IPs from a slice
func removeDuplicates(ipList []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, ip := range ipList {
		if !seen[ip] {
			seen[ip] = true
			result = append(result, ip)
		}
	}

	return result
}

// CheckDomainCDNStatus checks if a domain is protected by CDN
func (d *OriginIPDetector) CheckDomainCDNStatus(domain string) string {
	if d.isBehindCDN(domain) {
		// Try to identify which CDN
		url := fmt.Sprintf("https://%s", domain)
		resp, err := d.Client.Get(url)
		if err != nil {
			return "unknown"
		}
		defer resp.Body.Close()

		if d.checkCDNHeaders(resp.Header) {
			if resp.Header.Get("server") == "cloudflare" || 
				resp.Header.Get("cf-ray") != "" {
				return "cloudflare"
			} else if strings.Contains(resp.Header.Get("x-cache"), "cloudfront") {
				return "cloudfront"
			} else if resp.Header.Get("server") == "AkamaiGHost" {
				return "akamai"
			}
		}
	}

	return "none"
}

// GetCDNProviderDetails returns detailed information about CDN usage
func (d *OriginIPDetector) GetCDNProviderDetails(domain string) (string, map[string]string) {
	cdnStatus := d.CheckDomainCDNStatus(domain)
	details := make(map[string]string)

	if cdnStatus != "none" {
		url := fmt.Sprintf("https://%s", domain)
		resp, err := d.Client.Get(url)
		if err != nil {
			return cdnStatus, details
		}
		defer resp.Body.Close()

		// Extract CDN-specific headers
		if cdnStatus == "cloudflare" {
			if ray := resp.Header.Get("cf-ray"); ray != "" {
				details["cf-ray"] = ray
			}
			if country := resp.Header.Get("cf-ipcountry"); country != "" {
				details["country"] = country
			}
			if resp.Header.Get("server") == "cloudflare" {
				details["server"] = "cloudflare"
			}
		} else if cdnStatus == "cloudfront" {
			if pop := resp.Header.Get("x-amz-cf-pop"); pop != "" {
				details["pop"] = pop
			}
			if id := resp.Header.Get("x-amz-cf-id"); id != "" {
				details["id"] = id
			}
		}
	}

	return cdnStatus, details
}