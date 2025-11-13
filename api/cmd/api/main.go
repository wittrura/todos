package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"example.com/todos/pkg/db"
	"example.com/todos/pkg/handlers"

	"github.com/caarlos0/env/v11"
	"github.com/gorilla/mux"
)

func main() {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("cfg: %v\n", cfg)

	// Connect to the Postgres database
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.Db.User, cfg.Db.Pass, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)
	log.Printf("Connecting to database at url %s..\n", url)
	db := db.NewDB(context.Background(), url)
	handler := handlers.NewRouteHandler(db)
	router := setupRouter(handler)

	// Start the server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on port %d..\n", cfg.Port)
	log.Fatal(http.ListenAndServe(addr, router))
}

func setupRouter(h *handlers.RouteHandler) http.Handler {
	// Initialize the router
	r := mux.NewRouter()

	// Define API routes and their handlers
	r.HandleFunc("/", handlers.Healthy).Methods("GET")
	r.HandleFunc("/todos", h.GetTodos).Methods("GET")
	r.HandleFunc("/todos/{id}", h.GetTodo).Methods("GET")
	r.HandleFunc("/todos/{id}", h.UpdateTodo).Methods("PATCH")
	r.HandleFunc("/todos", h.CreateTodo).Methods("POST")
	r.HandleFunc("/todos/{id}", h.DeleteTodo).Methods("DELETE")
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
