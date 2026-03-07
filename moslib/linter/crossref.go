package linter

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
	"github.com/dpopsuev/mos/moslib/schema"
)

func validateCrossRefs(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	for id, path := range ctx.RuleIDs {
		diags = append(diags, validateRuleCrossRefs(id, path, ctx)...)
	}

	for _, path := range ctx.ContractIDs {
		diags = append(diags, validateContractCrossRefs(path, ctx)...)
	}

	diags = append(diags, validateLifecycleChainLinks(ctx)...)
	diags = append(diags, validateCriterionLinkage(ctx)...)
	diags = append(diags, validateLifecycleOrphans(ctx)...)
	diags = append(diags, validateScopeViolations(ctx)...)
	diags = append(diags, validatePreamble(ctx)...)
	diags = append(diags, validateSpecOnOriginating(ctx)...)
	diags = append(diags, validateIDCollisions(ctx)...)
	diags = append(diags, validatePolicies(ctx)...)
	diags = append(diags, validateStaleSprints(ctx)...)
	diags = append(diags, validateTerminalInActive(ctx)...)
	diags = append(diags, validateDirectiveAlignment(ctx)...)
	diags = append(diags, validateBlameRefs(ctx)...)

	return diags
}

func validateLifecycleChainLinks(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	allArtifacts := make(map[string]map[string]string)
	for kind, ids := range ctx.ArtifactIDs {
		allArtifacts[kind] = ids
	}
	for id, path := range ctx.ContractIDs {
		if allArtifacts["contract"] == nil {
			allArtifacts["contract"] = make(map[string]string)
		}
		allArtifacts["contract"][id] = path
	}

	allIDs := make(map[string]bool)
	for _, ids := range allArtifacts {
		for id := range ids {
			allIDs[id] = true
		}
	}

	for kind, ids := range allArtifacts {
		for id, path := range ids {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}

			linkFields := linkFieldsForKind(kind, ctx)
			for _, linkField := range linkFields {
				for _, target := range astFieldStringSlice(ab.Items, linkField) {
					if target != "" && !allIDs[target] {
						diags = append(diags, Diagnostic{
							File:     path,
							Severity: SeverityError,
							Rule:     "lifecycle-chain",
							Message:  fmt.Sprintf("%s %q: %s references %q which does not exist", kind, id, linkField, target),
						})
					}
				}
			}
		}
	}

	return diags
}

func validateLifecycleOrphans(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	allArtifacts := make(map[string]map[string]string)
	for kind, ids := range ctx.ArtifactIDs {
		allArtifacts[kind] = ids
	}
	for id, path := range ctx.ContractIDs {
		if allArtifacts["contract"] == nil {
			allArtifacts["contract"] = make(map[string]string)
		}
		allArtifacts["contract"][id] = path
	}

	reverseIndex := make(map[string]bool)
	for kind, ids := range allArtifacts {
		for _, path := range ids {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			linkFields := linkFieldsForKind(kind, ctx)
			for _, linkField := range linkFields {
				for _, target := range astFieldStringSlice(ab.Items, linkField) {
					if target != "" {
						reverseIndex[linkField+":"+target] = true
					}
				}
			}
		}
	}

	for _, schema := range ctx.CustomArtifacts {
		ed := schema.ExpectsDownstream
		if ed == nil {
			continue
		}
		ids, ok := ctx.ArtifactIDs[schema.Kind]
		if !ok {
			continue
		}
		for id, path := range ids {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			status, _ := astFieldString(ab.Items, "status")
			if !isPastThreshold(status, ed.After, schema) {
				continue
			}
			key := ed.Via + ":" + id
			if !reverseIndex[key] {
				sev := SeverityWarning
				if ed.Severity == "error" {
					sev = SeverityError
				}
				sev = promoteByUrgency(sev, ab, schema)
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: sev,
					Rule:     "lifecycle-orphan",
					Message:  fmt.Sprintf("%s %q has no downstream artifact referencing it via %q", schema.Kind, id, ed.Via),
				})
			}
		}
	}

	return diags
}

func isPastThreshold(status, threshold string, sch schema.ArtifactSchema) bool {
	if status == "" || threshold == "" {
		return false
	}
	allStates := append(sch.ActiveStates, sch.ArchiveStates...)
	thresholdIdx := -1
	statusIdx := -1
	for i, s := range allStates {
		if s == threshold {
			thresholdIdx = i
		}
		if s == status {
			statusIdx = i
		}
	}
	if thresholdIdx < 0 || statusIdx < 0 {
		return status == threshold
	}
	return statusIdx >= thresholdIdx
}

func validateRuleCrossRefs(id, path string, ctx *ProjectContext) []Diagnostic {
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		return nil
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil
	}

	var diags []Diagnostic

	if scope, ok := astFieldString(ab.Items, "scope"); ok && len(ctx.LayerSet) > 0 {
		if !ctx.LayerSet[scope] {
			diags = append(diags, Diagnostic{
				File: path, Severity: SeverityError, Rule: "layer-ref",
				Message: fmt.Sprintf("rule %q scope %q is not a declared resolution layer", id, scope),
			})
		}
	}

	if tags := astFieldStringSlice(ab.Items, "tags"); len(tags) > 0 && ctx.Lexicon != nil && len(ctx.Lexicon.Terms) > 0 {
		for _, tag := range tags {
			if _, ok := ctx.Lexicon.Terms[strings.ToLower(tag)]; !ok {
				diags = append(diags, Diagnostic{
					File: path, Severity: SeverityWarning, Rule: "vocab-term",
					Message: fmt.Sprintf("rule %q tag %q not found in lexicon", id, tag),
				})
			}
		}
	}

	return diags
}

func validateContractCrossRefs(path string, ctx *ProjectContext) []Diagnostic {
	f, err := parseDSLFile(path, ctx.Keywords)
	if err != nil {
		return nil
	}

	ab, ok := f.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return nil
	}

	var diags []Diagnostic

	execution := astFindBlock(ab.Items, "execution")
	if execution != nil {
		for _, fieldName := range []string{"rules_override", "rules_suspended"} {
			for _, ruleID := range astFieldStringSlice(execution.Items, fieldName) {
				if _, ok := ctx.RuleIDs[ruleID]; !ok {
					diags = append(diags, Diagnostic{
						File: path, Severity: SeverityError, Rule: "crossref-rule",
						Message: fmt.Sprintf("execution.%s references rule %q which does not exist", fieldName, ruleID),
					})
				}
			}
		}
	}

	if len(ctx.TemplateBlocks) > 0 {
		blockNames := make(map[string]bool)
		for _, item := range ab.Items {
			if blk, ok := item.(*dsl.Block); ok {
				blockNames[blk.Name] = true
			}
		}
		for _, expected := range ctx.TemplateBlocks {
			if !blockNames[expected] {
				diags = append(diags, Diagnostic{
					File: path, Severity: SeverityWarning, Rule: "template-conformance",
					Message: fmt.Sprintf("missing block '%s' expected by project template", expected),
				})
			}
		}
	}

	return diags
}

func validateCriterionLinkage(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	needCriteria := make(map[string]map[string]bool)
	needPaths := make(map[string]string)

	if needIDs, ok := ctx.ArtifactIDs["need"]; ok {
		for id, path := range needIDs {
			needPaths[id] = path
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			criteriaNames := make(map[string]bool)
			acceptBlk := astFindBlock(ab.Items, "acceptance")
			if acceptBlk != nil {
				for _, item := range acceptBlk.Items {
					sub, ok := item.(*dsl.Block)
					if !ok || sub.Name != "criterion" || sub.Title == "" {
						continue
					}
					criteriaNames[sub.Title] = true
				}
			}
			if len(criteriaNames) > 0 {
				needCriteria[id] = criteriaNames
			}
		}
	}

	if len(needCriteria) == 0 {
		return diags
	}

	coveredCriteria := make(map[string]map[string]bool)

	allSpecs := make(map[string]string)
	if specIDs, ok := ctx.ArtifactIDs["specification"]; ok {
		for id, path := range specIDs {
			allSpecs[id] = path
		}
	}

	for specID, path := range allSpecs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		satisfiesList := astFieldStringSlice(ab.Items, "satisfies")
		if len(satisfiesList) == 0 {
			continue
		}
		addresses := astFieldStringSlice(ab.Items, "addresses")
		for _, needID := range satisfiesList {
			criteria, hasCriteria := needCriteria[needID]
			for _, addr := range addresses {
				if hasCriteria && !criteria[addr] {
					diags = append(diags, Diagnostic{
						File:     path,
						Severity: SeverityError,
						Rule:     "criterion-exists",
						Message:  fmt.Sprintf("specification %q addresses criterion %q which does not exist on %s", specID, addr, needID),
					})
				}
				if hasCriteria && criteria[addr] {
					if coveredCriteria[needID] == nil {
						coveredCriteria[needID] = make(map[string]bool)
					}
					coveredCriteria[needID][addr] = true
				}
			}
		}
	}

	needSchema := findCustomSchema(ctx, "need")

	for needID, criteria := range needCriteria {
		needPath := needPaths[needID]
		f, err := parseDSLFile(needPath, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		status, _ := astFieldString(ab.Items, "status")
		if status != "validated" && status != "addressed" {
			continue
		}
		covered := coveredCriteria[needID]
		for name := range criteria {
			if covered == nil || !covered[name] {
				sev := SeverityWarning
				if needSchema != nil {
					sev = promoteByUrgency(sev, ab, *needSchema)
				}
				diags = append(diags, Diagnostic{
					File:     needPath,
					Severity: sev,
					Rule:     "criterion-coverage",
					Message:  fmt.Sprintf("need %q criterion %q is not addressed by any specification", needID, name),
				})
			}
		}
	}

	return diags
}

func validateScopeViolations(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	needExcludes := make(map[string][]string)
	if needIDs, ok := ctx.ArtifactIDs["need"]; ok {
		for id, path := range needIDs {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			scopeBlk := astFindBlock(ab.Items, "scope")
			if scopeBlk != nil {
				excludes := astFieldStringSlice(scopeBlk.Items, "excludes")
				if len(excludes) > 0 {
					needExcludes[id] = excludes
				}
			}
		}
	}

	allSpecs := make(map[string]string)
	if specIDs, ok := ctx.ArtifactIDs["specification"]; ok {
		for id, path := range specIDs {
			allSpecs[id] = path
		}
	}

	for specID, path := range allSpecs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		satisfiesList := astFieldStringSlice(ab.Items, "satisfies")
		if len(satisfiesList) == 0 {
			continue
		}
		title, _ := astFieldString(ab.Items, "title")
		titleTokens := tokenize(title)
		for _, needID := range satisfiesList {
			excludes, hasExcludes := needExcludes[needID]
			if !hasExcludes {
				continue
			}
		for _, excl := range excludes {
			exclTokens := tokenize(excl)
			allPresent := len(exclTokens) > 0
			for et := range exclTokens {
				if !titleTokens[et] {
					allPresent = false
					break
				}
			}
			if allPresent {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityError,
					Rule:     "scope-violation",
					Message:  fmt.Sprintf("specification %q title contains %q which is excluded by %s scope", specID, excl, needID),
				})
			}
		}
		}
	}

	if archIDs, ok := ctx.ArtifactIDs["architecture"]; ok {
		for archID, path := range archIDs {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			diags = append(diags, validateForbiddenEdges(archID, path, ab)...)
		}
	}

	return diags
}

func validateForbiddenEdges(archID, path string, ab *dsl.ArtifactBlock) []Diagnostic {
	var diags []Diagnostic
	var edges []struct{ from, to string }
	var forbidden []struct{ from, to, reason string }

	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok {
			continue
		}
		switch blk.Name {
		case "edge":
			from, _ := dsl.FieldString(blk.Items, "from")
			to, _ := dsl.FieldString(blk.Items, "to")
			if from != "" && to != "" {
				edges = append(edges, struct{ from, to string }{from, to})
			}
		case "forbidden":
			from, _ := dsl.FieldString(blk.Items, "from")
			to, _ := dsl.FieldString(blk.Items, "to")
			reason, _ := dsl.FieldString(blk.Items, "reason")
			if from != "" && to != "" {
				forbidden = append(forbidden, struct{ from, to, reason string }{from, to, reason})
			}
		}
	}

	for _, f := range forbidden {
		for _, e := range edges {
			if e.from == f.from && e.to == f.to {
				msg := fmt.Sprintf("architecture %q has edge %s->%s which is forbidden", archID, f.from, f.to)
				if f.reason != "" {
					msg += fmt.Sprintf(" (%s)", f.reason)
				}
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityError,
					Rule:     "scope-violation",
					Message:  msg,
				})
			}
		}
	}

	return diags
}

func linkFieldsForKind(kind string, ctx *ProjectContext) []string {
	for i := range ctx.CustomArtifacts {
		if ctx.CustomArtifacts[i].Kind == kind {
			if fields := ctx.CustomArtifacts[i].LinkFieldNames(); len(fields) > 0 {
				return fields
			}
		}
	}
	if sch := schema.DefaultCoreSchemas()[kind]; sch != nil {
		return sch.LinkFieldNames()
	}
	return nil
}

func findCustomSchema(ctx *ProjectContext, kind string) *schema.ArtifactSchema {
	for i := range ctx.CustomArtifacts {
		if ctx.CustomArtifacts[i].Kind == kind {
			return &ctx.CustomArtifacts[i]
		}
	}
	if sch := schema.DefaultCoreSchemas()[kind]; sch != nil {
		return sch
	}
	return nil
}

func promoteByUrgency(baseSev Severity, ab *dsl.ArtifactBlock, sch schema.ArtifactSchema) Severity {
	if sch.UrgencyPropagation == nil {
		return baseSev
	}
	urgency, _ := astFieldString(ab.Items, "urgency")
	if urgency == "" {
		return baseSev
	}
	promoted, ok := sch.UrgencyPropagation[urgency]
	if !ok {
		return baseSev
	}
	switch promoted {
	case "error":
		return SeverityError
	case "warn":
		if baseSev == SeverityInfo {
			return SeverityWarning
		}
		return baseSev
	case "info":
		return baseSev
	case "ignore":
		return baseSev
	}
	return baseSev
}

func validatePreamble(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	needIDs, ok := ctx.ArtifactIDs["need"]
	if !ok || len(needIDs) == 0 {
		return diags
	}

	var preambleIDs []string
	var preamblePaths []string

	reverseRefs := make(map[string]bool)
	for kind, ids := range ctx.ArtifactIDs {
		if kind == "need" {
			continue
		}
		linkFields := linkFieldsForKind(kind, ctx)
		for _, path := range ids {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			for _, lf := range linkFields {
				for _, target := range astFieldStringSlice(ab.Items, lf) {
					reverseRefs[target] = true
				}
			}
		}
	}

	for id, path := range needIDs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		originatingVal, _ := astFieldString(ab.Items, "originating")
		if originatingVal == "true" {
			preambleIDs = append(preambleIDs, id)
			preamblePaths = append(preamblePaths, path)
		}
	}

	if len(preambleIDs) > 1 {
		for i, path := range preamblePaths {
			diags = append(diags, Diagnostic{
				File:     path,
				Severity: SeverityError,
				Rule:     "originating",
				Message:  fmt.Sprintf("need %q is marked as originating, but multiple originating needs exist: %v", preambleIDs[i], preambleIDs),
			})
		}
	}

	if len(preambleIDs) == 0 && len(needIDs) > 0 {
		for id, path := range needIDs {
			if !reverseRefs[id] {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityInfo,
					Rule:     "originating",
					Message:  fmt.Sprintf("need %q is a top-level need (no upstream link) but is not marked as originating", id),
				})
			}
		}
	}

	return diags
}

func validateSpecOnOriginating(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	needIDs, ok := ctx.ArtifactIDs["need"]
	if !ok || len(needIDs) == 0 {
		return diags
	}

	originatingIDs := make(map[string]bool)
	for id, path := range needIDs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		val, _ := astFieldString(ab.Items, "originating")
		if val == "true" {
			originatingIDs[id] = true
		}
	}

	if len(originatingIDs) == 0 {
		return diags
	}

	hasDerived := false
	for id, path := range needIDs {
		if originatingIDs[id] {
			continue
		}
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		for _, target := range astFieldStringSlice(ab.Items, "derives_from") {
			if originatingIDs[target] {
				hasDerived = true
				break
			}
		}
		if hasDerived {
			break
		}
		_ = path
	}

	if !hasDerived {
		return diags
	}

	specIDs := ctx.ArtifactIDs["specification"]
	for id, path := range specIDs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		for _, target := range astFieldStringSlice(ab.Items, "satisfies") {
			if originatingIDs[target] {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityWarning,
					Rule:     "spec-on-originating",
					Message:  fmt.Sprintf("specification %q directly satisfies originating need %q -- consider a derived need", id, target),
				})
			}
		}
	}

	return diags
}

func validateIDCollisions(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	type dirPair struct {
		kind string
		base string
	}

	pairs := []dirPair{
		{"contract", filepath.Join(ctx.Root, "contracts")},
	}
	for _, schema := range ctx.CustomArtifacts {
		pairs = append(pairs, dirPair{schema.Kind, filepath.Join(ctx.Root, schema.Directory)})
	}

	for _, p := range pairs {
		activeIDs := dirNames(filepath.Join(p.base, "active"))
		archiveIDs := dirNames(filepath.Join(p.base, "archive"))
		for id := range activeIDs {
			if archiveIDs[id] {
				diags = append(diags, Diagnostic{
					File:     filepath.Join(p.base, "active", id),
					Severity: SeverityError,
					Rule:     "id-collision",
					Message:  fmt.Sprintf("%s %q exists in both active and archive directories", p.kind, id),
				})
			}
		}
	}

	return diags
}

func dirNames(dir string) map[string]bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	names := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			names[e.Name()] = true
		}
	}
	return names
}

func validateSlugs(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic
	type slugEntry struct {
		id   string
		path string
	}
	slugs := map[string]slugEntry{}
	isKebab := func(s string) bool {
		for _, c := range s {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
		return len(s) > 0 && s[0] != '-' && s[len(s)-1] != '-'
	}

	checkSlug := func(id, path string, items []dsl.Node, isArchive bool) {
		slug, hasSlug := astFieldString(items, "slug")
		if !hasSlug || slug == "" {
			if !isArchive {
				diags = append(diags, Diagnostic{
					File: path, Severity: SeverityWarning, Rule: "slug-missing",
					Message: fmt.Sprintf("artifact %q has no slug", id),
				})
			}
			return
		}
		if !isKebab(slug) {
			diags = append(diags, Diagnostic{
				File: path, Severity: SeverityError, Rule: "slug-format",
				Message: fmt.Sprintf("artifact %q slug %q must be lowercase kebab-case (a-z, 0-9, hyphens)", id, slug),
			})
		}
		if prev, dup := slugs[slug]; dup {
			diags = append(diags, Diagnostic{
				File: path, Severity: SeverityError, Rule: "slug-unique",
				Message: fmt.Sprintf("artifact %q slug %q collides with %q (%s)", id, slug, prev.id, prev.path),
			})
		} else {
			slugs[slug] = slugEntry{id: id, path: path}
		}
	}

	for _, sub := range []string{"active", "archive"} {
		isArchive := sub == "archive"
		contractDir := filepath.Join(ctx.Root, "contracts", sub)
		entries, _ := os.ReadDir(contractDir)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractDir, e.Name(), "contract.mos")
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			f, err := dsl.Parse(string(data), nil)
			if err != nil {
				continue
			}
			if ab, ok := f.Artifact.(*dsl.ArtifactBlock); ok {
				checkSlug(ab.Name, path, ab.Items, isArchive)
			}
		}
	}

	for _, schema := range ctx.CustomArtifacts {
		for _, sub := range []string{"active", "archive"} {
			isArchive := sub == "archive"
			dir := filepath.Join(ctx.Root, schema.Directory, sub)
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				path := filepath.Join(dir, e.Name(), schema.Kind+".mos")
				data, err := os.ReadFile(path)
				if err != nil {
					continue
				}
				f, err := dsl.Parse(string(data), nil)
				if err != nil {
					continue
				}
				if ab, ok := f.Artifact.(*dsl.ArtifactBlock); ok {
					checkSlug(ab.Name, path, ab.Items, isArchive)
				}
			}
		}
	}

	return diags
}

func tokenize(s string) map[string]bool {
	tokens := make(map[string]bool)
	s = strings.ToLower(s)
	for _, sep := range []string{" ", ",", ".", ";", ":", "-", "_", "/", "(", ")", "[", "]"} {
		s = strings.ReplaceAll(s, sep, " ")
	}
	for _, word := range strings.Fields(s) {
		if len(word) > 1 {
			tokens[word] = true
		}
	}
	return tokens
}

var idFormatRe = regexp.MustCompile(`^[A-Z]+(-\d{4})?-\d{1,}$`)

func validateIDFormat(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	prefixes := collectPrefixes(ctx)

	checkID := func(id, path string) {
		if !idFormatRe.MatchString(id) {
			diags = append(diags, Diagnostic{
				File: path, Severity: SeverityWarning, Rule: "id-format",
				Message: fmt.Sprintf("artifact ID %q does not match expected PREFIX-YYYY-NNN format", id),
			})
			return
		}
		if len(prefixes) > 0 {
			parts := strings.SplitN(id, "-", 2)
			if _, ok := prefixes[parts[0]]; !ok {
				diags = append(diags, Diagnostic{
					File: path, Severity: SeverityWarning, Rule: "id-prefix",
					Message: fmt.Sprintf("artifact ID %q has unknown prefix %q (known: %s)", id, parts[0], joinKeys(prefixes)),
				})
			}
		}
	}

	for _, sub := range []string{"active", "archive"} {
		contractDir := filepath.Join(ctx.Root, "contracts", sub)
		entries, _ := os.ReadDir(contractDir)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(contractDir, e.Name(), "contract.mos")
			if _, err := os.Stat(path); err != nil {
				continue
			}
			checkID(e.Name(), path)
		}
	}

	for _, schema := range ctx.CustomArtifacts {
		for _, sub := range []string{"active", "archive"} {
			dir := filepath.Join(ctx.Root, schema.Directory, sub)
			entries, _ := os.ReadDir(dir)
			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				path := filepath.Join(dir, e.Name(), schema.Kind+".mos")
				if _, err := os.Stat(path); err != nil {
					continue
				}
				checkID(e.Name(), path)
			}
		}
	}

	return diags
}

func collectPrefixes(ctx *ProjectContext) map[string]bool {
	prefixes := map[string]bool{}
	if ctx.Config == nil {
		return prefixes
	}
	ab, ok := ctx.Config.Artifact.(*dsl.ArtifactBlock)
	if !ok {
		return prefixes
	}
	for _, item := range ab.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "project" {
			continue
		}
		if p, ok := dsl.FieldString(blk.Items, "prefix"); ok && p != "" {
			prefixes[p] = true
		}
	}
	return prefixes
}

func joinKeys(m map[string]bool) string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

func validateStaleSprints(ctx *ProjectContext) []Diagnostic {
	sprintIDs, ok := ctx.ArtifactIDs["sprint"]
	if !ok || len(sprintIDs) == 0 {
		return nil
	}

	sprintSchema := findCustomSchema(ctx, "sprint")
	archiveStates := map[string]bool{"complete": true, "cancelled": true}
	if sprintSchema != nil {
		archiveStates = make(map[string]bool)
		for _, s := range sprintSchema.ArchiveStates {
			archiveStates[s] = true
		}
	}

	sprintStatuses := make(map[string]string)
	for id, path := range sprintIDs {
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		status, _ := astFieldString(ab.Items, "status")
		sprintStatuses[id] = status
	}

	var diags []Diagnostic
	for id, path := range ctx.ContractIDs {
		if !strings.Contains(path, "/active/") {
			continue
		}
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		sprintRef, _ := astFieldString(ab.Items, "sprint")
		if sprintRef == "" {
			continue
		}
		if status, exists := sprintStatuses[sprintRef]; exists && archiveStates[status] {
			diags = append(diags, Diagnostic{
				File:     path,
				Severity: SeverityWarning,
				Rule:     "stale-sprint-ref",
				Message:  fmt.Sprintf("contract %q references sprint %q which has status %q", id, sprintRef, status),
			})
		}
	}
	return diags
}

func validateDirectiveAlignment(ctx *ProjectContext) []Diagnostic {
	hasActiveDirective := false
	dirIDs, ok := ctx.ArtifactIDs["directive"]
	if ok {
		for _, path := range dirIDs {
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			status, _ := astFieldString(ab.Items, "status")
			if status == "active" || status == "declared" {
				hasActiveDirective = true
				break
			}
		}
	}
	if !hasActiveDirective {
		return nil
	}

	var diags []Diagnostic
	for id, path := range ctx.ContractIDs {
		if !strings.Contains(path, "/active/") {
			continue
		}
		f, err := parseDSLFile(path, ctx.Keywords)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}
		justifies := astFieldStringSlice(ab.Items, "justifies")
		if len(justifies) == 0 {
			diags = append(diags, Diagnostic{
				File:     path,
				Severity: SeverityInfo,
				Rule:     "directive-alignment",
				Message:  fmt.Sprintf("contract %q has no justifies link; consider aligning with the active directive", id),
			})
		}
	}
	return diags
}

func validateTerminalInActive(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	if contractSch := findCustomSchema(ctx, "contract"); contractSch != nil && len(contractSch.ArchiveStates) > 0 {
		archiveSet := make(map[string]bool)
		for _, s := range contractSch.ArchiveStates {
			archiveSet[s] = true
		}
		dir := filepath.Join(ctx.Root, "contracts", "active")
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), "contract.mos")
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			status, _ := astFieldString(ab.Items, "status")
			if archiveSet[status] {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityError,
					Rule:     "terminal-in-active",
					Message:  fmt.Sprintf("contract %q has terminal status %q but is still in active/ directory", ab.Name, status),
				})
			}
		}
	}

	for _, sch := range ctx.CustomArtifacts {
		if len(sch.ArchiveStates) == 0 {
			continue
		}
		archiveSet := make(map[string]bool)
		for _, s := range sch.ArchiveStates {
			archiveSet[s] = true
		}
		dir := filepath.Join(ctx.Root, sch.Directory, "active")
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			path := filepath.Join(dir, e.Name(), sch.Kind+".mos")
			f, err := parseDSLFile(path, ctx.Keywords)
			if err != nil {
				continue
			}
			ab, ok := f.Artifact.(*dsl.ArtifactBlock)
			if !ok {
				continue
			}
			status, _ := astFieldString(ab.Items, "status")
			if archiveSet[status] {
				diags = append(diags, Diagnostic{
					File:     path,
					Severity: SeverityError,
					Rule:     "terminal-in-active",
					Message:  fmt.Sprintf("%s %q has terminal status %q but is still in active/ directory", sch.Kind, ab.Name, status),
				})
			}
		}
	}

	return diags
}

func validateBlameRefs(ctx *ProjectContext) []Diagnostic {
	var diags []Diagnostic

	for _, artPath := range ctx.ContractIDs {
		data, err := os.ReadFile(artPath)
		if err != nil {
			continue
		}
		f, err := dsl.Parse(string(data), nil)
		if err != nil {
			continue
		}
		ab, ok := f.Artifact.(*dsl.ArtifactBlock)
		if !ok {
			continue
		}

		for _, item := range ab.Items {
			blk, ok := item.(*dsl.Block)
			if !ok || blk.Name != "blame" {
				continue
			}
			file, _ := astFieldString(blk.Items, "file")
			if file == "" {
				continue
			}
			projectRoot := filepath.Dir(ctx.Root)
			absPath := filepath.Join(projectRoot, file)
			if _, err := os.Stat(absPath); os.IsNotExist(err) {
				diags = append(diags, Diagnostic{
					File:     artPath,
					Severity: SeverityWarning,
					Rule:     "blame-ref-missing",
					Message:  fmt.Sprintf("blame block references %q but file does not exist", file),
				})
			}
		}
	}

	return diags
}
