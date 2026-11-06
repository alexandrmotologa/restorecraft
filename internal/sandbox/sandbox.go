package sandbox

import (
	"context"
)

// ConnectionDetails contains connection parameters to access the running sandbox.
type ConnectionDetails struct {
	Host         string
	Port         int
	User         string
	Password     string
	DatabaseName string
	FilePath     string // For SQLite databases
}

// Sandbox represents an ephemeral isolated execution environment for database restores.
type Sandbox interface {
	// Provision boots the isolated sandbox (e.g. starts container or creates temp workspace).
	Provision(ctx context.Context) error
	// ConnectionDetails returns the parameters to connect to the restored database.
	ConnectionDetails() ConnectionDetails
	// Cleanup destroys the sandbox, terminating containers and removing temporary disk artifacts.
	Cleanup(ctx context.Context) error
	// Type returns the sandbox driver kind ("docker" or "native").
	Type() string
}
