package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/netmon/netmon/internal/models"
	"github.com/netmon/netmon/internal/store"
	"go.uber.org/zap"
)

type API struct {
	store  *store.Store
	logger *zap.Logger
	router *gin.Engine
}

func NewAPI(store *store.Store, logger *zap.Logger) *API {
	api := &API{
		store:  store,
		logger: logger,
		router: gin.New(),
	}

	api.setupRoutes()
	return api
}

func (a *API) setupRoutes() {
	// Middleware
	a.router.Use(gin.Recovery())
	a.router.Use(a.corsMiddleware())
	a.router.Use(a.loggingMiddleware())

	// Health check
	a.router.GET("/health", a.healthCheck)

	// API v1
	v1 := a.router.Group("/api/v1")
	{
		// Probes
		probes := v1.Group("/probes")
		{
			probes.GET("", a.listProbes)
			probes.GET("/:id", a.getProbe)
			probes.POST("", a.createProbe)
			probes.POST("/:id/heartbeat", a.probeHeartbeat)
		}

		// Targets
		targets := v1.Group("/targets")
		{
			targets.GET("", a.listTargets)
			targets.GET("/:id", a.getTarget)
			targets.POST("", a.createTarget)
			targets.GET("/:id/stats", a.getTargetStats)
			targets.GET("/:id/measurements", a.getTargetMeasurements)
		}

		// Monitors
		monitors := v1.Group("/monitors")
		{
			monitors.GET("", a.listMonitors)
			monitors.GET("/:id", a.getMonitor)
			monitors.POST("", a.createMonitor)
		}

		// Measurements
		measurements := v1.Group("/measurements")
		{
			measurements.POST("", a.createMeasurement)
		}
	}
}

func (a *API) Router() *gin.Engine {
	return a.router
}

// Middleware
func (a *API) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		a.logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
		)
	}
}

func (a *API) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// Health check
func (a *API) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}

// Probe handlers
func (a *API) listProbes(c *gin.Context) {
	probes, err := a.store.ListProbes(c.Request.Context())
	if err != nil {
		a.logger.Error("failed to list probes", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, probes)
}

func (a *API) getProbe(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe ID"})
		return
	}

	probe, err := a.store.GetProbe(c.Request.Context(), id)
	if err != nil {
		a.logger.Error("failed to get probe", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "probe not found"})
		return
	}

	c.JSON(http.StatusOK, probe)
}

func (a *API) createProbe(c *gin.Context) {
	var probe models.Probe
	if err := c.ShouldBindJSON(&probe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	probe.ID = uuid.New()
	probe.Status = models.ProbeStatusInactive
	probe.APIKey = uuid.New().String()

	if err := a.store.CreateProbe(c.Request.Context(), &probe); err != nil {
		a.logger.Error("failed to create probe", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create probe"})
		return
	}

	c.JSON(http.StatusCreated, probe)
}

func (a *API) probeHeartbeat(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid probe ID"})
		return
	}

	if err := a.store.UpdateProbeHeartbeat(c.Request.Context(), id); err != nil {
		a.logger.Error("failed to update heartbeat", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update heartbeat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Target handlers
func (a *API) listTargets(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	targets, err := a.store.ListTargets(c.Request.Context(), enabledOnly)
	if err != nil {
		a.logger.Error("failed to list targets", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, targets)
}

func (a *API) getTarget(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target ID"})
		return
	}

	target, err := a.store.GetTarget(c.Request.Context(), id)
	if err != nil {
		a.logger.Error("failed to get target", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "target not found"})
		return
	}

	c.JSON(http.StatusOK, target)
}

func (a *API) createTarget(c *gin.Context) {
	var target models.Target
	if err := c.ShouldBindJSON(&target); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	target.ID = uuid.New()
	target.Enabled = true

	if err := a.store.CreateTarget(c.Request.Context(), &target); err != nil {
		a.logger.Error("failed to create target", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create target"})
		return
	}

	c.JSON(http.StatusCreated, target)
}

func (a *API) getTargetStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target ID"})
		return
	}

	// Default to last 24 hours
	since := time.Now().Add(-24 * time.Hour)

	stats, err := a.store.GetTargetStats(c.Request.Context(), id, since)
	if err != nil {
		a.logger.Error("failed to get target stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (a *API) getTargetMeasurements(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid target ID"})
		return
	}

	measurements, err := a.store.GetRecentMeasurements(c.Request.Context(), id, 100)
	if err != nil {
		a.logger.Error("failed to get measurements", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get measurements"})
		return
	}

	c.JSON(http.StatusOK, measurements)
}

// Monitor handlers
func (a *API) listMonitors(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	monitors, err := a.store.ListMonitors(c.Request.Context(), enabledOnly)
	if err != nil {
		a.logger.Error("failed to list monitors", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, monitors)
}

func (a *API) getMonitor(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitor ID"})
		return
	}

	monitor, err := a.store.GetMonitor(c.Request.Context(), id)
	if err != nil {
		a.logger.Error("failed to get monitor", zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": "monitor not found"})
		return
	}

	c.JSON(http.StatusOK, monitor)
}

func (a *API) createMonitor(c *gin.Context) {
	var monitor models.Monitor
	if err := c.ShouldBindJSON(&monitor); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	monitor.ID = uuid.New()
	monitor.Enabled = true

	if err := a.store.CreateMonitor(c.Request.Context(), &monitor); err != nil {
		a.logger.Error("failed to create monitor", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create monitor"})
		return
	}

	c.JSON(http.StatusCreated, monitor)
}

// Measurement handlers
func (a *API) createMeasurement(c *gin.Context) {
	var measurement models.Measurement
	if err := c.ShouldBindJSON(&measurement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	measurement.Time = time.Now()

	if err := a.store.InsertMeasurement(c.Request.Context(), &measurement); err != nil {
		a.logger.Error("failed to insert measurement", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert measurement"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}
