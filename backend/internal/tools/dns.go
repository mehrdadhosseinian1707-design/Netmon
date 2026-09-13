package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// DigTool wraps the dig command
type DigTool struct {
	*BaseTool
}

func NewDigTool() *DigTool {
	return &DigTool{
		BaseTool: NewBaseTool(
			"dig",
			"dig",
			CategoryDNS,
			"apt-get install dnsutils",
			"DNS lookup and analysis tool",
		),
	}
}

type DigAnswer struct {
	Name  string `json:"name"`
	TTL   int    `json:"ttl"`
	Class string `json:"class"`
	Type  string `json:"type"`
	Data  string `json:"data"`
}

type DigResult struct {
	Question    string      `json:"question"`
	QueryType   string      `json:"query_type"`
	Answers     []DigAnswer `json:"answers"`
	Authority   []DigAnswer `json:"authority,omitempty"`
	Additional  []DigAnswer `json:"additional,omitempty"`
	QueryTime   int         `json:"query_time_ms"`
	Server      string      `json:"server"`
	Size        int         `json:"size_bytes"`
	Flags       []string    `json:"flags"`
	Status      string      `json:"status"`
}

func (t *DigTool) ParseOutput(output string) (interface{}, error) {
	result := &DigResult{
		Answers:    make([]DigAnswer, 0),
		Authority:  make([]DigAnswer, 0),
		Additional: make([]DigAnswer, 0),
		Flags:      make([]string, 0),
	}

	lines := strings.Split(output, "\n")
	section := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Parse query time
		if strings.Contains(line, "Query time:") {
			timeRegex := regexp.MustCompile(`Query time:\s+(\d+)\s+msec`)
			if match := timeRegex.FindStringSubmatch(line); len(match) > 1 {
				result.QueryTime, _ = strconv.Atoi(match[1])
			}
		}

		// Parse server
		if strings.Contains(line, "SERVER:") {
			serverRegex := regexp.MustCompile(`SERVER:\s+([^\s#]+)`)
			if match := serverRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Server = match[1]
			}
		}

		// Parse size
		if strings.Contains(line, "MSG SIZE") {
			sizeRegex := regexp.MustCompile(`rcvd:\s+(\d+)`)
			if match := sizeRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Size, _ = strconv.Atoi(match[1])
			}
		}

		// Parse status
		if strings.Contains(line, "status:") {
			statusRegex := regexp.MustCompile(`status:\s+(\w+)`)
			if match := statusRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Status = match[1]
			}
		}

		// Parse flags
		if strings.Contains(line, "flags:") {
			flagRegex := regexp.MustCompile(`flags:\s+([^;]+)`)
			if match := flagRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Flags = strings.Fields(match[1])
			}
		}

		// Detect sections
		if strings.HasPrefix(line, ";; ANSWER SECTION:") {
			section = "answer"
			continue
		} else if strings.HasPrefix(line, ";; AUTHORITY SECTION:") {
			section = "authority"
			continue
		} else if strings.HasPrefix(line, ";; ADDITIONAL SECTION:") {
			section = "additional"
			continue
		} else if strings.HasPrefix(line, ";;") {
			section = ""
			continue
		}

		// Parse answer records
		if section != "" && !strings.HasPrefix(line, ";") && line != "" {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				answer := DigAnswer{
					Name:  fields[0],
					Class: fields[2],
					Type:  fields[3],
					Data:  strings.Join(fields[4:], " "),
				}
				answer.TTL, _ = strconv.Atoi(fields[1])

				switch section {
				case "answer":
					result.Answers = append(result.Answers, answer)
				case "authority":
					result.Authority = append(result.Authority, answer)
				case "additional":
					result.Additional = append(result.Additional, answer)
				}
			}
		}
	}

	return result, nil
}

func (t *DigTool) Dig(ctx context.Context, domain string, recordType string, server string) (*DigResult, error) {
	args := []string{domain}
	if recordType != "" {
		args = append(args, recordType)
	}
	if server != "" {
		args = append(args, "@"+server)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*DigResult), nil
}

func (t *DigTool) DigJSON(ctx context.Context, domain string, recordType string, server string) (map[string]interface{}, error) {
	args := []string{"+json", domain}
	if recordType != "" {
		args = append(args, recordType)
	}
	if server != "" {
		args = append(args, "@"+server)
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

// DrillTool wraps the drill command
type DrillTool struct {
	*BaseTool
}

func NewDrillTool() *DrillTool {
	return &DrillTool{
		BaseTool: NewBaseTool(
			"drill",
			"drill",
			CategoryDNS,
			"apt-get install ldnsutils",
			"DNS lookup tool similar to dig",
		),
	}
}

func (t *DrillTool) ParseOutput(output string) (interface{}, error) {
	// Drill output is similar to dig
	digTool := NewDigTool()
	return digTool.ParseOutput(output)
}

func (t *DrillTool) Drill(ctx context.Context, domain string, recordType string, server string) (*DigResult, error) {
	args := []string{domain}
	if recordType != "" {
		args = append(args, recordType)
	}
	if server != "" {
		args = append(args, "@"+server)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*DigResult), nil
}

// KDigTool wraps the kdig command (Knot DNS)
type KDigTool struct {
	*BaseTool
}

func NewKDigTool() *KDigTool {
	return &KDigTool{
		BaseTool: NewBaseTool(
			"kdig",
			"kdig",
			CategoryDNS,
			"apt-get install knot-dnsutils",
			"Advanced DNS lookup tool from Knot DNS",
		),
	}
}

func (t *KDigTool) ParseOutput(output string) (interface{}, error) {
	digTool := NewDigTool()
	return digTool.ParseOutput(output)
}

func (t *KDigTool) KDig(ctx context.Context, domain string, recordType string, server string) (*DigResult, error) {
	args := []string{domain}
	if recordType != "" {
		args = append(args, recordType)
	}
	if server != "" {
		args = append(args, "@"+server)
	}

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*DigResult), nil
}

// Register all DNS tools
func RegisterDNSTools(registry *ToolRegistry) {
	registry.Register(NewDigTool())
	registry.Register(NewDrillTool())
	registry.Register(NewKDigTool())
}
