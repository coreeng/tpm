package cmd

import "github.com/coreeng/tpm/pkg/module"

type modulePreviewPage struct {
	Kind             string                 `json:"kind"`
	Code             previewText            `json:"code"`
	Title            previewText            `json:"title"`
	Description      previewText            `json:"description"`
	ShortDescription previewText            `json:"shortDescription"`
	Level            previewText            `json:"level"`
	BannerImage      previewText            `json:"bannerImage"`
	BannerVideo      previewText            `json:"bannerVideo"`
	Stats            modulePreviewStats     `json:"stats"`
	Chapters         []modulePreviewChapter `json:"chapters"`
}

type modulePreviewStats struct {
	Chapters int `json:"chapters"`
	Sections int `json:"sections"`
	Quizzes  int `json:"quizzes"`
	Labs     int `json:"labs"`
}

type modulePreviewChapter struct {
	Code                      previewText         `json:"code"`
	Index                     int                 `json:"index"`
	Title                     previewText         `json:"title"`
	Description               previewText         `json:"description"`
	ShortDescription          previewText         `json:"shortDescription"`
	BannerImage               previewText         `json:"bannerImage"`
	BannerVideo               previewText         `json:"bannerVideo"`
	IsDraft                   bool                `json:"isDraft"`
	Sections                  []modulePreviewItem `json:"sections"`
	Labs                      []modulePreviewLab  `json:"labs"`
	MultipleChoiceAssessments []modulePreviewQuiz `json:"multipleChoiceAssessments"`
}

type modulePreviewItem struct {
	Code                 previewText `json:"code"`
	Index                int         `json:"index"`
	Title                previewText `json:"title"`
	Description          previewText `json:"description"`
	ShortDescription     previewText `json:"shortDescription"`
	ThumbnailDescription previewText `json:"thumbnailDescription"`
	Thumbnail            previewText `json:"thumbnail"`
	Video                previewText `json:"video"`
	EstimatedDuration    previewText `json:"estimatedDuration"`
}

type modulePreviewLab struct {
	Code              previewText              `json:"code"`
	Index             int                      `json:"index"`
	Title             previewText              `json:"title"`
	Description       previewText              `json:"description"`
	TimeLimit         previewText              `json:"timeLimit"`
	StarterImageURI   previewText              `json:"starterImageUri"`
	ValidatorImageURI previewText              `json:"validatorImageUri"`
	ImageVersion      previewText              `json:"imageVersion"`
	Video             previewText              `json:"video"`
	Challenges        []modulePreviewChallenge `json:"challenges"`
}

type modulePreviewChallenge struct {
	Code              previewText         `json:"code"`
	Index             int                 `json:"index"`
	Title             previewText         `json:"title"`
	Description       previewText         `json:"description"`
	SuccessMessage    previewText         `json:"successMessage"`
	EstimatedDuration previewText         `json:"estimatedDuration"`
	Video             previewText         `json:"video"`
	Goals             []modulePreviewGoal `json:"goals"`
}

type modulePreviewGoal struct {
	Code        previewText `json:"code"`
	Index       int         `json:"index"`
	Title       previewText `json:"title"`
	Description previewText `json:"description"`
}

type modulePreviewQuiz struct {
	Code         previewText             `json:"code"`
	Index        int                     `json:"index"`
	Title        previewText             `json:"title"`
	Description  previewText             `json:"description"`
	PassingScore previewNumber           `json:"passingScore"`
	Questions    []modulePreviewQuestion `json:"questions"`
}

type modulePreviewQuestion struct {
	Code     previewText           `json:"code"`
	Index    int                   `json:"index"`
	Question previewText           `json:"question"`
	Type     previewText           `json:"type"`
	Options  []modulePreviewOption `json:"options"`
}

type modulePreviewOption struct {
	Index   int         `json:"index"`
	Text    previewText `json:"text"`
	Correct bool        `json:"correct"`
}

func newModulePreviewPage(mod *module.Module) *modulePreviewPage {
	if mod == nil {
		return nil
	}
	page := &modulePreviewPage{
		Kind:             "module",
		Code:             sourcedText(mod.Code, mod.FilePath, "code"),
		Title:            sourcedText(mod.Title, mod.FilePath, "title"),
		Description:      sourcedText(mod.Description, siblingSource(mod.FilePath, "description.md"), ""),
		ShortDescription: sourcedText(mod.ShortDescription, mod.FilePath, "shortDescription"),
		Level:            sourcedText(mod.Level, mod.FilePath, "level"),
		BannerImage:      sourcedText(mod.BannerImage, mod.FilePath, "bannerImage"),
		BannerVideo:      sourcedText(mod.BannerVideo, mod.FilePath, "bannerVideo"),
		Chapters:         make([]modulePreviewChapter, 0, len(mod.Chapters)),
	}
	for _, chapter := range mod.Chapters {
		previewChapter := modulePreviewChapter{
			Code:             sourcedText(chapter.Code, chapter.FilePath, "code"),
			Index:            chapter.Index,
			Title:            sourcedText(chapter.Title, chapter.FilePath, "title"),
			Description:      sourcedText(chapter.Description, siblingSource(chapter.FilePath, "description.md"), ""),
			ShortDescription: sourcedText(chapter.ShortDescription, chapter.FilePath, "shortDescription"),
			BannerImage:      sourcedText(chapter.BannerImage, chapter.FilePath, "bannerImage"),
			BannerVideo:      sourcedText(chapter.BannerVideo, chapter.FilePath, "bannerVideo"),
			IsDraft:          chapter.IsDraft,
			Sections:         make([]modulePreviewItem, 0, len(chapter.Sections)),
			Labs:             make([]modulePreviewLab, 0, len(chapter.Assessments)),
			MultipleChoiceAssessments: make(
				[]modulePreviewQuiz,
				0,
				len(chapter.MultipleChoiceAssessments),
			),
		}
		for _, section := range chapter.Sections {
			previewChapter.Sections = append(previewChapter.Sections, modulePreviewItem{
				Code:                 sourcedText(section.Code, section.FilePath, "code"),
				Index:                section.Index,
				Title:                sourcedText(section.Title, section.FilePath, "title"),
				Description:          sourcedText(section.Description, siblingSource(section.FilePath, "description.md"), ""),
				ShortDescription:     sourcedText(section.ShortDescription, section.FilePath, "shortDescription"),
				ThumbnailDescription: sourcedText(section.ThumbnailDescription, section.FilePath, "thumbnailDescription"),
				Thumbnail:            sourcedText(section.Thumbnail, section.FilePath, "thumbnail"),
				Video:                sourcedText(section.Video, section.FilePath, "video"),
				EstimatedDuration:    sourcedText(section.EstimatedDuration, section.FilePath, "estimatedDuration"),
			})
		}
		for _, assessment := range chapter.Assessments {
			previewLab := modulePreviewLab{
				Code:              sourcedText(assessment.Code, assessment.FilePath, "code"),
				Index:             assessment.Index,
				Title:             sourcedText(assessment.Title, assessment.FilePath, "title"),
				Description:       sourcedText(assessment.Description, siblingSource(assessment.FilePath, "description.md"), ""),
				TimeLimit:         sourcedText(assessment.TimeLimit, assessment.FilePath, "timeLimit"),
				StarterImageURI:   sourcedText(assessment.StarterImageURI, assessment.FilePath, "starterImageUri"),
				ValidatorImageURI: sourcedText(assessment.ValidatorImageURI, assessment.FilePath, "validatorImageUri"),
				ImageVersion:      sourcedText(assessment.ImageVersion, assessment.FilePath, "imageVersion"),
				Video:             sourcedText(assessment.Video, assessment.FilePath, "video"),
				Challenges:        make([]modulePreviewChallenge, 0, len(assessment.Challenges)),
			}
			for _, challenge := range assessment.Challenges {
				previewChallenge := modulePreviewChallenge{
					Code:              sourcedText(challenge.Code, challenge.FilePath, "code"),
					Index:             challenge.Index,
					Title:             sourcedText(challenge.Title, challenge.FilePath, "title"),
					Description:       sourcedText(challenge.Description, siblingSource(challenge.FilePath, "description.md"), ""),
					SuccessMessage:    sourcedText(challenge.SuccessMessage, siblingSource(challenge.FilePath, "successMessage.md"), ""),
					EstimatedDuration: sourcedText(challenge.EstimatedDuration, challenge.FilePath, "estimatedDuration"),
					Video:             sourcedText(challenge.Video, challenge.FilePath, "video"),
					Goals:             make([]modulePreviewGoal, 0, len(challenge.Goals)),
				}
				for _, goal := range challenge.Goals {
					previewChallenge.Goals = append(previewChallenge.Goals, modulePreviewGoal{
						Code:        sourcedText(goal.Code, challenge.FilePath, "goals[].code"),
						Index:       goal.Index,
						Title:       sourcedText(goal.Title, challenge.FilePath, "goals[].title"),
						Description: sourcedText(goal.Description, challenge.FilePath, "goals[].description"),
					})
				}
				previewLab.Challenges = append(previewLab.Challenges, previewChallenge)
			}
			previewChapter.Labs = append(previewChapter.Labs, previewLab)
		}
		for _, quiz := range chapter.MultipleChoiceAssessments {
			previewQuiz := modulePreviewQuiz{
				Code:         sourcedText(quiz.Code, chapter.FilePath, "multipleChoiceAssessments[].code"),
				Index:        quiz.Index,
				Title:        sourcedText(quiz.Title, chapter.FilePath, "multipleChoiceAssessments[].title"),
				Description:  sourcedText(quiz.Description, chapter.FilePath, "multipleChoiceAssessments[].description"),
				PassingScore: sourcedNumber(quiz.PassingScore, chapter.FilePath, "multipleChoiceAssessments[].passingScore"),
				Questions:    make([]modulePreviewQuestion, 0, len(quiz.Questions)),
			}
			for _, question := range quiz.Questions {
				previewQuestion := modulePreviewQuestion{
					Code:     sourcedText(question.Code, chapter.FilePath, "multipleChoiceAssessments[].questions[].code"),
					Index:    question.Index,
					Question: sourcedText(question.Question, chapter.FilePath, "multipleChoiceAssessments[].questions[].question"),
					Type:     sourcedText(question.Type, chapter.FilePath, "multipleChoiceAssessments[].questions[].type"),
					Options:  make([]modulePreviewOption, 0, len(question.Options)),
				}
				for _, option := range question.Options {
					previewQuestion.Options = append(previewQuestion.Options, modulePreviewOption{
						Index:   option.Index,
						Text:    sourcedText(option.Text, chapter.FilePath, "multipleChoiceAssessments[].questions[].options[].text"),
						Correct: option.Correct,
					})
				}
				previewQuiz.Questions = append(previewQuiz.Questions, previewQuestion)
			}
			previewChapter.MultipleChoiceAssessments = append(previewChapter.MultipleChoiceAssessments, previewQuiz)
		}
		page.Stats.Chapters++
		page.Stats.Sections += len(previewChapter.Sections)
		page.Stats.Labs += len(previewChapter.Labs)
		page.Stats.Quizzes += len(previewChapter.MultipleChoiceAssessments)
		page.Chapters = append(page.Chapters, previewChapter)
	}
	return page
}
