package subsystem

import (
	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/cmd/mos/binder"
	"github.com/dpopsuev/mos/cmd/mos/config"
	"github.com/dpopsuev/mos/cmd/mos/contract"
	"github.com/dpopsuev/mos/cmd/mos/generic"
	"github.com/dpopsuev/mos/cmd/mos/govern"
	"github.com/dpopsuev/mos/cmd/mos/lexicon"
	"github.com/dpopsuev/mos/cmd/mos/rule"
	"github.com/dpopsuev/mos/cmd/mos/spec"
	"github.com/dpopsuev/mos/cmd/mos/tracecmd"
	"github.com/dpopsuev/mos/moslib/registry"
)

// GovCmd returns the "gov" subsystem command with all governance authoring
// subcommands registered.
func GovCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gov",
		Short: "Governance authoring & lifecycle",
		Long: `Artifact CRUD, field manipulation, formatting, project setup.

Resource-first: contract, rule, spec, binder, lexicon, config
Verb-first:     show, create, get, set, append, why, chain
Operations:     query, update, fmt, init, migrate, reclassify, archive, status`,
	}

	spec.Cmd.AddCommand(tracecmd.SpecGenCmd)

	cmd.AddCommand(
		contract.Cmd,
		rule.Cmd,
		spec.Cmd,
		binder.Cmd,
		lexicon.Cmd,
		config.Cmd,
		contract.ChainCmd,
	)

	cmd.AddCommand(
		govern.ShowCmd,
		govern.GetCmd,
		govern.SetCmd,
		govern.AppendCmd,
		govern.VerbCreateCmd,
		govern.WhyCmd,
	)

	cmd.AddCommand(
		govern.StatusCmd,
		govern.QueryCmd,
		govern.UpdateCmd,
		govern.MigrateCmd,
		govern.InitProjectCmd,
		govern.FmtCmd,
		govern.ReclassifyCmd,
		govern.ArchiveCmd,
	)

	reg, err := registry.LoadRegistry(".")
	if err == nil {
		for _, td := range reg.Types {
			if !registry.CoreKinds[td.Kind] {
				cmd.AddCommand(generic.NewCmd(td))
			}
		}
	}

	return cmd
}
