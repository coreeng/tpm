package module

// BuiltModule is the compiled module artifact shape. It includes markdown
// content and generated indexes that are not valid in source YAML files.
type BuiltModule struct {
	Code             string         `yaml:"code" json:"code"`
	Title            string         `yaml:"title" json:"title"`
	Description      string         `yaml:"description" json:"description"`
	ShortDescription string         `yaml:"shortDescription" json:"shortDescription"`
	BannerImage      string         `yaml:"bannerImage,omitempty" json:"bannerImage,omitempty"`
	BannerVideo      string         `yaml:"bannerVideo,omitempty" json:"bannerVideo,omitempty"`
	Tags             []string       `yaml:"tags,omitempty" json:"tags,omitempty"`
	Level            string         `yaml:"level" json:"level"`
	Chapters         []BuiltChapter `yaml:"chapters" json:"chapters"`
}

type BuiltChapter struct {
	Code                      string                          `yaml:"code" json:"code"`
	Index                     int                             `yaml:"index" json:"index"`
	Title                     string                          `yaml:"title" json:"title"`
	Description               string                          `yaml:"description" json:"description"`
	ShortDescription          string                          `yaml:"shortDescription" json:"shortDescription,omitempty"`
	BannerImage               string                          `yaml:"bannerImage" json:"bannerImage,omitempty"`
	BannerVideo               string                          `yaml:"bannerVideo" json:"bannerVideo,omitempty"`
	IsDraft                   bool                            `yaml:"isDraft" json:"isDraft"`
	Sections                  []BuiltSection                  `yaml:"sections" json:"sections,omitempty"`
	Assessments               []BuiltAssessment               `yaml:"assessments" json:"assessments,omitempty"`
	MultipleChoiceAssessments []BuiltMultipleChoiceAssessment `yaml:"multipleChoiceAssessments" json:"multipleChoiceAssessments,omitempty"`
}

type BuiltSection struct {
	Code                 string `yaml:"code" json:"code"`
	Index                int    `yaml:"index" json:"index"`
	Title                string `yaml:"title" json:"title"`
	Description          string `yaml:"description" json:"description"`
	ShortDescription     string `yaml:"shortDescription" json:"shortDescription,omitempty"`
	ThumbnailDescription string `yaml:"thumbnailDescription,omitempty" json:"thumbnailDescription,omitempty"`
	Thumbnail            string `yaml:"thumbnail,omitempty" json:"thumbnail,omitempty"`
	Video                string `yaml:"video,omitempty" json:"video,omitempty"`
	EstimatedDuration    string `yaml:"estimatedDuration,omitempty" json:"estimatedDuration,omitempty"`
}

type BuiltAssessment struct {
	Code              string           `yaml:"code" json:"code"`
	Index             int              `yaml:"index" json:"index"`
	Title             string           `yaml:"title" json:"title"`
	Description       string           `yaml:"description" json:"description"`
	TimeLimit         string           `yaml:"timeLimit" json:"timeLimit,omitempty"`
	StarterImageURI   string           `yaml:"starterImageUri" json:"starterImageUri"`
	ValidatorImageURI string           `yaml:"validatorImageUri" json:"validatorImageUri"`
	ImageVersion      string           `yaml:"imageVersion" json:"imageVersion"`
	Video             string           `yaml:"video" json:"video,omitempty"`
	Challenges        []BuiltChallenge `yaml:"challenges" json:"challenges"`
}

type BuiltChallenge struct {
	Code              string      `yaml:"code" json:"code"`
	Index             int         `yaml:"index" json:"index"`
	Title             string      `yaml:"title" json:"title"`
	Description       string      `yaml:"description" json:"description"`
	SuccessMessage    string      `yaml:"successMessage" json:"successMessage"`
	EstimatedDuration string      `yaml:"estimatedDuration" json:"estimatedDuration,omitempty"`
	Video             string      `yaml:"video" json:"video,omitempty"`
	Goals             []BuiltGoal `yaml:"goals" json:"goals"`
}

type BuiltGoal struct {
	Code        string `yaml:"code" json:"code"`
	Index       int    `yaml:"index" json:"index"`
	Title       string `yaml:"title" json:"title"`
	Description string `yaml:"description" json:"description"`
}

type BuiltMultipleChoiceAssessment struct {
	Code         string          `yaml:"code" json:"code"`
	Index        int             `yaml:"index" json:"index"`
	Title        string          `yaml:"title" json:"title"`
	Description  string          `yaml:"description" json:"description"`
	PassingScore int             `yaml:"passingScore" json:"passingScore"`
	Questions    []BuiltQuestion `yaml:"questions" json:"questions"`
}

type BuiltQuestion struct {
	Code     string        `yaml:"code" json:"code"`
	Index    int           `yaml:"index" json:"index"`
	Question string        `yaml:"question" json:"question"`
	Type     string        `yaml:"type" json:"type"`
	Options  []BuiltOption `yaml:"options" json:"options"`
}

type BuiltOption struct {
	Index   int    `yaml:"index" json:"index"`
	Text    string `yaml:"text" json:"text"`
	Correct bool   `yaml:"correct" json:"correct"`
}
