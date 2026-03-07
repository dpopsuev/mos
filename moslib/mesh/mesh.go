package mesh

// Node represents a vertex in the context mesh graph.
type Node struct {
	ID    string            `json:"id"`
	Kind  string            `json:"kind"`
	Label string            `json:"label"`
	Meta  map[string]string `json:"meta,omitempty"`
}

// Edge represents a directed relationship between two nodes.
type Edge struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

// Graph is the context mesh: a typed, directed graph linking code packages
// to governance artifacts (specs, contracts, needs, sprints).
type Graph struct {
	Nodes []Node `json:"nodes"`
	Edges []Edge `json:"edges"`

	nodeSet map[string]bool
	edgeSet map[edgeKey]bool
}

type edgeKey struct {
	from, to, rel string
}

// AddNode adds a node if one with the same ID doesn't already exist.
func (g *Graph) AddNode(n Node) {
	if g.nodeSet == nil {
		g.nodeSet = make(map[string]bool)
	}
	if g.nodeSet[n.ID] {
		return
	}
	g.nodeSet[n.ID] = true
	g.Nodes = append(g.Nodes, n)
}

// AddEdge adds an edge if an identical (from, to, relation) doesn't already exist.
func (g *Graph) AddEdge(e Edge) {
	if g.edgeSet == nil {
		g.edgeSet = make(map[edgeKey]bool)
	}
	k := edgeKey{e.From, e.To, e.Relation}
	if g.edgeSet[k] {
		return
	}
	g.edgeSet[k] = true
	g.Edges = append(g.Edges, e)
}

// NodesOfKind returns all nodes matching the given kind.
func (g *Graph) NodesOfKind(kind string) []Node {
	var out []Node
	for _, n := range g.Nodes {
		if n.Kind == kind {
			out = append(out, n)
		}
	}
	return out
}

// EdgesFrom returns all edges originating from the given node ID.
func (g *Graph) EdgesFrom(id string) []Edge {
	var out []Edge
	for _, e := range g.Edges {
		if e.From == id {
			out = append(out, e)
		}
	}
	return out
}

// EdgesTo returns all edges pointing to the given node ID.
func (g *Graph) EdgesTo(id string) []Edge {
	var out []Edge
	for _, e := range g.Edges {
		if e.To == id {
			out = append(out, e)
		}
	}
	return out
}
