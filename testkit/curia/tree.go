package main

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/model"
)

type nodeKind int

const (
	nodePackage nodeKind = iota
	nodeFile
	nodeSymbol
)

type treeNode struct {
	label  string
	depth  int
	kind   nodeKind
	pkgIdx int
	symbol *model.Symbol
}

func flattenTree(mod *model.Project, expanded map[int]bool) []treeNode {
	var nodes []treeNode
	for i, pkg := range mod.Namespaces {
		short := shortPath(mod.Path, pkg.ImportPath)
		arrow := "▸"
		if expanded[i] {
			arrow = "▾"
		}
		nodes = append(nodes, treeNode{
			label:  fmt.Sprintf("%s %s (%d files, %d syms)", arrow, short, len(pkg.Files), len(pkg.Symbols)),
			depth:  0,
			kind:   nodePackage,
			pkgIdx: i,
		})
		if !expanded[i] {
			continue
		}
		for _, f := range pkg.Files {
			parts := strings.Split(f.Path, "/")
			nodes = append(nodes, treeNode{
				label:  parts[len(parts)-1],
				depth:  1,
				kind:   nodeFile,
				pkgIdx: i,
			})
		}
		for _, s := range pkg.Symbols {
			vis := "+"
			if !s.Exported {
				vis = "-"
			}
			nodes = append(nodes, treeNode{
				label:  fmt.Sprintf("[%s] %s%s", s.Kind, vis, s.Name),
				depth:  1,
				kind:   nodeSymbol,
				pkgIdx: i,
				symbol: s,
			})
		}
	}
	return nodes
}

func renderTree(nodes []treeNode, cursor, scroll, width, height int, s *Styles) string {
	if len(nodes) == 0 {
		return s.NoData.Render("(no packages)")
	}

	end := scroll + height
	if end > len(nodes) {
		end = len(nodes)
	}
	if scroll > len(nodes) {
		scroll = len(nodes)
	}

	var lines []string
	for i := scroll; i < end; i++ {
		n := nodes[i]
		indent := strings.Repeat("  ", n.depth)

		prefix := "  "
		if i == cursor {
			prefix = s.Cursor.Render("▸ ")
		}

		text := indent + n.label
		runes := []rune(text)
		if len(runes) > width-3 && width > 6 {
			text = string(runes[:width-6]) + "…"
		}

		var styled string
		switch n.kind {
		case nodePackage:
			styled = s.Package.Render(text)
		case nodeFile:
			styled = s.File.Render(text)
		case nodeSymbol:
			if n.symbol != nil && n.symbol.Exported {
				styled = s.Exported.Render(text)
			} else {
				styled = s.Unexported.Render(text)
			}
		}
		lines = append(lines, prefix+styled)
	}
	return strings.Join(lines, "\n")
}

func shortPath(modPath, importPath string) string {
	if importPath == modPath {
		return "."
	}
	if strings.HasPrefix(importPath, modPath+"/") {
		return strings.TrimPrefix(importPath, modPath+"/")
	}
	return importPath
}

func ensureVisible(cursor, scroll, height int) int {
	if height <= 0 {
		return 0
	}
	if cursor < scroll {
		return cursor
	}
	if cursor >= scroll+height {
		return cursor - height + 1
	}
	return scroll
}
