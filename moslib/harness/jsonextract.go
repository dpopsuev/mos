package harness

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// extractFromJSON reads a JSON file and walks a dot-path to extract numeric values.
// Supports [*] for array iteration. Aggregates via max (default).
func extractFromJSON(projectRoot, source, jsonPath string) (float64, bool) {
	path := source
	if !filepath.IsAbs(path) {
		path = filepath.Join(projectRoot, source)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	var root interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return 0, false
	}

	values := walkJSONPath(root, strings.Split(jsonPath, "."))
	if len(values) == 0 {
		return 0, false
	}
	return aggregateMax(values), true
}

func walkJSONPath(node interface{}, segments []string) []float64 {
	if len(segments) == 0 {
		return toFloatSlice(node)
	}

	seg := segments[0]
	rest := segments[1:]

	if seg == "[*]" || strings.HasSuffix(seg, "[*]") {
		key := strings.TrimSuffix(seg, "[*]")
		target := node
		if key != "" {
			m, ok := node.(map[string]interface{})
			if !ok {
				return nil
			}
			target = m[key]
		}
		arr, ok := target.([]interface{})
		if !ok {
			return nil
		}
		var results []float64
		for _, item := range arr {
			results = append(results, walkJSONPath(item, rest)...)
		}
		return results
	}

	m, ok := node.(map[string]interface{})
	if !ok {
		return nil
	}
	child, exists := m[seg]
	if !exists {
		return nil
	}
	return walkJSONPath(child, rest)
}

func toFloatSlice(v interface{}) []float64 {
	switch val := v.(type) {
	case float64:
		return []float64{val}
	case json.Number:
		f, err := val.Float64()
		if err != nil {
			return nil
		}
		return []float64{f}
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return nil
		}
		return []float64{f}
	default:
		return nil
	}
}

func aggregateMax(values []float64) float64 {
	max := math.Inf(-1)
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}
