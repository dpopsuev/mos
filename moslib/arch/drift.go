package arch

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// LayerDef describes packages belonging to a named layer and their allowed dependencies.
type LayerDef struct {
	Name     string
	Level    int
	Packages []string
}

// SubToolDef maps a sub-tool name to its constituent packages.
type SubToolDef struct {
	Name     string
	Packages []string
}

// ExceptionDef allows a specific cross-boundary edge that would otherwise be forbidden.
type ExceptionDef struct {
	From string
	To   string
	Via  string // optional: through interface
}

// DesiredArch is the parsed desired architecture model.
type DesiredArch struct {
	Layers     []LayerDef
	SubTools   []SubToolDef
	Exceptions []ExceptionDef
}

// DriftViolation describes a single layer or sub-tool boundary violation.
type DriftViolation struct {
	From    string
	To      string
	Rule    string
	Message string
}

// ParseDesiredArch extracts structured layer/sub-tool rules from a desired
// architecture artifact. It first looks for structured blocks (layer, sub_tool,
// exception) and falls back to free-text section parsing for older artifacts.
func ParseDesiredArch(ab *dsl.ArtifactBlock) DesiredArch {
	var da DesiredArch

	dsl.WalkBlocks(ab.Items, func(b *dsl.Block) bool {
		switch b.Name {
		case "layer":
			da.Layers = append(da.Layers, parseLayerBlock(b))
		case "sub_tool":
			da.SubTools = append(da.SubTools, parseSubToolBlock(b))
		case "exception":
			da.Exceptions = append(da.Exceptions, parseExceptionBlock(b))
		}
		return false
	})

	if len(da.Layers) > 0 || len(da.SubTools) > 0 {
		return da
	}

	// Fallback: parse legacy free-text sections.
	dsl.WalkBlocks(ab.Items, func(b *dsl.Block) bool {
		if b.Name != "section" {
			return false
		}
		text, _ := dsl.FieldString(b.Items, "text")
		switch b.Title {
		case "Layer Rules":
			da.Layers = parseLayerRulesText(text)
		case "Sub-Tools":
			da.SubTools = parseSubToolsText(text)
		case "Forbidden Patterns":
			da.Exceptions = parseExceptionsText(text)
		}
		return false
	})

	return da
}

func parseLayerBlock(b *dsl.Block) LayerDef {
	ld := LayerDef{Name: b.Title}
	if lvl, ok := dsl.FieldString(b.Items, "level"); ok {
		ld.Level, _ = strconv.Atoi(lvl)
	}
	if pkgs, ok := dsl.FieldString(b.Items, "packages"); ok {
		for _, p := range strings.Split(pkgs, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				ld.Packages = append(ld.Packages, p)
			}
		}
	}
	return ld
}

func parseSubToolBlock(b *dsl.Block) SubToolDef {
	st := SubToolDef{Name: b.Title}
	if pkgs, ok := dsl.FieldString(b.Items, "packages"); ok {
		for _, p := range strings.Split(pkgs, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				st.Packages = append(st.Packages, p)
			}
		}
	}
	return st
}

func parseExceptionBlock(b *dsl.Block) ExceptionDef {
	ex := ExceptionDef{}
	ex.From, _ = dsl.FieldString(b.Items, "from")
	ex.To, _ = dsl.FieldString(b.Items, "to")
	ex.Via, _ = dsl.FieldString(b.Items, "via")
	return ex
}

// --- Legacy text parsers (fallback for old-format artifacts) ---

func parseLayerRulesText(text string) []LayerDef {
	var layers []LayerDef
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" || !strings.HasPrefix(sentence, "Layer ") {
			continue
		}
		var level int
		var labelAndRest string
		_, err := fmt.Sscanf(sentence, "Layer %d %s", &level, &labelAndRest)
		if err != nil {
			continue
		}
		colonIdx := strings.Index(sentence, ":")
		if colonIdx < 0 {
			continue
		}

		nameEnd := strings.Index(sentence, "(")
		var layerName string
		if nameEnd > 0 {
			parenEnd := strings.Index(sentence, ")")
			if parenEnd > nameEnd {
				layerName = strings.TrimSpace(sentence[nameEnd+1 : parenEnd])
			}
		}

		rest := sentence[colonIdx+1:]
		dashIdx := strings.Index(rest, "\u2014")
		if dashIdx < 0 {
			dashIdx = strings.Index(rest, "-")
		}
		pkgPart := rest
		if dashIdx > 0 {
			pkgPart = rest[:dashIdx]
		}

		var pkgs []string
		for _, p := range strings.Split(pkgPart, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				pkgs = append(pkgs, p)
			}
		}

		layers = append(layers, LayerDef{
			Name:     layerName,
			Level:    level,
			Packages: pkgs,
		})
	}
	return layers
}

func parseSubToolsText(text string) []SubToolDef {
	var tools []SubToolDef
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence == "" {
			continue
		}
		colonIdx := strings.Index(sentence, ":")
		if colonIdx < 0 {
			continue
		}
		toolName := strings.TrimSpace(sentence[:colonIdx])
		parenIdx := strings.Index(toolName, "(")
		if parenIdx > 0 {
			toolName = strings.TrimSpace(toolName[:parenIdx])
		}

		rest := sentence[colonIdx+1:]
		dashIdx := strings.Index(rest, "\u2014")
		if dashIdx > 0 {
			rest = rest[dashIdx+len("\u2014"):]
		}

		var pkgs []string
		for _, p := range strings.Split(rest, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				clean := strings.Fields(p)
				if len(clean) > 0 {
					pkgs = append(pkgs, clean[0])
				}
			}
		}

		tools = append(tools, SubToolDef{
			Name:     toolName,
			Packages: pkgs,
		})
	}
	return tools
}

func parseExceptionsText(text string) []ExceptionDef {
	var exceptions []ExceptionDef
	if idx := strings.Index(text, "except:"); idx >= 0 {
		rest := text[idx+len("except:"):]
		endIdx := strings.Index(rest, ")")
		if endIdx > 0 {
			rest = rest[:endIdx]
		}
		parts := strings.Split(rest, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			words := strings.Fields(p)
			if len(words) >= 4 {
				from := words[0]
				to := ""
				via := ""
				for i, w := range words {
					if w == "call" && i+1 < len(words) {
						to = words[i+1]
					}
					if w == "via" && i+1 < len(words) {
						via = strings.Join(words[i+1:], " ")
						break
					}
				}
				exceptions = append(exceptions, ExceptionDef{From: from, To: to, Via: via})
			}
		}
	}
	return exceptions
}

// DetectDrift compares a live architecture model against desired layer rules.
// Returns violations where imports cross layer or sub-tool boundaries illegally.
func DetectDrift(live ArchModel, desired DesiredArch) []DriftViolation {
	pkgLayer := make(map[string]int)
	for _, l := range desired.Layers {
		for _, p := range l.Packages {
			pkgLayer[p] = l.Level
		}
	}

	pkgTool := make(map[string]string)
	for _, t := range desired.SubTools {
		for _, p := range t.Packages {
			pkgTool[p] = t.Name
		}
	}

	exceptionSet := make(map[[2]string]bool)
	for _, ex := range desired.Exceptions {
		if ex.From != "" && ex.To != "" {
			exceptionSet[[2]string{ex.From, ex.To}] = true
		}
	}

	var violations []DriftViolation
	for _, e := range live.Edges {
		fromPkg := normalizePackage(e.From)
		toPkg := normalizePackage(e.To)

		fromLayer, fromKnown := pkgLayer[fromPkg]
		toLayer, toKnown := pkgLayer[toPkg]

		if !fromKnown || !toKnown {
			continue
		}

		if fromLayer < toLayer {
			violations = append(violations, DriftViolation{
				From:    e.From,
				To:      e.To,
				Rule:    "layer-inversion",
				Message: fmt.Sprintf("Layer %d (%s) imports Layer %d (%s) \u2014 lower layers must not depend on higher layers", fromLayer, fromPkg, toLayer, toPkg),
			})
			continue
		}

		if fromLayer == 2 && toLayer == 2 {
			fromTool := pkgTool[fromPkg]
			toTool := pkgTool[toPkg]
			if fromTool != "" && toTool != "" && fromTool != toTool {
				if exceptionSet[[2]string{fromPkg, toPkg}] {
					continue
				}
				violations = append(violations, DriftViolation{
					From:    e.From,
					To:      e.To,
					Rule:    "cross-tool",
					Message: fmt.Sprintf("%s (%s) imports %s (%s) \u2014 Layer 2 packages from different sub-tools must not directly depend on each other", fromPkg, fromTool, toPkg, toTool),
				})
			}
		}
	}

	return violations
}

// normalizePackage strips common prefixes from a package path to match layer definitions.
func normalizePackage(pkg string) string {
	parts := strings.Split(pkg, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return pkg
}

// ReadArchitectureFn reads an architecture artifact by ID. Injected to avoid
// import cycles between arch and artifact.
var ReadArchitectureFn func(root, id string) (*dsl.ArtifactBlock, error)

// LoadDesiredArch reads and parses the ARCH-desired artifact from a project root.
func LoadDesiredArch(root string) (*DesiredArch, error) {
	if ReadArchitectureFn == nil {
		return nil, fmt.Errorf("ReadArchitectureFn not initialized")
	}
	ab, err := ReadArchitectureFn(root, "ARCH-desired")
	if err != nil {
		return nil, fmt.Errorf("read ARCH-desired: %w", err)
	}

	da := ParseDesiredArch(ab)
	return &da, nil
}
