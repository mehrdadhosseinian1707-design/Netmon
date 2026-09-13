package monitors

import (
	"context"
	"fmt"
	"math"
	"net"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/netmon/netmon/internal/models"
)

// Internet-wide baseline percentiles (realistic 2024 global stats)
const (
	baselineLatencyP50 = 35.0  // ms
	baselineLatencyP95 = 180.0 // ms
	baselineLossP50    = 0.1   // %
	baselineLossP95    = 2.5   // %
	baselineJitterP50  = 5.0   // ms
	baselineJitterP95  = 30.0  // ms
)

// SLACorrelationMonitor implements cross-ISP correlation and SLA prediction.
type SLACorrelationMonitor struct {
	host             string
	port             int
	timeoutSeconds   float64
	slaLatencyMs     float64
	slaLossPercent   float64
	slaAvailability  float64
	recentLatencies  []float64
	recentLosses     []float64
	hopLatencies     []float64
	dnsLatencyMs     float64
	tlsLatencyMs     float64
	tcpConnectMs     float64
	ispName          string
	destination      string
}

func NewSLACorrelationMonitor(config map[string]interface{}) (*SLACorrelationMonitor, error) {
	m := &SLACorrelationMonitor{
		port:            443,
		timeoutSeconds:  20,
		slaLatencyMs:    100.0,
		slaLossPercent:  1.0,
		slaAvailability: 99.9,
	}

	if host, ok := config["host"].(string); ok {
		m.host = host
	}
	if port, ok := config["port"].(float64); ok {
		m.port = int(port)
	}
	if to, ok := config["timeout_seconds"].(float64); ok {
		m.timeoutSeconds = to
	}
	if v, ok := config["sla_latency_ms"].(float64); ok {
		m.slaLatencyMs = v
	}
	if v, ok := config["sla_loss_percent"].(float64); ok {
		m.slaLossPercent = v
	}
	if v, ok := config["sla_availability_percent"].(float64); ok {
		m.slaAvailability = v
	}
	if v, ok := config["dns_latency_ms"].(float64); ok {
		m.dnsLatencyMs = v
	}
	if v, ok := config["tls_latency_ms"].(float64); ok {
		m.tlsLatencyMs = v
	}
	if v, ok := config["tcp_connect_ms"].(float64); ok {
		m.tcpConnectMs = v
	}
	if v, ok := config["isp_name"].(string); ok {
		m.ispName = v
	}
	if v, ok := config["destination"].(string); ok {
		m.destination = v
	}

	// Parse recent_latencies
	if raw, ok := config["recent_latencies"].([]interface{}); ok {
		for _, item := range raw {
			if f, ok := item.(float64); ok {
				m.recentLatencies = append(m.recentLatencies, f)
			}
		}
	}
	// Parse recent_losses
	if raw, ok := config["recent_losses"].([]interface{}); ok {
		for _, item := range raw {
			if f, ok := item.(float64); ok {
				m.recentLosses = append(m.recentLosses, f)
			}
		}
	}
	// Parse hop_latencies
	if raw, ok := config["hop_latencies"].([]interface{}); ok {
		for _, item := range raw {
			if f, ok := item.(float64); ok {
				m.hopLatencies = append(m.hopLatencies, f)
			}
		}
	}

	return m, nil
}

func (m *SLACorrelationMonitor) Type() string {
	return "sla_correlation"
}

func (m *SLACorrelationMonitor) ValidateConfig(config map[string]interface{}) error {
	if host, ok := config["host"].(string); ok && host == "" {
		return fmt.Errorf("host must not be empty")
	}
	if to, ok := config["timeout_seconds"].(float64); ok {
		if to < 1 || to > 120 {
			return fmt.Errorf("timeout_seconds must be between 1 and 120")
		}
	}
	if v, ok := config["sla_latency_ms"].(float64); ok && v <= 0 {
		return fmt.Errorf("sla_latency_ms must be positive")
	}
	if v, ok := config["sla_loss_percent"].(float64); ok && (v < 0 || v > 100) {
		return fmt.Errorf("sla_loss_percent must be 0-100")
	}
	if v, ok := config["sla_availability_percent"].(float64); ok && (v < 0 || v > 100) {
		return fmt.Errorf("sla_availability_percent must be 0-100")
	}
	return nil
}

func (m *SLACorrelationMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	timeout := time.Duration(m.timeoutSeconds * float64(time.Second))

	host := m.host
	if host == "" && target != nil {
		host = target.Address
	}

	meta := make(map[string]interface{})

	// --- 1. Measure overall latency via TCP connect to host:port ---
	overallLatencyMs := m.measureTCPLatency(ctx, host, m.port, timeout)
	if overallLatencyMs < 0 {
		overallLatencyMs = m.tcpConnectMs // fall back to configured value if live test fails
	}

	// --- 2. SLA Breach Prediction ---
	m.computeSLAPrediction(meta, overallLatencyMs)

	// --- 3. Root-Cause Correlation ---
	packetLoss := 0.0
	if len(m.recentLosses) > 0 {
		packetLoss = m.recentLosses[len(m.recentLosses)-1]
	}
	m.computeRootCause(meta, overallLatencyMs, packetLoss)

	// --- 4. IPv4 vs IPv6 Quality Scoring ---
	m.computeIPVersionScores(ctx, host, m.port, timeout, meta)

	// --- 5. Internet-Wide Baseline Comparison ---
	jitterMs := m.computeJitter(m.recentLatencies)
	m.computeBaselineComparison(meta, overallLatencyMs, packetLoss, jitterMs)

	// --- 6. ISP Destination Performance Scoring ---
	dest := m.destination
	if dest == "" {
		dest = host
	}
	m.computeISPDestinationScore(ctx, dest, m.port, timeout, meta)

	// --- 7. Auto Root-Cause Diagnosis Summary ---
	diagnosis := m.buildDiagnosis(meta)
	if len(diagnosis) > 200 {
		diagnosis = diagnosis[:200]
	}
	meta["diagnosis"] = diagnosis

	// Build measurement
	latencyPtr := overallLatencyMs
	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "sla_correlation",
		Success:     true,
		LatencyMs:   &latencyPtr,
		Metadata:    meta,
	}

	if target != nil {
		measurement.TargetID = target.ID
	}

	return measurement, nil
}

// measureTCPLatency connects via TCP and returns latency in ms, or -1 on failure.
func (m *SLACorrelationMonitor) measureTCPLatency(ctx context.Context, host string, port int, timeout time.Duration) float64 {
	if host == "" {
		return -1
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	start := time.Now()
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return -1
	}
	conn.Close()
	return float64(time.Since(start).Microseconds()) / 1000.0
}

// linearRegressionSlope computes the slope of y over x = [0,1,...,n-1] using least squares.
func linearRegressionSlope(data []float64) float64 {
	n := float64(len(data))
	if n < 2 {
		return 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, y := range data {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0
	}
	return (n*sumXY - sumX*sumY) / denom
}

func (m *SLACorrelationMonitor) computeSLAPrediction(meta map[string]interface{}, currentLatency float64) {
	// --- Latency SLA ---
	latSlope := linearRegressionSlope(m.recentLatencies)
	predictedLatency := currentLatency + latSlope*30 // 30 samples * 10s = 5 minutes

	var latStatus string
	timeToBreach := -1.0

	if currentLatency > m.slaLatencyMs {
		latStatus = "breach_active"
	} else if predictedLatency > m.slaLatencyMs {
		latStatus = "breach_predicted"
		// Estimate time to breach: (threshold - current) / slope (in samples), convert to minutes
		if latSlope > 0 {
			samplesToBreath := (m.slaLatencyMs - currentLatency) / latSlope
			timeToBreach = samplesToBreath * 10.0 / 60.0 // 10s per sample -> minutes
		}
	} else if currentLatency > m.slaLatencyMs*0.8 {
		latStatus = "warning"
	} else {
		latStatus = "ok"
	}

	meta["sla_latency_status"] = latStatus
	meta["predicted_latency_ms"] = math.Round(predictedLatency*100) / 100
	meta["time_to_breach_minutes"] = math.Round(timeToBreach*100) / 100

	// --- Loss SLA ---
	currentLoss := 0.0
	if len(m.recentLosses) > 0 {
		currentLoss = m.recentLosses[len(m.recentLosses)-1]
	}
	lossSlope := linearRegressionSlope(m.recentLosses)
	predictedLoss := currentLoss + lossSlope*30

	var lossStatus string
	if currentLoss > m.slaLossPercent {
		lossStatus = "breach_active"
	} else if predictedLoss > m.slaLossPercent {
		lossStatus = "breach_predicted"
	} else if currentLoss > m.slaLossPercent*0.8 {
		lossStatus = "warning"
	} else {
		lossStatus = "ok"
	}
	meta["sla_loss_status"] = lossStatus

	// --- Availability (computed from loss history as proxy) ---
	availabilityPct := 100.0
	if len(m.recentLosses) > 0 {
		var totalLoss float64
		for _, l := range m.recentLosses {
			totalLoss += l
		}
		avgLoss := totalLoss / float64(len(m.recentLosses))
		availabilityPct = 100.0 - avgLoss
	}
	meta["sla_availability_percent_current"] = math.Round(availabilityPct*1000) / 1000
}

func (m *SLACorrelationMonitor) computeRootCause(meta map[string]interface{}, overallLatency, packetLoss float64) {
	rootCause := "none"
	rootCauseHop := -1
	confidence := 0.0

	// Check DNS
	if m.dnsLatencyMs > 200 {
		rootCause = "dns_slow"
		confidence = math.Min(1.0, m.dnsLatencyMs/500.0)
	}

	// Check TLS (only if DNS is not already the primary cause)
	if m.tlsLatencyMs > 500 && confidence < 0.7 {
		rootCause = "tls_slow"
		confidence = math.Min(1.0, m.tlsLatencyMs/1000.0)
	}

	// Check hop latency jumps
	badHop := -1
	if len(m.hopLatencies) >= 2 {
		for i := 1; i < len(m.hopLatencies); i++ {
			if m.hopLatencies[i]-m.hopLatencies[i-1] > 20.0 {
				badHop = i
				break
			}
		}
	}
	if badHop >= 0 && confidence < 0.8 {
		rootCause = fmt.Sprintf("routing_issue_at_hop_%d", badHop)
		rootCauseHop = badHop
		// Confidence based on magnitude of jump
		jump := m.hopLatencies[badHop] - m.hopLatencies[badHop-1]
		confidence = math.Min(1.0, jump/100.0)
	}

	// Packet loss check
	if packetLoss > 5.0 && confidence < 0.75 {
		rootCause = "packet_loss"
		confidence = math.Min(1.0, packetLoss/20.0)
		rootCauseHop = -1
	}

	// Fall-through: all component latencies normal but overall is high
	componentOK := m.dnsLatencyMs <= 200 && m.tlsLatencyMs <= 500 && m.tcpConnectMs <= 200
	if componentOK && badHop < 0 && packetLoss <= 5.0 && overallLatency > m.slaLatencyMs && confidence < 0.5 {
		rootCause = "server_slow"
		confidence = math.Min(1.0, overallLatency/m.slaLatencyMs*0.5)
	}

	if rootCause == "none" {
		confidence = 1.0
	}

	meta["root_cause"] = rootCause
	meta["root_cause_hop"] = rootCauseHop
	meta["root_cause_confidence"] = math.Round(confidence*100) / 100
}

type ipResult struct {
	proto     string
	latencyMs float64
	err       error
}

func (m *SLACorrelationMonitor) computeIPVersionScores(ctx context.Context, host string, port int, timeout time.Duration, meta map[string]interface{}) {
	var wg sync.WaitGroup
	results := make(chan ipResult, 2)

	for _, proto := range []string{"tcp4", "tcp6"} {
		wg.Add(1)
		go func(proto string) {
			defer wg.Done()
			if host == "" {
				results <- ipResult{proto: proto, latencyMs: -1, err: fmt.Errorf("no host")}
				return
			}
			addr := fmt.Sprintf("%s:%d", host, port)
			d := net.Dialer{Timeout: timeout}
			start := time.Now()
			conn, err := d.DialContext(ctx, proto, addr)
			if err != nil {
				results <- ipResult{proto: proto, latencyMs: -1, err: err}
				return
			}
			conn.Close()
			ms := float64(time.Since(start).Microseconds()) / 1000.0
			results <- ipResult{proto: proto, latencyMs: ms}
		}(proto)
	}

	wg.Wait()
	close(results)

	var ipv4Ms, ipv6Ms float64 = -1, -1
	for r := range results {
		if r.proto == "tcp4" {
			ipv4Ms = r.latencyMs
		} else {
			ipv6Ms = r.latencyMs
		}
	}

	ipv4Score := scoreFromLatency(ipv4Ms)
	ipv6Score := scoreFromLatency(ipv6Ms)

	preferred := "ipv4"
	switch {
	case ipv6Score < 0:
		preferred = "ipv4"
	case ipv4Score < 0:
		preferred = "ipv6"
	case ipv6Score > ipv4Score+5:
		preferred = "ipv6"
	case ipv4Score > ipv6Score+5:
		preferred = "ipv4"
	default:
		preferred = "equal"
	}

	meta["ipv4_score"] = ipv4Score
	meta["ipv6_score"] = ipv6Score
	meta["preferred_protocol"] = preferred
	meta["ipv4_connect_ms"] = math.Round(ipv4Ms*100) / 100
	meta["ipv6_connect_ms"] = math.Round(ipv6Ms*100) / 100
}

// scoreFromLatency returns a 0-100 score for a TCP connect latency (ms), or -1 if unavailable.
func scoreFromLatency(ms float64) int {
	if ms < 0 {
		return -1 // unavailable
	}
	score := 100
	switch {
	case ms > 200:
		score -= 30
	case ms > 100:
		score -= 20
	case ms > 50:
		score -= 10
	}
	return score
}

func (m *SLACorrelationMonitor) computeJitter(latencies []float64) float64 {
	if len(latencies) < 2 {
		return 0
	}
	// Jitter = mean of absolute successive differences
	var sum float64
	for i := 1; i < len(latencies); i++ {
		sum += math.Abs(latencies[i] - latencies[i-1])
	}
	return sum / float64(len(latencies)-1)
}

func (m *SLACorrelationMonitor) computeBaselineComparison(meta map[string]interface{}, latencyMs, lossPercent, jitterMs float64) {
	latPct := interpolatePercentile(latencyMs, baselineLatencyP50, baselineLatencyP95)
	lossPct := interpolatePercentile(lossPercent, baselineLossP50, baselineLossP95)
	jitterPct := interpolatePercentile(jitterMs, baselineJitterP50, baselineJitterP95)

	meta["latency_percentile"] = latPct
	meta["loss_percentile"] = lossPct
	meta["jitter_percentile"] = jitterPct

	// Overall baseline rating
	avgPct := (latPct + lossPct + jitterPct) / 3.0
	var vsBaseline string
	switch {
	case avgPct < 30:
		vsBaseline = "better_than_average"
	case avgPct < 60:
		vsBaseline = "average"
	case avgPct < 80:
		vsBaseline = "worse_than_average"
	default:
		vsBaseline = "poor"
	}
	meta["vs_internet_baseline"] = vsBaseline
}

// interpolatePercentile estimates the percentile of value given p50 and p95 reference points.
// Returns 0-100.
func interpolatePercentile(value, p50, p95 float64) float64 {
	if value <= 0 {
		return 0
	}
	if value <= p50 {
		// Scale linearly from 0 to 50
		return (value / p50) * 50.0
	}
	if value <= p95 {
		// Scale linearly from 50 to 95
		return 50.0 + ((value-p50)/(p95-p50))*45.0
	}
	// Above p95: extrapolate to 100
	excess := (value - p95) / p95
	pct := 95.0 + excess*5.0
	if pct > 100 {
		pct = 100
	}
	return math.Round(pct*10) / 10
}

func (m *SLACorrelationMonitor) computeISPDestinationScore(ctx context.Context, destination string, port int, timeout time.Duration, meta map[string]interface{}) {
	if destination == "" {
		meta["isp_destination_score"] = 0
		meta["isp_name"] = m.ispName
		meta["destination_latency_ms"] = -1.0
		meta["destination_latency_stddev"] = -1.0
		return
	}

	const numSamples = 5
	latencies := make([]float64, 0, numSamples)

	addr := fmt.Sprintf("%s:%d", destination, port)
	for i := 0; i < numSamples; i++ {
		d := net.Dialer{Timeout: timeout}
		start := time.Now()
		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			// count as loss
			continue
		}
		conn.Close()
		ms := float64(time.Since(start).Microseconds()) / 1000.0
		latencies = append(latencies, ms)
	}

	lossRate := float64(numSamples-len(latencies)) / float64(numSamples)

	avgLatency := 0.0
	stddev := 0.0

	if len(latencies) > 0 {
		var sum float64
		for _, l := range latencies {
			sum += l
		}
		avgLatency = sum / float64(len(latencies))

		if len(latencies) > 1 {
			var variance float64
			for _, l := range latencies {
				diff := l - avgLatency
				variance += diff * diff
			}
			stddev = math.Sqrt(variance / float64(len(latencies)))
		}
	}

	// Score: start at 100, deduct for latency, stddev, loss
	score := 100.0
	if avgLatency > 200 {
		score -= 30
	} else if avgLatency > 100 {
		score -= 20
	} else if avgLatency > 50 {
		score -= 10
	}
	if stddev > 50 {
		score -= 20
	} else if stddev > 20 {
		score -= 10
	}
	score -= lossRate * 50 // up to -50 for 100% loss

	if score < 0 {
		score = 0
	}

	// Sort latencies to compute median for reference (unused but validates slice)
	sort.Float64s(latencies)

	meta["isp_destination_score"] = int(math.Round(score))
	meta["isp_name"] = m.ispName
	meta["destination_latency_ms"] = math.Round(avgLatency*100) / 100
	meta["destination_latency_stddev"] = math.Round(stddev*100) / 100
}

func (m *SLACorrelationMonitor) buildDiagnosis(meta map[string]interface{}) string {
	var parts []string

	// Latency SLA status
	if latStatus, ok := meta["sla_latency_status"].(string); ok && latStatus != "ok" {
		switch latStatus {
		case "breach_active":
			parts = append(parts, "SLA latency breach active")
		case "breach_predicted":
			ttb, _ := meta["time_to_breach_minutes"].(float64)
			parts = append(parts, fmt.Sprintf("SLA latency breach predicted in %.1fmin", ttb))
		case "warning":
			parts = append(parts, "latency approaching SLA threshold")
		}
	}

	// Loss SLA
	if lossStatus, ok := meta["sla_loss_status"].(string); ok && lossStatus != "ok" {
		parts = append(parts, fmt.Sprintf("packet loss SLA: %s", lossStatus))
	}

	// Root cause
	if rc, ok := meta["root_cause"].(string); ok && rc != "none" {
		conf, _ := meta["root_cause_confidence"].(float64)
		parts = append(parts, fmt.Sprintf("root cause: %s (%.0f%% confidence)", rc, conf*100))
	}

	// Baseline
	if baseline, ok := meta["vs_internet_baseline"].(string); ok && baseline != "better_than_average" && baseline != "average" {
		parts = append(parts, fmt.Sprintf("vs internet: %s", baseline))
	}

	// Preferred protocol
	if proto, ok := meta["preferred_protocol"].(string); ok && proto == "ipv6" {
		parts = append(parts, "IPv6 preferred")
	}

	if len(parts) == 0 {
		return "All metrics within normal range; no issues detected"
	}

	return strings.Join(parts, "; ")
}
