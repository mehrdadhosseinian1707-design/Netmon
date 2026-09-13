package tools

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Installer helps with tool installation
type Installer struct {
	OS string
}

func NewInstaller() *Installer {
	return &Installer{
		OS: runtime.GOOS,
	}
}

// GetInstallCommands returns installation commands for missing tools
func (i *Installer) GetInstallCommands(toolNames []string) map[string]string {
	commands := make(map[string]string)

	for _, name := range toolNames {
		tool, err := GlobalRegistry.Get(name)
		if err != nil {
			continue
		}

		if !tool.IsInstalled() {
			commands[name] = tool.InstallCommand()
		}
	}

	return commands
}

// GetAllInstallCommands returns installation commands for all missing tools
func (i *Installer) GetAllInstallCommands() map[string]string {
	commands := make(map[string]string)

	for _, tool := range GlobalRegistry.List() {
		if !tool.IsInstalled() {
			commands[tool.Name()] = tool.InstallCommand()
		}
	}

	return commands
}

// GenerateInstallScript generates a shell script to install all missing tools
func (i *Installer) GenerateInstallScript() string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n")
	script.WriteString("# Network Monitoring Tools Installation Script\n")
	script.WriteString("# Generated for " + i.OS + "\n\n")

	if i.OS != "linux" {
		script.WriteString("# Warning: This script is designed for Linux\n")
		script.WriteString("# Please install tools manually on " + i.OS + "\n\n")
	}

	script.WriteString("set -e\n\n")

	// Detect package manager
	script.WriteString("# Detect package manager\n")
	script.WriteString("if command -v apt-get &> /dev/null; then\n")
	script.WriteString("    PM='apt-get'\n")
	script.WriteString("    UPDATE='sudo apt-get update'\n")
	script.WriteString("    INSTALL='sudo apt-get install -y'\n")
	script.WriteString("elif command -v yum &> /dev/null; then\n")
	script.WriteString("    PM='yum'\n")
	script.WriteString("    UPDATE='sudo yum update -y'\n")
	script.WriteString("    INSTALL='sudo yum install -y'\n")
	script.WriteString("elif command -v dnf &> /dev/null; then\n")
	script.WriteString("    PM='dnf'\n")
	script.WriteString("    UPDATE='sudo dnf update -y'\n")
	script.WriteString("    INSTALL='sudo dnf install -y'\n")
	script.WriteString("else\n")
	script.WriteString("    echo 'No supported package manager found'\n")
	script.WriteString("    exit 1\n")
	script.WriteString("fi\n\n")

	script.WriteString("echo 'Updating package lists...'\n")
	script.WriteString("$UPDATE\n\n")

	// Group tools by package
	packages := make(map[string][]string)

	for _, tool := range GlobalRegistry.List() {
		if !tool.IsInstalled() {
			installCmd := tool.InstallCommand()
			// Extract package name from install command
			if strings.Contains(installCmd, "apt-get install") {
				pkg := strings.TrimPrefix(installCmd, "apt-get install ")
				packages[pkg] = append(packages[pkg], tool.Name())
			}
		}
	}

	script.WriteString("echo 'Installing network monitoring tools...'\n\n")

	for pkg, toolNames := range packages {
		script.WriteString("# " + strings.Join(toolNames, ", ") + "\n")
		script.WriteString("$INSTALL " + pkg + "\n\n")
	}

	script.WriteString("echo 'Installation complete!'\n")
	script.WriteString("echo 'Installed tools:'\n")

	for pkg := range packages {
		script.WriteString("command -v " + pkg + " && echo '  ✓ " + pkg + "'\n")
	}

	return script.String()
}

// SaveInstallScript saves the installation script to a file
func (i *Installer) SaveInstallScript(filename string) error {
	script := i.GenerateInstallScript()

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create script file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(script)
	if err != nil {
		return fmt.Errorf("failed to write script: %w", err)
	}

	return nil
}

// DependencyChecker checks system dependencies
type DependencyChecker struct{}

func NewDependencyChecker() *DependencyChecker {
	return &DependencyChecker{}
}

// CheckAll checks all tool dependencies
func (c *DependencyChecker) CheckAll() *DependencyReport {
	report := &DependencyReport{
		Tools:     make(map[string]ToolStatus),
		Timestamp: "2026-09-11T04:57:07.551Z",
	}

	for _, tool := range GlobalRegistry.List() {
		status := ToolStatus{
			Name:      tool.Name(),
			Installed: tool.IsInstalled(),
			Command:   tool.InstallCommand(),
		}

		if status.Installed {
			// Try to get version
			if version := c.GetVersion(tool); version != "" {
				status.Version = version
			}
		}

		report.Tools[tool.Name()] = status

		if status.Installed {
			report.Installed++
		} else {
			report.Missing++
		}
	}

	report.Total = len(report.Tools)

	return report
}

// CheckCategory checks dependencies for a specific category
func (c *DependencyChecker) CheckCategory(category ToolCategory) *DependencyReport {
	report := &DependencyReport{
		Tools:     make(map[string]ToolStatus),
		Category:  string(category),
		Timestamp: "2026-09-11T04:57:07.551Z",
	}

	for _, tool := range GlobalRegistry.ListByCategory(category) {
		status := ToolStatus{
			Name:      tool.Name(),
			Installed: tool.IsInstalled(),
			Command:   tool.InstallCommand(),
		}

		if status.Installed {
			if version := c.GetVersion(tool); version != "" {
				status.Version = version
			}
		}

		report.Tools[tool.Name()] = status

		if status.Installed {
			report.Installed++
		} else {
			report.Missing++
		}
	}

	report.Total = len(report.Tools)

	return report
}

// GetVersion attempts to get tool version
func (c *DependencyChecker) GetVersion(tool Tool) string {
	cmd := exec.Command(tool.Name(), "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try -version
		cmd = exec.Command(tool.Name(), "-version")
		output, err = cmd.CombinedOutput()
		if err != nil {
			// Try version
			cmd = exec.Command(tool.Name(), "version")
			output, err = cmd.CombinedOutput()
			if err != nil {
				return ""
			}
		}
	}

	return ParseVersion(string(output))
}

// ToolStatus represents the status of a tool
type ToolStatus struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version,omitempty"`
	Command   string `json:"install_command,omitempty"`
}

// DependencyReport represents a dependency check report
type DependencyReport struct {
	Tools     map[string]ToolStatus `json:"tools"`
	Installed int                   `json:"installed"`
	Missing   int                   `json:"missing"`
	Total     int                   `json:"total"`
	Category  string                `json:"category,omitempty"`
	Timestamp string                `json:"timestamp"`
}

// PrintReport prints a formatted dependency report
func (r *DependencyReport) PrintReport() {
	fmt.Printf("\n=== Network Monitoring Tools Dependency Report ===\n")
	if r.Category != "" {
		fmt.Printf("Category: %s\n", r.Category)
	}
	fmt.Printf("Timestamp: %s\n", r.Timestamp)
	fmt.Printf("Total: %d | Installed: %d | Missing: %d\n\n", r.Total, r.Installed, r.Missing)

	if r.Installed > 0 {
		fmt.Println("✓ Installed Tools:")
		for name, status := range r.Tools {
			if status.Installed {
				version := ""
				if status.Version != "" {
					version = " (" + status.Version + ")"
				}
				fmt.Printf("  ✓ %s%s\n", name, version)
			}
		}
		fmt.Println()
	}

	if r.Missing > 0 {
		fmt.Println("✗ Missing Tools:")
		for name, status := range r.Tools {
			if !status.Installed {
				fmt.Printf("  ✗ %s\n", name)
				if status.Command != "" {
					fmt.Printf("    Install: %s\n", status.Command)
				}
			}
		}
	}

	fmt.Println()
}
