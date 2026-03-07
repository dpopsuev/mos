package generic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

// NewCmd creates a Cobra command for the given artifact type with all sub-commands.
func NewCmd(td registry.ArtifactTypeDef) *cobra.Command {
	cfg := factory.KindConfig{
		TD:     td,
		Blocks: factory.AllBlocks,
	}
	if td.Kind == names.KindNeed {
		cfg.Extra = append(cfg.Extra, criteriaCmd(td))
	}
	if td.Kind == names.KindSprint {
		cfg.Extra = append(cfg.Extra, addContractCmd(), removeContractCmd(), closeCmd(), planCmd())
	}

	return factory.Register(cfg)
}

// Re-export block command builders so existing consumers (spec, binder, contract)
// can use them without importing factory directly.

func AddSectionCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddSectionCmd(td)
}

func AddFeatureCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddFeatureCmd(td)
}

func AddScenarioCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddScenarioCmd(td)
}

func AddCriterionCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddCriterionCmd(td)
}

func RemoveBlockCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.RemoveBlockCmd(td)
}

func AddCoverageCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddCoverageCmd(td)
}

func AddBillCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddBillCmd(td)
}

func AddSpecCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return factory.AddSpecCmd(td)
}

// ApplyCmd returns an apply command for a given kind (used by spec, binder, lexicon).
func ApplyCmd(kind string) *cobra.Command {
	return factory.ApplyCmd(kind)
}

// EditCmd returns an edit command for a given kind (used by spec, binder, lexicon).
func EditCmd(kind string) *cobra.Command {
	return factory.EditCmd(kind)
}

// --- bespoke commands that stay in generic ---

func addContractCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-contract <sprint-id> <contract-id> [contract-id ...]",
		Short: "Add contracts to a sprint",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			sprintID := args[0]
			contractIDs := args[1:]
			if err := artifact.SprintAddContracts(rootDir, sprintID, contractIDs); err != nil {
				return fmt.Errorf("mos sprint add-contract: %w", err)
			}
			fmt.Printf("Added %d contract(s) to %s\n", len(contractIDs), sprintID)
			return nil
		},
	}
}

func removeContractCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-contract <sprint-id> <contract-id> [contract-id ...]",
		Short: "Remove contracts from a sprint",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			sprintID := args[0]
			contractIDs := args[1:]
			if err := artifact.SprintRemoveContracts(rootDir, sprintID, contractIDs); err != nil {
				return fmt.Errorf("mos sprint remove-contract: %w", err)
			}
			fmt.Printf("Removed %d contract(s) from %s\n", len(contractIDs), sprintID)
			return nil
		},
	}
}

func closeCmd() *cobra.Command {
	var (
		dryRun      bool
		closeFormat string
	)
	cmd := &cobra.Command{
		Use:   "close <sprint-id>",
		Short: "Mark all contracts complete and close the sprint",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			sprintID := args[0]
			if dryRun {
				result, err := artifact.SprintCloseDryRun(rootDir, sprintID)
				if err != nil {
					return err
				}
				if closeFormat == "json" {
					data, _ := json.MarshalIndent(result, "", "  ")
					fmt.Println(string(data))
					return nil
				}
				fmt.Printf("Dry run: sprint %s\n", sprintID)
				fmt.Printf("  Contracts: %d total\n", len(result.Contracts))
				fmt.Printf("  Would close:    %d\n", result.Closed)
				fmt.Printf("  Already done:   %d\n", result.AlreadyDone)
				for _, cid := range result.Contracts {
					fmt.Printf("    %s\n", cid)
				}
				return nil
			}
			result, err := artifact.SprintClose(rootDir, sprintID)
			if err != nil {
				return err
			}
			if closeFormat == "json" {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(data))
				return nil
			}
			fmt.Printf("Closed sprint %s: %d contract(s) marked complete, %d already done\n",
				sprintID, result.Closed, result.AlreadyDone)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview changes without applying")
	cmd.Flags().StringVar(&closeFormat, "format", "text", "Output format: text or json")
	return cmd
}

func planCmd() *cobra.Command {
	var (
		planMax    int
		planApply  bool
		planFormat string
	)
	cmd := &cobra.Command{
		Use:   "plan",
		Short: "Auto-suggest next sprint from unassigned backlog",
		RunE: func(c *cobra.Command, args []string) error {
			proposal, err := artifact.PlanSprint(rootDir, planMax)
			if err != nil {
				return err
			}

			if planFormat == "json" {
				data, _ := json.MarshalIndent(proposal, "", "  ")
				fmt.Println(string(data))
				if planApply {
					return applyProposal(proposal)
				}
				return nil
			}

			fmt.Printf("Proposed sprint: %s\n", proposal.Title)
			fmt.Printf("Contracts (%d):\n", len(proposal.ContractIDs))
			for _, cid := range proposal.ContractIDs {
				fmt.Printf("  %s\n", cid)
			}
			fmt.Printf("\n%s\n", proposal.Reasoning)

			if planApply {
				return applyProposal(proposal)
			}
			fmt.Println("\nUse --apply to create the sprint.")
			return nil
		},
	}
	cmd.Flags().IntVar(&planMax, "max", 8, "Maximum contracts per sprint")
	cmd.Flags().BoolVar(&planApply, "apply", false, "Create the sprint and assign contracts")
	cmd.Flags().StringVar(&planFormat, "format", "text", "Output format: text or json")
	return cmd
}

func applyProposal(proposal *artifact.SprintProposal) error {
	reg, err := registry.LoadRegistry(rootDir)
	if err != nil {
		return err
	}
	td, ok := reg.Types[names.KindSprint]
	if !ok {
		return fmt.Errorf("sprint type not configured")
	}
	fields := map[string]string{
		"title":     proposal.Title,
		"status":    "planned",
		"contracts": strings.Join(proposal.ContractIDs, ","),
	}
	path, err := artifact.GenericCreate(rootDir, td, "", fields)
	if err != nil {
		return fmt.Errorf("creating sprint: %w", err)
	}
	fmt.Printf("Created sprint: %s\n", filepath.Base(filepath.Dir(path)))
	return nil
}

func criteriaCmd(td registry.ArtifactTypeDef) *cobra.Command {
	return &cobra.Command{
		Use:   "criteria",
		Short: "Show acceptance criteria for a need",
		Args:  cobra.ArbitraryArgs,
		RunE: func(c *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: mos need criteria <NEED-ID>")
			}
			id := args[0]
			path, err := artifact.FindGenericArtifactPath(rootDir, td, id)
			if err != nil {
				return fmt.Errorf("mos need criteria: %w", err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("mos need criteria: %w", err)
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				return fmt.Errorf("mos need criteria: %w", err)
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				return fmt.Errorf("mos need criteria: invalid artifact")
			}
			criteria := artifact.ParseAcceptanceCriteria(ab)
			fmt.Print(artifact.FormatCriteria(criteria, nil))
			return nil
		},
	}
}
