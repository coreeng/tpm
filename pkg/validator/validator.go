package validator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
	"github.com/coreeng/tpm/pkg/schema"
)

// ValidationLevel represents the severity of a validation issue
type ValidationLevel int

const (
	ErrorLevel ValidationLevel = iota
	WarningLevel
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Level   ValidationLevel
	File    string
	Field   string
	Message string
}

// ValidationResult contains all validation issues found
type ValidationResult struct {
	Issues []ValidationIssue
}

// HasErrors returns true if there are any errors in the validation result
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Level == ErrorLevel {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of errors
func (r *ValidationResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Level == ErrorLevel {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warnings
func (r *ValidationResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Level == WarningLevel {
			count++
		}
	}
	return count
}

// AddError adds an error to the validation result
func (r *ValidationResult) AddError(file, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Level:   ErrorLevel,
		File:    file,
		Field:   field,
		Message: message,
	})
}

// AddWarning adds a warning to the validation result
func (r *ValidationResult) AddWarning(file, field, message string) {
	r.Issues = append(r.Issues, ValidationIssue{
		Level:   WarningLevel,
		File:    file,
		Field:   field,
		Message: message,
	})
}

// ValidateModule performs comprehensive validation on a module.
//
// Validation checks:
//   - JSON Schema validation against source schemas (structure, required fields, types, patterns)
//   - Required markdown files exist (description.md, successMessage.md)
//
// Note: Schema validation with additionalProperties:false automatically rejects
// description/successMessage fields in YAML (must be in .md files). Pattern validation
// in schemas enforces code formats (UUIDs, semantic codes, etc.).
//
// Returns a ValidationResult containing all validation issues found.
func ValidateModule(mod *module.Module, moduleName string, schemaDir string) *ValidationResult {
	result := &ValidationResult{}

	// Schema validation (runs first, fails fast on schema violations).
	// An empty schemaDir means "use the schemas embedded in the binary".
	validateModuleSchema(mod, schemaDir, result)
	// If schema validation failed, return early
	if result.HasErrors() {
		return result
	}

	// Validate required markdown files exist
	validateMarkdownFiles(mod, result)

	return result
}

// validateModuleSchema validates the module structure against JSON schemas
func validateModuleSchema(mod *module.Module, schemaDir string, result *ValidationResult) {
	// Create schema validator
	validator, err := schema.NewValidator(schemaDir)
	if err != nil {
		result.AddError("", "schema", fmt.Sprintf("failed to initialize schema validator: %v", err))
		return
	}

	// Validate module.yaml file
	if mod.FilePath != "" {
		errors, err := validator.ValidateYAMLFile(mod.FilePath, "module.schema.json")
		if err != nil {
			result.AddError(pathutil.GetRelativeModulePath(mod.FilePath), "schema", fmt.Sprintf("schema validation error: %v", err))
			return
		}
		for _, schemaErr := range errors {
			result.AddError(pathutil.GetRelativeModulePath(mod.FilePath), schemaErr.Field, schemaErr.Message)
		}
	}

	// Validate chapters
	for _, chapter := range mod.Chapters {
		validateChapterSchema(&chapter, validator, result)
	}
}

// validateChapterSchema validates a chapter against its schema
func validateChapterSchema(chapter *module.Chapter, validator *schema.Validator, result *ValidationResult) {
	if chapter.FilePath != "" {
		errors, err := validator.ValidateYAMLFile(chapter.FilePath, "chapter.schema.json")
		if err != nil {
			result.AddError(pathutil.GetRelativeModulePath(chapter.FilePath), "schema", fmt.Sprintf("schema validation error: %v", err))
			return
		}
		for _, schemaErr := range errors {
			result.AddError(pathutil.GetRelativeModulePath(chapter.FilePath), schemaErr.Field, schemaErr.Message)
		}
	}

	// Validate sections
	for _, section := range chapter.Sections {
		validateSectionSchema(&section, validator, result)
	}

	// Validate assessments
	for _, assessment := range chapter.Assessments {
		validateAssessmentSchema(&assessment, validator, result)
	}
}

// validateSectionSchema validates a section against its schema
func validateSectionSchema(section *module.Section, validator *schema.Validator, result *ValidationResult) {
	if section.FilePath != "" {
		errors, err := validator.ValidateYAMLFile(section.FilePath, "section.schema.json")
		if err != nil {
			result.AddError(pathutil.GetRelativeModulePath(section.FilePath), "schema", fmt.Sprintf("schema validation error: %v", err))
			return
		}
		for _, schemaErr := range errors {
			result.AddError(pathutil.GetRelativeModulePath(section.FilePath), schemaErr.Field, schemaErr.Message)
		}
	}
}

// validateAssessmentSchema validates an assessment against its schema
func validateAssessmentSchema(assessment *module.Assessment, validator *schema.Validator, result *ValidationResult) {
	if assessment.FilePath != "" {
		errors, err := validator.ValidateYAMLFile(assessment.FilePath, "assessment.schema.json")
		if err != nil {
			result.AddError(pathutil.GetRelativeModulePath(assessment.FilePath), "schema", fmt.Sprintf("schema validation error: %v", err))
			return
		}
		for _, schemaErr := range errors {
			result.AddError(pathutil.GetRelativeModulePath(assessment.FilePath), schemaErr.Field, schemaErr.Message)
		}
	}

	// Validate challenges
	for _, challenge := range assessment.Challenges {
		validateChallengeSchema(&challenge, validator, result)
	}
}

// validateChallengeSchema validates a challenge against its schema
func validateChallengeSchema(challenge *module.Challenge, validator *schema.Validator, result *ValidationResult) {
	if challenge.FilePath != "" {
		errors, err := validator.ValidateYAMLFile(challenge.FilePath, "challenge.schema.json")
		if err != nil {
			result.AddError(pathutil.GetRelativeModulePath(challenge.FilePath), "schema", fmt.Sprintf("schema validation error: %v", err))
			return
		}
		for _, schemaErr := range errors {
			result.AddError(pathutil.GetRelativeModulePath(challenge.FilePath), schemaErr.Field, schemaErr.Message)
		}
	}
}

// validateMarkdownFiles validates that required markdown files exist
func validateMarkdownFiles(mod *module.Module, result *ValidationResult) {
	// Module description.md
	if mod.FilePath != "" {
		moduleDir := filepath.Dir(mod.FilePath)
		checkMarkdownFile(moduleDir, "description.md", pathutil.GetRelativeModulePath(mod.FilePath), result)
	}

	// Chapters
	for _, chapter := range mod.Chapters {
		validateChapterMarkdown(&chapter, result)
	}
}

// validateChapterMarkdown validates markdown files for a chapter
func validateChapterMarkdown(chapter *module.Chapter, result *ValidationResult) {
	if chapter.FilePath != "" {
		chapterDir := filepath.Dir(chapter.FilePath)
		checkMarkdownFile(chapterDir, "description.md", pathutil.GetRelativeModulePath(chapter.FilePath), result)
	}

	// Sections
	for _, section := range chapter.Sections {
		if section.FilePath != "" {
			sectionDir := filepath.Dir(section.FilePath)
			checkMarkdownFile(sectionDir, "description.md", pathutil.GetRelativeModulePath(section.FilePath), result)
		}
	}

	// Assessments
	for _, assessment := range chapter.Assessments {
		validateAssessmentMarkdown(&assessment, result)
	}
}

// validateAssessmentMarkdown validates markdown files for an assessment
func validateAssessmentMarkdown(assessment *module.Assessment, result *ValidationResult) {
	if assessment.FilePath != "" {
		assessmentDir := filepath.Dir(assessment.FilePath)
		checkMarkdownFile(assessmentDir, "description.md", pathutil.GetRelativeModulePath(assessment.FilePath), result)
	}

	// Challenges
	for _, challenge := range assessment.Challenges {
		if challenge.FilePath != "" {
			challengeDir := filepath.Dir(challenge.FilePath)
			checkMarkdownFile(challengeDir, "description.md", pathutil.GetRelativeModulePath(challenge.FilePath), result)
			checkMarkdownFile(challengeDir, "successMessage.md", pathutil.GetRelativeModulePath(challenge.FilePath), result)
		}
	}
}

// checkMarkdownFile checks if a markdown file exists and adds an error if it doesn't
func checkMarkdownFile(dir, filename, yamlFile string, result *ValidationResult) {
	mdPath := filepath.Join(dir, filename)
	if _, err := os.Stat(mdPath); os.IsNotExist(err) {
		result.AddError(yamlFile, filename, fmt.Sprintf("required markdown file '%s' does not exist", filename))
	} else if err != nil {
		result.AddError(yamlFile, filename, fmt.Sprintf("error checking markdown file '%s': %v", filename, err))
	}
}
