package monitors

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/netmon/netmon/internal/models"
)

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

func getStringConfig(config map[string]interface{}, key, def string) string {
	if v, ok := config[key].(string); ok && v != "" {
		return v
	}
	return def
}

func getFloat64Config(config map[string]interface{}, key string, def float64) float64 {
	if v, ok := config[key].(float64); ok {
		return v
	}
	return def
}

// buildDNSQuery builds a minimal DNS wire-format query.
// qtype: 1=A, 28=AAAA, etc.
// withDNSSEC: if true, appends an EDNS0 OPT RR with the DO bit set.
func buildDNSQuery(domain string, qtype uint16, id uint16, withDNSSEC bool) []byte {
	buf := new(bytes.Buffer)

	// Header
	binary.Write(buf, binary.BigEndian, id)      // ID
	binary.Write(buf, binary.BigEndian, uint16(0x0100)) // Flags: QR=0 OPCODE=0 RD=1
	binary.Write(buf, binary.BigEndian, uint16(1))      // QDCOUNT
	binary.Write(buf, binary.BigEndian, uint16(0))      // ANCOUNT
	binary.Write(buf, binary.BigEndian, uint16(0))      // NSCOUNT
	if withDNSSEC {
		binary.Write(buf, binary.BigEndian, uint16(1)) // ARCOUNT (OPT RR)
	} else {
		binary.Write(buf, binary.BigEndian, uint16(0))
	}

	// Question section: encode domain name
	labels := strings.Split(strings.TrimSuffix(domain, "."), ".")
	for _, label := range labels {
		if label == "" {
			continue
		}
		buf.WriteByte(byte(len(label)))
		buf.WriteString(label)
	}
	buf.WriteByte(0) // root label
	binary.Write(buf, binary.BigEndian, qtype)    // QTYPE
	binary.Write(buf, binary.BigEndian, uint16(1)) // QCLASS IN

	// EDNS0 OPT RR with DO bit
	if withDNSSEC {
		buf.WriteByte(0)                                 // NAME: root
		binary.Write(buf, binary.BigEndian, uint16(41)) // TYPE: OPT
		binary.Write(buf, binary.BigEndian, uint16(4096)) // CLASS: requestor's UDP payload size
		binary.Write(buf, binary.BigEndian, uint32(0x00008000)) // TTL: extended RCODE + DO bit set
		binary.Write(buf, binary.BigEndian, uint16(0)) // RDLENGTH: 0
	}

	return buf.Bytes()
}

// sendDNSQueryUDP sends a raw DNS query over UDP and returns the raw response.
func sendDNSQueryUDP(ctx context.Context, server string, query []byte) ([]byte, error) {
	conn, err := net.Dial("udp", server)
	if err != nil {
		return nil, fmt.Errorf("udp dial: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	if _, err := conn.Write(query); err != nil {
		return nil, fmt.Errorf("udp write: %w", err)
	}

	resp := make([]byte, 4096)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("udp read: %w", err)
	}
	return resp[:n], nil
}

// sendDNSQueryTCP sends a raw DNS query over TCP (prefixes 2-byte length).
func sendDNSQueryTCP(ctx context.Context, server string, query []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", server)
	if err != nil {
		return nil, fmt.Errorf("tcp dial: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// TCP DNS: 2-byte length prefix
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(query)))
	if _, err := conn.Write(append(length, query...)); err != nil {
		return nil, fmt.Errorf("tcp write: %w", err)
	}

	// Read response length
	var respLen uint16
	if err := binary.Read(conn, binary.BigEndian, &respLen); err != nil {
		return nil, fmt.Errorf("tcp read length: %w", err)
	}

	resp := make([]byte, respLen)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return nil, fmt.Errorf("tcp read body: %w", err)
	}
	return resp, nil
}

// parseDNSFlags extracts flags and RCODE from a DNS response header.
// Returns: flags uint16, rcode uint8, truncated bool, authenticated bool.
func parseDNSFlags(resp []byte) (flags uint16, rcode uint8, truncated bool, authenticated bool) {
	if len(resp) < 4 {
		return 0, 0, false, false
	}
	flags = binary.BigEndian.Uint16(resp[2:4])
	rcode = uint8(flags & 0x000F)
	truncated = (flags & 0x0200) != 0
	authenticated = (flags & 0x0020) != 0
	return
}

// parseAnswerCount returns ANCOUNT from a DNS header.
func parseAnswerCount(resp []byte) uint16 {
	if len(resp) < 8 {
		return 0
	}
	return binary.BigEndian.Uint16(resp[6:8])
}

// parseAdditionalCount returns ARCOUNT.
func parseAdditionalCount(resp []byte) uint16 {
	if len(resp) < 12 {
		return 0
	}
	return binary.BigEndian.Uint16(resp[10:12])
}

// countRRSIG scans through the answer section and additional section of a DNS
// response counting RRSIG records (type 46 / 0x002e).
// This is a best-effort scan; it skips the question section and walks RRs.
func countRRSIG(resp []byte) int {
	if len(resp) < 12 {
		return 0
	}
	ancount := int(binary.BigEndian.Uint16(resp[6:8]))
	nscount := int(binary.BigEndian.Uint16(resp[8:10]))
	arcount := int(binary.BigEndian.Uint16(resp[10:12]))
	total := ancount + nscount + arcount

	offset := 12
	// Skip question section
	qdcount := int(binary.BigEndian.Uint16(resp[4:6]))
	for i := 0; i < qdcount; i++ {
		// Skip name
		o, ok := skipName(resp, offset)
		if !ok {
			return 0
		}
		offset = o + 4 // QTYPE + QCLASS
	}

	rrsigCount := 0
	for i := 0; i < total; i++ {
		o, ok := skipName(resp, offset)
		if !ok {
			break
		}
		if o+10 > len(resp) {
			break
		}
		rrtype := binary.BigEndian.Uint16(resp[o : o+2])
		rdlength := int(binary.BigEndian.Uint16(resp[o+8 : o+10]))
		if rrtype == 46 { // RRSIG
			rrsigCount++
		}
		offset = o + 10 + rdlength
	}
	return rrsigCount
}

// skipName advances past a DNS wire-format name (handles pointers).
// Returns new offset (pointing past the name) and ok.
func skipName(msg []byte, offset int) (int, bool) {
	for {
		if offset >= len(msg) {
			return 0, false
		}
		length := int(msg[offset])
		if length == 0 {
			return offset + 1, true
		}
		if length&0xC0 == 0xC0 {
			// Pointer: 2 bytes
			return offset + 2, true
		}
		offset += 1 + length
	}
}

// ---------------------------------------------------------------------------
// Monitor 1: dns_classify
// ---------------------------------------------------------------------------

// DNSClassifyMonitor performs DNS queries and classifies failure modes.
type DNSClassifyMonitor struct{}

func (m *DNSClassifyMonitor) Type() string { return "dns_classify" }

func (m *DNSClassifyMonitor) ValidateConfig(config map[string]interface{}) error {
	if ns, ok := config["nameserver"].(string); ok && ns != "" {
		host, _, err := net.SplitHostPort(ns)
		if err != nil {
			// might be just IP without port – acceptable
			if net.ParseIP(ns) == nil {
				return fmt.Errorf("invalid nameserver %q: must be IP or IP:port", ns)
			}
			_ = host
		}
	}
	if to, ok := config["timeout_seconds"].(float64); ok && to <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	return nil
}

func (m *DNSClassifyMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	domain := getStringConfig(config, "domain", target.Address)
	if domain == "" {
		domain = "google.com"
	}
	nameserver := getStringConfig(config, "nameserver", "8.8.8.8:53")
	// Ensure nameserver has a port
	if _, _, err := net.SplitHostPort(nameserver); err != nil {
		nameserver = nameserver + ":53"
	}
	timeoutSec := getFloat64Config(config, "timeout_seconds", 10)
	timeout := time.Duration(timeoutSec * float64(time.Second))

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "dns_classify",
		Metadata:    make(map[string]interface{}),
	}
	measurement.Metadata["nameserver_used"] = nameserver

	// Build and send UDP query
	query := buildDNSQuery(domain, 1, 0xABCD, false)
	start := time.Now()
	resp, err := sendDNSQueryUDP(queryCtx, nameserver, query)
	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	measurement.DNSLatencyMs = &elapsed

	if err != nil {
		class := classifyDNSError(err)
		measurement.Success = false
		measurement.ErrorMessage = err.Error()
		measurement.Metadata["failure_class"] = class
		measurement.Metadata["retry_tcp"] = false
		measurement.Metadata["cname_depth"] = 0
		return measurement, nil
	}

	// Check for truncation
	_, rcode, truncated, _ := parseDNSFlags(resp)
	if truncated {
		measurement.Metadata["retry_tcp"] = true
		// Retry over TCP
		tcpCtx, tcpCancel := context.WithTimeout(ctx, timeout)
		defer tcpCancel()
		tcpStart := time.Now()
		resp2, tcpErr := sendDNSQueryTCP(tcpCtx, nameserver, query)
		tcpElapsed := float64(time.Since(tcpStart).Microseconds()) / 1000.0
		measurement.DNSLatencyMs = &tcpElapsed
		if tcpErr != nil {
			measurement.Success = false
			measurement.ErrorMessage = tcpErr.Error()
			measurement.Metadata["failure_class"] = "TRUNCATED"
			measurement.Metadata["cname_depth"] = 0
			return measurement, nil
		}
		resp = resp2
		_, rcode, _, _ = parseDNSFlags(resp)
	} else {
		measurement.Metadata["retry_tcp"] = false
	}

	// Classify by RCODE
	switch rcode {
	case 0: // NOERROR
		ancount := parseAnswerCount(resp)
		if ancount == 0 {
			measurement.Success = false
			measurement.Metadata["failure_class"] = "EMPTY_RESPONSE"
			measurement.ErrorMessage = "NOERROR but zero records returned"
			measurement.Metadata["cname_depth"] = 0
			return measurement, nil
		}
		// Check CNAME chain depth
		cnameDepth := estimateCNAMEDepth(resp)
		measurement.Metadata["cname_depth"] = cnameDepth
		if cnameDepth > 8 {
			measurement.Success = false
			measurement.Metadata["failure_class"] = "LOOP_DETECTED"
			measurement.ErrorMessage = fmt.Sprintf("CNAME chain depth %d exceeds limit of 8", cnameDepth)
			return measurement, nil
		}
		measurement.Success = true
		measurement.Metadata["failure_class"] = ""
	case 1:
		measurement.Success = false
		measurement.Metadata["failure_class"] = "NETWORK_ERROR"
		measurement.ErrorMessage = "FORMERR: server could not interpret query"
		measurement.Metadata["cname_depth"] = 0
	case 2:
		measurement.Success = false
		measurement.Metadata["failure_class"] = "SERVFAIL"
		measurement.ErrorMessage = "SERVFAIL: server returned SERVFAIL"
		measurement.Metadata["cname_depth"] = 0
	case 3:
		measurement.Success = false
		measurement.Metadata["failure_class"] = "NXDOMAIN"
		measurement.ErrorMessage = "NXDOMAIN: domain does not exist"
		measurement.Metadata["cname_depth"] = 0
	case 5:
		measurement.Success = false
		measurement.Metadata["failure_class"] = "REFUSED"
		measurement.ErrorMessage = "REFUSED: server refused query"
		measurement.Metadata["cname_depth"] = 0
	default:
		measurement.Success = false
		measurement.Metadata["failure_class"] = "NETWORK_ERROR"
		measurement.ErrorMessage = fmt.Sprintf("unexpected RCODE %d", rcode)
		measurement.Metadata["cname_depth"] = 0
	}

	return measurement, nil
}

// classifyDNSError maps a low-level error to a DNS failure class string.
func classifyDNSError(err error) string {
	if err == nil {
		return ""
	}
	s := strings.ToLower(err.Error())
	if strings.Contains(s, "context deadline exceeded") || strings.Contains(s, "timeout") || strings.Contains(s, "i/o timeout") {
		return "TIMEOUT"
	}
	if strings.Contains(s, "no such host") || strings.Contains(s, "nxdomain") {
		return "NXDOMAIN"
	}
	if strings.Contains(s, "servfail") || strings.Contains(s, "server misbehaving") {
		return "SERVFAIL"
	}
	if strings.Contains(s, "refused") {
		return "REFUSED"
	}
	if strings.Contains(s, "connection refused") || strings.Contains(s, "no route") || strings.Contains(s, "network unreachable") {
		return "NO_SERVERS"
	}
	return "NETWORK_ERROR"
}

// estimateCNAMEDepth counts CNAME records (type 5) in the answer section.
func estimateCNAMEDepth(resp []byte) int {
	if len(resp) < 12 {
		return 0
	}
	ancount := int(binary.BigEndian.Uint16(resp[6:8]))
	offset := 12

	// Skip question section
	qdcount := int(binary.BigEndian.Uint16(resp[4:6]))
	for i := 0; i < qdcount; i++ {
		o, ok := skipName(resp, offset)
		if !ok {
			return 0
		}
		offset = o + 4
	}

	cnameCount := 0
	for i := 0; i < ancount; i++ {
		o, ok := skipName(resp, offset)
		if !ok {
			break
		}
		if o+10 > len(resp) {
			break
		}
		rrtype := binary.BigEndian.Uint16(resp[o : o+2])
		rdlength := int(binary.BigEndian.Uint16(resp[o+8 : o+10]))
		if rrtype == 5 { // CNAME
			cnameCount++
		}
		offset = o + 10 + rdlength
	}
	return cnameCount
}

// ---------------------------------------------------------------------------
// Monitor 2: dnssec
// ---------------------------------------------------------------------------

// DNSSECMonitor checks DNSSEC status via raw DNS queries with the DO bit set.
type DNSSECMonitor struct{}

func (m *DNSSECMonitor) Type() string { return "dnssec" }

func (m *DNSSECMonitor) ValidateConfig(config map[string]interface{}) error {
	if to, ok := config["timeout_seconds"].(float64); ok && to <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	return nil
}

func (m *DNSSECMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	domain := getStringConfig(config, "domain", target.Address)
	if domain == "" {
		domain = "google.com"
	}
	nameserver := getStringConfig(config, "nameserver", "8.8.8.8:53")
	if _, _, err := net.SplitHostPort(nameserver); err != nil {
		nameserver = nameserver + ":53"
	}
	timeoutSec := getFloat64Config(config, "timeout_seconds", 10)
	timeout := time.Duration(timeoutSec * float64(time.Second))

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "dnssec",
		Metadata:    make(map[string]interface{}),
	}

	// Build query with DO bit
	query := buildDNSQuery(domain, 1, 0x1234, true)
	start := time.Now()
	resp, err := sendDNSQueryUDP(queryCtx, nameserver, query)
	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	measurement.LatencyMs = &elapsed

	if err != nil {
		measurement.Success = false
		measurement.ErrorMessage = err.Error()
		measurement.Metadata["dnssec_enabled"] = false
		measurement.Metadata["dnssec_validated"] = false
		measurement.Metadata["dnssec_chain_valid"] = false
		measurement.Metadata["rrsig_count"] = 0
		return measurement, nil
	}

	_, rcode, truncated, authenticated := parseDNSFlags(resp)

	// Retry over TCP if truncated
	if truncated {
		tcpCtx, tcpCancel := context.WithTimeout(ctx, timeout)
		defer tcpCancel()
		tcpStart := time.Now()
		resp2, tcpErr := sendDNSQueryTCP(tcpCtx, nameserver, query)
		if tcpErr == nil {
			resp = resp2
			tcpElapsed := float64(time.Since(tcpStart).Microseconds()) / 1000.0
			measurement.LatencyMs = &tcpElapsed
			_, rcode, _, authenticated = parseDNSFlags(resp)
		}
	}

	rrsigCount := countRRSIG(resp)
	dnssecEnabled := rrsigCount > 0

	// dnssec_chain_valid: RRSIG present + AD bit set is the strongest indication
	// available without a full validation stack.
	chainValid := dnssecEnabled && authenticated

	measurement.Metadata["dnssec_enabled"] = dnssecEnabled
	measurement.Metadata["dnssec_validated"] = authenticated
	measurement.Metadata["dnssec_chain_valid"] = chainValid
	measurement.Metadata["rrsig_count"] = rrsigCount

	if rcode == 2 { // SERVFAIL may indicate DNSSEC validation failure upstream
		measurement.Success = false
		measurement.ErrorMessage = "SERVFAIL: possible DNSSEC validation failure"
		return measurement, nil
	}

	measurement.Success = rcode == 0
	if !measurement.Success {
		measurement.ErrorMessage = fmt.Sprintf("DNS RCODE %d", rcode)
	}
	return measurement, nil
}

// ---------------------------------------------------------------------------
// Monitor 3: doh_dot
// ---------------------------------------------------------------------------

// DoHDoTMonitor measures DNS-over-HTTPS and DNS-over-TLS performance.
type DoHDoTMonitor struct{}

func (m *DoHDoTMonitor) Type() string { return "doh_dot" }

func (m *DoHDoTMonitor) ValidateConfig(config map[string]interface{}) error {
	if to, ok := config["timeout_seconds"].(float64); ok && to <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}
	return nil
}

func (m *DoHDoTMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	domain := getStringConfig(config, "domain", target.Address)
	if domain == "" {
		domain = "google.com"
	}
	dohServer := getStringConfig(config, "doh_server", "https://1.1.1.1/dns-query")
	dotServer := getStringConfig(config, "dot_server", "1.1.1.1:853")
	if _, _, err := net.SplitHostPort(dotServer); err != nil {
		dotServer = dotServer + ":853"
	}
	timeoutSec := getFloat64Config(config, "timeout_seconds", 10)
	timeout := time.Duration(timeoutSec * float64(time.Second))

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "doh_dot",
		Metadata:    make(map[string]interface{}),
	}

	// Build a standard A query (no DNSSEC, just for latency measurement)
	query := buildDNSQuery(domain, 1, 0x5678, false)

	// Run DoH and DoT concurrently
	type dohResult struct {
		latencyMs  float64
		statusCode int
		err        error
	}
	type dotResult struct {
		latencyMs    float64
		tlsTimeMs    float64
		err          error
	}

	dohCh := make(chan dohResult, 1)
	dotCh := make(chan dotResult, 1)

	go func() {
		dohCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		lat, code, err := measureDoH(dohCtx, dohServer, query)
		dohCh <- dohResult{lat, code, err}
	}()

	go func() {
		dotCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		lat, tlsTime, err := measureDoT(dotCtx, dotServer, query)
		dotCh <- dotResult{lat, tlsTime, err}
	}()

	dohRes := <-dohCh
	dotRes := <-dotCh

	// DoH metrics
	measurement.Metadata["doh_status_code"] = dohRes.statusCode
	if dohRes.err != nil {
		measurement.Metadata["doh_latency_ms"] = nil
	} else {
		measurement.Metadata["doh_latency_ms"] = dohRes.latencyMs
	}

	// DoT metrics
	measurement.Metadata["dot_tls_time_ms"] = dotRes.tlsTimeMs
	if dotRes.err != nil {
		measurement.Metadata["dot_latency_ms"] = nil
	} else {
		measurement.Metadata["dot_latency_ms"] = dotRes.latencyMs
	}

	// Determine preferred protocol and overall latency
	dohOK := dohRes.err == nil
	dotOK := dotRes.err == nil

	switch {
	case dohOK && dotOK:
		if dohRes.latencyMs <= dotRes.latencyMs {
			measurement.Metadata["preferred"] = "doh"
			lat := dohRes.latencyMs
			measurement.LatencyMs = &lat
		} else {
			measurement.Metadata["preferred"] = "dot"
			lat := dotRes.latencyMs
			measurement.LatencyMs = &lat
		}
		measurement.Success = true
	case dohOK:
		measurement.Metadata["preferred"] = "doh"
		lat := dohRes.latencyMs
		measurement.LatencyMs = &lat
		measurement.Success = true
		measurement.ErrorMessage = fmt.Sprintf("DoT failed: %v", dotRes.err)
	case dotOK:
		measurement.Metadata["preferred"] = "dot"
		lat := dotRes.latencyMs
		measurement.LatencyMs = &lat
		measurement.Success = true
		measurement.ErrorMessage = fmt.Sprintf("DoH failed: %v", dohRes.err)
	default:
		measurement.Success = false
		measurement.Metadata["preferred"] = ""
		measurement.ErrorMessage = fmt.Sprintf("DoH: %v; DoT: %v", dohRes.err, dotRes.err)
	}

	return measurement, nil
}

// measureDoH sends a DoH POST query and returns (latencyMs, statusCode, error).
func measureDoH(ctx context.Context, server string, query []byte) (float64, int, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, server, bytes.NewReader(query))
	if err != nil {
		return 0, 0, fmt.Errorf("doh request build: %w", err)
	}
	req.Header.Set("Content-Type", "application/dns-message")
	req.Header.Set("Accept", "application/dns-message")

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return elapsed, 0, fmt.Errorf("doh request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode != http.StatusOK {
		return elapsed, resp.StatusCode, fmt.Errorf("doh status %d", resp.StatusCode)
	}
	return elapsed, resp.StatusCode, nil
}

// measureDoT dials TLS to a DoT server, sends a DNS query, and returns
// (total latency ms, TLS handshake ms, error).
func measureDoT(ctx context.Context, server string, query []byte) (float64, float64, error) {
	host, _, err := net.SplitHostPort(server)
	if err != nil {
		host = server
	}

	tlsCfg := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	dialStart := time.Now()

	// Use a net.Dialer with context for pure TCP connection first to measure separately
	dialer := &net.Dialer{}
	tcpConn, err := dialer.DialContext(ctx, "tcp", server)
	if err != nil {
		return 0, 0, fmt.Errorf("dot tcp dial: %w", err)
	}

	// TLS handshake
	tlsStart := time.Now()
	tlsConn := tls.Client(tcpConn, tlsCfg)
	if deadline, ok := ctx.Deadline(); ok {
		tlsConn.SetDeadline(deadline)
	}
	if err := tlsConn.Handshake(); err != nil {
		tlsConn.Close()
		return 0, 0, fmt.Errorf("dot tls handshake: %w", err)
	}
	tlsElapsed := float64(time.Since(tlsStart).Microseconds()) / 1000.0
	defer tlsConn.Close()

	// Send DNS query with 2-byte length prefix
	length := make([]byte, 2)
	binary.BigEndian.PutUint16(length, uint16(len(query)))
	if _, err := tlsConn.Write(append(length, query...)); err != nil {
		return 0, tlsElapsed, fmt.Errorf("dot write: %w", err)
	}

	// Read response length
	var respLen uint16
	if err := binary.Read(tlsConn, binary.BigEndian, &respLen); err != nil {
		return 0, tlsElapsed, fmt.Errorf("dot read length: %w", err)
	}

	respBody := make([]byte, respLen)
	if _, err := io.ReadFull(tlsConn, respBody); err != nil {
		return 0, tlsElapsed, fmt.Errorf("dot read body: %w", err)
	}

	totalElapsed := float64(time.Since(dialStart).Microseconds()) / 1000.0
	return totalElapsed, tlsElapsed, nil
}
