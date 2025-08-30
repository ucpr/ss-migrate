package schema

import (
	"errors"
	"os"

	"github.com/goccy/go-yaml"
)

type Schema struct {
	Resources []Resource `yaml:"resources"`
}

type Resource struct {
	Name         string  `yaml:"name"`
	Path         string  `yaml:"path"`
	HeaderRow    int     `yaml:"x-header-row"`
	HeaderColumn int     `yaml:"x-header-column"`
	Fields       []Field `yaml:"fields"`
}

type Field struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Format  string `yaml:"format"`
	Protect bool   `yaml:"x-protect"`
	Hidden  bool   `yaml:"x-hidden"`
}

func ParseYAML(data []byte) (*Schema, error) {
	var schema Schema
	err := yaml.Unmarshal(data, &schema)
	if err != nil {
		return nil, err
	}
	
	// Set default values if not specified
	for i := range schema.Resources {
		if schema.Resources[i].HeaderRow == 0 {
			schema.Resources[i].HeaderRow = 1
		}
		if schema.Resources[i].HeaderColumn == 0 {
			schema.Resources[i].HeaderColumn = 1
		}
	}
	
	return &schema, nil
}

func LoadFromFile(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseYAML(data)
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

func (s *Schema) Validate() error {
	if len(s.Resources) == 0 {
		return errors.New("at least one resource is required")
	}

	for _, resource := range s.Resources {
		if resource.Name == "" {
			return errors.New("resource name is required")
		}
		if resource.Path == "" {
			return errors.New("resource path is required")
		}
		if len(resource.Fields) == 0 {
			return errors.New("at least one field is required")
		}
		
		for _, field := range resource.Fields {
			if field.Name == "" {
				return errors.New("field name is required")
			}
			if field.Type == "" {
				return errors.New("field type is required")
			}
		}
	}
	
	return nil
}