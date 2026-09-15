import { useEffect, useState } from 'react';
import './Overview.css';

interface Stats {
  totalDevices: number;
  online: number;
  offline: number;
  avgLatency: number;
  dataCenters: number;
}

interface ServerInfo {
  hostname: string;
  os: string;
  kernel: string;
  uptime: string;
  cpuCores: number;
  ipAddress: string;
}

export const Overview = () => {
  const [stats] = useState<Stats>({
    totalDevices: 25,
    online: 25,
    offline: 0,
    avgLatency: 11,
    dataCenters: 3
  });

  const [serverInfo] = useState<ServerInfo>({
    hostname: 'dashboard',
    os: 'Debian GNU/Linux 13 (trixie)',
    kernel: '6.12.48+deb13-cloud-amd64',
    uptime: 'up 2 days, 3 hours, 46 minutes',
    cpuCores: 4,
    ipAddress: '192.168.203.151 (enp0s2)'
  });

  const [memoryUsed] = useState(81);
  const [storageUsed] = useState(14);
  const [cpuUsage, setCpuUsage] = useState(12.4);
  const [networkRx, setNetworkRx] = useState(143.2);
  const [networkTx, setNetworkTx] = useState(32.6);
  const [diskRead, setDiskRead] = useState(0);
  const [diskWrite, setDiskWrite] = useState(7.2);

  useEffect(() => {
    // Simulate real-time updates
    const interval = setInterval(() => {
      setCpuUsage(prev => Math.max(5, Math.min(100, prev + (Math.random() - 0.5) * 10)));
      setNetworkRx(prev => Math.max(0, prev + (Math.random() - 0.5) * 50));
      setNetworkTx(prev => Math.max(0, prev + (Math.random() - 0.5) * 20));
      setDiskRead(prev => Math.max(0, prev + (Math.random() - 0.5) * 5));
      setDiskWrite(prev => Math.max(0, prev + (Math.random() - 0.5) * 10));
    }, 3000);

    return () => clearInterval(interval);
  }, []);

  const getCircleProgress = (percentage: number) => {
    const radius = 70;
    const circumference = 2 * Math.PI * radius;
    const offset = circumference - (percentage / 100) * circumference;
    return { circumference, offset };
  };

  return (
    <div className="overview-page">
      {/* Page Header */}
      <div className="overview-header">
        <div>
          <div className="page-breadcrumb">NETWORK OPERATIONS CENTER</div>
          <h1 className="page-title">Overview</h1>
          <p className="page-subtitle">Manage and monitor your network infrastructure in real-time.</p>
        </div>
        <div className="header-actions">
          <button className="btn-primary">
            <span>📊</span> Add Data Center
          </button>
          <button className="btn-primary">
            <span>➕</span> Add Device
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">🖥️</div>
          <div className="stat-content">
            <div className="stat-label">TOTAL DEVICES</div>
            <div className="stat-value">{stats.totalDevices}</div>
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
          <div className="stat-icon">⚡</div>
          <div className="stat-content">
            <div className="stat-label">AVG LATENCY</div>
            <div className="stat-value">{stats.avgLatency}ms</div>
          </div>
        </div>

        <div className="stat-card stat-info">
          <div className="stat-icon">📡</div>
          <div className="stat-content">
            <div className="stat-label">DATA CENTERS</div>
            <div className="stat-value">{stats.dataCenters}</div>
          </div>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="overview-grid">
        {/* Server Info */}
        <div className="noc-card server-info-card">
          <div className="card-header">
            <span className="card-icon">💻</span>
            <span className="card-title">SERVER INFO</span>
          </div>
          <div className="server-details">
            <div className="detail-row">
              <span className="detail-label">Hostname</span>
              <span className="detail-value">{serverInfo.hostname}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">OS</span>
              <span className="detail-value">{serverInfo.os}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Kernel</span>
              <span className="detail-value">{serverInfo.kernel}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">Uptime</span>
              <span className="detail-value">{serverInfo.uptime}</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">CPU Cores</span>
              <span className="detail-value">{serverInfo.cpuCores} cores</span>
            </div>
            <div className="detail-row">
              <span className="detail-label">IP Addresses</span>
              <span className="detail-value">{serverInfo.ipAddress}</span>
            </div>
          </div>
        </div>

        {/* Memory Chart */}
        <div className="noc-card chart-card">
          <div className="card-header">
            <span className="card-icon">💾</span>
            <span className="card-title">MEMORY</span>
          </div>
          <div className="circular-chart">
            <svg width="200" height="200" viewBox="0 0 200 200">
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="rgba(255,255,255,0.05)"
                strokeWidth="12"
              />
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="var(--purple)"
                strokeWidth="12"
                strokeLinecap="round"
                strokeDasharray={getCircleProgress(memoryUsed).circumference}
                strokeDashoffset={getCircleProgress(memoryUsed).offset}
                transform="rotate(-90 100 100)"
                style={{ transition: 'stroke-dashoffset 1s ease' }}
              />
            </svg>
            <div className="chart-label">
              <div className="chart-percentage">{memoryUsed}%</div>
              <div className="chart-text">Used</div>
              <div className="chart-subtext">3.1 / 3.8 GB</div>
            </div>
          </div>
        </div>

        {/* Storage Chart */}
        <div className="noc-card chart-card">
          <div className="card-header">
            <span className="card-icon">💿</span>
            <span className="card-title">STORAGE (/)</span>
          </div>
          <div className="circular-chart">
            <svg width="200" height="200" viewBox="0 0 200 200">
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="rgba(255,255,255,0.05)"
                strokeWidth="12"
              />
              <circle
                cx="100"
                cy="100"
                r="70"
                fill="none"
                stroke="var(--orange)"
                strokeWidth="12"
                strokeLinecap="round"
                strokeDasharray={getCircleProgress(storageUsed).circumference}
                strokeDashoffset={getCircleProgress(storageUsed).offset}
                transform="rotate(-90 100 100)"
                style={{ transition: 'stroke-dashoffset 1s ease' }}
              />
            </svg>
            <div className="chart-label">
              <div className="chart-percentage">{storageUsed}%</div>
              <div className="chart-text">Used</div>
              <div className="chart-subtext">3.5 / 24.4 GB</div>
            </div>
          </div>
        </div>
      </div>

      {/* System Metrics Charts */}
      <div className="metrics-grid">
        {/* CPU Usage */}
        <div className="noc-card metric-card">
          <div className="metric-header">
            <div className="metric-title">
              <span className="metric-icon">⚡</span>
              <span>CPU USAGE</span>
            </div>
            <div className="metric-value">{cpuUsage.toFixed(1)}%</div>
          </div>
          <div className="metric-chart">
            <div className="chart-line" style={{ height: `${Math.min(cpuUsage, 100)}%` }}></div>
          </div>
        </div>

        {/* Network I/O */}
        <div className="noc-card metric-card">
          <div className="metric-header">
            <div className="metric-title">
              <span className="metric-icon">📶</span>
              <span>NETWORK I/O</span>
            </div>
            <div className="metric-value">{networkRx.toFixed(1)} KB/s <span className="rx-badge">RX</span> {networkTx.toFixed(1)} KB/s <span className="tx-badge">TX</span></div>
          </div>
          <div className="metric-chart dual-chart">
            <div className="chart-line rx-line" style={{ height: `${Math.min(networkRx / 2, 100)}%` }}></div>
            <div className="chart-line tx-line" style={{ height: `${Math.min(networkTx / 2, 100)}%` }}></div>
          </div>
        </div>

        {/* Disk I/O */}
        <div className="noc-card metric-card">
          <div className="metric-header">
            <div className="metric-title">
              <span className="metric-icon">💽</span>
              <span>DISK I/O</span>
            </div>
            <div className="metric-value">
              R:{diskRead.toFixed(1)} <span className="read-badge">Read</span> W:{diskWrite.toFixed(1)} <span className="write-badge">Write</span> MB/s
            </div>
          </div>
          <div className="metric-chart dual-chart">
            <div className="chart-line read-line" style={{ height: `${Math.min(diskRead * 10, 100)}%` }}></div>
            <div className="chart-line write-line" style={{ height: `${Math.min(diskWrite * 10, 100)}%` }}></div>
          </div>
        </div>
      </div>

      {/* Data Centers Section */}
      <div className="noc-card datacenters-card">
        <div className="card-header">
          <span className="card-title">Home DC</span>
          <span className="device-count">9 devices</span>
        </div>
      </div>
    </div>
  );
};
