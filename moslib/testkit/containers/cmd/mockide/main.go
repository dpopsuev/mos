package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dpopsuev/mos/moslib/store"
)

func main() {
	socket := flag.String("socket", "/data/mosbus.sock", "Unix socket path")
	workspace := flag.String("workspace", "/workspace/default", "Workspace root")
	tag := flag.String("tag", "ide-0", "IDE identifier tag for logging")
	flag.Parse()

	log.SetPrefix(fmt.Sprintf("[%s] ", *tag))

	var client store.Store
	for i := 0; i < 30; i++ {
		var err error
		client, err = store.Dial(*socket, []string{*workspace})
		if err == nil {
			break
		}
		log.Printf("waiting for daemon... (%v)", err)
		time.Sleep(time.Second)
	}
	if client == nil {
		log.Fatal("could not connect to daemon")
	}
	defer client.Close()

	ctx := context.Background()
	tag_ := *tag

	client.Put(ctx, "ide-data", tag_+"-key1", []byte(tag_+"-value1"))
	client.Put(ctx, "ide-data", tag_+"-key2", []byte(tag_+"-value2"))
	client.AddEdge(ctx, store.NodeID(tag_+"-A"), store.NodeID(tag_+"-B"), "depends_on", nil)

	val, _ := client.Get(ctx, "ide-data", tag_+"-key1")
	items, _ := client.List(ctx, "ide-data", tag_+"-")
	edges, _ := client.Neighbors(ctx, store.NodeID(tag_+"-A"), "depends_on", store.Outgoing)

	result := map[string]any{
		"tag":       tag_,
		"workspace": *workspace,
		"get_value": string(val),
		"list_count": len(items),
		"edge_count": len(edges),
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))

	os.WriteFile(fmt.Sprintf("/data/result-%s.json", tag_), data, 0o644)
	log.Printf("scenario complete: %d items, %d edges", len(items), len(edges))
}
