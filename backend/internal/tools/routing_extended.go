package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Traceroute6Tool wraps the traceroute6 command for IPv6 path discovery
type Traceroute6Tool struct {
	*BaseTool
}

func NewTraceroute6Tool() *Traceroute6Tool {
	return &Traceroute6Tool{
		BaseTool: NewBaseTool(
			"traceroute6",
			"traceroute6",
			CategoryTraceroute,
			"apt-get install iputils-tracepath",
			"IPv6 network path discovery",
		),
	}
}

func (t *Traceroute6Tool) ParseOutput(output string) (interface{}, error) {
	// Use same parser as regular traceroute
	tracerouteTool := NewTracerouteTool()
	return tracerouteTool.ParseOutput(output)
}

func (t *Traceroute6Tool) Traceroute6(ctx context.Context, host string, maxHops int) (*TracerouteResult, error) {
	args := []string{"-m", fmt.Sprintf("%d", maxHops), host}

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

// DumpcapTool wraps the dumpcap command for lightweight packet capture
type DumpcapTool struct {
	*BaseTool
}

func NewDumpcapTool() *DumpcapTool {
	return &DumpcapTool{
		BaseTool: NewBaseTool(
			"dumpcap",
			"dumpcap",
			CategoryPacketCapture,
			"apt-get install wireshark",
			"Lightweight network packet capture tool",
		),
	}
}

func (t *DumpcapTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

func (t *DumpcapTool) Capture(ctx context.Context, iface string, count int, output string) error {
	args := []string{"-i", iface, "-c", fmt.Sprintf("%d", count), "-w", output}

	_, err := t.Execute(ctx, args)
	return err
}

// BGPDumpTool wraps bgpdump for BGP routing data analysis
type BGPDumpTool struct {
	*BaseTool
}

func NewBGPDumpTool() *BGPDumpTool {
	return &BGPDumpTool{
		BaseTool: NewBaseTool(
			"bgpdump",
			"bgpdump",
			CategoryAdvanced,
			"apt-get install bgpdump",
			"BGP MRT dump file parser and analyzer",
		),
	}
}

func (t *BGPDumpTool) ParseOutput(output string) (interface{}, error) {
	// BGPDump outputs text or JSON based on flags
	if strings.HasPrefix(strings.TrimSpace(output), "{") {
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(output), &result); err == nil {
			return result, nil
		}
	}

	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

func (t *BGPDumpTool) DumpMRT(ctx context.Context, mrtFile string) (map[string]interface{}, error) {
	args := []string{"-m", mrtFile}

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

// PyBGPStreamTool placeholder for pybgpstream integration
type PyBGPStreamTool struct {
	*BaseTool
}

func NewPyBGPStreamTool() *PyBGPStreamTool {
	return &PyBGPStreamTool{
		BaseTool: NewBaseTool(
			"pybgpstream",
			"python3",
			CategoryAdvanced,
			"pip3 install pybgpstream",
			"Python library for BGP data stream analysis",
		),
	}
}

func (t *PyBGPStreamTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"note": "PyBGPStream requires Python integration",
		"raw_output": output,
	}, nil
}

// RouteViewsTool placeholder for RouteViews integration
type RouteViewsTool struct {
	*BaseTool
}

func NewRouteViewsTool() *RouteViewsTool {
	return &RouteViewsTool{
		BaseTool: NewBaseTool(
			"routeviews",
			"routeviews",
			CategoryAdvanced,
			"See http://www.routeviews.org/",
			"RouteViews BGP routing data archive (requires API integration)",
		),
	}
}

func (t *RouteViewsTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"note": "RouteViews requires API integration",
		"raw_output": output,
	}, nil
}

// RIPERISTool placeholder for RIPE RIS integration
type RIPERISTool struct {
	*BaseTool
}

func NewRIPERISTool() *RIPERISTool {
	return &RIPERISTool{
		BaseTool: NewBaseTool(
			"ripe-ris",
			"ripe-ris",
			CategoryAdvanced,
			"See https://www.ripe.net/analyse/internet-measurements/routing-information-service-ris",
			"RIPE Routing Information Service (requires API integration)",
		),
	}
}

func (t *RIPERISTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"note": "RIPE RIS requires API integration",
		"raw_output": output,
	}, nil
}

// Register extended routing and BGP tools
func RegisterRoutingToolsExtended(registry *ToolRegistry) {
	registry.Register(NewTraceroute6Tool())
}

func RegisterPacketCaptureToolsExtended(registry *ToolRegistry) {
	RegisterPacketCaptureTools(registry)
	registry.Register(NewDumpcapTool())
}

func RegisterBGPTools(registry *ToolRegistry) {
	registry.Register(NewBGPDumpTool())
	registry.Register(NewPyBGPStreamTool())
	registry.Register(NewRouteViewsTool())
	registry.Register(NewRIPERISTool())
}
