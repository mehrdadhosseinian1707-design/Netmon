import { useState } from 'react';
import './NetworkTools.css';

interface ToolResult {
  tool: string;
  data: any;
  timestamp: string;
}

export const NetworkTools = () => {
  const [activeTab, setActiveTab] = useState('whois');
  const [inputValue, setInputValue] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<ToolResult | null>(null);

  const tools = [
    { id: 'whois', name: 'WHOIS Lookup', icon: '🔍', placeholder: 'Enter domain (e.g., google.com)' },
    { id: 'port-scan', name: 'Port Scanner', icon: '🔌', placeholder: 'Enter IP or domain' },
    { id: 'ssl-check', name: 'SSL Certificate', icon: '🔒', placeholder: 'Enter domain (e.g., google.com)' },
    { id: 'mtr', name: 'MTR (Advanced Trace)', icon: '🛣️', placeholder: 'Enter IP or domain' },
    { id: 'asn-lookup', name: 'ASN Lookup', icon: '🌐', placeholder: 'Enter ASN or IP' },
    { id: 'geoip', name: 'GeoIP Location', icon: '📍', placeholder: 'Enter IP address' },
    { id: 'dns-propagation', name: 'DNS Propagation', icon: '🔄', placeholder: 'Enter domain' },
    { id: 'bgp-route', name: 'BGP Route', icon: '🗺️', placeholder: 'Enter IP prefix' },
    { id: 'packet-loss', name: 'Packet Loss Test', icon: '📊', placeholder: 'Enter target IP' },
    { id: 'bandwidth', name: 'Bandwidth Test', icon: '⚡', placeholder: 'Enter server URL' },
  ];

  const runTool = async () => {
    if (!inputValue.trim()) return;

    setIsLoading(true);
    setResult(null);

    try {
      const response = await fetch(`http://localhost:8080/api/v1/tools/${activeTab}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ input: inputValue }),
      });

      if (response.ok) {
        const data = await response.json();
        setResult({
          tool: activeTab,
          data,
          timestamp: new Date().toISOString(),
        });
      }
    } catch (error) {
      console.error('Tool execution failed:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const renderResult = () => {
    if (!result) return null;

    switch (result.tool) {
      case 'whois':
        return (
          <div className="result-whois">
            <h3>WHOIS Information</h3>
            <div className="info-grid">
              <div className="info-item">
                <span className="label">Domain:</span>
                <span className="value">{result.data.domain}</span>
              </div>
              <div className="info-item">
                <span className="label">Registrar:</span>
                <span className="value">{result.data.registrar}</span>
              </div>
              <div className="info-item">
                <span className="label">Created:</span>
                <span className="value">{result.data.created_date}</span>
              </div>
              <div className="info-item">
                <span className="label">Expires:</span>
                <span className="value">{result.data.expiry_date}</span>
              </div>
              <div className="info-item">
                <span className="label">Name Servers:</span>
                <span className="value">{result.data.name_servers?.join(', ')}</span>
              </div>
              <div className="info-item">
                <span className="label">Status:</span>
                <span className="value">{result.data.status}</span>
              </div>
            </div>
          </div>
        );

      case 'port-scan':
        return (
          <div className="result-portscan">
            <h3>Open Ports ({result.data.open_ports?.length || 0})</h3>
            <div className="ports-grid">
              {result.data.open_ports?.map((port: any) => (
                <div key={port.port} className="port-card">
                  <div className="port-number">{port.port}</div>
                  <div className="port-service">{port.service}</div>
                  <div className="port-protocol">{port.protocol}</div>
                  <div className={`port-status ${port.state}`}>{port.state}</div>
                </div>
              ))}
            </div>
          </div>
        );

      case 'ssl-check':
        return (
          <div className="result-ssl">
            <h3>SSL/TLS Certificate Information</h3>
            <div className="cert-status">
              <div className={`cert-badge ${result.data.valid ? 'valid' : 'invalid'}`}>
                {result.data.valid ? '✓ Valid Certificate' : '✗ Invalid Certificate'}
              </div>
            </div>
            <div className="info-grid">
              <div className="info-item">
                <span className="label">Issued To:</span>
                <span className="value">{result.data.common_name}</span>
              </div>
              <div className="info-item">
                <span className="label">Issued By:</span>
                <span className="value">{result.data.issuer}</span>
              </div>
              <div className="info-item">
                <span className="label">Valid From:</span>
                <span className="value">{result.data.valid_from}</span>
              </div>
              <div className="info-item">
                <span className="label">Valid Until:</span>
                <span className="value">{result.data.valid_until}</span>
              </div>
              <div className="info-item">
                <span className="label">Days Remaining:</span>
                <span className={`value ${result.data.days_remaining < 30 ? 'warning' : ''}`}>
                  {result.data.days_remaining} days
                </span>
              </div>
              <div className="info-item">
                <span className="label">Subject Alt Names:</span>
                <span className="value">{result.data.san?.join(', ')}</span>
              </div>
            </div>
          </div>
        );

      case 'mtr':
        return (
          <div className="result-mtr">
            <h3>MTR Report - {result.data.target}</h3>
            <div className="mtr-stats">
              <div className="stat">Loss: {result.data.packet_loss}%</div>
              <div className="stat">Avg: {result.data.avg_latency}ms</div>
              <div className="stat">Hops: {result.data.hops?.length}</div>
            </div>
            <table className="mtr-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Host</th>
                  <th>Loss%</th>
                  <th>Snt</th>
                  <th>Last</th>
                  <th>Avg</th>
                  <th>Best</th>
                  <th>Wrst</th>
                  <th>StDev</th>
                </tr>
              </thead>
              <tbody>
                {result.data.hops?.map((hop: any) => (
                  <tr key={hop.number}>
                    <td>{hop.number}</td>
                    <td><code>{hop.ip}</code></td>
                    <td>{hop.loss}%</td>
                    <td>{hop.sent}</td>
                    <td>{hop.last}ms</td>
                    <td>{hop.avg}ms</td>
                    <td>{hop.best}ms</td>
                    <td>{hop.worst}ms</td>
                    <td>{hop.stdev}ms</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        );

      case 'asn-lookup':
        return (
          <div className="result-asn">
            <h3>ASN Information</h3>
            <div className="asn-card">
              <div className="asn-number">AS{result.data.asn}</div>
              <div className="asn-name">{result.data.name}</div>
            </div>
            <div className="info-grid">
              <div className="info-item">
                <span className="label">Organization:</span>
                <span className="value">{result.data.organization}</span>
              </div>
              <div className="info-item">
                <span className="label">Country:</span>
                <span className="value">{result.data.country}</span>
              </div>
              <div className="info-item">
                <span className="label">Registry:</span>
                <span className="value">{result.data.registry}</span>
              </div>
              <div className="info-item">
                <span className="label">IP Prefixes:</span>
                <span className="value">{result.data.prefixes?.join(', ')}</span>
              </div>
            </div>
          </div>
        );

      case 'geoip':
        return (
          <div className="result-geoip">
            <h3>GeoIP Location - {result.data.ip}</h3>
            <div className="location-card">
              <div className="flag">{result.data.flag}</div>
              <div className="location-info">
                <div className="city">{result.data.city}</div>
                <div className="region">{result.data.region}, {result.data.country}</div>
              </div>
            </div>
            <div className="info-grid">
              <div className="info-item">
                <span className="label">ISP:</span>
                <span className="value">{result.data.isp}</span>
              </div>
              <div className="info-item">
                <span className="label">Organization:</span>
                <span className="value">{result.data.organization}</span>
              </div>
              <div className="info-item">
                <span className="label">Timezone:</span>
                <span className="value">{result.data.timezone}</span>
              </div>
              <div className="info-item">
                <span className="label">Coordinates:</span>
                <span className="value">{result.data.latitude}, {result.data.longitude}</span>
              </div>
            </div>
          </div>
        );

      case 'dns-propagation':
        return (
          <div className="result-dns-prop">
            <h3>DNS Propagation Check - {result.data.domain}</h3>
            <div className="propagation-grid">
              {result.data.servers?.map((server: any) => (
                <div key={server.location} className={`dns-server-card ${server.synced ? 'synced' : 'pending'}`}>
                  <div className="location">{server.location}</div>
                  <div className="ip-result">{server.result}</div>
                  <div className={`status ${server.synced ? 'synced' : 'pending'}`}>
                    {server.synced ? '✓ Synced' : '⏳ Pending'}
                  </div>
                </div>
              ))}
            </div>
          </div>
        );

      case 'packet-loss':
        return (
          <div className="result-packet-loss">
            <h3>Packet Loss Analysis</h3>
            <div className="loss-summary">
              <div className="loss-metric">
                <div className="metric-value">{result.data.packet_loss}%</div>
                <div className="metric-label">Packet Loss</div>
              </div>
              <div className="loss-metric">
                <div className="metric-value">{result.data.packets_sent}</div>
                <div className="metric-label">Packets Sent</div>
              </div>
              <div className="loss-metric">
                <div className="metric-value">{result.data.packets_received}</div>
                <div className="metric-label">Packets Received</div>
              </div>
            </div>
            <div className="loss-chart">
              <div className="chart-bar" style={{ width: `${100 - result.data.packet_loss}%` }}>
                Success: {100 - result.data.packet_loss}%
              </div>
              <div className="chart-bar loss" style={{ width: `${result.data.packet_loss}%` }}>
                Loss: {result.data.packet_loss}%
              </div>
            </div>
          </div>
        );

      default:
        return <div className="result-generic"><pre>{JSON.stringify(result.data, null, 2)}</pre></div>;
    }
  };

  const currentTool = tools.find(t => t.id === activeTab);

  return (
    <div className="network-tools">
      <div className="page-header">
        <h1>Professional Network Tools</h1>
        <p className="page-description">Enterprise-grade network diagnostics and analysis</p>
      </div>

      <div className="tools-container">
        <div className="tools-sidebar">
          {tools.map(tool => (
            <button
              key={tool.id}
              className={`tool-btn ${activeTab === tool.id ? 'active' : ''}`}
              onClick={() => {
                setActiveTab(tool.id);
                setResult(null);
                setInputValue('');
              }}
            >
              <span className="tool-icon">{tool.icon}</span>
              <span className="tool-name">{tool.name}</span>
            </button>
          ))}
        </div>

        <div className="tool-content">
          <div className="tool-input-section">
            <h2>{currentTool?.icon} {currentTool?.name}</h2>
            <div className="input-group">
              <input
                type="text"
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                placeholder={currentTool?.placeholder}
                onKeyPress={(e) => e.key === 'Enter' && runTool()}
              />
              <button
                className="btn-run"
                onClick={runTool}
                disabled={isLoading || !inputValue.trim()}
              >
                {isLoading ? '⏳ Running...' : '▶ Run'}
              </button>
            </div>
          </div>

          {result && (
            <div className="tool-result-section">
              <div className="result-header">
                <span className="result-timestamp">
                  {new Date(result.timestamp).toLocaleString()}
                </span>
              </div>
              {renderResult()}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
