package main

import (
	"os"
	"testing"
)

func TestVimSingleKey(t *testing.T) {
	km := VimKeymap()
	action, result := km.Match([]string{"j"})
	if result != MatchExact || action != ActionDown {
		t.Fatalf("vim j: got action=%d result=%d, want Down=%d Exact=%d", action, result, ActionDown, MatchExact)
	}
}

func TestVimMultiKeyPartial(t *testing.T) {
	km := VimKeymap()
	_, result := km.Match([]string{"g"})
	if result != MatchPartial {
		t.Fatalf("vim g: got result=%d, want Partial=%d", result, MatchPartial)
	}
}

func TestVimMultiKeyExact(t *testing.T) {
	km := VimKeymap()
	action, result := km.Match([]string{"g", "g"})
	if result != MatchExact || action != ActionHome {
		t.Fatalf("vim gg: got action=%d result=%d, want Home=%d Exact=%d", action, result, ActionHome, MatchExact)
	}
}

func TestVimShiftG(t *testing.T) {
	km := VimKeymap()
	action, result := km.Match([]string{"G"})
	if result != MatchExact || action != ActionEnd {
		t.Fatalf("vim G: got action=%d result=%d, want End=%d Exact=%d", action, result, ActionEnd, MatchExact)
	}
}

func TestEmacsMultiKeyPartial(t *testing.T) {
	km := EmacsKeymap()
	_, result := km.Match([]string{"ctrl+x"})
	if result != MatchPartial {
		t.Fatalf("emacs ctrl+x: got result=%d, want Partial=%d", result, MatchPartial)
	}
}

func TestEmacsMultiKeyQuit(t *testing.T) {
	km := EmacsKeymap()
	action, result := km.Match([]string{"ctrl+x", "ctrl+c"})
	if result != MatchExact || action != ActionQuit {
		t.Fatalf("emacs ctrl+x ctrl+c: got action=%d result=%d, want Quit=%d Exact=%d", action, result, ActionQuit, MatchExact)
	}
}

func TestEmacsCtrlN(t *testing.T) {
	km := EmacsKeymap()
	action, result := km.Match([]string{"ctrl+n"})
	if result != MatchExact || action != ActionDown {
		t.Fatalf("emacs ctrl+n: got action=%d result=%d, want Down=%d Exact=%d", action, result, ActionDown, MatchExact)
	}
}

func TestVSCodeDefaults(t *testing.T) {
	km := VSCodeKeymap()
	tests := []struct {
		keys   []string
		action Action
	}{
		{[]string{"up"}, ActionUp},
		{[]string{"down"}, ActionDown},
		{[]string{"enter"}, ActionExpand},
		{[]string{"tab"}, ActionSwitchPanel},
		{[]string{"q"}, ActionQuit},
	}
	for _, tt := range tests {
		action, result := km.Match(tt.keys)
		if result != MatchExact || action != tt.action {
			t.Errorf("vscode %v: got action=%d result=%d, want %d Exact", tt.keys, action, result, tt.action)
		}
	}
}

func TestMatchNone(t *testing.T) {
	km := VSCodeKeymap()
	_, result := km.Match([]string{"z"})
	if result != MatchNone {
		t.Fatalf("vscode z: got result=%d, want None=%d", result, MatchNone)
	}
}

func TestAutoDetectVim(t *testing.T) {
	os.Setenv("EDITOR", "nvim")
	defer os.Unsetenv("EDITOR")
	km := DetectKeymap()
	action, result := km.Match([]string{"j"})
	if result != MatchExact || action != ActionDown {
		t.Fatalf("auto(nvim) j: got action=%d result=%d, want Down Exact", action, result)
	}
	_, result = km.Match([]string{"g"})
	if result != MatchPartial {
		t.Fatalf("auto(nvim) g: got result=%d, want Partial", result)
	}
}

func TestAutoDetectEmacs(t *testing.T) {
	os.Setenv("EDITOR", "emacs")
	defer os.Unsetenv("EDITOR")
	km := DetectKeymap()
	action, result := km.Match([]string{"ctrl+n"})
	if result != MatchExact || action != ActionDown {
		t.Fatalf("auto(emacs) ctrl+n: got action=%d result=%d, want Down Exact", action, result)
	}
}

func TestAutoDetectFallback(t *testing.T) {
	os.Setenv("EDITOR", "nano")
	defer os.Unsetenv("EDITOR")
	km := DetectKeymap()
	action, result := km.Match([]string{"up"})
	if result != MatchExact || action != ActionUp {
		t.Fatalf("auto(nano) up: got action=%d result=%d, want Up Exact", action, result)
	}
}

func TestParseKeySeq(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"q", 1},
		{"g g", 2},
		{"ctrl+x ctrl+c", 2},
		{"", 0},
	}
	for _, tt := range tests {
		seq := ParseKeySeq(tt.input)
		if len(seq) != tt.want {
			t.Errorf("ParseKeySeq(%q): got len=%d, want %d", tt.input, len(seq), tt.want)
		}
	}
}

func TestApplyOverrides(t *testing.T) {
	km := VSCodeKeymap()
	km.ApplyOverrides(map[string][]string{
		"quit": {"x", "ctrl+q"},
		"home": {"g g"},
	})
	action, result := km.Match([]string{"x"})
	if result != MatchExact || action != ActionQuit {
		t.Fatalf("override x: got action=%d result=%d, want Quit Exact", action, result)
	}
	_, result = km.Match([]string{"g"})
	if result != MatchPartial {
		t.Fatalf("override g (prefix of gg): got result=%d, want Partial", result)
	}
	action, result = km.Match([]string{"g", "g"})
	if result != MatchExact || action != ActionHome {
		t.Fatalf("override gg: got action=%d result=%d, want Home Exact", action, result)
	}
}

func TestResolveKeymap(t *testing.T) {
	for _, preset := range []string{"vim", "emacs", "vscode", "auto"} {
		km := ResolveKeymap(preset)
		if km == nil {
			t.Errorf("ResolveKeymap(%q) returned nil", preset)
		}
	}
}
