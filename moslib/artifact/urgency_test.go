package artifact

import (
	"testing"
)

func TestCON042_UrgencyPropagationParsedFromRegistry(t *testing.T) {
	root := setupScaffold(t)
	reg := loadTestRegistry(t, root)
	td := reg.Types["need"]

	if td.Lifecycle.UrgencyPropagation == nil {
		t.Fatal("need lifecycle should have urgency_propagation")
	}
	if td.Lifecycle.UrgencyPropagation["critical"] != "error" {
		t.Errorf("critical should map to error, got %q", td.Lifecycle.UrgencyPropagation["critical"])
	}
	if td.Lifecycle.UrgencyPropagation["high"] != "warn" {
		t.Errorf("high should map to warn, got %q", td.Lifecycle.UrgencyPropagation["high"])
	}
	if td.Lifecycle.UrgencyPropagation["medium"] != "info" {
		t.Errorf("medium should map to info, got %q", td.Lifecycle.UrgencyPropagation["medium"])
	}
	if td.Lifecycle.UrgencyPropagation["low"] != "ignore" {
		t.Errorf("low should map to ignore, got %q", td.Lifecycle.UrgencyPropagation["low"])
	}
}

// --- CON-2026-047: Config management CLI ---
