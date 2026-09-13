import { Link } from 'react-router-dom';
import './Navigation.css';

interface NavigationProps {
  onLogout: () => void;
}

export const Navigation = ({ onLogout }: NavigationProps) => {
  const user = JSON.parse(localStorage.getItem('user') || '{}');

  return (
    <nav className="navbar">
      <div className="navbar-brand">
        <Link to="/">
          <div className="brand-text">
            <span className="brand-name">ONCIC</span>
            <span className="brand-subtitle">Network Monitoring</span>
          </div>
        </Link>
      </div>

      <div className="navbar-menu">
        <Link to="/" className="navbar-item">Dashboard</Link>
        <Link to="/targets" className="navbar-item">Targets</Link>
        <Link to="/dns-servers" className="navbar-item">DNS Servers</Link>
        <Link to="/network-performance" className="navbar-item">Speed Test</Link>
        <Link to="/traceroute" className="navbar-item">Traceroute</Link>
        <Link to="/probes" className="navbar-item">Probes</Link>
        <Link to="/monitors" className="navbar-item">Monitors</Link>
        <Link to="/alerts" className="navbar-item">Alerts</Link>
      </div>

      <div className="navbar-user">
        <div className="user-info">
          <span className="user-name">{user.username || 'User'}</span>
          <span className="user-role">Administrator</span>
        </div>
        <button onClick={onLogout} className="btn-logout">
          Logout
        </button>
      </div>
    </nav>
  );
};
