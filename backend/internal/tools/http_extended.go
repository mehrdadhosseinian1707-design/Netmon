package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// WgetTool wraps the wget command
type WgetTool struct {
	*BaseTool
}

func NewWgetTool() *WgetTool {
	return &WgetTool{
		BaseTool: NewBaseTool(
			"wget",
			"wget",
			CategoryHTTP,
			"apt-get install wget",
			"HTTP/HTTPS downloader with timing and retry capabilities",
		),
	}
}

type WgetResult struct {
	URL              string  `json:"url"`
	HTTPCode         int     `json:"http_code"`
	Size             int64   `json:"size_bytes"`
	DownloadSpeed    float64 `json:"download_speed_mbps"`
	TotalTime        float64 `json:"total_time_s"`
	DNSTime          float64 `json:"dns_time_s"`
	ConnectTime      float64 `json:"connect_time_s"`
	Success          bool    `json:"success"`
	RedirectURL      string  `json:"redirect_url,omitempty"`
	ErrorMessage     string  `json:"error_message,omitempty"`
}

func (t *WgetTool) ParseOutput(output string) (interface{}, error) {
	result := &WgetResult{}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Parse HTTP response code
		if strings.Contains(line, "HTTP request sent") || strings.Contains(line, "HTTP/") {
			codeRegex := regexp.MustCompile(`HTTP/[\d.]+\s+(\d+)`)
			if match := codeRegex.FindStringSubmatch(line); len(match) > 1 {
				result.HTTPCode, _ = strconv.Atoi(match[1])
			}
		}

		// Parse download speed
		if strings.Contains(line, "saved") {
			// Format: "2023-01-01 12:00:00 (1.23 MB/s) - 'file' saved [12345/12345]"
			speedRegex := regexp.MustCompile(`\(([\d.]+)\s+(MB|KB|B)/s\)`)
			if match := speedRegex.FindStringSubmatch(line); len(match) > 2 {
				speed, _ := strconv.ParseFloat(match[1], 64)
				unit := match[2]

				// Convert to Mbps
				switch unit {
				case "MB":
					result.DownloadSpeed = speed * 8
				case "KB":
					result.DownloadSpeed = speed * 8 / 1024
				case "B":
					result.DownloadSpeed = speed * 8 / 1024 / 1024
				}
			}

			// Parse size
			sizeRegex := regexp.MustCompile(`saved\s+\[(\d+)/(\d+)\]`)
			if match := sizeRegex.FindStringSubmatch(line); len(match) > 2 {
				result.Size, _ = strconv.ParseInt(match[1], 10, 64)
			}
		}

		// Parse redirect
		if strings.Contains(line, "Location:") {
			parts := strings.SplitN(line, "Location:", 2)
			if len(parts) > 1 {
				result.RedirectURL = strings.TrimSpace(parts[1])
			}
		}
	}

	result.Success = result.HTTPCode >= 200 && result.HTTPCode < 400
	return result, nil
}

func (t *WgetTool) Download(ctx context.Context, url string, output string) (*WgetResult, error) {
	args := []string{"--timeout=30", "-O", output}

	if output == "/dev/null" || output == "NUL" {
		args = []string{"--timeout=30", "-O", "/dev/null"}
	}

	args = append(args, url)

	result, err := t.Execute(ctx, args)

	parsed, parseErr := t.ParseOutput(result.Output)
	if parseErr != nil {
		return nil, parseErr
	}

	wgetResult := parsed.(*WgetResult)
	wgetResult.URL = url

	if err != nil && !wgetResult.Success {
		wgetResult.ErrorMessage = err.Error()
	}

	return wgetResult, nil
}

// XHTool wraps the xh command (modern HTTP client)
type XHTool struct {
	*BaseTool
}

func NewXHTool() *XHTool {
	return &XHTool{
		BaseTool: NewBaseTool(
			"xh",
			"xh",
			CategoryHTTP,
			"cargo install xh",
			"Modern, user-friendly HTTP client with JSON support",
		),
	}
}

func (t *XHTool) ParseOutput(output string) (interface{}, error) {
	// xh can output JSON
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

func (t *XHTool) Request(ctx context.Context, method string, url string, headers map[string]string) (map[string]interface{}, error) {
	args := []string{method, url, "--print=hb"}

	for key, value := range headers {
		args = append(args, fmt.Sprintf("%s:%s", key, value))
	}

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

// GnuTLSCliTool wraps gnutls-cli for TLS testing
type GnuTLSCliTool struct {
	*BaseTool
}

func NewGnuTLSCliTool() *GnuTLSCliTool {
	return &GnuTLSCliTool{
		BaseTool: NewBaseTool(
			"gnutls-cli",
			"gnutls-cli",
			CategoryHTTP,
			"apt-get install gnutls-bin",
			"GnuTLS command-line TLS client for certificate and handshake testing",
		),
	}
}

type GnuTLSResult struct {
	Connected       bool              `json:"connected"`
	TLSVersion      string            `json:"tls_version"`
	Cipher          string            `json:"cipher"`
	KeyExchange     string            `json:"key_exchange"`
	Certificate     map[string]string `json:"certificate"`
	HandshakeTime   float64           `json:"handshake_time_ms"`
	ErrorMessage    string            `json:"error_message,omitempty"`
}

func (t *GnuTLSCliTool) ParseOutput(output string) (interface{}, error) {
	result := &GnuTLSResult{
		Certificate: make(map[string]string),
	}

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check connection status
		if strings.Contains(line, "Handshake was completed") {
			result.Connected = true
		}

		// Parse TLS version
		if strings.Contains(line, "Version:") {
			versionRegex := regexp.MustCompile(`Version:\s+(.+)`)
			if match := versionRegex.FindStringSubmatch(line); len(match) > 1 {
				result.TLSVersion = match[1]
			}
		}

		// Parse cipher
		if strings.Contains(line, "Cipher:") {
			cipherRegex := regexp.MustCompile(`Cipher:\s+(.+)`)
			if match := cipherRegex.FindStringSubmatch(line); len(match) > 1 {
				result.Cipher = match[1]
			}
		}

		// Parse key exchange
		if strings.Contains(line, "Key Exchange:") {
			kexRegex := regexp.MustCompile(`Key Exchange:\s+(.+)`)
			if match := kexRegex.FindStringSubmatch(line); len(match) > 1 {
				result.KeyExchange = match[1]
			}
		}

		// Parse certificate info
		if strings.Contains(line, "Subject:") {
			result.Certificate["subject"] = strings.TrimSpace(strings.SplitN(line, "Subject:", 2)[1])
		}

		if strings.Contains(line, "Issuer:") {
			result.Certificate["issuer"] = strings.TrimSpace(strings.SplitN(line, "Issuer:", 2)[1])
		}
	}

	return result, nil
}

func (t *GnuTLSCliTool) TestTLS(ctx context.Context, host string, port int) (*GnuTLSResult, error) {
	args := []string{
		"--port", strconv.Itoa(port),
		host,
	}

	result, err := t.Execute(ctx, args)

	parsed, parseErr := t.ParseOutput(result.Output)
	if parseErr != nil {
		return nil, parseErr
	}

	tlsResult := parsed.(*GnuTLSResult)

	if err != nil && !tlsResult.Connected {
		tlsResult.ErrorMessage = err.Error()
	}

	return tlsResult, nil
}

// Register extended HTTP/TLS tools
func RegisterHTTPToolsExtended(registry *ToolRegistry) {
	RegisterHTTPTools(registry) // Register existing tools
	registry.Register(NewWgetTool())
	registry.Register(NewXHTool())
	registry.Register(NewGnuTLSCliTool())
}
