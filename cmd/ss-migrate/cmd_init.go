package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ucpr/ss-migrate/internal/schema"
)


func initCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ss-migrate init <schema-file-path>")
	}

	schemaPath := args[0]

	dir := filepath.Dir(schemaPath)
	if dir != "." && dir != "/" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	data := schema.GetDefaultSchemaBytes()
	if err := os.WriteFile(schemaPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	fmt.Printf("Schema file created: %s\n", schemaPath)
	return nil
}
