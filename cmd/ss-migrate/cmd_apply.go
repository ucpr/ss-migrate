package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ucpr/ss-migrate/internal/engine"
	"github.com/ucpr/ss-migrate/internal/schema"
	"github.com/ucpr/ss-migrate/internal/sheet"
)

func applyCommand(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ss-migrate apply <schema-file-path> [--dry-run] [--yes]")
	}

	var schemaPath string
	dryRun := false
	autoConfirm := false

	// Parse flags and find schema path
	for _, arg := range args {
		switch arg {
		case "--dry-run":
			dryRun = true
		case "--yes", "-y":
			autoConfirm = true
		default:
			if !strings.HasPrefix(arg, "-") && schemaPath == "" {
				schemaPath = arg
			}
		}
	}

	if schemaPath == "" {
		return fmt.Errorf("usage: ss-migrate apply <schema-file-path> [--dry-run] [--yes]")
	}

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

	// First, generate plan to show what will be changed
	planner := engine.NewPlanner(sheetClient)
	diffs, err := planner.PlanAll(ctx, schemaConfig)
	if err != nil {
		return fmt.Errorf("failed to generate plan: %w", err)
	}

	// Check if there are any changes
	hasChanges := false
	for _, diff := range diffs {
		if diff.HasChanges {
			hasChanges = true
			fmt.Println(diff.Format())
		}
	}

	if !hasChanges {
		fmt.Println("✓ All sheets are already up to date with the schema.")
		return nil
	}

	// Show dry run notice
	if dryRun {
		fmt.Println("\n=== DRY RUN MODE ===")
		fmt.Println("No actual changes will be made to the sheets.")
	}

	// Confirm before applying
	if !dryRun && !autoConfirm {
		fmt.Print("\nDo you want to apply these changes? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Apply cancelled.")
			return nil
		}
	}

	// Create applier
	applier := engine.NewApplier(sheetClient, dryRun)

	// Apply changes
	fmt.Println("\nApplying changes...")
	results, err := applier.ApplyAll(ctx, schemaConfig)
	if err != nil {
		return fmt.Errorf("failed to apply changes: %w", err)
	}

	// Display results
	totalApplied := 0
	totalErrors := 0
	for i, result := range results {
		resourceName := schemaConfig.Resources[i].Name
		if result.Success {
			if result.ChangesApplied > 0 {
				fmt.Printf("✓ %s: %s\n", resourceName, result.Message)
				totalApplied += result.ChangesApplied
			} else {
				fmt.Printf("- %s: %s\n", resourceName, result.Message)
			}
		} else {
			fmt.Printf("✗ %s: %s\n", resourceName, result.Message)
			for _, err := range result.Errors {
				fmt.Printf("  Error: %v\n", err)
			}
			totalErrors += len(result.Errors)
		}
	}

	// Summary
	fmt.Println("\n=== Summary ===")
	if dryRun {
		fmt.Printf("DRY RUN completed. Would apply %d changes.\n", totalApplied)
	} else if totalErrors > 0 {
		fmt.Printf("Applied %d changes with %d errors.\n", totalApplied, totalErrors)
		return fmt.Errorf("some changes failed to apply")
	} else {
		fmt.Printf("✓ Successfully applied %d changes.\n", totalApplied)
	}

	return nil
}