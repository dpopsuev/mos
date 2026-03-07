package names

import "io/fs"

const (
	MosDir     = ".mos"
	ConfigFile = "config.mos"

	ActiveDir  = "active"
	ArchiveDir = "archive"

	StatusDraft     = "draft"
	StatusActive    = "active"
	StatusComplete  = "complete"
	StatusAbandoned = "abandoned"

	FieldTitle  = "title"
	FieldStatus = "status"
	FieldGoal   = "goal"
	FieldKind   = "kind"
	FieldLabels = "labels"

	BlockScope      = "scope"
	BlockCoverage   = "coverage"
	BlockTestMatrix = "test_matrix"
	BlockPersonas   = "personas"

	FormatText = "text"
	FormatJSON = "json"

	KindContract      = "contract"
	KindSpecification = "specification"
	KindRule          = "rule"
	KindBinder        = "binder"
	KindLexicon       = "lexicon"
	KindSprint        = "sprint"
	KindBatch         = "batch"
	KindNeed          = "need"
	KindArchitecture  = "architecture"

	DirContracts      = "contracts"
	DirSpecifications = "specifications"
	DirRules          = "rules"
	DirBinders        = "binders"
	DirSprints        = "sprints"
	DirBatches        = "batches"
	DirNeeds          = "needs"
	DirArchitectures  = "architectures"

	DirPerm  fs.FileMode = 0755
	FilePerm fs.FileMode = 0644
)
