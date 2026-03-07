package mesh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// CommitTrace links a git commit to the contract IDs mentioned in its message
// and the Go packages changed by that commit.
type CommitTrace struct {
	Hash        string   `json:"hash"`
	ContractIDs []string `json:"contract_ids"`
	Packages    []string `json:"packages"`
}

var contractIDRe = regexp.MustCompile(`\b(CON|BUG|SPEC|NEED|SPR|DIR|WATCH|BAT|ARCH|DOC|BND)-\d{4}-\d+\b`)

// TraceCommitContracts scans git history and extracts contract ID references from commit messages.
func TraceCommitContracts(root string) ([]CommitTrace, error) {
	cmd := exec.Command("git", "log", "--format=%H %s", "--name-only")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	return parseGitLog(string(out)), nil
}

var commitLineRe = regexp.MustCompile(`^[0-9a-f]{40} .+$`)

func parseGitLog(output string) []CommitTrace {
	var traces []CommitTrace
	var current *CommitTrace

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if commitLineRe.MatchString(line) {
			if current != nil && len(current.ContractIDs) > 0 {
				current.Packages = uniqueStrings(current.Packages)
				traces = append(traces, *current)
			}

			hash := line[:40]
			msg := line[41:]
			ids := contractIDRe.FindAllString(msg, -1)
			if len(ids) > 0 {
				current = &CommitTrace{
					Hash:        hash,
					ContractIDs: uniqueStrings(ids),
				}
			} else {
				current = nil
			}
			continue
		}

		if current != nil {
			pkg := goPackageDir(line)
			if pkg != "" {
				current.Packages = append(current.Packages, pkg)
			}
		}
	}

	if current != nil && len(current.ContractIDs) > 0 {
		current.Packages = uniqueStrings(current.Packages)
		traces = append(traces, *current)
	}

	return traces
}

// MapContractsToSpecs builds a mapping of contractID -> []specID by
// scanning specs that have a satisfies field pointing to needs, then
// finding contracts that also justify those needs.
func MapContractsToSpecs(root string) (map[string][]string, error) {
	mosDir := filepath.Join(root, ".mos")

	specToNeeds := make(map[string][]string)
	needToSpecs := make(map[string][]string)

	specsDir := filepath.Join(mosDir, "specifications")
	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(specsDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "specification.mos")
			ab, err := dsl.ReadArtifact(path)
			if err != nil {
				continue
			}
			satisfies := dslFieldStringSlice(ab.Items, "satisfies")
			if len(satisfies) > 0 {
				specToNeeds[e.Name()] = satisfies
				for _, needID := range satisfies {
					needToSpecs[needID] = append(needToSpecs[needID], e.Name())
				}
			}
		}
	}

	result := make(map[string][]string)

	contractsDir := filepath.Join(mosDir, "contracts")
	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(contractsDir, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "contract.mos")
			ab, err := dsl.ReadArtifact(path)
			if err != nil {
				continue
			}
			justifies := dslFieldStringSlice(ab.Items, "justifies")
			for _, needID := range justifies {
				if specs, ok := needToSpecs[needID]; ok {
					result[e.Name()] = append(result[e.Name()], specs...)
				}
			}
		}
	}

	for k, v := range result {
		result[k] = uniqueStrings(v)
	}
	return result, nil
}

// SpecUpdate describes a proposed change to a spec's include directives.
type SpecUpdate struct {
	SpecID           string   `json:"spec_id"`
	CurrentIncludes  []string `json:"current_includes"`
	InferredIncludes []string `json:"inferred_includes"`
	Added            []string `json:"added"`
}

// InferSpecIncludes combines commit traces and contract-to-spec mappings
// to propose include directives for specs.
func InferSpecIncludes(root string) ([]SpecUpdate, error) {
	traces, err := TraceCommitContracts(root)
	if err != nil {
		return nil, err
	}

	contractToSpecs, err := MapContractsToSpecs(root)
	if err != nil {
		return nil, err
	}

	specPackages := make(map[string][]string)
	for _, trace := range traces {
		for _, conID := range trace.ContractIDs {
			if specs, ok := contractToSpecs[conID]; ok {
				for _, specID := range specs {
					specPackages[specID] = append(specPackages[specID], trace.Packages...)
				}
			}
		}
	}

	mosDir := filepath.Join(root, ".mos")
	var updates []SpecUpdate

	for specID, pkgs := range specPackages {
		pkgs = uniqueStrings(pkgs)
		if len(pkgs) == 0 {
			continue
		}

		current := readSpecIncludes(mosDir, specID)
		currentSet := make(map[string]bool)
		for _, p := range current {
			currentSet[p] = true
		}

		var added []string
		for _, pkg := range pkgs {
			if !currentSet[pkg] {
				added = append(added, pkg)
			}
		}

		if len(added) > 0 {
			updates = append(updates, SpecUpdate{
				SpecID:           specID,
				CurrentIncludes:  current,
				InferredIncludes: pkgs,
				Added:            added,
			})
		}
	}

	return updates, nil
}

func readSpecIncludes(mosDir, specID string) []string {
	for _, sub := range []string{"active", "archive"} {
		path := filepath.Join(mosDir, "specifications", sub, specID, "specification.mos")
		ab, err := dsl.ReadArtifact(path)
		if err != nil {
			continue
		}
		return dslFieldStringSlice(ab.Items, "includes")
	}
	return nil
}

func dslFieldStringSlice(items []dsl.Node, key string) []string {
	val, ok := dsl.FieldString(items, key)
	if !ok || val == "" {
		return nil
	}
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func goPackageDir(path string) string {
	if !strings.HasSuffix(path, ".go") {
		return ""
	}
	dir := filepath.Dir(path)
	if dir == "." {
		return ""
	}
	return dir
}

func uniqueStrings(ss []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range ss {
		if s != "" && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}
