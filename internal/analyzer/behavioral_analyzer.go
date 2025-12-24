package analyzer

import (
	"regexp"
	"strings"
	"time"

	"github.com/genesis410/fogger/internal/models"
)

// BehavioralAnalyzer performs behavioral and semantic analysis
type BehavioralAnalyzer struct {
	GamblingKeywords []string
	PaymentKeywords  []string
	RegexPatterns    map[string]*regexp.Regexp
}

// NewBehavioralAnalyzer creates a new instance of BehavioralAnalyzer
func NewBehavioralAnalyzer() *BehavioralAnalyzer {
	analyzer := &BehavioralAnalyzer{
		GamblingKeywords: []string{
			"gacor", "maxwin", "depo", "wd", "deposit", "withdraw", "bonus", 
			"slot", "bet", "win", "prize", "jackpot", "spin", "game",
			"casino", "poker", "roulette", "blackjack", "bingo",
			"togel", "lotto", "betting", "odds", "payout",
			"agen", "bandar", "daftar", "register", "login", "masuk",
			"rupiah", "idr", "rp", "withdrawal", "turnover",
			"raja", "sultan", "king", "vip", "premium", "gold", "silver",
			"tembak", "ikan", "tembak ikan", "fish", "fishing",
			"slot online", "judi online", "main judi", "bet online",
			"main slot", "daftar slot", "situs judi", "situs slot",
			"game slot", "slot gacor", "link alternatif",
			"free spin", "freespin", "promo", "promosi", "bonus",
			"cashback", "rekening", "bank", "transfer", "dana",
			"min deposit", "min depo", "deposit murah", "deposit kecil",
			"24 jam", "24jam", "layanan", "customer service",
			"live chat", "cs online", "operator", "layan",
			"menang besar", "kemenangan", "jackpot besar",
			"tembus", "hoki", "keberuntungan", "fortune",
			"mudah menang", "gampang menang", "gampang jp",
			"raih jp", "jp besar", "jp maxwin", "maxwin",
		},
		PaymentKeywords: []string{
			"qris", "qris2", "qris 2", "gopay", "ovo", "dana", "linkaja",
			"doku", "paypal", "bitcoin", "ethereum", "crypto", "wallet",
			"transfer", "bank", "bca", "bni", "mandiri", "bri", "permata",
			"deposit", "withdraw", "topup", "top up", "isi saldo", "saldo",
			"payment", "pay now", "pay", "pembayaran", "bayar",
			"duit", "uang", "money", "cash", "rupiah", "idr", "rp",
			"trx", "transaction", "transaksi", "kode", "unik", "kode unik",
			"ewallet", "e-wallet", "dompet digital", "dompet elektronik",
			"virtual account", "va", "virtual", "account",
			"pulsa", "pulsa telkomsel", "pulsa xl", "pulsa axis",
			"pulsa tri", "pulsa indosat", "pulsa smartfren",
			"coin", "token", "usdt", "usdc", "tether", "stablecoin",
			"dogecoin", "doge", "litecoin", "ltc", "bitcoin cash",
			"ripple", "xrp", "cardano", "ada", "solana", "sol",
			"monero", "xmr", "zcash", "zec", "dash", "dgb", "digibyte",
		},
		RegexPatterns: make(map[string]*regexp.Regexp),
	}

	// Compile regex patterns
	analyzer.compilePatterns()

	return analyzer
}

// compilePatterns compiles regex patterns for various checks
func (b *BehavioralAnalyzer) compilePatterns() {
	// Crypto address patterns
	b.RegexPatterns["bitcoin"] = regexp.MustCompile(`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`)
	b.RegexPatterns["ethereum"] = regexp.MustCompile(`0x[a-fA-F0-9]{40}`)
	b.RegexPatterns["ripple"] = regexp.MustCompile(`r[0-9a-zA-Z]{24,34}`)
	b.RegexPatterns["monero"] = regexp.MustCompile(`4[0-9AB][1-9A-HJ-NP-Za-km-z]{93}`)
	
	// Phone number patterns (sometimes used for contact in gambling sites)
	b.RegexPatterns["phone"] = regexp.MustCompile(`(\+62|62|0)8[1-9][0-9\s\-\+\(\)]{4,14}`)
	
	// ID patterns (sometimes used in gambling contexts)
	b.RegexPatterns["id_pattern"] = regexp.MustCompile(`[A-Za-z0-9]{8,20}`)
}

// AnalyzeContent performs behavioral and semantic analysis on content
func (b *BehavioralAnalyzer) AnalyzeContent(content string) []models.Signal {
	var signals []models.Signal

	// Convert to lowercase for matching
	lowerContent := strings.ToLower(content)

	// Check for gambling keywords
	gamblingSignals := b.checkGamblingKeywords(lowerContent)
	signals = append(signals, gamblingSignals...)

	// Check for payment keywords
	paymentSignals := b.checkPaymentKeywords(lowerContent)
	signals = append(signals, paymentSignals...)

	// Check for crypto addresses
	cryptoSignals := b.checkCryptoAddresses(content)
	signals = append(signals, cryptoSignals...)

	// Check for phone numbers (potentially customer service)
	phoneSignals := b.checkPhoneNumbers(content)
	signals = append(signals, phoneSignals...)

	// Check for suspicious patterns
	patternSignals := b.checkSuspiciousPatterns(content)
	signals = append(signals, patternSignals...)

	return signals
}

// checkGamblingKeywords checks for gambling-related keywords
func (b *BehavioralAnalyzer) checkGamblingKeywords(content string) []models.Signal {
	var signals []models.Signal

	for _, keyword := range b.GamblingKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			signal := models.Signal{
				SignalID:    "gambling_keyword_" + strings.ReplaceAll(keyword, " ", "_"),
				Category:    "UX",
				Description: "Found gambling keyword: " + keyword,
				Confidence:  0.6, // Adjust based on keyword importance
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found gambling keyword '" + keyword + "' in content",
						Timestamp: time.Now(),
					},
				},
			}
			
			// Increase confidence for more specific gambling terms
			if b.isHighValueGamblingKeyword(keyword) {
				signal.Confidence = 0.8
			}
			
			signals = append(signals, signal)
		}
	}

	return signals
}

// isHighValueGamblingKeyword checks if a keyword is particularly indicative of gambling
func (b *BehavioralAnalyzer) isHighValueGamblingKeyword(keyword string) bool {
	highValueKeywords := []string{
		"togel", "slot", "betting", "casino", "judi", "main judi",
		"slot gacor", "jp", "maxwin", "gacor", "raja", "bandar",
		"situs judi", "agen judi", "daftar judi",
	}
	
	for _, highValue := range highValueKeywords {
		if strings.Contains(strings.ToLower(keyword), strings.ToLower(highValue)) {
			return true
		}
	}
	
	return false
}

// checkPaymentKeywords checks for payment-related keywords
func (b *BehavioralAnalyzer) checkPaymentKeywords(content string) []models.Signal {
	var signals []models.Signal

	for _, keyword := range b.PaymentKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			signal := models.Signal{
				SignalID:    "payment_keyword_" + strings.ReplaceAll(keyword, " ", "_"),
				Category:    "PAYMENT",
				Description: "Found payment method reference: " + keyword,
				Confidence:  0.7, // Adjust based on keyword importance
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found payment method '" + keyword + "' in content",
						Timestamp: time.Now(),
					},
				},
			}
			
			// Increase confidence for specific payment methods
			if b.isHighValuePaymentMethod(keyword) {
				signal.Confidence = 0.9
			}
			
			signals = append(signals, signal)
		}
	}

	return signals
}

// isHighValuePaymentMethod checks if a payment method is particularly indicative of gambling sites
func (b *BehavioralAnalyzer) isHighValuePaymentMethod(keyword string) bool {
	highValuePayments := []string{
		"qris", "qris2", "gopay", "ovo", "dana", "linkaja",
		"doku", "ewallet", "e-wallet", "dompet digital",
		"virtual account", "va", "pulsa",
	}
	
	for _, highValue := range highValuePayments {
		if strings.Contains(strings.ToLower(keyword), strings.ToLower(highValue)) {
			return true
		}
	}
	
	return false
}

// checkCryptoAddresses checks for cryptocurrency addresses
func (b *BehavioralAnalyzer) checkCryptoAddresses(content string) []models.Signal {
	var signals []models.Signal

	for currency, pattern := range b.RegexPatterns {
		if currency == "phone" || currency == "id_pattern" {
			continue // Skip non-crypto patterns
		}
		
		matches := pattern.FindAllString(content, -1)
		for _, match := range matches {
			signal := models.Signal{
				SignalID:    "crypto_address_" + currency,
				Category:    "PAYMENT",
				Description: "Found " + currency + " cryptocurrency address: " + match,
				Confidence:  0.95,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found " + currency + " address: " + match,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// checkPhoneNumbers checks for Indonesian phone numbers
func (b *BehavioralAnalyzer) checkPhoneNumbers(content string) []models.Signal {
	var signals []models.Signal

	pattern := b.RegexPatterns["phone"]
	matches := pattern.FindAllString(content, -1)
	for _, match := range matches {
		signal := models.Signal{
			SignalID:    "indonesian_phone_number",
			Category:    "UX",
			Description: "Found Indonesian phone number pattern: " + match,
			Confidence:  0.4,
			Evidence: []models.Evidence{
				{
					Type:      "html",
					Reference: "Found phone number: " + match,
					Timestamp: time.Now(),
				},
			},
		}
		signals = append(signals, signal)
	}

	return signals
}

// checkSuspiciousPatterns checks for other suspicious patterns
func (b *BehavioralAnalyzer) checkSuspiciousPatterns(content string) []models.Signal {
	var signals []models.Signal

	// Check for ID patterns (potentially user IDs or account numbers)
	pattern := b.RegexPatterns["id_pattern"]
	matches := pattern.FindAllString(content, -1)
	for _, match := range matches {
		// Filter out likely false positives
		if len(match) > 15 { // Likely to be a real ID pattern
			signal := models.Signal{
				SignalID:    "suspicious_id_pattern",
				Category:    "INFRA",
				Description: "Found suspicious ID pattern: " + match,
				Confidence:  0.3,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found ID pattern: " + match,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// AnalyzeDOMStructure analyzes the structure of the DOM for gambling patterns
func (b *BehavioralAnalyzer) AnalyzeDOMStructure(html string) []models.Signal {
	var signals []models.Signal

	// Look for gambling-specific elements
	elements := []struct {
		pattern   string
		category  string
		confidence float64
		desc      string
	}{
		{`<input[^>]*type=["']password["'][^>]*name=["']*(pin|sandi|password|pass)["']*`, "UX", 0.5, "Password input field"},
		{`<input[^>]*type=["']text["'][^>]*name=["']*(username|user|id|uid)["']*`, "UX", 0.4, "Username input field"},
		{`<input[^>]*type=["']number["'][^>]*name=["']*(amount|jumlah|nominal|uang)["']*`, "PAYMENT", 0.6, "Amount input field"},
		{`<button[^>]*[Dd][Ee][Pp][Oo]|[Ww][Dd]|[Ww][Ii][Tt][Hh][Dd][Rr][Aa][Ww]|[Tt][Rr][Aa][Nn][Ss][Ff][Ee][Rr]`, "PAYMENT", 0.7, "Deposit/Withdraw button"},
		{`<img[^>]*[Ss][Ll][Oo][Tt]|[Cc][Aa][Ss][Ii][Nn][Oo]|[Gg][Aa][Cc][Oo][Rr]|[Mm][Aa][Xx][Ww][Ii][Nn]`, "UX", 0.6, "Gambling image"},
		{`<div[^>]*class=["']*[Ss][Ll][Oo][Tt]|[Gg][Aa][Mm][Ee]|[Cc][Aa][Ss][Ii][Nn][Oo]`, "UX", 0.5, "Gambling section"},
		{`<iframe[^>]*[Yy][Oo][Uu][Tt][Uu][Bb][Ee]|[Ff][Aa][Cc][Ee][Bb][Oo][Oo][Kk]|[Tt][Ww][Ii][Tt][Tt][Ee][Rr]`, "UX", 0.3, "Social media embed"},
	}

	for _, element := range elements {
		re := regexp.MustCompile(element.pattern)
		matches := re.FindAllString(html, -1)
		if len(matches) > 0 {
			signal := models.Signal{
				SignalID:    "dom_pattern_" + strings.ReplaceAll(element.desc, " ", "_"),
				Category:    element.category,
				Description: element.desc + " found in DOM",
				Confidence:  element.confidence,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found pattern: " + element.pattern,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// AnalyzePageSemantics analyzes the semantic meaning of page content
func (b *BehavioralAnalyzer) AnalyzePageSemantics(title, description, content string) []models.Signal {
	var signals []models.Signal

	// Analyze title
	titleSignals := b.analyzeTitle(title)
	signals = append(signals, titleSignals...)

	// Analyze meta description
	descSignals := b.analyzeDescription(description)
	signals = append(signals, descSignals...)

	// Analyze content
	contentSignals := b.analyzeContentSemantics(content)
	signals = append(signals, contentSignals...)

	return signals
}

// analyzeTitle analyzes the page title for gambling indicators
func (b *BehavioralAnalyzer) analyzeTitle(title string) []models.Signal {
	var signals []models.Signal

	if title == "" {
		return signals
	}

	lowerTitle := strings.ToLower(title)

	for _, keyword := range b.GamblingKeywords {
		if strings.Contains(lowerTitle, strings.ToLower(keyword)) {
			signal := models.Signal{
				SignalID:    "title_gambling_keyword_" + strings.ReplaceAll(keyword, " ", "_"),
				Category:    "UX",
				Description: "Gambling keyword found in title: " + keyword,
				Confidence:  0.8,
				Evidence: []models.Evidence{
					{
						Type:      "meta",
						Reference: "Title: " + title,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	// Check for common gambling title patterns
	gamblingTitlePatterns := []string{
		"slot", "casino", "togel", "bet", "judi", "game", 
		"online", "terpercaya", "terbaik", "gacor", "maxwin",
		"situs", "agen", "bandar", "raja", "sultan",
	}

	titleWords := strings.Fields(lowerTitle)
	for _, word := range titleWords {
		for _, pattern := range gamblingTitlePatterns {
			if strings.Contains(word, pattern) {
				signal := models.Signal{
					SignalID:    "title_pattern_" + pattern,
					Category:    "UX",
					Description: "Gambling-related pattern in title: " + pattern,
					Confidence:  0.7,
					Evidence: []models.Evidence{
						{
							Type:      "meta",
							Reference: "Title: " + title,
							Timestamp: time.Now(),
						},
					},
				}
				signals = append(signals, signal)
				break
			}
		}
	}

	return signals
}

// analyzeDescription analyzes the meta description for gambling indicators
func (b *BehavioralAnalyzer) analyzeDescription(description string) []models.Signal {
	var signals []models.Signal

	if description == "" {
		return signals
	}

	lowerDesc := strings.ToLower(description)

	for _, keyword := range b.GamblingKeywords {
		if strings.Contains(lowerDesc, strings.ToLower(keyword)) {
			signal := models.Signal{
				SignalID:    "desc_gambling_keyword_" + strings.ReplaceAll(keyword, " ", "_"),
				Category:    "UX",
				Description: "Gambling keyword found in description: " + keyword,
				Confidence:  0.7,
				Evidence: []models.Evidence{
					{
						Type:      "meta",
						Reference: "Description: " + description,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	return signals
}

// analyzeContentSemantics analyzes the semantic meaning of content
func (b *BehavioralAnalyzer) analyzeContentSemantics(content string) []models.Signal {
	var signals []models.Signal

	// Check for gambling-specific sentence patterns
	gamblingPatterns := []string{
		"daftar sekarang", "main sekarang", "daftar dan main", 
		"menang besar", "jackpot terbesar", "gampang menang",
		"bisa withdraw", "proses cepat", "customer service",
		"layanan 24 jam", "deposit murah", "bonus besar",
		"main slot", "main judi", "situs terpercaya",
		"terbukti bayar", "langsung main", "langsung dapat",
		"langsung dapat jp", "langsung jp", "langsung maxwin",
	}

	lowerContent := strings.ToLower(content)
	for _, pattern := range gamblingPatterns {
		if strings.Contains(lowerContent, pattern) {
			signal := models.Signal{
				SignalID:    "content_pattern_" + strings.ReplaceAll(pattern, " ", "_"),
				Category:    "UX",
				Description: "Gambling sentence pattern found: " + pattern,
				Confidence:  0.7,
				Evidence: []models.Evidence{
					{
						Type:      "html",
						Reference: "Found pattern: " + pattern,
						Timestamp: time.Now(),
					},
				},
			}
			signals = append(signals, signal)
		}
	}

	return signals
}