import { useEffect, useState, useCallback } from 'react';
import { useWebSocket } from '../services/websocket';
import { notificationManager } from '../services/notifications';
import { LatencyChart, PacketLossChart, BandwidthChart, UptimeChart } from '../components/Charts';
import apiService from '../services/api';
import type { DashboardStats } from '../types';
import './AdvancedDashboard.css';

export const AdvancedDashboard = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [liveMetrics, setLiveMetrics] = useState({
    latency: [] as Array<{ timestamp: string; value: number }>,
    packetLoss: [] as Array<{ timestamp: string; value: number }>,
    bandwidth: [] as Array<{ timestamp: string; download: number; upload: number }>,
    uptime: [] as Array<{ timestamp: string; uptime: number }>
  });

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await apiService.getDashboardStats();
      setStats(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load dashboard stats');
      notificationManager.error('Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
    const interval = setInterval(loadStats, 30000);
    return () => clearInterval(interval);
  }, []);

  // Handle WebSocket messages for real-time updates
  const handleWebSocketMessage = useCallback((message: any) => {
    const now = new Date().toLocaleTimeString();

    switch (message.type) {
      case 'measurement':
        if (message.data.latency_ms !== undefined) {
          setLiveMetrics(prev => ({
            ...prev,
            latency: [...prev.latency.slice(-49), { timestamp: now, value: message.data.latency_ms }]
          }));
        }
        if (message.data.packet_loss !== undefined) {
          setLiveMetrics(prev => ({
            ...prev,
            packetLoss: [...prev.packetLoss.slice(-49), { timestamp: now, value: message.data.packet_loss }]
          }));
        }
        break;

      case 'alert':
        notificationManager.warning(`Alert: ${message.data.title}`, 8000);
        break;

      case 'incident':
        notificationManager.error(`Incident: ${message.data.title}`, 10000);
        break;

      case 'target_status':
        if (message.data.status === 'down') {
          notificationManager.error(`Target ${message.data.name} is DOWN`);
        } else if (message.data.status === 'up') {
          notificationManager.success(`Target ${message.data.name} is UP`);
        }
        loadStats(); // Refresh stats on target status change
        break;
    }
  }, []);

  const { isConnected } = useWebSocket(handleWebSocketMessage);

  // Generate mock data for charts (will be replaced with real data)
  useEffect(() => {
    // Generate initial mock data
    const generateMockData = () => {
      const times = Array.from({ length: 50 }, (_, i) => {
        const d = new Date();
        d.setMinutes(d.getMinutes() - (49 - i));
        return d.toLocaleTimeString();
      });

      setLiveMetrics({
        latency: times.map(t => ({ timestamp: t, value: Math.random() * 50 + 10 })),
        packetLoss: times.map(t => ({ timestamp: t, value: Math.random() * 5 })),
        bandwidth: times.map(t => ({
          timestamp: t,
          download: Math.random() * 100 + 50,
          upload: Math.random() * 50 + 20
        })),
        uptime: times.slice(-10).map(t => ({ timestamp: t, uptime: 95 + Math.random() * 5 }))
      });
    };

    generateMockData();
  }, []);

  if (loading && !stats) {
    return (
      <div className="advanced-dashboard">
        <div className="loading">Loading dashboard...</div>
      </div>
    );
  }

  const upPercent = stats ? ((stats.targets_up / stats.total_targets) * 100).toFixed(1) : '0';
  const downPercent = stats ? ((stats.targets_down / stats.total_targets) * 100).toFixed(1) : '0';
  const degradedPercent = stats ? ((stats.targets_degraded / stats.total_targets) * 100).toFixed(1) : '0';
  const probePercent = stats ? ((stats.active_probes / stats.total_probes) * 100).toFixed(1) : '0';

  return (
    <div className="advanced-dashboard">
      {/* Header with connection status */}
      <div className="dashboard-header">
        <div>
          <h1>Network Operations Center</h1>
          <div className="dashboard-subtitle">
            Real-time monitoring and analytics
          </div>
        </div>
        <div className="header-status">
          <div className={`connection-indicator ${isConnected ? 'connected' : 'disconnected'}`}>
            <span className="connection-dot"></span>
            {isConnected ? 'Live' : 'Offline'}
          </div>
          <div className="last-updated">
            Updated: {new Date().toLocaleTimeString()}
          </div>
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <strong>Warning:</strong> {error}
        </div>
      )}

      {/* KPI Cards */}
      <div className="kpi-grid">
        <div className="kpi-card kpi-primary">
          <div className="kpi-icon">🎯</div>
          <div className="kpi-content">
            <div className="kpi-label">Total Targets</div>
            <div className="kpi-value">{stats?.total_targets || 0}</div>
            <div className="kpi-change positive">+2 this week</div>
          </div>
        </div>

        <div className="kpi-card kpi-success">
          <div className="kpi-icon">✓</div>
          <div className="kpi-content">
            <div className="kpi-label">Targets UP</div>
            <div className="kpi-value">{stats?.targets_up || 0}</div>
            <div className="kpi-subtitle">{upPercent}% availability</div>
          </div>
        </div>

        <div className="kpi-card kpi-warning">
          <div className="kpi-icon">⚠</div>
          <div className="kpi-content">
            <div className="kpi-label">Degraded</div>
            <div className="kpi-value">{stats?.targets_degraded || 0}</div>
            <div className="kpi-subtitle">{degradedPercent}% affected</div>
          </div>
        </div>

        <div className="kpi-card kpi-danger">
          <div className="kpi-icon">✕</div>
          <div className="kpi-content">
            <div className="kpi-label">Targets DOWN</div>
            <div className="kpi-value">{stats?.targets_down || 0}</div>
            <div className="kpi-subtitle">{downPercent}% offline</div>
          </div>
        </div>

        <div className="kpi-card kpi-info">
          <div className="kpi-icon">📡</div>
          <div className="kpi-content">
            <div className="kpi-label">Active Probes</div>
            <div className="kpi-value">
              {stats?.active_probes || 0} / {stats?.total_probes || 0}
            </div>
            <div className="kpi-subtitle">{probePercent}% online</div>
          </div>
        </div>

        <div className="kpi-card kpi-alert">
          <div className="kpi-icon">🔔</div>
          <div className="kpi-content">
            <div className="kpi-label">Active Alerts</div>
            <div className="kpi-value">{stats?.active_alerts || 0}</div>
            <div className="kpi-change negative">+3 today</div>
          </div>
        </div>

        <div className="kpi-card kpi-incident">
          <div className="kpi-icon">🚨</div>
          <div className="kpi-content">
            <div className="kpi-label">Incidents</div>
            <div className="kpi-value">{stats?.active_incidents || 0}</div>
            <div className="kpi-subtitle">Active issues</div>
          </div>
        </div>

        <div className="kpi-card kpi-metric">
          <div className="kpi-icon">📊</div>
          <div className="kpi-content">
            <div className="kpi-label">Avg Latency</div>
            <div className="kpi-value">24ms</div>
            <div className="kpi-change positive">-5ms vs yesterday</div>
          </div>
        </div>
      </div>

      {/* Real-time Charts */}
      <div className="charts-container">
        <div className="chart-card">
          <div className="chart-header">
            <h3>Network Latency</h3>
            <span className="chart-live-badge">LIVE</span>
          </div>
          <LatencyChart data={liveMetrics.latency} height={250} />
        </div>

        <div className="chart-card">
          <div className="chart-header">
            <h3>Packet Loss</h3>
            <span className="chart-live-badge">LIVE</span>
          </div>
          <PacketLossChart data={liveMetrics.packetLoss} height={250} />
        </div>

        <div className="chart-card chart-wide">
          <div className="chart-header">
            <h3>Bandwidth Usage</h3>
            <span className="chart-live-badge">LIVE</span>
          </div>
          <BandwidthChart data={liveMetrics.bandwidth} height={250} />
        </div>

        <div className="chart-card">
          <div className="chart-header">
            <h3>Uptime (Last 24h)</h3>
          </div>
          <UptimeChart data={liveMetrics.uptime} height={250} />
        </div>
      </div>

      {/* Activity Feed */}
      <div className="activity-section">
        <div className="activity-card">
          <h3>Recent Activity</h3>
          <div className="activity-feed">
            <div className="activity-item">
              <div className="activity-icon success">✓</div>
              <div className="activity-content">
                <div className="activity-title">Target "Production API" recovered</div>
                <div className="activity-time">2 minutes ago</div>
              </div>
            </div>
            <div className="activity-item">
              <div className="activity-icon warning">⚠</div>
              <div className="activity-content">
                <div className="activity-title">High latency detected on "EU Router"</div>
                <div className="activity-time">5 minutes ago</div>
              </div>
            </div>
            <div className="activity-item">
              <div className="activity-icon info">ℹ</div>
              <div className="activity-content">
                <div className="activity-title">New probe "US-West-02" connected</div>
                <div className="activity-time">12 minutes ago</div>
              </div>
            </div>
            <div className="activity-item">
              <div className="activity-icon error">✕</div>
              <div className="activity-content">
                <div className="activity-title">Target "Backup Server" went down</div>
                <div className="activity-time">1 hour ago</div>
              </div>
            </div>
          </div>
        </div>

        <div className="quick-actions-card">
          <h3>Quick Actions</h3>
          <div className="quick-actions">
            <button className="quick-action-btn">
              <span className="qa-icon">+</span>
              Add Target
            </button>
            <button className="quick-action-btn">
              <span className="qa-icon">⚡</span>
              Run Test
            </button>
            <button className="quick-action-btn">
              <span className="qa-icon">📊</span>
              View Reports
            </button>
            <button className="quick-action-btn">
              <span className="qa-icon">⚙</span>
              Settings
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
