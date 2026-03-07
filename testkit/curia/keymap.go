package main

import (
	"os"
	"strings"
)

type Action int

const (
	ActionNone Action = iota
	ActionQuit
	ActionUp
	ActionDown
	ActionPageUp
	ActionPageDown
	ActionHome
	ActionEnd
	ActionExpand
	ActionSwitchPanel
)

type KeySeq []string

type Keymap struct {
	Bindings map[Action][]KeySeq
}

type MatchResult int

const (
	MatchNone    MatchResult = iota
	MatchPartial
	MatchExact
)

// Match checks pending keystrokes against all bindings. Returns the matched
// action and whether the match is exact, partial (prefix of a longer
// sequence), or none.
func (km *Keymap) Match(pending []string) (Action, MatchResult) {
	if len(pending) == 0 {
		return ActionNone, MatchNone
	}
	bestAction := ActionNone
	hasPartial := false
	for action, seqs := range km.Bindings {
		for _, seq := range seqs {
			if len(seq) == 0 {
				continue
			}
			if seqEqual(pending, seq) {
				bestAction = action
				return bestAction, MatchExact
			}
			if isPrefix(pending, seq) {
				hasPartial = true
			}
		}
	}
	if hasPartial {
		return ActionNone, MatchPartial
	}
	return bestAction, MatchNone
}

func seqEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// isPrefix returns true if pending is a proper prefix of seq (shorter and matches).
func isPrefix(pending, seq []string) bool {
	if len(pending) >= len(seq) {
		return false
	}
	for i := range pending {
		if pending[i] != seq[i] {
			return false
		}
	}
	return true
}

func baseBindings() map[Action][]KeySeq {
	return map[Action][]KeySeq{
		ActionQuit:        {{"q"}, {"ctrl+c"}},
		ActionUp:          {{"up"}},
		ActionDown:        {{"down"}},
		ActionPageUp:      {{"pgup"}},
		ActionPageDown:    {{"pgdown"}},
		ActionHome:        {{"home"}},
		ActionEnd:         {{"end"}},
		ActionExpand:      {{"enter"}},
		ActionSwitchPanel: {{"tab"}},
	}
}

func mergeBindings(base, extra map[Action][]KeySeq) map[Action][]KeySeq {
	out := make(map[Action][]KeySeq, len(base)+len(extra))
	for a, seqs := range base {
		out[a] = append(out[a], seqs...)
	}
	for a, seqs := range extra {
		out[a] = append(out[a], seqs...)
	}
	return out
}

func VSCodeKeymap() *Keymap {
	b := baseBindings()
	b[ActionUp] = append(b[ActionUp], KeySeq{"k"})
	b[ActionDown] = append(b[ActionDown], KeySeq{"j"})
	b[ActionExpand] = append(b[ActionExpand], KeySeq{" "})
	return &Keymap{Bindings: b}
}

func VimKeymap() *Keymap {
	base := baseBindings()
	vim := map[Action][]KeySeq{
		ActionUp:          {{"k"}},
		ActionDown:        {{"j"}},
		ActionPageUp:      {{"ctrl+u"}},
		ActionPageDown:    {{"ctrl+d"}},
		ActionHome:        {{"g", "g"}},
		ActionEnd:         {{"G"}},
		ActionExpand:      {{"o"}, {" "}},
		ActionSwitchPanel: {{"ctrl+w", "ctrl+w"}},
	}
	return &Keymap{Bindings: mergeBindings(base, vim)}
}

func EmacsKeymap() *Keymap {
	base := baseBindings()
	emacs := map[Action][]KeySeq{
		ActionQuit:        {{"ctrl+x", "ctrl+c"}},
		ActionUp:          {{"ctrl+p"}},
		ActionDown:        {{"ctrl+n"}},
		ActionPageUp:      {{"alt+v"}},
		ActionPageDown:    {{"ctrl+v"}},
		ActionHome:        {{"alt+<"}},
		ActionEnd:         {{"alt+>"}},
		ActionExpand:      {{" "}},
		ActionSwitchPanel: {{"ctrl+o"}},
	}
	return &Keymap{Bindings: mergeBindings(base, emacs)}
}

func DetectKeymap() *Keymap {
	for _, env := range []string{"EDITOR", "VISUAL"} {
		v := strings.ToLower(os.Getenv(env))
		if v == "" {
			continue
		}
		if strings.Contains(v, "vim") || strings.Contains(v, "nvim") {
			return VimKeymap()
		}
		if strings.Contains(v, "emacs") {
			return EmacsKeymap()
		}
	}
	return VSCodeKeymap()
}

func ResolveKeymap(preset string) *Keymap {
	switch strings.ToLower(preset) {
	case "vim":
		return VimKeymap()
	case "emacs":
		return EmacsKeymap()
	case "vscode":
		return VSCodeKeymap()
	default:
		return DetectKeymap()
	}
}

// ParseKeySeq splits a binding string into a KeySeq. Space-separated tokens
// represent multi-key sequences: "g g" -> ["g","g"], "ctrl+x ctrl+c" ->
// ["ctrl+x","ctrl+c"]. A single token is a single-key sequence.
func ParseKeySeq(s string) KeySeq {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return nil
	}
	return KeySeq(parts)
}

var actionNames = map[string]Action{
	"quit":         ActionQuit,
	"up":           ActionUp,
	"down":         ActionDown,
	"page_up":      ActionPageUp,
	"page_down":    ActionPageDown,
	"home":         ActionHome,
	"end":          ActionEnd,
	"expand":       ActionExpand,
	"switch_panel": ActionSwitchPanel,
}

// ApplyOverrides replaces bindings for actions specified in the overrides map.
// Keys are action names (e.g. "quit"), values are lists of binding strings.
func (km *Keymap) ApplyOverrides(overrides map[string][]string) {
	for name, bindings := range overrides {
		action, ok := actionNames[name]
		if !ok {
			continue
		}
		var seqs []KeySeq
		for _, b := range bindings {
			if seq := ParseKeySeq(b); len(seq) > 0 {
				seqs = append(seqs, seq)
			}
		}
		if len(seqs) > 0 {
			km.Bindings[action] = seqs
		}
	}
}
