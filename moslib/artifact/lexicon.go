package artifact

import (
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sync"

	"github.com/dpopsuev/mos/moslib/dsl"
)

var (
	fileMuMap sync.Map // path -> *sync.Mutex
)

func fileMutex(path string) *sync.Mutex {
	actual, _ := fileMuMap.LoadOrStore(path, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

// LexiconTerm represents a single lexicon term with its description.
type LexiconTerm struct {
	Key         string
	Description string
}

// ListTerms returns all terms from the merged lexicon (default + project).
func ListTerms(root string) ([]LexiconTerm, error) {
	mosDir := filepath.Join(root, MosDir)
	if LoadLexicon == nil {
		return nil, fmt.Errorf("lexicon loader not configured")
	}
	terms, err := LoadLexicon(mosDir)
	if err != nil {
		return nil, fmt.Errorf("loading lexicon: %w", err)
	}
	keys := slices.Sorted(maps.Keys(terms))

	result := make([]LexiconTerm, len(keys))
	for i, k := range keys {
		result[i] = LexiconTerm{Key: k, Description: terms[k]}
	}
	return result, nil
}

// AddTerm adds or updates a term in .mos/lexicon/default.mos.
func AddTerm(root, key, description string) error {
	if key == "" {
		return fmt.Errorf("term key is required")
	}
	if description == "" {
		return fmt.Errorf("--description is required")
	}

	lexiconPath := filepath.Join(root, MosDir, "lexicon", "default.mos")

	mu := fileMutex(lexiconPath)
	mu.Lock()
	defer mu.Unlock()

	return dsl.WithArtifact(lexiconPath, func(ab *dsl.ArtifactBlock) error {
		termsBlock := findOrCreateTermsBlock(ab)
		dsl.SetField(&termsBlock.Items, key, &dsl.StringVal{Text: description})
		return nil
	})
}

// RemoveTerm removes a term from .mos/lexicon/default.mos.
func RemoveTerm(root, key string) error {
	if key == "" {
		return fmt.Errorf("term key is required")
	}

	lexiconPath := filepath.Join(root, MosDir, "lexicon", "default.mos")

	mu := fileMutex(lexiconPath)
	mu.Lock()
	defer mu.Unlock()

	return dsl.WithArtifact(lexiconPath, func(ab *dsl.ArtifactBlock) error {
		termsBlock := findTermsBlock(ab)
		if termsBlock == nil {
			return fmt.Errorf("no terms block found in lexicon")
		}
		if !dsl.HasField(termsBlock.Items, key) {
			return fmt.Errorf("term %q not found in lexicon", key)
		}
		filtered := make([]dsl.Node, 0, len(termsBlock.Items))
		for _, item := range termsBlock.Items {
			if field, ok := item.(*dsl.Field); ok && field.Key == key {
				continue
			}
			filtered = append(filtered, item)
		}
		termsBlock.Items = filtered
		return nil
	})
}

func findTermsBlock(ab *dsl.ArtifactBlock) *dsl.Block {
	return dsl.FindBlock(ab.Items, "terms")
}

func findOrCreateTermsBlock(ab *dsl.ArtifactBlock) *dsl.Block {
	if blk := findTermsBlock(ab); blk != nil {
		return blk
	}
	blk := &dsl.Block{Name: "terms", Items: []dsl.Node{}}
	ab.Items = append(ab.Items, blk)
	return blk
}
