package detector

import (
	"crypto/md5"
	"encoding/hex"
	"regexp"
	"strings"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// PaymentDetector detects payment methods and tracks affiliate relationships
type PaymentDetector struct {
	PaymentPatterns map[string]*regexp.Regexp
	AffiliateRegex  *regexp.Regexp
}

// NewPaymentDetector creates a new payment detector
func NewPaymentDetector() *PaymentDetector {
	pd := &PaymentDetector{
		PaymentPatterns: make(map[string]*regexp.Regexp),
	}
	
	// Compile payment method patterns
	pd.compilePaymentPatterns()
	
	// Compile affiliate tracking patterns
	pd.AffiliateRegex = regexp.MustCompile(`(ref|refer|affiliate|af|pid|aid|subid|campaign|source|medium|term|content)=[a-zA-Z0-9_-]+`)
	
	return pd
}

// compilePaymentPatterns compiles regex patterns for payment methods
func (pd *PaymentDetector) compilePaymentPatterns() {
	// Indonesian payment methods
	pd.PaymentPatterns["qris"] = regexp.MustCompile(`(?i)(qris|qris2)`)
	pd.PaymentPatterns["gopay"] = regexp.MustCompile(`(?i)(gopay|go-pay)`)
	pd.PaymentPatterns["ovo"] = regexp.MustCompile(`(?i)(ovo)`)
	pd.PaymentPatterns["dana"] = regexp.MustCompile(`(?i)(dana)`)
	pd.PaymentPatterns["linkaja"] = regexp.MustCompile(`(?i)(linkaja|link-aja)`)
	pd.PaymentPatterns["doku"] = regexp.MustCompile(`(?i)(doku)`)
	
	// Banks
	pd.PaymentPatterns["bca"] = regexp.MustCompile(`(?i)(bca|bank central asia)`)
	pd.PaymentPatterns["bni"] = regexp.MustCompile(`(?i)(bni|bank negara indonesia)`)
	pd.PaymentPatterns["mandiri"] = regexp.MustCompile(`(?i)(mandiri|bank mandiri)`)
	pd.PaymentPatterns["bri"] = regexp.MustCompile(`(?i)(bri|bank rakyat indonesia)`)
	pd.PaymentPatterns["permata"] = regexp.MustCompile(`(?i)(permata|bank permata)`)
	
	// Cryptocurrency patterns
	pd.PaymentPatterns["bitcoin"] = regexp.MustCompile(`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`)
	pd.PaymentPatterns["ethereum"] = regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	pd.PaymentPatterns["ripple"] = regexp.MustCompile(`r[0-9a-zA-Z]{24,34}`)
	
	// E-wallets
	pd.PaymentPatterns["paypal"] = regexp.MustCompile(`(?i)(paypal)`)
	pd.PaymentPatterns["payoneer"] = regexp.MustCompile(`(?i)(payoneer)`)
	
	// Payment-related keywords
	pd.PaymentPatterns["deposit"] = regexp.MustCompile(`(?i)(deposit|depo|isi saldo|top up|topup)`)
	pd.PaymentPatterns["withdraw"] = regexp.MustCompile(`(?i)(withdraw|wd|tarik dana|ambil dana)`)
	pd.PaymentPatterns["transfer"] = regexp.MustCompile(`(?i)(transfer|tf|kirim)`)
}

// DetectPaymentMethods detects payment methods in content
func (pd *PaymentDetector) DetectPaymentMethods(content string) []models.Signal {
	var signals []models.Signal

	for method, pattern := range pd.PaymentPatterns {
		matches := pattern.FindAllString(content, -1)
		for _, match := range matches {
			signal := models.Signal{
				SignalID:    "payment_method_" + method,
				Category:    "PAYMENT",
				Description: "Detected payment method: " + method + " (" + match + ")",
				Confidence:  pd.getPaymentConfidence(method),
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found payment method '" + method + "' in content: " + match,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}
	
	// Look for Indonesian-specific payment patterns
	idPaymentSignals := pd.detectIndonesianPaymentPatterns(content)
	signals = append(signals, idPaymentSignals...)

	// Look for crypto wallet addresses
	cryptoSignals := pd.detectCryptoWallets(content)
	signals = append(signals, cryptoSignals...)

	return signals
}

// detectIndonesianPaymentPatterns detects Indonesian-specific payment patterns
func (pd *PaymentDetector) detectIndonesianPaymentPatterns(content string) []models.Signal {
	var signals []models.Signal

	// QRIS patterns
	qrisRegex := regexp.MustCompile(`(?i)(qris.*2|qris2|qr.*2)`)
	qrisMatches := qrisRegex.FindAllString(content, -1)
	for _, match := range qrisMatches {
		signal := models.Signal{
			SignalID:    "payment_qris2",
			Category:    "PAYMENT",
			Description: "Detected QRIS 2.0 payment method: " + match,
			Confidence:  0.9,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found QRIS 2.0 pattern: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	// Pulsa (mobile credit) patterns
	pulsaRegex := regexp.MustCompile(`(?i)(pulsa|pulsa.*telkomsel|pulsa.*xl|pulsa.*axis|pulsa.*tri|pulsa.*indosat|pulsa.*smartfren)`)
	pulsaMatches := pulsaRegex.FindAllString(content, -1)
	for _, match := range pulsaMatches {
		signal := models.Signal{
			SignalID:    "payment_pulsa",
			Category:    "PAYMENT",
			Description: "Detected pulsa (mobile credit) payment method: " + match,
			Confidence:  0.8,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found pulsa pattern: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	return signals
}

// detectCryptoWallets detects cryptocurrency wallet addresses
func (pd *PaymentDetector) detectCryptoWallets(content string) []models.Signal {
	var signals []models.Signal
	
	// Bitcoin
	btcRegex := regexp.MustCompile(`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`)
	btcMatches := btcRegex.FindAllString(content, -1)
	for _, match := range btcMatches {
		signal := models.Signal{
			SignalID:    "crypto_bitcoin",
			Category:    "PAYMENT",
			Description: "Detected Bitcoin address: " + match,
			Confidence:  0.95,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found Bitcoin address: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}
	
	// Ethereum
	ethRegex := regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	ethMatches := ethRegex.FindAllString(content, -1)
	for _, match := range ethMatches {
		signal := models.Signal{
			SignalID:    "crypto_ethereum",
			Category:    "PAYMENT",
			Description: "Detected Ethereum address: " + match,
			Confidence:  0.95,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found Ethereum address: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}
	
	// USDT (Tether)
	usdtRegex := regexp.MustCompile(`[13][a-km-zA-HJ-NP-Z1-9]{33}|0x[a-fA-F0-9]{40}|T[A-Za-z1-9]{33}`)
	usdtMatches := usdtRegex.FindAllString(content, -1)
	for _, match := range usdtMatches {
		signal := models.Signal{
			SignalID:    "crypto_usdt",
			Category:    "PAYMENT",
			Description: "Detected USDT (Tether) address: " + match,
			Confidence:  0.95,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found USDT address: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}
	
	return signals
}

// getPaymentConfidence returns confidence level for different payment methods
func (pd *PaymentDetector) getPaymentConfidence(method string) float64 {
	switch method {
	case "bitcoin", "ethereum", "ripple", "usdt", "crypto_bitcoin", "crypto_ethereum", "crypto_usdt":
		return 0.95
	case "qris", "qris2", "gopay", "ovo", "dana", "linkaja", "doku", "bca", "bni", "mandiri", "bri", "permata":
		return 0.9
	case "paypal", "payoneer":
		return 0.8
	case "deposit", "withdraw", "transfer":
		return 0.7
	default:
		return 0.6
	}
}

// DetectAffiliateRelationships detects potential affiliate relationships
func (pd *PaymentDetector) DetectAffiliateRelationships(content string, url string) []models.Signal {
	var signals []models.Signal
	
	// Check URL for affiliate parameters
	affiliateMatches := pd.AffiliateRegex.FindAllString(url, -1)
	for _, match := range affiliateMatches {
		signal := models.Signal{
			SignalID:    "affiliate_parameter",
			Category:    "INFRA",
			Description: "Detected affiliate tracking parameter: " + match,
			Confidence:  0.6,
			Evidence: []models.Evidence{
				{
					Type:      "url",
					Reference: "Found affiliate parameter in URL: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}
	
	// Look for common affiliate tracking patterns in content
	affiliatePatterns := []string{
		"aff", "affiliate", "ref", "refer", "pid", "aid", "subid",
		"campaign", "source", "medium", "term", "content",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range affiliatePatterns {
		if strings.Contains(lowerContent, pattern) {
			signal := models.Signal{
				SignalID:    "affiliate_pattern_" + pattern,
				Category:    "INFRA",
				Description: "Detected potential affiliate pattern: " + pattern,
				Confidence:  0.5,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found affiliate pattern '" + pattern + "' in content",
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}
	
	// Look for referral links
	referralRegex := regexp.MustCompile(`https?://[^\s"']*ref=[a-zA-Z0-9_-]+`)
	referralMatches := referralRegex.FindAllString(content, -1)
	for _, match := range referralMatches {
		signal := models.Signal{
			SignalID:    "referral_link",
			Category:    "INFRA",
			Description: "Detected referral link: " + match,
			Confidence:  0.7,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found referral link: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}
	
	return signals
}

// DetectPaymentAPIs detects payment API integrations
func (pd *PaymentDetector) DetectPaymentAPIs(content string) []models.Signal {
	var signals []models.Signal
	
	// Look for common payment API patterns
	paymentAPIs := map[string]string{
		"midtrans":     `midtrans`,
		"stripe":       `stripe`,
		"paypal":       `paypal`,
		"razorpay":     `razorpay`,
		"payu":         `payu`,
		"doku":         `doku`,
		"xendit":       `xendit`,
		"iak":          `iak`, // Indonesian payment gateway
		"tripay":       `tripay`,
		"paymentku":    `paymentku`,
	}
	
	lowerContent := strings.ToLower(content)
	
	for apiName, pattern := range paymentAPIs {
		if strings.Contains(lowerContent, pattern) {
			signal := models.Signal{
				SignalID:    "payment_api_" + apiName,
				Category:    "PAYMENT",
				Description: "Detected payment API integration: " + apiName,
				Confidence:  pd.getPaymentAPIConfidence(apiName),
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found " + apiName + " API reference in content",
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}
	
	return signals
}

// getPaymentAPIConfidence returns confidence for different payment APIs
func (pd *PaymentDetector) getPaymentAPIConfidence(apiName string) float64 {
	switch apiName {
	case "doku", "xendit", "iak", "tripay", "paymentku":
		// Indonesian payment gateways - high confidence for local gambling sites
		return 0.85
	case "midtrans":
		// Popular in Indonesia
		return 0.8
	case "paypal", "stripe":
		// Common globally, less specific to Indonesian gambling
		return 0.6
	default:
		return 0.5
	}
}

// DetectPaymentFunnels detects payment flow patterns
func (pd *PaymentDetector) DetectPaymentFunnels(content string) []models.Signal {
	var signals []models.Signal
	
	// Look for payment flow patterns
	paymentFlowPatterns := []struct {
		pattern string
		desc    string
		conf    float64
	}{
		{`(?i)(deposit.*form|isi.*saldo.*form|form.*deposit|form.*isi.*saldo)`, "Deposit form detected", 0.8},
		{`(?i)(withdraw.*form|tarik.*dana.*form|form.*withdraw|form.*tarik.*dana)`, "Withdrawal form detected", 0.8},
		{`(?i)(payment.*method|pilih.*pembayaran|metode.*pembayaran)`, "Payment method selection", 0.7},
		{`(?i)(konfirmasi.*deposit|konfirmasi.*pembayaran|deposit.*confirm)`, "Deposit confirmation", 0.8},
		{`(?i)(minimal.*deposit|min.*dep|deposit.*min)`, "Minimum deposit requirement", 0.7},
		{`(?i)(promo.*deposit|bonus.*deposit|hadiah.*deposit)`, "Deposit bonus/promotion", 0.8},
		{`(?i)(customer.*service|cs.*24.*jam|layanan.*24.*jam)`, "Customer service for payments", 0.6},
	}
	
	for _, flow := range paymentFlowPatterns {
		regex := regexp.MustCompile(flow.pattern)
		matches := regex.FindAllString(content, -1)
		for _, match := range matches {
			signal := models.Signal{
				SignalID:    "payment_flow_" + strings.ReplaceAll(flow.desc, " ", "_"),
				Category:    "PAYMENT",
				Description: flow.desc + ": " + match,
				Confidence:  flow.conf,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found payment flow pattern: " + match,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}
	
	return signals
}

// TrackAffiliateNetwork identifies potential affiliate networks
func (pd *PaymentDetector) TrackAffiliateNetwork(baseDomain string, content string) string {
	// Create a hash of the domain and content to identify the network
	data := baseDomain + content
	
	hasher := md5.New()
	hasher.Write([]byte(data))
	
	// Return first 16 characters as a simple identifier
	return hex.EncodeToString(hasher.Sum(nil))[:16]
}

// GetPaymentFingerprint returns a comprehensive payment fingerprint
func (pd *PaymentDetector) GetPaymentFingerprint(content string, url string) map[string]interface{} {
	fingerprint := make(map[string]interface{})
	
	// Detect payment methods
	paymentSignals := pd.DetectPaymentMethods(content)
	paymentMethods := make([]string, 0)
	for _, signal := range paymentSignals {
		paymentMethods = append(paymentMethods, signal.Description)
	}
	fingerprint["payment_methods"] = paymentMethods
	
	// Detect affiliate relationships
	affiliateSignals := pd.DetectAffiliateRelationships(content, url)
	affiliateParams := make([]string, 0)
	for _, signal := range affiliateSignals {
		affiliateParams = append(affiliateParams, signal.Description)
	}
	fingerprint["affiliate_params"] = affiliateParams
	
	// Detect payment APIs
	apiSignals := pd.DetectPaymentAPIs(content)
	paymentAPIs := make([]string, 0)
	for _, signal := range apiSignals {
		paymentAPIs = append(paymentAPIs, signal.Description)
	}
	fingerprint["payment_apis"] = paymentAPIs
	
	// Detect payment funnels
	funnelSignals := pd.DetectPaymentFunnels(content)
	paymentFunnels := make([]string, 0)
	for _, signal := range funnelSignals {
		paymentFunnels = append(paymentFunnels, signal.Description)
	}
	fingerprint["payment_funnels"] = paymentFunnels
	
	// Calculate overall payment confidence
	totalConfidence := 0.0
	signalCount := 0
	
	for _, signal := range paymentSignals {
		totalConfidence += signal.Confidence
		signalCount++
	}
	for _, signal := range apiSignals {
		totalConfidence += signal.Confidence
		signalCount++
	}
	for _, signal := range funnelSignals {
		totalConfidence += signal.Confidence
		signalCount++
	}
	
	if signalCount > 0 {
		fingerprint["overall_payment_confidence"] = totalConfidence / float64(signalCount)
	} else {
		fingerprint["overall_payment_confidence"] = 0.0
	}
	
	return fingerprint
}