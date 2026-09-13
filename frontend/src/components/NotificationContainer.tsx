import { useState } from 'react';
import { useNotifications, notificationManager } from '../services/notifications';
import './NotificationContainer.css';

export const NotificationContainer = () => {
  const toasts = useNotifications();
  const [removingToasts, setRemovingToasts] = useState<Set<string>>(new Set());

  const handleClose = (id: string) => {
    setRemovingToasts(prev => new Set(prev).add(id));
    setTimeout(() => {
      notificationManager.remove(id);
      setRemovingToasts(prev => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }, 300);
  };

  const getIcon = (type: string) => {
    switch (type) {
      case 'success': return '✓';
      case 'error': return '✕';
      case 'warning': return '⚠';
      case 'info': return 'ℹ';
      default: return '•';
    }
  };

  return (
    <div className="notification-container">
      {toasts.map(toast => (
        <div
          key={toast.id}
          className={`notification-toast ${toast.type} ${removingToasts.has(toast.id) ? 'removing' : ''}`}
        >
          <div className="notification-header">
            <div className="notification-type">
              <span className="notification-icon">{getIcon(toast.type)}</span>
              {toast.type}
            </div>
            <button
              className="notification-close"
              onClick={() => handleClose(toast.id)}
              aria-label="Close notification"
            >
              ✕
            </button>
          </div>
          <div className="notification-message">{toast.message}</div>
          {toast.duration && toast.duration > 0 && (
            <div className="notification-progress">
              <div
                className="notification-progress-bar"
                style={{ animationDuration: `${toast.duration}ms` }}
              />
            </div>
          )}
        </div>
      ))}
    </div>
  );
};
