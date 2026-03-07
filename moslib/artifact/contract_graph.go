package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ContractGraph returns all contracts with DependsOn populated.
func ContractGraph(root string) ([]ContractInfo, error) {
	return ListContracts(root, ListOpts{})
}

// LinkContract adds dependsOnID to the depends_on list of contract id.
// Idempotent: if the dependency already exists, this is a no-op.
func LinkContract(root, id, dependsOnID string) error {
	if err := CheckSealForMutation(root, id); err != nil {
		return fmt.Errorf("LinkContract: %w", err)
	}
	if err := ValidateDependencies(root, []string{dependsOnID}); err != nil {
		return fmt.Errorf("LinkContract: %w", err)
	}

	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("LinkContract: %w", err)
	}

	all, _ := ContractGraph(root)
	simulated := slices.Clone(all)
	for i, c := range simulated {
		if c.ID == id {
			simulated[i].DependsOn = append(slices.Clone(c.DependsOn), dependsOnID)
		}
	}
	if cycles := DetectCycles(simulated); len(cycles) > 0 {
		return fmt.Errorf("adding dependency %s -> %s would create a cycle: %v", id, dependsOnID, cycles[0])
	}

	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		var depField *dsl.Field
		scopeBlock := dsl.FindBlock(ab.Items, BlockScope)
		if scopeBlock != nil {
			depField = dsl.FindField(scopeBlock.Items, "depends_on")
		}
		if depField == nil {
			depField = dsl.FindField(ab.Items, "depends_on")
		}
		if depField == nil {
			if scopeBlock == nil {
				scopeBlock = &dsl.Block{Name: BlockScope, Items: []dsl.Node{}}
				ab.Items = append(ab.Items, scopeBlock)
			}
			depField = &dsl.Field{Key: "depends_on", Value: &dsl.ListVal{Items: []dsl.Value{}}}
			scopeBlock.Items = append(scopeBlock.Items, depField)
		}

		lv, ok := depField.Value.(*dsl.ListVal)
		if !ok {
			return fmt.Errorf("depends_on is not a list")
		}

		for _, v := range lv.Items {
			if sv, ok := v.(*dsl.StringVal); ok && sv.Text == dependsOnID {
				return nil
			}
		}

		lv.Items = append(lv.Items, &dsl.StringVal{Text: dependsOnID})
		depField.Value = lv
		touchUpdatedAt(ab)
		return nil
	}); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}

	mosDir := filepath.Join(root, MosDir)
	if ValidateContract != nil {
		if err := ValidateContract(contractPath, mosDir); err != nil {
			return fmt.Errorf("LinkContract: %w", err)
		}
	}
	return nil
}

// UnlinkContract removes dependsOnID from the depends_on list of contract id.
// If the dependency is not present, this is a no-op.
func UnlinkContract(root, id, dependsOnID string) error {
	if err := CheckSealForMutation(root, id); err != nil {
		return fmt.Errorf("UnlinkContract: %w", err)
	}

	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("UnlinkContract: %w", err)
	}

	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		var depField *dsl.Field
		scopeBlock := dsl.FindBlock(ab.Items, BlockScope)
		if scopeBlock != nil {
			depField = dsl.FindField(scopeBlock.Items, "depends_on")
		}
		if depField == nil {
			depField = dsl.FindField(ab.Items, "depends_on")
		}
		if depField == nil {
			return nil
		}

		lv, ok := depField.Value.(*dsl.ListVal)
		if !ok {
			return nil
		}

		filtered := make([]dsl.Value, 0, len(lv.Items))
		for _, v := range lv.Items {
			if sv, ok := v.(*dsl.StringVal); ok && sv.Text == dependsOnID {
				continue
			}
			filtered = append(filtered, v)
		}

		if len(filtered) == len(lv.Items) {
			return nil
		}

		lv.Items = filtered
		depField.Value = lv
		touchUpdatedAt(ab)
		return nil
	}); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}
	return nil
}

// FindDependents returns the IDs of all contracts that depend on the given id.
func FindDependents(root, id string) ([]string, error) {
	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("FindDependents: %w", err)
	}
	var deps []string
	for _, c := range all {
		for _, d := range c.DependsOn {
			if d == id {
				deps = append(deps, c.ID)
				break
			}
		}
	}
	return deps, nil
}

// FindChildren returns direct children of a contract (contracts whose parent == id).
func FindChildren(root, id string) ([]ContractInfo, error) {
	return ListContracts(root, ListOpts{Parent: id})
}

// TopologicalSort returns contracts in dependency order (upstream first).
// If cycles exist, the remaining unsorted nodes are appended at the end
// and cycles are returned.
func TopologicalSort(contracts []ContractInfo) ([]ContractInfo, [][]string) {
	idxByID := make(map[string]int, len(contracts))
	for i, c := range contracts {
		idxByID[c.ID] = i
	}

	inDegree := make(map[string]int, len(contracts))
	downstream := make(map[string][]string, len(contracts))
	for _, c := range contracts {
		if _, ok := inDegree[c.ID]; !ok {
			inDegree[c.ID] = 0
		}
		for _, dep := range c.DependsOn {
			if _, ok := idxByID[dep]; ok {
				downstream[dep] = append(downstream[dep], c.ID)
				inDegree[c.ID]++
			}
		}
	}

	var queue []string
	for _, c := range contracts {
		if inDegree[c.ID] == 0 {
			queue = append(queue, c.ID)
		}
	}

	var sorted []ContractInfo
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		sorted = append(sorted, contracts[idxByID[id]])
		for _, child := range downstream[id] {
			inDegree[child]--
			if inDegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	cycles := DetectCycles(contracts)

	if len(sorted) < len(contracts) {
		seen := make(map[string]bool, len(sorted))
		for _, c := range sorted {
			seen[c.ID] = true
		}
		for _, c := range contracts {
			if !seen[c.ID] {
				sorted = append(sorted, c)
			}
		}
	}

	return sorted, cycles
}

// DetectCycles finds all cycles in the contract dependency graph.
// Returns a slice of cycles, each cycle being a slice of contract IDs.
func DetectCycles(contracts []ContractInfo) [][]string {
	byID := make(map[string]*ContractInfo, len(contracts))
	for i := range contracts {
		byID[contracts[i].ID] = &contracts[i]
	}

	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := make(map[string]int, len(contracts))
	parent := make(map[string]string, len(contracts))
	var cycles [][]string

	var dfs func(id string)
	dfs = func(id string) {
		color[id] = gray
		c := byID[id]
		if c == nil {
			color[id] = black
			return
		}
		for _, dep := range c.DependsOn {
			if _, ok := byID[dep]; !ok {
				continue
			}
			if color[dep] == gray {
				cycle := []string{dep, id}
				cur := id
				for cur != dep {
					cur = parent[cur]
					if cur == "" || cur == dep {
						break
					}
					cycle = append(cycle, cur)
				}
				cycles = append(cycles, cycle)
			} else if color[dep] == white {
				parent[dep] = id
				dfs(dep)
			}
		}
		color[id] = black
	}

	for _, c := range contracts {
		if color[c.ID] == white {
			dfs(c.ID)
		}
	}

	return cycles
}

// ContractGraphNode returns the neighborhood of a single contract: the contract
// itself, its direct dependencies (upstream), and its direct dependents (downstream).
func ContractGraphNode(root, id string) ([]ContractInfo, error) {
	if _, err := FindContractPath(root, id); err != nil {
		return nil, fmt.Errorf("ContractGraphNode: %w", err)
	}

	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("ContractGraphNode: %w", err)
	}

	byID := make(map[string]ContractInfo, len(all))
	for _, c := range all {
		byID[c.ID] = c
	}

	target, ok := byID[id]
	if !ok {
		return nil, fmt.Errorf("contract %q not found", id)
	}

	seen := map[string]bool{id: true}
	var result []ContractInfo

	for _, dep := range target.DependsOn {
		if c, ok := byID[dep]; ok && !seen[dep] {
			result = append(result, c)
			seen[dep] = true
		}
	}

	result = append(result, target)

	for _, c := range all {
		if seen[c.ID] {
			continue
		}
		for _, dep := range c.DependsOn {
			if dep == id {
				result = append(result, c)
				seen[c.ID] = true
				break
			}
		}
	}

	return result, nil
}

// ContractGraphSorted wraps ContractGraph to return topologically sorted results.
func ContractGraphSorted(root string) ([]ContractInfo, [][]string, error) {
	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		return nil, nil, fmt.Errorf("ContractGraphSorted: %w", err)
	}
	sorted, cycles := TopologicalSort(all)
	return sorted, cycles, nil
}

// RenameContract changes a contract's ID, moves its directory, updates the
// artifact name, and cascades the rename through all depends_on references.
func RenameContract(root, oldID, newID string) error {
	if _, err := FindContractPath(root, oldID); err != nil {
		return fmt.Errorf("contract %q not found", oldID)
	}
	if _, err := FindContractPath(root, newID); err == nil {
		return fmt.Errorf("contract %q already exists", newID)
	}

	if err := CheckSealForMutation(root, oldID); err != nil {
		return fmt.Errorf("RenameContract: %w", err)
	}

	oldPath, _ := FindContractPath(root, oldID)
	if err := dsl.WithArtifact(oldPath, func(ab *dsl.ArtifactBlock) error {
		ab.Name = newID
		touchUpdatedAt(ab)
		return nil
	}); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}

	oldDir := filepath.Dir(oldPath)
	newDir := filepath.Join(filepath.Dir(oldDir), newID)

	if err := os.MkdirAll(filepath.Dir(newDir), DirPerm); err != nil {
		return fmt.Errorf("creating target directory: %w", err)
	}
	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("moving contract directory: %w", err)
	}

	all, err := ListContracts(root, ListOpts{})
	if err != nil {
		return fmt.Errorf("RenameContract: %w", err)
	}
	for _, c := range all {
		for _, dep := range c.DependsOn {
			if dep == oldID {
				cPath, err := FindContractPath(root, c.ID)
				if err != nil {
					continue
				}
				if err := dsl.WithArtifact(cPath, func(ab *dsl.ArtifactBlock) error {
					updateDependsOnRef(ab, oldID, newID)
					return nil
				}); err != nil {
					continue
				}
				break
			}
		}
	}

	return nil
}

func updateDependsOnRef(ab *dsl.ArtifactBlock, oldID, newID string) {
	scopeBlock := dsl.FindBlock(ab.Items, BlockScope)
	if scopeBlock == nil {
		return
	}
	depField := dsl.FindField(scopeBlock.Items, "depends_on")
	if depField == nil {
		return
	}
	lv, ok := depField.Value.(*dsl.ListVal)
	if !ok {
		return
	}
	for _, v := range lv.Items {
		if sv, ok := v.(*dsl.StringVal); ok && sv.Text == oldID {
			sv.Text = newID
		}
	}
}
