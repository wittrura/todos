package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/todos/pkg/db"
	"example.com/todos/pkg/handlers"
	"example.com/todos/pkg/logging"
	"example.com/todos/pkg/middleware"

	"github.com/caarlos0/env/v11"
	"github.com/gorilla/mux"
)

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
	}

	if err := run(context.Background(), cfg, 3*time.Second); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func run(ctx context.Context, cfg Config, shutdownTimeout time.Duration) error {
	// Connect to the Postgres database
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.Db.User, cfg.Db.Pass, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)
	log.Printf("Connecting to database at url %s..\n", url)
	db := db.NewDB(context.Background(), url)
	defer func() {
		log.Println("Closing DB connection...")
		_ = db.Close()
	}()

	handler := handlers.NewRouteHandler(db)

	router := setupRouter(handler)
	server := createServer(cfg, router)

	serverErr := make(chan error, 1)

	// start server in a separate go routine which communicates via channel
	go func() {
		log.Printf("Starting server on %s\n", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// wait for a message from one of the channels
	select {
	case err := <-serverErr:
		return err
	case <-sigs:
		log.Println("Shutdown requested...")
	case <-ctx.Done():
		log.Println("Context cancelled...")
	}

	log.Println("Closing DB connection...")
	db.Close()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	log.Println("Shutdown complete...")
	return nil
}

func createServer(cfg Config, handler http.Handler) *http.Server {
	addr := fmt.Sprintf(":%d", cfg.Port)

	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}

func setupRouter(h *handlers.RouteHandler) http.Handler {
	// Initialize the router
	r := mux.NewRouter()

	logger := logging.NewLogger(os.Stdout)
	r.Use(
		func(next http.Handler) http.Handler {
			return middleware.MetricsMiddleware(func(r *http.Request) {
				handlers.HttpRequestCounter.WithLabelValues(r.URL.Path, r.Method).Inc()
			}, next)
		},
		func(next http.Handler) http.Handler {
			return middleware.LoggingMiddleware(logger, next)
		},
		func(next http.Handler) http.Handler {
			return middleware.RequestIDMiddleware(next)
		},
	)

	// Define API routes and their handlers
	r.Handle("/metrics", handlers.NewMetricsHandler())
	r.HandleFunc("/", handlers.Healthy).Methods("GET")
	r.HandleFunc("/todos", h.GetTodos).Methods("GET")
	r.HandleFunc("/todos/{id}", h.GetTodo).Methods("GET")
	r.HandleFunc("/todos/{id}", h.UpdateTodo).Methods("PATCH")
	r.HandleFunc("/todos", h.CreateTodo).Methods("POST")
	r.HandleFunc("/todos/{id}", h.DeleteTodo).Methods("DELETE")
	r.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Slow request started...")
		time.Sleep(8 * time.Second)
		fmt.Fprintf(w, "Slow request completed at %v\n", time.Now())
	})
	return r
}

type Config struct {
	Port int `env:"PORT" envDefault:"8080"`
	Db   DB  `envPrefix:"DB_"`
}

type DB struct {
	Host string `env:"HOST"`
	Port string `env:"PORT"`
	User string `env:"USER"`
	Pass string `env:"PASS"`
	Name string `env:"NAME"`
}
