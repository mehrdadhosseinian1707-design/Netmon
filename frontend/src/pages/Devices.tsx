import { useState } from 'react';
import './Devices.css';

interface Device {
  id: string;
  name: string;
  type: string;
  ip: string;
  location: string;
  status: 'online' | 'offline' | 'warning';
  uptime: string;
  lastSeen: string;
}

export const Devices = () => {
  const [devices] = useState<Device[]>([
    {
      id: '1',
      name: 'Core Router 01',
      type: 'Router',
      ip: '192.168.1.1',
      location: 'Data Center A',
      status: 'online',
      uptime: '45d 12h 30m',
      lastSeen: '2 minutes ago'
    },
    {
      id: '2',
      name: 'Switch-Floor-3',
      type: 'Switch',
      ip: '192.168.1.10',
      location: 'Building 1 - Floor 3',
      status: 'online',
      uptime: '30d 5h 15m',
      lastSeen: '1 minute ago'
    },
    {
      id: '3',
      name: 'Firewall-01',
      type: 'Firewall',
      ip: '192.168.1.254',
      location: 'Data Center A',
      status: 'online',
      uptime: '90d 22h 45m',
      lastSeen: '30 seconds ago'
    },
    {
      id: '4',
      name: 'Access Point 12',
      type: 'Access Point',
      ip: '192.168.2.12',
      location: 'Building 2 - Floor 2',
      status: 'warning',
      uptime: '15d 8h 20m',
      lastSeen: '5 minutes ago'
    },
    {
      id: '5',
      name: 'Server-DB-01',
      type: 'Server',
      ip: '10.0.1.50',
      location: 'Data Center B',
      status: 'online',
      uptime: '120d 15h 10m',
      lastSeen: '1 minute ago'
    }
  ]);

  const stats = {
    total: 25,
    online: 23,
    offline: 0,
    warning: 2
  };

  return (
    <div className="devices-page">
      {/* Page Header */}
      <div className="page-header">
        <div>
          <div className="page-breadcrumb">INFRASTRUCTURE</div>
          <h1 className="page-title">Devices</h1>
          <p className="page-subtitle">Manage and monitor all network devices.</p>
        </div>
        <div className="header-actions">
          <button className="btn-primary">
            <span>➕</span> Add Device
          </button>
          <button className="btn-primary">
            <span>🔄</span> Scan Network
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">🖥️</div>
          <div className="stat-content">
            <div className="stat-label">TOTAL DEVICES</div>
            <div className="stat-value">{stats.total}</div>
          </div>
        </div>

        <div className="stat-card stat-success">
          <div className="stat-icon">✅</div>
          <div className="stat-content">
            <div className="stat-label">ONLINE</div>
            <div className="stat-value">{stats.online}</div>
          </div>
        </div>

        <div className="stat-card stat-danger">
          <div className="stat-icon">❌</div>
          <div className="stat-content">
            <div className="stat-label">OFFLINE</div>
            <div className="stat-value">{stats.offline}</div>
          </div>
        </div>

        <div className="stat-card stat-warning">
          <div className="stat-icon">⚠️</div>
          <div className="stat-content">
            <div className="stat-label">WARNING</div>
            <div className="stat-value">{stats.warning}</div>
          </div>
        </div>
      </div>

      {/* Devices Table */}
      <div className="noc-card">
        <div className="card-header">
          <h2 className="card-title">All Devices</h2>
          <div className="card-actions">
            <input type="text" placeholder="Search devices..." className="search-input" />
            <select className="filter-select">
              <option>All Types</option>
              <option>Router</option>
              <option>Switch</option>
              <option>Firewall</option>
              <option>Server</option>
              <option>Access Point</option>
            </select>
          </div>
        </div>

        <div className="table-container">
          <table className="devices-table">
            <thead>
              <tr>
                <th>STATUS</th>
                <th>NAME</th>
                <th>TYPE</th>
                <th>IP ADDRESS</th>
                <th>LOCATION</th>
                <th>UPTIME</th>
                <th>LAST SEEN</th>
                <th>ACTIONS</th>
              </tr>
            </thead>
            <tbody>
              {devices.map((device) => (
                <tr key={device.id}>
                  <td>
                    <span className={`status-indicator ${device.status}`}>
                      {device.status === 'online' && '🟢'}
                      {device.status === 'offline' && '🔴'}
                      {device.status === 'warning' && '🟡'}
                    </span>
                  </td>
                  <td className="device-name">{device.name}</td>
                  <td>
                    <span className="device-type">{device.type}</span>
                  </td>
                  <td className="device-ip">{device.ip}</td>
                  <td className="device-location">{device.location}</td>
                  <td className="device-uptime">{device.uptime}</td>
                  <td className="device-lastseen">{device.lastSeen}</td>
                  <td>
                    <div className="action-buttons">
                      <button className="action-btn" title="View Details">👁️</button>
                      <button className="action-btn" title="Edit">✏️</button>
                      <button className="action-btn danger" title="Delete">🗑️</button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};
