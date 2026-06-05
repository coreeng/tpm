package codegen

import (
	"fmt"
	"os"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/yamlutil"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// GenerateResult contains the results of a code generation operation
type GenerateResult struct {
	FilesModified []string
	CodesAdded    int
	Errors        []error
}

// GenerateMissingCodes generates UUIDs for any entities that are missing a code field
// This function ONLY adds codes where they are missing - it never replaces existing codes
func GenerateMissingCodes(rootDir, moduleName string) (*GenerateResult, error) {
	result := &GenerateResult{
		FilesModified: make([]string, 0),
	}

	// Load the module
	mod, err := module.LoadModule(rootDir, moduleName)
	if err != nil {
		return nil, fmt.Errorf("failed to load module: %w", err)
	}

	// Generate module-level code if missing
	if mod.Code == "" && mod.FilePath != "" {
		if err := yamlutil.AddFieldToYAML(mod.FilePath, "code", uuid.New().String()); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to add code to %s: %w", mod.FilePath, err))
		} else {
			result.FilesModified = append(result.FilesModified, mod.FilePath)
			result.CodesAdded++
		}
	}

	// Generate codes for chapters
	for _, chapter := range mod.Chapters {
		if chapter.Code == "" && chapter.FilePath != "" {
			if err := yamlutil.AddFieldToYAML(chapter.FilePath, "code", uuid.New().String()); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to add code to %s: %w", chapter.FilePath, err))
			} else {
				result.FilesModified = append(result.FilesModified, chapter.FilePath)
				result.CodesAdded++
			}
		}

		// Generate codes for sections
		for _, section := range chapter.Sections {
			if section.Code == "" && section.FilePath != "" {
				if err := yamlutil.AddFieldToYAML(section.FilePath, "code", uuid.New().String()); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to add code to %s: %w", section.FilePath, err))
				} else {
					result.FilesModified = append(result.FilesModified, section.FilePath)
					result.CodesAdded++
				}
			}
		}

		// Generate codes for interactive assessments
		for _, assessment := range chapter.Assessments {
			if assessment.Code == "" && assessment.FilePath != "" {
				if err := yamlutil.AddFieldToYAML(assessment.FilePath, "code", uuid.New().String()); err != nil {
					result.Errors = append(result.Errors, fmt.Errorf("failed to add code to %s: %w", assessment.FilePath, err))
				} else {
					result.FilesModified = append(result.FilesModified, assessment.FilePath)
					result.CodesAdded++
				}
			}

			// Generate codes for challenges
			for _, challenge := range assessment.Challenges {
				if challenge.Code == "" && challenge.FilePath != "" {
					if err := yamlutil.AddFieldToYAML(challenge.FilePath, "code", uuid.New().String()); err != nil {
						result.Errors = append(result.Errors, fmt.Errorf("failed to add code to %s: %w", challenge.FilePath, err))
					} else {
						result.FilesModified = append(result.FilesModified, challenge.FilePath)
						result.CodesAdded++
					}
				}

				// Generate codes for goals
				if challenge.FilePath != "" {
					modified, count, err := generateGoalCodes(challenge.FilePath, challenge.Goals)
					if err != nil {
						result.Errors = append(result.Errors, fmt.Errorf("failed to generate goal codes in %s: %w", challenge.FilePath, err))
					} else if modified {
						if !contains(result.FilesModified, challenge.FilePath) {
							result.FilesModified = append(result.FilesModified, challenge.FilePath)
						}
						result.CodesAdded += count
					}
				}
			}
		}

		// Generate codes for multiple choice assessments
		if len(chapter.MultipleChoiceAssessments) > 0 && chapter.FilePath != "" {
			modified, count, err := generateMCQCodes(chapter.FilePath, chapter.MultipleChoiceAssessments)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to generate MCQ codes in %s: %w", chapter.FilePath, err))
			} else if modified {
				if !contains(result.FilesModified, chapter.FilePath) {
					result.FilesModified = append(result.FilesModified, chapter.FilePath)
				}
				result.CodesAdded += count
			}
		}
	}

	return result, nil
}

// generateGoalCodes generates missing codes for goals in a challenge file
// Returns (modified, count, error)
func generateGoalCodes(filePath string, goals []module.Goal) (bool, int, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as a node
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return false, 0, fmt.Errorf("failed to parse YAML: %w", err)
	}

	modified := false
	codesAdded := 0

	// Find the goals array in the YAML
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		mappingNode := node.Content[0]
		if mappingNode.Kind == yaml.MappingNode {
			for i := 0; i < len(mappingNode.Content); i += 2 {
				if mappingNode.Content[i].Value == "goals" {
					goalsNode := mappingNode.Content[i+1]
					if goalsNode.Kind == yaml.SequenceNode {
						// Process each goal
						for _, goalNode := range goalsNode.Content {
							if goalNode.Kind == yaml.MappingNode {
								// Check if goal has a code field
								hasCode := false
								for j := 0; j < len(goalNode.Content); j += 2 {
									if goalNode.Content[j].Value == "code" {
										hasCode = true
										break
									}
								}

								if !hasCode {
									// Add code field to goal
									newUUID := uuid.New().String()
									keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "code"}
									valueNode := &yaml.Node{Kind: yaml.ScalarNode, Value: newUUID}

									// Insert at the beginning
									newContent := make([]*yaml.Node, 0, len(goalNode.Content)+2)
									newContent = append(newContent, keyNode, valueNode)
									newContent = append(newContent, goalNode.Content...)
									goalNode.Content = newContent

									modified = true
									codesAdded++
								}
							}
						}
					}
					break
				}
			}
		}
	}

	if modified {
		// Write back to file
		output, err := yaml.Marshal(&node)
		if err != nil {
			return false, 0, fmt.Errorf("failed to marshal YAML: %w", err)
		}

		if err := os.WriteFile(filePath, output, 0644); err != nil {
			return false, 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return modified, codesAdded, nil
}

// generateMCQCodes generates missing codes for MCQs and questions
// IMPORTANT: Only adds codes where missing - does NOT replace existing semantic codes
// Returns (modified, count, error)
func generateMCQCodes(filePath string, mcqs []module.MultipleChoiceAssessment) (bool, int, error) {
	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return false, 0, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse as a node
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return false, 0, fmt.Errorf("failed to parse YAML: %w", err)
	}

	modified := false
	codesAdded := 0

	// Find and process MCQs
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		mappingNode := node.Content[0]
		if mappingNode.Kind == yaml.MappingNode {
			for i := 0; i < len(mappingNode.Content); i += 2 {
				if mappingNode.Content[i].Value == "multipleChoiceAssessments" {
					mcqArrayNode := mappingNode.Content[i+1]
					if mcqArrayNode.Kind == yaml.SequenceNode {
						// Process each MCQ
						for _, mcqNode := range mcqArrayNode.Content {
							if mcqNode.Kind == yaml.MappingNode {
								// Check if MCQ has a code field
								hasCode := false
								for j := 0; j < len(mcqNode.Content); j += 2 {
									if mcqNode.Content[j].Value == "code" {
										hasCode = true
										break
									}
								}

								if !hasCode {
									// Add code field to MCQ
									newUUID := uuid.New().String()
									keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "code"}
									valueNode := &yaml.Node{Kind: yaml.ScalarNode, Value: newUUID}

									// Insert after index field (or at the beginning)
									insertPos := 0
									for j := 0; j < len(mcqNode.Content); j += 2 {
										if mcqNode.Content[j].Value == "index" {
											insertPos = j + 2
											break
										}
									}

									newContent := make([]*yaml.Node, 0, len(mcqNode.Content)+2)
									newContent = append(newContent, mcqNode.Content[:insertPos]...)
									newContent = append(newContent, keyNode, valueNode)
									newContent = append(newContent, mcqNode.Content[insertPos:]...)
									mcqNode.Content = newContent

									modified = true
									codesAdded++
								}

								// Generate question codes (ONLY if missing)
								for j := 0; j < len(mcqNode.Content); j += 2 {
									if mcqNode.Content[j].Value == "questions" {
										questionsNode := mcqNode.Content[j+1]
										if questionsNode.Kind == yaml.SequenceNode {
											for _, questionNode := range questionsNode.Content {
												if questionNode.Kind == yaml.MappingNode {
													// Check if question has a code field
													hasQuestionCode := false
													for k := 0; k < len(questionNode.Content); k += 2 {
														if questionNode.Content[k].Value == "code" {
															hasQuestionCode = true
															break
														}
													}

													// Only add if missing (do NOT replace semantic codes)
													if !hasQuestionCode {
														newUUID := uuid.New().String()
														keyNode := &yaml.Node{Kind: yaml.ScalarNode, Value: "code"}
														valueNode := &yaml.Node{Kind: yaml.ScalarNode, Value: newUUID}

														// Insert after index field
														insertPos := 0
														for k := 0; k < len(questionNode.Content); k += 2 {
															if questionNode.Content[k].Value == "index" {
																insertPos = k + 2
																break
															}
														}

														newContent := make([]*yaml.Node, 0, len(questionNode.Content)+2)
														newContent = append(newContent, questionNode.Content[:insertPos]...)
														newContent = append(newContent, keyNode, valueNode)
														newContent = append(newContent, questionNode.Content[insertPos:]...)
														questionNode.Content = newContent

														modified = true
														codesAdded++
													}
												}
											}
										}
										break
									}
								}
							}
						}
					}
					break
				}
			}
		}
	}

	if modified {
		// Write back to file
		output, err := yaml.Marshal(&node)
		if err != nil {
			return false, 0, fmt.Errorf("failed to marshal YAML: %w", err)
		}

		if err := os.WriteFile(filePath, output, 0644); err != nil {
			return false, 0, fmt.Errorf("failed to write file: %w", err)
		}
	}

	return modified, codesAdded, nil
}

// contains checks if a string slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
