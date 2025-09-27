# Smokeping

A network monitoring tool that uses MTR (My Traceroute) to continuously measure latency and packet loss to specified targets, exposing metrics in Prometheus format.

Smokeping provides continuous network monitoring by:

- Running MTR measurements at regular intervals against specified targets
- Collecting latency and packet loss metrics
- Exposing metrics via HTTP endpoint for Prometheus scraping
- Supporting custom labels for flexible monitoring setups

## Dependencies

- **MTR** - Must be installed and available in PATH
  - Ubuntu/Debian: `sudo apt-get install mtr`
  - RHEL/CentOS: `sudo yum install mtr`
  - macOS: `brew install mtr`

## Installation

### Build from source

```bash
git clone https://github.com/elek/smokeping
cd smokeping
go build -o smokeping .
```

## Usage

### Basic usage

Monitor single target (label:ip_address)

```bash
./smokeping example.com:1.2.3.4
```

Monitor multiple targets:
```bash
./smokeping google.com:8.8.8.8 cloudflare.com:1.1.1.1
```

### Advanced usage

With custom port and interval:
```bash
./smokeping -p 9090 -i 30 example.com:1.2.3.4
```

With source IP and labels:
```bash
./smokeping -s 192.168.1.10 -l datacenter=us-east -l environment=prod ...
```
### Target specification

Targets can be specified as:
- `hostname` - Uses hostname as both label and target
- `label:ip_address` - Uses custom label with specific IP address

Examples:
- `google.com` - Label: "google.com", Target: "google.com"
- `google:8.8.8.8` - Label: "google", Target: "8.8.8.8"

## Metrics

The service exposes the following Prometheus metrics at `/metrics`:

- `smokeping_latency_ms` - Average latency to target in milliseconds
- `smokeping_packet_loss_percent` - Packet loss percentage to target

Both metrics include labels:
- `source` - Source IP used for measurements
- `target` - Target label specified in command line
- Any additional custom labels specified with `-l`

## Endpoints

- `/metrics` - Prometheus metrics endpoint
- `/health` - Health check endpoint (returns "OK")

## Example Prometheus configuration

```yaml
scrape_configs:
  - job_name: 'smokeping'
    static_configs:
      - targets: ['localhost:8080']
    scrape_interval: 60s
```

## Build instructions

### Prerequisites

- Go 1.25 or later
- MTR installed on the system

### Build

```bash
# Clone repository
git clone https://github.com/elek/smokeping
cd smokeping

docker buildx bakse --push
```
