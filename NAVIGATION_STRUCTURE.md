# NetMon Complete Navigation Structure

Based on the 50 capabilities documented in "NetMon - Complete Capabilities.txt"

## Navigation Menu Structure

### 1. OVERVIEW
- **Dashboard** - Main overview with real-time statistics
- **System Status** - System health and probe status

### 2. MONITORING
- **Targets** - Monitor targets management (hosts, servers, services)
- **Monitors** - Monitor configuration and management
- **Probes** - Distributed probe management and status
- **Real-time Dashboard** - Live monitoring data stream

### 3. NETWORK PERFORMANCE
- **Latency Monitoring** - RTT, jitter, packet loss
- **Bandwidth Testing** - Download/upload speed tests
- **Throughput Analysis** - TCP/UDP throughput
- **Quality Analysis** - Network quality metrics
- **SLA Monitoring** - Service Level Agreement compliance

### 4. PROTOCOL MONITORING
- **ICMP Monitoring** - Ping tests and reachability
- **TCP Monitoring** - TCP port connectivity and timing
- **UDP Monitoring** - UDP performance and packet loss
- **DNS Monitoring** - DNS query performance and resolution
- **HTTP/HTTPS Monitoring** - Web service performance
- **TLS/Certificate Monitoring** - Certificate health and expiration

### 5. NETWORK DIAGNOSTICS
- **Ping Tools** - Interactive ping diagnostics
- **Traceroute** - Path analysis and hop-by-hop latency
- **MTU Discovery** - Path MTU detection and analysis
- **Packet Capture** - Network traffic capture and analysis
- **Network Tools** - 60+ integrated diagnostic tools
- **Tool Management** - Tool discovery and installation

### 6. ROUTING & BGP
- **AS Path Analysis** - Autonomous system path tracking
- **Route Monitoring** - Route change detection
- **BGP Intelligence** - BGP prefix and path analysis
- **Routing History** - Historical route tracking

### 7. INFRASTRUCTURE
- **Devices** - Network device monitoring
- **SNMP Monitoring** - Router/switch monitoring via SNMP
- **Interface Monitoring** - Network interface statistics
- **Topology** - Network topology visualization
- **Data Centers** - Multi-location infrastructure

### 8. ISP & CONNECTIVITY
- **ISP Comparison** - Multi-ISP performance comparison
- **IPv4 Monitoring** - IPv4 connectivity and performance
- **IPv6 Monitoring** - IPv6 connectivity and performance
- **Multi-Location Analysis** - Geographic performance comparison
- **Connectivity Testing** - Protocol-specific connectivity tests

### 9. INCIDENTS & ALERTS
- **Alerts** - Active alerts and notifications
- **Incidents** - Incident management and tracking
- **Alert Rules** - Alert threshold configuration
- **Notification Settings** - Email/Slack/Webhook notifications

### 10. ANALYTICS & REPORTS
- **Historical Analysis** - Time-series data queries
- **Performance Trends** - Trend analysis and baselines
- **Availability Reports** - Uptime and SLA reports
- **Capacity Planning** - Traffic and utilization trends
- **Comparative Analysis** - ISP/probe/location comparisons

### 11. APPLICATION MONITORING
- **Website Monitor** - Website availability and performance
- **API Monitoring** - REST API endpoint monitoring
- **Database Monitoring** - Database connectivity tests
- **CDN Performance** - Content delivery network analysis

### 12. ADMINISTRATION
- **Users** - User management
- **Authentication** - Login and session management
- **Audit Logs** - User action tracking
- **Settings** - System configuration
- **Backup & Restore** - MikroTik backup management

### 13. SYSTEM TOOLS
- **Tool Categories**
  - ICMP/Reachability (ping, fping, hping3, arping, mtr)
  - Traceroute (traceroute, tracepath, mtr, paris-traceroute)
  - DNS (dig, drill, nslookup, dnsviz)
  - HTTP/TLS (curl, wget, httping, openssl)
  - TCP (nc, ncat, telnet)
  - Packet Analysis (tcpdump, tshark, hping3)
  - Bandwidth (iperf3, speedtest-cli, fast-cli)
  - Traffic Monitoring (iftop, nload, bmon, vnstat)
  - BGP/Routing (bgpdump, pybgpstream)

## Detailed Page Requirements

### Priority 1 - Core Monitoring (Existing + Enhancement)
1. ✅ Dashboard/Overview
2. ✅ Targets Management
3. ✅ Monitors Management
4. ✅ Devices
5. ✅ Network Performance/Bandwidth
6. ✅ DNS Servers
7. ✅ Diagnostics
8. ✅ Incidents
9. ✅ Website Monitor
10. ✅ NMS
11. ⚠️ Probes - NEEDS ENHANCEMENT

### Priority 2 - Protocol & Testing (New/Enhancement)
12. 🆕 ICMP Monitoring
13. 🆕 TCP Monitoring
14. 🆕 UDP Monitoring
15. ⚠️ HTTP/HTTPS Monitoring - ENHANCE
16. 🆕 TLS/Certificate Monitoring
17. ⚠️ Traceroute - ENHANCE (already has NetworkPerformance page)
18. 🆕 MTU Discovery
19. 🆕 Packet Capture

### Priority 3 - Network Intelligence (New)
20. 🆕 AS Path Analysis
21. 🆕 Route Monitoring
22. 🆕 BGP Intelligence
23. 🆕 Routing History

### Priority 4 - Infrastructure (New/Enhancement)
24. 🆕 SNMP Monitoring
25. 🆕 Interface Monitoring
26. ⚠️ Topology - ENHANCE
27. 🆕 Multi-Location Dashboard

### Priority 5 - ISP & Connectivity (New)
28. 🆕 ISP Comparison Dashboard
29. 🆕 IPv4 vs IPv6 Analysis
30. 🆕 Connectivity Testing

### Priority 6 - Alerts & Incidents (Enhancement)
31. 🆕 Alerts Dashboard
32. ⚠️ Incidents - ENHANCE
33. 🆕 Alert Rules Management
34. 🆕 Notification Settings

### Priority 7 - Analytics (New)
35. 🆕 Historical Analysis
36. 🆕 Performance Trends
37. 🆕 Availability Reports
38. 🆕 Capacity Planning
39. 🆕 Comparative Analysis

### Priority 8 - Application Monitoring (Enhancement)
40. ⚠️ Website Monitor - ENHANCE
41. 🆕 API Monitoring
42. 🆕 Database Monitoring
43. 🆕 CDN Performance

### Priority 9 - System & Admin (New/Enhancement)
44. 🆕 Users Management
45. 🆕 Audit Logs
46. 🆕 Settings
47. 🆕 Backup Management (MT Backup)
48. 🆕 Probe Management
49. 🆕 Tool Management
50. 🆕 Real-time Dashboard (WebSocket)

## Implementation Strategy

### Phase 1: Enhanced Navigation
- Update Navigation component with all categories
- Implement collapsible menu sections
- Add icons for all menu items
- Add badge counts (alerts, incidents, offline probes)

### Phase 2: Core Pages Enhancement
- Enhance existing pages with missing features
- Add Probes management page
- Add Real-time Dashboard with WebSocket

### Phase 3: New Protocol Monitoring Pages
- ICMP, TCP, UDP dedicated pages
- TLS/Certificate monitoring
- Enhanced Traceroute visualization
- MTU Discovery interface
- Packet Capture interface

### Phase 4: Network Intelligence Pages
- AS Path Analysis
- Route Monitoring
- BGP Intelligence dashboard
- Routing History viewer

### Phase 5: Complete Remaining Features
- ISP Comparison
- Analytics & Reports
- Enhanced Incident Management
- Alert Rules & Notifications
- Tool Management UI
