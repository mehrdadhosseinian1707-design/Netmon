package tools

import (
	"context"
	"encoding/json"
	"strings"
)

// NslookupTool wraps the nslookup command
type NslookupTool struct {
	*BaseTool
}

func NewNslookupTool() *NslookupTool {
	return &NslookupTool{
		BaseTool: NewBaseTool(
			"nslookup",
			"nslookup",
			CategoryDNS,
			"apt-get install dnsutils",
			"Basic DNS lookup and testing",
		),
	}
}

type NslookupResult struct {
	Server     string   `json:"server"`
	Address    string   `json:"address"`
	Query      string   `json:"query"`
	Answers    []string `json:"answers"`
	QueryTime  int      `json:"query_time_ms"`
	Authoritative bool  `json:"authoritative"`
}

func (t *NslookupTool) ParseOutput(output string) (interface{}, error) {
	result := &NslookupResult{
		Answers: make([]string, 0),
	}

	lines := strings.Split(output, "\n")
	inAnswerSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse server info
		if strings.HasPrefix(line, "Server:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				result.Server = parts[1]
			}
		}

		// Parse server address
		if strings.HasPrefix(line, "Address:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				result.Address = parts[1]
			}
		}

		// Parse authoritative answer
		if strings.Contains(line, "Authoritative answer") {
			result.Authoritative = true
			inAnswerSection = true
		}

		// Parse Name field
		if strings.HasPrefix(line, "Name:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				result.Query = parts[1]
			}
			inAnswerSection = true
		}

		// Parse addresses
		if inAnswerSection && strings.HasPrefix(line, "Address:") && result.Server != "" {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				result.Answers = append(result.Answers, parts[1])
			}
		}
	}

	return result, nil
}

func (t *NslookupTool) Lookup(ctx context.Context, domain string, recordType string, server string) (*NslookupResult, error) {
	args := []string{}

	if recordType != "" {
		args = append(args, "-type="+recordType)
	}

	args = append(args, domain)

	if server != "" {
		args = append(args, server)
	}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*NslookupResult), nil
}

// ResolvectlTool wraps the resolvectl command
type ResolvectlTool struct {
	*BaseTool
}

func NewResolvectlTool() *ResolvectlTool {
	return &ResolvectlTool{
		BaseTool: NewBaseTool(
			"resolvectl",
			"resolvectl",
			CategoryDNS,
			"apt-get install systemd",
			"Local DNS resolver testing and configuration",
		),
	}
}

func (t *ResolvectlTool) ParseOutput(output string) (interface{}, error) {
	return map[string]interface{}{
		"raw_output": output,
	}, nil
}

func (t *ResolvectlTool) Query(ctx context.Context, domain string) (map[string]interface{}, error) {
	args := []string{"query", domain}

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

func (t *ResolvectlTool) Status(ctx context.Context) (map[string]interface{}, error) {
	args := []string{"status"}

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

// DNSVizTool wraps dnsviz for DNS delegation analysis
type DNSVizTool struct {
	*BaseTool
}

func NewDNSVizTool() *DNSVizTool {
	return &DNSVizTool{
		BaseTool: NewBaseTool(
			"dnsviz",
			"dnsviz",
			CategoryDNS,
			"apt-get install dnsviz",
			"DNS configuration and delegation analysis with DNSSEC validation",
		),
	}
}

func (t *DNSVizTool) ParseOutput(output string) (interface{}, error) {
	// DNSViz outputs JSON by default
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

func (t *DNSVizTool) Probe(ctx context.Context, domain string) (map[string]interface{}, error) {
	args := []string{"probe", domain}

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

// Register extended DNS tools
func RegisterDNSToolsExtended(registry *ToolRegistry) {
	RegisterDNSTools(registry) // Register existing tools
	registry.Register(NewNslookupTool())
	registry.Register(NewResolvectlTool())
	registry.Register(NewDNSVizTool())
}
