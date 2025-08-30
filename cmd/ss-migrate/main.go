package main

import (
	"fmt"
	"os"

	"github.com/ucpr/ss-migrate/internal/cli"
)

func main() {
	c := cli.New("ss-migrate", "v0.0.1")

	// Register example commands
	c.RegisterCommand("migrate", func(args []string) error {
		fmt.Println("Running migration...")
		// TODO: Implement actual migration logic
		return nil
	})

	c.RegisterCommand("rollback", func(args []string) error {
		fmt.Println("Rolling back...")
		// TODO: Implement actual rollback logic
		return nil
	})

	if err := c.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
