package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/netmon/netmon/internal/config"
	"github.com/netmon/netmon/internal/models"
	"github.com/netmon/netmon/internal/monitors"
	"github.com/netmon/netmon/internal/store"
	"go.uber.org/zap"
)

type ProbeClient struct {
	config         *config.ProbeConfig
	serverURL      string
	apiKey         string
	logger         *zap.Logger
	store          *store.Store
	monitorFactory *monitors.MonitorFactory
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("starting netmon probe",
		zap.String("name", cfg.Probe.Name),
		zap.String("location", cfg.Probe.Location))

	// Initialize database connection
	db, err := store.NewStore(cfg.Database.DSN())
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize monitor factory
	monitorFactory := monitors.NewMonitorFactory()

	// Create probe client
	client := &ProbeClient{
		config:         &cfg.Probe,
		serverURL:      cfg.Probe.ServerURL,
		apiKey:         cfg.Probe.APIKey,
		logger:         logger,
		store:          db,
		monitorFactory: monitorFactory,
	}

	// Start probe
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.run(ctx)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("shutting down probe")
	cancel()

	logger.Info("probe stopped")
}

func (p *ProbeClient) run(ctx context.Context) {
	// Get probe ID
	probeID, err := uuid.Parse(p.config.ID)
	if err != nil {
		p.logger.Fatal("Invalid probe ID", zap.Error(err))
	}

	// Send heartbeat periodically
	heartbeatTicker := time.NewTicker(p.config.HeartbeatInterval)
	defer heartbeatTicker.Stop()

	// Monitor execution loop
	checkTicker := time.NewTicker(30 * time.Second)
	defer checkTicker.Stop()

	p.logger.Info("probe running", zap.String("probe_id", probeID.String()))

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			if err := p.sendHeartbeat(ctx, probeID); err != nil {
				p.logger.Error("failed to send heartbeat", zap.Error(err))
			}
		case <-checkTicker.C:
			if err := p.executeMonitors(ctx, probeID); err != nil {
				p.logger.Error("failed to execute monitors", zap.Error(err))
			}
		}
	}
}

func (p *ProbeClient) sendHeartbeat(ctx context.Context, probeID uuid.UUID) error {
	if p.serverURL == "" {
		// Local mode - update database directly
		return p.store.UpdateProbeHeartbeat(ctx, probeID)
	}

	// Remote mode - send HTTP request
	url := fmt.Sprintf("%s/api/v1/probes/%s/heartbeat", p.serverURL, probeID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed with status: %d", resp.StatusCode)
	}

	p.logger.Debug("heartbeat sent")
	return nil
}

func (p *ProbeClient) executeMonitors(ctx context.Context, probeID uuid.UUID) error {
	// Get monitors assigned to this probe
	monitors, err := p.store.ListMonitors(ctx, true)
	if err != nil {
		return err
	}

	// Filter monitors for this probe or unassigned monitors
	var relevantMonitors []*models.Monitor
	for _, mon := range monitors {
		if mon.ProbeID == nil || *mon.ProbeID == probeID {
			relevantMonitors = append(relevantMonitors, mon)
		}
	}

	p.logger.Debug("executing monitors", zap.Int("count", len(relevantMonitors)))

	// Execute each monitor
	for _, mon := range relevantMonitors {
		if err := p.executeMonitor(ctx, probeID, mon); err != nil {
			p.logger.Error("monitor execution failed",
				zap.String("monitor_id", mon.ID.String()),
				zap.Error(err))
		}
	}

	return nil
}

func (p *ProbeClient) executeMonitor(ctx context.Context, probeID uuid.UUID, monitor *models.Monitor) error {
	// Get target
	target, err := p.store.GetTarget(ctx, monitor.TargetID)
	if err != nil {
		return err
	}

	// Create monitor instance
	mon, err := p.monitorFactory.Create(monitor.MonitorType, monitor.Config)
	if err != nil {
		return err
	}

	// Execute check with timeout
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(monitor.TimeoutSeconds)*time.Second)
	defer cancel()

	measurement, err := mon.Check(checkCtx, target)
	if err != nil {
		return err
	}

	// Fill in probe and IDs
	measurement.ProbeID = probeID
	measurement.TargetID = target.ID
	measurement.MonitorID = monitor.ID
	measurement.Time = time.Now()

	// Store measurement
	if p.serverURL == "" {
		// Local mode - store directly
		return p.store.InsertMeasurement(ctx, measurement)
	}

	// Remote mode - send to server
	return p.sendMeasurement(ctx, measurement)
}

func (p *ProbeClient) sendMeasurement(ctx context.Context, measurement *models.Measurement) error {
	url := fmt.Sprintf("%s/api/v1/measurements", p.serverURL)

	data, err := json.Marshal(measurement)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send measurement, status: %d", resp.StatusCode)
	}

	return nil
}
