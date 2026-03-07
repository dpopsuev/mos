package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/dpopsuev/mos/cmd/internal/subsystem"
	"github.com/dpopsuev/mos/cmd/internal/wire"
	"github.com/dpopsuev/mos/cmd/mos/cliutil"
)

func main() {
	wire.Init()
	cmd := subsystem.TraceCmd()
	cmd.Use = "mtrace"

	if cliutil.IsAgentMode() {
		output, err := cliutil.CaptureStdout(func() error {
			return cmd.Execute()
		})
		cliutil.EmitAgentEnvelope(output, err)
		if err != nil {
			os.Exit(1)
		}
		return
	}

	if err := cmd.Execute(); err != nil {
		if errors.Is(err, cliutil.ErrInternalLint) {
			os.Exit(2)
		}
		if !errors.Is(err, cliutil.ErrNonZeroExit) {
			fmt.Fprintf(os.Stderr, "mtrace: %v\n", err)
		}
		os.Exit(1)
	}
}
