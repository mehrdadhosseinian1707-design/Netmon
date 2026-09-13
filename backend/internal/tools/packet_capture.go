package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TCPDumpTool wraps the tcpdump command
type TCPDumpTool struct {
	*BaseTool
}

func NewTCPDumpTool() *TCPDumpTool {
	return &TCPDumpTool{
		BaseTool: NewBaseTool(
			"tcpdump",
			"tcpdump",
			CategoryPacketCapture,
			"apt-get install tcpdump",
			"Packet capture and network traffic analysis",
		),
	}
}

type TCPDumpPacket struct {
	Timestamp string `json:"timestamp"`
	Protocol  string `json:"protocol"`
	Source    string `json:"source"`
	Dest      string `json:"dest"`
	Length    int    `json:"length"`
	Info      string `json:"info"`
}

type TCPDumpResult struct {
	Packets       []TCPDumpPacket `json:"packets"`
	PacketCount   int             `json:"packet_count"`
	PacketsCaptured int           `json:"packets_captured"`
	PacketsDropped  int           `json:"packets_dropped"`
}

func (t *TCPDumpTool) ParseOutput(output string) (interface{}, error) {
	result := &TCPDumpResult{
		Packets: make([]TCPDumpPacket, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse statistics line
		if strings.Contains(line, "packets captured") {
			statsRegex := regexp.MustCompile(`(\d+)\s+packets captured`)
			if match := statsRegex.FindStringSubmatch(line); len(match) > 1 {
				result.PacketsCaptured, _ = strconv.Atoi(match[1])
			}
		}

		if strings.Contains(line, "packets dropped") {
			droppedRegex := regexp.MustCompile(`(\d+)\s+packets dropped`)
			if match := droppedRegex.FindStringSubmatch(line); len(match) > 1 {
				result.PacketsDropped, _ = strconv.Atoi(match[1])
			}
		}

		// Parse packet line (basic format)
		// Format: "HH:MM:SS.microsec IP source > dest: proto..."
		if strings.Contains(line, "IP") || strings.Contains(line, "IP6") {
			packet := TCPDumpPacket{
				Info: line,
			}

			// Extract timestamp
			timeRegex := regexp.MustCompile(`^([\d:\.]+)`)
			if match := timeRegex.FindStringSubmatch(line); len(match) > 1 {
				packet.Timestamp = match[1]
			}

			// Extract source and destination
			addrRegex := regexp.MustCompile(`([^\s]+)\s+>\s+([^\s:]+)`)
			if match := addrRegex.FindStringSubmatch(line); len(match) > 2 {
				packet.Source = match[1]
				packet.Dest = match[2]
			}

			// Extract protocol
			if strings.Contains(line, "TCP") {
				packet.Protocol = "TCP"
			} else if strings.Contains(line, "UDP") {
				packet.Protocol = "UDP"
			} else if strings.Contains(line, "ICMP") {
				packet.Protocol = "ICMP"
			}

			result.Packets = append(result.Packets, packet)
		}
	}

	result.PacketCount = len(result.Packets)
	return result, nil
}

func (t *TCPDumpTool) TCPDump(ctx context.Context, iface string, count int, filter string) (*TCPDumpResult, error) {
	args := []string{"-n", "-c", strconv.Itoa(count)}

	if iface != "" {
		args = append(args, "-i", iface)
	}

	if filter != "" {
		args = append(args, filter)
	}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TCPDumpResult), nil
}

// TSharkTool wraps the tshark command (Wireshark CLI)
type TSharkTool struct {
	*BaseTool
}

func NewTSharkTool() *TSharkTool {
	return &TSharkTool{
		BaseTool: NewBaseTool(
			"tshark",
			"tshark",
			CategoryPacketCapture,
			"apt-get install tshark",
			"Wireshark command-line packet analyzer",
		),
	}
}

type TSharkPacket struct {
	Number      int               `json:"number"`
	Time        string            `json:"time"`
	Source      string            `json:"source"`
	Destination string            `json:"destination"`
	Protocol    string            `json:"protocol"`
	Length      int               `json:"length"`
	Info        string            `json:"info"`
	Layers      map[string]interface{} `json:"layers,omitempty"`
}

type TSharkResult struct {
	Packets     []TSharkPacket `json:"packets"`
	PacketCount int            `json:"packet_count"`
}

func (t *TSharkTool) ParseOutput(output string) (interface{}, error) {
	// TShark can output JSON with -T json
	if strings.HasPrefix(strings.TrimSpace(output), "[") {
		var packets []map[string]interface{}
		if err := json.Unmarshal([]byte(output), &packets); err == nil {
			result := &TSharkResult{
				Packets:     make([]TSharkPacket, 0),
				PacketCount: len(packets),
			}
			return result, nil
		}
	}

	// Parse text output
	result := &TSharkResult{
		Packets: make([]TSharkPacket, 0),
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse tshark text output
		// Format: "  1   0.000000 192.168.1.1 -> 192.168.1.2 TCP 66 12345 > 80 [SYN]"
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			packet := TSharkPacket{
				Info: line,
			}

			if num, err := strconv.Atoi(fields[0]); err == nil {
				packet.Number = num
			}

			packet.Time = fields[1]
			packet.Source = fields[2]
			// Skip "->" or similar
			if len(fields) > 4 {
				packet.Destination = fields[4]
			}
			if len(fields) > 5 {
				packet.Protocol = fields[5]
			}
			if len(fields) > 6 {
				if length, err := strconv.Atoi(fields[6]); err == nil {
					packet.Length = length
				}
			}

			result.Packets = append(result.Packets, packet)
		}
	}

	result.PacketCount = len(result.Packets)
	return result, nil
}

func (t *TSharkTool) TShark(ctx context.Context, iface string, count int, filter string) (*TSharkResult, error) {
	args := []string{"-c", strconv.Itoa(count)}

	if iface != "" {
		args = append(args, "-i", iface)
	}

	if filter != "" {
		args = append(args, "-Y", filter)
	}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TSharkResult), nil
}

func (t *TSharkTool) TSharkJSON(ctx context.Context, iface string, count int, filter string) ([]map[string]interface{}, error) {
	args := []string{"-T", "json", "-c", strconv.Itoa(count)}

	if iface != "" {
		args = append(args, "-i", iface)
	}

	if filter != "" {
		args = append(args, "-Y", filter)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	var packets []map[string]interface{}
	if err := json.Unmarshal([]byte(result.Output), &packets); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return packets, nil
}

// Register all packet capture tools
func RegisterPacketCaptureTools(registry *ToolRegistry) {
	registry.Register(NewTCPDumpTool())
	registry.Register(NewTSharkTool())
}
