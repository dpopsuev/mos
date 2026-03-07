package ci

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/governance/audit"
	"github.com/dpopsuev/mos/moslib/guard"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "ci",
	Short: "Run the local CI pipeline",
	Long: `Run the local CI pipeline. Stages run sequentially and fail fast.

Stages:
  build     go build ./...
  vet       go vet ./...
  test      go test -short ./... (skipped by --fast; use --stress for full suite)
  lint      mos lint             (internal)
  audit     mos audit            (internal, skipped by --fast)
  harness   mos harness run      (internal, skipped by --fast)`,
	RunE: runCI,
}

func init() {
	Cmd.Flags().Bool("fast", false, "Only run build + vet + lint (sub-second feedback)")
	Cmd.Flags().Bool("stress", false, "Run full test suite including stress/perf tests (without -short)")
	Cmd.Flags().Bool("fix", false, "Run mos fmt + go mod tidy before the pipeline")
	Cmd.Flags().String("format", names.FormatText, "Output format: text (default), json")
}

func runCI(cmd *cobra.Command, args []string) error {
	fast, _ := cmd.Flags().GetBool("fast")
	stress, _ := cmd.Flags().GetBool("stress")
	fix, _ := cmd.Flags().GetBool("fix")
	format, _ := cmd.Flags().GetString("format")

	if fix {
		if err := runFixSteps(); err != nil {
			return err
		}
	}

	testArgs := []string{"test", "-short", "./..."}
	if stress {
		testArgs = []string{"test", "./..."}
	}

	type ciStage struct {
		Name     string
		FastSkip bool
		Run      func() (bool, string)
	}

	stages := []ciStage{
		{"build", false, func() (bool, string) { return ciShell("go", "build", "./...") }},
		{"vet", false, func() (bool, string) { return ciShell("go", "vet", "./...") }},
		{"test", true, func() (bool, string) { return ciShell("go", testArgs...) }},
		{"lint", false, ciLint},
		{"audit", true, ciAudit},
		{"harness", true, ciHarness},
		{"vectors", true, ciVectors},
	}

	type ciResult struct {
		Stage    string        `json:"stage"`
		Pass     bool          `json:"pass"`
		Duration time.Duration `json:"duration_ms"`
		Output   string        `json:"output,omitempty"`
	}

	var results []ciResult
	allPass := true

	for _, s := range stages {
		if fast && s.FastSkip {
			continue
		}
		start := time.Now()
		pass, output := s.Run()
		dur := time.Since(start)
		results = append(results, ciResult{
			Stage:    s.Name,
			Pass:     pass,
			Duration: dur / time.Millisecond,
			Output:   output,
		})
		if format == names.FormatText {
			status := "PASS"
			if !pass {
				status = "FAIL"
			}
			fmt.Fprintf(os.Stderr, "%-10s %s  %s\n", s.Name, status, dur.Truncate(time.Millisecond))
		}
		if !pass {
			allPass = false
			if format == names.FormatText && output != "" {
				fmt.Fprintln(os.Stderr, output)
			}
			break
		}
	}

	if format == names.FormatJSON {
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	}

	if allPass {
		if format == names.FormatText {
			fmt.Fprintf(os.Stderr, "\nAll stages passed.\n")
		}
		return nil
	}
	return fmt.Errorf("CI pipeline failed")
}

func runFixSteps() error {
	fmt.Fprintln(os.Stderr, "fix: mos fmt .")
	if err := ciRunFmt("."); err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "fix: auto-archive")
	relocations, err := artifact.RelocateMisplacedArtifacts(".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "auto-archive: %v\n", err)
	}
	for _, r := range relocations {
		fmt.Fprintf(os.Stderr, "  relocated %s %s: %s → %s\n", r.Kind, r.ID, r.From, r.To)
	}
	fmt.Fprintln(os.Stderr, "fix: go mod tidy")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = "."
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ciRunFmt(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("mos fmt: %w", err)
	}
	var files []string
	if info.IsDir() {
		err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() && strings.HasSuffix(fi.Name(), names.MosDir) {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("mos fmt: %w", err)
		}
	} else {
		files = []string{path}
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mos fmt: reading %s: %v\n", f, err)
			continue
		}
		parsed, err := dsl.Parse(string(data), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "mos fmt: parsing %s: %v\n", f, err)
			continue
		}
		formatted := dsl.Format(parsed, nil)
		if string(data) != formatted {
			if err := os.WriteFile(f, []byte(formatted), names.FilePerm); err != nil {
				fmt.Fprintf(os.Stderr, "mos fmt: writing %s: %v\n", f, err)
				continue
			}
			fmt.Printf("formatted %s\n", f)
		}
	}
	return nil
}

func ciShell(name string, args ...string) (bool, string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = "."
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return err == nil, buf.String()
}

func ciLint() (bool, string) {
	result := guard.CIGate(".", "lint")
	return result.Pass, result.Output
}

func ciAudit() (bool, string) {
	report, err := audit.RunAudit(".", audit.AuditOpts{})
	if err != nil {
		return false, err.Error()
	}
	pass := report.LintErrors == 0 && len(report.Collisions) == 0
	return pass, audit.FormatReport(report, false)
}

func ciHarness() (bool, string) {
	result := guard.CIGate(".", "harness")
	return result.Pass, result.Output
}

func ciVectors() (bool, string) {
	result := guard.CIGate(".", "vectors")
	return result.Pass, result.Output
}
