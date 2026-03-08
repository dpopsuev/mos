package fsstore

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const auditLogPath = ".mos/store/audit.log"

// AuditEntry represents a single line in the audit log.
type AuditEntry struct {
	Timestamp string `json:"timestamp"`
	Actor     string `json:"actor"`
	Operation string `json:"operation"`
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
}

// AuditLogger wraps an ObjectStore and appends audit entries on writes.
type AuditLogger struct {
	Inner ObjectStore
	Root  string
	actor string
}

// NewAuditLogger returns an AuditLogger wrapping inner, rooted at root.
func NewAuditLogger(inner ObjectStore, root string) *AuditLogger {
	return &AuditLogger{Inner: inner, Root: root, actor: resolveActor()}
}

func resolveActor() string {
	if v := os.Getenv("MOS_ACTOR"); v != "" {
		return v
	}
	if v := os.Getenv("USER"); v != "" {
		return v
	}
	return "unknown"
}

func (a *AuditLogger) ReadFile(path string) ([]byte, error)                       { return a.Inner.ReadFile(path) }
func (a *AuditLogger) Stat(path string) (fs.FileInfo, error)                      { return a.Inner.Stat(path) }
func (a *AuditLogger) ReadDir(path string) ([]fs.DirEntry, error)                 { return a.Inner.ReadDir(path) }
func (a *AuditLogger) MkdirAll(path string, perm fs.FileMode) error               { return a.Inner.MkdirAll(path, perm) }
func (a *AuditLogger) Pack(root string) error                                     { return a.Inner.Pack(root) }
func (a *AuditLogger) Unpack(root string) error                                   { return a.Inner.Unpack(root) }
func (a *AuditLogger) Verify(root string) ([]IntegrityError, error)               { return a.Inner.Verify(root) }

func (a *AuditLogger) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if err := a.Inner.WriteFile(path, data, perm); err != nil {
		return err
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	a.appendEntry("write", path, hash)
	return nil
}

func (a *AuditLogger) RemoveAll(path string) error {
	if err := a.Inner.RemoveAll(path); err != nil {
		return err
	}
	a.appendEntry("remove", path, "-")
	return nil
}

func (a *AuditLogger) Rename(oldPath, newPath string) error {
	if err := a.Inner.Rename(oldPath, newPath); err != nil {
		return err
	}
	a.appendEntry("rename", oldPath+" -> "+newPath, "-")
	return nil
}

func (a *AuditLogger) appendEntry(operation, path, hash string) {
	logFile := filepath.Join(a.Root, auditLogPath)
	_ = os.MkdirAll(filepath.Dir(logFile), 0o755)
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(f, "%s %s %s %s %s\n", ts, a.actor, operation, path, hash)
}

// ReadAuditLog parses the audit log from disk.
func ReadAuditLog(root string) ([]AuditEntry, error) {
	logFile := filepath.Join(root, auditLogPath)
	f, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []AuditEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 5)
		if len(parts) < 5 {
			continue
		}
		entries = append(entries, AuditEntry{
			Timestamp: parts[0],
			Actor:     parts[1],
			Operation: parts[2],
			Path:      parts[3],
			SHA256:    parts[4],
		})
	}
	return entries, scanner.Err()
}
