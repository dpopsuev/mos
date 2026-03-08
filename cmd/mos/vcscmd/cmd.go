package vcscmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dpopsuev/mos/moslib/guard"
	"github.com/dpopsuev/mos/moslib/vcs"
	"github.com/dpopsuev/mos/moslib/vcs/history"
	"github.com/dpopsuev/mos/moslib/vcs/merge"
	"github.com/dpopsuev/mos/moslib/vcs/staging"
	"github.com/dpopsuev/mos/moslib/vcs/transport"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "vcs",
	Short: "Governance version control",
}

func init() {
	initCmd.Flags().String("store", "fs", "Backend store: git or fs")
	commitCmd.Flags().StringP("message", "m", "", "Commit message (required)")
	commitCmd.Flags().String("author", "mos-agent", "Author name")
	commitCmd.Flags().String("email", "agent@mos", "Author email")
	commitCmd.Flags().Bool("no-verify", false, "Skip pre-commit gates")
	_ = commitCmd.MarkFlagRequired("message")
	logCmd.Flags().IntP("max", "n", 0, "Maximum number of commits to show")
	logCmd.Flags().Bool("oneline", false, "Show each commit as a single line: <short-hash> <subject>")
	logCmd.Flags().String("format", "", "Output format: json")
	logCmd.Flags().Bool("stat", false, "Show changed file paths per commit")
	diffCmd.Flags().Bool("staged", false, "Show staged changes")
	resetCmd.Flags().Bool("soft", false, "Soft reset")
	resetCmd.Flags().Bool("hard", false, "Hard reset")
	resetCmd.Flags().Bool("mixed", false, "Mixed reset (default)")
	migrateCmd.Flags().String("to", "", "Target backend")
	_ = migrateCmd.MarkFlagRequired("to")
	branchCmd.Flags().StringP("delete", "d", "", "Delete branch by name")
	checkoutCmd.Flags().BoolP("branch", "b", false, "Create new branch and switch to it")
	tagCmd.Flags().StringP("delete", "d", "", "Delete tag by name")
	lsTreeCmd.Flags().BoolP("recursive", "r", false, "Recurse into subtrees")
	pushCmd.Flags().Bool("force", false, "Force push")
	pushCmd.Flags().Bool("all", false, "Push all governance branches")
	cloneCmd.Flags().String("dest", "", "Destination directory (default: inferred from URL)")
	stashCmd.Flags().StringP("message", "m", "", "Stash message")
	cleanCmd.Flags().Bool("dry-run", false, "Show what would be removed without deleting")
	cleanCmd.Flags().BoolP("force", "f", false, "Actually remove untracked files")
	cleanCmd.Flags().BoolP("dirs", "d", false, "Also remove empty untracked directories")
	grepCmd.Flags().IntP("context", "C", 0, "Lines of context around matches")
	grepCmd.Flags().Bool("committed", false, "Search HEAD tree instead of working tree")
	grepCmd.Flags().BoolP("ignore-case", "i", false, "Case-insensitive matching")
	Cmd.AddCommand(
		initCmd, addCmd, statusCmd, commitCmd, logCmd, diffCmd, showCmd, resetCmd,
		configCmd, migrateCmd, branchCmd, checkoutCmd, mergeCmd, rebaseCmd,
		tagCmd, hashObjectCmd, catFileCmd, lsTreeCmd, revParseCmd,
		remoteCmd, pushCmd, fetchCmd, pullCmd, cloneCmd,
		stashCmd, cleanCmd, grepCmd, blameCmd,
	)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize governance VCS",
	Long:  "usage: mos vcs init [--store git|fs]",
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	store, _ := cmd.Flags().GetString("store")
	repo, err := vcs.InitRepo(".", store)
	if err != nil {
		return fmt.Errorf("mos vcs init: %w", err)
	}
	fmt.Printf("Initialized governance VCS (backend: %s)\n", repo.Config.Backend)
	return nil
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Stage governance artifacts",
	Long:  "Stages governance artifacts. No args = stage all under .mos/.",
	RunE:  runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs add: %w", err)
	}
	paths := args
	if len(paths) == 0 {
		paths = []string{"."}
	}
	if err := staging.Add(repo, paths); err != nil {
		return fmt.Errorf("mos vcs add: %w", err)
	}
	idx, _ := staging.LoadIndex(".")
	fmt.Printf("Staged %d artifacts\n", len(idx.Entries))
	return nil
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show working tree status",
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs status: %w", err)
	}
	idx, err := staging.LoadIndex(".")
	if err != nil {
		return fmt.Errorf("mos vcs status: %w", err)
	}
	indexMap := staging.IndexToMap(idx)
	workEntries, err := staging.SnapshotWorkingTree(".", repo.Store)
	if err != nil {
		return fmt.Errorf("mos vcs status: %w", err)
	}
	workMap := map[string]vcs.Hash{}
	for _, e := range workEntries {
		workMap[e.Path] = e.Hash
	}
	workDiffs := staging.DiffTrees(indexMap, workMap)
	headHash, err := vcs.ResolveHead(repo)
	var headDiffs []staging.DiffEntry
	if err == nil && !headHash.IsZero() {
		cd, err := repo.Store.ReadCommit(headHash)
		if err == nil {
			headMap, err := staging.FlattenTree(repo.Store, cd.Tree, "")
			if err == nil {
				headDiffs = staging.DiffTrees(headMap, indexMap)
			}
		}
	}
	branch := vcs.CurrentBranch(".")
	if branch != "" {
		fmt.Printf("On branch %s\n", branch)
	} else {
		fmt.Printf("HEAD detached at %s\n", headHash.Short())
	}
	if len(headDiffs) == 0 && len(workDiffs) == 0 {
		fmt.Println("nothing to commit, working tree clean")
		return nil
	}
	if len(headDiffs) > 0 {
		fmt.Println("Changes staged for commit:")
		for _, d := range headDiffs {
			fmt.Printf("  %s: %s\n", d.Kind, d.Path)
		}
	}
	if len(workDiffs) > 0 {
		fmt.Println("Changes not staged for commit:")
		for _, d := range workDiffs {
			fmt.Printf("  %s: %s\n", d.Kind, d.Path)
		}
	}
	return nil
}

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commit staged changes",
	Long:  "usage: mos vcs commit -m <message> [--author <name>] [--email <email>] [--no-verify]",
	RunE:  runCommit,
}

func runCommit(cmd *cobra.Command, args []string) error {
	message, _ := cmd.Flags().GetString("message")
	author, _ := cmd.Flags().GetString("author")
	email, _ := cmd.Flags().GetString("email")
	noVerify, _ := cmd.Flags().GetBool("no-verify")
	if message == "" {
		return fmt.Errorf("mos vcs commit: -m <message> required")
	}
	if !noVerify {
		if failed := RunPreCommitGates("."); failed {
			return fmt.Errorf("mos vcs commit: pre-commit gates failed (use --no-verify to skip)")
		}
	}
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs commit: %w", err)
	}
	h, err := staging.Commit(repo, author, email, message)
	if err != nil {
		return fmt.Errorf("mos vcs commit: %w", err)
	}
	fmt.Printf("[%s] %s\n", h.Short(), message)

	return nil
}

// RunPreCommitGates delegates to guard.PreCommit for unified enforcement.
func RunPreCommitGates(root string) bool {
	result := guard.PreCommit(root)
	if !result.Pass {
		for _, d := range result.Diagnostics {
			if d.Severity == "error" {
				fmt.Fprintf(os.Stderr, "  [%s] %s: %s\n", d.Severity, d.File, d.Message)
			}
		}
	}
	return !result.Pass
}

var logCmd = &cobra.Command{
	Use:   "log",
	Short: "Show commit history",
	Long:  "usage: mos vcs log [-n <count>] [--oneline] [--format json] [--stat]",
	RunE:  runLog,
}

func runLog(cmd *cobra.Command, args []string) error {
	maxCount, _ := cmd.Flags().GetInt("max")
	oneline, _ := cmd.Flags().GetBool("oneline")
	format, _ := cmd.Flags().GetString("format")
	stat, _ := cmd.Flags().GetBool("stat")

	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs log: %w", err)
	}
	head, err := vcs.ResolveHead(repo)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mos vcs log: no commits yet")
		return nil
	}
	entries, err := history.Log(repo.Store, head, maxCount)
	if err != nil {
		return fmt.Errorf("mos vcs log: %w", err)
	}

	if format == "json" {
		return logFormatJSON(entries, repo, stat)
	}
	if oneline {
		return logFormatOneline(entries)
	}
	return logFormatDefault(entries, repo, stat)
}

func logFormatOneline(entries []history.LogEntry) error {
	for _, e := range entries {
		subject := firstLine(e.Commit.Message)
		fmt.Printf("%s %s\n", e.Hash.Short(), subject)
	}
	return nil
}

func logFormatDefault(entries []history.LogEntry, repo *vcs.Repository, stat bool) error {
	for _, e := range entries {
		fmt.Printf("commit %s\n", e.Hash)
		fmt.Printf("Author: %s <%s>\n", e.Commit.Author, e.Commit.Email)
		fmt.Printf("Date:   %s\n", e.Commit.Time.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Println()
		for _, line := range splitLines(e.Commit.Message) {
			fmt.Printf("    %s\n", line)
		}
		fmt.Println()
		if stat {
			diffs, err := history.CommitStat(repo.Store, e)
			if err == nil && len(diffs) > 0 {
				for _, d := range diffs {
					fmt.Printf(" %s | %s\n", d.Path, d.Kind)
				}
				fmt.Printf(" %d file(s) changed\n", len(diffs))
				fmt.Println()
			}
		}
	}
	return nil
}

type logJSONEntry struct {
	Hash    string   `json:"hash"`
	Author  string   `json:"author"`
	Email   string   `json:"email"`
	Date    string   `json:"date"`
	Message string   `json:"message"`
	Parents []string `json:"parents"`
	Files   []string `json:"files,omitempty"`
}

func logFormatJSON(entries []history.LogEntry, repo *vcs.Repository, stat bool) error {
	out := make([]logJSONEntry, 0, len(entries))
	for _, e := range entries {
		parents := make([]string, 0, len(e.Commit.Parents))
		for _, p := range e.Commit.Parents {
			parents = append(parents, p.String())
		}
		je := logJSONEntry{
			Hash:    e.Hash.String(),
			Author:  e.Commit.Author,
			Email:   e.Commit.Email,
			Date:    e.Commit.Time.Format("2006-01-02T15:04:05Z"),
			Message: e.Commit.Message,
			Parents: parents,
		}
		if stat {
			diffs, err := history.CommitStat(repo.Store, e)
			if err == nil {
				for _, d := range diffs {
					je.Files = append(je.Files, d.Path)
				}
			}
		}
		out = append(out, je)
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("mos vcs log: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}

func splitLines(s string) []string {
	if s == "" {
		return []string{""}
	}
	lines := []string{}
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show changes (working vs index vs HEAD)",
	Long:  "usage: mos vcs diff [--staged]",
	RunE:  runDiff,
}

func runDiff(cmd *cobra.Command, args []string) error {
	staged, _ := cmd.Flags().GetBool("staged")
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs diff: %w", err)
	}
	idx, _ := staging.LoadIndex(".")
	indexMap := staging.IndexToMap(idx)
	if staged {
		headHash, err := vcs.ResolveHead(repo)
		if err != nil {
			fmt.Println("(no commits yet — all staged files are new)")
			for path := range indexMap {
				fmt.Printf("  added: %s\n", path)
			}
			return nil
		}
		cd, _ := repo.Store.ReadCommit(headHash)
		headMap, _ := staging.FlattenTree(repo.Store, cd.Tree, "")
		diffs := staging.DiffTrees(headMap, indexMap)
		for _, d := range diffs {
			fmt.Printf("  %s: %s\n", d.Kind, d.Path)
		}
		return nil
	}
	workEntries, _ := staging.SnapshotWorkingTree(".", repo.Store)
	workMap := map[string]vcs.Hash{}
	for _, e := range workEntries {
		workMap[e.Path] = e.Hash
	}
	diffs := staging.DiffTrees(indexMap, workMap)
	if len(diffs) == 0 {
		fmt.Println("no changes")
		return nil
	}
	for _, d := range diffs {
		fmt.Printf("  %s: %s\n", d.Kind, d.Path)
	}
	return nil
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display a commit or object",
	Long:  "usage: mos vcs show <hash|ref>",
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs show: %w", err)
	}
	target := args[0]
	h, err := history.RevParse(repo.Store, ".", target)
	if err != nil {
		return fmt.Errorf("mos vcs show: cannot resolve %q", target)
	}
	typ, err := repo.Store.TypeOf(h)
	if err != nil {
		return fmt.Errorf("mos vcs show: %w", err)
	}
	switch typ {
	case vcs.ObjectBlob:
		data, err := repo.Store.ReadBlob(h)
		if err != nil {
			return fmt.Errorf("mos vcs show: %w", err)
		}
		fmt.Printf("blob %s (%d bytes)\n\n", h, len(data))
		os.Stdout.Write(data)
	case vcs.ObjectTree:
		entries, err := repo.Store.ReadTree(h)
		if err != nil {
			return fmt.Errorf("mos vcs show: %w", err)
		}
		fmt.Printf("tree %s\n\n", h)
		for _, e := range entries {
			kind := "blob"
			if e.Mode == vcs.ModeDir {
				kind = "tree"
			}
			fmt.Printf("  %06o %s %s\t%s\n", e.Mode, kind, e.Hash.Short(), e.Name)
		}
	case vcs.ObjectCommit:
		cd, err := repo.Store.ReadCommit(h)
		if err != nil {
			return fmt.Errorf("mos vcs show: %w", err)
		}
		fmt.Printf("commit %s\n", h)
		fmt.Printf("Tree:    %s\n", cd.Tree)
		for _, p := range cd.Parents {
			fmt.Printf("Parent:  %s\n", p)
		}
		fmt.Printf("Author:  %s <%s>\n", cd.Author, cd.Email)
		fmt.Printf("Date:    %s\n", cd.Time.Format("Mon Jan 2 15:04:05 2006 -0700"))
		fmt.Printf("\n    %s\n", cd.Message)
	}
	return nil
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset HEAD to a previous commit",
	Long:  "usage: mos vcs reset [--soft|--hard] [<ref>]",
	RunE:  runReset,
}

func runReset(cmd *cobra.Command, args []string) error {
	mode := "mixed"
	if cmd.Flags().Changed("soft") {
		mode = "soft"
	}
	if cmd.Flags().Changed("hard") {
		mode = "hard"
	}
	var target string
	if len(args) > 0 {
		target = args[0]
	}
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs reset: %w", err)
	}
	if target == "" {
		head, err := vcs.ResolveHead(repo)
		if err != nil {
			return fmt.Errorf("mos vcs reset: no commits")
		}
		cd, err := repo.Store.ReadCommit(head)
		if err != nil || len(cd.Parents) == 0 {
			return fmt.Errorf("mos vcs reset: already at root commit")
		}
		target = cd.Parents[0].String()
	}
	h, err := history.RevParse(repo.Store, ".", target)
	if err != nil {
		return fmt.Errorf("mos vcs reset: cannot resolve %q", target)
	}
	branch, detached := vcs.ReadSymbolicHead(".")
	if detached {
		if err := vcs.WriteDetachedHead(".", h); err != nil {
			return fmt.Errorf("mos vcs reset: %w", err)
		}
	} else {
		if err := repo.Store.UpdateRef("heads/"+branch, h); err != nil {
			return fmt.Errorf("mos vcs reset: %w", err)
		}
	}
	if mode == "mixed" || mode == "hard" {
		cd, err := repo.Store.ReadCommit(h)
		if err != nil {
			return fmt.Errorf("mos vcs reset: %w", err)
		}
		flatMap, err := staging.FlattenTree(repo.Store, cd.Tree, "")
		if err != nil {
			return fmt.Errorf("mos vcs reset: %w", err)
		}
		idx, _ := staging.LoadIndex(".")
		idx.Entries = nil
		for path, hash := range flatMap {
			idx.Set(path, hash, vcs.ModeRegular)
		}
		idx.Save()
	}
	fmt.Printf("HEAD is now at %s (%s reset)\n", h.Short(), mode)
	return nil
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Read/write VCS configuration",
	Long:  "usage: mos vcs config get|set <key> [<value>]",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runConfig,
}

func runConfig(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs config: %w", err)
	}
	switch args[0] {
	case "get":
		key := args[1]
		switch key {
		case "backend", "object_store.backend":
			fmt.Println(repo.Config.Backend)
		default:
			return fmt.Errorf("mos vcs config get: unknown key %q", key)
		}
	case "set":
		if len(args) < 3 {
			return fmt.Errorf("mos vcs config set: requires <key> <value>")
		}
		key, val := args[1], args[2]
		switch key {
		case "backend", "object_store.backend":
			repo.Config.Backend = val
		default:
			return fmt.Errorf("mos vcs config set: unknown key %q", key)
		}
		if err := writeVCSConfigCLI(repo.Config); err != nil {
			return fmt.Errorf("mos vcs config set: %w", err)
		}
		fmt.Printf("Updated %s = %s\n", key, val)
	default:
		return fmt.Errorf("mos vcs config: unknown operation %q", args[0])
	}
	return nil
}

func writeVCSConfigCLI(cfg vcs.VCSConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(".mos/vcs.json", data, 0644)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate between backends",
	Long:  "usage: mos vcs migrate --to <backend>",
	RunE:  runMigrate,
}

func runMigrate(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("to")
	if target == "" {
		return fmt.Errorf("mos vcs migrate: --to <backend> required")
	}
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs migrate: %w", err)
	}
	fmt.Printf("Migrating from %s to %s...\n", repo.Config.Backend, target)
	if err := vcs.Migrate(repo, target); err != nil {
		return fmt.Errorf("mos vcs migrate: %w", err)
	}
	fmt.Printf("Migration complete. Backend is now: %s\n", repo.Config.Backend)
	return nil
}

var branchCmd = &cobra.Command{
	Use:   "branch",
	Short: "Create, list, or delete branches",
	Long:  "No args: list branches. With name: create branch at HEAD (or at hash). -d name: delete branch.",
	RunE:  runBranch,
}

func runBranch(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs branch: %w", err)
	}
	deleteName, _ := cmd.Flags().GetString("delete")
	var createName, createAt string
	for _, a := range args {
		if createName == "" {
			createName = a
		} else {
			createAt = a
			break
		}
	}
	if deleteName != "" {
		if err := history.DeleteBranch(repo.Store, ".", deleteName); err != nil {
			return fmt.Errorf("mos vcs branch: %w", err)
		}
		fmt.Printf("Deleted branch %s\n", deleteName)
		return nil
	}
	if createName != "" {
		var target vcs.Hash
		if createAt != "" {
			target, err = history.RevParse(repo.Store, ".", createAt)
			if err != nil {
				return fmt.Errorf("mos vcs branch: %w", err)
			}
		} else {
			target, err = vcs.ResolveHead(repo)
			if err != nil {
				return fmt.Errorf("mos vcs branch: %w", err)
			}
		}
		if err := history.CreateBranch(repo.Store, createName, target); err != nil {
			return fmt.Errorf("mos vcs branch: %w", err)
		}
		fmt.Printf("Created branch %s at %s\n", createName, target.Short())
		return nil
	}
	current := vcs.CurrentBranch(".")
	branches, err := history.ListBranches(repo.Store)
	if err != nil {
		return fmt.Errorf("mos vcs branch: %w", err)
	}
	for _, b := range branches {
		marker := "  "
		if b.Name == current {
			marker = "* "
		}
		fmt.Printf("%s%s\t%s\n", marker, b.Name, b.Hash.Short())
	}
	return nil
}

var checkoutCmd = &cobra.Command{
	Use:   "checkout",
	Short: "Switch branches or detach HEAD",
	Long:  "Switch to a branch, tag, or detach at a hash. -b: create a new branch and switch to it.",
	RunE:  runCheckout,
}

func runCheckout(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs checkout: %w", err)
	}
	newBranch, _ := cmd.Flags().GetBool("branch")
	if len(args) == 0 {
		return fmt.Errorf("mos vcs checkout: requires a target")
	}
	target := args[0]
	if newBranch {
		result, err := history.CheckoutNewBranch(repo, target)
		if err != nil {
			return fmt.Errorf("mos vcs checkout: %w", err)
		}
		fmt.Printf("Switched to new branch '%s'\n", result.Branch)
		return nil
	}
	result, err := history.Checkout(repo, target)
	if err != nil {
		return fmt.Errorf("mos vcs checkout: %w", err)
	}
	if result.Detached {
		fmt.Printf("HEAD detached at %s\n", result.Hash.Short())
	} else {
		fmt.Printf("Switched to branch '%s'\n", result.Branch)
	}
	return nil
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge a branch into current",
	Args:  cobra.ExactArgs(1),
	RunE:  runMerge,
}

func runMerge(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs merge: %w", err)
	}
	result, err := merge.Merge(repo, args[0], "mos-agent", "agent@mos")
	if err != nil {
		return fmt.Errorf("mos vcs merge: %w", err)
	}
	if len(result.Conflicts) > 0 {
		fmt.Fprintf(os.Stderr, "CONFLICT: %d file(s) have merge conflicts:\n", len(result.Conflicts))
		for _, c := range result.Conflicts {
			fmt.Fprintf(os.Stderr, "  %s\n", c.Path)
		}
		return fmt.Errorf("merge conflicts")
	}
	if result.FastForward {
		fmt.Printf("Fast-forward to %s\n", result.CommitHash.Short())
	} else {
		fmt.Printf("Merge commit: %s\n", result.CommitHash.Short())
	}
	return nil
}

var rebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "Rebase current branch onto target",
	Args:  cobra.ExactArgs(1),
	RunE:  runRebase,
}

func runRebase(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs rebase: %w", err)
	}
	result, err := merge.Rebase(repo, args[0], "mos-agent", "agent@mos")
	if err != nil {
		return fmt.Errorf("mos vcs rebase: %w", err)
	}
	fmt.Printf("Rebased %d commit(s) onto %s\n", result.ReplayedCount, args[0])
	fmt.Printf("HEAD is now at %s\n", result.NewTip.Short())
	return nil
}

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Create, list, or delete tags",
	Long:  "No args: list tags. With name: create tag at HEAD (or at hash). -d name: delete tag.",
	RunE:  runTag,
}

func runTag(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs tag: %w", err)
	}
	deleteName, _ := cmd.Flags().GetString("delete")
	var createName, createAt string
	for _, a := range args {
		if createName == "" {
			createName = a
		} else {
			createAt = a
			break
		}
	}
	if deleteName != "" {
		if err := history.DeleteTag(repo.Store, deleteName); err != nil {
			return fmt.Errorf("mos vcs tag: %w", err)
		}
		fmt.Printf("Deleted tag %s\n", deleteName)
		return nil
	}
	if createName != "" {
		var target vcs.Hash
		if createAt != "" {
			target, err = history.RevParse(repo.Store, ".", createAt)
			if err != nil {
				return fmt.Errorf("mos vcs tag: %w", err)
			}
		} else {
			target, err = vcs.ResolveHead(repo)
			if err != nil {
				return fmt.Errorf("mos vcs tag: %w", err)
			}
		}
		if err := history.CreateTag(repo.Store, createName, target); err != nil {
			return fmt.Errorf("mos vcs tag: %w", err)
		}
		fmt.Printf("Created tag %s at %s\n", createName, target.Short())
		return nil
	}
	tags, err := history.ListTags(repo.Store)
	if err != nil {
		return fmt.Errorf("mos vcs tag: %w", err)
	}
	for _, t := range tags {
		fmt.Printf("%s\t%s\n", t.Name, t.Hash.Short())
	}
	return nil
}

var hashObjectCmd = &cobra.Command{
	Use:   "hash-object",
	Short: "Hash a file and store as blob",
	Args:  cobra.ExactArgs(1),
	RunE:  runHashObject,
}

func runHashObject(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs hash-object: %w", err)
	}
	h, err := history.HashObject(repo.Store, args[0])
	if err != nil {
		return fmt.Errorf("mos vcs hash-object: %w", err)
	}
	fmt.Println(h)
	return nil
}

var catFileCmd = &cobra.Command{
	Use:   "cat-file",
	Short: "Display object content by hash",
	Args:  cobra.ExactArgs(1),
	RunE:  runCatFile,
}

func runCatFile(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs cat-file: %w", err)
	}
	h, err := history.RevParse(repo.Store, ".", args[0])
	if err != nil {
		return fmt.Errorf("mos vcs cat-file: %w", err)
	}
	typ, data, err := history.CatFile(repo.Store, h)
	if err != nil {
		return fmt.Errorf("mos vcs cat-file: %w", err)
	}
	fmt.Printf("%s %s\n\n", typ, h)
	os.Stdout.Write(data)
	return nil
}

var lsTreeCmd = &cobra.Command{
	Use:   "ls-tree",
	Short: "List tree entries",
	Long:  "usage: mos vcs ls-tree <hash|ref> [-r]",
	Args:  cobra.ExactArgs(1),
	RunE:  runLsTree,
}

func runLsTree(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs ls-tree: %w", err)
	}
	recursive, _ := cmd.Flags().GetBool("recursive")
	h, err := history.RevParse(repo.Store, ".", args[0])
	if err != nil {
		return fmt.Errorf("mos vcs ls-tree: %w", err)
	}
	entries, err := history.LsTree(repo.Store, h, recursive)
	if err != nil {
		return fmt.Errorf("mos vcs ls-tree: %w", err)
	}
	for _, e := range entries {
		fmt.Printf("%06o %s %s\t%s\n", e.Mode, e.Type, e.Hash.Short(), e.Path)
	}
	return nil
}

var revParseCmd = &cobra.Command{
	Use:   "rev-parse",
	Short: "Resolve ref to hash",
	Args:  cobra.ExactArgs(1),
	RunE:  runRevParse,
}

func runRevParse(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs rev-parse: %w", err)
	}
	h, err := history.RevParse(repo.Store, ".", args[0])
	if err != nil {
		return fmt.Errorf("mos vcs rev-parse: %w", err)
	}
	fmt.Println(h)
	return nil
}

// --- Remote management ---

var remoteCmd = &cobra.Command{
	Use:   "remote [list|add|remove]",
	Short: "Manage remote repositories",
	Long:  "No args or 'list': show remotes. 'add <name> <url>': add remote. 'remove <name>': remove remote.",
	RunE:  runRemote,
}

func runRemote(cmd *cobra.Command, args []string) error {
	if len(args) == 0 || args[0] == "list" {
		return runRemoteList()
	}
	switch args[0] {
	case "add":
		if len(args) < 3 {
			return fmt.Errorf("usage: mos vcs remote add <name> <url>")
		}
		return runRemoteAdd(args[1], args[2])
	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: mos vcs remote remove <name>")
		}
		return runRemoteRemove(args[1])
	default:
		return fmt.Errorf("unknown remote subcommand %q (use list, add, or remove)", args[0])
	}
}

func runRemoteList() error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs remote: %w", err)
	}
	remotes, err := transport.ListRemotes(repo)
	if err != nil {
		return fmt.Errorf("mos vcs remote: %w", err)
	}
	for _, r := range remotes {
		fmt.Printf("%s\t%s\n", r.Name, r.URL)
	}
	return nil
}

func runRemoteAdd(name, url string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs remote add: %w", err)
	}
	if err := transport.AddRemote(repo, name, url); err != nil {
		return fmt.Errorf("mos vcs remote add: %w", err)
	}
	fmt.Printf("Added remote %s -> %s\n", name, url)
	return nil
}

func runRemoteRemove(name string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs remote remove: %w", err)
	}
	if err := transport.RemoveRemote(repo, name); err != nil {
		return fmt.Errorf("mos vcs remote remove: %w", err)
	}
	fmt.Printf("Removed remote %s\n", name)
	return nil
}

// --- Push ---

var pushCmd = &cobra.Command{
	Use:   "push [remote] [branch]",
	Short: "Push governance refs to a remote",
	Long:  "Defaults: remote=origin, branch=current. Use --force for force-push, --all for all branches.",
	RunE:  runPush,
}

func runPush(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs push: %w", err)
	}
	force, _ := cmd.Flags().GetBool("force")
	all, _ := cmd.Flags().GetBool("all")
	remote := "origin"
	var branch string
	if len(args) > 0 {
		remote = args[0]
	}
	if len(args) > 1 {
		branch = args[1]
	}
	if err := transport.Push(repo, remote, transport.PushOpts{Force: force, All: all, Branch: branch}); err != nil {
		return fmt.Errorf("mos vcs push: %w", err)
	}
	fmt.Printf("Pushed to %s\n", remote)
	return nil
}

// --- Fetch ---

var fetchCmd = &cobra.Command{
	Use:   "fetch [remote]",
	Short: "Fetch governance refs from a remote",
	Long:  "Downloads refs/mos/* from remote. Default remote: origin.",
	RunE:  runFetch,
}

func runFetch(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs fetch: %w", err)
	}
	remote := "origin"
	if len(args) > 0 {
		remote = args[0]
	}
	if err := transport.Fetch(repo, remote); err != nil {
		return fmt.Errorf("mos vcs fetch: %w", err)
	}
	fmt.Printf("Fetched from %s\n", remote)
	return nil
}

// --- Pull ---

var pullCmd = &cobra.Command{
	Use:   "pull [remote] [branch]",
	Short: "Fetch and merge governance refs",
	Long:  "Fetches from remote then merges. Defaults: remote=origin, branch=current.",
	RunE:  runPull,
}

func runPull(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs pull: %w", err)
	}
	remote := "origin"
	var branch string
	if len(args) > 0 {
		remote = args[0]
	}
	if len(args) > 1 {
		branch = args[1]
	}
	result, err := transport.Pull(repo, remote, branch, "mos-agent", "agent@mos")
	if err != nil {
		return fmt.Errorf("mos vcs pull: %w", err)
	}
	if len(result.Conflicts) > 0 {
		fmt.Fprintf(os.Stderr, "CONFLICT: %d file(s) have merge conflicts:\n", len(result.Conflicts))
		for _, c := range result.Conflicts {
			fmt.Fprintf(os.Stderr, "  %s\n", c.Path)
		}
		return fmt.Errorf("merge conflicts")
	}
	if result.FastForward {
		fmt.Printf("Fast-forward to %s\n", result.CommitHash.Short())
	} else {
		fmt.Printf("Merged from %s: %s\n", remote, result.CommitHash.Short())
	}
	return nil
}

// --- Clone ---

var cloneCmd = &cobra.Command{
	Use:   "clone <url> [dest]",
	Short: "Clone a remote governance repository",
	Long:  "Clones a Git repository and initializes governance VCS with git backend.",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runClone,
}

func runClone(cmd *cobra.Command, args []string) error {
	url := args[0]
	dest, _ := cmd.Flags().GetString("dest")
	if dest == "" && len(args) > 1 {
		dest = args[1]
	}
	result, err := transport.Clone(url, dest)
	if err != nil {
		return fmt.Errorf("mos vcs clone: %w", err)
	}
	fmt.Printf("Cloned into %s (branch: %s)\n", result.Root, result.Branch)
	return nil
}

// --- Stash ---

var stashCmd = &cobra.Command{
	Use:   "stash [save|pop|apply|list|drop]",
	Short: "Save and restore in-progress governance work",
	Long:  "No args or 'save': stash current state. 'pop': apply & remove latest. 'apply [n]': apply without removing. 'list': show stack. 'drop [n]': remove entry.",
	RunE:  runStash,
}

func runStash(cmd *cobra.Command, args []string) error {
	sub := "save"
	if len(args) > 0 {
		sub = args[0]
	}
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs stash: %w", err)
	}
	switch sub {
	case "save":
		msg, _ := cmd.Flags().GetString("message")
		if err := staging.Stash(repo, msg); err != nil {
			return fmt.Errorf("mos vcs stash save: %w", err)
		}
		fmt.Println("Saved working directory and index state")
	case "pop":
		entry, err := staging.StashPop(repo)
		if err != nil {
			return fmt.Errorf("mos vcs stash pop: %w", err)
		}
		fmt.Printf("Applied stash: %s\n", entry.Message)
	case "apply":
		idx := 0
		if len(args) > 1 {
			if _, err := fmt.Sscanf(args[1], "%d", &idx); err != nil {
				return fmt.Errorf("mos vcs stash apply: invalid index %q", args[1])
			}
		}
		entry, err := staging.StashApply(repo, idx)
		if err != nil {
			return fmt.Errorf("mos vcs stash apply: %w", err)
		}
		fmt.Printf("Applied stash@{%d}: %s\n", idx, entry.Message)
	case "list":
		entries, err := staging.StashList(repo)
		if err != nil {
			return fmt.Errorf("mos vcs stash list: %w", err)
		}
		for i, e := range entries {
			fmt.Printf("stash@{%d}: %s (%s)\n", i, e.Message, e.Time.Format("2006-01-02 15:04"))
		}
	case "drop":
		idx := 0
		if len(args) > 1 {
			if _, err := fmt.Sscanf(args[1], "%d", &idx); err != nil {
				return fmt.Errorf("mos vcs stash drop: invalid index %q", args[1])
			}
		}
		if err := staging.StashDrop(repo, idx); err != nil {
			return fmt.Errorf("mos vcs stash drop: %w", err)
		}
		fmt.Printf("Dropped stash@{%d}\n", idx)
	default:
		return fmt.Errorf("mos vcs stash: unknown subcommand %q (use save, pop, apply, list, drop)", sub)
	}
	return nil
}

// --- Clean ---

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove untracked governance artifacts",
	Long:  "Remove files from .mos/ that are not in the index. Requires --force or --dry-run.",
	RunE:  runClean,
}

func runClean(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs clean: %w", err)
	}
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	dirs, _ := cmd.Flags().GetBool("dirs")
	paths, err := staging.Clean(repo, staging.CleanOpts{
		DryRun: dryRun,
		Force:  force,
		Dirs:   dirs,
	})
	if err != nil {
		return fmt.Errorf("mos vcs clean: %w", err)
	}
	if len(paths) == 0 {
		fmt.Println("Nothing to clean")
		return nil
	}
	verb := "Removing"
	if dryRun {
		verb = "Would remove"
	}
	for _, p := range paths {
		fmt.Printf("%s %s\n", verb, p)
	}
	return nil
}

// --- Grep ---

var grepCmd = &cobra.Command{
	Use:   "grep <pattern>",
	Short: "Search governance artifacts by pattern",
	Long:  "Regex search across .mos/ artifact contents. Default: working tree. --committed: search HEAD tree.",
	Args:  cobra.ExactArgs(1),
	RunE:  runGrep,
}

func runGrep(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs grep: %w", err)
	}
	contextLines, _ := cmd.Flags().GetInt("context")
	committed, _ := cmd.Flags().GetBool("committed")
	ignoreCase, _ := cmd.Flags().GetBool("ignore-case")
	matches, err := history.Grep(repo, history.GrepOpts{
		Pattern:      args[0],
		ContextLines: contextLines,
		Committed:    committed,
		IgnoreCase:   ignoreCase,
	})
	if err != nil {
		return fmt.Errorf("mos vcs grep: %w", err)
	}
	if len(matches) == 0 {
		return nil
	}
	prevPath := ""
	for _, m := range matches {
		if m.Path != prevPath {
			if prevPath != "" {
				fmt.Println()
			}
			prevPath = m.Path
		}
		sep := ":"
		if m.IsContext {
			sep = "-"
		}
		fmt.Printf("%s%s%d%s%s\n", m.Path, sep, m.LineNumber, sep, m.Line)
	}
	return nil
}

// --- Blame ---

var blameCmd = &cobra.Command{
	Use:   "blame <path>",
	Short: "Show per-line commit attribution",
	Long:  "For each line in a governance artifact, show which commit last changed it.",
	Args:  cobra.ExactArgs(1),
	RunE:  runBlame,
}

func runBlame(cmd *cobra.Command, args []string) error {
	repo, err := vcs.OpenRepo(".")
	if err != nil {
		return fmt.Errorf("mos vcs blame: %w", err)
	}
	lines, err := history.Blame(repo, args[0])
	if err != nil {
		return fmt.Errorf("mos vcs blame: %w", err)
	}
	for _, bl := range lines {
		fmt.Printf("%s (%s %s) %s\n",
			bl.CommitHash.Short(),
			bl.Author,
			bl.Time.Format("2006-01-02"),
			bl.Content,
		)
	}
	return nil
}
