package store

import (
	"context"
	"fmt"
	"net/rpc"
)

type Client struct {
	rpc       *rpc.Client
	sessionID string
}

func Dial(socketPath string, workspaceRoots []string) (*Client, error) {
	rc, err := rpc.Dial("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("dial daemon: %w", err)
	}

	var reply ConnectReply
	if err := rc.Call("Store.Connect", &ConnectArgs{WorkspaceRoots: workspaceRoots}, &reply); err != nil {
		rc.Close()
		return nil, fmt.Errorf("handshake: %w", err)
	}

	return &Client{rpc: rc, sessionID: reply.SessionID}, nil
}

func (c *Client) SessionID() string { return c.sessionID }

func (c *Client) Get(_ context.Context, bucket, key string) ([]byte, error) {
	var reply GetReply
	err := c.rpc.Call("Store.Get", &GetArgs{
		SessionID: c.sessionID,
		Bucket:    bucket,
		Key:       key,
	}, &reply)
	return reply.Value, err
}

func (c *Client) Put(_ context.Context, bucket, key string, value []byte) error {
	var reply PutReply
	return c.rpc.Call("Store.Put", &PutArgs{
		SessionID: c.sessionID,
		Bucket:    bucket,
		Key:       key,
		Value:     value,
	}, &reply)
}

func (c *Client) Delete(_ context.Context, bucket, key string) error {
	var reply DeleteReply
	return c.rpc.Call("Store.Delete", &DeleteArgs{
		SessionID: c.sessionID,
		Bucket:    bucket,
		Key:       key,
	}, &reply)
}

func (c *Client) List(_ context.Context, bucket, prefix string) ([]KV, error) {
	var reply ListReply
	err := c.rpc.Call("Store.List", &ListArgs{
		SessionID: c.sessionID,
		Bucket:    bucket,
		Prefix:    prefix,
	}, &reply)
	return reply.Items, err
}

func (c *Client) AddEdge(_ context.Context, from, to NodeID, rel EdgeRel, meta []byte) error {
	var reply AddEdgeReply
	return c.rpc.Call("Store.AddEdge", &AddEdgeArgs{
		SessionID: c.sessionID,
		From:      from,
		To:        to,
		Rel:       rel,
		Meta:      meta,
	}, &reply)
}

func (c *Client) RemoveEdge(_ context.Context, from, to NodeID, rel EdgeRel) error {
	var reply RemoveEdgeReply
	return c.rpc.Call("Store.RemoveEdge", &RemoveEdgeArgs{
		SessionID: c.sessionID,
		From:      from,
		To:        to,
		Rel:       rel,
	}, &reply)
}

func (c *Client) Neighbors(_ context.Context, id NodeID, rel EdgeRel, dir Direction) ([]Edge, error) {
	var reply NeighborsReply
	err := c.rpc.Call("Store.Neighbors", &NeighborsArgs{
		SessionID: c.sessionID,
		ID:        id,
		Rel:       rel,
		Dir:       dir,
	}, &reply)
	return reply.Edges, err
}

func (c *Client) Walk(_ context.Context, root NodeID, rel EdgeRel, dir Direction, maxDepth int, fn WalkFn) error {
	var reply WalkReply
	err := c.rpc.Call("Store.Walk", &WalkArgs{
		SessionID: c.sessionID,
		Root:      root,
		Rel:       rel,
		Dir:       dir,
		MaxDepth:  maxDepth,
	}, &reply)
	if err != nil {
		return err
	}
	for _, entry := range reply.Entries {
		if !fn(entry.Depth, entry.Edge) {
			break
		}
	}
	return nil
}

func (c *Client) Close() error {
	var reply DisconnectReply
	_ = c.rpc.Call("Store.Disconnect", &DisconnectArgs{SessionID: c.sessionID}, &reply)
	return c.rpc.Close()
}
