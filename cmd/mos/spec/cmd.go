package spec

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/dpopsuev/mos/cmd/mos/cliutil"
	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/mesh"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

var specTD = registry.ArtifactTypeDef{Kind: names.KindSpecification, Directory: names.DirSpecifications}

// Cmd is the spec command, built via the artifact factory.
var Cmd = factory.Register(factory.KindConfig{
	TD:     specTD,
	Use:    "spec",
	Create: specCreateCmd(),
	List:   specListCmd(),
	Show:   specShowCmd(),
	Update: specUpdateCmd(),
	Blocks: factory.BlockOps{
		AddSection:  true,
		RemoveBlock: true,
		AddSpec:     true,
		SetSection:  true,
		SetField:    true,
	},
})

func specCreateCmd() *cobra.Command {
	var flags struct {
		title, enforcement, symbol, harness, project, satisfies, addresses string
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new specification",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			opts := artifact.SpecOpts{
				Title:       flags.title,
				Enforcement: flags.enforcement,
				Symbol:      flags.symbol,
				Harness:     flags.harness,
				Project:     flags.project,
				Satisfies:   flags.satisfies,
				Addresses:   flags.addresses,
			}
			overflow, positional := cliutil.ParseKVArgs(args)
			var id string
			if len(positional) > 0 {
				id = positional[0]
			}
			path, err := artifact.CreateSpec(rootDir, id, opts)
			if err != nil {
				return fmt.Errorf("mos spec create: %w", err)
			}
			if id == "" {
				id = filepath.Base(filepath.Dir(path))
			}
			if err := cliutil.ApplyOverflowFields(names.KindSpecification, id, overflow); err != nil {
				return err
			}
			fmt.Printf("Created specification: %s\n", path)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.title, "title", "", "Specification title")
	f.StringVar(&flags.enforcement, "enforcement", "", "Enforcement level")
	f.StringVar(&flags.symbol, "symbol", "", "Implementation binding")
	f.StringVar(&flags.harness, "harness", "", "Test binding")
	f.StringVar(&flags.project, "project", "", "Project for auto-ID")
	f.StringVar(&flags.satisfies, "satisfies", "", "Need ID this specification satisfies")
	f.StringVar(&flags.addresses, "addresses", "", "Requirement addressed")
	f.SetInterspersed(false)
	return cmd
}

func specListCmd() *cobra.Command {
	var format, enforcement string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all specifications",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			specs, err := artifact.ListSpecs(rootDir, enforcement)
			if err != nil {
				return fmt.Errorf("mos spec list: %w", err)
			}
			switch format {
			case names.FormatJSON:
				data, err := json.MarshalIndent(specs, "", "  ")
				if err != nil {
					return fmt.Errorf("mos spec list: %w", err)
				}
				fmt.Println(string(data))
			case names.FormatText:
				if len(specs) == 0 {
					fmt.Println("(no specifications found)")
					return nil
				}
				for _, s := range specs {
					fmt.Printf("  %-16s %-10s %-10s %s\n", s.ID, s.Status, s.Enforcement, s.Title)
				}
			default:
				return fmt.Errorf("mos spec list: unknown format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	cmd.Flags().StringVar(&enforcement, "enforcement", "", "Filter by enforcement level")
	return cmd
}

func specShowCmd() *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show specification content",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos spec show <id> [--short]")
			}
			id := args[0]
			if short {
				content, err := artifact.ShowSpec(rootDir, id)
				if err != nil {
					return fmt.Errorf("mos spec show: %w", err)
				}
				fmt.Print(content)
			} else {
				reg, loadErr := registry.LoadRegistry(rootDir)
				if loadErr != nil {
					return fmt.Errorf("mos spec show: %w", loadErr)
				}
				td := reg.Types[names.KindSpecification]
				content, err := artifact.GenericShow(rootDir, td, id)
				if err != nil {
					return fmt.Errorf("mos spec show: %w", err)
				}
				fmt.Print(content)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&short, "short", "s", false, "Show summary only")
	return cmd
}

var syncApply bool
var syncFormat string

// SpecSyncCmd infers spec include directives from git commit-contract tracing.
var SpecSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Infer spec include directives from git commit history",
	Long:  "Analyzes git commits for contract ID references and maps changed packages to specs.\nDefault is dry-run mode; use --apply to update spec includes.",
	RunE: func(cmd *cobra.Command, args []string) error {
		updates, err := mesh.InferSpecIncludes(rootDir)
		if err != nil {
			return fmt.Errorf("mos spec sync: %w", err)
		}

		if len(updates) == 0 {
			if syncFormat == "json" {
				fmt.Println("[]")
			} else {
				fmt.Println("No include updates inferred.")
			}
			return nil
		}

		if syncFormat == "json" {
			data, _ := json.MarshalIndent(updates, "", "  ")
			fmt.Println(string(data))
			if !syncApply {
				return nil
			}
		}

		if !syncApply {
			for _, u := range updates {
				fmt.Printf("%s:\n", u.SpecID)
				fmt.Printf("  current:  %v\n", u.CurrentIncludes)
				fmt.Printf("  inferred: %v\n", u.InferredIncludes)
				fmt.Printf("  added:    %v\n", u.Added)
			}
			fmt.Println("\nDry run — use --apply to update spec includes.")
			return nil
		}

		for _, u := range updates {
			if err := artifact.GenericAddSpec(rootDir, specTD, u.SpecID, u.Added); err != nil {
				fmt.Printf("  warning: could not update %s: %v\n", u.SpecID, err)
				continue
			}
			fmt.Printf("Updated %s: added %v\n", u.SpecID, u.Added)
		}
		return nil
	},
}

func init() {
	SpecSyncCmd.Flags().BoolVar(&syncApply, "apply", false, "Apply inferred changes (default: dry-run)")
	SpecSyncCmd.Flags().StringVar(&syncFormat, "format", "text", "Output format: text or json")
	Cmd.AddCommand(SpecSyncCmd)
}

func specUpdateCmd() *cobra.Command {
	var flags struct {
		title, enforcement, symbol, harness, satisfies, addresses string
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update specifications",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			var opts artifact.SpecUpdateOpts
			if c.Flags().Changed("title") {
				opts.Title = &flags.title
			}
			if c.Flags().Changed("enforcement") {
				opts.Enforcement = &flags.enforcement
			}
			if c.Flags().Changed("symbol") {
				opts.Symbol = &flags.symbol
			}
			if c.Flags().Changed("harness") {
				opts.Harness = &flags.harness
			}
			if c.Flags().Changed("satisfies") {
				opts.Satisfies = &flags.satisfies
			}
			if c.Flags().Changed("addresses") {
				opts.Addresses = &flags.addresses
			}
			overflow, ids := cliutil.ParseKVArgs(args)
			if len(ids) == 0 {
				return fmt.Errorf("usage: mos spec update <id>... [--title ...] [--enforcement ...] ...")
			}
			for _, id := range ids {
				if err := artifact.UpdateSpec(rootDir, id, opts); err != nil {
					return fmt.Errorf("mos spec update: %s: %w", id, err)
				}
				if err := cliutil.ApplyOverflowFields(names.KindSpecification, id, overflow); err != nil {
					return err
				}
				fmt.Printf("Updated specification %s\n", id)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.title, "title", "", "Specification title")
	f.StringVar(&flags.enforcement, "enforcement", "", "Enforcement level")
	f.StringVar(&flags.symbol, "symbol", "", "Implementation binding")
	f.StringVar(&flags.harness, "harness", "", "Test binding")
	f.StringVar(&flags.satisfies, "satisfies", "", "Need ID this specification satisfies")
	f.StringVar(&flags.addresses, "addresses", "", "Requirement addressed")
	f.SetInterspersed(false)
	return cmd
}
