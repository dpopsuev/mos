package artifact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

type pathSegment struct {
	Name  string
	Index int // -1 means "by name", >= 0 means "by index"
}

func parsePath(path string) ([]pathSegment, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}
	var segs []pathSegment
	parts := strings.Split(path, ".")
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("empty segment in path %q", path)
		}
		if idx := strings.Index(part, "["); idx >= 0 {
			name := part[:idx]
			rest := part[idx:]
			if !strings.HasSuffix(rest, "]") {
				return nil, fmt.Errorf("unclosed bracket in %q", part)
			}
			numStr := rest[1 : len(rest)-1]
			n, err := strconv.Atoi(numStr)
			if err != nil {
				return nil, fmt.Errorf("non-integer index in %q", part)
			}
			segs = append(segs, pathSegment{Name: name, Index: n})
		} else {
			segs = append(segs, pathSegment{Name: part, Index: -1})
		}
	}
	return segs, nil
}

// PathGet reads a value from a parsed DSL artifact at the given dot-path.
func PathGet(ab *dsl.ArtifactBlock, path string) (string, error) {
	segs, err := parsePath(path)
	if err != nil {
		return "", err
	}
	return resolveGet(ab.Items, segs)
}

func resolveGet(items []dsl.Node, segs []pathSegment) (string, error) {
	if len(segs) == 0 {
		return "", fmt.Errorf("path exhausted without reaching a value")
	}
	seg := segs[0]
	rest := segs[1:]

	if seg.Index >= 0 {
		matches := findAllBlocks(items, seg.Name)
		if seg.Index >= len(matches) {
			return "", fmt.Errorf("%s[%d]: index out of range (have %d)", seg.Name, seg.Index, len(matches))
		}
		blk := matches[seg.Index]
		if len(rest) == 0 {
			return blk.Title, nil
		}
		return resolveGet(blk.Items, rest)
	}

	if len(rest) == 0 {
		return getFieldValue(items, seg.Name)
	}

	blk := findPathBlock(items, seg.Name)
	if blk != nil {
		return resolveGet(blk.Items, rest)
	}

	fb := findFeatureBlock(items, seg.Name)
	if fb != nil {
		return resolveFeatureGet(fb, rest)
	}

	return "", fmt.Errorf("block %q not found", seg.Name)
}

func getFieldValue(items []dsl.Node, key string) (string, error) {
	f := dsl.FindField(items, key)
	if f == nil {
		return "", fmt.Errorf("field %q not found", key)
	}
	switch v := f.Value.(type) {
	case *dsl.StringVal:
		return v.Text, nil
	case *dsl.BoolVal:
		if v.Val {
			return "true", nil
		}
		return "false", nil
	case *dsl.IntegerVal:
		return strconv.FormatInt(v.Val, 10), nil
	case *dsl.ListVal:
		parts := dsl.FieldStringSlice(items, key)
		return strings.Join(parts, ","), nil
	default:
		return "", nil
	}
}

func findPathBlock(items []dsl.Node, name string) *dsl.Block {
	return dsl.FindBlock(items, name)
}

func findAllBlocks(items []dsl.Node, name string) []*dsl.Block {
	var result []*dsl.Block
	for _, item := range items {
		if blk, ok := item.(*dsl.Block); ok && blk.Name == name {
			result = append(result, blk)
		}
	}
	return result
}

func findFeatureBlock(items []dsl.Node, name string) *dsl.FeatureBlock {
	for _, item := range items {
		if fb, ok := item.(*dsl.FeatureBlock); ok && fb.Name == name {
			return fb
		}
	}
	return nil
}

func resolveFeatureGet(fb *dsl.FeatureBlock, segs []pathSegment) (string, error) {
	if len(segs) == 0 {
		return fb.Name, nil
	}
	seg := segs[0]
	if seg.Name == "name" && len(segs) == 1 {
		return fb.Name, nil
	}
	return "", fmt.Errorf("cannot traverse into feature block with path %q", seg.Name)
}

// PathSet writes a value into a parsed DSL artifact at the given dot-path.
// It creates intermediate blocks as needed.
func PathSet(ab *dsl.ArtifactBlock, path, value string) error {
	segs, err := parsePath(path)
	if err != nil {
		return err
	}
	ab.Items = resolveSet(ab.Items, segs, value)
	return nil
}

func resolveSet(items []dsl.Node, segs []pathSegment, value string) []dsl.Node {
	if len(segs) == 1 {
		seg := segs[0]
		dsl.SetField(&items, seg.Name, parseValue(value))
		return items
	}

	seg := segs[0]
	rest := segs[1:]

	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != seg.Name {
			continue
		}
		blk.Items = resolveSet(blk.Items, rest, value)
		return items
	}

	newBlk := &dsl.Block{Name: seg.Name}
	newBlk.Items = resolveSet(nil, rest, value)
	items = append(items, newBlk)
	return items
}

func parseValue(s string) dsl.Value {
	if s == "true" {
		return &dsl.BoolVal{Val: true}
	}
	if s == "false" {
		return &dsl.BoolVal{Val: false}
	}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return &dsl.IntegerVal{Val: n}
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return &dsl.StringVal{Text: s}
}

// PathAppend appends a value to a list field at the given dot-path.
func PathAppend(ab *dsl.ArtifactBlock, path, value string) error {
	segs, err := parsePath(path)
	if err != nil {
		return err
	}
	return resolveAppend(ab.Items, segs, value)
}

func resolveAppend(items []dsl.Node, segs []pathSegment, value string) error {
	if len(segs) == 1 {
		seg := segs[0]
		for _, item := range items {
			f, ok := item.(*dsl.Field)
			if !ok || f.Key != seg.Name {
				continue
			}
			lv, ok := f.Value.(*dsl.ListVal)
			if !ok {
				return fmt.Errorf("field %q is not a list", seg.Name)
			}
			lv.Items = append(lv.Items, &dsl.StringVal{Text: value})
			return nil
		}
		return fmt.Errorf("field %q not found", seg.Name)
	}

	seg := segs[0]
	rest := segs[1:]

	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != seg.Name {
			continue
		}
		return resolveAppend(blk.Items, rest, value)
	}
	return fmt.Errorf("block %q not found", seg.Name)
}

func findArtifactPathByID(root, id string) (string, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return "", fmt.Errorf("load registry: %w", err)
	}
	kind, err := reg.ResolveKindFromID(id)
	if err != nil {
		return "", err
	}
	td, ok := reg.Types[kind]
	if !ok {
		return "", fmt.Errorf("no type definition for kind %q", kind)
	}
	return FindGenericArtifactPath(root, td, id)
}

func findArtifactFileByID(root, id string) (string, *dsl.File, error) {
	artPath, err := findArtifactPathByID(root, id)
	if err != nil {
		return "", nil, err
	}
	data, err := storeReadFile(artPath)
	if err != nil {
		return "", nil, err
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return "", nil, fmt.Errorf("parse %s: %w", artPath, err)
	}
	return artPath, f, nil
}

// GetArtifactField reads a field from an artifact on disk via its ID and path.
func GetArtifactField(root, id, path string) (string, error) {
	_, f, err := findArtifactFileByID(root, id)
	if err != nil {
		return "", err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return "", fmt.Errorf("file does not contain a valid artifact")
	}
	return PathGet(ab, path)
}

// SetArtifactField writes a field in an artifact on disk via its ID and path.
func SetArtifactField(root, id, path, value string) error {
	if path == "kind" {
		if err := validateKindChange(root, id, value); err != nil {
			return err
		}
	}
	artPath, err := findArtifactPathByID(root, id)
	if err != nil {
		return err
	}
	return dsl.WithArtifact(artPath, func(ab *dsl.ArtifactBlock) error {
		return PathSet(ab, path, value)
	})
}

// validateKindChange rejects kind changes that would create a prefix/kind
// mismatch. E.g. BUG-2026-023 (project "bugs") cannot have kind = "feature".
func validateKindChange(root, id, newKind string) error {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) < 2 {
		return nil
	}
	prefix := parts[0]

	projects, err := LoadProjects(root)
	if err != nil {
		return nil
	}

	proj := FindProjectByPrefix(projects, prefix)
	if proj == nil || proj.Default {
		return nil
	}

	// Project name "bugs" implies kind "bug" (strip trailing 's')
	impliedKind := strings.TrimSuffix(proj.Name, "s")
	if newKind != impliedKind {
		return fmt.Errorf(
			"cannot set kind=%q on %s: prefix %s- belongs to project %q which implies kind=%q; "+
				"create a new artifact with the correct prefix instead",
			newKind, id, prefix, proj.Name, impliedKind,
		)
	}
	return nil
}

// AppendArtifactField appends a value to a list field on an artifact on disk.
func AppendArtifactField(root, id, path, value string) error {
	artPath, err := findArtifactPathByID(root, id)
	if err != nil {
		return err
	}
	return dsl.WithArtifact(artPath, func(ab *dsl.ArtifactBlock) error {
		return PathAppend(ab, path, value)
	})
}
