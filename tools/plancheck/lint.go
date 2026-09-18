package main

import (
	"fmt"
	"regexp"
	"strings"
)

// Required fields per lane, in the order the planning workflow defines them.
var (
	agentFields = []string{
		"Serves", "Agent status", "Dependencies", "User-visible outcome", "Scope boundary",
		"Data and artifact paths", "Execution path", "Acceptance gates", "Documentation target",
		"Review checkpoint",
	}
	humanFields = []string{
		"Serves", "Human status", "Dependencies", "Requested input or decision", "Unblocks",
	}
	agentStatuses = map[string]bool{"CLEAR": true, "RUN NEEDED": true}
	humanStatuses = map[string]bool{"HUMAN-GATED": true, "BLOCKED BY HUMAN": true}
)

var (
	recordLinkRe = regexp.MustCompile(`\(records/([0-9]{4}-[a-z0-9-]+)\.md(?:#[^)]*)?\)`)
	recordIDRe   = regexp.MustCompile("(?m)^- Id / capability / checkpoint: `([a-z0-9-]+)`")
)

// Lint returns every integrity problem found; an empty slice means the documents agree.
func Lint(in Inputs) []string {
	var errs []string
	add := func(format string, a ...any) { errs = append(errs, fmt.Sprintf(format, a...)) }

	order := map[string]int{}
	for i, c := range in.Registry {
		if _, dup := order[c.ID]; dup {
			add("registry: capability %q listed twice", c.ID)
		}
		order[c.ID] = i
		if c.Status != "planned" && c.Status != "shipped" {
			add("registry: capability %q has status %q, want planned or shipped", c.ID, c.Status)
		}
		if c.Status == "shipped" && !strings.Contains(c.Implementation, "impl/current/") {
			add("registry: shipped capability %q must link its current-state page", c.ID)
		}
		if _, ok := in.Groups[c.ID]; !ok {
			add("registry: capability %q has no record group in the planning workflow table", c.ID)
		}
	}
	if len(in.Registry) == 0 {
		add("registry: no capabilities parsed from the specification")
	}

	lastByLane := map[string]int{}
	for _, g := range in.Plan.Groups {
		idx, ok := order[g.Capability]
		if !ok {
			add("plan:%d: group %q is not in the registry", g.Line, g.Capability)
			continue
		}
		if prev, seen := lastByLane[g.Lane]; seen && idx <= prev {
			add("plan:%d: group %q is out of registry order in the %s lane", g.Line, g.Capability, g.Lane)
		}
		lastByLane[g.Lane] = idx
	}

	open := map[string]Task{}
	tasksPerCap := map[string]int{}
	for _, t := range in.Plan.Tasks {
		if _, dup := open[t.ID]; dup {
			add("plan:%d: duplicate task id %q", t.Line, t.ID)
		}
		open[t.ID] = t
		tasksPerCap[t.Capability]++
	}

	records := map[string]bool{}
	sequences := map[string]string{}
	for _, r := range in.Records {
		records[strings.TrimSuffix(r.File, ".md")] = true
		if prev, dup := sequences[r.Sequence]; dup {
			add("record %s: sequence %s is already used by %s", r.File, r.Sequence, prev)
		}
		sequences[r.Sequence] = r.File
		if m := recordIDRe.FindStringSubmatch(r.Body); m == nil || m[1] != r.TaskID {
			add("record %s: first scope line must be \"- Id / capability / checkpoint: `%s` ...\"", r.File, r.TaskID)
		}
		if !strings.Contains(in.RecordIndex, "("+r.File+")") {
			add("record %s: not linked from the records README", r.File)
		}
		if _, still := open[r.TaskID]; still {
			add("record %s: task %q is still in the plan", r.File, r.TaskID)
		}
	}

	for _, t := range in.Plan.Tasks {
		errs = append(errs, lintTask(t, open, records)...)
	}

	for _, c := range in.Registry {
		switch {
		case c.Status == "planned" && tasksPerCap[c.ID] == 0:
			add("registry: planned capability %q has no open tasks", c.ID)
		case c.Status == "shipped" && tasksPerCap[c.ID] > 0:
			add("registry: shipped capability %q still has %d open task(s)", c.ID, tasksPerCap[c.ID])
		}
	}

	if cycle := findCycle(in.Plan.Tasks); cycle != nil {
		add("plan: dependency cycle %s", strings.Join(cycle, " -> "))
	}
	return errs
}

// lintTask checks one task; records holds record file names without ".md".
func lintTask(t Task, open map[string]Task, records map[string]bool) []string {
	var errs []string
	add := func(format string, a ...any) {
		errs = append(errs, fmt.Sprintf("plan:%d: task %q: ", t.Line, t.ID)+fmt.Sprintf(format, a...))
	}
	required, statusField, statuses := agentFields, "Agent status", agentStatuses
	if t.Lane == LaneHuman {
		required, statusField, statuses = humanFields, "Human status", humanStatuses
	}
	if t.Capability == "" {
		add("not under a capability group heading")
	}
	for _, f := range required {
		if strings.TrimSpace(t.Fields[f]) == "" {
			add("missing field %q", f)
		}
	}
	if s := t.Fields[statusField]; s != "" && !statuses[s] {
		add("%s %q is not valid in the %s lane", statusField, s, t.Lane)
	}
	if serves := t.Fields["Serves"]; serves != "" && !strings.HasPrefix(serves, "`"+t.Capability+"`") {
		add("Serves must start with `%s`", t.Capability)
	}
	for _, id := range dependencyIDs(t.Fields["Dependencies"]) {
		if _, ok := open[id]; !ok {
			add("dependency `%s` is not an open task (link accepted work as records/...)", id)
		}
	}
	for _, m := range recordLinkRe.FindAllStringSubmatch(t.Fields["Dependencies"], -1) {
		if !records[m[1]] {
			add("dependency record %s.md does not exist", m[1])
		}
	}
	if t.Lane == LaneHuman {
		for _, id := range dependencyIDs(t.Fields["Unblocks"]) {
			if _, ok := open[id]; !ok {
				add("unblocks `%s`, which is not an open task", id)
			}
		}
	}
	return errs
}

// dependencyIDs returns the backticked task ids in a field, ignoring record links.
func dependencyIDs(field string) []string {
	var ids []string
	for _, m := range backtickRe.FindAllStringSubmatch(recordLinkRe.ReplaceAllString(field, ""), -1) {
		ids = append(ids, m[1])
	}
	return ids
}

// findCycle returns one dependency cycle among open tasks, or nil.
func findCycle(tasks []Task) []string {
	deps := map[string][]string{}
	for _, t := range tasks {
		deps[t.ID] = dependencyIDs(t.Fields["Dependencies"])
	}
	const (
		visiting = iota + 1 // the zero value means unvisited
		done
	)
	state := map[string]int{}
	var stack []string
	var visit func(string) []string
	visit = func(id string) []string {
		switch state[id] {
		case visiting:
			for i, s := range stack {
				if s == id {
					return append(append([]string{}, stack[i:]...), id)
				}
			}
		case done:
			return nil
		}
		state[id] = visiting
		stack = append(stack, id)
		for _, d := range deps[id] {
			if _, known := deps[d]; known {
				if c := visit(d); c != nil {
					return c
				}
			}
		}
		stack = stack[:len(stack)-1]
		state[id] = done
		return nil
	}
	for _, t := range tasks {
		if c := visit(t.ID); c != nil {
			return c
		}
	}
	return nil
}
