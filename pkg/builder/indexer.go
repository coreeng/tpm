package builder

import (
	"github.com/coreeng/tpm/pkg/module"
)

// assignIndices sets sequential indices on all entities
// Preserves existing non-zero indices, assigns 1-based indices to entities with index 0
// Modifies the module in place
func assignIndices(mod *module.Module) {
	// Chapters
	for i := range mod.Chapters {
		ch := &mod.Chapters[i]
		if ch.Index == 0 {
			ch.Index = i + 1 // 1-based indexing
		}

		// Sections
		for j := range ch.Sections {
			sec := &ch.Sections[j]
			if sec.Index == 0 {
				sec.Index = j + 1 // 1-based indexing
			}
		}

		// Interactive Assessments
		for j := range ch.Assessments {
			assessment := &ch.Assessments[j]
			if assessment.Index == 0 {
				assessment.Index = j + 1 // 1-based indexing
			}

			// Challenges
			for k := range assessment.Challenges {
				challenge := &assessment.Challenges[k]
				if challenge.Index == 0 {
					challenge.Index = k + 1 // 1-based indexing
				}

				// Goals - keep 0-based for inline arrays
				for l := range challenge.Goals {
					goal := &challenge.Goals[l]
					goal.Index = l // Goals remain 0-based (no directories)
				}
			}
		}

		// Multiple Choice Assessments - keep 0-based for inline arrays
		for j := range ch.MultipleChoiceAssessments {
			mca := &ch.MultipleChoiceAssessments[j]
			mca.Index = j // Quizzes remain 0-based (no directories)

			// Questions
			for k := range mca.Questions {
				q := &mca.Questions[k]
				q.Index = k // Questions remain 0-based

				// Options
				for l := range q.Options {
					opt := &q.Options[l]
					opt.Index = l // Options remain 0-based
				}
			}
		}
	}
}
