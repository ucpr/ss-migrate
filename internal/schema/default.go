package schema

import "strings"

const DefaultSchemaTemplate = `resources:
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

func GetDefaultSchemaBytes() []byte {
	return []byte(strings.TrimSpace(DefaultSchemaTemplate) + "\n")
}