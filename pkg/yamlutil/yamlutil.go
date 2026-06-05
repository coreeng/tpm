package yamlutil

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// AddFieldToYAML adds a field to a YAML file at the beginning of the mapping
// Preserves formatting and comments using yaml.v3's Node API
func AddFieldToYAML(filePath, fieldName, fieldValue string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as a node to preserve formatting
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Add field at the top of the document
	if err := addFieldToNode(&node, fieldName, fieldValue); err != nil {
		return fmt.Errorf("failed to add field: %w", err)
	}

	// Write back to file
	output, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// RemoveFieldFromYAML removes a field from a YAML file
// Preserves formatting and comments using yaml.v3's Node API
func RemoveFieldFromYAML(filePath, fieldName string) error {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as a node to preserve formatting
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Remove the field from the document
	if err := removeFieldFromNode(&node, fieldName); err != nil {
		return fmt.Errorf("failed to remove field: %w", err)
	}

	// Write back to file
	output, err := yaml.Marshal(&node)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if err := os.WriteFile(filePath, output, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// addFieldToNode adds a field to a YAML node at the beginning
func addFieldToNode(node *yaml.Node, fieldName, fieldValue string) error {
	// The root node is a Document node, we need to get to the MappingNode
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("expected document node")
	}

	if len(node.Content) == 0 {
		return fmt.Errorf("empty document")
	}

	mappingNode := node.Content[0]
	if mappingNode.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node")
	}

	// Check if field already exists
	for i := 0; i < len(mappingNode.Content); i += 2 {
		if mappingNode.Content[i].Value == fieldName {
			// Field already exists, don't modify
			return nil
		}
	}

	// Create new field nodes
	keyNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: fieldName,
	}
	valueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Value: fieldValue,
	}

	// Prepend field to the mapping (add at the beginning)
	newContent := make([]*yaml.Node, 0, len(mappingNode.Content)+2)
	newContent = append(newContent, keyNode, valueNode)
	newContent = append(newContent, mappingNode.Content...)
	mappingNode.Content = newContent

	return nil
}

// removeFieldFromNode removes a field from a YAML node
func removeFieldFromNode(node *yaml.Node, fieldName string) error {
	// The root node is a Document node, we need to get to the MappingNode
	if node.Kind != yaml.DocumentNode {
		return fmt.Errorf("expected document node")
	}

	if len(node.Content) == 0 {
		return fmt.Errorf("empty document")
	}

	mappingNode := node.Content[0]
	if mappingNode.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node")
	}

	// Find and remove the field
	// MappingNode content is key-value pairs: [key1, value1, key2, value2, ...]
	for i := 0; i < len(mappingNode.Content); i += 2 {
		if i+1 < len(mappingNode.Content) && mappingNode.Content[i].Value == fieldName {
			// Found the field, remove both key and value nodes
			newContent := make([]*yaml.Node, 0, len(mappingNode.Content)-2)
			newContent = append(newContent, mappingNode.Content[:i]...)
			newContent = append(newContent, mappingNode.Content[i+2:]...)
			mappingNode.Content = newContent
			return nil
		}
	}

	// Field not found, that's OK (might already be removed)
	return nil
}
