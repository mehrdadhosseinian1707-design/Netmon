import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import './Auth.css';

interface SignInProps {
  onLogin: (token: string, user: any) => void;
}

export const SignIn = ({ onLogin }: SignInProps) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      // TODO: Replace with actual API call to backend
      // For now, simulate authentication
      if (username && password) {
        // Simulate API delay
        await new Promise(resolve => setTimeout(resolve, 800));

        // Mock successful login
        const mockToken = `token_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
        const mockUser = {
          id: '1',
          username: username,
          email: `${username}@oncic.local`,
          created_at: new Date().toISOString()
        };

        onLogin(mockToken, mockUser);
        navigate('/');
      } else {
        setError('Please enter username and password');
      }
    } catch (err: any) {
      setError(err.message || 'Failed to sign in');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="auth-container">
      <div className="auth-background"></div>

      <div className="auth-card">
        <div className="auth-header">
          <h1 className="auth-title">ONCIC</h1>
          <p className="auth-subtitle">Network Monitoring</p>
          <p className="auth-creator">Created by Medo</p>
        </div>

        <form onSubmit={handleSubmit} className="auth-form">
          <h2 className="form-title">Sign In</h2>

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="username">Username</label>
            <input
              id="username"
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Enter your username"
              disabled={loading}
              autoComplete="username"
            />
          </div>

          <div className="form-group">
            <label htmlFor="password">Password</label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Enter your password"
              disabled={loading}
              autoComplete="current-password"
            />
          </div>

          <button
            type="submit"
            className="btn-submit"
            disabled={loading}
          >
            {loading ? 'Signing in...' : 'Sign In'}
          </button>

          <div className="auth-footer">
            <p>
              Don't have an account?{' '}
              <Link to="/signup" className="auth-link">
                Sign Up
              </Link>
            </p>
          </div>
        </form>
      </div>
    </div>
  );
};
