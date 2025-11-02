package main

import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    _ "github.com/jackc/pgx/v5/stdlib"

    "github.com/yourname/searchsvc/internal/cache"
    "github.com/yourname/searchsvc/internal/handlers"
    "github.com/yourname/searchsvc/internal/provider"
    "github.com/yourname/searchsvc/internal/rate"
    "github.com/yourname/searchsvc/internal/services"
    "github.com/yourname/searchsvc/internal/storage"
)

func main() {
    dsn := mustEnv("DB_DSN")
    p1 := mustEnv("PROVIDER1_URL")
    p2 := mustEnv("PROVIDER2_URL")
    ttlSec, _ := strconv.Atoi(getEnv("CACHE_TTL_SECONDS", "120"))
    ratePerMin, _ := strconv.Atoi(getEnv("RATE_LIMIT_PER_MINUTE", "60"))
    migDir := getEnv("MIGRATIONS_DIR", "/app/migrations")
    webDir := getEnv("WEB_DIR", "/app/web")

    db, err := sql.Open("pgx", dsn)
    if err != nil { log.Fatalf("open db: %v", err) }
    defer db.Close()
    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(30 * time.Minute)

    if err := storage.RunMigrations(context.Background(), db, migDir); err != nil {
        log.Fatalf("migrations: %v", err)
    }

    limiter := rate.NewRegistry(ratePerMin)
    p1c := provider.NewJSONProvider("provider1", p1, limiter)
    p2c := provider.NewXMLProvider("provider2", p2, limiter)

    repo := storage.NewRepo(db)
    c := cache.NewTTLCache(time.Duration(ttlSec) * time.Second)
    svc := services.NewSearchService([]provider.Provider{p1c, p2c}, repo, c)

    r := chi.NewRouter()
    r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer)
    r.Use(middleware.Timeout(10 * time.Second))

    r.Get("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
    r.Mount("/search", handlers.NewSearchRouter(svc))

    // static files from disk
    fs := http.FileServer(http.Dir(webDir))
    r.Handle("/dashboard/*", http.StripPrefix("/dashboard", fs))

    port := getEnv("PORT", "8080")
    log.Printf("listening on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}

func getEnv(k, def string) string { if v := os.Getenv(k); v != "" { return v }; return def }
func mustEnv(k string) string { v := os.Getenv(k); if v == "" { log.Fatalf("%s is required", k) }; return v }
