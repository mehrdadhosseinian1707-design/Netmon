import { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import apiService from '../services/api';
import type { Target } from '../types';
import './TargetsList.css';

export const TargetsList = () => {
  const [targets, setTargets] = useState<Target[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showModal, setShowModal] = useState(false);
  const [formData, setFormData] = useState({
    name: '',
    type: 'host',
    address: '',
    port: '',
    description: ''
  });
  const [submitting, setSubmitting] = useState(false);

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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      const targetData: any = {
        name: formData.name,
        type: formData.type,
        address: formData.address,
        description: formData.description
      };

      if (formData.port) {
        targetData.port = parseInt(formData.port);
      }

      await apiService.createTarget(targetData);
      setShowModal(false);
      setFormData({ name: '', type: 'host', address: '', port: '', description: '' });
      loadTargets();
    } catch (err: any) {
      alert('Failed to create target: ' + (err.message || 'Unknown error'));
    } finally {
      setSubmitting(false);
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
          <button className="btn-primary" onClick={() => setShowModal(true)}>+ Add Target</button>
        </div>
      </div>

      {error && <div className="error-banner">Error: {error}</div>}

      {targets.length === 0 ? (
        <div className="empty-state">
          <p>No targets configured yet.</p>
          <button className="btn-primary" onClick={() => setShowModal(true)}>Create Your First Target</button>
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
                <div className="target-type">{target.type}</div>
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

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h2>Add New Target</h2>
              <button className="modal-close" onClick={() => setShowModal(false)}>×</button>
            </div>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Name *</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({...formData, name: e.target.value})}
                  placeholder="e.g., Production Server"
                  required
                />
              </div>

              <div className="form-group">
                <label>Type *</label>
                <select
                  value={formData.type}
                  onChange={(e) => setFormData({...formData, type: e.target.value})}
                  required
                >
                  <option value="host">Host</option>
                  <option value="server">Server</option>
                  <option value="router">Router</option>
                  <option value="switch">Switch</option>
                  <option value="firewall">Firewall</option>
                  <option value="load_balancer">Load Balancer</option>
                </select>
              </div>

              <div className="form-group">
                <label>Address *</label>
                <input
                  type="text"
                  value={formData.address}
                  onChange={(e) => setFormData({...formData, address: e.target.value})}
                  placeholder="e.g., 192.168.1.1 or example.com"
                  required
                />
              </div>

              <div className="form-group">
                <label>Port (optional)</label>
                <input
                  type="number"
                  value={formData.port}
                  onChange={(e) => setFormData({...formData, port: e.target.value})}
                  placeholder="e.g., 80, 443, 3306"
                />
              </div>

              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({...formData, description: e.target.value})}
                  placeholder="Optional description"
                  rows={3}
                />
              </div>

              <div className="modal-actions">
                <button type="button" className="btn-secondary" onClick={() => setShowModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn-primary" disabled={submitting}>
                  {submitting ? 'Creating...' : 'Create Target'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};
