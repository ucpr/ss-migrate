package main

import (
	"fmt"
	"os"

	"github.com/ucpr/ss-migrate/internal/cli"
)

func main() {
	c := cli.New("ss-migrate", "v0.0.1")

	// Register commands
	c.RegisterCommand("init", initCommand)

	if err := c.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
