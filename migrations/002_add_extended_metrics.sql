-- Migration: 002_add_extended_metrics
-- Add extended network monitoring metrics to measurements table

-- Add new metric columns to measurements table
ALTER TABLE measurements
ADD COLUMN IF NOT EXISTS download_speed_mbps DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS upload_speed_mbps DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS tcp_throughput_mbps DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS udp_throughput_mbps DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS dns_latency_ms DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS tcp_connect_time_ms DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS tls_handshake_time_ms DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS tcp_retransmissions INTEGER,
ADD COLUMN IF NOT EXISTS http_ttfb_ms DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS http_total_time_ms DOUBLE PRECISION,
ADD COLUMN IF NOT EXISTS mtu INTEGER,
ADD COLUMN IF NOT EXISTS hop_count INTEGER,
ADD COLUMN IF NOT EXISTS route_changes INTEGER,
ADD COLUMN IF NOT EXISTS as_path TEXT,
ADD COLUMN IF NOT EXISTS rx_errors BIGINT,
ADD COLUMN IF NOT EXISTS tx_errors BIGINT,
ADD COLUMN IF NOT EXISTS rx_drops BIGINT,
ADD COLUMN IF NOT EXISTS tx_drops BIGINT,
ADD COLUMN IF NOT EXISTS ipv4_connectivity BOOLEAN,
ADD COLUMN IF NOT EXISTS ipv6_connectivity BOOLEAN;

-- Create indexes for new metrics that will be commonly queried
CREATE INDEX IF NOT EXISTS idx_measurements_download_speed ON measurements(download_speed_mbps, time DESC) WHERE download_speed_mbps IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_upload_speed ON measurements(upload_speed_mbps, time DESC) WHERE upload_speed_mbps IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_tcp_throughput ON measurements(tcp_throughput_mbps, time DESC) WHERE tcp_throughput_mbps IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_udp_throughput ON measurements(udp_throughput_mbps, time DESC) WHERE udp_throughput_mbps IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_dns_latency ON measurements(dns_latency_ms, time DESC) WHERE dns_latency_ms IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_tcp_connect ON measurements(tcp_connect_time_ms, time DESC) WHERE tcp_connect_time_ms IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_tls_handshake ON measurements(tls_handshake_time_ms, time DESC) WHERE tls_handshake_time_ms IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_http_ttfb ON measurements(http_ttfb_ms, time DESC) WHERE http_ttfb_ms IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_mtu ON measurements(mtu, time DESC) WHERE mtu IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_hop_count ON measurements(hop_count, time DESC) WHERE hop_count IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_ipv4_connectivity ON measurements(ipv4_connectivity, time DESC) WHERE ipv4_connectivity IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_ipv6_connectivity ON measurements(ipv6_connectivity, time DESC) WHERE ipv6_connectivity IS NOT NULL;

-- Update continuous aggregate to include new metrics
DROP MATERIALIZED VIEW IF EXISTS measurements_1min;

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
    AVG(jitter_ms) AS avg_jitter_ms,
    AVG(download_speed_mbps) AS avg_download_speed_mbps,
    AVG(upload_speed_mbps) AS avg_upload_speed_mbps,
    AVG(tcp_throughput_mbps) AS avg_tcp_throughput_mbps,
    AVG(udp_throughput_mbps) AS avg_udp_throughput_mbps,
    AVG(dns_latency_ms) AS avg_dns_latency_ms,
    AVG(tcp_connect_time_ms) AS avg_tcp_connect_time_ms,
    AVG(tls_handshake_time_ms) AS avg_tls_handshake_time_ms,
    AVG(tcp_retransmissions) AS avg_tcp_retransmissions,
    AVG(http_ttfb_ms) AS avg_http_ttfb_ms,
    AVG(http_total_time_ms) AS avg_http_total_time_ms,
    AVG(mtu) AS avg_mtu,
    AVG(hop_count) AS avg_hop_count,
    AVG(route_changes) AS avg_route_changes,
    SUM(rx_errors) AS total_rx_errors,
    SUM(tx_errors) AS total_tx_errors,
    SUM(rx_drops) AS total_rx_drops,
    SUM(tx_drops) AS total_tx_drops,
    SUM(CASE WHEN ipv4_connectivity THEN 1 ELSE 0 END)::DOUBLE PRECISION / NULLIF(COUNT(ipv4_connectivity), 0) AS ipv4_availability,
    SUM(CASE WHEN ipv6_connectivity THEN 1 ELSE 0 END)::DOUBLE PRECISION / NULLIF(COUNT(ipv6_connectivity), 0) AS ipv6_availability
FROM measurements
GROUP BY bucket, probe_id, target_id, monitor_id, monitor_type
WITH NO DATA;

-- Re-add refresh policy for continuous aggregate
SELECT add_continuous_aggregate_policy('measurements_1min',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- Create hourly aggregate view
CREATE MATERIALIZED VIEW measurements_1hour
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', bucket) AS bucket,
    probe_id,
    target_id,
    monitor_id,
    monitor_type,
    SUM(measurement_count) AS measurement_count,
    AVG(availability) AS availability,
    AVG(avg_latency_ms) AS avg_latency_ms,
    MIN(min_latency_ms) AS min_latency_ms,
    MAX(max_latency_ms) AS max_latency_ms,
    AVG(avg_packet_loss) AS avg_packet_loss,
    AVG(avg_jitter_ms) AS avg_jitter_ms,
    AVG(avg_download_speed_mbps) AS avg_download_speed_mbps,
    AVG(avg_upload_speed_mbps) AS avg_upload_speed_mbps,
    AVG(avg_tcp_throughput_mbps) AS avg_tcp_throughput_mbps,
    AVG(avg_udp_throughput_mbps) AS avg_udp_throughput_mbps,
    AVG(avg_dns_latency_ms) AS avg_dns_latency_ms,
    AVG(avg_tcp_connect_time_ms) AS avg_tcp_connect_time_ms,
    AVG(avg_tls_handshake_time_ms) AS avg_tls_handshake_time_ms,
    AVG(avg_tcp_retransmissions) AS avg_tcp_retransmissions,
    AVG(avg_http_ttfb_ms) AS avg_http_ttfb_ms,
    AVG(avg_http_total_time_ms) AS avg_http_total_time_ms,
    AVG(avg_mtu) AS avg_mtu,
    AVG(avg_hop_count) AS avg_hop_count,
    AVG(avg_route_changes) AS avg_route_changes,
    SUM(total_rx_errors) AS total_rx_errors,
    SUM(total_tx_errors) AS total_tx_errors,
    SUM(total_rx_drops) AS total_rx_drops,
    SUM(total_tx_drops) AS total_tx_drops,
    AVG(ipv4_availability) AS ipv4_availability,
    AVG(ipv6_availability) AS ipv6_availability
FROM measurements_1min
GROUP BY bucket, probe_id, target_id, monitor_id, monitor_type
WITH NO DATA;

-- Add refresh policy for hourly aggregate
SELECT add_continuous_aggregate_policy('measurements_1hour',
    start_offset => INTERVAL '7 days',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour');

-- Create daily aggregate view
CREATE MATERIALIZED VIEW measurements_1day
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 day', bucket) AS bucket,
    probe_id,
    target_id,
    monitor_id,
    monitor_type,
    SUM(measurement_count) AS measurement_count,
    AVG(availability) AS availability,
    AVG(avg_latency_ms) AS avg_latency_ms,
    MIN(min_latency_ms) AS min_latency_ms,
    MAX(max_latency_ms) AS max_latency_ms,
    AVG(avg_packet_loss) AS avg_packet_loss,
    AVG(avg_jitter_ms) AS avg_jitter_ms,
    AVG(avg_download_speed_mbps) AS avg_download_speed_mbps,
    AVG(avg_upload_speed_mbps) AS avg_upload_speed_mbps,
    AVG(avg_tcp_throughput_mbps) AS avg_tcp_throughput_mbps,
    AVG(avg_udp_throughput_mbps) AS avg_udp_throughput_mbps,
    AVG(avg_dns_latency_ms) AS avg_dns_latency_ms,
    AVG(avg_tcp_connect_time_ms) AS avg_tcp_connect_time_ms,
    AVG(avg_tls_handshake_time_ms) AS avg_tls_handshake_time_ms,
    AVG(avg_tcp_retransmissions) AS avg_tcp_retransmissions,
    AVG(avg_http_ttfb_ms) AS avg_http_ttfb_ms,
    AVG(avg_http_total_time_ms) AS avg_http_total_time_ms,
    AVG(avg_mtu) AS avg_mtu,
    AVG(avg_hop_count) AS avg_hop_count,
    AVG(avg_route_changes) AS avg_route_changes,
    SUM(total_rx_errors) AS total_rx_errors,
    SUM(total_tx_errors) AS total_tx_errors,
    SUM(total_rx_drops) AS total_rx_drops,
    SUM(total_tx_drops) AS total_tx_drops,
    AVG(ipv4_availability) AS ipv4_availability,
    AVG(ipv6_availability) AS ipv6_availability
FROM measurements_1hour
GROUP BY bucket, probe_id, target_id, monitor_id, monitor_type
WITH NO DATA;

-- Add refresh policy for daily aggregate
SELECT add_continuous_aggregate_policy('measurements_1day',
    start_offset => INTERVAL '30 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day');

-- Create view for route history tracking
CREATE TABLE IF NOT EXISTS route_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    probe_id UUID NOT NULL REFERENCES probes(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    hop_count INTEGER,
    as_path TEXT,
    route_hash VARCHAR(64), -- MD5 hash of the route for change detection
    hops JSONB,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_route_history_target ON route_history(target_id, timestamp DESC);
CREATE INDEX idx_route_history_probe ON route_history(probe_id, timestamp DESC);
CREATE INDEX idx_route_history_hash ON route_history(route_hash, timestamp DESC);

-- Create view for bandwidth statistics
CREATE TABLE IF NOT EXISTS bandwidth_tests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    target_id UUID NOT NULL REFERENCES targets(id) ON DELETE CASCADE,
    probe_id UUID NOT NULL REFERENCES probes(id) ON DELETE CASCADE,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    download_speed_mbps DOUBLE PRECISION,
    upload_speed_mbps DOUBLE PRECISION,
    download_bytes BIGINT,
    upload_bytes BIGINT,
    test_duration_ms INTEGER,
    success BOOLEAN NOT NULL,
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX idx_bandwidth_tests_target ON bandwidth_tests(target_id, timestamp DESC);
CREATE INDEX idx_bandwidth_tests_probe ON bandwidth_tests(probe_id, timestamp DESC);

-- Add comments for documentation
COMMENT ON COLUMN measurements.download_speed_mbps IS 'Download speed in Megabits per second';
COMMENT ON COLUMN measurements.upload_speed_mbps IS 'Upload speed in Megabits per second';
COMMENT ON COLUMN measurements.tcp_throughput_mbps IS 'TCP throughput in Megabits per second';
COMMENT ON COLUMN measurements.udp_throughput_mbps IS 'UDP throughput in Megabits per second';
COMMENT ON COLUMN measurements.dns_latency_ms IS 'DNS resolution latency in milliseconds';
COMMENT ON COLUMN measurements.tcp_connect_time_ms IS 'TCP connection establishment time in milliseconds';
COMMENT ON COLUMN measurements.tls_handshake_time_ms IS 'TLS handshake time in milliseconds';
COMMENT ON COLUMN measurements.tcp_retransmissions IS 'Number of TCP packet retransmissions';
COMMENT ON COLUMN measurements.http_ttfb_ms IS 'HTTP Time To First Byte in milliseconds';
COMMENT ON COLUMN measurements.http_total_time_ms IS 'HTTP total request time in milliseconds';
COMMENT ON COLUMN measurements.mtu IS 'Maximum Transmission Unit in bytes';
COMMENT ON COLUMN measurements.hop_count IS 'Number of network hops to destination';
COMMENT ON COLUMN measurements.route_changes IS 'Number of route changes detected since last measurement';
COMMENT ON COLUMN measurements.as_path IS 'Autonomous System path to destination';
COMMENT ON COLUMN measurements.rx_errors IS 'Network interface receive errors';
COMMENT ON COLUMN measurements.tx_errors IS 'Network interface transmit errors';
COMMENT ON COLUMN measurements.rx_drops IS 'Network interface receive packet drops';
COMMENT ON COLUMN measurements.tx_drops IS 'Network interface transmit packet drops';
COMMENT ON COLUMN measurements.ipv4_connectivity IS 'IPv4 connectivity status';
COMMENT ON COLUMN measurements.ipv6_connectivity IS 'IPv6 connectivity status';
