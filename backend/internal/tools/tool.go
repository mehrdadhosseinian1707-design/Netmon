package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Tool represents a network analysis tool
type Tool interface {
	Name() string
	IsInstalled() bool
	InstallCommand() string
	Execute(ctx context.Context, args []string) (*ToolResult, error)
	ParseOutput(output string) (interface{}, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Tool          string                 `json:"tool"`
	Success       bool                   `json:"success"`
	ExitCode      int                    `json:"exit_code"`
	Output        string                 `json:"output"`
	Error         string                 `json:"error,omitempty"`
	ParsedData    interface{}            `json:"parsed_data,omitempty"`
	ExecutionTime time.Duration          `json:"execution_time"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// ToolCategory represents a category of tools
type ToolCategory string

const (
	CategoryICMP           ToolCategory = "icmp"
	CategoryTraceroute     ToolCategory = "traceroute"
	CategoryDNS            ToolCategory = "dns"
	CategoryHTTP           ToolCategory = "http"
	CategoryThroughput     ToolCategory = "throughput"
	CategoryPacketAnalysis ToolCategory = "packet_analysis"
	CategoryPacketCapture  ToolCategory = "packet_capture"
	CategorySystem         ToolCategory = "system"
	CategoryAdvanced       ToolCategory = "advanced"
)

// BaseTool provides common functionality for all tools
type BaseTool struct {
	name           string
	binary         string
	category       ToolCategory
	installCommand string
	description    string
}

func NewBaseTool(name, binary string, category ToolCategory, installCmd, description string) *BaseTool {
	return &BaseTool{
		name:           name,
		binary:         binary,
		category:       category,
		installCommand: installCmd,
		description:    description,
	}
}

func (t *BaseTool) Name() string {
	return t.name
}

func (t *BaseTool) Binary() string {
	return t.binary
}

func (t *BaseTool) Category() ToolCategory {
	return t.category
}

func (t *BaseTool) Description() string {
	return t.description
}

func (t *BaseTool) IsInstalled() bool {
	_, err := exec.LookPath(t.binary)
	return err == nil
}

func (t *BaseTool) InstallCommand() string {
	return t.installCommand
}

func (t *BaseTool) Execute(ctx context.Context, args []string) (*ToolResult, error) {
	if !t.IsInstalled() {
		return nil, fmt.Errorf("tool %s is not installed. Install with: %s", t.name, t.installCommand)
	}

	result := &ToolResult{
		Tool:      t.name,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	cmd := exec.CommandContext(ctx, t.binary, args...)

	outputBytes, err := cmd.CombinedOutput()
	result.Output = string(outputBytes)
	result.EndTime = time.Now()
	result.ExecutionTime = result.EndTime.Sub(result.StartTime)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		}
		result.Success = false
		result.Error = err.Error()
		return result, err
	}

	result.Success = true
	result.ExitCode = 0

	return result, nil
}

// ToolRegistry manages all available tools
type ToolRegistry struct {
	tools map[string]Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

func (r *ToolRegistry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

func (r *ToolRegistry) Get(name string) (Tool, error) {
	tool, exists := r.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool %s not found", name)
	}
	return tool, nil
}

func (r *ToolRegistry) List() []Tool {
	tools := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

func (r *ToolRegistry) ListInstalled() []Tool {
	installed := make([]Tool, 0)
	for _, tool := range r.tools {
		if tool.IsInstalled() {
			installed = append(installed, tool)
		}
	}
	return installed
}

// CategorizedTool is an interface for tools that have a category
type CategorizedTool interface {
	Category() ToolCategory
}

func (r *ToolRegistry) ListByCategory(category ToolCategory) []Tool {
	tools := make([]Tool, 0)
	for _, tool := range r.tools {
		// Check if the tool has a Category method by checking its embedded BaseTool
		if ct, ok := interface{}(tool).(CategorizedTool); ok && ct.Category() == category {
			tools = append(tools, tool)
		}
	}
	return tools
}

func (r *ToolRegistry) CheckDependencies() map[string]bool {
	deps := make(map[string]bool)
	for name, tool := range r.tools {
		deps[name] = tool.IsInstalled()
	}
	return deps
}

// ExecuteWithTimeout executes a tool with a timeout
func ExecuteWithTimeout(tool Tool, args []string, timeout time.Duration) (*ToolResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return tool.Execute(ctx, args)
}

// ParseVersion extracts version from tool output
func ParseVersion(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "version") {
			return strings.TrimSpace(line)
		}
	}
	return "unknown"
}

// ToolInfo provides information about a tool
type ToolInfo struct {
	Name           string       `json:"name"`
	Binary         string       `json:"binary"`
	Category       ToolCategory `json:"category"`
	Description    string       `json:"description"`
	Installed      bool         `json:"installed"`
	InstallCommand string       `json:"install_command,omitempty"`
	Version        string       `json:"version,omitempty"`
}

// DetailedTool is an interface for tools with detailed information
type DetailedTool interface {
	Binary() string
	Category() ToolCategory
	Description() string
}

func (r *ToolRegistry) GetToolInfo(name string) (*ToolInfo, error) {
	tool, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	info := &ToolInfo{
		Name:      tool.Name(),
		Installed: tool.IsInstalled(),
	}

	// Try to get detailed info if the tool supports it
	if dt, ok := interface{}(tool).(DetailedTool); ok {
		info.Binary = dt.Binary()
		info.Category = dt.Category()
		info.Description = dt.Description()
		info.InstallCommand = tool.InstallCommand()
	}

	// Try to get version if installed
	if info.Installed {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		result, err := tool.Execute(ctx, []string{"--version"})
		if err == nil {
			info.Version = ParseVersion(result.Output)
		}
	}

	return info, nil
}

func (r *ToolRegistry) GetAllToolsInfo() []ToolInfo {
	infos := make([]ToolInfo, 0, len(r.tools))
	for name := range r.tools {
		if info, err := r.GetToolInfo(name); err == nil {
			infos = append(infos, *info)
		}
	}
	return infos
}
