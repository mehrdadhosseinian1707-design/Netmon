import { useEffect, useState } from 'react';
import './Traceroute.css';

interface Hop {
  number: number;
  ip: string;
  hostname: string;
  rtt1: number;
  rtt2: number;
  rtt3: number;
  avgRtt: number;
}

interface TracerouteResult {
  target: string;
  targetName: string;
  provider: string;
  hops: Hop[];
  totalHops: number;
  status: 'running' | 'completed' | 'failed';
}

export const Traceroute = () => {
  const [traceroutes, setTraceroutes] = useState<TracerouteResult[]>([]);
  const [isRunning, setIsRunning] = useState(false);
  const [expandedTarget, setExpandedTarget] = useState<string | null>(null);

  const dnsServers = [
    { target: '8.8.8.8', name: 'Google DNS Primary', provider: 'Google' },
    { target: '1.1.1.1', name: 'Cloudflare DNS', provider: 'Cloudflare' },
    { target: '208.67.222.222', name: 'OpenDNS Primary', provider: 'Cisco OpenDNS' },
    { target: '9.9.9.9', name: 'Quad9 DNS', provider: 'Quad9' },
    { target: '4.2.2.2', name: 'Level3 DNS', provider: 'Level3 (Microsoft)' },
  ];

  const runTraceroute = async (server: typeof dnsServers[0]) => {
    const result: TracerouteResult = {
      target: server.target,
      targetName: server.name,
      provider: server.provider,
      hops: [],
      totalHops: 0,
      status: 'running',
    };

    // Update state to show running
    setTraceroutes(prev => {
      const filtered = prev.filter(t => t.target !== server.target);
      return [...filtered, result];
    });

    try {
      const response = await fetch('http://localhost:8080/api/v1/tools/traceroute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: server.target }),
      });

      if (response.ok) {
        const data = await response.json();
        result.hops = data.hops || [];
        result.totalHops = result.hops.length;
        result.status = 'completed';
      } else {
        result.status = 'failed';
      }
    } catch (error) {
      result.status = 'failed';
    }

    // Update with results
    setTraceroutes(prev => {
      const filtered = prev.filter(t => t.target !== server.target);
      return [...filtered, result];
    });
  };

  const runAllTraceroutes = async () => {
    setIsRunning(true);
    setTraceroutes([]);

    // Run all traceroutes in parallel
    await Promise.all(dnsServers.map(server => runTraceroute(server)));

    setIsRunning(false);
  };

  useEffect(() => {
    // Auto-run on mount
    runAllTraceroutes();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const toggleExpand = (target: string) => {
    setExpandedTarget(expandedTarget === target ? null : target);
  };

  const getLatencyClass = (rtt: number) => {
    if (rtt < 50) return 'excellent';
    if (rtt < 100) return 'good';
    if (rtt < 200) return 'fair';
    return 'poor';
  };

  // Sort traceroutes by provider name
  const sortedTraceroutes = [...traceroutes].sort((a, b) =>
    a.provider.localeCompare(b.provider)
  );

  return (
    <div className="traceroute-page">
      <div className="page-header">
        <div>
          <h1>DNS Traceroute Monitor</h1>
          <p className="page-description">Network path analysis to major DNS providers</p>
        </div>
        <button
          className="btn-trace"
          onClick={runAllTraceroutes}
          disabled={isRunning}
        >
          {isRunning ? '🔄 Running Traceroutes...' : '▶ Run Traceroute'}
        </button>
      </div>

      <div className="traceroute-grid">
        {sortedTraceroutes.map((trace, index) => (
          <div
            key={trace.target}
            className={`traceroute-card ${trace.status}`}
            style={{ animationDelay: `${index * 0.1}s` }}
          >
            <div className="trace-header" onClick={() => toggleExpand(trace.target)}>
              <div className="trace-info">
                <h3>{trace.targetName}</h3>
                <div className="trace-meta">
                  <span className="trace-provider">{trace.provider}</span>
                  <code className="trace-ip">{trace.target}</code>
                </div>
              </div>

              <div className="trace-summary">
                {trace.status === 'running' && (
                  <div className="status-badge running">
                    <span className="spinner"></span> Running...
                  </div>
                )}
                {trace.status === 'completed' && (
                  <>
                    <div className="hop-count">
                      <span className="count-value">{trace.totalHops}</span>
                      <span className="count-label">Hops</span>
                    </div>
                    <button className="expand-btn">
                      {expandedTarget === trace.target ? '▼' : '▶'}
                    </button>
                  </>
                )}
                {trace.status === 'failed' && (
                  <div className="status-badge failed">✗ Failed</div>
                )}
              </div>
            </div>

            {expandedTarget === trace.target && trace.status === 'completed' && (
              <div className="trace-hops">
                <div className="hops-header">
                  <div className="hop-col-number">#</div>
                  <div className="hop-col-ip">IP Address</div>
                  <div className="hop-col-hostname">Hostname</div>
                  <div className="hop-col-rtts">RTT (ms)</div>
                </div>

                {trace.hops.map((hop) => (
                  <div key={hop.number} className="hop-row">
                    <div className="hop-col-number">
                      <span className="hop-number">{hop.number}</span>
                    </div>
                    <div className="hop-col-ip">
                      <code>{hop.ip}</code>
                    </div>
                    <div className="hop-col-hostname">
                      {hop.hostname !== hop.ip ? hop.hostname : '—'}
                    </div>
                    <div className="hop-col-rtts">
                      <span className={`rtt ${getLatencyClass(hop.rtt1)}`}>
                        {hop.rtt1.toFixed(1)}
                      </span>
                      <span className={`rtt ${getLatencyClass(hop.rtt2)}`}>
                        {hop.rtt2.toFixed(1)}
                      </span>
                      <span className={`rtt ${getLatencyClass(hop.rtt3)}`}>
                        {hop.rtt3.toFixed(1)}
                      </span>
                      <span className="rtt-avg">
                        avg: {hop.avgRtt.toFixed(1)} ms
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>

      {traceroutes.length === 0 && !isRunning && (
        <div className="empty-state">
          <p>Click "Run Traceroute" to trace the network path to DNS servers</p>
        </div>
      )}
    </div>
  );
};
