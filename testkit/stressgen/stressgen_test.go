package stressgen

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/governance/audit"
	"github.com/dpopsuev/mos/moslib/linter"
	"github.com/dpopsuev/mos/moslib/lsp"
	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/history"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
	"github.com/dpopsuev/mos/moslib/vcs/transport"
	"github.com/dpopsuev/mos/testkit/forge"
)

// --- Functional tests ---

func TestGenerateKubernetesSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	small := Profile{
		Name:              "k8s-smoke",
		Rules:             10,
		ActiveContracts:   5,
		ArchivedContracts: 5,
		ResolutionLayers:  3,
		LexiconTerms:      20,
		IncludeFiles:      5,
	}

	root := t.TempDir()
	if err := Generate(root, small); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	var parseCount int
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".mos" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := dsl.Parse(string(data), nil); err != nil {
			t.Errorf("parse %s: %v", path, err)
		}
		parseCount++
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	expected := small.Rules + small.ActiveContracts + small.ArchivedContracts + 4
	if parseCount != expected {
		t.Errorf("parsed %d files, expected %d", parseCount, expected)
	}

	l := &linter.Linter{}
	diags, lintErr := l.Lint(root)
	if lintErr != nil {
		t.Fatalf("Lint: %v", lintErr)
	}

	for _, d := range diags {
		if d.Severity == linter.SeverityError {
			t.Errorf("unexpected error diagnostic: %s [%s] %s", d.File, d.Rule, d.Message)
		}
	}
}

// --- Kubernetes-scale benchmarks ---

func benchSetupKubernetes(b *testing.B) string {
	b.Helper()
	root := b.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		b.Fatalf("Generate: %v", err)
	}
	return root
}

func BenchmarkLoadContextKubernetes(b *testing.B) {
	root := benchSetupKubernetes(b)
	mosDir := root + "/.mos"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			b.Fatalf("LoadContext: %v", err)
		}
		if len(ctx.RuleIDs) != Kubernetes.Rules {
			b.Fatalf("expected %d rules, got %d", Kubernetes.Rules, len(ctx.RuleIDs))
		}
	}
}

func BenchmarkLintFullSweepKubernetes(b *testing.B) {
	root := benchSetupKubernetes(b)
	l := &linter.Linter{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := l.Lint(root)
		if err != nil {
			b.Fatalf("Lint: %v", err)
		}
	}
}

func BenchmarkCompletionResponseKubernetes(b *testing.B) {
	path := "/project/.mos/rules/mechanical/test.mos"
	content := "rule \"test\" {\n"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		items := lsp.Complete(path, content, 1, 0)
		if len(items) == 0 {
			b.Fatal("expected completion items")
		}
	}
}

// --- Linux Kernel scale benchmarks ---

func benchSetupLinuxKernel(b *testing.B) string {
	b.Helper()
	root := b.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		b.Fatalf("Generate: %v", err)
	}
	return root
}

func BenchmarkLoadContextLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)
	mosDir := root + "/.mos"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			b.Fatalf("LoadContext: %v", err)
		}
		if len(ctx.RuleIDs) != LinuxKernel.Rules {
			b.Fatalf("expected %d rules, got %d", LinuxKernel.Rules, len(ctx.RuleIDs))
		}
	}
}

func BenchmarkLintFullSweepLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)
	l := &linter.Linter{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := l.Lint(root)
		if err != nil {
			b.Fatalf("Lint: %v", err)
		}
	}
}

func BenchmarkDiagnosticLatencyLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)
	mosDir := root + "/.mos"

	ctx, err := linter.LoadContext(mosDir)
	if err != nil {
		b.Fatalf("LoadContext: %v", err)
	}

	var sampleRule string
	for _, path := range ctx.RuleIDs {
		sampleRule = path
		break
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = linter.ValidateRuleFile(sampleRule, ctx)
	}
}

// --- DSL parse benchmarks ---

func parseDSLDir(b *testing.B, root string) {
	b.Helper()
	mosDir := filepath.Join(root, ".mos")
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		if filepath.Ext(path) != ".mos" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := dsl.Parse(string(data), nil); err != nil {
			b.Fatalf("parse %s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		b.Fatalf("Walk: %v", err)
	}
}

func BenchmarkDSLParseKubernetes(b *testing.B) {
	root := b.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		b.Fatalf("Generate: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDSLDir(b, root)
	}
}

func BenchmarkDSLParseLinuxKernel(b *testing.B) {
	root := b.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		b.Fatalf("Generate: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseDSLDir(b, root)
	}
}

// --- Time budget assertion tests ---

func TestTimeBudgetLoadContextKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	start := time.Now()
	ctx, err := linter.LoadContext(root + "/.mos")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if len(ctx.RuleIDs) != Kubernetes.Rules {
		t.Errorf("expected %d rules, got %d", Kubernetes.Rules, len(ctx.RuleIDs))
	}

	budget := 1 * time.Second
	if elapsed > budget {
		t.Errorf("LoadContext took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("LoadContext completed in %v (budget: %v)", elapsed, budget)
	}
}

func TestTimeBudgetLoadContextLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	start := time.Now()
	ctx, err := linter.LoadContext(root + "/.mos")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if len(ctx.RuleIDs) != LinuxKernel.Rules {
		t.Errorf("expected %d rules, got %d", LinuxKernel.Rules, len(ctx.RuleIDs))
	}

	budget := 2 * time.Second
	if elapsed > budget {
		t.Errorf("LoadContext took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("LoadContext completed in %v (budget: %v)", elapsed, budget)
	}
}

// --- Fmt benchmarks ---

func fmtDSLDir(b *testing.B, root string) {
	b.Helper()
	mosDir := filepath.Join(root, ".mos")
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".mos" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := dsl.Parse(string(data), nil)
		if err != nil {
			b.Fatalf("parse %s: %v", path, err)
		}
		_ = dsl.Format(parsed, nil)
		return nil
	})
	if err != nil {
		b.Fatalf("Walk: %v", err)
	}
}

func BenchmarkFmtKubernetes(b *testing.B) {
	root := benchSetupKubernetes(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmtDSLDir(b, root)
	}
}

func BenchmarkFmtLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmtDSLDir(b, root)
	}
}

// --- Query benchmarks ---

func BenchmarkQueryKubernetes(b *testing.B) {
	root := benchSetupKubernetes(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := artifact.QueryArtifacts(root, artifact.QueryOpts{})
		if err != nil {
			b.Fatalf("QueryArtifacts: %v", err)
		}
	}
}

func BenchmarkQueryLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := artifact.QueryArtifacts(root, artifact.QueryOpts{})
		if err != nil {
			b.Fatalf("QueryArtifacts: %v", err)
		}
	}
}

// --- Status (audit) benchmarks ---

func BenchmarkStatusKubernetes(b *testing.B) {
	root := benchSetupKubernetes(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := audit.RunAudit(root, audit.AuditOpts{})
		if err != nil {
			b.Fatalf("RunAudit: %v", err)
		}
	}
}

func BenchmarkStatusLinuxKernel(b *testing.B) {
	root := benchSetupLinuxKernel(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := audit.RunAudit(root, audit.AuditOpts{})
		if err != nil {
			b.Fatalf("RunAudit: %v", err)
		}
	}
}

// --- Fmt time budget tests ---

func timeBudgetFmt(t *testing.T, root string, budget time.Duration) {
	t.Helper()
	mosDir := filepath.Join(root, ".mos")
	start := time.Now()
	var fileCount int
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".mos" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		parsed, err := dsl.Parse(string(data), nil)
		if err != nil {
			return nil
		}
		_ = dsl.Format(parsed, nil)
		fileCount++
		return nil
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
	if elapsed > budget {
		t.Errorf("fmt took %v, budget is %v (%d files)", elapsed, budget, fileCount)
	} else {
		t.Logf("fmt completed in %v (budget: %v, %d files)", elapsed, budget, fileCount)
	}
}

func TestTimeBudgetFmtKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	timeBudgetFmt(t, root, 2*time.Second)
}

func TestTimeBudgetFmtLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	timeBudgetFmt(t, root, 4*time.Second)
}

// --- Query time budget tests ---

func TestTimeBudgetQueryKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	start := time.Now()
	results, err := artifact.QueryArtifacts(root, artifact.QueryOpts{})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	budget := 1 * time.Second
	if elapsed > budget {
		t.Errorf("query took %v, budget is %v (%d results)", elapsed, budget, len(results))
	} else {
		t.Logf("query completed in %v (budget: %v, %d results)", elapsed, budget, len(results))
	}
}

func TestTimeBudgetQueryLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	start := time.Now()
	results, err := artifact.QueryArtifacts(root, artifact.QueryOpts{})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("QueryArtifacts: %v", err)
	}
	budget := 2 * time.Second
	if elapsed > budget {
		t.Errorf("query took %v, budget is %v (%d results)", elapsed, budget, len(results))
	} else {
		t.Logf("query completed in %v (budget: %v, %d results)", elapsed, budget, len(results))
	}
}

// --- Status (audit) time budget tests ---

func TestTimeBudgetStatusKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	start := time.Now()
	report, err := audit.RunAudit(root, audit.AuditOpts{})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	budget := 3 * time.Second
	if elapsed > budget {
		t.Errorf("status took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("status completed in %v (budget: %v, lint_errors=%d, lint_warnings=%d)",
			elapsed, budget, report.LintErrors, report.LintWarnings)
	}
}

func TestTimeBudgetStatusLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := Generate(root, LinuxKernel); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	start := time.Now()
	report, err := audit.RunAudit(root, audit.AuditOpts{})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("RunAudit: %v", err)
	}
	budget := 5 * time.Second
	if elapsed > budget {
		t.Errorf("status took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("status completed in %v (budget: %v, lint_errors=%d, lint_warnings=%d)",
			elapsed, budget, report.LintErrors, report.LintWarnings)
	}
}

// =============================================================================
// Dunix — Cross-planetary distributed microkernel OS (Terra + Mars)
// =============================================================================
//
// Dunix is a hypothetical Distributed Unix descended from a cross-breed of
// the Linux kernel and Kubernetes, operating across Terra and Mars with
// light-speed communication delays of 4-24 minutes. Its governance uses
// two human-protocol languages:
//
//   Terran English — standard keywords (feature, rule, scenario, ...)
//   Martian English — localized keywords (capability, mandate, protocol, ...)
//
// The Martian dialect evolved under the influence of capability-based
// microkernel culture (seL4, L4), high-latency protocol-centric comms,
// and Mars Habitat Authority governance norms.

var dunixSmoke = Profile{
	Name:              "dunix-smoke",
	Rules:             20,
	ActiveContracts:   10,
	ArchivedContracts: 10,
	ResolutionLayers:  10,
	LexiconTerms:      50,
	IncludeFiles:      10,
}

// TestDunixTerranSmoke verifies Terran English generation at reduced scale.
func TestDunixTerranSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized(Terran): %v", err)
	}

	assertAllCstFilesParse(t, root, nil)
	assertLintClean(t, root)
}

// TestDunixMartianSmoke verifies Martian English generation at reduced scale.
// Files use localized keywords (mandate, charter, capability, protocol, etc.)
// but parse correctly via the lexicon-driven two-phase flow.
func TestDunixMartianSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized(Martian): %v", err)
	}

	ctx, err := linter.LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	if ctx.Keywords == nil {
		t.Fatal("expected non-nil KeywordMap from Martian lexicon")
	}
	if got := ctx.Keywords.ToMachine["capability"]; got != "feature" {
		t.Errorf("expected capability->feature, got %q", got)
	}
	if got := ctx.Keywords.ToMachine["mandate"]; got != "rule" {
		t.Errorf("expected mandate->rule, got %q", got)
	}

	if len(ctx.RuleIDs) != dunixSmoke.Rules {
		t.Errorf("expected %d rules, got %d", dunixSmoke.Rules, len(ctx.RuleIDs))
	}
	if len(ctx.ContractIDs) != dunixSmoke.ActiveContracts+dunixSmoke.ArchivedContracts {
		t.Errorf("expected %d contracts, got %d",
			dunixSmoke.ActiveContracts+dunixSmoke.ArchivedContracts, len(ctx.ContractIDs))
	}

	assertAllCstFilesParse(t, root, ctx.Keywords)
}

// TestDunixMartianKeywordRoundTrip verifies that Martian-keyword files
// produce AST nodes with machine-protocol kinds (rule, contract, etc.)
// even though the source uses localized keywords (mandate, charter, etc.).
func TestDunixMartianKeywordRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	ctx, err := linter.LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	rulePath := ""
	for _, p := range ctx.RuleIDs {
		rulePath = p
		break
	}
	if rulePath == "" {
		t.Fatal("no rules found")
	}

	data, err := os.ReadFile(rulePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	src := string(data)
	if !strings.Contains(src, "mandate") {
		t.Error("expected Martian keyword 'mandate' in rule source")
	}

	f, err := dsl.Parse(src, ctx.Keywords)
	if err != nil {
		t.Fatalf("Parse with Martian keywords: %v", err)
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		t.Fatal("expected ArtifactBlock")
	}
	if ab.Kind != "rule" {
		t.Errorf("expected machine Kind 'rule', got %q", ab.Kind)
	}
}

// TestDunixLayerNames verifies Dunix-specific resolution layer names.
func TestDunixLayerNames(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	ctx, err := linter.LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	expectedLayers := []string{
		"kernel", "subsystem", "module", "driver", "habitat",
		"planet", "orbital", "federation", "organization", "interplanetary",
	}
	for _, name := range expectedLayers {
		if !ctx.LayerSet[name] {
			t.Errorf("missing expected layer %q", name)
		}
	}
}

// --- Dunix time budget tests ---

func TestTimeBudgetDunixTerran(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := GenerateLocalized(root, Dunix, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	start := time.Now()
	ctx, err := linter.LoadContext(root + "/.mos")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if len(ctx.RuleIDs) != Dunix.Rules {
		t.Errorf("expected %d rules, got %d", Dunix.Rules, len(ctx.RuleIDs))
	}
	if len(ctx.Layers) != Dunix.ResolutionLayers {
		t.Errorf("expected %d layers, got %d", Dunix.ResolutionLayers, len(ctx.Layers))
	}

	budget := 5 * time.Second
	if elapsed > budget {
		t.Errorf("LoadContext took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("Dunix Terran LoadContext: %v (budget: %v, rules: %d, contracts: %d)",
			elapsed, budget, len(ctx.RuleIDs), len(ctx.ContractIDs))
	}
}

func TestTimeBudgetDunixMartian(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := GenerateLocalized(root, Dunix, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	start := time.Now()
	ctx, err := linter.LoadContext(root + "/.mos")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	if len(ctx.RuleIDs) != Dunix.Rules {
		t.Errorf("expected %d rules, got %d", Dunix.Rules, len(ctx.RuleIDs))
	}

	if ctx.Keywords.ToMachine["mandate"] != "rule" {
		t.Error("expected Martian keyword mapping in loaded context")
	}

	budget := 5 * time.Second
	if elapsed > budget {
		t.Errorf("LoadContext took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("Dunix Martian LoadContext: %v (budget: %v, rules: %d, contracts: %d)",
			elapsed, budget, len(ctx.RuleIDs), len(ctx.ContractIDs))
	}
}

// --- Dunix benchmarks ---

func benchSetupDunix(b *testing.B, loc Locale) string {
	b.Helper()
	root := b.TempDir()
	if err := GenerateLocalized(root, Dunix, loc); err != nil {
		b.Fatalf("GenerateLocalized: %v", err)
	}
	return root
}

func BenchmarkLoadContextDunixTerran(b *testing.B) {
	root := benchSetupDunix(b, TerranEnglish)
	mosDir := root + "/.mos"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			b.Fatalf("LoadContext: %v", err)
		}
		if len(ctx.RuleIDs) != Dunix.Rules {
			b.Fatalf("expected %d rules, got %d", Dunix.Rules, len(ctx.RuleIDs))
		}
	}
}

func BenchmarkLoadContextDunixMartian(b *testing.B) {
	root := benchSetupDunix(b, MartianEnglish)
	mosDir := root + "/.mos"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, err := linter.LoadContext(mosDir)
		if err != nil {
			b.Fatalf("LoadContext: %v", err)
		}
		if len(ctx.RuleIDs) != Dunix.Rules {
			b.Fatalf("expected %d rules, got %d", Dunix.Rules, len(ctx.RuleIDs))
		}
	}
}

func BenchmarkDSLParseDunixMartian(b *testing.B) {
	root := b.TempDir()
	if err := GenerateLocalized(root, Dunix, MartianEnglish); err != nil {
		b.Fatalf("GenerateLocalized: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	vocabData, err := os.ReadFile(filepath.Join(mosDir, "lexicon", "default.mos"))
	if err != nil {
		b.Fatalf("ReadFile lexicon: %v", err)
	}
	kw, err := dsl.LoadKeywords(string(vocabData))
	if err != nil {
		b.Fatalf("LoadKeywords: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() || filepath.Ext(path) != ".mos" {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := dsl.Parse(string(data), kw); err != nil {
				b.Fatalf("parse %s: %v", path, err)
			}
			return nil
		})
	}
}

// --- Nested contract tests ---

func TestDunixNestedContractsSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	var umbrella, flat, totalFiles int
	for _, sub := range []string{"active", "archive"} {
		contractsDir := filepath.Join(mosDir, "contracts", sub)
		entries, err := os.ReadDir(contractsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			totalFiles++
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			hasNested := false
			for _, item := range ab.Items {
				if blk, ok := item.(*dsl.Block); ok && blk.Name == "contract" {
					hasNested = true
					break
				}
			}
			if hasNested {
				umbrella++
			} else {
				flat++
			}
		}
	}

	if umbrella == 0 {
		t.Error("expected at least one umbrella contract-of-contracts")
	}
	if flat == 0 {
		t.Error("expected at least one flat contract")
	}
	t.Logf("contracts: %d total (%d umbrella, %d flat)", totalFiles, umbrella, flat)
}

func TestDunixNestedContractsMartianParse(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	ctx, err := linter.LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	// Pick an umbrella contract and verify its nested structure parses with Martian keywords
	var umbrellaPath string
	for _, sub := range []string{"active", "archive"} {
		contractsDir := filepath.Join(root, ".mos", "contracts", sub)
		entries, _ := os.ReadDir(contractsDir)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, _ := os.ReadFile(path)
			if strings.Contains(string(data), "ordering") {
				umbrellaPath = path
				break
			}
		}
		if umbrellaPath != "" {
			break
		}
	}

	if umbrellaPath == "" {
		t.Fatal("no umbrella contract found in generated output")
	}

	data, err := os.ReadFile(umbrellaPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	src := string(data)
	if !strings.Contains(src, "charter") {
		t.Error("expected Martian keyword 'charter' in umbrella contract")
	}

	f, err := dsl.Parse(src, ctx.Keywords)
	if err != nil {
		t.Fatalf("Parse umbrella with Martian keywords: %v", err)
	}

	ab := f.Artifact.(*dsl.ArtifactBlock)
	if ab.Kind != "contract" {
		t.Errorf("umbrella Kind = %q, want 'contract' (machine keyword)", ab.Kind)
	}

	var nestedCount int
	for _, item := range ab.Items {
		if blk, ok := item.(*dsl.Block); ok && blk.Name == "charter" {
			nestedCount++
			// Verify each nested charter has a depends_on field
			hasDeps := false
			for _, bi := range blk.Items {
				if fld, ok := bi.(*dsl.Field); ok && fld.Key == "depends_on" {
					hasDeps = true
					break
				}
			}
			if !hasDeps {
				t.Errorf("nested charter %q missing depends_on", blk.Title)
			}
		}
	}

	if nestedCount != 3 {
		t.Errorf("expected 3 nested charters, got %d", nestedCount)
	}
	t.Logf("umbrella contract has %d nested charters with dependency DAG", nestedCount)
}

func TestDunixDependencyDAGTopology(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	var linearDeps, diamondDeps, rootContracts int

	for _, sub := range []string{"active", "archive"} {
		contractsDir := filepath.Join(mosDir, "contracts", sub)
		entries, err := os.ReadDir(contractsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			classifyDeps(ab.Items, &linearDeps, &diamondDeps, &rootContracts)

			for _, item := range ab.Items {
				if blk, ok := item.(*dsl.Block); ok && blk.Name == "contract" {
					classifyDeps(blk.Items, &linearDeps, &diamondDeps, &rootContracts)
				}
			}
		}
	}

	if rootContracts == 0 {
		t.Error("expected root contracts (depends_on = [])")
	}
	if linearDeps == 0 {
		t.Error("expected linear dependencies (depends_on with 1 item)")
	}
	if diamondDeps == 0 {
		t.Error("expected diamond/non-linear dependencies (depends_on with 2+ items)")
	}

	t.Logf("dependency topology: %d root, %d linear, %d diamond/non-linear",
		rootContracts, linearDeps, diamondDeps)
}

func classifyDeps(items []dsl.Node, linear, diamond, root *int) {
	for _, item := range items {
		fld, ok := item.(*dsl.Field)
		if !ok || fld.Key != "depends_on" {
			continue
		}
		lv, ok := fld.Value.(*dsl.ListVal)
		if !ok {
			continue
		}
		switch len(lv.Items) {
		case 0:
			*root++
		case 1:
			*linear++
		default:
			*diamond++
		}
	}
}

func TestDunixTrackerBlocks(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	var withGH, withJira, withBoth, total int

	for _, sub := range []string{"active", "archive"} {
		contractsDir := filepath.Join(mosDir, "contracts", sub)
		entries, err := os.ReadDir(contractsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			total++
			gh, jira := countTrackers(ab.Items)
			if gh > 0 {
				withGH++
			}
			if jira > 0 {
				withJira++
			}
			if gh > 0 && jira > 0 {
				withBoth++
			}
		}
	}

	if withGH == 0 {
		t.Error("expected at least one contract with GitHub tracker")
	}
	if withJira == 0 {
		t.Error("expected at least one contract with Jira tracker")
	}
	if withBoth == 0 {
		t.Error("expected at least one contract with both GitHub and Jira trackers")
	}
	t.Logf("tracker coverage: %d/%d GitHub, %d/%d Jira, %d/%d both",
		withGH, total, withJira, total, withBoth, total)
}

func TestDunixMartianTrackerBlocks(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	ctx, err := linter.LoadContext(filepath.Join(root, ".mos"))
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}

	// Pick a contract file and verify trackers survive Martian two-phase parse
	for _, sub := range []string{"active"} {
		contractsDir := filepath.Join(root, ".mos", "contracts", sub)
		entries, _ := os.ReadDir(contractsDir)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, _ := os.ReadFile(path)
			f, err := dsl.Parse(string(data), ctx.Keywords)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			gh, jira := countTrackers(ab.Items)
			if gh > 0 || jira > 0 {
				t.Logf("Martian contract %s: %d github, %d jira trackers",
					ab.Name, gh, jira)
				return
			}
		}
	}
	t.Error("no Martian contracts with tracker blocks found")
}

func TestDunixUmbrellaTrackerHierarchy(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	if err := GenerateLocalized(root, dunixSmoke, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	mosDir := filepath.Join(root, ".mos")
	for _, sub := range []string{"active", "archive"} {
		contractsDir := filepath.Join(mosDir, "contracts", sub)
		entries, err := os.ReadDir(contractsDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractsDir, e.Name(), "contract.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			// Only check umbrella contracts
			var nested []*dsl.Block
			for _, item := range ab.Items {
				if blk, ok := item.(*dsl.Block); ok && blk.Name == "contract" {
					nested = append(nested, blk)
				}
			}
			if len(nested) == 0 {
				continue
			}

			// Parent must have tracker with type=epic
			parentGH, parentJira := countTrackers(ab.Items)
			if parentGH == 0 && parentJira == 0 {
				t.Errorf("umbrella %s has no tracker blocks", ab.Name)
			}

			// Each nested sub-contract must have a Jira tracker with epic reference
			for _, sub := range nested {
				_, subJira := countTrackers(sub.Items)
				if subJira == 0 {
					t.Errorf("nested sub-contract %q has no Jira tracker", sub.Title)
				}

				for _, item := range sub.Items {
					blk, ok := item.(*dsl.Block)
					if !ok || blk.Name != "tracker" || blk.Title != "jira" {
						continue
					}
					hasEpic := false
					for _, bi := range blk.Items {
						if fld, ok := bi.(*dsl.Field); ok && fld.Key == "epic" {
							hasEpic = true
						}
					}
					if !hasEpic {
						t.Errorf("nested sub-contract %q: Jira tracker missing epic field", sub.Title)
					}
				}
			}

			t.Logf("umbrella %s: parent has %d GH + %d Jira trackers, %d nested subs with Jira epic refs",
				ab.Name, parentGH, parentJira, len(nested))
			return
		}
	}
	t.Error("no umbrella contract found")
}

func countTrackers(items []dsl.Node) (github, jira int) {
	for _, item := range items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "tracker" {
			continue
		}
		switch blk.Title {
		case "github":
			github++
		case "jira":
			jira++
		}
	}
	return
}

// --- Federation model tests ---

func TestDunixUpstreamBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	p := Profile{
		Name: "dunix-upstream-test", Rules: 10, ActiveContracts: 5,
		ArchivedContracts: 5, ResolutionLayers: 3, LexiconTerms: 50, IncludeFiles: 5,
	}

	if err := GenerateLocalized(root, p, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".mos", "config.mos"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	doc, err := dsl.Parse(string(data), nil)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}

	ab, ok := doc.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		t.Fatal("config.mos artifact is not an ArtifactBlock")
	}

	var foundUpstream bool
	var upstreamScope, upstreamURL string
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "upstream" {
			continue
		}
		foundUpstream = true
		if blk.Title != "dunix-federation" {
			t.Errorf("upstream title = %q, want %q", blk.Title, "dunix-federation")
		}
		for _, sub := range blk.Items {
			f, ok := sub.(*dsl.Field)
			if !ok {
				continue
			}
			switch f.Key {
			case "scope":
				upstreamScope = dsl.StringValue(f.Value)
			case "url":
				upstreamURL = dsl.StringValue(f.Value)
			}
		}
	}
	if !foundUpstream {
		t.Fatal("config.mos missing upstream block")
	}
	if upstreamScope != "interplanetary" {
		t.Errorf("upstream scope = %q, want %q", upstreamScope, "interplanetary")
	}
	if !strings.Contains(upstreamURL, "dunix/federation-mos") {
		t.Errorf("upstream url = %q, missing expected repo path", upstreamURL)
	}

	var govScope string
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "governance" {
			continue
		}
		for _, sub := range blk.Items {
			f, ok := sub.(*dsl.Field)
			if !ok {
				continue
			}
			if f.Key == "scope" {
				govScope = dsl.StringValue(f.Value)
			}
		}
	}
	if govScope != "planet" {
		t.Errorf("governance scope = %q, want %q", govScope, "planet")
	}
	t.Logf("upstream: scope=%s url=%s; governance: scope=%s", upstreamScope, upstreamURL, govScope)
}

func TestDunixMartianUpstreamBlock(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	p := Profile{
		Name: "dunix-martian-upstream-test", Rules: 10, ActiveContracts: 5,
		ArchivedContracts: 5, ResolutionLayers: 3, LexiconTerms: 50, IncludeFiles: 5,
	}

	if err := GenerateLocalized(root, p, MartianEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	vocData, err := os.ReadFile(filepath.Join(root, ".mos", "lexicon", "default.mos"))
	if err != nil {
		t.Fatalf("read lexicon: %v", err)
	}
	vocDoc, err := dsl.Parse(string(vocData), nil)
	if err != nil {
		t.Fatalf("parse lexicon: %v", err)
	}
	kw := dsl.ExtractKeywords(vocDoc)

	data, err := os.ReadFile(filepath.Join(root, ".mos", "config.mos"))
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	doc, err := dsl.Parse(string(data), kw)
	if err != nil {
		t.Fatalf("parse config with Martian vocab: %v", err)
	}

	ab, ok := doc.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		t.Fatal("config.mos artifact is not an ArtifactBlock")
	}
	if ab.Kind != "config" {
		t.Errorf("config artifact kind = %q, want %q (machine protocol)", ab.Kind, "config")
	}
	var foundUpstream bool
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "upstream" {
			continue
		}
		foundUpstream = true
	}
	if !foundUpstream {
		t.Fatal("Martian config.mos missing upstream block after two-phase parse")
	}
}

func TestDunixJurisdictionDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	p := Profile{
		Name: "dunix-jurisdiction-test", Rules: 30, ActiveContracts: 5,
		ArchivedContracts: 5, ResolutionLayers: 3, LexiconTerms: 50, IncludeFiles: 10,
	}

	if err := GenerateLocalized(root, p, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	jurisdictionCounts := map[string]int{}
	extendsCount := 0

	mosDir := filepath.Join(root, ".mos")
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".mos" {
			return err
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		doc, parseErr := dsl.Parse(string(data), nil)
		if parseErr != nil {
			return nil
		}
		ab, ok := doc.Artifact.(*dsl.ArtifactBlock)
		if !ok || ab.Kind != "rule" {
			return nil
		}
		for _, item := range ab.Items {
			f, ok := item.(*dsl.Field)
			if !ok {
				continue
			}
			if f.Key == "jurisdiction" {
				jurisdictionCounts[dsl.StringValue(f.Value)]++
			}
			if f.Key == "extends" {
				extendsCount++
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}

	expected := []string{
		"interplanetary", "planet", "division",
		"subsystem", "team", "collective", "cabinet",
	}
	for _, j := range expected {
		if jurisdictionCounts[j] == 0 {
			t.Errorf("no %s-jurisdiction rules found", j)
		}
	}
	if extendsCount == 0 {
		t.Error("no extends references found (rule inheritance)")
	}
	if len(jurisdictionCounts) != len(expected) {
		t.Errorf("expected %d jurisdiction tiers, got %d: %v", len(expected), len(jurisdictionCounts), jurisdictionCounts)
	}

	t.Logf("7-tier jurisdiction distribution (%d total rules): %v; extends=%d",
		p.Rules, jurisdictionCounts, extendsCount)
}

func TestDunixFederationDeclaration(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root := t.TempDir()
	p := Profile{
		Name: "dunix-decl-test", Rules: 5, ActiveContracts: 5,
		ArchivedContracts: 5, ResolutionLayers: 3, LexiconTerms: 50, IncludeFiles: 2,
	}

	if err := GenerateLocalized(root, p, TerranEnglish); err != nil {
		t.Fatalf("GenerateLocalized: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".mos", "declaration.mos"))
	if err != nil {
		t.Fatalf("read declaration: %v", err)
	}
	doc, err := dsl.Parse(string(data), nil)
	if err != nil {
		t.Fatalf("parse declaration: %v", err)
	}

	ab, ok := doc.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		t.Fatal("declaration.mos artifact is not an ArtifactBlock")
	}

	var foundFederation bool
	var overridePolicy string
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "federation" {
			continue
		}
		foundFederation = true
		for _, sub := range blk.Items {
			f, ok := sub.(*dsl.Field)
			if !ok {
				continue
			}
			if f.Key == "override_policy" {
				overridePolicy = dsl.StringValue(f.Value)
			}
		}
	}
	if !foundFederation {
		t.Fatal("declaration.mos missing federation block")
	}
	if overridePolicy != "extend-only" {
		t.Errorf("override_policy = %q, want %q", overridePolicy, "extend-only")
	}
	t.Logf("federation: override_policy=%s", overridePolicy)
}

// --- Dunix helpers ---

func assertAllCstFilesParse(t *testing.T, root string, kw *dsl.KeywordMap) {
	t.Helper()
	mosDir := filepath.Join(root, ".mos")
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".mos" {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := dsl.Parse(string(data), kw); err != nil {
			t.Errorf("parse %s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Walk: %v", err)
	}
}

func assertLintClean(t *testing.T, root string) {
	t.Helper()
	l := &linter.Linter{}
	diags, err := l.Lint(root)
	if err != nil {
		t.Fatalf("Lint: %v", err)
	}
	for _, d := range diags {
		if d.Severity == linter.SeverityError {
			t.Errorf("unexpected error diagnostic: %s [%s] %s", d.File, d.Rule, d.Message)
		}
	}
}

// =============================================================================
// VCS Benchmarks — SnapshotWorkingTree, BuildTree+Commit, DiffTrees
// =============================================================================

func benchSetupVCS(b *testing.B, p Profile) (string, *vcs.Repository) {
	b.Helper()
	root := b.TempDir()
	if err := Generate(root, p); err != nil {
		b.Fatalf("Generate: %v", err)
	}
	repo, err := vcs.InitRepo(root, "git")
	if err != nil {
		b.Fatalf("InitRepo: %v", err)
	}
	if err := staging.Add(repo, []string{"."}); err != nil {
		b.Fatalf("Add: %v", err)
	}
	if _, err := staging.Commit(repo, "bench", "bench@mos.dev", "baseline"); err != nil {
		b.Fatalf("Commit: %v", err)
	}
	return root, repo
}

func testSetupVCS(t *testing.T, p Profile) (string, *vcs.Repository) {
	t.Helper()
	root := t.TempDir()
	if err := Generate(root, p); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	repo, err := vcs.InitRepo(root, "git")
	if err != nil {
		t.Fatalf("InitRepo: %v", err)
	}
	if err := staging.Add(repo, []string{"."}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := staging.Commit(repo, "bench", "bench@mos.dev", "baseline"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	return root, repo
}

// --- Snapshot benchmarks ---

func BenchmarkVCSSnapshotKubernetes(b *testing.B) {
	root, repo := benchSetupVCS(b, Kubernetes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err := staging.SnapshotWorkingTree(root, repo.Store)
		if err != nil {
			b.Fatalf("SnapshotWorkingTree: %v", err)
		}
		if len(entries) == 0 {
			b.Fatal("empty snapshot")
		}
	}
}

func BenchmarkVCSSnapshotLinuxKernel(b *testing.B) {
	root, repo := benchSetupVCS(b, LinuxKernel)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entries, err := staging.SnapshotWorkingTree(root, repo.Store)
		if err != nil {
			b.Fatalf("SnapshotWorkingTree: %v", err)
		}
		if len(entries) == 0 {
			b.Fatal("empty snapshot")
		}
	}
}

// --- Commit benchmarks (BuildTree + StoreCommit) ---

func BenchmarkVCSCommitKubernetes(b *testing.B) {
	_, repo := benchSetupVCS(b, Kubernetes)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := staging.Add(repo, []string{"."}); err != nil {
			b.Fatalf("Add: %v", err)
		}
		if _, err := staging.Commit(repo, "bench", "bench@mos.dev", "iteration"); err != nil {
			b.Fatalf("Commit: %v", err)
		}
	}
}

func BenchmarkVCSCommitLinuxKernel(b *testing.B) {
	_, repo := benchSetupVCS(b, LinuxKernel)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := staging.Add(repo, []string{"."}); err != nil {
			b.Fatalf("Add: %v", err)
		}
		if _, err := staging.Commit(repo, "bench", "bench@mos.dev", "iteration"); err != nil {
			b.Fatalf("Commit: %v", err)
		}
	}
}

// --- Diff benchmarks ---

func BenchmarkVCSDiffKubernetes(b *testing.B) {
	root, repo := benchSetupVCS(b, Kubernetes)

	idx, err := staging.LoadIndex(root)
	if err != nil {
		b.Fatalf("LoadIndex: %v", err)
	}
	baseMap := staging.IndexToMap(idx)

	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	if err != nil {
		b.Fatalf("SnapshotWorkingTree: %v", err)
	}
	workMap := make(map[string]vcs.Hash, len(entries))
	for _, e := range entries {
		workMap[e.Path] = e.Hash
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diffs := staging.DiffTrees(baseMap, workMap)
		_ = diffs
	}
}

func BenchmarkVCSDiffLinuxKernel(b *testing.B) {
	root, repo := benchSetupVCS(b, LinuxKernel)

	idx, err := staging.LoadIndex(root)
	if err != nil {
		b.Fatalf("LoadIndex: %v", err)
	}
	baseMap := staging.IndexToMap(idx)

	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	if err != nil {
		b.Fatalf("SnapshotWorkingTree: %v", err)
	}
	workMap := make(map[string]vcs.Hash, len(entries))
	for _, e := range entries {
		workMap[e.Path] = e.Hash
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		diffs := staging.DiffTrees(baseMap, workMap)
		_ = diffs
	}
}

// --- VCS Time budget tests ---

func TestTimeBudgetVCSSnapshotKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, Kubernetes)

	start := time.Now()
	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("SnapshotWorkingTree: %v", err)
	}

	budget := 2 * time.Second
	if elapsed > budget {
		t.Errorf("SnapshotWorkingTree took %v, budget is %v (%d entries)", elapsed, budget, len(entries))
	} else {
		t.Logf("SnapshotWorkingTree completed in %v (budget: %v, %d entries)", elapsed, budget, len(entries))
	}
}

func TestTimeBudgetVCSSnapshotLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, LinuxKernel)

	start := time.Now()
	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("SnapshotWorkingTree: %v", err)
	}

	budget := 4 * time.Second
	if elapsed > budget {
		t.Errorf("SnapshotWorkingTree took %v, budget is %v (%d entries)", elapsed, budget, len(entries))
	} else {
		t.Logf("SnapshotWorkingTree completed in %v (budget: %v, %d entries)", elapsed, budget, len(entries))
	}
}

func TestTimeBudgetVCSCommitKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, Kubernetes)
	_ = root

	start := time.Now()
	if err := staging.Add(repo, []string{"."}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err := staging.Commit(repo, "bench", "bench@mos.dev", "budget-test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	budget := 3 * time.Second
	if elapsed > budget {
		t.Errorf("Add+Commit took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("Add+Commit completed in %v (budget: %v)", elapsed, budget)
	}
}

func TestTimeBudgetVCSCommitLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, LinuxKernel)
	_ = root

	start := time.Now()
	if err := staging.Add(repo, []string{"."}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err := staging.Commit(repo, "bench", "bench@mos.dev", "budget-test")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	budget := 5 * time.Second
	if elapsed > budget {
		t.Errorf("Add+Commit took %v, budget is %v", elapsed, budget)
	} else {
		t.Logf("Add+Commit completed in %v (budget: %v)", elapsed, budget)
	}
}

func TestTimeBudgetVCSDiffKubernetes(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, Kubernetes)

	idx, err := staging.LoadIndex(root)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	baseMap := staging.IndexToMap(idx)

	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	if err != nil {
		t.Fatalf("SnapshotWorkingTree: %v", err)
	}
	workMap := make(map[string]vcs.Hash, len(entries))
	for _, e := range entries {
		workMap[e.Path] = e.Hash
	}

	start := time.Now()
	diffs := staging.DiffTrees(baseMap, workMap)
	elapsed := time.Since(start)

	budget := 50 * time.Millisecond
	if elapsed > budget {
		t.Errorf("DiffTrees took %v, budget is %v (%d diffs)", elapsed, budget, len(diffs))
	} else {
		t.Logf("DiffTrees completed in %v (budget: %v, %d diffs)", elapsed, budget, len(diffs))
	}
}

func TestTimeBudgetVCSDiffLinuxKernel(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, LinuxKernel)

	idx, err := staging.LoadIndex(root)
	if err != nil {
		t.Fatalf("LoadIndex: %v", err)
	}
	baseMap := staging.IndexToMap(idx)

	entries, err := staging.SnapshotWorkingTree(root, repo.Store)
	if err != nil {
		t.Fatalf("SnapshotWorkingTree: %v", err)
	}
	workMap := make(map[string]vcs.Hash, len(entries))
	for _, e := range entries {
		workMap[e.Path] = e.Hash
	}

	start := time.Now()
	diffs := staging.DiffTrees(baseMap, workMap)
	elapsed := time.Since(start)

	budget := 100 * time.Millisecond
	if elapsed > budget {
		t.Errorf("DiffTrees took %v, budget is %v (%d diffs)", elapsed, budget, len(diffs))
	} else {
		t.Logf("DiffTrees completed in %v (budget: %v, %d diffs)", elapsed, budget, len(diffs))
	}
}

// =============================================================================
// Deep VCS Stress — history depth, incremental commits, clone transport
// =============================================================================

func collectMosFiles(root string) ([]string, error) {
	mosDir := filepath.Join(root, ".mos")
	var files []string
	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".mos") {
			rel, _ := filepath.Rel(root, path)
			files = append(files, rel)
		}
		return nil
	})
	return files, err
}

func mutateFiles(root string, paths []string) error {
	for _, p := range paths {
		abs := filepath.Join(root, p)
		f, err := os.OpenFile(abs, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(f, "\n# mutated %d\n", time.Now().UnixNano())
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func buildHistory(t testing.TB, root string, repo *vcs.Repository, numCommits, filesPerCommit int) string {
	t.Helper()
	allFiles, err := collectMosFiles(root)
	if err != nil {
		t.Fatalf("collectMosFiles: %v", err)
	}
	if len(allFiles) == 0 {
		t.Fatal("no .mos files found")
	}

	rng := rand.New(rand.NewSource(42))

	targetFile := allFiles[0]

	for i := 0; i < numCommits; i++ {
		picked := make([]string, 0, filesPerCommit)
		picked = append(picked, targetFile)

		perm := rng.Perm(len(allFiles))
		for _, idx := range perm {
			if len(picked) >= filesPerCommit {
				break
			}
			if allFiles[idx] != targetFile {
				picked = append(picked, allFiles[idx])
			}
		}

		if err := mutateFiles(root, picked); err != nil {
			t.Fatalf("mutateFiles (commit %d): %v", i, err)
		}
		if err := staging.Add(repo, picked); err != nil {
			t.Fatalf("Add (commit %d): %v", i, err)
		}
		if _, err := staging.Commit(repo, "stress", "stress@mos.dev", fmt.Sprintf("commit-%d", i)); err != nil {
			t.Fatalf("Commit (commit %d): %v", i, err)
		}
	}

	return targetFile
}

func dirSizeBytes(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// --- CON-2026-248: History Depth Stress ---

func TestVCSHistoryDepthLog(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	for _, depth := range []int{50, 500} {
		depth := depth
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			root := t.TempDir()
			if err := Generate(root, Kubernetes); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			repo, err := vcs.InitRepo(root, "git")
			if err != nil {
				t.Fatalf("InitRepo: %v", err)
			}
			if err := staging.Add(repo, []string{"."}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			if _, err := staging.Commit(repo, "stress", "stress@mos.dev", "baseline"); err != nil {
				t.Fatalf("Commit: %v", err)
			}

			buildHistory(t, root, repo, depth, 5)

			head, err := vcs.ResolveHead(repo)
			if err != nil {
				t.Fatalf("ResolveHead: %v", err)
			}

			start := time.Now()
			entries, err := history.Log(repo.Store, head, 0)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Log: %v", err)
			}
			t.Logf("Log(%d commits): %v (%d entries returned)", depth, elapsed, len(entries))
		})
	}
}

func TestVCSHistoryDepthBlame(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	for _, depth := range []int{50, 500} {
		depth := depth
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			root := t.TempDir()
			if err := Generate(root, Kubernetes); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			repo, err := vcs.InitRepo(root, "git")
			if err != nil {
				t.Fatalf("InitRepo: %v", err)
			}
			if err := staging.Add(repo, []string{"."}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			if _, err := staging.Commit(repo, "stress", "stress@mos.dev", "baseline"); err != nil {
				t.Fatalf("Commit: %v", err)
			}

			targetFile := buildHistory(t, root, repo, depth, 5)

			start := time.Now()
			lines, err := history.Blame(repo, targetFile)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Blame: %v", err)
			}
			t.Logf("Blame(%d commits, %d lines): %v", depth, len(lines), elapsed)
		})
	}
}

func TestVCSHistoryDepthCommitCost(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	root := t.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	repo, err := vcs.InitRepo(root, "git")
	if err != nil {
		t.Fatalf("InitRepo: %v", err)
	}
	if err := staging.Add(repo, []string{"."}); err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := staging.Commit(repo, "stress", "stress@mos.dev", "baseline"); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	allFiles, err := collectMosFiles(root)
	if err != nil {
		t.Fatalf("collectMosFiles: %v", err)
	}
	rng := rand.New(rand.NewSource(99))

	checkpoints := map[int]bool{50: true, 100: true, 250: true, 500: true}

	for i := 1; i <= 500; i++ {
		perm := rng.Perm(len(allFiles))
		picked := make([]string, 0, 5)
		for _, idx := range perm {
			if len(picked) >= 5 {
				break
			}
			picked = append(picked, allFiles[idx])
		}

		if err := mutateFiles(root, picked); err != nil {
			t.Fatalf("mutateFiles %d: %v", i, err)
		}

		start := time.Now()
		if err := staging.Add(repo, picked); err != nil {
			t.Fatalf("Add %d: %v", i, err)
		}
		_, err := staging.Commit(repo, "stress", "stress@mos.dev", fmt.Sprintf("c-%d", i))
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Commit %d: %v", i, err)
		}

		if checkpoints[i] {
			t.Logf("Commit #%d: %v", i, elapsed)
		}
	}
}

// --- CON-2026-249: Incremental Commit Stress ---

func BenchmarkVCSIncremental1(b *testing.B) {
	benchIncrementalCommit(b, 1)
}

func BenchmarkVCSIncremental10(b *testing.B) {
	benchIncrementalCommit(b, 10)
}

func BenchmarkVCSIncremental100(b *testing.B) {
	benchIncrementalCommit(b, 100)
}

func benchIncrementalCommit(b *testing.B, deltaSize int) {
	b.Helper()
	root := b.TempDir()
	if err := Generate(root, Kubernetes); err != nil {
		b.Fatalf("Generate: %v", err)
	}
	repo, err := vcs.InitRepo(root, "git")
	if err != nil {
		b.Fatalf("InitRepo: %v", err)
	}
	if err := staging.Add(repo, []string{"."}); err != nil {
		b.Fatalf("Add: %v", err)
	}
	if _, err := staging.Commit(repo, "bench", "bench@mos.dev", "baseline"); err != nil {
		b.Fatalf("Commit: %v", err)
	}

	allFiles, err := collectMosFiles(root)
	if err != nil {
		b.Fatalf("collectMosFiles: %v", err)
	}
	rng := rand.New(rand.NewSource(7))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		perm := rng.Perm(len(allFiles))
		picked := make([]string, 0, deltaSize)
		for _, idx := range perm {
			if len(picked) >= deltaSize {
				break
			}
			picked = append(picked, allFiles[idx])
		}

		if err := mutateFiles(root, picked); err != nil {
			b.Fatalf("mutateFiles: %v", err)
		}
		if err := staging.Add(repo, picked); err != nil {
			b.Fatalf("Add: %v", err)
		}
		if _, err := staging.Commit(repo, "bench", "bench@mos.dev", fmt.Sprintf("incr-%d", i)); err != nil {
			b.Fatalf("Commit: %v", err)
		}
	}
}

func TestTimeBudgetVCSIncremental1(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, Kubernetes)
	allFiles, err := collectMosFiles(root)
	if err != nil {
		t.Fatalf("collectMosFiles: %v", err)
	}

	picked := allFiles[:1]
	if err := mutateFiles(root, picked); err != nil {
		t.Fatalf("mutateFiles: %v", err)
	}

	start := time.Now()
	if err := staging.Add(repo, picked); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = staging.Commit(repo, "bench", "bench@mos.dev", "incr-1")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	budget := 500 * time.Millisecond
	if elapsed > budget {
		t.Errorf("Incremental commit (1 file) took %v, budget %v", elapsed, budget)
	} else {
		t.Logf("Incremental commit (1 file) completed in %v (budget: %v)", elapsed, budget)
	}
}

func TestTimeBudgetVCSIncremental100(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}
	root, repo := testSetupVCS(t, Kubernetes)
	allFiles, err := collectMosFiles(root)
	if err != nil {
		t.Fatalf("collectMosFiles: %v", err)
	}

	picked := allFiles[:100]
	if err := mutateFiles(root, picked); err != nil {
		t.Fatalf("mutateFiles: %v", err)
	}

	start := time.Now()
	if err := staging.Add(repo, picked); err != nil {
		t.Fatalf("Add: %v", err)
	}
	_, err = staging.Commit(repo, "bench", "bench@mos.dev", "incr-100")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Commit: %v", err)
	}

	budget := 1 * time.Second
	if elapsed > budget {
		t.Errorf("Incremental commit (100 files) took %v, budget %v", elapsed, budget)
	} else {
		t.Logf("Incremental commit (100 files) completed in %v (budget: %v)", elapsed, budget)
	}
}

// --- CON-2026-250: Clone & Transport Stress ---

func TestVCSCloneTransport(t *testing.T) {
	if testing.Short() {
		t.Skip("stress test skipped in short mode")
	}

	f := forge.InProcess(t)

	for _, depth := range []int{50, 500} {
		depth := depth
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			root := t.TempDir()
			if err := Generate(root, Kubernetes); err != nil {
				t.Fatalf("Generate: %v", err)
			}
			repo, err := vcs.InitRepo(root, "git")
			if err != nil {
				t.Fatalf("InitRepo: %v", err)
			}
			if err := staging.Add(repo, []string{"."}); err != nil {
				t.Fatalf("Add: %v", err)
			}
			if _, err := staging.Commit(repo, "stress", "stress@mos.dev", "baseline"); err != nil {
				t.Fatalf("Commit: %v", err)
			}

			buildHistory(t, root, repo, depth, 5)

			repoName := fmt.Sprintf("stress-%d", depth)
			remoteURL, err := f.CreateRepo(repoName)
			if err != nil {
				t.Fatalf("CreateRepo: %v", err)
			}
			if err := transport.AddRemote(repo, "origin", remoteURL); err != nil {
				t.Fatalf("AddRemote: %v", err)
			}

			pushStart := time.Now()
			if err := transport.Push(repo, "origin", transport.PushOpts{}); err != nil {
				t.Fatalf("Push: %v", err)
			}
			pushElapsed := time.Since(pushStart)

			cloneDest := t.TempDir()
			cloneStart := time.Now()
			_, err = transport.Clone(remoteURL, cloneDest)
			cloneElapsed := time.Since(cloneStart)
			if err != nil {
				t.Fatalf("Clone: %v", err)
			}

			srcSize, _ := dirSizeBytes(filepath.Join(root, ".git"))
			dstSize, _ := dirSizeBytes(filepath.Join(cloneDest, ".git"))

			t.Logf("depth=%d: push=%v, clone=%v, src_git=%dKB, dst_git=%dKB",
				depth, pushElapsed, cloneElapsed, srcSize/1024, dstSize/1024)
		})
	}
}
