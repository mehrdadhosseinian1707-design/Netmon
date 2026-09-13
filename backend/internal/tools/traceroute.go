package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TracerouteTool wraps the traceroute command
type TracerouteTool struct {
	*BaseTool
}

func NewTracerouteTool() *TracerouteTool {
	return &TracerouteTool{
		BaseTool: NewBaseTool(
			"traceroute",
			"traceroute",
			CategoryTraceroute,
			"apt-get install traceroute",
			"Standard traceroute for network path discovery",
		),
	}
}

type TracerouteHop struct {
	Hop      int      `json:"hop"`
	IP       string   `json:"ip"`
	Hostname string   `json:"hostname,omitempty"`
	RTT1     *float64 `json:"rtt1_ms,omitempty"`
	RTT2     *float64 `json:"rtt2_ms,omitempty"`
	RTT3     *float64 `json:"rtt3_ms,omitempty"`
	Timeout  bool     `json:"timeout"`
}

type TracerouteResult struct {
	Target   string          `json:"target"`
	Hops     []TracerouteHop `json:"hops"`
	MaxHops  int             `json:"max_hops"`
	Complete bool            `json:"complete"`
}

func (t *TracerouteTool) ParseOutput(output string) (interface{}, error) {
	result := &TracerouteResult{
		Hops: make([]TracerouteHop, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "traceroute to") {
			continue
		}

		// Parse hop line
		hop := TracerouteHop{}

		// Extract hop number
		hopRegex := regexp.MustCompile(`^\s*(\d+)\s+`)
		if match := hopRegex.FindStringSubmatch(line); len(match) > 1 {
			hop.Hop, _ = strconv.Atoi(match[1])
		} else {
			continue
		}

		// Check for timeout
		if strings.Contains(line, "* * *") {
			hop.Timeout = true
			result.Hops = append(result.Hops, hop)
			continue
		}

		// Extract hostname/IP and RTTs
		// Format: "1  router.local (192.168.1.1)  0.345 ms  0.312 ms  0.298 ms"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		// Extract hostname and IP
		for i := 1; i < len(parts); i++ {
			if strings.Contains(parts[i], "(") && strings.Contains(parts[i], ")") {
				// Format: (192.168.1.1)
				hop.IP = strings.Trim(parts[i], "()")
				if i > 1 {
					hop.Hostname = parts[i-1]
				}
				break
			} else if regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`).MatchString(parts[i]) {
				hop.IP = parts[i]
				break
			}
		}

		// Extract RTT values
		rttRegex := regexp.MustCompile(`([\d.]+)\s+ms`)
		matches := rttRegex.FindAllStringSubmatch(line, -1)
		if len(matches) > 0 {
			if len(matches) > 0 {
				val, _ := strconv.ParseFloat(matches[0][1], 64)
				hop.RTT1 = &val
			}
			if len(matches) > 1 {
				val, _ := strconv.ParseFloat(matches[1][1], 64)
				hop.RTT2 = &val
			}
			if len(matches) > 2 {
				val, _ := strconv.ParseFloat(matches[2][1], 64)
				hop.RTT3 = &val
			}
		}

		result.Hops = append(result.Hops, hop)
	}

	if len(result.Hops) > 0 {
		result.Complete = !result.Hops[len(result.Hops)-1].Timeout
	}

	return result, nil
}

func (t *TracerouteTool) Traceroute(ctx context.Context, host string, maxHops int) (*TracerouteResult, error) {
	args := []string{"-m", strconv.Itoa(maxHops), host}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TracerouteResult), nil
}

// TracepathTool wraps the tracepath command
type TracepathTool struct {
	*BaseTool
}

func NewTracepathTool() *TracepathTool {
	return &TracepathTool{
		BaseTool: NewBaseTool(
			"tracepath",
			"tracepath",
			CategoryTraceroute,
			"apt-get install iputils-tracepath",
			"Traceroute with automatic MTU discovery",
		),
	}
}

type TracepathHop struct {
	Hop      int     `json:"hop"`
	IP       string  `json:"ip"`
	Hostname string  `json:"hostname,omitempty"`
	RTT      float64 `json:"rtt_ms"`
	PMTU     int     `json:"pmtu,omitempty"`
	Asymm    int     `json:"asymm,omitempty"`
}

type TracepathResult struct {
	Target string         `json:"target"`
	Hops   []TracepathHop `json:"hops"`
	MTU    int            `json:"mtu"`
}

func (t *TracepathTool) ParseOutput(output string) (interface{}, error) {
	result := &TracepathResult{
		Hops: make([]TracepathHop, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse MTU line: "     pmtu 1500"
		if strings.Contains(line, "pmtu") {
			mtuRegex := regexp.MustCompile(`pmtu\s+(\d+)`)
			if match := mtuRegex.FindStringSubmatch(line); len(match) > 1 {
				result.MTU, _ = strconv.Atoi(match[1])
			}
			continue
		}

		// Parse hop line: " 1:  router.local  0.345ms"
		hopRegex := regexp.MustCompile(`^\s*(\d+):\s+`)
		match := hopRegex.FindStringSubmatch(line)
		if len(match) < 2 {
			continue
		}

		hop := TracepathHop{}
		hop.Hop, _ = strconv.Atoi(match[1])

		// Extract IP/hostname and RTT
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			hop.Hostname = parts[1]
			// Check if there's an IP in parentheses
			for _, part := range parts {
				if strings.Contains(part, "(") && strings.Contains(part, ")") {
					hop.IP = strings.Trim(part, "()")
				}
			}
		}

		// Extract RTT
		rttRegex := regexp.MustCompile(`([\d.]+)ms`)
		if rttMatch := rttRegex.FindStringSubmatch(line); len(rttMatch) > 1 {
			hop.RTT, _ = strconv.ParseFloat(rttMatch[1], 64)
		}

		// Extract asymm value
		if strings.Contains(line, "asymm") {
			asymmRegex := regexp.MustCompile(`asymm\s+(\d+)`)
			if asymmMatch := asymmRegex.FindStringSubmatch(line); len(asymmMatch) > 1 {
				hop.Asymm, _ = strconv.Atoi(asymmMatch[1])
			}
		}

		result.Hops = append(result.Hops, hop)
	}

	return result, nil
}

func (t *TracepathTool) Tracepath(ctx context.Context, host string) (*TracepathResult, error) {
	args := []string{host}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TracepathResult), nil
}

// TCPTracerouteTool wraps the tcptraceroute command
type TCPTracerouteTool struct {
	*BaseTool
}

func NewTCPTracerouteTool() *TCPTracerouteTool {
	return &TCPTracerouteTool{
		BaseTool: NewBaseTool(
			"tcptraceroute",
			"tcptraceroute",
			CategoryTraceroute,
			"apt-get install tcptraceroute",
			"TCP-based traceroute for firewall traversal",
		),
	}
}

func (t *TCPTracerouteTool) ParseOutput(output string) (interface{}, error) {
	// Similar to TracerouteTool parsing
	tool := NewTracerouteTool()
	return tool.ParseOutput(output)
}

func (t *TCPTracerouteTool) TCPTraceroute(ctx context.Context, host string, port int, maxHops int) (*TracerouteResult, error) {
	args := []string{"-m", strconv.Itoa(maxHops), host, strconv.Itoa(port)}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TracerouteResult), nil
}

// DublinTracerouteTool wraps the dublin-traceroute command
type DublinTracerouteTool struct {
	*BaseTool
}

func NewDublinTracerouteTool() *DublinTracerouteTool {
	return &DublinTracerouteTool{
		BaseTool: NewBaseTool(
			"dublin-traceroute",
			"dublin-traceroute",
			CategoryTraceroute,
			"apt-get install dublin-traceroute",
			"Multi-path traceroute using UDP/TCP/ICMP with NAT detection",
		),
	}
}

func (t *DublinTracerouteTool) ParseOutput(output string) (interface{}, error) {
	// Dublin-traceroute outputs JSON by default
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return nil, fmt.Errorf("failed to parse dublin-traceroute JSON: %w", err)
	}
	return result, nil
}

func (t *DublinTracerouteTool) DublinTraceroute(ctx context.Context, host string) (map[string]interface{}, error) {
	args := []string{host}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(map[string]interface{}), nil
}

// ParisTracerouteTool wraps paris-traceroute command
type ParisTracerouteTool struct {
	*BaseTool
}

func NewParisTracerouteTool() *ParisTracerouteTool {
	return &ParisTracerouteTool{
		BaseTool: NewBaseTool(
			"paris-traceroute",
			"paris-traceroute",
			CategoryTraceroute,
			"apt-get install paris-traceroute",
			"Flow-aware traceroute to detect load-balanced paths",
		),
	}
}

func (t *ParisTracerouteTool) ParseOutput(output string) (interface{}, error) {
	tool := NewTracerouteTool()
	return tool.ParseOutput(output)
}

func (t *ParisTracerouteTool) ParisTraceroute(ctx context.Context, host string, maxHops int) (*TracerouteResult, error) {
	args := []string{"-m", strconv.Itoa(maxHops), host}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TracerouteResult), nil
}

// ScamperTool wraps the scamper command for advanced measurements
type ScamperTool struct {
	*BaseTool
}

func NewScamperTool() *ScamperTool {
	return &ScamperTool{
		BaseTool: NewBaseTool(
			"scamper",
			"scamper",
			CategoryTraceroute,
			"apt-get install scamper",
			"Advanced Internet measurement tool with multiple techniques",
		),
	}
}

func (t *ScamperTool) ParseOutput(output string) (interface{}, error) {
	// Scamper can output in various formats (JSON, warts, text)
	// For JSON output, parse directly
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err == nil {
			return result, nil
		}
	}

	// Otherwise return raw output
	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

func (t *ScamperTool) Scamper(ctx context.Context, command string, target string) (map[string]interface{}, error) {
	args := []string{"-i", target, "-c", command}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(map[string]interface{}), nil
}

// Register all traceroute tools
func RegisterTracerouteTools(registry *ToolRegistry) {
	registry.Register(NewTracerouteTool())
	registry.Register(NewTracepathTool())
	registry.Register(NewTCPTracerouteTool())
	registry.Register(NewDublinTracerouteTool())
	registry.Register(NewParisTracerouteTool())
	registry.Register(NewScamperTool())
}
