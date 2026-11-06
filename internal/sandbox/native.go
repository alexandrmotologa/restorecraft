package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// NativeSandbox manages an isolated ephemeral filesystem directory for file-based database drills.
type NativeSandbox struct {
	tempDir  string
	dbName   string
	filePath string
}

// NewNativeSandbox creates a sandbox using the local filesystem.
func NewNativeSandbox(dbName string) *NativeSandbox {
	if dbName == "" {
		dbName = "drill_sandbox.db"
	}
	return &NativeSandbox{
		dbName: dbName,
	}
}

func (n *NativeSandbox) Provision(ctx context.Context) error {
	dir, err := os.MkdirTemp("", "restorecraft_sandbox_*")
	if err != nil {
		return fmt.Errorf("create native sandbox directory: %w", err)
	}

	n.tempDir = dir
	n.filePath = filepath.Join(dir, n.dbName)
	return nil
}

func (n *NativeSandbox) ConnectionDetails() ConnectionDetails {
	return ConnectionDetails{
		FilePath:     n.filePath,
		DatabaseName: n.dbName,
	}
}

func (n *NativeSandbox) Cleanup(ctx context.Context) error {
	if n.tempDir != "" {
		err := os.RemoveAll(n.tempDir)
		n.tempDir = ""
		n.filePath = ""
		return err
	}
	return nil
}

func (n *NativeSandbox) Type() string {
	return "native"
}
