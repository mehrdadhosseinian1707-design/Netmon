# 🔧 ONCIC Network Monitoring - Tools

Advanced network diagnostic and testing tools for the ONCIC platform.

## Available Tools

### 1. DNS Reachability Tester (`dns-reachability.py`) ⭐ NEW

Tests reachability and response time of 20 major public DNS servers.

**Features:**
- Tests 20 DNS providers (40+ servers)
- Measures latency for each server
- Validates DNS resolution
- Identifies fastest DNS servers
- Statistics and success rate
- Beautiful Rich terminal UI
- JSON export capability

**Quick Start:**
```bash
cd tools
pip install rich
python dns-reachability.py
```

**Usage:**
```bash
# Basic test
python dns-reachability.py

# Custom domain
python dns-reachability.py --domain youtube.com

# Custom timeout
python dns-reachability.py --timeout 10

# Export results
python dns-reachability.py --json results.json
```

**DNS Providers:**
- Google, Cloudflare, Quad9
- OpenDNS, AdGuard, CleanBrowsing
- Level3, Yandex, AliDNS
- 20 providers worldwide

---

### 2. DoH Server Tester (`doh-tester.py`)

Tests DNS over HTTPS (DoH) endpoints worldwide and validates domain resolution.

**Features:**
- Tests 100+ DoH servers globally
- Measures latency and HTTP/2 support
- Validates domain resolution
- Identifies blocked/filtered servers
- Beautiful Rich terminal UI
- JSON export capability

**Quick Start:**
```bash
cd tools
pip install httpx rich
python doh-tester.py
```

**Usage:**
```bash
# Basic test
python doh-tester.py

# Custom domains
python doh-tester.py --domains google.com youtube.com

# Export results
python doh-tester.py --json results.json
```

**Providers Tested:**
- Cloudflare, Google, Quad9
- AdGuard, OpenDNS, DNS.SB
- Mullvad, LibreDNS, BlahDNS
- 100+ servers across all continents

---

## Installation

```bash
# Install dependencies
pip install -r requirements.txt
```

## Integration

These tools can be:
- ✅ Run standalone for diagnostics
- ✅ Integrated into ONCIC backend
- ✅ Exposed via REST API endpoints
- ✅ Automated via scheduler/cron
- ✅ Used in monitoring workflows

## Future Tools

Planned additions:
- [ ] Bandwidth tester (iperf3 wrapper)
- [ ] Traceroute analyzer
- [ ] BGP route explorer
- [ ] TLS/SSL certificate checker
- [ ] Network path analyzer

---

**Part of ONCIC Network Monitoring**  
Created by Medo
