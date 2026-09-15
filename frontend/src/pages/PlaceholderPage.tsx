import './PlaceholderPage.css';

interface PlaceholderPageProps {
  title: string;
  subtitle: string;
  icon: string;
  description?: string;
}

export const PlaceholderPage = ({ title, subtitle, icon, description }: PlaceholderPageProps) => {
  return (
    <div className="placeholder-page">
      <div className="placeholder-header">
        <div>
          <div className="page-breadcrumb">NETWORK OPERATIONS CENTER</div>
          <h1 className="page-title">
            <span className="title-icon">{icon}</span>
            {title}
          </h1>
          <p className="page-subtitle">{subtitle}</p>
        </div>
      </div>

      <div className="placeholder-content">
        <div className="placeholder-card">
          <div className="placeholder-icon">{icon}</div>
          <h2>Coming Soon</h2>
          <p className="placeholder-description">
            {description || `The ${title} feature is currently under development and will be available soon.`}
          </p>
          <div className="placeholder-features">
            <div className="feature-badge">📊 Real-time Monitoring</div>
            <div className="feature-badge">📈 Advanced Analytics</div>
            <div className="feature-badge">⚡ High Performance</div>
            <div className="feature-badge">🔔 Smart Alerts</div>
          </div>
        </div>
      </div>
    </div>
  );
};