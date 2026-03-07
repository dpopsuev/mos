package artifact

import (
	"fmt"
	"os"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ShowContract reads and returns the formatted content of a contract.
func ShowContract(root, id string) (string, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return "", fmt.Errorf("ShowContract: %w", err)
	}
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return "", fmt.Errorf("reading contract: %w", err)
	}
	f, err := dsl.Parse(string(data), nil) // needs full File for dsl.Format
	if err != nil {
		return "", fmt.Errorf("parsing contract: %w", err)
	}
	return dsl.Format(f, nil), nil
}

// ShowContractVerbose returns the full formatted content of a contract.
func ShowContractVerbose(root, id string) (string, error) {
	return ShowContract(root, id)
}

// Summary holds the short-form view of a contract.
type Summary struct {
	ID             string
	Title          string
	Status         string
	Progress       string
	Prev           string
	Current        string
	Next           string
	DependsOn      []string
	DependedOnBy   []string
	LockedBy       string
	CreatedAt      string
	UpdatedAt      string
	Kind           string
	Labels         []string
	Priority       string
	Parent         string
	Branches       []string
	Children       []string
	RollupProgress string
	Specs          []string
	Goal           string
	Precondition   string
	Justifies      string
	Sprint         string
	Batch          string
}

// ContractSummary builds a short-form summary of a contract.
func ContractSummary(root, id string) (*Summary, error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return nil, fmt.Errorf("ContractSummary: %w", err)
	}
	ab, err := dsl.ReadArtifact(contractPath)
	if err != nil {
		return nil, fmt.Errorf("ContractSummary: %w", err)
	}

	items := ab.Items
	s := &Summary{ID: id}
	s.Title, _ = dsl.FieldString(items, FieldTitle)
	s.Status, _ = dsl.FieldString(items, FieldStatus)
	s.Kind, _ = dsl.FieldString(items, "kind")
	s.Priority, _ = dsl.FieldString(items, "priority")
	s.Parent, _ = dsl.FieldString(items, "parent")
	s.Goal, _ = dsl.FieldString(items, FieldGoal)
	s.Labels = dsl.FieldStringSlice(items, "labels")
	s.Branches = dsl.FieldStringSlice(items, "branches")
	s.Specs = dsl.FieldStringSlice(items, "specs")
	if f := dsl.FindField(items, "created_at"); f != nil {
		if dv, ok := f.Value.(*dsl.DateTimeVal); ok {
			s.CreatedAt = dv.Raw
		}
	}
	if f := dsl.FindField(items, "updated_at"); f != nil {
		if dv, ok := f.Value.(*dsl.DateTimeVal); ok {
			s.UpdatedAt = dv.Raw
		}
	}
	s.Precondition = strings.Join(dsl.FieldStringSlice(items, "precondition"), ", ")
	s.Justifies = strings.Join(dsl.FieldStringSlice(items, "justifies"), ", ")
	s.Sprint = strings.Join(dsl.FieldStringSlice(items, "sprint"), ", ")
	s.Batch = strings.Join(dsl.FieldStringSlice(items, "batch"), ", ")
	if scope := dsl.FindBlock(items, BlockScope); scope != nil {
		s.DependsOn = extractDependsOn(scope.Items)
	}
	if len(s.DependsOn) == 0 {
		s.DependsOn = extractDependsOn(items)
	}

	done, total, current, names := computeProgress(ab)
	if total > 0 {
		s.Progress = fmt.Sprintf("%d/%d", done, total)
	}
	s.Current = current

	for i, name := range names {
		if name == current {
			if i > 0 {
				s.Prev = names[i-1]
			}
			if i+1 < len(names) {
				s.Next = names[i+1]
			}
			break
		}
	}

	seal, _ := CheckSeal(root, id)
	if seal != nil {
		s.LockedBy = seal.Operator
	}

	dependents, _ := FindDependents(root, id)
	s.DependedOnBy = dependents

	childInfos, _ := FindChildren(root, id)
	for _, ch := range childInfos {
		s.Children = append(s.Children, ch.ID)
	}

		if len(childInfos) > 0 {
		rollupDone, rollupTotal := 0, 0
		for _, ch := range childInfos {
			chPath, err := FindContractPath(root, ch.ID)
			if err != nil {
				continue
			}
			chAB, err := dsl.ReadArtifact(chPath)
			if err != nil {
				continue
			}
			d, t, _, _ := computeProgress(chAB)
			if t > 0 {
				rollupDone += d
				rollupTotal += t
			} else {
				rollupTotal++
				if ch.Status == StatusComplete {
					rollupDone++
				}
			}
		}
		if rollupTotal > 0 {
			s.RollupProgress = fmt.Sprintf("%d/%d", rollupDone, rollupTotal)
		}
	}

	return s, nil
}

// ContractProgress returns progress info for a contract.
func ContractProgress(root, id string) (done, total int, current string, err error) {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return 0, 0, "", fmt.Errorf("ContractProgress: %w", err)
	}
	ab, err := dsl.ReadArtifact(contractPath)
	if err != nil {
		return 0, 0, "", fmt.Errorf("ContractProgress: %w", err)
	}
	done, total, current, _ = computeProgress(ab)
	return done, total, current, nil
}

func computeProgress(ab *dsl.ArtifactBlock) (done, total int, current string, names []string) {
	walkBlocks(ab.Items, &done, &total, &current, &names)
	return
}

func walkBlocks(items []dsl.Node, done, total *int, current *string, names *[]string) {
	for _, item := range items {
		switch n := item.(type) {
		case *dsl.Block:
			status := blockStatus(n.Items)
			if status != "" {
				name := n.Title
				if name == "" {
					name = n.Name
				}
				*names = append(*names, name)
				*total++
				if status == "done" {
					*done++
				}
				if status != "done" && *current == "" {
					*current = name
				}
			}
			walkBlocks(n.Items, done, total, current, names)
		case *dsl.FeatureBlock:
			walkFeatureBlock(n, done, total, current, names)
		}
	}
}

func walkFeatureBlock(fb *dsl.FeatureBlock, done, total *int, current *string, names *[]string) {
	for _, group := range fb.Groups {
		switch g := group.(type) {
		case *dsl.Scenario:
			status := scenarioStatus(g)
			if status != "" {
				*names = append(*names, g.Name)
				*total++
				if status == "done" {
					*done++
				}
				if status != "done" && *current == "" {
					*current = g.Name
				}
			}
		case *dsl.Group:
			for _, sc := range g.Scenarios {
				status := scenarioStatus(sc)
				if status != "" {
					*names = append(*names, sc.Name)
					*total++
					if status == "done" {
						*done++
					}
					if status != "done" && *current == "" {
						*current = sc.Name
					}
				}
			}
		}
	}
}

func blockStatus(items []dsl.Node) string {
	s, _ := dsl.FieldString(items, FieldStatus)
	return s
}

func scenarioStatus(s *dsl.Scenario) string {
	items := make([]dsl.Node, len(s.Fields))
	for i := range s.Fields {
		items[i] = s.Fields[i]
	}
	st, _ := dsl.FieldString(items, FieldStatus)
	return st
}

// ShowContractShort returns a short-form summary of a contract.
func ShowContractShort(root, id string) (string, error) {
	s, err := ContractSummary(root, id)
	if err != nil {
		return "", fmt.Errorf("ShowContractShort: %w", err)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s: %s [%s]\n", s.ID, s.Title, s.Status)
	if s.Kind != "" {
		fmt.Fprintf(&b, "Kind: %s\n", s.Kind)
	}
	if s.Priority != "" {
		fmt.Fprintf(&b, "Priority: %s\n", s.Priority)
	}
	if s.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", s.Goal)
	}
	if s.Precondition != "" {
		fmt.Fprintf(&b, "Precondition: %s\n", s.Precondition)
	}
	if s.Justifies != "" {
		fmt.Fprintf(&b, "Justifies: %s\n", s.Justifies)
	}
	if s.Sprint != "" {
		fmt.Fprintf(&b, "Sprint: %s\n", s.Sprint)
	}
	if s.Batch != "" {
		fmt.Fprintf(&b, "Batch: %s\n", s.Batch)
	}
	if len(s.Labels) > 0 {
		fmt.Fprintf(&b, "Labels: %s\n", strings.Join(s.Labels, ", "))
	}
	if s.Parent != "" {
		fmt.Fprintf(&b, "Parent: %s\n", s.Parent)
	}
	if len(s.Branches) > 0 {
		fmt.Fprintf(&b, "Branches: %s\n", strings.Join(s.Branches, ", "))
	}
	if len(s.Children) > 0 {
		fmt.Fprintf(&b, "Children: %s\n", strings.Join(s.Children, ", "))
	}
	if len(s.Specs) > 0 {
		fmt.Fprintf(&b, "Specs: %s\n", strings.Join(s.Specs, ", "))
	}
	if s.RollupProgress != "" {
		fmt.Fprintf(&b, "Rollup: %s\n", s.RollupProgress)
	}
	if s.Progress != "" {
		fmt.Fprintf(&b, "Progress: %s\n", s.Progress)
	}
	if s.Current != "" {
		if s.Prev != "" {
			fmt.Fprintf(&b, "Prev: %s\n", s.Prev)
		}
		fmt.Fprintf(&b, "Current: %s\n", s.Current)
		if s.Next != "" {
			fmt.Fprintf(&b, "Next: %s\n", s.Next)
		}
	} else if s.Progress != "" {
		fmt.Fprintf(&b, "Status: all complete\n")
	}
	if len(s.DependsOn) > 0 {
		fmt.Fprintf(&b, "Depends on: %s\n", strings.Join(s.DependsOn, ", "))
	}
	if len(s.DependedOnBy) > 0 {
		fmt.Fprintf(&b, "Depended on by: %s\n", strings.Join(s.DependedOnBy, ", "))
	}
	if s.LockedBy != "" {
		fmt.Fprintf(&b, "Locked by: %s\n", s.LockedBy)
	}
	if s.CreatedAt != "" {
		fmt.Fprintf(&b, "Created: %s\n", s.CreatedAt)
	}
	if s.UpdatedAt != "" {
		fmt.Fprintf(&b, "Updated: %s\n", s.UpdatedAt)
	}
	return b.String(), nil
}
