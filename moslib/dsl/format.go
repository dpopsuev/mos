package dsl

import (
	"fmt"
	"slices"
	"strings"
)

// Format produces canonical DSL text from an AST.
// If kw is nil, English defaults are used. The KeywordMap's ToHuman side
// translates machine-protocol names back to human-protocol keywords.
func Format(f *File, kw *KeywordMap) string {
	if kw == nil {
		kw = DefaultKeywords()
	}
	var b strings.Builder
	fmtr := &formatter{w: &b, kw: kw}
	fmtr.file(f)
	return b.String()
}

type formatter struct {
	w     *strings.Builder
	kw    *KeywordMap
	depth int
}

func (f *formatter) indent() string {
	return strings.Repeat("  ", f.depth)
}

func (f *formatter) file(file *File) {
	if file.Artifact != nil {
		f.node(file.Artifact)
	}
	f.w.WriteString("\n")
}

func (f *formatter) node(n Node) {
	switch v := n.(type) {
	case *ArtifactBlock:
		f.artifactBlock(v)
	case *Field:
		f.field(v)
	case *Block:
		f.block(v)
	case *SpecBlock:
		f.specBlock(v)
	case *FeatureBlock:
		f.featureBlock(v)
	}
}

func (f *formatter) artifactBlock(ab *ArtifactBlock) {
	humanKind := f.kw.humanKeyword(ab.Kind)
	f.w.WriteString(humanKind)
	if ab.Name != "" {
		fmt.Fprintf(f.w, " %q", ab.Name)
	}
	f.w.WriteString(" {\n")
	f.depth++
	f.blockItems(ab.Items)
	f.depth--
	f.w.WriteString("}\n")
}

func (f *formatter) blockItems(items []Node) {
	prevWasBlock := false
	for _, item := range items {
		switch item.(type) {
		case *Block, *SpecBlock, *FeatureBlock:
			if prevWasBlock || len(items) > 1 {
				f.w.WriteString("\n")
			}
			prevWasBlock = true
		default:
			prevWasBlock = false
		}
		f.node(item)
	}
}

func (f *formatter) field(fld *Field) {
	fmt.Fprintf(f.w, "%s%s = %s\n", f.indent(), fld.Key, f.formatValue(fld.Value))
}

func (f *formatter) formatValue(v Value) string {
	switch val := v.(type) {
	case *StringVal:
		if val.Triple || strings.Contains(val.Text, "\n") {
			text := strings.TrimRight(val.Text, "\n")
			return `"""` + "\n" + text + `"""`
		}
		return fmt.Sprintf("%q", val.Text)
	case *IntegerVal:
		return val.Raw
	case *FloatVal:
		return val.Raw
	case *BoolVal:
		if val.Val {
			return "true"
		}
		return "false"
	case *DateTimeVal:
		return val.Raw
	case *ListVal:
		return f.formatList(val)
	case *InlineTableVal:
		return f.formatInlineTable(val)
	default:
		return "???"
	}
}

func (f *formatter) formatList(l *ListVal) string {
	if len(l.Items) == 0 {
		return "[]"
	}

	parts := make([]string, len(l.Items))
	totalLen := 2 // brackets
	for i, item := range l.Items {
		parts[i] = f.formatValue(item)
		totalLen += len(parts[i]) + 2
	}

	if totalLen < 80 && !slices.ContainsFunc(parts, func(s string) bool { return strings.Contains(s, "\n") }) {
		return "[" + strings.Join(parts, ", ") + "]"
	}

	var b strings.Builder
	b.WriteString("[\n")
	for _, p := range parts {
		fmt.Fprintf(&b, "%s  %s,\n", f.indent(), p)
	}
	fmt.Fprintf(&b, "%s]", f.indent())
	return b.String()
}

func (f *formatter) formatInlineTable(t *InlineTableVal) string {
	if len(t.Fields) == 0 {
		return "{}"
	}

	parts := make([]string, len(t.Fields))
	totalLen := 4 // braces + spaces
	for i, fld := range t.Fields {
		parts[i] = fmt.Sprintf("%s = %s", fld.Key, f.formatValue(fld.Value))
		totalLen += len(parts[i]) + 2
	}

	if totalLen < 80 {
		return "{ " + strings.Join(parts, ", ") + " }"
	}

	var b strings.Builder
	b.WriteString("{\n")
	for _, p := range parts {
		fmt.Fprintf(&b, "%s  %s,\n", f.indent(), p)
	}
	fmt.Fprintf(&b, "%s}", f.indent())
	return b.String()
}

func (f *formatter) block(blk *Block) {
	if blk.Title != "" {
		fmt.Fprintf(f.w, "%s%s %q {\n", f.indent(), blk.Name, blk.Title)
	} else {
		fmt.Fprintf(f.w, "%s%s {\n", f.indent(), blk.Name)
	}
	f.depth++
	if blk.Name == "acceptance" {
		f.acceptanceItems(blk.Items)
	} else {
		f.blockItems(blk.Items)
	}
	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) acceptanceItems(items []Node) {
	prevWasBlock := false
	for _, item := range items {
		switch v := item.(type) {
		case *Block:
			if prevWasBlock || len(items) > 1 {
				f.w.WriteString("\n")
			}
			prevWasBlock = true
			f.criterionBlock(v)
		default:
			prevWasBlock = false
			f.node(item)
		}
	}
}

func (f *formatter) criterionBlock(blk *Block) {
	name := blk.Name
	title := blk.Title
	if name != "criterion" && title == "" {
		title = name
		name = "criterion"
	}
	if title != "" {
		fmt.Fprintf(f.w, "%s%s %q {\n", f.indent(), name, title)
	} else {
		fmt.Fprintf(f.w, "%s%s {\n", f.indent(), name)
	}
	f.depth++
	f.blockItems(blk.Items)
	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) specBlock(sb *SpecBlock) {
	humanSpec := f.kw.humanKeyword("spec")
	humanInclude := f.kw.humanKeyword("include")
	fmt.Fprintf(f.w, "%s%s {\n", f.indent(), humanSpec)
	f.depth++

	for _, inc := range sb.Includes {
		fmt.Fprintf(f.w, "%s%s %q\n", f.indent(), humanInclude, inc.Path)
	}
	for i, feat := range sb.Features {
		if i > 0 || len(sb.Includes) > 0 {
			f.w.WriteString("\n")
		}
		f.featureBlock(feat)
	}

	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) featureBlock(fb *FeatureBlock) {
	humanFeature := f.kw.humanKeyword("feature")
	fmt.Fprintf(f.w, "%s%s %q {\n", f.indent(), humanFeature, fb.Name)
	f.depth++

	for _, desc := range fb.Description {
		fmt.Fprintf(f.w, "%s%s\n", f.indent(), desc)
	}

	if fb.Background != nil {
		if len(fb.Description) > 0 {
			f.w.WriteString("\n")
		}
		f.backgroundBlock(fb.Background)
	}

	for _, g := range fb.Groups {
		f.w.WriteString("\n")
		switch v := g.(type) {
		case *Scenario:
			f.scenarioBlock(v)
		case *Group:
			f.groupBlock(v)
		}
	}

	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) backgroundBlock(bg *Background) {
	humanBg := f.kw.humanKeyword("background")
	fmt.Fprintf(f.w, "%s%s {\n", f.indent(), humanBg)
	f.depth++
	if bg.Given != nil {
		f.stepBlock("given", bg.Given)
	}
	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) groupBlock(g *Group) {
	humanGroup := f.kw.humanKeyword("group")
	fmt.Fprintf(f.w, "%s%s %q {\n", f.indent(), humanGroup, g.Name)
	f.depth++
	for i, sc := range g.Scenarios {
		if i > 0 {
			f.w.WriteString("\n")
		}
		f.scenarioBlock(sc)
	}
	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) scenarioBlock(sc *Scenario) {
	humanScenario := f.kw.humanKeyword("scenario")
	fmt.Fprintf(f.w, "%s%s %q {\n", f.indent(), humanScenario, sc.Name)
	f.depth++

	for _, fld := range sc.Fields {
		f.field(fld)
	}

	if sc.Given != nil {
		f.stepBlock("given", sc.Given)
	}
	if sc.When != nil {
		f.stepBlock("when", sc.When)
	}
	if sc.Then != nil {
		f.stepBlock("then", sc.Then)
	}

	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}

func (f *formatter) stepBlock(machineKeyword string, sb *StepBlock) {
	human := f.kw.humanKeyword(machineKeyword)
	fmt.Fprintf(f.w, "%s%s {\n", f.indent(), human)
	f.depth++
	for _, line := range sb.Lines {
		fmt.Fprintf(f.w, "%s%s\n", f.indent(), line)
	}
	f.depth--
	fmt.Fprintf(f.w, "%s}\n", f.indent())
}
