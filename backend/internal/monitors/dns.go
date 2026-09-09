package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type DNSMonitor struct {
	config DNSConfig
}

type DNSConfig struct {
	RecordType string   // A, AAAA, CNAME, MX, NS, TXT, SOA, PTR
	Nameserver string   // Optional: specific nameserver to query
	Timeout    time.Duration
}

type DNSResult struct {
	Success       bool
	QueryTime     float64
	Answers       []string
	ResponseCode  string
	Nameserver    string
	ErrorMessage  string
	DNSSECEnabled bool
}

func NewDNSMonitor(config map[string]interface{}) (*DNSMonitor, error) {
	cfg := DNSConfig{
		RecordType: "A",
		Timeout:    10 * time.Second,
		Nameserver: "", // Use system default
	}

	if recordType, ok := config["record_type"].(string); ok {
		cfg.RecordType = recordType
	}
	if nameserver, ok := config["nameserver"].(string); ok {
		cfg.Nameserver = nameserver
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}

	return &DNSMonitor{config: cfg}, nil
}

func (m *DNSMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.query(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeDNS,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		measurement.LatencyMs = &result.QueryTime
		measurement.Metadata["answers"] = result.Answers
		measurement.Metadata["response_code"] = result.ResponseCode
		measurement.Metadata["nameserver"] = result.Nameserver
		measurement.Metadata["record_type"] = m.config.RecordType
		measurement.Metadata["dnssec_enabled"] = result.DNSSECEnabled
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *DNSMonitor) query(ctx context.Context, domain string) DNSResult {
	result := DNSResult{
		Success:    false,
		Nameserver: m.config.Nameserver,
	}

	// Create resolver
	resolver := &net.Resolver{}
	if m.config.Nameserver != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: m.config.Timeout}
				return d.DialContext(ctx, "udp", m.config.Nameserver+":53")
			},
		}
	}

	// Create timeout context
	queryCtx, cancel := context.WithTimeout(ctx, m.config.Timeout)
	defer cancel()

	start := time.Now()

	switch m.config.RecordType {
	case "A":
		ips, err := resolver.LookupIP(queryCtx, "ip4", domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("A lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		for _, ip := range ips {
			if ip.To4() != nil {
				result.Answers = append(result.Answers, ip.String())
			}
		}
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	case "AAAA":
		ips, err := resolver.LookupIP(queryCtx, "ip6", domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("AAAA lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		for _, ip := range ips {
			if ip.To16() != nil && ip.To4() == nil {
				result.Answers = append(result.Answers, ip.String())
			}
		}
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	case "CNAME":
		cname, err := resolver.LookupCNAME(queryCtx, domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("CNAME lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		result.Answers = []string{cname}
		result.Success = true
		result.ResponseCode = "NOERROR"

	case "MX":
		mxs, err := resolver.LookupMX(queryCtx, domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("MX lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		for _, mx := range mxs {
			result.Answers = append(result.Answers, fmt.Sprintf("%s (priority %d)", mx.Host, mx.Pref))
		}
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	case "NS":
		nss, err := resolver.LookupNS(queryCtx, domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("NS lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		for _, ns := range nss {
			result.Answers = append(result.Answers, ns.Host)
		}
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	case "TXT":
		txts, err := resolver.LookupTXT(queryCtx, domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("TXT lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		result.Answers = txts
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	case "PTR":
		names, err := resolver.LookupAddr(queryCtx, domain)
		result.QueryTime = float64(time.Since(start).Microseconds()) / 1000.0
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("PTR lookup failed: %v", err)
			result.ResponseCode = getDNSErrorCode(err)
			return result
		}
		result.Answers = names
		result.Success = len(result.Answers) > 0
		result.ResponseCode = "NOERROR"

	default:
		result.ErrorMessage = fmt.Sprintf("unsupported record type: %s", m.config.RecordType)
		return result
	}

	if result.Nameserver == "" {
		result.Nameserver = "system_default"
	}

	return result
}

func getDNSErrorCode(err error) string {
	if err == nil {
		return "NOERROR"
	}

	errStr := err.Error()
	if contains(errStr, "no such host") || contains(errStr, "NXDOMAIN") {
		return "NXDOMAIN"
	}
	if contains(errStr, "server misbehaving") || contains(errStr, "SERVFAIL") {
		return "SERVFAIL"
	}
	if contains(errStr, "connection refused") || contains(errStr, "REFUSED") {
		return "REFUSED"
	}
	if contains(errStr, "timeout") {
		return "TIMEOUT"
	}

	return "ERROR"
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

func (m *DNSMonitor) Type() string {
	return models.MonitorTypeDNS
}

func (m *DNSMonitor) ValidateConfig(config map[string]interface{}) error {
	if recordType, ok := config["record_type"].(string); ok {
		validTypes := map[string]bool{
			"A": true, "AAAA": true, "CNAME": true, "MX": true,
			"NS": true, "TXT": true, "SOA": true, "PTR": true,
		}
		if !validTypes[recordType] {
			return fmt.Errorf("invalid record_type: %s", recordType)
		}
	}
	return nil
}
