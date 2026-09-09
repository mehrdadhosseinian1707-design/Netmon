import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import { Dashboard } from './pages/Dashboard';
import { TargetsList } from './pages/TargetsList';
import { TargetDetail } from './pages/TargetDetail';
import './App.css';

function App() {
  return (
    <Router>
      <div className="app">
        <nav className="navbar">
          <div className="navbar-brand">
            <Link to="/">NetMon</Link>
          </div>
          <div className="navbar-menu">
            <Link to="/" className="navbar-item">Dashboard</Link>
            <Link to="/targets" className="navbar-item">Targets</Link>
            <Link to="/probes" className="navbar-item">Probes</Link>
            <Link to="/monitors" className="navbar-item">Monitors</Link>
            <Link to="/alerts" className="navbar-item">Alerts</Link>
          </div>
        </nav>

        <main className="main-content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/targets" element={<TargetsList />} />
            <Route path="/targets/:id" element={<TargetDetail />} />
            <Route path="*" element={<div className="not-found">Page not found</div>} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
