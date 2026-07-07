package lab

type Format string

const (
	FormatStandalone   Format = "standalone"
	FormatModuleBacked Format = "module-backed"
)

type Lab struct {
	Format        Format      `yaml:"-"`
	RootPath      string      `yaml:"-"`
	RuntimePath   string      `yaml:"-"`
	MetadataPath  string      `yaml:"-"`
	Title         string      `yaml:"title"`
	Code          string      `yaml:"code"`
	Description   string      `yaml:"description"`
	TimeLimit     string      `yaml:"timeLimit"`
	StarterPath   string      `yaml:"-"`
	SolutionPath  string      `yaml:"-"`
	ValidatorPath string      `yaml:"-"`
	Challenges    []Challenge `yaml:"challenges"`
}

type Challenge struct {
	Code           string `yaml:"code"`
	Title          string `yaml:"title"`
	Description    string `yaml:"description"`
	SuccessMessage string `yaml:"successMessage"`
	Goals          []Goal `yaml:"goals"`
}

type Goal struct {
	Code        string `yaml:"code"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
}
