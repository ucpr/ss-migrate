package cli

import (
	"bytes"
	"testing"
)

func TestNewCLI(t *testing.T) {
	cli := New()
	if cli == nil {
		t.Error("expected CLI instance, got nil")
	}
}

func TestRun(t *testing.T) {
	cli := New()
	err := cli.Run([]string{"cmd"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestFlagParsing(t *testing.T) {
	var output bytes.Buffer
	cli := NewWithOutput(&output)
	
	err := cli.Run([]string{"cmd", "-version"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	if output.String() == "" {
		t.Error("expected version output, got empty string")
	}
}

func TestSubcommand(t *testing.T) {
	var output bytes.Buffer
	cli := NewWithOutput(&output)
	
	// Register a test subcommand
	cli.RegisterCommand("test", func(args []string) error {
		output.WriteString("test command executed")
		return nil
	})
	
	err := cli.Run([]string{"cmd", "test"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	expected := "test command executed"
	if output.String() != expected {
		t.Errorf("expected %q, got %q", expected, output.String())
	}
}

func TestHelpFlag(t *testing.T) {
	var output bytes.Buffer
	cli := NewWithOutput(&output)
	
	cli.RegisterCommand("migrate", func(args []string) error {
		return nil
	})
	
	err := cli.Run([]string{"cmd", "-help"})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	if !bytes.Contains(output.Bytes(), []byte("Usage:")) {
		t.Error("expected help text to contain 'Usage:'")
	}
	
	if !bytes.Contains(output.Bytes(), []byte("migrate")) {
		t.Error("expected help text to list 'migrate' command")
	}
}