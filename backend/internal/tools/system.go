package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// SSTool wraps the ss command (socket statistics)
type SSTool struct {
	*BaseTool
}

func NewSSTool() *SSTool {
	return &SSTool{
		BaseTool: NewBaseTool(
			"ss",
			"ss",
			CategorySystem,
			"apt-get install iproute2",
			"Socket statistics and TCP connection analysis",
		),
	}
}

type SSConnection struct {
	State       string            `json:"state"`
	RecvQ       int               `json:"recv_q"`
	SendQ       int               `json:"send_q"`
	LocalAddr   string            `json:"local_addr"`
	LocalPort   int               `json:"local_port"`
	PeerAddr    string            `json:"peer_addr"`
	PeerPort    int               `json:"peer_port"`
	Process     string            `json:"process,omitempty"`
	TcpInfo     map[string]string `json:"tcp_info,omitempty"`
}

type SSResult struct {
	Connections []SSConnection `json:"connections"`
	Total       int            `json:"total"`
}

func (t *SSTool) ParseOutput(output string) (interface{}, error) {
	result := &SSResult{
		Connections: make([]SSConnection, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Netid") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		conn := SSConnection{
			TcpInfo: make(map[string]string),
		}

		// Parse: State Recv-Q Send-Q Local-Address:Port Peer-Address:Port
		if len(fields) >= 5 {
			conn.State = fields[0]
			conn.RecvQ, _ = strconv.Atoi(fields[1])
			conn.SendQ, _ = strconv.Atoi(fields[2])

			// Parse local address:port
			if parts := strings.Split(fields[3], ":"); len(parts) == 2 {
				conn.LocalAddr = parts[0]
				conn.LocalPort, _ = strconv.Atoi(parts[1])
			}

			// Parse peer address:port
			if parts := strings.Split(fields[4], ":"); len(parts) == 2 {
				conn.PeerAddr = parts[0]
				conn.PeerPort, _ = strconv.Atoi(parts[1])
			}
		}

		result.Connections = append(result.Connections, conn)
	}

	result.Total = len(result.Connections)
	return result, nil
}

func (t *SSTool) SS(ctx context.Context, protocol string, state string) (*SSResult, error) {
	args := []string{"-n"} // numeric, no name resolution

	if protocol != "" {
		args = append(args, "-"+protocol)
	}

	if state != "" {
		args = append(args, "state", state)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*SSResult), nil
}

// IPTool wraps the ip command
type IPTool struct {
	*BaseTool
}

func NewIPTool() *IPTool {
	return &IPTool{
		BaseTool: NewBaseTool(
			"ip",
			"ip",
			CategorySystem,
			"apt-get install iproute2",
			"Network configuration and routing analysis",
		),
	}
}

func (t *IPTool) ParseOutput(output string) (interface{}, error) {
	// Try to parse as JSON first
	if strings.HasPrefix(strings.TrimSpace(output), "[") {
		var result []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err == nil {
			return result, nil
		}
	}

	// Otherwise return raw output
	return map[string]interface{}{"raw": output}, nil
}

func (t *IPTool) IPAddr(ctx context.Context) ([]map[string]interface{}, error) {
	args := []string{"-j", "addr", "show"}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	if jsonResult, ok := parsed.([]map[string]interface{}); ok {
		return jsonResult, nil
	}

	return nil, fmt.Errorf("unexpected output format")
}

func (t *IPTool) IPRoute(ctx context.Context) ([]map[string]interface{}, error) {
	args := []string{"-j", "route", "show"}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	if jsonResult, ok := parsed.([]map[string]interface{}); ok {
		return jsonResult, nil
	}

	return nil, fmt.Errorf("unexpected output format")
}

func (t *IPTool) IPLink(ctx context.Context) ([]map[string]interface{}, error) {
	args := []string{"-j", "link", "show"}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	if jsonResult, ok := parsed.([]map[string]interface{}); ok {
		return jsonResult, nil
	}

	return nil, fmt.Errorf("unexpected output format")
}

// EthtoolTool wraps the ethtool command
type EthtoolTool struct {
	*BaseTool
}

func NewEthtoolTool() *EthtoolTool {
	return &EthtoolTool{
		BaseTool: NewBaseTool(
			"ethtool",
			"ethtool",
			CategorySystem,
			"apt-get install ethtool",
			"Ethernet device statistics and configuration",
		),
	}
}

type EthtoolStats struct {
	Interface string            `json:"interface"`
	Speed     string            `json:"speed"`
	Duplex    string            `json:"duplex"`
	LinkUp    bool              `json:"link_up"`
	Stats     map[string]int64  `json:"stats"`
}

func (t *EthtoolTool) ParseOutput(output string) (interface{}, error) {
	result := &EthtoolStats{
		Stats: make(map[string]int64),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse speed
		if strings.Contains(line, "Speed:") {
			speedRegex := regexp.MustCompile(`Speed:\s+(.+)`)
			if match := speedRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Speed = match[1]
			}
		}

		// Parse duplex
		if strings.Contains(line, "Duplex:") {
			duplexRegex := regexp.MustCompile(`Duplex:\s+(.+)`)
			if match := duplexRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Duplex = match[1]
			}
		}

		// Parse link status
		if strings.Contains(line, "Link detected:") {
			result.LinkUp = strings.Contains(line, "yes")
		}

		// Parse statistics (key: value format)
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				valueStr := strings.TrimSpace(parts[1])
				if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
					result.Stats[key] = value
				}
			}
		}
	}

	return result, nil
}

func (t *EthtoolTool) Ethtool(ctx context.Context, iface string) (*EthtoolStats, error) {
	args := []string{iface}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	stats := parsed.(*EthtoolStats)
	stats.Interface = iface
	return stats, nil
}

func (t *EthtoolTool) EthtoolStats(ctx context.Context, iface string) (*EthtoolStats, error) {
	args := []string{"-S", iface}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	stats := parsed.(*EthtoolStats)
	stats.Interface = iface
	return stats, nil
}

// ConntrackTool wraps the conntrack command
type ConntrackTool struct {
	*BaseTool
}

func NewConntrackTool() *ConntrackTool {
	return &ConntrackTool{
		BaseTool: NewBaseTool(
			"conntrack",
			"conntrack",
			CategorySystem,
			"apt-get install conntrack",
			"Connection tracking table analysis",
		),
	}
}

func (t *ConntrackTool) ParseOutput(output string) (interface{}, error) {
	connections := make([]map[string]string, 0)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		conn := make(map[string]string)
		conn["raw"] = line

		// Basic parsing of conntrack output
		// Format varies, typically: protocol src=x.x.x.x dst=y.y.y.y ...
		fields := strings.Fields(line)
		for _, field := range fields {
			if strings.Contains(field, "=") {
				parts := strings.SplitN(field, "=", 2)
				if len(parts) == 2 {
					conn[parts[0]] = parts[1]
				}
			}
		}

		connections = append(connections, conn)
	}

	return map[string]interface{}{
		"connections": connections,
		"count":       len(connections),
	}, nil
}

func (t *ConntrackTool) Conntrack(ctx context.Context, protocol string) (map[string]interface{}, error) {
	args := []string{"-L"}

	if protocol != "" {
		args = append(args, "-p", protocol)
	}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(map[string]interface{}), nil
}

// Register all system tools
func RegisterSystemTools(registry *ToolRegistry) {
	registry.Register(NewSSTool())
	registry.Register(NewIPTool())
	registry.Register(NewEthtoolTool())
	registry.Register(NewConntrackTool())
}
