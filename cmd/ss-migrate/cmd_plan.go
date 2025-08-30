package main

import (
	"context"
	"fmt"

	"github.com/ucpr/ss-migrate/internal/engine"
	"github.com/ucpr/ss-migrate/internal/schema"
	"github.com/ucpr/ss-migrate/internal/sheet"
)

func planCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ss-migrate plan <schema-file-path>")
	}

	schemaPath := args[0]

	// Load schema from file
	schemaConfig, err := schema.LoadFromFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Validate schema
	if err := schemaConfig.Validate(); err != nil {
		return fmt.Errorf("invalid schema: %w", err)
	}

	// Create context
	ctx := context.Background()

	// Create sheet client
	sheetClient, err := sheet.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create sheet client: %w", err)
	}

	// Create planner
	planner := engine.NewPlanner(sheetClient)

	// Generate plan for all resources
	results, err := planner.PlanAll(ctx, schemaConfig)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Display results
	hasAnyChanges := false
	for _, result := range results {
		fmt.Println(result.Format())
		if result.HasChanges {
			hasAnyChanges = true
		}
	}

	if !hasAnyChanges {
		fmt.Println("\n✓ All sheets are up to date with the schema.")
	} else {
		fmt.Println("\nRun 'ss-migrate apply' to apply these changes.")
	}

	return nil
}