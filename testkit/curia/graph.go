package main

import (
	"strings"

	"github.com/dpopsuev/mos/moslib/model"
)

// renderGraphForSymbol shows the selected symbol's dependency edges.
// Returns empty string if the symbol has no dependencies, signaling
// the caller to fall back to package-level rendering.
func renderGraphForSymbol(mod *model.Project, sym *model.Symbol, scroll, width, height int, s *Styles) string {
	if sym == nil || len(sym.Dependencies) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, s.SectionHeader.Render("Dependencies of "+sym.Name))
	for _, dep := range sym.Dependencies {
		target := shortPath(mod.Path, dep)
		lines = append(lines, s.ExternalEdge.Render("  → "+target))
	}

	if scroll > len(lines) {
		scroll = len(lines)
	}
	end := scroll + height
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[scroll:end], "\n")
}

func renderGraph(mod *model.Project, pkgIdx, scroll, width, height int, s *Styles) string {
	if pkgIdx < 0 || pkgIdx >= len(mod.Namespaces) {
		return s.NoData.Render("(select a package)")
	}
	pkg := mod.Namespaces[pkgIdx]
	graph := mod.DependencyGraph
	if graph == nil {
		return s.NoData.Render("(no import data)")
	}

	outgoing := graph.EdgesFrom(pkg.ImportPath)
	var incoming []model.DependencyEdge
	for _, e := range graph.Edges {
		if e.To == pkg.ImportPath {
			incoming = append(incoming, e)
		}
	}

	if len(outgoing) == 0 && len(incoming) == 0 {
		return s.NoData.Render("(no imports)")
	}

	var lines []string

	if len(outgoing) > 0 {
		lines = append(lines, s.SectionHeader.Render("Imports"))
		for _, e := range outgoing {
			target := shortPath(mod.Path, e.To)
			if e.External {
				lines = append(lines, s.ExternalEdge.Render("  → "+target)+" "+s.EdgeTag.Render("ext"))
			} else {
				lines = append(lines, s.InternalEdge.Render("  → "+target))
			}
		}
	}

	if len(incoming) > 0 {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, s.SectionHeader.Render("Imported by"))
		for _, e := range incoming {
			source := shortPath(mod.Path, e.From)
			lines = append(lines, s.InternalEdge.Render("  ← "+source))
		}
	}

	if scroll > len(lines) {
		scroll = len(lines)
	}
	end := scroll + height
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[scroll:end], "\n")
}

func graphLineCount(mod *model.Project, pkgIdx int) int {
	if pkgIdx < 0 || pkgIdx >= len(mod.Namespaces) || mod.DependencyGraph == nil {
		return 0
	}
	pkg := mod.Namespaces[pkgIdx]
	n := 0

	out := mod.DependencyGraph.EdgesFrom(pkg.ImportPath)
	if len(out) > 0 {
		n += 1 + len(out) // header + edges
	}

	incoming := 0
	for _, e := range mod.DependencyGraph.Edges {
		if e.To == pkg.ImportPath {
			incoming++
		}
	}
	if incoming > 0 {
		if n > 0 {
			n++ // blank separator
		}
		n += 1 + incoming // header + edges
	}
	return n
}
