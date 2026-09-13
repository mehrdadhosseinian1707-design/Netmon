package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/netmon/netmon/internal/tools"
)

func main() {
	// Define subcommands
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	checkCmd := flag.NewFlagSet("check", flag.ExitOnError)
	installCmd := flag.NewFlagSet("install", flag.ExitOnError)
	executeCmd := flag.NewFlagSet("execute", flag.ExitOnError)
	infoCmd := flag.NewFlagSet("info", flag.ExitOnError)

	// List command flags
	listInstalled := listCmd.Bool("installed", false, "List only installed tools")
	listCategory := listCmd.String("category", "", "List tools by category")
	listJSON := listCmd.Bool("json", false, "Output as JSON")

	// Execute command flags
	executeTool := executeCmd.String("tool", "", "Tool to execute (required)")
	executeTimeout := executeCmd.Int("timeout", 30, "Timeout in seconds")
	executeJSON := executeCmd.Bool("json", false, "Output as JSON")

	// Install command flags
	installScript := installCmd.String("script", "install-tools.sh", "Output script filename")
	installGenerate := installCmd.Bool("generate", false, "Generate installation script")

	// Info command flags
	infoTool := infoCmd.String("tool", "", "Tool name (required)")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		handleList(*listInstalled, *listCategory, *listJSON)

	case "check":
		checkCmd.Parse(os.Args[2:])
		handleCheck()

	case "install":
		installCmd.Parse(os.Args[2:])
		handleInstall(*installGenerate, *installScript)

	case "execute":
		executeCmd.Parse(os.Args[2:])
		if *executeTool == "" {
			fmt.Println("Error: -tool flag is required")
			executeCmd.Usage()
			os.Exit(1)
		}
		handleExecute(*executeTool, executeCmd.Args(), *executeTimeout, *executeJSON)

	case "info":
		infoCmd.Parse(os.Args[2:])
		if *infoTool == "" {
			fmt.Println("Error: -tool flag is required")
			infoCmd.Usage()
			os.Exit(1)
		}
		handleInfo(*infoTool)

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("NetMon Network Analysis Tools CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  netmon-tools <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  list      List available network analysis tools")
	fmt.Println("  check     Check tool dependencies")
	fmt.Println("  install   Generate installation script for missing tools")
	fmt.Println("  execute   Execute a network analysis tool")
	fmt.Println("  info      Get information about a specific tool")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  netmon-tools list --installed")
	fmt.Println("  netmon-tools list --category icmp")
	fmt.Println("  netmon-tools check")
	fmt.Println("  netmon-tools install --generate --script install.sh")
	fmt.Println("  netmon-tools execute --tool ping 8.8.8.8 -c 4")
	fmt.Println("  netmon-tools info --tool mtr")
}

func handleList(installed bool, category string, asJSON bool) {
	var toolsList []tools.ToolInfo

	if category != "" {
		// List by category
		categoryTools := tools.GlobalRegistry.ListByCategory(tools.ToolCategory(category))
		for _, tool := range categoryTools {
			if info, err := tools.GlobalRegistry.GetToolInfo(tool.Name()); err == nil {
				toolsList = append(toolsList, *info)
			}
		}
	} else if installed {
		// List only installed
		for _, tool := range tools.GlobalRegistry.ListInstalled() {
			if info, err := tools.GlobalRegistry.GetToolInfo(tool.Name()); err == nil {
				toolsList = append(toolsList, *info)
			}
		}
	} else {
		// List all
		toolsList = tools.GlobalRegistry.GetAllToolsInfo()
	}

	if asJSON {
		data, _ := json.MarshalIndent(map[string]interface{}{
			"tools": toolsList,
			"count": len(toolsList),
		}, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Total tools: %d\n\n", len(toolsList))
		for _, info := range toolsList {
			status := "✗"
			if info.Installed {
				status = "✓"
			}
			fmt.Printf("%s %-20s [%s] %s\n", status, info.Name, info.Category, info.Description)
			if info.Installed && info.Version != "" {
				fmt.Printf("  Version: %s\n", info.Version)
			}
			if !info.Installed && info.InstallCommand != "" {
				fmt.Printf("  Install: %s\n", info.InstallCommand)
			}
		}
	}
}

func handleCheck() {
	checker := tools.NewDependencyChecker()
	report := checker.CheckAll()
	report.PrintReport()
}

func handleInstall(generate bool, scriptPath string) {
	installer := tools.NewInstaller()

	if generate {
		fmt.Printf("Generating installation script: %s\n", scriptPath)
		if err := installer.SaveInstallScript(scriptPath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Installation script generated successfully!")
		fmt.Printf("Run: chmod +x %s && ./%s\n", scriptPath, scriptPath)
	} else {
		// Just show missing tools
		missing := installer.GetAllInstallCommands()
		if len(missing) == 0 {
			fmt.Println("All tools are installed!")
			return
		}

		fmt.Printf("Missing %d tools:\n\n", len(missing))
		for name, cmd := range missing {
			fmt.Printf("%-20s: %s\n", name, cmd)
		}
		fmt.Printf("\nRun with --generate to create installation script\n")
	}
}

func handleExecute(toolName string, args []string, timeout int, asJSON bool) {
	tool, err := tools.GlobalRegistry.Get(toolName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if !tool.IsInstalled() {
		fmt.Printf("Tool '%s' is not installed\n", toolName)
		fmt.Printf("Install with: %s\n", tool.InstallCommand())
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	fmt.Printf("Executing: %s %v\n", toolName, args)
	result, err := tool.Execute(ctx, args)

	if err != nil && result == nil {
		fmt.Printf("Execution error: %v\n", err)
		os.Exit(1)
	}

	if asJSON {
		// Try to parse output
		if result.Success {
			parsed, err := tool.ParseOutput(result.Output)
			if err == nil {
				result.ParsedData = parsed
			}
		}

		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("\n--- Output ---\n")
		fmt.Println(result.Output)
		fmt.Printf("\n--- Result ---\n")
		fmt.Printf("Success: %v\n", result.Success)
		fmt.Printf("Exit Code: %d\n", result.ExitCode)
		fmt.Printf("Execution Time: %v\n", result.ExecutionTime)
		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}
	}
}

func handleInfo(toolName string) {
	info, err := tools.GlobalRegistry.GetToolInfo(toolName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Tool: %s\n", info.Name)
	fmt.Printf("Binary: %s\n", info.Binary)
	fmt.Printf("Category: %s\n", info.Category)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Printf("Installed: %v\n", info.Installed)

	if info.Installed {
		if info.Version != "" {
			fmt.Printf("Version: %s\n", info.Version)
		}
	} else {
		fmt.Printf("Install Command: %s\n", info.InstallCommand)
	}
}
