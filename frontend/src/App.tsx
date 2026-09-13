import { useState, useEffect } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { SignIn } from './pages/SignIn';
import { SignUp } from './pages/SignUp';
import { AdvancedDashboard } from './pages/AdvancedDashboard';
import { TargetsList } from './pages/TargetsList';
import { TargetDetail } from './pages/TargetDetail';
import { DNSServers } from './pages/DNSServers';
import { NetworkPerformance } from './pages/NetworkPerformance';
import { Traceroute } from './pages/Traceroute';
import { Navigation } from './components/Navigation';
import { NotificationContainer } from './components/NotificationContainer';
import './App.css';

function App() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Check if user is authenticated (check localStorage for token)
    const token = localStorage.getItem('auth_token');
    const user = localStorage.getItem('user');

    if (token && user) {
      setIsAuthenticated(true);
    }
    setIsLoading(false);
  }, []);

  const handleLogin = (token: string, user: any) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('user', JSON.stringify(user));
    setIsAuthenticated(true);
  };

  const handleLogout = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('user');
    setIsAuthenticated(false);
  };

  if (isLoading) {
    return (
      <div className="app-loading">
        <div className="loading-spinner"></div>
        <p>Loading ONCIC...</p>
      </div>
    );
  }

  return (
    <Router>
      <div className="app">
        {!isAuthenticated ? (
          <Routes>
            <Route path="/signin" element={<SignIn onLogin={handleLogin} />} />
            <Route path="/signup" element={<SignUp onSignUp={handleLogin} />} />
            <Route path="*" element={<Navigate to="/signin" replace />} />
          </Routes>
        ) : (
          <>
            <Navigation onLogout={handleLogout} />
            <NotificationContainer />
            <main className="main-content">
              <Routes>
                <Route path="/" element={<AdvancedDashboard />} />
                <Route path="/targets" element={<TargetsList />} />
                <Route path="/targets/:id" element={<TargetDetail />} />
                <Route path="/dns-servers" element={<DNSServers />} />
                <Route path="/network-performance" element={<NetworkPerformance />} />
                <Route path="/traceroute" element={<Traceroute />} />
                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </main>
          </>
        )}
      </div>
    </Router>
  );
}

export default App;
