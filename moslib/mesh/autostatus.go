package mesh

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
)

// StatusChange describes a proposed contract status transition inferred from git history.
type StatusChange struct {
	ContractID string `json:"contract_id"`
	From       string `json:"from"`
	To         string `json:"to"`
	CommitHash string `json:"commit_hash"`
	Reason     string `json:"reason"`
}

// InferStatusChanges scans git commits and proposes status transitions
// for referenced contracts based on heuristics:
//   - A draft contract mentioned in any commit -> active
//   - An active contract mentioned in a merge commit on the main branch -> complete
func InferStatusChanges(root string) ([]StatusChange, error) {
	traces, err := TraceCommitContracts(root)
	if err != nil {
		return nil, err
	}

	mainBranch := detectMainBranch(root)
	mainCommits := collectMainCommits(root, mainBranch)

	seen := make(map[string]bool)
	var changes []StatusChange

	for _, trace := range traces {
		for _, cid := range trace.ContractIDs {
			if !strings.HasPrefix(cid, "CON-") && !strings.HasPrefix(cid, "BUG-") {
				continue
			}
			if seen[cid] {
				continue
			}

			status, err := artifact.GetContractStatus(root, cid)
			if err != nil {
				continue
			}

			switch status {
			case "draft":
				changes = append(changes, StatusChange{
					ContractID: cid,
					From:       "draft",
					To:         "active",
					CommitHash: trace.Hash,
					Reason:     "first commit reference",
				})
				seen[cid] = true
			case "active":
				if mainCommits[trace.Hash] {
					changes = append(changes, StatusChange{
						ContractID: cid,
						From:       "active",
						To:         artifact.StatusComplete,
						CommitHash: trace.Hash,
						Reason:     fmt.Sprintf("commit on %s branch", mainBranch),
					})
					seen[cid] = true
				}
			}
		}
	}

	return changes, nil
}

// ApplyStatusChanges applies the proposed transitions.
func ApplyStatusChanges(root string, changes []StatusChange) error {
	for _, c := range changes {
		if err := artifact.UpdateContractStatus(root, c.ContractID, c.To); err != nil {
			return fmt.Errorf("applying %s -> %s for %s: %w", c.From, c.To, c.ContractID, err)
		}
	}
	return nil
}

func detectMainBranch(root string) string {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "--short")
	cmd.Dir = root
	out, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(out))
		branch = strings.TrimPrefix(branch, "origin/")
		return branch
	}
	for _, candidate := range []string{"main", "master"} {
		cmd := exec.Command("git", "rev-parse", "--verify", candidate)
		cmd.Dir = root
		if err := cmd.Run(); err == nil {
			return candidate
		}
	}
	return "main"
}

func collectMainCommits(root, branch string) map[string]bool {
	cmd := exec.Command("git", "log", branch, "--format=%H", "--max-count=500")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	commits := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			commits[line] = true
		}
	}
	return commits
}
