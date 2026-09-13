# ONCIC Network Monitoring - DNS Reachability Tester

Tests reachability and response time of 20 major public DNS servers worldwide.

## Features

- Tests 20 DNS providers (40+ individual servers)
- Measures latency for each server
- Validates DNS resolution capability
- Identifies fastest DNS servers
- Calculates success rate and statistics
- Exports results to JSON
- Beautiful Rich terminal UI

## DNS Providers Tested

1. **Google Public DNS** - 8.8.8.8, 8.8.4.4
2. **Cloudflare DNS** - 1.1.1.1, 1.0.0.1
3. **Quad9** - 9.9.9.9, 149.112.112.112
4. **OpenDNS (Cisco)** - 208.67.222.222, 208.67.220.220
5. **AdGuard DNS** - 94.140.14.14, 94.140.15.15
6. **CleanBrowsing** - 185.228.168.168, 185.228.169.168
7. **Comodo Secure DNS** - 8.26.56.26, 8.20.247.20
8. **Quad101** - 101.101.101.101, 101.102.103.104
9. **OpenNIC** - 217.160.70.42
10. **SWITCH DNS** - 130.59.31.248
11. **DNS.SB** - 185.222.222.222, 45.11.45.11
12. **Verisign Public DNS** - 64.6.64.6, 64.6.65.6
13. **Level3** - 4.2.2.1, 4.2.2.2
14. **Neustar UltraRecursive** - 64.6.64.6, 64.6.65.6
15. **Control D** - 76.76.2.0, 76.76.10.0
16. **Alternate DNS** - 76.76.19.19, 76.223.122.150
17. **Yandex DNS** - 77.88.8.8, 77.88.8.1
18. **AliDNS** - 223.5.5.5, 223.6.6.6
19. **DNSPod** - 119.29.29.29, 182.254.116.116
20. **114DNS** - 114.114.114.114, 114.114.115.115

## Requirements

```bash
pip install rich
```

## Usage

### Basic Test
```bash
python dns-reachability.py
```

### Custom Domain
```bash
python dns-reachability.py --domain youtube.com
```

### Custom Timeout
```bash
python dns-reachability.py --timeout 10
```

### Export to JSON
```bash
python dns-reachability.py --json results.json
```

## Output

The tool provides:

1. **Summary Table** - All providers with primary/secondary server status
2. **Statistics** - Success rate, average latency, fastest server
3. **Working Servers** - Only reachable servers sorted by speed (fastest first)

### Example Output

```
DNS Server Reachability Test Results
┌────┬─────────────────────┬─────────────────┬────────────┬──────────┬─────────────────┬────────────┬──────────┐
│ #  │ Provider            │ Primary IP      │ Status     │ Latency  │ Secondary IP    │ Status     │ Latency  │
├────┼─────────────────────┼─────────────────┼────────────┼──────────┼─────────────────┼────────────┼──────────│
│ 1  │ Google Public DNS   │ 8.8.8.8         │ ✓ ONLINE   │ 15 ms    │ 8.8.4.4         │ ✓ ONLINE   │ 16 ms    │
│ 2  │ Cloudflare DNS      │ 1.1.1.1         │ ✓ ONLINE   │ 8 ms     │ 1.0.0.1         │ ✓ ONLINE   │ 9 ms     │
└────┴─────────────────────┴─────────────────┴────────────┴──────────┴─────────────────┴────────────┴──────────┘

Statistics Summary
┌─────────────────────────┬────────────┐
│ Metric                  │ Value      │
├─────────────────────────┼────────────┤
│ Total DNS Providers     │ 20         │
│ Total Servers Tested    │ 38         │
│ Working Servers         │ 35         │
│ Failed Servers          │ 3          │
│ Success Rate            │ 92.1%      │
│ Average Latency         │ 45.3 ms    │
│ Fastest Server          │ 8.2 ms     │
└─────────────────────────┴────────────┘

✔ WORKING DNS SERVERS (Sorted by Speed)
┌──────┬─────────────────────┬─────────────────┬───────────┬──────────┬─────────────────────────┐
│ Rank │ Provider            │ Server IP       │ Type      │ Latency  │ Resolved IPs            │
├──────┼─────────────────────┼─────────────────┼───────────┼──────────┼─────────────────────────┤
│ 2    │ Cloudflare DNS      │ 1.1.1.1         │ Primary   │ 8 ms     │ 142.250.185.14          │
│ 2    │ Cloudflare DNS      │ 1.0.0.1         │ Secondary │ 9 ms     │ 142.250.185.14          │
│ 1    │ Google Public DNS   │ 8.8.8.8         │ Primary   │ 15 ms    │ 142.250.185.14          │
└──────┴─────────────────────┴─────────────────┴───────────┴──────────┴─────────────────────────┘
```

## JSON Export Format

```json
{
  "rank": 1,
  "name": "Google Public DNS",
  "primary": {
    "ip": "8.8.8.8",
    "reachable": true,
    "latency_ms": 15.3,
    "resolved_ips": ["142.250.185.14"],
    "error": null
  },
  "secondary": {
    "ip": "8.8.4.4",
    "reachable": true,
    "latency_ms": 16.1,
    "resolved_ips": ["142.250.185.14"],
    "error": null
  }
}
```

## Integration with ONCIC

This tool is part of the ONCIC Network Monitoring platform and can be:
- Integrated into the backend monitoring engine
- Used for DNS health checks
- Automated via cron/scheduler
- Exposed via REST API endpoint

## Use Cases

- **Network Diagnostics** - Identify which DNS servers are accessible
- **Performance Testing** - Find the fastest DNS servers for your location
- **Availability Monitoring** - Check if public DNS is being filtered/blocked
- **Network Troubleshooting** - Diagnose DNS connectivity issues

## Created by Medo
Part of ONCIC Network Monitoring Platform
