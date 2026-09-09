package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type TCPMonitor struct {
	config TCPConfig
}

type TCPConfig struct {
	Timeout time.Duration
}

type TCPResult struct {
	Success         bool
	ConnectTime     float64
	DNSTime         float64
	ErrorMessage    string
	LocalAddr       string
	RemoteAddr      string
}

func NewTCPMonitor(config map[string]interface{}) (*TCPMonitor, error) {
	cfg := TCPConfig{
		Timeout: 10 * time.Second,
	}

	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}

	return &TCPMonitor{config: cfg}, nil
}

func (m *TCPMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	if target.Port == nil {
		return nil, fmt.Errorf("TCP monitor requires port")
	}

	result := m.connect(ctx, target.Address, *target.Port)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeTCP,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.LatencyMs = &result.ConnectTime
		measurement.Metadata["dns_time_ms"] = result.DNSTime
		measurement.Metadata["local_addr"] = result.LocalAddr
		measurement.Metadata["remote_addr"] = result.RemoteAddr
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *TCPMonitor) connect(ctx context.Context, address string, port int) TCPResult {
	result := TCPResult{
		Success: false,
	}

	// Measure DNS resolution time
	dnsStart := time.Now()
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", address)
	dnsDuration := time.Since(dnsStart)
	result.DNSTime = float64(dnsDuration.Microseconds()) / 1000.0

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("DNS resolution failed: %v", err)
		return result
	}

	if len(ips) == 0 {
		result.ErrorMessage = "no IP addresses found"
		return result
	}

	// Use the first resolved IP
	targetAddr := fmt.Sprintf("%s:%d", ips[0].String(), port)

	// Measure connection time
	connectStart := time.Now()
	dialer := &net.Dialer{
		Timeout: m.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", targetAddr)
	connectDuration := time.Since(connectStart)
	result.ConnectTime = float64(connectDuration.Microseconds()) / 1000.0

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			result.ErrorMessage = "connection timeout"
		} else {
			result.ErrorMessage = fmt.Sprintf("connection failed: %v", err)
		}
		return result
	}
	defer conn.Close()

	result.Success = true
	result.LocalAddr = conn.LocalAddr().String()
	result.RemoteAddr = conn.RemoteAddr().String()

	return result
}

func (m *TCPMonitor) Type() string {
	return models.MonitorTypeTCP
}

func (m *TCPMonitor) ValidateConfig(config map[string]interface{}) error {
	if timeout, ok := config["timeout"].(float64); ok {
		if timeout < 1 || timeout > 300 {
			return fmt.Errorf("timeout must be between 1 and 300 seconds")
		}
	}
	return nil
}
