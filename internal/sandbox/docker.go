package sandbox

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/alexandrmotologa/restorecraft/internal/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// DockerSandbox provisions ephemeral PostgreSQL or MySQL containers via the Docker Engine API.
type DockerSandbox struct {
	cfg          config.DockerSandboxConfig
	engine       string
	drillID      string
	cli          *client.Client
	containerID  string
	connDetails  ConnectionDetails
}

// NewDockerSandbox creates a Docker-backed sandbox for relational database engines.
func NewDockerSandbox(drillID string, engine string, cfg config.DockerSandboxConfig) *DockerSandbox {
	if cfg.Image == "" {
		if engine == "postgres" {
			cfg.Image = "postgres:16-alpine"
		} else if engine == "mysql" {
			cfg.Image = "mysql:8.0"
		}
	}
	if cfg.MemoryMb == 0 {
		cfg.MemoryMb = 2048
	}

	return &DockerSandbox{
		cfg:     cfg,
		engine:  engine,
		drillID: drillID,
	}
}

func (d *DockerSandbox) Provision(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("connect to docker engine: %w", err)
	}
	d.cli = cli

	// Verify Docker daemon ping
	_, err = cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker daemon ping failed: %w", err)
	}

	// Ensure image exists locally or pull it
	_, _, err = cli.ImageInspectWithRaw(ctx, d.cfg.Image)
	if err != nil {
		// Pull image
		reader, pullErr := cli.ImagePull(ctx, d.cfg.Image, image.PullOptions{})
		if pullErr != nil {
			return fmt.Errorf("pull docker image %s: %w", d.cfg.Image, pullErr)
		}
		_ = reader.Close()
	}

	// Pick an available local port
	freePort, err := getFreePort()
	if err != nil {
		return fmt.Errorf("find free host port: %w", err)
	}

	dbName := "restorecraft_drill"
	dbUser := "postgres"
	dbPass := "drill_sandbox_secret"
	targetPort := "5432/tcp"

	env := []string{
		"POSTGRES_DB=" + dbName,
		"POSTGRES_USER=" + dbUser,
		"POSTGRES_PASSWORD=" + dbPass,
	}

	if d.engine == "mysql" {
		dbUser = "root"
		targetPort = "3306/tcp"
		env = []string{
			"MYSQL_DATABASE=" + dbName,
			"MYSQL_ROOT_PASSWORD=" + dbPass,
		}
	}

	containerConfig := &container.Config{
		Image: d.cfg.Image,
		Env:   env,
		Labels: map[string]string{
			"restorecraft-managed": "true",
			"restorecraft-drill-id": d.drillID,
		},
		ExposedPorts: nat.PortSet{
			nat.Port(targetPort): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(targetPort): []nat.PortBinding{
				{
					HostIP:   "127.0.0.1",
					HostPort: strconv.Itoa(freePort),
				},
			},
		},
		Resources: container.Resources{
			Memory: int64(d.cfg.MemoryMb) * 1024 * 1024,
		},
		Tmpfs: map[string]string{
			"/var/lib/postgresql/data": "rw,noexec,nosuid,size=1024m",
			"/var/lib/mysql":           "rw,noexec,nosuid,size=1024m",
		},
		AutoRemove: false,
	}

	created, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "restorecraft-"+d.drillID)
	if err != nil {
		return fmt.Errorf("create docker container: %w", err)
	}
	d.containerID = created.ID

	if err := cli.ContainerStart(ctx, d.containerID, container.StartOptions{}); err != nil {
		_ = d.Cleanup(ctx)
		return fmt.Errorf("start docker container: %w", err)
	}

	d.connDetails = ConnectionDetails{
		Host:         "127.0.0.1",
		Port:         freePort,
		User:         dbUser,
		Password:     dbPass,
		DatabaseName: dbName,
	}

	// Wait for container TCP port readiness
	if err := waitForPort(ctx, "127.0.0.1", freePort, 45*time.Second); err != nil {
		_ = d.Cleanup(ctx)
		return fmt.Errorf("sandbox port readiness timeout: %w", err)
	}

	return nil
}

func (d *DockerSandbox) ConnectionDetails() ConnectionDetails {
	return d.connDetails
}

func (d *DockerSandbox) Cleanup(ctx context.Context) error {
	if d.cli == nil || d.containerID == "" {
		return nil
	}

	timeout := 5
	stopOpts := container.StopOptions{Timeout: &timeout}
	_ = d.cli.ContainerStop(ctx, d.containerID, stopOpts)

	removeOpts := container.RemoveOptions{
		Force:         true,
		RemoveVolumes: true,
	}
	err := d.cli.ContainerRemove(ctx, d.containerID, removeOpts)
	d.containerID = ""
	_ = d.cli.Close()
	return err
}

func (d *DockerSandbox) Type() string {
	return "docker"
}

func getFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func waitForPort(ctx context.Context, host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	target := fmt.Sprintf("%s:%d", host, port)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			conn, err := net.DialTimeout("tcp", target, 500*time.Millisecond)
			if err == nil {
				_ = conn.Close()
				return nil
			}
			time.Sleep(300 * time.Millisecond)
		}
	}
	return fmt.Errorf("timeout waiting for %s", target)
}
