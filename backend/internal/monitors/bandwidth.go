package monitors

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/netmon/netmon/internal/models"
)

type BandwidthMonitor struct {
	config BandwidthConfig
}

type BandwidthConfig struct {
	DownloadURL     string
	UploadURL       string
	TestDuration    time.Duration
	SampleSize      int64
	Timeout         time.Duration
	MeasureBoth     bool
}

type BandwidthResult struct {
	Success           bool
	DownloadSpeedMbps float64
	UploadSpeedMbps   float64
	DownloadBytes     int64
	UploadBytes       int64
	DownloadDuration  time.Duration
	UploadDuration    time.Duration
	ErrorMessage      string
	Metadata          map[string]interface{}
}

func NewBandwidthMonitor(config map[string]interface{}) (*BandwidthMonitor, error) {
	cfg := BandwidthConfig{
		TestDuration: 10 * time.Second,
		SampleSize:   10 * 1024 * 1024, // 10MB default
		Timeout:      30 * time.Second,
		MeasureBoth:  true,
	}

	if downloadURL, ok := config["download_url"].(string); ok {
		cfg.DownloadURL = downloadURL
	}
	if uploadURL, ok := config["upload_url"].(string); ok {
		cfg.UploadURL = uploadURL
	}
	if duration, ok := config["test_duration"].(float64); ok {
		cfg.TestDuration = time.Duration(duration) * time.Second
	}
	if size, ok := config["sample_size"].(float64); ok {
		cfg.SampleSize = int64(size)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if measureBoth, ok := config["measure_both"].(bool); ok {
		cfg.MeasureBoth = measureBoth
	}

	return &BandwidthMonitor{config: cfg}, nil
}

func (m *BandwidthMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.measureBandwidth(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeBandwidth,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		if result.DownloadSpeedMbps > 0 {
			measurement.DownloadSpeedMbps = &result.DownloadSpeedMbps
			measurement.Metadata["download_bytes"] = result.DownloadBytes
			measurement.Metadata["download_duration_ms"] = result.DownloadDuration.Milliseconds()
		}
		if result.UploadSpeedMbps > 0 {
			measurement.UploadSpeedMbps = &result.UploadSpeedMbps
			measurement.Metadata["upload_bytes"] = result.UploadBytes
			measurement.Metadata["upload_duration_ms"] = result.UploadDuration.Milliseconds()
		}
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *BandwidthMonitor) measureBandwidth(ctx context.Context, baseURL string) BandwidthResult {
	result := BandwidthResult{
		Success: false,
	}

	// Ensure URL has scheme
	if len(baseURL) > 0 && baseURL[0] != 'h' {
		baseURL = "http://" + baseURL
	}

	// Measure download speed
	if m.config.DownloadURL != "" || m.config.MeasureBoth {
		downloadURL := m.config.DownloadURL
		if downloadURL == "" {
			downloadURL = baseURL + "/download"
		}

		downloadSpeed, bytes, duration, err := m.measureDownload(ctx, downloadURL)
		if err != nil {
			result.ErrorMessage = fmt.Sprintf("download test failed: %v", err)
			return result
		}
		result.DownloadSpeedMbps = downloadSpeed
		result.DownloadBytes = bytes
		result.DownloadDuration = duration
	}

	// Measure upload speed
	if m.config.UploadURL != "" || m.config.MeasureBoth {
		uploadURL := m.config.UploadURL
		if uploadURL == "" {
			uploadURL = baseURL + "/upload"
		}

		uploadSpeed, bytes, duration, err := m.measureUpload(ctx, uploadURL)
		if err != nil {
			if result.DownloadSpeedMbps == 0 {
				result.ErrorMessage = fmt.Sprintf("upload test failed: %v", err)
				return result
			}
			// If download succeeded but upload failed, still mark as partial success
			result.Metadata = map[string]interface{}{
				"upload_error": err.Error(),
			}
		} else {
			result.UploadSpeedMbps = uploadSpeed
			result.UploadBytes = bytes
			result.UploadDuration = duration
		}
	}

	result.Success = result.DownloadSpeedMbps > 0 || result.UploadSpeedMbps > 0
	return result
}

func (m *BandwidthMonitor) measureDownload(ctx context.Context, url string) (float64, int64, time.Duration, error) {
	client := &http.Client{
		Timeout: m.config.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to create request: %w", err)
	}

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read response body and measure
	bytesRead := int64(0)
	buffer := make([]byte, 32*1024) // 32KB buffer

	for {
		n, err := resp.Body.Read(buffer)
		bytesRead += int64(n)

		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, 0, 0, fmt.Errorf("read error: %w", err)
		}

		// Stop after test duration or sample size
		if time.Since(startTime) > m.config.TestDuration || bytesRead >= m.config.SampleSize {
			break
		}
	}

	duration := time.Since(startTime)
	if duration == 0 {
		return 0, bytesRead, duration, fmt.Errorf("test duration too short")
	}

	// Calculate speed in Mbps
	speedMbps := (float64(bytesRead) * 8) / (float64(duration.Milliseconds()) / 1000.0) / 1_000_000

	return speedMbps, bytesRead, duration, nil
}

func (m *BandwidthMonitor) measureUpload(ctx context.Context, url string) (float64, int64, time.Duration, error) {
	// Create a reader that generates data
	dataSize := m.config.SampleSize
	dataReader := &limitedReader{
		size:      dataSize,
		remaining: dataSize,
	}

	client := &http.Client{
		Timeout: m.config.Timeout,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, dataReader)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.ContentLength = dataSize

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)
	bytesUploaded := dataSize - dataReader.remaining

	if duration == 0 {
		return 0, bytesUploaded, duration, fmt.Errorf("test duration too short")
	}

	// Calculate speed in Mbps
	speedMbps := (float64(bytesUploaded) * 8) / (float64(duration.Milliseconds()) / 1000.0) / 1_000_000

	return speedMbps, bytesUploaded, duration, nil
}

func (m *BandwidthMonitor) Type() string {
	return models.MonitorTypeBandwidth
}

func (m *BandwidthMonitor) ValidateConfig(config map[string]interface{}) error {
	if duration, ok := config["test_duration"].(float64); ok {
		if duration < 1 || duration > 300 {
			return fmt.Errorf("test_duration must be between 1 and 300 seconds")
		}
	}
	if size, ok := config["sample_size"].(float64); ok {
		if size < 1024 || size > 1024*1024*1024 {
			return fmt.Errorf("sample_size must be between 1KB and 1GB")
		}
	}
	return nil
}

// limitedReader generates dummy data for upload tests
type limitedReader struct {
	size      int64
	remaining int64
}

func (r *limitedReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}

	n = len(p)
	if int64(n) > r.remaining {
		n = int(r.remaining)
	}

	// Fill with zeros (could use random data if needed)
	for i := 0; i < n; i++ {
		p[i] = 0
	}

	r.remaining -= int64(n)
	return n, nil
}
