# ONCIC Network Monitoring - DoH Server Tester

Tests DNS over HTTPS (DoH) endpoints worldwide and validates domain resolution.

## Features

- Tests 100+ DoH servers across multiple providers
- Validates domain resolution for popular sites
- Measures latency and HTTP version support
- Identifies blocked or filtered servers
- Exports results to JSON
- Beautiful Rich terminal UI

## Requirements

```bash
pip install httpx rich
```

## Usage

### Basic Test
```bash
python doh-tester.py
```

### Custom Domains
```bash
python doh-tester.py --domains google.com youtube.com facebook.com
```

### Export to JSON
```bash
python doh-tester.py --json results.json
```

## Output

The tool provides:

1. **Reachability Summary** - All servers with status, latency, HTTP version
2. **Domain Resolution Details** - Resolved IPs for each domain per server
3. **Working Servers Only** - Final table with only accessible servers

## DoH Servers Tested

- **100+ servers** across worldwide providers
- **Major Providers**: Cloudflare, Google, Quad9, AdGuard, OpenDNS
- **Privacy-focused**: Mullvad, LibreDNS, BlahDNS
- **Regional**: US, EU, Asia, Australia coverage
- **Filtered Options**: Family filters, ad-blocking, malware protection

## Results Format

### Console Output
- ✅ Color-coded status indicators
- 📊 Sorted by latency (fastest first)
- 🌍 Regional information
- 🔍 Detailed per-domain resolution

### JSON Export
```json
{
  "name": "Cloudflare",
  "url": "https://cloudflare-dns.com/dns-query",
  "region": "Main",
  "reachable": true,
  "latency_ms": 15.3,
  "http_version": "HTTP/2",
  "domains": {
    "youtube.com": {
      "ips": ["142.250.185.14"],
      "error": null
    }
  }
}
```

## Integration with ONCIC

This tool is part of the ONCIC Network Monitoring platform and can be:
- Integrated into the backend monitoring engine
- Used for DNS health checks
- Automated via cron/scheduler
- Exposed via REST API

## Created by Medo
Part of ONCIC Network Monitoring Platform
