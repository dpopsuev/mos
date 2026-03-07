package gatecmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/harness"
)

const maquetteHistoryPath = ".mos/maquettes/history.json"

type maquetteRun struct {
	Timestamp string   `json:"timestamp"`
	Mode      string   `json:"mode"`
	Pattern   string   `json:"pattern"`
	Value     float64  `json:"value"`
	Found     bool     `json:"found"`
	Ceiling   *float64 `json:"ceiling,omitempty"`
	Floor     *float64 `json:"floor,omitempty"`
	Pass      bool     `json:"pass"`
	Unit      string   `json:"unit,omitempty"`
}

var (
	mqPattern string
	mqInput   string
	mqFile    string
	mqStdin   bool
	mqCeiling float64
	mqFloor   float64
	mqUnit    string
	mqFormat  string
	mqMode    string
	mqHistory bool
	mqSave    int
	mqName    string
	mqRule    string

	mqHasCeiling bool
	mqHasFloor   bool
)

var MaquetteCmd = &cobra.Command{
	Use:   "maquette",
	Short: "Dry-run metric extraction for pattern validation",
	Long: `Interactive dry-run tool for testing metric extraction patterns.
Supports regex (stdout) and json (file) modes.

Examples:
  mos harness maquette --pattern 'latency:\s*(?P<value>[\d.]+)' --input 'latency: 42.5ms' --ceiling 50
  mos harness maquette --mode json --pattern 'result.median' --file bench.json
  mos harness maquette --history
  mos harness maquette --save 0 --name latency_p99 --rule RULE-2026-001`,
	RunE: runMaquette,
}

func init() {
	f := MaquetteCmd.Flags()
	f.StringVar(&mqPattern, "pattern", "", "Regex pattern (with named group 'value') or JSON dot-path")
	f.StringVar(&mqInput, "input", "", "Input string to parse (regex mode)")
	f.StringVar(&mqFile, "file", "", "Path to JSON file (json mode)")
	f.BoolVar(&mqStdin, "stdin", false, "Read input from stdin")
	f.Float64Var(&mqCeiling, "ceiling", 0, "Upper bound threshold")
	f.Float64Var(&mqFloor, "floor", 0, "Lower bound threshold")
	f.StringVar(&mqUnit, "unit", "", "Unit label")
	f.StringVar(&mqFormat, "format", "text", "Output format: text or json")
	f.StringVar(&mqMode, "mode", "regex", "Extraction mode: regex or json")
	f.BoolVar(&mqHistory, "history", false, "Show previous maquette runs")
	f.IntVar(&mqSave, "save", -1, "Promote history entry at index to a rule's harness block")
	f.StringVar(&mqName, "name", "", "Metric name for --save promotion")
	f.StringVar(&mqRule, "rule", "", "Rule ID for --save promotion")

	HarnessCmd.AddCommand(MaquetteCmd)
}

func runMaquette(cmd *cobra.Command, args []string) error {
	if mqHistory {
		return showHistory(mqFormat)
	}

	if mqSave >= 0 {
		return promoteMaquette(mqSave, mqName, mqRule)
	}

	if mqPattern == "" {
		return fmt.Errorf("--pattern is required")
	}

	mqHasCeiling = cmd.Flags().Changed("ceiling")
	mqHasFloor = cmd.Flags().Changed("floor")

	switch mqMode {
	case "regex":
		return runRegexMaquette()
	case "json":
		return runJSONMaquette()
	default:
		return fmt.Errorf("unknown mode %q, use 'regex' or 'json'", mqMode)
	}
}

func runRegexMaquette() error {
	input, err := resolveInput()
	if err != nil {
		return err
	}

	var ceil, flr *float64
	if mqHasCeiling {
		ceil = &mqCeiling
	}
	if mqHasFloor {
		flr = &mqFloor
	}

	th := harness.MetricThreshold{
		Name:    "maquette",
		Pattern: mqPattern,
		Ceiling: ceil,
		Floor:   flr,
		Unit:    mqUnit,
	}

	results := harness.ExtractMetrics("", input, []harness.MetricThreshold{th})

	run := maquetteRun{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Mode:      "regex",
		Pattern:   mqPattern,
		Unit:      mqUnit,
		Ceiling:   ceil,
		Floor:     flr,
	}

	if len(results) == 0 {
		run.Found = false
		run.Pass = false
	} else {
		run.Found = true
		run.Value = results[0].Value
		run.Pass = results[0].Pass
	}

	appendHistory(run)
	return outputRun(run)
}

func runJSONMaquette() error {
	if mqFile == "" {
		return fmt.Errorf("--file is required in json mode")
	}

	var ceil, flr *float64
	if mqHasCeiling {
		ceil = &mqCeiling
	}
	if mqHasFloor {
		flr = &mqFloor
	}

	th := harness.MetricThreshold{
		Name:     "maquette",
		Source:   mqFile,
		JSONPath: mqPattern,
		Ceiling:  ceil,
		Floor:    flr,
		Unit:     mqUnit,
	}

	results := harness.ExtractMetrics(".", "", []harness.MetricThreshold{th})

	run := maquetteRun{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Mode:      "json",
		Pattern:   mqPattern,
		Unit:      mqUnit,
		Ceiling:   ceil,
		Floor:     flr,
	}

	if len(results) == 0 {
		run.Found = false
		run.Pass = false
	} else {
		run.Found = true
		run.Value = results[0].Value
		run.Pass = results[0].Pass
	}

	appendHistory(run)
	return outputRun(run)
}

func resolveInput() (string, error) {
	if mqStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading stdin: %w", err)
		}
		return string(data), nil
	}
	if mqInput != "" {
		return mqInput, nil
	}
	return "", fmt.Errorf("--input or --stdin is required in regex mode")
}

func outputRun(run maquetteRun) error {
	if mqFormat == "json" {
		data, _ := json.MarshalIndent(run, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if !run.Found {
		fmt.Println("Result:  NOT FOUND")
		fmt.Printf("Pattern: %s\n", run.Pattern)
		return nil
	}

	verdict := "PASS"
	if !run.Pass {
		verdict = "FAIL"
	}
	fmt.Printf("Result:  %s\n", verdict)
	fmt.Printf("Value:   %.6g\n", run.Value)
	if run.Ceiling != nil {
		fmt.Printf("Ceiling: %.6g\n", *run.Ceiling)
	}
	if run.Floor != nil {
		fmt.Printf("Floor:   %.6g\n", *run.Floor)
	}
	if run.Unit != "" {
		fmt.Printf("Unit:    %s\n", run.Unit)
	}
	fmt.Printf("Pattern: %s\n", run.Pattern)
	return nil
}

func showHistory(format string) error {
	runs, err := loadHistory()
	if err != nil {
		return err
	}
	if len(runs) == 0 {
		fmt.Println("No maquette history.")
		return nil
	}

	if format == "json" {
		data, _ := json.MarshalIndent(runs, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	for i, r := range runs {
		verdict := "PASS"
		if !r.Found {
			verdict = "NOT_FOUND"
		} else if !r.Pass {
			verdict = "FAIL"
		}
		fmt.Printf("[%d] %s  mode=%-5s  value=%.6g  %s  pattern=%s\n",
			i, r.Timestamp, r.Mode, r.Value, verdict, r.Pattern)
	}
	return nil
}

func promoteMaquette(index int, name, ruleID string) error {
	if name == "" || ruleID == "" {
		return fmt.Errorf("--name and --rule are required with --save")
	}
	runs, err := loadHistory()
	if err != nil {
		return err
	}
	if index < 0 || index >= len(runs) {
		return fmt.Errorf("index %d out of range (0..%d)", index, len(runs)-1)
	}
	r := runs[index]

	snippet := fmt.Sprintf(`    metric %q {`, name)
	if r.Mode == "regex" {
		snippet += fmt.Sprintf("\n      pattern = %q", r.Pattern)
	} else {
		snippet += fmt.Sprintf("\n      json_path = %q", r.Pattern)
	}
	if r.Ceiling != nil {
		snippet += fmt.Sprintf("\n      ceiling = %.6g", *r.Ceiling)
	}
	if r.Floor != nil {
		snippet += fmt.Sprintf("\n      floor = %.6g", *r.Floor)
	}
	if r.Unit != "" {
		snippet += fmt.Sprintf("\n      unit = %q", r.Unit)
	}
	snippet += "\n    }"

	fmt.Printf("Promoted maquette entry [%d] to metric %q for rule %s:\n\n%s\n\n", index, name, ruleID, snippet)
	fmt.Println("Paste the above block into the harness section of the rule file.")
	return nil
}

func loadHistory() ([]maquetteRun, error) {
	data, err := os.ReadFile(maquetteHistoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var runs []maquetteRun
	if err := json.Unmarshal(data, &runs); err != nil {
		return nil, err
	}
	return runs, nil
}

func appendHistory(run maquetteRun) {
	runs, _ := loadHistory()
	runs = append([]maquetteRun{run}, runs...)

	const maxHistory = 50
	if len(runs) > maxHistory {
		runs = runs[:maxHistory]
	}

	_ = os.MkdirAll(filepath.Dir(maquetteHistoryPath), 0o755)
	data, _ := json.MarshalIndent(runs, "", "  ")
	_ = os.WriteFile(maquetteHistoryPath, data, 0o644)
}
