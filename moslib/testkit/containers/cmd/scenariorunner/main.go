package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/dpopsuev/mos/moslib/store"
)

func main() {
	workspaceRoots := strings.Split(os.Getenv("WORKSPACE_ROOTS"), ",")
	socketPath := os.Getenv("MOSBUS_SOCKET")
	scenarioID := os.Getenv("SCENARIO_ID")

	if socketPath == "" || len(workspaceRoots) == 0 || workspaceRoots[0] == "" {
		fmt.Fprintln(os.Stderr, "MOSBUS_SOCKET and WORKSPACE_ROOTS required")
		os.Exit(1)
	}

	client, err := store.Dial(socketPath, workspaceRoots)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dial: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	ctx := context.Background()

	client.Put(ctx, "scenario", "id", []byte(scenarioID))
	client.Put(ctx, "scenario", "workspace", []byte(strings.Join(workspaceRoots, ",")))

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("data-%s-%d", scenarioID, i)
		val := fmt.Sprintf("value-from-%s-%d", scenarioID, i)
		if err := client.Put(ctx, "testdata", key, []byte(val)); err != nil {
			fmt.Fprintf(os.Stderr, "put: %v\n", err)
			os.Exit(1)
		}
	}

	client.AddEdge(ctx, store.NodeID("node-"+scenarioID), store.NodeID("root"), "belongs_to", nil)

	items, err := client.List(ctx, "testdata", "data-"+scenarioID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "list: %v\n", err)
		os.Exit(1)
	}

	result := map[string]any{
		"scenario":   scenarioID,
		"workspace":  strings.Join(workspaceRoots, ","),
		"items_read": len(items),
		"status":     "ok",
	}
	json.NewEncoder(os.Stdout).Encode(result)
}
