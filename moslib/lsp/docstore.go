package lsp

import (
	"net/url"
	"strings"
	"sync"
)

// DocumentStore tracks open documents by URI.
type DocumentStore struct {
	mu   sync.RWMutex
	docs map[string]string // URI -> content
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{docs: make(map[string]string)}
}

func (ds *DocumentStore) Open(uri, content string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.docs[uri] = content
}

func (ds *DocumentStore) Update(uri, content string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.docs[uri] = content
}

func (ds *DocumentStore) Close(uri string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	delete(ds.docs, uri)
}

func (ds *DocumentStore) Get(uri string) (string, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	c, ok := ds.docs[uri]
	return c, ok
}

// URIToPath converts a file:// URI to a filesystem path.
func URIToPath(uri string) string {
	if strings.HasPrefix(uri, "file://") {
		u, err := url.Parse(uri)
		if err == nil {
			return u.Path
		}
	}
	return uri
}

// PathToURI converts a filesystem path to a file:// URI.
func PathToURI(path string) string {
	if strings.HasPrefix(path, "/") {
		return "file://" + path
	}
	return "file:///" + path
}
