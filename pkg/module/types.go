package module

// Module represents a training platform module
type Module struct {
	Code             string    `yaml:"code" json:"code"`
	Title            string    `yaml:"title" json:"title"`
	Description      string    `yaml:"-" json:"-"`
	ShortDescription string    `yaml:"shortDescription" json:"shortDescription"`
	BannerImage      string    `yaml:"bannerImage,omitempty" json:"bannerImage,omitempty"`
	BannerVideo      string    `yaml:"bannerVideo,omitempty" json:"bannerVideo,omitempty"`
	Tags             []string  `yaml:"tags,omitempty" json:"tags,omitempty"`
	Level            string    `yaml:"level" json:"level"`
	Chapters         []Chapter `yaml:"chapters" json:"chapters"`
	FilePath         string    `yaml:"-" json:"-"` // Path to module.yaml
}

// Chapter represents a chapter within a module
type Chapter struct {
	Code                      string                     `yaml:"code" json:"code"`
	Index                     int                        `yaml:"-" json:"-"`
	Title                     string                     `yaml:"title" json:"title"`
	Description               string                     `yaml:"-" json:"-"`
	IsDraft                   bool                       `yaml:"isDraft" json:"isDraft"`
	Sections                  []Section                  `yaml:"sections" json:"sections,omitempty"`
	Assessments               []Assessment               `yaml:"assessments" json:"assessments,omitempty"`
	MultipleChoiceAssessments []MultipleChoiceAssessment `yaml:"multipleChoiceAssessments" json:"multipleChoiceAssessments,omitempty"`
	FilePath                  string                     `yaml:"-" json:"-"` // Path to chapter.yml
}

// Section represents a learning section within a chapter
type Section struct {
	Code              string   `yaml:"code" json:"code"`
	Index             int      `yaml:"-" json:"-"`
	Title             string   `yaml:"title" json:"title"`
	Description       string   `yaml:"-" json:"-"`
	ShortDescription  string   `yaml:"shortDescription" json:"shortDescription,omitempty"`
	Video             string   `yaml:"video,omitempty" json:"video,omitempty"`
	EstimatedDuration string   `yaml:"estimatedDuration,omitempty" json:"estimatedDuration,omitempty"`
	Prerequisites     []string `yaml:"prerequisites,omitempty" json:"prerequisites,omitempty"`
	FilePath          string   `yaml:"-" json:"-"` // Path to section.yaml
}

// Assessment represents an interactive assessment
type Assessment struct {
	Code              string      `yaml:"code" json:"code"`
	Index             int         `yaml:"-" json:"-"`
	Title             string      `yaml:"title" json:"title"`
	Description       string      `yaml:"-" json:"-"`
	TimeLimit         string      `yaml:"timeLimit" json:"timeLimit,omitempty"`
	StarterImageURI   string      `yaml:"starterImageUri" json:"starterImageUri"`
	ValidatorImageURI string      `yaml:"validatorImageUri" json:"validatorImageUri"`
	ImageVersion      string      `yaml:"imageVersion" json:"imageVersion"`
	Video             string      `yaml:"video" json:"video,omitempty"`
	Challenges        []Challenge `yaml:"challenges" json:"challenges"`
	FilePath          string      `yaml:"-" json:"-"` // Path to assessment.yaml
}

// Challenge represents a challenge within an assessment
type Challenge struct {
	Code              string `yaml:"code" json:"code"`
	Index             int    `yaml:"-" json:"-"`
	Title             string `yaml:"title" json:"title"`
	Description       string `yaml:"-" json:"-"`
	SuccessMessage    string `yaml:"-" json:"-"`
	EstimatedDuration string `yaml:"estimatedDuration" json:"estimatedDuration,omitempty"`
	Video             string `yaml:"video" json:"video,omitempty"`
	Goals             []Goal `yaml:"goals" json:"goals"`
	FilePath          string `yaml:"-" json:"-"` // Path to challenge.yaml
}

// Goal represents a goal within a challenge
type Goal struct {
	Code        string `yaml:"code" json:"code"`
	Index       int    `yaml:"-" json:"-"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
}

// MultipleChoiceAssessment represents a quiz-based assessment
type MultipleChoiceAssessment struct {
	Code         string     `yaml:"code" json:"code"`
	Index        int        `yaml:"index" json:"index"`
	Title        string     `yaml:"title" json:"title"`
	Description  string     `yaml:"description" json:"description"`
	PassingScore int        `yaml:"passingScore" json:"passingScore"`
	Questions    []Question `yaml:"questions" json:"questions"`
}

// Question represents a quiz question
type Question struct {
	Code     string   `yaml:"code" json:"code"`
	Index    int      `yaml:"index" json:"index"`
	Question string   `yaml:"question" json:"question"`
	Type     string   `yaml:"type" json:"type"` // SINGLE or MULTIPLE
	Options  []Option `yaml:"options" json:"options"`
}

// Option represents an answer option for a quiz question
type Option struct {
	Index   int    `yaml:"index" json:"index"`
	Text    string `yaml:"text" json:"text"`
	Correct bool   `yaml:"correct" json:"correct"`
}
