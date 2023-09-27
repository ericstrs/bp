package storage

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLStorage implements a YAML storage solution
type YAMLStorage struct {
	Filename string
}

// Save saves Storable data into a YAML file.
func (ys *YAMLStorage) Save(name string, data Storable) error {
	// Serialize the Storable object to YAML format
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to serialize data: %v", err)
	}

	// Generate a filename based on the name
	filename := fmt.Sprintf("%s_%s.yaml", ys.Filename, name)

	// Write the serialized data to a file
	if err := ioutil.WriteFile(filename, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write to file: %v", err)
	}

	return nil
}

// Load loads Storable data from a YAML file.
func (ys *YAMLStorage) Load(name string, into Storable) error {
	// Generate a filename based on the name
	filename := fmt.Sprintf("%s_%s.yaml", ys.Filename, name)

	// Open or create the file with read-only permissions
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open or create file: %v", err)
	}
	defer file.Close()

	// Read the file into a byte slice
	yamlData, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read from file: %v", err)
	}

	// Deserialize the data into the Storable object
	if err := yaml.Unmarshal(yamlData, into); err != nil {
		return fmt.Errorf("failed to deserialize data: %v", err)
	}

	return nil
}
