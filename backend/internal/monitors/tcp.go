package monitors

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"syscall"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type TCPMonitor struct {
	config TCPConfig
}

type TCPConfig struct {
	Timeout           time.Duration
	MeasureThroughput bool
	TransferSize      int64
}

type TCPResult struct {
	Success         bool
	ConnectTime     float64
	DNSTime         float64
	ThroughputMbps  float64
	Retransmissions int
	ErrorMessage    string
	LocalAddr       string
	RemoteAddr      string
	BytesTransferred int64
}

func NewTCPMonitor(config map[string]interface{}) (*TCPMonitor, error) {
	cfg := TCPConfig{
		Timeout:           10 * time.Second,
		MeasureThroughput: false,
		TransferSize:      1024 * 1024, // 1MB default
	}

	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if measure, ok := config["measure_throughput"].(bool); ok {
		cfg.MeasureThroughput = measure
	}
	if size, ok := config["transfer_size"].(float64); ok {
		cfg.TransferSize = int64(size)
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
		connectTime := result.ConnectTime
		measurement.TCPConnectTimeMs = &connectTime
		measurement.LatencyMs = &result.ConnectTime

		dnsTime := result.DNSTime
		measurement.DNSLatencyMs = &dnsTime

		if result.ThroughputMbps > 0 {
			measurement.TCPThroughputMbps = &result.ThroughputMbps
			measurement.Metadata["bytes_transferred"] = result.BytesTransferred
		}

		if result.Retransmissions >= 0 {
			measurement.TCPRetransmissions = &result.Retransmissions
		}

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
		Success:         false,
		Retransmissions: -1,
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

	// Get TCP retransmissions if available
	result.Retransmissions = m.getTCPRetransmissions(conn)

	// Measure throughput if enabled
	if m.config.MeasureThroughput {
		throughput, bytes, err := m.measureThroughput(ctx, conn)
		if err == nil {
			result.ThroughputMbps = throughput
			result.BytesTransferred = bytes
		}
	}

	return result
}

func (m *TCPMonitor) measureThroughput(ctx context.Context, conn net.Conn) (float64, int64, error) {
	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(m.config.Timeout))

	// Try to read data from connection
	buffer := make([]byte, 32*1024) // 32KB buffer
	totalBytes := int64(0)
	startTime := time.Now()

	for totalBytes < m.config.TransferSize {
		n, err := conn.Read(buffer)
		totalBytes += int64(n)

		if err == io.EOF {
			break
		}
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return 0, 0, err
		}

		// Check context
		select {
		case <-ctx.Done():
			return 0, 0, ctx.Err()
		default:
		}
	}

	duration := time.Since(startTime)
	if duration == 0 || totalBytes == 0 {
		return 0, totalBytes, nil
	}

	// Calculate throughput in Mbps
	throughputMbps := (float64(totalBytes) * 8) / (float64(duration.Milliseconds()) / 1000.0) / 1_000_000

	return throughputMbps, totalBytes, nil
}

func (m *TCPMonitor) getTCPRetransmissions(conn net.Conn) int {
	// This is platform-specific and may not work on all systems
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return -1
	}

	// Get raw connection
	rawConn, err := tcpConn.SyscallConn()
	if err != nil {
		return -1
	}

	var retransmits int = -1

	// Try to get TCP_INFO (Linux-specific)
	if runtime.GOOS == "linux" {
		rawConn.Control(func(fd uintptr) {
			// TCP_INFO syscall - platform specific
			// This is a simplified version; full implementation would use syscall.GetsockoptTCPInfo
			retransmits = 0 // Placeholder - actual implementation requires syscall
		})
	} else if runtime.GOOS == "windows" {
		// Windows TCP statistics
		rawConn.Control(func(fd uintptr) {
			// Get TCP statistics on Windows using GetTcpStatistics or similar
			// This requires platform-specific code
			retransmits = 0 // Placeholder
		})
	}

	return retransmits
}

// Helper function to get TCP info on Linux (requires unsafe and syscall)
func getTCPInfoLinux(fd uintptr) (int, error) {
	// This would use syscall.GetsockoptTCPInfo on Linux
	// For now, return -1 to indicate unavailable
	_ = syscall.TCP_NODELAY // Reference syscall package
	return -1, fmt.Errorf("not implemented")
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
