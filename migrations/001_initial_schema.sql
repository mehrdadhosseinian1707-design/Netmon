-- Migration: 001_initial_schema
-- Create core tables for network monitoring platform

-- Probes table
CREATE TABLE IF NOT EXISTS probes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    location VARCHAR(255),
    description TEXT,
    api_key VARCHAR(255) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    version VARCHAR(50),
    last_heartbeat TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_probes_status ON probes(status);
CREATE INDEX idx_probes_last_heartbeat ON probes(last_heartbeat);

-- Targets table
CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    port INTEGER,
    enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_targets_enabled ON targets(enabled);
CREATE INDEX idx_targets_type ON targets(target_type);
CREATE INDEX idx_targets_address ON targets(address);

-- Monitor configurations table
CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    probe_id UUID REFERENCES probes(id) ON DELETE SET NULL,
    monitor_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    retries INTEGER NOT NULL DEFAULT 3,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitors_target ON monitors(target_id);
CREATE INDEX idx_monitors_probe ON monitors(probe_id);
CREATE INDEX idx_monitors_type ON monitors(monitor_type);
CREATE INDEX idx_monitors_enabled ON monitors(enabled);

-- Measurements table (time-series data)
CREATE TABLE IF NOT EXISTS measurements (
    time TIMESTAMPTZ NOT NULL,
    probe_id UUID NOT NULL,
    target_id UUID NOT NULL,
    monitor_id UUID NOT NULL,
    monitor_type VARCHAR(50) NOT NULL,
    success BOOLEAN NOT NULL,
    latency_ms DOUBLE PRECISION,
    packet_loss DOUBLE PRECISION,
    jitter_ms DOUBLE PRECISION,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Convert measurements to hypertable
SELECT create_hypertable('measurements', 'time', if_not_exists => TRUE);

-- Create indexes on measurements
CREATE INDEX idx_measurements_probe ON measurements(probe_id, time DESC);
CREATE INDEX idx_measurements_target ON measurements(target_id, time DESC);
CREATE INDEX idx_measurements_monitor ON measurements(monitor_id, time DESC);
CREATE INDEX idx_measurements_type ON measurements(monitor_type, time DESC);
CREATE INDEX idx_measurements_success ON measurements(success, time DESC);

-- Alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    monitor_id UUID REFERENCES monitors(id) ON DELETE CASCADE,
    severity VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    first_occurrence TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_occurrence TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    occurrence_count INTEGER NOT NULL DEFAULT 1,
    resolved_at TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_alerts_target ON alerts(target_id);
CREATE INDEX idx_alerts_monitor ON alerts(monitor_id);
CREATE INDEX idx_alerts_status ON alerts(status);
CREATE INDEX idx_alerts_severity ON alerts(severity);
CREATE INDEX idx_alerts_first_occurrence ON alerts(first_occurrence DESC);

-- Incidents table (correlated alerts)
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    affected_targets INTEGER DEFAULT 0,
    affected_probes INTEGER DEFAULT 0,
    root_cause TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_incidents_status ON incidents(status);
CREATE INDEX idx_incidents_severity ON incidents(severity);
CREATE INDEX idx_incidents_started ON incidents(started_at DESC);

-- Alert-Incident mapping
CREATE TABLE IF NOT EXISTS incident_alerts (
    incident_id UUID NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    alert_id UUID NOT NULL REFERENCES alerts(id) ON DELETE CASCADE,
    added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (incident_id, alert_id)
);

-- Traceroute results
CREATE TABLE IF NOT EXISTS traceroutes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    probe_id UUID NOT NULL REFERENCES probes(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    success BOOLEAN NOT NULL,
    total_hops INTEGER,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_traceroutes_probe ON traceroutes(probe_id, timestamp DESC);
CREATE INDEX idx_traceroutes_target ON traceroutes(target_id, timestamp DESC);

-- Traceroute hops
CREATE TABLE IF NOT EXISTS traceroute_hops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    traceroute_id UUID NOT NULL REFERENCES traceroutes(id) ON DELETE CASCADE,
    hop_number INTEGER NOT NULL,
    ip_address INET,
    hostname VARCHAR(255),
    rtt_ms DOUBLE PRECISION,
    timeout BOOLEAN NOT NULL DEFAULT false,
    asn INTEGER,
    organization VARCHAR(255),
    country VARCHAR(2)
);

CREATE INDEX idx_traceroute_hops_traceroute ON traceroute_hops(traceroute_id, hop_number);

-- ASN information
CREATE TABLE IF NOT EXISTS asns (
    asn INTEGER PRIMARY KEY,
    name VARCHAR(255),
    description TEXT,
    country VARCHAR(2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ISPs table
CREATE TABLE IF NOT EXISTS isps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    country VARCHAR(2),
    asn INTEGER REFERENCES asns(asn),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_isps_asn ON isps(asn);

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'viewer',
    enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);

-- Audit log
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action VARCHAR(255) NOT NULL,
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    previous_value JSONB,
    new_value JSONB,
    source_ip INET,
    result VARCHAR(50) NOT NULL
);

CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, timestamp DESC);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id, timestamp DESC);
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);

-- Create continuous aggregates for rollups (1-minute averages)
CREATE MATERIALIZED VIEW measurements_1min
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 minute', time) AS bucket,
    probe_id,
    target_id,
    monitor_id,
    monitor_type,
    COUNT(*) AS measurement_count,
    SUM(CASE WHEN success THEN 1 ELSE 0 END)::DOUBLE PRECISION / COUNT(*) AS availability,
    AVG(latency_ms) AS avg_latency_ms,
    MIN(latency_ms) AS min_latency_ms,
    MAX(latency_ms) AS max_latency_ms,
    AVG(packet_loss) AS avg_packet_loss,
    AVG(jitter_ms) AS avg_jitter_ms
FROM measurements
GROUP BY bucket, probe_id, target_id, monitor_id, monitor_type
WITH NO DATA;

-- Add refresh policy for continuous aggregate
SELECT add_continuous_aggregate_policy('measurements_1min',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- Enable compression on measurements hypertable
ALTER TABLE measurements SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'probe_id, target_id, monitor_id'
);

-- Create compression policy (compress data older than 7 days)
SELECT add_compression_policy('measurements', INTERVAL '7 days');

-- Create retention policy (keep raw data for 30 days)
SELECT add_retention_policy('measurements', INTERVAL '30 days');
