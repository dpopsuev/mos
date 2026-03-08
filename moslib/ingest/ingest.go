package ingest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/store"
)

const (
	filesBucket = "files"
	metaBucket  = "meta"
)

type Request struct {
	Path        string   `json:"path"`
	Kind        string   `json:"kind"`
	Description string   `json:"description"`
	RelatesTo   []string `json:"relates_to,omitempty"`
}

type Result struct {
	NodeID     store.NodeID `json:"node_id"`
	Kind       string       `json:"kind"`
	Size       int          `json:"size"`
	EdgesAdded int          `json:"edges_added"`
}

type fileMeta struct {
	Kind        string    `json:"kind"`
	Description string    `json:"description"`
	IngestedAt  time.Time `json:"ingested_at"`
	Size        int       `json:"size"`
}

var dependsOnRe = regexp.MustCompile(`depends_on\s*=\s*\[([^\]]*)\]`)

func Ingest(ctx context.Context, s store.Store, req Request) (*Result, error) {
	absPath, err := filepath.Abs(req.Path)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	nodeID := store.NodeID(absPath)

	if err := s.Put(ctx, filesBucket, absPath, content); err != nil {
		return nil, fmt.Errorf("store content: %w", err)
	}

	meta := fileMeta{
		Kind:        req.Kind,
		Description: req.Description,
		IngestedAt:  time.Now(),
		Size:        len(content),
	}
	metaJSON, _ := json.Marshal(meta)
	if err := s.Put(ctx, metaBucket, absPath, metaJSON); err != nil {
		return nil, fmt.Errorf("store meta: %w", err)
	}

	edgesAdded := 0

	for _, rel := range req.RelatesTo {
		if err := s.AddEdge(ctx, nodeID, store.NodeID(rel), "relates_to", nil); err != nil {
			return nil, fmt.Errorf("add relates_to edge: %w", err)
		}
		edgesAdded++
	}

	if req.Kind == "contract" {
		deps := parseDependsOn(string(content))
		for _, dep := range deps {
			if err := s.AddEdge(ctx, nodeID, store.NodeID(dep), "depends_on", nil); err != nil {
				return nil, fmt.Errorf("add depends_on edge: %w", err)
			}
			edgesAdded++
		}
	}

	return &Result{
		NodeID:     nodeID,
		Kind:       req.Kind,
		Size:       len(content),
		EdgesAdded: edgesAdded,
	}, nil
}

func parseDependsOn(content string) []string {
	m := dependsOnRe.FindStringSubmatch(content)
	if len(m) < 2 {
		return nil
	}
	raw := m[1]
	var deps []string
	for _, part := range strings.Split(raw, ",") {
		s := strings.TrimSpace(part)
		s = strings.Trim(s, `"'`)
		if s != "" {
			deps = append(deps, s)
		}
	}
	return deps
}
