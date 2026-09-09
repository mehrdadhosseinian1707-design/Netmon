import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import apiService from '../services/api';
import { LatencyChart } from '../components/LatencyChart';
import { PacketLossChart } from '../components/PacketLossChart';
import type { Target, TargetStats, Measurement } from '../types';
import './TargetDetail.css';

export const TargetDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const [target, setTarget] = useState<Target | null>(null);
  const [stats, setStats] = useState<TargetStats | null>(null);
  const [measurements, setMeasurements] = useState<Measurement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    if (!id) return;

    try {
      setLoading(true);
      const [targetData, statsData, measurementsData] = await Promise.all([
        apiService.getTarget(id),
        apiService.getTargetStats(id),
        apiService.getTargetMeasurements(id),
      ]);

      setTarget(targetData);
      setStats(statsData);
      setMeasurements(measurementsData);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load target data');
      console.error('Failed to load target:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
    const interval = setInterval(loadData, 30000);
    return () => clearInterval(interval);
  }, [id]);

  if (loading && !target) {
    return <div className="target-detail"><div className="loading">Loading...</div></div>;
  }

  if (error && !target) {
    return <div className="target-detail"><div className="error">Error: {error}</div></div>;
  }

  if (!target) {
    return <div className="target-detail"><div className="error">Target not found</div></div>;
  }

  const availability = stats?.availability ?? 0;
  const statusClass = availability >= 99 ? 'status-up' : availability >= 50 ? 'status-degraded' : 'status-down';

  return (
    <div className="target-detail">
      <div className="breadcrumb">
        <Link to="/targets">← Back to Targets</Link>
      </div>

      <div className="target-header">
        <div>
          <h1>{target.name}</h1>
          <div className="target-meta">
            <span className="target-address">{target.address}</span>
            {target.port && <span className="target-port">:{target.port}</span>}
            <span className={`target-status ${statusClass}`}>
              {availability >= 99 ? '● UP' : availability >= 50 ? '● DEGRADED' : '● DOWN'}
            </span>
          </div>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-box">
          <div className="stat-label">Availability</div>
          <div className="stat-value">{availability.toFixed(2)}%</div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Avg Latency</div>
          <div className="stat-value">
            {stats?.avg_latency_ms ? `${stats.avg_latency_ms.toFixed(2)} ms` : 'N/A'}
          </div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Min Latency</div>
          <div className="stat-value">
            {stats?.min_latency_ms ? `${stats.min_latency_ms.toFixed(2)} ms` : 'N/A'}
          </div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Max Latency</div>
          <div className="stat-value">
            {stats?.max_latency_ms ? `${stats.max_latency_ms.toFixed(2)} ms` : 'N/A'}
          </div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Packet Loss</div>
          <div className="stat-value">
            {stats?.avg_packet_loss ? `${stats.avg_packet_loss.toFixed(2)}%` : 'N/A'}
          </div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Total Checks</div>
          <div className="stat-value">{stats?.total_checks || 0}</div>
        </div>
      </div>

      <div className="charts-section">
        <div className="chart-container">
          <h3>Latency (Last 100 Checks)</h3>
          <LatencyChart measurements={measurements} />
        </div>

        <div className="chart-container">
          <h3>Packet Loss (Last 100 Checks)</h3>
          <PacketLossChart measurements={measurements} />
        </div>
      </div>

      <div className="recent-checks">
        <h3>Recent Checks</h3>
        <div className="checks-table">
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Status</th>
                <th>Latency</th>
                <th>Packet Loss</th>
                <th>Jitter</th>
                <th>Error</th>
              </tr>
            </thead>
            <tbody>
              {measurements.slice(0, 20).map((m, idx) => (
                <tr key={idx} className={m.success ? 'check-success' : 'check-failure'}>
                  <td>{new Date(m.time).toLocaleString()}</td>
                  <td>
                    <span className={`status-badge ${m.success ? 'success' : 'failure'}`}>
                      {m.success ? 'SUCCESS' : 'FAILURE'}
                    </span>
                  </td>
                  <td>{m.latency_ms ? `${m.latency_ms.toFixed(2)} ms` : '-'}</td>
                  <td>{m.packet_loss !== undefined ? `${m.packet_loss.toFixed(2)}%` : '-'}</td>
                  <td>{m.jitter_ms ? `${m.jitter_ms.toFixed(2)} ms` : '-'}</td>
                  <td className="error-cell">{m.error_message || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
