package harness

import (
	"path/filepath"
	"time"

	"github.com/dpopsuev/mos/moslib/store"
)

type Harness struct {
	dir        string
	socketPath string
	daemon     *store.Daemon
	doneCh     chan error
}

func Start(dir string) (*Harness, error) {
	socketPath := filepath.Join(dir, "test.sock")
	d := store.NewDaemon(socketPath, 30*time.Minute)

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- d.ListenAndServe()
	}()

	time.Sleep(50 * time.Millisecond)

	return &Harness{
		dir:        dir,
		socketPath: socketPath,
		daemon:     d,
		doneCh:     doneCh,
	}, nil
}

func (h *Harness) SocketPath() string {
	return h.socketPath
}

func (h *Harness) ConnectClient(workspaceRoots []string) (store.Store, error) {
	return store.Dial(h.socketPath, workspaceRoots)
}

func (h *Harness) Cleanup() {
	h.daemon.Shutdown()
}
