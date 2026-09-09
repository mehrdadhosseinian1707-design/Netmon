package monitors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/netmon/netmon/internal/models"
	"github.com/netmon/netmon/internal/monitors"
)

func TestICMPMonitor(t *testing.T) {
	config := map[string]interface{}{
		"count":   float64(4),
		"timeout": float64(5),
	}

	monitor, err := monitors.NewICMPMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create ICMP monitor: %v", err)
	}

	target := &models.Target{
		Address: "8.8.8.8",
		Name:    "Google DNS",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	measurement, err := monitor.Check(ctx, target)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	fmt.Printf("\nICMP Test Results:\n")
	fmt.Printf("Target: %s\n", target.Address)
	fmt.Printf("Success: %v\n", measurement.Success)
	if measurement.Success {
		fmt.Printf("Latency: %.2f ms\n", *measurement.LatencyMs)
		fmt.Printf("Packet Loss: %.2f%%\n", *measurement.PacketLoss)
		fmt.Printf("Jitter: %.2f ms\n", *measurement.JitterMs)
		fmt.Printf("Min RTT: %.2f ms\n", measurement.Metadata["min_rtt_ms"])
		fmt.Printf("Max RTT: %.2f ms\n", measurement.Metadata["max_rtt_ms"])
	} else {
		fmt.Printf("Error: %s\n", measurement.ErrorMessage)
	}
}

func TestTCPMonitor(t *testing.T) {
	config := map[string]interface{}{
		"timeout": float64(10),
	}

	monitor, err := monitors.NewTCPMonitor(config)
	if err != nil {
		t.Fatalf("Failed to create TCP monitor: %v", err)
	}

	port := 443
	target := &models.Target{
		Address: "1.1.1.1",
		Name:    "Cloudflare DNS",
		Port:    &port,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	measurement, err := monitor.Check(ctx, target)
	if err != nil {
		t.Fatalf("Check failed: %v", err)
	}

	fmt.Printf("\nTCP Test Results:\n")
	fmt.Printf("Target: %s:%d\n", target.Address, *target.Port)
	fmt.Printf("Success: %v\n", measurement.Success)
	if measurement.Success {
		fmt.Printf("Connect Time: %.2f ms\n", *measurement.LatencyMs)
		fmt.Printf("DNS Time: %.2f ms\n", measurement.Metadata["dns_time_ms"])
		fmt.Printf("Remote Addr: %s\n", measurement.Metadata["remote_addr"])
	} else {
		fmt.Printf("Error: %s\n", measurement.ErrorMessage)
	}
}
