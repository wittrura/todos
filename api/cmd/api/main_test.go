package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"testing"
	"time"

	"example.com/todos/pkg/handlers"
	"example.com/todos/pkg/models"
)

func TestHandler(t *testing.T) {
	handler := setupRouter(handlers.NewRouteHandler(newInMemoryDB()))

	// health
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// create Todo
	todo1 := models.Todo{
		Id:    "1",
		Title: "something new",
	}

	todo1Bytes, err := json.Marshal(todo1)
	if err != nil {
		panic("ahhh")
	}

	rr = httptest.NewRecorder()
	create := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(todo1Bytes))
	handler.ServeHTTP(rr, create)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// get Todo
	rr = httptest.NewRecorder()
	get := httptest.NewRequest(http.MethodGet, "/todos/1", nil)
	handler.ServeHTTP(rr, get)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var newTodo models.Todo
	err = json.NewDecoder(rr.Body).Decode(&newTodo)
	if err != nil {
		panic("ahhh")
	}
	if newTodo.Id != todo1.Id {
		t.Errorf("newly created todo has wrong ID: got %v want %v",
			newTodo.Id, todo1.Id)
	}
	if newTodo.Title != todo1.Title {
		t.Errorf("newly created todo has wrong title: got %v want %v",
			newTodo.Title, todo1.Title)
	}

	// get non-existant Todo
	rr = httptest.NewRecorder()
	dne := httptest.NewRequest(http.MethodGet, "/todos/1986", nil)
	handler.ServeHTTP(rr, dne)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}

	// create another Todo and get all
	todo2 := models.Todo{
		Id:    "2",
		Title: "something else",
	}

	todo2Bytes, err := json.Marshal(todo2)
	if err != nil {
		panic("ahhh")
	}

	rr = httptest.NewRecorder()
	create2 := httptest.NewRequest(http.MethodPost, "/todos", bytes.NewBuffer(todo2Bytes))
	handler.ServeHTTP(rr, create2)
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	rr = httptest.NewRecorder()
	getAll := httptest.NewRequest(http.MethodGet, "/todos", nil)
	handler.ServeHTTP(rr, getAll)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var allTodos []models.Todo
	err = json.NewDecoder(rr.Body).Decode(&allTodos)
	if err != nil {
		panic("ahhh")
	}

	if len(allTodos) != 2 {
		t.Errorf("did not get all todos: got length %v expected length %v",
			len(allTodos), 2)
	}

	// update Todo
	todo2Done := models.Todo{
		Done: true,
	}
	todo2Bytes, err = json.Marshal(todo2Done)
	if err != nil {
		panic("ahhh")
	}

	rr = httptest.NewRecorder()
	patch := httptest.NewRequest(http.MethodPatch, "/todos/2", bytes.NewBuffer(todo2Bytes))
	handler.ServeHTTP(rr, patch)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// get updated Todo
	rr = httptest.NewRecorder()
	get = httptest.NewRequest(http.MethodGet, "/todos/2", nil)
	handler.ServeHTTP(rr, get)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	var updatedTodo models.Todo
	err = json.NewDecoder(rr.Body).Decode(&updatedTodo)
	if err != nil {
		panic("ahhh")
	}
	if updatedTodo.Id != todo2.Id {
		t.Errorf("newly created todo has wrong ID: got %v want %v",
			updatedTodo.Id, todo2.Id)
	}
	if updatedTodo.Title != todo2.Title {
		t.Errorf("newly created todo has wrong title: got %v want %v",
			updatedTodo.Title, todo2.Title)
	}
	if updatedTodo.Done != true {
		t.Errorf("newly created todo has wrong done: got %v want true", updatedTodo.Title)
	}

	// delete original Todo and get all
	rr = httptest.NewRecorder()
	delete := httptest.NewRequest(http.MethodDelete, "/todos/1", nil)
	handler.ServeHTTP(rr, delete)
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	rr = httptest.NewRecorder()
	getAll = httptest.NewRequest(http.MethodGet, "/todos", nil)
	handler.ServeHTTP(rr, getAll)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	err = json.NewDecoder(rr.Body).Decode(&allTodos)
	if err != nil {
		panic("ahhh")
	}

	if len(allTodos) != 1 {
		t.Errorf("did not get all todos: got length %v expected length %v",
			len(allTodos), 1)
	}
	if allTodos[0].Id != "2" {
		t.Errorf("todo did not have correct ID: got ID %v expected ID %v",
			allTodos[0].Id, 2)
	}

	// delete non-existing Todo
	rr = httptest.NewRecorder()
	dne = httptest.NewRequest(http.MethodDelete, "/todos/1986", nil)
	handler.ServeHTTP(rr, dne)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusNotFound)
	}
}

type InMemoryDB struct {
	todos []models.Todo
	id    int
}

func newInMemoryDB() Database {
	return &InMemoryDB{
		todos: []models.Todo{},
		id:    0,
	}
}

// Create implements Database.
func (db *InMemoryDB) Create(todo models.Todo) (id string, err error) {
	db.id++
	todo.Id = strconv.Itoa(db.id)
	todo.CreatedAt = time.Now()
	db.todos = append(db.todos, todo)
	return id, nil
}

// Get implements Database.
func (db *InMemoryDB) Get(id string) (todo models.Todo, err error) {
	for _, todo := range db.todos {
		if todo.Id == id {
			return todo, nil
		}
	}
	return models.Todo{}, fmt.Errorf("not found")
}

// Update implements Database.
func (db *InMemoryDB) Update(id string, todo models.Todo) (count int64, err error) {
	for i, t := range db.todos {
		if t.Id == id {
			title := todo.Title
			if title == "" {
				title = t.Title
			}
			db.todos[i] = models.Todo{
				Id:        t.Id,
				Title:     title,
				Done:      todo.Done,
				CreatedAt: t.CreatedAt,
			}
			return 1, nil
		}
	}
	return 0, fmt.Errorf("not found")
}

// Delete implements Database.
func (db *InMemoryDB) Delete(id string) (count int64, err error) {
	for i, todo := range db.todos {
		if todo.Id == id {
			db.todos = slices.Delete(db.todos, i, i+1)
			return 1, nil
		}
	}
	return 0, fmt.Errorf("not found")
}

// GetAll implements Database.
func (db *InMemoryDB) GetAll() (todos []models.Todo, err error) {
	return db.todos, nil
}

var _ Database = (*InMemoryDB)(nil)

type Database interface {
	Create(todo models.Todo) (id string, err error)
	Get(id string) (todo models.Todo, err error)
	GetAll() (todos []models.Todo, err error)
	Update(id string, todo models.Todo) (count int64, err error)
	Delete(id string) (count int64, err error)
}
