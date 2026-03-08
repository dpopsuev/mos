package mos_test

import (
	"fmt"
	"testing"

	"github.com/dpopsuev/mos/moslib/survey"
)

func TestCouplingDebug(t *testing.T) {
	sc := &survey.PackagesScanner{Fallback: &survey.GoScanner{}}
	proj, err := sc.Scan(".")
	if err != nil {
		t.Fatal(err)
	}
	if proj.DependencyGraph == nil {
		t.Fatal("no dependency graph")
	}

	withCoupling := 0
	for _, e := range proj.DependencyGraph.Edges {
		if e.CallSites > 0 || e.LOCSurface > 0 {
			withCoupling++
			if withCoupling <= 5 {
				fmt.Printf("  %s -> %s  weight=%d call_sites=%d loc_surface=%d\n",
					e.From, e.To, e.Weight, e.CallSites, e.LOCSurface)
			}
		}
	}
	fmt.Printf("Total edges: %d, with coupling: %d\n", len(proj.DependencyGraph.Edges), withCoupling)
	if withCoupling == 0 {
		t.Error("expected some edges with coupling data")
	}
}
