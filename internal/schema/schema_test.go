package schema

import (
	"testing"
)

func TestLoadSchemaFromFile(t *testing.T) {
	// Create a temporary YAML file
	tmpFile := t.TempDir() + "/test_schema.yaml"
	yamlContent := `resources:
  - name: test_table
    path: https://docs.google.com/spreadsheets/d/test-id
    fields:
      - name: id
        type: integer
      - name: name
        type: string`

	err := WriteFile(tmpFile, []byte(yamlContent))
	if err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	schema, err := LoadFromFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to load schema from file: %v", err)
	}

	if len(schema.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(schema.Resources))
	}

	if schema.Resources[0].Name != "test_table" {
		t.Errorf("Expected resource name 'test_table', got %s", schema.Resources[0].Name)
	}
}

func TestParseSchemaFromYAML(t *testing.T) {
	yamlContent := `
resources:
  - name: example_table
    path: https://docs.google.com/spreadsheets/d/1_XXXXXXXXXXXXXXXX-xXXXXXXXXXXXX
    x-header-row: 1
    x-header-column: 1
    fields:
      - name: id
        type: integer
        x-protect: true
      - name: name
        type: string
      - name: created_at
        type: datetime
        format: default
`

	schema, err := ParseYAML([]byte(yamlContent))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if len(schema.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(schema.Resources))
	}

	resource := schema.Resources[0]
	if resource.Name != "example_table" {
		t.Errorf("Expected resource name 'example_table', got %s", resource.Name)
	}

	if resource.Path != "https://docs.google.com/spreadsheets/d/1_XXXXXXXXXXXXXXXX-xXXXXXXXXXXXX" {
		t.Errorf("Expected path to be spreadsheet URL, got %s", resource.Path)
	}

	if resource.HeaderRow != 1 {
		t.Errorf("Expected header row 1, got %d", resource.HeaderRow)
	}

	if resource.HeaderColumn != 1 {
		t.Errorf("Expected header column 1, got %d", resource.HeaderColumn)
	}

	if len(resource.Fields) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(resource.Fields))
	}

	// Test first field
	field := resource.Fields[0]
	if field.Name != "id" {
		t.Errorf("Expected field name 'id', got %s", field.Name)
	}
	if field.Type != "integer" {
		t.Errorf("Expected field type 'integer', got %s", field.Type)
	}
	if !field.Protect {
		t.Errorf("Expected field 'id' to be protected")
	}

	// Test second field
	field = resource.Fields[1]
	if field.Name != "name" {
		t.Errorf("Expected field name 'name', got %s", field.Name)
	}
	if field.Type != "string" {
		t.Errorf("Expected field type 'string', got %s", field.Type)
	}
	if field.Protect {
		t.Errorf("Expected field 'name' to not be protected")
	}

	// Test third field
	field = resource.Fields[2]
	if field.Name != "created_at" {
		t.Errorf("Expected field name 'created_at', got %s", field.Name)
	}
	if field.Type != "datetime" {
		t.Errorf("Expected field type 'datetime', got %s", field.Type)
	}
	if field.Format != "default" {
		t.Errorf("Expected field format 'default', got %s", field.Format)
	}
}

func TestSchemaValidation(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid schema",
			yaml: `resources:
  - name: users
    path: https://docs.google.com/spreadsheets/d/valid-id
    fields:
      - name: id
        type: integer`,
			wantErr: false,
		},
		{
			name: "missing resource name",
			yaml: `resources:
  - path: https://docs.google.com/spreadsheets/d/valid-id
    fields:
      - name: id
        type: integer`,
			wantErr: true,
			errMsg:  "resource name is required",
		},
		{
			name: "missing resource path",
			yaml: `resources:
  - name: users
    fields:
      - name: id
        type: integer`,
			wantErr: true,
			errMsg:  "resource path is required",
		},
		{
			name: "empty fields",
			yaml: `resources:
  - name: users
    path: https://docs.google.com/spreadsheets/d/valid-id
    fields: []`,
			wantErr: true,
			errMsg:  "at least one field is required",
		},
		{
			name: "missing field name",
			yaml: `resources:
  - name: users
    path: https://docs.google.com/spreadsheets/d/valid-id
    fields:
      - type: integer`,
			wantErr: true,
			errMsg:  "field name is required",
		},
		{
			name: "missing field type",
			yaml: `resources:
  - name: users
    path: https://docs.google.com/spreadsheets/d/valid-id
    fields:
      - name: id`,
			wantErr: true,
			errMsg:  "field type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := ParseYAML([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			err = schema.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected validation error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Expected error message %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected validation error: %v", err)
				}
			}
		})
	}
}
