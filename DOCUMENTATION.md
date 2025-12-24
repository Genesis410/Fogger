# fogger - Full Implementation Guide

## Table of Contents
1. [Overview](#overview)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Command Reference](#command-reference)
5. [Configuration](#configuration)
6. [Advanced Usage](#advanced-usage)
7. [Examples](#examples)
8. [Troubleshooting](#troubleshooting)

## Overview

fogger is a cybersecurity tool designed to identify and analyze illicit online gambling ("judol") operations that hide behind CDNs like Cloudflare. It provides intelligence on gambling sites without attempting to bypass CDN protections, focusing on ecosystem-level abuse patterns.

### Key Features
- **CDN-Aware Detection**: Identifies sites protected by Cloudflare and other CDNs
- **Multi-Vector Analysis**: Behavioral, semantic, and infrastructure correlation
- **Judol Likelihood Index (JLI)**: Composite risk scoring system
- **Payment Method Detection**: Identifies local payment methods (Qris, OVO, DANA, etc.)
- **Clustering Engine**: Groups related domains into operator clusters
- **Export Functionality**: JSON/CSV export for integration systems

## Installation

### Prerequisites
- Go 1.21 or higher
- Git

### Method 1: Using Go Install (Recommended)
```bash
go install github.com/genesis410/fogger@latest
```

### Method 2: From Source
```bash
# Clone the repository
git clone https://github.com/genesis410/fogger.git
cd fogger

# Build the binary
go build -o fogger main.go

# Install dependencies
go mod tidy

# Move to PATH (optional)
sudo mv fogger /usr/local/bin/
```

### Method 3: Pre-built Binaries (Future)
When available, pre-built binaries can be downloaded from the releases page.

## Quick Start

### Basic Domain Scan
```bash
fogger scan example.com
```

### Scan with JSON Output
```bash
fogger scan example.com --json
```

### View Available Commands
```bash
fogger --help
```

## Command Reference

### `fogger scan <domain>`

Analyzes a domain and produces a Judol Likelihood Index (JLI) with evidence.

**Flags:**
- `--json`: Output JSON only
- `--csv`: Output CSV
- `--no-color`: Disable ANSI coloring
- `--timeout <sec>`: Network timeout (default: 10)
- `--profile <name>`: Scoring profile (default: standard)
- `--save`: Persist result to local DB

**Example:**
```bash
fogger scan suspicious-site.com --profile intensive --timeout 30
```

### `fogger cluster <cluster-id>`

View all domains and evidence connected to an operator/campaign.

**Flags:**
- `--graph`: ASCII graph visualization
- `--json`: Output JSON
- `--since <days>`: Time filter

**Example:**
```bash
fogger cluster abc123def456 --graph
```

### `fogger lookup <domain>`

Quick confidence check (cached-first, no deep analysis).

**Example:**
```bash
fogger lookup quick-check.com
```

### `fogger monitor <domain>`

Continuously monitor a domain for changes.

**Flags:**
- `--interval <duration>`: Monitoring interval (default: 5m)
- `--duration <duration>`: Total monitoring time (default: 1h)

**Example:**
```bash
fogger monitor watch-this.com --interval 2m --duration 4h
```

### `fogger export`

Export data for integration with other systems.

**Flags:**
- `--format <json|csv>`: Export format (default: json)
- `--since <period>`: Time period (default: 30d)
- `--domain <domain>`: Specific domain to export
- `--cluster <cluster-id>`: Specific cluster to export
- `--output <file>`: Output file path

**Example:**
```bash
fogger export --format csv --since 7d --output results.csv
```

### `fogger config`

Manage configuration settings.

**Subcommands:**
- `show`: Show current configuration
- `validate`: Validate current configuration

**Example:**
```bash
fogger config show
fogger config validate
```

## Configuration

fogger uses a YAML configuration file located at `~/.fogger.yaml` or `./.fogger.yaml`.

### Default Configuration
```yaml
scoring:
  gambling_ui: 0.30
  payment_signal: 0.25
  infra_correlation: 0.20
  domain_churn: 0.15
  cdn_pattern: 0.10

thresholds:
  high: 0.75
  medium: 0.50
```

### Configuration Parameters

#### Scoring Weights
- `gambling_ui`: Weight for gambling UX and semantic patterns
- `payment_signal`: Weight for payment and monetization indicators
- `infra_correlation`: Weight for infrastructure reuse and churn
- `domain_churn`: Weight for DNS patterns and domain behavior
- `cdn_pattern`: Weight for CDN usage patterns

**Note:** All weights must sum to 1.0

#### Thresholds
- `high`: Threshold for HIGH risk classification
- `medium`: Threshold for MEDIUM risk classification

### Available Scoring Profiles

1. **standard** (default): Balanced weights for general use
2. **intensive**: Higher weight on gambling and payment signals
3. **conservative**: Higher thresholds, fewer false positives
4. **aggressive**: Lower thresholds, more sensitive detection

## Advanced Usage

### Custom Scoring Profiles

Create a custom configuration file:

```yaml
# custom-profile.yaml
scoring:
  gambling_ui: 0.40
  payment_signal: 0.30
  infra_correlation: 0.15
  domain_churn: 0.10
  cdn_pattern: 0.05

thresholds:
  high: 0.60
  medium: 0.30
```

Use with scan command:
```bash
fogger scan target.com --config custom-profile.yaml
```

### Batch Domain Analysis

For analyzing multiple domains, create a text file with one domain per line:

```bash
# domains.txt
site1.com
site2.com
site3.com
```

Then process with a script:
```bash
while read domain; do
  fogger scan "$domain" --json >> results.json
done < domains.txt
```

### Integration with Other Tools

#### Output to JSON for Processing
```bash
fogger scan target.com --json | jq '.jli_score'
```

#### Export for CSV Analysis
```bash
fogger export --format csv --since 7d --output weekly_report.csv
```

## Examples

### Example 1: Basic Analysis
```bash
$ fogger scan gambling-site.com
┌─────────────────┬──────────┬──────────┬──────────────┐
│ Domain          │ JLI Score│ JLI Level│ CDN Provider │
├─────────────────┼──────────┼──────────┼──────────────┤
│ gambling-site.com│   0.842  │   HIGH   │ cloudflare   │
└─────────────────┴──────────┴──────────┴──────────────┘

┌──────────┬───────┬────────┬─────────────┐
│ Category │ Score │ Weight │ Contribution│
├──────────┼───────┼────────┼─────────────┤
│ UX       │ 0.800 │ 0.300  │   0.240     │
│ PAYMENT  │ 0.900 │ 0.250  │   0.225     │
│ INFRA    │ 0.700 │ 0.200  │   0.140     │
│ DNS      │ 0.500 │ 0.150  │   0.075     │
│ CDN      │ 0.600 │ 0.100  │   0.060     │
├──────────┼───────┼────────┼─────────────┤
│ TOTAL    │       │        │   0.740     │
└──────────┴───────┴────────┴─────────────┘
Judol Likelihood Level: HIGH
```

### Example 2: Cluster Analysis
```bash
$ fogger cluster abc123def456
Cluster ID: abc123def456
Domains: 45
Confidence: 0.89
First Seen: 2023-01-15
Last Seen: 2023-12-20
Shared Resources:
  - IP: 1.2.3.4
  - Wallet: 1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa
  - Payment API: doku
```

### Example 3: Export Data
```bash
$ fogger export --format csv --since 30d --output monthly_results.csv
Exported 1247 domains to monthly_results.csv
```

## Troubleshooting

### Common Issues

#### 1. "Connection timeout" errors
- **Cause**: Target domain is not responding or blocking requests
- **Solution**: Increase timeout value: `fogger scan domain.com --timeout 30`

#### 2. "No signals found" for known gambling sites
- **Cause**: Scoring profile may be too conservative
- **Solution**: Use intensive profile: `fogger scan domain.com --profile intensive`

#### 3. Slow performance
- **Cause**: Large number of requests or slow network
- **Solution**: Reduce parallelism or increase timeout values

#### 4. Permission errors during installation
- **Cause**: Insufficient permissions to install to system directories
- **Solution**: Use local installation or `go install` to user directory

### Debugging Tips

1. **Enable verbose output** (if available):
   ```bash
   fogger scan domain.com -v
   ```

2. **Check configuration validity**:
   ```bash
   fogger config validate
   ```

3. **Test with known good domains** first:
   ```bash
   fogger scan google.com  # Should return LOW risk
   ```

### Performance Considerations

- **Rate Limiting**: The tool implements built-in rate limiting to avoid overwhelming targets
- **Timeout Handling**: Default 10-second timeout per request, adjustable via `--timeout`
- **Memory Usage**: Results are processed in memory; large batch operations may require more RAM

## Legal and Ethical Usage

### Permitted Use Cases
- Government cybercrime and digital enforcement units
- ISP and DNS operator abuse departments
- Payment processor compliance teams
- Academic research (with appropriate ethics approval)
- Bug bounty and security research (within scope)

### Prohibited Use Cases
- Circumvention of access controls
- Traffic interception or manipulation
- Active exploitation or scanning
- Any activity violating local laws

### Data Handling
- All data collection is passive and OSINT-only
- No private or non-consensual data collection
- Data minimization principles applied
- Audit logging for analyst actions

## Support and Community

### Getting Help
- Check the documentation first
- Search existing issues
- Open a new issue for bugs or feature requests

### Contributing
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

### Reporting Issues
When reporting issues, please include:
- Version information: `fogger version` (when available)
- Operating system and Go version
- Steps to reproduce
- Expected vs actual behavior
- Any relevant error messages

## Roadmap

### Phase 1 (Current) - MVP
- ✅ Core detection and clustering
- ✅ Analyst dashboard MVP
- ✅ CLI-first interface

### Phase 2 - Analyst Productivity
- 🔲 API and automation
- 🔲 Multi-CDN expansion
- 🔲 Continuous monitoring

### Phase 3 - Ecosystem Intelligence
- 🔲 Predictive campaign modeling
- 🔲 Cross-border intelligence sharing
- 🔲 Machine learning enhancement

---

## Appendix A: Judol Likelihood Index (JLI) Scoring

The JLI is calculated as follows:

```
JLI_raw = Σ (Category_Weight × Category_Score)
JLI = JLI_raw × Confidence_Factor
```

Where:
- Category_Score = max(signal_scores in category) (to prevent spamming)
- Confidence_Factor = function of independent signal categories

### Default Weights:
- Gambling UX & semantic patterns: 30%
- Payment and monetization indicators: 25%
- Infrastructure reuse and churn: 20%
- DNS patterns: 15%
- CDN usage patterns: 10%

### Classification:
- HIGH: JLI ≥ 0.75
- MEDIUM: JLI ≥ 0.50
- LOW: JLI < 0.50

## Appendix B: Signal Categories

### UX (User Experience)
- Gambling-specific keywords ("gacor", "maxwin", "slot", etc.)
- UI patterns and design elements
- Navigation and interaction patterns

### PAYMENT
- Local payment methods (Qris, OVO, DANA, etc.)
- Cryptocurrency addresses
- Payment flow patterns
- Transaction forms

### INFRA
- Shared infrastructure indicators
- Certificate reuse
- Server configurations
- Network patterns

### DNS
- Domain registration patterns
- DNS record similarities
- Subdomain structures
- Registrar information

### CDN
- CDN provider detection
- CDN usage patterns
- Bypass attempts
- Security configurations

---

*This documentation is for fogger version 1.0.0. For the latest documentation, visit the project repository.*