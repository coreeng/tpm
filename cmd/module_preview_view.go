package cmd

import (
	"github.com/coreeng/tpm/pkg/module"
)

type modulePreviewPage struct {
	*module.Module
	Source            string
	DescriptionSource string
	Chapters          []modulePreviewChapter
}

type modulePreviewChapter struct {
	module.Chapter
	Source                    string
	DescriptionSource         string
	Sections                  []modulePreviewSection
	Assessments               []modulePreviewAssessment
	MultipleChoiceAssessments []modulePreviewQuiz
}

type modulePreviewSection struct {
	module.Section
	Source            string
	DescriptionSource string
}

type modulePreviewAssessment struct {
	module.Assessment
	Source            string
	DescriptionSource string
	Challenges        []modulePreviewChallenge
}

type modulePreviewChallenge struct {
	module.Challenge
	Source               string
	DescriptionSource    string
	SuccessMessageSource string
	Goals                []modulePreviewGoal
}

type modulePreviewGoal struct {
	module.Goal
	Source string
}

type modulePreviewQuiz struct {
	module.MultipleChoiceAssessment
	Source    string
	Questions []modulePreviewQuestion
}

type modulePreviewQuestion struct {
	Code    string
	Index   int
	Text    string
	Type    string
	Source  string
	Options []modulePreviewOption
}

type modulePreviewOption struct {
	module.Option
	Source string
}

func newModulePreviewPage(mod *module.Module) *modulePreviewPage {
	if mod == nil {
		return nil
	}
	page := &modulePreviewPage{
		Module:            mod,
		Source:            previewSourceLabel(mod.FilePath),
		DescriptionSource: previewSourceLabel(siblingSource(mod.FilePath, "description.md")),
	}
	for _, chapter := range mod.Chapters {
		previewChapter := modulePreviewChapter{
			Chapter:           chapter,
			Source:            previewSourceLabel(chapter.FilePath),
			DescriptionSource: previewSourceLabel(siblingSource(chapter.FilePath, "description.md")),
		}
		for _, section := range chapter.Sections {
			previewChapter.Sections = append(previewChapter.Sections, modulePreviewSection{
				Section:           section,
				Source:            previewSourceLabel(section.FilePath),
				DescriptionSource: previewSourceLabel(siblingSource(section.FilePath, "description.md")),
			})
		}
		for _, assessment := range chapter.Assessments {
			previewAssessment := modulePreviewAssessment{
				Assessment:        assessment,
				Source:            previewSourceLabel(assessment.FilePath),
				DescriptionSource: previewSourceLabel(siblingSource(assessment.FilePath, "description.md")),
				Challenges:        make([]modulePreviewChallenge, 0, len(assessment.Challenges)),
			}
			for _, challenge := range assessment.Challenges {
				previewChallenge := modulePreviewChallenge{
					Challenge:            challenge,
					Source:               previewSourceLabel(challenge.FilePath),
					DescriptionSource:    previewSourceLabel(siblingSource(challenge.FilePath, "description.md")),
					SuccessMessageSource: previewSourceLabel(siblingSource(challenge.FilePath, "successMessage.md")),
					Goals:                make([]modulePreviewGoal, 0, len(challenge.Goals)),
				}
				for _, goal := range challenge.Goals {
					previewChallenge.Goals = append(previewChallenge.Goals, modulePreviewGoal{
						Goal:   goal,
						Source: previewSourceLabel(challenge.FilePath),
					})
				}
				previewAssessment.Challenges = append(previewAssessment.Challenges, previewChallenge)
			}
			previewChapter.Assessments = append(previewChapter.Assessments, previewAssessment)
		}
		for _, quiz := range chapter.MultipleChoiceAssessments {
			previewQuiz := modulePreviewQuiz{
				MultipleChoiceAssessment: quiz,
				Source:                   previewSourceLabel(chapter.FilePath),
				Questions:                make([]modulePreviewQuestion, 0, len(quiz.Questions)),
			}
			for _, question := range quiz.Questions {
				previewQuestion := modulePreviewQuestion{
					Code:    question.Code,
					Index:   question.Index,
					Text:    question.Question,
					Type:    question.Type,
					Source:  previewSourceLabel(chapter.FilePath),
					Options: make([]modulePreviewOption, 0, len(question.Options)),
				}
				for _, option := range question.Options {
					previewQuestion.Options = append(previewQuestion.Options, modulePreviewOption{
						Option: option,
						Source: previewSourceLabel(chapter.FilePath),
					})
				}
				previewQuiz.Questions = append(previewQuiz.Questions, previewQuestion)
			}
			previewChapter.MultipleChoiceAssessments = append(previewChapter.MultipleChoiceAssessments, previewQuiz)
		}
		page.Chapters = append(page.Chapters, previewChapter)
	}
	return page
}
