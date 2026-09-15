import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { SignIn } from './pages/SignIn';
import { SignUp } from './pages/SignUp';
import { Overview } from './pages/Overview';
import { Devices } from './pages/Devices';
import { Topology } from './pages/Topology';
import { WebsiteMonitor } from './pages/WebsiteMonitor';
import { NMS } from './pages/NMS';
import { Probes } from './pages/Probes';
import { TargetsList } from './pages/TargetsList';
import { TargetDetail } from './pages/TargetDetail';
import { DNSServers } from './pages/DNSServers';
import { NetworkPerformance } from './pages/NetworkPerformance';
import { NetworkTools } from './pages/NetworkTools';
import { Diagnostics } from './pages/Diagnostics';
import { Incidents } from './pages/Incidents';
import { Navigation } from './components/Navigation';
import { NotificationContainer } from './components/NotificationContainer';
import {
  // Network Performance
  LatencyMonitoring,
  ThroughputAnalysis,
  QualityAnalysis,
  SLAMonitoring,
  // Protocol Monitoring
  ICMPMonitoring,
  TCPMonitoring,
  UDPMonitoring,
  HTTPMonitoring,
  TLSCertificates,
  // Diagnostics
  PingTools,
  TraceroutePage,
  MTUDiscovery,
  PacketCapture,
  // Routing & BGP
  ASPathAnalysis,
  RouteMonitoring,
  BGPIntelligence,
  RoutingHistory,
  // Infrastructure
  SNMPMonitoring,
  InterfaceMonitoring,
  DataCenters,
  // ISP & Connectivity
  ISPComparison,
  IPv4Monitoring,
  IPv6Monitoring,
  MultiLocation,
  ConnectivityTesting,
  // Incidents & Alerts
  AlertsPage,
  AlertRules,
  NotificationSettings,
  // Analytics & Reports
  HistoricalAnalysis,
  PerformanceTrends,
  AvailabilityReports,
  CapacityPlanning,
  ComparativeAnalysis,
  // Application Monitoring
  APIMonitoring,
  DatabaseMonitoring,
  CDNPerformance,
  // System
  UsersPage,
  AuditLogs,
  SettingsPage,
  BackupManagement,
  ToolManagement,
  RealtimeDashboard,
} from './pages/PlaceholderPages';
import './App.css';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated (check localStorage for token)
    const token = localStorage.getItem('auth_token');
    const user = localStorage.getItem('user');

    if (token && user) {
      setIsAuthenticated(true);
    }
    setIsLoading(false);
  }, []);

  const handleLogin = (token: string, user: any) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('user', JSON.stringify(user));
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
    setIsAuthenticated(false);
  };

  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="loading-spinner"></div>
        <p>Loading ONCIC...</p>
      </div>
    );
  }

  return (
    <Router>
      <div className="app">
        {!isAuthenticated ? (
          <Routes>
            <Route path="/signin" element={<SignIn onLogin={handleLogin} />} />
            <Route path="/signup" element={<SignUp onSignUp={handleLogin} />} />
            <Route path="*" element={<Navigate to="/signin" replace />} />
          </Routes>
        ) : (
          <>
            <Navigation onLogout={handleLogout} />
            <NotificationContainer />
            <main className="main-content">
              <Routes>
                {/* Overview */}
                <Route path="/" element={<Overview />} />
                <Route path="/realtime" element={<RealtimeDashboard />} />

                {/* Monitoring */}
                <Route path="/targets" element={<TargetsList />} />
                <Route path="/targets/:id" element={<TargetDetail />} />
                <Route path="/monitors" element={<NMS />} />
                <Route path="/probes" element={<Probes />} />
                <Route path="/website-monitor" element={<WebsiteMonitor />} />

                {/* Network Performance */}
                <Route path="/latency" element={<LatencyMonitoring />} />
                <Route path="/network-performance" element={<NetworkPerformance />} />
                <Route path="/throughput" element={<ThroughputAnalysis />} />
                <Route path="/quality" element={<QualityAnalysis />} />
                <Route path="/sla" element={<SLAMonitoring />} />

                {/* Protocol Monitoring */}
                <Route path="/icmp" element={<ICMPMonitoring />} />
                <Route path="/tcp" element={<TCPMonitoring />} />
                <Route path="/udp" element={<UDPMonitoring />} />
                <Route path="/dns-servers" element={<DNSServers />} />
                <Route path="/http-monitoring" element={<HTTPMonitoring />} />
                <Route path="/tls-certificates" element={<TLSCertificates />} />

                {/* Diagnostics */}
                <Route path="/diagnostics" element={<Diagnostics />} />
                <Route path="/ping-tools" element={<PingTools />} />
                <Route path="/traceroute" element={<TraceroutePage />} />
                <Route path="/mtu-discovery" element={<MTUDiscovery />} />
                <Route path="/packet-capture" element={<PacketCapture />} />
                <Route path="/network-tools" element={<NetworkTools />} />

                {/* Routing & BGP */}
                <Route path="/as-path" element={<ASPathAnalysis />} />
                <Route path="/route-monitoring" element={<RouteMonitoring />} />
                <Route path="/bgp-intelligence" element={<BGPIntelligence />} />
                <Route path="/routing-history" element={<RoutingHistory />} />

                {/* Infrastructure */}
                <Route path="/devices" element={<Devices />} />
                <Route path="/snmp" element={<SNMPMonitoring />} />
                <Route path="/interfaces" element={<InterfaceMonitoring />} />
                <Route path="/topology" element={<Topology />} />
                <Route path="/datacenters" element={<DataCenters />} />

                {/* ISP & Connectivity */}
                <Route path="/isp-comparison" element={<ISPComparison />} />
                <Route path="/ipv4-monitoring" element={<IPv4Monitoring />} />
                <Route path="/ipv6-monitoring" element={<IPv6Monitoring />} />
                <Route path="/multi-location" element={<MultiLocation />} />
                <Route path="/connectivity" element={<ConnectivityTesting />} />

                {/* Incidents & Alerts */}
                <Route path="/alerts" element={<AlertsPage />} />
                <Route path="/incidents" element={<Incidents />} />
                <Route path="/alert-rules" element={<AlertRules />} />
                <Route path="/notifications" element={<NotificationSettings />} />

                {/* Analytics & Reports */}
                <Route path="/historical" element={<HistoricalAnalysis />} />
                <Route path="/trends" element={<PerformanceTrends />} />
                <Route path="/availability" element={<AvailabilityReports />} />
                <Route path="/capacity-planning" element={<CapacityPlanning />} />
                <Route path="/comparative" element={<ComparativeAnalysis />} />

                {/* Application Monitoring */}
                <Route path="/api-monitoring" element={<APIMonitoring />} />
                <Route path="/database-monitoring" element={<DatabaseMonitoring />} />
                <Route path="/cdn-performance" element={<CDNPerformance />} />

                {/* System */}
                <Route path="/users" element={<UsersPage />} />
                <Route path="/audit-logs" element={<AuditLogs />} />
                <Route path="/settings" element={<SettingsPage />} />
                <Route path="/backup" element={<BackupManagement />} />
                <Route path="/tool-management" element={<ToolManagement />} />

                {/* Fallback */}
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </main>
          </>
        )}
      </div>
    </Router>
  );
}

export default App;
