// API types matching backend models
export interface Probe {
  id: string;
  name: string;
  location: string;
  description: string;
  status: 'active' | 'inactive' | 'error';
  version: string;
  last_heartbeat: string | null;
  created_at: string;
  updated_at: string;
  metadata: Record<string, any>;
}

export interface Target {
  id: string;
  name: string;
  type: string;
  address: string;
  port?: number;
  enabled: boolean;
  status: 'up' | 'down' | 'degraded' | 'unknown';
  last_check?: string;
  description: string;
  created_at: string;
  updated_at: string;
  metadata?: Record<string, any>;
}

export interface Monitor {
  id: string;
  target_id: string;
  probe_id?: string;
  monitor_type: 'icmp' | 'tcp' | 'udp' | 'dns' | 'http' | 'https' | 'traceroute' | 'snmp';
  enabled: boolean;
  interval_seconds: number;
  timeout_seconds: number;
  retries: number;
  config: Record<string, any>;
  created_at: string;
  updated_at: string;
}

export interface Measurement {
  time: string;
  probe_id: string;
  target_id: string;
  monitor_id: string;
  monitor_type: string;
  success: boolean;
  latency_ms?: number;
  packet_loss?: number;
  jitter_ms?: number;
  error_message?: string;
  metadata: Record<string, any>;
}

export interface TargetStats {
  total_checks: number;
  availability?: number;
  avg_latency_ms?: number;
  min_latency_ms?: number;
  max_latency_ms?: number;
  avg_packet_loss?: number;
}

export interface Alert {
  id: string;
  target_id: string;
  monitor_id?: string;
  severity: 'info' | 'warning' | 'critical';
  title: string;
  message: string;
  status: 'active' | 'resolved';
  first_occurrence: string;
  last_occurrence: string;
  occurrence_count: number;
  resolved_at?: string;
  metadata: Record<string, any>;
}

export interface Incident {
  id: string;
  title: string;
  description: string;
  severity: 'info' | 'warning' | 'critical';
  status: 'active' | 'resolved';
  started_at: string;
  resolved_at?: string;
  affected_targets: number;
  affected_probes: number;
  root_cause: string;
  metadata: Record<string, any>;
}

export interface DashboardStats {
  total_targets: number;
  targets_up: number;
  targets_down: number;
  targets_degraded: number;
  active_alerts: number;
  active_incidents: number;
  total_probes: number;
  active_probes: number;
}

export interface WebSocketMessage {
  type: 'measurement' | 'alert' | 'incident' | 'probe_status' | 'target_status';
  data: any;
  timestamp: string;
}
