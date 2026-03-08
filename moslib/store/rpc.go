package store

type ConnectArgs struct {
	WorkspaceRoots []string
}
type ConnectReply struct {
	SessionID     string
	WorkspaceHash string
}

type GetArgs struct {
	SessionID string
	Bucket    string
	Key       string
}
type GetReply struct {
	Value []byte
}

type PutArgs struct {
	SessionID string
	Bucket    string
	Key       string
	Value     []byte
}
type PutReply struct{}

type DeleteArgs struct {
	SessionID string
	Bucket    string
	Key       string
}
type DeleteReply struct{}

type ListArgs struct {
	SessionID string
	Bucket    string
	Prefix    string
}
type ListReply struct {
	Items []KV
}

type AddEdgeArgs struct {
	SessionID string
	From      NodeID
	To        NodeID
	Rel       EdgeRel
	Meta      []byte
}
type AddEdgeReply struct{}

type RemoveEdgeArgs struct {
	SessionID string
	From      NodeID
	To        NodeID
	Rel       EdgeRel
}
type RemoveEdgeReply struct{}

type NeighborsArgs struct {
	SessionID string
	ID        NodeID
	Rel       EdgeRel
	Dir       Direction
}
type NeighborsReply struct {
	Edges []Edge
}

type WalkArgs struct {
	SessionID string
	Root      NodeID
	Rel       EdgeRel
	Dir       Direction
	MaxDepth  int
}
type WalkEntry struct {
	Depth int
	Edge  Edge
}
type WalkReply struct {
	Entries []WalkEntry
}

type DisconnectArgs struct {
	SessionID string
}
type DisconnectReply struct{}
