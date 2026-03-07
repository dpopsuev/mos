package rule

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/cmd/mos/factory"
	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

var ruleTD = registry.ArtifactTypeDef{Kind: names.KindRule, Directory: names.DirRules}

// Cmd is the rule command, built via the artifact factory.
var Cmd = factory.Register(factory.KindConfig{
	TD:     ruleTD,
	Create: ruleCreateCmd(),
	List:   ruleListCmd(),
	Show:   ruleShowCmd(),
	Update: ruleUpdateCmd(),
	Delete: ruleDeleteCmd(),
	Apply:  ruleApplyCmd(),
	Edit:   ruleEditCmd(),
	Extra: []*cobra.Command{
		ruleAddSectionCmd(), ruleSetHarnessCmd(), ruleRemoveBlockCmd(), ruleAddWhenCmd(),
	},
})

func parseKeyValuePairs(s string) map[string]string {
	m := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return m
}

func applyOverflowFields(id string, overflow map[string]string) error {
	if len(overflow) == 0 {
		return nil
	}
	return artifact.UpdateRuleFields(rootDir, id, overflow)
}

func parseOverflowFromCreateSet(setPairs []string) map[string]string {
	overflow := make(map[string]string)
	for _, pair := range setPairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			overflow[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return overflow
}

// --- CRUD commands ---

func ruleCreateCmd() *cobra.Command {
	var flags struct {
		name, ruleType, enforcement, scope, glob          string
		appliesTo, harnessCmd, harnessTimeout, harnessReq string
		set                                               []string
	}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rule",
		Long:  "usage: mos rule create <id> --name <name> --type <type> [--scope <scope>] [--enforcement <level>]",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			opts := artifact.RuleOpts{
				Name:           flags.name,
				Type:           flags.ruleType,
				Enforcement:    flags.enforcement,
				Scope:          flags.scope,
				Glob:           flags.glob,
				HarnessCmd:     flags.harnessCmd,
				HarnessTimeout: flags.harnessTimeout,
			}
			if flags.appliesTo != "" {
				opts.AppliesTo = strings.Split(flags.appliesTo, ",")
			}
			if flags.harnessReq != "" {
				opts.HarnessRequires = parseKeyValuePairs(flags.harnessReq)
			}
			overflow := parseOverflowFromCreateSet(flags.set)
			rulePath, err := artifact.CreateRule(rootDir, id, opts)
			if err != nil {
				return fmt.Errorf("mos rule create: %w", err)
			}
			if id == "" {
				id = filepath.Base(filepath.Dir(rulePath))
			}
			if err := applyOverflowFields(id, overflow); err != nil {
				return fmt.Errorf("mos rule create: %w", err)
			}
			fmt.Printf("Created rule: %s\n", rulePath)
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.name, "name", "", "Rule name (required)")
	f.StringVar(&flags.ruleType, "type", "", "Rule type: mechanical or interpretive (required)")
	f.StringVar(&flags.enforcement, "enforcement", "", "Enforcement level")
	f.StringVar(&flags.scope, "scope", "", "Scope")
	f.StringVar(&flags.glob, "glob", "", "File glob pattern")
	f.StringVar(&flags.appliesTo, "applies-to", "", "Comma-separated kinds")
	f.StringVar(&flags.harnessCmd, "harness-cmd", "", "Harness command")
	f.StringVar(&flags.harnessTimeout, "harness-timeout", "", "Harness timeout")
	f.StringVar(&flags.harnessReq, "harness-requires", "", "Harness requirements (key=value,...)")
	f.StringArrayVar(&flags.set, "set", nil, "Set field (key=value, repeatable)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")
	return cmd
}

func ruleListCmd() *cobra.Command {
	var typeFilter, format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List rules",
		Long:  "usage: mos rule list [--type <type>] [--format text|json]",
		RunE: func(c *cobra.Command, args []string) error {
			rules, err := artifact.ListRules(rootDir, typeFilter)
			if err != nil {
				return fmt.Errorf("mos rule list: %w", err)
			}
			switch format {
			case names.FormatJSON:
				data, err := json.MarshalIndent(rules, "", "  ")
				if err != nil {
					return fmt.Errorf("mos rule list: %w", err)
				}
				fmt.Println(string(data))
			case names.FormatText:
				if len(rules) == 0 {
					fmt.Println("(no rules found)")
					return nil
				}
				for _, r := range rules {
					fmt.Printf("  %-25s %-14s %s\n", r.ID, r.Type, r.Name)
				}
			default:
				return fmt.Errorf("mos rule list: unknown format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&typeFilter, "type", "", "Filter by rule type")
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	return cmd
}

func ruleShowCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show a rule",
		Long:  "usage: mos rule show <id> [--format text|json]",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			content, err := artifact.ShowRule(rootDir, id)
			if err != nil {
				return fmt.Errorf("mos rule show: %w", err)
			}
			switch format {
			case names.FormatText:
				fmt.Print(content)
			case names.FormatJSON:
				data, err := json.Marshal(map[string]string{"id": id, "content": content})
				if err != nil {
					return fmt.Errorf("mos rule show: %w", err)
				}
				fmt.Println(string(data))
			default:
				return fmt.Errorf("mos rule show: unknown format %q", format)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "Output format: text or json")
	return cmd
}

func ruleUpdateCmd() *cobra.Command {
	var flags struct {
		name, ruleType, enforcement, scope, glob string
		harnessCmd, harnessTimeout, harnessReq   string
		set                                      []string
	}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update rule fields",
		Long:  "usage: mos rule update <id>... [--name ...] [--type ...] [--scope ...] [--enforcement ...] [--<field> ...]",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			opts := artifact.RuleUpdateOpts{}
			if flags.name != "" {
				opts.Name = &flags.name
			}
			if flags.ruleType != "" {
				opts.Type = &flags.ruleType
			}
			if flags.enforcement != "" {
				opts.Enforcement = &flags.enforcement
			}
			if flags.scope != "" {
				opts.Scope = &flags.scope
			}
			if flags.glob != "" {
				opts.Glob = &flags.glob
			}
			if flags.harnessCmd != "" {
				opts.HarnessCmd = &flags.harnessCmd
			}
			if flags.harnessTimeout != "" {
				opts.HarnessTimeout = &flags.harnessTimeout
			}
			if flags.harnessReq != "" {
				hr := parseKeyValuePairs(flags.harnessReq)
				opts.HarnessRequires = hr
			}
			overflow := parseOverflowFromCreateSet(flags.set)
			for _, id := range args {
				if err := artifact.UpdateRule(rootDir, id, opts); err != nil {
					return fmt.Errorf("mos rule update: %s: %w", id, err)
				}
				if err := applyOverflowFields(id, overflow); err != nil {
					return fmt.Errorf("mos rule update: %w", err)
				}
				fmt.Printf("Updated rule %s\n", id)
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&flags.name, "name", "", "Rule name")
	f.StringVar(&flags.ruleType, "type", "", "Rule type")
	f.StringVar(&flags.enforcement, "enforcement", "", "Enforcement level")
	f.StringVar(&flags.scope, "scope", "", "Scope")
	f.StringVar(&flags.glob, "glob", "", "File glob pattern")
	f.StringVar(&flags.harnessCmd, "harness-cmd", "", "Harness command")
	f.StringVar(&flags.harnessTimeout, "harness-timeout", "", "Harness timeout")
	f.StringVar(&flags.harnessReq, "harness-requires", "", "Harness requirements (key=value,...)")
	f.StringArrayVar(&flags.set, "set", nil, "Set field (key=value, repeatable)")
	return cmd
}

func ruleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if err := artifact.DeleteRule(rootDir, id); err != nil {
				return fmt.Errorf("mos rule delete: %w", err)
			}
			fmt.Printf("Deleted rule %s\n", id)
			return nil
		},
	}
}

func ruleApplyCmd() *cobra.Command {
	var filePath string
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply artifact from file or stdin",
		Long:  "usage: mos rule apply -f <path|->",
		RunE: func(c *cobra.Command, args []string) error {
			var content []byte
			var err error
			if filePath == "-" {
				content, err = io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("mos rule apply: reading stdin: %w", err)
				}
			} else {
				content, err = os.ReadFile(filePath)
				if err != nil {
					return fmt.Errorf("mos rule apply: reading file: %w", err)
				}
			}
			resultPath, err := artifact.ApplyArtifact(rootDir, content)
			if err != nil {
				return fmt.Errorf("mos rule apply: %w", err)
			}
			fmt.Printf("Applied rule: %s\n", resultPath)
			return nil
		},
	}
	cmd.Flags().StringVarP(&filePath, "file", "f", "", "File path or - for stdin (required)")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func ruleEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit a rule in the configured editor",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if err := artifact.EditArtifact(rootDir, names.KindRule, id); err != nil {
				return fmt.Errorf("mos rule edit: %w", err)
			}
			fmt.Println("Edited rule")
			return nil
		},
	}
}

// --- bespoke extras ---

func ruleAddSectionCmd() *cobra.Command {
	var text string
	var fromStdin bool
	cmd := &cobra.Command{
		Use:   "add-section",
		Short: "Add a section to a rule",
		Long:  "usage: mos rule add-section <id> <section-name> [--text <text> | --stdin]",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id, name := args[0], args[1]
			if fromStdin {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("mos rule add-section: reading stdin: %w", err)
				}
				text = string(data)
			}
			if text == "" {
				return fmt.Errorf("mos rule add-section: provide content via --text or --stdin")
			}
			if err := artifact.AddRuleSection(rootDir, id, name, text); err != nil {
				return fmt.Errorf("mos rule add-section: %w", err)
			}
			fmt.Printf("Added section %q to rule %s\n", name, id)
			return nil
		},
	}
	cmd.Flags().StringVar(&text, "text", "", "Section text content")
	cmd.Flags().BoolVar(&fromStdin, "stdin", false, "Read content from stdin")
	return cmd
}

func ruleSetHarnessCmd() *cobra.Command {
	var command, timeout string
	cmd := &cobra.Command{
		Use:   "set-harness",
		Short: "Set harness on a rule",
		Long:  "usage: mos rule set-harness <id> --command <cmd> [--timeout <dur>]",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			if err := artifact.SetRuleHarness(rootDir, id, command, timeout); err != nil {
				return fmt.Errorf("mos rule set-harness: %w", err)
			}
			fmt.Printf("Set harness on rule %s\n", id)
			return nil
		},
	}
	cmd.Flags().StringVar(&command, "command", "", "Harness command (required)")
	cmd.Flags().StringVar(&timeout, "timeout", "", "Harness timeout")
	_ = cmd.MarkFlagRequired("command")
	return cmd
}

func ruleRemoveBlockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove-block",
		Short: "Remove a block from a rule",
		Long:  "usage: mos rule remove-block <id> <block-type> [block-name]",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id, blockType := args[0], args[1]
			blockName := ""
			if len(args) > 2 {
				blockName = args[2]
			}
			if err := artifact.RemoveRuleBlock(rootDir, id, blockType, blockName); err != nil {
				return fmt.Errorf("mos rule remove-block: %w", err)
			}
			if blockName != "" {
				fmt.Printf("Removed %s block %q from rule %s\n", blockType, blockName, id)
			} else {
				fmt.Printf("Removed %s block from rule %s\n", blockType, id)
			}
			return nil
		},
	}
}

func ruleAddWhenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-when",
		Short: "Add a when block to a rule",
		Long:  "usage: mos rule add-when <id> <field=value> [field=value ...]",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			id := args[0]
			fields := make(map[string]string)
			for _, arg := range args[1:] {
				parts := strings.SplitN(arg, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("mos rule add-when: invalid field %q (expected key=value)", arg)
				}
				fields[parts[0]] = parts[1]
			}
			if err := artifact.AddRuleWhen(rootDir, id, fields); err != nil {
				return fmt.Errorf("mos rule add-when: %w", err)
			}
			fmt.Println("Added when block to rule")
			return nil
		},
	}
}
