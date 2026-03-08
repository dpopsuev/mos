package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dpopsuev/mos/moslib/store"
)

func main() {
	sock := os.Getenv("MOSBUS_SOCKET")
	if sock == "" {
		sock = "/data/mosbus.sock"
	}

	d := store.NewDaemon(sock, 30*time.Minute)
	fmt.Fprintf(os.Stderr, "mosbus daemon listening on %s\n", sock)
	if err := d.ListenAndServe(); err != nil {
		fmt.Fprintf(os.Stderr, "daemon: %v\n", err)
		os.Exit(1)
	}
}
