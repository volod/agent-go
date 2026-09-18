package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

const testSpec = "# Spec\n\n## Capability Registry\n\n" +
	"| # | Capability | Status | Implementation |\n| --- | --- | --- | --- |\n" +
	"| 1 | `alpha` | shipped | [Current](../impl/current/alpha.md) |\n" +
	"| 2 | `beta` | planned | [Open work](../impl/plan.md) |\n\n## After\n"

const testWorkflow = "# Workflow\n\n| Capability id | Group abbrev |\n| --- | --- |\n" +
	"| `alpha` | `al` |\n| `beta` | `be` |\n"

func agentTask(id, deps string) string {
	return "#### " + id + "\n\nDo it.\n\n" +
		"- Serves: `beta` -- [Spec](../openspec/spec.md)\n- Agent status: CLEAR\n" +
		"- Dependencies: " + deps + "\n- User-visible outcome: o\n- Scope boundary: s\n" +
		"- Data and artifact paths: p\n- Execution path:\n  multi-line\n  execution\n- Acceptance gates: g\n" +
		"- Documentation target: d\n- Review checkpoint: none.\n\n"
}

func validPlan() string {
	return "# Plan\n\n## Agent Implementation Tasks\n\n### Beta -- `beta`\n\n" +
		agentTask("build-beta", "[First](records/0001-al-first-task.md).") +
		agentTask("finish-beta", "`build-beta`; `wait-human`.") +
		"## Human-Assisted Tasks\n\n### Beta -- `beta`\n\n#### wait-human\n\nDecide.\n\n" +
		"- Serves: `beta` -- x\n- Human status: HUMAN-GATED\n- Dependencies: none.\n" +
		"- Requested input or decision: d\n- Unblocks: `finish-beta`.\n"
}

func testRepo(plan string) fstest.MapFS {
	return fstest.MapFS{
		SpecPath:                              {Data: []byte(testSpec)},
		WorkflowPath:                          {Data: []byte(testWorkflow)},
		PlanPath:                              {Data: []byte(plan)},
		RecordsDir + "/README.md":             {Data: []byte("| [0001](0001-al-first-task.md) | x |\n")},
		RecordsDir + "/template.md":           {Data: []byte("# Template\n")},
		RecordsDir + "/0001-al-first-task.md": {Data: []byte("# First\n\n- Id / capability / checkpoint: `first-task` / `alpha` / none\n")},
	}
}

func lintRepo(t *testing.T, fsys fstest.MapFS) []string {
	t.Helper()
	in, problems, err := Load(fsys)
	if err != nil {
		t.Fatal(err)
	}
	return append(problems, Lint(in)...)
}

func TestLintAcceptsConsistentDocuments(t *testing.T) {
	if errs := lintRepo(t, testRepo(validPlan())); len(errs) != 0 {
		t.Fatalf("unexpected problems:\n%s", strings.Join(errs, "\n"))
	}
}

func TestParsePlanJoinsMultilineFields(t *testing.T) {
	plan := ParsePlan(validPlan())
	if got := plan.Tasks[0].Fields["Execution path"]; got != "multi-line execution" {
		t.Fatalf("Execution path = %q, want %q", got, "multi-line execution")
	}
}

func TestLintDetectsDefects(t *testing.T) {
	replace := func(old, new string) func() fstest.MapFS {
		return func() fstest.MapFS { return testRepo(strings.Replace(validPlan(), old, new, 1)) }
	}
	withFile := func(name, data string) func() fstest.MapFS {
		return func() fstest.MapFS {
			fsys := testRepo(validPlan())
			fsys[name] = &fstest.MapFile{Data: []byte(data)}
			return fsys
		}
	}
	cases := []struct {
		name string
		fsys func() fstest.MapFS
		want string
	}{
		{"unknown dependency", replace("`build-beta`;", "`ghost-task`;"), "dependency `ghost-task` is not an open task"},
		{"missing record", replace("0001-al-first-task", "0009-al-gone"), "0009-al-gone.md does not exist"},
		{"cycle", replace("[First](records/0001-al-first-task.md).", "`finish-beta`."), "dependency cycle"},
		{"missing field", replace("- Scope boundary: s\n", ""), `missing field "Scope boundary"`},
		{"bad status", replace("Agent status: CLEAR", "Agent status: DONE"), `Agent status "DONE"`},
		{"wrong serves", replace("- Serves: `beta`", "- Serves: `alpha`"), "Serves must start with `beta`"},
		{"group not in registry", replace("### Beta -- `beta`", "### Gamma -- `gamma`"), `group "gamma" is not in the registry`},
		{"accepted task still planned", replace("#### build-beta", "#### first-task"), `task "first-task" is still in the plan`},
		{"planned capability without tasks", func() fstest.MapFS {
			return testRepo("# Plan\n\n## Agent Implementation Tasks\n")
		}, `planned capability "beta" has no open tasks`},
		{"shipped capability without current page", withFile(SpecPath,
			strings.Replace(testSpec, "../impl/current/alpha.md", "../impl/plan.md", 1)),
			`shipped capability "alpha" must link its current-state page`},
		{"record not indexed", withFile(RecordsDir+"/README.md", "empty\n"), "not linked from the records README"},
		{"record with unknown group", withFile(RecordsDir+"/0002-zz-other.md", "x"), "unknown group"},
		{"reused record sequence", withFile(RecordsDir+"/0001-be-other-task.md",
			"- Id / capability / checkpoint: `other-task` / `beta` / none\n"), "sequence 0001 is already used"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := lintRepo(t, tc.fsys())
			if !strings.Contains(strings.Join(errs, "\n"), tc.want) {
				t.Fatalf("want a problem containing %q, got:\n%s", tc.want, strings.Join(errs, "\n"))
			}
		})
	}
}

func TestStatusReportsEligibleWork(t *testing.T) {
	s := ComputeStatus(ParsePlan(validPlan()))
	if s.AgentTasks != 2 || s.HumanTasks != 1 {
		t.Fatalf("counts = %d agent, %d human, want 2 and 1", s.AgentTasks, s.HumanTasks)
	}
	if len(s.Eligible) != 1 || s.Eligible[0].ID != "build-beta" {
		t.Fatalf("eligible = %+v, want only build-beta", s.Eligible)
	}
	if len(s.WaitingHuman) != 1 || s.WaitingHuman[0].ID != "wait-human" {
		t.Fatalf("waiting human = %+v, want only wait-human", s.WaitingHuman)
	}
}

func TestCheckLinks(t *testing.T) {
	fsys := fstest.MapFS{
		"README.md":                   {Data: []byte("[ok](docs/a.md#archive----core) [web](https://x.y) [self](#top)\n# Top\n")},
		"docs/a.md":                   {Data: []byte("# A\n## Archive -- `core`\n[bad](missing.md)\n[anchor](#nope)\n```\n[fenced](ignored.md)\n```\n")},
		".git/config.md":              {Data: []byte("[x](nowhere.md)")},
		"test/testdata/sample/out.md": {Data: []byte("[file](../../../../mnt/clip.mp4)\n")},
	}
	errs, err := CheckLinks(fsys)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(errs, "\n")
	for _, want := range []string{`docs/a.md:3: broken link "missing.md"`, `docs/a.md:4: missing anchor "#nope"`} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing problem %q in:\n%s", want, joined)
		}
	}
	if len(errs) != 2 {
		t.Errorf("want exactly 2 problems, got:\n%s", joined)
	}
}

func TestSlugMatchesGitHub(t *testing.T) {
	cases := map[string]string{
		"Project identity -- `project-identity`": "project-identity----project-identity",
		"Run lock":                               "run-lock",
		"Video -- Previews (v2.0)!":              "video----previews-v20",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRun(t *testing.T) {
	root := t.TempDir()
	if err := os.CopyFS(root, testRepo(validPlan())); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		args       []string
		wantCode   int
		wantStdout string
	}{
		{[]string{"-root", root, "lint"}, 0, "lint-spec-plan: ok"},
		{[]string{"-root", root, "status"}, 0, "next agent task: build-beta (beta, CLEAR)"},
		{[]string{"-root", root, "links"}, 1, ""}, // the fixture links to files it does not contain
		{[]string{"-root", root, "bogus"}, 2, ""},
		{nil, 2, ""},
	}
	for _, tc := range cases {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run(tc.args, &stdout, &stderr); got != tc.wantCode {
				t.Errorf("exit code = %d, want %d; stderr:\n%s", got, tc.wantCode, stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tc.wantStdout)
			}
		})
	}
}
