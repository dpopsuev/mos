package harness

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// MetricThreshold defines a named performance bound extracted from a harness block.
type MetricThreshold struct {
	Name     string   // "ns_op", "allocs_op", "bytes_op", or custom
	Ceiling  *float64 // upper bound (fail if value > ceiling)
	Floor    *float64 // lower bound (fail if value < floor)
	Unit     string   // "ns", "B", "allocs", etc.
	Pattern  string   // optional regex with named group "value" for custom extraction
	Source   string   // optional path to JSON file (relative to project root)
	JSONPath string   // optional dot-path for JSON extraction (e.g. "benchmarks[*].stats.mean")
}

// MetricResult is the outcome of extracting and checking one metric.
type MetricResult struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Unit    string   `json:"unit"`
	Ceiling *float64 `json:"ceiling,omitempty"`
	Floor   *float64 `json:"floor,omitempty"`
	Pass    bool     `json:"pass"`
	Message string   `json:"message,omitempty"`
}

// Go benchmark output regex: matches lines like
//   BenchmarkFoo-8   1000   1234.0 ns/op   56 B/op   3 allocs/op
var goBenchRe = regexp.MustCompile(
	`^Benchmark\S+\s+\d+\s+([\d.]+)\s+ns/op` +
		`(?:\s+([\d.]+)\s+B/op)?` +
		`(?:\s+(\d+)\s+allocs/op)?`,
)

// ExtractMetrics parses stdout (and optionally JSON files) for metric values and checks them against thresholds.
func ExtractMetrics(projectRoot, stdout string, thresholds []MetricThreshold) []MetricResult {
	if len(thresholds) == 0 {
		return nil
	}

	benchmarks := parseGoBenchmarks(stdout)

	var results []MetricResult
	for _, th := range thresholds {
		val, found := extractValue(th, projectRoot, stdout, benchmarks)
		if !found {
			continue
		}
		mr := MetricResult{
			Name:    th.Name,
			Value:   val,
			Unit:    th.Unit,
			Ceiling: th.Ceiling,
			Floor:   th.Floor,
			Pass:    true,
		}
		if th.Ceiling != nil && val > *th.Ceiling {
			mr.Pass = false
			mr.Message = fmt.Sprintf("%s=%.2f exceeds ceiling %.2f %s", th.Name, val, *th.Ceiling, th.Unit)
		}
		if th.Floor != nil && val < *th.Floor {
			mr.Pass = false
			mr.Message = fmt.Sprintf("%s=%.2f below floor %.2f %s", th.Name, val, *th.Floor, th.Unit)
		}
		results = append(results, mr)
	}
	return results
}

type benchResult struct {
	NsOp     float64
	BytesOp  float64
	AllocsOp float64
}

func parseGoBenchmarks(stdout string) []benchResult {
	var results []benchResult
	for _, line := range strings.Split(stdout, "\n") {
		m := goBenchRe.FindStringSubmatch(strings.TrimSpace(line))
		if m == nil {
			continue
		}
		br := benchResult{}
		br.NsOp, _ = strconv.ParseFloat(m[1], 64)
		if m[2] != "" {
			br.BytesOp, _ = strconv.ParseFloat(m[2], 64)
		}
		if m[3] != "" {
			br.AllocsOp, _ = strconv.ParseFloat(m[3], 64)
		}
		results = append(results, br)
	}
	return results
}

func extractValue(th MetricThreshold, projectRoot, stdout string, benchmarks []benchResult) (float64, bool) {
	// 1. Custom regex pattern takes highest precedence
	if th.Pattern != "" {
		return extractCustom(th.Pattern, stdout)
	}

	// 2. JSON file source (with optional json_path or registry-provided path)
	if th.Source != "" {
		jsonPath := th.JSONPath
		if jsonPath == "" {
			if ext, ok := LookupExtractor(th.Name); ok && ext.Kind == "json" && ext.JSONPath != "" {
				jsonPath = ext.JSONPath
			}
		}
		if jsonPath != "" {
			return extractFromJSON(projectRoot, th.Source, jsonPath)
		}
	}

	// 3. Registry lookup for regex extractors
	if ext, ok := LookupExtractor(th.Name); ok && ext.Kind == "regex" && ext.Pattern != "" {
		return extractCustom(ext.Pattern, stdout)
	}

	// 4. Go benchmark fallback
	if len(benchmarks) == 0 {
		return 0, false
	}
	switch th.Name {
	case "ns_op":
		return worstCase(benchmarks, func(b benchResult) float64 { return b.NsOp }), true
	case "bytes_op", "B_op":
		return worstCase(benchmarks, func(b benchResult) float64 { return b.BytesOp }), true
	case "allocs_op":
		return worstCase(benchmarks, func(b benchResult) float64 { return b.AllocsOp }), true
	default:
		return 0, false
	}
}

func extractCustom(pattern, stdout string) (float64, bool) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0, false
	}
	names := re.SubexpNames()
	valueIdx := -1
	for i, n := range names {
		if n == "value" {
			valueIdx = i
			break
		}
	}
	if valueIdx < 0 {
		return 0, false
	}
	m := re.FindStringSubmatch(stdout)
	if m == nil || valueIdx >= len(m) {
		return 0, false
	}
	v, err := strconv.ParseFloat(m[valueIdx], 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

func worstCase(benchmarks []benchResult, extract func(benchResult) float64) float64 {
	max := extract(benchmarks[0])
	for _, b := range benchmarks[1:] {
		if v := extract(b); v > max {
			max = v
		}
	}
	return max
}
