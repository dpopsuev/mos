package factory

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

// BlockOps controls which block-manipulation subcommands are included.
type BlockOps struct {
	AddSection     bool
	AddFeature     bool
	AddScenario    bool
	AddCriterion   bool
	RemoveBlock    bool
	SetHarness     bool
	AddCoverage    bool
	AddBill        bool
	AddSpec        bool
	AddBlame       bool
	RemoveScenario bool
	SetSection     bool
	SetField       bool
}

// AllBlocks enables every block-manipulation subcommand.
var AllBlocks = BlockOps{
	AddSection: true, AddFeature: true, AddScenario: true,
	AddCriterion: true, RemoveBlock: true, SetHarness: true,
	AddCoverage: true, AddBill: true, AddSpec: true, AddBlame: true,
	RemoveScenario: true, SetSection: true, SetField: true,
}

// KindConfig declares how a CLI command tree is built for an artifact kind.
// Nil command fields fall through to the generic implementation.
type KindConfig struct {
	TD registry.ArtifactTypeDef

	// Use overrides the command name (defaults to TD.Kind).
	Use string

	Create *cobra.Command
	Update *cobra.Command
	List   *cobra.Command
	Show   *cobra.Command
	Status *cobra.Command
	Delete *cobra.Command
	Apply  *cobra.Command
	Edit   *cobra.Command

	Blocks BlockOps
	Extra  []*cobra.Command
}

// Register builds a complete Cobra command tree from cfg.
// For each CRUD slot, it uses the custom command if provided,
// otherwise it generates the standard generic implementation.
func Register(cfg KindConfig) *cobra.Command {
	td := cfg.TD
	use := cfg.Use
	if use == "" {
		use = td.Kind
	}
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Manage %s artifacts", td.Kind),
	}

	cmd.AddCommand(pick(cfg.Create, defaultCreate(td)))
	cmd.AddCommand(pick(cfg.Update, defaultUpdate(td)))
	cmd.AddCommand(pick(cfg.List, defaultList(td)))
	cmd.AddCommand(pick(cfg.Show, defaultShow(td)))
	cmd.AddCommand(pick(cfg.Status, defaultStatus(td)))
	cmd.AddCommand(pick(cfg.Delete, defaultDelete(td)))
	cmd.AddCommand(pick(cfg.Apply, defaultApply(td)))
	cmd.AddCommand(pick(cfg.Edit, defaultEdit(td)))

	addBlockCmds(cmd, td, cfg.Blocks)

	for _, extra := range cfg.Extra {
		cmd.AddCommand(extra)
	}

	return cmd
}

func pick(custom, generic *cobra.Command) *cobra.Command {
	if custom != nil {
		return custom
	}
	return generic
}

// --- default CRUD commands (mirror the existing generic implementations) ---

func defaultCreate(td registry.ArtifactTypeDef) *cobra.Command {
	var project, format, fromTemplate string
	cmd := &cobra.Command{
		Use:   "create",
		Short: fmt.Sprintf("Create a new %s", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			fields, positional := cliutil.ParseKVArgs(args)
			var id string
			if len(positional) > 0 {
				id = positional[0]
			}
			if id == "" && project != "" {
				generated, err := registry.NextID(rootDir, project)
				if err != nil {
					return fmt.Errorf("mos %s create: auto-ID from project %q: %w", td.Kind, project, err)
				}
				id = generated
			}
			if id == "" && td.Prefix != "" {
				generated, err := registry.NextIDForType(rootDir, td.Prefix, td.Directory)
				if err != nil {
					return fmt.Errorf("mos %s create: auto-ID: %w", td.Kind, err)
				}
				id = generated
			}
			if id == "" {
				return fmt.Errorf("mos %s create: id or --project required", td.Kind)
			}
			path, err := artifact.GenericCreateWithTemplate(rootDir, td, id, fields, fromTemplate)
			if err != nil {
				return fmt.Errorf("mos %s create: %w", td.Kind, err)
			}
			if format == names.FormatJSON {
				data, _ := json.MarshalIndent(map[string]string{
					"id": id, "kind": td.Kind, "path": path,
				}, "", "  ")
				fmt.Println(string(data))
			} else {
				fmt.Printf("Created %s: %s\n", td.Kind, path)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&project, "project", "", "Project name for auto-ID generation")
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	cmd.Flags().StringVar(&fromTemplate, "from-template", "", "Template name from .mos/templates/")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func defaultUpdate(td registry.ArtifactTypeDef) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update a %s", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			fields, ids := cliutil.ParseKVArgs(args)
			if len(ids) == 0 {
				return fmt.Errorf("usage: mos %s update <id>... [--field value ...]", td.Kind)
			}
			var results []map[string]string
			for _, id := range ids {
				if err := artifact.GenericUpdate(rootDir, td, id, fields); err != nil {
					return fmt.Errorf("mos %s update: %s: %w", td.Kind, id, err)
				}
				if format == names.FormatJSON {
					results = append(results, map[string]string{"id": id, "kind": td.Kind})
				} else {
					fmt.Printf("Updated %s %s\n", td.Kind, id)
				}
			}
			if format == names.FormatJSON {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(data))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	cmd.Flags().SetInterspersed(false)
	return cmd
}

func defaultList(td registry.ArtifactTypeDef) *cobra.Command {
	var format, statusFilter string
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List %s artifacts", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			items, err := artifact.GenericList(rootDir, td, statusFilter)
			if err != nil {
				return fmt.Errorf("mos %s list: %w", td.Kind, err)
			}
			switch format {
			case names.FormatJSON:
				data, err := json.MarshalIndent(items, "", "  ")
				if err != nil {
					return fmt.Errorf("mos %s list: %w", td.Kind, err)
				}
				fmt.Println(string(data))
			case names.FormatText:
				if len(items) == 0 {
					fmt.Printf("(no %s found)\n", td.Directory)
					return nil
				}
				for _, item := range items {
					fmt.Printf("  %-20s %-12s %s\n", item.ID, item.Status, item.Title)
				}
			default:
				return fmt.Errorf("mos %s list: unknown format %q", td.Kind, format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	cmd.Flags().StringVar(&statusFilter, "status", "", "Filter by status")
	return cmd
}

func defaultShow(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: fmt.Sprintf("Show %s content", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var id string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					id = arg
					break
				}
			}
			if id == "" {
				return fmt.Errorf("usage: mos %s show <id>", td.Kind)
			}
			content, err := artifact.GenericShow(rootDir, td, id)
			if err != nil {
				return fmt.Errorf("mos %s show: %w", td.Kind, err)
			}
			fmt.Print(content)
			return nil
		},
	}
}

func defaultStatus(td registry.ArtifactTypeDef) *cobra.Command {
	var format, ids string
	cmd := &cobra.Command{
		Use:   "status",
		Short: fmt.Sprintf("Update %s status", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var idList []string
			var newStatus string
			if ids != "" {
				if len(args) < 1 {
					return fmt.Errorf("usage: mos %s status --ids <id,id,...> <new-status>", td.Kind)
				}
				newStatus = args[0]
				for _, id := range strings.Split(ids, ",") {
					id = strings.TrimSpace(id)
					if id != "" {
						idList = append(idList, id)
					}
				}
			} else {
				if len(args) < 2 {
					return fmt.Errorf("usage: mos %s status <id> <new-status>", td.Kind)
				}
				newStatus = args[len(args)-1]
				idList = args[:len(args)-1]
			}
			var results []map[string]string
			for _, id := range idList {
				if err := artifact.GenericUpdateStatus(rootDir, td, id, newStatus); err != nil {
					return fmt.Errorf("mos %s status: %s: %w", td.Kind, id, err)
				}
				if format == names.FormatJSON {
					results = append(results, map[string]string{"id": id, "kind": td.Kind, "status": newStatus})
				} else {
					fmt.Printf("Updated %s %s status to %s\n", td.Kind, id, newStatus)
				}
			}
			if format == names.FormatJSON {
				data, _ := json.MarshalIndent(results, "", "  ")
				fmt.Println(string(data))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated artifact IDs for bulk operation")
	return cmd
}

func defaultDelete(td registry.ArtifactTypeDef) *cobra.Command {
	var ids string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: fmt.Sprintf("Delete a %s", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var idList []string
			if ids != "" {
				for _, id := range strings.Split(ids, ",") {
					id = strings.TrimSpace(id)
					if id != "" {
						idList = append(idList, id)
					}
				}
			}
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					idList = append(idList, arg)
				}
			}
			if len(idList) == 0 {
				return fmt.Errorf("usage: mos %s delete <id>...", td.Kind)
			}
			for _, id := range idList {
				if err := artifact.GenericDelete(rootDir, td, id); err != nil {
					return fmt.Errorf("mos %s delete: %s: %w", td.Kind, id, err)
				}
				fmt.Printf("Deleted %s %s\n", td.Kind, id)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&ids, "ids", "", "Comma-separated artifact IDs for bulk operation")
	return cmd
}

func defaultApply(td registry.ArtifactTypeDef) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply artifact from file",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("usage: mos %s apply -f <path|->", td.Kind)
			}
			var content []byte
			var err error
			if filePath == "-" {
				content, err = io.ReadAll(os.Stdin)
			} else {
				content, err = os.ReadFile(filePath)
			}
			if err != nil {
				return fmt.Errorf("mos %s apply: reading: %w", td.Kind, err)
			}
			resultPath, err := artifact.ApplyArtifact(rootDir, content)
			if err != nil {
				return fmt.Errorf("mos %s apply: %w", td.Kind, err)
			}
			fmt.Printf("Applied %s: %s\n", td.Kind, resultPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File path or - for stdin")
	return cmd
}

func defaultEdit(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: fmt.Sprintf("Edit %s in $EDITOR", td.Kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var id string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					id = arg
					break
				}
			}
			if td.Kind != names.KindLexicon && id == "" {
				return fmt.Errorf("usage: mos %s edit <id>", td.Kind)
			}
			if err := artifact.EditArtifact(rootDir, td.Kind, id); err != nil {
				return fmt.Errorf("mos %s edit: %w", td.Kind, err)
			}
			fmt.Printf("Edited %s\n", td.Kind)
			return nil
		},
	}
}

// --- kind-string convenience commands (used by spec, binder, lexicon) ---

// ApplyCmd returns an apply command for a given kind string.
func ApplyCmd(kind string) *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply artifact from file",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if filePath == "" {
				return fmt.Errorf("usage: mos %s apply -f <path|->", kind)
			}
			var content []byte
			var err error
			if filePath == "-" {
				content, err = io.ReadAll(os.Stdin)
			} else {
				content, err = os.ReadFile(filePath)
			}
			if err != nil {
				return fmt.Errorf("mos %s apply: reading: %w", kind, err)
			}
			resultPath, err := artifact.ApplyArtifact(rootDir, content)
			if err != nil {
				return fmt.Errorf("mos %s apply: %w", kind, err)
			}
			fmt.Printf("Applied %s: %s\n", kind, resultPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File path or - for stdin")
	return cmd
}

// EditCmd returns an edit command for a given kind string.
func EditCmd(kind string) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: fmt.Sprintf("Edit %s in $EDITOR", kind),
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var id string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					id = arg
					break
				}
			}
			if kind != names.KindLexicon && id == "" {
				return fmt.Errorf("usage: mos %s edit <id>", kind)
			}
			if err := artifact.EditArtifact(rootDir, kind, id); err != nil {
				return fmt.Errorf("mos %s edit: %w", kind, err)
			}
			fmt.Printf("Edited %s\n", kind)
			return nil
		},
	}
}

// --- block commands ---

func addBlockCmds(parent *cobra.Command, td registry.ArtifactTypeDef, b BlockOps) {
	if b.AddSection {
		parent.AddCommand(AddSectionCmd(td))
	}
	if b.AddFeature {
		parent.AddCommand(AddFeatureCmd(td))
	}
	if b.AddScenario {
		parent.AddCommand(AddScenarioCmd(td))
	}
	if b.AddCriterion {
		parent.AddCommand(AddCriterionCmd(td))
	}
	if b.RemoveBlock {
		parent.AddCommand(RemoveBlockCmd(td))
	}
	if b.SetHarness {
		parent.AddCommand(SetHarnessCmd(td))
	}
	if b.AddCoverage {
		parent.AddCommand(AddCoverageCmd(td))
	}
	if b.AddBill {
		parent.AddCommand(AddBillCmd(td))
	}
	if b.AddBlame {
		parent.AddCommand(AddBlameCmd(td))
	}
	if b.AddSpec {
		parent.AddCommand(AddSpecCmd(td))
	}
	if b.RemoveScenario {
		parent.AddCommand(RemoveScenarioCmd(td))
	}
	if b.SetSection {
		parent.AddCommand(SetSectionCmd(td))
	}
	if b.SetField {
		parent.AddCommand(SetFieldCmd(td))
	}
}

// AddSectionCmd returns a command for adding a section.
func AddSectionCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var name, text string
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "add-section",
		Short: "Add a section to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos %s add-section <id> <section-name> [--text <text> | --stdin]", td.Kind)
			}
			id := args[0]
			if name == "" && len(args) >= 2 {
				name = args[1]
			}
			if name == "" {
				return fmt.Errorf("mos %s add-section: section name required (positional or --name)", td.Kind)
			}
			if fromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("mos %s add-section: reading stdin: %w", td.Kind, err)
				}
				text = string(data)
			}
			if text == "" {
				return fmt.Errorf("mos %s add-section: provide content via --text or --stdin", td.Kind)
			}
			if err := artifact.GenericAddSection(rootDir, td, id, name, text); err != nil {
				return fmt.Errorf("mos %s add-section: %w", td.Kind, err)
			}
			fmt.Printf("Added section %q to %s %s\n", name, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Section name")
	cmd.Flags().StringVar(&text, "text", "", "Section content")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read content from stdin")
	return cmd
}

// AddFeatureCmd returns a command for adding a feature.
func AddFeatureCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var name, description string
	cmd := &cobra.Command{
		Use:   "add-feature",
		Short: "Add a feature to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos %s add-feature <id> <feature-name> [--description <text>]", td.Kind)
			}
			id := args[0]
			if name == "" && len(args) >= 2 {
				name = args[1]
			}
			if name == "" {
				return fmt.Errorf("mos %s add-feature: feature name required (positional or --name)", td.Kind)
			}
			if err := artifact.GenericAddFeature(rootDir, td, id, name, description); err != nil {
				return fmt.Errorf("mos %s add-feature: %w", td.Kind, err)
			}
			fmt.Printf("Added feature %q to %s %s\n", name, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Feature name")
	cmd.Flags().StringVar(&description, "description", "", "Feature description")
	return cmd
}

// AddScenarioCmd returns a command for adding a scenario.
func AddScenarioCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var scenarioName, given, when, then string
	cmd := &cobra.Command{
		Use:   "add-scenario",
		Short: "Add a scenario to a feature",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mos %s add-scenario <id> <feature-name> [scenario-name] --given <text> --when <text> --then <text>", td.Kind)
			}
			id, featureName := args[0], args[1]
			if scenarioName == "" && len(args) >= 3 {
				scenarioName = args[2]
			}
			if scenarioName == "" {
				return fmt.Errorf("mos %s add-scenario: scenario name required (positional or --scenario)", td.Kind)
			}
			if given == "" || when == "" || then == "" {
				return fmt.Errorf("mos %s add-scenario: --given, --when, and --then are all required", td.Kind)
			}
			if err := artifact.GenericAddScenario(rootDir, td, id, featureName, scenarioName, given, when, then); err != nil {
				return fmt.Errorf("mos %s add-scenario: %w", td.Kind, err)
			}
			fmt.Printf("Added scenario %q to feature %q on %s %s\n", scenarioName, featureName, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&scenarioName, "scenario", "", "Scenario name")
	cmd.Flags().StringVar(&given, "given", "", "Given step")
	cmd.Flags().StringVar(&when, "when", "", "When step")
	cmd.Flags().StringVar(&then, "then", "", "Then step")
	return cmd
}

// AddCriterionCmd returns a command for adding a criterion.
// Exported so that packages with custom command trees can selectively include it.
func AddCriterionCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "add-criterion",
		Short: "Add a criterion to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mos %s add-criterion <id> <criterion-name> [--description <text>]", td.Kind)
			}
			id, criterionName := args[0], args[1]
			if err := artifact.GenericAddCriterion(rootDir, td, id, criterionName, description); err != nil {
				return fmt.Errorf("mos %s add-criterion: %w", td.Kind, err)
			}
			fmt.Printf("Added criterion %q to %s %s\n", criterionName, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&description, "description", "", "Criterion description")
	return cmd
}

// RemoveBlockCmd returns a command for removing a block.
func RemoveBlockCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-block",
		Short: "Remove a block from an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mos %s remove-block <id> <block-type> [block-name]", td.Kind)
			}
			id, blockType := args[0], args[1]
			blockName := ""
			if len(args) > 2 {
				blockName = args[2]
			}
			if err := artifact.GenericRemoveBlock(rootDir, td, id, blockType, blockName); err != nil {
				return fmt.Errorf("mos %s remove-block: %w", td.Kind, err)
			}
			if blockName != "" {
				fmt.Printf("Removed %s block %q from %s %s\n", blockType, blockName, td.Kind, id)
			} else {
				fmt.Printf("Removed %s block from %s %s\n", blockType, td.Kind, id)
			}
			return nil
		},
	}
}

// AddCoverageCmd returns a command for adding coverage.
func AddCoverageCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "add-coverage",
		Short: "Add coverage to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mos %s add-coverage <id> <key=value> [key=value ...]", td.Kind)
			}
			id := args[0]
			fields := make(map[string]string)
			for _, kv := range args[1:] {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return fmt.Errorf("mos %s add-coverage: invalid key=value pair %q", td.Kind, kv)
				}
				fields[k] = v
			}
			if err := artifact.GenericAddCoverage(rootDir, td, id, fields); err != nil {
				return fmt.Errorf("mos %s add-coverage: %w", td.Kind, err)
			}
			fmt.Printf("Set coverage on %s %s\n", td.Kind, id)
			return nil
		},
	}
}

// AddBillCmd returns a command for adding a bill.
func AddBillCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var introducedBy, intent string
	cmd := &cobra.Command{
		Use:   "add-bill",
		Short: "Add a bill to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos %s add-bill <id> --introduced-by <identity> --intent <text>", td.Kind)
			}
			id := args[0]
			if introducedBy == "" || intent == "" {
				return fmt.Errorf("mos %s add-bill: --introduced-by and --intent are required", td.Kind)
			}
			if err := artifact.GenericAddBill(rootDir, td, id, introducedBy, intent); err != nil {
				return fmt.Errorf("mos %s add-bill: %w", td.Kind, err)
			}
			fmt.Printf("Added bill to %s %s\n", td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&introducedBy, "introduced-by", "", "Introduced by")
	cmd.Flags().StringVar(&intent, "intent", "", "Intent")
	return cmd
}

// AddSpecCmd returns a command for adding spec include paths.
func AddSpecCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "add-spec",
		Short: "Add spec include paths to an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: mos %s add-spec <id> <include-path> [include-path ...]", td.Kind)
			}
			id := args[0]
			includes := args[1:]
			if err := artifact.GenericAddSpec(rootDir, td, id, includes); err != nil {
				return fmt.Errorf("mos %s add-spec: %w", td.Kind, err)
			}
			fmt.Printf("Set spec block on %s %s\n", td.Kind, id)
			return nil
		},
	}
}

// AddBlameCmd returns a command for adding a blame block (source reference) to an artifact.
func AddBlameCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var lines, symbol string
	cmd := &cobra.Command{
		Use:   "add-blame <id> <file>",
		Short: "Add a source reference (blame block) to a bug contract",
		Long: `Link a contract to the specific source file and line range where the
defect lives. Multiple blame blocks are supported (one per file).

Examples:
  mos contract add-blame CON-2026-225 cmd/mos/ci/cmd_test.go --lines 56-67
  mos contract add-blame CON-2026-225 cmd/mos/ci/cmd_test.go --lines 62 --symbol runCICmd`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id, file := args[0], args[1]
			if err := artifact.GenericAddBlame(rootDir, td, id, file, lines, symbol); err != nil {
				return fmt.Errorf("mos %s add-blame: %w", td.Kind, err)
			}
			fmt.Printf("Added blame %s to %s %s\n", file, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&lines, "lines", "", "Line or line range (e.g. 42, 56-67)")
	cmd.Flags().StringVar(&symbol, "symbol", "", "Go symbol name (e.g. runCICmd)")
	return cmd
}

// SetHarnessCmd returns a command for setting the harness.
func SetHarnessCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var command, timeout string
	cmd := &cobra.Command{
		Use:   "set-harness",
		Short: "Set harness command",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos %s set-harness <id> --command <cmd> [--timeout <dur>]", td.Kind)
			}
			id := args[0]
			if command == "" {
				return fmt.Errorf("mos %s set-harness: --command is required", td.Kind)
			}
			if err := artifact.GenericSetHarness(rootDir, td, id, command, timeout); err != nil {
				return fmt.Errorf("mos %s set-harness: %w", td.Kind, err)
			}
			fmt.Printf("Set harness on %s %s\n", td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "Harness command")
	cmd.Flags().StringVar(&timeout, "timeout", "", "Timeout duration")
	return cmd
}

// RemoveScenarioCmd returns a command for removing a scenario.
func RemoveScenarioCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "remove-scenario",
		Short: "Remove a scenario from a feature",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("usage: mos %s remove-scenario <id> <feature-name> <scenario-name>", td.Kind)
			}
			id, featureName, scenarioName := args[0], args[1], args[2]
			if err := artifact.GenericRemoveScenario(rootDir, td, id, featureName, scenarioName); err != nil {
				return fmt.Errorf("mos %s remove-scenario: %w", td.Kind, err)
			}
			fmt.Printf("Removed scenario %q from feature %q on %s %s\n", scenarioName, featureName, td.Kind, id)
			return nil
		},
	}
}

// SetSectionCmd returns a command for setting (add-or-replace) a section.
func SetSectionCmd(td registry.ArtifactTypeDef) *cobra.Command {
	var name, text string
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "set-section",
		Short: "Set (add or replace) a section on an artifact",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos %s set-section <id> <section-name> [--text <text> | --stdin]", td.Kind)
			}
			id := args[0]
			if name == "" && len(args) >= 2 {
				name = args[1]
			}
			if name == "" {
				return fmt.Errorf("mos %s set-section: section name required (positional or --name)", td.Kind)
			}
			if fromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("mos %s set-section: reading stdin: %w", td.Kind, err)
				}
				text = string(data)
			}
			if text == "" {
				return fmt.Errorf("mos %s set-section: provide content via --text or --stdin", td.Kind)
			}
			if err := artifact.GenericAddSection(rootDir, td, id, name, text); err != nil {
				return fmt.Errorf("mos %s set-section: %w", td.Kind, err)
			}
			fmt.Printf("Set section %q on %s %s\n", name, td.Kind, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Section name")
	cmd.Flags().StringVar(&text, "text", "", "Section content")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read content from stdin")
	return cmd
}

// SetFieldCmd returns a command for setting a field via dot-path.
func SetFieldCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "set-field",
		Short: "Set a field value via dot-path",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("usage: mos %s set-field <id> <path> <value>", td.Kind)
			}
			id, fieldPath, value := args[0], args[1], args[2]
			if err := artifact.SetArtifactField(rootDir, id, fieldPath, value); err != nil {
				return fmt.Errorf("mos %s set-field: %w", td.Kind, err)
			}
			fmt.Printf("Set %s = %s on %s %s\n", fieldPath, value, td.Kind, id)
			return nil
		},
	}
}
