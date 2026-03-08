package store

import (
	"context"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"
)

type RPCHandler struct {
	mgr *SessionManager
}

func (h *RPCHandler) Connect(args *ConnectArgs, reply *ConnectReply) error {
	sess, err := h.mgr.Connect(args.WorkspaceRoots)
	if err != nil {
		return err
	}
	reply.SessionID = sess.ID
	reply.WorkspaceHash = sess.WorkspaceHash
	return nil
}

func (h *RPCHandler) Disconnect(args *DisconnectArgs, reply *DisconnectReply) error {
	h.mgr.Disconnect(args.SessionID)
	return nil
}

func (h *RPCHandler) Get(args *GetArgs, reply *GetReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	val, err := s.Get(context.Background(), args.Bucket, args.Key)
	if err != nil {
		return err
	}
	reply.Value = val
	return nil
}

func (h *RPCHandler) Put(args *PutArgs, reply *PutReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	return s.Put(context.Background(), args.Bucket, args.Key, args.Value)
}

func (h *RPCHandler) Delete(args *DeleteArgs, reply *DeleteReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	return s.Delete(context.Background(), args.Bucket, args.Key)
}

func (h *RPCHandler) List(args *ListArgs, reply *ListReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	items, err := s.List(context.Background(), args.Bucket, args.Prefix)
	if err != nil {
		return err
	}
	reply.Items = items
	return nil
}

func (h *RPCHandler) AddEdge(args *AddEdgeArgs, reply *AddEdgeReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	return s.AddEdge(context.Background(), args.From, args.To, args.Rel, args.Meta)
}

func (h *RPCHandler) RemoveEdge(args *RemoveEdgeArgs, reply *RemoveEdgeReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	return s.RemoveEdge(context.Background(), args.From, args.To, args.Rel)
}

func (h *RPCHandler) Neighbors(args *NeighborsArgs, reply *NeighborsReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	edges, err := s.Neighbors(context.Background(), args.ID, args.Rel, args.Dir)
	if err != nil {
		return err
	}
	reply.Edges = edges
	return nil
}

func (h *RPCHandler) Walk(args *WalkArgs, reply *WalkReply) error {
	s, err := h.mgr.Route(args.SessionID)
	if err != nil {
		return err
	}
	var entries []WalkEntry
	err = s.Walk(context.Background(), args.Root, args.Rel, args.Dir, args.MaxDepth, func(depth int, edge Edge) bool {
		entries = append(entries, WalkEntry{Depth: depth, Edge: edge})
		return true
	})
	if err != nil {
		return err
	}
	reply.Entries = entries
	return nil
}

type Daemon struct {
	SocketPath  string
	IdleTimeout time.Duration

	mgr      *SessionManager
	listener net.Listener
	srv      *rpc.Server
	wg       sync.WaitGroup
}

func NewDaemon(socketPath string, idleTimeout time.Duration) *Daemon {
	return &Daemon{
		SocketPath:  socketPath,
		IdleTimeout: idleTimeout,
	}
}

func (d *Daemon) ListenAndServe() error {
	if err := os.MkdirAll(filepath.Dir(d.SocketPath), 0o755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	os.Remove(d.SocketPath)

	ln, err := net.Listen("unix", d.SocketPath)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	d.listener = ln

	d.mgr = NewSessionManager(d.IdleTimeout)
	d.srv = rpc.NewServer()

	handler := &RPCHandler{mgr: d.mgr}
	if err := d.srv.RegisterName("Store", handler); err != nil {
		ln.Close()
		return fmt.Errorf("register rpc: %w", err)
	}

	if err := d.writePID(); err != nil {
		ln.Close()
		return fmt.Errorf("write pid: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		d.Shutdown()
	}()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.srv.Accept(ln)
	}()

	d.wg.Wait()
	return nil
}

func (d *Daemon) Shutdown() {
	if d.listener != nil {
		d.listener.Close()
	}
	if d.mgr != nil {
		d.mgr.Close()
	}
	os.Remove(d.SocketPath)
	os.Remove(DefaultPIDPath())
}

func (d *Daemon) writePID() error {
	pidPath := DefaultPIDPath()
	if err := os.MkdirAll(filepath.Dir(pidPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(pidPath, []byte(strconv.Itoa(os.Getpid())), 0o644)
}
