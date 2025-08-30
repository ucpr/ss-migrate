package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type CommandFunc func(args []string) error

type CLI struct {
	output   io.Writer
	commands map[string]CommandFunc
}

func New() *CLI {
	return &CLI{
		output:   os.Stdout,
		commands: make(map[string]CommandFunc),
	}
}

func NewWithOutput(output io.Writer) *CLI {
	return &CLI{
		output:   output,
		commands: make(map[string]CommandFunc),
	}
}

func (c *CLI) RegisterCommand(name string, fn CommandFunc) {
	c.commands[name] = fn
}

func (c *CLI) Run(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no command specified")
	}

	// Check for subcommand first (before parsing flags)
	if len(args) > 1 && !strings.HasPrefix(args[1], "-") {
		cmdName := args[1]
		if cmdFunc, exists := c.commands[cmdName]; exists {
			return cmdFunc(args[2:])
		}
	}

	// Parse global flags
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fs.SetOutput(c.output)

	version := fs.Bool("version", false, "show version")
	help := fs.Bool("help", false, "show help")

	if err := fs.Parse(args[1:]); err != nil {
		if err == flag.ErrHelp {
			c.showHelp()
			return nil
		}
		return err
	}

	if *help {
		c.showHelp()
		return nil
	}

	if *version {
		fmt.Fprintln(c.output, "ss-migrate version 0.1.0")
		return nil
	}

	return nil
}

func (c *CLI) showHelp() {
	fmt.Fprintln(c.output, "Usage: ss-migrate [OPTIONS] COMMAND [ARGS...]")
	fmt.Fprintln(c.output, "")
	fmt.Fprintln(c.output, "Options:")
	fmt.Fprintln(c.output, "  -help       Show this help message")
	fmt.Fprintln(c.output, "  -version    Show version information")
	fmt.Fprintln(c.output, "")
	fmt.Fprintln(c.output, "Commands:")
	for name := range c.commands {
		fmt.Fprintf(c.output, "  %s\n", name)
	}
}

