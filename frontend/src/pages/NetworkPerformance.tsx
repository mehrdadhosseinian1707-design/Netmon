import { useEffect, useState } from 'react';
import './NetworkPerformance.css';

interface PerformanceTest {
  timestamp: string;
  downloadSpeed: number; // Mbps
  uploadSpeed: number; // Mbps
  jitter: number; // ms
  publicIP: string;
  location: string;
  isp: string;
  status: 'testing' | 'completed' | 'failed';
}

export const NetworkPerformance = () => {
  const [currentTest, setCurrentTest] = useState<PerformanceTest | null>(null);
  const [testHistory, setTestHistory] = useState<PerformanceTest[]>([]);
  const [isTesting, setIsTesting] = useState(false);
  const [nextScheduledTest, setNextScheduledTest] = useState<string>('');

  const runSpeedTest = async () => {
    setIsTesting(true);

    const testData: PerformanceTest = {
      timestamp: new Date().toISOString(),
      downloadSpeed: 0,
      uploadSpeed: 0,
      jitter: 0,
      publicIP: '',
      location: '',
      isp: '',
      status: 'testing',
    };

    setCurrentTest(testData);

    try {
      // Get public IP and location
      const ipResponse = await fetch('http://localhost:8080/api/v1/tools/ip-info');
      if (ipResponse.ok) {
        const ipData = await ipResponse.json();
        testData.publicIP = ipData.ip;
        testData.location = ipData.location || 'Unknown';
        testData.isp = ipData.isp || 'Unknown';
      }

      setCurrentTest({...testData});

      // Test download speed
      const downloadResponse = await fetch('http://localhost:8080/api/v1/tools/speed-test/download');
      if (downloadResponse.ok) {
        const downloadData = await downloadResponse.json();
        testData.downloadSpeed = downloadData.speed_mbps;
      }

      setCurrentTest({...testData});

      // Test upload speed
      const uploadResponse = await fetch('http://localhost:8080/api/v1/tools/speed-test/upload', {
        method: 'POST',
        body: new Blob([new ArrayBuffer(1024 * 1024)]), // 1MB upload test
      });
      if (uploadResponse.ok) {
        const uploadData = await uploadResponse.json();
        testData.uploadSpeed = uploadData.speed_mbps;
      }

      setCurrentTest({...testData});

      // Test jitter
      const jitterResponse = await fetch('http://localhost:8080/api/v1/tools/jitter-test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ target: '8.8.8.8', count: 10 }),
      });
      if (jitterResponse.ok) {
        const jitterData = await jitterResponse.json();
        testData.jitter = jitterData.jitter;
      }

      testData.status = 'completed';
      setCurrentTest({...testData});

      // Save to history
      const newHistory = [testData, ...testHistory].slice(0, 20); // Keep last 20 tests
      setTestHistory(newHistory);
      localStorage.setItem('speedtest_history', JSON.stringify(newHistory));

    } catch (error) {
      testData.status = 'failed';
      setCurrentTest({...testData});
    } finally {
      setIsTesting(false);
    }
  };

  const calculateNextTest = () => {
    const now = new Date();
    const hours = now.getHours();

    // Schedule for 8 AM and 8 PM
    let nextTest = new Date(now);

    if (hours < 8) {
      nextTest.setHours(8, 0, 0, 0);
    } else if (hours < 20) {
      nextTest.setHours(20, 0, 0, 0);
    } else {
      nextTest.setDate(nextTest.getDate() + 1);
      nextTest.setHours(8, 0, 0, 0);
    }

    return nextTest;
  };

  useEffect(() => {
    // Load history from localStorage
    const savedHistory = localStorage.getItem('speedtest_history');
    if (savedHistory) {
      setTestHistory(JSON.parse(savedHistory));
    }

    // Calculate next scheduled test
    const nextTest = calculateNextTest();
    setNextScheduledTest(nextTest.toLocaleString());

    // Check if we should run a test now
    const lastTestTime = localStorage.getItem('last_speedtest_time');
    const now = new Date();
    const shouldRunTest = !lastTestTime ||
      (now.getTime() - new Date(lastTestTime).getTime() > 12 * 60 * 60 * 1000); // 12 hours

    if (shouldRunTest) {
      const testTime = calculateNextTest();
      const timeUntilTest = testTime.getTime() - now.getTime();

      if (timeUntilTest <= 60000) { // Within 1 minute
        runSpeedTest();
      }
    }

    // Set up scheduled tests (every 12 hours)
    const interval = setInterval(() => {
      const now = new Date();
      const hours = now.getHours();
      const minutes = now.getMinutes();

      // Run at 8:00 AM and 8:00 PM
      if ((hours === 8 || hours === 20) && minutes === 0) {
        runSpeedTest();
        localStorage.setItem('last_speedtest_time', now.toISOString());
      }
    }, 60000); // Check every minute

    return () => clearInterval(interval);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const getSpeedClass = (speed: number, type: 'download' | 'upload') => {
    const threshold = type === 'download' ? { good: 50, fair: 25 } : { good: 20, fair: 10 };
    if (speed >= threshold.good) return 'excellent';
    if (speed >= threshold.fair) return 'good';
    return 'poor';
  };

  const getJitterClass = (jitter: number) => {
    if (jitter < 5) return 'excellent';
    if (jitter < 15) return 'good';
    return 'poor';
  };

  return (
    <div className="network-performance">
      <div className="page-header">
        <div>
          <h1>Network Performance Monitor</h1>
          <p className="next-test">Next scheduled test: {nextScheduledTest}</p>
        </div>
        <button
          className="btn-test"
          onClick={runSpeedTest}
          disabled={isTesting}
        >
          {isTesting ? '🔄 Testing...' : '▶ Run Speed Test'}
        </button>
      </div>

      {/* Current Test Results */}
      {currentTest && (
        <section className="current-test">
          <h2>Current Test Results</h2>
          <div className="metrics-grid">
            <div className="metric-card">
              <div className="metric-icon">⬇️</div>
              <div className="metric-content">
                <div className="metric-label">Download Speed</div>
                <div className={`metric-value ${getSpeedClass(currentTest.downloadSpeed, 'download')}`}>
                  {currentTest.status === 'testing' && currentTest.downloadSpeed === 0 ? (
                    <span className="loading-dots">Testing...</span>
                  ) : (
                    <>{currentTest.downloadSpeed.toFixed(2)} <span className="unit">Mbps</span></>
                  )}
                </div>
              </div>
            </div>

            <div className="metric-card">
              <div className="metric-icon">⬆️</div>
              <div className="metric-content">
                <div className="metric-label">Upload Speed</div>
                <div className={`metric-value ${getSpeedClass(currentTest.uploadSpeed, 'upload')}`}>
                  {currentTest.status === 'testing' && currentTest.uploadSpeed === 0 ? (
                    <span className="loading-dots">Testing...</span>
                  ) : (
                    <>{currentTest.uploadSpeed.toFixed(2)} <span className="unit">Mbps</span></>
                  )}
                </div>
              </div>
            </div>

            <div className="metric-card">
              <div className="metric-icon">📊</div>
              <div className="metric-content">
                <div className="metric-label">Jitter</div>
                <div className={`metric-value ${getJitterClass(currentTest.jitter)}`}>
                  {currentTest.status === 'testing' && currentTest.jitter === 0 ? (
                    <span className="loading-dots">Testing...</span>
                  ) : (
                    <>{currentTest.jitter.toFixed(2)} <span className="unit">ms</span></>
                  )}
                </div>
              </div>
            </div>

            <div className="metric-card">
              <div className="metric-icon">🌐</div>
              <div className="metric-content">
                <div className="metric-label">Public IP</div>
                <div className="metric-value ip-value">
                  {currentTest.publicIP || <span className="loading-dots">Detecting...</span>}
                </div>
                {currentTest.location && (
                  <div className="metric-sub">{currentTest.location}</div>
                )}
                {currentTest.isp && (
                  <div className="metric-sub">ISP: {currentTest.isp}</div>
                )}
              </div>
            </div>
          </div>
        </section>
      )}

      {/* Test History */}
      <section className="test-history">
        <h2>Test History (Last 20 Tests)</h2>
        {testHistory.length === 0 ? (
          <div className="empty-state">
            <p>No test history yet. Run your first speed test!</p>
          </div>
        ) : (
          <div className="history-table">
            <div className="table-header">
              <div className="col-time">Date & Time</div>
              <div className="col-download">Download</div>
              <div className="col-upload">Upload</div>
              <div className="col-jitter">Jitter</div>
              <div className="col-ip">Public IP</div>
            </div>
            {testHistory.map((test, index) => (
              <div key={index} className="table-row" style={{ animationDelay: `${index * 0.05}s` }}>
                <div className="col-time">
                  {new Date(test.timestamp).toLocaleString()}
                </div>
                <div className={`col-download ${getSpeedClass(test.downloadSpeed, 'download')}`}>
                  {test.downloadSpeed.toFixed(2)} Mbps
                </div>
                <div className={`col-upload ${getSpeedClass(test.uploadSpeed, 'upload')}`}>
                  {test.uploadSpeed.toFixed(2)} Mbps
                </div>
                <div className={`col-jitter ${getJitterClass(test.jitter)}`}>
                  {test.jitter.toFixed(2)} ms
                </div>
                <div className="col-ip">
                  <code>{test.publicIP}</code>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
};
