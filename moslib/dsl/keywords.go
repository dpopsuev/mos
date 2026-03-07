package dsl

// KeywordMap bridges human-protocol keywords (what developers write in .mos
// files) and machine-protocol keywords (internal Go identifiers). The
// lexicon file defines the human protocol; the machine protocol is fixed
// English.
type KeywordMap struct {
	ToMachine map[string]string // human keyword -> machine keyword
	ToHuman   map[string]string // machine keyword -> human keyword
}

// allMachineKeywords is the canonical set of translatable keywords.
var allMachineKeywords = []string{
	"feature", "background", "scenario", "given", "when", "then",
	"group", "spec", "include",
	"rule", "contract", "config", "declaration", "lexicon", "layers",
	"layer",
}

// DefaultKeywords returns the English identity mapping where human and
// machine keywords are identical.
func DefaultKeywords() *KeywordMap {
	kw := &KeywordMap{
		ToMachine: make(map[string]string, len(allMachineKeywords)),
		ToHuman:   make(map[string]string, len(allMachineKeywords)),
	}
	for _, k := range allMachineKeywords {
		kw.ToMachine[k] = k
		kw.ToHuman[k] = k
	}
	return kw
}

// AddKeywords registers additional machine keywords in a KeywordMap
// so that Format() can round-trip custom artifact type names.
func AddKeywords(kw *KeywordMap, keywords ...string) {
	for _, k := range keywords {
		if _, exists := kw.ToMachine[k]; !exists {
			kw.ToMachine[k] = k
			kw.ToHuman[k] = k
		}
	}
}

// machineKeyword translates a human keyword to its machine equivalent.
// Returns the word unchanged if no mapping exists.
func (kw *KeywordMap) machineKeyword(human string) string {
	if kw == nil {
		return human
	}
	if m, ok := kw.ToMachine[human]; ok {
		return m
	}
	return human
}

// humanKeyword translates a machine keyword to its human equivalent.
// Returns the word unchanged if no mapping exists.
func (kw *KeywordMap) humanKeyword(machine string) string {
	if kw == nil {
		return machine
	}
	if h, ok := kw.ToHuman[machine]; ok {
		return h
	}
	return machine
}

// ExtractKeywords builds a KeywordMap from a parsed lexicon artifact.
// The lexicon must contain a "keywords" block where keys are machine
// names and values are human-protocol strings:
//
//	lexicon {
//	  keywords {
//	    feature = "característica"
//	    rule    = "regla"
//	  }
//	}
//
// Missing entries fall back to English defaults (identity mapping).
func ExtractKeywords(lexicon *File) *KeywordMap {
	kw := DefaultKeywords()
	if lexicon == nil {
		return kw
	}

	ab, ok := lexicon.Artifact.(*ArtifactBlock)
	if !ok {
		return kw
	}

	for _, item := range ab.Items {
		blk, ok := item.(*Block)
		if !ok || blk.Name != "keywords" {
			continue
		}
		for _, bi := range blk.Items {
			f, ok := bi.(*Field)
			if !ok {
				continue
			}
			sv, ok := f.Value.(*StringVal)
			if !ok {
				continue
			}
			machineKey := f.Key
			humanKey := sv.Text
			kw.ToMachine[humanKey] = machineKey
			kw.ToHuman[machineKey] = humanKey
		}
		break
	}

	return kw
}
