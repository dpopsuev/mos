package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dpopsuev/mos/moslib/model"
)

const keyTimeoutDuration = 300 * time.Millisecond

type panel int

const (
	panelTree panel = iota
	panelGraph
)

type keyTimeoutMsg struct {
	seq int
}

type appModel struct {
	mod         *model.Project
	nodes       []treeNode
	cursor      int
	treeScroll  int
	expanded    map[int]bool
	activePanel panel
	graphScroll int
	width       int
	height      int
	quitting    bool

	keymap     *Keymap
	styles     *Styles
	pending    []string
	pendingSeq int
}

func newAppModel(mod *model.Project, km *Keymap, styles *Styles) appModel {
	m := appModel{
		mod:      mod,
		expanded: make(map[int]bool),
		width:    80,
		height:   24,
		keymap:   km,
		styles:   styles,
	}
	m.nodes = flattenTree(mod, m.expanded)
	return m
}

func (m appModel) Init() tea.Cmd {
	return nil
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.processKey(msg)
	case keyTimeoutMsg:
		if msg.seq == m.pendingSeq && len(m.pending) > 0 {
			m.pending = nil
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}
	return m, nil
}

func (m appModel) processKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	m.pending = append(m.pending, key)
	m.pendingSeq++

	action, result := m.keymap.Match(m.pending)
	switch result {
	case MatchExact:
		m.pending = nil
		return m.handleAction(action)
	case MatchPartial:
		seq := m.pendingSeq
		return m, func() tea.Msg {
			time.Sleep(keyTimeoutDuration)
			return keyTimeoutMsg{seq: seq}
		}
	default:
		// No match with full pending buffer. Try just the current key alone.
		m.pending = []string{key}
		action, result = m.keymap.Match(m.pending)
		if result == MatchExact {
			m.pending = nil
			return m.handleAction(action)
		}
		if result == MatchPartial {
			seq := m.pendingSeq
			return m, func() tea.Msg {
				time.Sleep(keyTimeoutDuration)
				return keyTimeoutMsg{seq: seq}
			}
		}
		m.pending = nil
		return m, nil
	}
}

func (m appModel) handleAction(action Action) (tea.Model, tea.Cmd) {
	switch action {
	case ActionQuit:
		m.quitting = true
		return m, tea.Quit

	case ActionSwitchPanel:
		if m.activePanel == panelTree {
			m.activePanel = panelGraph
		} else {
			m.activePanel = panelTree
		}
		return m, nil

	case ActionUp:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			if m.cursor > 0 {
				m.cursor--
			}
			m.treeScroll = ensureVisible(m.cursor, m.treeScroll, m.treeBodyHeight())
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			if m.graphScroll > 0 {
				m.graphScroll--
			}
		}
		return m, nil

	case ActionDown:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			if m.cursor < len(m.nodes)-1 {
				m.cursor++
			}
			m.treeScroll = ensureVisible(m.cursor, m.treeScroll, m.treeBodyHeight())
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			pkgIdx := m.selectedPkgIdx()
			maxScroll := graphLineCount(m.mod, pkgIdx) - m.graphBodyHeight()
			if maxScroll < 0 {
				maxScroll = 0
			}
			if m.graphScroll < maxScroll {
				m.graphScroll++
			}
		}
		return m, nil

	case ActionPageUp:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			m.cursor -= m.treeBodyHeight()
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.treeScroll = ensureVisible(m.cursor, m.treeScroll, m.treeBodyHeight())
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			m.graphScroll -= m.graphBodyHeight()
			if m.graphScroll < 0 {
				m.graphScroll = 0
			}
		}
		return m, nil

	case ActionPageDown:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			m.cursor += m.treeBodyHeight()
			if m.cursor >= len(m.nodes) {
				m.cursor = len(m.nodes) - 1
			}
			m.treeScroll = ensureVisible(m.cursor, m.treeScroll, m.treeBodyHeight())
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			pkgIdx := m.selectedPkgIdx()
			maxScroll := graphLineCount(m.mod, pkgIdx) - m.graphBodyHeight()
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.graphScroll += m.graphBodyHeight()
			if m.graphScroll > maxScroll {
				m.graphScroll = maxScroll
			}
		}
		return m, nil

	case ActionHome:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			m.cursor = 0
			m.treeScroll = 0
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			m.graphScroll = 0
		}
		return m, nil

	case ActionEnd:
		if m.activePanel == panelTree {
			prevPkg := m.selectedPkgIdx()
			if len(m.nodes) > 0 {
				m.cursor = len(m.nodes) - 1
			}
			m.treeScroll = ensureVisible(m.cursor, m.treeScroll, m.treeBodyHeight())
			if m.selectedPkgIdx() != prevPkg {
				m.graphScroll = 0
			}
		} else {
			pkgIdx := m.selectedPkgIdx()
			maxScroll := graphLineCount(m.mod, pkgIdx) - m.graphBodyHeight()
			if maxScroll < 0 {
				maxScroll = 0
			}
			m.graphScroll = maxScroll
		}
		return m, nil

	case ActionExpand:
		if m.activePanel == panelTree && len(m.nodes) > 0 {
			n := m.nodes[m.cursor]
			if n.kind == nodePackage {
				m.expanded[n.pkgIdx] = !m.expanded[n.pkgIdx]
				m.nodes = flattenTree(m.mod, m.expanded)
				if m.cursor >= len(m.nodes) {
					m.cursor = len(m.nodes) - 1
				}
			}
		}
		m.graphScroll = 0
		return m, nil
	}
	return m, nil
}

func (m appModel) selectedPkgIdx() int {
	if len(m.nodes) == 0 || m.cursor < 0 || m.cursor >= len(m.nodes) {
		return -1
	}
	return m.nodes[m.cursor].pkgIdx
}

func (m appModel) selectedSymbol() *model.Symbol {
	if len(m.nodes) == 0 || m.cursor < 0 || m.cursor >= len(m.nodes) {
		return nil
	}
	n := m.nodes[m.cursor]
	if n.kind != nodeSymbol {
		return nil
	}
	return n.symbol
}

func (m appModel) leftWidth() int {
	w := m.width * 3 / 5
	if w < 20 {
		w = m.width / 2
	}
	return w
}

func (m appModel) rightWidth() int {
	return m.width - m.leftWidth()
}

func (m appModel) panelTotalHeight() int {
	h := m.height - 1
	if h < 3 {
		h = 3
	}
	return h
}

func (m appModel) contentHeight() int {
	h := m.panelTotalHeight() - 2
	if h < 1 {
		h = 1
	}
	return h
}

func (m appModel) treeBodyHeight() int {
	h := m.contentHeight() - 1
	if h < 1 {
		h = 1
	}
	return h
}

func (m appModel) graphBodyHeight() int {
	h := m.contentHeight() - 1
	if h < 1 {
		h = 1
	}
	return h
}

func (m appModel) View() string {
	if m.quitting {
		return ""
	}

	s := m.styles
	lw := m.leftWidth()
	rw := m.rightWidth()
	cH := m.contentHeight()
	cWL := lw - 2
	cWR := rw - 2
	if cWL < 1 {
		cWL = 1
	}
	if cWR < 1 {
		cWR = 1
	}

	bodyH := cH - 1
	if bodyH < 1 {
		bodyH = 1
	}

	edgeCount := 0
	if m.mod.DependencyGraph != nil {
		edgeCount = len(m.mod.DependencyGraph.Edges)
	}
	treeTitle := s.Title.Render(fmt.Sprintf(" %s (%d pkgs, %d edges)", m.mod.Path, len(m.mod.Namespaces), edgeCount))
	treeBody := renderTree(m.nodes, m.cursor, m.treeScroll, cWL, bodyH, s)
	treeContent := treeTitle + "\n" + treeBody

	pkgIdx := m.selectedPkgIdx()
	var graphTitle string
	if pkgIdx >= 0 && pkgIdx < len(m.mod.Namespaces) {
		pkg := m.mod.Namespaces[pkgIdx]
		out := 0
		in := 0
		if m.mod.DependencyGraph != nil {
			out = len(m.mod.DependencyGraph.EdgesFrom(pkg.ImportPath))
			for _, e := range m.mod.DependencyGraph.Edges {
				if e.To == pkg.ImportPath {
					in++
				}
			}
		}
		graphTitle = s.Title.Render(fmt.Sprintf(" %s (%d→ %d←)", shortPath(m.mod.Path, pkg.ImportPath), out, in))
	} else {
		graphTitle = s.Title.Render(" Import Graph")
	}
	var graphBody string
	if sym := m.selectedSymbol(); sym != nil && len(sym.Dependencies) > 0 {
		graphBody = renderGraphForSymbol(m.mod, sym, m.graphScroll, cWR, bodyH, s)
	}
	if graphBody == "" {
		graphBody = renderGraph(m.mod, pkgIdx, m.graphScroll, cWR, bodyH, s)
	}
	graphContent := graphTitle + "\n" + graphBody

	var leftStyle, rightStyle lipgloss.Style
	if m.activePanel == panelTree {
		leftStyle = s.ActiveBorder.Width(cWL).Height(cH)
		rightStyle = s.InactiveBorder.Width(cWR).Height(cH)
	} else {
		leftStyle = s.InactiveBorder.Width(cWL).Height(cH)
		rightStyle = s.ActiveBorder.Width(cWR).Height(cH)
	}

	left := leftStyle.Render(treeContent)
	right := rightStyle.Render(graphContent)
	panels := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	help := m.renderHelp()
	return panels + "\n" + help
}

func (m appModel) renderHelp() string {
	s := m.styles
	type binding struct{ key, desc string }
	bindings := []binding{
		{"q", "quit"},
		{"tab", "panel"},
		{"↑↓", "navigate"},
		{"pgup/dn", "page"},
	}
	if m.activePanel == panelTree {
		bindings = append(bindings, binding{"enter", "expand/collapse"})
	}
	var parts []string
	for _, b := range bindings {
		parts = append(parts, s.HelpKey.Render(b.key)+" "+s.HelpDesc.Render(b.desc))
	}
	return " " + strings.Join(parts, "  │  ")
}
