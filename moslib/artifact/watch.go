package artifact

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// TriggerResult describes a watch whose trigger condition has been met.
type TriggerResult struct {
	WatchID string `json:"watch_id"`
	Action  string `json:"action"` // "triggered" or "expired"
	Message string `json:"message"`
}

var triggerRe = regexp.MustCompile(`^when\s+(\S+)\s+(\S+)$`)

// EvaluateWatchTriggers scans active watches and checks trigger conditions.
func EvaluateWatchTriggers(root string, now time.Time) ([]TriggerResult, error) {
	reg, err := LoadRegistry(root)
	if err != nil {
		return nil, nil
	}
	td, ok := reg.Types["watch"]
	if !ok {
		return nil, nil
	}

	watches, err := GenericList(root, td, "active")
	if err != nil {
		return nil, err
	}

	var results []TriggerResult
	for _, w := range watches {
		ab, err := dsl.ReadArtifact(w.Path)
		if err != nil {
			continue
		}

		if expires, ok := dsl.FieldString(ab.Items, "expires"); ok && expires != "" {
			t, err := time.Parse(time.RFC3339, expires)
			if err == nil && now.After(t) {
				results = append(results, TriggerResult{
					WatchID: w.ID,
					Action:  "expired",
					Message: fmt.Sprintf("watch %s expired at %s", w.ID, expires),
				})
				continue
			}
		}

		if trigger, ok := dsl.FieldString(ab.Items, "trigger"); ok && trigger != "" {
			if tr := evaluateTrigger(root, w.ID, trigger); tr != nil {
				results = append(results, *tr)
			}
		}
	}
	return results, nil
}

func evaluateTrigger(root, watchID, trigger string) *TriggerResult {
	m := triggerRe.FindStringSubmatch(strings.TrimSpace(trigger))
	if m == nil {
		return nil
	}
	targetID := m[1]
	expectedStatus := m[2]

	actualStatus := resolveArtifactStatus(root, targetID)
	if actualStatus == "" {
		return nil
	}

	if strings.EqualFold(actualStatus, expectedStatus) {
		return &TriggerResult{
			WatchID: watchID,
			Action:  "triggered",
			Message: fmt.Sprintf("watch %s triggered: %s reached status %q", watchID, targetID, expectedStatus),
		}
	}
	return nil
}

// resolveArtifactStatus finds an artifact by ID across all known types and returns its status.
func resolveArtifactStatus(root, id string) string {
	mosDir := filepath.Join(root, MosDir)

	coreKinds := map[string]string{
		"contract": "contracts",
	}

	for kind, dir := range coreKinds {
		for _, sub := range []string{ActiveDir, ArchiveDir} {
			p := filepath.Join(mosDir, dir, sub, id, kind+".mos")
			if _, err := os.Stat(p); err != nil {
				continue
			}
			ab, err := dsl.ReadArtifact(p)
			if err != nil {
				continue
			}
			if status, ok := dsl.FieldString(ab.Items, "status"); ok {
				return status
			}
		}
	}

	reg, err := LoadRegistry(root)
	if err != nil {
		return ""
	}
	for _, td := range reg.Types {
		for _, sub := range []string{ActiveDir, ArchiveDir} {
			p := filepath.Join(mosDir, td.Directory, sub, id, td.Kind+".mos")
			if _, err := os.Stat(p); err != nil {
				continue
			}
			ab, err := dsl.ReadArtifact(p)
			if err != nil {
				continue
			}
			if status, ok := dsl.FieldString(ab.Items, "status"); ok {
				return status
			}
		}
	}
	return ""
}
