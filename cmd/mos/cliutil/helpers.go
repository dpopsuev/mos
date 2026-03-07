package cliutil

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/dpopsuev/mos/moslib/registry"
)

// ErrNonZeroExit signals a non-zero exit code without printing an error message.
// Used by lint to indicate "findings present" (exit code 1).
var ErrNonZeroExit = errors.New("")

// ErrInternalLint signals an internal linter error (exit code 2).
// Distinguished from ErrNonZeroExit so the CLI can use a different exit code.
var ErrInternalLint = errors.New("internal linter error")

// CoreKindDirs maps core kind names to their on-disk directory names.
var CoreKindDirs = map[string]string{
	names.KindContract:      names.DirContracts,
	names.KindSpecification: names.DirSpecifications,
	names.KindRule:          names.DirRules,
}

// ApplyOverflowFields applies CAD-driven overflow fields via GenericUpdate.
// Used by dedicated handlers to pass through unknown --flags to the generic layer.
func ApplyOverflowFields(kind, id string, overflow map[string]string) error {
	if len(overflow) == 0 {
		return nil
	}
	if kind == names.KindRule {
		if err := artifact.UpdateRuleFields(".", id, overflow); err != nil {
			return fmt.Errorf("mos %s update: %w", kind, err)
		}
		return nil
	}
	reg, err := registry.LoadRegistry(".")
	if err != nil {
		return fmt.Errorf("mos %s update: loading registry for overflow fields: %w", kind, err)
	}
	td, ok := reg.Types[kind]
	if !ok {
		dir, known := CoreKindDirs[kind]
		if !known {
			return fmt.Errorf("mos %s update: kind %q not found in registry", kind, kind)
		}
		td = registry.ArtifactTypeDef{Kind: kind, Directory: dir}
	}
	if err := artifact.GenericUpdate(".", td, id, overflow); err != nil {
		return fmt.Errorf("mos %s update: %w", kind, err)
	}
	return nil
}

// ParseKVArgs extracts --key value pairs and positional arguments from a raw
// arg slice. Both "--key value" and "--key=value" forms are accepted.
func ParseKVArgs(args []string) (flags map[string]string, positional []string) {
	flags = make(map[string]string)
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			key := strings.TrimPrefix(args[i], "--")
			if idx := strings.IndexByte(key, '='); idx >= 0 {
				flags[key[:idx]] = key[idx+1:]
			} else if i+1 < len(args) {
				flags[key] = args[i+1]
				i++
			}
		} else if !strings.HasPrefix(args[i], "-") {
			positional = append(positional, args[i])
		}
	}
	return
}

// ExtractPositional extracts the first positional argument, respecting "--" as end-of-flags.
func ExtractPositional(args []string, usage string) (string, error) {
	endOfFlags := false
	for _, arg := range args {
		if arg == "--" {
			endOfFlags = true
			continue
		}
		if !endOfFlags && (arg == "--help" || arg == "-h") {
			return "", fmt.Errorf("usage: %s", usage)
		}
		if !endOfFlags && strings.HasPrefix(arg, "-") {
			return "", fmt.Errorf("unknown flag %q\n  usage: %s", arg, usage)
		}
		return arg, nil
	}
	return "", fmt.Errorf("usage: %s", usage)
}
