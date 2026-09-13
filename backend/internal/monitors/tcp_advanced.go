package monitors

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/netmon/netmon/internal/models"
)

// TCPAdvancedMonitor implements advanced TCP analysis measuring handshake timing,
// retransmissions, congestion window, and bufferbloat detection.
type TCPAdvancedMonitor struct{}

// NewTCPAdvancedMonitor creates a new TCPAdvancedMonitor.
func NewTCPAdvancedMonitor() *TCPAdvancedMonitor {
	return &TCPAdvancedMonitor{}
}

// Type returns the monitor type identifier.
func (m *TCPAdvancedMonitor) Type() string {
	return "tcp_advanced"
}

// ValidateConfig validates the monitor configuration.
func (m *TCPAdvancedMonitor) ValidateConfig(config map[string]interface{}) error {
	if _, ok := config["host"]; !ok {
		return fmt.Errorf("tcp_advanced: missing required config field 'host'")
	}
	if _, ok := config["port"]; !ok {
		return fmt.Errorf("tcp_advanced: missing required config field 'port'")
	}
	return nil
}

// Check performs a full TCP analysis measurement.
func (m *TCPAdvancedMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	host, _ := config["host"].(string)
	if host == "" {
		host = target.Address
	}

	var port int
	switch v := config["port"].(type) {
	case int:
		port = v
	case float64:
		port = int(v)
	default:
		if target.Port != nil {
			port = *target.Port
		} else {
			port = 80
		}
	}

	timeoutSeconds := 10.0
	if ts, ok := config["timeout_seconds"].(float64); ok && ts > 0 {
		timeoutSeconds = ts
	}
	timeout := time.Duration(timeoutSeconds * float64(time.Second))

	measurement := &models.Measurement{
		Time:        time.Now().UTC(),
		MonitorType: "tcp_advanced",
		Metadata:    make(map[string]interface{}),
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", host, port)
	dialer := &net.Dialer{}

	// TCP handshake timing:
	// t0 = before dial (before SYN sent)
	// t1 = after DialContext returns (connection established, SYN-ACK received and ACK sent by OS)
	t0 := time.Now()
	conn, dialErr := dialer.DialContext(dialCtx, "tcp", addr)
	t1 := time.Now()

	if dialErr != nil {
		measurement.Success = false
		errClass := classifyTCPError(dialErr)
		measurement.Metadata["error_class"] = errClass
		measurement.ErrorMessage = dialErr.Error()
		return measurement, nil
	}

	// handshake_syn_ms: time from before SYN to connection established (SYN-ACK received + ACK sent)
	handshakeSynMs := float64(t1.Sub(t0).Microseconds()) / 1000.0
	measurement.Metadata["handshake_syn_ms"] = handshakeSynMs

	connectMs := handshakeSynMs
	measurement.TCPConnectTimeMs = &connectMs

	// handshake_ack_ms: latency of the first application-level Write after connect
	// This represents the time for our first data write (post-handshake ACK timing)
	tWriteBefore := time.Now()
	conn.Write([]byte("HEAD / HTTP/1.0\r\nHost: " + host + "\r\n\r\n"))
	tWriteAfter := time.Now()
	handshakeAckMs := float64(tWriteAfter.Sub(tWriteBefore).Microseconds()) / 1000.0
	measurement.Metadata["handshake_ack_ms"] = handshakeAckMs

	// Platform-specific TCP info: retransmissions and congestion window
	retrans, cwnd := getTCPInfo(conn)
	measurement.Metadata["cwnd"] = cwnd

	if runtime.GOOS == "windows" {
		measurement.Metadata["tcp_info_note"] = "tcp_info not available on windows; retransmissions and cwnd reported as 0"
	}

	conn.Close()

	retransVal := retrans
	measurement.TCPRetransmissions = &retransVal

	// Bufferbloat detection
	bbDialCtx, bbCancel := context.WithTimeout(ctx, timeout)
	defer bbCancel()

	bbRatio, bbStatus, bbErr := measureBufferbloat(bbDialCtx, addr, timeout)
	if bbErr == nil {
		measurement.Metadata["bufferbloat_ratio"] = bbRatio
		measurement.Metadata["bufferbloat_status"] = bbStatus
	} else {
		measurement.Metadata["bufferbloat_status"] = "error"
		measurement.Metadata["bufferbloat_error"] = bbErr.Error()
	}

	// Overall latency is the TCP connect time (SYN→established)
	latency := handshakeSynMs
	measurement.LatencyMs = &latency
	measurement.Success = true

	return measurement, nil
}

// classifyTCPError categorizes a TCP dial error into a known class:
// "rst", "refused", "timeout", "no_route", "host_unreachable", or "generic".
func classifyTCPError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "reset by peer") ||
		strings.Contains(msg, "forcibly closed") ||
		strings.Contains(msg, "econnreset") {
		return "rst"
	}
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "econnrefused") {
		return "refused"
	}
	if strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "deadline exceeded") ||
		strings.Contains(msg, "timed out") ||
		strings.Contains(msg, "etimedout") {
		return "timeout"
	}
	if strings.Contains(msg, "no route to host") ||
		strings.Contains(msg, "network unreachable") ||
		strings.Contains(msg, "enetunreach") {
		return "no_route"
	}
	if strings.Contains(msg, "host unreachable") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "ehostunreach") {
		return "host_unreachable"
	}
	return "generic"
}

// measureBufferbloat sends 10 small (64-byte) probe packets then 1 large (1400-byte) packet
// over a new TCP connection, measuring send latency for each. If large_lat/avg_small_lat > 1.5
// bufferbloat is classified as "detected".
func measureBufferbloat(ctx context.Context, addr string, timeout time.Duration) (float64, string, error) {
	dialer := &net.Dialer{}

	connCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, err := dialer.DialContext(connCtx, "tcp", addr)
	if err != nil {
		return 0, "", fmt.Errorf("bufferbloat connect: %w", err)
	}
	defer conn.Close()

	smallPkt := make([]byte, 64)
	largePkt := make([]byte, 1400)

	// Measure send latency for 10 small probes
	const smallCount = 10
	var smallTotal time.Duration
	successCount := 0
	for i := 0; i < smallCount; i++ {
		if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
			break
		}
		ts := time.Now()
		if _, werr := conn.Write(smallPkt); werr != nil {
			break
		}
		smallTotal += time.Since(ts)
		successCount++
	}

	if successCount == 0 {
		return 0, "unknown", nil
	}

	avgSmall := smallTotal / time.Duration(successCount)

	// Measure send latency for the large probe
	conn.SetDeadline(time.Now().Add(5 * time.Second))
	tl := time.Now()
	conn.Write(largePkt)
	largeLat := time.Since(tl)
	conn.SetDeadline(time.Time{})

	if avgSmall == 0 {
		return 0, "unknown", nil
	}

	ratio := float64(largeLat) / float64(avgSmall)
	status := "not_detected"
	if ratio > 1.5 {
		status = "detected"
	}

	return ratio, status, nil
}
