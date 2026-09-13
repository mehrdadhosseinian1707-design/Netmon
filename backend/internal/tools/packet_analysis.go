package tools

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

// Hping3Tool wraps the hping3 command
type Hping3Tool struct {
	*BaseTool
}

func NewHping3Tool() *Hping3Tool {
	return &Hping3Tool{
		BaseTool: NewBaseTool(
			"hping3",
			"hping3",
			CategoryPacketAnalysis,
			"apt-get install hping3",
			"Advanced TCP/UDP/ICMP packet crafting and analysis",
		),
	}
}

type Hping3Result struct {
	Target       string  `json:"target"`
	PacketsSent  int     `json:"packets_sent"`
	PacketsRecv  int     `json:"packets_received"`
	PacketLoss   float64 `json:"packet_loss_percent"`
	MinRTT       float64 `json:"min_rtt_ms"`
	AvgRTT       float64 `json:"avg_rtt_ms"`
	MaxRTT       float64 `json:"max_rtt_ms"`
}

func (t *Hping3Tool) ParseOutput(output string) (interface{}, error) {
	result := &Hping3Result{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Parse statistics line
		// "--- hping statistic ---"
		// "5 packets transmitted, 5 packets received, 0% packet loss"
		// "round-trip min/avg/max = 0.3/0.5/0.8 ms"

		if strings.Contains(line, "packets transmitted") {
			statsRegex := regexp.MustCompile(`(\d+)\s+packets transmitted,\s+(\d+)\s+packets received,\s+([\d.]+)%`)
			if match := statsRegex.FindStringSubmatch(line); len(match) > 3 {
				result.PacketsSent, _ = strconv.Atoi(match[1])
				result.PacketsRecv, _ = strconv.Atoi(match[2])
				result.PacketLoss, _ = strconv.ParseFloat(match[3], 64)
			}
		}

		if strings.Contains(line, "round-trip min/avg/max") {
			rttRegex := regexp.MustCompile(`min/avg/max\s+=\s+([\d.]+)/([\d.]+)/([\d.]+)`)
			if match := rttRegex.FindStringSubmatch(line); len(match) > 3 {
				result.MinRTT, _ = strconv.ParseFloat(match[1], 64)
				result.AvgRTT, _ = strconv.ParseFloat(match[2], 64)
				result.MaxRTT, _ = strconv.ParseFloat(match[3], 64)
			}
		}
	}

	return result, nil
}

func (t *Hping3Tool) Hping3(ctx context.Context, target string, count int, protocol string, port int) (*Hping3Result, error) {
	args := []string{"-c", strconv.Itoa(count)}

	switch protocol {
	case "tcp":
		args = append(args, "-S")
	case "udp":
		args = append(args, "--udp")
	case "icmp":
		args = append(args, "--icmp")
	}

	if port > 0 {
		args = append(args, "-p", strconv.Itoa(port))
	}

	args = append(args, target)

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	hpingResult := parsed.(*Hping3Result)
	hpingResult.Target = target
	return hpingResult, nil
}

// NpingTool wraps the nping command from nmap
type NpingTool struct {
	*BaseTool
}

func NewNpingTool() *NpingTool {
	return &NpingTool{
		BaseTool: NewBaseTool(
			"nping",
			"nping",
			CategoryPacketAnalysis,
			"apt-get install nmap",
			"Network packet generation and analysis tool",
		),
	}
}

type NpingResult struct {
	Target      string  `json:"target"`
	Protocol    string  `json:"protocol"`
	PacketsSent int     `json:"packets_sent"`
	PacketsRecv int     `json:"packets_received"`
	LossRate    float64 `json:"loss_rate_percent"`
	AvgRTT      float64 `json:"avg_rtt_ms"`
}

func (t *NpingTool) ParseOutput(output string) (interface{}, error) {
	result := &NpingResult{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Parse statistics
		if strings.Contains(line, "Raw packets sent:") {
			sentRegex := regexp.MustCompile(`sent:\s+(\d+)`)
			if match := sentRegex.FindStringSubmatch(line); len(match) > 1 {
				result.PacketsSent, _ = strconv.Atoi(match[1])
			}
		}

		if strings.Contains(line, "Rcvd:") {
			rcvdRegex := regexp.MustCompile(`Rcvd:\s+(\d+)`)
			if match := rcvdRegex.FindStringSubmatch(line); len(match) > 1 {
				result.PacketsRecv, _ = strconv.Atoi(match[1])
			}
		}

		if strings.Contains(line, "Lost:") {
			lostRegex := regexp.MustCompile(`Lost:\s+\d+\s+\(([\d.]+)%\)`)
			if match := lostRegex.FindStringSubmatch(line); len(match) > 1 {
				result.LossRate, _ = strconv.ParseFloat(match[1], 64)
			}
		}
	}

	return result, nil
}

func (t *NpingTool) Nping(ctx context.Context, target string, count int, protocol string, port int) (*NpingResult, error) {
	args := []string{"-c", strconv.Itoa(count)}

	switch protocol {
	case "tcp":
		args = append(args, "--tcp")
	case "udp":
		args = append(args, "--udp")
	case "icmp":
		args = append(args, "--icmp")
	}

	if port > 0 {
		args = append(args, "-p", strconv.Itoa(port))
	}

	args = append(args, target)

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	npingResult := parsed.(*NpingResult)
	npingResult.Target = target
	npingResult.Protocol = protocol
	return npingResult, nil
}

// Register all packet analysis tools
func RegisterPacketAnalysisTools(registry *ToolRegistry) {
	registry.Register(NewHping3Tool())
	registry.Register(NewNpingTool())
}
