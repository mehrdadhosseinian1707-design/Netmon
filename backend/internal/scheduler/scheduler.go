package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/netmon/netmon/internal/models"
	"github.com/netmon/netmon/internal/monitors"
	"github.com/netmon/netmon/internal/store"
	"go.uber.org/zap"
)

type Scheduler struct {
	store          *store.Store
	monitorFactory *monitors.MonitorFactory
	logger         *zap.Logger
	jobs           map[uuid.UUID]*Job
	mu             sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

type Job struct {
	Monitor    *models.Monitor
	Target     *models.Target
	Ticker     *time.Ticker
	CancelFunc context.CancelFunc
}

func NewScheduler(store *store.Store, monitorFactory *monitors.MonitorFactory, logger *zap.Logger) *Scheduler {
	return &Scheduler{
		store:          store,
		monitorFactory: monitorFactory,
		logger:         logger,
		jobs:           make(map[uuid.UUID]*Job),
		stopChan:       make(chan struct{}),
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.logger.Info("starting scheduler")

	// Load all enabled monitors
	if err := s.loadMonitors(ctx); err != nil {
		return fmt.Errorf("failed to load monitors: %w", err)
	}

	// Start refresh loop to pick up new monitors
	s.wg.Add(1)
	go s.refreshLoop(ctx)

	return nil
}

func (s *Scheduler) Stop() {
	s.logger.Info("stopping scheduler")
	close(s.stopChan)

	// Stop all jobs
	s.mu.Lock()
	for _, job := range s.jobs {
		job.Ticker.Stop()
		job.CancelFunc()
	}
	s.mu.Unlock()

	s.wg.Wait()
	s.logger.Info("scheduler stopped")
}

func (s *Scheduler) loadMonitors(ctx context.Context) error {
	monitors, err := s.store.ListMonitors(ctx, true)
	if err != nil {
		return err
	}

	for _, mon := range monitors {
		if err := s.scheduleMonitor(ctx, mon); err != nil {
			s.logger.Error("failed to schedule monitor",
				zap.String("monitor_id", mon.ID.String()),
				zap.Error(err))
		}
	}

	s.logger.Info("loaded monitors", zap.Int("count", len(monitors)))
	return nil
}

func (s *Scheduler) scheduleMonitor(ctx context.Context, monitor *models.Monitor) error {
	// Get the target
	target, err := s.store.GetTarget(ctx, monitor.TargetID)
	if err != nil {
		return fmt.Errorf("failed to get target: %w", err)
	}

	if !target.Enabled {
		return nil
	}

	// Check if already scheduled
	s.mu.RLock()
	_, exists := s.jobs[monitor.ID]
	s.mu.RUnlock()

	if exists {
		return nil
	}

	// Create monitor instance
	mon, err := s.monitorFactory.Create(monitor.MonitorType, monitor.Config)
	if err != nil {
		return fmt.Errorf("failed to create monitor: %w", err)
	}

	// Create ticker
	ticker := time.NewTicker(time.Duration(monitor.IntervalSeconds) * time.Second)

	// Create context for this job
	jobCtx, cancel := context.WithCancel(ctx)

	job := &Job{
		Monitor:    monitor,
		Target:     target,
		Ticker:     ticker,
		CancelFunc: cancel,
	}

	s.mu.Lock()
	s.jobs[monitor.ID] = job
	s.mu.Unlock()

	// Start job goroutine
	s.wg.Add(1)
	go s.runJob(jobCtx, job, mon)

	s.logger.Info("scheduled monitor",
		zap.String("monitor_id", monitor.ID.String()),
		zap.String("target", target.Address),
		zap.String("type", monitor.MonitorType),
		zap.Int("interval", monitor.IntervalSeconds))

	return nil
}

func (s *Scheduler) runJob(ctx context.Context, job *Job, monitor monitors.Monitor) {
	defer s.wg.Done()

	// Run immediately on start
	s.executeCheck(ctx, job, monitor)

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-job.Ticker.C:
			s.executeCheck(ctx, job, monitor)
		}
	}
}

func (s *Scheduler) executeCheck(ctx context.Context, job *Job, monitor monitors.Monitor) {
	start := time.Now()

	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(job.Monitor.TimeoutSeconds)*time.Second)
	defer cancel()

	// Execute check with retries
	var measurement *models.Measurement
	var err error

	for attempt := 0; attempt <= job.Monitor.Retries; attempt++ {
		measurement, err = monitor.Check(checkCtx, job.Target)
		if err == nil && measurement.Success {
			break
		}

		if attempt < job.Monitor.Retries {
			time.Sleep(1 * time.Second)
		}
	}

	if err != nil {
		s.logger.Error("check failed",
			zap.String("monitor_id", job.Monitor.ID.String()),
			zap.String("target", job.Target.Address),
			zap.Error(err))
		return
	}

	// Fill in IDs
	measurement.ProbeID = uuid.Nil // Will be set by probe
	measurement.TargetID = job.Target.ID
	measurement.MonitorID = job.Monitor.ID

	// Store measurement
	if err := s.store.InsertMeasurement(ctx, measurement); err != nil {
		s.logger.Error("failed to store measurement",
			zap.String("monitor_id", job.Monitor.ID.String()),
			zap.Error(err))
		return
	}

	duration := time.Since(start)

	s.logger.Debug("check completed",
		zap.String("monitor_id", job.Monitor.ID.String()),
		zap.String("target", job.Target.Address),
		zap.Bool("success", measurement.Success),
		zap.Duration("duration", duration))
}

func (s *Scheduler) refreshLoop(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			if err := s.loadMonitors(ctx); err != nil {
				s.logger.Error("failed to refresh monitors", zap.Error(err))
			}
		}
	}
}

func (s *Scheduler) GetJobCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.jobs)
}
