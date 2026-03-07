package artifact

import (
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func sortedKeys(m map[string]string) []string {
	return slices.Sorted(maps.Keys(m))
}

// RuleOpts configures the mos rule create command.
type RuleOpts struct {
	Name            string            // human-readable name
	Type            string            // mechanical | interpretive
	Enforcement     string            // error | warning | info
	Scope           string            // resolution layer (default: "project")
	Glob            string            // file glob pattern
	AppliesTo       []string          // contract kinds this rule applies to
	HarnessCmd      string            // harness command
	HarnessTimeout  string            // harness timeout (e.g. "2m")
	HarnessRequires map[string]string // environment requirements (key=value pairs)
}

// CreateRule creates a rule artifact at .mos/rules/<type>/<id>.mos.
// It validates the produced file against the linter and returns an error
// if any SeverityError diagnostics are found.
func CreateRule(root, id string, opts RuleOpts) (string, error) {
	if id == "" {
		return "", fmt.Errorf("rule id is required")
	}
	if opts.Type == "" {
		return "", fmt.Errorf("--type is required (mechanical or interpretive)")
	}
	if opts.Type != "mechanical" && opts.Type != "interpretive" {
		return "", fmt.Errorf("--type must be \"mechanical\" or \"interpretive\", got %q", opts.Type)
	}
	if opts.Enforcement == "" {
		opts.Enforcement = "error"
	}
	if opts.Scope == "" {
		opts.Scope = "project"
	}
	if opts.Name == "" {
		opts.Name = id
	}

	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return "", fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	rulePath := filepath.Join(mosDir, "rules", opts.Type, id+".mos")

	items := []dsl.Node{
		&dsl.Field{Key: "name", Value: &dsl.StringVal{Text: opts.Name}},
		&dsl.Field{Key: "type", Value: &dsl.StringVal{Text: opts.Type}},
		&dsl.Field{Key: "scope", Value: &dsl.StringVal{Text: opts.Scope}},
		&dsl.Field{Key: "enforcement", Value: &dsl.StringVal{Text: opts.Enforcement}},
	}

	if opts.Glob != "" {
		items = append(items, &dsl.Field{Key: "glob", Value: &dsl.StringVal{Text: opts.Glob}})
	}

	if len(opts.AppliesTo) > 0 {
		listItems := make([]dsl.Value, len(opts.AppliesTo))
		for i, v := range opts.AppliesTo {
			listItems[i] = &dsl.StringVal{Text: v}
		}
		items = append(items, &dsl.Field{Key: "applies_to", Value: &dsl.ListVal{Items: listItems}})
	}

	if opts.HarnessCmd != "" {
		harnessItems := []dsl.Node{
			&dsl.Field{Key: "command", Value: &dsl.StringVal{Text: opts.HarnessCmd}},
		}
		if opts.HarnessTimeout != "" {
			harnessItems = append(harnessItems, &dsl.Field{Key: "timeout", Value: &dsl.StringVal{Text: opts.HarnessTimeout}})
		}
		if len(opts.HarnessRequires) > 0 {
			reqItems := make([]dsl.Node, 0, len(opts.HarnessRequires))
			keys := sortedKeys(opts.HarnessRequires)
			for _, k := range keys {
				reqItems = append(reqItems, &dsl.Field{
					Key:   k,
					Value: &dsl.StringVal{Text: opts.HarnessRequires[k]},
				})
			}
			harnessItems = append(harnessItems, &dsl.Block{Name: "requires", Items: reqItems})
		}
		items = append(items, &dsl.Block{Name: "harness", Items: harnessItems})
	}

	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  "rule",
			Name:  id,
			Items: items,
		},
	}

	if err := writeArtifact(rulePath, file); err != nil {
		return "", fmt.Errorf("writing rule: %w", err)
	}

	if ValidateRule != nil {
		if err := ValidateRule(rulePath, mosDir); err != nil {
			os.Remove(rulePath)
			return "", err
		}
	}

	return rulePath, nil
}

// RuleInfo holds metadata about a discovered rule.
type RuleInfo struct {
	ID          string
	Name        string
	Type        string
	Enforcement string
	AppliesTo   []string
	Path        string
}

// ListRules returns all rules found under .mos/rules/.
// If typeFilter is non-empty, only rules matching that type are returned.
func ListRules(root, typeFilter string) ([]RuleInfo, error) {
	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return nil, fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	var rules []RuleInfo
	for _, ruleType := range []string{"mechanical", "interpretive"} {
		if typeFilter != "" && ruleType != typeFilter {
			continue
		}
		base := filepath.Join(mosDir, "rules", ruleType)
		entries, err := os.ReadDir(base)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".mos" {
				continue
			}
			rulePath := filepath.Join(base, entry.Name())
			info, err := readRuleInfo(rulePath)
			if err != nil {
				continue
			}
			rules = append(rules, info)
		}
	}
	return rules, nil
}

// ShowRule reads and returns the formatted content of a rule.
func ShowRule(root, id string) (string, error) {
	rulePath, err := findRulePath(root, id)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(rulePath)
	if err != nil {
		return "", fmt.Errorf("reading rule: %w", err)
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return "", fmt.Errorf("parsing rule: %w", err)
	}
	return dsl.Format(f, nil), nil
}

// RuleUpdateOpts configures the mos rule update command.
// Nil pointer fields are left unchanged; non-nil fields are applied.
type RuleUpdateOpts struct {
	Name            *string
	Type            *string
	Enforcement     *string
	Scope           *string
	Glob            *string
	AppliesTo       *[]string // nil = don't change, empty slice = clear
	HarnessCmd      *string
	HarnessTimeout  *string
	HarnessRequires map[string]string // nil = don't change, empty map = clear
}

// UpdateRule applies partial updates to an existing rule.
func UpdateRule(root, id string, opts RuleUpdateOpts) error {
	rulePath, err := findRulePath(root, id)
	if err != nil {
		return err
	}

	if opts.Type != nil {
		if *opts.Type != "mechanical" && *opts.Type != "interpretive" {
			return fmt.Errorf("type must be mechanical or interpretive; got %q", *opts.Type)
		}
	}

	if err := dsl.WithArtifact(rulePath, func(ab *dsl.ArtifactBlock) error {
		if opts.Name != nil {
			dsl.SetField(&ab.Items, "name", &dsl.StringVal{Text: *opts.Name})
		}
		if opts.Type != nil {
			dsl.SetField(&ab.Items, "type", &dsl.StringVal{Text: *opts.Type})
		}
		if opts.Enforcement != nil {
			dsl.SetField(&ab.Items, "enforcement", &dsl.StringVal{Text: *opts.Enforcement})
		}
		if opts.Scope != nil {
			dsl.SetField(&ab.Items, "scope", &dsl.StringVal{Text: *opts.Scope})
		}
		if opts.Glob != nil {
			dsl.SetField(&ab.Items, "glob", &dsl.StringVal{Text: *opts.Glob})
		}
		if opts.AppliesTo != nil {
			if len(*opts.AppliesTo) == 0 {
				dsl.SetField(&ab.Items, "applies_to", &dsl.ListVal{Items: nil})
			} else {
				listItems := make([]dsl.Value, len(*opts.AppliesTo))
				for i, v := range *opts.AppliesTo {
					listItems[i] = &dsl.StringVal{Text: v}
				}
				dsl.SetField(&ab.Items, "applies_to", &dsl.ListVal{Items: listItems})
			}
		}
		updateHarnessBlock(ab, opts)
		return nil
	}); err != nil {
		return fmt.Errorf("updating rule: %w", err)
	}

	mosDir := filepath.Join(root, MosDir)
	if opts.Type != nil {
		oldType := filepath.Base(filepath.Dir(rulePath))
		if oldType != *opts.Type {
			newPath := filepath.Join(mosDir, "rules", *opts.Type, id+".mos")
			if err := os.MkdirAll(filepath.Dir(newPath), DirPerm); err != nil {
				return fmt.Errorf("creating target directory: %w", err)
			}
			if err := os.Rename(rulePath, newPath); err != nil {
				return fmt.Errorf("moving rule file: %w", err)
			}
			rulePath = newPath
		}
	}

	if ValidateRule != nil {
		if err := ValidateRule(rulePath, mosDir); err != nil {
			return err
		}
	}

	return nil
}

func updateHarnessBlock(ab *dsl.ArtifactBlock, opts RuleUpdateOpts) {
	if opts.HarnessCmd == nil && opts.HarnessTimeout == nil && opts.HarnessRequires == nil {
		return
	}

	harnessBlock := dsl.FindBlock(ab.Items, "harness")
	if harnessBlock == nil {
		harnessBlock = &dsl.Block{Name: "harness", Items: []dsl.Node{}}
		ab.Items = append(ab.Items, harnessBlock)
	}

	if opts.HarnessCmd != nil {
		dsl.SetField(&harnessBlock.Items, "command", &dsl.StringVal{Text: *opts.HarnessCmd})
	}
	if opts.HarnessTimeout != nil {
		dsl.SetField(&harnessBlock.Items, "timeout", &dsl.StringVal{Text: *opts.HarnessTimeout})
	}

	if opts.HarnessRequires != nil {
		harnessBlock.Items = filterNodes(harnessBlock.Items, func(n dsl.Node) bool {
			blk, ok := n.(*dsl.Block)
			return ok && blk.Name == "requires"
		})
		if len(opts.HarnessRequires) > 0 {
			reqItems := make([]dsl.Node, 0, len(opts.HarnessRequires))
			keys := sortedKeys(opts.HarnessRequires)
			for _, k := range keys {
				reqItems = append(reqItems, &dsl.Field{
					Key:   k,
					Value: &dsl.StringVal{Text: opts.HarnessRequires[k]},
				})
			}
			harnessBlock.Items = append(harnessBlock.Items, &dsl.Block{Name: "requires", Items: reqItems})
		}
	}
}

// DeleteRule removes a rule file from .mos/rules/.
func DeleteRule(root, id string) error {
	rulePath, err := findRulePath(root, id)
	if err != nil {
		return err
	}
	if err := os.Remove(rulePath); err != nil {
		return fmt.Errorf("removing rule file: %w", err)
	}
	return nil
}

func readRuleInfo(path string) (RuleInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return RuleInfo{}, err
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return RuleInfo{}, err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return RuleInfo{}, fmt.Errorf("not an artifact block")
	}
	info := RuleInfo{ID: ab.Name, Path: path}
	info.Name, _ = dsl.FieldString(ab.Items, "name")
	info.Type, _ = dsl.FieldString(ab.Items, "type")
	info.Enforcement, _ = dsl.FieldString(ab.Items, "enforcement")
	info.AppliesTo = dsl.FieldStringSlice(ab.Items, "applies_to")
	return info, nil
}

func findRulePath(root, id string) (string, error) {
	mosDir := filepath.Join(root, MosDir)
	for _, ruleType := range []string{"mechanical", "interpretive"} {
		p := filepath.Join(mosDir, "rules", ruleType, id+".mos")
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("rule %q not found in .mos/rules/", id)
}

// UpdateRuleFields updates arbitrary fields on a rule artifact (for overflow/CAD fields).
func UpdateRuleFields(root, id string, fields map[string]string) error {
	path, err := findRulePath(root, id)
	if err != nil {
		return err
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		for k, v := range fields {
			dsl.SetField(&ab.Items, k, &dsl.StringVal{Text: v})
		}
		return nil
	})
}

// SetRuleHarness adds or replaces the harness block on a rule artifact.
func SetRuleHarness(root, id, command, timeout string) error {
	path, err := findRulePath(root, id)
	if err != nil {
		return err
	}
	harnessItems := []dsl.Node{&dsl.Field{Key: "command", Value: &dsl.StringVal{Text: command}}}
	if timeout != "" {
		harnessItems = append(harnessItems, &dsl.Field{Key: "timeout", Value: &dsl.StringVal{Text: timeout}})
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		dsl.RemoveBlock(&ab.Items, "harness")
		dsl.AppendToBlock(&ab.Items, "harness", harnessItems...)
		return nil
	})
}

// RemoveRuleBlock removes a block from a rule artifact by type and optional name.
func RemoveRuleBlock(root, id, blockType, blockName string) error {
	path, err := findRulePath(root, id)
	if err != nil {
		return err
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		if blockType != "feature" {
			removed := false
			if blockName == "" {
				removed = dsl.RemoveBlock(&ab.Items, blockType)
			} else {
				removed = dsl.RemoveNamedBlock(&ab.Items, blockType, blockName)
			}
			if !removed {
				if blockName != "" {
					return fmt.Errorf("%s block %q not found on rule %q", blockType, blockName, id)
				}
				return fmt.Errorf("%s block not found on rule %q", blockType, id)
			}
			return nil
		}
		// FeatureBlock: use filterNodes
		found := false
		var filtered []dsl.Node
		for _, item := range ab.Items {
			match := false
			if fb, ok := item.(*dsl.FeatureBlock); ok {
				match = blockName == "" || fb.Name == blockName
			}
			if match {
				found = true
			} else {
				filtered = append(filtered, item)
			}
		}
		if !found {
			if blockName != "" {
				return fmt.Errorf("%s block %q not found on rule %q", blockType, blockName, id)
			}
			return fmt.Errorf("%s block not found on rule %q", blockType, id)
		}
		ab.Items = filtered
		return nil
	})
}

// AddRuleSection adds or replaces a named section block on a rule artifact.
func AddRuleSection(root, id, name, text string) error {
	path, err := findRulePath(root, id)
	if err != nil {
		return err
	}
	sectionBlk := &dsl.Block{
		Name:  "section",
		Title: name,
		Items: []dsl.Node{
			&dsl.Field{Key: "text", Value: &dsl.StringVal{Text: text, Triple: true}},
		},
	}
	return dsl.WithArtifact(path, func(ab *dsl.ArtifactBlock) error {
		dsl.RemoveNamedBlock(&ab.Items, "section", name)
		ab.Items = append(ab.Items, sectionBlk)
		return nil
	})
}
