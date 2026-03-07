package gatecmd

import (
	"testing"

	"github.com/dpopsuev/mos/moslib/linter"
)

func TestFilterBySeverity(t *testing.T) {
	diags := []linter.Diagnostic{
		{Rule: "r1", Severity: linter.SeverityError},
		{Rule: "r2", Severity: linter.SeverityWarning},
		{Rule: "r3", Severity: linter.SeverityInfo},
		{Rule: "r4", Severity: linter.SeverityError},
	}

	tests := []struct {
		severity string
		want     int
	}{
		{"error", 2},
		{"warning", 1},
		{"info", 1},
		{"Error", 2},
	}
	for _, tt := range tests {
		got := filterBySeverity(diags, tt.severity)
		if len(got) != tt.want {
			t.Errorf("filterBySeverity(%q) returned %d diags, want %d", tt.severity, len(got), tt.want)
		}
	}
}

func TestFilterBySeverity_Empty(t *testing.T) {
	got := filterBySeverity(nil, "error")
	if len(got) != 0 {
		t.Errorf("expected empty result, got %d", len(got))
	}
}

func TestSeverityRank(t *testing.T) {
	if severityRank("error") >= severityRank("warning") {
		t.Error("error should rank before warning")
	}
	if severityRank("warning") >= severityRank("info") {
		t.Error("warning should rank before info")
	}
}
