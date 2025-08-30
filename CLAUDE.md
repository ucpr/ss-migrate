# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ss-migrate is a CLI tool for managing Google Sheets schemas using Frictionless Table Schema (FTS) as the source of truth. It provides plan/diff/apply/check operations similar to Terraform but for spreadsheet schema management.

## Build and Test Commands

```bash
# Build the CLI
go build ./cmd/ss-migrate

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/schema
go test ./internal/engine
go test ./internal/sheet

# Run tests with verbose output
go test -v ./...

# Check and update dependencies
go mod tidy
```

## Architecture

### Core Components

1. **CLI Layer** (`internal/cli/`)
   - Custom CLI framework with command registration pattern
   - Commands registered in `cmd/ss-migrate/main.go`

2. **Schema Management** (`internal/schema/`)
   - Handles Frictionless Table Schema parsing and validation
   - YAML-based schema format with custom extensions (`x-header-row`, `x-protect`)
   - Resources contain spreadsheet URL and field definitions

3. **Google Sheets Client** (`internal/sheet/`)
   - Wraps Google Sheets API v4
   - Uses Application Default Credentials (ADC) for authentication
   - Provides operations: read headers, analyze columns, insert columns, update values

4. **Migration Engine** (`internal/engine/`)
   - **Planner**: Compares current sheet state with schema, generates diff
   - **Applier**: Executes migrations based on diff result
   - **Diff**: Core diffing logic for field comparison and ordering

### Command Flow

1. **init**: Creates a new schema file with default structure
2. **plan**: Analyzes differences between schema and actual sheet
3. **apply**: Executes the migration plan to update the sheet
4. **check**: Validates that sheet matches schema (not yet implemented)

### Schema Format

```yaml
resources:
  - name: "Sheet1"
    path: "https://docs.google.com/spreadsheets/d/{spreadsheet-id}/..."
    x-header-row: 1     # Row number containing headers (default: 1)
    x-header-column: 1  # Starting column (default: 1)
    fields:
      - name: "ID"
        type: "integer"
        x-protect: true  # Protect column from edits
      - name: "Name"
        type: "string"
      - name: "InternalNotes"
        type: "string"
        x-hidden: true   # Hide column in spreadsheet
```

## Authentication

The tool uses Google Application Default Credentials (ADC). Set up authentication via:
- `gcloud auth application-default login` for local development
- Service account JSON via `GOOGLE_APPLICATION_CREDENTIALS` environment variable

## Key Implementation Details

- Field ordering is preserved based on schema definition
- Type inference from existing data when analyzing sheets
- Column insertion maintains proper positioning
- Column deletion removes entire columns (all rows)
- Column visibility can be controlled via `x-hidden` field property
- Batch operations for performance when updating multiple columns