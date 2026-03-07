package harness

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// QualityConfig describes a quality metric extracted from a rule's config block.
type QualityConfig struct {
	RuleID      string
	Metric      string
	Ceiling     *int
	Floor       *int
	Glob        string
	Enforcement string
	Trigger     string // "manual", "pre-commit", "post-commit"; empty defaults to "manual"
	Vector      string // "functional", "structural", "performance", or "" (untagged)
	Overrides   []LanguageOverride
}

// LanguageOverride allows per-language threshold adjustments.
type LanguageOverride struct {
	Language string
	Ceiling  *int
	Floor    *int
}

// EffectiveCeiling returns the ceiling for a given language, falling back to the default.
func (qc *QualityConfig) EffectiveCeiling(lang string) (int, bool) {
	for _, o := range qc.Overrides {
		if strings.EqualFold(o.Language, lang) && o.Ceiling != nil {
			return *o.Ceiling, true
		}
	}
	if qc.Ceiling != nil {
		return *qc.Ceiling, true
	}
	return 0, false
}

// EffectiveFloor returns the floor for a given language, falling back to the default.
func (qc *QualityConfig) EffectiveFloor(lang string) (int, bool) {
	for _, o := range qc.Overrides {
		if strings.EqualFold(o.Language, lang) && o.Floor != nil {
			return *o.Floor, true
		}
	}
	if qc.Floor != nil {
		return *qc.Floor, true
	}
	return 0, false
}

// DiscoverQualityConfigs scans rule artifacts in mosDir for config blocks.
func DiscoverQualityConfigs(mosDir string) ([]QualityConfig, error) {
	var configs []QualityConfig

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
			qc, err := extractQualityConfig(path)
			if err != nil {
				return nil, fmt.Errorf("parsing %s: %w", path, err)
			}
			if qc != nil {
				configs = append(configs, *qc)
			}
		}
	}

	return configs, nil
}

func extractQualityConfig(path string) (*QualityConfig, error) {
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

	configBlk := dsl.FindBlock(ab.Items, "config")
	if configBlk == nil {
		return nil, nil
	}

	metric, _ := dsl.FieldString(configBlk.Items, "metric")
	if metric == "" {
		return nil, nil
	}

	enforcement, _ := dsl.FieldString(ab.Items, "enforcement")
	if enforcement == "" {
		enforcement = "warning"
	}

	glob, _ := dsl.FieldString(ab.Items, "glob")

	trigger, _ := dsl.FieldString(ab.Items, "trigger")
	if trigger == "" {
		trigger = "manual"
	}

	vector, _ := dsl.FieldString(ab.Items, "vector")

	qc := &QualityConfig{
		RuleID:      ab.Name,
		Metric:      metric,
		Glob:        glob,
		Enforcement: enforcement,
		Trigger:     trigger,
		Vector:      vector,
	}

	if ceil, ok := dsl.FieldInt(configBlk.Items, "ceiling"); ok {
		v := int(ceil)
		qc.Ceiling = &v
	}
	if flr, ok := dsl.FieldInt(configBlk.Items, "floor"); ok {
		v := int(flr)
		qc.Floor = &v
	}

	for _, item := range configBlk.Items {
		blk, ok := item.(*dsl.Block)
		if !ok || blk.Name != "language" || blk.Title == "" {
			continue
		}
		override := LanguageOverride{Language: blk.Title}
		if ceil, ok := dsl.FieldInt(blk.Items, "ceiling"); ok {
			v := int(ceil)
			override.Ceiling = &v
		}
		if flr, ok := dsl.FieldInt(blk.Items, "floor"); ok {
			v := int(flr)
			override.Floor = &v
		}
		qc.Overrides = append(qc.Overrides, override)
	}

	return qc, nil
}
