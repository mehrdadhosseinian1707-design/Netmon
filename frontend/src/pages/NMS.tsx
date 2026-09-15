import { useState } from 'react';
import './NMS.css';

interface SNMPDevice {
  id: string;
  name: string;
  ip: string;
  type: string;
  status: 'online' | 'offline' | 'warning';
  uptime: string;
  cpuUsage: number;
  memoryUsage: number;
  interfaces: number;
}

export const NMS = () => {
  const [devices] = useState<SNMPDevice[]>([
    {
      id: '1',
      name: 'Core Router',
      ip: '192.168.1.1',
      type: 'Router',
      status: 'online',
      uptime: '45d 12h',
      cpuUsage: 23,
      memoryUsage: 45,
      interfaces: 24
    },
    {
      id: '2',
      name: 'Distribution Switch',
      ip: '192.168.1.2',
      type: 'Switch',
      status: 'online',
      uptime: '30d 5h',
      cpuUsage: 15,
      memoryUsage: 32,
      interfaces: 48
    },
    {
      id: '3',
      name: 'Firewall',
      ip: '192.168.1.254',
      type: 'Firewall',
      status: 'online',
      uptime: '90d 22h',
      cpuUsage: 56,
      memoryUsage: 78,
      interfaces: 4
    }
  ]);

  return (
    <div className="nms-page">
      {/* Page Header */}
      <div className="page-header">
        <div>
          <div className="page-breadcrumb">MONITORING</div>
          <h1 className="page-title">Network Management System</h1>
          <p className="page-subtitle">SNMP-based device monitoring and management.</p>
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
          <div className="stat-icon">📊</div>
          <div className="stat-content">
            <div className="stat-label">SNMP DEVICES</div>
            <div className="stat-value">15</div>
          </div>
        </div>

        <div className="stat-card stat-success">
          <div className="stat-icon">✅</div>
          <div className="stat-content">
            <div className="stat-label">ONLINE</div>
            <div className="stat-value">13</div>
          </div>
        </div>

        <div className="stat-card stat-danger">
          <div className="stat-icon">❌</div>
          <div className="stat-content">
            <div className="stat-label">OFFLINE</div>
            <div className="stat-value">2</div>
          </div>
        </div>

        <div className="stat-card stat-info">
          <div className="stat-icon">📡</div>
          <div className="stat-content">
            <div className="stat-label">INTERFACES</div>
            <div className="stat-value">284</div>
          </div>
        </div>
      </div>

      {/* Devices Grid */}
      <div className="devices-grid">
        {devices.map((device) => (
          <div key={device.id} className="device-card noc-card">
            <div className="device-header">
              <div className="device-info">
                <h3 className="device-name">{device.name}</h3>
                <p className="device-ip">{device.ip}</p>
              </div>
              <span className={`status-indicator ${device.status}`}>
                {device.status === 'online' && '🟢'}
                {device.status === 'offline' && '🔴'}
                {device.status === 'warning' && '🟡'}
              </span>
            </div>

            <div className="device-type-badge">{device.type}</div>

            <div className="device-metrics">
              <div className="metric-row">
                <span className="metric-label">Uptime</span>
                <span className="metric-value">{device.uptime}</span>
              </div>
              <div className="metric-row">
                <span className="metric-label">CPU Usage</span>
                <span className="metric-value">
                  {device.cpuUsage}%
                  <div className="progress-bar">
                    <div className="progress-fill" style={{ width: `${device.cpuUsage}%` }}></div>
                  </div>
                </span>
              </div>
              <div className="metric-row">
                <span className="metric-label">Memory</span>
                <span className="metric-value">
                  {device.memoryUsage}%
                  <div className="progress-bar">
                    <div className="progress-fill" style={{ width: `${device.memoryUsage}%` }}></div>
                  </div>
                </span>
              </div>
              <div className="metric-row">
                <span className="metric-label">Interfaces</span>
                <span className="metric-value">{device.interfaces}</span>
              </div>
            </div>

            <div className="device-actions">
              <button className="action-btn">📊 View Stats</button>
              <button className="action-btn">⚙️ Configure</button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};
