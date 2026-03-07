package linter

import (
	"os"

	"github.com/dpopsuev/mos/moslib/dsl"
)

func parseDSLFile(path string, kw *dsl.KeywordMap) (*dsl.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return dsl.Parse(string(data), kw)
}

func astFindBlock(items []dsl.Node, name string) *dsl.Block {
	return dsl.FindBlock(items, name)
}

func astFieldString(items []dsl.Node, key string) (string, bool) {
	return dsl.FieldString(items, key)
}

func astFieldInt(items []dsl.Node, key string) (int64, bool) {
	return dsl.FieldInt(items, key)
}

func astFieldStringSlice(items []dsl.Node, key string) []string {
	return dsl.FieldStringSlice(items, key)
}

func astHasField(items []dsl.Node, key string) bool {
	return dsl.HasField(items, key)
}

func astFindFeatures(items []dsl.Node) []*dsl.FeatureBlock {
	var features []*dsl.FeatureBlock
	for _, item := range items {
		if fb, ok := item.(*dsl.FeatureBlock); ok {
			features = append(features, fb)
		}
		if sb, ok := item.(*dsl.SpecBlock); ok {
			features = append(features, sb.Features...)
		}
	}
	return features
}

func astFindIncludes(items []dsl.Node) []*dsl.IncludeDirective {
	var includes []*dsl.IncludeDirective
	for _, item := range items {
		if sb, ok := item.(*dsl.SpecBlock); ok {
			includes = append(includes, sb.Includes...)
		}
	}
	return includes
}

// astCountTitledBlocks counts nested blocks that are sub-contracts
// (titled blocks with a "status" field), distinguishing them from
// other titled blocks like tracker adapters.
func astCountTitledBlocks(items []dsl.Node) int {
	n := 0
	for _, item := range items {
		if b, ok := item.(*dsl.Block); ok && b.Title != "" && dsl.HasField(b.Items, "status") {
			n++
		}
	}
	return n
}
