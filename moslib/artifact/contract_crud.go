package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ValidStatuses is the set of valid contract status values.
var ValidStatuses = map[string]bool{StatusDraft: true, StatusActive: true, StatusComplete: true, StatusAbandoned: true}

// ContractOpts configures the mos contract create command.
type ContractOpts struct {
	Title        string   // contract title
	Status       string   // draft | active | complete | abandoned
	Goal         string   // contract goal (desired state)
	Precondition []string // precondition artifact IDs
	DependsOn    []string // contract IDs this depends on
	SpecFile     string   // path to file containing feature/scenario blocks to inline
	CoverageFile string   // path to file containing a coverage block to inline
	Kind         string   // free-form kind (bug, feature, task, etc.)
	Labels       []string // free-form labels
	Priority     string   // free-form priority (p1, p2, etc.)
	Project      string   // project name from config.mos for auto-ID generation
	Parent       string   // parent contract ID for hierarchy
	Branches     []string // branch/version scope annotations
	Specs        []string // referenced specification IDs
	Justifies    []string // need IDs this contract justifies
	Sprint       []string // sprint IDs this contract belongs to
	Batch        []string // batch IDs this contract belongs to
	Template     string   // template name from .mos/templates/
}

// CreateContract creates a contract artifact under .mos/contracts/.
// Active/draft contracts go in active/<id>/contract.mos;
// complete/abandoned contracts go in archive/<id>/contract.mos.
func CreateContract(root, id string, opts ContractOpts) (string, error) {
	if id == "" && opts.Project != "" {
		generated, err := NextID(root, opts.Project)
		if err != nil {
			return "", fmt.Errorf("auto-ID from project %q: %w", opts.Project, err)
		}
		id = generated
	} else if id == "" {
		projects, err := LoadProjects(root)
		if err == nil {
			var proj *ProjectDef
			if opts.Kind != "" {
				proj = FindProjectByPrefix(projects, opts.Kind)
			}
			if proj == nil {
				proj = FindDefaultProject(projects)
			}
			if proj != nil {
				generated, err := NextID(root, proj.Name)
				if err != nil {
					return "", fmt.Errorf("auto-ID from project %q: %w", proj.Name, err)
				}
				id = generated
			}
		}
		if id == "" {
			return "", fmt.Errorf("contract id is required (no default project configured)")
		}
	}
	if opts.Title == "" {
		opts.Title = id
	}
	if opts.Status == "" {
		opts.Status = StatusDraft
	}

	if !ValidStatuses[opts.Status] {
		return "", fmt.Errorf("--status must be draft, active, complete, or abandoned; got %q", opts.Status)
	}

	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return "", fmt.Errorf(".mos/ directory not found; run mos init first")
	}

	subDir := ActiveDir
	if opts.Status == StatusComplete || opts.Status == StatusAbandoned {
		subDir = ArchiveDir
	}

	contractDir := filepath.Join(mosDir, "contracts", subDir, id)
	if err := os.MkdirAll(contractDir, DirPerm); err != nil {
		return "", fmt.Errorf("creating contract directory: %w", err)
	}

	contractPath := filepath.Join(contractDir, "contract.mos")

	now := time.Now().UTC().Format(time.RFC3339)
	items := []dsl.Node{
		&dsl.Field{Key: FieldTitle, Value: &dsl.StringVal{Text: opts.Title}},
		&dsl.Field{Key: FieldStatus, Value: &dsl.StringVal{Text: opts.Status}},
		&dsl.Field{Key: "created_at", Value: &dsl.DateTimeVal{Raw: now}},
		&dsl.Field{Key: "updated_at", Value: &dsl.DateTimeVal{Raw: now}},
	}

	if opts.Goal != "" {
		items = append(items, &dsl.Field{Key: FieldGoal, Value: &dsl.StringVal{Text: opts.Goal}})
	}
	if len(opts.Precondition) > 0 {
		items = append(items, &dsl.Field{Key: "precondition", Value: sliceRefValue(opts.Precondition)})
	}
	if len(opts.Justifies) > 0 {
		items = append(items, &dsl.Field{Key: "justifies", Value: sliceRefValue(opts.Justifies)})
	}
	if len(opts.Sprint) > 0 {
		items = append(items, &dsl.Field{Key: "sprint", Value: sliceRefValue(opts.Sprint)})
	}
	if len(opts.Batch) > 0 {
		items = append(items, &dsl.Field{Key: "batch", Value: sliceRefValue(opts.Batch)})
	}
	if opts.Kind != "" {
		items = append(items, &dsl.Field{Key: "kind", Value: &dsl.StringVal{Text: opts.Kind}})
	}
	if len(opts.Labels) > 0 {
		labelValues := make([]dsl.Value, len(opts.Labels))
		for i, l := range opts.Labels {
			labelValues[i] = &dsl.StringVal{Text: l}
		}
		items = append(items, &dsl.Field{Key: "labels", Value: &dsl.ListVal{Items: labelValues}})
	}
	if opts.Priority != "" {
		items = append(items, &dsl.Field{Key: "priority", Value: &dsl.StringVal{Text: opts.Priority}})
	}
	if opts.Parent != "" {
		items = append(items, &dsl.Field{Key: "parent", Value: &dsl.StringVal{Text: opts.Parent}})
	}
	if len(opts.Branches) > 0 {
		branchValues := make([]dsl.Value, len(opts.Branches))
		for i, b := range opts.Branches {
			branchValues[i] = &dsl.StringVal{Text: b}
		}
		items = append(items, &dsl.Field{Key: "branches", Value: &dsl.ListVal{Items: branchValues}})
	}
	if len(opts.Specs) > 0 {
		specValues := make([]dsl.Value, len(opts.Specs))
		for i, s := range opts.Specs {
			specValues[i] = &dsl.StringVal{Text: s}
		}
		items = append(items, &dsl.Field{Key: "specs", Value: &dsl.ListVal{Items: specValues}})
	}
	if len(opts.DependsOn) > 0 {
		if err := ValidateDependencies(root, opts.DependsOn); err != nil {
			return "", fmt.Errorf("CreateContract: %w", err)
		}
		depValues := make([]dsl.Value, len(opts.DependsOn))
		for i, dep := range opts.DependsOn {
			depValues[i] = &dsl.StringVal{Text: dep}
		}
		items = append(items, &dsl.Block{
			Name: "scope",
			Items: []dsl.Node{
				&dsl.Field{Key: "depends_on", Value: &dsl.ListVal{Items: depValues}},
			},
		})
	}

	if opts.CoverageFile != "" {
		coverageNodes, err := parseCoverageFile(opts.CoverageFile)
		if err != nil {
			return "", fmt.Errorf("parsing --coverage-file: %w", err)
		}
		items = append(items, coverageNodes...)
	}

	if opts.SpecFile != "" {
		specNodes, err := parseSpecFile(opts.SpecFile)
		if err != nil {
			return "", fmt.Errorf("parsing --spec-file: %w", err)
		}
		items = append(items, specNodes...)
	}

	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  "contract",
			Name:  id,
			Items: items,
		},
	}

	if err := writeArtifact(contractPath, file); err != nil {
		return "", fmt.Errorf("writing contract: %w", err)
	}

	if opts.Template != "" {
		tmpl, err := LoadTemplate(root, opts.Template)
		if err != nil {
			os.RemoveAll(contractDir)
			return "", fmt.Errorf("CreateContract: %w", err)
		}
		if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
			MergeTemplate(ab, tmpl)
			return nil
		}); err != nil {
			os.RemoveAll(contractDir)
			return "", fmt.Errorf("CreateContract template merge: %w", err)
		}
	}

	if ValidateContract != nil {
		if err := ValidateContract(contractPath, mosDir); err != nil {
			os.RemoveAll(contractDir)
			return "", fmt.Errorf("CreateContract: %w", err)
		}
	}

	return contractPath, nil
}

// UpdateContractStatus updates the status field and moves the contract
// between active/ and archive/ as needed.
func UpdateContractStatus(root, id, newStatus string) error {
	if !ValidStatuses[newStatus] {
		return fmt.Errorf("status must be draft, active, complete, or abandoned; got %q", newStatus)
	}

	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("UpdateContractStatus: %w", err)
	}

	var oldStatus string
	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		oldStatus, _ = dsl.FieldString(ab.Items, FieldStatus)
		dsl.SetField(&ab.Items, FieldStatus, &dsl.StringVal{Text: newStatus})
		return nil
	}); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}

	oldDir := filepath.Dir(contractPath)
	newSubDir := ActiveDir
	if newStatus == StatusComplete || newStatus == StatusAbandoned {
		newSubDir = ArchiveDir
	}

	mosDir := filepath.Join(root, MosDir)
	newDir := filepath.Join(mosDir, "contracts", newSubDir, id)

	if oldDir != newDir {
		if err := os.MkdirAll(filepath.Dir(newDir), DirPerm); err != nil {
			return fmt.Errorf("creating target directory: %w", err)
		}
		if err := os.Rename(oldDir, newDir); err != nil {
			return fmt.Errorf("moving contract directory: %w", err)
		}
		contractPath = filepath.Join(newDir, "contract.mos")
	}

	if ValidateContract != nil {
		if err := ValidateContract(contractPath, mosDir); err != nil {
			return fmt.Errorf("UpdateContractStatus: %w", err)
		}
	}

	AppendContractLedger(root, id, LedgerEntry{
		Event:    "status_changed",
		Field:    FieldStatus,
		OldValue: oldStatus,
		NewValue: newStatus,
	})

	return nil
}

// ContractUpdateOpts configures the mos contract update command.
// Nil pointer fields are left unchanged; non-nil fields are applied.
type ContractUpdateOpts struct {
	Title        *string
	Goal         *string
	Precondition *[]string
	Status       *string
	SpecFile     *string
	CoverageFile *string
	DependsOn    *[]string
	Parent       *string
	Specs        *[]string
	Justifies    *[]string
	Sprint       *[]string
	Batch        *[]string
}

// UpdateContract applies partial updates to an existing contract.
func UpdateContract(root, id string, opts ContractUpdateOpts) error {
	if err := CheckSealForMutation(root, id); err != nil {
		return fmt.Errorf("UpdateContract: %w", err)
	}

	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("UpdateContract: %w", err)
	}

	if opts.Status != nil {
		if !ValidStatuses[*opts.Status] {
			return fmt.Errorf("status must be draft, active, complete, or abandoned; got %q", *opts.Status)
		}
	}

	if err := dsl.WithArtifact(contractPath, func(ab *dsl.ArtifactBlock) error {
		if opts.Title != nil {
			dsl.SetField(&ab.Items, FieldTitle, &dsl.StringVal{Text: *opts.Title})
		}
		if opts.Goal != nil {
			dsl.SetField(&ab.Items, FieldGoal, &dsl.StringVal{Text: *opts.Goal})
		}
		if opts.Precondition != nil {
			dsl.SetField(&ab.Items, "precondition", sliceRefValue(*opts.Precondition))
		}
		if opts.Sprint != nil {
			dsl.SetField(&ab.Items, "sprint", sliceRefValue(*opts.Sprint))
		}
		if opts.Batch != nil {
			dsl.SetField(&ab.Items, "batch", sliceRefValue(*opts.Batch))
		}
		if opts.Status != nil {
			dsl.SetField(&ab.Items, FieldStatus, &dsl.StringVal{Text: *opts.Status})
		}

		if opts.CoverageFile != nil {
			for dsl.RemoveBlock(&ab.Items, "coverage") {
			}
			if *opts.CoverageFile != "" {
				nodes, err := parseCoverageFile(*opts.CoverageFile)
				if err != nil {
					return fmt.Errorf("parsing --coverage-file: %w", err)
				}
				ab.Items = append(ab.Items, nodes...)
			}
		}

		if opts.SpecFile != nil {
			ab.Items = filterNodes(ab.Items, func(n dsl.Node) bool {
				_, ok := n.(*dsl.FeatureBlock)
				return ok
			})
			if *opts.SpecFile != "" {
				nodes, err := parseSpecFile(*opts.SpecFile)
				if err != nil {
					return fmt.Errorf("parsing --spec-file: %w", err)
				}
				ab.Items = append(ab.Items, nodes...)
			}
		}

		if opts.Parent != nil {
			if *opts.Parent != "" {
				if err := ValidateParent(root, id, *opts.Parent); err != nil {
					return err
				}
			}
			dsl.SetField(&ab.Items, "parent", &dsl.StringVal{Text: *opts.Parent})
		}
		if opts.Justifies != nil {
			dsl.SetField(&ab.Items, "justifies", sliceRefValue(*opts.Justifies))
		}
		if opts.Specs != nil {
			var specVal dsl.Value
			if len(*opts.Specs) > 0 {
				specValues := make([]dsl.Value, len(*opts.Specs))
				for i, s := range *opts.Specs {
					specValues[i] = &dsl.StringVal{Text: s}
				}
				specVal = &dsl.ListVal{Items: specValues}
			} else {
				specVal = &dsl.ListVal{Items: []dsl.Value{}}
			}
			dsl.SetField(&ab.Items, "specs", specVal)
		}

		if opts.DependsOn != nil {
			if err := ValidateDependencies(root, *opts.DependsOn); err != nil {
				return err
			}
			for dsl.RemoveBlock(&ab.Items, BlockScope) {
			}
			if len(*opts.DependsOn) > 0 {
				depValues := make([]dsl.Value, len(*opts.DependsOn))
				for i, dep := range *opts.DependsOn {
					depValues[i] = &dsl.StringVal{Text: dep}
				}
				ab.Items = append(ab.Items, &dsl.Block{
					Name: BlockScope,
					Items: []dsl.Node{
						&dsl.Field{Key: "depends_on", Value: &dsl.ListVal{Items: depValues}},
					},
				})
			}
		}

		touchUpdatedAt(ab)
		return nil
	}); err != nil {
		return fmt.Errorf("updating contract: %w", err)
	}

	mosDir := filepath.Join(root, MosDir)

	if opts.Status != nil {
		oldDir := filepath.Dir(contractPath)
		newSubDir := ActiveDir
		if *opts.Status == StatusComplete || *opts.Status == StatusAbandoned {
			newSubDir = ArchiveDir
		}
		newDir := filepath.Join(mosDir, "contracts", newSubDir, id)
		if oldDir != newDir {
			if err := os.MkdirAll(filepath.Dir(newDir), DirPerm); err != nil {
				return fmt.Errorf("creating target directory: %w", err)
			}
			if err := os.Rename(oldDir, newDir); err != nil {
				return fmt.Errorf("moving contract directory: %w", err)
			}
			contractPath = filepath.Join(newDir, "contract.mos")
		}
	}

	if ValidateContract != nil {
		if err := ValidateContract(contractPath, mosDir); err != nil {
			return fmt.Errorf("UpdateContract: %w", err)
		}
	}

	return nil
}

// DeleteContract removes a contract and its directory from .mos/contracts/.
// If force is false, it refuses to delete a contract that others depend on.
func DeleteContract(root, id string, force bool) error {
	contractPath, err := FindContractPath(root, id)
	if err != nil {
		return fmt.Errorf("DeleteContract: %w", err)
	}

	if !force {
		if err := CheckSealForMutation(root, id); err != nil {
			return fmt.Errorf("DeleteContract: %w", err)
		}
		dependents, err := FindDependents(root, id)
		if err != nil {
			return fmt.Errorf("DeleteContract: %w", err)
		}
		if len(dependents) > 0 {
			return fmt.Errorf("cannot delete %s: depended on by %s; use --force to override", id, strings.Join(dependents, ", "))
		}
	}

	contractDir := filepath.Dir(contractPath)
	if err := os.RemoveAll(contractDir); err != nil {
		return fmt.Errorf("removing contract directory: %w", err)
	}
	return nil
}

func parseSpecFile(path string) ([]dsl.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec file: %w", err)
	}
	return extractNodes(string(data), func(n dsl.Node) bool {
		_, ok := n.(*dsl.FeatureBlock)
		return ok
	})
}

func parseCoverageFile(path string) ([]dsl.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading coverage file: %w", err)
	}
	return extractNodes(string(data), func(n dsl.Node) bool {
		blk, ok := n.(*dsl.Block)
		return ok && blk.Name == "coverage"
	})
}

func extractNodes(content string, match func(dsl.Node) bool) ([]dsl.Node, error) {
	wrapper := fmt.Sprintf("%s \"tmp\" {\n  %s = \"tmp\"\n  %s = %q\n\n%s\n}\n", "contract", FieldTitle, FieldStatus, StatusDraft, content)
	parsed, err := dsl.Parse(wrapper, nil) // in-memory string: cannot migrate to WithArtifact
	if err != nil {
		return nil, fmt.Errorf("parsing content: %w", err)
	}
	ab, ok := parsed.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil, fmt.Errorf("unexpected AST structure")
	}
	var nodes []dsl.Node
	for _, item := range ab.Items {
		if match(item) {
			nodes = append(nodes, item)
		}
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no matching nodes found in content")
	}
	return nodes, nil
}
