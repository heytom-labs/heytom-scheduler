package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"task-scheduler/internal/domain"
	"task-scheduler/pkg/logger"
	"task-scheduler/pkg/metrics"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server represents the HTTP server
type Server struct {
	port       int
	handler    *Handler
	httpServer *http.Server
}

// NewServer creates a new HTTP server
func NewServer(port int, taskService domain.TaskService, scheduleService domain.ScheduleService) *Server {
	handler := NewHandler(taskService)
	return &Server{
		port:    port,
		handler: handler,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(s.traceIDMiddleware())
	router.Use(s.metricsMiddleware())
	router.Use(s.loggingMiddleware())
	router.Use(s.corsMiddleware())

	// Setup API routes
	s.handler.SetupRoutes(router)

	// Add metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Serve static files for web UI
	s.setupStaticFiles(router)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", s.port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting HTTP server", zap.String("address", addr))

	// Start server
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	logger.Info("Stopping HTTP server")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to stop HTTP server: %w", err)
	}

	return nil
}

// traceIDMiddleware adds trace ID to requests
func (s *Server) traceIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if trace ID is provided in header
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			// Generate new trace ID
			traceID = uuid.New().String()
		}

		// Check if request ID is provided in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = uuid.New().String()
		}

		// Add to context
		ctx := logger.WithTraceID(c.Request.Context(), traceID)
		ctx = logger.WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Add to response headers
		c.Header("X-Trace-ID", traceID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// metricsMiddleware records HTTP request metrics
func (s *Server) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := c.Writer.Status()
		method := c.Request.Method

		metrics.RecordHTTPRequest(method, path, fmt.Sprintf("%d", statusCode))
		metrics.RecordHTTPRequestDuration(method, path, duration)
	}
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Use context-aware logging
		logger.InfoContext(c.Request.Context(), "HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// setupStaticFiles configures static file serving for the web UI
func (s *Server) setupStaticFiles(router *gin.Engine) {
	// Serve static files from web-dist directory
	router.Static("/assets", "./web-dist/assets")

	// Serve index.html for root and SPA routes
	router.NoRoute(func(c *gin.Context) {
		// Check if the request is for an API endpoint
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Check if the request is for metrics or health
		if c.Request.URL.Path == "/metrics" || c.Request.URL.Path == "/health" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Endpoint not found"})
			return
		}

		// Serve index.html for all other routes (SPA)
		c.File("./web-dist/index.html")
	})

	logger.Info("Static file serving configured", zap.String("path", "./web-dist"))
}
