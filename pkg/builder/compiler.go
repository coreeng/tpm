package builder

import "github.com/coreeng/tpm/pkg/module"

func toBuiltModule(mod *module.Module) *module.BuiltModule {
	built := &module.BuiltModule{
		Code:             mod.Code,
		Title:            mod.Title,
		Description:      mod.Description,
		ShortDescription: mod.ShortDescription,
		BannerImage:      mod.BannerImage,
		BannerVideo:      mod.BannerVideo,
		Tags:             append([]string(nil), mod.Tags...),
		Level:            mod.Level,
		Chapters:         make([]module.BuiltChapter, 0, len(mod.Chapters)),
	}

	for _, chapter := range mod.Chapters {
		built.Chapters = append(built.Chapters, toBuiltChapter(chapter))
	}
	return built
}

func toBuiltChapter(chapter module.Chapter) module.BuiltChapter {
	built := module.BuiltChapter{
		Code:                      chapter.Code,
		Index:                     chapter.Index,
		Title:                     chapter.Title,
		Description:               chapter.Description,
		IsDraft:                   chapter.IsDraft,
		Sections:                  make([]module.BuiltSection, 0, len(chapter.Sections)),
		Assessments:               make([]module.BuiltAssessment, 0, len(chapter.Assessments)),
		MultipleChoiceAssessments: make([]module.BuiltMultipleChoiceAssessment, 0, len(chapter.MultipleChoiceAssessments)),
	}
	for _, section := range chapter.Sections {
		built.Sections = append(built.Sections, module.BuiltSection{
			Code:              section.Code,
			Index:             section.Index,
			Title:             section.Title,
			Description:       section.Description,
			ShortDescription:  section.ShortDescription,
			Video:             section.Video,
			EstimatedDuration: section.EstimatedDuration,
			Prerequisites:     append([]string(nil), section.Prerequisites...),
		})
	}
	for _, assessment := range chapter.Assessments {
		built.Assessments = append(built.Assessments, toBuiltAssessment(assessment))
	}
	for _, quiz := range chapter.MultipleChoiceAssessments {
		built.MultipleChoiceAssessments = append(built.MultipleChoiceAssessments, toBuiltQuiz(quiz))
	}
	return built
}

func toBuiltAssessment(assessment module.Assessment) module.BuiltAssessment {
	built := module.BuiltAssessment{
		Code:              assessment.Code,
		Index:             assessment.Index,
		Title:             assessment.Title,
		Description:       assessment.Description,
		TimeLimit:         assessment.TimeLimit,
		StarterImageURI:   assessment.StarterImageURI,
		ValidatorImageURI: assessment.ValidatorImageURI,
		ImageVersion:      assessment.ImageVersion,
		Video:             assessment.Video,
		Challenges:        make([]module.BuiltChallenge, 0, len(assessment.Challenges)),
	}
	for _, challenge := range assessment.Challenges {
		built.Challenges = append(built.Challenges, toBuiltChallenge(challenge))
	}
	return built
}

func toBuiltChallenge(challenge module.Challenge) module.BuiltChallenge {
	built := module.BuiltChallenge{
		Code:              challenge.Code,
		Index:             challenge.Index,
		Title:             challenge.Title,
		Description:       challenge.Description,
		SuccessMessage:    challenge.SuccessMessage,
		EstimatedDuration: challenge.EstimatedDuration,
		Video:             challenge.Video,
		Goals:             make([]module.BuiltGoal, 0, len(challenge.Goals)),
	}
	for _, goal := range challenge.Goals {
		built.Goals = append(built.Goals, module.BuiltGoal(goal))
	}
	return built
}

func toBuiltQuiz(quiz module.MultipleChoiceAssessment) module.BuiltMultipleChoiceAssessment {
	built := module.BuiltMultipleChoiceAssessment{
		Code:         quiz.Code,
		Index:        quiz.Index,
		Title:        quiz.Title,
		Description:  quiz.Description,
		PassingScore: quiz.PassingScore,
		Questions:    make([]module.BuiltQuestion, 0, len(quiz.Questions)),
	}
	for _, question := range quiz.Questions {
		built.Questions = append(built.Questions, toBuiltQuestion(question))
	}
	return built
}

func toBuiltQuestion(question module.Question) module.BuiltQuestion {
	built := module.BuiltQuestion{
		Code:     question.Code,
		Index:    question.Index,
		Question: question.Question,
		Type:     question.Type,
		Options:  make([]module.BuiltOption, 0, len(question.Options)),
	}
	for _, option := range question.Options {
		built.Options = append(built.Options, module.BuiltOption(option))
	}
	return built
}
