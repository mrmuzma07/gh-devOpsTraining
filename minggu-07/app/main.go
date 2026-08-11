package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// =============================================================
// 1. Setup OTel SDK + slog JSON logger
// =============================================================
var (
	tracer trace.Tracer
	logger *slog.Logger
)

func initLogger() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func initTracer(ctx context.Context) (func(context.Context) error, error) {
	// a. Resource: metadata service
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("go-app"),
			semconv.ServiceVersion("v1.0.0"),
			semconv.DeploymentEnvironment("dev"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// b. Exporter OTLP gRPC → Alloy
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		otlpEndpoint = "grafana-alloy-otel.mini-prod.svc.cluster.local:4317"
	}
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// c. TracerProvider: AlwaysOn sampler (kirim semua trace untuk dev)
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(100),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.AlwaysSample())),
	)
	otel.SetTracerProvider(tp)

	// d. Propagator W3C Trace Context (untuk antar-service)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp.Tracer("go-app")
	return tp.Shutdown, nil
}

// =============================================================
// 2. Database & Redis clients (stub untuk demo lokal)
// =============================================================

type DB struct {
	conn *sql.DB
}

type Cache struct {
	client *redis.Client
}

func NewDB() *DB {
	// Untuk demo, return stub DB agar tidak butuh Postgres berjalan
	// Di production nyata, gunakan:
	//   db, _ := sql.Open("postgres", "postgres://user:pass@postgres:5432/dbname?sslmode=disable")
	return &DB{}
}

func (d *DB) GetProduct(ctx context.Context, id string) (string, error) {
	// Simulasi query DB ~50ms
	time.Sleep(50 * time.Millisecond)
	return fmt.Sprintf("Product-%s", id), nil
}

func NewCache() *Cache {
	return &Cache{}
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	// Simulasi lookup Redis ~10ms
	time.Sleep(10 * time.Millisecond)
	return fmt.Sprintf("cached:%s", key), nil
}

var (
	db    = NewDB()
	cache = NewCache()
)

// =============================================================
// 3. Instrumented handlers (HTTP → DB → Redis)
// =============================================================

func handleOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, span := tracer.Start(ctx, "GET /order",
		trace.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", "/order"),
		),
	)
	defer span.End()

	productID := r.URL.Query().Get("product")
	if productID == "" {
		productID = "default-1"
	}

	logger.InfoContext(ctx, "handling order request",
		"product_id", productID,
		"remote", r.RemoteAddr,
	)

	// Step 1: Query Postgres (akan jadi child span)
	ctx, dbSpan := tracer.Start(ctx, "postgres.query",
		trace.WithAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
		),
	)
	product, err := db.GetProduct(ctx, productID)
	dbSpan.End()
	if err != nil {
		dbSpan.RecordError(err)
		logger.ErrorContext(ctx, "db query failed", "error", err.Error())
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	// Step 2: Lookup Redis (child span lain)
	ctx, cacheSpan := tracer.Start(ctx, "redis.lookup",
		trace.WithAttributes(
			attribute.String("db.system", "redis"),
			attribute.String("cache.key", productID),
		),
	)
	cached, _ := cache.Get(ctx, productID)
	cacheSpan.SetAttributes(attribute.String("cache.value", cached))
	cacheSpan.End()

	// Final response
	response := fmt.Sprintf("Order for %s — %s", product, cached)
	span.SetAttributes(
		attribute.String("order.product", product),
		attribute.Int("http.status_code", http.StatusOK),
	)
	fmt.Fprint(w, response)

	logger.InfoContext(ctx, "order handled successfully",
		"product", product,
		"trace_id", span.SpanContext().TraceID().String(),
	)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(ctx := r.Context(), "GET /health")
	defer span.End()
	fmt.Fprint(w, "OK")
	span.SetAttributes(attribute.Int("http.status_code", 200))
}

// =============================================================
// 4. Main: bootstrap + graceful shutdown
// =============================================================

func main() {
	initLogger()
	ctx := context.Background()

	shutdown, err := initTracer(ctx)
	if err != nil {
		logger.Error("failed to init tracer", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdown(shutdownCtx); err != nil {
			logger.Error("tracer shutdown failed", "error", err.Error())
		}
	}()

	http.HandleFunc("/order", handleOrder)
	http.HandleFunc("/health", handleHealth)

	port := ":8080"
	logger.Info("server starting with OTel enabled",
		"port", port,
		"otel_endpoint", os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
	)
	if err := http.ListenAndServe(port, nil); err != nil {
		logger.Error("server failed", "error", err.Error())
		os.Exit(1)
	}
}