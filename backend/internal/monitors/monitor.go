package monitors

import (
	"context"
	"fmt"

	"github.com/netmon/netmon/internal/models"
)

// Monitor interface that all monitoring plugins must implement
type Monitor interface {
	Check(ctx context.Context, target *models.Target) (*models.Measurement, error)
	Type() string
	ValidateConfig(config map[string]interface{}) error
}

// MonitorFactory creates monitors based on type
type MonitorFactory struct {
	creators map[string]func(map[string]interface{}) (Monitor, error)
}

func NewMonitorFactory() *MonitorFactory {
	factory := &MonitorFactory{
		creators: make(map[string]func(map[string]interface{}) (Monitor, error)),
	}

	// Register all monitors
	factory.Register(models.MonitorTypeICMP, func(config map[string]interface{}) (Monitor, error) {
		return NewICMPMonitor(config)
	})
	factory.Register(models.MonitorTypeTCP, func(config map[string]interface{}) (Monitor, error) {
		return NewTCPMonitor(config)
	})
	factory.Register(models.MonitorTypeUDP, func(config map[string]interface{}) (Monitor, error) {
		return NewUDPMonitor(config)
	})
	factory.Register(models.MonitorTypeDNS, func(config map[string]interface{}) (Monitor, error) {
		return NewDNSMonitor(config)
	})
	factory.Register(models.MonitorTypeHTTP, func(config map[string]interface{}) (Monitor, error) {
		return NewHTTPMonitor(config)
	})
	factory.Register(models.MonitorTypeHTTPS, func(config map[string]interface{}) (Monitor, error) {
		if config == nil {
			config = make(map[string]interface{})
		}
		config["validate_tls"] = true
		return NewHTTPMonitor(config)
	})
	factory.Register(models.MonitorTypeTraceroute, func(config map[string]interface{}) (Monitor, error) {
		return NewTracerouteMonitor(config)
	})
	factory.Register(models.MonitorTypeSNMP, func(config map[string]interface{}) (Monitor, error) {
		return NewSNMPMonitor(config)
	})
	factory.Register(models.MonitorTypeBandwidth, func(config map[string]interface{}) (Monitor, error) {
		return NewBandwidthMonitor(config)
	})
	factory.Register(models.MonitorTypeConnectivity, func(config map[string]interface{}) (Monitor, error) {
		return NewConnectivityMonitor(config)
	})
	factory.Register(models.MonitorTypeInterface, func(config map[string]interface{}) (Monitor, error) {
		return NewInterfaceMonitor(config)
	})

	return factory
}

func (f *MonitorFactory) Register(monitorType string, creator func(map[string]interface{}) (Monitor, error)) {
	f.creators[monitorType] = creator
}

func (f *MonitorFactory) Create(monitorType string, config map[string]interface{}) (Monitor, error) {
	creator, exists := f.creators[monitorType]
	if !exists {
		return nil, fmt.Errorf("unknown monitor type: %s", monitorType)
	}

	return creator(config)
}

func (f *MonitorFactory) SupportedTypes() []string {
	types := make([]string, 0, len(f.creators))
	for t := range f.creators {
		types = append(types, t)
	}
	return types
}
