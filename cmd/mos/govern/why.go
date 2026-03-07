package govern

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/artifact/chain"
	"github.com/dpopsuev/mos/moslib/registry"
)

var WhyCmd = &cobra.Command{
	Use:   "why <ID>",
	Short: "Trace any artifact to its root justification",
	Long: `Show the full lineage chain for an artifact — upward to its root need,
downward to implementations. Answers "why does this exist?"

Use --negative to also show scope exclusions and non-goals.`,
	Args: cobra.ExactArgs(1),
	RunE: runWhy,
}

var (
	whyFormat   string
	whyNegative bool
	whyKind     string
)

func init() {
	WhyCmd.Flags().StringVar(&whyFormat, "format", "text", "Output format: text or json")
	WhyCmd.Flags().BoolVar(&whyNegative, "negative", false, "Include negative-space chain")
	WhyCmd.Flags().StringVar(&whyKind, "kind", "", "Artifact kind override")
}

func runWhy(cmd *cobra.Command, args []string) error {
	id := args[0]
	kind := whyKind

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

	c, err := chain.WalkChain(".", kind, id)
	if err != nil {
		return err
	}

	if whyFormat == "json" {
		payload := map[string]interface{}{"chain": c}
		if whyNegative {
			nc, err := chain.WalkNegativeChain(".", kind, id)
			if err != nil {
				return err
			}
			payload["negative_space"] = nc
		}
		data, _ := json.MarshalIndent(payload, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	fmt.Print(chain.FormatChain(c))

	if whyNegative {
		nc, err := chain.WalkNegativeChain(".", kind, id)
		if err != nil {
			return err
		}
		fmt.Println()
		fmt.Print(chain.FormatNegativeChain(nc))
	}

	return nil
}
