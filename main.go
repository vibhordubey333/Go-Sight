package main

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	customRegistry.MustRegister(HttpRequestErrorTotal)
	customRegistry.MustRegister(HttpRequestTotal)
}

var (
	HttpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestErrorTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_error_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "path", "status"},
	)
)

var customRegistry = prometheus.NewRegistry()

func main() {
	router := gin.Default()
	router.GET("/metrics", PrometheusHandler())

	router.Use(RequestMetricsMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.GET("/v1/users", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Hello World"})
	})
	router.Run(":8000")
}

// Custom metrics with custom regsitry
func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.HandlerFor(customRegistry, promhttp.HandlerOpts{EnableOpenMetrics: true})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RequestMetricsMiddleware records incoming request metrics
func RequestMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//Before request
		c.Next()
		//After request
		status := strconv.Itoa(c.Writer.Status())
		intStatus, errAtoi := strconv.Atoi(status)
		if errAtoi != nil {
			fmt.Println("Error while Atoi op", errAtoi)
		}
		if intStatus < 400 {
			HttpRequestTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
		} else {
			HttpRequestErrorTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, status).Inc()
		}

	}
}
