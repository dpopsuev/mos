package binder

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

var binderTD = registry.ArtifactTypeDef{Kind: names.KindBinder, Directory: names.DirBinders}

// Cmd is the binder command, built via the artifact factory.
var Cmd = factory.Register(factory.KindConfig{
	TD:     binderTD,
	Create: binderCreateCmd(),
	List:   binderListCmd(),
	Show:   binderShowCmd(),
	Extra:  []*cobra.Command{bindCmd(), unbindCmd(), traceCmd()},
})

func binderCreateCmd() *cobra.Command {
	var title, project string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new binder",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			opts := artifact.BinderOpts{
				Title:   title,
				Project: project,
			}
			var id string
			for _, arg := range args {
				if !strings.HasPrefix(arg, "-") {
					id = arg
					break
				}
			}
			path, err := artifact.CreateBinder(rootDir, id, opts)
			if err != nil {
				return fmt.Errorf("mos binder create: %w", err)
			}
			fmt.Printf("Created binder: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Binder title (required)")
	cmd.Flags().StringVar(&project, "project", "", "Owning project")
	return cmd
}

func binderListCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all binders",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			binders, err := artifact.ListBinders(rootDir)
			if err != nil {
				return fmt.Errorf("mos binder list: %w", err)
			}
			switch format {
			case names.FormatJSON:
				data, err := json.MarshalIndent(binders, "", "  ")
				if err != nil {
					return fmt.Errorf("mos binder list: %w", err)
				}
				fmt.Println(string(data))
			case names.FormatText:
				if len(binders) == 0 {
					fmt.Println("(no binders found)")
					return nil
				}
				for _, b := range binders {
					fmt.Printf("  %-16s %s (%d specs)\n", b.ID, b.Title, len(b.Specs))
				}
			default:
				return fmt.Errorf("mos binder list: unknown format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", names.FormatText, "Output format: text, json")
	return cmd
}

func binderShowCmd() *cobra.Command {
	var verbose bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show binder details",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos binder show <id> [--verbose]")
			}
			id := args[0]
			if verbose {
				reg, loadErr := registry.LoadRegistry(rootDir)
				if loadErr != nil {
					return fmt.Errorf("mos binder show: %w", loadErr)
				}
				td := reg.Types[names.KindBinder]
				content, err := artifact.GenericShow(rootDir, td, id)
				if err != nil {
					return fmt.Errorf("mos binder show: %w", err)
				}
				fmt.Print(content)
			} else {
				content, err := artifact.ShowBinder(rootDir, id)
				if err != nil {
					return fmt.Errorf("mos binder show: %w", err)
				}
				fmt.Print(content)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full artifact content")
	return cmd
}

func bindCmd() *cobra.Command {
	var specID string
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "Bind a specification to a binder",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos binder bind <id> --spec <spec-id>")
			}
			id := args[0]
			if specID == "" {
				return fmt.Errorf("usage: mos binder bind <id> --spec <spec-id>")
			}
			if err := artifact.BinderBind(rootDir, id, specID); err != nil {
				return fmt.Errorf("mos binder bind: %w", err)
			}
			fmt.Printf("Bound %s to binder %s\n", specID, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "", "Specification ID to bind (required)")
	return cmd
}

func unbindCmd() *cobra.Command {
	var specID string
	cmd := &cobra.Command{
		Use:   "unbind",
		Short: "Unbind a specification from a binder",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos binder unbind <id> --spec <spec-id>")
			}
			id := args[0]
			if specID == "" {
				return fmt.Errorf("usage: mos binder unbind <id> --spec <spec-id>")
			}
			if err := artifact.BinderUnbind(rootDir, id, specID); err != nil {
				return fmt.Errorf("mos binder unbind: %w", err)
			}
			fmt.Printf("Unbound %s from binder %s\n", specID, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&specID, "spec", "", "Specification ID to unbind (required)")
	return cmd
}

func traceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "trace",
		Short: "Show traceability report for a binder",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos binder trace <id>")
			}
			id := args[0]
			report, err := artifact.BinderTrace(rootDir, id)
			if err != nil {
				return fmt.Errorf("mos binder trace: %w", err)
			}
			fmt.Print(report)
			return nil
		},
	}
}
