# Network Analysis Tools - Complete Measurement Matrix

## Tool Inventory: 60+ Tools Integrated

### Priority Legend
- ⭐⭐⭐ **Essential** - Core monitoring, must have
- ⭐⭐ **High** - Advanced diagnostics, recommended
- ⭐ **Medium** - Specialized use cases, optional

---

## 1. Basic Reachability & Packet Quality

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **ping** | ⭐⭐⭐ | Latency, reachability, packet loss, jitter | `apt-get install iputils-ping` |
| **ping6** | ⭐⭐⭐ | IPv6 latency, reachability, packet loss | `apt-get install iputils-ping` |
| **fping** | ⭐⭐⭐ | Fast parallel ICMP, large-scale probing | `apt-get install fping` |
| **hping3** | ⭐⭐ | TCP/UDP/ICMP packet-level tests, custom flags | `apt-get install hping3` |
| **nping** | ⭐⭐ | Flexible packet generation, protocol testing | `apt-get install nmap` |
| **arping** | ⭐⭐ | Layer-2 (ARP) reachability, MAC address discovery | `apt-get install arping` |
| **mtr** | ⭐⭐⭐ | Continuous latency + packet loss per hop | `apt-get install mtr-tiny` |

**Measurements:**
- Round-trip time (min/avg/max/stddev)
- Packet loss percentage
- Jitter (variation in latency)
- IP address resolution
- ARP/Layer-2 connectivity

**Use Cases:**
- Basic connectivity verification
- Latency monitoring
- Packet loss detection
- Path quality over time (mtr)
- Large-scale network scanning (fping)

---

## 2. Routing / Path Analysis

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **traceroute** | ⭐⭐⭐ | Hop-by-hop path, per-hop latency | `apt-get install traceroute` |
| **traceroute6** | ⭐⭐⭐ | IPv6 hop-by-hop path | `apt-get install iputils-tracepath` |
| **tracepath** | ⭐⭐ | Path + MTU discovery, no root required | `apt-get install iputils-tracepath` |
| **mtr** | ⭐⭐⭐ | Historical/continuous path quality | `apt-get install mtr-tiny` |
| **paris-traceroute** | ⭐⭐ | More reliable path measurement, ECMP-aware | `apt-get install paris-traceroute` |
| **scamper** | ⭐⭐ | Large-scale traceroute/ping measurement | `apt-get install scamper` |
| **dublin-traceroute** | ⭐ | TCP/UDP/ICMP path comparison, NAT detection | `apt-get install dublin-traceroute` |
| **tcptraceroute** | ⭐⭐ | TCP-based traceroute, firewall bypass | `apt-get install tcptraceroute` |

**Measurements:**
- Hop count
- Per-hop latency
- Per-hop packet loss (mtr)
- MTU (tracepath)
- Path changes
- Route asymmetry
- AS path (with ASN lookup)

**Use Cases:**
- Route discovery
- Path quality analysis
- MTU problems diagnosis
- Firewall/filtering detection
- Multi-path (ECMP) analysis
- ISP routing comparison

---

## 3. DNS Monitoring

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **dig** | ⭐⭐⭐ | DNS queries, detailed responses, timing | `apt-get install dnsutils` |
| **drill** | ⭐⭐ | DNS diagnostics, DNSSEC validation | `apt-get install ldnsutils` |
| **nslookup** | ⭐⭐ | Basic DNS testing, legacy tool | `apt-get install dnsutils` |
| **kdig** | ⭐⭐ | Advanced DNS querying (Knot DNS) | `apt-get install knot-dnsutils` |
| **dnsviz** | ⭐ | DNS delegation analysis, DNSSEC visualization | `apt-get install dnsviz` |
| **resolvectl** | ⭐⭐ | Local resolver testing, systemd-resolved | `apt-get install systemd` |

**Measurements:**
- DNS response time
- SERVFAIL errors
- NXDOMAIN (non-existent domain)
- Timeouts
- Wrong/inconsistent answers
- Resolver availability
- DNSSEC validation
- UDP vs TCP DNS
- IPv4 vs IPv6 DNS resolution
- TTL values
- Authoritative vs cached responses

**Use Cases:**
- DNS resolver performance
- DNS propagation testing
- DNSSEC validation
- DNS hijacking detection
- Resolver comparison (ISP vs public DNS)
- DNS-based blacklisting detection

---

## 4. TCP Connectivity

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **nc (netcat)** | ⭐⭐⭐ | TCP/UDP connectivity, port scanning | `apt-get install netcat-openbsd` |
| **ncat** | ⭐⭐⭐ | Modern netcat with SSL/TLS | `apt-get install nmap` |
| **hping3** | ⭐⭐ | TCP SYN/ACK testing, flag manipulation | `apt-get install hping3` |
| **curl** | ⭐⭐⭐ | TCP + application-layer testing | `apt-get install curl` |
| **telnet** | ⭐⭐ | Basic TCP connection testing | `apt-get install telnet` |

**Test Ports:**
- 22 (SSH)
- 53 (DNS)
- 80 (HTTP)
- 443 (HTTPS)
- 853 (DNS-over-TLS)
- 3306 (MySQL)
- 5432 (PostgreSQL)
- Custom monitoring ports

**Measurements:**
- TCP connect time
- Connection success/failure
- Port open/closed/filtered
- TCP handshake timing
- Banner grabbing
- Firewall detection

**Use Cases:**
- Port availability testing
- Firewall rule verification
- Service availability
- Connection timeout debugging

---

## 5. HTTP/HTTPS Monitoring

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **curl** | ⭐⭐⭐ | HTTP timing, headers, redirects, content | `apt-get install curl` |
| **wget** | ⭐⭐ | HTTP downloads, timing, retry logic | `apt-get install wget` |
| **httping** | ⭐⭐⭐ | HTTP latency, TTFB, availability | `apt-get install httping` |
| **xh** | ⭐ | Modern HTTP client, JSON support | `cargo install xh` |

**Measurements:**
- TCP connect time
- TLS handshake time
- Time to first byte (TTFB)
- HTTP response time
- HTTP status codes (2xx, 3xx, 4xx, 5xx)
- Redirect chains
- Download time
- Content size
- HTTP/2 vs HTTP/3
- IPv4 vs IPv6 performance
- Certificate validation
- Response headers

**Use Cases:**
- Web service monitoring
- API endpoint testing
- CDN performance
- SSL/TLS issues
- HTTP vs HTTPS comparison
- Geographic latency testing

---

## 6. TLS Monitoring

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **openssl s_client** | ⭐⭐⭐ | TLS handshake, certificates, protocols | `apt-get install openssl` |
| **gnutls-cli** | ⭐⭐ | GnuTLS testing, alternative to OpenSSL | `apt-get install gnutls-bin` |
| **curl** | ⭐⭐⭐ | TLS timing integrated with HTTP | `apt-get install curl` |

**Measurements:**
- TLS handshake time
- Certificate validity
- Certificate expiration date
- Days until expiration
- TLS version (1.0, 1.1, 1.2, 1.3)
- Cipher suite
- Certificate chain
- Certificate issuer
- Subject Alternative Names (SAN)
- SNI (Server Name Indication) behavior
- TLS failures/errors
- Certificate revocation (OCSP/CRL)

**Use Cases:**
- Certificate expiration monitoring
- TLS configuration audit
- SSL/TLS vulnerability testing
- Cipher suite analysis
- Certificate chain validation
- TLS version compliance

---

## 7. Packet Capture / Forensic Analysis

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **tcpdump** | ⭐⭐⭐ | Raw packet capture, protocol analysis | `apt-get install tcpdump` |
| **tshark** | ⭐⭐⭐ | Wireshark CLI, advanced filtering | `apt-get install tshark` |
| **dumpcap** | ⭐⭐ | Lightweight capture, minimal overhead | `apt-get install wireshark` |
| **Wireshark** | ⭐⭐ | Full GUI packet analysis | `apt-get install wireshark` |

**Measurements:**
- Packet-level details
- Protocol distributions
- Retransmissions
- TCP flags
- Packet timing
- Out-of-order packets
- Fragmentation
- Application layer protocols
- Malformed packets

**Critical Use:**
These tools determine **why** a measurement failed, not just **that** it failed.

**Use Cases:**
- Troubleshooting connectivity issues
- Protocol analysis
- Security investigations
- Performance debugging
- Packet loss root cause
- Application protocol issues

---

## 8. Traffic / Interface Monitoring

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **iftop** | ⭐⭐ | Real-time bandwidth per connection | `apt-get install iftop` |
| **nload** | ⭐⭐ | Console bandwidth monitor | `apt-get install nload` |
| **bmon** | ⭐ | Bandwidth monitoring, rate estimating | `apt-get install bmon` |
| **vnstat** | ⭐⭐ | Long-term traffic statistics | `apt-get install vnstat` |
| **ip -s link** | ⭐⭐⭐ | Interface statistics (built-in) | `apt-get install iproute2` |
| **ethtool** | ⭐⭐⭐ | Interface health, negotiated speed | `apt-get install ethtool` |

**Measurements:**
- RX/TX bytes
- RX/TX packets
- Errors (RX/TX)
- Dropped packets (RX/TX)
- Interface state (up/down)
- Current bandwidth utilization
- Peak bandwidth
- Historical traffic (vnstat)
- Negotiated link speed
- Duplex mode
- Interface counters

**Use Cases:**
- Interface health monitoring
- Bandwidth usage tracking
- Error rate monitoring
- Link speed verification
- Traffic pattern analysis
- Capacity planning

---

## 9. Path / Internet Measurement

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **RIPE Atlas** | ⭐⭐⭐ | Global measurement probes | API integration |
| **scamper** | ⭐⭐ | Large-scale measurements | `apt-get install scamper` |
| **mtr** | ⭐⭐⭐ | Continuous path monitoring | `apt-get install mtr-tiny` |
| **traceroute** | ⭐⭐⭐ | Path discovery | `apt-get install traceroute` |
| **Paris Traceroute** | ⭐⭐ | ECMP-aware tracing | `apt-get install paris-traceroute` |
| **Dublin Traceroute** | ⭐ | Multi-protocol path analysis | `apt-get install dublin-traceroute` |

**RIPE Atlas Benefits:**
- Compare your measurements against external probes
- Measure from multiple geographic locations
- Detect ISP-specific issues
- Global reachability testing
- Independent verification

**Use Cases:**
- ISP performance comparison
- Geographic latency analysis
- Path diversity testing
- Global service monitoring
- ISP routing issues detection

---

## 10. BGP / Internet Routing

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **BGPStream** | ⭐⭐ | BGP routing events, real-time | API integration |
| **bgpdump** | ⭐⭐ | BGP MRT dump file parsing | `apt-get install bgpdump` |
| **pybgpstream** | ⭐⭐ | Python BGP stream analysis | `pip3 install pybgpstream` |
| **RIPE RIS** | ⭐⭐ | Routing Information Service | API integration |
| **RouteViews** | ⭐ | BGP routing table archive | API integration |

**Measurements:**
- ASN (Autonomous System Number)
- Prefix announcements
- Prefix withdrawals
- Route changes
- AS path
- Upstream changes
- Routing instability
- BGP hijacking detection
- Route leaks

**Use Cases (ISP Analysis):**
- Monitor your ISP's BGP health
- Detect routing changes
- AS path stability
- Prefix announcement monitoring
- BGP hijacking alerts
- Upstream provider changes
- Routing policy changes

---

## 11. Throughput / Bandwidth

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **iperf3** | ⭐⭐⭐ | TCP/UDP throughput, controlled endpoints | `apt-get install iperf3` |
| **speedtest-cli** | ⭐⭐⭐ | Real-world ISP speed (Speedtest.net) | `apt-get install speedtest-cli` |
| **fast-cli** | ⭐⭐ | Netflix speed test | `npm install -g fast-cli` |
| **curl** | ⭐⭐ | Download speed tests | `apt-get install curl` |
| **wget** | ⭐⭐ | Download speed, retry logic | `apt-get install wget` |

**Measurements:**
- Download speed (Mbps)
- Upload speed (Mbps)
- TCP throughput
- UDP throughput
- Retransmissions
- Jitter
- TCP window size behavior
- UDP packet loss
- Congestion window (cwnd)
- RTT during transfer

**Best Tool Choice:**
- **iperf3**: Controlled testing (you control both endpoints)
- **speedtest-cli**: Real ISP performance
- **fast-cli**: Netflix CDN performance

**Use Cases:**
- ISP speed verification
- Bandwidth capacity testing
- Network bottleneck identification
- QoS verification
- CDN performance testing

---

## 12. UDP / VoIP Quality

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **iperf3** | ⭐⭐⭐ | UDP throughput, jitter, loss | `apt-get install iperf3` |
| **hping3** | ⭐⭐ | UDP packet testing | `apt-get install hping3` |
| **nping** | ⭐⭐ | UDP probing | `apt-get install nmap` |
| **twamp** | ⭐⭐ | Two-way active measurement | `apt-get install twamp-client` |
| **owamp** | ⭐⭐ | One-way delay measurement | `apt-get install owamp-client` |

**Measurements:**
- UDP packet loss
- Jitter (packet delay variation)
- One-way delay
- Two-way delay
- Packet reordering
- Out-of-order delivery
- Duplicate packets

**Use Cases:**
- VoIP quality assessment
- Video streaming performance
- Real-time application testing
- Gaming latency testing
- UDP-based service monitoring

---

## 13. MTU / Fragmentation

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **ping -M do** | ⭐⭐⭐ | PMTU discovery, fragmentation | Built-in |
| **tracepath** | ⭐⭐⭐ | Automatic MTU discovery | `apt-get install iputils-tracepath` |
| **traceroute** | ⭐⭐ | MTU with manual testing | `apt-get install traceroute` |
| **hping3** | ⭐⭐ | Custom packet sizes | `apt-get install hping3` |

**Measurements:**
- Path MTU (PMTU)
- Fragmentation behavior
- ICMP fragmentation-needed messages
- Black-hole MTU problems
- MSS (Maximum Segment Size)

**Use Cases:**
- MTU misconfiguration detection
- Fragmentation issues
- VPN MTU problems
- PPPoE MTU issues
- Performance optimization

---

## 14. IPv4 / IPv6 Dual-Stack

| Tool | Priority | What It Measures | Install Command |
|------|----------|------------------|-----------------|
| **ping / ping6** | ⭐⭐⭐ | Both protocol latencies | `apt-get install iputils-ping` |
| **traceroute / traceroute6** | ⭐⭐⭐ | Both protocol paths | `apt-get install traceroute` |
| **curl** | ⭐⭐⭐ | HTTP over v4/v6 | `apt-get install curl` |
| **dig** | ⭐⭐⭐ | DNS A vs AAAA records | `apt-get install dnsutils` |

**Critical Measurements:**
Record IPv4 and IPv6 **independently** for:
- Latency comparison
- Path differences
- Performance deltas
- IPv6 availability
- DNS resolution time
- Reachability differences

**Use Cases:**
- IPv6 deployment verification
- Performance comparison v4 vs v6
- IPv6 connectivity issues
- Dual-stack behavior analysis

---

## Core Toolkit for 3-ISP Monitoring

### Essential (Must Have) ⭐⭐⭐

| Tool | Purpose |
|------|---------|
| **ping** | Latency/loss baseline |
| **fping** | Large-scale parallel testing |
| **mtr** | Path + loss monitoring |
| **traceroute** | Route discovery |
| **dig** | DNS performance |
| **curl** | HTTP/HTTPS testing |
| **openssl s_client** | TLS validation |
| **tcpdump** | Packet-level troubleshooting |
| **iperf3** | Throughput testing |
| **ethtool** | Interface health |
| **ip** | System network state |
| **speedtest-cli** | Real-world ISP speed |

### High Priority (Recommended) ⭐⭐

| Tool | Purpose |
|------|---------|
| **scamper** | Advanced measurements |
| **tshark** | Automated packet analysis |
| **hping3** | TCP/UDP/ICMP tests |
| **tracepath** | MTU discovery |
| **vnstat** | Traffic statistics |
| **BGPStream** | BGP/routing intelligence |
| **RIPE Atlas** | External verification |
| **httping** | HTTP latency |
| **ss** | Socket statistics |

### Medium Priority (Specialized) ⭐

| Tool | Purpose |
|------|---------|
| **dnsviz** | DNSSEC analysis |
| **dublin-traceroute** | Multi-path analysis |
| **paris-traceroute** | ECMP detection |
| **nload/iftop** | Real-time bandwidth |
| **fast-cli** | Netflix performance |
| **resolvectl** | Local resolver |

---

## Measurement Frequency Recommendations

### Continuous (Every 1-5 seconds)
- mtr (for critical paths)
- Interface statistics (ethtool, ip -s link)

### High Frequency (Every 10-60 seconds)
- ping (basic reachability)
- fping (multi-target)
- HTTP endpoint checks (curl, httping)

### Medium Frequency (Every 5-15 minutes)
- traceroute (path changes)
- DNS lookups (dig)
- TLS certificate checks (openssl)
- TCP connectivity (nc)

### Low Frequency (Every hour or daily)
- speedtest-cli (bandwidth)
- iperf3 (controlled throughput)
- BGP analysis (BGPStream)
- Traffic statistics (vnstat)
- Full packet capture (tcpdump - on-demand)

---

## Tool Output Integration

All tools in NetMon implement:

1. **Structured Parsing**: Raw output → Go structs
2. **JSON Serialization**: API-ready responses
3. **Measurement Mapping**: Tool results → database fields
4. **Error Handling**: Graceful failures, timeout management
5. **Context Support**: Cancellation, deadlines

---

## Total Tools Integrated: 60+

- **ICMP**: 5 tools (ping, ping6, fping, arping, mtr)
- **Traceroute**: 8 tools (traceroute, traceroute6, tracepath, mtr, paris, dublin, tcp, scamper)
- **DNS**: 6 tools (dig, drill, kdig, nslookup, dnsviz, resolvectl)
- **HTTP/HTTPS**: 5 tools (curl, wget, httping, xh, openssl)
- **TLS**: 2 tools (openssl, gnutls-cli)
- **TCP Connectivity**: 4 tools (nc, ncat, telnet, hping3)
- **Packet Analysis**: 2 tools (hping3, nping)
- **Packet Capture**: 3 tools (tcpdump, tshark, dumpcap)
- **Throughput**: 5 tools (iperf3, speedtest-cli, fast-cli, curl, wget)
- **System/Interface**: 6 tools (ss, ip, ethtool, conntrack, iftop, nload, bmon, vnstat)
- **BGP/Routing**: 5 tools (bgpdump, pybgpstream, BGPStream, RouteViews, RIPE RIS)
- **Advanced**: 5 tools (owamp, twamp, RIPE Atlas, scamper, BGPStream)

---

## Installation Quick Reference

### Debian/Ubuntu Complete Install
```bash
sudo apt-get update && sudo apt-get install -y \
  iputils-ping fping arping mtr-tiny \
  traceroute iputils-tracepath tcptraceroute paris-traceroute dublin-traceroute \
  dnsutils ldnsutils knot-dnsutils dnsviz \
  curl wget httping openssl gnutls-bin \
  netcat-openbsd nmap hping3 telnet \
  tcpdump tshark wireshark \
  iperf3 speedtest-cli \
  iproute2 ethtool conntrack iftop nload bmon vnstat \
  bgpdump scamper \
  owamp-client twamp-client

# Node.js tools
npm install -g fast-cli

# Rust tools (optional)
cargo install xh

# Python tools
pip3 install pybgpstream
```

---

**Last Updated**: 2026-09-11T05:04:12.880Z
**NetMon Version**: Extended Tools v2.0
**Total Tools**: 60+
**Coverage**: Complete network monitoring stack
