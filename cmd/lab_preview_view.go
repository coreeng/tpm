package cmd

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/coreeng/tpm/pkg/lab"
)

type labPreviewChallenge struct {
	lab.Challenge
	Source               string
	DescriptionSource    string
	SuccessMessageSource string
	Goals                []labPreviewGoal
}

type labPreviewGoal struct {
	lab.Goal
	Source string
}

type labPreviewChallengeSource struct {
	dirName            string
	yamlPath           string
	descriptionPath    string
	successMessagePath string
}

func newLabPreviewPage(loaded *lab.Lab, state *lab.RunState) labPreviewPage {
	page := labPreviewPage{Lab: loaded, State: state}
	if loaded == nil {
		return page
	}

	page.Source = labPreviewSourceLabel(loaded, loaded.MetadataPath)
	page.DescriptionSource = page.Source
	page.TimeLimitSource = page.Source
	if loaded.Format == lab.FormatModuleBacked {
		descriptionPath := existingSourceOrFallback(filepath.Join(loaded.RootPath, "description.md"), loaded.MetadataPath)
		page.DescriptionSource = labPreviewSourceLabel(loaded, descriptionPath)
	}

	challengeSources := labPreviewChallengeSources(loaded)
	for i, challenge := range loaded.Challenges {
		yamlPath := loaded.MetadataPath
		descriptionPath := loaded.MetadataPath
		successMessagePath := loaded.MetadataPath
		if i < len(challengeSources) {
			yamlPath = challengeSources[i].yamlPath
			descriptionPath = existingSourceOrFallback(challengeSources[i].descriptionPath, yamlPath)
			successMessagePath = existingSourceOrFallback(challengeSources[i].successMessagePath, yamlPath)
		}
		previewChallenge := labPreviewChallenge{
			Challenge:            challenge,
			Source:               labPreviewSourceLabel(loaded, yamlPath),
			DescriptionSource:    labPreviewSourceLabel(loaded, descriptionPath),
			SuccessMessageSource: labPreviewSourceLabel(loaded, successMessagePath),
			Goals:                make([]labPreviewGoal, 0, len(challenge.Goals)),
		}
		for _, goal := range challenge.Goals {
			previewChallenge.Goals = append(previewChallenge.Goals, labPreviewGoal{
				Goal:   goal,
				Source: labPreviewSourceLabel(loaded, yamlPath),
			})
		}
		page.Challenges = append(page.Challenges, previewChallenge)
	}
	return page
}

func labPreviewSourceLabel(loaded *lab.Lab, path string) string {
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
