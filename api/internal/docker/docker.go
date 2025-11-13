package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func Command(args ...string) (string, error) {
	cmd := exec.Command("docker", args...)
	out := bytes.Buffer{}
	cmd.Stdout = &out
	fmt.Println("$", strings.Join(cmd.Args, " "))
	err := cmd.Run()

	return strings.TrimSpace(out.String()), err
}

type ContainerSpec struct {
	Name           string
	Env            []string
	Ports          []string
	Volumes        []string
	HealthCmd      string
	HealthInterval string
	Image          string
	Args           []string
}

func CreateContainer(spec ContainerSpec) (string, error) {
	args := []string{"create"}
	if spec.Name != "" {
		args = append(args, "--name="+spec.Name)
	}

	for _, e := range spec.Env {
		args = append(args, "--env="+e)
	}

	for _, p := range spec.Ports {
		args = append(args, "--publish="+p)
	}

	for _, v := range spec.Volumes {
		args = append(args, "--volume="+v)
	}

	if spec.HealthCmd != "" {
		args = append(args, "--health-cmd="+spec.HealthCmd)
	}
	if spec.HealthInterval != "" {
		args = append(args, "--health-interval="+spec.HealthInterval)
	}

	args = append(args, spec.Image)
	args = append(args, spec.Args...)

	return Command(args...)
}

func StartContainer(containerID string) (string, error) {
	_, err := Command("start", containerID)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}
	return containerID, nil
}

func RemoveContainer(name string) error {
	_, err := Command("container", "rm", "--force", "--volumes", name)
	return err
}

func PostgresURL(containerName string) (string, error) {
	postgresPort, err := ContainerPort(containerName, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get database port: %v\n", err)
		return "", err
	}

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	username := "postgres"
	password := "mysecretpassword"
	databaseName := "testdb"
	return fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", username, password, postgresPort, databaseName), nil
}

func PostgresDB(containerName string) (*pgx.Conn, error) {
	postgresPort, err := ContainerPort(containerName, "5432")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to get database port: %v\n", err)
		return nil, err
	}

	// urlExample := "postgres://username:password@localhost:5432/database_name"
	username := "postgres"
	password := "mysecretpassword"
	databaseName := "testdb"
	url := fmt.Sprintf("postgres://%s:%s@localhost:%s/%s", username, password, postgresPort, databaseName)

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		return nil, err
	}
	return conn, nil
}

func ContainerPort(containerID, portSpec string) (string, error) {
	hostPort, err := Command("port", containerID, portSpec)
	if err != nil {
		return "", err
	}
	// example output
	// ❯ docker container port 93e7d30919cc
	// 5432/tcp -> 0.0.0.0:55637       IPv4
	// 5432/tcp -> [::]:55637.         IPv6

	_, port, err := net.SplitHostPort(strings.Fields(hostPort)[0])
	if err != nil {
		return "", err
	}

	return port, nil
}

func WaitHealthy(container string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timer.C:
			return errors.New("timeout waiting for healthy container")
		case <-ticker.C:
			result, err := Command("container", "inspect", "--format={{.State.Status}} {{.State.Health.Status}}", container)
			if err != nil {
				return fmt.Errorf("error inspecting container: %v", err)
			}

			if strings.HasPrefix(result, "exited ") {
				return fmt.Errorf("container exited")
			}

			if result == "running healthy" {
				return nil
			}
		}
	}
}
