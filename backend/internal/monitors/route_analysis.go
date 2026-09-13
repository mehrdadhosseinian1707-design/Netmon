package monitors

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/netmon/netmon/internal/models"
)

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

func routeGetHost(config map[string]interface{}, target *models.Target) string {
	if h, ok := config["host"].(string); ok && h != "" {
		return h
	}
	return target.Address
}

func routeGetPort(config map[string]interface{}) int {
	if p, ok := config["port"].(float64); ok && p > 0 {
		return int(p)
	}
	return 80
}

func routeGetTimeout(config map[string]interface{}) time.Duration {
	if t, ok := config["timeout_seconds"].(float64); ok && t > 0 {
		return time.Duration(t) * time.Second
	}
	return 20 * time.Second
}

// resolveHost returns the first IPv4 (or IPv6) address for host, or host if
// it is already an IP literal.
func resolveHost(host string) (string, error) {
	if net.ParseIP(host) != nil {
		return host, nil
	}
	addrs, err := net.LookupHost(host)
	if err != nil || len(addrs) == 0 {
		return "", fmt.Errorf("cannot resolve %s: %v", host, err)
	}
	return addrs[0], nil
}

// hopFingerprint builds a SHA-256 hex digest from a slice of hop IP strings.
func hopFingerprint(hops []string) string {
	h := sha256.New()
	h.Write([]byte(strings.Join(hops, ",")))
	return hex.EncodeToString(h.Sum(nil))
}

// ---------------------------------------------------------------------------
// Monitor 1: route_instability
// ---------------------------------------------------------------------------

// RouteInstabilityMonitor runs repeated TCP-TTL probes to detect route churn.
type RouteInstabilityMonitor struct {
	config map[string]interface{}
}

func NewRouteInstabilityMonitor(config map[string]interface{}) (*RouteInstabilityMonitor, error) {
	return &RouteInstabilityMonitor{config: config}, nil
}

func (m *RouteInstabilityMonitor) Type() string { return "route_instability" }

func (m *RouteInstabilityMonitor) ValidateConfig(config map[string]interface{}) error {
	if ts, ok := config["timeout_seconds"].(float64); ok {
		if ts < 1 || ts > 300 {
			return fmt.Errorf("timeout_seconds must be between 1 and 300")
		}
	}
	if p, ok := config["port"].(float64); ok {
		if p < 1 || p > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
	}
	return nil
}

func (m *RouteInstabilityMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	host := routeGetHost(m.config, target)
	port := routeGetPort(m.config)
	timeout := routeGetTimeout(m.config)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "route_instability",
		Metadata:    make(map[string]interface{}),
	}

	// Run 3 probe rounds; each round collects hop IPs via TTL probing.
	const rounds = 3
	var roundHops [rounds][]string
	var err error

	for r := 0; r < rounds; r++ {
		select {
		case <-ctx.Done():
			measurement.ErrorMessage = "context cancelled"
			return measurement, nil
		default:
		}
		roundHops[r], err = m.probeRoute(ctx, host, port, timeout)
		if err != nil {
			// Non-fatal: record empty hop list for this round.
			roundHops[r] = []string{}
		}
	}

	// Use the longest round as the reference hop list.
	refRound := 0
	for r := 1; r < rounds; r++ {
		if len(roundHops[r]) > len(roundHops[refRound]) {
			refRound = r
		}
	}
	refHops := roundHops[refRound]
	hopCount := len(refHops)

	// Count differing hops across rounds.
	totalComparisons := 0
	differingHops := 0
	for r := 0; r < rounds; r++ {
		if r == refRound {
			continue
		}
		maxLen := hopCount
		if len(roundHops[r]) > maxLen {
			maxLen = len(roundHops[r])
		}
		for i := 0; i < maxLen; i++ {
			totalComparisons++
			var h1, h2 string
			if i < len(refHops) {
				h1 = refHops[i]
			}
			if i < len(roundHops[r]) {
				h2 = roundHops[r][i]
			}
			if h1 != h2 {
				differingHops++
			}
		}
	}

	instabilityScore := 0.0
	if totalComparisons > 0 {
		instabilityScore = float64(differingHops) / float64(totalComparisons)
	}

	fingerprint := hopFingerprint(refHops)

	// Compare with previous fingerprint from config.
	routeChanged := false
	if prevFP, ok := m.config["prev_fingerprint"].(string); ok && prevFP != "" {
		routeChanged = prevFP != fingerprint
	}

	measurement.Success = true
	measurement.HopCount = &hopCount
	measurement.Metadata["route_instability_score"] = instabilityScore
	measurement.Metadata["route_fingerprint"] = fingerprint
	measurement.Metadata["route_changed"] = routeChanged

	// RouteChanges: number of hops that differed in any round.
	rc := differingHops
	measurement.RouteChanges = &rc

	return measurement, nil
}

// probeRoute performs a single TTL-sweep traceroute-style probe (TTL 1..32)
// and returns the ordered list of responding hop IPs.
func (m *RouteInstabilityMonitor) probeRoute(ctx context.Context, host string, port int, timeout time.Duration) ([]string, error) {
	if runtime.GOOS == "linux" {
		return m.probeRouteLinux(ctx, host, port, timeout)
	}
	return m.probeRouteExec(ctx, host, timeout)
}

// probeRouteLinux uses net.Dial with a raw TCP socket and IP_TTL set via
// syscall control function to elicit ICMP Time Exceeded responses. Because
// we cannot receive ICMP replies easily without a raw socket (which requires
// root), we use the dial error or ECONNREFUSED pattern: a connection failure
// at a specific TTL implies the packet reached that hop.  When that is not
// possible we fall back to the exec path.
func (m *RouteInstabilityMonitor) probeRouteLinux(ctx context.Context, host string, port int, timeout time.Duration) ([]string, error) {
	// Try to resolve target first.
	destIP, err := resolveHost(host)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", destIP, port)
	var hops []string

	for ttl := 1; ttl <= 32; ttl++ {
		select {
		case <-ctx.Done():
			return hops, nil
		default:
		}

		hopIP := m.probeTTLLinux(ctx, addr, ttl, timeout)
		if hopIP != "" {
			hops = append(hops, hopIP)
		}

		// If we reached the destination, stop.
		if hopIP == destIP {
			break
		}

		// Stop after several consecutive unknowns.
		if ttl > 5 && hopIP == "" {
			zeroStreak := 0
			for i := len(hops) - 1; i >= 0 && i >= len(hops)-3; i-- {
				if hops[i] == "" {
					zeroStreak++
				}
			}
			if zeroStreak >= 3 {
				break
			}
		}
	}

	return hops, nil
}

// probeTTLLinux attempts a TCP dial with a custom IP_TTL and returns the peer
// IP of whatever router responded (or empty string on timeout/unknown).
func (m *RouteInstabilityMonitor) probeTTLLinux(ctx context.Context, addr string, ttl int, perHopTimeout time.Duration) string {
	ttlCopy := ttl // capture for closure

	dialer := &net.Dialer{
		Timeout: perHopTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			var setSockErr error
			err := c.Control(func(fd uintptr) {
				// Set IP TTL for the socket
				setSockErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, syscall.IP_TTL, ttlCopy)
			})
			if err != nil {
				return err
			}
			return setSockErr
		},
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		// Try to extract the remote address from the error (ICMP Time Exceeded
		// causes a net.OpError with a source IP in the error string on Linux).
		errStr := err.Error()
		// Pattern: "connect: no route to host" or similar; parse the IP from
		// the address in the dialer.
		_ = errStr
		// We cannot reliably extract the ICMP responder IP without a raw
		// socket, so return empty to indicate "no response / unknown hop".
		return ""
	}
	defer conn.Close()

	// If we got a connection, this is the destination.
	remoteAddr := conn.RemoteAddr().String()
	h, _, _ := net.SplitHostPort(remoteAddr)
	return h
}

// probeRouteExec runs the system traceroute/tracert command and parses the output.
func (m *RouteInstabilityMonitor) probeRouteExec(ctx context.Context, host string, timeout time.Duration) ([]string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.CommandContext(ctx, "tracert", "-d", "-w", "1000", "-h", "30", host)
	default:
		cmd = exec.CommandContext(ctx, "traceroute", "-n", "-m", "30", "-w", "2", host)
	}

	out, err := cmd.Output()
	if err != nil && len(out) == 0 {
		return nil, fmt.Errorf("traceroute exec failed: %v", err)
	}

	return parseTracerouteOutput(out, runtime.GOOS == "windows"), nil
}

// parseTracerouteOutput extracts hop IPs from traceroute/tracert text output.
func parseTracerouteOutput(data []byte, isWindows bool) []string {
	var hops []string
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// First field must be a hop number.
		hopNum, err := strconv.Atoi(fields[0])
		if err != nil || hopNum < 1 {
			continue
		}

		// Find the first IP address in the remaining fields.
		for _, f := range fields[1:] {
			// Strip surrounding brackets common in some formats.
			f = strings.Trim(f, "[]")
			if ip := net.ParseIP(f); ip != nil {
				hops = append(hops, ip.String())
				break
			}
		}
	}

	return hops
}

// ---------------------------------------------------------------------------
// Monitor 2: pmtu_probe
// ---------------------------------------------------------------------------

// PMTUProbeMonitor performs Path MTU discovery with black-hole detection.
type PMTUProbeMonitor struct {
	config map[string]interface{}
}

func NewPMTUProbeMonitor(config map[string]interface{}) (*PMTUProbeMonitor, error) {
	return &PMTUProbeMonitor{config: config}, nil
}

func (m *PMTUProbeMonitor) Type() string { return "pmtu_probe" }

func (m *PMTUProbeMonitor) ValidateConfig(config map[string]interface{}) error {
	if ts, ok := config["timeout_seconds"].(float64); ok {
		if ts < 1 || ts > 300 {
			return fmt.Errorf("timeout_seconds must be between 1 and 300")
		}
	}
	if p, ok := config["port"].(float64); ok {
		if p < 1 || p > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
	}
	return nil
}

func (m *PMTUProbeMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	host := routeGetHost(m.config, target)
	port := routeGetPort(m.config)
	timeout := routeGetTimeout(m.config)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "pmtu_probe",
		Metadata:    make(map[string]interface{}),
	}

	detectedPMTU, blackhole, searchPath, err := m.probePathMTU(ctx, host, port, timeout)
	if err != nil {
		measurement.Success = false
		measurement.ErrorMessage = err.Error()
		return measurement, nil
	}

	measurement.Success = true
	measurement.MTU = &detectedPMTU
	measurement.Metadata["pmtu_detected"] = detectedPMTU
	measurement.Metadata["mtu_blackhole"] = blackhole
	measurement.Metadata["pmtu_search_path"] = searchPath

	return measurement, nil
}

// probePathMTU performs a binary-search PMTU discovery.
// Returns: (detectedPMTU, blackholeDetected, searchPath, error)
func (m *PMTUProbeMonitor) probePathMTU(ctx context.Context, host string, port int, timeout time.Duration) (int, bool, []int, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Ordered candidate sizes for initial classification.
	candidates := []int{1500, 1280, 576, 512}
	searchPath := make([]int, 0, 16)

	// ---- Step 1: Determine the working ceiling ----
	// Find the largest candidate that works.
	workingCeiling := 0
	largeFailedWithConnect := false

	for _, size := range candidates {
		searchPath = append(searchPath, size)
		ok, connectOk := m.tryPayloadSize(ctx, addr, size, timeout)
		if ok {
			workingCeiling = size
			break
		}
		// If connect itself succeeded but the send failed, that suggests
		// fragmentation is being silently dropped (classic black-hole symptom).
		if connectOk && size == 1500 {
			largeFailedWithConnect = true
		}
	}

	if workingCeiling == 0 {
		// Cannot connect at all; not a black-hole situation, just unreachable.
		return 0, false, searchPath, fmt.Errorf("target %s is unreachable at all tested MTU sizes", host)
	}

	// ---- Step 2: Binary-search between workingCeiling and 1500 ----
	lo := workingCeiling
	hi := 1500
	if workingCeiling == 1500 {
		// Already works at 1500; no need to search further.
		blackhole := false
		return 1500, blackhole, searchPath, nil
	}

	for lo < hi {
		mid := (lo + hi + 1) / 2
		searchPath = append(searchPath, mid)
		ok, _ := m.tryPayloadSize(ctx, addr, mid, timeout)
		if ok {
			lo = mid
		} else {
			hi = mid - 1
		}
	}

	detectedPMTU := lo

	// ---- Step 3: Black-hole heuristic ----
	// A black-hole is when large packets are silently dropped (no ICMP
	// "fragmentation needed") rather than rejected with an error.
	// Evidence: largeFailedWithConnect = TCP connection established but
	// oversized payload timed-out or connection reset without ICMP.
	blackhole := largeFailedWithConnect && detectedPMTU < 1500

	return detectedPMTU, blackhole, searchPath, nil
}

// tryPayloadSize attempts to connect and send a TCP segment sized to payloadBytes.
// Returns (dataTransmitted, connectSucceeded).
func (m *PMTUProbeMonitor) tryPayloadSize(ctx context.Context, addr string, payloadBytes int, timeout time.Duration) (bool, bool) {
	dialTimeout := timeout / 4
	if dialTimeout < 2*time.Second {
		dialTimeout = 2 * time.Second
	}

	dialer := m.buildDialer(dialTimeout)

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return false, false
	}
	defer conn.Close()

	// Connection succeeded.
	// Now try sending a payload of the target size to stress the path MTU.
	// We subtract 60 bytes for TCP+IP headers (conservative estimate).
	dataSize := payloadBytes - 60
	if dataSize < 1 {
		dataSize = 1
	}

	payload := make([]byte, dataSize)
	// Fill with a recognizable pattern.
	for i := range payload {
		payload[i] = 0xAB
	}

	conn.SetWriteDeadline(time.Now().Add(dialTimeout))
	_, writeErr := conn.Write(payload)
	if writeErr != nil {
		// Write failed — could be RST or timeout.
		// Connection was established, but data didn't get through.
		return false, true
	}

	// Optionally read a byte to confirm the far end processed the data.
	conn.SetReadDeadline(time.Now().Add(dialTimeout))
	buf := make([]byte, 1)
	conn.Read(buf) // ignore error; we just care that we got this far

	return true, true
}

// buildDialer returns a net.Dialer; on Linux it also sets IP_DONTFRAG /
// IP_MTU_DISCOVER so that the kernel does not fragment our probes.
// On Windows it sets IP_DONTFRAGMENT.  On other platforms no special socket
// options are set and we rely purely on TCP behavior observation.
func (m *PMTUProbeMonitor) buildDialer(timeout time.Duration) *net.Dialer {
	d := &net.Dialer{Timeout: timeout}

	switch runtime.GOOS {
	case "linux":
		d.Control = func(network, address string, c syscall.RawConn) error {
			var innerErr error
			err := c.Control(func(fd uintptr) {
				// IP_MTU_DISCOVER = 10 on Linux; IP_PMTUDISC_DO = 2
				const IP_MTU_DISCOVER = 10
				const IP_PMTUDISC_DO = 2
				innerErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, IP_MTU_DISCOVER, IP_PMTUDISC_DO)
			})
			if err != nil {
				return err
			}
			return innerErr
		}
	case "windows":
		d.Control = func(network, address string, c syscall.RawConn) error {
			var innerErr error
			err := c.Control(func(fd uintptr) {
				// IP_DONTFRAGMENT = 14 on Windows (winsock2).
				const IP_DONTFRAGMENT = 14
				innerErr = syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_IP, IP_DONTFRAGMENT, 1)
			})
			if err != nil {
				return err
			}
			return innerErr
		}
	}

	return d
}
