package artifact

import (
	"io/fs"

	"github.com/dpopsuev/mos/moslib/names"
)

const (
	MosDir     = names.MosDir
	ConfigFile = names.ConfigFile

	ActiveDir  = names.ActiveDir
	ArchiveDir = names.ArchiveDir

	StatusDraft     = names.StatusDraft
	StatusActive    = names.StatusActive
	StatusComplete  = names.StatusComplete
	StatusAbandoned = names.StatusAbandoned

	FieldTitle  = names.FieldTitle
	FieldStatus = names.FieldStatus
	FieldGoal   = names.FieldGoal
	FieldKind   = names.FieldKind
	FieldLabels = names.FieldLabels

	BlockScope      = names.BlockScope
	BlockCoverage   = names.BlockCoverage
	BlockTestMatrix = names.BlockTestMatrix
	BlockPersonas   = names.BlockPersonas

	FormatText = names.FormatText
	FormatJSON = names.FormatJSON

	KindContract      = names.KindContract
	KindSpecification = names.KindSpecification
	KindRule          = names.KindRule
	KindBinder        = names.KindBinder
	KindLexicon       = names.KindLexicon
	KindSprint        = names.KindSprint
	KindBatch         = names.KindBatch
	KindNeed          = names.KindNeed
	KindArchitecture  = names.KindArchitecture

	DirContracts      = names.DirContracts
	DirSpecifications = names.DirSpecifications
	DirRules          = names.DirRules
	DirBinders        = names.DirBinders
	DirSprints        = names.DirSprints
	DirBatches        = names.DirBatches
	DirNeeds          = names.DirNeeds
	DirArchitectures  = names.DirArchitectures

	DirPerm  fs.FileMode = names.DirPerm
	FilePerm fs.FileMode = names.FilePerm
)
