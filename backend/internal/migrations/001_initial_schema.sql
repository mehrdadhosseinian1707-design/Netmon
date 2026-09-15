-- Create probes table
CREATE TABLE IF NOT EXISTS probes (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    location VARCHAR(255),
    description TEXT,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'inactive',
    version VARCHAR(50),
    last_heartbeat TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_probes_status ON probes(status);
CREATE INDEX IF NOT EXISTS idx_probes_api_key ON probes(api_key);

-- Create targets table
CREATE TABLE IF NOT EXISTS targets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL,
    address VARCHAR(255) NOT NULL,
    port INTEGER,
    enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_targets_enabled ON targets(enabled);
CREATE INDEX IF NOT EXISTS idx_targets_type ON targets(target_type);

-- Create monitors table
CREATE TABLE IF NOT EXISTS monitors (
    id UUID PRIMARY KEY,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    probe_id UUID REFERENCES probes(id) ON DELETE SET NULL,
    monitor_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    interval_seconds INTEGER NOT NULL DEFAULT 60,
    timeout_seconds INTEGER NOT NULL DEFAULT 10,
    retries INTEGER NOT NULL DEFAULT 3,
    config JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_monitors_target ON monitors(target_id);
CREATE INDEX IF NOT EXISTS idx_monitors_probe ON monitors(probe_id);
CREATE INDEX IF NOT EXISTS idx_monitors_enabled ON monitors(enabled);

-- Create measurements table (TimescaleDB hypertable)
CREATE TABLE IF NOT EXISTS measurements (
    time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    probe_id UUID,
    target_id UUID NOT NULL,
    monitor_id UUID NOT NULL,
    monitor_type VARCHAR(50) NOT NULL,
    success BOOLEAN NOT NULL,
    latency_ms FLOAT,
    packet_loss FLOAT,
    jitter_ms FLOAT,
    error_message TEXT,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Convert measurements to hypertable if TimescaleDB is available and not already a hypertable
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM timescaledb_information.hypertables
        WHERE hypertable_name = 'measurements'
    ) THEN
        PERFORM create_hypertable('measurements', 'time', if_not_exists => TRUE);
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        -- TimescaleDB might not be available, continue without it
        RAISE NOTICE 'Could not create hypertable, continuing without TimescaleDB: %', SQLERRM;
END $$;

-- Create indexes on measurements
CREATE INDEX IF NOT EXISTS idx_measurements_target ON measurements(target_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_measurements_monitor ON measurements(monitor_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_measurements_probe ON measurements(probe_id, time DESC);
CREATE INDEX IF NOT EXISTS idx_measurements_success ON measurements(success, time DESC);

-- Create incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    monitor_id UUID REFERENCES monitors(id) ON DELETE SET NULL,
    severity VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_incidents_target ON incidents(target_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);

-- Create alerts table
CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY,
    incident_id UUID REFERENCES incidents(id) ON DELETE CASCADE,
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    alert_type VARCHAR(50) NOT NULL,
    severity VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    notified BOOLEAN NOT NULL DEFAULT false,
    acknowledged BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_alerts_incident ON alerts(incident_id);
CREATE INDEX IF NOT EXISTS idx_alerts_target ON alerts(target_id);
CREATE INDEX IF NOT EXISTS idx_alerts_notified ON alerts(notified);
CREATE INDEX IF NOT EXISTS idx_alerts_created ON alerts(created_at DESC);
