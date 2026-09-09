import React, { useEffect, useState } from 'react';
import apiService from '../services/api';
import type { DashboardStats } from '../types';
import './Dashboard.css';

export const Dashboard: React.FC = () => {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await apiService.getDashboardStats();
      setStats(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load dashboard stats');
      console.error('Failed to load stats:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadStats();
    const interval = setInterval(loadStats, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, []);

  if (loading && !stats) {
    return (
      <div className="dashboard">
        <div className="loading">Loading dashboard...</div>
      </div>
    );
  }

  if (error && !stats) {
    return (
      <div className="dashboard">
        <div className="error">Error: {error}</div>
      </div>
    );
  }

  const upPercent = stats ? ((stats.targets_up / stats.total_targets) * 100).toFixed(1) : '0';
  const downPercent = stats ? ((stats.targets_down / stats.total_targets) * 100).toFixed(1) : '0';
  const degradedPercent = stats ? ((stats.targets_degraded / stats.total_targets) * 100).toFixed(1) : '0';

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Network Monitoring Dashboard</h1>
        <div className="last-updated">
          Last updated: {new Date().toLocaleTimeString()}
        </div>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-label">Total Targets</div>
          <div className="stat-value">{stats?.total_targets || 0}</div>
        </div>

        <div className="stat-card stat-success">
          <div className="stat-label">Targets UP</div>
          <div className="stat-value">{stats?.targets_up || 0}</div>
          <div className="stat-subtitle">{upPercent}%</div>
        </div>

        <div className="stat-card stat-warning">
          <div className="stat-label">Targets Degraded</div>
          <div className="stat-value">{stats?.targets_degraded || 0}</div>
          <div className="stat-subtitle">{degradedPercent}%</div>
        </div>

        <div className="stat-card stat-danger">
          <div className="stat-label">Targets DOWN</div>
          <div className="stat-value">{stats?.targets_down || 0}</div>
          <div className="stat-subtitle">{downPercent}%</div>
        </div>

        <div className="stat-card">
          <div className="stat-label">Active Probes</div>
          <div className="stat-value">
            {stats?.active_probes || 0} / {stats?.total_probes || 0}
          </div>
        </div>

        <div className="stat-card stat-warning">
          <div className="stat-label">Active Alerts</div>
          <div className="stat-value">{stats?.active_alerts || 0}</div>
        </div>

        <div className="stat-card stat-danger">
          <div className="stat-label">Active Incidents</div>
          <div className="stat-value">{stats?.active_incidents || 0}</div>
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <strong>Warning:</strong> {error}
        </div>
      )}
    </div>
  );
};
