package detector

import (
	"crypto/tls"
	"net/http"
	"strings"
	"time"
)

// CDNDetector provides advanced CDN detection capabilities
type CDNDetector struct {
	Client *http.Client
}

// NewCDNDetector creates a new instance of CDNDetector
func NewCDNDetector() *CDNDetector {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	return &CDNDetector{
		Client: client,
	}
}

// CDNInfo holds information about detected CDN
type CDNInfo struct {
	Name     string            `json:"name"`
	Version  string            `json:"version"`
	Features map[string]string `json:"features"`
}

// DetectCDN identifies which CDN is being used by a domain
func (c *CDNDetector) DetectCDN(domain string) *CDNInfo {
	// Ensure domain has proper scheme
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	resp, err := c.Client.Get(url)
	if err != nil {
		// Try HTTP if HTTPS fails
		url = strings.Replace(url, "https://", "http://", 1)
		resp, err = c.Client.Get(url)
		if err != nil {
			return &CDNInfo{Name: "unknown", Features: make(map[string]string)}
		}
	}
	defer resp.Body.Close()

	return c.analyzeResponse(resp)
}

// analyzeResponse analyzes HTTP response to detect CDN
func (c *CDNDetector) analyzeResponse(resp *http.Response) *CDNInfo {
	headers := resp.Header
	cdnInfo := &CDNInfo{
		Name:     "none",
		Features: make(map[string]string),
	}

	// Check for Cloudflare
	if c.isCloudflare(headers, resp.TLS) {
		cdnInfo.Name = "cloudflare"
		cdnInfo.Features["server"] = headers.Get("server")
		cdnInfo.Features["cf-ray"] = headers.Get("cf-ray")
		cdnInfo.Features["cf-request-id"] = headers.Get("cf-request-id")
		cdnInfo.Features["cf-cache-status"] = headers.Get("cf-cache-status")
		return cdnInfo
	}

	// Check for CloudFront
	if c.isCloudFront(headers) {
		cdnInfo.Name = "cloudfront"
		cdnInfo.Features["x-cache"] = headers.Get("x-cache")
		cdnInfo.Features["x-amz-cf-pop"] = headers.Get("x-amz-cf-pop")
		cdnInfo.Features["x-amz-cf-id"] = headers.Get("x-amz-cf-id")
		return cdnInfo
	}

	// Check for Akamai
	if c.isAkamai(headers) {
		cdnInfo.Name = "akamai"
		cdnInfo.Features["server"] = headers.Get("server")
		cdnInfo.Features["x-akamai-transformed"] = headers.Get("x-akamai-transformed")
		return cdnInfo
	}

	// Check for other common CDNs
	if c.isFastly(headers) {
		cdnInfo.Name = "fastly"
		cdnInfo.Features["x-served-by"] = headers.Get("x-served-by")
		cdnInfo.Features["x-cache"] = headers.Get("x-cache")
		return cdnInfo
	}

	if c.isSquarespace(headers) {
		cdnInfo.Name = "squarespace"
		cdnInfo.Features["server"] = headers.Get("server")
		cdnInfo.Features["x-served-by"] = headers.Get("x-served-by")
		return cdnInfo
	}

	if c.isNetlify(headers) {
		cdnInfo.Name = "netlify"
		cdnInfo.Features["server"] = headers.Get("server")
		cdnInfo.Features["x-nf-request-id"] = headers.Get("x-nf-request-id")
		return cdnInfo
	}

	if c.isGithubPages(headers) {
		cdnInfo.Name = "github-pages"
		cdnInfo.Features["x-github-request-id"] = headers.Get("x-github-request-id")
		cdnInfo.Features["x-proxy-response"] = headers.Get("x-proxy-response")
		return cdnInfo
	}

	return cdnInfo
}

// isCloudflare checks for Cloudflare-specific indicators
func (c *CDNDetector) isCloudflare(headers http.Header, tlsState *tls.ConnectionState) bool {
	// Check headers
	if headers.Get("server") == "cloudflare" ||
		headers.Get("cf-ray") != "" ||
		headers.Get("cf-request-id") != "" ||
		strings.Contains(headers.Get("via"), "cloudflare") {
		return true
	}

	// Check TLS certificate for Cloudflare indicators
	if tlsState != nil {
		for _, cert := range tlsState.PeerCertificates {
			if strings.Contains(strings.ToLower(cert.Subject.CommonName), "cloudflaressl") ||
				strings.Contains(strings.ToLower(cert.Subject.CommonName), "sni.cloudflaressl.com") {
				return true
			}
			for _, name := range cert.DNSNames {
				if strings.Contains(strings.ToLower(name), "cloudflare") {
					return true
				}
			}
		}
	}

	return false
}

// isCloudFront checks for CloudFront-specific indicators
func (c *CDNDetector) isCloudFront(headers http.Header) bool {
	return strings.Contains(headers.Get("x-cache"), "CloudFront") ||
		headers.Get("x-amz-cf-pop") != "" ||
		headers.Get("x-amz-cf-id") != "" ||
		strings.Contains(headers.Get("via"), "CloudFront")
}

// isAkamai checks for Akamai-specific indicators
func (c *CDNDetector) isAkamai(headers http.Header) bool {
	return headers.Get("server") == "AkamaiGHost" ||
		headers.Get("x-akamai-transformed") != "" ||
		strings.Contains(headers.Get("via"), "akamai")
}

// isFastly checks for Fastly-specific indicators
func (c *CDNDetector) isFastly(headers http.Header) bool {
	return headers.Get("x-served-by") != "" ||
		headers.Get("x-cache") != "" ||
		strings.Contains(headers.Get("via"), "fastly")
}

// isSquarespace checks for Squarespace-specific indicators
func (c *CDNDetector) isSquarespace(headers http.Header) bool {
	return strings.Contains(headers.Get("server"), "Squarespace") ||
		headers.Get("x-served-by") != ""
}

// isNetlify checks for Netlify-specific indicators
func (c *CDNDetector) isNetlify(headers http.Header) bool {
	return strings.Contains(headers.Get("server"), "Netlify") ||
		headers.Get("x-nf-request-id") != ""
}

// isGithubPages checks for GitHub Pages-specific indicators
func (c *CDNDetector) isGithubPages(headers http.Header) bool {
	return strings.Contains(headers.Get("server"), "GitHub.com") ||
		headers.Get("x-github-request-id") != "" ||
		headers.Get("x-proxy-response") != ""
}

// GetCDNFingerprint returns detailed fingerprint of CDN usage
func (c *CDNDetector) GetCDNFingerprint(domain string) map[string]interface{} {
	info := c.DetectCDN(domain)
	
	fingerprint := make(map[string]interface{})
	fingerprint["domain"] = domain
	fingerprint["cdn_detected"] = info.Name
	fingerprint["features"] = info.Features
	fingerprint["is_protected"] = info.Name != "none" && info.Name != "unknown"
	
	// Additional checks
	fingerprint["has_ssl"] = c.hasSSL(domain)
	fingerprint["response_time"] = c.getResponseTime(domain)
	
	return fingerprint
}

// hasSSL checks if the domain has SSL/TLS enabled
func (c *CDNDetector) hasSSL(domain string) bool {
	url := domain
	if !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	resp, err := c.Client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	
	return resp.TLS != nil
}

// getResponseTime measures the response time of a domain
func (c *CDNDetector) getResponseTime(domain string) string {
	start := time.Now()
	
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	resp, err := c.Client.Get(url)
	if err != nil {
		return "error"
	}
	defer resp.Body.Close()
	
	elapsed := time.Since(start)
	return elapsed.String()
}

// GetCDNUsagePatterns identifies usage patterns common in gambling sites
func (c *CDNDetector) GetCDNUsagePatterns(domain string) []string {
	var patterns []string
	
	info := c.DetectCDN(domain)
	
	// Check for common patterns in gambling sites
	if info.Name == "cloudflare" {
		// Check for specific Cloudflare features often used by gambling sites
		headers := c.getHeaders(domain)
		
		// Check for security level headers
		if headers.Get("cf-security-level") != "" {
			patterns = append(patterns, "uses-cloudflare-security")
		}
		
		// Check for caching headers
		if strings.Contains(strings.ToLower(headers.Get("cf-cache-status")), "hit") {
			patterns = append(patterns, "aggressive-caching")
		}
		
		// Check for specific page rules that gambling sites might use
		if headers.Get("cf-polished-version") != "" {
			patterns = append(patterns, "image-optimization")
		}
	}
	
	// Check for CDN fingerprinting bypass attempts
	body := c.getBody(domain)
	if c.hasBypassIndicators(body) {
		patterns = append(patterns, "bypass-attempts")
	}
	
	return patterns
}

// getHeaders gets headers for a domain
func (c *CDNDetector) getHeaders(domain string) http.Header {
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	resp, err := c.Client.Get(url)
	if err != nil {
		return http.Header{}
	}
	defer resp.Body.Close()
	
	return resp.Header
}

// getBody gets the response body for a domain
func (c *CDNDetector) getBody(domain string) string {
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	
	resp, err := c.Client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	
	body := make([]byte, 1024) // Read only first 1KB for performance
	resp.Body.Read(body)
	
	return string(body)
}

// hasBypassIndicators checks if the page has indicators of bypass attempts
func (c *CDNDetector) hasBypassIndicators(body string) bool {
	bypassIndicators := []string{
		"bypass",
		"cloudflare",
		"captcha",
		"checking your browser",
		"please enable javascript",
		"enable cookies",
		"you are being redirected",
		"checking your connection",
	}
	
	lowerBody := strings.ToLower(body)
	for _, indicator := range bypassIndicators {
		if strings.Contains(lowerBody, indicator) {
			return true
		}
	}
	
	return false
}

// GetCDNProviderDetails returns detailed information about CDN usage
func (c *CDNDetector) GetCDNProviderDetails(domain string) *CDNInfo {
	return c.DetectCDN(domain)
}