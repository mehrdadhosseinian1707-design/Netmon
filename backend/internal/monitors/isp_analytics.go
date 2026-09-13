package monitors

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/netmon/netmon/internal/models"
)

// ISPAnalyticsMonitor implements the Monitor interface for ISP intelligence analytics.
type ISPAnalyticsMonitor struct {
	config ispAnalyticsConfig
}

type ispAnalyticsConfig struct {
	TestURL              string
	TimeoutSeconds       float64
	SampleCount          int
	BaselineLatencyMs    float64
	BaselineLoss         float64
	BaselineJitterMs     float64
	BaselineSpeedMbps    float64
	BaselineStddevLatency float64
	PeakLatencyMs        float64
	OffpeakLatencyMs     float64
}

// NewISPAnalyticsMonitor creates a new ISPAnalyticsMonitor from a config map.
func NewISPAnalyticsMonitor(config map[string]interface{}) (*ISPAnalyticsMonitor, error) {
	cfg := ispAnalyticsConfig{
		TestURL:        "http://speed.cloudflare.com/__down?bytes=1000000",
		TimeoutSeconds: 30,
		SampleCount:    10,
	}
	if v, ok := config["test_url"].(string); ok && v != "" {
		cfg.TestURL = v
	}
	if v, ok := config["timeout_seconds"].(float64); ok && v > 0 {
		cfg.TimeoutSeconds = v
	}
	if v, ok := config["sample_count"].(float64); ok && v > 0 {
		cfg.SampleCount = int(v)
	}
	if v, ok := config["baseline_latency_ms"].(float64); ok {
		cfg.BaselineLatencyMs = v
	}
	if v, ok := config["baseline_loss"].(float64); ok {
		cfg.BaselineLoss = v
	}
	if v, ok := config["baseline_jitter_ms"].(float64); ok {
		cfg.BaselineJitterMs = v
	}
	if v, ok := config["baseline_speed_mbps"].(float64); ok {
		cfg.BaselineSpeedMbps = v
	}
	if v, ok := config["baseline_stddev_latency"].(float64); ok {
		cfg.BaselineStddevLatency = v
	}
	if v, ok := config["peak_latency_ms"].(float64); ok {
		cfg.PeakLatencyMs = v
	}
	if v, ok := config["offpeak_latency_ms"].(float64); ok {
		cfg.OffpeakLatencyMs = v
	}
	return &ISPAnalyticsMonitor{config: cfg}, nil
}

func (m *ISPAnalyticsMonitor) Type() string {
	return "isp_analytics"
}

func (m *ISPAnalyticsMonitor) ValidateConfig(config map[string]interface{}) error {
	if v, ok := config["timeout_seconds"].(float64); ok {
		if v < 1 || v > 300 {
			return fmt.Errorf("timeout_seconds must be between 1 and 300")
		}
	}
	if v, ok := config["sample_count"].(float64); ok {
		if v < 3 || v > 50 {
			return fmt.Errorf("sample_count must be between 3 and 50")
		}
	}
	return nil
}

// speedSample holds the result of a single HTTP download chunk measurement.
type speedSample struct {
	ThroughputMbps float64
	LatencyMs      float64
	Success        bool
}

// downloadChunk performs an HTTP GET and measures throughput and latency.
func downloadChunk(ctx context.Context, client *http.Client, url string) speedSample {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return speedSample{Success: false}
	}
	ttfb := time.Duration(0)
	resp, err := client.Do(req)
	if err != nil {
		return speedSample{Success: false}
	}
	defer resp.Body.Close()

	// Time-to-first-byte approximation: measure after Do() returns headers
	ttfb = time.Since(start)

	n, err := io.Copy(io.Discard, resp.Body)
	if err != nil || n == 0 {
		return speedSample{Success: false}
	}
	total := time.Since(start)
	seconds := total.Seconds()
	if seconds <= 0 {
		seconds = 0.001
	}
	mbps := (float64(n) * 8) / (seconds * 1e6)
	latMs := float64(ttfb.Milliseconds())
	return speedSample{ThroughputMbps: mbps, LatencyMs: latMs, Success: true}
}

// collectSamples gathers n speed samples and returns them in order.
func collectSamples(ctx context.Context, client *http.Client, url string, n int) []speedSample {
	samples := make([]speedSample, 0, n)
	for i := 0; i < n; i++ {
		s := downloadChunk(ctx, client, url)
		samples = append(samples, s)
		// Bail early if context is done
		select {
		case <-ctx.Done():
			return samples
		default:
		}
	}
	return samples
}

// avg returns the mean of a float64 slice; 0 if empty.
func avg(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// stddevWithMean returns the population standard deviation given a pre-computed mean.
func stddevWithMean(vals []float64, mean float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		d := v - mean
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}

// clamp restricts v to [lo, hi].
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ---- 1. ISP Congestion Detection ----

type congestionResult struct {
	Detected bool
	Trend    string // "stable" | "degrading" | "recovering"
}

func detectCongestion(samples []speedSample) congestionResult {
	speeds := make([]float64, 0, len(samples))
	for _, s := range samples {
		if s.Success {
			speeds = append(speeds, s.ThroughputMbps)
		}
	}

	res := congestionResult{Detected: false, Trend: "stable"}
	if len(speeds) < 6 {
		return res
	}

	first3 := avg(speeds[:3])
	last3 := avg(speeds[len(speeds)-3:])

	if first3 <= 0 {
		return res
	}

	ratio := last3 / first3
	if ratio < 0.8 { // >20% degradation
		res.Detected = true
		res.Trend = "degrading"
	} else if ratio > 1.1 {
		res.Trend = "recovering"
	} else {
		res.Trend = "stable"
	}
	return res
}

// ---- 2. Network Technology Detection ----

type techResult struct {
	Technology string
	Confidence float64
}

func detectNetworkTechnology(rttMs, jitterMs, speedMbps float64) techResult {
	type rule struct {
		tech       string
		confidence float64
		matches    func() bool
	}

	rules := []rule{
		{
			tech:       "fiber",
			confidence: 0.95,
			matches:    func() bool { return rttMs < 5 && speedMbps > 100 },
		},
		{
			tech:       "cable",
			confidence: 0.85,
			matches:    func() bool { return rttMs < 20 && speedMbps > 50 },
		},
		{
			tech:       "wifi",
			confidence: 0.70,
			matches:    func() bool { return rttMs < 5 && (speedMbps <= 100 || jitterMs > 2) },
		},
		{
			tech:       "4g/lte",
			confidence: 0.80,
			matches:    func() bool { return rttMs >= 20 && rttMs <= 80 && speedMbps >= 10 && speedMbps <= 50 },
		},
		{
			tech:       "3g",
			confidence: 0.75,
			matches:    func() bool { return rttMs > 80 && rttMs <= 200 && speedMbps >= 1 && speedMbps <= 10 },
		},
		{
			tech:       "2g",
			confidence: 0.70,
			matches:    func() bool { return rttMs > 200 || speedMbps < 1 },
		},
	}

	for _, r := range rules {
		if r.matches() {
			return techResult{Technology: r.tech, Confidence: r.confidence}
		}
	}
	// Fallback
	return techResult{Technology: "unknown", Confidence: 0.3}
}

// ---- 3. ISP Health Score ----

type healthResult struct {
	Score int
	Grade string
}

func computeHealthScore(latencyMs, packetLoss, jitterMs, speedMbps float64, congestionDetected bool, technology string) healthResult {
	score := 100.0

	if packetLoss > 5 {
		score -= 20 // >1% and >5% both apply: subtract 20+15=35 but we sequence them
		score -= 15
	} else if packetLoss > 1 {
		score -= 20
	}

	if latencyMs > 200 {
		score -= 10 // >100ms and >200ms
		score -= 10
	} else if latencyMs > 100 {
		score -= 10
	}

	if jitterMs > 20 {
		score -= 10
	}

	if congestionDetected {
		score -= 15
	}

	if (technology == "fiber" || technology == "cable") && speedMbps < 10 {
		score -= 20
	}

	s := int(clamp(score, 0, 100))
	var grade string
	switch {
	case s >= 90:
		grade = "A"
	case s >= 75:
		grade = "B"
	case s >= 60:
		grade = "C"
	case s >= 40:
		grade = "D"
	default:
		grade = "F"
	}
	return healthResult{Score: s, Grade: grade}
}

// ---- 4. Anomaly Detection ----

type anomalyResult struct {
	LatencyAnomaly string
	LossAnomaly    string
	JitterAnomaly  string
	SpeedAnomaly   string
	AnomalyScore   float64
}

func anomalyLevel(z float64) string {
	az := math.Abs(z)
	switch {
	case az > 3.0:
		return "anomaly"
	case az >= 2.0:
		return "warning"
	default:
		return "normal"
	}
}

func detectAnomalies(latencyMs, loss, jitterMs, speedMbps float64, cfg ispAnalyticsConfig) anomalyResult {
	res := anomalyResult{}

	safeDivStddev := cfg.BaselineStddevLatency
	if safeDivStddev == 0 {
		safeDivStddev = 1
	}

	zLatency := (latencyMs - cfg.BaselineLatencyMs) / safeDivStddev
	res.LatencyAnomaly = anomalyLevel(zLatency)

	// For loss, jitter, speed we use a fixed stddev approximation of 20% of baseline
	// when dedicated stddev is not configured.
	lossDev := cfg.BaselineLoss * 0.2
	if lossDev == 0 {
		lossDev = 0.5
	}
	zLoss := (loss - cfg.BaselineLoss) / lossDev
	res.LossAnomaly = anomalyLevel(zLoss)

	jitterDev := cfg.BaselineJitterMs * 0.2
	if jitterDev == 0 {
		jitterDev = 2
	}
	zJitter := (jitterMs - cfg.BaselineJitterMs) / jitterDev
	res.JitterAnomaly = anomalyLevel(zJitter)

	speedDev := cfg.BaselineSpeedMbps * 0.2
	if speedDev == 0 {
		speedDev = 1
	}
	// Speed: anomaly if much LOWER than baseline (negative z is significant)
	zSpeed := (cfg.BaselineSpeedMbps - speedMbps) / speedDev
	res.SpeedAnomaly = anomalyLevel(zSpeed)

	maxZ := math.Abs(zLatency)
	for _, z := range []float64{math.Abs(zLoss), math.Abs(zJitter), math.Abs(zSpeed)} {
		if z > maxZ {
			maxZ = z
		}
	}
	res.AnomalyScore = maxZ
	return res
}

// ---- 5. Peak vs Off-Peak Congestion Analysis ----

type peakResult struct {
	TimePeriod           string
	CongestionVsBaseline float64
	PeakDegradationRatio float64
}

func analyzePeakOffPeak(currentLatencyMs float64, cfg ispAnalyticsConfig) peakResult {
	hour := time.Now().Hour()
	period := "peak"
	if hour < 7 || hour >= 23 {
		period = "off_peak"
	}

	expectedLatency := cfg.PeakLatencyMs
	if period == "off_peak" {
		expectedLatency = cfg.OffpeakLatencyMs
	}

	congestionRatio := 1.0
	if expectedLatency > 0 {
		congestionRatio = currentLatencyMs / expectedLatency
	}

	peakDegRatio := 1.0
	if cfg.OffpeakLatencyMs > 0 && cfg.PeakLatencyMs > 0 {
		peakDegRatio = cfg.PeakLatencyMs / cfg.OffpeakLatencyMs
	}

	return peakResult{
		TimePeriod:           period,
		CongestionVsBaseline: congestionRatio,
		PeakDegradationRatio: peakDegRatio,
	}
}

// ---- 6. Measurement Confidence / Quality Score ----

type confidenceResult struct {
	Score  float64
	Reason string
}

func computeConfidence(samples []speedSample, packetLoss float64, testDuration time.Duration) confidenceResult {
	reasons := []string{}
	score := 1.0

	successCount := 0
	speeds := []float64{}
	for _, s := range samples {
		if s.Success {
			successCount++
			speeds = append(speeds, s.ThroughputMbps)
		}
	}

	// Sample completeness
	if len(samples) > 0 {
		completeness := float64(successCount) / float64(len(samples))
		if completeness < 1.0 {
			penalty := (1 - completeness) * 0.4
			score -= penalty
			reasons = append(reasons, fmt.Sprintf("%.0f%% sample failure rate", (1-completeness)*100))
		}
	}
	if successCount < 3 {
		score -= 0.3
		reasons = append(reasons, "fewer than 3 successful samples")
	}

	// Variance check
	if len(speeds) >= 2 {
		meanSpeed := avg(speeds)
		sd := stddevWithMean(speeds, meanSpeed)
		cv := 0.0
		if meanSpeed > 0 {
			cv = sd / meanSpeed
		}
		if cv > 0.5 {
			score -= 0.2
			reasons = append(reasons, "high speed variance")
		} else if cv > 0.3 {
			score -= 0.1
			reasons = append(reasons, "moderate speed variance")
		}
	}

	// Packet loss during test
	if packetLoss > 5 {
		score -= 0.2
		reasons = append(reasons, "high packet loss during test")
	} else if packetLoss > 1 {
		score -= 0.1
		reasons = append(reasons, "elevated packet loss during test")
	}

	// Test duration adequacy (expect at least 10s for 10 samples)
	minAdequate := 10 * time.Second
	if testDuration < minAdequate {
		score -= 0.1
		reasons = append(reasons, "test duration too short")
	}

	score = clamp(score, 0, 1)

	reason := "high quality measurement"
	if len(reasons) > 0 {
		reason = "reduced by: "
		for i, r := range reasons {
			if i > 0 {
				reason += "; "
			}
			reason += r
		}
	}
	return confidenceResult{Score: score, Reason: reason}
}

// ---- 7. Active-Test Scheduling Hint ----

type scheduleResult struct {
	IntervalSeconds int
	Reason          string
}

func computeSchedule(anomalyScore, confidence float64) scheduleResult {
	switch {
	case anomalyScore > 3.0:
		return scheduleResult{IntervalSeconds: 30, Reason: "anomaly detected - increasing test frequency"}
	case anomalyScore >= 2.0:
		return scheduleResult{IntervalSeconds: 60, Reason: "warning condition - elevated test frequency"}
	case confidence >= 0.8:
		return scheduleResult{IntervalSeconds: 300, Reason: "normal conditions with high confidence"}
	default:
		return scheduleResult{IntervalSeconds: 120, Reason: "normal conditions with low confidence - more data needed"}
	}
}

// ---- Check() - main entry point ----

func (m *ISPAnalyticsMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	timeout := time.Duration(m.config.TimeoutSeconds * float64(time.Second))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	client := &http.Client{
		Timeout: timeout,
	}

	testURL := m.config.TestURL
	if target != nil && target.Address != "" {
		// Allow target address to override test URL if it looks like a URL
		if len(target.Address) > 4 && (target.Address[:4] == "http") {
			testURL = target.Address
		}
	}

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "isp_analytics",
		Success:     false,
		Metadata:    make(map[string]interface{}),
	}

	testStart := time.Now()

	// Collect samples for congestion detection (up to SampleCount)
	samples := collectSamples(ctx, client, testURL, m.config.SampleCount)

	testDuration := time.Since(testStart)

	// Collect latency values and speed values from successful samples
	successSpeeds := []float64{}
	successLatencies := []float64{}
	for _, s := range samples {
		if s.Success {
			successSpeeds = append(successSpeeds, s.ThroughputMbps)
			successLatencies = append(successLatencies, s.LatencyMs)
		}
	}

	if len(successSpeeds) == 0 {
		measurement.ErrorMessage = "all speed samples failed"
		return measurement, nil
	}

	avgSpeed := avg(successSpeeds)
	avgLatency := avg(successLatencies)

	// Jitter = stddev of latencies
	jitter := stddevWithMean(successLatencies, avgLatency)

	// Packet loss: fraction of failed samples expressed as percentage
	failCount := 0
	for _, s := range samples {
		if !s.Success {
			failCount++
		}
	}
	packetLoss := 0.0
	if len(samples) > 0 {
		packetLoss = float64(failCount) / float64(len(samples)) * 100.0
	}

	// ---- 1. Congestion Detection ----
	congestion := detectCongestion(samples)
	measurement.Metadata["isp_congestion_detected"] = congestion.Detected
	measurement.Metadata["congestion_trend"] = congestion.Trend

	// ---- 2. Network Technology Detection ----
	// Use 3 fresh samples for tech detection (reuse already collected latencies/speeds)
	techLatency := avgLatency
	techSpeed := avgSpeed
	// If we have at least 3 samples, average first 3 for tech baseline
	if len(successLatencies) >= 3 {
		techLatency = avg(successLatencies[:3])
		techSpeed = avg(successSpeeds[:3])
	}
	tech := detectNetworkTechnology(techLatency, jitter, techSpeed)
	measurement.Metadata["network_technology"] = tech.Technology
	measurement.Metadata["tech_confidence"] = tech.Confidence

	// ---- 3. ISP Health Score ----
	health := computeHealthScore(avgLatency, packetLoss, jitter, avgSpeed, congestion.Detected, tech.Technology)
	measurement.Metadata["isp_health_score"] = health.Score
	measurement.Metadata["health_grade"] = health.Grade

	// ---- 4. Anomaly Detection ----
	anomalies := detectAnomalies(avgLatency, packetLoss, jitter, avgSpeed, m.config)
	measurement.Metadata["latency_anomaly"] = anomalies.LatencyAnomaly
	measurement.Metadata["loss_anomaly"] = anomalies.LossAnomaly
	measurement.Metadata["jitter_anomaly"] = anomalies.JitterAnomaly
	measurement.Metadata["speed_anomaly"] = anomalies.SpeedAnomaly
	measurement.Metadata["anomaly_score"] = anomalies.AnomalyScore

	// ---- 5. Peak vs Off-Peak ----
	peak := analyzePeakOffPeak(avgLatency, m.config)
	measurement.Metadata["time_period"] = peak.TimePeriod
	measurement.Metadata["congestion_vs_baseline"] = peak.CongestionVsBaseline
	measurement.Metadata["peak_degradation_ratio"] = peak.PeakDegradationRatio

	// ---- 6. Confidence Score ----
	conf := computeConfidence(samples, packetLoss, testDuration)
	measurement.Metadata["measurement_confidence"] = conf.Score
	measurement.Metadata["confidence_reason"] = conf.Reason

	// ---- 7. Scheduling Hint ----
	sched := computeSchedule(anomalies.AnomalyScore, conf.Score)
	measurement.Metadata["recommended_interval_seconds"] = sched.IntervalSeconds
	measurement.Metadata["schedule_reason"] = sched.Reason

	// Additional diagnostics
	sortedSpeeds := make([]float64, len(successSpeeds))
	copy(sortedSpeeds, successSpeeds)
	sort.Float64s(sortedSpeeds)
	if len(sortedSpeeds) > 0 {
		measurement.Metadata["speed_min_mbps"] = sortedSpeeds[0]
		measurement.Metadata["speed_max_mbps"] = sortedSpeeds[len(sortedSpeeds)-1]
		measurement.Metadata["speed_stddev_mbps"] = stddevWithMean(successSpeeds, avgSpeed)
	}
	measurement.Metadata["samples_collected"] = len(successSpeeds)
	measurement.Metadata["samples_attempted"] = len(samples)
	measurement.Metadata["test_duration_ms"] = testDuration.Milliseconds()

	// Set standard Measurement fields
	measurement.Success = true
	latPtr := avgLatency
	measurement.LatencyMs = &latPtr

	lossPtr := packetLoss
	measurement.PacketLoss = &lossPtr

	jitterPtr := jitter
	measurement.JitterMs = &jitterPtr

	speedPtr := avgSpeed
	measurement.DownloadSpeedMbps = &speedPtr

	return measurement, nil
}
