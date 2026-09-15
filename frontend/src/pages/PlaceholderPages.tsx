import { PlaceholderPage } from './PlaceholderPage';

// Network Performance Pages
export const LatencyMonitoring = () => (
  <PlaceholderPage
    title="Latency Monitoring"
    subtitle="Real-time latency, jitter, and packet loss monitoring"
    icon="⚡"
    description="Monitor network latency with RTT measurements, jitter analysis, and packet loss detection across all your targets."
  />
);

export const ThroughputAnalysis = () => (
  <PlaceholderPage
    title="Throughput Analysis"
    subtitle="TCP and UDP throughput measurement and analysis"
    icon="📊"
    description="Analyze network throughput performance with comprehensive TCP and UDP testing capabilities."
  />
);

export const QualityAnalysis = () => (
  <PlaceholderPage
    title="Quality Analysis"
    subtitle="Network quality metrics and performance analysis"
    icon="✨"
    description="Comprehensive network quality analysis including latency, jitter, packet loss, and overall network health."
  />
);

export const SLAMonitoring = () => (
  <PlaceholderPage
    title="SLA Monitoring"
    subtitle="Service Level Agreement compliance tracking"
    icon="📋"
    description="Track and report on SLA compliance with automated threshold monitoring and violation alerts."
  />
);

// Protocol Monitoring Pages
export const ICMPMonitoring = () => (
  <PlaceholderPage
    title="ICMP Monitoring"
    subtitle="Ping tests and ICMP reachability monitoring"
    icon="🏓"
    description="Execute ping tests with configurable packet size, count, and interval. Monitor ICMP reachability across your network."
  />
);

export const TCPMonitoring = () => (
  <PlaceholderPage
    title="TCP Monitoring"
    subtitle="TCP port connectivity and performance testing"
    icon="🔌"
    description="Monitor TCP service availability, connection timing, and throughput across all your TCP services."
  />
);

export const UDPMonitoring = () => (
  <PlaceholderPage
    title="UDP Monitoring"
    subtitle="UDP connectivity and packet loss testing"
    icon="📦"
    description="Test UDP connectivity, measure packet loss, and monitor UDP-based services like VoIP and video streaming."
  />
);

export const HTTPMonitoring = () => (
  <PlaceholderPage
    title="HTTP/HTTPS Monitoring"
    subtitle="Web service performance and availability monitoring"
    icon="🌐"
    description="Monitor HTTP/HTTPS endpoints with detailed timing breakdown including DNS, TCP, TLS handshake, and TTFB."
  />
);

export const TLSCertificates = () => (
  <PlaceholderPage
    title="TLS/Certificate Monitoring"
    subtitle="SSL/TLS certificate monitoring and expiration tracking"
    icon="🔒"
    description="Monitor TLS certificates, track expiration dates, validate certificate chains, and ensure secure connections."
  />
);

// Diagnostics Pages
export const PingTools = () => (
  <PlaceholderPage
    title="Ping Tools"
    subtitle="Interactive ping diagnostic tools"
    icon="🏓"
    description="Run interactive ping tests with real-time results, packet size options, and detailed statistics."
  />
);

export const TraceroutePage = () => (
  <PlaceholderPage
    title="Traceroute"
    subtitle="Network path tracing and hop-by-hop analysis"
    icon="🗺️"
    description="Trace network routes with detailed hop information, latency measurements, and path visualization."
  />
);

export const MTUDiscovery = () => (
  <PlaceholderPage
    title="MTU Discovery"
    subtitle="Path MTU discovery and fragmentation analysis"
    icon="📏"
    description="Discover maximum transmission unit (MTU) along network paths and identify fragmentation issues."
  />
);

export const PacketCapture = () => (
  <PlaceholderPage
    title="Packet Capture"
    subtitle="Network traffic capture and analysis"
    icon="📸"
    description="Capture and analyze network packets using tcpdump, tshark, and other packet analysis tools."
  />
);

// Routing & BGP Pages
export const ASPathAnalysis = () => (
  <PlaceholderPage
    title="AS Path Analysis"
    subtitle="Autonomous system path tracking and analysis"
    icon="🌐"
    description="Track and analyze BGP AS paths, monitor route changes, and understand network routing decisions."
  />
);

export const RouteMonitoring = () => (
  <PlaceholderPage
    title="Route Monitoring"
    subtitle="Network route tracking and change detection"
    icon="🛤️"
    description="Monitor network routes in real-time, detect route changes, and analyze routing stability."
  />
);

export const BGPIntelligence = () => (
  <PlaceholderPage
    title="BGP Intelligence"
    subtitle="BGP prefix monitoring and route analysis"
    icon="🧠"
    description="Advanced BGP intelligence with prefix monitoring, route leak detection, and hijack prevention."
  />
);

export const RoutingHistory = () => (
  <PlaceholderPage
    title="Routing History"
    subtitle="Historical route tracking and analysis"
    icon="📚"
    description="Review historical routing data, analyze route changes over time, and identify patterns."
  />
);

// Infrastructure Pages
export const SNMPMonitoring = () => (
  <PlaceholderPage
    title="SNMP Monitoring"
    subtitle="SNMP-based device monitoring and statistics"
    icon="📡"
    description="Monitor network devices via SNMP, collect interface statistics, and track device health."
  />
);

export const InterfaceMonitoring = () => (
  <PlaceholderPage
    title="Interface Monitoring"
    subtitle="Network interface statistics and performance"
    icon="🔌"
    description="Monitor network interface statistics including RX/TX bytes, packets, errors, and drops."
  />
);

export const DataCenters = () => (
  <PlaceholderPage
    title="Data Centers"
    subtitle="Multi-location data center monitoring"
    icon="🏢"
    description="Monitor infrastructure across multiple data centers and geographic locations."
  />
);

// ISP & Connectivity Pages
export const ISPComparison = () => (
  <PlaceholderPage
    title="ISP Comparison"
    subtitle="Multi-ISP performance comparison and analysis"
    icon="⚖️"
    description="Compare performance across multiple ISPs including latency, packet loss, bandwidth, and routing."
  />
);

export const IPv4Monitoring = () => (
  <PlaceholderPage
    title="IPv4 Monitoring"
    subtitle="IPv4 connectivity and performance monitoring"
    icon="4️⃣"
    description="Monitor IPv4 connectivity, performance, and availability across your network."
  />
);

export const IPv6Monitoring = () => (
  <PlaceholderPage
    title="IPv6 Monitoring"
    subtitle="IPv6 connectivity and performance monitoring"
    icon="6️⃣"
    description="Monitor IPv6 connectivity, compare with IPv4 performance, and track dual-stack availability."
  />
);

export const MultiLocation = () => (
  <PlaceholderPage
    title="Multi-Location Analysis"
    subtitle="Geographic performance comparison and analysis"
    icon="🌍"
    description="Compare network performance across multiple geographic locations and data centers."
  />
);

export const ConnectivityTesting = () => (
  <PlaceholderPage
    title="Connectivity Testing"
    subtitle="Protocol-specific connectivity tests"
    icon="🔗"
    description="Test network connectivity using various protocols including ICMP, TCP, UDP, and HTTP."
  />
);

// Incidents & Alerts Pages
export const AlertsPage = () => (
  <PlaceholderPage
    title="Alerts"
    subtitle="Active alerts and notification center"
    icon="🔔"
    description="View and manage active alerts with severity levels, acknowledgment, and resolution tracking."
  />
);

export const AlertRules = () => (
  <PlaceholderPage
    title="Alert Rules"
    subtitle="Alert threshold configuration and rules"
    icon="📏"
    description="Configure alert thresholds, define rules, and set up automated notifications."
  />
);

export const NotificationSettings = () => (
  <PlaceholderPage
    title="Notification Settings"
    subtitle="Email, Slack, and webhook notifications"
    icon="📧"
    description="Configure notification channels including email, Slack, webhooks, and SMS."
  />
);

// Analytics & Reports Pages
export const HistoricalAnalysis = () => (
  <PlaceholderPage
    title="Historical Analysis"
    subtitle="Time-series data analysis and queries"
    icon="📊"
    description="Analyze historical monitoring data with custom time ranges and detailed metrics."
  />
);

export const PerformanceTrends = () => (
  <PlaceholderPage
    title="Performance Trends"
    subtitle="Trend analysis and performance baselines"
    icon="📈"
    description="Identify performance trends, establish baselines, and detect anomalies over time."
  />
);

export const AvailabilityReports = () => (
  <PlaceholderPage
    title="Availability Reports"
    subtitle="Uptime tracking and SLA reports"
    icon="✅"
    description="Generate availability reports, track uptime percentages, and monitor SLA compliance."
  />
);

export const CapacityPlanning = () => (
  <PlaceholderPage
    title="Capacity Planning"
    subtitle="Traffic analysis and capacity forecasting"
    icon="📊"
    description="Analyze traffic patterns, forecast capacity needs, and plan infrastructure growth."
  />
);

export const ComparativeAnalysis = () => (
  <PlaceholderPage
    title="Comparative Analysis"
    subtitle="Multi-dimensional performance comparison"
    icon="⚖️"
    description="Compare performance across ISPs, probes, locations, protocols, and time periods."
  />
);

// Application Monitoring Pages
export const APIMonitoring = () => (
  <PlaceholderPage
    title="API Monitoring"
    subtitle="REST API endpoint monitoring and testing"
    icon="🔌"
    description="Monitor API endpoints, track response times, validate status codes, and test API availability."
  />
);

export const DatabaseMonitoring = () => (
  <PlaceholderPage
    title="Database Monitoring"
    subtitle="Database connectivity and performance testing"
    icon="🗄️"
    description="Monitor database connectivity, test query performance, and track database availability."
  />
);

export const CDNPerformance = () => (
  <PlaceholderPage
    title="CDN Performance"
    subtitle="Content delivery network performance analysis"
    icon="🌐"
    description="Analyze CDN performance across multiple locations and compare delivery speeds."
  />
);

// System Pages
export const UsersPage = () => (
  <PlaceholderPage
    title="Users"
    subtitle="User management and access control"
    icon="👥"
    description="Manage users, roles, permissions, and access control for the monitoring platform."
  />
);

export const AuditLogs = () => (
  <PlaceholderPage
    title="Audit Logs"
    subtitle="User activity tracking and audit trails"
    icon="📝"
    description="Track user actions, configuration changes, and maintain security audit trails."
  />
);

export const SettingsPage = () => (
  <PlaceholderPage
    title="Settings"
    subtitle="System configuration and preferences"
    icon="⚙️"
    description="Configure system settings, preferences, and global monitoring parameters."
  />
);

export const BackupManagement = () => (
  <PlaceholderPage
    title="MikroTik Backup"
    subtitle="MikroTik device backup management"
    icon="💾"
    description="Automated backup and restore for MikroTik devices with scheduling and version control."
  />
);

export const ToolManagement = () => (
  <PlaceholderPage
    title="Tool Management"
    subtitle="Network diagnostic tool discovery and management"
    icon="🛠️"
    description="Discover, install, and manage 60+ network diagnostic tools across your monitoring probes."
  />
);

// Real-time Dashboard
export const RealtimeDashboard = () => (
  <PlaceholderPage
    title="Real-time Monitor"
    subtitle="Live monitoring data stream via WebSocket"
    icon="📡"
    description="Watch live monitoring data with real-time updates, instant alerts, and streaming metrics."
  />
);
