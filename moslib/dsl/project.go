package dsl

import "fmt"

// LoadKeywords parses a lexicon source with English defaults and extracts
// a KeywordMap. Returns DefaultKeywords() if src is empty.
func LoadKeywords(lexiconSrc string) (*KeywordMap, error) {
	if lexiconSrc == "" {
		return DefaultKeywords(), nil
	}
	lexicon, err := Parse(lexiconSrc, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing lexicon: %w", err)
	}
	return ExtractKeywords(lexicon), nil
}

// ParseFiles orchestrates the two-phase lexicon-driven parse. vocabSrc
// is the lexicon file source (may be empty for English-only projects).
// Each entry in sources is parsed with the resulting KeywordMap.
func ParseFiles(lexiconSrc string, sources map[string]string) (map[string]*File, *KeywordMap, error) {
	kw, err := LoadKeywords(lexiconSrc)
	if err != nil {
		return nil, nil, err
	}
	files := make(map[string]*File, len(sources))
	for name, src := range sources {
		f, err := Parse(src, kw)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing %s: %w", name, err)
		}
		files[name] = f
	}
	return files, kw, nil
}
