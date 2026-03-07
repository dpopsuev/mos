package dsl

import (
	"fmt"
	"os"
	"slices"
	"strings"
)

// FindBlock returns the first child block with the given name, or nil.
func FindBlock(items []Node, name string) *Block {
	for _, item := range items {
		if b, ok := item.(*Block); ok && b.Name == name {
			return b
		}
	}
	return nil
}

// FieldString returns the string value of the first field with the given key.
func FieldString(items []Node, key string) (string, bool) {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			switch v := f.Value.(type) {
			case *StringVal:
				return v.Text, true
			case *ListVal:
				if len(v.Items) > 0 {
					if sv, ok := v.Items[0].(*StringVal); ok {
						return sv.Text, true
					}
				}
				return "", true
			}
		}
	}
	return "", false
}

// FieldInt returns the integer value of the first field with the given key.
func FieldInt(items []Node, key string) (int64, bool) {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			if iv, ok := f.Value.(*IntegerVal); ok {
				return iv.Val, true
			}
		}
	}
	return 0, false
}

// FieldFloat returns the float64 value of the first field with the given key.
// Accepts both FloatVal and IntegerVal (promoted to float64).
func FieldFloat(items []Node, key string) (float64, bool) {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			switch v := f.Value.(type) {
			case *FloatVal:
				return v.Val, true
			case *IntegerVal:
				return float64(v.Val), true
			}
		}
	}
	return 0, false
}

// FieldStringSlice returns string list values for the first field with the given key.
func FieldStringSlice(items []Node, key string) []string {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			if lv, ok := f.Value.(*ListVal); ok {
				var out []string
				for _, elem := range lv.Items {
					if sv, ok := elem.(*StringVal); ok {
						out = append(out, sv.Text)
					}
				}
				return out
			}
			if sv, ok := f.Value.(*StringVal); ok && sv.Text != "" {
				if strings.Contains(sv.Text, ",") {
					parts := strings.Split(sv.Text, ",")
					for i := range parts {
						parts[i] = strings.TrimSpace(parts[i])
					}
					return parts
				}
				return []string{sv.Text}
			}
		}
	}
	return nil
}

// FieldBool returns the boolean value of the first field with the given key, defaulting to false.
func FieldBool(items []Node, key string) bool {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			if bv, ok := f.Value.(*BoolVal); ok {
				return bv.Val
			}
		}
	}
	return false
}

// HasField returns true if a field with the given key exists in items.
func HasField(items []Node, key string) bool {
	for _, item := range items {
		if f, ok := item.(*Field); ok && f.Key == key {
			return true
		}
	}
	return false
}

// ToMap converts an ArtifactBlock into a map[string]any suitable for JSON marshaling.
func ToMap(ab *ArtifactBlock) map[string]any {
	m := map[string]any{
		"kind": ab.Kind,
	}
	if ab.Name != "" {
		m["id"] = ab.Name
	}
	nodeListToMap(ab.Items, m)
	return m
}

// SetField creates or updates a field in items. If a field with the given key
// already exists, its value is replaced; otherwise a new field is appended.
func SetField(items *[]Node, key string, val Value) {
	idx := slices.IndexFunc(*items, func(n Node) bool {
		f, ok := n.(*Field)
		return ok && f.Key == key
	})
	if idx >= 0 {
		(*items)[idx].(*Field).Value = val
	} else {
		*items = append(*items, &Field{Key: key, Value: val})
	}
}

// RemoveBlock removes the first block with the given name from items.
// Returns true if a block was removed.
func RemoveBlock(items *[]Node, name string) bool {
	idx := slices.IndexFunc(*items, func(n Node) bool {
		b, ok := n.(*Block)
		return ok && b.Name == name
	})
	if idx < 0 {
		return false
	}
	*items = slices.Delete(*items, idx, idx+1)
	return true
}

// RemoveNamedBlock removes the first block with the given name and title from items.
// Returns true if a block was removed.
func RemoveNamedBlock(items *[]Node, name, title string) bool {
	idx := slices.IndexFunc(*items, func(n Node) bool {
		b, ok := n.(*Block)
		return ok && b.Name == name && b.Title == title
	})
	if idx < 0 {
		return false
	}
	*items = slices.Delete(*items, idx, idx+1)
	return true
}

// WalkBlocks visits all Block nodes in items recursively. The callback
// returns true to descend into the block's children, false to skip them.
// Follows the go/ast.Inspect pattern.
func WalkBlocks(items []Node, fn func(*Block) bool) {
	for _, item := range items {
		if b, ok := item.(*Block); ok {
			if fn(b) {
				WalkBlocks(b.Items, fn)
			}
		}
	}
}

// AppendToBlock appends nodes to the first block with the given name.
// If no such block exists, a new block is created and appended to items.
func AppendToBlock(items *[]Node, blockName string, nodes ...Node) {
	b := FindBlock(*items, blockName)
	if b == nil {
		b = &Block{Name: blockName}
		*items = append(*items, b)
	}
	b.Items = append(b.Items, nodes...)
}

// WithArtifact encapsulates the load-parse-modify-format-write cycle.
// It reads the file at path, parses it, passes the ArtifactBlock to fn,
// then formats and writes the result back. If fn returns an error, the
// file is not written.
func WithArtifact(path string, fn func(*ArtifactBlock) error) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	f, err := Parse(string(data), nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	ab, ok := f.Artifact.(*ArtifactBlock)
	if !ok {
		return fmt.Errorf("%s: no artifact block", path)
	}
	if err := fn(ab); err != nil {
		return err
	}
	content := Format(f, nil)
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadArtifact reads a file at path and returns its ArtifactBlock.
// Returns an error if the file cannot be read, parsed, or is not an artifact.
func ReadArtifact(path string) (*ArtifactBlock, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	f, err := Parse(string(data), nil)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	ab, ok := f.Artifact.(*ArtifactBlock)
	if !ok {
		return nil, fmt.Errorf("%s: no artifact block", path)
	}
	return ab, nil
}

// FindField returns the first field with the given key, or nil.
func FindField(items []Node, key string) *Field {
	idx := slices.IndexFunc(items, func(n Node) bool {
		f, ok := n.(*Field)
		return ok && f.Key == key
	})
	if idx < 0 {
		return nil
	}
	return items[idx].(*Field)
}

func nodeListToMap(items []Node, m map[string]any) {
	for _, item := range items {
		switch v := item.(type) {
		case *Field:
			m[v.Key] = valueToAny(v.Value)
		case *Block:
			sub := make(map[string]any)
			if v.Title != "" {
				sub["_name"] = v.Title
			}
			nodeListToMap(v.Items, sub)
			appendBlock(m, v.Name, sub)
		case *FeatureBlock:
			fb := featureToMap(v)
			appendBlock(m, "feature", fb)
		case *SpecBlock:
			sb := specBlockToMap(v)
			m["spec"] = sb
		}
	}
}

func valueToAny(v Value) any {
	switch val := v.(type) {
	case *StringVal:
		return val.Text
	case *IntegerVal:
		return val.Val
	case *FloatVal:
		return val.Val
	case *BoolVal:
		return val.Val
	case *DateTimeVal:
		return val.Raw
	case *ListVal:
		var out []any
		for _, item := range val.Items {
			out = append(out, valueToAny(item))
		}
		return out
	case *InlineTableVal:
		sub := make(map[string]any)
		for _, f := range val.Fields {
			sub[f.Key] = valueToAny(f.Value)
		}
		return sub
	default:
		return nil
	}
}

func appendBlock(m map[string]any, key string, sub map[string]any) {
	if existing, ok := m[key]; ok {
		switch ex := existing.(type) {
		case []any:
			m[key] = append(ex, sub)
		default:
			m[key] = []any{ex, sub}
		}
	} else {
		m[key] = sub
	}
}

func featureToMap(fb *FeatureBlock) map[string]any {
	m := map[string]any{"_name": fb.Name}
	if len(fb.Description) > 0 {
		m["description"] = fb.Description
	}
	if fb.Background != nil && fb.Background.Given != nil {
		m["background"] = map[string]any{"given": fb.Background.Given.Lines}
	}
	var scenarios []any
	for _, g := range fb.Groups {
		switch gc := g.(type) {
		case *Scenario:
			scenarios = append(scenarios, scenarioToMap(gc))
		case *Group:
			gm := map[string]any{"_name": gc.Name}
			var ss []any
			for _, s := range gc.Scenarios {
				ss = append(ss, scenarioToMap(s))
			}
			gm["scenarios"] = ss
			scenarios = append(scenarios, gm)
		}
	}
	if len(scenarios) > 0 {
		m["scenarios"] = scenarios
	}
	return m
}

func scenarioToMap(s *Scenario) map[string]any {
	m := map[string]any{"_name": s.Name}
	for _, f := range s.Fields {
		m[f.Key] = valueToAny(f.Value)
	}
	if s.Given != nil {
		m["given"] = s.Given.Lines
	}
	if s.When != nil {
		m["when"] = s.When.Lines
	}
	if s.Then != nil {
		m["then"] = s.Then.Lines
	}
	return m
}

func specBlockToMap(sb *SpecBlock) map[string]any {
	m := make(map[string]any)
	if len(sb.Includes) > 0 {
		var inc []string
		for _, i := range sb.Includes {
			inc = append(inc, i.Path)
		}
		m["includes"] = inc
	}
	if len(sb.Features) > 0 {
		var features []any
		for _, fb := range sb.Features {
			features = append(features, featureToMap(fb))
		}
		m["features"] = features
	}
	return m
}
