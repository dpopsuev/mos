package harness

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// ScenarioMatch records whether a contract scenario has a matching test.
type ScenarioMatch struct {
	ContractID    string `json:"contract_id"`
	ScenarioTitle string `json:"scenario_title"`
	TestFunc      string `json:"test_func,omitempty"`
	Matched       bool   `json:"matched"`
}

// MatchScenarios reads contract artifacts from mosDir, extracts scenario
// titles, and scans Go test function names in root to find matches.
func MatchScenarios(root, mosDir string) ([]ScenarioMatch, error) {
	scenarios := collectScenarios(mosDir)
	if len(scenarios) == 0 {
		return nil, nil
	}

	testFuncs, err := collectTestFuncs(root)
	if err != nil {
		return nil, err
	}

	normalIndex := make(map[string]string, len(testFuncs))
	for _, fn := range testFuncs {
		normalIndex[strings.ToLower(fn)] = fn
	}

	var matches []ScenarioMatch
	for _, sc := range scenarios {
		normalized := normalizeToTestName(sc.ScenarioTitle)
		m := ScenarioMatch{
			ContractID:    sc.ContractID,
			ScenarioTitle: sc.ScenarioTitle,
		}
		if fn, ok := normalIndex[strings.ToLower(normalized)]; ok {
			m.Matched = true
			m.TestFunc = fn
		}
		matches = append(matches, m)
	}
	return matches, nil
}

// ScenarioCoverage returns (matched, total) counts from scenario matches.
func ScenarioCoverage(matches []ScenarioMatch) (int, int) {
	matched := 0
	for _, m := range matches {
		if m.Matched {
			matched++
		}
	}
	return matched, len(matches)
}

type rawScenario struct {
	ContractID    string
	ScenarioTitle string
}

func collectScenarios(mosDir string) []rawScenario {
	contractDir := filepath.Join(mosDir, "contracts", "active")
	entries, err := os.ReadDir(contractDir)
	if err != nil {
		return nil
	}

	var out []rawScenario
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		path := filepath.Join(contractDir, e.Name(), "contract.mos")
		ab, err := dsl.ReadArtifact(path)
		if err != nil {
			continue
		}
		contractID := ab.Name
		for _, item := range ab.Items {
			fb, ok := item.(*dsl.FeatureBlock)
			if !ok {
				continue
			}
			for _, g := range fb.Groups {
				switch sc := g.(type) {
				case *dsl.Scenario:
					out = append(out, rawScenario{ContractID: contractID, ScenarioTitle: sc.Name})
				case *dsl.Group:
					for _, s := range sc.Scenarios {
						out = append(out, rawScenario{ContractID: contractID, ScenarioTitle: s.Name})
					}
				}
			}
		}
	}
	return out
}

var nonAlphaNum = regexp.MustCompile(`[^a-zA-Z0-9]+`)

// normalizeToTestName converts a scenario title like
// "mos audit reports convergence rate per axis" to "TestMosAuditReportsConvergenceRatePerAxis".
func normalizeToTestName(title string) string {
	words := nonAlphaNum.Split(title, -1)
	var b strings.Builder
	b.WriteString("Test")
	for _, w := range words {
		if w == "" {
			continue
		}
		runes := []rune(w)
		runes[0] = unicode.ToUpper(runes[0])
		b.WriteString(string(runes))
	}
	return b.String()
}

func collectTestFuncs(root string) ([]string, error) {
	var funcs []string
	fset := token.NewFileSet()

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == ".git" || (strings.HasPrefix(d.Name(), ".") && d.Name() != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			if strings.HasPrefix(fn.Name.Name, "Test") {
				funcs = append(funcs, fn.Name.Name)
			}
		}
		return nil
	})
	return funcs, err
}
