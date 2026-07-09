package authoring

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/coreeng/tpm/pkg/module"
	"gopkg.in/yaml.v3"
)

type BreakingPolicy string

const (
	BreakingPolicyError  BreakingPolicy = "error"
	BreakingPolicyWarn   BreakingPolicy = "warn"
	BreakingPolicyIgnore BreakingPolicy = "ignore"
)

type Options struct {
	At             int
	From           int
	To             int
	Chapter        int
	Section        int
	Lab            int
	Challenge      int
	Goal           int
	Quiz           int
	Question       int
	Option         int
	Sets           []string
	Yes            bool
	BreakingPolicy BreakingPolicy
}

func Add(modulePath, resource string, opts Options) error {
	resolved, mod, err := load(modulePath)
	if err != nil {
		return err
	}
	if opts.At < 1 {
		return fmt.Errorf("--at must be 1 or greater")
	}
	fields, err := parseSets(resource, opts.Sets)
	if err != nil {
		return err
	}
	switch resource {
	case "chapter":
		return addDirResource(resolved.SourcePath, "chapter.yml", opts.At, fields, "chapter")
	case "section":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return addDirResource(filepath.Dir(chapter.FilePath), "section.yaml", opts.At, fields, "section")
	case "lab":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		parent := filepath.Join(filepath.Dir(chapter.FilePath), "assessments")
		return addDirResource(parent, "assessment.yaml", opts.At, fields, "lab")
	case "challenge":
		lab, err := selectLab(mod, opts.Chapter, opts.Lab)
		if err != nil {
			return err
		}
		return addDirResource(filepath.Dir(lab.FilePath), "challenge.yaml", opts.At, fields, "challenge")
	case "goal":
		challenge, err := selectChallenge(mod, opts.Chapter, opts.Lab, opts.Challenge)
		if err != nil {
			return err
		}
		return addInlineResource(challenge.FilePath, []inlineSelector{{key: "goals", index: 0}}, opts.At, resource, fields)
	case "quiz":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return addInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: 0}}, opts.At, resource, fields)
	case "question":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return addInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: 0}}, opts.At, resource, fields)
	case "option":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return addInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.Question}, {key: "options", index: 0}}, opts.At, resource, fields)
	default:
		return unknownResource(resource)
	}
}

func Edit(modulePath, resource string, opts Options) error {
	resolved, mod, err := load(modulePath)
	if err != nil {
		return err
	}
	fields, err := parseSets(resource, opts.Sets)
	if err != nil {
		return err
	}
	switch resource {
	case "module":
		return editMappingFile(resolved.ModuleFilePath, resource, fields, opts.BreakingPolicy)
	case "chapter":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return editMappingFile(chapter.FilePath, resource, fields, opts.BreakingPolicy)
	case "section":
		section, err := selectSection(mod, opts.Chapter, opts.Section)
		if err != nil {
			return err
		}
		return editMappingFile(section.FilePath, resource, fields, opts.BreakingPolicy)
	case "lab":
		lab, err := selectLab(mod, opts.Chapter, opts.Lab)
		if err != nil {
			return err
		}
		return editMappingFile(lab.FilePath, resource, fields, opts.BreakingPolicy)
	case "challenge":
		challenge, err := selectChallenge(mod, opts.Chapter, opts.Lab, opts.Challenge)
		if err != nil {
			return err
		}
		return editMappingFile(challenge.FilePath, resource, fields, opts.BreakingPolicy)
	case "goal":
		challenge, err := selectChallenge(mod, opts.Chapter, opts.Lab, opts.Challenge)
		if err != nil {
			return err
		}
		return editInlineResource(challenge.FilePath, []inlineSelector{{key: "goals", index: opts.Goal}}, resource, fields, opts.BreakingPolicy)
	case "quiz":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return editInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}}, resource, fields, opts.BreakingPolicy)
	case "question":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return editInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.Question}}, resource, fields, opts.BreakingPolicy)
	case "option":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return editInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.Question}, {key: "options", index: opts.Option}}, resource, fields, opts.BreakingPolicy)
	default:
		return unknownResource(resource)
	}
}

func Remove(modulePath, resource string, opts Options) error {
	_, mod, err := load(modulePath)
	if err != nil {
		return err
	}
	if !opts.Yes {
		return fmt.Errorf("--yes is required for remove")
	}
	if err := requireBreakingAllowed(opts.BreakingPolicy, "remove"); err != nil {
		return err
	}
	switch resource {
	case "chapter":
		return removeDirResource(filepath.Dir(mod.FilePath), "chapter.yml", opts.From)
	case "section":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return removeDirResource(filepath.Dir(chapter.FilePath), "section.yaml", opts.From)
	case "lab":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return removeDirResource(filepath.Join(filepath.Dir(chapter.FilePath), "assessments"), "assessment.yaml", opts.From)
	case "challenge":
		lab, err := selectLab(mod, opts.Chapter, opts.Lab)
		if err != nil {
			return err
		}
		return removeDirResource(filepath.Dir(lab.FilePath), "challenge.yaml", opts.From)
	case "goal":
		challenge, err := selectChallenge(mod, opts.Chapter, opts.Lab, opts.Challenge)
		if err != nil {
			return err
		}
		return removeInlineResource(challenge.FilePath, []inlineSelector{{key: "goals", index: opts.From}})
	case "quiz":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return removeInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.From}})
	case "question":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return removeInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.From}})
	case "option":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return removeInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.Question}, {key: "options", index: opts.From}})
	default:
		return unknownResource(resource)
	}
}

func Move(modulePath, resource string, opts Options) error {
	_, mod, err := load(modulePath)
	if err != nil {
		return err
	}
	if opts.From < 1 || opts.To < 1 {
		return fmt.Errorf("--from and --to must be 1 or greater")
	}
	switch resource {
	case "chapter":
		return moveDirResource(filepath.Dir(mod.FilePath), "chapter.yml", opts.From, opts.To)
	case "section":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return moveDirResource(filepath.Dir(chapter.FilePath), "section.yaml", opts.From, opts.To)
	case "lab":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return moveDirResource(filepath.Join(filepath.Dir(chapter.FilePath), "assessments"), "assessment.yaml", opts.From, opts.To)
	case "challenge":
		lab, err := selectLab(mod, opts.Chapter, opts.Lab)
		if err != nil {
			return err
		}
		return moveDirResource(filepath.Dir(lab.FilePath), "challenge.yaml", opts.From, opts.To)
	case "goal":
		challenge, err := selectChallenge(mod, opts.Chapter, opts.Lab, opts.Challenge)
		if err != nil {
			return err
		}
		return moveInlineResource(challenge.FilePath, []inlineSelector{{key: "goals", index: 0}}, opts.From, opts.To)
	case "quiz":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return moveInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: 0}}, opts.From, opts.To)
	case "question":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return moveInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: 0}}, opts.From, opts.To)
	case "option":
		chapter, err := selectChapter(mod, opts.Chapter)
		if err != nil {
			return err
		}
		return moveInlineResource(chapter.FilePath, []inlineSelector{{key: "multipleChoiceAssessments", index: opts.Quiz}, {key: "questions", index: opts.Question}, {key: "options", index: 0}}, opts.From, opts.To)
	default:
		return unknownResource(resource)
	}
}

func load(modulePath string) (module.ResolvedPath, *module.Module, error) {
	mod, resolved, err := module.LoadPath(modulePath)
	if err != nil {
		return module.ResolvedPath{}, nil, err
	}
	return resolved, mod, nil
}

var allowedFields = map[string]map[string]bool{
	"module":    fieldSet("code", "title", "shortDescription", "bannerImage", "bannerVideo", "tags", "level"),
	"chapter":   fieldSet("code", "title", "isDraft"),
	"section":   fieldSet("code", "title", "shortDescription", "estimatedDuration", "video", "prerequisites"),
	"lab":       fieldSet("code", "title", "timeLimit", "starterImageUri", "validatorImageUri", "imageVersion", "video"),
	"challenge": fieldSet("code", "title", "estimatedDuration", "video"),
	"goal":      fieldSet("code", "title", "description"),
	"quiz":      fieldSet("code", "title", "description", "passingScore"),
	"question":  fieldSet("code", "question", "type"),
	"option":    fieldSet("text", "correct"),
}

func fieldSet(fields ...string) map[string]bool {
	set := map[string]bool{}
	for _, field := range fields {
		set[field] = true
	}
	return set
}

func parseSets(resource string, sets []string) (map[string]*yaml.Node, error) {
	allowed, ok := allowedFields[resource]
	if !ok {
		return nil, unknownResource(resource)
	}
	fields := map[string]*yaml.Node{}
	for _, set := range sets {
		key, value, ok := strings.Cut(set, "=")
		if !ok || strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("--set must be field=value")
		}
		key = strings.TrimSpace(key)
		if !allowed[key] {
			return nil, fmt.Errorf("%s.%s is not editable YAML through the CLI", resource, key)
		}
		node, err := scalarNode(resource, key, value)
		if err != nil {
			return nil, err
		}
		fields[key] = node
	}
	return fields, nil
}

func scalarNode(resource, key, value string) (*yaml.Node, error) {
	switch fieldType(resource, key) {
	case "sequence":
		seq := &yaml.Node{Kind: yaml.SequenceNode}
		for _, item := range strings.Split(value, ",") {
			seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: strings.TrimSpace(item)})
		}
		return seq, nil
	case "bool":
		if value != "true" && value != "false" {
			return nil, fmt.Errorf("%s.%s must be true or false", resource, key)
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: value}, nil
	case "int":
		if _, err := strconv.Atoi(value); err != nil {
			return nil, fmt.Errorf("%s.%s must be an integer", resource, key)
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: value}, nil
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: value}, nil
	}
}

func fieldType(resource, key string) string {
	switch key {
	case "tags", "prerequisites":
		return "sequence"
	case "isDraft", "correct":
		return "bool"
	case "passingScore":
		return "int"
	default:
		return "string"
	}
}

func addDirResource(parent, fileName string, at int, fields map[string]*yaml.Node, resource string) error {
	if err := requireFields(fields, "code", "title"); err != nil {
		return err
	}
	if resource == "chapter" {
		if _, ok := fields["isDraft"]; !ok {
			fields["isDraft"] = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "false"}
		}
	}
	if resource == "challenge" {
		if _, ok := fields["goals"]; !ok {
			fields["goals"] = &yaml.Node{Kind: yaml.SequenceNode}
		}
	}
	if err := os.MkdirAll(parent, 0750); err != nil {
		return err
	}
	dirs, err := listResourceDirs(parent, fileName)
	if err != nil {
		return err
	}
	if at > len(dirs)+1 {
		return fmt.Errorf("--at %d is outside valid range 1-%d", at, len(dirs)+1)
	}
	dirs = append(dirs, resourceDir{})
	copy(dirs[at:], dirs[at-1:])
	slug := slugify(fields["title"].Value)
	if slug == "" {
		slug = slugify(fields["code"].Value)
	}
	dirs[at-1] = resourceDir{path: filepath.Join(parent, fmt.Sprintf("%02d-%s", at, slug)), suffix: slug}
	if err := rewriteResourceDirs(parent, dirs); err != nil {
		return err
	}
	if err := os.MkdirAll(dirs[at-1].path, 0750); err != nil {
		return err
	}
	return writeYAML(filepath.Join(dirs[at-1].path, fileName), mappingNode(fields))
}

func editMappingFile(path, resource string, fields map[string]*yaml.Node, policy BreakingPolicy) error {
	node, err := readYAMLNode(path)
	if err != nil {
		return err
	}
	mapping, err := documentMapping(node)
	if err != nil {
		return err
	}
	if codeNode := fields["code"]; codeNode != nil {
		if existing := mappingValue(mapping, "code"); existing != nil && existing.Value != "" && existing.Value != codeNode.Value {
			if err := requireBreakingAllowed(policy, "edit code"); err != nil {
				return err
			}
		}
	}
	for key, value := range fields {
		setMappingValue(mapping, key, value)
	}
	return writeYAML(path, node)
}

func removeDirResource(parent, fileName string, from int) error {
	dirs, err := listResourceDirs(parent, fileName)
	if err != nil {
		return err
	}
	if from < 1 || from > len(dirs) {
		return fmt.Errorf("--from %d is outside valid range 1-%d", from, len(dirs))
	}
	if err := os.RemoveAll(dirs[from-1].path); err != nil {
		return err
	}
	dirs = append(dirs[:from-1], dirs[from:]...)
	return rewriteResourceDirs(parent, dirs)
}

func moveDirResource(parent, fileName string, from, to int) error {
	dirs, err := listResourceDirs(parent, fileName)
	if err != nil {
		return err
	}
	if from < 1 || from > len(dirs) || to < 1 || to > len(dirs) {
		return fmt.Errorf("--from and --to must be within 1-%d", len(dirs))
	}
	item := dirs[from-1]
	dirs = append(dirs[:from-1], dirs[from:]...)
	dirs = append(dirs[:to-1], append([]resourceDir{item}, dirs[to-1:]...)...)
	return rewriteResourceDirs(parent, dirs)
}

type resourceDir struct {
	path   string
	suffix string
}

func listResourceDirs(parent, fileName string) ([]resourceDir, error) {
	entries, err := os.ReadDir(parent)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var dirs []resourceDir
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		path := filepath.Join(parent, entry.Name())
		if !anyResourceFileExists(path, fileName) {
			continue
		}
		dirs = append(dirs, resourceDir{path: path, suffix: suffix(entry.Name())})
	}
	sort.Slice(dirs, func(i, j int) bool {
		return filepath.Base(dirs[i].path) < filepath.Base(dirs[j].path)
	})
	return dirs, nil
}

func anyResourceFileExists(dir, fileName string) bool {
	for _, name := range resourceFileNames(fileName) {
		if fileExists(filepath.Join(dir, name)) {
			return true
		}
	}
	return false
}

func resourceFileNames(fileName string) []string {
	switch ext := filepath.Ext(fileName); ext {
	case ".yaml":
		return []string{fileName, strings.TrimSuffix(fileName, ext) + ".yml"}
	case ".yml":
		return []string{fileName, strings.TrimSuffix(fileName, ext) + ".yaml"}
	default:
		return []string{fileName}
	}
}

func rewriteResourceDirs(parent string, dirs []resourceDir) error {
	tempPaths := make([]string, len(dirs))
	for i, dir := range dirs {
		if dir.path == "" || !fileExists(filepath.Join(dir.path, ".")) && !dirExists(dir.path) {
			continue
		}
		temp := filepath.Join(parent, fmt.Sprintf(".tpm-move-%d-%s", i, filepath.Base(dir.path)))
		if err := os.Rename(dir.path, temp); err != nil {
			return err
		}
		tempPaths[i] = temp
	}
	for i, dir := range dirs {
		final := filepath.Join(parent, fmt.Sprintf("%02d-%s", i+1, dir.suffix))
		if tempPaths[i] == "" {
			dirs[i].path = final
			continue
		}
		if err := os.Rename(tempPaths[i], final); err != nil {
			return err
		}
		dirs[i].path = final
	}
	return nil
}

func addInlineResource(path string, selectors []inlineSelector, at int, resource string, fields map[string]*yaml.Node) error {
	required := []string{"code"}
	if resource == "option" {
		required = []string{"text"}
	}
	if err := requireFields(fields, required...); err != nil {
		return err
	}
	node, seq, err := inlineSequence(path, selectors, true)
	if err != nil {
		return err
	}
	if at > len(seq.Content)+1 {
		return fmt.Errorf("--at %d is outside valid range 1-%d", at, len(seq.Content)+1)
	}
	item := mappingNode(fields)
	seq.Content = append(seq.Content, nil)
	copy(seq.Content[at:], seq.Content[at-1:])
	seq.Content[at-1] = item
	return writeYAML(path, node)
}

func editInlineResource(path string, selectors []inlineSelector, resource string, fields map[string]*yaml.Node, policy BreakingPolicy) error {
	node, item, err := inlineItem(path, selectors)
	if err != nil {
		return err
	}
	if codeNode := fields["code"]; codeNode != nil {
		if existing := mappingValue(item, "code"); existing != nil && existing.Value != "" && existing.Value != codeNode.Value {
			if err := requireBreakingAllowed(policy, "edit code"); err != nil {
				return err
			}
		}
	}
	for key, value := range fields {
		setMappingValue(item, key, value)
	}
	return writeYAML(path, node)
}

func removeInlineResource(path string, selectors []inlineSelector) error {
	node, seq, index, err := inlineSequenceAndIndex(path, selectors)
	if err != nil {
		return err
	}
	seq.Content = append(seq.Content[:index], seq.Content[index+1:]...)
	return writeYAML(path, node)
}

func moveInlineResource(path string, selectors []inlineSelector, from, to int) error {
	node, seq, err := inlineSequence(path, selectors, false)
	if err != nil {
		return err
	}
	if from < 1 || from > len(seq.Content) || to < 1 || to > len(seq.Content) {
		return fmt.Errorf("--from and --to must be within 1-%d", len(seq.Content))
	}
	item := seq.Content[from-1]
	seq.Content = append(seq.Content[:from-1], seq.Content[from:]...)
	seq.Content = append(seq.Content[:to-1], append([]*yaml.Node{item}, seq.Content[to-1:]...)...)
	return writeYAML(path, node)
}

type inlineSelector struct {
	key   string
	index int
}

func inlineItem(path string, selectors []inlineSelector) (*yaml.Node, *yaml.Node, error) {
	node, seq, index, err := inlineSequenceAndIndex(path, selectors)
	if err != nil {
		return nil, nil, err
	}
	item := seq.Content[index]
	if item.Kind != yaml.MappingNode {
		return nil, nil, fmt.Errorf("selected item is not a YAML mapping")
	}
	return node, item, nil
}

func inlineSequenceAndIndex(path string, selectors []inlineSelector) (*yaml.Node, *yaml.Node, int, error) {
	if len(selectors) == 0 {
		return nil, nil, 0, fmt.Errorf("missing inline selector")
	}
	last := selectors[len(selectors)-1]
	node, seq, err := inlineSequence(path, selectors, false)
	if err != nil {
		return nil, nil, 0, err
	}
	index := last.index - 1
	if index < 0 || index >= len(seq.Content) {
		return nil, nil, 0, fmt.Errorf("%s index %d is outside valid range 1-%d", last.key, last.index, len(seq.Content))
	}
	return node, seq, index, nil
}

func inlineSequence(path string, selectors []inlineSelector, create bool) (*yaml.Node, *yaml.Node, error) {
	node, err := readYAMLNode(path)
	if err != nil {
		return nil, nil, err
	}
	current, err := documentMapping(node)
	if err != nil {
		return nil, nil, err
	}
	var seq *yaml.Node
	for i, selector := range selectors {
		value := mappingValue(current, selector.key)
		if value == nil {
			if !create {
				return nil, nil, fmt.Errorf("missing %s", selector.key)
			}
			value = &yaml.Node{Kind: yaml.SequenceNode}
			setMappingValue(current, selector.key, value)
		}
		if value.Kind != yaml.SequenceNode {
			return nil, nil, fmt.Errorf("%s is not a sequence", selector.key)
		}
		seq = value
		if i == len(selectors)-1 {
			break
		}
		index := selector.index - 1
		if index < 0 || index >= len(value.Content) {
			return nil, nil, fmt.Errorf("%s index %d is outside valid range 1-%d", selector.key, selector.index, len(value.Content))
		}
		current = value.Content[index]
		if current.Kind != yaml.MappingNode {
			return nil, nil, fmt.Errorf("selected %s item is not a mapping", selector.key)
		}
	}
	return node, seq, nil
}

func readYAMLNode(path string) (*yaml.Node, error) {
	// #nosec G304 -- authoring commands intentionally edit local YAML paths selected by the CLI user.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

func writeYAML(path string, node *yaml.Node) error {
	data, err := yaml.Marshal(node)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func mappingNode(fields map[string]*yaml.Node) *yaml.Node {
	node := &yaml.Node{Kind: yaml.MappingNode}
	preferred := []string{"code", "title", "question", "type", "text", "correct", "description", "passingScore", "isDraft", "shortDescription", "timeLimit", "starterImageUri", "validatorImageUri", "imageVersion", "estimatedDuration", "video", "bannerImage", "bannerVideo", "tags", "prerequisites", "goals", "questions", "options"}
	seen := map[string]bool{}
	for _, key := range preferred {
		if value, ok := fields[key]; ok {
			node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
			seen[key] = true
		}
	}
	var rest []string
	for key := range fields {
		if !seen[key] {
			rest = append(rest, key)
		}
	}
	sort.Strings(rest)
	for _, key := range rest {
		node.Content = append(node.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, fields[key])
	}
	return node
}

func documentMapping(node *yaml.Node) (*yaml.Node, error) {
	if node.Kind != yaml.DocumentNode || len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected YAML document mapping")
	}
	return node.Content[0], nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

func setMappingValue(mapping *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content[i+1] = value
			return
		}
	}
	mapping.Content = append(mapping.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
}

func requireFields(fields map[string]*yaml.Node, keys ...string) error {
	for _, key := range keys {
		if fields[key] == nil || strings.TrimSpace(fields[key].Value) == "" {
			return fmt.Errorf("--set %s=... is required", key)
		}
	}
	return nil
}

func requireBreakingAllowed(policy BreakingPolicy, action string) error {
	if policy == BreakingPolicyError || policy == "" {
		return fmt.Errorf("%s is a breaking change; set --breaking-policy=warn or --breaking-policy=ignore", action)
	}
	return nil
}

func selectChapter(mod *module.Module, index int) (module.Chapter, error) {
	if index < 1 || index > len(mod.Chapters) {
		return module.Chapter{}, fmt.Errorf("--chapter %d is outside valid range 1-%d", index, len(mod.Chapters))
	}
	return mod.Chapters[index-1], nil
}

func selectSection(mod *module.Module, chapterIndex, sectionIndex int) (module.Section, error) {
	chapter, err := selectChapter(mod, chapterIndex)
	if err != nil {
		return module.Section{}, err
	}
	if sectionIndex < 1 || sectionIndex > len(chapter.Sections) {
		return module.Section{}, fmt.Errorf("--section %d is outside valid range 1-%d", sectionIndex, len(chapter.Sections))
	}
	return chapter.Sections[sectionIndex-1], nil
}

func selectLab(mod *module.Module, chapterIndex, labIndex int) (module.Assessment, error) {
	chapter, err := selectChapter(mod, chapterIndex)
	if err != nil {
		return module.Assessment{}, err
	}
	if labIndex < 1 || labIndex > len(chapter.Assessments) {
		return module.Assessment{}, fmt.Errorf("--lab %d is outside valid range 1-%d", labIndex, len(chapter.Assessments))
	}
	return chapter.Assessments[labIndex-1], nil
}

func selectChallenge(mod *module.Module, chapterIndex, labIndex, challengeIndex int) (module.Challenge, error) {
	lab, err := selectLab(mod, chapterIndex, labIndex)
	if err != nil {
		return module.Challenge{}, err
	}
	if challengeIndex < 1 || challengeIndex > len(lab.Challenges) {
		return module.Challenge{}, fmt.Errorf("--challenge %d is outside valid range 1-%d", challengeIndex, len(lab.Challenges))
	}
	return lab.Challenges[challengeIndex-1], nil
}

func suffix(name string) string {
	parts := strings.SplitN(name, "-", 2)
	if len(parts) == 2 {
		if _, err := strconv.Atoi(parts[0]); err == nil {
			return parts[1]
		}
	}
	return name
}

var nonSlug = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(value string) string {
	slug := strings.ToLower(strings.TrimSpace(value))
	slug = nonSlug.ReplaceAllString(slug, "-")
	return strings.Trim(slug, "-")
}

func unknownResource(resource string) error {
	return fmt.Errorf("unknown resource %q", resource)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
