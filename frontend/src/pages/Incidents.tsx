import { useState } from 'react';
import './Incidents.css';

interface Incident {
  id: string;
  type: string;
  resource: string;
  description: string;
  status: 'RESOLVED' | 'OPEN' | 'INVESTIGATING';
  severity: 'HIGH' | 'MEDIUM' | 'LOW' | 'CRITICAL';
  method: string;
  started: string;
  duration: string;
}

export const Incidents = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [typeFilter, setTypeFilter] = useState('All Types');
  const [statusFilter, setStatusFilter] = useState('All Status');
  const [dateFrom, setDateFrom] = useState('2026-02-12');
  const [dateTo, setDateTo] = useState('2026-03-14');
  const [perPage, setPerPage] = useState(100);

  const [incidents] = useState<Incident[]>([
    {
      id: '1',
      type: 'Website',
      resource: 'PB Djarum',
      description: 'Website PB Djarum went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 10:16:45 PM',
      duration: '0.0h'
    },
    {
      id: '2',
      type: 'Website',
      resource: 'IndoBSD',
      description: 'Website IndoBSD went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 3:40:46 PM',
      duration: '0.0h'
    },
    {
      id: '3',
      type: 'Website',
      resource: 'idve.cloud',
      description: 'Website idve.cloud went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 2:28:36 AM',
      duration: '0.0h'
    },
    {
      id: '4',
      type: 'Website',
      resource: 'IndoBSD',
      description: 'Website IndoBSD went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 2:28:36 AM',
      duration: '0.0h'
    },
    {
      id: '5',
      type: 'Website',
      resource: 'PB Djarum',
      description: 'Website PB Djarum went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 2:28:36 AM',
      duration: '0.0h'
    },
    {
      id: '6',
      type: 'Website',
      resource: 'NOC-System',
      description: 'Website NOC-System went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/14/2026, 2:28:36 AM',
      duration: '0.0h'
    },
    {
      id: '7',
      type: 'Website',
      resource: 'idve.cloud',
      description: 'Website idve.cloud went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/13/2026, 11:26:10 PM',
      duration: '0.0h'
    },
    {
      id: '8',
      type: 'Website',
      resource: 'PB Djarum',
      description: 'Website PB Djarum went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/13/2026, 11:26:10 PM',
      duration: '0.0h'
    },
    {
      id: '9',
      type: 'Website',
      resource: 'IndoBSD',
      description: 'Website IndoBSD went down',
      status: 'RESOLVED',
      severity: 'HIGH',
      method: 'WEBSITE',
      started: '3/13/2026, 11:26:10 PM',
      duration: '0.0h'
    }
  ]);

  const stats = {
    total: 15,
    open: 5,
    resolvedToday: 6,
    avgResolutionTime: '0.0h'
  };

  const filteredIncidents = incidents.filter(incident => {
    if (searchQuery && !incident.description.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false;
    }
    if (typeFilter !== 'All Types' && incident.type !== typeFilter) {
      return false;
    }
    if (statusFilter !== 'All Status' && incident.status !== statusFilter) {
      return false;
    }
    return true;
  });

  return (
    <div className="incidents-page">
      {/* Page Header */}
      <div className="incidents-header">
        <div>
          <h1 className="page-title">Incident Reports</h1>
          <p className="page-subtitle">Monitor and report on system incidents and outages.</p>
        </div>
        <div className="header-actions">
          <button className="btn-primary">
            <span>📄</span> Generate Report
          </button>
          <button className="btn-success">
            <span>📥</span> Export CSV
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="incident-stats-grid">
        <div className="stat-card">
          <div className="stat-icon danger">⚠️</div>
          <div className="stat-content">
            <div className="stat-label">Total Incidents</div>
            <div className="stat-value">{stats.total}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon warning">⏳</div>
          <div className="stat-content">
            <div className="stat-label">Open Incidents</div>
            <div className="stat-value">{stats.open}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon success">✅</div>
          <div className="stat-content">
            <div className="stat-label">Resolved Today</div>
            <div className="stat-value">{stats.resolvedToday}</div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon info">⏱️</div>
          <div className="stat-content">
            <div className="stat-label">Avg Resolution Time</div>
            <div className="stat-value">{stats.avgResolutionTime}</div>
          </div>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="search-bar">
        <div className="search-input-group">
          <input
            type="text"
            placeholder="Search incidents..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
          <button className="btn-search">🔍 Search</button>
        </div>

        <div className="filter-group">
          <div className="filter-item">
            <label>Type:</label>
            <select value={typeFilter} onChange={(e) => setTypeFilter(e.target.value)}>
              <option>All Types</option>
              <option>Website</option>
              <option>Network</option>
              <option>Server</option>
            </select>
          </div>

          <div className="filter-item">
            <label>Status:</label>
            <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
              <option>All Status</option>
              <option>OPEN</option>
              <option>RESOLVED</option>
              <option>INVESTIGATING</option>
            </select>
          </div>

          <div className="filter-item">
            <label>Date Range:</label>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
            />
            <span className="date-separator">to</span>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
            />
          </div>

          <button className="btn-apply">Apply Filters</button>
        </div>
      </div>

      {/* Incidents Table */}
      <div className="incidents-table-container">
        <div className="table-header">
          <h2 className="table-title">Incident History</h2>
          <div className="table-controls">
            <label>
              Per Page:
              <select value={perPage} onChange={(e) => setPerPage(Number(e.target.value))}>
                <option value={10}>10</option>
                <option value={25}>25</option>
                <option value={50}>50</option>
                <option value={100}>100</option>
              </select>
            </label>
          </div>
        </div>

        <table className="incidents-table">
          <thead>
            <tr>
              <th>TYPE</th>
              <th>RESOURCE</th>
              <th>DESCRIPTION</th>
              <th>STATUS</th>
              <th>SEVERITY</th>
              <th>METHOD</th>
              <th>STARTED</th>
              <th>DURATION</th>
              <th>PDF</th>
              <th>ACTIONS</th>
            </tr>
          </thead>
          <tbody>
            {filteredIncidents.map((incident) => (
              <tr key={incident.id}>
                <td>
                  <span className="type-badge">
                    <span className="type-icon">🌐</span>
                    {incident.type}
                  </span>
                </td>
                <td className="resource-cell">{incident.resource}</td>
                <td className="description-cell">{incident.description}</td>
                <td>
                  <span className={`status-badge ${incident.status.toLowerCase()}`}>
                    {incident.status}
                  </span>
                </td>
                <td>
                  <span className={`severity-badge ${incident.severity.toLowerCase()}`}>
                    {incident.severity}
                  </span>
                </td>
                <td className="method-cell">{incident.method}</td>
                <td className="time-cell">{incident.started}</td>
                <td className="duration-cell">{incident.duration}</td>
                <td>
                  <button className="icon-btn pdf-btn" title="Download PDF">
                    📄
                  </button>
                </td>
                <td>
                  <button className="icon-btn view-btn" title="View Details">
                    👁️
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};
