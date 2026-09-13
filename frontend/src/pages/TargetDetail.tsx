import { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import apiService from '../services/api';
import type { Target } from '../types';
import './TargetDetail.css';

export const TargetDetail = () => {
  const { id } = useParams<{ id: string }>();
  const [target, setTarget] = useState<Target | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadData = async () => {
    if (!id) return;

    try {
      setLoading(true);
      const targetData = await apiService.getTarget(id);
      setTarget(targetData);
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
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

  const getStatusClass = () => {
    if (target.status === 'up') return 'status-up';
    if (target.status === 'degraded') return 'status-degraded';
    if (target.status === 'down') return 'status-down';
    return 'status-down';
  };

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
            {target.port && <span className="target-port">Port: {target.port}</span>}
            <span className={`target-status ${getStatusClass()}`}>
              {target.status.toUpperCase()}
            </span>
          </div>
        </div>
      </div>

      <div className="stats-row">
        <div className="stat-box">
          <div className="stat-label">Status</div>
          <div className="stat-value">{target.status.toUpperCase()}</div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Type</div>
          <div className="stat-value">{target.type}</div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Enabled</div>
          <div className="stat-value">{target.enabled ? 'Yes' : 'No'}</div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Last Check</div>
          <div className="stat-value">
            {target.last_check ? new Date(target.last_check).toLocaleTimeString() : 'Never'}
          </div>
        </div>
        <div className="stat-box">
          <div className="stat-label">Created</div>
          <div className="stat-value">
            {new Date(target.created_at).toLocaleDateString()}
          </div>
        </div>
      </div>

      <div className="charts-section">
        <div className="chart-container">
          <h3>Latency Chart</h3>
          <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
            Chart visualization coming soon...
          </p>
        </div>

        <div className="chart-container">
          <h3>Packet Loss Chart</h3>
          <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
            Chart visualization coming soon...
          </p>
        </div>
      </div>

      <div className="recent-checks">
        <h3>Recent Checks</h3>
        <p style={{ color: 'var(--text-muted)', textAlign: 'center', padding: '2rem' }}>
          Recent checks data coming soon...
        </p>
      </div>
    </div>
  );
};
