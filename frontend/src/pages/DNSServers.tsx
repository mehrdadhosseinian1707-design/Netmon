import { useEffect, useState } from 'react';
import './DNSServers.css';

interface DNSServer {
  name: string;
  ip: string;
  provider: string;
  location: string;
  ping?: number;
  status: 'checking' | 'online' | 'offline';
}

interface DoHServer {
  name: string;
  url: string;
  provider: string;
  status: 'checking' | 'accessible' | 'unavailable';
  responseTime?: number;
}

export const DNSServers = () => {
  const [dnsServers, setDnsServers] = useState<DNSServer[]>([
    { name: 'Google DNS Primary', ip: '8.8.8.8', provider: 'Google', location: 'Global', status: 'checking' },
    { name: 'Google DNS Secondary', ip: '8.8.4.4', provider: 'Google', location: 'Global', status: 'checking' },
    { name: 'Cloudflare DNS Primary', ip: '1.1.1.1', provider: 'Cloudflare', location: 'Global', status: 'checking' },
    { name: 'Cloudflare DNS Secondary', ip: '1.0.0.1', provider: 'Cloudflare', location: 'Global', status: 'checking' },
    { name: 'Quad9 DNS Primary', ip: '9.9.9.9', provider: 'Quad9', location: 'Global', status: 'checking' },
    { name: 'Quad9 DNS Secondary', ip: '149.112.112.112', provider: 'Quad9', location: 'Global', status: 'checking' },
    { name: 'OpenDNS Primary', ip: '208.67.222.222', provider: 'Cisco OpenDNS', location: 'Global', status: 'checking' },
    { name: 'OpenDNS Secondary', ip: '208.67.220.220', provider: 'Cisco OpenDNS', location: 'Global', status: 'checking' },
    { name: 'AdGuard DNS Primary', ip: '94.140.14.14', provider: 'AdGuard', location: 'Global', status: 'checking' },
    { name: 'AdGuard DNS Secondary', ip: '94.140.15.15', provider: 'AdGuard', location: 'Global', status: 'checking' },
  ]);

  const [dohServers, setDohServers] = useState<DoHServer[]>([
    { name: 'Cloudflare DoH', url: 'https://cloudflare-dns.com/dns-query', provider: 'Cloudflare', status: 'checking' },
    { name: 'Google DoH', url: 'https://dns.google/dns-query', provider: 'Google', status: 'checking' },
    { name: 'Quad9 DoH', url: 'https://dns.quad9.net/dns-query', provider: 'Quad9', status: 'checking' },
    { name: 'AdGuard DoH', url: 'https://dns.adguard.com/dns-query', provider: 'AdGuard', status: 'checking' },
    { name: 'OpenDNS DoH', url: 'https://doh.opendns.com/dns-query', provider: 'Cisco OpenDNS', status: 'checking' },
    { name: 'CleanBrowsing DoH', url: 'https://doh.cleanbrowsing.org/doh/family-filter', provider: 'CleanBrowsing', status: 'checking' },
  ]);

  const [loading, setLoading] = useState(true);

  const checkDNSServers = async () => {
    const updatedServers = await Promise.all(
      dnsServers.map(async (server) => {
        try {
          const startTime = Date.now();
          const response = await fetch(`http://localhost:8080/api/v1/tools/ping`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ target: server.ip, count: 3 }),
          });

          if (response.ok) {
            const data = await response.json();
            const ping = Math.round(data.avg_rtt || data.latency || (Date.now() - startTime));
            return { ...server, ping, status: 'online' as const };
          }
          return { ...server, status: 'offline' as const };
        } catch (error) {
          return { ...server, status: 'offline' as const };
        }
      })
    );
    setDnsServers(updatedServers);
  };

  const checkDoHServers = async () => {
    const updatedServers = await Promise.all(
      dohServers.map(async (server) => {
        try {
          const startTime = Date.now();
          const response = await fetch(`http://localhost:8080/api/v1/tools/doh-check`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: server.url }),
          });

          const responseTime = Date.now() - startTime;

          if (response.ok) {
            return { ...server, status: 'accessible' as const, responseTime };
          }
          return { ...server, status: 'unavailable' as const };
        } catch (error) {
          return { ...server, status: 'unavailable' as const };
        }
      })
    );
    setDohServers(updatedServers);
  };

  useEffect(() => {
    const init = async () => {
      setLoading(true);
      await Promise.all([checkDNSServers(), checkDoHServers()]);
      setLoading(false);
    };

    init();

    // Refresh every 30 seconds
    const interval = setInterval(() => {
      checkDNSServers();
      checkDoHServers();
    }, 30000);

    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const sortedDnsServers = [...dnsServers].sort((a, b) => {
    if (a.status === 'online' && b.status !== 'online') return -1;
    if (a.status !== 'online' && b.status === 'online') return 1;
    if (a.ping && b.ping) return a.ping - b.ping;
    return 0;
  });

  const accessibleDoH = dohServers.filter(s => s.status === 'accessible');

  return (
    <div className="dns-servers">
      <div className="page-header">
        <h1>DNS & DoH Servers Monitor</h1>
        <div className="header-stats">
          <div className="stat-pill">
            <span className="stat-label">DNS Servers Online:</span>
            <span className="stat-value">{dnsServers.filter(s => s.status === 'online').length}/{dnsServers.length}</span>
          </div>
          <div className="stat-pill">
            <span className="stat-label">DoH Accessible:</span>
            <span className="stat-value">{accessibleDoH.length}/{dohServers.length}</span>
          </div>
        </div>
      </div>

      {/* DNS Servers Section */}
      <section className="dns-section">
        <div className="section-header">
          <h2>🌐 DNS Servers</h2>
          <p>Traditional DNS servers with ping latency</p>
        </div>

        {loading ? (
          <div className="loading-state">Checking DNS servers...</div>
        ) : (
          <div className="servers-grid">
            {sortedDnsServers.map((server, index) => (
              <div key={server.ip} className={`server-card ${server.status}`} style={{ animationDelay: `${index * 0.05}s` }}>
                <div className="server-header">
                  <div className="server-info">
                    <h3>{server.name}</h3>
                    <span className="server-provider">{server.provider}</span>
                  </div>
                  <div className="server-status">
                    {server.status === 'checking' && <span className="status-badge checking">⏳ Checking</span>}
                    {server.status === 'online' && <span className="status-badge online">✓ Online</span>}
                    {server.status === 'offline' && <span className="status-badge offline">✗ Offline</span>}
                  </div>
                </div>

                <div className="server-details">
                  <div className="detail-row">
                    <span className="detail-label">IP Address:</span>
                    <code className="detail-value">{server.ip}</code>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">Location:</span>
                    <span className="detail-value">{server.location}</span>
                  </div>
                  {server.ping !== undefined && (
                    <div className="detail-row ping-row">
                      <span className="detail-label">Ping:</span>
                      <span className={`ping-value ${server.ping < 50 ? 'excellent' : server.ping < 100 ? 'good' : 'poor'}`}>
                        {server.ping} ms
                      </span>
                    </div>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* DoH Servers Section */}
      <section className="doh-section">
        <div className="section-header">
          <h2>🔒 DNS over HTTPS (DoH) Servers</h2>
          <p>Encrypted DNS servers - Only accessible servers shown</p>
        </div>

        {loading ? (
          <div className="loading-state">Checking DoH servers...</div>
        ) : (
          <>
            {accessibleDoH.length === 0 ? (
              <div className="empty-state">
                <p>No DoH servers are currently accessible</p>
              </div>
            ) : (
              <div className="servers-grid">
                {accessibleDoH.map((server, index) => (
                  <div key={server.url} className="server-card accessible" style={{ animationDelay: `${index * 0.05}s` }}>
                    <div className="server-header">
                      <div className="server-info">
                        <h3>{server.name}</h3>
                        <span className="server-provider">{server.provider}</span>
                      </div>
                      <span className="status-badge accessible">✓ Accessible</span>
                    </div>

                    <div className="server-details">
                      <div className="detail-row">
                        <span className="detail-label">URL:</span>
                        <code className="detail-value url-value">{server.url}</code>
                      </div>
                      {server.responseTime !== undefined && (
                        <div className="detail-row ping-row">
                          <span className="detail-label">Response Time:</span>
                          <span className={`ping-value ${server.responseTime < 200 ? 'excellent' : server.responseTime < 500 ? 'good' : 'poor'}`}>
                            {server.responseTime} ms
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
};
