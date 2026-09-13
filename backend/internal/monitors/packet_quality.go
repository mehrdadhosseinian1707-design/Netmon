package monitors

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/netmon/netmon/internal/models"
)

const (
	packetQualityMagic   = uint32(0xDEADBEEF)
	packetQualityType    = "packet_quality"
	reorderProbeCount    = 20
	reorderCollectMs     = 200
	burstProbeCount      = 50
	burstSpacingMs       = 5
	burstResponseTimeout = 50
	burstThreshold       = 3
	asymmetryProbeCount  = 10
	asymmetryStddevPct   = 0.20
)

// PacketQualityMonitor measures packet reordering, burst loss, and path asymmetry.
type PacketQualityMonitor struct {
	host            string
	port            int
	packetCount     int
	timeoutSeconds  float64
}

// NewPacketQualityMonitor creates a new PacketQualityMonitor from config.
func NewPacketQualityMonitor(config map[string]interface{}) (*PacketQualityMonitor, error) {
	m := &PacketQualityMonitor{
		port:           9001,
		packetCount:    50,
		timeoutSeconds: 15,
	}

	if host, ok := config["host"].(string); ok {
		m.host = host
	}
	if port, ok := config["port"].(float64); ok {
		m.port = int(port)
	} else if port, ok := config["port"].(int); ok {
		m.port = port
	}
	if pc, ok := config["packet_count"].(float64); ok {
		m.packetCount = int(pc)
	}
	if ts, ok := config["timeout_seconds"].(float64); ok {
		m.timeoutSeconds = ts
	}

	return m, nil
}

func (m *PacketQualityMonitor) Type() string {
	return packetQualityType
}

func (m *PacketQualityMonitor) ValidateConfig(config map[string]interface{}) error {
	if host, ok := config["host"].(string); !ok || host == "" {
		return fmt.Errorf("packet_quality: 'host' is required")
	}
	if port, ok := config["port"].(float64); ok {
		if port < 1 || port > 65535 {
			return fmt.Errorf("packet_quality: port must be between 1 and 65535")
		}
	}
	if pc, ok := config["packet_count"].(float64); ok {
		if pc < 1 || pc > 10000 {
			return fmt.Errorf("packet_quality: packet_count must be between 1 and 10000")
		}
	}
	if ts, ok := config["timeout_seconds"].(float64); ok {
		if ts < 1 || ts > 300 {
			return fmt.Errorf("packet_quality: timeout_seconds must be between 1 and 300")
		}
	}
	return nil
}

func (m *PacketQualityMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	host := m.host
	if host == "" {
		host = target.Address
	}
	port := m.port
	if target.Port != nil {
		port = *target.Port
	}

	deadline := time.Duration(m.timeoutSeconds * float64(time.Second))
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	metadata := make(map[string]interface{})

	// --- 1. Packet Reordering Detection ---
	reorderMeta, reorderRTTs, err := m.detectReordering(ctx, host, port)
	if err != nil {
		// non-fatal: record partial metadata
		metadata["reorder_error"] = err.Error()
	} else {
		for k, v := range reorderMeta {
			metadata[k] = v
		}
	}

	// --- 2. Burst Packet Loss Detection ---
	burstMeta, burstRTTs, err := m.detectBurstLoss(ctx, host, port)
	if err != nil {
		metadata["burst_loss_error"] = err.Error()
	} else {
		for k, v := range burstMeta {
			metadata[k] = v
		}
	}

	// --- 3. Path Asymmetry Detection ---
	asymmetryMeta, asymmetryRTTs, err := m.detectPathAsymmetry(ctx, host, port)
	if err != nil {
		metadata["asymmetry_error"] = err.Error()
	} else {
		for k, v := range asymmetryMeta {
			metadata[k] = v
		}
	}

	// Aggregate all RTT samples for top-level metrics.
	allRTTs := make([]float64, 0, len(reorderRTTs)+len(burstRTTs)+len(asymmetryRTTs))
	allRTTs = append(allRTTs, reorderRTTs...)
	allRTTs = append(allRTTs, burstRTTs...)
	allRTTs = append(allRTTs, asymmetryRTTs...)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: packetQualityType,
		Success:     len(allRTTs) > 0,
		Metadata:    metadata,
	}

	if len(allRTTs) > 0 {
		avgRTT := meanPQ(allRTTs)
		stddev := stdDev(allRTTs)

		// Overall packet loss: count all probes across the three sub-tests.
		totalSent := reorderProbeCount + burstProbeCount + asymmetryProbeCount
		totalReceived := len(allRTTs)
		lossRatio := 0.0
		if totalSent > 0 {
			lossRatio = float64(totalSent-totalReceived) / float64(totalSent)
		}

		measurement.LatencyMs = &avgRTT
		measurement.PacketLoss = &lossRatio
		measurement.JitterMs = &stddev
	}

	return measurement, nil
}

// buildProbePacket creates an 8-byte probe payload: 4-byte magic + 4-byte seq.
func buildProbePacket(seq uint32) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint32(buf[0:4], packetQualityMagic)
	binary.BigEndian.PutUint32(buf[4:8], seq)
	return buf
}

// parseProbePacket validates magic and returns the sequence number.
func parseProbePacket(buf []byte) (uint32, bool) {
	if len(buf) < 8 {
		return 0, false
	}
	magic := binary.BigEndian.Uint32(buf[0:4])
	if magic != packetQualityMagic {
		return 0, false
	}
	seq := binary.BigEndian.Uint32(buf[4:8])
	return seq, true
}

// detectReordering sends reorderProbeCount UDP probes and checks arrival order.
// Returns metadata map and slice of per-probe RTTs (ms) for successfully received probes.
func (m *PacketQualityMonitor) detectReordering(ctx context.Context, host string, port int) (map[string]interface{}, []float64, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, nil, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	n := reorderProbeCount

	// sendTimes[seq] = send timestamp
	sendTimes := make([]time.Time, n)
	recvBuf := make([]byte, 64)

	// Send all probes first, then collect.
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}
		sendTimes[i] = time.Now()
		pkt := buildProbePacket(uint32(i))
		conn.SetWriteDeadline(time.Now().Add(50 * time.Millisecond))
		conn.Write(pkt) //nolint:errcheck — best-effort send
	}

	// Collect responses for up to reorderCollectMs.
	deadline := time.Now().Add(reorderCollectMs * time.Millisecond)
	conn.SetReadDeadline(deadline)

	type recv struct {
		seq   uint32
		rttMs float64
	}
	var received []recv

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			goto doneReorder
		default:
		}
		conn.SetReadDeadline(time.Now().Add(10 * time.Millisecond))
		nr, err2 := conn.Read(recvBuf)
		if err2 != nil {
			// timeout or error — keep looping until outer deadline.
			continue
		}
		now := time.Now()
		seq, ok := parseProbePacket(recvBuf[:nr])
		if !ok || int(seq) >= n {
			continue
		}
		rttMs := float64(now.Sub(sendTimes[seq]).Microseconds()) / 1000.0
		received = append(received, recv{seq: seq, rttMs: rttMs})
	}

doneReorder:
	totalReceived := len(received)
	reorderCount := 0
	expected := uint32(0)
	for _, r := range received {
		if r.seq != expected {
			reorderCount++
		}
		expected = r.seq + 1
	}

	reorderRatio := 0.0
	if totalReceived > 0 {
		reorderRatio = float64(reorderCount) / float64(totalReceived)
	}

	rtts := make([]float64, 0, totalReceived)
	for _, r := range received {
		rtts = append(rtts, r.rttMs)
	}

	meta := map[string]interface{}{
		"reorder_count":    reorderCount,
		"reorder_ratio":    reorderRatio,
		"reorder_detected": reorderRatio > 0.02,
	}
	return meta, rtts, nil
}

// detectBurstLoss sends burstProbeCount UDP probes spaced 5ms apart.
// Returns metadata map and slice of RTTs for successfully received probes.
func (m *PacketQualityMonitor) detectBurstLoss(ctx context.Context, host string, port int) (map[string]interface{}, []float64, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	raddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return nil, nil, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	n := burstProbeCount
	sendTimes := make([]time.Time, n)
	// received[i] = true if we got a reply for probe i
	received := make([]bool, n)
	rtts := make([]float64, 0, n)

	var mu sync.Mutex
	var wg sync.WaitGroup

	// Receiver goroutine.
	wg.Add(1)
	go func() {
		defer wg.Done()
		recvBuf := make([]byte, 64)
		// Read until context done or we've been waiting long enough.
		overallDeadline := time.Duration(n)*burstSpacingMs*time.Millisecond + burstResponseTimeout*time.Millisecond + 200*time.Millisecond
		conn.SetReadDeadline(time.Now().Add(overallDeadline))
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			conn.SetReadDeadline(time.Now().Add(20 * time.Millisecond))
			nr, err2 := conn.Read(recvBuf)
			if err2 != nil {
				// Check if we're past the reasonable wait window.
				if time.Now().After(time.Now().Add(overallDeadline)) {
					return
				}
				continue
			}
			now := time.Now()
			seq, ok := parseProbePacket(recvBuf[:nr])
			if !ok || int(seq) >= n {
				continue
			}

			mu.Lock()
			if !received[seq] {
				received[seq] = true
				rttMs := float64(now.Sub(sendTimes[seq]).Microseconds()) / 1000.0
				if rttMs >= 0 {
					rtts = append(rtts, rttMs)
				}
			}
			mu.Unlock()
		}
	}()

	// Send probes with 5ms spacing.
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			goto doneSend
		default:
		}
		mu.Lock()
		sendTimes[i] = time.Now()
		mu.Unlock()
		pkt := buildProbePacket(uint32(i))
		conn.SetWriteDeadline(time.Now().Add(20 * time.Millisecond))
		conn.Write(pkt) //nolint:errcheck
		time.Sleep(burstSpacingMs * time.Millisecond)
	}

doneSend:
	// Wait for remaining responses.
	wait := time.NewTimer(burstResponseTimeout * time.Millisecond)
	select {
	case <-wait.C:
	case <-ctx.Done():
	}
	conn.Close() // unblocks receiver
	wg.Wait()

	// Identify lost probes and burst runs.
	// A probe is "lost" if no response received within the overall window.
	// We use the received[] array (set by receiver goroutine).
	mu.Lock()
	defer mu.Unlock()

	burstLossCount := 0
	type lossRun [2]int
	var burstRuns []lossRun

	inRun := false
	runStart := 0
	runLen := 0

	for i := 0; i < n; i++ {
		if !received[i] {
			burstLossCount++
			if !inRun {
				inRun = true
				runStart = i
				runLen = 1
			} else {
				runLen++
			}
		} else {
			if inRun {
				if runLen >= burstThreshold {
					burstRuns = append(burstRuns, lossRun{runStart, runStart + runLen - 1})
				}
				inRun = false
				runLen = 0
			}
		}
	}
	if inRun && runLen >= burstThreshold {
		burstRuns = append(burstRuns, lossRun{runStart, runStart + runLen - 1})
	}

	// Convert burst runs to [][]int for JSON-friendliness.
	runsJSON := make([][]int, 0, len(burstRuns))
	for _, r := range burstRuns {
		runsJSON = append(runsJSON, []int{r[0], r[1]})
	}

	meta := map[string]interface{}{
		"burst_loss_count":    burstLossCount,
		"burst_loss_runs":     runsJSON,
		"burst_loss_detected": len(burstRuns) > 0,
	}
	return meta, rtts, nil
}

// detectPathAsymmetry sends asymmetryProbeCount probes (ICMP preferred, TCP fallback)
// and computes RTT variance to detect path asymmetry.
func (m *PacketQualityMonitor) detectPathAsymmetry(ctx context.Context, host string, port int) (map[string]interface{}, []float64, error) {
	n := asymmetryProbeCount
	rtts := make([]float64, 0, n)

	// Try ICMP first; fall back to TCP connect timing.
	icmpRTTs, icmpErr := m.probeICMP(ctx, host, n)
	if icmpErr == nil && len(icmpRTTs) > 0 {
		rtts = icmpRTTs
	} else {
		tcpRTTs, tcpErr := m.probeTCP(ctx, host, port, n)
		if tcpErr != nil {
			return nil, nil, fmt.Errorf("asymmetry probe failed (icmp: %v, tcp: %v)", icmpErr, tcpErr)
		}
		rtts = tcpRTTs
	}

	if len(rtts) == 0 {
		return nil, nil, fmt.Errorf("no RTT samples for asymmetry detection")
	}

	avg := meanPQ(rtts)
	sd := stdDev(rtts)
	asymmetryRatio := 0.0
	if avg > 0 {
		asymmetryRatio = sd / avg
	}
	asymmetryDetected := asymmetryRatio > asymmetryStddevPct

	meta := map[string]interface{}{
		"asymmetry_ratio":    asymmetryRatio,
		"asymmetry_detected": asymmetryDetected,
		"rtt_variance_ms":    sd * sd, // variance = stddev^2 (in ms^2)
	}
	return meta, rtts, nil
}

// probeICMP sends n ICMP echo requests using raw ip4:icmp socket and returns RTTs.
func (m *PacketQualityMonitor) probeICMP(ctx context.Context, host string, n int) ([]float64, error) {
	conn, err := net.DialTimeout("ip4:icmp", host, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rtts := make([]float64, 0, n)
	icmpID := uint16(time.Now().UnixNano() & 0xFFFF)

	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return rtts, ctx.Err()
		default:
		}

		seq := uint16(i)
		pkt := buildICMPEcho(icmpID, seq)

		conn.SetDeadline(time.Now().Add(500 * time.Millisecond))
		start := time.Now()
		_, err := conn.Write(pkt)
		if err != nil {
			continue
		}

		reply := make([]byte, 100)
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		for {
			nr, err2 := conn.Read(reply)
			if err2 != nil {
				break
			}
			// IP header is 20 bytes; ICMP starts at offset 20.
			if nr < 28 {
				continue
			}
			icmpData := reply[20:nr]
			if len(icmpData) < 8 {
				continue
			}
			// Type=0 (echo reply), Code=0
			if icmpData[0] != 0 {
				continue
			}
			rid := binary.BigEndian.Uint16(icmpData[4:6])
			rseq := binary.BigEndian.Uint16(icmpData[6:8])
			if rid == icmpID && rseq == seq {
				rttMs := float64(time.Since(start).Microseconds()) / 1000.0
				rtts = append(rtts, rttMs)
				break
			}
		}

		time.Sleep(10 * time.Millisecond)
	}

	if len(rtts) == 0 {
		return nil, fmt.Errorf("no ICMP replies received")
	}
	return rtts, nil
}

// buildICMPEcho constructs a minimal ICMP echo request packet.
func buildICMPEcho(id, seq uint16) []byte {
	pkt := make([]byte, 8)
	pkt[0] = 8 // type: echo request
	pkt[1] = 0 // code
	// checksum placeholder at [2:4]
	binary.BigEndian.PutUint16(pkt[4:6], id)
	binary.BigEndian.PutUint16(pkt[6:8], seq)
	cs := icmpChecksum(pkt)
	binary.BigEndian.PutUint16(pkt[2:4], cs)
	return pkt
}

func icmpChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i+1 < len(data); i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	if len(data)%2 != 0 {
		sum += uint32(data[len(data)-1]) << 8
	}
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}

// probeTCP measures TCP connect latency to host:port for n attempts.
func (m *PacketQualityMonitor) probeTCP(ctx context.Context, host string, port int, n int) ([]float64, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	rtts := make([]float64, 0, n)

	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			return rtts, ctx.Err()
		default:
		}

		start := time.Now()
		conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		if err == nil {
			conn.Close()
			rtts = append(rtts, elapsed)
		}

		time.Sleep(20 * time.Millisecond)
	}

	if len(rtts) == 0 {
		return nil, fmt.Errorf("no TCP connections succeeded to %s", addr)
	}
	return rtts, nil
}

// meanPQ returns the arithmetic mean of a slice of float64.
func meanPQ(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

// stdDev returns the population standard deviation.
func stdDev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	avg := meanPQ(vals)
	sumSq := 0.0
	for _, v := range vals {
		d := v - avg
		sumSq += d * d
	}
	return math.Sqrt(sumSq / float64(len(vals)))
}
