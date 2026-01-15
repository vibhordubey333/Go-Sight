package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
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
	if err := setupOTelMetrics(customRegistry); err != nil {
		log.Fatalf("failed to setup otel metrics: %v", err)
	}

	router := gin.Default()
	router.GET("/metrics", PrometheusHandler())

	router.Use(RequestMetricsMiddleware())
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/v1/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})

	router.GET("/v1/compute", computeHandler)
	router.Run(":8000")
}

func setupOTelMetrics(reg *prometheus.Registry) error {
	exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return err
	}

	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(5 * time.Second)); err != nil {
		return err
	}

	if err := host.Start(host.WithMeterProvider(provider)); err != nil {
		return err
	}

	return nil
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

func computeHandler(c *gin.Context) {
	n, err := strconv.Atoi(c.DefaultQuery("n", "10000"))
	if err != nil || n < 1 {
		n = 10000
	}
	if n > 200000 {
		n = 200000
	}

	var sum int64
	for i := 1; i <= n; i++ {
		sum += int64(i * i)
	}

	c.JSON(http.StatusOK, gin.H{
		"n":           n,
		"sum_squares": sum,
	})
}
