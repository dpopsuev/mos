package export

import (
	"github.com/dpopsuev/mos/moslib/mesh"
)

// GraphDelta captures the difference between two graph snapshots.
type GraphDelta struct {
	AddedNodes   []mesh.Node `json:"added_nodes,omitempty"`
	RemovedNodes []mesh.Node `json:"removed_nodes,omitempty"`
	ModifiedNodes []NodeDiff `json:"modified_nodes,omitempty"`
	AddedEdges   []mesh.Edge `json:"added_edges,omitempty"`
	RemovedEdges []mesh.Edge `json:"removed_edges,omitempty"`
}

// NodeDiff records a node whose metadata changed between snapshots.
type NodeDiff struct {
	ID      string            `json:"id"`
	OldMeta map[string]string `json:"old_meta,omitempty"`
	NewMeta map[string]string `json:"new_meta,omitempty"`
}

// DiffGraphs computes the delta between two graph snapshots.
// A node is "added" if present in newG but not oldG, "removed" if vice versa,
// and "modified" if present in both but with different Meta maps.
func DiffGraphs(oldG, newG *mesh.Graph) *GraphDelta {
	d := &GraphDelta{}

	oldNodes := indexNodes(oldG)
	newNodes := indexNodes(newG)

	for id, n := range newNodes {
		old, exists := oldNodes[id]
		if !exists {
			d.AddedNodes = append(d.AddedNodes, n)
			continue
		}
		if !metaEqual(old.Meta, n.Meta) {
			d.ModifiedNodes = append(d.ModifiedNodes, NodeDiff{
				ID:      id,
				OldMeta: old.Meta,
				NewMeta: n.Meta,
			})
		}
	}
	for id, n := range oldNodes {
		if _, exists := newNodes[id]; !exists {
			d.RemovedNodes = append(d.RemovedNodes, n)
		}
	}

	type ek struct{ from, to, rel string }
	oldEdges := map[ek]mesh.Edge{}
	newEdges := map[ek]mesh.Edge{}
	for _, e := range oldG.Edges {
		oldEdges[ek{e.From, e.To, e.Relation}] = e
	}
	for _, e := range newG.Edges {
		newEdges[ek{e.From, e.To, e.Relation}] = e
	}

	for k, e := range newEdges {
		if _, exists := oldEdges[k]; !exists {
			d.AddedEdges = append(d.AddedEdges, e)
		}
	}
	for k, e := range oldEdges {
		if _, exists := newEdges[k]; !exists {
			d.RemovedEdges = append(d.RemovedEdges, e)
		}
	}

	return d
}

func indexNodes(g *mesh.Graph) map[string]mesh.Node {
	m := make(map[string]mesh.Node, len(g.Nodes))
	for _, n := range g.Nodes {
		m[n.ID] = n
	}
	return m
}

func metaEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
