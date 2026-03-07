package harness

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeToTestName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"mos audit reports convergence rate per axis", "TestMosAuditReportsConvergenceRatePerAxis"},
		{"user logs in", "TestUserLogsIn"},
		{"simple", "TestSimple"},
		{"with-dashes_and underscores", "TestWithDashesAndUnderscores"},
	}
	for _, tc := range tests {
		got := normalizeToTestName(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeToTestName(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestMatchScenarios(t *testing.T) {
	root := t.TempDir()

	contractDir := filepath.Join(root, ".mos", "contracts", "active", "CON-TEST-001")
	if err := os.MkdirAll(contractDir, 0o755); err != nil {
		t.Fatal(err)
	}

	contract := `contract "CON-TEST-001" {
  title = "Test Contract"
  status = "active"

  feature "Login" {
    scenario "user logs in" {
      given {
        the user exists
      }
      when {
        they enter credentials
      }
      then {
        they are authenticated
      }
    }
  }
}
`
	if err := os.WriteFile(filepath.Join(contractDir, "contract.mos"), []byte(contract), 0o644); err != nil {
		t.Fatal(err)
	}

	testFile := `package main

import "testing"

func TestUserLogsIn(t *testing.T) {}
func TestSomethingElse(t *testing.T) {}
`
	if err := os.WriteFile(filepath.Join(root, "login_test.go"), []byte(testFile), 0o644); err != nil {
		t.Fatal(err)
	}

	mosDir := filepath.Join(root, ".mos")
	matches, err := MatchScenarios(root, mosDir)
	if err != nil {
		t.Fatalf("MatchScenarios: %v", err)
	}

	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if !matches[0].Matched {
		t.Error("expected scenario to be matched")
	}
	if matches[0].TestFunc != "TestUserLogsIn" {
		t.Errorf("expected TestFunc=TestUserLogsIn, got %s", matches[0].TestFunc)
	}
	if matches[0].ContractID != "CON-TEST-001" {
		t.Errorf("expected ContractID=CON-TEST-001, got %s", matches[0].ContractID)
	}
}

func TestScenarioCoverage(t *testing.T) {
	matches := []ScenarioMatch{
		{Matched: true},
		{Matched: false},
		{Matched: true},
	}
	matched, total := ScenarioCoverage(matches)
	if matched != 2 || total != 3 {
		t.Errorf("expected (2, 3), got (%d, %d)", matched, total)
	}
}
