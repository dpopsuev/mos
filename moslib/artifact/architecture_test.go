package artifact

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/model"
)

func TestCON034_ArchitectureTypeInDefaultConfig(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)

	arch, ok := reg.Types["architecture"]
	if !ok {
		t.Fatal("architecture type not in registry")
	}
	if arch.Directory != "architectures" {
		t.Errorf("architecture.Directory = %q, want architectures", arch.Directory)
	}

	hasResolution := false
	for _, f := range arch.Fields {
		if f.Name == "resolution" && len(f.Enum) == 2 {
			hasResolution = true
		}
	}
	if !hasResolution {
		t.Error("architecture missing resolution enum field with service/component")
	}
}

func TestCON034_CreateServiceLevelArchitecture(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["architecture"]

	path, err := GenericCreate(root, td, "ARCH-2026-001", map[string]string{
		"title":      "Mos Service Map",
		"resolution": "service",
		"status":     "draft",
		"implements": "SPEC-2026-001",
	})
	if err != nil {
		t.Fatalf("GenericCreate architecture: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("architecture file not found: %v", err)
	}
	assertParses(t, path)
}

func TestCON034_CreateComponentLevelArchitecture(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["architecture"]

	path, err := GenericCreate(root, td, "ARCH-COMP-001", map[string]string{
		"title":      "Governance Internals",
		"resolution": "component",
		"status":     "draft",
	})
	if err != nil {
		t.Fatalf("GenericCreate architecture: %v", err)
	}
	assertParses(t, path)

	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `resolution = "component"`) {
		t.Error("architecture should contain resolution = component")
	}
}

func TestCON034_ParseArchModelFromDSL(t *testing.T) {
	src := `architecture "Mos Service Map" {
  resolution = "service"
  implements = "SPEC-2026-001"

  service "CLI" {
    package = "cmd/mos"
    trust_zone = "public"
    exposes = ["mos init", "mos contract"]
  }

  service "Governance Engine" {
    package = "moslib/governance"
    trust_zone = "internal"
  }

  edge "cli-to-governance" {
    from = "CLI"
    to = "Governance Engine"
    protocol = "function-call"
  }

  forbidden "no-governance-to-cli" {
    from = "Governance Engine"
    to = "CLI"
    reason = "engine must not depend on CLI layer"
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	m := ParseArchModel(ab)

	if m.Resolution != "service" {
		t.Errorf("resolution = %q, want service", m.Resolution)
	}
	if m.Implements != "SPEC-2026-001" {
		t.Errorf("implements = %q, want SPEC-2026-001", m.Implements)
	}
	if len(m.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(m.Services))
	}
	if m.Services[0].Name != "CLI" || m.Services[0].Package != "cmd/mos" {
		t.Errorf("service[0] = %+v, want CLI at cmd/mos", m.Services[0])
	}
	if m.Services[0].TrustZone != "public" {
		t.Errorf("trust_zone = %q, want public", m.Services[0].TrustZone)
	}
	if len(m.Services[0].Exposes) != 2 {
		t.Errorf("exposes = %v, want 2 items", m.Services[0].Exposes)
	}
	if len(m.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(m.Edges))
	}
	if m.Edges[0].From != "CLI" || m.Edges[0].To != "Governance Engine" {
		t.Errorf("edge = %+v", m.Edges[0])
	}
	if len(m.Forbidden) != 1 {
		t.Fatalf("expected 1 forbidden, got %d", len(m.Forbidden))
	}
	if m.Forbidden[0].From != "Governance Engine" || m.Forbidden[0].To != "CLI" {
		t.Errorf("forbidden = %+v", m.Forbidden[0])
	}
	if m.Forbidden[0].Reason != "engine must not depend on CLI layer" {
		t.Errorf("forbidden reason = %q", m.Forbidden[0].Reason)
	}
}

func TestCON034_RenderMermaidOutput(t *testing.T) {
	m := ArchModel{
		Title:      "Test Map",
		Resolution: "service",
		Services: []ArchService{
			{Name: "CLI", Package: "cmd/mos", TrustZone: "public"},
			{Name: "Engine", Package: "moslib/governance", TrustZone: "internal"},
		},
		Edges: []ArchEdge{
			{Name: "cli-to-engine", From: "CLI", To: "Engine", Protocol: "function-call"},
		},
		Forbidden: []ArchForbidden{
			{Name: "no-reverse", From: "Engine", To: "CLI", Reason: "no reverse dependency"},
		},
	}

	mermaid := RenderMermaid(m)

	if !strings.Contains(mermaid, "graph TD") {
		t.Error("missing graph TD header")
	}
	if !strings.Contains(mermaid, "CLI[") {
		t.Error("missing CLI node")
	}
	if !strings.Contains(mermaid, "Engine[") {
		t.Error("missing Engine node")
	}
	if !strings.Contains(mermaid, `-->|"function-call"|`) {
		t.Error("missing edge with protocol label")
	}
	if !strings.Contains(mermaid, `-.-x|"no reverse dependency"|`) {
		t.Error("missing forbidden edge rendering")
	}
}

func TestCON034_ComponentWithSymbols(t *testing.T) {
	src := `architecture "Governance Internals" {
  resolution = "component"

  component "Contract CRUD" {
    package = "moslib/governance"
    symbols = ["CreateContract", "UpdateContract", "DeleteContract"]
  }

  component "Hook Evaluator" {
    package = "moslib/governance"
    symbols = ["EvaluateHooks"]
  }

  edge "scenario-triggers-hooks" {
    from = "Contract CRUD"
    to = "Hook Evaluator"
    trigger = "after scenario status change"
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	m := ParseArchModel(ab)

	if m.Resolution != "component" {
		t.Errorf("resolution = %q, want component", m.Resolution)
	}
	if len(m.Services) != 2 {
		t.Fatalf("expected 2 components, got %d", len(m.Services))
	}
	if len(m.Services[0].Symbols) != 3 {
		t.Errorf("expected 3 symbols on first component, got %d", len(m.Services[0].Symbols))
	}
	if len(m.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(m.Edges))
	}
	if m.Edges[0].Trigger != "after scenario status change" {
		t.Errorf("edge trigger = %q", m.Edges[0].Trigger)
	}
}

func TestCON034_ArchitectureLifecycleTransitions(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["architecture"]

	GenericCreate(root, td, "ARCH-LC-001", map[string]string{
		"title": "Lifecycle test", "resolution": "service", "status": "draft",
	})

	if err := GenericUpdateStatus(root, td, "ARCH-LC-001", "active"); err != nil {
		t.Fatalf("transition to active: %v", err)
	}
	if err := GenericUpdateStatus(root, td, "ARCH-LC-001", "superseded"); err != nil {
		t.Fatalf("transition to superseded: %v", err)
	}

	items, _ := GenericList(root, td, "superseded")
	if len(items) != 1 {
		t.Error("expected superseded architecture in archive")
	}
}

func TestCON034_TrustZoneForbiddenBlock(t *testing.T) {
	src := `architecture "Zone Test" {
  resolution = "service"

  service "Internal API" {
    package = "internal/api"
    trust_zone = "internal"
  }

  service "Public Gateway" {
    package = "cmd/gateway"
    trust_zone = "public"
  }

  forbidden "no-internal-to-public" {
    from_trust_zone = "internal"
    to_trust_zone = "public"
    reason = "internal services must not call public surface"
  }
}
`
	f, err := dsl.Parse(src, nil)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	ab := f.Artifact.(*dsl.ArtifactBlock)
	m := ParseArchModel(ab)

	if len(m.Forbidden) != 1 {
		t.Fatalf("expected 1 forbidden, got %d", len(m.Forbidden))
	}
	fb := m.Forbidden[0]
	if fb.FromTrustZone != "internal" || fb.ToTrustZone != "public" {
		t.Errorf("forbidden trust zones = %q -> %q", fb.FromTrustZone, fb.ToTrustZone)
	}
}

func TestCON034_ArchitectureDirectoriesCreatedByInit(t *testing.T) {
	root := setupScaffold(t)
	for _, sub := range []string{"active", "archive"} {
		dir := filepath.Join(root, ".mos", "architectures", sub)
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("architectures/%s directory not created by init: %v", sub, err)
		}
	}
}

// --- CON-2026-097: Self-emerging architecture views ---

func TestProjectToArchModel_PackageLevel(t *testing.T) {
	proj := &model.Project{
		Path: "github.com/example/app",
		Namespaces: []*model.Namespace{
			model.NewNamespace("dsl", "github.com/example/app/moslib/dsl"),
			model.NewNamespace("governance", "github.com/example/app/moslib/governance"),
			model.NewNamespace("main", "github.com/example/app/cmd/mos"),
		},
		DependencyGraph: model.NewDependencyGraph(),
	}
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/cmd/mos",
		"github.com/example/app/moslib/governance", false)
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/moslib/governance",
		"github.com/example/app/moslib/dsl", false)
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/cmd/mos",
		"fmt", true)

	m := ProjectToArchModel(proj, SyncOptions{ModulePath: "github.com/example/app"})

	if len(m.Services) != 3 {
		t.Fatalf("expected 3 services, got %d", len(m.Services))
	}
	if len(m.Edges) != 2 {
		t.Fatalf("expected 2 internal edges (external filtered), got %d", len(m.Edges))
	}
	for _, e := range m.Edges {
		if e.Protocol != "import" {
			t.Errorf("edge protocol = %q, want import", e.Protocol)
		}
	}
}

func TestProjectToArchModel_GroupLevel(t *testing.T) {
	proj := &model.Project{
		Path: "github.com/example/app",
		Namespaces: []*model.Namespace{
			model.NewNamespace("dsl", "github.com/example/app/moslib/dsl"),
			model.NewNamespace("antlrgen", "github.com/example/app/moslib/dsl/antlrgen"),
			model.NewNamespace("governance", "github.com/example/app/moslib/governance"),
			model.NewNamespace("main", "github.com/example/app/cmd/mos"),
			model.NewNamespace("contract", "github.com/example/app/cmd/mos/contract"),
		},
		DependencyGraph: model.NewDependencyGraph(),
	}
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/cmd/mos",
		"github.com/example/app/moslib/governance", false)
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/cmd/mos/contract",
		"github.com/example/app/moslib/governance", false)
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/moslib/governance",
		"github.com/example/app/moslib/dsl", false)
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/moslib/dsl/antlrgen",
		"github.com/example/app/moslib/dsl", false)

	groups := []ComponentGroup{
		{Name: "CLI", Packages: []string{"cmd/mos", "cmd/mos/contract"}},
		{Name: "Parser", Packages: []string{"moslib/dsl", "moslib/dsl/antlrgen"}},
	}
	m := ProjectToArchModel(proj, SyncOptions{
		ModulePath: "github.com/example/app",
		Groups:     groups,
	})

	if len(m.Services) != 3 {
		t.Fatalf("expected 3 groups (CLI, Parser, moslib/governance), got %d: %v", len(m.Services), m.Services)
	}

	edgeMap := make(map[string]bool)
	for _, e := range m.Edges {
		edgeMap[e.From+" -> "+e.To] = true
	}
	if !edgeMap["CLI -> moslib/governance"] {
		t.Error("missing CLI -> moslib/governance edge")
	}
	if !edgeMap["moslib/governance -> Parser"] {
		t.Error("missing moslib/governance -> Parser edge")
	}
	if edgeMap["Parser -> Parser"] {
		t.Error("intra-group edge should be suppressed")
	}
}

func TestProjectToArchModel_ExcludeTests(t *testing.T) {
	proj := &model.Project{
		Path: "github.com/example/app",
		Namespaces: []*model.Namespace{
			model.NewNamespace("governance", "github.com/example/app/moslib/governance"),
			model.NewNamespace("forge", "github.com/example/app/testkit/forge"),
		},
		DependencyGraph: model.NewDependencyGraph(),
	}
	proj.DependencyGraph.AddEdge(
		"github.com/example/app/testkit/forge",
		"github.com/example/app/moslib/governance", false)

	m := ProjectToArchModel(proj, SyncOptions{
		ModulePath:   "github.com/example/app",
		ExcludeTests: true,
	})

	if len(m.Services) != 1 {
		t.Fatalf("expected 1 service (testkit excluded), got %d", len(m.Services))
	}
	if len(m.Edges) != 0 {
		t.Fatalf("expected 0 edges (testkit edges excluded), got %d", len(m.Edges))
	}
}

func TestCheckForbiddenEdges(t *testing.T) {
	live := ArchModel{
		Edges: []ArchEdge{
			{From: "governance", To: "dsl"},
			{From: "governance", To: "linter"},
		},
	}
	declared := ArchModel{
		Forbidden: []ArchForbidden{
			{From: "governance", To: "linter", Reason: "breaks import cycle"},
			{From: "dsl", To: "governance", Reason: "leaf package"},
		},
	}

	violations := CheckForbiddenEdges(live, declared)
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if !strings.Contains(violations[0], "governance -> linter") {
		t.Errorf("violation = %q, want governance -> linter", violations[0])
	}
}

func TestRenderArchMos(t *testing.T) {
	m := ArchModel{
		Title: "Test",
		Services: []ArchService{
			{Name: "cmd/mos", Package: "github.com/example/app/cmd/mos"},
			{Name: "moslib/dsl"},
		},
		Edges: []ArchEdge{
			{From: "cmd/mos", To: "moslib/dsl", Protocol: "import"},
		},
	}
	out := RenderArchMos(m)
	if !strings.Contains(out, `architecture "Test"`) {
		t.Error("missing architecture block")
	}
	if !strings.Contains(out, `component "cmd/mos"`) {
		t.Error("missing component block")
	}
	if !strings.Contains(out, `edge "cmd/mos -> moslib/dsl"`) {
		t.Error("missing edge block")
	}
}

func TestRenderArchMarkdown(t *testing.T) {
	m := ArchModel{
		Title: "Test App",
		Services: []ArchService{
			{Name: "cli", Package: "cmd/mos"},
		},
		Edges: []ArchEdge{
			{From: "cli", To: "engine"},
		},
	}
	md := RenderArchMarkdown(m)
	if !strings.Contains(md, "# Architecture: Test App") {
		t.Error("missing title")
	}
	if !strings.Contains(md, "Auto-generated") {
		t.Error("missing auto-generated notice")
	}
	if !strings.Contains(md, "```mermaid") {
		t.Error("missing mermaid block")
	}
}

func TestLoadComponentGroups(t *testing.T) {
	root := t.TempDir()
	mosDir := filepath.Join(root, ".mos")
	os.MkdirAll(mosDir, 0o755)
	os.WriteFile(filepath.Join(mosDir, "config.mos"), []byte(`config {
  component_group "CLI" {
    packages = "cmd/mos, cmd/mos/contract, cmd/mos/rule"
  }

  component_group "Parser" {
    packages = "moslib/dsl, moslib/dsl/antlrgen"
  }
}
`), 0o644)

	groups, err := LoadComponentGroups(root)
	if err != nil {
		t.Fatalf("LoadComponentGroups: %v", err)
	}
	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}

	groupMap := make(map[string][]string)
	for _, g := range groups {
		groupMap[g.Name] = g.Packages
	}

	cli, ok := groupMap["CLI"]
	if !ok {
		t.Fatal("missing CLI group")
	}
	if len(cli) != 3 {
		t.Errorf("CLI group has %d packages, want 3", len(cli))
	}

	parser, ok := groupMap["Parser"]
	if !ok {
		t.Fatal("missing Parser group")
	}
	if len(parser) != 2 {
		t.Errorf("Parser group has %d packages, want 2", len(parser))
	}
}

func TestLoadComponentGroups_NoConfig(t *testing.T) {
	root := t.TempDir()
	groups, err := LoadComponentGroups(root)
	if err != nil {
		t.Fatalf("LoadComponentGroups: %v", err)
	}
	if len(groups) != 0 {
		t.Errorf("expected 0 groups with no config, got %d", len(groups))
	}
}

// --- CON-2026-035: Docs -- The Externalization Primitive ---
