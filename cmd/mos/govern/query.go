package govern

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dpopsuev/mos/moslib/artifact"
	"github.com/dpopsuev/mos/moslib/names"
	"github.com/spf13/cobra"
)

var QueryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query artifact data",
	Long: `Query artifacts across all kinds.

Examples:
  mos query --kind contract --status draft
  mos query --kind specification --where satisfies=NEED-2026-001
  mos query --kind need -v
  mos query --kind contract --count --group-by status --format json`,
	RunE: runQuery,
}

var (
	queryKind       string
	queryStatus     string
	queryLabels     string
	queryReferences string
	queryCount      bool
	queryGroupBy    string
	queryFormat     string
	queryRich       bool
	queryVerbose    bool
	queryJustifies  string
	querySprint     string
	queryDependsOn  string
	queryPriority   string
	queryUnlinked   bool
	queryState      string
	queryGroup      string
	queryWhere      []string
)

func init() {
	QueryCmd.Flags().StringVar(&queryKind, "kind", "", "Filter by artifact kind")
	QueryCmd.Flags().StringVar(&queryStatus, "status", "", "Filter by status")
	QueryCmd.Flags().StringVar(&queryLabels, "labels", "", "Filter by comma-separated labels")
	QueryCmd.Flags().StringVar(&queryReferences, "references", "", "Filter artifacts referencing a given ID")
	QueryCmd.Flags().BoolVar(&queryCount, "count", false, "Show count instead of listing")
	QueryCmd.Flags().StringVar(&queryGroupBy, "group-by", "", "Group counts by field (implies --count)")
	QueryCmd.Flags().StringVar(&queryFormat, "format", "text", "Output format: text or json")
	QueryCmd.Flags().BoolVar(&queryRich, "rich", false, "Return full artifact content in JSON output")
	QueryCmd.Flags().BoolVarP(&queryVerbose, "verbose", "v", false, "Show key metadata fields in text output")
	QueryCmd.Flags().StringArrayVar(&queryWhere, "where", nil, "Filter by field=value (repeatable, AND logic)")
	QueryCmd.Flags().StringVar(&queryJustifies, "justifies", "", "Filter by justifies field")
	QueryCmd.Flags().StringVar(&querySprint, "sprint", "", "Filter by sprint assignment")
	QueryCmd.Flags().StringVar(&queryDependsOn, "depends-on", "", "Filter by depends_on field")
	QueryCmd.Flags().StringVar(&queryPriority, "priority", "", "Filter by priority field")
	QueryCmd.Flags().BoolVar(&queryUnlinked, "unlinked", false, "Only show artifacts with no sprint assignment")
	QueryCmd.Flags().StringVar(&queryState, "state", "", "Filter by state field (current, desired, both)")
	QueryCmd.Flags().StringVar(&queryGroup, "group", "", "Filter by group field")
}

func runQuery(cmd *cobra.Command, args []string) error {
	opts := artifact.QueryOpts{
		Kind:       queryKind,
		Status:     queryStatus,
		References: queryReferences,
		Count:      queryCount,
		GroupBy:    queryGroupBy,
		Format:     queryFormat,
		Rich:       queryRich,
		Verbose:    queryVerbose,
		Justifies:  queryJustifies,
		SprintEq:   querySprint,
		DependsOn:  queryDependsOn,
		Priority:   queryPriority,
		Unlinked:   queryUnlinked,
		State:      queryState,
		Group:      queryGroup,
		Where:      queryWhere,
	}
	if queryLabels != "" {
		opts.Labels = strings.Split(queryLabels, ",")
	}
	if opts.GroupBy != "" {
		opts.Count = true
	}

	results, err := artifact.QueryArtifacts(".", opts)
	if err != nil {
		return err
	}

	if opts.Format == names.FormatJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		switch {
		case opts.Count && opts.GroupBy != "":
			enc.Encode(artifact.GroupResults(results, opts.GroupBy))
		case opts.Count:
			enc.Encode(map[string]int{"count": len(results)})
		case opts.Rich:
			enc.Encode(artifact.RichQueryResults(results))
		default:
			enc.Encode(results)
		}
		return nil
	}

	fmt.Print(artifact.FormatQueryResults(results, opts))
	return nil
}

var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Bulk-update artifacts",
	Long: `Bulk-update multiple artifacts read from stdin (one ID per line).

Examples:
  echo "CON-2026-101" | mos update --stdin --status active
  mos query --kind contract --status draft --format json | jq -r '.[].id' | mos update --stdin --sprint SPR-2026-004`,
	RunE: runBulkUpdate,
}

var (
	updateStdin     bool
	updateStatus    string
	updateSprint    string
	updateJustifies string
)

func init() {
	UpdateCmd.Flags().BoolVar(&updateStdin, "stdin", false, "Read artifact IDs from stdin (required)")
	UpdateCmd.Flags().StringVar(&updateStatus, "status", "", "Set status on all listed artifacts")
	UpdateCmd.Flags().StringVar(&updateSprint, "sprint", "", "Assign sprint (comma-separated)")
	UpdateCmd.Flags().StringVar(&updateJustifies, "justifies", "", "Set justifies field (comma-separated)")
}

func runBulkUpdate(cmd *cobra.Command, args []string) error {
	if !updateStdin {
		return fmt.Errorf("--stdin flag is required")
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}

	var status *string
	var sprint, justifies *[]string

	if updateStatus != "" {
		status = &updateStatus
	}
	if updateSprint != "" {
		vals := strings.Split(updateSprint, ",")
		sprint = &vals
	}
	if updateJustifies != "" {
		vals := strings.Split(updateJustifies, ",")
		justifies = &vals
	}

	ids := strings.Fields(strings.TrimSpace(string(data)))
	errCount := 0
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		updateOpts := artifact.ContractUpdateOpts{
			Status:    status,
			Sprint:    sprint,
			Justifies: justifies,
		}
		if err := artifact.UpdateContract(".", id, updateOpts); err != nil {
			fmt.Fprintf(os.Stderr, "mos update %s: %v\n", id, err)
			errCount++
			continue
		}
		fmt.Printf("Updated %s\n", id)
	}
	if errCount > 0 {
		return fmt.Errorf("%d update(s) failed", errCount)
	}
	return nil
}
