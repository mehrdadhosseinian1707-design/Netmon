import './Topology.css';

export const Topology = () => {
  return (
    <div className="topology-page">
      {/* Page Header */}
      <div className="page-header">
        <div>
          <div className="page-breadcrumb">INFRASTRUCTURE</div>
          <h1 className="page-title">Network Topology</h1>
          <p className="page-subtitle">Visualize your network infrastructure and device connections.</p>
        </div>
        <div className="header-actions">
          <button className="btn-primary">
            <span>🔄</span> Refresh
          </button>
          <button className="btn-primary">
            <span>📥</span> Export
          </button>
        </div>
      </div>

      {/* Topology Canvas */}
      <div className="noc-card topology-card">
        <div className="topology-canvas">
          <div className="topology-placeholder">
            <div className="placeholder-icon">🌐</div>
            <h3>Network Topology Visualization</h3>
            <p>Network topology map will be displayed here.</p>
            <p className="placeholder-note">Connect devices and define relationships to see the topology graph.</p>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="noc-card legend-card">
        <h3 className="legend-title">Device Types</h3>
        <div className="legend-grid">
          <div className="legend-item">
            <span className="legend-icon router">🔷</span>
            <span>Router</span>
          </div>
          <div className="legend-item">
            <span className="legend-icon switch">🔶</span>
            <span>Switch</span>
          </div>
          <div className="legend-item">
            <span className="legend-icon firewall">🔴</span>
            <span>Firewall</span>
          </div>
          <div className="legend-item">
            <span className="legend-icon server">🟦</span>
            <span>Server</span>
          </div>
          <div className="legend-item">
            <span className="legend-icon ap">🟢</span>
            <span>Access Point</span>
          </div>
          <div className="legend-item">
            <span className="legend-icon client">⚪</span>
            <span>Client Device</span>
          </div>
        </div>
      </div>
    </div>
  );
};
