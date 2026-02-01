package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/host"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
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
	traceShutdown, err := setupOTelTracing()
	if err != nil {
		log.Fatalf("failed to setup otel tracing: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := traceShutdown(ctx); err != nil {
			log.Printf("failed to shutdown otel tracing: %v", err)
		}
	}()

	router := gin.Default()
	router.Use(otelgin.Middleware("go-sight-api"))
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

func setupOTelTracing() (func(context.Context) error, error) {
	ctx := context.Background()

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("go-sight-api"),
		),
	)
	if err != nil {
		return nil, err
	}

	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	return provider.Shutdown, nil
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
