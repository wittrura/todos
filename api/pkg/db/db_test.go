package db_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"example.com/todos/internal/docker"
	. "example.com/todos/pkg/db"
	"example.com/todos/pkg/models"
)

func TestDB(t *testing.T) {
	// start postgres docker container
	containerName := startPostgresContainer(t)

	// cleanup
	defer cleanupContainer(t, containerName)

	// allow for the container to be healthy before trying to connect
	err := docker.WaitHealthy(containerName, time.Second*5)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatalf("failed to connect to Postgres db, %v", err)
	}

	url, err := docker.PostgresURL(containerName)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatalf("failed to get Postgres URL, %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	sut := NewDB(ctx, url)
	defer cancel()

	t.Run("happy path", func(t *testing.T) {
		todo := models.Todo{
			Title: "a newly created todo",
		}

		id, err := sut.Create(todo)
		if err != nil {
			t.Fatalf("failed to create new todo, %v", err)
		}

		newTodo, err := sut.Get(id)
		if err != nil {
			t.Fatalf("failed to get new todo, %v", err)
		}

		if newTodo.Title != todo.Title {
			t.Fatalf("newly created todo has bad data, expected: %s, got: %s", todo.Title, newTodo.Title)
		}

		newTodo.Done = true
		updatedRecords, err := sut.Update(newTodo.Id, newTodo)
		if updatedRecords != 1 {
			t.Fatalf("updated wrong number of records, expected: 1, got: %d", updatedRecords)
		}

		completedTodo, err := sut.Get(id)
		if err != nil {
			t.Fatalf("failed to get new todo, %v", err)
		}

		if completedTodo.Done != true {
			t.Fatalf("updated todo has bad data for 'done', expected: %t, got: %t", true, completedTodo.Done)
		}
		if completedTodo.Title != todo.Title {
			t.Fatalf("updated todo has bad data for 'title', expected: %s, got: %s", todo.Title, completedTodo.Title)
		}

		allTodos, err := sut.GetAll()
		if err != nil {
			t.Fatalf("failed to get all todos, %v", err)
		}
		if len(allTodos) != 1 {
			t.Fatalf("wrong number of total todos, expected: 1, got: %d", len(allTodos))
		}

		deletedRecords, err := sut.Delete(completedTodo.Id)
		if err != nil {
			t.Fatalf("failed to delete todo: %s, %v", completedTodo.Id, err)
		}
		if deletedRecords != 1 {
			t.Fatalf("wrong number of deleted todos, expected: 1, got: %d", deletedRecords)
		}

		allTodos, err = sut.GetAll()
		if err != nil {
			t.Fatalf("failed to get all todos, %v", err)
		}
		if len(allTodos) != 0 {
			t.Fatalf("wrong number of total todos, expected: 0, got: %d", len(allTodos))
		}
	})
}

func startPostgresContainer(t *testing.T) string {
	containerID, err := docker.CreateContainer(docker.ContainerSpec{
		Image:          "postgres",
		Name:           "some-postgres",
		Env:            []string{"POSTGRES_PASSWORD=mysecretpassword"},
		Ports:          []string{"5432"},
		Volumes:        []string{"./fixtures/seed.sql:/docker-entrypoint-initdb.d/seed.sql"},
		HealthCmd:      "pg_isready",
		HealthInterval: "1s",
	})

	if err != nil {
		fmt.Println("Error:", err)
		t.Fatalf("failed to create postgres container, %v", err)
	}

	containerID, err = docker.StartContainer(containerID)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatalf("failed to start postgres container, %v", err)
	}

	return containerID
}

func cleanupContainer(t *testing.T, id string) {
	err := docker.RemoveContainer(id)
	if err != nil {
		fmt.Println("Error:", err)
		t.Fatalf("failed to remove postgres container, %v", err)
	}
}
