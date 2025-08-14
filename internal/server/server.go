package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"filesystem-exporter/internal/config"
	"filesystem-exporter/internal/metrics"
	"filesystem-exporter/internal/version"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	config  *config.Config
	metrics *metrics.Registry
	server  *http.Server
	router  *gin.Engine
}

// customGinLogger creates a custom Gin logger that uses slog
func customGinLogger() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Use slog to log the request
		slog.Info("HTTP request",
			"method", param.Method,
			"path", param.Path,
			"status", param.StatusCode,
			"latency", param.Latency,
			"client_ip", param.ClientIP,
			"user_agent", param.Request.UserAgent(),
		)

		return "" // Return empty string since slog handles the output
	})
}

func New(cfg *config.Config, registry *metrics.Registry) *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(customGinLogger(), gin.Recovery())

	srv := &Server{
		config:  cfg,
		metrics: registry,
		router:  router,
	}

	srv.setupRoutes()

	return srv
}

func (s *Server) setupRoutes() {
	// Root endpoint with HTML dashboard
	s.router.GET("/", s.handleRoot)

	// Metrics endpoint - use our custom registry
	s.router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(s.metrics.GetRegistry(), promhttp.HandlerOpts{})))

	// Health endpoint
	s.router.GET("/health", s.handleHealth)
}

func (s *Server) handleRoot(c *gin.Context) {
	versionInfo := version.Get()
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Filesystem Exporter ` + versionInfo.Version + `</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 800px;
            margin: 0 auto;
            padding: 2rem;
            line-height: 1.6;
            color: #333;
        }
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 0.5rem;
        }
        .endpoint {
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .endpoint h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .endpoint a {
            color: #007bff;
            text-decoration: none;
            font-weight: 500;
        }
        .endpoint a:hover {
            text-decoration: underline;
        }
        .description {
            color: #6c757d;
            font-size: 0.9rem;
        }
        .status {
            display: inline-block;
            padding: 0.25rem 0.5rem;
            border-radius: 4px;
            font-size: 0.8rem;
            font-weight: 500;
        }
        .status.healthy {
            background: #d4edda;
            color: #155724;
        }
        .status.metrics {
            background: #d1ecf1;
            color: #0c5460;
        }
        .status.ready {
            background: #d4edda;
            color: #155724;
        }
        .status.not-ready {
            background: #f8d7da;
            color: #721c24;
        }
        .status.fast {
            background: #fff3cd;
            color: #856404;
        }
        .status.slow {
            background: #e2e3e5;
            color: #383d41;
        }
        .service-status {
            background: #e9ecef;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .service-status h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .service-status p {
            margin: 0.25rem 0;
            color: #6c757d;
        }
        .metrics-info {
            background: #e9ecef;
            border: 1px solid #dee2e6;
            border-radius: 8px;
            padding: 1rem;
            margin: 1rem 0;
        }
        .metrics-info h3 {
            margin: 0 0 0.5rem 0;
            color: #495057;
        }
        .metrics-info ul {
            margin: 0.5rem 0;
            padding-left: 1.5rem;
        }
        .metrics-info li {
            margin: 0.25rem 0;
            color: #6c757d;
        }
        .footer {
            margin-top: 2rem;
            padding-top: 1rem;
            border-top: 1px solid #dee2e6;
            text-align: center;
            color: #6c757d;
            font-size: 0.9rem;
        }
        .footer a {
            color: #007bff;
            text-decoration: none;
        }
        .footer a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <h1>Filesystem Exporter</h1>
    
    <div class="endpoint">
        <h3><a href="/metrics">📊 Metrics</a></h3>
        <p class="description">Prometheus metrics endpoint</p>
        <span class="status metrics">Available</span>
    </div>

    <div class="endpoint">
        <h3><a href="/health">❤️ Health Check</a></h3>
        <p class="description">Service health status</p>
        <span class="status healthy">Healthy</span>
    </div>

    <div class="service-status">
        <h3>Service Status</h3>
        <p><strong>Status:</strong> <span class="status ready">Ready</span></p>
        <p><strong>Filesystem Metrics:</strong> <span class="status fast">Collection</span></p>
        <p><strong>Directory Metrics:</strong> <span class="status slow">Collection</span></p>
    </div>

    <div class="metrics-info">
        <h3>Version Information</h3>
        <ul>
            <li><strong>Version:</strong> ` + versionInfo.Version + `</li>
            <li><strong>Commit:</strong> ` + versionInfo.Commit + `</li>
            <li><strong>Build Date:</strong> ` + versionInfo.BuildDate + `</li>
        </ul>
    </div>

    <div class="metrics-info">
        <h3>Configuration</h3>
        <ul>
            <li><strong>Filesystems:</strong> ` + fmt.Sprintf("%d", len(s.config.Filesystems)) + ` configured</li>
            <li><strong>Directories:</strong> ` + fmt.Sprintf("%d", len(s.config.Directories)) + ` configured</li>
        </ul>
    </div>

    <div class="footer">
        <p>Copyright © 2024 Dougal Matthews. Licensed under <a href="https://opensource.org/licenses/MIT" target="_blank">MIT License</a>.</p>
        <p><a href="https://github.com/d0ugal/filesystem-exporter" target="_blank">GitHub Repository</a> | <a href="https://github.com/d0ugal/filesystem-exporter/issues" target="_blank">Report Issues</a></p>
    </div>
</body>
</html>`

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (s *Server) handleHealth(c *gin.Context) {
	versionInfo := version.Get()
	c.JSON(http.StatusOK, gin.H{
		"status":     "healthy",
		"timestamp":  time.Now().Unix(),
		"service":    "filesystem-exporter",
		"version":    versionInfo.Version,
		"commit":     versionInfo.Commit,
		"build_date": versionInfo.BuildDate,
	})
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.router,
		ReadHeaderTimeout: 30 * time.Second,
	}

	slog.Info("Starting server", "address", addr)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.server.Shutdown(ctx); err != nil {
			slog.Error("Server shutdown error", "error", err)
		} else {
			slog.Info("Server shutdown gracefully")
		}
	}
}
