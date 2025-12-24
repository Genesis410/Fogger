# fogger

fogger is a cybersecurity tool designed to identify and analyze illicit online gambling ("judol") operations that hide behind CDNs like Cloudflare. It provides intelligence on gambling sites without attempting to bypass CDN protections, focusing on ecosystem-level abuse patterns rather than infrastructure-level suppression.

## 🚀 Features

- **CDN-Aware Detection**: Identifies sites protected by Cloudflare and other CDNs
- **Multi-Vector Analysis**: Behavioral, semantic, and infrastructure correlation
- **Judol Likelihood Index (JLI)**: Composite risk scoring system with explainable factors
- **Payment Method Detection**: Identifies local payment methods (Qris, OVO, DANA, Gopay, etc.)
- **Clustering Engine**: Groups related domains into operator clusters
- **Export Functionality**: JSON/CSV export for integration systems
- **CLI-First Design**: Optimized for automation and scripting
- **Courtroom-Safe**: Explainable scoring with evidence breakdown

## 📋 Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Commands](#commands)
- [Configuration](#configuration)
- [Judol Likelihood Index](#judol-likelihood-index)
- [Examples](#examples)
- [Legal & Ethical Usage](#legal--ethical-usage)
- [Contributing](#contributing)
- [License](#license)

## Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Quick Install

```bash
go install github.com/genesis410/fogger@latest
```

### From Source

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

## Usage

### Basic Scanning

```bash
fogger scan example.com
```

### Advanced Scanning

```bash
fogger scan example.com --profile intensive --timeout 30 --json
```

### Other Commands

```bash
# View cluster information
fogger cluster <cluster-id>

# Quick domain lookup
fogger lookup example.com

# Monitor domain continuously
fogger monitor example.com --interval 5m --duration 2h

# Export data
fogger export --format json --since 30d --output results.json

# View configuration
fogger config show
fogger config validate
```

## Commands

### `scan` - Domain Analysis

Analyzes a domain and produces a Judol Likelihood Index (JLI) with evidence.

```bash
fogger scan <domain> [flags]
```

**Flags:**
- `--json`: Output JSON only
- `--csv`: Output CSV
- `--no-color`: Disable ANSI coloring
- `--timeout <sec>`: Network timeout (default: 10)
- `--profile <name>`: Scoring profile (default: standard)
- `--save`: Persist result to local DB

### `cluster` - Campaign Analysis

View all domains and evidence connected to an operator/campaign.

```bash
fogger cluster <cluster-id> [flags]
```

**Flags:**
- `--graph`: ASCII graph visualization
- `--json`: Output JSON
- `--since <days>`: Time filter

### `lookup` - Quick Check

Quick confidence check (cached-first, no deep analysis).

```bash
fogger lookup <domain>
```

### `monitor` - Continuous Monitoring

Continuously monitor a domain for changes.

```bash
fogger monitor <domain> [flags]
```

**Flags:**
- `--interval <duration>`: Monitoring interval (default: 5m)
- `--duration <duration>`: Total monitoring time (default: 1h)

### `export` - Data Export

Export data for integration with other systems.

```bash
fogger export [flags]
```

**Flags:**
- `--format <json|csv>`: Export format (default: json)
- `--since <period>`: Time period (default: 30d)
- `--domain <domain>`: Specific domain to export
- `--cluster <cluster-id>`: Specific cluster to export
- `--output <file>`: Output file path

### `config` - Configuration Management

Manage configuration settings.

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

- `scoring`: Weight distribution for different signal categories (must sum to 1.0)
- `thresholds`: Classification thresholds for risk levels

### Available Profiles

- `standard`: Balanced weights for general use
- `intensive`: Higher weight on gambling and payment signals
- `conservative`: Higher thresholds, fewer false positives
- `aggressive`: Lower thresholds, more sensitive detection

## Judol Likelihood Index

The JLI is a composite confidence score derived from weighted signals:

- **Gambling UX & semantic patterns**: 30%
- **Payment and monetization indicators**: 25%
- **Infrastructure reuse and churn**: 20%
- **DNS patterns**: 15%
- **CDN usage patterns**: 10%

Scores are classified as:
- **HIGH**: ≥ 0.75
- **MEDIUM**: ≥ 0.50
- **LOW**: < 0.50

## Examples

### Basic Analysis

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

### Export Data

```bash
$ fogger export --format csv --since 7d --output weekly_results.csv
Exported 156 domains to weekly_results.csv
```

### Monitor Domain

```bash
$ fogger monitor suspicious-site.com --interval 2m --duration 1h
Monitoring suspicious-site.com every 2m0s for 1h0m0s
Scanning suspicious-site.com at 2023-12-24T10:30:00Z...
JLI Score: 0.78, Level: HIGH
Scanning suspicious-site.com at 2023-12-24T10:32:00Z...
JLI Score: 0.82, Level: HIGH
Monitoring completed
```

## Legal & Ethical Usage

### Permitted Use Cases

- Government cybercrime and digital enforcement units
- ISP and DNS operator abuse departments
- Payment processor compliance teams
- Academic research (with appropriate ethics approval)

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

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests if applicable
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Setup

```bash
# Clone the repository
git clone https://github.com/genesis410/fogger.git
cd fogger

# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build the binary
go build -o fogger main.go
```

## Architecture

### Tech Stack

- **Core Language**: Go (Golang)
- **CLI Framework**: Cobra
- **Configuration**: Viper
- **Output Formatting**: go-pretty/table
- **HTTP Client**: Built-in net/http with resty

### Design Philosophy

- **CLI-First**: Designed for analysts, researchers, and engineers
- **Scriptable**: Automatable and pipeline-friendly
- **Deterministic**: Idempotent and reproducible results
- **Explainable**: Transparent scoring with evidence breakdown

### Data Model

- **Domain Entity**: Core domain information with JLI score
- **Signal Entity**: Atomic, explainable indicators
- **Evidence Entity**: Human-auditable evidence
- **Cluster Entity**: Grouped domains by operator/campaign

## Testing

Run the test suite:

```bash
go test ./...
```

Run specific tests:

```bash
go test -v ./internal/analyzer
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built for government cybercrime and digital enforcement units
- Focused on ecosystem-level disruption rather than infrastructure suppression
- Designed to enable targeted, scalable enforcement without disrupting legitimate internet infrastructure

---

*This tool is designed to assist legitimate law enforcement and regulatory agencies in identifying and analyzing illegal gambling operations while respecting the legitimate use of CDN services.*