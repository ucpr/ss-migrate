package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

type (
	// CommandFunc defines the signature for command functions.
	CommandFunc func(args []string) error
	// Option for configuring the CLI.
	Option func(*CLI)
)

// WithOutput sets the output writer for the CLI.
func WithOutput(w io.Writer) Option {
	return func(c *CLI) {
		c.output = w
	}
}

type CLI struct {
	serviceName string
	version     string
	output      io.Writer
	commands    map[string]CommandFunc
}

func New(serviceName, version string, opts ...Option) *CLI {
	c := &CLI{
		serviceName: serviceName,
		version:     version,
		output:      os.Stdout,
		commands:    make(map[string]CommandFunc),
	}
	for _, opt := range opts {
		opt(c)
	}

	return c
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
		fmt.Fprintf(c.output, "%s version %s\n", c.serviceName, c.version)
		return nil
	}

	return nil
}

func (c *CLI) showHelp() {
	fmt.Fprintf(c.output, "Usage: %s [OPTIONS] COMMAND [ARGS...]\n", c.serviceName)
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
