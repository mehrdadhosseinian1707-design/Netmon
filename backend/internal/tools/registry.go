package tools

import (
	"context"
	"time"
)

// GlobalRegistry is the default tool registry
var GlobalRegistry *ToolRegistry

func init() {
	GlobalRegistry = NewToolRegistry()
	RegisterAllTools(GlobalRegistry)
}

// RegisterAllTools registers all available network analysis tools
func RegisterAllTools(registry *ToolRegistry) {
	// Core tools
	RegisterICMPTools(registry)
	RegisterTracerouteTools(registry)
	RegisterDNSTools(registry)
	RegisterHTTPTools(registry)
	RegisterPacketAnalysisTools(registry)
	RegisterPacketCaptureTools(registry)
	RegisterThroughputTools(registry)
	RegisterSystemTools(registry)
	RegisterAdvancedTools(registry)

	// Extended tools
	RegisterICMPToolsExtended(registry)
	RegisterDNSToolsExtended(registry)
	RegisterHTTPToolsExtended(registry)
	RegisterRoutingToolsExtended(registry)
	RegisterPacketCaptureToolsExtended(registry)
	RegisterConnectivityTools(registry)
	RegisterBandwidthTestingTools(registry)
	RegisterTrafficMonitoringTools(registry)
	RegisterBGPTools(registry)
}

// GetTool returns a tool by name from the global registry
func GetTool(name string) (Tool, error) {
	return GlobalRegistry.Get(name)
}

// ListTools returns all tools from the global registry
func ListTools() []Tool {
	return GlobalRegistry.List()
}

// ListInstalledTools returns all installed tools
func ListInstalledTools() []Tool {
	return GlobalRegistry.ListInstalled()
}

// CheckDependencies checks all tool dependencies
func CheckDependencies() map[string]bool {
	return GlobalRegistry.CheckDependencies()
}

// ExecuteTool executes a tool by name with given arguments
func ExecuteTool(ctx context.Context, name string, args []string) (*ToolResult, error) {
	tool, err := GetTool(name)
	if err != nil {
		return nil, err
	}
	return tool.Execute(ctx, args)
}

// ExecuteToolWithTimeout executes a tool with a timeout
func ExecuteToolWithTimeout(name string, args []string, timeout time.Duration) (*ToolResult, error) {
	tool, err := GetTool(name)
	if err != nil {
		return nil, err
	}
	return ExecuteWithTimeout(tool, args, timeout)
}

// GetToolsByCategory returns all tools in a category
func GetToolsByCategory(category ToolCategory) []Tool {
	return GlobalRegistry.ListByCategory(category)
}

// GetAllToolsInfo returns information about all tools
func GetAllToolsInfo() []ToolInfo {
	return GlobalRegistry.GetAllToolsInfo()
}
