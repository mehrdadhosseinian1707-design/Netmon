package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// CurlTool wraps the curl command
type CurlTool struct {
	*BaseTool
}

func NewCurlTool() *CurlTool {
	return &CurlTool{
		BaseTool: NewBaseTool(
			"curl",
			"curl",
			CategoryHTTP,
			"apt-get install curl",
			"HTTP/HTTPS transfer tool with detailed timing",
		),
	}
}

type CurlResult struct {
	URL              string  `json:"url"`
	HTTPCode         int     `json:"http_code"`
	Size             int64   `json:"size_bytes"`
	TimeNameLookup   float64 `json:"time_namelookup_s"`
	TimeConnect      float64 `json:"time_connect_s"`
	TimeAppConnect   float64 `json:"time_appconnect_s"`
	TimePreTransfer  float64 `json:"time_pretransfer_s"`
	TimeStartTransfer float64 `json:"time_starttransfer_s"`
	TimeTotal        float64 `json:"time_total_s"`
	SpeedDownload    float64 `json:"speed_download_bps"`
	SpeedUpload      float64 `json:"speed_upload_bps"`
}

func (t *CurlTool) ParseOutput(output string) (interface{}, error) {
	// When using -w flag with format string, curl outputs timing data
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		// If not JSON, return raw output
		return map[string]interface{}{"raw": output}, nil
	}
	return result, nil
}

func (t *CurlTool) Curl(ctx context.Context, url string, method string, headers map[string]string) (*CurlResult, error) {
	// Use curl with timing format output
	formatStr := `{
		"url": "%{url}",
		"http_code": %{http_code},
		"size": %{size_download},
		"time_namelookup": %{time_namelookup},
		"time_connect": %{time_connect},
		"time_appconnect": %{time_appconnect},
		"time_pretransfer": %{time_pretransfer},
		"time_starttransfer": %{time_starttransfer},
		"time_total": %{time_total},
		"speed_download": %{speed_download},
		"speed_upload": %{speed_upload}
	}`

	args := []string{"-s", "-o", "/dev/null", "-w", formatStr}

	if method != "" && method != "GET" {
		args = append(args, "-X", method)
	}

	for key, value := range headers {
		args = append(args, "-H", fmt.Sprintf("%s: %s", key, value))
	}

	args = append(args, url)

	result, err := t.Execute(ctx, args)
	if err != nil {
		return nil, err
	}

	var curlResult CurlResult
	if err := json.Unmarshal([]byte(result.Output), &curlResult); err != nil {
		return nil, fmt.Errorf("failed to parse curl output: %w", err)
	}

	return &curlResult, nil
}

// HTTPingTool wraps the httping command
type HTTPingTool struct {
	*BaseTool
}

func NewHTTPingTool() *HTTPingTool {
	return &HTTPingTool{
		BaseTool: NewBaseTool(
			"httping",
			"httping",
			CategoryHTTP,
			"apt-get install httping",
			"HTTP latency measurement tool",
		),
	}
}

type HTTPingResult struct {
	URL       string  `json:"url"`
	Count     int     `json:"count"`
	MinTime   float64 `json:"min_time_ms"`
	AvgTime   float64 `json:"avg_time_ms"`
	MaxTime   float64 `json:"max_time_ms"`
	StdDev    float64 `json:"stddev_ms"`
	Succeeded int     `json:"succeeded"`
	Failed    int     `json:"failed"`
}

func (t *HTTPingTool) ParseOutput(output string) (interface{}, error) {
	result := &HTTPingResult{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Parse statistics line
		// "--- http://example.com ping statistics ---"
		// "5 connects, 5 ok, 0.00% failed, time 5023ms"
		// "round-trip min/avg/max = 234.5/256.3/289.1 ms"

		if strings.Contains(line, "connects") {
			statsRegex := regexp.MustCompile(`(\d+)\s+connects,\s+(\d+)\s+ok,\s+([\d.]+)%\s+failed`)
			if match := statsRegex.FindStringSubmatch(line); len(match) > 3 {
				result.Count, _ = strconv.Atoi(match[1])
				result.Succeeded, _ = strconv.Atoi(match[2])
				result.Failed = result.Count - result.Succeeded
			}
		}

		if strings.Contains(line, "round-trip min/avg/max") {
			rttRegex := regexp.MustCompile(`min/avg/max\s+=\s+([\d.]+)/([\d.]+)/([\d.]+)`)
			if match := rttRegex.FindStringSubmatch(line); len(match) > 3 {
				result.MinTime, _ = strconv.ParseFloat(match[1], 64)
				result.AvgTime, _ = strconv.ParseFloat(match[2], 64)
				result.MaxTime, _ = strconv.ParseFloat(match[3], 64)
			}
		}
	}

	return result, nil
}

func (t *HTTPingTool) HTTPing(ctx context.Context, url string, count int) (*HTTPingResult, error) {
	args := []string{"-c", strconv.Itoa(count), "-g", url}

	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	httpingResult := parsed.(*HTTPingResult)
	httpingResult.URL = url
	return httpingResult, nil
}

// OpenSSLTool wraps openssl s_client for TLS analysis
type OpenSSLTool struct {
	*BaseTool
}

func NewOpenSSLTool() *OpenSSLTool {
	return &OpenSSLTool{
		BaseTool: NewBaseTool(
			"openssl",
			"openssl",
			CategoryHTTP,
			"apt-get install openssl",
			"TLS/SSL analysis and certificate inspection",
		),
	}
}

type TLSInfo struct {
	Protocol        string            `json:"protocol"`
	Cipher          string            `json:"cipher"`
	ServerCert      map[string]string `json:"server_cert"`
	CertChain       []string          `json:"cert_chain"`
	ValidFrom       string            `json:"valid_from"`
	ValidTo         string            `json:"valid_to"`
	Issuer          string            `json:"issuer"`
	Subject         string            `json:"subject"`
	SubjectAltNames []string          `json:"subject_alt_names"`
}

func (t *OpenSSLTool) ParseOutput(output string) (interface{}, error) {
	info := &TLSInfo{
		ServerCert:      make(map[string]string),
		CertChain:       make([]string, 0),
		SubjectAltNames: make([]string, 0),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Protocol
		if strings.Contains(line, "Protocol  :") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info.Protocol = strings.TrimSpace(parts[1])
			}
		}

		// Cipher
		if strings.Contains(line, "Cipher    :") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				info.Cipher = strings.TrimSpace(parts[1])
			}
		}

		// Subject
		if strings.HasPrefix(line, "subject=") {
			info.Subject = strings.TrimPrefix(line, "subject=")
		}

		// Issuer
		if strings.HasPrefix(line, "issuer=") {
			info.Issuer = strings.TrimPrefix(line, "issuer=")
		}

		// Validity dates
		if strings.Contains(line, "notBefore=") {
			dateRegex := regexp.MustCompile(`notBefore=(.+)`)
			if match := dateRegex.FindStringSubmatch(line); len(match) > 1 {
				info.ValidFrom = match[1]
			}
		}

		if strings.Contains(line, "notAfter=") {
			dateRegex := regexp.MustCompile(`notAfter=(.+)`)
			if match := dateRegex.FindStringSubmatch(line); len(match) > 1 {
				info.ValidTo = match[1]
			}
		}
	}

	return info, nil
}

func (t *OpenSSLTool) AnalyzeTLS(ctx context.Context, host string, port int) (*TLSInfo, error) {
	args := []string{
		"s_client",
		"-connect", fmt.Sprintf("%s:%d", host, port),
		"-servername", host,
		"-showcerts",
	}

	// Send empty input to close connection quickly
	result, err := t.Execute(ctx, args)
	if err != nil && result == nil {
		return nil, err
	}

	parsed, err := t.ParseOutput(result.Output)
	if err != nil {
		return nil, err
	}

	return parsed.(*TLSInfo), nil
}

// Register all HTTP/TLS tools
func RegisterHTTPTools(registry *ToolRegistry) {
	registry.Register(NewCurlTool())
	registry.Register(NewHTTPingTool())
	registry.Register(NewOpenSSLTool())
}
