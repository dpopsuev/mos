package harness

// Extractor defines a named metric extraction strategy.
type Extractor struct {
	Name        string // unique name, e.g. "ns_op", "jmh_ns_op", "criterion_ns"
	Kind        string // "regex" or "json"
	Pattern     string // regex with named group "value" (for Kind="regex")
	JSONPath    string // dot-path with [*] array support (for Kind="json")
	Unit        string // default unit label
	Aggregation string // "max" (default), "min", "avg", "last"
}

var builtinExtractors = map[string]Extractor{}

func init() {
	for _, e := range []Extractor{
		// Go benchmark (fallback path, no Pattern — handled by parseGoBenchmarks)
		{Name: "ns_op", Kind: "regex", Unit: "ns"},
		{Name: "bytes_op", Kind: "regex", Unit: "B"},
		{Name: "B_op", Kind: "regex", Unit: "B"},
		{Name: "allocs_op", Kind: "regex", Unit: "allocs"},

		// JMH (Java Microbenchmark Harness)
		{Name: "jmh_ns_op", Kind: "regex", Unit: "ns/op",
			Pattern: `(?m)^\S+\s+(?:thrpt|avgt|sample|ss)\s+\d+\s+(?P<value>[\d.]+)\s+.*ns/op`},
		{Name: "jmh_us_op", Kind: "regex", Unit: "us/op",
			Pattern: `(?m)^\S+\s+(?:thrpt|avgt|sample|ss)\s+\d+\s+(?P<value>[\d.]+)\s+.*us/op`},
		{Name: "jmh_ms_op", Kind: "regex", Unit: "ms/op",
			Pattern: `(?m)^\S+\s+(?:thrpt|avgt|sample|ss)\s+\d+\s+(?P<value>[\d.]+)\s+.*ms/op`},
		{Name: "jmh_ops_s", Kind: "regex", Unit: "ops/s",
			Pattern: `(?m)^\S+\s+thrpt\s+\d+\s+(?P<value>[\d.]+)\s+.*ops/s`},

		// hyperfine
		{Name: "hyperfine_s", Kind: "regex", Unit: "s",
			Pattern: `Time \(mean[^)]*\):\s+(?P<value>[\d.]+)\s+s`},
		{Name: "hyperfine_ms", Kind: "regex", Unit: "ms",
			Pattern: `Time \(mean[^)]*\):\s+(?P<value>[\d.]+)\s+ms`},

		// Rust Criterion (JSON file)
		{Name: "criterion_ns", Kind: "json", Unit: "ns",
			JSONPath: "estimates.median.point_estimate"},

		// Python pytest-benchmark (JSON file)
		{Name: "pytest_mean_ns", Kind: "json", Unit: "ns",
			JSONPath: "benchmarks[*].stats.mean"},
		{Name: "pytest_median_ns", Kind: "json", Unit: "ns",
			JSONPath: "benchmarks[*].stats.median"},

		// C++ Google Benchmark (JSON file)
		{Name: "gbench_ns", Kind: "json", Unit: "ns",
			JSONPath: "benchmarks[*].real_time"},

		// .NET BenchmarkDotNet (JSON file)
		{Name: "dotnet_ns", Kind: "json", Unit: "ns",
			JSONPath: "Benchmarks[*].Statistics.Median"},

		// Go test coverage
		{Name: "go_coverage", Kind: "regex", Unit: "%",
			Pattern: `coverage:\s+(?P<value>[\d.]+)%`},
	} {
		builtinExtractors[e.Name] = e
	}
}

// RegisterExtractor adds or replaces an extractor in the built-in registry.
func RegisterExtractor(e Extractor) {
	builtinExtractors[e.Name] = e
}

// LookupExtractor returns the named extractor if registered.
func LookupExtractor(name string) (Extractor, bool) {
	e, ok := builtinExtractors[name]
	return e, ok
}

// ListExtractors returns all registered extractor names.
func ListExtractors() []string {
	names := make([]string, 0, len(builtinExtractors))
	for n := range builtinExtractors {
		names = append(names, n)
	}
	return names
}
