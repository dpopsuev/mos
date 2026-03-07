package govern

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dpopsuev/mos/moslib/artifact"
)

var ArchiveCmd = &cobra.Command{
	Use:   "archive [ID...]",
	Short: "Move artifacts from active/ to archive/",
	Long:  "Move one or more artifacts from active/ to archive/ by ID, or use --all-terminal to relocate all artifacts with terminal status.",
	RunE:  runArchive,
}

var archiveAllTerminal bool

func init() {
	ArchiveCmd.Flags().BoolVar(&archiveAllTerminal, "all-terminal", false, "Move all artifacts whose status is terminal to archive/")
}

func runArchive(cmd *cobra.Command, args []string) error {
	root, _ := os.Getwd()

	if archiveAllTerminal {
		relocations, err := artifact.RelocateMisplacedArtifacts(root)
		if err != nil {
			return fmt.Errorf("archive --all-terminal: %w", err)
		}
		for _, r := range relocations {
			fmt.Printf("archived %s (%s)\n", r.ID, r.Kind)
		}
		if len(relocations) == 0 {
			fmt.Println("no misplaced artifacts found")
		}
		return nil
	}

	if len(args) == 0 {
		return fmt.Errorf("provide at least one artifact ID, or use --all-terminal")
	}

	reg, err := artifact.LoadRegistry(root)
	if err != nil {
		return fmt.Errorf("archive: %w", err)
	}

	var errs []string
	for _, id := range args {
		moved := false
		for _, td := range reg.Types {
			activeDir := filepath.Join(root, artifact.MosDir, td.Directory, artifact.ActiveDir, id)
			if _, err := os.Stat(activeDir); err != nil {
				continue
			}
			archiveDir := filepath.Join(root, artifact.MosDir, td.Directory, artifact.ArchiveDir)
			if err := os.MkdirAll(archiveDir, 0o755); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", id, err))
				break
			}
			dest := filepath.Join(archiveDir, id)
			if err := os.Rename(activeDir, dest); err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", id, err))
				break
			}
			fmt.Printf("archived %s (%s)\n", id, td.Kind)
			moved = true
			break
		}
		if !moved && !containsErr(errs, id) {
			errs = append(errs, fmt.Sprintf("%s: not found in any active/ directory", id))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("archive errors:\n  %s", strings.Join(errs, "\n  "))
	}
	return nil
}

func containsErr(errs []string, id string) bool {
	for _, e := range errs {
		if strings.HasPrefix(e, id+":") {
			return true
		}
	}
	return false
}
