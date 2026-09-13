package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// PingTool wraps the ping command
type PingTool struct {
	*BaseTool
}

func NewPingTool() *PingTool {
	return &PingTool{
		BaseTool: NewBaseTool(
			"ping",
			"ping",
			CategoryICMP,
			"apt-get install iputils-ping",
			"ICMP echo request tool for latency and reachability testing",
		),
	}
}

type PingResult struct {
	Host           string  `json:"host"`
	PacketsSent    int     `json:"packets_sent"`
	PacketsReceived int    `json:"packets_received"`
	PacketLoss     float64 `json:"packet_loss"`
	MinRTT         float64 `json:"min_rtt_ms"`
	AvgRTT         float64 `json:"avg_rtt_ms"`
	MaxRTT         float64 `json:"max_rtt_ms"`
	StdDevRTT      float64 `json:"stddev_rtt_ms"`
}

func (t *PingTool) ParseOutput(output string) (interface{}, error) {
	result := &PingResult{}

	// Extract host
	hostRegex := regexp.MustCompile(`PING\s+([^\s]+)\s+`)
	if match := hostRegex.FindStringSubmatch(output); len(match) > 1 {
		result.Host = match[1]
	}

	// Extract packet statistics
	statsRegex := regexp.MustCompile(`(\d+)\s+packets transmitted,\s+(\d+)\s+received,\s+(\d+(?:\.\d+)?|\d+)%\s+packet loss`)
	if match := statsRegex.FindStringSubmatch(output); len(match) > 3 {
		result.PacketsSent, _ = strconv.Atoi(match[1])
		result.PacketsReceived, _ = strconv.Atoi(match[2])
		result.PacketLoss, _ = strconv.ParseFloat(match[3], 64)
	}

	// Extract RTT statistics
	rttRegex := regexp.MustCompile(`rtt min/avg/max/mdev = ([\d.]+)/([\d.]+)/([\d.]+)/([\d.]+)`)
	if match := rttRegex.FindStringSubmatch(output); len(match) > 4 {
		result.MinRTT, _ = strconv.ParseFloat(match[1], 64)
		result.AvgRTT, _ = strconv.ParseFloat(match[2], 64)
		result.MaxRTT, _ = strconv.ParseFloat(match[3], 64)
		result.StdDevRTT, _ = strconv.ParseFloat(match[4], 64)
	}

	return result, nil
}

func (t *PingTool) Ping(ctx context.Context, host string, count int) (*PingResult, error) {
	args := []string{"-c", strconv.Itoa(count), host}
	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*PingResult), nil
}

// FPingTool wraps the fping command for high-volume parallel ICMP testing
type FPingTool struct {
	*BaseTool
}

func NewFPingTool() *FPingTool {
	return &FPingTool{
		BaseTool: NewBaseTool(
			"fping",
			"fping",
			CategoryICMP,
			"apt-get install fping",
			"High-performance parallel ICMP ping tool",
		),
	}
}

type FPingResult struct {
	Host      string  `json:"host"`
	Alive     bool    `json:"alive"`
	MinRTT    float64 `json:"min_rtt_ms"`
	AvgRTT    float64 `json:"avg_rtt_ms"`
	MaxRTT    float64 `json:"max_rtt_ms"`
	Sent      int     `json:"sent"`
	Received  int     `json:"received"`
	LossRate  float64 `json:"loss_rate"`
}

func (t *FPingTool) ParseOutput(output string) (interface{}, error) {
	results := make([]FPingResult, 0)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		result := FPingResult{}

		// Parse alive/unreachable
		if strings.Contains(line, "is alive") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				result.Host = parts[0]
				result.Alive = true
			}
		} else if strings.Contains(line, "is unreachable") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				result.Host = parts[0]
				result.Alive = false
			}
		}

		// Parse detailed stats: host : xmt/rcv/%loss = 4/4/0%, min/avg/max = 0.04/0.06/0.08
		statsRegex := regexp.MustCompile(`([^\s:]+)\s*:\s*xmt/rcv/%loss\s*=\s*(\d+)/(\d+)/(\d+)%,\s*min/avg/max\s*=\s*([\d.]+)/([\d.]+)/([\d.]+)`)
		if match := statsRegex.FindStringSubmatch(line); len(match) > 7 {
			result.Host = match[1]
			result.Sent, _ = strconv.Atoi(match[2])
			result.Received, _ = strconv.Atoi(match[3])
			result.LossRate, _ = strconv.ParseFloat(match[4], 64)
			result.MinRTT, _ = strconv.ParseFloat(match[5], 64)
			result.AvgRTT, _ = strconv.ParseFloat(match[6], 64)
			result.MaxRTT, _ = strconv.ParseFloat(match[7], 64)
			result.Alive = result.Received > 0
		}

		if result.Host != "" {
			results = append(results, result)
		}
	}

	return results, nil
}

func (t *FPingTool) FPing(ctx context.Context, hosts []string, count int) ([]FPingResult, error) {
	args := []string{"-c", strconv.Itoa(count), "-q"}
	args = append(args, hosts...)

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.([]FPingResult), nil
}

// ArpingTool wraps the arping command
type ArpingTool struct {
	*BaseTool
}

func NewArpingTool() *ArpingTool {
	return &ArpingTool{
		BaseTool: NewBaseTool(
			"arping",
			"arping",
			CategoryICMP,
			"apt-get install arping",
			"ARP-level reachability testing",
		),
	}
}

type ArpingResult struct {
	Host         string  `json:"host"`
	MACAddress   string  `json:"mac_address"`
	PacketsSent  int     `json:"packets_sent"`
	PacketsRecv  int     `json:"packets_received"`
	AvgRTT       float64 `json:"avg_rtt_ms"`
}

func (t *ArpingTool) ParseOutput(output string) (interface{}, error) {
	result := &ArpingResult{}

	// Extract MAC address
	macRegex := regexp.MustCompile(`\[([0-9A-Fa-f:]{17})\]`)
	if match := macRegex.FindStringSubmatch(output); len(match) > 1 {
		result.MACAddress = match[1]
	}

	// Extract statistics
	statsRegex := regexp.MustCompile(`Sent\s+(\d+)\s+probes.*Received\s+(\d+)\s+response`)
	if match := statsRegex.FindStringSubmatch(output); len(match) > 2 {
		result.PacketsSent, _ = strconv.Atoi(match[1])
		result.PacketsRecv, _ = strconv.Atoi(match[2])
	}

	return result, nil
}

func (t *ArpingTool) Arping(ctx context.Context, host string, count int, iface string) (*ArpingResult, error) {
	args := []string{"-c", strconv.Itoa(count)}
	if iface != "" {
		args = append(args, "-I", iface)
	}
	args = append(args, host)

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*ArpingResult), nil
}

// MTRTool wraps the mtr command for continuous network diagnostics
type MTRTool struct {
	*BaseTool
}

func NewMTRTool() *MTRTool {
	return &MTRTool{
		BaseTool: NewBaseTool(
			"mtr",
			"mtr",
			CategoryTraceroute,
			"apt-get install mtr-tiny",
			"Continuous traceroute with packet loss per hop",
		),
	}
}

type MTRHop struct {
	Hop        int     `json:"hop"`
	Host       string  `json:"host"`
	IP         string  `json:"ip"`
	Loss       float64 `json:"loss_percent"`
	Sent       int     `json:"sent"`
	Last       float64 `json:"last_ms"`
	Avg        float64 `json:"avg_ms"`
	Best       float64 `json:"best_ms"`
	Worst      float64 `json:"worst_ms"`
	StdDev     float64 `json:"stddev_ms"`
}

type MTRResult struct {
	Target string   `json:"target"`
	Hops   []MTRHop `json:"hops"`
}

func (t *MTRTool) ParseOutput(output string) (interface{}, error) {
	result := &MTRResult{
		Hops: make([]MTRHop, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "HOST:") || strings.HasPrefix(line, "Start:") {
			continue
		}

		// Parse hop line format: " 1.|-- host (ip) Loss% Snt Last Avg Best Wrst StDev"
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		hop := MTRHop{}

		// Extract hop number
		hopStr := strings.TrimRight(fields[0], ".|--")
		hop.Hop, _ = strconv.Atoi(hopStr)

		// Extract hostname and IP
		hop.Host = fields[1]
		if strings.Contains(fields[1], "(") {
			parts := strings.Split(fields[1], "(")
			hop.Host = parts[0]
			hop.IP = strings.TrimRight(parts[1], ")")
		}

		// Extract statistics
		if len(fields) >= 8 {
			hop.Loss, _ = strconv.ParseFloat(strings.TrimSuffix(fields[2], "%"), 64)
			hop.Sent, _ = strconv.Atoi(fields[3])
			hop.Last, _ = strconv.ParseFloat(fields[4], 64)
			hop.Avg, _ = strconv.ParseFloat(fields[5], 64)
			hop.Best, _ = strconv.ParseFloat(fields[6], 64)
			hop.Worst, _ = strconv.ParseFloat(fields[7], 64)
			if len(fields) >= 9 {
				hop.StdDev, _ = strconv.ParseFloat(fields[8], 64)
			}
		}

		result.Hops = append(result.Hops, hop)
	}

	return result, nil
}

func (t *MTRTool) MTR(ctx context.Context, host string, count int, reportMode bool) (*MTRResult, error) {
	args := []string{"--report", "--report-cycles", strconv.Itoa(count), host}
	if reportMode {
		args = append([]string{"--no-dns"}, args...)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*MTRResult), nil
}

func (t *MTRTool) MTRJSON(ctx context.Context, host string, count int) (map[string]interface{}, error) {
	args := []string{"--json", "--report-cycles", strconv.Itoa(count), host}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &jsonResult); err != nil {
		return nil, fmt.Errorf("failed to parse JSON output: %w", err)
	}

	return jsonResult, nil
}

// Register all ICMP tools
func RegisterICMPTools(registry *ToolRegistry) {
	registry.Register(NewPingTool())
	registry.Register(NewFPingTool())
	registry.Register(NewArpingTool())
	registry.Register(NewMTRTool())
}
