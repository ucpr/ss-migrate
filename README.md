## ss-migrate

A CLI tool for managing Google SpreadSheets schemas using YAML-based schema definitions.

## Features

- **Schema as Code**: Define your Google SpreadSheets structure in YAML format based on Frictionless Table Schema
- **Plan & Apply**: Preview changes before applying them (similar to Terraform workflow)
- **Type Management**: Automatic type detection and formatting for integer, number, datetime, and string types
- **Column Operations**: Add, remove, reorder, and modify columns automatically
- **Format Preservation**: Apply and maintain number formats, date formats, and text formats

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/ucpr/ss-migrate.git
cd ss-migrate

# Build the binary
go build -o ss-migrate ./cmd/ss-migrate

# Optional: Install to your PATH
go install ./cmd/ss-migrate
```

### Using Go install

```bash
go install github.com/ucpr/ss-migrate/cmd/ss-migrate@latest
```

## Authentication

ss-migrate uses Google Application Default Credentials (ADC). Set up authentication using one of these methods:

### Method 1: gcloud CLI

```bash
gcloud auth login --enable-gdrive-access
gcloud auth application-default login
```

### Method 2: Service Account

```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account-key.json"
```

## Usage

### Basic Commands

```bash
# Initialize a new schema file
ss-migrate init schema.yaml

# Plan changes (preview what will be changed)
ss-migrate plan schema.yaml

# Apply changes to the spreadsheet
ss-migrate apply schema.yaml

# Dry run (preview changes without applying)
ss-migrate apply schema.yaml --dry-run

# Auto-confirm without prompting
ss-migrate apply schema.yaml --yes
```

### Schema Format

Create a `schema.yaml` file defining your spreadsheet structure:

```yaml
resources:
  - name: "Sheet1"
    path: "https://docs.google.com/spreadsheets/d/{spreadsheet-id}/edit"
    x-header-row: 1       # Row containing headers (default: 1)
    x-header-column: 1    # Starting column (default: 1)
    fields:
      - name: "ID"
        type: "integer"
        x-protect: true   # Protect column from edits (future feature)
      
      - name: "Name"
        type: "string"
      
      - name: "Email"
        type: "string"
        format: "email"
      
      - name: "Age"
        type: "integer"
      
      - name: "Score"
        type: "number"    # Decimal numbers
      
      - name: "RegisteredAt"
        type: "datetime"
        format: "default" # Full datetime format
      
      - name: "BirthDate"
        type: "datetime"
        format: "date"    # Date only format
      
      - name: "Active"
        type: "boolean"
      
      - name: "InternalNotes"
        type: "string"
        x-hidden: true    # Hide this column in the spreadsheet
```

### Supported Data Types

| Type | Description | Google Sheets Format |
|------|-------------|---------------------|
| `string` | Text data | Text format (@) |
| `integer` | Whole numbers | Number format (0) |
| `number` | Decimal numbers | Number format (0.00) |
| `boolean` | True/False values | No special format |
| `datetime` | Date and time values | Date/time format |

#### DateTime Formats

- `default`: Full datetime (yyyy-mm-dd hh:mm:ss)
- `date`: Date only (yyyy-mm-dd)
- `time`: Time only (hh:mm:ss)

### Example Workflow

1. **Create a schema file**:

```yaml
resources:
  - name: "Users"
    path: "https://docs.google.com/spreadsheets/d/abc123/edit"
    fields:
      - name: "UserID"
        type: "integer"
      - name: "Username"
        type: "string"
      - name: "CreatedAt"
        type: "datetime"
```

2. **Preview changes**:

```bash
$ ss-migrate plan schema.yaml

=== Schema Migration Plan ===

Sheet 'Users': 2 field(s) to add

Changes to be applied:
  + Users.Username: Add new field 'Username' of type string
  + Users.CreatedAt: Add new field 'CreatedAt' of type datetime
```

3. **Apply changes**:

```bash
$ ss-migrate apply schema.yaml

✓ Applied changes to Users
Added field 'Username' to column B (type: string)
Added field 'CreatedAt' to column C (type: datetime(default))

✓ Users: Successfully applied 2 changes
```

### Advanced Features

#### Column Reordering

Fields in the schema define the expected column order. If columns exist in a different order, ss-migrate will reorder them:

```yaml
fields:
  - name: "ID"        # Should be column A
  - name: "Name"      # Should be column B
  - name: "Email"     # Should be column C
```

#### Hidden Columns

Use `x-hidden: true` to hide sensitive or internal columns:

```yaml
fields:
  - name: "PublicData"
    type: "string"
  - name: "InternalNotes"
    type: "string"
    x-hidden: true  # This column will be hidden
```

#### Type Changes

ss-migrate can change column types by applying appropriate formatting:

```yaml
fields:
  - name: "UserID"
    type: "integer"  # Changed from string to integer
```

## Limitations

- Currently supports Google SpreadSheets only
- Type changes apply formatting only (doesn't convert existing data)
- Column protection (`x-protect`) is planned but not yet implemented
- Formula preservation during column operations may require manual intervention

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) file for details
