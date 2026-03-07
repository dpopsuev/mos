// Package governance provides the top-level project initialization entry
// points for Mos. All other symbols have moved to their canonical packages:
//
//   - artifact: contracts, rules, specs, binders, lexicon, architecture, generic CRUD
//   - registry: type definitions, lifecycle, projects, ID generation
//   - store: persistence layer
//   - names: constants (directories, statuses, kinds, formats)
//   - arch: architecture model, rendering, churn
package governance

import "github.com/dpopsuev/mos/moslib/artifact"

func Init(root string, opts artifact.InitOpts) error { return artifact.Init(root, opts) }
func InitDynamicRegistry(root string) error           { return artifact.InitDynamicRegistry(root) }
