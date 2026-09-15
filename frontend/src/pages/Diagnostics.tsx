import { useState } from 'react';
import './Diagnostics.css';

interface SystemEvent {
  id: string;
  severity: 'error' | 'warning' | 'info';
  source: string;
  message: string;
  time: string;
}

export const Diagnostics = () => {
  const [activeTab, setActiveTab] = useState('system-events');
  const [filter, setFilter] = useState('all');
  const [hideResolved, setHideResolved] = useState(false);

  const [events] = useState<SystemEvent[]>([
    // Currently no events for demo
  ]);

  const [eventStats] = useState({
    errors: 0,
    warnings: 0,
    info: 0,
    unresolved: 0
  });

  const filteredEvents = events.filter(event => {
    if (filter === 'all') return true;
    return event.severity === filter;
  });

  return (
    <div className="diagnostics-page">
      {/* Page Header */}
      <div className="diagnostics-header">
        <div>
          <h1 className="page-title">Diagnostics & Troubleshooting</h1>
          <p className="page-subtitle">System event log and network diagnostic tools.</p>
        </div>
        <button className="btn-refresh">
          <span>🔄</span> Refresh
        </button>
      </div>

      <div className="diagnostics-layout">
        {/* Left Sidebar - Tools */}
        <div className="tools-sidebar">
          <div className="tools-header">TOOLS</div>

          <div className="tools-list">
            <button
              className={`tool-item ${activeTab === 'system-events' ? 'active' : ''}`}
              onClick={() => setActiveTab('system-events')}
            >
              <span className="tool-icon">⚡</span>
              <span className="tool-name">System Events</span>
            </button>

            <button
              className={`tool-item ${activeTab === 'ping' ? 'active' : ''}`}
              onClick={() => setActiveTab('ping')}
            >
              <span className="tool-icon">📡</span>
              <span className="tool-name">Ping Test</span>
            </button>

            <button
              className={`tool-item ${activeTab === 'traceroute' ? 'active' : ''}`}
              onClick={() => setActiveTab('traceroute')}
            >
              <span className="tool-icon">🔀</span>
              <span className="tool-name">Traceroute</span>
            </button>

            <button
              className={`tool-item ${activeTab === 'whois' ? 'active' : ''}`}
              onClick={() => setActiveTab('whois')}
            >
              <span className="tool-icon">🔍</span>
              <span className="tool-name">Whois</span>
            </button>
          </div>

          {/* Event Stats */}
          <div className="event-stats">
            <div className="stats-header">EVENT STATS</div>
            <div className="stat-item">
              <span className="stat-label">Errors</span>
              <span className="stat-value error">{eventStats.errors}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Warnings</span>
              <span className="stat-value warning">{eventStats.warnings}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Info</span>
              <span className="stat-value info">{eventStats.info}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">Unresolved</span>
              <span className="stat-value">{eventStats.unresolved}</span>
            </div>
          </div>
        </div>

        {/* Main Content Area */}
        <div className="diagnostics-content">
          {activeTab === 'system-events' && (
            <>
              {/* Filter Bar */}
              <div className="filter-bar">
                <div className="filter-group">
                  <span className="filter-label">Filter:</span>
                  <button
                    className={`filter-btn ${filter === 'all' ? 'active' : ''}`}
                    onClick={() => setFilter('all')}
                  >
                    All
                  </button>
                  <button
                    className={`filter-btn filter-error ${filter === 'error' ? 'active' : ''}`}
                    onClick={() => setFilter('error')}
                  >
                    Error
                  </button>
                  <button
                    className={`filter-btn filter-warning ${filter === 'warning' ? 'active' : ''}`}
                    onClick={() => setFilter('warning')}
                  >
                    Warning
                  </button>
                  <button
                    className={`filter-btn filter-info ${filter === 'info' ? 'active' : ''}`}
                    onClick={() => setFilter('info')}
                  >
                    Info
                  </button>
                </div>

                <div className="filter-actions">
                  <label className="checkbox-label">
                    <input
                      type="checkbox"
                      checked={hideResolved}
                      onChange={(e) => setHideResolved(e.target.checked)}
                    />
                    <span>Hide resolved</span>
                  </label>
                  <button className="clear-btn">Clear all</button>
                </div>
              </div>

              {/* Events Table */}
              <div className="events-table-container">
                <table className="events-table">
                  <thead>
                    <tr>
                      <th>SEVERITY</th>
                      <th>SOURCE</th>
                      <th>MESSAGE</th>
                      <th>TIME</th>
                      <th>ACTION</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredEvents.length === 0 ? (
                      <tr>
                        <td colSpan={5}>
                          <div className="no-events">
                            <span className="success-icon">✅</span>
                            <span>No events match filter.</span>
                          </div>
                        </td>
                      </tr>
                    ) : (
                      filteredEvents.map(event => (
                        <tr key={event.id}>
                          <td>
                            <span className={`severity-badge ${event.severity}`}>
                              {event.severity.toUpperCase()}
                            </span>
                          </td>
                          <td>{event.source}</td>
                          <td>{event.message}</td>
                          <td>{event.time}</td>
                          <td>
                            <button className="action-btn">View</button>
                          </td>
                        </tr>
                      ))
                    )}
                  </tbody>
                </table>
              </div>
            </>
          )}

          {activeTab === 'ping' && (
            <div className="tool-panel">
              <h2 className="tool-title">Ping Test</h2>
              <p className="tool-description">Test network connectivity to a host.</p>

              <div className="tool-form">
                <div className="form-group">
                  <label>Host or IP Address</label>
                  <input type="text" placeholder="e.g., google.com or 8.8.8.8" />
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label>Count</label>
                    <input type="number" defaultValue={4} min={1} max={100} />
                  </div>
                  <div className="form-group">
                    <label>Timeout (seconds)</label>
                    <input type="number" defaultValue={5} min={1} max={60} />
                  </div>
                </div>
                <button className="btn-primary">Run Ping Test</button>
              </div>

              <div className="tool-results">
                <div className="results-header">Results</div>
                <div className="results-content">
                  <p className="placeholder-text">Run a ping test to see results here.</p>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'traceroute' && (
            <div className="tool-panel">
              <h2 className="tool-title">Traceroute</h2>
              <p className="tool-description">Trace the network path to a destination.</p>

              <div className="tool-form">
                <div className="form-group">
                  <label>Destination Host</label>
                  <input type="text" placeholder="e.g., google.com" />
                </div>
                <div className="form-row">
                  <div className="form-group">
                    <label>Max Hops</label>
                    <input type="number" defaultValue={30} min={1} max={64} />
                  </div>
                  <div className="form-group">
                    <label>Timeout (seconds)</label>
                    <input type="number" defaultValue={5} min={1} max={60} />
                  </div>
                </div>
                <button className="btn-primary">Run Traceroute</button>
              </div>

              <div className="tool-results">
                <div className="results-header">Results</div>
                <div className="results-content">
                  <p className="placeholder-text">Run a traceroute to see the path here.</p>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'whois' && (
            <div className="tool-panel">
              <h2 className="tool-title">Whois Lookup</h2>
              <p className="tool-description">Query domain registration information.</p>

              <div className="tool-form">
                <div className="form-group">
                  <label>Domain or IP Address</label>
                  <input type="text" placeholder="e.g., example.com" />
                </div>
                <button className="btn-primary">Lookup</button>
              </div>

              <div className="tool-results">
                <div className="results-header">Results</div>
                <div className="results-content">
                  <p className="placeholder-text">Enter a domain to see registration details.</p>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
