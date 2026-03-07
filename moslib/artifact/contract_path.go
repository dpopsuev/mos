package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func FindContractPath(root, id string) (string, error) {
	mosDir := filepath.Join(root, MosDir)
	for _, subDir := range []string{ActiveDir, ArchiveDir} {
		p := filepath.Join(mosDir, DirContracts, subDir, id, "contract.mos")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if resolved, err := resolveSlug(root, DirContracts, KindContract, id); err == nil {
		return resolved, nil
	}
	return "", fmt.Errorf("contract %q not found in .mos/contracts/", id)
}

// ContractInfo holds metadata about a discovered contract.
type ContractInfo struct {
	ID        string
	Title     string
	Status    string
	Path      string
	DependsOn []string
	CreatedAt string
	UpdatedAt string
	Progress  string
	Kind      string
	Labels    []string
	Priority  string
	Parent    string
	Branches  []string
	Specs     []string
}

// ListOpts configures contract list filtering.
type ListOpts struct {
	Status    string
	Project   string // filter by project name (prefix match on ID)
	Kind      string
	Label     string // single label; matches if contract has this label
	Priority  string
	Parent    string // filter by direct parent
	Branch    string // filter by branch scope (contains match)
	Roots     bool   // only return contracts with no parent
	Recursive bool   // when Parent is set, include all descendants
}

// ListContracts returns all contracts found under .mos/contracts/.
func ListContracts(root string, opts ListOpts) ([]ContractInfo, error) {
	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return nil, fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	var projectPrefix string
	if opts.Project != "" {
		projects, err := LoadProjects(root)
		if err == nil {
			for _, p := range projects {
				if p.Name == opts.Project {
					projectPrefix = p.Prefix + "-"
					break
				}
			}
		}
		if projectPrefix == "" {
			projectPrefix = opts.Project + "-"
		}
	}

	var contracts []ContractInfo
	for _, subDir := range []string{ActiveDir, ArchiveDir} {
		base := filepath.Join(mosDir, DirContracts, subDir)
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("ListContracts: %w", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			contractPath := filepath.Join(base, entry.Name(), "contract.mos")
			info, err := readContractInfo(entry.Name(), contractPath)
			if err != nil {
				continue
			}
			if opts.Status != "" && info.Status != opts.Status {
				continue
			}
			if projectPrefix != "" && !strings.HasPrefix(info.ID, projectPrefix) {
				continue
			}
			if opts.Kind != "" && info.Kind != opts.Kind {
				continue
			}
			if opts.Priority != "" && info.Priority != opts.Priority {
				continue
			}
			if opts.Label != "" {
				found := false
				for _, l := range info.Labels {
					if l == opts.Label {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if opts.Branch != "" {
				found := false
				for _, b := range info.Branches {
					if b == opts.Branch {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			if opts.Roots && info.Parent != "" {
				continue
			}
			if opts.Parent != "" && !opts.Recursive && info.Parent != opts.Parent {
				continue
			}
			contracts = append(contracts, info)
		}
	}

	if opts.Parent != "" && opts.Recursive {
		contracts = filterDescendants(contracts, opts.Parent)
	}

	return contracts, nil
}

func filterDescendants(all []ContractInfo, ancestor string) []ContractInfo {
	childMap := map[string][]string{}
	infoMap := map[string]ContractInfo{}
	for _, c := range all {
		infoMap[c.ID] = c
		if c.Parent != "" {
			childMap[c.Parent] = append(childMap[c.Parent], c.ID)
		}
	}

	var result []ContractInfo
	var walk func(parentID string)
	walk = func(parentID string) {
		for _, childID := range childMap[parentID] {
			if info, ok := infoMap[childID]; ok {
				result = append(result, info)
				walk(childID)
			}
		}
	}
	walk(ancestor)
	return result
}

func readContractInfo(id, path string) (ContractInfo, error) {
	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return ContractInfo{}, err
	}
	items := ab.Items
	info := ContractInfo{
		ID:        id,
		Path:      path,
		Title:     FieldStr(items, FieldTitle),
		Status:    FieldStr(items, FieldStatus),
		CreatedAt: FieldStr(items, "created_at"),
		UpdatedAt: FieldStr(items, "updated_at"),
		Kind:      FieldStr(items, "kind"),
		Priority:  FieldStr(items, "priority"),
		Parent:    FieldStr(items, "parent"),
		Labels:    dsl.FieldStringSlice(items, "labels"),
		Branches:  dsl.FieldStringSlice(items, "branches"),
		Specs:     dsl.FieldStringSlice(items, "specs"),
	}
	if scope := dsl.FindBlock(items, BlockScope); scope != nil {
		info.DependsOn = extractDependsOn(scope.Items)
	}
	if len(info.DependsOn) == 0 {
		info.DependsOn = extractDependsOn(items)
	}
	done, total, _, _ := computeProgress(ab)
	if total > 0 {
		info.Progress = fmt.Sprintf("%d/%d", done, total)
	}
	return info, nil
}

func extractDependsOn(items []dsl.Node) []string {
	return dsl.FieldStringSlice(items, "depends_on")
}

// ValidateDependencies checks that all dependency IDs refer to existing contracts.
func ValidateDependencies(root string, deps []string) error {
	for _, dep := range deps {
		if _, err := FindContractPath(root, dep); err != nil {
			return fmt.Errorf("dependency %q not found", dep)
		}
	}
	return nil
}

// ValidateParent checks that setting parentID as the parent of childID
// does not create a cycle in the hierarchy.
func ValidateParent(root, childID, parentID string) error {
	if parentID == "" {
		return nil
	}
	if parentID == childID {
		return fmt.Errorf("contract cannot be its own parent")
	}
	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		return fmt.Errorf("ValidateParent: %w", err)
	}
	parentMap := map[string]string{}
	for _, c := range all {
		if c.Parent != "" {
			parentMap[c.ID] = c.Parent
		}
	}
	parentMap[childID] = parentID

	visited := map[string]bool{}
	cur := childID
	for {
		if visited[cur] {
			return fmt.Errorf("cycle detected: setting parent of %s to %s creates a loop", childID, parentID)
		}
		visited[cur] = true
		next, ok := parentMap[cur]
		if !ok || next == "" {
			break
		}
		cur = next
	}
	return nil
}

// sliceRefValue converts a string slice to a DSL Value.
// Single-element slices produce a StringVal for backward compatibility.
func sliceRefValue(vals []string) dsl.Value {
	if len(vals) == 1 {
		return &dsl.StringVal{Text: vals[0]}
	}
	items := make([]dsl.Value, len(vals))
	for i, v := range vals {
		items[i] = &dsl.StringVal{Text: v}
	}
	return &dsl.ListVal{Items: items}
}

func filterNodes(items []dsl.Node, isMatch func(dsl.Node) bool) []dsl.Node {
	filtered := make([]dsl.Node, 0, len(items))
	for _, item := range items {
		if !isMatch(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func touchUpdatedAt(ab *dsl.ArtifactBlock) {
	now := time.Now().UTC().Format(time.RFC3339)
	dsl.SetField(&ab.Items, "updated_at", &dsl.DateTimeVal{Raw: now})
}
