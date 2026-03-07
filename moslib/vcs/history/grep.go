package history

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
)

type GrepOpts struct {
	Pattern      string
	ContextLines int
	Committed    bool
	IgnoreCase   bool
}

type GrepMatch struct {
	Path       string
	LineNumber int
	Line       string
	IsContext  bool
}

func Grep(repo *vcs.Repository, opts GrepOpts) ([]GrepMatch, error) {
	flags := ""
	if opts.IgnoreCase {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + opts.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	if opts.Committed {
		return grepCommitted(repo, re, opts.ContextLines)
	}
	return grepWorkingTree(repo.Root, re, opts.ContextLines)
}

func grepWorkingTree(root string, re *regexp.Regexp, contextLines int) ([]GrepMatch, error) {
	mosDir := filepath.Join(root, ".mos")
	var allMatches []GrepMatch

	err := filepath.Walk(mosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vcs" || base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if staging.ShouldSkipFile(rel) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		matches := grepLines(rel, string(data), re, contextLines)
		allMatches = append(allMatches, matches...)
		return nil
	})
	return allMatches, err
}

func grepCommitted(repo *vcs.Repository, re *regexp.Regexp, contextLines int) ([]GrepMatch, error) {
	head, err := vcs.ResolveHead(repo)
	if err != nil {
		return nil, fmt.Errorf("no commits yet")
	}
	cd, err := repo.Store.ReadCommit(head)
	if err != nil {
		return nil, err
	}
	flatMap, err := staging.FlattenTree(repo.Store, cd.Tree, "")
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(flatMap))
	for p := range flatMap {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var allMatches []GrepMatch
	for _, path := range paths {
		data, err := repo.Store.ReadBlob(flatMap[path])
		if err != nil {
			continue
		}
		matches := grepLines(path, string(data), re, contextLines)
		allMatches = append(allMatches, matches...)
	}
	return allMatches, nil
}

func grepLines(path, content string, re *regexp.Regexp, contextLines int) []GrepMatch {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	matchSet := map[int]bool{}
	for i, line := range lines {
		if re.MatchString(line) {
			matchSet[i] = true
		}
	}
	if len(matchSet) == 0 {
		return nil
	}

	includeSet := map[int]bool{}
	for i := range matchSet {
		for c := i - contextLines; c <= i+contextLines; c++ {
			if c >= 0 && c < len(lines) {
				includeSet[c] = true
			}
		}
	}

	indices := make([]int, 0, len(includeSet))
	for i := range includeSet {
		indices = append(indices, i)
	}
	sort.Ints(indices)

	var matches []GrepMatch
	for _, i := range indices {
		matches = append(matches, GrepMatch{
			Path:       path,
			LineNumber: i + 1,
			Line:       lines[i],
			IsContext:  !matchSet[i],
		})
	}
	return matches
}
