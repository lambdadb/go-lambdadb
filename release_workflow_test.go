package lambdadb_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// These tests exercise the release guards used by .github/workflows/publish.yaml
// against temporary files and Git repositories.

func requireTools(t *testing.T, tools ...string) {
	t.Helper()
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not available: %v", tool, err)
		}
	}
}

func scriptPath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("scripts", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

func readRepoFile(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// runCommand executes a command in dir with Git isolated from user and system config.
func runCommand(t *testing.T, dir string, env []string, name string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=Release Test",
		"GIT_AUTHOR_EMAIL=release-test@example.com",
		"GIT_COMMITTER_NAME=Release Test",
		"GIT_COMMITTER_EMAIL=release-test@example.com",
	)
	cmd.Env = append(cmd.Env, env...)
	output, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

func mustRunCommand(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	output, err := runCommand(t, dir, nil, name, args...)
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, output)
	}
	return output
}

func writeReleaseFiles(t *testing.T, dir, version, changelogVersion string) {
	t.Helper()
	files := map[string]string{
		"lambdadb.go":  "package lambdadb\n\nconst Version = \"" + version + "\"\n",
		"CHANGELOG.md": "# Changelog\n\n## [Unreleased]\n\n## [" + changelogVersion + "] - 2026-10-29\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestValidateReleaseMetadata(t *testing.T) {
	requireTools(t, "bash")
	script := scriptPath(t, "validate-release-metadata.sh")

	tests := []struct {
		name      string
		version   string
		changelog string
		tag       string
		want      string
		wantErr   string
	}{
		{name: "stable", version: "0.8.0", changelog: "0.8.0", tag: "v0.8.0",
			want: "prerelease=false\nmake_latest=true"},
		{name: "release candidate", version: "0.8.0-rc.2", changelog: "0.8.0-rc.2", tag: "v0.8.0-rc.2",
			want: "prerelease=true\nmake_latest=false"},
		{name: "rc zero", version: "0.8.0-rc.0", changelog: "0.8.0-rc.0", tag: "v0.8.0-rc.0",
			wantErr: "invalid release tag"},
		{name: "dev tag", version: "0.8.0-dev.1", changelog: "0.8.0-dev.1", tag: "v0.8.0-dev.1",
			wantErr: "invalid release tag"},
		{name: "missing leading v", version: "0.8.0", changelog: "0.8.0", tag: "0.8.0",
			wantErr: "invalid release tag"},
		{name: "version mismatch", version: "0.7.2", changelog: "0.8.0", tag: "v0.8.0",
			wantErr: "does not match SDK Version 0.7.2"},
		{name: "missing changelog", version: "0.8.0", changelog: "0.7.2", tag: "v0.8.0",
			wantErr: "no release section for 0.8.0"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			writeReleaseFiles(t, dir, test.version, test.changelog)
			output, err := runCommand(t, dir, nil, "bash", script, test.tag)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(output, test.wantErr) {
					t.Fatalf("got err=%v output=%q, want failure containing %q", err, output, test.wantErr)
				}
				return
			}
			if err != nil || output != test.want {
				t.Fatalf("got err=%v output=%q, want %q", err, output, test.want)
			}
		})
	}
}

func TestVerifyMainAncestry(t *testing.T) {
	requireTools(t, "bash", "git")
	script := scriptPath(t, "verify-main-ancestry.sh")

	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	work := filepath.Join(root, "work")
	mustRunCommand(t, root, "git", "init", "--quiet", "--bare", origin)
	mustRunCommand(t, root, "git", "init", "--quiet", "-b", "main", work)
	mustRunCommand(t, work, "git", "remote", "add", "origin", origin)

	mustRunCommand(t, work, "git", "commit", "--quiet", "--allow-empty", "-m", "release")
	released := mustRunCommand(t, work, "git", "rev-parse", "HEAD")
	mustRunCommand(t, work, "git", "tag", "-a", "-m", "v0.8.0", "v0.8.0")
	mustRunCommand(t, work, "git", "push", "--quiet", "origin", "main")

	mustRunCommand(t, work, "git", "switch", "--quiet", "-c", "feature")
	mustRunCommand(t, work, "git", "commit", "--quiet", "--allow-empty", "-m", "unmerged")
	mustRunCommand(t, work, "git", "tag", "v0.8.1-rc.1")

	tests := []struct {
		name    string
		target  []string
		wantErr string
	}{
		{name: "main commit", target: []string{released}},
		{name: "annotated tag on main", target: []string{"v0.8.0"}},
		{name: "unmerged head", target: nil, wantErr: "is not contained in origin/main"},
		{name: "tag outside main", target: []string{"v0.8.1-rc.1"}, wantErr: "is not contained in origin/main"},
		{name: "unknown target", target: []string{"v9.9.9"}, wantErr: "unknown commit: v9.9.9"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output, err := runCommand(t, work, nil, "bash", append([]string{script}, test.target...)...)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(output, test.wantErr) {
					t.Fatalf("got err=%v output=%q, want failure containing %q", err, output, test.wantErr)
				}
				return
			}
			if err != nil || !strings.Contains(output, "is contained in origin/main") {
				t.Fatalf("got err=%v output=%q, want success", err, output)
			}
		})
	}

	// A main push that later lands on main passes once it is pushed.
	mustRunCommand(t, work, "git", "push", "--quiet", "origin", "feature:main")
	if output, err := runCommand(t, work, nil, "bash", script); err != nil {
		t.Fatalf("pushed head: %v\n%s", err, output)
	}
}

// workflowStep returns the run script of the named step in workflow.
func workflowStep(t *testing.T, workflow, name string) string {
	t.Helper()
	pattern := regexp.MustCompile(`(?s)- name: ` + regexp.QuoteMeta(name) + `\n.*?run: \|\n(.*?)\n\n`)
	match := pattern.FindStringSubmatch(workflow)
	if match == nil {
		t.Fatalf("workflow step %q not found", name)
	}
	lines := strings.Split(match[1], "\n")
	for i, line := range lines {
		lines[i] = strings.TrimPrefix(line, "          ")
	}
	return strings.Join(lines, "\n")
}

func TestPublishWorkflowReleaseTagSelection(t *testing.T) {
	requireTools(t, "bash")
	workflow := readRepoFile(t, filepath.Join(".github", "workflows", "publish.yaml"))
	step := workflowStep(t, workflow, "Validate release metadata")

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	validator := readRepoFile(t, filepath.Join("scripts", "validate-release-metadata.sh"))
	if err := os.WriteFile(filepath.Join(dir, "scripts", "validate-release-metadata.sh"), []byte(validator), 0o755); err != nil {
		t.Fatal(err)
	}
	writeReleaseFiles(t, dir, "0.8.0-rc.1", "0.8.0-rc.1")

	validate := func(refType, refName string) (string, error) {
		outputFile := filepath.Join(dir, "github-output")
		_ = os.Remove(outputFile)
		output, err := runCommand(t, dir, []string{
			"GITHUB_REF_TYPE=" + refType,
			"GITHUB_REF_NAME=" + refName,
			"GITHUB_OUTPUT=" + outputFile,
		}, "bash", "-e", "-c", step)
		if err != nil {
			return output, err
		}
		data, err := os.ReadFile(outputFile)
		return strings.TrimSpace(string(data)), err
	}

	// A main push validates the committed Version as its future tag.
	if output, err := validate("branch", "main"); err != nil || output != "prerelease=true\nmake_latest=false" {
		t.Fatalf("main push: err=%v output=%q", err, output)
	}
	// A tag push validates the pushed tag against the committed Version.
	if output, err := validate("tag", "v0.8.0-rc.1"); err != nil || output != "prerelease=true\nmake_latest=false" {
		t.Fatalf("matching tag: err=%v output=%q", err, output)
	}
	if output, err := validate("tag", "v0.8.0"); err == nil || !strings.Contains(output, "does not match SDK Version") {
		t.Fatalf("mismatched tag: err=%v output=%q", err, output)
	}
	// The real SDK Version must be readable by the same expression.
	if !regexp.MustCompile(`(?m)^const Version = "[^"]+"$`).MatchString(readRepoFile(t, "lambdadb.go")) {
		t.Fatal("lambdadb.go Version declaration no longer matches the workflow expression")
	}
}

func TestReleaseWorkflowsTargetMainOnly(t *testing.T) {
	publish := readRepoFile(t, filepath.Join(".github", "workflows", "publish.yaml"))
	ci := readRepoFile(t, filepath.Join(".github", "workflows", "ci.yaml"))

	checks := []struct {
		name     string
		workflow string
		pattern  string
	}{
		{"publish triggers on main pushes and v* tags", publish,
			`on:\n  push:\n    branches: \[main\]\n    tags:\n      - "v\*"\n\n`},
		{"publish checks main ancestry before building", publish,
			`- name: Verify the commit is on main\n        run: \./scripts/verify-main-ancestry\.sh HEAD\n`},
		{"publish checks out full history", publish,
			`- uses: actions/checkout@v7\n        with:\n          fetch-depth: 0\n\n      # A tag push`},
		{"only tag pushes create a GitHub Release", publish,
			`release:\n    needs: test\n    if: github\.ref_type == 'tag'\n`},
		{"CI runs only for main", ci,
			`on:\n  push:\n    branches: \[main\]\n  pull_request:\n    branches: \[main\]\n\n`},
	}
	for _, check := range checks {
		if !regexp.MustCompile(check.pattern).MatchString(check.workflow) {
			t.Errorf("%s: pattern %q not found", check.name, check.pattern)
		}
	}
	for name, workflow := range map[string]string{"publish.yaml": publish, "ci.yaml": ci} {
		if strings.Contains(workflow, "develop") {
			t.Errorf("%s still references develop", name)
		}
	}
}
