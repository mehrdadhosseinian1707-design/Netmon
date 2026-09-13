package monitors

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ocsp"

	"github.com/netmon/netmon/internal/models"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func ptrFloat(v float64) *float64 { return &v }

func mean(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := mean(vals)
	sum := 0.0
	for _, v := range vals {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}

// ---------------------------------------------------------------------------
// Monitor 1: http_advanced
// ---------------------------------------------------------------------------

type HTTPAdvancedMonitor struct{}

func NewHTTPAdvancedMonitor(_ map[string]interface{}) (*HTTPAdvancedMonitor, error) {
	return &HTTPAdvancedMonitor{}, nil
}

func (m *HTTPAdvancedMonitor) Type() string { return "http_advanced" }

func (m *HTTPAdvancedMonitor) ValidateConfig(config map[string]interface{}) error {
	if u, ok := config["url"].(string); ok && u != "" {
		if _, err := url.ParseRequestURI(u); err != nil {
			return fmt.Errorf("invalid url: %v", err)
		}
	}
	if ts, ok := config["timeout_seconds"].(float64); ok && ts <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0")
	}
	return nil
}

func (m *HTTPAdvancedMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	rawURL, _ := config["url"].(string)
	if rawURL == "" {
		rawURL = "http://" + target.Address
	}

	timeoutSecs := 15.0
	if v, ok := config["timeout_seconds"].(float64); ok && v > 0 {
		timeoutSecs = v
	}
	reqCount := 5
	if v, ok := config["request_count"].(float64); ok && v > 0 {
		reqCount = int(v)
	}
	timeout := time.Duration(timeoutSecs * float64(time.Second))

	transport := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: false},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too_many_redirects")
			}
			return nil
		},
	}

	type reqResult struct {
		ttfbMs    float64
		totalMs   float64
		size      float64
		status    int
		errClass  string
		errMsg    string
	}

	results := make([]reqResult, 0, reqCount)
	var lastErrClass, lastErrMsg string
	var lastStatus int

	for i := 0; i < reqCount; i++ {
		r := doHTTPRequest(ctx, client, rawURL)
		results = append(results, reqResult{
			ttfbMs:   r.ttfbMs,
			totalMs:  r.totalMs,
			size:     r.size,
			status:   r.status,
			errClass: r.errClass,
			errMsg:   r.errMsg,
		})
		lastErrClass = r.errClass
		lastErrMsg = r.errMsg
		lastStatus = r.status
	}

	// Aggregate
	ttfbs := make([]float64, 0, reqCount)
	totals := make([]float64, 0, reqCount)
	sizes := make([]float64, 0, reqCount)
	sizeInts := make([]interface{}, 0, reqCount)
	successCount := 0

	for _, r := range results {
		if r.errClass == "2xx_ok" {
			ttfbs = append(ttfbs, r.ttfbMs)
			totals = append(totals, r.totalMs)
			sizes = append(sizes, r.size)
			sizeInts = append(sizeInts, r.size)
			successCount++
		}
	}

	success := successCount > 0
	avgTTFB := mean(ttfbs)
	avgTotal := mean(totals)

	// Download consistency
	downloadCV := 0.0
	consistent := true
	if len(sizes) > 1 {
		m2 := mean(sizes)
		if m2 > 0 {
			downloadCV = stddev(sizes) / m2
			if downloadCV > 0.1 {
				consistent = false
			}
		}
	}

	errClass := lastErrClass
	if success {
		errClass = "2xx_ok"
	}

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "http_advanced",
		Success:     success,
		Metadata:    make(map[string]interface{}),
	}

	if !success {
		measurement.ErrorMessage = lastErrMsg
	}

	measurement.Metadata["http_error_class"] = errClass
	measurement.Metadata["status_code"] = lastStatus
	measurement.Metadata["download_cv"] = downloadCV
	measurement.Metadata["download_consistent"] = consistent
	measurement.Metadata["response_sizes"] = sizeInts

	if success {
		measurement.HTTPTTFBMs = ptrFloat(avgTTFB)
		measurement.HTTPTotalTimeMs = ptrFloat(avgTotal)
		measurement.LatencyMs = ptrFloat(avgTotal)
	}

	return measurement, nil
}

type httpSingleResult struct {
	ttfbMs   float64
	totalMs  float64
	size     float64
	status   int
	errClass string
	errMsg   string
}

func doHTTPRequest(ctx context.Context, client *http.Client, rawURL string) httpSingleResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return httpSingleResult{errClass: "connection_reset", errMsg: err.Error()}
	}

	start := time.Now()
	resp, err := client.Do(req)
	ttfbMs := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return httpSingleResult{errClass: classifyHTTPError(err), errMsg: err.Error()}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	totalMs := float64(time.Since(start).Microseconds()) / 1000.0
	if err != nil {
		return httpSingleResult{
			status:   resp.StatusCode,
			errClass: "connection_reset",
			errMsg:   err.Error(),
			ttfbMs:   ttfbMs,
			totalMs:  totalMs,
		}
	}

	size := float64(len(body))
	code := resp.StatusCode
	errClass := httpStatusClass(code)

	return httpSingleResult{
		ttfbMs:   ttfbMs,
		totalMs:  totalMs,
		size:     size,
		status:   code,
		errClass: errClass,
	}
}

func classifyHTTPError(err error) string {
	if err == nil {
		return "2xx_ok"
	}
	msg := err.Error()
	if strings.Contains(msg, "too_many_redirects") || strings.Contains(msg, "stopped after") {
		return "too_many_redirects"
	}
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "timeout") {
		return "timeout"
	}
	if strings.Contains(msg, "no such host") || strings.Contains(msg, "lookup") {
		return "dns_failure"
	}
	if strings.Contains(msg, "tls") || strings.Contains(msg, "certificate") || strings.Contains(msg, "x509") {
		return "tls_error"
	}
	if strings.Contains(msg, "connection reset") || strings.Contains(msg, "EOF") {
		return "connection_reset"
	}
	if strings.Contains(msg, "response body too large") {
		return "response_too_large"
	}
	return "connection_reset"
}

func httpStatusClass(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx_ok"
	case code >= 300 && code < 400:
		return "3xx_redirect"
	case code >= 400 && code < 500:
		return "4xx_client_error"
	case code >= 500:
		return "5xx_server_error"
	default:
		return "connection_reset"
	}
}

// ---------------------------------------------------------------------------
// Monitor 2: tls_advanced
// ---------------------------------------------------------------------------

type TLSAdvancedMonitor struct{}

func NewTLSAdvancedMonitor(_ map[string]interface{}) (*TLSAdvancedMonitor, error) {
	return &TLSAdvancedMonitor{}, nil
}

func (m *TLSAdvancedMonitor) Type() string { return "tls_advanced" }

func (m *TLSAdvancedMonitor) ValidateConfig(config map[string]interface{}) error {
	if p, ok := config["port"].(float64); ok && (p < 1 || p > 65535) {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}

func (m *TLSAdvancedMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	host, _ := config["host"].(string)
	if host == "" {
		host = target.Address
		// strip port if present
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}
	}

	port := 443
	if v, ok := config["port"].(float64); ok && v > 0 {
		port = int(v)
	}

	timeoutSecs := 15.0
	if v, ok := config["timeout_seconds"].(float64); ok && v > 0 {
		timeoutSecs = v
	}
	timeout := time.Duration(timeoutSecs * float64(time.Second))

	addr := fmt.Sprintf("%s:%d", host, port)

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "tls_advanced",
		Metadata:    make(map[string]interface{}),
	}

	// Primary TLS attempt (default: best available)
	tlsResult := doTLSHandshake(ctx, addr, host, timeout, 0, 0)

	measurement.Metadata["tls_failure_class"] = tlsResult.failureClass
	measurement.Metadata["cert_expiry_days"] = tlsResult.certExpiryDays
	measurement.Metadata["ocsp_status"] = tlsResult.ocspStatus

	if tlsResult.handshakeMs > 0 {
		measurement.TLSHandshakeTimeMs = ptrFloat(tlsResult.handshakeMs)
	}

	measurement.Success = tlsResult.failureClass == "ok"
	if !measurement.Success {
		measurement.ErrorMessage = tlsResult.errMsg
	}

	// Test individual TLS versions
	type versionTest struct {
		name    string
		version uint16
	}
	versions := []versionTest{
		{"TLS1.0", tls.VersionTLS10},
		{"TLS1.1", tls.VersionTLS11},
		{"TLS1.2", tls.VersionTLS12},
		{"TLS1.3", tls.VersionTLS13},
	}

	supported := []string{}
	for _, vt := range versions {
		r := doTLSHandshake(ctx, addr, host, timeout, vt.version, vt.version)
		if r.failureClass == "ok" {
			supported = append(supported, vt.name)
		}
	}
	measurement.Metadata["tls_versions_supported"] = supported

	return measurement, nil
}

type tlsHandshakeResult struct {
	failureClass   string
	handshakeMs    float64
	certExpiryDays int
	ocspStatus     string
	errMsg         string
}

func doTLSHandshake(ctx context.Context, addr, host string, timeout time.Duration, minVer, maxVer uint16) tlsHandshakeResult {
	cfg := &tls.Config{
		ServerName: host,
	}
	if minVer != 0 {
		cfg.MinVersion = minVer
	}
	if maxVer != 0 {
		cfg.MaxVersion = maxVer
	}

	dialCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	dialer := &net.Dialer{}
	tcpConn, err := dialer.DialContext(dialCtx, "tcp", addr)
	if err != nil {
		if isTimeout(err) {
			return tlsHandshakeResult{failureClass: "handshake_timeout", errMsg: err.Error()}
		}
		return tlsHandshakeResult{failureClass: "connection_reset", errMsg: err.Error()}
	}
	defer tcpConn.Close()

	tlsConn := tls.Client(tcpConn, cfg)
	_ = tlsConn.SetDeadline(time.Now().Add(timeout))

	start := time.Now()
	err = tlsConn.HandshakeContext(dialCtx)
	handshakeMs := float64(time.Since(start).Microseconds()) / 1000.0

	if err != nil {
		return tlsHandshakeResult{
			failureClass: classifyTLSError(err),
			handshakeMs:  handshakeMs,
			errMsg:       err.Error(),
		}
	}

	state := tlsConn.ConnectionState()

	// Cert expiry
	certExpiryDays := 0
	if len(state.PeerCertificates) > 0 {
		cert := state.PeerCertificates[0]
		now := time.Now()

		if cert.NotAfter.Before(now) {
			return tlsHandshakeResult{
				failureClass:   "cert_expired",
				handshakeMs:    handshakeMs,
				certExpiryDays: int(cert.NotAfter.Sub(now).Hours() / 24),
				ocspStatus:     "unknown",
			}
		}
		if cert.NotBefore.After(now) {
			return tlsHandshakeResult{
				failureClass:   "cert_not_yet_valid",
				handshakeMs:    handshakeMs,
				certExpiryDays: 0,
				ocspStatus:     "unknown",
			}
		}
		certExpiryDays = int(cert.NotAfter.Sub(now).Hours() / 24)
	}

	// OCSP stapling check
	ocspStatus := "unknown"
	if len(state.OCSPResponse) > 0 {
		ocspStatus = parseOCSPResponse(state.OCSPResponse)
	}

	return tlsHandshakeResult{
		failureClass:   "ok",
		handshakeMs:    handshakeMs,
		certExpiryDays: certExpiryDays,
		ocspStatus:     ocspStatus,
	}
}

func classifyTLSError(err error) string {
	if err == nil {
		return "ok"
	}
	if isTimeout(err) {
		return "handshake_timeout"
	}

	msg := err.Error()

	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return "hostname_mismatch"
	}
	var unknownAuthErr x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthErr) {
		return "untrusted_ca"
	}
	var certErr x509.CertificateInvalidError
	if errors.As(err, &certErr) {
		if certErr.Reason == x509.Expired {
			return "cert_expired"
		}
	}

	lmsg := strings.ToLower(msg)
	if strings.Contains(lmsg, "protocol version") || strings.Contains(lmsg, "no supported versions") {
		return "protocol_version"
	}
	if strings.Contains(lmsg, "no cipher") || strings.Contains(lmsg, "handshake failure") {
		return "cipher_mismatch"
	}
	if strings.Contains(lmsg, "certificate has expired") || strings.Contains(lmsg, "certificate is not yet valid") {
		if strings.Contains(lmsg, "not yet valid") {
			return "cert_not_yet_valid"
		}
		return "cert_expired"
	}
	if strings.Contains(lmsg, "revoked") {
		return "cert_revoked"
	}
	return "tls_error"
}

func parseOCSPResponse(raw []byte) string {
	resp, err := ocsp.ParseResponse(raw, nil)
	if err != nil {
		return "parse_error"
	}
	switch resp.Status {
	case ocsp.Good:
		return "good"
	case ocsp.Revoked:
		return "revoked"
	case ocsp.Unknown:
		return "unknown"
	default:
		return "unknown"
	}
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	return strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "timeout")
}

// ---------------------------------------------------------------------------
// Monitor 3: one_way_delay
// ---------------------------------------------------------------------------

type OneWayDelayMonitor struct{}

func NewOneWayDelayMonitor(_ map[string]interface{}) (*OneWayDelayMonitor, error) {
	return &OneWayDelayMonitor{}, nil
}

func (m *OneWayDelayMonitor) Type() string { return "one_way_delay" }

func (m *OneWayDelayMonitor) ValidateConfig(config map[string]interface{}) error {
	if ts, ok := config["timeout_seconds"].(float64); ok && ts <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0")
	}
	return nil
}

func (m *OneWayDelayMonitor) Check(ctx context.Context, target *models.Target) (*models.Measurement, error) {
	config := map[string]interface{}{}
	if target.Metadata != nil {
		config = target.Metadata
	}

	host := target.Address
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}

	port := 80
	if v, ok := config["port"].(float64); ok && v > 0 {
		port = int(v)
	}

	timeoutSecs := 15.0
	if v, ok := config["timeout_seconds"].(float64); ok && v > 0 {
		timeoutSecs = v
	}
	timeout := time.Duration(timeoutSecs * float64(time.Second))

	probeCount := 20

	measurement := &models.Measurement{
		Time:        time.Now(),
		MonitorType: "one_way_delay",
		Metadata:    make(map[string]interface{}),
	}

	// Method 1: OWAMP-style UDP echo
	owdForward, owdReverse, owdAsymmetry, rttMs, jitterMs, ok := tryOWAMPStyle(ctx, host, timeout, probeCount)
	method := "owamp"

	if !ok {
		// Method 2: fallback TCP RTT split
		owdForward, owdReverse, owdAsymmetry, rttMs, jitterMs = tcpRTTSplit(ctx, host, port, timeout, probeCount)
		method = "rtt_estimate"
	}

	measurement.Success = rttMs > 0
	if !measurement.Success {
		measurement.ErrorMessage = "unable to measure delay"
	}

	measurement.LatencyMs = ptrFloat(rttMs)
	measurement.JitterMs = ptrFloat(jitterMs)
	measurement.Metadata["owd_forward_ms"] = owdForward
	measurement.Metadata["owd_reverse_ms"] = owdReverse
	measurement.Metadata["owd_asymmetry_ms"] = owdAsymmetry
	measurement.Metadata["owd_method"] = method

	return measurement, nil
}

// tryOWAMPStyle attempts UDP-based OWD measurement using echo.
// It listens on a local UDP port, sends packets with embedded timestamps to
// the target's echo service (port 7 / UDP echo), and computes OWD from RTT.
// Returns (fwd, rev, asymmetry, avgRTT, jitter, success).
func tryOWAMPStyle(ctx context.Context, host string, timeout time.Duration, count int) (float64, float64, float64, float64, float64, bool) {
	// Resolve target
	targetUDP := fmt.Sprintf("%s:7", host) // UDP echo port
	raddr, err := net.ResolveUDPAddr("udp", targetUDP)
	if err != nil {
		return 0, 0, 0, 0, 0, false
	}

	conn, err := net.DialUDP("udp", nil, raddr)
	if err != nil {
		return 0, 0, 0, 0, 0, false
	}
	defer conn.Close()

	rtts := make([]float64, 0, count)
	buf := make([]byte, 16)

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			break
		default:
		}

		sendTs := time.Now().UnixNano()
		binary.BigEndian.PutUint64(buf[:8], uint64(sendTs))
		binary.BigEndian.PutUint64(buf[8:16], uint64(i))

		_ = conn.SetDeadline(time.Now().Add(timeout / time.Duration(count)))
		_, err := conn.Write(buf)
		if err != nil {
			continue
		}

		resp := make([]byte, 16)
		n, err := conn.Read(resp)
		recvTs := time.Now().UnixNano()
		if err != nil || n < 8 {
			continue
		}

		origTs := int64(binary.BigEndian.Uint64(resp[:8]))
		if origTs != sendTs {
			continue
		}
		rttNs := recvTs - sendTs
		if rttNs > 0 {
			rtts = append(rtts, float64(rttNs)/1e6)
		}
	}

	if len(rtts) == 0 {
		return 0, 0, 0, 0, 0, false
	}

	avgRTT := mean(rtts)
	jitter := stddev(rtts)

	// OWD approximation: symmetric split
	owdFwd := avgRTT / 2.0
	owdRev := avgRTT / 2.0

	// Asymmetry: stddev of deviation from symmetric split across samples
	minRTT := rtts[0]
	for _, r := range rtts {
		if r < minRTT {
			minRTT = r
		}
	}
	asymVals := make([]float64, len(rtts))
	for i, r := range rtts {
		asymVals[i] = r - minRTT
	}
	asymmetry := stddev(asymVals)

	return owdFwd, owdRev, asymmetry, avgRTT, jitter, true
}

// tcpRTTSplit estimates OWD by measuring TCP connect RTT repeatedly.
// Sends probeCount TCP SYN probes (by dialing and timing to connect),
// uses min RTT as baseline and computes asymmetry from RTT variability.
func tcpRTTSplit(ctx context.Context, host string, port int, timeout time.Duration, count int) (float64, float64, float64, float64, float64) {
	addr := fmt.Sprintf("%s:%d", host, port)
	rtts := make([]float64, 0, count)
	perProbeTimeout := timeout / time.Duration(count)
	if perProbeTimeout < 500*time.Millisecond {
		perProbeTimeout = 500 * time.Millisecond
	}

	dialer := &net.Dialer{}
	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			goto done
		default:
		}

		probeCtx, cancel := context.WithTimeout(ctx, perProbeTimeout)
		start := time.Now()
		conn, err := dialer.DialContext(probeCtx, "tcp", addr)
		elapsed := float64(time.Since(start).Microseconds()) / 1000.0
		cancel()
		if err == nil {
			conn.Close()
			rtts = append(rtts, elapsed)
		}
	}

done:
	if len(rtts) == 0 {
		return 0, 0, 0, 0, 0
	}

	avgRTT := mean(rtts)
	jitter := stddev(rtts)

	minRTT := rtts[0]
	for _, r := range rtts {
		if r < minRTT {
			minRTT = r
		}
	}

	owdFwd := minRTT / 2.0
	owdRev := minRTT / 2.0

	// Asymmetry: stddev of (rtt - minRTT) across probes
	asymVals := make([]float64, len(rtts))
	for i, r := range rtts {
		asymVals[i] = r - minRTT
	}
	asymmetry := stddev(asymVals)

	return owdFwd, owdRev, asymmetry, avgRTT, jitter
}
