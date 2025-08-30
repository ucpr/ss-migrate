package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const deafultSchema = `
resources:
  - name: example_table # name is sheet name
    # path to your Google Spreadsheets URL
    path: https://docs.google.com/spreadsheets/d/1_XXXXXXXXXXXXXXXX-xXXXXXXXXXXXX
    # optional: specify a specific tab within the spreadsheet (default is 1)
    # x-header-row: 1
    # optional: specify a specific column within the spreadsheet (default is 1)
    # x-header-column: 1
    fields:
      - name: id
        type: integer
        # optional: set to true to protect this field from being overwritten
        # x-protect: true
      - name: name
        type: string
      - name: created_at
        type: datetime
        format: default
`

func getDefaultSchema() []byte {
	return []byte(strings.TrimSpace(deafultSchema) + "\n")
}

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

	data := getDefaultSchema()
	if err := os.WriteFile(schemaPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write schema file: %w", err)
	}

	fmt.Printf("Schema file created: %s\n", schemaPath)
	return nil
}
