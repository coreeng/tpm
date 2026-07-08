package cmd

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/coreeng/tpm/pkg/lab"
)

type labPreviewPage struct {
	Kind        string                `json:"kind"`
	Format      string                `json:"format"`
	Code        previewText           `json:"code"`
	Title       previewText           `json:"title"`
	Description previewText           `json:"description"`
	TimeLimit   previewText           `json:"timeLimit"`
	Progress    labPreviewProgress    `json:"progress"`
	StatusError string                `json:"statusError,omitempty"`
	Runtime     *labPreviewRuntime    `json:"runtime,omitempty"`
	Challenges  []labPreviewChallenge `json:"challenges"`
}

type labPreviewRuntime struct {
	RunID              string `json:"runId"`
	SystemNamespace    string `json:"systemNamespace"`
	WorkspaceNamespace string `json:"workspaceNamespace"`
	RegistryURL        string `json:"registryUrl"`
	RegistryUsername   string `json:"registryUsername"`
}

type labPreviewChallenge struct {
	Code           previewText        `json:"code"`
	Title          previewText        `json:"title"`
	Description    previewText        `json:"description"`
	SuccessMessage previewText        `json:"successMessage"`
	Progress       labPreviewProgress `json:"progress"`
	Goals          []labPreviewGoal   `json:"goals"`
}

type labPreviewGoal struct {
	Code        previewText        `json:"code"`
	Title       previewText        `json:"title"`
	Description previewText        `json:"description"`
	Progress    labPreviewProgress `json:"progress"`
}

type labPreviewProgress struct {
	ConditionType string `json:"conditionType"`
	State         string `json:"state"`
	Label         string `json:"label"`
	Status        string `json:"status,omitempty"`
	Reason        string `json:"reason,omitempty"`
	Message       string `json:"message,omitempty"`
}

type labPreviewChallengeSource struct {
	dirName            string
	yamlPath           string
	descriptionPath    string
	successMessagePath string
}

func newLabPreviewPage(loaded *lab.Lab, state *lab.RunState, conditions []lab.ProgressCondition, statusErr error) labPreviewPage {
	page := labPreviewPage{Kind: "lab"}
	if loaded == nil {
		return page
	}
	conditionMap := labPreviewConditionMap(conditions)

	titleSource := labPreviewSourcePath(loaded, loaded.MetadataPath)
	descriptionPath := loaded.MetadataPath
	descriptionProperty := "description"
	if loaded.Format == lab.FormatModuleBacked {
		descriptionPath = existingSourceOrFallback(filepath.Join(loaded.RootPath, "description.md"), loaded.MetadataPath)
		descriptionProperty = ""
	}

	page.Format = string(loaded.Format)
	page.Code = sourcedText(loaded.Code, titleSource, "code")
	page.Title = sourcedText(loaded.Title, titleSource, "title")
	page.Description = sourcedText(loaded.Description, labPreviewSourcePath(loaded, descriptionPath), descriptionProperty)
	page.TimeLimit = sourcedText(loaded.TimeLimit, titleSource, "timeLimit")
	page.Progress = labPreviewProgressFor(conditionMap, "IA_Completed", state != nil)
	if statusErr != nil {
		page.StatusError = statusErr.Error()
		page.Progress = labPreviewProgressUnknown("IA_Completed", "Status unavailable", statusErr.Error())
	}
	page.Challenges = make([]labPreviewChallenge, 0, len(loaded.Challenges))
	if state != nil {
		page.Runtime = &labPreviewRuntime{
			RunID:              state.RunID,
			SystemNamespace:    state.SystemNamespace,
			WorkspaceNamespace: state.WorkspaceNamespace,
			RegistryURL:        state.RegistryURL,
			RegistryUsername:   state.RegistryUsername,
		}
	}

	challengeSources := labPreviewChallengeSources(loaded)
	for i, challenge := range loaded.Challenges {
		yamlPath := loaded.MetadataPath
		descriptionPath := loaded.MetadataPath
		successMessagePath := loaded.MetadataPath
		challengePropertyPrefix := "challenges[]."
		goalPropertyPrefix := "challenges[].goals[]."
		if i < len(challengeSources) {
			yamlPath = challengeSources[i].yamlPath
			descriptionPath = existingSourceOrFallback(challengeSources[i].descriptionPath, yamlPath)
			successMessagePath = existingSourceOrFallback(challengeSources[i].successMessagePath, yamlPath)
			challengePropertyPrefix = ""
			goalPropertyPrefix = "goals[]."
		}
		yamlSource := labPreviewSourcePath(loaded, yamlPath)
		previewChallenge := labPreviewChallenge{
			Code:           sourcedText(challenge.Code, yamlSource, challengePropertyPrefix+"code"),
			Title:          sourcedText(challenge.Title, yamlSource, challengePropertyPrefix+"title"),
			Description:    sourcedText(challenge.Description, labPreviewSourcePath(loaded, descriptionPath), challengePropertyPrefix+"description"),
			SuccessMessage: sourcedText(challenge.SuccessMessage, labPreviewSourcePath(loaded, successMessagePath), challengePropertyPrefix+"successMessage"),
			Progress:       labPreviewProgressFor(conditionMap, "IAC_"+challenge.Code, state != nil),
			Goals:          make([]labPreviewGoal, 0, len(challenge.Goals)),
		}
		for _, goal := range challenge.Goals {
			previewChallenge.Goals = append(previewChallenge.Goals, labPreviewGoal{
				Code:        sourcedText(goal.Code, yamlSource, goalPropertyPrefix+"code"),
				Title:       sourcedText(goal.Title, yamlSource, goalPropertyPrefix+"title"),
				Description: sourcedText(goal.Description, yamlSource, goalPropertyPrefix+"description"),
				Progress:    labPreviewProgressFor(conditionMap, "IAG_"+challenge.Code+"_"+goal.Code, state != nil),
			})
		}
		page.Challenges = append(page.Challenges, previewChallenge)
	}
	return page
}

func labPreviewConditionMap(conditions []lab.ProgressCondition) map[string]lab.ProgressCondition {
	conditionMap := make(map[string]lab.ProgressCondition, len(conditions))
	for _, condition := range conditions {
		conditionMap[condition.Type] = condition
	}
	return conditionMap
}

func labPreviewProgressFor(conditionMap map[string]lab.ProgressCondition, conditionType string, runtimeActive bool) labPreviewProgress {
	if condition, ok := conditionMap[conditionType]; ok {
		switch condition.Status {
		case "True":
			return labPreviewProgress{
				ConditionType: conditionType,
				State:         "complete",
				Label:         "Complete",
				Status:        condition.Status,
				Reason:        condition.Reason,
				Message:       condition.Message,
			}
		case "False":
			return labPreviewProgress{
				ConditionType: conditionType,
				State:         "incomplete",
				Label:         "Incomplete",
				Status:        condition.Status,
				Reason:        condition.Reason,
				Message:       condition.Message,
			}
		default:
			return labPreviewProgress{
				ConditionType: conditionType,
				State:         "unknown",
				Label:         "Unknown",
				Status:        condition.Status,
				Reason:        condition.Reason,
				Message:       condition.Message,
			}
		}
	}
	if runtimeActive {
		return labPreviewProgressUnknown(conditionType, "Not reported", "The validator has not emitted this condition yet.")
	}
	return labPreviewProgressUnknown(conditionType, "Not running", "Start the lab runtime to report live progress.")
}

func labPreviewProgressUnknown(conditionType, label, message string) labPreviewProgress {
	return labPreviewProgress{
		ConditionType: conditionType,
		State:         "unknown",
		Label:         label,
		Message:       message,
	}
}

func labPreviewSourcePath(loaded *lab.Lab, path string) string {
	if loaded == nil {
		return previewSourceLabel(path)
	}
	return previewSourceLabelRelative(loaded.RootPath, path)
}

func labPreviewChallengeSources(loaded *lab.Lab) []labPreviewChallengeSource {
	if loaded == nil || loaded.Format != lab.FormatModuleBacked || loaded.RootPath == "" {
		return nil
	}
	entries, err := os.ReadDir(loaded.RootPath)
	if err != nil {
		return nil
	}

	var sources []labPreviewChallengeSource
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		challengeRoot := filepath.Join(loaded.RootPath, entry.Name())
		yamlPath := findLabPreviewChallengeFile(challengeRoot)
		if yamlPath == "" {
			continue
		}
		sources = append(sources, labPreviewChallengeSource{
			dirName:            entry.Name(),
			yamlPath:           yamlPath,
			descriptionPath:    filepath.Join(challengeRoot, "description.md"),
			successMessagePath: filepath.Join(challengeRoot, "successMessage.md"),
		})
	}
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].dirName < sources[j].dirName
	})
	return sources
}

func findLabPreviewChallengeFile(root string) string {
	for _, name := range []string{"challenge.yaml", "challenge.yml"} {
		path := filepath.Join(root, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
