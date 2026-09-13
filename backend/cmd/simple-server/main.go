package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// CORS middleware
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Data types
type Target struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Address     string    `json:"address"`
	Port        int       `json:"port,omitempty"`
	Enabled     bool      `json:"enabled"`
	Status      string    `json:"status"`
	LastCheck   string    `json:"last_check,omitempty"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Probe struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Location    string    `json:"location"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	LastSeen    string    `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Monitor struct {
	ID              string                 `json:"id"`
	TargetID        string                 `json:"target_id"`
	MonitorType     string                 `json:"monitor_type"`
	IntervalSeconds int                    `json:"interval_seconds"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	Enabled         bool                   `json:"enabled"`
	Config          map[string]interface{} `json:"config"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type TargetStats struct {
	TargetID       string  `json:"target_id"`
	Availability   float64 `json:"availability"`
	AvgLatency     float64 `json:"avg_latency"`
	MinLatency     float64 `json:"min_latency"`
	MaxLatency     float64 `json:"max_latency"`
	PacketLoss     float64 `json:"packet_loss"`
	TotalChecks    int     `json:"total_checks"`
	SuccessfulChecks int   `json:"successful_checks"`
	FailedChecks   int     `json:"failed_checks"`
}

type Measurement struct {
	ID          string                 `json:"id"`
	TargetID    string                 `json:"target_id"`
	MonitorID   string                 `json:"monitor_id"`
	ProbeID     string                 `json:"probe_id"`
	Success     bool                   `json:"success"`
	Latency     float64                `json:"latency"`
	PacketLoss  float64                `json:"packet_loss"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
}

// In-memory storage
var (
	targets  = make(map[string]*Target)
	probes   = make(map[string]*Probe)
	monitors = make(map[string]*Monitor)
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket handler
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	log.Println("WebSocket client connected")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			msg := map[string]interface{}{
				"type": "measurement",
				"data": map[string]interface{}{
					"latency_ms":    20 + (time.Now().Unix() % 30),
					"packet_loss":   float64(time.Now().Unix()%5) * 0.5,
					"download_mbps": 80 + float64(time.Now().Unix()%40),
					"upload_mbps":   40 + float64(time.Now().Unix()%30),
				},
				"timestamp": time.Now().Format(time.RFC3339),
			}

			if err := conn.WriteJSON(msg); err != nil {
				log.Println("WebSocket write error:", err)
				return
			}
		}
	}
}

// Probes handlers
func getProbes(w http.ResponseWriter, r *http.Request) {
	probeList := make([]*Probe, 0, len(probes))
	for _, p := range probes {
		probeList = append(probeList, p)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(probeList)
}

func createProbe(w http.ResponseWriter, r *http.Request) {
	var probe Probe
	if err := json.NewDecoder(r.Body).Decode(&probe); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	probe.ID = uuid.New().String()
	probe.Status = "active"
	probe.LastSeen = time.Now().Format(time.RFC3339)
	probe.CreatedAt = time.Now()
	probe.UpdatedAt = time.Now()

	probes[probe.ID] = &probe

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(probe)
}

func getProbe(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	probe, exists := probes[id]
	if !exists {
		http.Error(w, "Probe not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(probe)
}

// Targets handlers
func getTargets(w http.ResponseWriter, r *http.Request) {
	targetList := make([]*Target, 0, len(targets))
	for _, t := range targets {
		targetList = append(targetList, t)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targetList)
}

func createTarget(w http.ResponseWriter, r *http.Request) {
	var target Target
	if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	target.ID = uuid.New().String()
	target.Enabled = true
	target.Status = "up"
	target.LastCheck = time.Now().Format(time.RFC3339)
	target.CreatedAt = time.Now()
	target.UpdatedAt = time.Now()

	targets[target.ID] = &target

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}

func getTarget(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	target, exists := targets[id]
	if !exists {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(target)
}

func getTargetStats(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, exists := targets[id]
	if !exists {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	stats := TargetStats{
		TargetID:         id,
		Availability:     99.5,
		AvgLatency:       25.3,
		MinLatency:       15.2,
		MaxLatency:       45.8,
		PacketLoss:       0.5,
		TotalChecks:      1000,
		SuccessfulChecks: 995,
		FailedChecks:     5,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func getTargetMeasurements(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	_, exists := targets[id]
	if !exists {
		http.Error(w, "Target not found", http.StatusNotFound)
		return
	}

	measurements := make([]Measurement, 0)
	now := time.Now()

	for i := 0; i < 50; i++ {
		measurements = append(measurements, Measurement{
			ID:         uuid.New().String(),
			TargetID:   id,
			MonitorID:  uuid.New().String(),
			ProbeID:    uuid.New().String(),
			Success:    i%10 != 0,
			Latency:    20 + float64(i%30),
			PacketLoss: float64(i%5) * 0.5,
			Data:       map[string]interface{}{},
			Timestamp:  now.Add(-time.Duration(i) * time.Minute),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(measurements)
}

// Monitors handlers
func getMonitors(w http.ResponseWriter, r *http.Request) {
	monitorList := make([]*Monitor, 0, len(monitors))
	for _, m := range monitors {
		monitorList = append(monitorList, m)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitorList)
}

func createMonitor(w http.ResponseWriter, r *http.Request) {
	var monitor Monitor
	if err := json.NewDecoder(r.Body).Decode(&monitor); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	monitor.ID = uuid.New().String()
	monitor.Enabled = true
	monitor.CreatedAt = time.Now()
	monitor.UpdatedAt = time.Now()

	monitors[monitor.ID] = &monitor

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(monitor)
}

func getMonitor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	monitor, exists := monitors[id]
	if !exists {
		http.Error(w, "Monitor not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(monitor)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func pingTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
		Count  int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Count == 0 {
		req.Count = 3
	}

	// Simulate ping with random latency
	minLatency := 10.0 + float64(time.Now().Unix()%20)
	maxLatency := minLatency + float64(time.Now().Unix()%30)
	avgLatency := (minLatency + maxLatency) / 2

	result := map[string]interface{}{
		"target":      req.Target,
		"packets_sent": req.Count,
		"packets_recv": req.Count,
		"packet_loss":  0.0,
		"min_rtt":      minLatency,
		"avg_rtt":      avgLatency,
		"max_rtt":      maxLatency,
		"latency":      avgLatency,
		"success":      true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func dohCheck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Try to access the DoH server with a proper DNS query
	startTime := time.Now()
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}

	// Send a simple DoH query (query for cloudflare.com A record)
	dohQuery := req.URL + "?name=cloudflare.com&type=A"
	httpReq, err := http.NewRequest("GET", dohQuery, nil)
	if err != nil {
		// If query fails, just check base URL
		resp, err := client.Head(req.URL)
		responseTime := time.Since(startTime).Milliseconds()

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"url":           req.URL,
				"accessible":    false,
				"response_time": responseTime,
			})
			return
		}
		defer resp.Body.Close()

		result := map[string]interface{}{
			"url":           req.URL,
			"accessible":    resp.StatusCode < 500,
			"response_time": responseTime,
			"status_code":   resp.StatusCode,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	httpReq.Header.Set("Accept", "application/dns-json")
	resp, err := client.Do(httpReq)
	responseTime := time.Since(startTime).Milliseconds()

	accessible := false
	statusCode := 0

	if err == nil {
		defer resp.Body.Close()
		statusCode = resp.StatusCode
		// DoH servers typically return 200, 400, or 415
		// Even 400/415 means the server is accessible
		accessible = statusCode < 500
	}

	result := map[string]interface{}{
		"url":           req.URL,
		"accessible":    accessible,
		"response_time": responseTime,
		"status_code":   statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func getIPInfo(w http.ResponseWriter, r *http.Request) {
	// Simulate IP info (in production, you'd use a real IP geolocation service)
	result := map[string]interface{}{
		"ip":       "203.0.113." + string(rune(50+time.Now().Unix()%200)),
		"location": "United States",
		"city":     "New York",
		"region":   "NY",
		"country":  "US",
		"isp":      "Example ISP",
		"timezone": "America/New_York",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func downloadSpeedTest(w http.ResponseWriter, r *http.Request) {
	// Simulate download speed test
	// Generate random data between 30-120 Mbps
	speedMbps := 30.0 + float64(time.Now().Unix()%90)

	result := map[string]interface{}{
		"speed_mbps": speedMbps,
		"test_size":  "10MB",
		"duration":   (10.0 / speedMbps) * 8.0, // seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func uploadSpeedTest(w http.ResponseWriter, r *http.Request) {
	// Simulate upload speed test
	// Generate random data between 10-50 Mbps (typically lower than download)
	speedMbps := 10.0 + float64(time.Now().Unix()%40)

	result := map[string]interface{}{
		"speed_mbps": speedMbps,
		"test_size":  "5MB",
		"duration":   (5.0 / speedMbps) * 8.0, // seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func jitterTest(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
		Count  int    `json:"count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Count == 0 {
		req.Count = 10
	}

	// Simulate jitter calculation (variation in ping times)
	// Generate random jitter between 1-20ms
	jitter := 1.0 + float64(time.Now().Unix()%19)

	result := map[string]interface{}{
		"target":       req.Target,
		"packets_sent": req.Count,
		"jitter":       jitter,
		"avg_latency":  20 + float64(time.Now().Unix()%30),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func tracerouteTool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Target string `json:"target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Simulate realistic traceroute with 8-15 hops
	numHops := 8 + int(time.Now().Unix()%8)
	hops := make([]map[string]interface{}, 0, numHops)

	// Common network patterns for realistic simulation
	ipPrefixes := []string{
		"192.168.1",   // Local network
		"10.0.0",      // Private network
		"172.16.0",    // ISP gateway
		"203.0.113",   // ISP backbone
		"198.51.100",  // Internet backbone
		"8.8.8",       // Near destination
	}

	for i := 1; i <= numHops; i++ {
		// Calculate realistic RTT that increases with hop count
		baseRTT := 5.0 + float64(i)*3.5
		variation := float64(time.Now().Unix()%10) / 2.0

		rtt1 := baseRTT + variation
		rtt2 := baseRTT + variation + float64(time.Now().Unix()%5)
		rtt3 := baseRTT + variation - float64(time.Now().Unix()%3)
		avgRTT := (rtt1 + rtt2 + rtt3) / 3.0

		// Generate realistic IP based on hop position
		var ip string
		var hostname string

		if i < len(ipPrefixes) {
			ip = fmt.Sprintf("%s.%d", ipPrefixes[i-1], 1+int(time.Now().Unix()%254))

			// Generate hostname based on hop type
			switch {
			case i == 1:
				hostname = "gateway.local"
			case i == 2:
				hostname = "router.local"
			case i <= 4:
				hostname = fmt.Sprintf("isp-gateway-%d.net", i)
			case i <= 6:
				hostname = fmt.Sprintf("backbone-%d.transit.net", i)
			default:
				hostname = fmt.Sprintf("edge-%d.cdn.net", i)
			}
		} else {
			// Close to destination
			ip = req.Target
			hostname = req.Target
		}

		hop := map[string]interface{}{
			"number":   i,
			"ip":       ip,
			"hostname": hostname,
			"rtt1":     rtt1,
			"rtt2":     rtt2,
			"rtt3":     rtt3,
			"avgRtt":   avgRTT,
		}

		hops = append(hops, hop)
	}

	result := map[string]interface{}{
		"target":     req.Target,
		"hops":       hops,
		"total_hops": numHops,
		"success":    true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func initMockData() {
	// Add default probes
	now := time.Now()

	probe1 := &Probe{
		ID:          uuid.New().String(),
		Name:        "Main Probe",
		Location:    "US-East",
		Description: "Primary monitoring probe",
		Status:      "active",
		LastSeen:    now.Format(time.RFC3339),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	probes[probe1.ID] = probe1

	probe2 := &Probe{
		ID:          uuid.New().String(),
		Name:        "EU Probe",
		Location:    "EU-West",
		Description: "European monitoring probe",
		Status:      "active",
		LastSeen:    now.Format(time.RFC3339),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	probes[probe2.ID] = probe2

	// Add default targets
	target1 := &Target{
		ID:          uuid.New().String(),
		Name:        "Google DNS",
		Type:        "host",
		Address:     "8.8.8.8",
		Enabled:     true,
		Status:      "up",
		LastCheck:   now.Format(time.RFC3339),
		Description: "Google Public DNS",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	targets[target1.ID] = target1

	target2 := &Target{
		ID:          uuid.New().String(),
		Name:        "Production API",
		Type:        "server",
		Address:     "api.example.com",
		Port:        443,
		Enabled:     true,
		Status:      "up",
		LastCheck:   now.Format(time.RFC3339),
		Description: "Main production API",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	targets[target2.ID] = target2

	target3 := &Target{
		ID:          uuid.New().String(),
		Name:        "Database Server",
		Type:        "server",
		Address:     "db.example.com",
		Port:        5432,
		Enabled:     true,
		Status:      "up",
		LastCheck:   now.Format(time.RFC3339),
		Description: "PostgreSQL Database",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	targets[target3.ID] = target3
}

func main() {
	// Initialize mock data
	initMockData()

	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api/v1").Subrouter()

	// Probes
	api.HandleFunc("/probes", getProbes).Methods("GET")
	api.HandleFunc("/probes", createProbe).Methods("POST")
	api.HandleFunc("/probes/{id}", getProbe).Methods("GET")

	// Targets
	api.HandleFunc("/targets", getTargets).Methods("GET")
	api.HandleFunc("/targets", createTarget).Methods("POST")
	api.HandleFunc("/targets/{id}", getTarget).Methods("GET")
	api.HandleFunc("/targets/{id}/stats", getTargetStats).Methods("GET")
	api.HandleFunc("/targets/{id}/measurements", getTargetMeasurements).Methods("GET")

	// Monitors
	api.HandleFunc("/monitors", getMonitors).Methods("GET")
	api.HandleFunc("/monitors", createMonitor).Methods("POST")
	api.HandleFunc("/monitors/{id}", getMonitor).Methods("GET")

	// Health
	api.HandleFunc("/health", healthCheck).Methods("GET")

	// Tools
	api.HandleFunc("/tools/ping", pingTool).Methods("POST")
	api.HandleFunc("/tools/doh-check", dohCheck).Methods("POST")
	api.HandleFunc("/tools/ip-info", getIPInfo).Methods("GET")
	api.HandleFunc("/tools/speed-test/download", downloadSpeedTest).Methods("GET")
	api.HandleFunc("/tools/speed-test/upload", uploadSpeedTest).Methods("POST")
	api.HandleFunc("/tools/jitter-test", jitterTest).Methods("POST")
	api.HandleFunc("/tools/traceroute", tracerouteTool).Methods("POST")

	// Professional network tools - TODO: Implement these handlers
	// api.HandleFunc("/tools/whois", whoisLookup).Methods("POST")
	// api.HandleFunc("/tools/port-scan", portScanner).Methods("POST")
	// api.HandleFunc("/tools/ssl-check", sslChecker).Methods("POST")
	// api.HandleFunc("/tools/mtr", mtrTool).Methods("POST")
	// api.HandleFunc("/tools/asn-lookup", asnLookup).Methods("POST")
	// api.HandleFunc("/tools/geoip", geoipLookup).Methods("POST")
	// api.HandleFunc("/tools/dns-propagation", dnsPropagation).Methods("POST")
	// api.HandleFunc("/tools/bgp-route", bgpRoute).Methods("POST")
	// api.HandleFunc("/tools/packet-loss", packetLossTest).Methods("POST")
	// api.HandleFunc("/tools/bandwidth", bandwidthTest).Methods("POST")

	// WebSocket route
	r.HandleFunc("/ws", handleWebSocket)

	// Root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"service": "NetMon API",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	// Apply CORS middleware
	handler := corsMiddleware(r)

	// Start server
	addr := ":8080"
	log.Printf("🚀 NetMon Backend Server starting on %s", addr)
	log.Printf("📊 API Endpoint: http://localhost:8080/api/v1")
	log.Printf("🔌 WebSocket: ws://localhost:8080/ws")
	log.Printf("✅ CORS: Enabled for all origins")
	log.Printf("✅ Probes: %d initialized", len(probes))
	log.Printf("✅ Targets: %d initialized", len(targets))

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}
