package linter

import (
	"bufio"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/harness"
)

// NestingViolation records a function that exceeds the allowed nesting depth.
type NestingViolation struct {
	File    string
	Func    string
	Line    int
	Depth   int
	Ceiling int
}

// AnalyzeNestingDepth parses Go files under root, computes max nesting depth per
// function (counting if/for/range/switch/select/case), and returns violations
// where depth exceeds ceiling. The glob parameter filters files (e.g. "**/*.go");
// empty or "**/*.go" means all .go files recursively.
func AnalyzeNestingDepth(root string, ceiling int, glob string) ([]NestingViolation, error) {
	files, err := collectGoFiles(root, glob)
	if err != nil {
		return nil, err
	}

	var violations []NestingViolation
	fset := token.NewFileSet()

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			relPath = path
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			maxDepth := maxNestingInBlock(fn.Body)
			if maxDepth > ceiling {
				line := fset.Position(fn.Pos()).Line
				name := funcName(fn)
				violations = append(violations, NestingViolation{
					File:    relPath,
					Func:    name,
					Line:    line,
					Depth:   maxDepth,
					Ceiling: ceiling,
				})
			}
		}
	}

	return violations, nil
}

func collectGoFiles(root string, glob string) ([]string, error) {
	var out []string
	root = filepath.Clean(root)
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		if strings.HasSuffix(root, ".go") {
			return []string{root}, nil
		}
		return nil, nil
	}

	// Treat empty or "**/*.go" as "all .go files recursively"
	useRecursive := glob == "" || glob == "**/*.go"

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == ".git" || strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		if useRecursive {
			out = append(out, path)
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		matched, err := filepath.Match(glob, rel)
		if err != nil {
			return err
		}
		if matched {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func funcName(fn *ast.FuncDecl) string {
	if fn.Recv != nil && len(fn.Recv.List) > 0 {
		recv := fn.Recv.List[0]
		if len(recv.Names) > 0 {
			return recv.Names[0].Name + "." + fn.Name.Name
		}
	}
	return fn.Name.Name
}

// maxNestingInBlock returns the maximum nesting depth inside a block statement.
// Depth is counted for: if, for, range, switch, type switch, select, and case clauses.
func maxNestingInBlock(block *ast.BlockStmt) int {
	return walkStmtList(block.List, 0)
}

func walkStmtList(list []ast.Stmt, depth int) int {
	max := depth
	for _, stmt := range list {
		if m := walkStmt(stmt, depth); m > max {
			max = m
		}
	}
	return max
}

func walkStmt(stmt ast.Stmt, depth int) int {
	if stmt == nil {
		return depth
	}
	nextDepth := depth + 1
	var bodyMax int

	switch s := stmt.(type) {
	case *ast.IfStmt:
		bodyMax = walkStmtList([]ast.Stmt{s.Body}, nextDepth)
		if s.Else != nil {
			if elseBlock, ok := s.Else.(*ast.BlockStmt); ok {
				if m := walkStmtList(elseBlock.List, nextDepth); m > bodyMax {
					bodyMax = m
				}
			} else {
				// else if
				if m := walkStmt(s.Else, nextDepth); m > bodyMax {
					bodyMax = m
				}
			}
		}
		return bodyMax

	case *ast.ForStmt:
		return walkStmtList([]ast.Stmt{s.Body}, nextDepth)

	case *ast.RangeStmt:
		return walkStmtList([]ast.Stmt{s.Body}, nextDepth)

	case *ast.SwitchStmt:
		return walkSwitchBody(s.Body, nextDepth)

	case *ast.TypeSwitchStmt:
		return walkSwitchBody(s.Body, nextDepth)

	case *ast.SelectStmt:
		return walkSwitchBody(s.Body, nextDepth)

	case *ast.BlockStmt:
		return walkStmtList(s.List, depth)

	default:
		return depth
	}
}

func walkSwitchBody(body *ast.BlockStmt, depth int) int {
	if body == nil {
		return depth
	}
	max := depth
	for _, stmt := range body.List {
		clause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		// Case clause adds one level; walk its body at depth+1
		if m := walkStmtList(clause.Body, depth+1); m > max {
			max = m
		}
	}
	return max
}

// LengthViolation records a function that violates length boundaries.
type LengthViolation struct {
	File      string
	Func      string
	Line      int
	Length    int
	Threshold int
	Kind      string // "ceiling" or "floor"
}

// AnalyzeFunctionLength parses Go files under root, computes function body
// line counts, and returns violations where length exceeds ceiling or falls
// below floor.
func AnalyzeFunctionLength(root string, ceiling, floor int, glob string) ([]LengthViolation, error) {
	files, err := collectGoFiles(root, glob)
	if err != nil {
		return nil, err
	}

	var violations []LengthViolation
	fset := token.NewFileSet()

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			relPath = path
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			startLine := fset.Position(fn.Body.Lbrace).Line
			endLine := fset.Position(fn.Body.Rbrace).Line
			length := endLine - startLine - 1
			if length < 0 {
				length = 0
			}

			name := funcName(fn)
			line := fset.Position(fn.Pos()).Line

			if ceiling > 0 && length > ceiling {
				violations = append(violations, LengthViolation{
					File: relPath, Func: name, Line: line,
					Length: length, Threshold: ceiling, Kind: "ceiling",
				})
			}
			if floor > 0 && length < floor {
				violations = append(violations, LengthViolation{
					File: relPath, Func: name, Line: line,
					Length: length, Threshold: floor, Kind: "floor",
				})
			}
		}
	}
	return violations, nil
}

// RunStructuralChecks discovers quality configs in mosDir, runs metric
// analyzers for each, and returns diagnostics. Supports bidirectional
// thresholds (floor + ceiling) for nesting_depth and function_length.
func RunStructuralChecks(root string, mosDir string) []Diagnostic {
	configs, err := harness.DiscoverQualityConfigs(mosDir)
	if err != nil {
		return nil
	}

	var diags []Diagnostic
	for _, qc := range configs {
		glob := qc.Glob
		if glob == "" {
			glob = "**/*.go"
		}
		sev := SeverityWarning
		if qc.Enforcement == "error" {
			sev = SeverityError
		}

		switch qc.Metric {
		case "nesting_depth":
			diags = append(diags, runNestingDepth(root, qc, glob, sev)...)
		case "function_length":
			diags = append(diags, runFunctionLength(root, qc, glob, sev)...)
		case "params_per_function":
			diags = append(diags, runParamsPerFunction(root, qc, glob, sev)...)
		case "fan_out":
			diags = append(diags, runFanOut(root, qc, glob, sev)...)
		case "loc_per_file":
			diags = append(diags, runLOCPerFile(root, qc, glob, sev)...)
		}
	}
	return diags
}

func runNestingDepth(root string, qc harness.QualityConfig, glob string, sev Severity) []Diagnostic {
	ceiling, ok := qc.EffectiveCeiling("go")
	if !ok {
		return nil
	}
	violations, err := AnalyzeNestingDepth(root, ceiling, glob)
	if err != nil {
		return []Diagnostic{{
			Severity: SeverityWarning, Rule: qc.RuleID,
			Message: fmt.Sprintf("structural check failed: %v", err),
		}}
	}
	var diags []Diagnostic
	for _, v := range violations {
		diags = append(diags, Diagnostic{
			File: v.File, Line: v.Line, Severity: sev, Rule: qc.RuleID,
			Message: fmt.Sprintf("%s: nesting depth %d exceeds ceiling %d", v.Func, v.Depth, v.Ceiling),
		})
	}
	return diags
}

func runFunctionLength(root string, qc harness.QualityConfig, glob string, sev Severity) []Diagnostic {
	ceiling, _ := qc.EffectiveCeiling("go")
	floor, _ := qc.EffectiveFloor("go")
	if ceiling == 0 && floor == 0 {
		return nil
	}
	violations, err := AnalyzeFunctionLength(root, ceiling, floor, glob)
	if err != nil {
		return []Diagnostic{{
			Severity: SeverityWarning, Rule: qc.RuleID,
			Message: fmt.Sprintf("structural check failed: %v", err),
		}}
	}
	var diags []Diagnostic
	for _, v := range violations {
		var msg string
		if v.Kind == "ceiling" {
			msg = fmt.Sprintf("%s: function length %d exceeds ceiling %d", v.Func, v.Length, v.Threshold)
		} else {
			msg = fmt.Sprintf("%s: function length %d below floor %d", v.Func, v.Length, v.Threshold)
		}
		diags = append(diags, Diagnostic{
			File: v.File, Line: v.Line, Severity: sev, Rule: qc.RuleID,
			Message: msg,
		})
	}
	return diags
}

// ParamViolation records a function with too many parameters.
type ParamViolation struct {
	File    string
	Func    string
	Line    int
	Count   int
	Ceiling int
}

// AnalyzeParamsPerFunction counts function parameters and returns violations
// where the count exceeds ceiling.
func AnalyzeParamsPerFunction(root string, ceiling int, glob string) ([]ParamViolation, error) {
	files, err := collectGoFiles(root, glob)
	if err != nil {
		return nil, err
	}

	var violations []ParamViolation
	fset := token.NewFileSet()

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			relPath = path
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Type.Params == nil {
				continue
			}

			count := 0
			for _, field := range fn.Type.Params.List {
				if len(field.Names) == 0 {
					count++
				} else {
					count += len(field.Names)
				}
			}

			if count > ceiling {
				line := fset.Position(fn.Pos()).Line
				violations = append(violations, ParamViolation{
					File:    relPath,
					Func:    funcName(fn),
					Line:    line,
					Count:   count,
					Ceiling: ceiling,
				})
			}
		}
	}
	return violations, nil
}

func runParamsPerFunction(root string, qc harness.QualityConfig, glob string, sev Severity) []Diagnostic {
	ceiling, ok := qc.EffectiveCeiling("go")
	if !ok {
		return nil
	}
	violations, err := AnalyzeParamsPerFunction(root, ceiling, glob)
	if err != nil {
		return []Diagnostic{{
			Severity: SeverityWarning, Rule: qc.RuleID,
			Message: fmt.Sprintf("structural check failed: %v", err),
		}}
	}
	var diags []Diagnostic
	for _, v := range violations {
		diags = append(diags, Diagnostic{
			File: v.File, Line: v.Line, Severity: sev, Rule: qc.RuleID,
			Message: fmt.Sprintf("%s: %d params exceeds ceiling %d", v.Func, v.Count, v.Ceiling),
		})
	}
	return diags
}

// FanOutViolation records a file with too many unique import paths.
type FanOutViolation struct {
	File    string
	Line    int
	Count   int
	Ceiling int
}

// AnalyzeFanOut counts unique import paths per file and returns violations
// where the count exceeds ceiling.
func AnalyzeFanOut(root string, ceiling int, glob string) ([]FanOutViolation, error) {
	files, err := collectGoFiles(root, glob)
	if err != nil {
		return nil, err
	}

	var violations []FanOutViolation
	fset := token.NewFileSet()

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			relPath = path
		}

		count := len(f.Imports)
		if count > ceiling {
			violations = append(violations, FanOutViolation{
				File:    relPath,
				Line:    1,
				Count:   count,
				Ceiling: ceiling,
			})
		}
	}
	return violations, nil
}

func runFanOut(root string, qc harness.QualityConfig, glob string, sev Severity) []Diagnostic {
	ceiling, ok := qc.EffectiveCeiling("go")
	if !ok {
		return nil
	}
	violations, err := AnalyzeFanOut(root, ceiling, glob)
	if err != nil {
		return []Diagnostic{{
			Severity: SeverityWarning, Rule: qc.RuleID,
			Message: fmt.Sprintf("structural check failed: %v", err),
		}}
	}
	var diags []Diagnostic
	for _, v := range violations {
		diags = append(diags, Diagnostic{
			File: v.File, Line: v.Line, Severity: sev, Rule: qc.RuleID,
			Message: fmt.Sprintf("fan-out %d imports exceeds ceiling %d", v.Count, v.Ceiling),
		})
	}
	return diags
}

// LOCViolation records a file with too many lines of code.
type LOCViolation struct {
	File    string
	Count   int
	Ceiling int
}

// AnalyzeLOCPerFile counts non-blank, non-comment lines per Go file and
// returns violations where the count exceeds ceiling.
func AnalyzeLOCPerFile(root string, ceiling int, glob string) ([]LOCViolation, error) {
	files, err := collectGoFiles(root, glob)
	if err != nil {
		return nil, err
	}

	var violations []LOCViolation

	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		relPath, _ := filepath.Rel(root, path)
		if relPath == "" || strings.HasPrefix(relPath, "..") {
			relPath = path
		}

		loc := countLOC(string(data))
		if loc > ceiling {
			violations = append(violations, LOCViolation{
				File:    relPath,
				Count:   loc,
				Ceiling: ceiling,
			})
		}
	}
	return violations, nil
}

func countLOC(source string) int {
	count := 0
	inBlock := false
	for _, line := range strings.Split(source, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "/*") {
			inBlock = true
		}
		if inBlock {
			if strings.Contains(trimmed, "*/") {
				inBlock = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		count++
	}
	return count
}

func runLOCPerFile(root string, qc harness.QualityConfig, glob string, sev Severity) []Diagnostic {
	ceiling, ok := qc.EffectiveCeiling("go")
	if !ok {
		return nil
	}
	violations, err := AnalyzeLOCPerFile(root, ceiling, glob)
	if err != nil {
		return []Diagnostic{{
			Severity: SeverityWarning, Rule: qc.RuleID,
			Message: fmt.Sprintf("structural check failed: %v", err),
		}}
	}
	var diags []Diagnostic
	for _, v := range violations {
		diags = append(diags, Diagnostic{
			File: v.File, Severity: sev, Rule: qc.RuleID,
			Message: fmt.Sprintf("LOC %d exceeds ceiling %d", v.Count, v.Ceiling),
		})
	}
	return diags
}

// SymbolResolution describes the result of resolving a symbol reference.
type SymbolResolution struct {
	PackageExists bool
	SymbolExists  bool
}

// ResolveSymbol checks whether a "pkg/path.SymbolName" reference resolves
// to a real exported identifier in the project rooted at projectRoot.
// modulePath is the Go module path (from go.mod).
func ResolveSymbol(projectRoot, modulePath, symbolRef string) SymbolResolution {
	dot := strings.LastIndex(symbolRef, ".")
	if dot < 0 {
		return SymbolResolution{}
	}
	pkgPath := symbolRef[:dot]
	symbolName := symbolRef[dot+1:]
	if pkgPath == "" || symbolName == "" {
		return SymbolResolution{}
	}

	var pkgDir string
	if strings.HasPrefix(pkgPath, modulePath+"/") {
		pkgDir = filepath.Join(projectRoot, strings.TrimPrefix(pkgPath, modulePath+"/"))
	} else if pkgPath == modulePath {
		pkgDir = projectRoot
	} else {
		pkgDir = filepath.Join(projectRoot, pkgPath)
	}

	info, err := os.Stat(pkgDir)
	if err != nil || !info.IsDir() {
		return SymbolResolution{PackageExists: false}
	}

	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, pkgDir, func(fi os.FileInfo) bool {
		return !strings.HasSuffix(fi.Name(), "_test.go")
	}, 0)
	if err != nil || len(pkgs) == 0 {
		return SymbolResolution{PackageExists: true, SymbolExists: false}
	}

	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Recv == nil && d.Name.Name == symbolName {
						return SymbolResolution{PackageExists: true, SymbolExists: true}
					}
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.Name == symbolName {
								return SymbolResolution{PackageExists: true, SymbolExists: true}
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.Name == symbolName {
									return SymbolResolution{PackageExists: true, SymbolExists: true}
								}
							}
						}
					}
				}
			}
		}
	}

	return SymbolResolution{PackageExists: true, SymbolExists: false}
}

// ReadGoModulePath reads the module path from a go.mod file.
func ReadGoModulePath(goModPath string) (string, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s: no module directive found", goModPath)
}
