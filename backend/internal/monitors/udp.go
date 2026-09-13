package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type UDPMonitor struct {
	config UDPConfig
}

type UDPConfig struct {
	Port         int
	PacketSize   int
	PacketCount  int
	TestDuration time.Duration
	Timeout      time.Duration
}

type UDPResult struct {
	Success           bool
	ThroughputMbps    float64
	PacketsSent       int
	PacketsReceived   int
	PacketLoss        float64
	BytesSent         int64
	BytesReceived     int64
	AvgLatencyMs      float64
	ErrorMessage      string
}

func NewUDPMonitor(config map[string]interface{}) (*UDPMonitor, error) {
	cfg := UDPConfig{
		Port:         9000,
		PacketSize:   1024,
		PacketCount:  100,
		TestDuration: 10 * time.Second,
		Timeout:      30 * time.Second,
	}

	if port, ok := config["port"].(float64); ok {
		cfg.Port = int(port)
	}
	if size, ok := config["packet_size"].(float64); ok {
		cfg.PacketSize = int(size)
	}
	if count, ok := config["packet_count"].(float64); ok {
		cfg.PacketCount = int(count)
	}
	if duration, ok := config["test_duration"].(float64); ok {
		cfg.TestDuration = time.Duration(duration) * time.Second
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}

	return &UDPMonitor{config: cfg}, nil
}

func (m *UDPMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	port := m.config.Port
	if target.Port != nil {
		port = *target.Port
	}

	result := m.measureUDPThroughput(ctx, target.Address, port)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeUDP,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.UDPThroughputMbps = &result.ThroughputMbps
		packetLoss := result.PacketLoss
		measurement.PacketLoss = &packetLoss
		avgLatency := result.AvgLatencyMs
		measurement.LatencyMs = &avgLatency

		measurement.Metadata["packets_sent"] = result.PacketsSent
		measurement.Metadata["packets_received"] = result.PacketsReceived
		measurement.Metadata["bytes_sent"] = result.BytesSent
		measurement.Metadata["bytes_received"] = result.BytesReceived
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *UDPMonitor) measureUDPThroughput(ctx context.Context, address string, port int) UDPResult {
	result := UDPResult{
		Success: false,
	}

	// Resolve address
	raddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", address, port))
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to resolve address: %v", err)
		return result
	}

	// Create UDP connection
	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to connect: %v", err)
		return result
	}
	defer conn.Close()

	// Set deadline
	deadline := time.Now().Add(m.config.Timeout)
	conn.SetDeadline(deadline)

	// Prepare packet data
	packetData := make([]byte, m.config.PacketSize)
	for i := range packetData {
		packetData[i] = byte(i % 256)
	}

	startTime := time.Now()
	packetsSent := 0
	bytesSent := int64(0)
	latencies := []time.Duration{}

	// Send packets
	for i := 0; i < m.config.PacketCount; i++ {
		if time.Since(startTime) > m.config.TestDuration {
			break
		}

		sendStart := time.Now()
		n, err := conn.Write(packetData)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("failed to send packet: %v", err)
			return result
		}

		packetsSent++
		bytesSent += int64(n)

		// Try to receive echo/response (if server echoes back)
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		receiveBuffer := make([]byte, m.config.PacketSize)
		_, err = conn.Read(receiveBuffer)

		if err == nil {
			latency := time.Since(sendStart)
			latencies = append(latencies, latency)
		}

		// Small delay between packets
		time.Sleep(10 * time.Millisecond)

		// Check context cancellation
		select {
		case <-ctx.Done():
			result.ErrorMessage = "context cancelled"
			return result
		default:
		}
	}

	duration := time.Since(startTime)
	result.PacketsSent = packetsSent
	result.BytesSent = bytesSent
	result.PacketsReceived = len(latencies)
	result.BytesReceived = int64(len(latencies) * m.config.PacketSize)

	if packetsSent > 0 {
		result.PacketLoss = float64(packetsSent-len(latencies)) / float64(packetsSent) * 100.0
	}

	if duration > 0 {
		// Calculate throughput in Mbps
		result.ThroughputMbps = (float64(bytesSent) * 8) / (float64(duration.Milliseconds()) / 1000.0) / 1_000_000
	}

	// Calculate average latency
	if len(latencies) > 0 {
		totalLatency := time.Duration(0)
		for _, lat := range latencies {
			totalLatency += lat
		}
		result.AvgLatencyMs = float64(totalLatency.Microseconds()) / float64(len(latencies)) / 1000.0
	}

	result.Success = packetsSent > 0
	return result
}

func (m *UDPMonitor) Type() string {
	return models.MonitorTypeUDP
}

func (m *UDPMonitor) ValidateConfig(config map[string]interface{}) error {
	if port, ok := config["port"].(float64); ok {
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
	}
	if size, ok := config["packet_size"].(float64); ok {
		if size < 8 || size > 65507 {
			return fmt.Errorf("packet_size must be between 8 and 65507")
		}
	}
	if count, ok := config["packet_count"].(float64); ok {
		if count < 1 || count > 10000 {
			return fmt.Errorf("packet_count must be between 1 and 10000")
		}
	}
	return nil
}
