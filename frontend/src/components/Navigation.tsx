import { Link, useLocation } from 'react-router-dom';
import { useState } from 'react';
import './Navigation.css';

interface NavigationProps {
  onLogout: () => void;
}

export const Navigation = ({ onLogout }: NavigationProps) => {
  const user = JSON.parse(localStorage.getItem('user') || '{}');
  const location = useLocation();

  // Collapsible sections state
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({
    monitoring: true,
    performance: false,
    protocols: false,
    diagnostics: false,
    routing: false,
    infrastructure: false,
    isp: false,
    incidents: false,
    analytics: false,
    applications: false,
    system: false,
  });

  const isActive = (path: string) => location.pathname === path;

  const toggleSection = (section: string) => {
    setExpandedSections(prev => ({ ...prev, [section]: !prev[section] }));
  };

  return (
    <nav className="noc-sidebar">
      {/* Logo Section */}
      <div className="sidebar-logo">
        <div className="logo-icon">
          <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
            <circle cx="16" cy="16" r="14" stroke="currentColor" strokeWidth="2"/>
            <circle cx="16" cy="16" r="8" fill="currentColor"/>
            <circle cx="16" cy="16" r="4" fill="var(--bg-sidebar)"/>
          </svg>
        </div>
        <div className="logo-text">
          <span className="logo-name">NOC System</span>
          <span className="logo-subtitle">PT. Indonesia Bisnis Digital</span>
        </div>
      </div>

      {/* Navigation Menu */}
      <div className="sidebar-menu">
        {/* OVERVIEW Section */}
        <div className="menu-section">
          <div className="menu-label">OVERVIEW</div>
          <Link to="/" className={`menu-item ${isActive('/') ? 'active' : ''}`}>
            <span className="menu-icon">📊</span>
            <span className="menu-text">Dashboard</span>
          </Link>
          <Link to="/realtime" className={`menu-item ${isActive('/realtime') ? 'active' : ''}`}>
            <span className="menu-icon">📡</span>
            <span className="menu-text">Real-time Monitor</span>
          </Link>
        </div>

        {/* MONITORING Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('monitoring')}
          >
            MONITORING {expandedSections.monitoring ? '▼' : '▶'}
          </div>
          {expandedSections.monitoring && (
            <>
              <Link to="/targets" className={`menu-item ${isActive('/targets') ? 'active' : ''}`}>
                <span className="menu-icon">🎯</span>
                <span className="menu-text">Targets</span>
              </Link>
              <Link to="/monitors" className={`menu-item ${isActive('/monitors') ? 'active' : ''}`}>
                <span className="menu-icon">📈</span>
                <span className="menu-text">Monitors</span>
              </Link>
              <Link to="/probes" className={`menu-item ${isActive('/probes') ? 'active' : ''}`}>
                <span className="menu-icon">🔬</span>
                <span className="menu-text">Probes</span>
              </Link>
              <Link to="/website-monitor" className={`menu-item ${isActive('/website-monitor') ? 'active' : ''}`}>
                <span className="menu-icon">🌐</span>
                <span className="menu-text">Website Monitor</span>
              </Link>
            </>
          )}
        </div>

        {/* NETWORK PERFORMANCE Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('performance')}
          >
            NETWORK PERFORMANCE {expandedSections.performance ? '▼' : '▶'}
          </div>
          {expandedSections.performance && (
            <>
              <Link to="/latency" className={`menu-item ${isActive('/latency') ? 'active' : ''}`}>
                <span className="menu-icon">⚡</span>
                <span className="menu-text">Latency Monitoring</span>
              </Link>
              <Link to="/network-performance" className={`menu-item ${isActive('/network-performance') ? 'active' : ''}`}>
                <span className="menu-icon">🚀</span>
                <span className="menu-text">Bandwidth Testing</span>
              </Link>
              <Link to="/throughput" className={`menu-item ${isActive('/throughput') ? 'active' : ''}`}>
                <span className="menu-icon">📊</span>
                <span className="menu-text">Throughput Analysis</span>
              </Link>
              <Link to="/quality" className={`menu-item ${isActive('/quality') ? 'active' : ''}`}>
                <span className="menu-icon">✨</span>
                <span className="menu-text">Quality Analysis</span>
              </Link>
              <Link to="/sla" className={`menu-item ${isActive('/sla') ? 'active' : ''}`}>
                <span className="menu-icon">📋</span>
                <span className="menu-text">SLA Monitoring</span>
              </Link>
            </>
          )}
        </div>

        {/* PROTOCOL MONITORING Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('protocols')}
          >
            PROTOCOLS {expandedSections.protocols ? '▼' : '▶'}
          </div>
          {expandedSections.protocols && (
            <>
              <Link to="/icmp" className={`menu-item ${isActive('/icmp') ? 'active' : ''}`}>
                <span className="menu-icon">🏓</span>
                <span className="menu-text">ICMP / Ping</span>
              </Link>
              <Link to="/tcp" className={`menu-item ${isActive('/tcp') ? 'active' : ''}`}>
                <span className="menu-icon">🔌</span>
                <span className="menu-text">TCP Monitoring</span>
              </Link>
              <Link to="/udp" className={`menu-item ${isActive('/udp') ? 'active' : ''}`}>
                <span className="menu-icon">📦</span>
                <span className="menu-text">UDP Monitoring</span>
              </Link>
              <Link to="/dns-servers" className={`menu-item ${isActive('/dns-servers') ? 'active' : ''}`}>
                <span className="menu-icon">🌍</span>
                <span className="menu-text">DNS Monitoring</span>
              </Link>
              <Link to="/http-monitoring" className={`menu-item ${isActive('/http-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">🌐</span>
                <span className="menu-text">HTTP/HTTPS</span>
              </Link>
              <Link to="/tls-certificates" className={`menu-item ${isActive('/tls-certificates') ? 'active' : ''}`}>
                <span className="menu-icon">🔒</span>
                <span className="menu-text">TLS/Certificates</span>
              </Link>
            </>
          )}
        </div>

        {/* DIAGNOSTICS Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('diagnostics')}
          >
            DIAGNOSTICS {expandedSections.diagnostics ? '▼' : '▶'}
          </div>
          {expandedSections.diagnostics && (
            <>
              <Link to="/diagnostics" className={`menu-item ${isActive('/diagnostics') ? 'active' : ''}`}>
                <span className="menu-icon">🔧</span>
                <span className="menu-text">Diagnostics Hub</span>
              </Link>
              <Link to="/ping-tools" className={`menu-item ${isActive('/ping-tools') ? 'active' : ''}`}>
                <span className="menu-icon">🏓</span>
                <span className="menu-text">Ping Tools</span>
              </Link>
              <Link to="/traceroute" className={`menu-item ${isActive('/traceroute') ? 'active' : ''}`}>
                <span className="menu-icon">🗺️</span>
                <span className="menu-text">Traceroute</span>
              </Link>
              <Link to="/mtu-discovery" className={`menu-item ${isActive('/mtu-discovery') ? 'active' : ''}`}>
                <span className="menu-icon">📏</span>
                <span className="menu-text">MTU Discovery</span>
              </Link>
              <Link to="/packet-capture" className={`menu-item ${isActive('/packet-capture') ? 'active' : ''}`}>
                <span className="menu-icon">📸</span>
                <span className="menu-text">Packet Capture</span>
              </Link>
              <Link to="/network-tools" className={`menu-item ${isActive('/network-tools') ? 'active' : ''}`}>
                <span className="menu-icon">🛠️</span>
                <span className="menu-text">Network Tools</span>
              </Link>
            </>
          )}
        </div>

        {/* ROUTING & BGP Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('routing')}
          >
            ROUTING & BGP {expandedSections.routing ? '▼' : '▶'}
          </div>
          {expandedSections.routing && (
            <>
              <Link to="/as-path" className={`menu-item ${isActive('/as-path') ? 'active' : ''}`}>
                <span className="menu-icon">🌐</span>
                <span className="menu-text">AS Path Analysis</span>
              </Link>
              <Link to="/route-monitoring" className={`menu-item ${isActive('/route-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">🛤️</span>
                <span className="menu-text">Route Monitoring</span>
              </Link>
              <Link to="/bgp-intelligence" className={`menu-item ${isActive('/bgp-intelligence') ? 'active' : ''}`}>
                <span className="menu-icon">🧠</span>
                <span className="menu-text">BGP Intelligence</span>
              </Link>
              <Link to="/routing-history" className={`menu-item ${isActive('/routing-history') ? 'active' : ''}`}>
                <span className="menu-icon">📚</span>
                <span className="menu-text">Routing History</span>
              </Link>
            </>
          )}
        </div>

        {/* INFRASTRUCTURE Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('infrastructure')}
          >
            INFRASTRUCTURE {expandedSections.infrastructure ? '▼' : '▶'}
          </div>
          {expandedSections.infrastructure && (
            <>
              <Link to="/devices" className={`menu-item ${isActive('/devices') ? 'active' : ''}`}>
                <span className="menu-icon">🖥️</span>
                <span className="menu-text">Devices</span>
              </Link>
              <Link to="/snmp" className={`menu-item ${isActive('/snmp') ? 'active' : ''}`}>
                <span className="menu-icon">📡</span>
                <span className="menu-text">SNMP Monitoring</span>
              </Link>
              <Link to="/interfaces" className={`menu-item ${isActive('/interfaces') ? 'active' : ''}`}>
                <span className="menu-icon">🔌</span>
                <span className="menu-text">Interface Monitoring</span>
              </Link>
              <Link to="/topology" className={`menu-item ${isActive('/topology') ? 'active' : ''}`}>
                <span className="menu-icon">🔗</span>
                <span className="menu-text">Topology</span>
              </Link>
              <Link to="/datacenters" className={`menu-item ${isActive('/datacenters') ? 'active' : ''}`}>
                <span className="menu-icon">🏢</span>
                <span className="menu-text">Data Centers</span>
              </Link>
            </>
          )}
        </div>

        {/* ISP & CONNECTIVITY Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('isp')}
          >
            ISP & CONNECTIVITY {expandedSections.isp ? '▼' : '▶'}
          </div>
          {expandedSections.isp && (
            <>
              <Link to="/isp-comparison" className={`menu-item ${isActive('/isp-comparison') ? 'active' : ''}`}>
                <span className="menu-icon">⚖️</span>
                <span className="menu-text">ISP Comparison</span>
              </Link>
              <Link to="/ipv4-monitoring" className={`menu-item ${isActive('/ipv4-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">4️⃣</span>
                <span className="menu-text">IPv4 Monitoring</span>
              </Link>
              <Link to="/ipv6-monitoring" className={`menu-item ${isActive('/ipv6-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">6️⃣</span>
                <span className="menu-text">IPv6 Monitoring</span>
              </Link>
              <Link to="/multi-location" className={`menu-item ${isActive('/multi-location') ? 'active' : ''}`}>
                <span className="menu-icon">🌍</span>
                <span className="menu-text">Multi-Location</span>
              </Link>
              <Link to="/connectivity" className={`menu-item ${isActive('/connectivity') ? 'active' : ''}`}>
                <span className="menu-icon">🔗</span>
                <span className="menu-text">Connectivity Tests</span>
              </Link>
            </>
          )}
        </div>

        {/* INCIDENTS & ALERTS Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('incidents')}
          >
            INCIDENTS & ALERTS {expandedSections.incidents ? '▼' : '▶'}
          </div>
          {expandedSections.incidents && (
            <>
              <Link to="/alerts" className={`menu-item ${isActive('/alerts') ? 'active' : ''}`}>
                <span className="menu-icon">🔔</span>
                <span className="menu-text">Alerts</span>
              </Link>
              <Link to="/incidents" className={`menu-item ${isActive('/incidents') ? 'active' : ''}`}>
                <span className="menu-icon">⚠️</span>
                <span className="menu-text">Incidents</span>
              </Link>
              <Link to="/alert-rules" className={`menu-item ${isActive('/alert-rules') ? 'active' : ''}`}>
                <span className="menu-icon">📏</span>
                <span className="menu-text">Alert Rules</span>
              </Link>
              <Link to="/notifications" className={`menu-item ${isActive('/notifications') ? 'active' : ''}`}>
                <span className="menu-icon">📧</span>
                <span className="menu-text">Notifications</span>
              </Link>
            </>
          )}
        </div>

        {/* ANALYTICS & REPORTS Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('analytics')}
          >
            ANALYTICS {expandedSections.analytics ? '▼' : '▶'}
          </div>
          {expandedSections.analytics && (
            <>
              <Link to="/historical" className={`menu-item ${isActive('/historical') ? 'active' : ''}`}>
                <span className="menu-icon">📊</span>
                <span className="menu-text">Historical Analysis</span>
              </Link>
              <Link to="/trends" className={`menu-item ${isActive('/trends') ? 'active' : ''}`}>
                <span className="menu-icon">📈</span>
                <span className="menu-text">Performance Trends</span>
              </Link>
              <Link to="/availability" className={`menu-item ${isActive('/availability') ? 'active' : ''}`}>
                <span className="menu-icon">✅</span>
                <span className="menu-text">Availability Reports</span>
              </Link>
              <Link to="/capacity-planning" className={`menu-item ${isActive('/capacity-planning') ? 'active' : ''}`}>
                <span className="menu-icon">📊</span>
                <span className="menu-text">Capacity Planning</span>
              </Link>
              <Link to="/comparative" className={`menu-item ${isActive('/comparative') ? 'active' : ''}`}>
                <span className="menu-icon">⚖️</span>
                <span className="menu-text">Comparative Analysis</span>
              </Link>
            </>
          )}
        </div>

        {/* APPLICATION MONITORING Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('applications')}
          >
            APPLICATIONS {expandedSections.applications ? '▼' : '▶'}
          </div>
          {expandedSections.applications && (
            <>
              <Link to="/api-monitoring" className={`menu-item ${isActive('/api-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">🔌</span>
                <span className="menu-text">API Monitoring</span>
              </Link>
              <Link to="/database-monitoring" className={`menu-item ${isActive('/database-monitoring') ? 'active' : ''}`}>
                <span className="menu-icon">🗄️</span>
                <span className="menu-text">Database Monitoring</span>
              </Link>
              <Link to="/cdn-performance" className={`menu-item ${isActive('/cdn-performance') ? 'active' : ''}`}>
                <span className="menu-icon">🌐</span>
                <span className="menu-text">CDN Performance</span>
              </Link>
            </>
          )}
        </div>

        {/* SYSTEM Section */}
        <div className="menu-section">
          <div
            className="menu-label collapsible"
            onClick={() => toggleSection('system')}
          >
            SYSTEM {expandedSections.system ? '▼' : '▶'}
          </div>
          {expandedSections.system && (
            <>
              <Link to="/users" className={`menu-item ${isActive('/users') ? 'active' : ''}`}>
                <span className="menu-icon">👥</span>
                <span className="menu-text">Users</span>
              </Link>
              <Link to="/audit-logs" className={`menu-item ${isActive('/audit-logs') ? 'active' : ''}`}>
                <span className="menu-icon">📝</span>
                <span className="menu-text">Audit Logs</span>
              </Link>
              <Link to="/settings" className={`menu-item ${isActive('/settings') ? 'active' : ''}`}>
                <span className="menu-icon">⚙️</span>
                <span className="menu-text">Settings</span>
              </Link>
              <Link to="/backup" className={`menu-item ${isActive('/backup') ? 'active' : ''}`}>
                <span className="menu-icon">💾</span>
                <span className="menu-text">MT Backup</span>
              </Link>
              <Link to="/tool-management" className={`menu-item ${isActive('/tool-management') ? 'active' : ''}`}>
                <span className="menu-icon">🛠️</span>
                <span className="menu-text">Tool Management</span>
              </Link>
            </>
          )}
        </div>
      </div>

      {/* User Profile Section */}
      <div className="sidebar-user">
        <div className="user-avatar">
          <span className="avatar-initials">
            {(user.username || 'DI').substring(0, 2).toUpperCase()}
          </span>
        </div>
        <div className="user-details">
          <div className="user-name">{user.username || 'Dion IPe'}</div>
          <div className="user-role">Admin</div>
        </div>
        <button className="user-logout" onClick={onLogout} title="Logout">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
            <path d="M6 14H3C2.44772 14 2 13.5523 2 13V3C2 2.44772 2.44772 2 3 2H6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
            <path d="M11 11L14 8L11 5" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round"/>
            <path d="M14 8H6" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round"/>
          </svg>
        </button>
      </div>
    </nav>
  );
};
