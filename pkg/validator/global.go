package validator

import (
	"fmt"

	"github.com/coreeng/tpm/pkg/module"
	"github.com/coreeng/tpm/pkg/pathutil"
)

// CodeLocation represents where a code was found
type CodeLocation struct {
	Code       string
	EntityType string // "module", "chapter", "section", "assessment", "challenge", "goal", "mcq", "question"
	ModuleName string
	FilePath   string
}

// DuplicateCode represents a code that appears in multiple locations
type DuplicateCode struct {
	Code      string
	Locations []CodeLocation
}

// GlobalValidationResult contains the results of global validation across all modules
type GlobalValidationResult struct {
	Duplicates     []DuplicateCode
	TotalCodes     int
	ModulesScanned int
}

// HasDuplicates returns true if there are any duplicate codes
func (r *GlobalValidationResult) HasDuplicates() bool {
	return len(r.Duplicates) > 0
}

// ValidateAllModules validates uniqueness of codes across all modules
func ValidateAllModules(modules []*module.Module, moduleNames []string) *GlobalValidationResult {
	result := &GlobalValidationResult{
		Duplicates:     make([]DuplicateCode, 0),
		ModulesScanned: len(modules),
	}

	// Collect all codes with their locations
	codeMap := make(map[string][]CodeLocation)

	for i, mod := range modules {
		moduleName := moduleNames[i]
		collectCodes(mod, moduleName, codeMap)
	}

	result.TotalCodes = len(codeMap)

	// Find duplicates
	for code, locations := range codeMap {
		if len(locations) > 1 {
			result.Duplicates = append(result.Duplicates, DuplicateCode{
				Code:      code,
				Locations: locations,
			})
		}
	}

	return result
}

// collectCodes collects all codes from a module into the codeMap
func collectCodes(mod *module.Module, moduleName string, codeMap map[string][]CodeLocation) {
	// Module code
	if mod.Code != "" {
		codeMap[mod.Code] = append(codeMap[mod.Code], CodeLocation{
			Code:       mod.Code,
			EntityType: "module",
			ModuleName: moduleName,
			FilePath:   pathutil.GetRelativeModulePath(mod.FilePath),
		})
	}

	// Chapter codes
	for _, chapter := range mod.Chapters {
		if chapter.Code != "" {
			codeMap[chapter.Code] = append(codeMap[chapter.Code], CodeLocation{
				Code:       chapter.Code,
				EntityType: "chapter",
				ModuleName: moduleName,
				FilePath:   pathutil.GetRelativeModulePath(chapter.FilePath),
			})
		}

		// Section codes
		for _, section := range chapter.Sections {
			if section.Code != "" {
				codeMap[section.Code] = append(codeMap[section.Code], CodeLocation{
					Code:       section.Code,
					EntityType: "section",
					ModuleName: moduleName,
					FilePath:   pathutil.GetRelativeModulePath(section.FilePath),
				})
			}
		}

		// Interactive assessment codes
		for _, assessment := range chapter.Assessments {
			if assessment.Code != "" {
				codeMap[assessment.Code] = append(codeMap[assessment.Code], CodeLocation{
					Code:       assessment.Code,
					EntityType: "assessment",
					ModuleName: moduleName,
					FilePath:   pathutil.GetRelativeModulePath(assessment.FilePath),
				})
			}

			// Challenge codes
			for _, challenge := range assessment.Challenges {
				if challenge.Code != "" {
					codeMap[challenge.Code] = append(codeMap[challenge.Code], CodeLocation{
						Code:       challenge.Code,
						EntityType: "challenge",
						ModuleName: moduleName,
						FilePath:   pathutil.GetRelativeModulePath(challenge.FilePath),
					})
				}

				// Goal codes
				for i, goal := range challenge.Goals {
					if goal.Code != "" {
						codeMap[goal.Code] = append(codeMap[goal.Code], CodeLocation{
							Code:       goal.Code,
							EntityType: "goal",
							ModuleName: moduleName,
							FilePath:   fmt.Sprintf("%s (goal #%d)", pathutil.GetRelativeModulePath(challenge.FilePath), i+1),
						})
					}
				}
			}
		}

		// Multiple choice assessment codes
		for i, mcq := range chapter.MultipleChoiceAssessments {
			if mcq.Code != "" {
				codeMap[mcq.Code] = append(codeMap[mcq.Code], CodeLocation{
					Code:       mcq.Code,
					EntityType: "multiple-choice-assessment",
					ModuleName: moduleName,
					FilePath:   fmt.Sprintf("%s (MCQ #%d)", pathutil.GetRelativeModulePath(chapter.FilePath), i+1),
				})
			}

			// Question codes
			for j, question := range mcq.Questions {
				if question.Code != "" {
					codeMap[question.Code] = append(codeMap[question.Code], CodeLocation{
						Code:       question.Code,
						EntityType: "question",
						ModuleName: moduleName,
						FilePath:   fmt.Sprintf("%s (MCQ #%d, question #%d)", pathutil.GetRelativeModulePath(chapter.FilePath), i+1, j+1),
					})
				}
			}
		}
	}
}
