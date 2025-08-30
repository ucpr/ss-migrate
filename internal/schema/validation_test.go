package schema

import (
	"testing"
)

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