import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import apiService from '../services/api';
import type { Target } from '../types';
import './TargetsList.css';

export const TargetsList: React.FC = () => {
  const [targets, setTargets] = useState<Target[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const loadTargets = async () => {
    try {
      setLoading(true);
      const data = await apiService.getTargets();
      setTargets(data);
      setError(null);
    } catch (err: any) {
      setError(err.message || 'Failed to load targets');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadTargets();
    const interval = setInterval(loadTargets, 30000);
    return () => clearInterval(interval);
  }, []);

  if (loading && targets.length === 0) {
    return <div className="targets-list"><div className="loading">Loading targets...</div></div>;
  }

  return (
    <div className="targets-list">
      <div className="list-header">
        <h1>Monitoring Targets</h1>
        <div className="list-actions">
          <button className="btn-primary">+ Add Target</button>
        </div>
      </div>

      {error && <div className="error-banner">Error: {error}</div>}

      {targets.length === 0 ? (
        <div className="empty-state">
          <p>No targets configured yet.</p>
          <button className="btn-primary">Create Your First Target</button>
        </div>
      ) : (
        <div className="targets-grid">
          {targets.map(target => (
            <Link key={target.id} to={`/targets/${target.id}`} className="target-card">
              <div className="target-card-header">
                <h3>{target.name}</h3>
                <span className={`status-indicator ${target.enabled ? 'enabled' : 'disabled'}`}>
                  {target.enabled ? '●' : '○'}
                </span>
              </div>
              <div className="target-card-body">
                <div className="target-address">{target.address}</div>
                {target.port && <div className="target-port">Port: {target.port}</div>}
                <div className="target-type">{target.target_type}</div>
              </div>
              {target.description && (
                <div className="target-card-footer">
                  {target.description}
                </div>
              )}
            </Link>
          ))}
        </div>
      )}
    </div>
  );
};
