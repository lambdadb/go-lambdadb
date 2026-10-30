package lambdadb_test

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// These tests exercise the deployed live smoke scripts used by
// .github/workflows/publish.yaml with fake go and live test commands.

const syntheticSmokeKey = "synthetic%smoke-key"

// copyScript copies a repository script into dir/scripts and returns its path.
func copyScript(t *testing.T, dir, name string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "scripts", name)
	if err := os.WriteFile(path, []byte(readRepoFile(t, filepath.Join("scripts", name))), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeExecutable(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatal(err)
	}
}

func readRecord(t *testing.T, path string) (string, bool) {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", false
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(data), true
}

// liveSmokeTests returns the live smoke tests and their opt-in flags declared
// by the repository's integration test files.
func liveSmokeTests(t *testing.T) (tests, flags []string) {
	t.Helper()
	files, err := filepath.Glob("*_integration_test.go")
	if err != nil {
		t.Fatal(err)
	}
	testPattern := regexp.MustCompile(`(?m)^func (TestIntegration\w+)\(t \*testing\.T\)`)
	flagPattern := regexp.MustCompile(`os\.Getenv\("(LAMBDADB_RUN_\w+)"\)`)
	for _, file := range files {
		source := readRepoFile(t, file)
		for _, match := range testPattern.FindAllStringSubmatch(source, -1) {
			tests = append(tests, match[1])
		}
		for _, match := range flagPattern.FindAllStringSubmatch(source, -1) {
			flags = append(flags, match[1])
		}
	}
	sort.Strings(tests)
	sort.Strings(flags)
	if len(tests) == 0 || len(flags) != len(tests) {
		t.Fatalf("found live tests %v with opt-in flags %v", tests, flags)
	}
	return tests, flags
}

func TestLiveTestScriptRunsEveryLiveTestSequentially(t *testing.T) {
	requireTools(t, "bash", "grep", "tee")
	liveTests, flags := liveSmokeTests(t)

	dir := t.TempDir()
	script := copyScript(t, dir, "live-test.sh")
	bin := filepath.Join(dir, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	// The fake go command records its arguments and environment, prints
	// GO_OUTPUT, and exits with GO_STATUS.
	writeExecutable(t, filepath.Join(bin, "go"), `#!/usr/bin/env bash
printf '%s\n' "$@" > "${RECORD_DIR}/go-args"
env > "${RECORD_DIR}/go-env"
printf '%s' "${GO_OUTPUT}"
exit "${GO_STATUS:-0}"
`)

	var passing strings.Builder
	for _, test := range liveTests {
		passing.WriteString("=== RUN   " + test + "\n--- PASS: " + test + " (1.00s)\n")
	}
	passing.WriteString("PASS\nok  \tgithub.com/lambdadb/go-lambdadb\t5.000s\n")

	baseEnv := func(record string) []string {
		env := []string{
			"PATH=" + bin + string(os.PathListSeparator) + os.Getenv("PATH"),
			"RECORD_DIR=" + record,
			"LAMBDADB_BASE_URL=https://test.example",
			"LAMBDADB_PROJECT_NAME=go-sdk-ci",
			"LAMBDADB_PROJECT_API_KEY=" + syntheticSmokeKey,
		}
		for _, flag := range flags {
			env = append(env, flag+"=")
		}
		return env
	}

	tests := []struct {
		name     string
		env      []string
		output   string
		status   string
		wantErr  string
		wantNoGo bool
	}{
		{name: "all pass", output: passing.String()},
		{name: "missing base URL", env: []string{"LAMBDADB_BASE_URL="},
			wantErr: "Missing live test configuration: LAMBDADB_BASE_URL", wantNoGo: true},
		{name: "blank project name", env: []string{"LAMBDADB_PROJECT_NAME= "},
			wantErr: "Missing live test configuration: LAMBDADB_PROJECT_NAME", wantNoGo: true},
		{name: "missing key", env: []string{"LAMBDADB_PROJECT_API_KEY="},
			wantErr: "Missing live test configuration: LAMBDADB_PROJECT_API_KEY", wantNoGo: true},
		{name: "go test failure", output: passing.String(), status: "1"},
		{name: "skipped test", wantErr: "skipped tests are not validation",
			output: strings.Replace(passing.String(), "--- PASS: "+liveTests[0], "--- SKIP: "+liveTests[0], 1)},
		{name: "skipped subtest", wantErr: "skipped tests are not validation",
			output: "    --- SKIP: " + liveTests[0] + "/case (0.00s)\n" + passing.String()},
		{name: "test did not run", wantErr: "Live smoke test " + liveTests[len(liveTests)-1] + " did not pass",
			output: strings.Replace(passing.String(), "--- PASS: "+liveTests[len(liveTests)-1]+" ", "", 1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			record := t.TempDir()
			env := append(baseEnv(record), "GO_OUTPUT="+test.output, "GO_STATUS="+test.status)
			output, err := runCommand(t, dir, append(env, test.env...), "bash", script)
			goArgs, ranGo := readRecord(t, filepath.Join(record, "go-args"))
			if test.wantNoGo && ranGo {
				t.Fatal("go test ran despite missing configuration")
			}
			if test.wantErr != "" || test.status != "" {
				if err == nil || !strings.Contains(output, test.wantErr) {
					t.Fatalf("got err=%v output=%q, want failure containing %q", err, output, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("got err=%v output=%q, want success", err, output)
			}
			wantArgs := "test\n-p\n1\n-count=1\n-timeout\n45m\n-v\n-run\n^(" + strings.Join(liveTests, "|") + ")$\n.\n"
			if goArgs != wantArgs {
				t.Fatalf("go arguments = %q, want %q", goArgs, wantArgs)
			}
			goEnv, _ := readRecord(t, filepath.Join(record, "go-env"))
			for _, flag := range flags {
				if !regexp.MustCompile(`(?m)^` + flag + `=1$`).MatchString(goEnv) {
					t.Errorf("go test environment does not enable %s", flag)
				}
			}
		})
	}
}

func TestRunLiveSmokeChecksTheKeyAndRunsTheSuite(t *testing.T) {
	requireTools(t, "bash")
	dir := t.TempDir()
	script := copyScript(t, dir, "run-live-smoke.sh")
	// The fake live test records its arguments and environment and exits with
	// LIVE_STATUS.
	writeExecutable(t, filepath.Join(dir, "scripts", "live-test.sh"), `#!/usr/bin/env bash
printf '%s\n' "$@" > "${RECORD_DIR}/live-args"
env > "${RECORD_DIR}/live-env"
echo "live tests ran"
exit "${LIVE_STATUS:-0}"
`)

	baseEnv := func(record string) []string {
		return []string{
			"RECORD_DIR=" + record,
			"LAMBDADB_BASE_URL=https://test.example",
			"LAMBDADB_PROJECT_NAME=go-sdk-ci",
			"LAMBDADB_PROJECT_API_KEY=" + syntheticSmokeKey,
		}
	}
	run := func(t *testing.T, extra ...string) (record, output string, err error) {
		record = t.TempDir()
		output, err = runCommand(t, dir, append(baseEnv(record), extra...), "bash", script)
		return record, output, err
	}

	t.Run("key is passed in the live test environment and never printed", func(t *testing.T) {
		record, output, err := run(t)
		if err != nil {
			t.Fatalf("got err=%v output=%q", err, output)
		}
		if strings.Contains(output, "smoke-key") || strings.Contains(output, "::add-mask::") {
			t.Fatalf("key printed: %q", output)
		}
		liveArgs, _ := readRecord(t, filepath.Join(record, "live-args"))
		liveEnv, _ := readRecord(t, filepath.Join(record, "live-env"))
		keyLine := regexp.MustCompile(`(?m)^LAMBDADB_PROJECT_API_KEY=` + regexp.QuoteMeta(syntheticSmokeKey) + `$`)
		if liveArgs != "\n" || !keyLine.MatchString(liveEnv) {
			t.Fatalf("live test args=%q; key not passed in its environment", liveArgs)
		}
	})

	failures := []struct {
		name    string
		env     []string
		wantErr string
	}{
		{name: "missing base URL", env: []string{"LAMBDADB_BASE_URL="}, wantErr: "Missing smoke configuration: LAMBDADB_BASE_URL"},
		{name: "blank project", env: []string{"LAMBDADB_PROJECT_NAME= "}, wantErr: "Missing smoke configuration: LAMBDADB_PROJECT_NAME"},
		{name: "missing key", env: []string{"LAMBDADB_PROJECT_API_KEY="}, wantErr: "Missing smoke configuration: LAMBDADB_PROJECT_API_KEY"},
		{name: "blank key", env: []string{"LAMBDADB_PROJECT_API_KEY= \t"}, wantErr: "Missing smoke configuration: LAMBDADB_PROJECT_API_KEY"},
		{name: "multiple lines", env: []string{"LAMBDADB_PROJECT_API_KEY=key\nother-key"}, wantErr: "single project API key"},
		{name: "carriage return", env: []string{"LAMBDADB_PROJECT_API_KEY=key\r"}, wantErr: "single project API key"},
		{name: "live test failure", env: []string{"LIVE_STATUS=1"}, wantErr: "Deployed live smoke failed"},
	}
	for _, test := range failures {
		t.Run(test.name, func(t *testing.T) {
			record, output, err := run(t, test.env...)
			if err == nil || !strings.Contains(output, test.wantErr) {
				t.Fatalf("got err=%v output=%q, want failure containing %q", err, output, test.wantErr)
			}
			_, ranLive := readRecord(t, filepath.Join(record, "live-args"))
			if ranLive != (test.name == "live test failure") {
				t.Fatalf("live test ran=%v before its preconditions held", ranLive)
			}
		})
	}
}

func TestSelectSmokeEnvironment(t *testing.T) {
	requireTools(t, "bash", "git")
	script := scriptPath(t, "select-smoke-environment.sh")

	root := t.TempDir()
	origin := filepath.Join(root, "origin.git")
	work := filepath.Join(root, "work")
	mustRunCommand(t, root, "git", "init", "--quiet", "--bare", origin)
	mustRunCommand(t, root, "git", "init", "--quiet", "-b", "main", work)
	mustRunCommand(t, work, "git", "remote", "add", "origin", origin)
	writeReleaseFiles(t, work, "0.8.0", "0.8.0")
	mustRunCommand(t, work, "git", "add", ".")
	mustRunCommand(t, work, "git", "commit", "--quiet", "-m", "release 0.8.0")
	mustRunCommand(t, work, "git", "push", "--quiet", "origin", "main")

	selectEnv := func(args ...string) (string, error) {
		return runCommand(t, work, []string{"GITHUB_REF_TYPE=", "GITHUB_REF_NAME="}, "bash", append([]string{script}, args...)...)
	}
	expect := func(name, wantPrefix, wantEnvironment string, args ...string) {
		t.Helper()
		output, err := selectEnv(args...)
		want := "\nsmoke_prefix=" + wantPrefix + "\nenvironment=" + wantEnvironment
		if err != nil || !strings.HasSuffix(output, want) {
			t.Fatalf("%s: got err=%v output=%q, want suffix %q", name, err, output, want)
		}
	}

	// The committed Version is not tagged yet: a release commit gates its tag on production.
	expect("untagged release commit", "LAMBDADB_PROD_", "prod", "branch", "main")
	// A local tag that was not pushed does not make the commit a development commit.
	mustRunCommand(t, work, "git", "tag", "v0.8.0")
	expect("tag only in the local checkout", "LAMBDADB_PROD_", "prod", "branch", "main")
	// After the tag is pushed, later main commits keep that Version and are development commits.
	mustRunCommand(t, work, "git", "push", "--quiet", "origin", "v0.8.0")
	expect("development commit after the release", "LAMBDADB_DEV_", "dev", "branch", "main")
	// A release tag push always targets production.
	expect("release tag", "LAMBDADB_PROD_", "prod", "tag", "v0.8.0")

	if output, err := selectEnv("branch", "feature"); err == nil || !strings.Contains(output, "only for main pushes") {
		t.Fatalf("feature branch: err=%v output=%q", err, output)
	}
	mustRunCommand(t, work, "git", "remote", "set-url", "origin", filepath.Join(root, "missing.git"))
	if output, err := selectEnv("branch", "main"); err == nil || !strings.Contains(output, "could not query origin") {
		t.Fatalf("unreachable origin: err=%v output=%q", err, output)
	}
}

func TestPublishWorkflowRunsLiveSmokeBeforeRelease(t *testing.T) {
	publish := readRepoFile(t, filepath.Join(".github", "workflows", "publish.yaml"))
	checks := []struct {
		name    string
		pattern string
	}{
		{"an earlier job selects the environment with the script", `\n  smoke-environment:\n    runs-on: ubuntu-latest\n    outputs:\n` +
			`      smoke_prefix: \$\{\{ steps\.smoke\.outputs\.smoke_prefix \}\}\n      environment: \$\{\{ steps\.smoke\.outputs\.environment \}\}\n` +
			`(?s:.*?)- name: Select the live smoke environment\n        id: smoke\n        run: \./scripts/select-smoke-environment\.sh >> "\$\{GITHUB_OUTPUT\}"\n\n  test:\n`},
		{"test job runs in the selected GitHub Environment", `\n  test:\n    needs: smoke-environment\n    runs-on: ubuntu-latest\n` +
			`    environment: \$\{\{ needs\.smoke-environment\.outputs\.environment \}\}\n    permissions:\n      contents: read\n    outputs:\n`},
		{"base URL comes from the selected Variables", `LAMBDADB_BASE_URL: \$\{\{ vars\[format\('\{0\}BASE_URL', needs\.smoke-environment\.outputs\.smoke_prefix\)\] \}\}\n`},
		{"project comes from the selected Variables", `LAMBDADB_PROJECT_NAME: \$\{\{ vars\[format\('\{0\}PROJECT_NAME', needs\.smoke-environment\.outputs\.smoke_prefix\)\] \}\}\n`},
		{"key comes from the Environment secret", `LAMBDADB_PROJECT_API_KEY: \$\{\{ secrets\.LAMBDADB_PROJECT_API_KEY \}\}\n`},
		{"smoke runs after the build in the test job", `- name: Build\n        run: go build -v \./\.\.\.\n\n(?s:.*)\./scripts/run-live-smoke\.sh\n\n  release:\n    needs: test\n`},
	}
	for _, check := range checks {
		if !regexp.MustCompile(check.pattern).MatchString(publish) {
			t.Errorf("%s: pattern %q not found", check.name, check.pattern)
		}
	}
	if count := strings.Count(publish, "LAMBDADB_PROJECT_API_KEY"); count != 2 {
		t.Errorf("publish.yaml references LAMBDADB_PROJECT_API_KEY %d times, want only the smoke step's secret", count)
	}
	for _, removed := range []string{"configure-aws-credentials", "SMOKE_AWS_ROLE_ARN", "SMOKE_SECRET_ARN", "id-token", "steps.smoke.outputs.smoke_prefix)"} {
		if strings.Contains(publish, removed) {
			t.Errorf("publish.yaml still contains %q", removed)
		}
	}
}
