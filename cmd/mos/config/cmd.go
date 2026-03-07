package config

import (
	"fmt"
	"strings"

	"github.com/dpopsuev/mos/moslib/registry"
	"github.com/spf13/cobra"
)

const rootDir = "."

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage project configuration in .mos/config.mos",
	Long: `Manage project configuration in .mos/config.mos.

Sub-commands:
  add-project      Add a governed project
  remove-project   Remove a governed project
  add-type         Register a custom artifact type
  remove-type      Remove a custom artifact type
  add-field        Add a field to an artifact type
  set-field-enum   Set enum values for a field
  set-lifecycle    Configure lifecycle statuses for an artifact type
  set-directory    Set the directory for an artifact type

Examples:
  mos config add-project myapp --prefix APP
  mos config add-type epic --directory epics
  mos config set-lifecycle epic --active draft,active --archive complete,abandoned`,
}

func init() {
	Cmd.AddCommand(addProjectCmd, removeProjectCmd, addTypeCmd, removeTypeCmd,
		addFieldCmd, setLifecycleCmd, setDirectoryCmd, setFieldEnumCmd)
}

// add-project
var addProjectCmd = &cobra.Command{
	Use:   "add-project",
	Short: "Add a governed project",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddProject,
}

var addProjectPrefix string

func init() {
	addProjectCmd.Flags().StringVar(&addProjectPrefix, "prefix", "", "Project prefix (required)")
	_ = addProjectCmd.MarkFlagRequired("prefix")
}

func runAddProject(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := registry.AddProject(rootDir, name, addProjectPrefix); err != nil {
		return fmt.Errorf("mos config add-project: %w", err)
	}
	fmt.Printf("Added project %q with prefix %q\n", name, addProjectPrefix)
	return nil
}

// remove-project
var removeProjectCmd = &cobra.Command{
	Use:   "remove-project",
	Short: "Remove a governed project",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemoveProject,
}

func runRemoveProject(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := registry.RemoveProject(rootDir, name); err != nil {
		return fmt.Errorf("mos config remove-project: %w", err)
	}
	fmt.Printf("Removed project %q\n", name)
	return nil
}

// add-type
var addTypeCmd = &cobra.Command{
	Use:   "add-type",
	Short: "Register a custom artifact type",
	Args:  cobra.ExactArgs(1),
	RunE:  runAddType,
}

var addTypeDirectory string

func init() {
	addTypeCmd.Flags().StringVar(&addTypeDirectory, "directory", "", "Directory for artifact type")
}

func runAddType(cmd *cobra.Command, args []string) error {
	kind := args[0]
	if err := registry.AddArtifactType(rootDir, kind, addTypeDirectory); err != nil {
		return fmt.Errorf("mos config add-type: %w", err)
	}
	reg, _ := registry.LoadRegistry(rootDir)
	directory := addTypeDirectory
	if td, ok := reg.Types[kind]; ok {
		directory = td.Directory
	}
	fmt.Printf("Added artifact_type %q (directory: %s)\n", kind, directory)
	return nil
}

// remove-type
var removeTypeCmd = &cobra.Command{
	Use:   "remove-type",
	Short: "Remove a custom artifact type",
	Args:  cobra.ExactArgs(1),
	RunE:  runRemoveType,
}

func runRemoveType(cmd *cobra.Command, args []string) error {
	kind := args[0]
	if err := registry.RemoveArtifactType(rootDir, kind); err != nil {
		return fmt.Errorf("mos config remove-type: %w", err)
	}
	fmt.Printf("Removed artifact_type %q\n", kind)
	return nil
}

// add-field
var addFieldCmd = &cobra.Command{
	Use:   "add-field",
	Short: "Add a field to an artifact type definition",
	Long: `Add a field to an artifact type definition.

Flags:
  --required            Mark the field as required
  --enum <val1,val2>    Set allowed enum values for the field

Examples:
  mos config add-field contract priority --enum low,medium,high
  mos config add-field epic effort --required`,
	Args: cobra.ExactArgs(2),
	RunE: runAddField,
}

var addFieldRequired bool
var addFieldEnum string

func init() {
	addFieldCmd.Flags().BoolVar(&addFieldRequired, "required", false, "Mark the field as required")
	addFieldCmd.Flags().StringVar(&addFieldEnum, "enum", "", "Comma-separated enum values")
}

func runAddField(cmd *cobra.Command, args []string) error {
	kind, fieldName := args[0], args[1]
	var enum []string
	if addFieldEnum != "" {
		enum = strings.Split(addFieldEnum, ",")
	}
	opts := registry.FieldOpts{Name: fieldName, Required: addFieldRequired, Enum: enum}
	if err := registry.AddFieldToType(rootDir, kind, opts); err != nil {
		return fmt.Errorf("mos config add-field: %w", err)
	}
	fmt.Printf("Added field %q to artifact_type %q\n", fieldName, kind)
	return nil
}

// set-lifecycle
var setLifecycleCmd = &cobra.Command{
	Use:   "set-lifecycle",
	Short: "Configure lifecycle statuses for an artifact type",
	Long: `Configure lifecycle statuses for an artifact type. Artifacts in active
statuses are stored in the active directory; archive statuses move them
to the archive directory.

Flags:
  --active <s1,s2>    Comma-separated active statuses (required)
  --archive <s3,s4>   Comma-separated archive statuses (required)

Examples:
  mos config set-lifecycle contract --active draft,active --archive complete,abandoned
  mos config set-lifecycle epic --active planned,in-progress --archive done`,
	Args: cobra.ExactArgs(1),
	RunE: runSetLifecycle,
}

var setLifecycleActive string
var setLifecycleArchive string

func init() {
	setLifecycleCmd.Flags().StringVar(&setLifecycleActive, "active", "", "Comma-separated active statuses (required)")
	setLifecycleCmd.Flags().StringVar(&setLifecycleArchive, "archive", "", "Comma-separated archive statuses (required)")
	_ = setLifecycleCmd.MarkFlagRequired("active")
	_ = setLifecycleCmd.MarkFlagRequired("archive")
}

func runSetLifecycle(cmd *cobra.Command, args []string) error {
	kind := args[0]
	active := strings.Split(setLifecycleActive, ",")
	archive := strings.Split(setLifecycleArchive, ",")
	if err := registry.SetTypeLifecycle(rootDir, kind, active, archive); err != nil {
		return fmt.Errorf("mos config set-lifecycle: %w", err)
	}
	fmt.Printf("Updated lifecycle for artifact_type %q\n", kind)
	return nil
}

// set-directory
var setDirectoryCmd = &cobra.Command{
	Use:   "set-directory",
	Short: "Set the on-disk directory for an artifact type",
	Long: `Set the on-disk directory for an artifact type.

Examples:
  mos config set-directory epic epics
  mos config set-directory need needs`,
	Args: cobra.ExactArgs(2),
	RunE: runSetDirectory,
}

func runSetDirectory(cmd *cobra.Command, args []string) error {
	kind, directory := args[0], args[1]
	if err := registry.SetTypeDirectory(rootDir, kind, directory); err != nil {
		return fmt.Errorf("mos config set-directory: %w", err)
	}
	fmt.Printf("Set directory for artifact_type %q to %q\n", kind, directory)
	return nil
}

// set-field-enum
var setFieldEnumCmd = &cobra.Command{
	Use:   "set-field-enum",
	Short: "Set the allowed enum values for a field on an artifact type",
	Long: `Set the allowed enum values for a field on an artifact type.

Examples:
  mos config set-field-enum contract priority low,medium,high,critical`,
	Args: cobra.ExactArgs(3),
	RunE: runSetFieldEnum,
}

func runSetFieldEnum(cmd *cobra.Command, args []string) error {
	kind, fieldName := args[0], args[1]
	enum := strings.Split(args[2], ",")
	if err := registry.SetFieldEnum(rootDir, kind, fieldName, enum); err != nil {
		return fmt.Errorf("mos config set-field-enum: %w", err)
	}
	fmt.Printf("Set enum for field %q in artifact_type %q to %v\n", fieldName, kind, enum)
	return nil
}
