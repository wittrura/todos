package handlers

import (
	"encoding/json"
	"net/http"

	"example.com/todos/pkg/models"

	"github.com/gorilla/mux"
)

type Database interface {
	Create(todo models.Todo) (id string, err error)
	Get(id string) (todo models.Todo, err error)
	GetAll() (todos []models.Todo, err error)
	Update(id string, todo models.Todo) (count int64, err error)
	Delete(id string) (count int64, err error)
}

type RouteHandler struct {
	db Database
}

func NewRouteHandler(db Database) *RouteHandler {
	return &RouteHandler{
		db: db,
	}
}

// Handlers for the API endpoints
func Healthy(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// •	POST /todos {title:string} → 201 with {id,title,done:false}
// •	GET /todos → list
// •	PATCH /todos/:id {done:bool} → 200
// •	DELETE /todos/:id → 204
func (h *RouteHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	todos, _ := h.db.GetAll()
	json.NewEncoder(w).Encode(todos)
}

func (h *RouteHandler) GetTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	todo, err := h.db.Get(params["id"])
	if err != nil {
		http.NotFound(w, r)
	}
	json.NewEncoder(w).Encode(todo)
}

func (h *RouteHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	var todo models.Todo
	_ = json.NewDecoder(r.Body).Decode(&todo)
	_, err := h.db.Update(params["id"], todo)
	if err != nil {
		http.NotFound(w, r)
	}
}

func (h *RouteHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var todo models.Todo
	_ = json.NewDecoder(r.Body).Decode(&todo)

	id, _ := h.db.Create(todo)
	todo.Id = id

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

func (h *RouteHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	_, err := h.db.Delete(params["id"])
	if err != nil {
		http.NotFound(w, r)
	}

	w.WriteHeader(http.StatusNoContent)
}
