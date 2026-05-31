package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldPackCreatesValidPack(t *testing.T) {
	root := filepath.Join(t.TempDir(), "story-pack")

	written, err := ScaffoldPack(root, ScaffoldOptions{
		Title:        "测试故事",
		InitialScene: "front_door",
	})
	if err != nil {
		t.Fatalf("scaffold pack: %v", err)
	}
	if len(written) != len(scaffoldFiles) {
		t.Fatalf("expected %d written files, got %d", len(scaffoldFiles), len(written))
	}

	for _, fileName := range requiredPackFiles {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(fileName))); err != nil {
			t.Fatalf("expected scaffold file %s: %v", fileName, err)
		}
	}

	report, err := ValidatePack(root)
	if err != nil {
		t.Fatalf("validate pack: %v", err)
	}
	if len(report.Problems) > 0 {
		t.Fatalf("expected valid scaffold, got problems: %#v", report.Problems)
	}
	if report.Title != "测试故事" {
		t.Fatalf("expected title from metadata, got %q", report.Title)
	}
}

func TestScaffoldPackRefusesExistingFilesWithoutForce(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "metadata.json"), []byte("{}"), 0o600); err != nil {
		t.Fatalf("write existing metadata: %v", err)
	}

	if _, err := ScaffoldPack(root, ScaffoldOptions{}); err == nil || !strings.Contains(err.Error(), "overwrite") {
		t.Fatalf("expected force hint for existing file, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "scene.md")); !os.IsNotExist(err) {
		t.Fatalf("scaffold should preflight before writing scene.md, stat err=%v", err)
	}
}

func TestValidatePackReportsMissingRequiredFile(t *testing.T) {
	root := t.TempDir()
	if _, err := ScaffoldPack(root, ScaffoldOptions{Force: true}); err != nil {
		t.Fatalf("scaffold pack: %v", err)
	}
	if err := os.Remove(filepath.Join(root, "true.md")); err != nil {
		t.Fatalf("remove true.md: %v", err)
	}

	report, err := ValidatePack(root)
	if err != nil {
		t.Fatalf("validate pack: %v", err)
	}
	if !containsIssue(report.Problems, "missing file: true.md") {
		t.Fatalf("expected missing true.md problem, got %#v", report.Problems)
	}
}

func TestValidatePackReportsInvalidJSON(t *testing.T) {
	root := t.TempDir()
	if _, err := ScaffoldPack(root, ScaffoldOptions{Force: true}); err != nil {
		t.Fatalf("scaffold pack: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "metadata.json"), []byte(`{"title":`), 0o600); err != nil {
		t.Fatalf("write invalid metadata: %v", err)
	}

	report, err := ValidatePack(root)
	if err != nil {
		t.Fatalf("validate pack: %v", err)
	}
	if !containsIssue(report.Problems, "metadata.json parse failed") {
		t.Fatalf("expected metadata parse problem, got %#v", report.Problems)
	}
}

func TestRunInitAndValidateCommands(t *testing.T) {
	root := filepath.Join(t.TempDir(), "story-pack")
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	if code := run([]string{"init", "--title", "命令测试", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("init code=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"validate", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("validate code=%d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	if !strings.Contains(stdout.String(), "status: ok") {
		t.Fatalf("expected ok status, got stdout=%s", stdout.String())
	}
}

func containsIssue(issues []string, want string) bool {
	for _, issue := range issues {
		if strings.Contains(issue, want) {
			return true
		}
	}
	return false
}
