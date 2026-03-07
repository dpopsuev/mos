package artifact

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// GenericAddSection adds or replaces a named section block on a custom artifact.
func GenericAddSection(root string, td ArtifactTypeDef, id, name, text string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddSection: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		sectionBlk := &dsl.Block{
			Name:  "section",
			Title: name,
			Items: []dsl.Node{
				&dsl.Field{Key: "text", Value: &dsl.StringVal{Text: text, Triple: true}},
			},
		}
		dsl.RemoveNamedBlock(&ab.Items, "section", name)
		ab.Items = append(ab.Items, sectionBlk)
		return nil
	})
}

// GenericAddFeature adds or replaces a named feature block on an artifact.
func GenericAddFeature(root string, td ArtifactTypeDef, id, name, description string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddFeature: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		fb := &dsl.FeatureBlock{Name: name}
		replaced := false
		for i, item := range ab.Items {
			if existing, ok := item.(*dsl.FeatureBlock); ok && existing.Name == name {
				ab.Items[i] = fb
				replaced = true
				break
			}
		}
		if !replaced {
			ab.Items = append(ab.Items, fb)
		}
		return nil
	})
}

// GenericAddScenario adds a scenario to an existing feature block on an artifact.
func GenericAddScenario(root string, td ArtifactTypeDef, id, featureName, scenarioName, given, when, then string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddScenario: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		var target *dsl.FeatureBlock
		for _, item := range ab.Items {
			if fb, ok := item.(*dsl.FeatureBlock); ok && fb.Name == featureName {
				target = fb
				break
			}
		}
		if target == nil {
			return fmt.Errorf("feature %q not found on %s %q", featureName, td.Kind, id)
		}
		scenario := &dsl.Scenario{
			Name:  scenarioName,
			Given: &dsl.StepBlock{Lines: []string{given}},
			When:  &dsl.StepBlock{Lines: []string{when}},
			Then:  &dsl.StepBlock{Lines: []string{then}},
		}
		for i, g := range target.Groups {
			if s, ok := g.(*dsl.Scenario); ok && s.Name == scenarioName {
				target.Groups[i] = scenario
				return nil
			}
		}
		target.Groups = append(target.Groups, scenario)
		return nil
	})
}

// GenericAddCriterion adds or replaces a criterion inside an acceptance block.
// If the acceptance block doesn't exist, it is created.
func GenericAddCriterion(root string, td ArtifactTypeDef, id, criterionName, description string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddCriterion: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		acceptance := dsl.FindBlock(ab.Items, "acceptance")
		if acceptance == nil {
			acceptance = &dsl.Block{Name: "acceptance"}
			ab.Items = append(ab.Items, acceptance)
		}
		crit := &dsl.Block{
			Name:  "criterion",
			Title: criterionName,
			Items: []dsl.Node{
				&dsl.Field{Key: "description", Value: &dsl.StringVal{Text: description}},
			},
		}
		dsl.RemoveNamedBlock(&acceptance.Items, "criterion", criterionName)
		acceptance.Items = append(acceptance.Items, crit)
		return nil
	})
}

// GenericRemoveBlock removes a block from an artifact by type and optional name.
func GenericRemoveBlock(root string, td ArtifactTypeDef, id, blockType, blockName string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericRemoveBlock: %w", err)
	}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		var filtered []dsl.Node
		found := false
		for _, item := range ab.Items {
			match := false
			switch v := item.(type) {
			case *dsl.Block:
				match = v.Name == blockType && (blockName == "" || v.Title == blockName)
			case *dsl.FeatureBlock:
				match = blockType == "feature" && (blockName == "" || v.Name == blockName)
			case *dsl.SpecBlock:
				match = blockType == "spec"
			}
			if match {
				found = true
			} else {
				filtered = append(filtered, item)
			}
		}
		if !found {
			if blockName != "" {
				return fmt.Errorf("%s block %q not found on %s %q", blockType, blockName, td.Kind, id)
			}
			return fmt.Errorf("%s block not found on %s %q", blockType, td.Kind, id)
		}
		ab.Items = filtered
		return nil
	})
}

// GenericSetHarness adds or replaces the harness block on an artifact.
func GenericSetHarness(root string, td ArtifactTypeDef, id, command, timeout string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericSetHarness: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		harnessItems := []dsl.Node{&dsl.Field{Key: "command", Value: &dsl.StringVal{Text: command}}}
		if timeout != "" {
			harnessItems = append(harnessItems, &dsl.Field{Key: "timeout", Value: &dsl.StringVal{Text: timeout}})
		}
		harnessBlk := &dsl.Block{Name: "harness", Items: harnessItems}
		dsl.RemoveBlock(&ab.Items, "harness")
		ab.Items = append(ab.Items, harnessBlk)
		return nil
	})
}

// GenericAddCoverage adds or replaces a coverage block on an artifact with key=value fields.
func GenericAddCoverage(root string, td ArtifactTypeDef, id string, fields map[string]string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddCoverage: %w", err)
	}

	var items []dsl.Node
	for _, k := range slices.Sorted(maps.Keys(fields)) {
		items = append(items, &dsl.Field{Key: k, Value: &dsl.StringVal{Text: fields[k]}})
	}
	covBlk := &dsl.Block{Name: "coverage", Items: items}

	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		dsl.RemoveBlock(&ab.Items, "coverage")
		ab.Items = append(ab.Items, covBlk)
		return nil
	})
}

// GenericAddBill adds or replaces a bill block on an artifact.
func GenericAddBill(root string, td ArtifactTypeDef, id, introducedBy, intent string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddBill: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		billBlk := &dsl.Block{
			Name: "bill",
			Items: []dsl.Node{
				&dsl.Field{Key: "introduced_by", Value: &dsl.StringVal{Text: introducedBy}},
				&dsl.Field{Key: "intent", Value: &dsl.StringVal{Text: intent}},
			},
		}
		dsl.RemoveBlock(&ab.Items, "bill")
		ab.Items = append(ab.Items, billBlk)
		return nil
	})
}

// GenericAddSpec adds or replaces a spec block on an artifact with include directives.
func GenericAddSpec(root string, td ArtifactTypeDef, id string, includes []string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddSpec: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		var directives []*dsl.IncludeDirective
		for _, inc := range includes {
			directives = append(directives, &dsl.IncludeDirective{Path: inc})
		}
		specBlk := &dsl.SpecBlock{Includes: directives}
		idx := slices.IndexFunc(ab.Items, func(n dsl.Node) bool { _, ok := n.(*dsl.SpecBlock); return ok })
		if idx >= 0 {
			ab.Items = slices.Delete(ab.Items, idx, idx+1)
		}
		ab.Items = append(ab.Items, specBlk)
		return nil
	})
}

// GenericRemoveScenario removes a named scenario from a feature block on an artifact.
func GenericRemoveScenario(root string, td ArtifactTypeDef, id, featureName, scenarioName string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericRemoveScenario: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		var target *dsl.FeatureBlock
		for _, item := range ab.Items {
			if fb, ok := item.(*dsl.FeatureBlock); ok && fb.Name == featureName {
				target = fb
				break
			}
		}
		if target == nil {
			return fmt.Errorf("feature %q not found on %s %q", featureName, td.Kind, id)
		}
		var filtered []dsl.ScenarioContainer
		found := false
		for _, g := range target.Groups {
			if s, ok := g.(*dsl.Scenario); ok && s.Name == scenarioName {
				found = true
				continue
			}
			filtered = append(filtered, g)
		}
		if !found {
			return fmt.Errorf("scenario %q not found in feature %q on %s %q", scenarioName, featureName, td.Kind, id)
		}
		target.Groups = filtered
		return nil
	})
}

// BlameEntry represents a structured source reference on a bug contract.
type BlameEntry struct {
	File   string `json:"file"`
	Lines  string `json:"lines,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}

// GenericAddBlame adds or replaces a blame block on a contract.
func GenericAddBlame(root string, td ArtifactTypeDef, id, file, lines, symbol string) error {
	path, err := FindGenericPath(root, td, id)
	if err != nil {
		return fmt.Errorf("GenericAddBlame: %w", err)
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		items := []dsl.Node{
			&dsl.Field{Key: "file", Value: &dsl.StringVal{Text: file}},
		}
		if lines != "" {
			items = append(items, &dsl.Field{Key: "lines", Value: &dsl.StringVal{Text: lines}})
		}
		if symbol != "" {
			items = append(items, &dsl.Field{Key: "symbol", Value: &dsl.StringVal{Text: symbol}})
		}
		blameBlk := &dsl.Block{Name: "blame", Title: file, Items: items}
		dsl.RemoveNamedBlock(&ab.Items, "blame", file)
		ab.Items = append(ab.Items, blameBlk)
		return nil
	})
}

// GenericRemoveBlame removes a blame block by file path.
func GenericRemoveBlame(root string, td ArtifactTypeDef, id, file string) error {
	return GenericRemoveBlock(root, td, id, "blame", file)
}

// ParseBlameEntries extracts all blame blocks from an artifact.
func ParseBlameEntries(ab *dsl.ArtifactBlock) []BlameEntry {
	var entries []BlameEntry
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "blame" {
			continue
		}
		entry := BlameEntry{}
		if f, ok := dsl.FieldString(blk.Items, "file"); ok {
			entry.File = f
		}
		if l, ok := dsl.FieldString(blk.Items, "lines"); ok {
			entry.Lines = l
		}
		if s, ok := dsl.FieldString(blk.Items, "symbol"); ok {
			entry.Symbol = s
		}
		if entry.File != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

func evaluateTransitionGates(root string, td ArtifactTypeDef, id string, ab *dsl.ArtifactBlock, from, to string) error {
	for _, gate := range td.Lifecycle.Gates {
		if gate.From != from || gate.To != to {
			continue
		}
		switch gate.Gate {
		case "criteria_coverage":
			return evaluateCriteriaCoverageGate(root, td, id, ab)
		case "test_matrix_coverage":
			return evaluateTestMatrixGate(root, id, ab)
		default:
			return fmt.Errorf("unknown gate %q on transition %s -> %s", gate.Gate, from, to)
		}
	}
	return nil
}

func evaluateCriteriaCoverageGate(root string, td ArtifactTypeDef, id string, ab *dsl.ArtifactBlock) error {
	criteria := ParseAcceptanceCriteria(ab)
	if len(criteria) == 0 {
		return nil
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("gate check: %w", err)
	}

	specTD, ok := reg.Types["specification"]
	if !ok {
		return fmt.Errorf("gate check: specification type not found in registry")
	}

	specDir := filepath.Join(root, MosDir, specTD.Directory, ActiveDir)
	entries, _ := os.ReadDir(specDir)

	addressed := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		specPath := filepath.Join(specDir, entry.Name(), "specification.mos")
		sab, err := dsl.ReadArtifact(specPath)
		if err != nil {
			continue
		}
		satisfiesList := dsl.FieldStringSlice(sab.Items, "satisfies")
		if !slices.Contains(satisfiesList, id) {
			continue
		}
		addrs := dsl.FieldStringSlice(sab.Items, "addresses")
		if len(addrs) == 0 {
			if single, ok := dsl.FieldString(sab.Items, "addresses"); ok && single != "" {
				addrs = []string{single}
			}
		}
		for _, a := range addrs {
			addressed[a] = true
		}
	}

	var uncovered []string
	for _, c := range criteria {
		if !addressed[c.Name] {
			uncovered = append(uncovered, c.Name)
		}
	}

	if len(uncovered) > 0 {
		return fmt.Errorf("transition blocked: criteria not covered by any specification: %s",
			strings.Join(uncovered, ", "))
	}
	return nil
}

// evaluateTestMatrixGate checks that the justified specification has a
// test_matrix block covering the contract's applicable coverage layers.
// Contracts without a justifies field pass the gate freely.
func evaluateTestMatrixGate(root, id string, ab *dsl.ArtifactBlock) error {
	justifies := FieldStr(ab.Items, "justifies")
	if justifies == "" {
		return nil
	}

	applicableLayers := contractCoverageLayers(ab)
	if len(applicableLayers) == 0 {
		return nil
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return nil
	}
	specTD, ok := reg.Types["specification"]
	if !ok {
		return nil
	}
	specPath, err := FindGenericPath(root, specTD, justifies)
	if err != nil {
		return nil
	}
	specAB, err := dsl.ReadArtifact(specPath)
	if err != nil {
		return nil
	}

	matrixLayers := specTestMatrixLayers(specAB)
	var missing []string
	for _, layer := range applicableLayers {
		if !matrixLayers[layer] {
			missing = append(missing, layer)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("transition blocked: specification %s is missing test_matrix coverage for layers: %s",
			justifies, strings.Join(missing, ", "))
	}
	return nil
}

// contractCoverageLayers returns the layer names from the contract's coverage
// block(s) where applies = true.
func contractCoverageLayers(ab *dsl.ArtifactBlock) []string {
	var layers []string
	dsl.WalkBlocks(ab.Items, func(b *dsl.Block) bool {
		if b.Name != "coverage" {
			return true
		}
		dsl.WalkBlocks(b.Items, func(sub *dsl.Block) bool {
			if dsl.FieldBool(sub.Items, "applies") {
				layers = append(layers, sub.Name)
			}
			return false
		})
		return false
	})
	return layers
}

// specTestMatrixLayers returns the set of layer names in a spec's test_matrix
// block(s) that have a non-empty symbol binding.
func specTestMatrixLayers(ab *dsl.ArtifactBlock) map[string]bool {
	result := make(map[string]bool)
	dsl.WalkBlocks(ab.Items, func(b *dsl.Block) bool {
		if b.Name != "test_matrix" {
			return true
		}
		dsl.WalkBlocks(b.Items, func(sub *dsl.Block) bool {
			if sym, ok := dsl.FieldString(sub.Items, "symbol"); ok && sym != "" {
				result[sub.Name] = true
			}
			return false
		})
		return false
	})
	return result
}
