package main

import (
	"bytes"
	"testing"

	"github.com/dpopsuev/mos/cmd/mos/govern"
	"github.com/spf13/cobra"
)

func executeCmd(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err := root.Execute()
	return buf.String(), err
}

func TestQueryCmdRequiresNoArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.AddCommand(govern.QueryCmd)
	_, err := executeCmd(cmd, "query")
	if err != nil {
		t.Fatalf("query with no args should not error (returns empty results): %v", err)
	}
}

func TestUpdateCmdRequiresStdin(t *testing.T) {
	cmd := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	cmd.AddCommand(govern.UpdateCmd)
	_, err := executeCmd(cmd, "update")
	if err == nil {
		t.Fatal("expected error when --stdin not provided")
	}
}

func TestGetCmdRequiresArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	cmd.AddCommand(govern.GetCmd)
	_, err := executeCmd(cmd, "get")
	if err == nil {
		t.Fatal("expected error when no args provided")
	}
}

func TestSetCmdRequiresArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	cmd.AddCommand(govern.SetCmd)
	_, err := executeCmd(cmd, "set", "ID")
	if err == nil {
		t.Fatal("expected error when too few args")
	}
}

func TestAppendCmdRequiresArgs(t *testing.T) {
	cmd := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true}
	cmd.AddCommand(govern.AppendCmd)
	_, err := executeCmd(cmd, "append", "ID", "path")
	if err == nil {
		t.Fatal("expected error when too few args")
	}
}
