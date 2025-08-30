package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/ucpr/ss-migrate/internal/schema"
)

func TestInitCommand(t *testing.T) {
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "test-schema.yaml")

	err := initCommand([]string{schemaPath})
	if err != nil {
		t.Fatalf("initCommand failed: %v", err)
	}

	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Fatal("schema file was not created")
	}

	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	expected := schema.GetDefaultSchemaBytes()
	if !bytes.Equal(data, expected) {
		t.Errorf("schema content does not match default schema\ngot:\n%s\nwant:\n%s", data, expected)
	}
}

func TestInitCommandWithSubdirectory(t *testing.T) {
	tempDir := t.TempDir()
	schemaPath := filepath.Join(tempDir, "subdir", "schema.json")

	err := initCommand([]string{schemaPath})
	if err != nil {
		t.Fatalf("initCommand failed: %v", err)
	}

	if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
		t.Fatal("schema file was not created in subdirectory")
	}
}

func TestInitCommandWithoutArgs(t *testing.T) {
	err := initCommand([]string{})
	if err == nil {
		t.Fatal("expected error when no arguments provided")
	}
}
