import { useState } from 'react';
import './WebsiteMonitor.css';

interface Website {
  id: string;
  name: string;
  url: string;
  status: 'up' | 'down' | 'degraded';
  uptime: number;
  responseTime: number;
  lastCheck: string;
  ssl: boolean;
  sslExpiry?: string;
}

export const WebsiteMonitor = () => {
  const [websites] = useState<Website[]>([
    {
      id: '1',
      name: 'Company Website',
      url: 'https://example.com',
      status: 'up',
      uptime: 99.9,
      responseTime: 245,
      lastCheck: '30 seconds ago',
      ssl: true,
      sslExpiry: '90 days'
    },
    {
      id: '2',
      name: 'API Gateway',
      url: 'https://api.example.com',
      status: 'up',
      uptime: 99.95,
      responseTime: 180,
      lastCheck: '1 minute ago',
      ssl: true,
      sslExpiry: '120 days'
    },
    {
      id: '3',
      name: 'Customer Portal',
      url: 'https://portal.example.com',
      status: 'degraded',
      uptime: 98.5,
      responseTime: 1200,
      lastCheck: '45 seconds ago',
      ssl: true,
      sslExpiry: '45 days'
    }
  ]);

  const stats = {
    total: 12,
    up: 10,
    down: 0,
    degraded: 2,
    avgUptime: 99.2,
    avgResponseTime: 320
  };

  return (
    <div className="website-monitor-page">
      {/* Page Header */}
      <div className="page-header">
        <div>
          <div className="page-breadcrumb">MONITORING</div>
          <h1 className="page-title">Website Monitor</h1>
          <p className="page-subtitle">Monitor website uptime, performance, and SSL certificates.</p>
        </div>
        <div className="header-actions">
          <button className="btn-primary">
            <span>➕</span> Add Website
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">🌐</div>
          <div className="stat-content">
            <div className="stat-label">TOTAL WEBSITES</div>
            <div className="stat-value">{stats.total}</div>
          </div>
        </div>

        <div className="stat-card stat-success">
          <div className="stat-icon">✅</div>
          <div className="stat-content">
            <div className="stat-label">UP</div>
            <div className="stat-value">{stats.up}</div>
          </div>
        </div>

        <div className="stat-card stat-danger">
          <div className="stat-icon">❌</div>
          <div className="stat-content">
            <div className="stat-label">DOWN</div>
            <div className="stat-value">{stats.down}</div>
          </div>
        </div>

        <div className="stat-card stat-warning">
          <div className="stat-icon">⚠️</div>
          <div className="stat-content">
            <div className="stat-label">DEGRADED</div>
            <div className="stat-value">{stats.degraded}</div>
          </div>
        </div>

        <div className="stat-card stat-info">
          <div className="stat-icon">📊</div>
          <div className="stat-content">
            <div className="stat-label">AVG UPTIME</div>
            <div className="stat-value">{stats.avgUptime}%</div>
          </div>
        </div>

        <div className="stat-card stat-info">
          <div className="stat-icon">⚡</div>
          <div className="stat-content">
            <div className="stat-label">AVG RESPONSE</div>
            <div className="stat-value">{stats.avgResponseTime}ms</div>
          </div>
        </div>
      </div>

      {/* Websites Grid */}
      <div className="websites-grid">
        {websites.map((website) => (
          <div key={website.id} className="website-card noc-card">
            <div className="website-header">
              <div className="website-info">
                <h3 className="website-name">{website.name}</h3>
                <a href={website.url} target="_blank" rel="noopener noreferrer" className="website-url">
                  {website.url}
                </a>
              </div>
              <span className={`status-badge ${website.status}`}>
                {website.status === 'up' && '🟢'}
                {website.status === 'down' && '🔴'}
                {website.status === 'degraded' && '🟡'}
                <span>{website.status.toUpperCase()}</span>
              </span>
            </div>

            <div className="website-metrics">
              <div className="metric-item">
                <div className="metric-label">Uptime</div>
                <div className="metric-value success">{website.uptime}%</div>
              </div>
              <div className="metric-item">
                <div className="metric-label">Response Time</div>
                <div className="metric-value">{website.responseTime}ms</div>
              </div>
              <div className="metric-item">
                <div className="metric-label">SSL Status</div>
                <div className="metric-value">
                  {website.ssl ? '🔒 Valid' : '⚠️ No SSL'}
                </div>
              </div>
            </div>

            {website.ssl && website.sslExpiry && (
              <div className="ssl-info">
                <span className="ssl-label">SSL expires in:</span>
                <span className="ssl-expiry">{website.sslExpiry}</span>
              </div>
            )}

            <div className="website-footer">
              <span className="last-check">Last checked: {website.lastCheck}</span>
              <div className="website-actions">
                <button className="action-btn">📊 View Stats</button>
                <button className="action-btn">⚙️ Settings</button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
