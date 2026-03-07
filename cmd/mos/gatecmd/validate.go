package gatecmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/names"
)

var ValidateCmd = &cobra.Command{
	Use:   "validate [path]",
	Short: "Parse-only validation of .mos files (no cross-reference or lint analysis)",
	RunE:  runValidate,
}

var validateFormat string

func init() {
	ValidateCmd.Flags().StringVar(&validateFormat, "format", "text", "Output format: text or json")
}

type validateResult struct {
	File     string   `json:"file"`
	Line     int      `json:"line,omitempty"`
	Col      int      `json:"col,omitempty"`
	Message  string   `json:"message"`
	Expected []string `json:"expected,omitempty"`
	Got      string   `json:"got,omitempty"`
}

func runValidate(cmd *cobra.Command, args []string) error {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	mosDir := filepath.Join(root, names.MosDir)
	if _, err := os.Stat(mosDir); err != nil {
		return fmt.Errorf("no .mos directory found at %s", root)
	}

	var results []validateResult
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".mos") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		_, parseErr := dsl.Parse(string(data), nil)
		if parseErr != nil {
			var pe *dsl.ParseError
			r := validateResult{File: path, Message: parseErr.Error()}
			if errors.As(parseErr, &pe) {
				r.Line = pe.Line
				r.Col = pe.Col
				r.Expected = pe.Expected
				r.Got = pe.Got
			}
			results = append(results, r)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(results) == 0 {
		return nil
	}

	if validateFormat == names.FormatJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for _, r := range results {
		if r.Line > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "%s:%d: %s\n", r.File, r.Line, r.Message)
		} else {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", r.File, r.Message)
		}
	}
	return fmt.Errorf("%d parse error(s) found", len(results))
}
