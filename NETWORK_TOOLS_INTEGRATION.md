# Network Analysis Tools Integration

## Overview
Comprehensive integration of 40+ network analysis tools for the NetMon monitoring platform.

## Tool Categories

### ICMP Tools
- **ping** - Standard ICMP echo for latency and reachability
- **fping** - High-performance parallel ICMP testing
- **arping** - ARP-level reachability testing
- **mtr** - Continuous traceroute with packet loss per hop

### Traceroute Tools
- **traceroute** - Standard path discovery
- **tracepath** - Path discovery with automatic MTU detection
- **tcptraceroute** - TCP-based traceroute for firewall traversal
- **paris-traceroute** - Flow-aware traceroute for load-balanced paths
- **dublin-traceroute** - Multi-path traceroute with NAT detection
- **scamper** - Advanced Internet measurement toolkit

### DNS Tools
- **dig** - DNS lookup and analysis
- **drill** - DNS lookup (ldnsutils)
- **kdig** - Advanced DNS queries (Knot DNS)

### HTTP/TLS Tools
- **curl** - HTTP/HTTPS transfer with detailed timing
- **httping** - HTTP latency measurement
- **openssl** - TLS/SSL analysis and certificate inspection

### Packet Analysis Tools
- **hping3** - Advanced TCP/UDP/ICMP packet crafting
- **nping** - Network packet generation (nmap)

### Packet Capture Tools
- **tcpdump** - Packet capture and analysis
- **tshark** - Wireshark CLI packet analyzer

### Throughput Tools
- **iperf3** - Bandwidth, jitter, and packet loss testing

### System Tools
- **ss** - Socket statistics and TCP connection analysis
- **ip** - Network configuration and routing
- **ethtool** - Ethernet device statistics
- **conntrack** - Connection tracking analysis

### Advanced Measurement
- **owamp** - One-way active measurement protocol
- **twamp** - Two-way active measurement protocol
- **bgpstream** - BGP routing data analysis
- **ripe-atlas** - Internet-wide measurement platform

## Installation

### Quick Install (Linux)
```bash
# Generate installation script
cd backend
go run cmd/tools/main.go install --generate --script /tmp/install-tools.sh

# Run installation
chmod +x /tmp/install-tools.sh
sudo /tmp/install-tools.sh
```

### Manual Installation
```bash
# Debian/Ubuntu
sudo apt-get update
sudo apt-get install -y \
  iputils-ping fping arping mtr-tiny \
  traceroute iputils-tracepath tcptraceroute \
  paris-traceroute dublin-traceroute scamper \
  dnsutils ldnsutils knot-dnsutils \
  curl httping openssl \
  hping3 nmap \
  tcpdump tshark \
  iperf3 \
  iproute2 ethtool conntrack \
  owamp-client twamp-client

# RHEL/CentOS/Fedora
sudo dnf install -y \
  iputils fping arping mtr \
  traceroute tcptraceroute \
  bind-utils \
  curl httping openssl \
  hping3 nmap \
  tcpdump wireshark-cli \
  iperf3 \
  iproute ethtool conntrack-tools
```

## CLI Usage

### Check Dependencies
```bash
# Check all tools
go run cmd/tools/main.go check

# Output:
# === Network Monitoring Tools Dependency Report ===
# Total: 40 | Installed: 25 | Missing: 15
# 
# ✓ Installed Tools:
#   ✓ ping (iputils-20211215)
#   ✓ curl (7.81.0)
#   ...
# 
# ✗ Missing Tools:
#   ✗ dublin-traceroute
#     Install: apt-get install dublin-traceroute
```

### List Tools
```bash
# List all tools
go run cmd/tools/main.go list

# List installed only
go run cmd/tools/main.go list --installed

# List by category
go run cmd/tools/main.go list --category icmp

# JSON output
go run cmd/tools/main.go list --json
```

### Execute Tools
```bash
# Ping
go run cmd/tools/main.go execute --tool ping 8.8.8.8 -c 4

# MTR
go run cmd/tools/main.go execute --tool mtr google.com --report --report-cycles 10

# Traceroute
go run cmd/tools/main.go execute --tool traceroute google.com -m 20

# DNS lookup
go run cmd/tools/main.go execute --tool dig google.com A

# HTTP timing
go run cmd/tools/main.go execute --tool curl https://google.com

# JSON output
go run cmd/tools/main.go execute --tool ping 8.8.8.8 -c 4 --json
```

### Tool Information
```bash
go run cmd/tools/main.go info --tool mtr

# Output:
# Tool: mtr
# Binary: mtr
# Category: traceroute
# Description: Continuous traceroute with packet loss per hop
# Installed: true
# Version: mtr 0.95
```

## API Endpoints

### List All Tools
```bash
GET /api/tools

Response:
{
  "tools": [
    {
      "name": "ping",
      "binary": "ping",
      "category": "icmp",
      "description": "ICMP echo request tool",
      "installed": true,
      "version": "iputils-20211215"
    }
  ],
  "count": 40
}
```

### List Installed Tools
```bash
GET /api/tools/installed
```

### Check Dependencies
```bash
GET /api/tools/dependencies

Response:
{
  "dependencies": {
    "ping": true,
    "fping": true,
    "mtr": false,
    ...
  },
  "installed": 25,
  "missing": 15,
  "total": 40
}
```

### Get Tool Info
```bash
GET /api/tools/{name}

Example: GET /api/tools/mtr
```

### Execute Tool
```bash
POST /api/tools/{name}/execute

Body:
{
  "args": ["google.com", "--report", "--report-cycles", "10"],
  "timeout": 30
}

Response:
{
  "tool": "mtr",
  "success": true,
  "exit_code": 0,
  "output": "...",
  "parsed_data": {
    "target": "google.com",
    "hops": [...]
  },
  "execution_time": 10234567890,
  "start_time": "2026-09-11T04:57:00Z",
  "end_time": "2026-09-11T04:57:10Z"
}
```

### Get Tools by Category
```bash
GET /api/tools/categories/{category}

Example: GET /api/tools/categories/icmp
```

## Specialized API Endpoints

### Ping
```bash
POST /api/tools/ping

Body:
{
  "host": "8.8.8.8",
  "count": 4
}

Response:
{
  "host": "8.8.8.8",
  "packets_sent": 4,
  "packets_received": 4,
  "packet_loss": 0,
  "min_rtt_ms": 10.5,
  "avg_rtt_ms": 12.3,
  "max_rtt_ms": 15.1,
  "stddev_rtt_ms": 1.8
}
```

### Traceroute
```bash
POST /api/tools/traceroute

Body:
{
  "host": "google.com",
  "max_hops": 30
}

Response:
{
  "target": "google.com",
  "hops": [
    {
      "hop": 1,
      "ip": "192.168.1.1",
      "hostname": "router.local",
      "rtt1_ms": 0.5,
      "rtt2_ms": 0.4,
      "rtt3_ms": 0.6,
      "timeout": false
    }
  ],
  "complete": true
}
```

### DNS Lookup
```bash
POST /api/tools/dns

Body:
{
  "domain": "google.com",
  "record_type": "A",
  "server": "8.8.8.8"
}

Response:
{
  "question": "google.com",
  "query_type": "A",
  "answers": [
    {
      "name": "google.com",
      "ttl": 300,
      "class": "IN",
      "type": "A",
      "data": "142.250.185.78"
    }
  ],
  "query_time_ms": 15,
  "server": "8.8.8.8:53",
  "status": "NOERROR"
}
```

## Programmatic Usage

### Go Code Example
```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/netmon/netmon/internal/tools"
)

func main() {
    // Get a tool
    tool, err := tools.GetTool("ping")
    if err != nil {
        panic(err)
    }
    
    // Check if installed
    if !tool.IsInstalled() {
        fmt.Println("Install:", tool.InstallCommand())
        return
    }
    
    // Execute with context
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    result, err := tool.Execute(ctx, []string{"-c", "4", "8.8.8.8"})
    if err != nil {
        panic(err)
    }
    
    // Parse output
    parsed, err := tool.ParseOutput(result.Output)
    if err != nil {
        panic(err)
    }
    
    pingResult := parsed.(*tools.PingResult)
    fmt.Printf("Average RTT: %.2f ms\n", pingResult.AvgRTT)
}
```

### Using Specific Tools
```go
// MTR
mtrTool := tools.NewMTRTool()
result, err := mtrTool.MTR(ctx, "google.com", 10, true)
for _, hop := range result.Hops {
    fmt.Printf("Hop %d: %s (%.2f ms, %.1f%% loss)\n", 
        hop.Hop, hop.Host, hop.Avg, hop.Loss)
}

// Iperf3
iperf3Tool := tools.NewIperf3Tool()
result, err := iperf3Tool.Iperf3Client(ctx, "iperf.example.com", 5201, 10, "tcp", false)
fmt.Printf("Throughput: %.2f Mbps\n", 
    result.End.SumReceived.BitsPerSecond / 1_000_000)

// Dig
digTool := tools.NewDigTool()
result, err := digTool.Dig(ctx, "google.com", "A", "8.8.8.8")
for _, answer := range result.Answers {
    fmt.Printf("%s -> %s\n", answer.Name, answer.Data)
}
```

## Output Parsing

All tools implement structured output parsing:

- **Ping/FPing/Arping**: Packet statistics, RTT measurements
- **MTR**: Per-hop latency and packet loss
- **Traceroute**: Hop-by-hop path with RTT
- **DNS**: Parsed resource records
- **HTTP**: Timing breakdown (DNS, connect, TLS, TTFB, total)
- **Iperf3**: JSON output with detailed throughput stats
- **TShark/TCPDump**: Packet-level analysis
- **System tools**: Structured connection/interface stats

## Integration with Monitors

Network analysis tools can be integrated into existing monitors:

```go
// Create a monitor that uses external tools
type ExternalToolMonitor struct {
    tool     tools.Tool
    args     []string
    parser   func(interface{}) *models.Measurement
}

func (m *ExternalToolMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
    result, err := m.tool.Execute(ctx, m.args)
    if err != nil {
        return nil, err
    }
    
    parsed, err := m.tool.ParseOutput(result.Output)
    if err != nil {
        return nil, err
    }
    
    return m.parser(parsed), nil
}
```

## Performance Considerations

- Tools execute as subprocesses - use sparingly in high-frequency monitors
- Prefer native Go implementations for high-volume testing
- Use tool results for detailed analysis and troubleshooting
- Cache tool availability checks
- Set appropriate timeouts
- Consider rate limiting for API endpoints

## Security Considerations

- Many tools require elevated privileges (raw sockets)
- Validate all user inputs before passing to tools
- Sanitize command arguments to prevent injection
- Use context timeouts to prevent resource exhaustion
- Limit concurrent tool executions
- Audit tool execution logs

## Platform Support

### Linux (Full Support)
All tools available via package managers.

### Windows (Limited)
- ping, tracert (built-in)
- curl (if installed)
- Most Unix tools require WSL or Cygwin

### macOS (Partial)
- Basic tools (ping, traceroute, curl) built-in
- Install others via Homebrew

## Files Created

```
backend/internal/tools/
├── tool.go               # Base tool interface and registry
├── icmp.go              # Ping, fping, arping, mtr
├── traceroute.go        # All traceroute variants
├── dns.go               # Dig, drill, kdig
├── http.go              # Curl, httping, openssl
├── packet_analysis.go   # Hping3, nping
├── packet_capture.go    # Tcpdump, tshark
├── throughput.go        # Iperf3
├── system.go            # ss, ip, ethtool, conntrack
├── advanced.go          # OWAMP, TWAMP, etc.
├── installer.go         # Dependency checker and installer
└── registry.go          # Global registry

backend/internal/api/
└── tools_handler.go     # REST API handlers

backend/cmd/tools/
└── main.go              # CLI tool
```

## Next Steps

1. Add tools to monitoring workflows
2. Create dashboard visualizations for tool results
3. Implement tool result storage in database
4. Add scheduled tool execution
5. Create tool result comparison/trending
6. Implement tool-based alerting
7. Add more specialized parsers
8. Integrate with existing monitors

## Documentation

- API documentation: `/api/tools` endpoints
- CLI help: `go run cmd/tools/main.go --help`
- Tool-specific docs: `go run cmd/tools/main.go info --tool <name>`
