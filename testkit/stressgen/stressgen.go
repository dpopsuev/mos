package stressgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Profile defines the scale parameters for a synthetic .mos/ directory.
type Profile struct {
	Name              string
	Rules             int
	ActiveContracts   int
	ArchivedContracts int
	ResolutionLayers  int
	LexiconTerms   int
	IncludeFiles      int
}

// LinuxKernel returns a profile modeled after Linux kernel governance scale.
// ~1,700 MAINTAINERS subsystems, 40M LOC, 2,134 devs/release, 1,780 orgs.
var LinuxKernel = Profile{
	Name:              "linux-kernel",
	Rules:             2000,
	ActiveContracts:   200,
	ArchivedContracts: 500,
	ResolutionLayers:  6,
	LexiconTerms:   2000,
	IncludeFiles:      1000,
}

// Kubernetes returns a profile modeled after Kubernetes governance scale.
// 30+ SIGs, 4,149+ KEPs, 135K commits, 200+ sub-repos.
var Kubernetes = Profile{
	Name:              "kubernetes",
	Rules:             500,
	ActiveContracts:   300,
	ArchivedContracts: 1000,
	ResolutionLayers:  5,
	LexiconTerms:   800,
	IncludeFiles:      400,
}

// Dunix is a hypothetical cross-planetary distributed microkernel OS developed
// on Terra and Mars. Combines Linux kernel + Kubernetes governance at
// interplanetary scale: redundancy, robustness, multi-tenancy, capability-based
// security, and relativistic-delay-tolerant consensus.
//
// Scale rationale:
//   - 5,000 rules: kernel (2k) + orchestration (500) + networking (500) +
//     redundancy (500) + multi-tenancy (500) + habitat-specific (500) + security (500)
//   - 1,500 active contracts: ~750 per planet, concurrent cross-planetary development
//   - 5,000 archived: decades of cross-planetary history
//   - 10 resolution layers: kernel → subsystem → module → driver → habitat →
//     planet → orbital → federation → organization → interplanetary
//   - 8,000 lexicon terms: two full locales (Terran + Martian English) + OS domain
//   - 3,000 include files: massive spec decomposition across subsystems
var Dunix = Profile{
	Name:              "dunix",
	Rules:             5000,
	ActiveContracts:   1500,
	ArchivedContracts: 5000,
	ResolutionLayers:  10,
	LexiconTerms:   8000,
	IncludeFiles:      3000,
}

// Locale defines human-protocol keyword translations for a language variant.
// The lexicon file (always written in Terran English for bootstrap) contains
// a keywords block that maps machine keywords to human keywords. All other
// .mos files in the project then use the human keywords.
type Locale struct {
	Name     string
	Keywords map[string]string // machine keyword -> human keyword (nil = identity)
}

// TerranEnglish is the default locale where human and machine keywords match.
var TerranEnglish = Locale{
	Name:     "terran-english",
	Keywords: nil,
}

// MartianEnglish is the Martian colonist dialect. Evolved from Terran English
// under the influence of capability-based microkernel culture, high-latency
// protocol-centric communication, and Mars Habitat Authority governance norms.
var MartianEnglish = Locale{
	Name: "martian-english",
	Keywords: map[string]string{
		"feature":     "capability",
		"scenario":    "protocol",
		"given":       "assuming",
		"when":        "upon",
		"then":        "verify",
		"group":       "division",
		"background":  "baseline",
		"spec":        "manifest",
		"include":     "import",
		"rule":        "mandate",
		"contract":    "charter",
		"config":      "settings",
		"declaration": "proclamation",
		"lexicon":  "lexicon",
		"layers":      "strata",
		"layer":       "stratum",
	},
}

// Generate creates a synthetic .mos/ directory at the given root using DSL format.
func Generate(root string, p Profile) error {
	mos := filepath.Join(root, ".mos")

	if err := writeConfig(mos); err != nil {
		return err
	}
	if err := writeDeclaration(mos); err != nil {
		return err
	}
	if err := writeLexicon(mos, p); err != nil {
		return err
	}
	if err := writeLayers(mos, p); err != nil {
		return err
	}
	if err := writeRules(mos, p); err != nil {
		return err
	}
	if err := writeContracts(mos, "active", p.ActiveContracts, p); err != nil {
		return err
	}
	if err := writeContracts(mos, "archive", p.ArchivedContracts, p); err != nil {
		return err
	}

	return nil
}

// GenerateLocalized creates a synthetic .mos/ directory where non-lexicon
// files use localized human-protocol keywords. The lexicon file itself is
// always in Terran English (bootstrap contract) but contains a keywords block
// that defines the locale mapping. All rule, contract, config, declaration,
// and layers files use the locale's human keywords.
func GenerateLocalized(root string, p Profile, loc Locale) error {
	mos := filepath.Join(root, ".mos")
	kw := func(machine string) string {
		if loc.Keywords == nil {
			return machine
		}
		if h, ok := loc.Keywords[machine]; ok {
			return h
		}
		return machine
	}

	if err := writeLocalizedConfig(mos, kw); err != nil {
		return err
	}
	if err := writeLocalizedDeclaration(mos, kw); err != nil {
		return err
	}
	if err := writeLocalizedLexicon(mos, p, loc); err != nil {
		return err
	}
	if err := writeLocalizedLayers(mos, p, kw); err != nil {
		return err
	}
	if err := writeLocalizedRules(mos, p, kw); err != nil {
		return err
	}
	if err := writeLocalizedContracts(mos, "active", p.ActiveContracts, p, kw); err != nil {
		return err
	}
	if err := writeLocalizedContracts(mos, "archive", p.ArchivedContracts, p, kw); err != nil {
		return err
	}

	return nil
}

func writeLocalizedConfig(mos string, kw func(string) string) error {
	return mkdirWrite(filepath.Join(mos, "config.mos"), fmt.Sprintf(`%s {
  mos {
    version = 1
  }

  backend {
    type = "git"
  }

  upstream "dunix-federation" {
    url   = "https://github.com/dunix/federation-mos"
    ref   = "main"
    scope = "interplanetary"
  }

  governance {
    model = "federation"
    scope = "planet"
    jurisdiction = "terra"
    ratification_authority = ["terra-council", "mars-authority"]
  }
}
`, kw("config")))
}

func writeLocalizedDeclaration(mos string, kw func(string) string) error {
	return mkdirWrite(filepath.Join(mos, "declaration.mos"), fmt.Sprintf(`%s {
  name = "Dunix"
  created = "2026-01-01"
  authors = ["terra-kernel-collective", "mars-systems-authority"]

  federation {
    upstream = "dunix-federation"
    inherits_rules = true
    inherits_lexicon = true
    inherits_layers = true
    override_policy = "extend-only"
  }

  principles {
    microkernel_purity = "mechanism only; all policy in userspace"
    interplanetary_consistency = "eventual consistency under light-speed delay"
    capability_security = "unforgeable object capabilities for all IPC"
    redundancy = "no single point of failure across planetary boundaries"
  }
}
`, kw("declaration")))
}

func writeLocalizedLexicon(mos string, p Profile, loc Locale) error {
	var b strings.Builder
	b.WriteString("lexicon {\n")

	if loc.Keywords != nil {
		b.WriteString("  keywords {\n")
		for machine, human := range loc.Keywords {
			fmt.Fprintf(&b, "    %s = %q\n", machine, human)
		}
		b.WriteString("  }\n\n")
	}

	b.WriteString("  terms {\n")
	dunixTerms := []struct{ k, v string }{
		{"ipc_gate", "Unforgeable capability for inter-process communication"},
		{"capability", "Object reference granting specific access rights"},
		{"address_space", "Virtual memory region owned by a process"},
		{"microkernel", "Minimal privileged kernel providing only mechanism"},
		{"scheduler", "Policy module selecting next thread to execute"},
		{"habitat", "Self-contained Mars colony module"},
		{"light_delay", "Signal propagation time between Terra and Mars (4-24 min)"},
		{"orbital_relay", "Communication satellite in planetary orbit"},
		{"failover", "Automatic switch to redundant component on failure"},
		{"quorum", "Minimum participants required for consensus"},
		{"split_brain", "Network partition creating independent decision groups"},
		{"tenant", "Isolated governance domain within shared infrastructure"},
		{"namespace", "Hierarchical isolation boundary for resources"},
		{"tcb", "Trusted Computing Base — minimized in microkernel design"},
		{"formal_verification", "Mathematical proof of implementation correctness"},
		{"confinement", "Restriction preventing information leakage"},
	}
	for _, t := range dunixTerms {
		fmt.Fprintf(&b, "    %s = %q\n", t.k, t.v)
	}
	for i := 0; i < p.LexiconTerms-len(dunixTerms); i++ {
		fmt.Fprintf(&b, "    term_%04d = \"Definition of term %d\"\n", i, i)
	}
	b.WriteString("  }\n\n")

	b.WriteString("  artifact_labels {\n")
	b.WriteString("    rule = \"Mandate\"\n")
	b.WriteString("    contract = \"Charter\"\n")
	b.WriteString("  }\n}\n")

	return mkdirWrite(filepath.Join(mos, "lexicon", "default.mos"), b.String())
}

func writeLocalizedLayers(mos string, p Profile, kw func(string) string) error {
	dunixLayers := []string{
		"kernel", "subsystem", "module", "driver", "habitat",
		"planet", "orbital", "federation", "organization", "interplanetary",
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s {\n", kw("layers"))
	for i := 0; i < p.ResolutionLayers; i++ {
		name := fmt.Sprintf("layer_%d", i)
		if i < len(dunixLayers) {
			name = dunixLayers[i]
		}
		fmt.Fprintf(&b, "  %s %q {\n    level = %d\n  }\n", kw("layer"), name, i+1)
	}
	b.WriteString("}\n")
	return mkdirWrite(filepath.Join(mos, "resolution", "layers.mos"), b.String())
}

func writeLocalizedRules(mos string, p Profile, kw func(string) string) error {
	dunixLayers := []string{
		"kernel", "subsystem", "module", "driver", "habitat",
		"planet", "orbital", "federation", "organization", "interplanetary",
	}

	includeIdx := 0
	for i := 0; i < p.Rules; i++ {
		ruleType := "mechanical"
		if i%3 == 0 {
			ruleType = "interpretive"
		}

		layerIdx := i % p.ResolutionLayers
		layerName := fmt.Sprintf("layer_%d", layerIdx)
		if layerIdx < len(dunixLayers) {
			layerName = dunixLayers[layerIdx]
		}

		var featureSection string

		if includeIdx < p.IncludeFiles && i%2 == 0 {
			featureName := fmt.Sprintf("spec_%04d.feature", i)
			featurePath := filepath.Join(mos, "rules", ruleType, featureName)
			featureContent := fmt.Sprintf(`%s "Generated mandate" {
  %s "baseline" {
    sut = "rules/%s"
    test = "mandate_%04d_test.go"
    %s {
      a Dunix subsystem
    }
    %s {
      the mandate is evaluated
    }
    %s {
      it passes
    }
  }
}
`, kw("feature"), kw("scenario"), ruleType, i,
				kw("given"), kw("when"), kw("then"))
			if err := mkdirWrite(featurePath, featureContent); err != nil {
				return err
			}
			featureSection = fmt.Sprintf("  %s {\n    %s %q\n  }",
				kw("spec"), kw("include"), featureName)
			includeIdx++
		} else {
			featureSection = fmt.Sprintf(`  %s "Generated mandate" {
    %s "baseline" {
      sut = "rules/%s"
      test = "mandate_%04d_test.go"
      %s {
        a Dunix subsystem
      }
      %s {
        the mandate is evaluated
      }
      %s {
        it passes
      }
    }
  }`, kw("feature"), kw("scenario"), ruleType, i,
				kw("given"), kw("when"), kw("then"))
		}

		jurisdictions := []string{
			"interplanetary", "planet", "division",
			"subsystem", "team", "collective", "cabinet",
		}
		jurisdiction := jurisdictions[i%len(jurisdictions)]

		var extendsLine string
		if jurisdiction != "interplanetary" && i >= len(jurisdictions) {
			parentIdx := i - (i % len(jurisdictions))
			extendsLine = fmt.Sprintf("\n  extends = \"R-%04d\"", parentIdx)
		}

		content := fmt.Sprintf(`%s "R-%04d" {
  name = "Generated mandate %d"
  type = "%s"
  scope = "%s"
  jurisdiction = "%s"%s
  enforcement = "error"

%s
}
`, kw("rule"), i, i, ruleType, layerName, jurisdiction, extendsLine, featureSection)

		path := filepath.Join(mos, "rules", ruleType, fmt.Sprintf("rule-%04d.mos", i))
		if err := mkdirWrite(path, content); err != nil {
			return err
		}
	}

	return nil
}

func writeLocalizedContracts(mos, subdir string, count int, p Profile, kw func(string) string) error {
	i := 0
	for i < count {
		// Every 5th contract is an umbrella contract-of-contracts with 2-4
		// nested sub-contracts and a dependency DAG.
		if i%5 == 0 && i+3 < count {
			if err := writeUmbrellaContract(mos, subdir, i, p, kw); err != nil {
				return err
			}
			i += 4 // umbrella consumes 4 IDs (parent + 3 children)
			continue
		}

		if err := writeFlatContract(mos, subdir, i, p, kw); err != nil {
			return err
		}
		i++
	}

	return nil
}

func writeFlatContract(mos, subdir string, i int, p Profile, kw func(string) string) error {
	id := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), i)
	ruleRef := fmt.Sprintf("R-%04d", i%p.Rules)

	status := "active"
	if subdir == "archive" {
		status = "complete"
	}

	// Flat contracts reference previous contracts for linear dependency chains
	var depsField string
	if i > 0 && i%3 != 0 {
		prevID := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), i-1)
		depsField = fmt.Sprintf("  depends_on = [\"%s\"]\n", prevID)
	} else {
		depsField = "  depends_on = []\n"
	}

	planet := "terra"
	if i%2 == 1 {
		planet = "mars"
	}
	ghIssue := 10000 + i
	jiraKey := fmt.Sprintf("DUNIX-%d", 20000+i)

	content := fmt.Sprintf(`%s "%s" {
  title = "Generated charter %d"
  status = "%s"
%s
  bill {
    introduced_by = "terra-mars-joint-committee"
    introduced_at = "2026-01-01"
    intent = "Cross-planetary governance"
  }

  tracker "github" {
    repo   = "dunix/%s-kernel"
    issue  = %d
    labels = ["%s", "charter"]
    sync   = "bidirectional"
  }

  tracker "jira" {
    project = "DUNIX"
    key     = "%s"
    type    = "story"
    sync    = "bidirectional"
  }

  %s "Generated charter" {
    %s "acceptance" {
      sut = "charters/%s"
      test = "charter_%04d_test.go"
      %s {
        the charter is active
      }
      %s {
        deliverables are reviewed
      }
      %s {
        the charter passes
      }
    }
  }

  execution {
    rules_override = ["%s"]
  }
}
`, kw("contract"), id, i, status, depsField,
		planet, ghIssue, planet,
		jiraKey,
		kw("feature"), kw("scenario"), subdir, i,
		kw("given"), kw("when"), kw("then"), ruleRef)

	dir := filepath.Join(mos, "contracts", subdir, id)
	path := filepath.Join(dir, "contract.mos")
	return mkdirWrite(path, content)
}

// writeUmbrellaContract generates a contract-of-contracts with 3 nested
// sub-contracts forming a non-linear dependency DAG:
//
//	PARENT (umbrella, ordering = "non-linear")
//	├── CHILD-A (no deps, can start immediately)
//	├── CHILD-B (no deps, can start in parallel with A)
//	└── CHILD-C (depends on A and B -- diamond join)
func writeUmbrellaContract(mos, subdir string, baseIdx int, p Profile, kw func(string) string) error {
	parentID := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), baseIdx)
	childA := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), baseIdx+1)
	childB := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), baseIdx+2)
	childC := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), baseIdx+3)

	status := "active"
	if subdir == "archive" {
		status = "complete"
	}

	childStatus := status
	ruleRef := fmt.Sprintf("R-%04d", baseIdx%p.Rules)

	makeFeature := func(title string, idx int) string {
		return fmt.Sprintf(`    %s "%s" {
      %s "acceptance" {
        sut = "charters/%s"
        test = "charter_%04d_test.go"
        %s {
          the sub-charter is active
        }
        %s {
          deliverables are reviewed
        }
        %s {
          the sub-charter passes
        }
      }
    }`, kw("feature"), title, kw("scenario"), subdir, idx,
			kw("given"), kw("when"), kw("then"))
	}

	ghEpic := 30000 + baseIdx
	jiraEpic := fmt.Sprintf("DUNIX-%d", 40000+baseIdx)
	jiraA := fmt.Sprintf("DUNIX-%d", 40000+baseIdx+1)
	jiraB := fmt.Sprintf("DUNIX-%d", 40000+baseIdx+2)
	jiraC := fmt.Sprintf("DUNIX-%d", 40000+baseIdx+3)

	content := fmt.Sprintf(`%s "%s" {
  title = "Umbrella charter %d"
  status = "%s"
  ordering = "non-linear"
  depends_on = []

  bill {
    introduced_by = "terra-mars-joint-committee"
    introduced_at = "2026-01-01"
    intent = "Cross-planetary work package"
  }

  tracker "github" {
    repo   = "dunix/kernel"
    issue  = %d
    labels = ["epic", "cross-planetary"]
    sync   = "bidirectional"
  }

  tracker "jira" {
    project = "DUNIX"
    key     = "%s"
    type    = "epic"
    sync    = "bidirectional"
  }

  execution {
    rules_override = ["%s"]
  }

  # Sub-charter A: independent, can start immediately
  %s "%s" {
    title = "Sub-charter A (independent)"
    status = "%s"
    depends_on = []
    sequence = 1

    tracker "jira" {
      project = "DUNIX"
      key     = "%s"
      type    = "story"
      epic    = "%s"
    }

%s
  }

  # Sub-charter B: independent, can run in parallel with A
  %s "%s" {
    title = "Sub-charter B (parallel with A)"
    status = "%s"
    depends_on = []
    sequence = 1

    tracker "jira" {
      project = "DUNIX"
      key     = "%s"
      type    = "story"
      epic    = "%s"
    }

%s
  }

  # Sub-charter C: diamond join, depends on both A and B
  %s "%s" {
    title = "Sub-charter C (diamond join)"
    status = "%s"
    depends_on = ["%s", "%s"]
    sequence = 2

    tracker "jira" {
      project = "DUNIX"
      key     = "%s"
      type    = "story"
      epic    = "%s"
      blocked_by = ["%s", "%s"]
    }

%s
  }
}
`,
		kw("contract"), parentID, baseIdx, status,
		ghEpic,
		jiraEpic,
		ruleRef,
		kw("contract"), childA, childStatus,
		jiraA, jiraEpic,
		makeFeature("Sub-charter A", baseIdx+1),
		kw("contract"), childB, childStatus,
		jiraB, jiraEpic,
		makeFeature("Sub-charter B", baseIdx+2),
		kw("contract"), childC, childStatus, childA, childB,
		jiraC, jiraEpic, jiraA, jiraB,
		makeFeature("Sub-charter C (join)", baseIdx+3))

	dir := filepath.Join(mos, "contracts", subdir, parentID)
	path := filepath.Join(dir, "contract.mos")
	return mkdirWrite(path, content)
}

func mkdirWrite(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func writeConfig(mos string) error {
	return mkdirWrite(filepath.Join(mos, "config.mos"), `config {
  mos {
    version = 1
  }

  backend {
    type = "git"
  }

  governance {
    model = "committee"
  }
}
`)
}

func writeDeclaration(mos string) error {
	return mkdirWrite(filepath.Join(mos, "declaration.mos"), `declaration {
  name = "stress-test-project"
  created = "2026-01-01"
  authors = ["generator"]
}
`)
}

func writeLexicon(mos string, p Profile) error {
	var b strings.Builder
	b.WriteString("lexicon {\n  terms {\n")
	for i := 0; i < p.LexiconTerms; i++ {
		fmt.Fprintf(&b, "    term_%04d = \"Definition of term %d\"\n", i, i)
	}
	b.WriteString("  }\n\n  artifact_labels {\n    rule = \"Rule\"\n    contract = \"Contract\"\n  }\n}\n")
	return mkdirWrite(filepath.Join(mos, "lexicon", "default.mos"), b.String())
}

func writeLayers(mos string, p Profile) error {
	var b strings.Builder
	b.WriteString("layers {\n")
	for i := 0; i < p.ResolutionLayers; i++ {
		fmt.Fprintf(&b, "  layer \"layer_%d\" {\n    level = %d\n  }\n", i, i+1)
	}
	b.WriteString("}\n")
	return mkdirWrite(filepath.Join(mos, "resolution", "layers.mos"), b.String())
}

func writeRules(mos string, p Profile) error {
	includeIdx := 0
	for i := 0; i < p.Rules; i++ {
		ruleType := "mechanical"
		if i%3 == 0 {
			ruleType = "interpretive"
		}

		layerIdx := i % p.ResolutionLayers
		var featureSection string

		if includeIdx < p.IncludeFiles && i%2 == 0 {
			featureName := fmt.Sprintf("spec_%04d.feature", i)
			featurePath := filepath.Join(mos, "rules", ruleType, featureName)
			featureContent := fmt.Sprintf(`feature "Generated rule" {
  scenario "baseline" {
    sut = "rules/%s"
    test = "rule_%04d_test.go"
    given {
      a project
    }
    when {
      the rule is evaluated
    }
    then {
      it passes
    }
  }
}
`, ruleType, i)
			if err := mkdirWrite(featurePath, featureContent); err != nil {
				return err
			}
			featureSection = fmt.Sprintf("  spec {\n    include %q\n  }", featureName)
			includeIdx++
		} else {
			featureSection = fmt.Sprintf(`  feature "Generated rule" {
    scenario "baseline" {
      sut = "rules/%s"
      test = "rule_%04d_test.go"
      given {
        a project
      }
      when {
        the rule is evaluated
      }
      then {
        it passes
      }
    }
  }`, ruleType, i)
		}

		content := fmt.Sprintf(`rule "R-%04d" {
  name = "Generated rule %d"
  type = "%s"
  scope = "layer_%d"
  enforcement = "error"

%s
}
`, i, i, ruleType, layerIdx, featureSection)

		path := filepath.Join(mos, "rules", ruleType, fmt.Sprintf("rule-%04d.mos", i))
		if err := mkdirWrite(path, content); err != nil {
			return err
		}
	}

	return nil
}

func writeContracts(mos, subdir string, count int, p Profile) error {
	for i := 0; i < count; i++ {
		id := fmt.Sprintf("CON-%s-%04d", strings.ToUpper(subdir[:3]), i)
		ruleRef := fmt.Sprintf("R-%04d", i%p.Rules)

		status := "active"
		if subdir == "archive" {
			status = "complete"
		}

		content := fmt.Sprintf(`contract "%s" {
  title = "Generated contract %d"
  status = "%s"

  bill {
    introduced_by = "generator"
    introduced_at = "2026-01-01"
    intent = "Stress testing"
  }

  feature "Generated contract" {
    scenario "acceptance" {
      sut = "contracts/%s"
      test = "contract_%04d_test.go"
      given {
        the contract is active
      }
      when {
        deliverables are reviewed
      }
      then {
        the contract passes
      }
    }
  }

  execution {
    rules_override = ["%s"]
  }
}
`, id, i, status, subdir, i, ruleRef)

		dir := filepath.Join(mos, "contracts", subdir, id)
		path := filepath.Join(dir, "contract.mos")
		if err := mkdirWrite(path, content); err != nil {
			return err
		}
	}

	return nil
}
