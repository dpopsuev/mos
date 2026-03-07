package harness

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFormatTextEmpty(t *testing.T) {
	out := FormatText(nil)
	if !strings.Contains(out, "No harness specs found") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestFormatTextPass(t *testing.T) {
	results := []Evidence{
		{RuleID: "build-pass", Command: "go build ./...", ExitCode: 0, Duration: 500 * time.Millisecond, Pass: true},
		{RuleID: "test-pass", Command: "go test ./...", ExitCode: 0, Duration: 1200 * time.Millisecond, Pass: true},
	}
	out := FormatText(results)
	if !strings.Contains(out, "PASS") {
		t.Errorf("expected PASS in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2/2 passed") {
		t.Errorf("expected '2/2 passed', got:\n%s", out)
	}
}

func TestFormatTextFail(t *testing.T) {
	results := []Evidence{
		{RuleID: "build-pass", Command: "go build ./...", ExitCode: 0, Duration: 100 * time.Millisecond, Pass: true},
		{RuleID: "bad-rule", Command: "false", ExitCode: 1, Duration: 10 * time.Millisecond, Pass: false},
	}
	out := FormatText(results)
	if !strings.Contains(out, "FAIL(1)") {
		t.Errorf("expected FAIL(1) in output, got:\n%s", out)
	}
	if !strings.Contains(out, "1/2 passed") {
		t.Errorf("expected '1/2 passed', got:\n%s", out)
	}
}

func TestFormatTextTimeout(t *testing.T) {
	results := []Evidence{
		{RuleID: "slow", Command: "sleep 60", ExitCode: -1, Duration: 1 * time.Second, Pass: false, TimedOut: true},
	}
	out := FormatText(results)
	if !strings.Contains(out, "TIMEOUT") {
		t.Errorf("expected TIMEOUT in output, got:\n%s", out)
	}
}

func TestFormatJSON(t *testing.T) {
	results := []Evidence{
		{RuleID: "build-pass", Command: "go build ./...", ExitCode: 0, Duration: 500 * time.Millisecond, Pass: true},
	}
	data, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}

	var parsed []jsonEvidence
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(parsed))
	}
	if parsed[0].RuleID != "build-pass" {
		t.Errorf("rule_id = %q, want build-pass", parsed[0].RuleID)
	}
	if parsed[0].DurationMs != 500 {
		t.Errorf("duration_ms = %d, want 500", parsed[0].DurationMs)
	}
	if !parsed[0].Pass {
		t.Error("expected pass = true")
	}
}

func TestDiscoverAndRun(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	rulesDir := filepath.Join(mosDir, "rules", "mechanical")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}

	ruleContent := `rule "echo-test" {
  name = "Echo Test"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo hello"
    timeout = "10s"
  }
}
`
	if err := os.WriteFile(filepath.Join(rulesDir, "echo-test.mos"), []byte(ruleContent), 0644); err != nil {
		t.Fatal(err)
	}

	specs, err := Discover(mosDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if specs[0].RuleID != "echo-test" {
		t.Errorf("rule ID = %q, want echo-test", specs[0].RuleID)
	}
	if specs[0].Command != "echo hello" {
		t.Errorf("command = %q, want 'echo hello'", specs[0].Command)
	}
	if specs[0].Timeout != 10*time.Second {
		t.Errorf("timeout = %v, want 10s", specs[0].Timeout)
	}

	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pass {
		t.Errorf("expected pass, got fail: exit=%d stderr=%q", results[0].ExitCode, results[0].Stderr)
	}
	if !strings.Contains(results[0].Stdout, "hello") {
		t.Errorf("expected 'hello' in stdout, got: %q", results[0].Stdout)
	}
}

func TestRunFailingCommand(t *testing.T) {
	root := t.TempDir()
	specs := []HarnessSpec{
		{RuleID: "fail", Command: "exit 1", Timeout: 10 * time.Second},
	}
	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("expected fail, got pass")
	}
	if results[0].ExitCode != 1 {
		t.Errorf("exit code = %d, want 1", results[0].ExitCode)
	}
}

func TestRunTimeout(t *testing.T) {
	root := t.TempDir()
	specs := []HarnessSpec{
		{RuleID: "slow", Command: "sleep 60", Timeout: 500 * time.Millisecond},
	}
	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("expected fail due to timeout, got pass")
	}
	if !results[0].TimedOut {
		t.Error("expected TimedOut = true")
	}
}

func TestDiscoverSkipsInterpretiveWithoutHarness(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	mechDir := filepath.Join(mosDir, "rules", "mechanical")
	interpDir := filepath.Join(mosDir, "rules", "interpretive")
	os.MkdirAll(mechDir, 0755)
	os.MkdirAll(interpDir, 0755)

	interpRule := `rule "bdd-specs" {
  name = "BDD Specs Required"
  type = "interpretive"
  scope = "project"
  enforcement = "warning"
}
`
	os.WriteFile(filepath.Join(interpDir, "bdd-specs.mos"), []byte(interpRule), 0644)

	mechRule := `rule "build" {
  name = "Build"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo ok"
    timeout = "5s"
  }
}
`
	os.WriteFile(filepath.Join(mechDir, "build.mos"), []byte(mechRule), 0644)

	specs, err := Discover(mosDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec (skipping interpretive without harness), got %d", len(specs))
	}
	if specs[0].RuleID != "build" {
		t.Errorf("rule ID = %q, want build", specs[0].RuleID)
	}
}

// --- Metric extraction and threshold tests ---

func TestExtractMetricsGoBenchmark(t *testing.T) {
	stdout := `goos: linux
goarch: amd64
pkg: github.com/dpopsuev/mos/moslib/harness
BenchmarkFoo-8   	 1000000	      1234.0 ns/op	     128 B/op	       3 allocs/op
BenchmarkBar-8   	  500000	      2500.0 ns/op	      64 B/op	       1 allocs/op
PASS
ok  	github.com/dpopsuev/mos/moslib/harness	3.456s
`
	ceil := 2000.0
	thresholds := []MetricThreshold{
		{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 1 {
		t.Fatalf("expected 1 metric result, got %d", len(results))
	}
	if results[0].Value != 2500.0 {
		t.Errorf("ns_op worst-case = %f, want 2500.0", results[0].Value)
	}
	if results[0].Pass {
		t.Error("expected ns_op to fail (2500 > ceiling 2000)")
	}
	if results[0].Message == "" {
		t.Error("expected breach message")
	}
}

func TestExtractMetricsGoBenchmarkPass(t *testing.T) {
	stdout := `BenchmarkFoo-8   1000   800.0 ns/op   32 B/op   2 allocs/op
`
	ceil := 1000.0
	thresholds := []MetricThreshold{
		{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pass {
		t.Errorf("expected pass (800 < 1000), got fail: %s", results[0].Message)
	}
}

func TestExtractMetricsFloor(t *testing.T) {
	stdout := `BenchmarkFoo-8   1000   50.0 ns/op
`
	floor := 100.0
	thresholds := []MetricThreshold{
		{Name: "ns_op", Floor: &floor, Unit: "ns"},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("expected fail (50 < floor 100)")
	}
}

func TestExtractMetricsBytesAndAllocs(t *testing.T) {
	stdout := `BenchmarkFoo-8   1000   500.0 ns/op   256 B/op   5 allocs/op
`
	bytesCeil := 128.0
	allocsCeil := 3.0
	thresholds := []MetricThreshold{
		{Name: "bytes_op", Ceiling: &bytesCeil, Unit: "B"},
		{Name: "allocs_op", Ceiling: &allocsCeil, Unit: "allocs"},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("bytes_op should fail (256 > 128)")
	}
	if results[1].Pass {
		t.Error("allocs_op should fail (5 > 3)")
	}
}

func TestExtractMetricsCustomPattern(t *testing.T) {
	stdout := `latency_p99: 42.5ms
`
	ceil := 50.0
	thresholds := []MetricThreshold{
		{Name: "latency_p99", Ceiling: &ceil, Unit: "ms", Pattern: `latency_p99:\s*(?P<value>[\d.]+)ms`},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 42.5 {
		t.Errorf("value = %f, want 42.5", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass (42.5 < 50)")
	}
}

func TestExtractMetricsNoMatchReturnsEmpty(t *testing.T) {
	stdout := "no benchmark output here\n"
	ceil := 100.0
	thresholds := []MetricThreshold{
		{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
	}
	results := ExtractMetrics("", stdout, thresholds)
	if len(results) != 0 {
		t.Errorf("expected 0 results when no benchmarks found, got %d", len(results))
	}
}

func TestExtractMetricsNilThresholds(t *testing.T) {
	results := ExtractMetrics("", "anything", nil)
	if results != nil {
		t.Errorf("expected nil, got %v", results)
	}
}

func TestDiscoverMetricThresholds(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	rulesDir := filepath.Join(mosDir, "rules", "mechanical")
	os.MkdirAll(rulesDir, 0755)

	ruleContent := `rule "perf-bench" {
  name = "Performance Benchmark"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "go test -bench=. ./..."
    timeout = "2m"

    metric "ns_op" {
      ceiling = 5000
      unit = "ns"
    }

    metric "allocs_op" {
      ceiling = 10
      unit = "allocs"
    }
  }
}
`
	os.WriteFile(filepath.Join(rulesDir, "perf-bench.mos"), []byte(ruleContent), 0644)

	specs, err := Discover(mosDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if len(specs[0].Thresholds) != 2 {
		t.Fatalf("expected 2 thresholds, got %d", len(specs[0].Thresholds))
	}
	th0 := specs[0].Thresholds[0]
	if th0.Name != "ns_op" {
		t.Errorf("threshold[0].Name = %q, want ns_op", th0.Name)
	}
	if th0.Ceiling == nil || *th0.Ceiling != 5000 {
		t.Errorf("threshold[0].Ceiling = %v, want 5000", th0.Ceiling)
	}
	if th0.Unit != "ns" {
		t.Errorf("threshold[0].Unit = %q, want ns", th0.Unit)
	}
	th1 := specs[0].Thresholds[1]
	if th1.Name != "allocs_op" {
		t.Errorf("threshold[1].Name = %q, want allocs_op", th1.Name)
	}
	if th1.Ceiling == nil || *th1.Ceiling != 10 {
		t.Errorf("threshold[1].Ceiling = %v, want 10", th1.Ceiling)
	}
}

func TestRunWithMetricThresholdBreach(t *testing.T) {
	root := t.TempDir()
	ceil := 500.0
	specs := []HarnessSpec{
		{
			RuleID:      "bench",
			Command:     `printf 'BenchmarkX-8\t1000\t1234.0 ns/op\t64 B/op\t2 allocs/op\n'`,
			Timeout:     10 * time.Second,
			Enforcement: "error",
			Thresholds: []MetricThreshold{
				{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
			},
		},
	}
	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Pass {
		t.Error("expected fail due to metric breach (1234 > 500)")
	}
	if len(results[0].Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(results[0].Metrics))
	}
	if results[0].Metrics[0].Pass {
		t.Error("metric should show breach")
	}
}

func TestRunWithMetricThresholdWarningDoesNotFail(t *testing.T) {
	root := t.TempDir()
	ceil := 500.0
	specs := []HarnessSpec{
		{
			RuleID:      "bench",
			Command:     `printf 'BenchmarkX-8\t1000\t1234.0 ns/op\n'`,
			Timeout:     10 * time.Second,
			Enforcement: "warning",
			Thresholds: []MetricThreshold{
				{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
			},
		},
	}
	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pass {
		t.Error("warning enforcement should not fail the harness, only report breach")
	}
	if len(results[0].Metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(results[0].Metrics))
	}
	if results[0].Metrics[0].Pass {
		t.Error("metric should show breach even with warning enforcement")
	}
}

func TestFormatTextWithMetrics(t *testing.T) {
	ceil := 100.0
	results := []Evidence{
		{
			RuleID: "bench", Command: "go test -bench=.", Pass: true, Duration: 1 * time.Second,
			Metrics: []MetricResult{
				{Name: "ns_op", Value: 50.0, Unit: "ns", Ceiling: &ceil, Pass: true},
			},
		},
	}
	out := FormatText(results)
	if !strings.Contains(out, "ns_op") {
		t.Errorf("expected metric name in text output, got:\n%s", out)
	}
	if !strings.Contains(out, "OK") {
		t.Errorf("expected OK tag in text output, got:\n%s", out)
	}
}

func TestFormatTextWithMetricBreach(t *testing.T) {
	ceil := 100.0
	results := []Evidence{
		{
			RuleID: "bench", Command: "go test -bench=.", Pass: false, Duration: 1 * time.Second,
			Metrics: []MetricResult{
				{Name: "ns_op", Value: 150.0, Unit: "ns", Ceiling: &ceil, Pass: false, Message: "ns_op=150.00 exceeds ceiling 100.00 ns"},
			},
		},
	}
	out := FormatText(results)
	if !strings.Contains(out, "BREACH") {
		t.Errorf("expected BREACH tag in text output, got:\n%s", out)
	}
	if !strings.Contains(out, "exceeds ceiling") {
		t.Errorf("expected breach message in text output, got:\n%s", out)
	}
}

func TestFormatJSONWithMetrics(t *testing.T) {
	ceil := 100.0
	results := []Evidence{
		{
			RuleID: "bench", Command: "go test", Pass: true, Duration: 1 * time.Second,
			Metrics: []MetricResult{
				{Name: "ns_op", Value: 50.0, Unit: "ns", Ceiling: &ceil, Pass: true},
			},
		},
	}
	data, err := FormatJSON(results)
	if err != nil {
		t.Fatalf("FormatJSON: %v", err)
	}
	var parsed []jsonEvidence
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(parsed[0].Metrics) != 1 {
		t.Fatalf("expected 1 metric in JSON, got %d", len(parsed[0].Metrics))
	}
	if parsed[0].Metrics[0].Name != "ns_op" {
		t.Errorf("metric name = %q, want ns_op", parsed[0].Metrics[0].Name)
	}
}

func TestIntegrationInitCreateRuleDiscoverRun(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module test\n\ngo 1.25.7\n"), 0644); err != nil {
		t.Fatal(err)
	}

	mosDir := filepath.Join(root, ".mos")
	dirs := []string{
		mosDir,
		filepath.Join(mosDir, "lexicon"),
		filepath.Join(mosDir, "resolution"),
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "rules", "interpretive"),
		filepath.Join(mosDir, "contracts", "active"),
		filepath.Join(mosDir, "contracts", "archive"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	configContent := `config {
  mos {
    version = 1
  }
  backend {
    type = "git"
  }
  governance {
    model = "bdfl"
    scope = "cabinet"
  }
}
`
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte(configContent), 0644)

	lexiconContent := `lexicon {
  terms {
  }
}
`
	os.WriteFile(filepath.Join(mosDir, "lexicon", "default.mos"), []byte(lexiconContent), 0644)

	layersContent := `layers {
  layer "project" {
    level = 1
  }
}
`
	os.WriteFile(filepath.Join(mosDir, "resolution", "layers.mos"), []byte(layersContent), 0644)

	declContent := `declaration {
  name = "test"
  created = "2026-03-02T00:00:00Z"
}
`
	os.WriteFile(filepath.Join(mosDir, "declaration.mos"), []byte(declContent), 0644)

	ruleContent := `rule "echo-pass" {
  name = "Echo Pass"
  type = "mechanical"
  scope = "project"
  enforcement = "error"

  harness {
    command = "echo integration-test"
    timeout = "10s"
  }
}
`
	os.WriteFile(filepath.Join(mosDir, "rules", "mechanical", "echo-pass.mos"), []byte(ruleContent), 0644)

	specs, err := Discover(mosDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}

	results := Run(root, specs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Pass {
		t.Errorf("expected pass, got fail: exit=%d stderr=%q", results[0].ExitCode, results[0].Stderr)
	}
	if results[0].RuleID != "echo-pass" {
		t.Errorf("rule ID = %q, want echo-pass", results[0].RuleID)
	}
}

func TestRegistryLookupBuiltins(t *testing.T) {
	for _, name := range []string{"ns_op", "bytes_op", "B_op", "allocs_op"} {
		if _, ok := LookupExtractor(name); !ok {
			t.Errorf("expected built-in extractor %q to be registered", name)
		}
	}
}

func TestRegistryLookupMissing(t *testing.T) {
	if _, ok := LookupExtractor("nonexistent_extractor"); ok {
		t.Error("expected LookupExtractor to return false for unknown name")
	}
}

func TestRegisterExtractorCustom(t *testing.T) {
	e := Extractor{Name: "test_custom_ext", Kind: "regex", Pattern: `val:\s*(?P<value>[\d.]+)`, Unit: "ms"}
	RegisterExtractor(e)
	defer delete(builtinExtractors, "test_custom_ext")

	got, ok := LookupExtractor("test_custom_ext")
	if !ok {
		t.Fatal("expected custom extractor to be found")
	}
	if got.Pattern != e.Pattern {
		t.Errorf("pattern = %q, want %q", got.Pattern, e.Pattern)
	}
}

func TestExtractMetricsRegistryRegex(t *testing.T) {
	RegisterExtractor(Extractor{
		Name:    "test_reg_metric",
		Kind:    "regex",
		Pattern: `throughput:\s*(?P<value>[\d.]+)\s*ops/s`,
		Unit:    "ops/s",
	})
	defer delete(builtinExtractors, "test_reg_metric")

	stdout := "throughput: 1234.56 ops/s\n"
	ceil := 2000.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "test_reg_metric", Ceiling: &ceil, Unit: "ops/s"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 1234.56 {
		t.Errorf("value = %f, want 1234.56", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass")
	}
}

func TestExtractMetricsResolutionOrder(t *testing.T) {
	RegisterExtractor(Extractor{
		Name:    "ns_op",
		Kind:    "regex",
		Pattern: `custom_ns:\s*(?P<value>[\d.]+)`,
		Unit:    "ns",
	})
	defer func() {
		builtinExtractors["ns_op"] = Extractor{Name: "ns_op", Kind: "regex", Unit: "ns"}
	}()

	stdout := "BenchmarkFoo-8   1000   500.0 ns/op\ncustom_ns: 999.0\n"
	ceil := 2000.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "ns_op", Ceiling: &ceil, Unit: "ns"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Registry regex should take precedence over Go bench fallback when it has a pattern
	if results[0].Value != 999.0 {
		t.Errorf("value = %f, want 999.0 (registry regex should win)", results[0].Value)
	}
}

func TestExtractFromJSONBasic(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"result": {"median": 42.5}}`
	os.WriteFile(filepath.Join(dir, "bench.json"), []byte(jsonContent), 0644)

	ceil := 50.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "custom_json", Ceiling: &ceil, Unit: "ns", Source: "bench.json", JSONPath: "result.median"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 42.5 {
		t.Errorf("value = %f, want 42.5", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass")
	}
}

func TestExtractFromJSONArrayWildcard(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"benchmarks": [{"time": 10.0}, {"time": 25.0}, {"time": 15.0}]}`
	os.WriteFile(filepath.Join(dir, "results.json"), []byte(jsonContent), 0644)

	ceil := 30.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "array_metric", Ceiling: &ceil, Unit: "ns", Source: "results.json", JSONPath: "benchmarks[*].time"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 25.0 {
		t.Errorf("value = %f, want 25.0 (max aggregation)", results[0].Value)
	}
}

func TestExtractJMH_NsOp(t *testing.T) {
	stdout := `Benchmark                        Mode  Cnt       Score      Error  Units
c.e.MyBenchmark.testMethod         avgt   10    1234.567 ±   12.34  ns/op
`
	ceil := 2000.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "jmh_ns_op", Ceiling: &ceil, Unit: "ns/op"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 1234.567 {
		t.Errorf("value = %f, want 1234.567", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass")
	}
}

func TestExtractJMH_MsOp(t *testing.T) {
	stdout := `Benchmark                        Mode  Cnt       Score      Error  Units
c.e.MyBenchmark.testMethod         avgt   10      5.678 ±    0.12  ms/op
`
	ceil := 10.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "jmh_ms_op", Ceiling: &ceil, Unit: "ms/op"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 5.678 {
		t.Errorf("value = %f, want 5.678", results[0].Value)
	}
}

func TestExtractJMH_OpsS(t *testing.T) {
	stdout := `Benchmark                        Mode  Cnt       Score      Error  Units
c.e.MyBenchmark.testMethod        thrpt   10  987654.321 ± 1234.56  ops/s
`
	floor := 500000.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "jmh_ops_s", Floor: &floor, Unit: "ops/s"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 987654.321 {
		t.Errorf("value = %f, want 987654.321", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass")
	}
}

func TestExtractHyperfine_Ms(t *testing.T) {
	stdout := `Benchmark 1: sleep 0.1
  Time (mean ± σ):     102.3 ms ±   1.2 ms    [User: 1.0 ms, System: 1.0 ms]
  Range (min … max):   100.1 ms … 105.5 ms    10 runs
`
	ceil := 200.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "hyperfine_ms", Ceiling: &ceil, Unit: "ms"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 102.3 {
		t.Errorf("value = %f, want 102.3", results[0].Value)
	}
}

func TestExtractHyperfine_S(t *testing.T) {
	stdout := `Benchmark 1: make build
  Time (mean ± σ):      2.345 s ±  0.123 s    [User: 1.234 s, System: 0.567 s]
  Range (min … max):    2.100 s …  2.600 s    10 runs
`
	ceil := 5.0
	results := ExtractMetrics("", stdout, []MetricThreshold{
		{Name: "hyperfine_s", Ceiling: &ceil, Unit: "s"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 2.345 {
		t.Errorf("value = %f, want 2.345", results[0].Value)
	}
}

func TestExtractFromJSONNestedPath(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"estimates": {"median": {"point_estimate": 1500.0, "confidence_interval": {"lower": 1400.0, "upper": 1600.0}}}}`
	os.WriteFile(filepath.Join(dir, "estimates.json"), []byte(jsonContent), 0644)

	ceil := 2000.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "deep_nested", Ceiling: &ceil, Unit: "ns", Source: "estimates.json", JSONPath: "estimates.median.point_estimate"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 1500.0 {
		t.Errorf("value = %f, want 1500.0", results[0].Value)
	}
}

func TestExtractFromJSONMissingFile(t *testing.T) {
	dir := t.TempDir()
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "missing", Source: "nonexistent.json", JSONPath: "foo.bar"},
	})
	if len(results) != 0 {
		t.Errorf("expected 0 results for missing file, got %d", len(results))
	}
}

func TestExtractFromJSONInvalidPath(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"foo": {"bar": 42.0}}`
	os.WriteFile(filepath.Join(dir, "test.json"), []byte(jsonContent), 0644)

	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "bad_path", Source: "test.json", JSONPath: "foo.nonexistent.deep"},
	})
	if len(results) != 0 {
		t.Errorf("expected 0 results for invalid path, got %d", len(results))
	}
}

func TestExtractFromJSONMultipleArrayElements(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{
		"benchmarks": [
			{"stats": {"mean": 0.001, "median": 0.0009}},
			{"stats": {"mean": 0.005, "median": 0.004}},
			{"stats": {"mean": 0.002, "median": 0.0019}}
		]
	}`
	os.WriteFile(filepath.Join(dir, "pytest.json"), []byte(jsonContent), 0644)

	ceil := 0.01
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "pytest_mean", Ceiling: &ceil, Unit: "s", Source: "pytest.json", JSONPath: "benchmarks[*].stats.mean"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 0.005 {
		t.Errorf("value = %f, want 0.005 (max of array)", results[0].Value)
	}
}

func TestExtractFromJSONTopLevelArray(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `[{"time": 100.0}, {"time": 200.0}, {"time": 150.0}]`
	os.WriteFile(filepath.Join(dir, "arr.json"), []byte(jsonContent), 0644)

	ceil := 300.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "top_arr", Ceiling: &ceil, Unit: "ns", Source: "arr.json", JSONPath: "[*].time"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 200.0 {
		t.Errorf("value = %f, want 200.0", results[0].Value)
	}
}

func TestExtractCriterionFromRegistry(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"estimates": {"median": {"point_estimate": 4567.89}}}`
	os.WriteFile(filepath.Join(dir, "estimates.json"), []byte(jsonContent), 0644)

	ceil := 5000.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "criterion_ns", Ceiling: &ceil, Unit: "ns", Source: "estimates.json"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 4567.89 {
		t.Errorf("value = %f, want 4567.89", results[0].Value)
	}
	if !results[0].Pass {
		t.Error("expected pass")
	}
}

func TestExtractPytestMeanFromRegistry(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"benchmarks": [{"stats": {"mean": 0.001}}, {"stats": {"mean": 0.003}}]}`
	os.WriteFile(filepath.Join(dir, "pytest.json"), []byte(jsonContent), 0644)

	ceil := 0.01
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "pytest_mean_ns", Ceiling: &ceil, Unit: "ns", Source: "pytest.json"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 0.003 {
		t.Errorf("value = %f, want 0.003 (max)", results[0].Value)
	}
}

func TestExtractGBenchFromRegistry(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"benchmarks": [{"name": "BM_Sort", "real_time": 12345.0}, {"name": "BM_Search", "real_time": 678.0}]}`
	os.WriteFile(filepath.Join(dir, "gbench.json"), []byte(jsonContent), 0644)

	ceil := 20000.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "gbench_ns", Ceiling: &ceil, Unit: "ns", Source: "gbench.json"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 12345.0 {
		t.Errorf("value = %f, want 12345.0", results[0].Value)
	}
}

func TestExtractDotNetFromRegistry(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"Benchmarks": [{"Statistics": {"Median": 500.5}}, {"Statistics": {"Median": 750.3}}]}`
	os.WriteFile(filepath.Join(dir, "dotnet.json"), []byte(jsonContent), 0644)

	ceil := 1000.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "dotnet_ns", Ceiling: &ceil, Unit: "ns", Source: "dotnet.json"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != 750.3 {
		t.Errorf("value = %f, want 750.3", results[0].Value)
	}
}

func TestCustomJSONPathOverridesRegistry(t *testing.T) {
	dir := t.TempDir()
	jsonContent := `{"estimates": {"median": {"point_estimate": 1000.0}, "mean": {"point_estimate": 2000.0}}}`
	os.WriteFile(filepath.Join(dir, "estimates.json"), []byte(jsonContent), 0644)

	ceil := 3000.0
	results := ExtractMetrics(dir, "", []MetricThreshold{
		{Name: "criterion_ns", Ceiling: &ceil, Unit: "ns",
			Source:   "estimates.json",
			JSONPath: "estimates.mean.point_estimate"},
	})
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	// Custom json_path should override the registry's default path
	if results[0].Value != 2000.0 {
		t.Errorf("value = %f, want 2000.0 (custom json_path should override registry)", results[0].Value)
	}
}

func TestDiscoverParsesSourceAndJSONPath(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	rulesDir := filepath.Join(mosDir, "rules", "mechanical")
	os.MkdirAll(rulesDir, 0755)

	ruleContent := `rule "json-metric-test" {
  enforcement = "error"
  harness {
    command = "echo done"
    metric "criterion_ns" {
      ceiling = 5000
      unit = "ns"
      source = "target/criterion/bench/new/estimates.json"
      json_path = "estimates.median.point_estimate"
    }
  }
}
`
	os.WriteFile(filepath.Join(rulesDir, "json-test.mos"), []byte(ruleContent), 0644)

	specs, err := Discover(mosDir)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}
	if len(specs[0].Thresholds) != 1 {
		t.Fatalf("expected 1 threshold, got %d", len(specs[0].Thresholds))
	}
	th := specs[0].Thresholds[0]
	if th.Source != "target/criterion/bench/new/estimates.json" {
		t.Errorf("source = %q, want criterion path", th.Source)
	}
	if th.JSONPath != "estimates.median.point_estimate" {
		t.Errorf("json_path = %q, want estimates.median.point_estimate", th.JSONPath)
	}
}

func TestListExtractors(t *testing.T) {
	names := ListExtractors()
	if len(names) < 3 {
		t.Errorf("expected at least 3 built-in extractors, got %d", len(names))
	}
}
