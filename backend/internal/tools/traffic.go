package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// IftopTool wraps the iftop command for real-time bandwidth monitoring
type IftopTool struct {
	*BaseTool
}

func NewIftopTool() *IftopTool {
	return &IftopTool{
		BaseTool: NewBaseTool(
			"iftop",
			"iftop",
			CategorySystem,
			"apt-get install iftop",
			"Real-time interface bandwidth monitoring",
		),
	}
}

type IftopResult struct {
	Interface      string             `json:"interface"`
	TotalRX        float64            `json:"total_rx_mbps"`
	TotalTX        float64            `json:"total_tx_mbps"`
	PeakRX         float64            `json:"peak_rx_mbps"`
	PeakTX         float64            `json:"peak_tx_mbps"`
	Connections    []IftopConnection  `json:"connections"`
}

type IftopConnection struct {
	Source      string  `json:"source"`
	Destination string  `json:"destination"`
	RXRate      float64 `json:"rx_rate_kbps"`
	TXRate      float64 `json:"tx_rate_kbps"`
}

func (t *IftopTool) ParseOutput(output string) (interface{}, error) {
	result := &IftopResult{
		Connections: make([]IftopConnection, 0),
	}

	// Iftop text output parsing
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Parse total rates
		if strings.Contains(line, "Total send rate:") {
			rateRegex := regexp.MustCompile(`([\d.]+)\s*(Kb|Mb|Gb)`)
			if match := rateRegex.FindStringSubmatch(line); len(match) > 2 {
				rate, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if unit == "Kb" {
					result.TotalTX = rate / 1024
				} else if unit == "Mb" {
					result.TotalTX = rate
				} else if unit == "Gb" {
					result.TotalTX = rate * 1024
				}
			}
		}

		if strings.Contains(line, "Total receive rate:") {
			rateRegex := regexp.MustCompile(`([\d.]+)\s*(Kb|Mb|Gb)`)
			if match := rateRegex.FindStringSubmatch(line); len(match) > 2 {
				rate, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				if unit == "Kb" {
					result.TotalRX = rate / 1024
				} else if unit == "Mb" {
					result.TotalRX = rate
				} else if unit == "Gb" {
					result.TotalRX = rate * 1024
				}
			}
		}
	}

	return result, nil
}

func (t *IftopTool) Monitor(ctx context.Context, iface string, seconds int) (*IftopResult, error) {
	args := []string{"-t", "-s", strconv.Itoa(seconds)}
	if iface != "" {
		args = append(args, "-i", iface)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	iftopResult := parsed.(*IftopResult)
	iftopResult.Interface = iface
	return iftopResult, nil
}

// NloadTool wraps the nload command
type NloadTool struct {
	*BaseTool
}

func NewNloadTool() *NloadTool {
	return &NloadTool{
		BaseTool: NewBaseTool(
			"nload",
			"nload",
			CategorySystem,
			"apt-get install nload",
			"Console network traffic and bandwidth monitor",
		),
	}
}

func (t *NloadTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

// BmonTool wraps the bmon command
type BmonTool struct {
	*BaseTool
}

func NewBmonTool() *BmonTool {
	return &BmonTool{
		BaseTool: NewBaseTool(
			"bmon",
			"bmon",
			CategorySystem,
			"apt-get install bmon",
			"Bandwidth monitoring and rate estimating tool",
		),
	}
}

func (t *BmonTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

// VnstatTool wraps the vnstat command
type VnstatTool struct {
	*BaseTool
}

func NewVnstatTool() *VnstatTool {
	return &VnstatTool{
		BaseTool: NewBaseTool(
			"vnstat",
			"vnstat",
			CategorySystem,
			"apt-get install vnstat",
			"Network traffic statistics and history",
		),
	}
}

type VnstatResult struct {
	Interface    string             `json:"interface"`
	TotalRX      int64              `json:"total_rx_bytes"`
	TotalTX      int64              `json:"total_tx_bytes"`
	Today        VnstatDayStats     `json:"today"`
	CurrentMonth VnstatMonthStats   `json:"current_month"`
}

type VnstatDayStats struct {
	RX int64 `json:"rx_bytes"`
	TX int64 `json:"tx_bytes"`
}

type VnstatMonthStats struct {
	RX int64 `json:"rx_bytes"`
	TX int64 `json:"tx_bytes"`
}

func (t *VnstatTool) ParseOutput(output string) (interface{}, error) {
	result := &VnstatResult{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse interface name
		if strings.Contains(line, "Database updated:") {
			continue
		}

		// Parse rx/tx data
		if strings.Contains(line, "rx:") {
			dataRegex := regexp.MustCompile(`rx:\s+([\d.]+)\s+([KMGT]iB)`)
			if match := dataRegex.FindStringSubmatch(line); len(match) > 2 {
				value, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				bytes := convertToBytes(value, unit)
				result.TotalRX = bytes
			}
		}

		if strings.Contains(line, "tx:") {
			dataRegex := regexp.MustCompile(`tx:\s+([\d.]+)\s+([KMGT]iB)`)
			if match := dataRegex.FindStringSubmatch(line); len(match) > 2 {
				value, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]
				bytes := convertToBytes(value, unit)
				result.TotalTX = bytes
			}
		}
	}

	return result, nil
}

func convertToBytes(value float64, unit string) int64 {
	multiplier := 1.0
	switch unit {
	case "KiB":
		multiplier = 1024
	case "MiB":
		multiplier = 1024 * 1024
	case "GiB":
		multiplier = 1024 * 1024 * 1024
	case "TiB":
		multiplier = 1024 * 1024 * 1024 * 1024
	}
	return int64(value * multiplier)
}

func (t *VnstatTool) GetStats(ctx context.Context, iface string) (*VnstatResult, error) {
	args := []string{}
	if iface != "" {
		args = append(args, "-i", iface)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	vnstatResult := parsed.(*VnstatResult)
	vnstatResult.Interface = iface
	return vnstatResult, nil
}

func (t *VnstatTool) GetJSON(ctx context.Context, iface string) (map[string]interface{}, error) {
	args := []string{"--json"}
	if iface != "" {
		args = append(args, "-i", iface)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	var jsonResult map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &jsonResult); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return jsonResult, nil
}

// Register traffic monitoring tools
func RegisterTrafficMonitoringTools(registry *ToolRegistry) {
	registry.Register(NewIftopTool())
	registry.Register(NewNloadTool())
	registry.Register(NewBmonTool())
	registry.Register(NewVnstatTool())
}
