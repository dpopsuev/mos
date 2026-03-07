package govern

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
)

var ShowCmd = &cobra.Command{
	Use:   "show <ID>",
	Short: "Show any artifact (infers kind from prefix)",
	Long:  "Displays any artifact. The kind is inferred from the ID prefix.\nUse --kind to disambiguate slugs or unknown prefixes.",
	Args:  cobra.ExactArgs(1),
	RunE:  runVerbShow,
}

var showKindFlag string
var showFormatFlag string

func init() {
	ShowCmd.Flags().StringVar(&showKindFlag, "kind", "", "Artifact kind override")
	ShowCmd.Flags().StringVar(&showFormatFlag, "format", "text", "Output format: text or json")
}

func runVerbShow(cmd *cobra.Command, args []string) error {
	id := args[0]
	kind := showKindFlag

	if kind == "" {
		reg, err := registry.LoadRegistry(".")
		if err != nil {
			return fmt.Errorf("loading registry: %w", err)
		}
		resolved, err := reg.ResolveKindFromID(id)
		if err != nil {
			return fmt.Errorf("%v — use --kind to specify", err)
		}
		kind = resolved
	}

	if showFormatFlag == names.FormatJSON {
		return showArtifactJSON(id, kind)
	}

	switch kind {
	case names.KindContract:
		content, err := artifact.ShowContractVerbose(".", id)
		if err != nil {
			return err
		}
		fmt.Print(content)
	case names.KindSpecification:
		content, err := artifact.ShowSpec(".", id)
		if err != nil {
			return err
		}
		fmt.Print(content)
	case names.KindBinder:
		content, err := artifact.ShowBinder(".", id)
		if err != nil {
			return err
		}
		fmt.Print(content)
	default:
		reg, err := registry.LoadRegistry(".")
		if err != nil {
			return err
		}
		td, ok := reg.Types[kind]
		if !ok {
			return fmt.Errorf("no show handler for kind %q", kind)
		}
		content, err := artifact.GenericShow(".", td, id)
		if err != nil {
			return err
		}
		fmt.Print(content)
	}
	return nil
}

func showArtifactJSON(id, kind string) error {
	reg, err := registry.LoadRegistry(".")
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}
	td, ok := reg.Types[kind]
	if !ok {
		return fmt.Errorf("no type definition for kind %q", kind)
	}

	var path string
	switch kind {
	case names.KindContract:
		path, err = artifact.FindContractPath(".", id)
	case names.KindSpecification:
		path, err = artifact.FindSpecPath(".", id)
	default:
		path, err = artifact.FindGenericPath(".", td, id)
	}
	if err != nil {
		return err
	}

	ab, err := dsl.ReadArtifact(path)
	if err != nil {
		return err
	}
	m := dsl.ToMap(ab)
	m["_path"] = path
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

var GetCmd = &cobra.Command{
	Use:   "get <ID> <path>",
	Short: "Read a field via dot-path",
	Long:  "Reads a field value at the given dot-path.\nExamples:\n  mos get CON-2026-125 title\n  mos get CON-2026-125 coverage.unit.applies",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		val, err := artifact.GetArtifactField(".", args[0], args[1])
		if err != nil {
			return err
		}
		fmt.Println(val)
		return nil
	},
}

var SetCmd = &cobra.Command{
	Use:   "set <ID> <path> <value>",
	Short: "Write a field via dot-path",
	Long: `Writes a field value at the given dot-path.
Creates intermediate blocks if they don't exist.

Bulk modes:
  --ids CON-1,CON-2 status draft       Same value for multiple artifacts
  --map ID1=VAL1,ID2=VAL2 field        Per-artifact values
  --dry-run                             Preview without writing`,
	Args: cobra.RangeArgs(1, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, _ := cmd.Flags().GetString("ids")
		mapFlag, _ := cmd.Flags().GetString("map")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if mapFlag != "" {
			if len(args) < 1 {
				return fmt.Errorf("requires at least 1 arg: <field-path>")
			}
			fieldPath := args[0]
			for _, pair := range strings.Split(mapFlag, ",") {
				pair = strings.TrimSpace(pair)
				if pair == "" {
					continue
				}
				eqIdx := strings.Index(pair, "=")
				if eqIdx < 0 {
					return fmt.Errorf("invalid --map entry %q: expected ID=VALUE", pair)
				}
				id := strings.TrimSpace(pair[:eqIdx])
				value := strings.TrimSpace(pair[eqIdx+1:])
				if id == "" || value == "" {
					return fmt.Errorf("invalid --map entry %q: empty ID or value", pair)
				}
				if dryRun {
					fmt.Printf("[dry-run] would set %s %s = %s\n", id, fieldPath, value)
					continue
				}
				if err := artifact.SetArtifactField(".", id, fieldPath, value); err != nil {
					return fmt.Errorf("%s: %w", id, err)
				}
				fmt.Printf("Updated %s %s = %s\n", id, fieldPath, value)
			}
			return nil
		}

		if ids != "" {
			if len(args) < 2 {
				return fmt.Errorf("requires 2 args with --ids: <field-path> <value>")
			}
			fieldPath, value := args[0], args[1]
			for _, id := range strings.Split(ids, ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				if dryRun {
					fmt.Printf("[dry-run] would set %s %s = %s\n", id, fieldPath, value)
					continue
				}
				if err := artifact.SetArtifactField(".", id, fieldPath, value); err != nil {
					return fmt.Errorf("%s: %w", id, err)
				}
				fmt.Printf("Updated %s %s = %s\n", id, fieldPath, value)
			}
			return nil
		}

		if len(args) < 3 {
			return fmt.Errorf("requires 3 args: <ID> <path> <value>")
		}
		if dryRun {
			fmt.Printf("[dry-run] would set %s %s = %s\n", args[0], args[1], args[2])
			return nil
		}
		if err := artifact.SetArtifactField(".", args[0], args[1], args[2]); err != nil {
			return err
		}
		fmt.Printf("Updated %s %s = %s\n", args[0], args[1], args[2])
		return nil
	},
}

var AppendCmd = &cobra.Command{
	Use:   "append <ID> <path> <value>",
	Short: "Append to a list field via dot-path",
	Long:  "Appends a value to a list field.\nUse --ids to apply to multiple artifacts: mos append --ids CON-1,CON-2 labels review",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids, _ := cmd.Flags().GetString("ids")
		if ids != "" {
			fieldPath, value := args[0], args[1]
			for _, id := range strings.Split(ids, ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				if err := artifact.AppendArtifactField(".", id, fieldPath, value); err != nil {
					return fmt.Errorf("%s: %w", id, err)
				}
				fmt.Printf("Appended %q to %s %s\n", value, id, fieldPath)
			}
			return nil
		}
		if len(args) < 3 {
			return fmt.Errorf("requires 3 args: <ID> <path> <value>")
		}
		if err := artifact.AppendArtifactField(".", args[0], args[1], args[2]); err != nil {
			return err
		}
		fmt.Printf("Appended %q to %s %s\n", args[2], args[0], args[1])
		return nil
	},
}

func init() {
	SetCmd.Flags().String("ids", "", "Comma-separated artifact IDs for same-value bulk operation")
	SetCmd.Flags().String("map", "", "Per-artifact values: ID1=VAL1,ID2=VAL2")
	SetCmd.Flags().Bool("dry-run", false, "Preview changes without writing to disk")
	AppendCmd.Flags().String("ids", "", "Comma-separated artifact IDs for bulk operation")
}

var VerbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new artifact (--title required)",
	Long:  "Creates an artifact. Defaults to contract unless --kind specifies otherwise.",
	RunE:  runVerbCreate,
}

var verbCreateFlags struct {
	kind, title, status, goal, project string
	justifies, sprint, dependsOn       string
	fromFile, fromTemplate             string
}

func init() {
	f := VerbCreateCmd.Flags()
	f.StringVar(&verbCreateFlags.kind, "kind", "", "Artifact kind")
	f.StringVar(&verbCreateFlags.title, "title", "", "Artifact title")
	f.StringVar(&verbCreateFlags.status, "status", "", "Initial status")
	f.StringVar(&verbCreateFlags.goal, "goal", "", "Contract goal")
	f.StringVar(&verbCreateFlags.project, "project", "", "Project for auto-ID")
	f.StringVar(&verbCreateFlags.justifies, "justifies", "", "Comma-separated need IDs")
	f.StringVar(&verbCreateFlags.sprint, "sprint", "", "Comma-separated sprint IDs")
	f.StringVar(&verbCreateFlags.dependsOn, "depends-on", "", "Comma-separated dependency IDs")
	f.StringVar(&verbCreateFlags.fromFile, "from-file", "", "JSON file with array of artifact definitions for bulk import")
	f.StringVar(&verbCreateFlags.fromTemplate, "from-template", "", "Template name from .mos/templates/")
	f.SetInterspersed(false)
}

func runVerbCreate(cmd *cobra.Command, args []string) error {
	if verbCreateFlags.fromFile != "" {
		return runBulkCreate(verbCreateFlags.fromFile)
	}

	kind := verbCreateFlags.kind

	isContractKind := kind == "" || kind == names.KindContract || kind == "bug" || kind == "feature" || kind == "task"
	if isContractKind {
		opts := artifact.ContractOpts{
			Title:    verbCreateFlags.title,
			Status:   verbCreateFlags.status,
			Kind:     kind,
			Goal:     verbCreateFlags.goal,
			Project:  verbCreateFlags.project,
			Template: verbCreateFlags.fromTemplate,
		}
		if verbCreateFlags.justifies != "" {
			opts.Justifies = strings.Split(verbCreateFlags.justifies, ",")
		}
		if verbCreateFlags.sprint != "" {
			opts.Sprint = strings.Split(verbCreateFlags.sprint, ",")
		}
		if verbCreateFlags.dependsOn != "" {
			opts.DependsOn = strings.Split(verbCreateFlags.dependsOn, ",")
		}

		overflow, positional := cliutil.ParseKVArgs(args)
		var id string
		if len(positional) > 0 {
			id = positional[0]
		}

		if opts.Title == "" {
			return fmt.Errorf("--title is required")
		}

		createdID, err := artifact.CreateContract(".", id, opts)
		if err != nil {
			return err
		}
		if len(overflow) > 0 {
			if err := cliutil.ApplyOverflowFields(names.KindContract, createdID, overflow); err != nil {
				return err
			}
		}
		fmt.Println(createdID)
		return nil
	}

	reg, err := registry.LoadRegistry(".")
	if err != nil {
		return err
	}
	td, ok := reg.Types[kind]
	if !ok {
		return fmt.Errorf("kind %q not found in registry", kind)
	}

	fields, positional := cliutil.ParseKVArgs(args)
	if verbCreateFlags.title != "" {
		fields["title"] = verbCreateFlags.title
	}
	if verbCreateFlags.status != "" {
		fields["status"] = verbCreateFlags.status
	}
	if verbCreateFlags.goal != "" {
		fields["goal"] = verbCreateFlags.goal
	}
	if verbCreateFlags.sprint != "" {
		fields["sprint"] = verbCreateFlags.sprint
	}
	var gid string
	if len(positional) > 0 {
		gid = positional[0]
	}
	if fields["title"] == "" {
		return fmt.Errorf("--title is required")
	}
	if gid == "" && td.Prefix != "" {
		autoID, idErr := registry.NextIDForType(".", td.Prefix, td.Directory)
		if idErr != nil {
			return fmt.Errorf("auto-generating ID for %s: %w", kind, idErr)
		}
		gid = autoID
	}
	if gid == "" {
		return fmt.Errorf("--kind %s requires an explicit ID (no prefix configured for auto-generation)", kind)
	}
	createdPath, err := artifact.GenericCreateWithTemplate(".", td, gid, fields, verbCreateFlags.fromTemplate)
	if err != nil {
		return err
	}
	fmt.Println(createdPath)
	return nil
}

type bulkArtifact struct {
	Kind      string            `json:"kind"`
	Title     string            `json:"title"`
	Status    string            `json:"status"`
	Goal      string            `json:"goal"`
	Justifies string            `json:"justifies"`
	Sprint    string            `json:"sprint"`
	DependsOn string            `json:"depends_on"`
	Fields    map[string]string `json:"fields"`
}

func runBulkCreate(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	var artifacts []bulkArtifact
	if err := json.Unmarshal(data, &artifacts); err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	for i, a := range artifacts {
		kind := a.Kind
		if kind == "" {
			kind = names.KindContract
		}

		isContract := kind == names.KindContract || kind == "bug" || kind == "feature" || kind == "task"
		if isContract {
			opts := artifact.ContractOpts{
				Title:  a.Title,
				Status: a.Status,
				Kind:   kind,
				Goal:   a.Goal,
			}
			if a.Justifies != "" {
				opts.Justifies = strings.Split(a.Justifies, ",")
			}
			if a.Sprint != "" {
				opts.Sprint = strings.Split(a.Sprint, ",")
			}
			if a.DependsOn != "" {
				opts.DependsOn = strings.Split(a.DependsOn, ",")
			}
			if opts.Title == "" {
				return fmt.Errorf("artifact %d: title is required", i)
			}
			id, err := artifact.CreateContract(".", "", opts)
			if err != nil {
				return fmt.Errorf("artifact %d: %w", i, err)
			}
			for k, v := range a.Fields {
				if err := artifact.SetArtifactField(".", id, k, v); err != nil {
					return fmt.Errorf("artifact %d: setting %s: %w", i, k, err)
				}
			}
			fmt.Println(id)
		} else {
			reg, err := registry.LoadRegistry(".")
			if err != nil {
				return fmt.Errorf("artifact %d: %w", i, err)
			}
			td, ok := reg.Types[kind]
			if !ok {
				return fmt.Errorf("artifact %d: kind %q not found", i, kind)
			}
			fields := map[string]string{"title": a.Title}
			if a.Status != "" {
				fields["status"] = a.Status
			}
			for k, v := range a.Fields {
				fields[k] = v
			}
			id, err := artifact.GenericCreate(".", td, "", fields)
			if err != nil {
				return fmt.Errorf("artifact %d: %w", i, err)
			}
			fmt.Println(id)
		}
	}
	fmt.Printf("Created %d artifact(s)\n", len(artifacts))
	return nil
}
