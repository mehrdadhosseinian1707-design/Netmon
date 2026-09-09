package monitors

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type HTTPMonitor struct {
	config HTTPConfig
}

type HTTPConfig struct {
	Method          string
	Headers         map[string]string
	Body            string
	FollowRedirects bool
	ValidateTLS     bool
	Timeout         time.Duration
	ExpectedStatus  []int
}

type HTTPResult struct {
	Success          bool
	StatusCode       int
	DNSTime          float64
	ConnectTime      float64
	TLSTime          float64
	TTFBTime         float64
	TotalTime        float64
	ResponseSize     int64
	RedirectCount    int
	TLSVersion       string
	TLSCipher        string
	CertExpiry       time.Time
	CertIssuer       string
	CertSubject      string
	CertSANs         []string
	ErrorMessage     string
}

func NewHTTPMonitor(config map[string]interface{}) (*HTTPMonitor, error) {
	cfg := HTTPConfig{
		Method:          "GET",
		Headers:         make(map[string]string),
		FollowRedirects: true,
		ValidateTLS:     true,
		Timeout:         30 * time.Second,
		ExpectedStatus:  []int{200},
	}

	if method, ok := config["method"].(string); ok {
		cfg.Method = method
	}
	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if str, ok := v.(string); ok {
				cfg.Headers[k] = str
			}
		}
	}
	if body, ok := config["body"].(string); ok {
		cfg.Body = body
	}
	if follow, ok := config["follow_redirects"].(bool); ok {
		cfg.FollowRedirects = follow
	}
	if validate, ok := config["validate_tls"].(bool); ok {
		cfg.ValidateTLS = validate
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if statuses, ok := config["expected_status"].([]interface{}); ok {
		cfg.ExpectedStatus = []int{}
		for _, s := range statuses {
			if status, ok := s.(float64); ok {
				cfg.ExpectedStatus = append(cfg.ExpectedStatus, int(status))
			}
		}
	}

	return &HTTPMonitor{config: cfg}, nil
}

func (m *HTTPMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.makeRequest(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeHTTP,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.LatencyMs = &result.TotalTime
		measurement.Metadata["status_code"] = result.StatusCode
		measurement.Metadata["dns_time_ms"] = result.DNSTime
		measurement.Metadata["connect_time_ms"] = result.ConnectTime
		measurement.Metadata["tls_time_ms"] = result.TLSTime
		measurement.Metadata["ttfb_ms"] = result.TTFBTime
		measurement.Metadata["response_size"] = result.ResponseSize
		measurement.Metadata["redirect_count"] = result.RedirectCount

		if result.TLSVersion != "" {
			measurement.Metadata["tls_version"] = result.TLSVersion
			measurement.Metadata["tls_cipher"] = result.TLSCipher
			measurement.Metadata["cert_expiry"] = result.CertExpiry
			measurement.Metadata["cert_issuer"] = result.CertIssuer
			measurement.Metadata["cert_subject"] = result.CertSubject
			measurement.Metadata["cert_sans"] = result.CertSANs

			// Calculate days until cert expiry
			daysUntilExpiry := int(time.Until(result.CertExpiry).Hours() / 24)
			measurement.Metadata["cert_days_until_expiry"] = daysUntilExpiry
		}
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *HTTPMonitor) makeRequest(ctx context.Context, url string) HTTPResult {
	result := HTTPResult{
		Success: false,
	}

	// Add scheme if not present
	if len(url) > 0 && url[0] != 'h' {
		url = "http://" + url
	}

	// Timing variables
	var dnsStart, connectStart, tlsStart, ttfbStart time.Time
	var dnsTime, connectTime, tlsTime time.Duration

	// Create custom transport with timing
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !m.config.ValidateTLS,
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dnsStart = time.Now()
			conn, err := (&net.Dialer{
				Timeout: m.config.Timeout,
			}).DialContext(ctx, network, addr)
			dnsTime = time.Since(dnsStart)
			connectStart = time.Now()
			if err != nil {
				return nil, err
			}
			connectTime = time.Since(connectStart)
			return conn, nil
		},
	}

	// Track TLS handshake
	transport.DialTLSContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dnsStart = time.Now()
		conn, err := tls.Dial(network, addr, transport.TLSClientConfig)
		if err != nil {
			return nil, err
		}
		dnsTime = time.Since(dnsStart)

		tlsStart = time.Now()
		err = conn.Handshake()
		tlsTime = time.Since(tlsStart)

		// Extract TLS info
		state := conn.ConnectionState()
		result.TLSVersion = getTLSVersion(state.Version)
		result.TLSCipher = tls.CipherSuiteName(state.CipherSuite)

		if len(state.PeerCertificates) > 0 {
			cert := state.PeerCertificates[0]
			result.CertExpiry = cert.NotAfter
			result.CertIssuer = cert.Issuer.String()
			result.CertSubject = cert.Subject.String()
			result.CertSANs = cert.DNSNames
		}

		return conn, err
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   m.config.Timeout,
	}

	if !m.config.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, m.config.Method, url, nil)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to create request: %v", err)
		return result
	}

	// Add headers
	for k, v := range m.config.Headers {
		req.Header.Set(k, v)
	}

	// Make request
	startTime := time.Now()
	ttfbStart = time.Now()

	resp, err := client.Do(req)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("request failed: %v", err)
		return result
	}
	defer resp.Body.Close()

	ttfbTime := time.Since(ttfbStart)

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to read response: %v", err)
		return result
	}

	totalTime := time.Since(startTime)

	// Populate result
	result.StatusCode = resp.StatusCode
	result.DNSTime = float64(dnsTime.Microseconds()) / 1000.0
	result.ConnectTime = float64(connectTime.Microseconds()) / 1000.0
	result.TLSTime = float64(tlsTime.Microseconds()) / 1000.0
	result.TTFBTime = float64(ttfbTime.Microseconds()) / 1000.0
	result.TotalTime = float64(totalTime.Microseconds()) / 1000.0
	result.ResponseSize = int64(len(body))

	// Check if status code is expected
	result.Success = false
	for _, expected := range m.config.ExpectedStatus {
		if result.StatusCode == expected {
			result.Success = true
			break
		}
	}

	if !result.Success {
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", result.StatusCode)
	}

	return result
}

func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04X)", version)
	}
}

func (m *HTTPMonitor) Type() string {
	return models.MonitorTypeHTTP
}

func (m *HTTPMonitor) ValidateConfig(config map[string]interface{}) error {
	if method, ok := config["method"].(string); ok {
		validMethods := map[string]bool{
			"GET": true, "HEAD": true, "POST": true,
			"PUT": true, "DELETE": true, "PATCH": true,
		}
		if !validMethods[method] {
			return fmt.Errorf("invalid HTTP method: %s", method)
		}
	}
	return nil
}
