package monitors

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type InterfaceMonitor struct {
	config InterfaceConfig
}

type InterfaceConfig struct {
	InterfaceName string
	Timeout       time.Duration
}

type InterfaceResult struct {
	Success       bool
	RXErrors      int64
	TXErrors      int64
	RXDrops       int64
	TXDrops       int64
	RXBytes       int64
	TXBytes       int64
	RXPackets     int64
	TXPackets     int64
	ErrorMessage  string
}

func NewInterfaceMonitor(config map[string]interface{}) (*InterfaceMonitor, error) {
	cfg := InterfaceConfig{
		InterfaceName: "",
		Timeout:       10 * time.Second,
	}

	if ifName, ok := config["interface_name"].(string); ok {
		cfg.InterfaceName = ifName
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}

	return &InterfaceMonitor{config: cfg}, nil
}

func (m *InterfaceMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	// If interface name not specified in config, try to use target address
	interfaceName := m.config.InterfaceName
	if interfaceName == "" {
		// Try to determine interface from target address
		interfaceName = m.getInterfaceForTarget(target.Address)
	}

	result := m.getInterfaceStats(ctx, interfaceName)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeInterface,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.RXErrors = &result.RXErrors
		measurement.TXErrors = &result.TXErrors
		measurement.RXDrops = &result.RXDrops
		measurement.TXDrops = &result.TXDrops

		measurement.Metadata["interface_name"] = interfaceName
		measurement.Metadata["rx_bytes"] = result.RXBytes
		measurement.Metadata["tx_bytes"] = result.TXBytes
		measurement.Metadata["rx_packets"] = result.RXPackets
		measurement.Metadata["tx_packets"] = result.TXPackets
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *InterfaceMonitor) getInterfaceStats(ctx context.Context, interfaceName string) InterfaceResult {
	result := InterfaceResult{
		Success: false,
	}

	if interfaceName == "" {
		result.ErrorMessage = "no interface name specified"
		return result
	}

	// Get network interface by name
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to get interface: %v", err)
		return result
	}

	// Platform-specific statistics retrieval
	stats, err := m.getInterfaceStatsPlatform(iface)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to get stats: %v", err)
		return result
	}

	result.RXErrors = stats.RXErrors
	result.TXErrors = stats.TXErrors
	result.RXDrops = stats.RXDrops
	result.TXDrops = stats.TXDrops
	result.RXBytes = stats.RXBytes
	result.TXBytes = stats.TXBytes
	result.RXPackets = stats.RXPackets
	result.TXPackets = stats.TXPackets
	result.Success = true

	return result
}

func (m *InterfaceMonitor) getInterfaceStatsPlatform(iface *net.Interface) (*InterfaceResult, error) {
	// Platform-specific implementation
	// This is a simplified version - full implementation would use:
	// - Linux: /proc/net/dev or netlink
	// - Windows: Performance Counters or WMI
	// - macOS/BSD: sysctl

	switch runtime.GOOS {
	case "linux":
		return m.getInterfaceStatsLinux(iface)
	case "windows":
		return m.getInterfaceStatsWindows(iface)
	case "darwin":
		return m.getInterfaceStatsDarwin(iface)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func (m *InterfaceMonitor) getInterfaceStatsLinux(iface *net.Interface) (*InterfaceResult, error) {
	// Read from /proc/net/dev
	// Format: Interface|RX bytes packets errs drop|TX bytes packets errs drop
	// This is a placeholder - actual implementation would parse /proc/net/dev

	result := &InterfaceResult{
		RXErrors:  0,
		TXErrors:  0,
		RXDrops:   0,
		TXDrops:   0,
		RXBytes:   0,
		TXBytes:   0,
		RXPackets: 0,
		TXPackets: 0,
	}

	// Placeholder: Would read from /proc/net/dev
	// Example line: "eth0: 1234567 8901 0 0 0 0 0 0 2345678 9012 0 0 0 0 0 0"

	return result, nil
}

func (m *InterfaceMonitor) getInterfaceStatsWindows(iface *net.Interface) (*InterfaceResult, error) {
	// Use Windows Performance Counters or WMI
	// This is a placeholder - actual implementation would use syscalls or WMI

	result := &InterfaceResult{
		RXErrors:  0,
		TXErrors:  0,
		RXDrops:   0,
		TXDrops:   0,
		RXBytes:   0,
		TXBytes:   0,
		RXPackets: 0,
		TXPackets: 0,
	}

	// Placeholder: Would use Windows API

	return result, nil
}

func (m *InterfaceMonitor) getInterfaceStatsDarwin(iface *net.Interface) (*InterfaceResult, error) {
	// Use sysctl on macOS/BSD
	// This is a placeholder - actual implementation would use syscalls

	result := &InterfaceResult{
		RXErrors:  0,
		TXErrors:  0,
		RXDrops:   0,
		TXDrops:   0,
		RXBytes:   0,
		TXBytes:   0,
		RXPackets: 0,
		TXPackets: 0,
	}

	// Placeholder: Would use sysctl

	return result, nil
}

func (m *InterfaceMonitor) getInterfaceForTarget(targetAddr string) string {
	// Try to determine which interface would be used to reach target
	// This is a simplified approach

	// Parse target address
	host := targetAddr
	if h, _, err := net.SplitHostPort(targetAddr); err == nil {
		host = h
	}

	// Resolve to IP
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return ""
	}

	// Get all interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// Try to find interface with route to target
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check if any address is on same network as target
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				if ipnet.Contains(ips[0]) {
					return iface.Name
				}
			}
		}
	}

	// If no specific match, return first non-loopback interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface.Name
		}
	}

	return ""
}

func (m *InterfaceMonitor) Type() string {
	return models.MonitorTypeInterface
}

func (m *InterfaceMonitor) ValidateConfig(config map[string]interface{}) error {
	if timeout, ok := config["timeout"].(float64); ok {
		if timeout < 1 || timeout > 300 {
			return fmt.Errorf("timeout must be between 1 and 300 seconds")
		}
	}
	return nil
}
