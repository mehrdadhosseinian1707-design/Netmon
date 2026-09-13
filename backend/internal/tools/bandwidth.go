package tools

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// SpeedtestCLITool wraps the speedtest-cli command
type SpeedtestCLITool struct {
	*BaseTool
}

func NewSpeedtestCLITool() *SpeedtestCLITool {
	return &SpeedtestCLITool{
		BaseTool: NewBaseTool(
			"speedtest-cli",
			"speedtest-cli",
			CategoryThroughput,
			"apt-get install speedtest-cli",
			"Internet bandwidth testing using Speedtest.net",
		),
	}
}

type SpeedtestResult struct {
	Download    float64 `json:"download_mbps"`
	Upload      float64 `json:"upload_mbps"`
	Ping        float64 `json:"ping_ms"`
	Server      string  `json:"server"`
	ServerID    string  `json:"server_id"`
	Sponsor     string  `json:"sponsor"`
	Distance    float64 `json:"distance_km"`
	Latency     float64 `json:"latency_ms"`
	BytesSent   int64   `json:"bytes_sent"`
	BytesRecv   int64   `json:"bytes_received"`
	Timestamp   string  `json:"timestamp"`
}

func (t *SpeedtestCLITool) ParseOutput(output string) (interface{}, error) {
	// speedtest-cli can output JSON
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		var jsonResult map[string]interface{}
		if err := json.Unmarshal([]byte(output), &jsonResult); err == nil {
			result := &SpeedtestResult{}

			if download, ok := jsonResult["download"].(float64); ok {
				result.Download = download / 1_000_000 // Convert to Mbps
			}
			if upload, ok := jsonResult["upload"].(float64); ok {
				result.Upload = upload / 1_000_000 // Convert to Mbps
			}
			if ping, ok := jsonResult["ping"].(float64); ok {
				result.Ping = ping
			}

			if server, ok := jsonResult["server"].(map[string]interface{}); ok {
				if name, ok := server["name"].(string); ok {
					result.Server = name
				}
				if sponsor, ok := server["sponsor"].(string); ok {
					result.Sponsor = sponsor
				}
				if distance, ok := server["d"].(float64); ok {
					result.Distance = distance
				}
			}

			return result, nil
		}
	}

	// Parse text output
	result := &SpeedtestResult{}
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Download:") {
			downloadRegex := regexp.MustCompile(`Download:\s+([\d.]+)\s+Mbit/s`)
			if match := downloadRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Download, _ = strconv.ParseFloat(match[1], 64)
			}
		}

		if strings.Contains(line, "Upload:") {
			uploadRegex := regexp.MustCompile(`Upload:\s+([\d.]+)\s+Mbit/s`)
			if match := uploadRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Upload, _ = strconv.ParseFloat(match[1], 64)
			}
		}

		if strings.Contains(line, "Hosted by") {
			serverRegex := regexp.MustCompile(`Hosted by\s+(.+?)\s+\[`)
			if match := serverRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Sponsor = match[1]
			}
		}
	}

	return result, nil
}

func (t *SpeedtestCLITool) Test(ctx context.Context, serverID string) (*SpeedtestResult, error) {
	args := []string{"--json"}

	if serverID != "" {
		args = append(args, "--server", serverID)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*SpeedtestResult), nil
}

// FastCLITool wraps the fast-cli command (Netflix speed test)
type FastCLITool struct {
	*BaseTool
}

func NewFastCLITool() *FastCLITool {
	return &FastCLITool{
		BaseTool: NewBaseTool(
			"fast",
			"fast",
			CategoryThroughput,
			"npm install -g fast-cli",
			"Netflix Fast.com speed test CLI",
		),
	}
}

type FastResult struct {
	Download float64 `json:"download_mbps"`
	Upload   float64 `json:"upload_mbps"`
	Latency  float64 `json:"latency_ms"`
}

func (t *FastCLITool) ParseOutput(output string) (interface{}, error) {
	result := &FastResult{}

	// Parse download speed
	downloadRegex := regexp.MustCompile(`([\d.]+)\s+Mbps`)
	matches := downloadRegex.FindAllStringSubmatch(output, -1)

	if len(matches) > 0 {
		result.Download, _ = strconv.ParseFloat(matches[0][1], 64)
	}

	if len(matches) > 1 {
		result.Upload, _ = strconv.ParseFloat(matches[1][1], 64)
	}

	// Parse latency
	latencyRegex := regexp.MustCompile(`([\d.]+)\s+ms`)
	if match := latencyRegex.FindStringSubmatch(output); len(match) > 1 {
		result.Latency, _ = strconv.ParseFloat(match[1], 64)
	}

	return result, nil
}

func (t *FastCLITool) Test(ctx context.Context, upload bool) (*FastResult, error) {
	args := []string{}

	if upload {
		args = append(args, "--upload")
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*FastResult), nil
}

// Register bandwidth testing tools
func RegisterBandwidthTestingTools(registry *ToolRegistry) {
	registry.Register(NewSpeedtestCLITool())
	registry.Register(NewFastCLITool())
}
