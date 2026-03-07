package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
)

// ReadMosFile reads a file from the .mos/ directory. relPath is relative to .mos/.
func ReadMosFile(root, relPath string) ([]byte, error) {
	return os.ReadFile(filepath.Join(root, names.MosDir, relPath))
}

// ReadMosDirEntries reads directory entries from .mos/. relDir is relative to .mos/.
func ReadMosDirEntries(root, relDir string) ([]os.DirEntry, error) {
	return os.ReadDir(filepath.Join(root, names.MosDir, relDir))
}

// ReadConfig reads and parses .mos/config.mos.
func ReadConfig(root string) (*dsl.File, error) {
	path := filepath.Join(root, names.MosDir, "config.mos")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), nil)
}

// ReadLexiconFile reads and parses a lexicon file (e.g. "default.mos" or "project.mos").
func ReadLexiconFile(root, name string) (*dsl.File, error) {
	path := filepath.Join(root, names.MosDir, "lexicon", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), nil)
}

// ReadLayers reads and parses .mos/resolution/layers.mos.
func ReadLayers(root string) (*dsl.File, error) {
	path := filepath.Join(root, names.MosDir, "resolution", "layers.mos")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), nil)
}

// ReadTemplate reads and parses .mos/templates/contract.mos.
func ReadTemplate(root string) (*dsl.File, error) {
	path := filepath.Join(root, names.MosDir, "templates", "contract.mos")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), nil)
}

// ReadRuleInventory returns a map of rule ID to file path for all rules
// under .mos/rules/{mechanical,interpretive}/.
func ReadRuleInventory(root string, kw *dsl.KeywordMap) map[string]string {
	result := make(map[string]string)
	for _, sub := range []string{"mechanical", "interpretive"} {
		dir := filepath.Join(root, names.MosDir, "rules", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".mos") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			f, err := ReadDSLFile(path, kw)
			if err != nil {
				continue
			}
			if ab, ok := f.Artifact.(*dsl.ArtifactBlock); ok && ab.Name != "" {
				result[ab.Name] = path
			}
		}
	}
	return result
}

// ReadContractInventory returns a map of contract ID to file path for all
// contracts under .mos/contracts/{active,archive}/.
func ReadContractInventory(root string, kw *dsl.KeywordMap) map[string]string {
	result := make(map[string]string)
	for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
		dir := filepath.Join(root, names.MosDir, "contracts", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "contract.mos")
			f, err := ReadDSLFile(path, kw)
			if err != nil {
				continue
			}
			if ab, ok := f.Artifact.(*dsl.ArtifactBlock); ok && ab.Name != "" {
				result[ab.Name] = path
			}
		}
	}
	return result
}

// ReadArtifactInventory returns a kind -> id -> path map for all instances
// of custom artifact types.
func ReadArtifactInventory(root string, directory, kind string) map[string]string {
	ids := make(map[string]string)
	for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
		dir := filepath.Join(root, names.MosDir, directory, sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), kind+".mos")
			if _, err := os.Stat(path); err == nil {
				ids[e.Name()] = path
			}
		}
	}
	return ids
}

// ReadArchitecture reads a specific architecture artifact by ID.
func ReadArchitecture(root, id string) (*dsl.ArtifactBlock, error) {
	path := filepath.Join(root, names.MosDir, names.DirArchitectures,
		names.ActiveDir, id, "architecture.mos")
	return dsl.ReadArtifact(path)
}

// ReadSpecificationInventory returns a map of spec include references found
// in all specification artifacts under .mos/specifications/{active,archive}/.
func ReadSpecificationIncludes(root string) map[string]bool {
	result := make(map[string]bool)
	for _, sub := range []string{names.ActiveDir, names.ArchiveDir} {
		dir := filepath.Join(root, names.MosDir, "specifications", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "specification.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			extractSpecIncludes(ab.Items, result)
		}
	}
	return result
}

func extractSpecIncludes(items []dsl.Node, result map[string]bool) {
	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok {
			continue
		}
		if blk.Name == "include" || blk.Name == "includes" {
			if blk.Title != "" {
				result[blk.Title] = true
			}
			for _, sub := range blk.Items {
				if f, ok := sub.(*dsl.Field); ok {
					result[f.Key] = true
				}
			}
		}
		extractSpecIncludes(blk.Items, result)
	}
}

// ReadDSLFile reads and parses a .mos file at an absolute path with optional keywords.
func ReadDSLFile(path string, kw *dsl.KeywordMap) (*dsl.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), kw)
}

// ReadConfigBlock reads .mos/config.mos and returns its artifact block.
func ReadConfigBlock(root string) (*dsl.ArtifactBlock, error) {
	f, err := ReadConfig(root)
	if err != nil {
		return nil, err
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil, fmt.Errorf("config.mos has no artifact block")
	}
	return ab, nil
}
