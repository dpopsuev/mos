package artifact

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// Artifact defines the per-kind routing and validation logic used by
// ApplyArtifact and EditArtifact.
type Artifact interface {
	Kind() string
	ResolvePath(root string, ab *dsl.ArtifactBlock) (string, error)
	FindExisting(root string, id string) (string, error)
	Validate(path, mosDir string) error
}

var artifactRegistry = map[string]Artifact{
	KindContract: contractArtifact{},
	KindRule:     ruleArtifact{},
	KindLexicon:  lexiconArtifact{},
}

// RegisterArtifact adds an artifact implementation to the global registry.
func RegisterArtifact(kind string, impl Artifact) {
	artifactRegistry[kind] = impl
}

// InitDynamicRegistry loads all artifact types (CADs) from config.mos and
// registers those without dedicated Artifact implementations as generic artifacts.
func InitDynamicRegistry(root string) error {
	reg, err := LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("InitDynamicRegistry: %w", err)
	}
	for kind, td := range reg.Types {
		if _, hasSpecialized := artifactRegistry[kind]; !hasSpecialized {
			RegisterArtifact(kind, genericArtifact{typeDef: td})
		}
	}
	return nil
}

func getArtifact(kind string) (Artifact, error) {
	a, ok := artifactRegistry[kind]
	if !ok {
		return nil, fmt.Errorf("unknown artifact kind %q", kind)
	}
	return a, nil
}

func artifactField(ab *dsl.ArtifactBlock, key string) string {
	s, _ := dsl.FieldString(ab.Items, key)
	return s
}

// --- contract ---

type contractArtifact struct{}

func (contractArtifact) Kind() string { return KindContract }

func (contractArtifact) ResolvePath(root string, ab *dsl.ArtifactBlock) (string, error) {
	id := ab.Name
	if id == "" {
		return "", fmt.Errorf("contract artifact must have an ID")
	}
	status := artifactField(ab, FieldStatus)
	if status == "" {
		status = StatusDraft
	}
	subDir := ActiveDir
	if status == StatusComplete || status == StatusAbandoned {
		subDir = ArchiveDir
	}
	return filepath.Join(root, MosDir, DirContracts, subDir, id, "contract.mos"), nil
}

func (contractArtifact) FindExisting(root string, id string) (string, error) {
	return FindContractPath(root, id)
}

func (contractArtifact) Validate(path, mosDir string) error {
	if ValidateContract == nil {
		return nil
	}
	return ValidateContract(path, mosDir)
}

// --- rule ---

type ruleArtifact struct{}

func (ruleArtifact) Kind() string { return KindRule }

func (ruleArtifact) ResolvePath(root string, ab *dsl.ArtifactBlock) (string, error) {
	id := ab.Name
	if id == "" {
		return "", fmt.Errorf("rule artifact must have an ID")
	}
	ruleType := artifactField(ab, "type")
	if ruleType == "" {
		return "", fmt.Errorf("rule artifact must have a type field")
	}
	if ruleType != "mechanical" && ruleType != "interpretive" {
		return "", fmt.Errorf("rule type must be mechanical or interpretive; got %q", ruleType)
	}
	return filepath.Join(root, MosDir, DirRules, ruleType, id+".mos"), nil
}

func (ruleArtifact) FindExisting(root string, id string) (string, error) {
	return findRulePath(root, id)
}

func (ruleArtifact) Validate(path, mosDir string) error {
	if ValidateRule == nil {
		return nil
	}
	return ValidateRule(path, mosDir)
}

// --- lexicon ---

type lexiconArtifact struct{}

func (lexiconArtifact) Kind() string { return KindLexicon }

func (lexiconArtifact) ResolvePath(root string, _ *dsl.ArtifactBlock) (string, error) {
	return filepath.Join(root, MosDir, "lexicon", "default.mos"), nil
}

func (lexiconArtifact) FindExisting(root string, _ string) (string, error) {
	p := filepath.Join(root, MosDir, "lexicon", "default.mos")
	if _, err := os.Stat(p); err != nil {
		return "", fmt.Errorf("lexicon file not found at %s", p)
	}
	return p, nil
}

func (lexiconArtifact) Validate(path, mosDir string) error {
	_, err := dsl.ReadArtifact(path)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

func mustRead(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// ApplyArtifact creates or updates an artifact from raw DSL content.
// The artifact kind and ID are extracted from the parsed content.
func ApplyArtifact(root string, content []byte) (string, error) {
	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return "", fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	f, err := dsl.Parse(string(content), nil) // parses content bytes, not file: cannot migrate to WithArtifact
	if err != nil {
		return "", fmt.Errorf("parsing artifact: %w", err)
	}
	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return "", fmt.Errorf("content is not a valid artifact")
	}

	impl, err := getArtifact(ab.Kind)
	if err != nil {
		if loadErr := InitDynamicRegistry(root); loadErr == nil {
			impl, err = getArtifact(ab.Kind)
		}
		if err != nil {
			return "", fmt.Errorf("ApplyArtifact: %w", err)
		}
	}

	targetPath, err := impl.ResolvePath(root, ab)
	if err != nil {
		return "", fmt.Errorf("ApplyArtifact: %w", err)
	}

	// Find old path for move detection.
	var oldPath string
	if existing, err := impl.FindExisting(root, ab.Name); err == nil {
		oldPath = existing
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), DirPerm); err != nil {
		return "", fmt.Errorf("creating directory: %w", err)
	}
	if err := writeArtifact(targetPath, f); err != nil {
		return "", fmt.Errorf("writing artifact: %w", err)
	}

	if err := impl.Validate(targetPath, mosDir); err != nil {
		os.Remove(targetPath)
		return "", fmt.Errorf("ApplyArtifact: %w", err)
	}

	// Clean up old location if the artifact moved.
	if oldPath != "" && oldPath != targetPath {
		oldDir := filepath.Dir(oldPath)
		targetDir := filepath.Dir(targetPath)
		if oldDir != targetDir {
			// Rules use flat files; everything else uses directory-per-instance
			if ab.Kind == KindRule {
				os.Remove(oldPath)
			} else {
				os.RemoveAll(oldDir)
			}
		}
	}

	return targetPath, nil
}

// EditArtifact opens an artifact in $EDITOR, then applies the result.
func EditArtifact(root, kind, id string) error {
	impl, err := getArtifact(kind)
	if err != nil {
		return fmt.Errorf("EditArtifact: %w", err)
	}

	existingPath, err := impl.FindExisting(root, id)
	if err != nil {
		return fmt.Errorf("EditArtifact: %w", err)
	}

	data, err := os.ReadFile(existingPath)
	if err != nil {
		return fmt.Errorf("reading artifact: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "mos-edit-*.mos")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing temp file: %w", err)
	}
	tmpFile.Close()

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command("sh", "-c", editor+" "+tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	edited, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("reading edited file: %w", err)
	}

	if string(edited) == string(data) {
		return nil
	}

	_, err = ApplyArtifact(root, edited)
	if err != nil {
		return fmt.Errorf("EditArtifact: %w", err)
	}
	return nil
}
