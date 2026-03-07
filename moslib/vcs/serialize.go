package vcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"time"
)

// serializeTree produces a canonical byte representation of tree entries.
// Format: sorted by name, each entry is: mode(4 bytes) + nameLen(2 bytes) + name + hash(32 bytes)
func serializeTree(entries []TreeEntry) []byte {
	sorted := make([]TreeEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	var buf bytes.Buffer
	buf.WriteString("tree\n")
	for _, e := range sorted {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, e.Mode)
		buf.Write(b)
		nameBytes := []byte(e.Name)
		nb := make([]byte, 2)
		binary.BigEndian.PutUint16(nb, uint16(len(nameBytes)))
		buf.Write(nb)
		buf.Write(nameBytes)
		buf.Write(e.Hash[:])
	}
	return buf.Bytes()
}

// deserializeTree reverses serializeTree.
func deserializeTree(data []byte) ([]TreeEntry, error) {
	if !bytes.HasPrefix(data, []byte("tree\n")) {
		return nil, fmt.Errorf("not a tree object")
	}
	data = data[5:] // skip "tree\n"

	var entries []TreeEntry
	for len(data) > 0 {
		if len(data) < 6 {
			return nil, fmt.Errorf("truncated tree entry header")
		}
		mode := binary.BigEndian.Uint32(data[:4])
		nameLen := int(binary.BigEndian.Uint16(data[4:6]))
		data = data[6:]

		if len(data) < nameLen+32 {
			return nil, fmt.Errorf("truncated tree entry body")
		}
		name := string(data[:nameLen])
		data = data[nameLen:]
		var h Hash
		copy(h[:], data[:32])
		data = data[32:]

		entries = append(entries, TreeEntry{Name: name, Hash: h, Mode: mode})
	}
	return entries, nil
}

// serializeCommit produces a canonical byte representation of a commit.
func serializeCommit(c CommitData) []byte {
	var buf bytes.Buffer
	buf.WriteString("commit\n")
	fmt.Fprintf(&buf, "tree %s\n", c.Tree)
	for _, p := range c.Parents {
		fmt.Fprintf(&buf, "parent %s\n", p)
	}
	fmt.Fprintf(&buf, "author %s <%s> %d\n", c.Author, c.Email, c.Time.Unix())
	fmt.Fprintf(&buf, "\n%s\n", c.Message)
	return buf.Bytes()
}

// deserializeCommit reverses serializeCommit.
func deserializeCommit(data []byte) (*CommitData, error) {
	lines := bytes.Split(data, []byte("\n"))
	if len(lines) < 3 || string(lines[0]) != "commit" {
		return nil, fmt.Errorf("not a commit object")
	}

	c := &CommitData{}
	i := 1
	for ; i < len(lines); i++ {
		line := string(lines[i])
		if line == "" {
			i++
			break
		}

		switch {
		case len(line) > 5 && line[:5] == "tree ":
			h, err := ParseHash(line[5:])
			if err != nil {
				return nil, fmt.Errorf("invalid tree hash: %w", err)
			}
			c.Tree = h

		case len(line) > 7 && line[:7] == "parent ":
			h, err := ParseHash(line[7:])
			if err != nil {
				return nil, fmt.Errorf("invalid parent hash: %w", err)
			}
			c.Parents = append(c.Parents, h)

		case len(line) > 7 && line[:7] == "author ":
			author, email, ts, err := parseAuthorLine(line[7:])
			if err != nil {
				return nil, err
			}
			c.Author = author
			c.Email = email
			c.Time = ts
		}
	}

	var msgLines []string
	for ; i < len(lines); i++ {
		msgLines = append(msgLines, string(lines[i]))
	}
	msg := ""
	if len(msgLines) > 0 {
		msg = joinTrimRight(msgLines)
	}
	c.Message = msg
	return c, nil
}

func parseAuthorLine(s string) (name, email string, t time.Time, err error) {
	ltIdx := bytes.IndexByte([]byte(s), '<')
	gtIdx := bytes.IndexByte([]byte(s), '>')
	if ltIdx < 0 || gtIdx < 0 || gtIdx < ltIdx {
		return "", "", time.Time{}, fmt.Errorf("malformed author line: %s", s)
	}
	name = s[:ltIdx]
	if len(name) > 0 && name[len(name)-1] == ' ' {
		name = name[:len(name)-1]
	}
	email = s[ltIdx+1 : gtIdx]

	rest := s[gtIdx+1:]
	if len(rest) > 0 && rest[0] == ' ' {
		rest = rest[1:]
	}
	var unix int64
	_, err = fmt.Sscanf(rest, "%d", &unix)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("malformed timestamp in author line: %w", err)
	}
	t = time.Unix(unix, 0).UTC()
	return name, email, t, nil
}

func joinTrimRight(ss []string) string {
	for len(ss) > 0 && ss[len(ss)-1] == "" {
		ss = ss[:len(ss)-1]
	}
	var buf bytes.Buffer
	for i, s := range ss {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(s)
	}
	return buf.String()
}
