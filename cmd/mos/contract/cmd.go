package contract

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	chainpkg "github.com/dpopsuev/mos/moslib/governance/chain"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

var contractTD = registry.ArtifactTypeDef{Kind: names.KindContract, Directory: names.DirContracts}

// Cmd is the contract command, built via the artifact factory.
var Cmd = factory.Register(factory.KindConfig{
	TD:     contractTD,
	Create: contractCreateCmd(),
	List:   contractListCmd(),
	Show:   contractShowCmd(),
	Status: contractStatusCmd(),
	Update: contractUpdateCmd(),
	Delete: contractDeleteCmd(),
	Apply:  contractApplyCmd(),
	Edit:   contractEditCmd(),
	Blocks: factory.BlockOps{AddSection: true, AddCriterion: true, AddBlame: true, SetSection: true, SetField: true},
	Extra: []*cobra.Command{
		graphCmd(), linkCmd(), unlinkCmd(), lockCmd(), unlockCmd(),
		renameCmd(), scenarioCmd(), contextCmd(), verifyCmd(),
		historyCmd(), chainCommand(),
	},
})

// ChainCmd is the chain command, exported for use as top-level "mos chain".
var ChainCmd = chainCommand()

// InferArtifactKind infers artifact kind from ID prefix using registry.
func InferArtifactKind(reg *registry.Registry, id string) string {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) >= 1 {
		upper := strings.ToUpper(parts[0])
		if kind, ok := reg.PrefixKind[upper]; ok {
			return kind
		}
	}
	return names.KindContract
}

// RunApply applies artifact content from a file or stdin. Exported for use by other packages.
func RunApply(kind, filePath string, out io.Writer) error {
	var content []byte
	var err error
	if filePath == "-" {
		content, err = io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("mos %s apply: reading stdin: %w", kind, err)
		}
	} else {
		content, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("mos %s apply: reading file: %w", kind, err)
		}
	}
	resultPath, err := artifact.ApplyArtifact(".", content)
	if err != nil {
		return fmt.Errorf("mos %s apply: %w", kind, err)
	}
	fmt.Fprintf(out, "Applied %s: %s\n", kind, resultPath)
	return nil
}

// RunEdit opens the artifact in $EDITOR. Exported for use by other packages.
func RunEdit(kind, id string) error {
	if err := artifact.EditArtifact(".", kind, id); err != nil {
		return fmt.Errorf("mos %s edit: %w", kind, err)
	}
	fmt.Printf("Edited %s\n", kind)
	return nil
}

// --- CRUD commands ---

func contractCreateCmd() *cobra.Command {
	var flags struct {
		title, status, goal, project, kind, priority, parent string
		specFile, coverageFile                               string
		dependsOn, labels, branches, specs                   []string
		justifies, precondition, sprint, batch               []string
	}
	cmd := &cobra.Command{
		Use:   "create [id]",
		Short: "Create a new contract",
		Long: `Create a new contract. ID is auto-generated from the default project (or --kind/--project prefix).
Provide <id> explicitly only when overriding auto-generation.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			var id string
			if len(args) > 0 {
				id = args[0]
			}
			opts := artifact.ContractOpts{
				Title:        flags.title,
				Status:       flags.status,
				Goal:         flags.goal,
				DependsOn:    flags.dependsOn,
				SpecFile:     flags.specFile,
				CoverageFile: flags.coverageFile,
				Project:      flags.project,
				Kind:         flags.kind,
				Labels:       flags.labels,
				Priority:     flags.priority,
				Parent:       flags.parent,
				Branches:     flags.branches,
				Specs:        flags.specs,
				Justifies:    flags.justifies,
				Precondition: flags.precondition,
				Sprint:       flags.sprint,
				Batch:        flags.batch,
			}
			contractPath, err := artifact.CreateContract(".", id, opts)
			if err != nil {
				return fmt.Errorf("mos contract create: %w", err)
			}
			if id == "" {
				id = filepath.Base(filepath.Dir(contractPath))
			}
			if err := cliutil.ApplyOverflowFields(names.KindContract, id, map[string]string{}); err != nil {
				return err
			}
			fmt.Printf("Created contract: %s\n", contractPath)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.title, "title", "", "Contract title (required)")
	f.StringVar(&flags.status, "status", "draft", "Initial status")
	f.StringVar(&flags.goal, "goal", "", "Goal description")
	f.StringSliceVar(&flags.dependsOn, "depends-on", nil, "Dependency ID(s)")
	f.StringVar(&flags.specFile, "spec-file", "", "Path to feature/scenario file to inline")
	f.StringVar(&flags.coverageFile, "coverage-file", "", "Path to coverage block file to inline")
	f.StringVar(&flags.project, "project", "", "Project name for auto-ID")
	f.StringVar(&flags.kind, "kind", "", "Maps to project prefix for auto-ID")
	f.StringSliceVar(&flags.labels, "labels", nil, "Labels")
	f.StringVar(&flags.priority, "priority", "", "Priority")
	f.StringVar(&flags.parent, "parent", "", "Parent contract ID")
	f.StringSliceVar(&flags.branches, "branches", nil, "Branch scope annotations")
	f.StringSliceVar(&flags.specs, "specs", nil, "Referenced specification IDs")
	f.StringSliceVar(&flags.justifies, "justifies", nil, "Need ID(s) this contract justifies")
	f.StringSliceVar(&flags.precondition, "precondition", nil, "Precondition artifact IDs")
	f.StringSliceVar(&flags.sprint, "sprint", nil, "Sprint ID(s)")
	f.StringSliceVar(&flags.batch, "batch", nil, "Batch ID(s)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func contractListCmd() *cobra.Command {
	var flags struct {
		format, status, project, kind, label, priority, parent, branch string
		roots, recursive                                               bool
	}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List contracts",
		RunE: func(c *cobra.Command, args []string) error {
			listOpts := artifact.ListOpts{
				Status:    flags.status,
				Project:   flags.project,
				Kind:      flags.kind,
				Label:     flags.label,
				Priority:  flags.priority,
				Parent:    flags.parent,
				Branch:    flags.branch,
				Roots:     flags.roots,
				Recursive: flags.recursive,
			}
			contracts, err := artifact.ListContracts(".", listOpts)
			if err != nil {
				return fmt.Errorf("mos contract list: %w", err)
			}
			switch flags.format {
			case names.FormatJSON:
				data, err := json.MarshalIndent(contracts, "", "  ")
				if err != nil {
					return fmt.Errorf("mos contract list: %w", err)
				}
				fmt.Println(string(data))
			case names.FormatText:
				if len(contracts) == 0 {
					fmt.Println("(no contracts found)")
					return nil
				}
				for _, ct := range contracts {
					fmt.Printf("  %-20s %-12s %s\n", ct.ID, ct.Status, ct.Title)
				}
			default:
				return fmt.Errorf("mos contract list: unknown format %q", flags.format)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.format, "format", names.FormatText, "Output format (text|json)")
	f.StringVar(&flags.status, "status", "", "Filter by status")
	f.StringVar(&flags.project, "project", "", "Filter by project")
	f.StringVar(&flags.kind, "kind", "", "Filter by kind")
	f.StringVar(&flags.label, "label", "", "Filter by label")
	f.StringVar(&flags.priority, "priority", "", "Filter by priority")
	f.StringVar(&flags.parent, "parent", "", "Filter by parent")
	f.StringVar(&flags.branch, "branch", "", "Filter by branch")
	f.BoolVar(&flags.roots, "roots", false, "Only return contracts with no parent")
	f.BoolVar(&flags.recursive, "recursive", false, "When parent is set, include all descendants")
	return cmd
}

func contractShowCmd() *cobra.Command {
	var format string
	var short bool
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show full contract content",
		Long: `Show full contract content (description, features, scenarios, sections,
coverage, and all metadata). Use --short for a summary-only view.`,
		Args: cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			var content string
			var err error
			if short {
				content, err = artifact.ShowContractShort(".", id)
			} else {
				content, err = artifact.ShowContractVerbose(".", id)
			}
			if err != nil {
				return fmt.Errorf("mos contract show: %w", err)
			}
			switch format {
			case names.FormatText:
				fmt.Print(content)
			case names.FormatJSON:
				data, err := json.Marshal(map[string]string{"id": id, "content": content})
				if err != nil {
					return fmt.Errorf("mos contract show: %w", err)
				}
				fmt.Println(string(data))
			default:
				return fmt.Errorf("mos contract show: unknown format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format (text|json)")
	cmd.Flags().BoolVarP(&short, "short", "s", false, "Show summary only")
	return cmd
}

func contractStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <id> <new-status>",
		Short: "Update contract status",
		Long:  "Update contract status.\nUse --ids for bulk: mos contract status --ids CON-1,CON-2 complete",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(c *cobra.Command, args []string) error {
			ids, _ := c.Flags().GetString("ids")
			if ids != "" {
				newStatus := args[0]
				for _, id := range strings.Split(ids, ",") {
					id = strings.TrimSpace(id)
					if id == "" {
						continue
					}
					if err := artifact.UpdateContractStatus(".", id, newStatus); err != nil {
						return fmt.Errorf("mos contract status %s: %w", id, err)
					}
					fmt.Printf("Updated %s status to %s\n", id, newStatus)
				}
				return nil
			}
			if len(args) < 2 {
				return fmt.Errorf("requires 2 args: <id> <new-status>")
			}
			id, newStatus := args[0], args[1]
			if err := artifact.UpdateContractStatus(".", id, newStatus); err != nil {
				return fmt.Errorf("mos contract status: %w", err)
			}
			fmt.Printf("Updated %s status to %s\n", id, newStatus)
			return nil
		},
	}
	cmd.Flags().String("ids", "", "Comma-separated contract IDs for bulk status update")
	return cmd
}

func contractUpdateCmd() *cobra.Command {
	var flags struct {
		title, goal, status, specFile, coverageFile, parent string
		dependsOn, specs, justifies, precondition           []string
		sprint, batch                                       []string
	}
	cmd := &cobra.Command{
		Use:   "update <id>...",
		Short: "Update contract fields",
		Long:  "Update one or more contracts with the given field values.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := artifact.ContractUpdateOpts{}
			if flags.title != "" {
				opts.Title = &flags.title
			}
			if flags.goal != "" {
				opts.Goal = &flags.goal
			}
			if flags.status != "" {
				opts.Status = &flags.status
			}
			if flags.specFile != "" {
				opts.SpecFile = &flags.specFile
			}
			if flags.coverageFile != "" {
				opts.CoverageFile = &flags.coverageFile
			}
			if len(flags.dependsOn) > 0 {
				deps := flags.dependsOn
				if len(deps) == 1 && deps[0] == "" {
					deps = []string{}
				}
				opts.DependsOn = &deps
			}
			if flags.parent != "" {
				opts.Parent = &flags.parent
			}
			if len(flags.specs) > 0 {
				specs := flags.specs
				if len(specs) == 1 && specs[0] == "" {
					specs = []string{}
				}
				opts.Specs = &specs
			}
			if len(flags.justifies) > 0 {
				opts.Justifies = &flags.justifies
			}
			if len(flags.precondition) > 0 {
				opts.Precondition = &flags.precondition
			}
			if len(flags.sprint) > 0 {
				opts.Sprint = &flags.sprint
			}
			if len(flags.batch) > 0 {
				opts.Batch = &flags.batch
			}
			for _, id := range args {
				if err := artifact.UpdateContract(".", id, opts); err != nil {
					return fmt.Errorf("mos contract update %s: %w", id, err)
				}
				if err := cliutil.ApplyOverflowFields(names.KindContract, id, map[string]string{}); err != nil {
					return err
				}
				fmt.Printf("Updated contract %s\n", id)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.title, "title", "", "Contract title")
	f.StringVar(&flags.goal, "goal", "", "Goal")
	f.StringVar(&flags.status, "status", "", "Status")
	f.StringVar(&flags.specFile, "spec-file", "", "Spec file path")
	f.StringVar(&flags.coverageFile, "coverage-file", "", "Coverage file path")
	f.StringSliceVar(&flags.dependsOn, "depends-on", nil, "Dependencies")
	f.StringVar(&flags.parent, "parent", "", "Parent ID")
	f.StringSliceVar(&flags.specs, "specs", nil, "Specification IDs")
	f.StringSliceVar(&flags.justifies, "justifies", nil, "Justifies")
	f.StringSliceVar(&flags.precondition, "precondition", nil, "Precondition IDs")
	f.StringSliceVar(&flags.sprint, "sprint", nil, "Sprint IDs")
	f.StringSliceVar(&flags.batch, "batch", nil, "Batch IDs")
	return cmd
}

func contractDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <id>...",
		Short: "Delete one or more contracts",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			for _, id := range args {
				if err := artifact.DeleteContract(".", id, force); err != nil {
					return fmt.Errorf("mos contract delete: %s: %w", id, err)
				}
				fmt.Printf("Deleted contract %s\n", id)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force delete")
	return cmd
}

func contractApplyCmd() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply artifact from file",
		Long:  "Apply a contract from a file (or stdin with -f -).",
		Args:  cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			return RunApply(names.KindContract, filePath, c.OutOrStdout())
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File path or - for stdin")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func contractEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit <id>",
		Short: "Edit contract in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			return RunEdit(names.KindContract, args[0])
		},
	}
}

// --- bespoke extras ---

func graphCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "graph [id]",
		Short: "Show contract dependency graph",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			var nodeID string
			if len(args) > 0 {
				nodeID = args[0]
			}
			var contracts []artifact.ContractInfo
			var err error
			if nodeID != "" {
				contracts, err = artifact.ContractGraphNode(".", nodeID)
			} else {
				contracts, err = artifact.ContractGraph(".")
			}
			if err != nil {
				return fmt.Errorf("mos contract graph: %w", err)
			}
			if len(contracts) == 0 {
				fmt.Println("(no contracts found)")
				return nil
			}
			dependents := make(map[string][]string)
			for _, ct := range contracts {
				for _, dep := range ct.DependsOn {
					dependents[dep] = append(dependents[dep], ct.ID)
				}
			}
			switch format {
			case names.FormatText:
				for _, ct := range contracts {
					marker := " "
					if ct.Status == names.StatusActive || ct.Status == names.StatusDraft {
						marker = "*"
					}
					fmt.Printf("%s %-20s [%-10s] %s\n", marker, ct.ID, ct.Status, ct.Title)
					if len(ct.DependsOn) > 0 {
						fmt.Printf("    depends on: %s\n", strings.Join(ct.DependsOn, ", "))
					}
					if deps, ok := dependents[ct.ID]; ok {
						fmt.Printf("    depended on by: %s\n", strings.Join(deps, ", "))
					}
				}
			case "mermaid":
				fmt.Println("graph TD")
				for _, ct := range contracts {
					label := ct.ID + ": " + ct.Title
					if ct.Status == names.StatusComplete || ct.Status == names.StatusAbandoned {
						fmt.Printf("    %s[\"%s\"]\n", sanitizeMermaidID(ct.ID), label)
					} else {
						fmt.Printf("    %s([\"%s\"])\n", sanitizeMermaidID(ct.ID), label)
					}
				}
				for _, ct := range contracts {
					for _, dep := range ct.DependsOn {
						fmt.Printf("    %s --> %s\n", sanitizeMermaidID(dep), sanitizeMermaidID(ct.ID))
					}
				}
			case "dot":
				fmt.Println("digraph contracts {")
				fmt.Println("    rankdir=LR;")
				for _, ct := range contracts {
					shape := "box"
					if ct.Status == names.StatusComplete || ct.Status == names.StatusAbandoned {
						shape = "ellipse"
					}
					fmt.Printf("    %q [label=%q, shape=%s];\n", ct.ID, ct.ID+": "+ct.Title, shape)
				}
				for _, ct := range contracts {
					for _, dep := range ct.DependsOn {
						fmt.Printf("    %q -> %q;\n", dep, ct.ID)
					}
				}
				fmt.Println("}")
			default:
				return fmt.Errorf("mos contract graph: unknown format %q (use text, mermaid, or dot)", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format (text|mermaid|dot)")
	return cmd
}

func sanitizeMermaidID(id string) string {
	return strings.ReplaceAll(id, "-", "_")
}

func linkCmd() *cobra.Command {
	var dependsOn string
	cmd := &cobra.Command{
		Use:   "link <id>",
		Short: "Link contract to a dependency",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if dependsOn == "" {
				return fmt.Errorf("usage: mos contract link <id> --depends-on <dependency-id>")
			}
			if err := artifact.LinkContract(".", id, dependsOn); err != nil {
				return fmt.Errorf("mos contract link: %w", err)
			}
			fmt.Printf("Linked: %s depends on %s\n", id, dependsOn)
			return nil
		},
	}
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "Dependency ID")
	_ = cmd.MarkFlagRequired("depends-on")
	return cmd
}

func unlinkCmd() *cobra.Command {
	var dependsOn string
	cmd := &cobra.Command{
		Use:   "unlink <id>",
		Short: "Unlink contract from a dependency",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if dependsOn == "" {
				return fmt.Errorf("usage: mos contract unlink <id> --depends-on <dependency-id>")
			}
			if err := artifact.UnlinkContract(".", id, dependsOn); err != nil {
				return fmt.Errorf("mos contract unlink: %w", err)
			}
			fmt.Printf("Unlinked: %s no longer depends on %s\n", id, dependsOn)
			return nil
		},
	}
	cmd.Flags().StringVar(&dependsOn, "depends-on", "", "Dependency ID to remove")
	_ = cmd.MarkFlagRequired("depends-on")
	return cmd
}

func lockCmd() *cobra.Command {
	var message string
	cmd := &cobra.Command{
		Use:   "lock <id>",
		Short: "Lock (seal) a contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if err := artifact.SealContract(".", id, "lock", message); err != nil {
				return fmt.Errorf("mos contract lock: %w", err)
			}
			fmt.Printf("Locked contract %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVarP(&message, "message", "m", "", "Lock message")
	return cmd
}

func unlockCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "unlock <id>",
		Short: "Unlock a contract",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if err := artifact.UnsealContract(".", id, force); err != nil {
				return fmt.Errorf("mos contract unlock: %w", err)
			}
			fmt.Printf("Unlocked contract %s\n", id)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Force unlock")
	return cmd
}

func renameCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <old-id> <new-id>",
		Short: "Rename a contract",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			oldID, newID := args[0], args[1]
			if err := artifact.RenameContract(".", oldID, newID); err != nil {
				return fmt.Errorf("mos contract rename: %w", err)
			}
			fmt.Printf("Renamed contract %s -> %s\n", oldID, newID)
			return nil
		},
	}
}

func scenarioCmd() *cobra.Command {
	var filter string
	var allDone bool
	cmd := &cobra.Command{
		Use:   "scenario <id> [done|pending \"Scenario Name\"]",
		Short: "List or update scenario status",
		Long:  "List scenarios, mark one as done/pending, or mark all as done with --all-done.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			positional := args[1:]
			if allDone {
				count, err := artifact.SetAllScenariosStatus(".", id, "done")
				if err != nil {
					return fmt.Errorf("mos contract scenario: %w", err)
				}
				fmt.Printf("Marked %d scenarios as done in %s\n", count, id)
				return nil
			}
			if len(positional) >= 3 {
				action := positional[0]
				name := strings.Join(positional[1:], " ")
				switch action {
				case "done":
					if err := artifact.SetScenarioStatus(".", id, name, "done"); err != nil {
						return fmt.Errorf("mos contract scenario: %w", err)
					}
					fmt.Printf("Marked scenario %q as done in %s\n", name, id)
				case "pending":
					if err := artifact.SetScenarioStatus(".", id, name, "pending"); err != nil {
						return fmt.Errorf("mos contract scenario: %w", err)
					}
					fmt.Printf("Marked scenario %q as pending in %s\n", name, id)
				default:
					return fmt.Errorf("mos contract scenario: unknown action %q (expected done or pending)", action)
				}
				return nil
			}
			scenarios, err := artifact.ListScenarios(".", id)
			if err != nil {
				return fmt.Errorf("mos contract scenario: %w", err)
			}
			for _, s := range scenarios {
				if filter != "" && s.Status != filter {
					continue
				}
				fmt.Printf("  [%s] %s\n", s.Status, s.Name)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Filter by status (done|pending)")
	cmd.Flags().BoolVar(&allDone, "all-done", false, "Mark all scenarios as done")
	return cmd
}

func contextCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "context <id>",
		Short: "Show contract context",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			ctx, err := artifact.ContractContext(".", id)
			if err != nil {
				return fmt.Errorf("mos contract context: %w", err)
			}
			fmt.Print(artifact.FormatContext(ctx))
			return nil
		},
	}
}

func chainCommand() *cobra.Command {
	var negative bool
	var format string
	cmd := &cobra.Command{
		Use:   "chain <id>",
		Short: "Show justification chain for an artifact",
		Long:  "Walk and display the justification chain. Also available as top-level 'mos chain'.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			reg, err := registry.LoadRegistry(".")
			if err != nil {
				return fmt.Errorf("mos chain: %w", err)
			}
			kind := InferArtifactKind(reg, id)
			ch, err := chainpkg.WalkChain(".", kind, id)
			if err != nil {
				return fmt.Errorf("mos chain: %w", err)
			}
			if format == names.FormatJSON {
				result := map[string]any{"chain": ch}
				if negative {
					nc, err := chainpkg.WalkNegativeChain(".", kind, id)
					if err != nil {
						return fmt.Errorf("mos chain --negative: %w", err)
					}
					result["negative"] = nc
				}
				data, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					return fmt.Errorf("mos chain: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}
			fmt.Print(chainpkg.FormatChain(ch))
			if negative {
				nc, err := chainpkg.WalkNegativeChain(".", kind, id)
				if err != nil {
					return fmt.Errorf("mos chain --negative: %w", err)
				}
				fmt.Println()
				fmt.Print(chainpkg.FormatNegativeChain(nc))
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&negative, "negative", false, "Include negative chain")
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format (text|json)")
	return cmd
}

func verifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "verify <id>",
		Short: "Verify contract scenarios",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			results, err := artifact.VerifyContract(".", id)
			if err != nil {
				return fmt.Errorf("mos contract verify: %w", err)
			}
			if len(results) == 0 {
				fmt.Println("No scenarios in 'implemented' state to verify.")
				return nil
			}
			hasFail := false
			for _, r := range results {
				status := "PASS"
				if !r.Pass {
					status = "FAIL"
					hasFail = true
				}
				fmt.Printf("  [%s] %s", status, r.Scenario)
				if r.RuleID != "" {
					fmt.Printf(" (rule: %s)", r.RuleID)
				}
				fmt.Println()
			}
			if hasFail {
				return fmt.Errorf("verification failed")
			}
			return nil
		},
	}
}

func historyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "history <id>",
		Short: "Show contract history",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			ledgerPath, err := artifact.LedgerPathForContract(".", id)
			if err != nil {
				return fmt.Errorf("mos contract history: %w", err)
			}
			entries, err := artifact.ReadLedger(ledgerPath)
			if err != nil {
				return fmt.Errorf("mos contract history: %w", err)
			}
			fmt.Print(artifact.FormatHistory(entries))
			return nil
		},
	}
}
