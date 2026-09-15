package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/netmon/netmon/internal/models"
	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DB() *sql.DB {
	return s.db
}

// Probe operations
func (s *Store) CreateProbe(ctx context.Context, probe *models.Probe) error {
	metadata, err := json.Marshal(probe.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO probes (id, name, location, description, api_key, status, version, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRowContext(ctx, query,
		probe.ID, probe.Name, probe.Location, probe.Description,
		probe.APIKey, probe.Status, probe.Version, metadata,
	).Scan(&probe.CreatedAt, &probe.UpdatedAt)
}

func (s *Store) GetProbe(ctx context.Context, id uuid.UUID) (*models.Probe, error) {
	query := `
		SELECT id, name, location, description, api_key, status, version,
		       last_heartbeat, created_at, updated_at, metadata
		FROM probes WHERE id = $1
	`

	probe := &models.Probe{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&probe.ID, &probe.Name, &probe.Location, &probe.Description,
		&probe.APIKey, &probe.Status, &probe.Version,
		&probe.LastHeartbeat, &probe.CreatedAt, &probe.UpdatedAt, &metadata,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadata, &probe.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return probe, nil
}

func (s *Store) GetProbeByAPIKey(ctx context.Context, apiKey string) (*models.Probe, error) {
	query := `
		SELECT id, name, location, description, api_key, status, version,
		       last_heartbeat, created_at, updated_at, metadata
		FROM probes WHERE api_key = $1
	`

	probe := &models.Probe{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, query, apiKey).Scan(
		&probe.ID, &probe.Name, &probe.Location, &probe.Description,
		&probe.APIKey, &probe.Status, &probe.Version,
		&probe.LastHeartbeat, &probe.CreatedAt, &probe.UpdatedAt, &metadata,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadata, &probe.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return probe, nil
}

func (s *Store) ListProbes(ctx context.Context) ([]*models.Probe, error) {
	query := `
		SELECT id, name, location, description, api_key, status, version,
		       last_heartbeat, created_at, updated_at, metadata
		FROM probes ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	probes := make([]*models.Probe, 0)
	for rows.Next() {
		probe := &models.Probe{}
		var metadata []byte

		err := rows.Scan(
			&probe.ID, &probe.Name, &probe.Location, &probe.Description,
			&probe.APIKey, &probe.Status, &probe.Version,
			&probe.LastHeartbeat, &probe.CreatedAt, &probe.UpdatedAt, &metadata,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &probe.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		probes = append(probes, probe)
	}

	return probes, rows.Err()
}

func (s *Store) UpdateProbeHeartbeat(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE probes
		SET last_heartbeat = NOW(), status = $2, updated_at = NOW()
		WHERE id = $1
	`
	_, err := s.db.ExecContext(ctx, query, id, models.ProbeStatusActive)
	return err
}

// Target operations
func (s *Store) CreateTarget(ctx context.Context, target *models.Target) error {
	metadata, err := json.Marshal(target.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO targets (id, name, target_type, address, port, enabled, description, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRowContext(ctx, query,
		target.ID, target.Name, target.TargetType, target.Address,
		target.Port, target.Enabled, target.Description, metadata,
	).Scan(&target.CreatedAt, &target.UpdatedAt)
}

func (s *Store) GetTarget(ctx context.Context, id uuid.UUID) (*models.Target, error) {
	query := `
		SELECT id, name, target_type, address, port, enabled, description,
		       created_at, updated_at, metadata
		FROM targets WHERE id = $1
	`

	target := &models.Target{}
	var metadata []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&target.ID, &target.Name, &target.TargetType, &target.Address,
		&target.Port, &target.Enabled, &target.Description,
		&target.CreatedAt, &target.UpdatedAt, &metadata,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(metadata, &target.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	return target, nil
}

func (s *Store) ListTargets(ctx context.Context, enabledOnly bool) ([]*models.Target, error) {
	query := `
		SELECT id, name, target_type, address, port, enabled, description,
		       created_at, updated_at, metadata
		FROM targets
	`
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY name"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	targets := make([]*models.Target, 0)
	for rows.Next() {
		target := &models.Target{}
		var metadata []byte

		err := rows.Scan(
			&target.ID, &target.Name, &target.TargetType, &target.Address,
			&target.Port, &target.Enabled, &target.Description,
			&target.CreatedAt, &target.UpdatedAt, &metadata,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &target.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		targets = append(targets, target)
	}

	return targets, rows.Err()
}

// Monitor operations
func (s *Store) CreateMonitor(ctx context.Context, monitor *models.Monitor) error {
	config, err := json.Marshal(monitor.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	query := `
		INSERT INTO monitors (id, target_id, probe_id, monitor_type, enabled,
		                      interval_seconds, timeout_seconds, retries, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, updated_at
	`

	return s.db.QueryRowContext(ctx, query,
		monitor.ID, monitor.TargetID, monitor.ProbeID, monitor.MonitorType,
		monitor.Enabled, monitor.IntervalSeconds, monitor.TimeoutSeconds,
		monitor.Retries, config,
	).Scan(&monitor.CreatedAt, &monitor.UpdatedAt)
}

func (s *Store) GetMonitor(ctx context.Context, id uuid.UUID) (*models.Monitor, error) {
	query := `
		SELECT id, target_id, probe_id, monitor_type, enabled,
		       interval_seconds, timeout_seconds, retries, config,
		       created_at, updated_at
		FROM monitors WHERE id = $1
	`

	monitor := &models.Monitor{}
	var config []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&monitor.ID, &monitor.TargetID, &monitor.ProbeID, &monitor.MonitorType,
		&monitor.Enabled, &monitor.IntervalSeconds, &monitor.TimeoutSeconds,
		&monitor.Retries, &config, &monitor.CreatedAt, &monitor.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(config, &monitor.Config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return monitor, nil
}

func (s *Store) ListMonitors(ctx context.Context, enabledOnly bool) ([]*models.Monitor, error) {
	query := `
		SELECT id, target_id, probe_id, monitor_type, enabled,
		       interval_seconds, timeout_seconds, retries, config,
		       created_at, updated_at
		FROM monitors
	`
	if enabledOnly {
		query += " WHERE enabled = true"
	}
	query += " ORDER BY created_at"

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	monitors := make([]*models.Monitor, 0)
	for rows.Next() {
		monitor := &models.Monitor{}
		var config []byte

		err := rows.Scan(
			&monitor.ID, &monitor.TargetID, &monitor.ProbeID, &monitor.MonitorType,
			&monitor.Enabled, &monitor.IntervalSeconds, &monitor.TimeoutSeconds,
			&monitor.Retries, &config, &monitor.CreatedAt, &monitor.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(config, &monitor.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}

		monitors = append(monitors, monitor)
	}

	return monitors, rows.Err()
}

// Measurement operations
func (s *Store) InsertMeasurement(ctx context.Context, measurement *models.Measurement) error {
	metadata, err := json.Marshal(measurement.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO measurements (time, probe_id, target_id, monitor_id, monitor_type,
		                          success, latency_ms, packet_loss, jitter_ms,
		                          error_message, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = s.db.ExecContext(ctx, query,
		measurement.Time, measurement.ProbeID, measurement.TargetID,
		measurement.MonitorID, measurement.MonitorType, measurement.Success,
		measurement.LatencyMs, measurement.PacketLoss, measurement.JitterMs,
		measurement.ErrorMessage, metadata,
	)

	return err
}

func (s *Store) GetRecentMeasurements(ctx context.Context, targetID uuid.UUID, limit int) ([]*models.Measurement, error) {
	query := `
		SELECT time, probe_id, target_id, monitor_id, monitor_type,
		       success, latency_ms, packet_loss, jitter_ms,
		       error_message, metadata
		FROM measurements
		WHERE target_id = $1
		ORDER BY time DESC
		LIMIT $2
	`

	rows, err := s.db.QueryContext(ctx, query, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var measurements []*models.Measurement
	for rows.Next() {
		m := &models.Measurement{}
		var metadata []byte

		err := rows.Scan(
			&m.Time, &m.ProbeID, &m.TargetID, &m.MonitorID, &m.MonitorType,
			&m.Success, &m.LatencyMs, &m.PacketLoss, &m.JitterMs,
			&m.ErrorMessage, &metadata,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(metadata, &m.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}

		measurements = append(measurements, m)
	}

	return measurements, rows.Err()
}

// GetTargetStats returns availability and latency statistics for a target
func (s *Store) GetTargetStats(ctx context.Context, targetID uuid.UUID, since time.Time) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_checks,
			SUM(CASE WHEN success THEN 1 ELSE 0 END)::FLOAT / COUNT(*) * 100 as availability,
			AVG(latency_ms) as avg_latency,
			MIN(latency_ms) as min_latency,
			MAX(latency_ms) as max_latency,
			AVG(packet_loss) as avg_packet_loss
		FROM measurements
		WHERE target_id = $1 AND time >= $2
	`

	var totalChecks int
	var availability, avgLatency, minLatency, maxLatency, avgPacketLoss sql.NullFloat64

	err := s.db.QueryRowContext(ctx, query, targetID, since).Scan(
		&totalChecks, &availability, &avgLatency, &minLatency, &maxLatency, &avgPacketLoss,
	)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total_checks": totalChecks,
	}

	if availability.Valid {
		stats["availability"] = availability.Float64
	}
	if avgLatency.Valid {
		stats["avg_latency_ms"] = avgLatency.Float64
	}
	if minLatency.Valid {
		stats["min_latency_ms"] = minLatency.Float64
	}
	if maxLatency.Valid {
		stats["max_latency_ms"] = maxLatency.Float64
	}
	if avgPacketLoss.Valid {
		stats["avg_packet_loss"] = avgPacketLoss.Float64
	}

	return stats, nil
}
