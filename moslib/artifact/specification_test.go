package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCON024_CreateStandaloneSpecification(t *testing.T) {
	root := setupScaffold(t)

	path, err := CreateSpec(root, "SPEC-001", SpecOpts{
		Title:       "API backward compatibility",
		Enforcement: "warn",
	})
	if err != nil {
		t.Fatalf("CreateSpec: %v", err)
	}

	expectedDir := filepath.Join(root, ".mos", "specifications", "active", "SPEC-001")
	expectedPath := filepath.Join(expectedDir, "specification.mos")
	if path != expectedPath {
		t.Errorf("path = %q, want %q", path, expectedPath)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("specification file does not exist: %v", err)
	}
	assertParses(t, path)

	info := readSpecInfoByID(root, loadTestRegistry(t, root).Types["specification"], "SPEC-001")
	if info.Title != "API backward compatibility" {
		t.Errorf("title = %q, want %q", info.Title, "API backward compatibility")
	}
	if info.Enforcement != "warn" {
		t.Errorf("enforcement = %q, want %q", info.Enforcement, "warn")
	}
	if info.Status != "active" {
		t.Errorf("status = %q, want %q", info.Status, "active")
	}
}

func TestCON024_SpecificationPersistsBeyondContractCompletion(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{Title: "Persisting spec", Enforcement: "warn"})
	CreateContract(root, "CON-PERSIST", ContractOpts{
		Title: "Contract with spec",
		Specs: []string{"SPEC-001"},
	})

	if err := UpdateContractStatus(root, "CON-PERSIST", "complete"); err != nil {
		t.Fatalf("UpdateContractStatus: %v", err)
	}

	specPath := filepath.Join(root, ".mos", "specifications", "active", "SPEC-001", "specification.mos")
	if _, err := os.Stat(specPath); err != nil {
		t.Errorf("specification should still exist after contract completion: %v", err)
	}

	info := readSpecInfoByID(root, loadTestRegistry(t, root).Types["specification"], "SPEC-001")
	if info.Status != "active" {
		t.Errorf("spec status = %q, want %q", info.Status, "active")
	}
}

func TestCON024_SpecificationEnforcementLevelTransitions(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{Title: "Transitioning spec", Enforcement: "disabled"})

	td := loadTestRegistry(t, root).Types["specification"]

	info := readSpecInfoByID(root, td, "SPEC-001")
	if info.Enforcement != "disabled" {
		t.Fatalf("initial enforcement = %q, want disabled", info.Enforcement)
	}

	warn := "warn"
	if err := UpdateSpec(root, "SPEC-001", SpecUpdateOpts{Enforcement: &warn}); err != nil {
		t.Fatalf("UpdateSpec to warn: %v", err)
	}
	info = readSpecInfoByID(root, td, "SPEC-001")
	if info.Enforcement != "warn" {
		t.Errorf("enforcement after warn update = %q, want warn", info.Enforcement)
	}

	enforced := "enforced"
	if err := UpdateSpec(root, "SPEC-001", SpecUpdateOpts{Enforcement: &enforced}); err != nil {
		t.Fatalf("UpdateSpec to enforced: %v", err)
	}
	info = readSpecInfoByID(root, td, "SPEC-001")
	if info.Enforcement != "enforced" {
		t.Errorf("enforcement after enforced update = %q, want enforced", info.Enforcement)
	}
}

// Feature: Contract-Specification Binding

func TestCON024_ContractReferencesSpecifications(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{Title: "Spec A", Enforcement: "warn"})
	CreateSpec(root, "SPEC-002", SpecOpts{Title: "Spec B", Enforcement: "enforced"})

	_, err := CreateContract(root, "CON-SPECREF", ContractOpts{
		Title: "Contract with specs",
		Specs: []string{"SPEC-001", "SPEC-002"},
	})
	if err != nil {
		t.Fatalf("CreateContract: %v", err)
	}

	info, err := readContractInfo("CON-SPECREF", filepath.Join(root, ".mos", "contracts", "active", "CON-SPECREF", "contract.mos"))
	if err != nil {
		t.Fatalf("readContractInfo: %v", err)
	}
	if len(info.Specs) != 2 {
		t.Fatalf("specs count = %d, want 2", len(info.Specs))
	}
	if info.Specs[0] != "SPEC-001" || info.Specs[1] != "SPEC-002" {
		t.Errorf("specs = %v, want [SPEC-001, SPEC-002]", info.Specs)
	}

	summary, err := ContractSummary(root, "CON-SPECREF")
	if err != nil {
		t.Fatalf("ContractSummary: %v", err)
	}
	if len(summary.Specs) != 2 {
		t.Errorf("summary.Specs count = %d, want 2", len(summary.Specs))
	}

	short, err := ShowContractShort(root, "CON-SPECREF")
	if err != nil {
		t.Fatalf("ShowContractShort: %v", err)
	}
	if !strings.Contains(short, "SPEC-001") || !strings.Contains(short, "SPEC-002") {
		t.Errorf("short show should contain spec refs, got:\n%s", short)
	}
}

func TestCON024_SpecificationCreatedWithoutStandingContract(t *testing.T) {
	root := setupScaffold(t)

	path, err := CreateSpec(root, "SPEC-STANDALONE", SpecOpts{Title: "Standalone spec"})
	if err != nil {
		t.Fatalf("CreateSpec: %v", err)
	}

	info := readSpecInfoByID(root, loadTestRegistry(t, root).Types["specification"], "SPEC-STANDALONE")
	if info.Enforcement != "disabled" {
		t.Errorf("enforcement should default to disabled, got %q", info.Enforcement)
	}
	assertParses(t, path)
}

// Feature: Binders

func TestCON024_CreateBinderToGroupSpecifications(t *testing.T) {
	root := setupScaffold(t)

	path, err := CreateBinder(root, "BND-001", BinderOpts{Title: "PTP API Standards"})
	if err != nil {
		t.Fatalf("CreateBinder: %v", err)
	}

	expectedPath := filepath.Join(root, ".mos", "binders", "active", "BND-001", "binder.mos")
	if path != expectedPath {
		t.Errorf("path = %q, want %q", path, expectedPath)
	}
	assertParses(t, path)
}

func TestCON024_BindSpecificationsToBinder(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{Title: "Spec A", Enforcement: "warn"})
	CreateSpec(root, "SPEC-002", SpecOpts{Title: "Spec B", Enforcement: "enforced"})
	CreateBinder(root, "BND-001", BinderOpts{Title: "API Binder"})

	if err := BinderBind(root, "BND-001", "SPEC-001"); err != nil {
		t.Fatalf("BinderBind SPEC-001: %v", err)
	}
	if err := BinderBind(root, "BND-001", "SPEC-002"); err != nil {
		t.Fatalf("BinderBind SPEC-002: %v", err)
	}

	// Idempotent
	if err := BinderBind(root, "BND-001", "SPEC-001"); err != nil {
		t.Fatalf("duplicate BinderBind: %v", err)
	}

	td := loadTestRegistry(t, root).Types["binder"]
	path, _ := FindGenericPath(root, td, "BND-001")
	info := readBinderInfo(path, "BND-001")

	if len(info.Specs) != 2 {
		t.Errorf("specs count = %d, want 2", len(info.Specs))
	}

	showOut, err := ShowBinder(root, "BND-001")
	if err != nil {
		t.Fatalf("ShowBinder: %v", err)
	}
	if !strings.Contains(showOut, "SPEC-001") || !strings.Contains(showOut, "SPEC-002") {
		t.Errorf("ShowBinder should list both specs, got:\n%s", showOut)
	}
}

func TestCON024_BinderShowsEnforcementRollup(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{Title: "Disabled spec", Enforcement: "disabled"})
	CreateSpec(root, "SPEC-002", SpecOpts{Title: "Warn spec", Enforcement: "warn"})
	CreateSpec(root, "SPEC-003", SpecOpts{Title: "Enforced spec", Enforcement: "enforced"})
	CreateBinder(root, "BND-001", BinderOpts{Title: "Mixed Enforcement"})

	BinderBind(root, "BND-001", "SPEC-001")
	BinderBind(root, "BND-001", "SPEC-002")
	BinderBind(root, "BND-001", "SPEC-003")

	showOut, err := ShowBinder(root, "BND-001")
	if err != nil {
		t.Fatalf("ShowBinder: %v", err)
	}
	if !strings.Contains(showOut, "disabled=1") {
		t.Errorf("ShowBinder should show disabled=1, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "warn=1") {
		t.Errorf("ShowBinder should show warn=1, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "enforced=1") {
		t.Errorf("ShowBinder should show enforced=1, got:\n%s", showOut)
	}
}

// Feature: Traceability Triangle

func TestCON024_SpecificationBindsToSymbol(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{
		Title:       "API handler spec",
		Enforcement: "warn",
		Symbol:      "pkg/api.Handler",
	})

	showOut, err := ShowSpec(root, "SPEC-001")
	if err != nil {
		t.Fatalf("ShowSpec: %v", err)
	}
	if !strings.Contains(showOut, "Symbol: pkg/api.Handler") {
		t.Errorf("ShowSpec should display symbol binding, got:\n%s", showOut)
	}
}

func TestCON024_SpecificationBindsToHarness(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{
		Title:       "Full traceability spec",
		Enforcement: "enforced",
		Symbol:      "pkg/api.Handler",
		Harness:     "tests/api_test.go:TestBackwardCompat",
	})

	showOut, err := ShowSpec(root, "SPEC-001")
	if err != nil {
		t.Fatalf("ShowSpec: %v", err)
	}
	if !strings.Contains(showOut, "Symbol: pkg/api.Handler") {
		t.Errorf("ShowSpec should display symbol, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Harness: tests/api_test.go:TestBackwardCompat") {
		t.Errorf("ShowSpec should display harness, got:\n%s", showOut)
	}
	if !strings.Contains(showOut, "Traceability: complete") {
		t.Errorf("ShowSpec should show traceability complete, got:\n%s", showOut)
	}
}

func TestCON024_TraceabilityReportAcrossBinder(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-001", SpecOpts{
		Title:       "Bound spec",
		Enforcement: "enforced",
		Symbol:      "pkg/api.Handler",
		Harness:     "tests/api_test.go:TestA",
	})
	CreateSpec(root, "SPEC-002", SpecOpts{
		Title:       "Partially bound",
		Enforcement: "warn",
		Symbol:      "pkg/cache.Store",
	})
	CreateSpec(root, "SPEC-003", SpecOpts{
		Title:       "Unbound spec",
		Enforcement: "disabled",
	})

	CreateBinder(root, "BND-001", BinderOpts{Title: "Trace Binder"})
	BinderBind(root, "BND-001", "SPEC-001")
	BinderBind(root, "BND-001", "SPEC-002")
	BinderBind(root, "BND-001", "SPEC-003")

	report, err := BinderTrace(root, "BND-001")
	if err != nil {
		t.Fatalf("BinderTrace: %v", err)
	}

	if !strings.Contains(report, "SPEC-001") {
		t.Errorf("trace report should include SPEC-001")
	}
	if !strings.Contains(report, "pkg/api.Handler") {
		t.Errorf("trace report should include symbol for SPEC-001")
	}
	if !strings.Contains(report, "INCOMPLETE") {
		t.Errorf("trace report should flag incomplete specs, got:\n%s", report)
	}
}

// Feature: Enforcement Semantics

func TestCON024_DisabledSpecificationProducesNoDiagnostics(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-DISABLED", SpecOpts{
		Title:       "Disabled spec with no bindings",
		Enforcement: "disabled",
	})

	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	for _, d := range diags {
		if d.Rule == "spec-traceability" && strings.Contains(d.Message, "SPEC-DISABLED") {
			t.Errorf("disabled spec should produce no traceability diagnostics, got: %s", d.Message)
		}
	}
}

func TestCON024_WarnSpecificationProducesWarnings(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-WARN", SpecOpts{
		Title:       "Warn spec missing harness",
		Enforcement: "warn",
		Symbol:      "pkg/api.Handler",
	})

	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	foundWarning := false
	for _, d := range diags {
		if d.Rule == "spec-traceability" && strings.Contains(d.Message, "SPEC-WARN") {
			if d.Severity != "warning" {
				t.Errorf("expected warning severity for warn spec, got %s", d.Severity)
			}
			foundWarning = true
		}
	}
	if !foundWarning {
		t.Errorf("expected at least one warning for warn spec with missing harness")
	}

	hasErrors := false
	for _, d := range diags {
		if d.Severity == "error" && d.Rule == "spec-traceability" {
			hasErrors = true
		}
	}
	if hasErrors {
		t.Errorf("warn-level enforcement should not produce errors")
	}
}

func TestCON024_EnforcedSpecificationBlocksOnViolations(t *testing.T) {
	root := setupScaffold(t)

	CreateSpec(root, "SPEC-ENFORCED", SpecOpts{
		Title:       "Enforced spec missing harness",
		Enforcement: "enforced",
		Symbol:      "pkg/api.Handler",
	})

	diags, err := LintAll(root)
	if err != nil {
		t.Fatalf("lint: %v", err)
	}

	foundError := false
	for _, d := range diags {
		if d.Rule == "spec-traceability" && strings.Contains(d.Message, "SPEC-ENFORCED") {
			if d.Severity != "error" {
				t.Errorf("expected error severity for enforced spec, got %s", d.Severity)
			}
			foundError = true
		}
	}
	if !foundError {
		t.Errorf("expected at least one error for enforced spec with missing harness")
	}
}

// --- CON-2026-025: Contract Scenario CLI ---

func TestCON038_SpecificationCADIncludesAddressesField(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["specification"]

	hasAddresses := false
	hasSatisfies := false
	for _, f := range td.Fields {
		if f.Name == "addresses" {
			hasAddresses = true
		}
		if f.Name == "satisfies" {
			hasSatisfies = true
		}
	}
	if !hasAddresses {
		t.Error("specification CAD should include addresses field")
	}
	if !hasSatisfies {
		t.Error("specification CAD should include satisfies field")
	}
}

// --- CON-2026-041: Need Transition Guards ---
