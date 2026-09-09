package monitors

import (
	"context"
	"fmt"
	"time"

	"github.com/gosnmp/gosnmp"
	"github.com/netmon/netmon/internal/models"
)

type SNMPMonitor struct {
	config SNMPConfig
}

type SNMPConfig struct {
	Version        string            // v2c or v3
	Community      string            // For v2c
	Username       string            // For v3
	AuthPassword   string            // For v3
	PrivPassword   string            // For v3
	AuthProtocol   string            // MD5 or SHA (v3)
	PrivProtocol   string            // DES or AES (v3)
	SecurityLevel  string            // noAuthNoPriv, authNoPriv, authPriv (v3)
	Timeout        time.Duration
	Retries        int
	OIDs           []string          // OIDs to query
	OIDDescriptions map[string]string // Human-readable names for OIDs
}

type SNMPResult struct {
	Success      bool
	QueryTime    float64
	Values       map[string]interface{}
	ErrorMessage string
}

// Common SNMP OIDs
var CommonOIDs = map[string]string{
	// System Information
	"sysDescr":       "1.3.6.1.2.1.1.1.0",
	"sysUpTime":      "1.3.6.1.2.1.1.3.0",
	"sysContact":     "1.3.6.1.2.1.1.4.0",
	"sysName":        "1.3.6.1.2.1.1.5.0",
	"sysLocation":    "1.3.6.1.2.1.1.6.0",

	// Interface Statistics
	"ifNumber":       "1.3.6.1.2.1.2.1.0",
	"ifDescr":        "1.3.6.1.2.1.2.2.1.2",
	"ifType":         "1.3.6.1.2.1.2.2.1.3",
	"ifSpeed":        "1.3.6.1.2.1.2.2.1.5",
	"ifAdminStatus":  "1.3.6.1.2.1.2.2.1.7",
	"ifOperStatus":   "1.3.6.1.2.1.2.2.1.8",
	"ifInOctets":     "1.3.6.1.2.1.2.2.1.10",
	"ifOutOctets":    "1.3.6.1.2.1.2.2.1.16",
	"ifInErrors":     "1.3.6.1.2.1.2.2.1.14",
	"ifOutErrors":    "1.3.6.1.2.1.2.2.1.20",
	"ifInDiscards":   "1.3.6.1.2.1.2.2.1.13",
	"ifOutDiscards":  "1.3.6.1.2.1.2.2.1.19",

	// CPU and Memory
	"hrProcessorLoad": "1.3.6.1.2.1.25.3.3.1.2",
	"hrMemorySize":    "1.3.6.1.2.1.25.2.2.0",
	"hrStorageSize":   "1.3.6.1.2.1.25.2.3.1.5",
	"hrStorageUsed":   "1.3.6.1.2.1.25.2.3.1.6",

	// Network Statistics
	"ipInReceives":   "1.3.6.1.2.1.4.3.0",
	"ipInDelivers":   "1.3.6.1.2.1.4.9.0",
	"ipOutRequests":  "1.3.6.1.2.1.4.10.0",
	"icmpInMsgs":     "1.3.6.1.2.1.5.1.0",
	"icmpOutMsgs":    "1.3.6.1.2.1.5.14.0",
	"tcpActiveOpens": "1.3.6.1.2.1.6.5.0",
	"tcpCurrEstab":   "1.3.6.1.2.1.6.9.0",
}

func NewSNMPMonitor(config map[string]interface{}) (*SNMPMonitor, error) {
	cfg := SNMPConfig{
		Version:         "v2c",
		Community:       "public",
		Timeout:         10 * time.Second,
		Retries:         3,
		OIDs:            []string{},
		OIDDescriptions: make(map[string]string),
	}

	if version, ok := config["version"].(string); ok {
		cfg.Version = version
	}
	if community, ok := config["community"].(string); ok {
		cfg.Community = community
	}
	if username, ok := config["username"].(string); ok {
		cfg.Username = username
	}
	if authPass, ok := config["auth_password"].(string); ok {
		cfg.AuthPassword = authPass
	}
	if privPass, ok := config["priv_password"].(string); ok {
		cfg.PrivPassword = privPass
	}
	if authProto, ok := config["auth_protocol"].(string); ok {
		cfg.AuthProtocol = authProto
	}
	if privProto, ok := config["priv_protocol"].(string); ok {
		cfg.PrivProtocol = privProto
	}
	if secLevel, ok := config["security_level"].(string); ok {
		cfg.SecurityLevel = secLevel
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if retries, ok := config["retries"].(float64); ok {
		cfg.Retries = int(retries)
	}
	if oids, ok := config["oids"].([]interface{}); ok {
		for _, oid := range oids {
			if oidStr, ok := oid.(string); ok {
				cfg.OIDs = append(cfg.OIDs, oidStr)
			}
		}
	}

	return &SNMPMonitor{config: cfg}, nil
}

func (m *SNMPMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.query(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeSNMP,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.LatencyMs = &result.QueryTime
		measurement.Metadata["values"] = result.Values
		measurement.Metadata["oid_count"] = len(result.Values)
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *SNMPMonitor) query(ctx context.Context, address string) SNMPResult {
	result := SNMPResult{
		Success: false,
		Values:  make(map[string]interface{}),
	}

	// Create SNMP client
	params := &gosnmp.GoSNMP{
		Target:    address,
		Port:      161,
		Transport: "udp",
		Timeout:   m.config.Timeout,
		Retries:   m.config.Retries,
	}

	// Configure version-specific settings
	switch m.config.Version {
	case "v2c":
		params.Version = gosnmp.Version2c
		params.Community = m.config.Community

	case "v3":
		params.Version = gosnmp.Version3
		params.SecurityModel = gosnmp.UserSecurityModel

		msgFlags := gosnmp.NoAuthNoPriv
		authProtocol := gosnmp.NoAuth
		privProtocol := gosnmp.NoPriv

		// Set authentication
		if m.config.AuthPassword != "" {
			switch m.config.AuthProtocol {
			case "MD5":
				authProtocol = gosnmp.MD5
			case "SHA":
				authProtocol = gosnmp.SHA
			}
			msgFlags = gosnmp.AuthNoPriv
		}

		// Set privacy
		if m.config.PrivPassword != "" {
			switch m.config.PrivProtocol {
			case "DES":
				privProtocol = gosnmp.DES
			case "AES":
				privProtocol = gosnmp.AES
			}
			msgFlags = gosnmp.AuthPriv
		}

		params.MsgFlags = msgFlags
		params.SecurityParameters = &gosnmp.UsmSecurityParameters{
			UserName:                 m.config.Username,
			AuthenticationProtocol:   authProtocol,
			AuthenticationPassphrase: m.config.AuthPassword,
			PrivacyProtocol:         privProtocol,
			PrivacyPassphrase:       m.config.PrivPassword,
		}

	default:
		result.ErrorMessage = fmt.Sprintf("unsupported SNMP version: %s", m.config.Version)
		return result
	}

	// Connect
	err := params.Connect()
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("connection failed: %v", err)
		return result
	}
	defer params.Conn.Close()

	start := time.Now()

	// Query OIDs
	if len(m.config.OIDs) == 0 {
		// Default: query system information
		m.config.OIDs = []string{
			CommonOIDs["sysDescr"],
			CommonOIDs["sysUpTime"],
			CommonOIDs["sysName"],
		}
	}

	for _, oid := range m.config.OIDs {
		pdu, err := params.Get([]string{oid})
		if err != nil {
			result.Values[oid] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}

		if len(pdu.Variables) > 0 {
			variable := pdu.Variables[0]
			result.Values[oid] = m.parseVariable(variable)
		}
	}

	result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
	result.Success = len(result.Values) > 0

	if !result.Success {
		result.ErrorMessage = "no values retrieved"
	}

	return result
}

func (m *SNMPMonitor) parseVariable(variable gosnmp.SnmpPDU) interface{} {
	switch variable.Type {
	case gosnmp.OctetString:
		return string(variable.Value.([]byte))
	case gosnmp.Integer:
		return variable.Value
	case gosnmp.Counter32, gosnmp.Counter64:
		return variable.Value
	case gosnmp.Gauge32:
		return variable.Value
	case gosnmp.TimeTicks:
		ticks := variable.Value.(uint32)
		return map[string]interface{}{
			"ticks":   ticks,
			"seconds": ticks / 100,
			"uptime":  formatUptime(ticks),
		}
	case gosnmp.IPAddress:
		return variable.Value
	case gosnmp.Null:
		return nil
	default:
		return fmt.Sprintf("%v", variable.Value)
	}
}

func formatUptime(ticks uint32) string {
	seconds := ticks / 100
	days := seconds / 86400
	hours := (seconds % 86400) / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, secs)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}

func (m *SNMPMonitor) Type() string {
	return models.MonitorTypeSNMP
}

func (m *SNMPMonitor) ValidateConfig(config map[string]interface{}) error {
	if version, ok := config["version"].(string); ok {
		if version != "v2c" && version != "v3" {
			return fmt.Errorf("version must be 'v2c' or 'v3'")
		}

		if version == "v3" {
			if _, ok := config["username"].(string); !ok {
				return fmt.Errorf("username required for SNMPv3")
			}
		}
	}

	return nil
}

// GetInterfaceStats queries interface statistics for all interfaces
func (m *SNMPMonitor) GetInterfaceStats(address string) ([]map[string]interface{}, error) {
	// This would be called separately to get detailed interface stats
	// Implementation would walk the interface table
	return nil, fmt.Errorf("not implemented yet")
}
