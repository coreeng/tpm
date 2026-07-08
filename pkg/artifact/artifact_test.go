package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

const validBuiltModule = `code: test-module-123
title: Test Module
description: A test module
shortDescription: Test
bannerImage: https://example.com/banner.png
bannerVideo: https://example.com/video.mp4
tags:
  - test
level: BEGINNER
chapters:
  - code: test-chapter-456
    index: 1
    title: Test Chapter
    description: A test chapter
    shortDescription: Test chapter
    bannerImage: https://example.com/chapter.png
    bannerVideo: https://example.com/chapter.mp4
    isDraft: true
    sections:
      - code: test-section-789
        index: 1
        title: Test Section
        description: A test section
        shortDescription: A brief test section
        video: https://example.com/video.mp4
        estimatedDuration: 30m
    assessments: []
    multipleChoiceAssessments: []
`

func TestResolveModuleArtifactPathAcceptsFileAndDirectory(t *testing.T) {
	dir := t.TempDir()
	artifactPath := filepath.Join(dir, ModuleArtifactFile)
	if err := os.WriteFile(artifactPath, []byte(validBuiltModule), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	resolvedFile, err := ResolveModuleArtifactPath(artifactPath)
	if err != nil {
		t.Fatalf("resolve artifact file: %v", err)
	}
	if resolvedFile != artifactPath {
		t.Fatalf("resolved file = %q, want %q", resolvedFile, artifactPath)
	}

	resolvedDir, err := ResolveModuleArtifactPath(dir)
	if err != nil {
		t.Fatalf("resolve artifact directory: %v", err)
	}
	if resolvedDir != artifactPath {
		t.Fatalf("resolved dir = %q, want %q", resolvedDir, artifactPath)
	}
}

func TestValidateModuleArtifactUsesEmbeddedBuiltSchema(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ModuleArtifactFile), []byte(validBuiltModule), 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	result, err := ValidateModuleArtifact(dir, "")
	if err != nil {
		t.Fatalf("validate artifact: %v", err)
	}
	if result.HasErrors() {
		t.Fatalf("expected no validation errors, got %#v", result.Issues)
	}
}

func TestValidateModuleArtifactReportsSchemaErrors(t *testing.T) {
	dir := t.TempDir()
	invalid := []byte(`code: test-module-123
title: Test Module
`)
	if err := os.WriteFile(filepath.Join(dir, ModuleArtifactFile), invalid, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	result, err := ValidateModuleArtifact(dir, "")
	if err != nil {
		t.Fatalf("validate artifact: %v", err)
	}
	if !result.HasErrors() {
		t.Fatalf("expected validation errors")
	}
}
