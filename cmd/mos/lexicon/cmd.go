package lexicon

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/spf13/cobra"
)

const rootDir = "."

// Cmd is the lexicon command.
var Cmd = &cobra.Command{
	Use:   "lexicon",
	Short: "Manage the project lexicon",
	Long: `Manage the project lexicon (shared vocabulary of defined terms).

Sub-commands:
  list      List all defined terms
  add       Add a new term
  remove    Remove a term
  apply     Apply lexicon from a .mos file
  edit      Open the lexicon in $EDITOR`,
}

func init() {
	Cmd.AddCommand(
		listCmd,
		addCmd,
		removeCmd,
		factory.ApplyCmd(names.KindLexicon),
		factory.EditCmd(names.KindLexicon),
	)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all defined terms",
	Args:  cobra.ArbitraryArgs,
	RunE:  runList,
}

var listFormat string

func init() {
	listCmd.Flags().StringVar(&listFormat, "format", names.FormatText, "Output format: text, json")
}

func runList(cmd *cobra.Command, args []string) error {
	terms, err := artifact.ListTerms(rootDir)
	if err != nil {
		return fmt.Errorf("mos lexicon list: %w", err)
	}
	switch listFormat {
	case names.FormatJSON:
		data, err := json.MarshalIndent(terms, "", "  ")
		if err != nil {
			return fmt.Errorf("mos lexicon list: %w", err)
		}
		fmt.Println(string(data))
	case names.FormatText:
		if len(terms) == 0 {
			fmt.Println("(no terms defined)")
			return nil
		}
		for _, t := range terms {
			fmt.Printf("  %s = %q\n", t.Key, t.Description)
		}
	default:
		return fmt.Errorf("mos lexicon list: unknown format %q", listFormat)
	}
	return nil
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new term",
	Args:  cobra.ArbitraryArgs,
	RunE:  runAdd,
}

var addDescription string

func init() {
	addCmd.Flags().StringVar(&addDescription, "description", "", "Term definition (required)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	var key string
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			key = arg
			break
		}
	}
	if key == "" {
		return fmt.Errorf("usage: mos lexicon add <key> --description \"...\"")
	}
	if err := artifact.AddTerm(rootDir, key, addDescription); err != nil {
		return fmt.Errorf("mos lexicon add: %w", err)
	}
	fmt.Printf("Added term: %s\n", key)
	return nil
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a term",
	Args:  cobra.ArbitraryArgs,
	RunE:  runRemove,
}

func runRemove(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: mos lexicon remove <key>")
	}
	key := args[0]
	if err := artifact.RemoveTerm(rootDir, key); err != nil {
		return fmt.Errorf("mos lexicon remove: %w", err)
	}
	fmt.Printf("Removed term: %s\n", key)
	return nil
}
