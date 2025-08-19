package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

type MetricInfo struct {
	Name         string
	Help         string
	ExampleValue string
	Labels       map[string]string
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

	// Metrics info endpoint
	s.router.GET("/metrics-info", s.handleMetricsInfo)
}

func (s *Server) getMetricsInfo() []MetricInfo {
	var metricsInfo []MetricInfo

	// Define all metrics manually since reflection approach is complex with Prometheus metrics
	metrics := []struct {
		name  string
		field string
	}{
		{"filesystem_exporter_info", "VersionInfo"},
		{"filesystem_exporter_volume_size_bytes", "VolumeSize"},
		{"filesystem_exporter_volume_available_bytes", "VolumeAvailable"},
		{"filesystem_exporter_volume_used_ratio", "VolumeUsedRatio"},
		{"filesystem_exporter_directory_size_bytes", "DirectorySize"},
		{"filesystem_exporter_collection_duration_seconds", "CollectionDuration"},
		{"filesystem_exporter_collection_timestamp", "CollectionTimestamp"},
		{"filesystem_exporter_collection_interval_seconds", "CollectionInterval"},
		{"filesystem_exporter_collection_success_total", "CollectionSuccess"},
		{"filesystem_exporter_collection_failed_total", "CollectionFailed"},
		{"filesystem_exporter_directories_processed_total", "DirectoriesProcessed"},
		{"filesystem_exporter_directories_failed_total", "DirectoriesFailed"},
		{"filesystem_exporter_du_lock_wait_duration_seconds", "DuLockWaitDuration"},
	}

	for _, metric := range metrics {
		metricsInfo = append(metricsInfo, MetricInfo{
			Name:         metric.name,
			Help:         s.getMetricHelp(metric.field),
			ExampleValue: s.getExampleValue(metric.field),
			Labels:       s.getExampleLabels(metric.field),
		})
	}

	return metricsInfo
}

func (s *Server) getExampleLabels(metricName string) map[string]string {
	switch metricName {
	case "VersionInfo":
		return map[string]string{"version": "v1.7.0", "commit": "abc123", "build_date": "2024-01-01"}
	case "VolumeSize", "VolumeAvailable", "VolumeUsedRatio":
		return map[string]string{"volume": "root", "mount_point": "/", "device": "/dev/sda1"}
	case "DirectorySize":
		return map[string]string{"group": "home", "directory": "/home/hoose", "mode": "recursive", "subdirectory_level": "0"}
	case "CollectionDuration", "CollectionTimestamp", "CollectionSuccess", "CollectionFailed":
		return map[string]string{"type": "filesystem", "group": "default", "interval_seconds": "30"}
	case "CollectionInterval":
		return map[string]string{"type": "filesystem", "group": "default"}
	case "DirectoriesProcessed", "DirectoriesFailed":
		return map[string]string{"group": "home", "mode": "recursive"}
	case "DuLockWaitDuration":
		return map[string]string{"group": "home", "directory": "/home/hoose"}
	default:
		return map[string]string{}
	}
}

func (s *Server) getExampleValue(metricName string) string {
	switch metricName {
	case "VersionInfo":
		return "1"
	case "VolumeSize":
		return "107374182400"
	case "VolumeAvailable":
		return "53687091200"
	case "VolumeUsedRatio":
		return "0.5"
	case "DirectorySize":
		return "2147483648"
	case "CollectionDuration":
		return "2.5"
	case "CollectionTimestamp":
		return "1704067200"
	case "CollectionInterval":
		return "30"
	case "CollectionSuccess":
		return "42"
	case "CollectionFailed":
		return "3"
	case "DirectoriesProcessed":
		return "15"
	case "DirectoriesFailed":
		return "1"
	case "DuLockWaitDuration":
		return "0.125"
	default:
		return "0"
	}
}

func (s *Server) getMetricHelp(metricName string) string {
	switch metricName {
	case "VersionInfo":
		return "Information about the filesystem exporter"
	case "VolumeSize":
		return "Total size of volume in bytes"
	case "VolumeAvailable":
		return "Available space on volume in bytes"
	case "VolumeUsedRatio":
		return "Ratio of used space on volume (0.0 to 1.0)"
	case "DirectorySize":
		return "Size of directory in bytes"
	case "CollectionDuration":
		return "Duration of collection in seconds"
	case "CollectionTimestamp":
		return "Timestamp of last collection"
	case "CollectionInterval":
		return "Configured collection interval in seconds"
	case "CollectionSuccess":
		return "Total number of successful collections"
	case "CollectionFailed":
		return "Total number of failed collections"
	case "DirectoriesProcessed":
		return "Total number of directories processed"
	case "DirectoriesFailed":
		return "Total number of directories that failed to process"
	case "DuLockWaitDuration":
		return "Time spent waiting for du mutex lock in seconds"
	default:
		return "Filesystem exporter metric"
	}
}

func (s *Server) handleMetricsInfo(c *gin.Context) {
	metricsInfo := s.getMetricsInfo()

	// Generate JSON response
	response := gin.H{
		"metrics":      metricsInfo,
		"total_count":  len(metricsInfo),
		"generated_at": time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) handleRoot(c *gin.Context) {
	versionInfo := version.Get()
	metricsInfo := s.getMetricsInfo()

	// Generate metrics HTML dynamically
	metricsHTML := ""

	for i, metric := range metricsInfo {
		labelsStr := ""

		if len(metric.Labels) > 0 {
			var labelPairs []string
			for k, v := range metric.Labels {
				labelPairs = append(labelPairs, fmt.Sprintf(`%s="%s"`, k, v))
			}

			labelsStr = "{" + strings.Join(labelPairs, ", ") + "}"
		}

		// Create clickable metric with hidden details
		metricsHTML += fmt.Sprintf(`
            <div class="metric-item" onclick="toggleMetricDetails(%d)">
                <div class="metric-header">
                    <span class="metric-name">%s</span>
                    <span class="metric-toggle">▼</span>
                </div>
                <div class="metric-details" id="metric-%d">
                    <div class="metric-help"><strong>Description:</strong> %s</div>
                    <div class="metric-example"><strong>Example:</strong> %s = %s</div>
                    <div class="metric-labels"><strong>Labels:</strong> %s</div>
                </div>
            </div>`,
			i,
			metric.Name,
			i,
			metric.Help,
			metric.Name,
			metric.ExampleValue,
			labelsStr)
	}

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
        h1 .version {
            font-size: 0.6em;
            color: #6c757d;
            font-weight: normal;
            margin-left: 0.5rem;
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
        .metrics-list {
            margin: 0.5rem 0;
        }
        .metric-item {
            border: 1px solid #dee2e6;
            border-radius: 6px;
            margin: 0.5rem 0;
            cursor: pointer;
            transition: all 0.2s ease;
        }
        .metric-item:hover {
            border-color: #007bff;
            background-color: #f8f9fa;
        }
        .metric-header {
            padding: 0.75rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-weight: 500;
            color: #495057;
        }
        .metric-name {
            font-family: 'Courier New', monospace;
            font-size: 0.9rem;
        }
        .metric-toggle {
            font-size: 0.8rem;
            color: #6c757d;
            transition: transform 0.2s ease;
        }
        .metric-details {
            display: none;
            padding: 0.75rem;
            border-top: 1px solid #dee2e6;
            background-color: #f8f9fa;
            font-size: 0.85rem;
            line-height: 1.4;
        }
        .metric-details.show {
            display: block;
        }
        .metric-help, .metric-example, .metric-labels {
            margin: 0.5rem 0;
        }
        .metric-example {
            font-family: 'Courier New', monospace;
            background-color: #e9ecef;
            padding: 0.25rem 0.5rem;
            border-radius: 3px;
        }
        .metric-labels {
            color: #6c757d;
        }
    </style>
    <script>
        function toggleMetricDetails(id) {
            const details = document.getElementById('metric-' + id);
            const toggle = details.previousElementSibling.querySelector('.metric-toggle');
            
            if (details.classList.contains('show')) {
                details.classList.remove('show');
                toggle.textContent = '▼';
            } else {
                details.classList.add('show');
                toggle.textContent = '▲';
            }
        }
    </script>
</head>
<body>
    <h1>Filesystem Exporter<span class="version">` + versionInfo.Version + `</span></h1>
    
    <div class="endpoint">
        <h3><a href="/metrics">📊 Metrics</a></h3>
        <p class="description">Prometheus metrics endpoint</p>
        <span class="status metrics">Available</span>
    </div>

    <div class="endpoint">
        <h3><a href="/metrics-info">📋 Metrics Info</a></h3>
        <p class="description">Detailed metrics information with examples</p>
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

    <div class="metrics-info">
        <h3>Available Metrics</h3>
        <div class="metrics-list">` + metricsHTML + `
        </div>
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
