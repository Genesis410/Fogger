package scanner

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/genesis410/fogger/internal/detector"
	"github.com/genesis410/fogger/internal/models"
)

// ScanResult holds the result of a domain scan
type ScanResult struct {
	Domain      string
	CDNProvider string
	Signals     []models.Signal
	StatusCode  int
	Headers     http.Header
	Body        string
}

// ScanDomain performs a scan of the given domain
func ScanDomain(domain string, timeout time.Duration) *ScanResult {
	result := &ScanResult{
		Domain:  domain,
		Signals: []models.Signal{},
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Ensure domain has proper scheme
	url := domain
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		// If HTTPS fails, try HTTP
		url = strings.Replace(url, "https://", "http://", 1)
		resp, err = client.Get(url)
		if err != nil {
			fmt.Printf("Error connecting to %s: %v\n", domain, err)
			return result
		}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		body = []byte{}
	}

	result.StatusCode = resp.StatusCode
	result.Headers = resp.Header
	result.Body = string(body)

	// Detect CDN
	result.CDNProvider = detectCDN(resp.Header)

	// Add signals based on analysis
	result.Signals = append(result.Signals, detectCDNSignals(result.CDNProvider)...)
	result.Signals = append(result.Signals, detectGamblingUXSignals(result.Body)...)
	result.Signals = append(result.Signals, detectPaymentSignals(result.Body)...)
	result.Signals = append(result.Signals, detectInfrastructureSignals(resp.Header)...)

	// Try to detect origin IPs behind CDN
	originIPs, originEvidence, err := detectOriginIPs(domain)
	if err == nil && len(originIPs) > 0 {
		// Add signals for detected origin IPs
		for _, ip := range originIPs {
			signal := models.Signal{
				SignalID:    "origin_ip_detected",
				Category:    "INFRA",
				Description: fmt.Sprintf("Potential origin IP detected behind CDN: %s", ip),
				Confidence:  0.8,
				Evidence:    originEvidence,
			}
			result.Signals = append(result.Signals, signal)
			break // Only add one to avoid spamming
		}
	}

	return result
}

// detectCDN detects which CDN is being used
func detectCDN(headers http.Header) string {
	// Check for Cloudflare headers
	if headers.Get("server") == "cloudflare" ||
		headers.Get("cf-ray") != "" ||
		headers.Get("cf-request-id") != "" {
		return "cloudflare"
	}

	// Check for other common CDN headers
	if headers.Get("x-cache") != "" && strings.Contains(headers.Get("x-cache"), "cloudfront") {
		return "cloudfront"
	}

	if headers.Get("server") == "AkamaiGHost" ||
		headers.Get("x-akamai-transformed") != "" {
		return "akamai"
	}

	// Check for other CDNs
	if headers.Get("x-served-by") != "" || strings.Contains(headers.Get("via"), "fastly") {
		return "fastly"
	}

	if strings.Contains(headers.Get("server"), "Netlify") || headers.Get("x-nf-request-id") != "" {
		return "netlify"
	}

	return "none"
}

// detectCDNSignals returns signals related to CDN usage
func detectCDNSignals(cdnProvider string) []models.Signal {
	signals := []models.Signal{}

	if cdnProvider == "cloudflare" {
		signal := models.Signal{
			SignalID:    "cdn_cloudflare",
			Category:    "CDN",
			Description: "Domain is protected by Cloudflare CDN",
			Confidence:  0.2,
			Evidence: []models.Evidence{
				{
					Type:      "header",
					Reference: "server: cloudflare",
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	return signals
}

// detectGamblingUXSignals detects gambling-related UX patterns
func detectGamblingUXSignals(body string) []models.Signal {
	signals := []models.Signal{}

	// Define gambling-related keywords
	gamblingKeywords := []string{
		"gacor", "maxwin", "depo", "wd", "deposit", "withdraw", "bonus", 
		"slot", "bet", "win", "prize", "jackpot", "spin", "game",
		"casino", "poker", "roulette", "blackjack", "bingo",
		"togel", "lotto", "betting", "odds", "payout",
		"agen", "bandar", "daftar", "register", "login", "masuk",
		"rupiah", "idr", "rp", "withdrawal", "turnover",
		"raja", "sultan", "king", "vip", "premium", "gold", "silver",
		"tembak", "ikan", "tembak ikan", "fish", "fishing",
		"slot online", "judi online", "main judi",
	}

	// Convert body to lowercase for matching
	lowerBody := strings.ToLower(body)

	for _, keyword := range gamblingKeywords {
		if strings.Contains(lowerBody, strings.ToLower(keyword)) {
			signal := models.Signal{
				SignalID:    fmt.Sprintf("ux_%s", strings.ReplaceAll(keyword, " ", "_")),
				Category:    "UX",
				Description: fmt.Sprintf("Found gambling keyword: %s", keyword),
				Confidence:  0.7,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: fmt.Sprintf("Found keyword '%s' in page content", keyword),
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
			break // Only add one UX signal to avoid spamming
		}
	}

	return signals
}

// detectPaymentSignals detects payment-related signals
func detectPaymentSignals(body string) []models.Signal {
	// Use the payment detector for comprehensive payment method detection
	paymentDetector := detector.NewPaymentDetector()
	signals := paymentDetector.DetectPaymentMethods(body)

	// Add payment funnels detection
	funnelSignals := paymentDetector.DetectPaymentFunnels(body)
	signals = append(signals, funnelSignals...)

	// If no payment signals found, fall back to keyword matching
	if len(signals) == 0 {
		// Define payment-related keywords
		paymentKeywords := []string{
			"qris", "qris2", "qris 2", "gopay", "ovo", "dana", "linkaja",
			"doku", "paypal", "bitcoin", "ethereum", "crypto", "wallet",
			"transfer", "bank", "bca", "bni", "mandiri", "bri", "permata",
			"deposit", "withdraw", "topup", "top up", "isi saldo", "saldo",
			"payment", "pay now", "pay", "pembayaran", "bayar",
			"duit", "uang", "money", "cash", "rupiah", "idr", "rp",
			"trx", "transaction", "transaksi", "kode", "unik", "kode unik",
		}

		// Convert body to lowercase for matching
		lowerBody := strings.ToLower(body)

		for _, keyword := range paymentKeywords {
			if strings.Contains(lowerBody, strings.ToLower(keyword)) {
				signal := models.Signal{
					SignalID:    fmt.Sprintf("payment_%s", strings.ReplaceAll(keyword, " ", "_")),
					Category:    "PAYMENT",
					Description: fmt.Sprintf("Found payment method reference: %s", keyword),
					Confidence:  0.8,
					Evidence: []models.Evidence{
						{
							Type:      "html",
							Reference: fmt.Sprintf("Found payment reference '%s' in page content", keyword),
							Timestamp: time.Now(),
						},
					},
				}
				signals = append(signals, signal)
				break // Only add one payment signal to avoid spamming
			}
		}

		// Check for cryptocurrency addresses
		cryptoPatterns := []string{
			`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`, // Bitcoin
			`0x[a-fA-F0-9]{40}`,               // Ethereum
			`R[a-zA-Z0-9]{25,34}`,             // Ripple
		}

		for _, pattern := range cryptoPatterns {
			re := regexp.MustCompile(pattern)
			matches := re.FindAllString(body, -1)
			if len(matches) > 0 {
				signal := models.Signal{
					SignalID:    "payment_crypto_address",
					Category:    "PAYMENT",
					Description: "Found cryptocurrency address pattern",
					Confidence:  0.9,
					Evidence: []models.Evidence{
						{
							Type:      "html",
							Reference: fmt.Sprintf("Found crypto address: %s", matches[0]),
							Timestamp: time.Now(),
						},
					},
				}
				signals = append(signals, signal)
				break // Only add one crypto signal to avoid spamming
			}
		}
	}

	return signals
}

// detectInfrastructureSignals detects infrastructure-related signals
func detectInfrastructureSignals(headers http.Header) []models.Signal {
	signals := []models.Signal{}

	// Check for specific infrastructure headers
	if headers.Get("x-powered-by") != "" {
		signal := models.Signal{
			SignalID:    "infra_x_powered_by",
			Category:    "INFRA",
			Description: fmt.Sprintf("Found x-powered-by header: %s", headers.Get("x-powered-by")),
			Confidence:  0.3,
			Evidence: []models.Evidence{
				{
					Type:      "header",
					Reference: fmt.Sprintf("x-powered-by: %s", headers.Get("x-powered-by")),
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	if headers.Get("x-generator") != "" {
		signal := models.Signal{
			SignalID:    "infra_x_generator",
			Category:    "INFRA",
			Description: fmt.Sprintf("Found x-generator header: %s", headers.Get("x-generator")),
			Confidence:  0.3,
			Evidence: []models.Evidence{
				{
					Type:      "header",
					Reference: fmt.Sprintf("x-generator: %s", headers.Get("x-generator")),
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	return signals
}

// detectOriginIPs attempts to find origin IPs behind CDN
func detectOriginIPs(domain string) ([]string, []models.Evidence, error) {
	detector := detector.NewOriginIPDetector()
	return detector.DetectOriginIPs(domain)
}

// GetIPFromDomain attempts to get the origin IP of a domain
// This is a simplified approach that won't work for Cloudflare-protected domains
func GetIPFromDomain(domain string) (string, error) {
	ips, err := net.LookupIP(domain)
	if err != nil {
		return "", err
	}

	if len(ips) > 0 {
		// Return the first IP found
		return ips[0].String(), nil
	}

	return "", fmt.Errorf("no IP found for domain %s", domain)
}