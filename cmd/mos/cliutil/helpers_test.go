package cliutil

import (
	"strings"
	"testing"
)

func TestParseKVArgs(t *testing.T) {
	t.Run("flags only", func(t *testing.T) {
		flags, pos := ParseKVArgs([]string{"--foo", "bar", "--baz", "qux"})
		if len(pos) != 0 {
			t.Errorf("expected no positional; got %v", pos)
		}
		if flags["foo"] != "bar" || flags["baz"] != "qux" {
			t.Errorf("expected foo=bar, baz=qux; got %v", flags)
		}
	})

	t.Run("positional only", func(t *testing.T) {
		flags, pos := ParseKVArgs([]string{"a", "b", "c"})
		if len(flags) != 0 {
			t.Errorf("expected no flags; got %v", flags)
		}
		if want := []string{"a", "b", "c"}; strings.Join(pos, ",") != strings.Join(want, ",") {
			t.Errorf("expected positional %v; got %v", want, pos)
		}
	})

	t.Run("mixed", func(t *testing.T) {
		flags, pos := ParseKVArgs([]string{"--key", "val", "pos1", "--other", "v2", "pos2"})
		if flags["key"] != "val" || flags["other"] != "v2" {
			t.Errorf("expected key=val, other=v2; got %v", flags)
		}
		if want := []string{"pos1", "pos2"}; strings.Join(pos, ",") != strings.Join(want, ",") {
			t.Errorf("expected positional %v; got %v", want, pos)
		}
	})

	t.Run("key=value format", func(t *testing.T) {
		flags, pos := ParseKVArgs([]string{"--name=alice", "--id=42", "extra"})
		if flags["name"] != "alice" || flags["id"] != "42" {
			t.Errorf("expected name=alice, id=42; got %v", flags)
		}
		if want := []string{"extra"}; strings.Join(pos, ",") != strings.Join(want, ",") {
			t.Errorf("expected positional %v; got %v", want, pos)
		}
	})
}

func TestExtractPositional(t *testing.T) {
	t.Run("valid args", func(t *testing.T) {
		got, err := ExtractPositional([]string{"CON-001"}, "mos contract show <id>")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "CON-001" {
			t.Errorf("expected CON-001; got %q", got)
		}
	})

	t.Run("no args", func(t *testing.T) {
		_, err := ExtractPositional([]string{}, "mos contract show <id>")
		if err == nil {
			t.Fatal("expected error when no args")
		}
		if !strings.Contains(err.Error(), "usage") {
			t.Errorf("expected usage in error; got %v", err)
		}
	})

	t.Run("help flag", func(t *testing.T) {
		_, err := ExtractPositional([]string{"--help"}, "mos contract show <id>")
		if err == nil {
			t.Fatal("expected error when --help passed")
		}
		if !strings.Contains(err.Error(), "usage") {
			t.Errorf("expected usage in error; got %v", err)
		}
	})
}

func TestApplyOverflowFieldsEmpty(t *testing.T) {
	err := ApplyOverflowFields("contract", "CON-001", nil)
	if err != nil {
		t.Errorf("expected nil for empty overflow; got %v", err)
	}
	err = ApplyOverflowFields("contract", "CON-001", map[string]string{})
	if err != nil {
		t.Errorf("expected nil for empty overflow map; got %v", err)
	}
}
