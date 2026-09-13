package models

import (
	"time"

	"github.com/google/uuid"
)

type Probe struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name"`
	Location      string                 `json:"location" db:"location"`
	Description   string                 `json:"description" db:"description"`
	APIKey        string                 `json:"api_key,omitempty" db:"api_key"`
	Status        string                 `json:"status" db:"status"`
	Version       string                 `json:"version" db:"version"`
	LastHeartbeat *time.Time             `json:"last_heartbeat" db:"last_heartbeat"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	Metadata      map[string]interface{} `json:"metadata" db:"metadata"`
}

type Target struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	TargetType  string                 `json:"target_type" db:"target_type"`
	Address     string                 `json:"address" db:"address"`
	Port        *int                   `json:"port,omitempty" db:"port"`
	Enabled     bool                   `json:"enabled" db:"enabled"`
	Description string                 `json:"description" db:"description"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
}

type Monitor struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TargetID        uuid.UUID              `json:"target_id" db:"target_id"`
	ProbeID         *uuid.UUID             `json:"probe_id,omitempty" db:"probe_id"`
	MonitorType     string                 `json:"monitor_type" db:"monitor_type"`
	Enabled         bool                   `json:"enabled" db:"enabled"`
	IntervalSeconds int                    `json:"interval_seconds" db:"interval_seconds"`
	TimeoutSeconds  int                    `json:"timeout_seconds" db:"timeout_seconds"`
	Retries         int                    `json:"retries" db:"retries"`
	Config          map[string]interface{} `json:"config" db:"config"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

type Measurement struct {
	Time         time.Time              `json:"time" db:"time"`
	ProbeID      uuid.UUID              `json:"probe_id" db:"probe_id"`
	TargetID     uuid.UUID              `json:"target_id" db:"target_id"`
	MonitorID    uuid.UUID              `json:"monitor_id" db:"monitor_id"`
	MonitorType  string                 `json:"monitor_type" db:"monitor_type"`
	Success      bool                   `json:"success" db:"success"`

	// Basic network metrics
	LatencyMs    *float64               `json:"latency_ms,omitempty" db:"latency_ms"`
	PacketLoss   *float64               `json:"packet_loss,omitempty" db:"packet_loss"`
	JitterMs     *float64               `json:"jitter_ms,omitempty" db:"jitter_ms"`

	// Speed and throughput metrics (Mbps)
	DownloadSpeedMbps *float64           `json:"download_speed_mbps,omitempty" db:"download_speed_mbps"`
	UploadSpeedMbps   *float64           `json:"upload_speed_mbps,omitempty" db:"upload_speed_mbps"`
	TCPThroughputMbps *float64           `json:"tcp_throughput_mbps,omitempty" db:"tcp_throughput_mbps"`
	UDPThroughputMbps *float64           `json:"udp_throughput_mbps,omitempty" db:"udp_throughput_mbps"`

	// DNS metrics
	DNSLatencyMs      *float64           `json:"dns_latency_ms,omitempty" db:"dns_latency_ms"`

	// TCP metrics
	TCPConnectTimeMs  *float64           `json:"tcp_connect_time_ms,omitempty" db:"tcp_connect_time_ms"`
	TLSHandshakeTimeMs *float64          `json:"tls_handshake_time_ms,omitempty" db:"tls_handshake_time_ms"`
	TCPRetransmissions *int              `json:"tcp_retransmissions,omitempty" db:"tcp_retransmissions"`

	// HTTP metrics
	HTTPTTFBMs        *float64           `json:"http_ttfb_ms,omitempty" db:"http_ttfb_ms"`
	HTTPTotalTimeMs   *float64           `json:"http_total_time_ms,omitempty" db:"http_total_time_ms"`

	// Path metrics
	MTU               *int               `json:"mtu,omitempty" db:"mtu"`
	HopCount          *int               `json:"hop_count,omitempty" db:"hop_count"`
	RouteChanges      *int               `json:"route_changes,omitempty" db:"route_changes"`
	ASPath            *string            `json:"as_path,omitempty" db:"as_path"`

	// Interface metrics
	RXErrors          *int64             `json:"rx_errors,omitempty" db:"rx_errors"`
	TXErrors          *int64             `json:"tx_errors,omitempty" db:"tx_errors"`
	RXDrops           *int64             `json:"rx_drops,omitempty" db:"rx_drops"`
	TXDrops           *int64             `json:"tx_drops,omitempty" db:"tx_drops"`

	// Connectivity
	IPv4Connectivity  *bool              `json:"ipv4_connectivity,omitempty" db:"ipv4_connectivity"`
	IPv6Connectivity  *bool              `json:"ipv6_connectivity,omitempty" db:"ipv6_connectivity"`

	ErrorMessage string                 `json:"error_message,omitempty" db:"error_message"`
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"`
}

type Alert struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TargetID        uuid.UUID              `json:"target_id" db:"target_id"`
	MonitorID       *uuid.UUID             `json:"monitor_id,omitempty" db:"monitor_id"`
	Severity        string                 `json:"severity" db:"severity"`
	Title           string                 `json:"title" db:"title"`
	Message         string                 `json:"message" db:"message"`
	Status          string                 `json:"status" db:"status"`
	FirstOccurrence time.Time              `json:"first_occurrence" db:"first_occurrence"`
	LastOccurrence  time.Time              `json:"last_occurrence" db:"last_occurrence"`
	OccurrenceCount int                    `json:"occurrence_count" db:"occurrence_count"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

type Incident struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	Title           string                 `json:"title" db:"title"`
	Description     string                 `json:"description" db:"description"`
	Severity        string                 `json:"severity" db:"severity"`
	Status          string                 `json:"status" db:"status"`
	StartedAt       time.Time              `json:"started_at" db:"started_at"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty" db:"resolved_at"`
	AffectedTargets int                    `json:"affected_targets" db:"affected_targets"`
	AffectedProbes  int                    `json:"affected_probes" db:"affected_probes"`
	RootCause       string                 `json:"root_cause" db:"root_cause"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

const (
	ProbeStatusActive   = "active"
	ProbeStatusInactive = "inactive"
	ProbeStatusError    = "error"

	TargetTypeHost   = "host"
	TargetTypeRouter = "router"
	TargetTypeSwitch = "switch"
	TargetTypeServer = "server"

	MonitorTypeICMP         = "icmp"
	MonitorTypeTCP          = "tcp"
	MonitorTypeUDP          = "udp"
	MonitorTypeDNS          = "dns"
	MonitorTypeHTTP         = "http"
	MonitorTypeHTTPS        = "https"
	MonitorTypeTraceroute   = "traceroute"
	MonitorTypeSNMP         = "snmp"
	MonitorTypeBandwidth    = "bandwidth"
	MonitorTypeConnectivity = "connectivity"
	MonitorTypeInterface    = "interface"

	AlertSeverityInfo     = "info"
	AlertSeverityWarning  = "warning"
	AlertSeverityCritical = "critical"

	AlertStatusActive   = "active"
	AlertStatusResolved = "resolved"
)
