package db

import (
	"context"
	"fmt"
	"log"
	"os"

	"example.com/todos/pkg/models"
	"github.com/jackc/pgx/v5"
)

func NewDB(ctx context.Context, url string) *DB {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	return &DB{
		conn: conn,
	}
}

type DB struct {
	conn *pgx.Conn
}

func (db *DB) Create(todo models.Todo) (id string, err error) {
	err = db.conn.QueryRow(context.Background(), "INSERT INTO todos (title) VALUES ($1) RETURNING id", todo.Title).Scan(&id)
	return id, err
}

func (db *DB) Get(id string) (todo models.Todo, err error) {
	err = db.conn.QueryRow(context.Background(), "SELECT * from todos WHERE id = $1", id).Scan(&todo.Id, &todo.Title, &todo.Done, &todo.CreatedAt)
	return todo, err
}

func (db *DB) Update(id string, todo models.Todo) (count int64, err error) {
	commandTag, err := db.conn.Exec(context.Background(), "UPDATE todos SET title = $1, done = $2 WHERE id = $3", todo.Title, todo.Done, id)
	if err != nil {
		return -1, err
	}

	return commandTag.RowsAffected(), err
}

func (db *DB) GetAll() (todos []models.Todo, err error) {
	rows, err := db.conn.Query(context.Background(), "SELECT * from todos")
	if err != nil {
		return nil, fmt.Errorf("Error executing get all query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var todo models.Todo
		if err := rows.Scan(&todo.Id, &todo.Title, &todo.Done, &todo.CreatedAt); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue // Or handle the error as appropriate
		}
		todos = append(todos, todo)
	}

	return todos, err
}

func (db *DB) Delete(id string) (count int64, err error) {
	commandTag, err := db.conn.Exec(context.Background(), "DELETE from todos WHERE id = $1", id)
	if err != nil {
		return -1, err
	}

	return commandTag.RowsAffected(), err
}
