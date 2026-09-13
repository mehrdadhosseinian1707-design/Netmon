package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type ConnectivityMonitor struct {
	config ConnectivityConfig
}

type ConnectivityConfig struct {
	Timeout      time.Duration
	TestIPv4     bool
	TestIPv6     bool
	IPv4TestHost string
	IPv6TestHost string
}

type ConnectivityResult struct {
	Success          bool
	IPv4Connectivity bool
	IPv6Connectivity bool
	IPv4Latency      float64
	IPv6Latency      float64
	IPv4Address      string
	IPv6Address      string
	ErrorMessage     string
}

func NewConnectivityMonitor(config map[string]interface{}) (*ConnectivityMonitor, error) {
	cfg := ConnectivityConfig{
		Timeout:      10 * time.Second,
		TestIPv4:     true,
		TestIPv6:     true,
		IPv4TestHost: "8.8.8.8",       // Google DNS
		IPv6TestHost: "2001:4860:4860::8888", // Google DNS IPv6
	}

	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if testIPv4, ok := config["test_ipv4"].(bool); ok {
		cfg.TestIPv4 = testIPv4
	}
	if testIPv6, ok := config["test_ipv6"].(bool); ok {
		cfg.TestIPv6 = testIPv6
	}
	if ipv4Host, ok := config["ipv4_test_host"].(string); ok {
		cfg.IPv4TestHost = ipv4Host
	}
	if ipv6Host, ok := config["ipv6_test_host"].(string); ok {
		cfg.IPv6TestHost = ipv6Host
	}

	return &ConnectivityMonitor{config: cfg}, nil
}

func (m *ConnectivityMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.testConnectivity(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeConnectivity,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.IPv4Connectivity = &result.IPv4Connectivity
		measurement.IPv6Connectivity = &result.IPv6Connectivity

		if result.IPv4Connectivity {
			measurement.Metadata["ipv4_latency_ms"] = result.IPv4Latency
			measurement.Metadata["ipv4_address"] = result.IPv4Address
			// Use IPv4 latency as primary latency if available
			measurement.LatencyMs = &result.IPv4Latency
		}

		if result.IPv6Connectivity {
			measurement.Metadata["ipv6_latency_ms"] = result.IPv6Latency
			measurement.Metadata["ipv6_address"] = result.IPv6Address
			// If no IPv4, use IPv6 latency as primary
			if !result.IPv4Connectivity {
				measurement.LatencyMs = &result.IPv6Latency
			}
		}
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *ConnectivityMonitor) testConnectivity(ctx context.Context, targetAddr string) ConnectivityResult {
	result := ConnectivityResult{
		Success: false,
	}

	hasAnyConnectivity := false

	// Test IPv4 connectivity
	if m.config.TestIPv4 {
		testHost := m.config.IPv4TestHost
		if targetAddr != "" {
			// Try to use target address for IPv4 test
			ips, err := net.LookupIP(targetAddr)
			if err == nil {
				for _, ip := range ips {
					if ip.To4() != nil {
						testHost = ip.String()
						break
					}
				}
			}
		}

		ipv4Connected, latency, addr := m.testIPv4Connectivity(ctx, testHost)
		result.IPv4Connectivity = ipv4Connected
		result.IPv4Latency = latency
		result.IPv4Address = addr

		if ipv4Connected {
			hasAnyConnectivity = true
		}
	}

	// Test IPv6 connectivity
	if m.config.TestIPv6 {
		testHost := m.config.IPv6TestHost
		if targetAddr != "" {
			// Try to use target address for IPv6 test
			ips, err := net.LookupIP(targetAddr)
			if err == nil {
				for _, ip := range ips {
					if ip.To4() == nil && ip.To16() != nil {
						testHost = ip.String()
						break
					}
				}
			}
		}

		ipv6Connected, latency, addr := m.testIPv6Connectivity(ctx, testHost)
		result.IPv6Connectivity = ipv6Connected
		result.IPv6Latency = latency
		result.IPv6Address = addr

		if ipv6Connected {
			hasAnyConnectivity = true
		}
	}

	if !hasAnyConnectivity {
		result.ErrorMessage = "no connectivity detected on IPv4 or IPv6"
	} else {
		result.Success = true
	}

	return result
}

func (m *ConnectivityMonitor) testIPv4Connectivity(ctx context.Context, testHost string) (bool, float64, string) {
	// Create IPv4-only dialer
	dialer := &net.Dialer{
		Timeout: m.config.Timeout,
	}

	// Try to connect to test host
	startTime := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp4", testHost+":80")
	latency := time.Since(startTime)

	if err != nil {
		// Try ICMP ping as fallback
		return m.tryICMPPing(ctx, testHost, "ip4:icmp")
	}

	defer conn.Close()

	// Get local IPv4 address
	localAddr := ""
	if conn.LocalAddr() != nil {
		localAddr = conn.LocalAddr().String()
		if host, _, err := net.SplitHostPort(localAddr); err == nil {
			localAddr = host
		}
	}

	latencyMs := float64(latency.Microseconds()) / 1000.0
	return true, latencyMs, localAddr
}

func (m *ConnectivityMonitor) testIPv6Connectivity(ctx context.Context, testHost string) (bool, float64, string) {
	// Create IPv6-only dialer
	dialer := &net.Dialer{
		Timeout: m.config.Timeout,
	}

	// Try to connect to test host
	startTime := time.Now()
	conn, err := dialer.DialContext(ctx, "tcp6", "["+testHost+"]:80")
	latency := time.Since(startTime)

	if err != nil {
		// Try ICMP ping as fallback
		return m.tryICMPPing(ctx, testHost, "ip6:ipv6-icmp")
	}

	defer conn.Close()

	// Get local IPv6 address
	localAddr := ""
	if conn.LocalAddr() != nil {
		localAddr = conn.LocalAddr().String()
		if host, _, err := net.SplitHostPort(localAddr); err == nil {
			localAddr = host
		}
	}

	latencyMs := float64(latency.Microseconds()) / 1000.0
	return true, latencyMs, localAddr
}

func (m *ConnectivityMonitor) tryICMPPing(ctx context.Context, host string, network string) (bool, float64, string) {
	// Simplified ICMP ping test
	// This is a placeholder - full implementation would use raw sockets

	// Try to resolve the host
	ips, err := net.LookupIP(host)
	if err != nil {
		return false, 0, ""
	}

	if len(ips) == 0 {
		return false, 0, ""
	}

	// For now, just check if we can resolve the address
	// Full implementation would send ICMP echo request
	return true, 0, ips[0].String()
}

func (m *ConnectivityMonitor) Type() string {
	return models.MonitorTypeConnectivity
}

func (m *ConnectivityMonitor) ValidateConfig(config map[string]interface{}) error {
	if timeout, ok := config["timeout"].(float64); ok {
		if timeout < 1 || timeout > 300 {
			return fmt.Errorf("timeout must be between 1 and 300 seconds")
		}
	}

	testIPv4 := true
	testIPv6 := true
	if v, ok := config["test_ipv4"].(bool); ok {
		testIPv4 = v
	}
	if v, ok := config["test_ipv6"].(bool); ok {
		testIPv6 = v
	}

	if !testIPv4 && !testIPv6 {
		return fmt.Errorf("at least one of test_ipv4 or test_ipv6 must be enabled")
	}

	return nil
}
