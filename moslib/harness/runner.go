package harness

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

const defaultTimeout = 5 * time.Minute

// Discover walks .mos/rules/ and returns a HarnessSpec for each rule
// that contains a harness block.
func Discover(mosDir string) ([]HarnessSpec, error) {
	var specs []HarnessSpec

	for _, sub := range []string{"mechanical", "interpretive"} {
		dir := filepath.Join(mosDir, "rules", sub)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, fmt.Errorf("reading %s: %w", dir, err)
		}

		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".mos") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			spec, err := extractHarness(path)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			if spec != nil {
				specs = append(specs, *spec)
			}
		}
	}

	return specs, nil
}

func extractHarness(path string) (*HarnessSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := dsl.Parse(string(data), nil)
	if err != nil {
		return nil, err
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil, nil
	}

	harness := dsl.FindBlock(ab.Items, "harness")
	if harness == nil {
		return nil, nil
	}

	command, _ := dsl.FieldString(harness.Items, "command")
	if command == "" {
		return nil, nil
	}

	timeout := defaultTimeout
	if ts, ok := dsl.FieldString(harness.Items, "timeout"); ok {
		parsed, err := time.ParseDuration(ts)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout %q in %s: %w", ts, path, err)
		}
		timeout = parsed
	}

	enforcement, _ := dsl.FieldString(ab.Items, "enforcement")
	if enforcement == "" {
		enforcement = "error"
	}

	trigger, _ := dsl.FieldString(harness.Items, "trigger")
	if trigger == "" {
		trigger = "manual"
	}

	vector, _ := dsl.FieldString(ab.Items, "vector")

	spec := &HarnessSpec{
		RuleID:      ab.Name,
		Command:     command,
		Timeout:     timeout,
		Enforcement: enforcement,
		Trigger:     trigger,
		Vector:      vector,
	}

	for _, item := range harness.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "metric" || blk.Title == "" {
			continue
		}
		th := MetricThreshold{Name: blk.Title}
		if c, ok := dsl.FieldFloat(blk.Items, "ceiling"); ok {
			v := c
			th.Ceiling = &v
		}
		if f, ok := dsl.FieldFloat(blk.Items, "floor"); ok {
			v := f
			th.Floor = &v
		}
		th.Unit, _ = dsl.FieldString(blk.Items, "unit")
		th.Pattern, _ = dsl.FieldString(blk.Items, "pattern")
		th.Source, _ = dsl.FieldString(blk.Items, "source")
		th.JSONPath, _ = dsl.FieldString(blk.Items, "json_path")
		spec.Thresholds = append(spec.Thresholds, th)
	}

	return spec, nil
}

// Run executes each HarnessSpec sequentially from the given root directory
// and returns the collected evidence.
func Run(root string, specs []HarnessSpec) []Evidence {
	results := make([]Evidence, 0, len(specs))
	for _, spec := range specs {
		ev := execute(root, spec)
		if len(spec.Thresholds) > 0 {
			ev.Metrics = ExtractMetrics(root, ev.Stdout, spec.Thresholds)
			for _, mr := range ev.Metrics {
				if !mr.Pass {
					if spec.Enforcement == "error" {
						ev.Pass = false
					}
					break
				}
			}
		}
		results = append(results, ev)
	}
	return results
}

func execute(root string, spec HarnessSpec) Evidence {
	ctx, cancel := context.WithTimeout(context.Background(), spec.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", spec.Command)
	cmd.Dir = root
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)

	ev := Evidence{
		RuleID:   spec.RuleID,
		Command:  spec.Command,
		Duration: dur,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}

	switch {
	case err == nil:
		ev.ExitCode = 0
		ev.Pass = true
	case ctx.Err() == context.DeadlineExceeded:
		ev.TimedOut = true
		ev.ExitCode = -1
		ev.Pass = false
	default:
		if exitErr, ok := err.(*exec.ExitError); ok {
			ev.ExitCode = exitErr.ExitCode()
		} else {
			ev.ExitCode = -1
			ev.Stderr = ev.Stderr + "\n" + err.Error()
		}
		ev.Pass = false
	}

	return ev
}
