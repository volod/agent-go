package main

import (
	"cmp"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
)

// Repository-relative locations of the planning documents.
const (
	SpecPath     = "docs/openspec/spec.md"
	PlanPath     = "docs/impl/plan.md"
	WorkflowPath = "docs/guide/planning-workflow.md"
	RecordsDir   = "docs/impl/records"
)

// Lane names of the plan's top-level sections.
const (
	LaneAgent = "agent"
	LaneHuman = "human"
)

var laneHeadings = map[string]string{
	"## Agent Implementation Tasks": LaneAgent,
	"## Human-Assisted Tasks":       LaneHuman,
}

// nonRecordFiles live in the records directory but are not task records.
var nonRecordFiles = map[string]bool{"README.md": true, "template.md": true}

var (
	groupHeadingRe = regexp.MustCompile("^### .+ -- `([a-z0-9-]+)`\\s*$")
	taskHeadingRe  = regexp.MustCompile(`^#### ([a-z0-9-]+)(\s+\(optional\))?\s*$`)
	fieldRe        = regexp.MustCompile(`^- ([A-Z][A-Za-z -]+): ?(.*)$`)
	backtickRe     = regexp.MustCompile("`([a-z0-9][a-z0-9-]*)`")
)

// Task is one "#### task-id" block of the plan.
type Task struct {
	ID         string
	Capability string
	Lane       string
	Line       int
	Fields     map[string]string
}

// Group is one "### Title -- `capability`" heading inside a lane.
type Group struct {
	Capability string
	Lane       string
	Line       int
}

// Plan is the parsed forward plan.
type Plan struct {
	Groups []Group
	Tasks  []Task
}

// Capability is one row of the specification registry.
type Capability struct {
	ID             string
	Status         string
	Implementation string
}

// Record is one task record file under RecordsDir.
type Record struct {
	File     string
	Sequence string
	Group    string
	TaskID   string
	Body     string
}

// Inputs are the documents a lint run reads.
type Inputs struct {
	Registry    []Capability
	Plan        Plan
	Groups      map[string]string // capability id -> record group abbreviation
	Records     []Record
	RecordIndex string // content of the records README
}

// Load reads every planning document from the repository root fsys. Record
// naming errors are returned as lint problems rather than failures.
func Load(fsys fs.FS) (Inputs, []string, error) {
	var in Inputs
	texts := map[string]string{}
	for _, p := range []string{SpecPath, PlanPath, WorkflowPath, path.Join(RecordsDir, "README.md")} {
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return in, nil, err
		}
		texts[p] = string(b)
	}
	in.Registry = ParseRegistry(texts[SpecPath])
	in.Plan = ParsePlan(texts[PlanPath])
	in.Groups = ParseGroupTable(texts[WorkflowPath])
	in.RecordIndex = texts[path.Join(RecordsDir, "README.md")]

	entries, err := fs.ReadDir(fsys, RecordsDir)
	if err != nil {
		return in, nil, err
	}
	var problems []string
	for _, e := range entries {
		if e.IsDir() || nonRecordFiles[e.Name()] || path.Ext(e.Name()) != ".md" {
			continue
		}
		r, err := ParseRecordName(e.Name(), in.Groups)
		if err != nil {
			problems = append(problems, err.Error())
			continue
		}
		b, err := fs.ReadFile(fsys, path.Join(RecordsDir, e.Name()))
		if err != nil {
			return in, nil, err
		}
		r.Body = string(b)
		in.Records = append(in.Records, r)
	}
	slices.SortFunc(in.Records, func(a, b Record) int { return cmp.Compare(a.File, b.File) })
	return in, problems, nil
}

// ParsePlan reads plan.md content. Field values may continue on indented lines.
func ParsePlan(text string) Plan {
	var (
		plan       Plan
		lane       string
		capability string
		cur        *Task
		field      string
	)
	flush := func() {
		if cur != nil {
			plan.Tasks = append(plan.Tasks, *cur)
		}
		cur, field = nil, ""
	}
	for i, line := range strings.Split(text, "\n") {
		n := i + 1
		switch {
		case strings.HasPrefix(line, "## "):
			flush()
			capability, lane = "", laneHeadings[strings.TrimSpace(line)]
		case strings.HasPrefix(line, "### "):
			flush()
			capability = ""
			if m := groupHeadingRe.FindStringSubmatch(line); m != nil && lane != "" {
				capability = m[1]
				plan.Groups = append(plan.Groups, Group{Capability: capability, Lane: lane, Line: n})
			}
		case strings.HasPrefix(line, "#### "):
			flush()
			if m := taskHeadingRe.FindStringSubmatch(line); m != nil && lane != "" {
				cur = &Task{ID: m[1], Capability: capability, Lane: lane, Line: n, Fields: map[string]string{}}
			}
		case cur == nil: // prose outside a task block
		case fieldRe.MatchString(line):
			m := fieldRe.FindStringSubmatch(line)
			field = m[1]
			cur.Fields[field] = strings.TrimSpace(m[2])
		case strings.TrimSpace(line) == "":
			field = ""
		case field != "":
			cur.Fields[field] = strings.TrimSpace(cur.Fields[field] + " " + strings.TrimSpace(line))
		}
	}
	flush()
	return plan
}

// ParseRegistry reads the "## Capability Registry" table of spec.md. The
// capability id is the first cell that is a single backticked id; the status
// and implementation cells are found by their column headings.
func ParseRegistry(text string) []Capability {
	var caps []Capability
	in := false
	statusCol, implCol := -1, -1
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "## ") {
			in = strings.TrimSpace(line) == "## Capability Registry"
			continue
		}
		if !in || !strings.HasPrefix(line, "|") {
			continue
		}
		cells := splitRow(line)
		if statusCol < 0 {
			statusCol, implCol = slices.Index(cells, "Status"), slices.Index(cells, "Implementation")
			continue
		}
		for _, c := range cells {
			if m := backtickRe.FindStringSubmatch(c); m != nil && m[0] == c {
				caps = append(caps, Capability{ID: m[1], Status: cell(cells, statusCol), Implementation: cell(cells, implCol)})
				break
			}
		}
	}
	return caps
}

// ParseGroupTable reads the capability-to-record-group table of the planning
// workflow: rows of the form "| `capability` | `abbrev` |".
func ParseGroupTable(text string) map[string]string {
	groups := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		if !strings.HasPrefix(line, "|") {
			continue
		}
		cells := splitRow(line)
		if len(cells) != 2 {
			continue
		}
		c, a := backtickRe.FindStringSubmatch(cells[0]), backtickRe.FindStringSubmatch(cells[1])
		if c != nil && a != nil {
			groups[c[1]] = a[1]
		}
	}
	return groups
}

var recordNameRe = regexp.MustCompile(`^([0-9]{4})-([a-z0-9-]+)\.md$`)

// ParseRecordName splits "NNNN-group-task-id.md" using the known group
// abbreviations, longest first.
func ParseRecordName(name string, groups map[string]string) (Record, error) {
	m := recordNameRe.FindStringSubmatch(name)
	if m == nil {
		return Record{}, fmt.Errorf("record %s: name must be NNNN-<group>-<task-id>.md", name)
	}
	abbrevs := slices.Collect(maps.Values(groups))
	slices.SortFunc(abbrevs, func(a, b string) int { return cmp.Compare(len(b), len(a)) })
	for _, a := range abbrevs {
		if rest, ok := strings.CutPrefix(m[2], a+"-"); ok && rest != "" {
			return Record{File: name, Sequence: m[1], Group: a, TaskID: rest}, nil
		}
	}
	return Record{}, fmt.Errorf("record %s: unknown group; add it to the planning workflow table", name)
}

func splitRow(line string) []string {
	parts := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func cell(cells []string, i int) string {
	if i < 0 || i >= len(cells) {
		return ""
	}
	return cells[i]
}
