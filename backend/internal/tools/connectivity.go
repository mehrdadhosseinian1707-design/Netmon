package tools

import (
	"context"
	"strconv"
	"strings"
)

// Ping6Tool wraps the ping6 command for IPv6 ICMP testing
type Ping6Tool struct {
	*BaseTool
}

func NewPing6Tool() *Ping6Tool {
	return &Ping6Tool{
		BaseTool: NewBaseTool(
			"ping6",
			"ping6",
			CategoryICMP,
			"apt-get install iputils-ping",
			"IPv6 ICMP echo request for latency and reachability testing",
		),
	}
}

func (t *Ping6Tool) ParseOutput(output string) (interface{}, error) {
	// Use same parser as regular ping
	pingTool := NewPingTool()
	return pingTool.ParseOutput(output)
}

func (t *Ping6Tool) Ping6(ctx context.Context, host string, count int) (*PingResult, error) {
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

// NcatTool wraps the ncat command (modern netcat)
type NcatTool struct {
	*BaseTool
}

func NewNcatTool() *NcatTool {
	return &NcatTool{
		BaseTool: NewBaseTool(
			"ncat",
			"ncat",
			CategorySystem,
			"apt-get install nmap",
			"Modern network connectivity tool for TCP/UDP testing",
		),
	}
}

type NcatResult struct {
	Success      bool    `json:"success"`
	Host         string  `json:"host"`
	Port         int     `json:"port"`
	Protocol     string  `json:"protocol"`
	Connected    bool    `json:"connected"`
	ConnectTime  float64 `json:"connect_time_ms"`
	BytesSent    int64   `json:"bytes_sent"`
	BytesRecv    int64   `json:"bytes_received"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

func (t *NcatTool) ParseOutput(output string) (interface{}, error) {
	result := &NcatResult{
		Connected: strings.Contains(output, "Connected") || strings.Contains(output, "succeeded"),
	}

	// Parse connection info from output
	if strings.Contains(output, "Connection refused") {
		result.ErrorMessage = "Connection refused"
	} else if strings.Contains(output, "Connection timed out") {
		result.ErrorMessage = "Connection timed out"
	}

	result.Success = result.Connected
	return result, nil
}

func (t *NcatTool) TestConnection(ctx context.Context, host string, port int, protocol string) (*NcatResult, error) {
	args := []string{"-v", "-w", "5"}

	if protocol == "udp" {
		args = append(args, "-u")
	}

	args = append(args, host, strconv.Itoa(port))

	result, err := t.Execute(ctx, args)

	parsed, parseErr := t.ParseOutput(result.Output)
	if parseErr != nil {
		return nil, parseErr
	}

	ncatResult := parsed.(*NcatResult)
	ncatResult.Host = host
	ncatResult.Port = port
	ncatResult.Protocol = protocol

	if err != nil && !ncatResult.Connected {
		ncatResult.ErrorMessage = err.Error()
	}

	return ncatResult, nil
}

// NetcatTool wraps the nc (netcat) command
type NetcatTool struct {
	*BaseTool
}

func NewNetcatTool() *NetcatTool {
	return &NetcatTool{
		BaseTool: NewBaseTool(
			"nc",
			"nc",
			CategorySystem,
			"apt-get install netcat-openbsd",
			"Traditional netcat for TCP/UDP connectivity testing",
		),
	}
}

func (t *NetcatTool) ParseOutput(output string) (interface{}, error) {
	ncatTool := NewNcatTool()
	return ncatTool.ParseOutput(output)
}

func (t *NetcatTool) TestConnection(ctx context.Context, host string, port int, protocol string) (*NcatResult, error) {
	args := []string{"-v", "-w", "5", "-z"}

	if protocol == "udp" {
		args = append(args, "-u")
	}

	args = append(args, host, strconv.Itoa(port))

	result, err := t.Execute(ctx, args)

	parsed, parseErr := t.ParseOutput(result.Output)
	if parseErr != nil {
		return nil, parseErr
	}

	ncResult := parsed.(*NcatResult)
	ncResult.Host = host
	ncResult.Port = port
	ncResult.Protocol = protocol

	if err != nil && !ncResult.Connected {
		ncResult.ErrorMessage = err.Error()
	}

	return ncResult, nil
}

// TelnetTool wraps the telnet command
type TelnetTool struct {
	*BaseTool
}

func NewTelnetTool() *TelnetTool {
	return &TelnetTool{
		BaseTool: NewBaseTool(
			"telnet",
			"telnet",
			CategorySystem,
			"apt-get install telnet",
			"Basic TCP connection testing and terminal access",
		),
	}
}

type TelnetResult struct {
	Success      bool    `json:"success"`
	Host         string  `json:"host"`
	Port         int     `json:"port"`
	Connected    bool    `json:"connected"`
	ConnectTime  float64 `json:"connect_time_ms"`
	Banner       string  `json:"banner,omitempty"`
	ErrorMessage string  `json:"error_message,omitempty"`
}

func (t *TelnetTool) ParseOutput(output string) (interface{}, error) {
	result := &TelnetResult{}

	// Check for successful connection
	if strings.Contains(output, "Connected to") || strings.Contains(output, "Escape character") {
		result.Connected = true
		result.Success = true

		// Extract banner (first few lines after connection)
		lines := strings.Split(output, "\n")
		banner := make([]string, 0)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, "Trying") && !strings.Contains(line, "Connected") && !strings.Contains(line, "Escape") {
				banner = append(banner, line)
				if len(banner) >= 3 {
					break
				}
			}
		}
		result.Banner = strings.Join(banner, "\n")
	} else if strings.Contains(output, "Connection refused") {
		result.ErrorMessage = "Connection refused"
	} else if strings.Contains(output, "Connection timed out") || strings.Contains(output, "Unable to connect") {
		result.ErrorMessage = "Connection timed out"
	} else if strings.Contains(output, "Name or service not known") {
		result.ErrorMessage = "Host not found"
	}

	return result, nil
}

func (t *TelnetTool) TestConnection(ctx context.Context, host string, port int) (*TelnetResult, error) {
	args := []string{host, strconv.Itoa(port)}

	// Telnet needs special handling - send quit command immediately
	result, err := t.Execute(ctx, args)

	parsed, parseErr := t.ParseOutput(result.Output)
	if parseErr != nil {
		return nil, parseErr
	}

	telnetResult := parsed.(*TelnetResult)
	telnetResult.Host = host
	telnetResult.Port = port

	if err != nil && !telnetResult.Connected {
		telnetResult.ErrorMessage = err.Error()
	}

	return telnetResult, nil
}

// Update ICMP tools registration
func RegisterICMPToolsExtended(registry *ToolRegistry) {
	RegisterICMPTools(registry) // Register existing tools
	registry.Register(NewPing6Tool())
}

// Register connectivity tools
func RegisterConnectivityTools(registry *ToolRegistry) {
	registry.Register(NewNcatTool())
	registry.Register(NewNetcatTool())
	registry.Register(NewTelnetTool())
}
