package history

import (
	"fmt"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type BlameLine struct {
	CommitHash vcs.Hash
	Author     string
	Time       time.Time
	LineNumber int
	Content    string
}

func Blame(repo *vcs.Repository, path string) ([]BlameLine, error) {
	head, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("no commits yet")
	}

	logEntries, err := Log(repo.Store, head, 0)
	if err != nil {
		return nil, fmt.Errorf("walk log: %w", err)
	}
	if len(logEntries) == 0 {
		return nil, fmt.Errorf("no commits in history")
	}

	type versionEntry struct {
		hash   vcs.Hash
		commit vcs.CommitData
		lines  []string
		blob   vcs.Hash
	}
	var chain []versionEntry
	var prevBlob vcs.Hash
	for _, e := range logEntries {
		flatMap, err := staging.FlattenTree(repo.Store, e.Commit.Tree, "")
		if err != nil {
			continue
		}
		blobHash, exists := flatMap[path]
		if !exists {
			break
		}
		if len(chain) > 0 && blobHash == prevBlob {
			continue
		}
		data, err := repo.Store.ReadBlob(blobHash)
		if err != nil {
			break
		}
		chain = append(chain, versionEntry{
			hash:   e.Hash,
			commit: e.Commit,
			lines:  splitLines(string(data)),
			blob:   blobHash,
		})
		prevBlob = blobHash
	}

	if len(chain) == 0 {
		return nil, fmt.Errorf("file %q not found in any commit", path)
	}

	headLines := chain[0].lines
	blame := make([]int, len(headLines))
	for i := range blame {
		blame[i] = -1
	}

	if len(chain) == 1 {
		for i := range blame {
			blame[i] = 0
		}
	} else {
		headMap := make(map[int]int, len(headLines))
		for i := range headLines {
			headMap[i] = i
		}

		for i := 0; i < len(chain)-1; i++ {
			newer := chain[i].lines
			older := chain[i+1].lines
			lcs := computeLCS(older, newer)

			lcsNewSet := make(map[int]bool, len(lcs))
			newToOld := make(map[int]int, len(lcs))
			for _, m := range lcs {
				lcsNewSet[m.newIdx] = true
				newToOld[m.newIdx] = m.oldIdx
			}

			for j := 0; j < len(newer); j++ {
				if !lcsNewSet[j] {
					if headIdx, ok := headMap[j]; ok {
						blame[headIdx] = i
					}
				}
			}

			nextMap := make(map[int]int)
			for j, headIdx := range headMap {
				if oldIdx, ok := newToOld[j]; ok {
					nextMap[oldIdx] = headIdx
				}
			}
			headMap = nextMap
		}

		for _, headIdx := range headMap {
			if blame[headIdx] == -1 {
				blame[headIdx] = len(chain) - 1
			}
		}
	}

	result := make([]BlameLine, len(headLines))
	for i, line := range headLines {
		ci := blame[i]
		if ci < 0 {
			ci = len(chain) - 1
		}
		result[i] = BlameLine{
			CommitHash: chain[ci].hash,
			Author:     chain[ci].commit.Author,
			Time:       chain[ci].commit.Time,
			LineNumber: i + 1,
			Content:    line,
		}
	}
	return result, nil
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

type lcsMatch struct {
	oldIdx int
	newIdx int
}

func computeLCS(old, new []string) []lcsMatch {
	m, n := len(old), len(new)
	if m == 0 || n == 0 {
		return nil
	}

	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if old[i] == new[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	var matches []lcsMatch
	i, j := 0, 0
	for i < m && j < n {
		if old[i] == new[j] {
			matches = append(matches, lcsMatch{oldIdx: i, newIdx: j})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			i++
		} else {
			j++
		}
	}
	return matches
}
