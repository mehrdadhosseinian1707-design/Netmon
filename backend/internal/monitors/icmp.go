package monitors

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/go-ping/ping"
	"github.com/netmon/netmon/internal/models"
)

type ICMPMonitor struct {
	config ICMPConfig
}

type ICMPConfig struct {
	Count      int
	PacketSize int
	Timeout    time.Duration
	Interval   time.Duration
}

type ICMPResult struct {
	Success      bool
	PacketsSent  int
	PacketsRecv  int
	PacketLoss   float64
	MinRTT       float64
	MaxRTT       float64
	AvgRTT       float64
	StdDevRTT    float64
	Jitter       float64
	ErrorMessage string
	IPAddr       string
}

func NewICMPMonitor(config map[string]interface{}) (*ICMPMonitor, error) {
	cfg := ICMPConfig{
		Count:      4,
		PacketSize: 64,
		Timeout:    10 * time.Second,
		Interval:   1 * time.Second,
	}

	if count, ok := config["count"].(float64); ok {
		cfg.Count = int(count)
	}
	if size, ok := config["packet_size"].(float64); ok {
		cfg.PacketSize = int(size)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}

	return &ICMPMonitor{config: cfg}, nil
}

func (m *ICMPMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.ping(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeICMP,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.LatencyMs = &result.AvgRTT
		measurement.PacketLoss = &result.PacketLoss
		measurement.JitterMs = &result.Jitter
		measurement.Metadata["min_rtt_ms"] = result.MinRTT
		measurement.Metadata["max_rtt_ms"] = result.MaxRTT
		measurement.Metadata["stddev_rtt_ms"] = result.StdDevRTT
		measurement.Metadata["packets_sent"] = result.PacketsSent
		measurement.Metadata["packets_recv"] = result.PacketsRecv
		measurement.Metadata["ip_addr"] = result.IPAddr
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *ICMPMonitor) ping(ctx context.Context, address string) ICMPResult {
	result := ICMPResult{
		Success: false,
	}

	pinger, err := ping.NewPinger(address)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to create pinger: %v", err)
		return result
	}

	pinger.Count = m.config.Count
	pinger.Size = m.config.PacketSize
	pinger.Timeout = m.config.Timeout
	pinger.Interval = m.config.Interval

	// Run privileged mode on Windows
	pinger.SetPrivileged(true)

	// Create a done channel for context cancellation
	done := make(chan bool)
	go func() {
		select {
		case <-ctx.Done():
			pinger.Stop()
		case <-done:
		}
	}()

	err = pinger.Run()
	close(done)

	if err != nil {
		result.ErrorMessage = fmt.Sprintf("ping failed: %v", err)
		return result
	}

	stats := pinger.Statistics()

	result.PacketsSent = stats.PacketsSent
	result.PacketsRecv = stats.PacketsRecv
	result.PacketLoss = stats.PacketLoss

	if stats.PacketsRecv > 0 {
		result.Success = true
		result.MinRTT = float64(stats.MinRtt.Microseconds()) / 1000.0
		result.MaxRTT = float64(stats.MaxRtt.Microseconds()) / 1000.0
		result.AvgRTT = float64(stats.AvgRtt.Microseconds()) / 1000.0
		result.StdDevRTT = float64(stats.StdDevRtt.Microseconds()) / 1000.0
		result.IPAddr = stats.IPAddr.String()

		// Calculate jitter from stddev
		result.Jitter = result.StdDevRTT
	} else {
		result.ErrorMessage = "no packets received"
	}

	return result
}

func (m *ICMPMonitor) Type() string {
	return models.MonitorTypeICMP
}

func (m *ICMPMonitor) ValidateConfig(config map[string]interface{}) error {
	if count, ok := config["count"].(float64); ok {
		if count < 1 || count > 100 {
			return fmt.Errorf("count must be between 1 and 100")
		}
	}
	if size, ok := config["packet_size"].(float64); ok {
		if size < 8 || size > 65507 {
			return fmt.Errorf("packet_size must be between 8 and 65507")
		}
	}
	return nil
}

// Calculate jitter from RTT samples
func calculateJitter(rtts []time.Duration) float64 {
	if len(rtts) < 2 {
		return 0
	}

	var diffs []float64
	for i := 1; i < len(rtts); i++ {
		diff := math.Abs(float64(rtts[i].Microseconds()) - float64(rtts[i-1].Microseconds()))
		diffs = append(diffs, diff)
	}

	sum := 0.0
	for _, diff := range diffs {
		sum += diff
	}

	return (sum / float64(len(diffs))) / 1000.0 // Convert to milliseconds
}
