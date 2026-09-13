package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/netmon/netmon/internal/models"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

type TracerouteMonitor struct {
	config TracerouteConfig
}

type TracerouteConfig struct {
	MaxHops       int
	Timeout       time.Duration
	PacketSize    int
	Queries       int // Number of queries per hop
	DetectMTU     bool
	TrackRoutes   bool
	LookupASN     bool
}

type TracerouteResult struct {
	Success       bool
	TotalHops     int
	Hops          []HopResult
	MTU           int
	RouteChanges  int
	ASPath        string
	ErrorMessage  string
}

type HopResult struct {
	HopNumber int
	IPAddress string
	Hostname  string
	RTT       float64
	Timeout   bool
	ASN       string
}

func NewTracerouteMonitor(config map[string]interface{}) (*TracerouteMonitor, error) {
	cfg := TracerouteConfig{
		MaxHops:     30,
		Timeout:     5 * time.Second,
		PacketSize:  52,
		Queries:     3,
		DetectMTU:   true,
		TrackRoutes: true,
		LookupASN:   false,
	}

	if maxHops, ok := config["max_hops"].(float64); ok {
		cfg.MaxHops = int(maxHops)
	}
	if timeout, ok := config["timeout"].(float64); ok {
		cfg.Timeout = time.Duration(timeout) * time.Second
	}
	if packetSize, ok := config["packet_size"].(float64); ok {
		cfg.PacketSize = int(packetSize)
	}
	if queries, ok := config["queries"].(float64); ok {
		cfg.Queries = int(queries)
	}
	if detectMTU, ok := config["detect_mtu"].(bool); ok {
		cfg.DetectMTU = detectMTU
	}
	if trackRoutes, ok := config["track_routes"].(bool); ok {
		cfg.TrackRoutes = trackRoutes
	}
	if lookupASN, ok := config["lookup_asn"].(bool); ok {
		cfg.LookupASN = lookupASN
	}

	return &TracerouteMonitor{config: cfg}, nil
}

func (m *TracerouteMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	result := m.trace(ctx, target.Address)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: models.MonitorTypeTraceroute,
		Success:     result.Success,
		Metadata:    make(map[string]interface{}),
	}

	if result.Success {
		// Calculate average RTT from all successful hops
		totalRTT := 0.0
		hopCount := 0
		for _, hop := range result.Hops {
			if !hop.Timeout && hop.RTT > 0 {
				totalRTT += hop.RTT
				hopCount++
			}
		}
		if hopCount > 0 {
			avgRTT := totalRTT / float64(hopCount)
			measurement.LatencyMs = &avgRTT
		}

		// Set hop count
		measurement.HopCount = &result.TotalHops

		// Set MTU if detected
		if result.MTU > 0 {
			measurement.MTU = &result.MTU
		}

		// Set route changes if tracked
		if result.RouteChanges >= 0 {
			measurement.RouteChanges = &result.RouteChanges
		}

		// Set AS path if available
		if result.ASPath != "" {
			measurement.ASPath = &result.ASPath
		}

		measurement.Metadata["total_hops"] = result.TotalHops
		measurement.Metadata["hops"] = result.Hops
	} else {
		measurement.ErrorMessage = result.ErrorMessage
	}

	return measurement, nil
}

func (m *TracerouteMonitor) trace(ctx context.Context, destination string) TracerouteResult {
	result := TracerouteResult{
		Success:      false,
		Hops:         make([]HopResult, 0),
		RouteChanges: -1,
	}

	// Resolve destination
	ips, err := net.LookupIP(destination)
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to resolve destination: %v", err)
		return result
	}
	if len(ips) == 0 {
		result.ErrorMessage = "no IP addresses found"
		return result
	}

	destIP := ips[0].String()

	// Create ICMP connection
	conn, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		result.ErrorMessage = fmt.Sprintf("failed to create ICMP connection: %v", err)
		return result
	}
	defer conn.Close()

	// Perform traceroute
	for ttl := 1; ttl <= m.config.MaxHops; ttl++ {
		select {
		case <-ctx.Done():
			result.ErrorMessage = "context cancelled"
			return result
		default:
		}

		hop := m.probeHop(conn, destIP, ttl)

		// Lookup ASN if enabled
		if m.config.LookupASN && hop.IPAddress != "" {
			hop.ASN = m.lookupASN(hop.IPAddress)
		}

		result.Hops = append(result.Hops, hop)

		// Check if we reached the destination
		if hop.IPAddress == destIP {
			result.Success = true
			result.TotalHops = ttl
			break
		}

		// Stop if we got consecutive timeouts (likely unreachable)
		if ttl > 5 && hop.Timeout {
			consecutiveTimeouts := 0
			for i := len(result.Hops) - 1; i >= 0 && i >= len(result.Hops)-3; i-- {
				if result.Hops[i].Timeout {
					consecutiveTimeouts++
				}
			}
			if consecutiveTimeouts >= 3 {
				result.ErrorMessage = "destination unreachable (consecutive timeouts)"
				return result
			}
		}
	}

	if !result.Success {
		result.ErrorMessage = "max hops reached without reaching destination"
		result.TotalHops = m.config.MaxHops
	}

	// Detect MTU if enabled
	if m.config.DetectMTU && result.Success {
		result.MTU = m.detectMTU(ctx, destIP)
	}

	// Build AS path
	if m.config.LookupASN {
		result.ASPath = m.buildASPath(result.Hops)
	}

	// Calculate route changes (placeholder - would need historical data)
	if m.config.TrackRoutes {
		result.RouteChanges = 0 // Would compare with previous routes
	}

	return result
}

func (m *TracerouteMonitor) probeHop(conn *icmp.PacketConn, destIP string, ttl int) HopResult {
	hop := HopResult{
		HopNumber: ttl,
		Timeout:   true,
	}

	// Set TTL
	if err := conn.IPv4PacketConn().SetTTL(ttl); err != nil {
		return hop
	}

	// Create ICMP echo request
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Code: 0,
		Body: &icmp.Echo{
			ID:   1234,
			Seq:  ttl,
			Data: make([]byte, m.config.PacketSize),
		},
	}

	msgBytes, err := msg.Marshal(nil)
	if err != nil {
		return hop
	}

	// Send packet
	start := time.Now()
	dest := &net.IPAddr{IP: net.ParseIP(destIP)}

	if _, err := conn.WriteTo(msgBytes, dest); err != nil {
		return hop
	}

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(m.config.Timeout))

	// Read response
	reply := make([]byte, 1500)
	n, peer, err := conn.ReadFrom(reply)
	if err != nil {
		return hop
	}

	rtt := time.Since(start)
	hop.RTT = float64(rtt.Microseconds()) / 1000.0
	hop.Timeout = false

	// Extract IP address
	if peerIP, ok := peer.(*net.IPAddr); ok {
		hop.IPAddress = peerIP.IP.String()

		// Reverse DNS lookup
		names, err := net.LookupAddr(hop.IPAddress)
		if err == nil && len(names) > 0 {
			hop.Hostname = names[0]
		}
	}

	// Parse ICMP response
	rm, err := icmp.ParseMessage(1, reply[:n])
	if err == nil {
		switch rm.Type {
		case ipv4.ICMPTypeTimeExceeded:
			// Expected response from intermediate hop
		case ipv4.ICMPTypeEchoReply:
			// Reached destination
		}
	}

	return hop
}

func (m *TracerouteMonitor) detectMTU(ctx context.Context, destIP string) int {
	// MTU detection using binary search with DF (Don't Fragment) flag
	// Start with common MTU sizes
	minMTU := 576   // IPv4 minimum
	maxMTU := 1500  // Ethernet standard
	detectedMTU := minMTU

	for minMTU <= maxMTU {
		testMTU := (minMTU + maxMTU) / 2

		// Try to send packet with this size
		if m.testMTUSize(ctx, destIP, testMTU) {
			detectedMTU = testMTU
			minMTU = testMTU + 1
		} else {
			maxMTU = testMTU - 1
		}
	}

	return detectedMTU
}

func (m *TracerouteMonitor) testMTUSize(ctx context.Context, destIP string, size int) bool {
	// Simplified MTU test - would need raw sockets for DF flag
	// This is a placeholder implementation
	conn, err := net.DialTimeout("tcp", destIP+":80", 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (m *TracerouteMonitor) lookupASN(ipAddr string) string {
	// ASN lookup - would typically use a service like Team Cymru or RIPEstat
	// This is a placeholder that would need external API integration
	// Example: whois -h whois.cymru.com " -v 8.8.8.8"
	return "" // Placeholder - requires external service
}

func (m *TracerouteMonitor) buildASPath(hops []HopResult) string {
	// Build AS path from hop ASNs
	var asPath string
	prevASN := ""

	for _, hop := range hops {
		if hop.ASN != "" && hop.ASN != prevASN {
			if asPath != "" {
				asPath += " "
			}
			asPath += hop.ASN
			prevASN = hop.ASN
		}
	}

	return asPath
}

func (m *TracerouteMonitor) Type() string {
	return models.MonitorTypeTraceroute
}

func (m *TracerouteMonitor) ValidateConfig(config map[string]interface{}) error {
	if maxHops, ok := config["max_hops"].(float64); ok {
		if maxHops < 1 || maxHops > 64 {
			return fmt.Errorf("max_hops must be between 1 and 64")
		}
	}
	if queries, ok := config["queries"].(float64); ok {
		if queries < 1 || queries > 10 {
			return fmt.Errorf("queries must be between 1 and 10")
		}
	}
	return nil
}
