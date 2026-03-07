package artifact

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dpopsuev/mos/moslib/dsl"
)

// InitOpts configures the mos init command.
type InitOpts struct {
	Model   string // governance.model (default: "bdfl")
	Scope   string // governance.scope (default: "cabinet")
	Name    string // project name (auto-detect from go.mod if empty)
	Purpose string // declaration description
}

// Init creates a .mos/ scaffold under root with config, lexicon,
// layers, declaration, and empty rules/contracts directories.
func Init(root string, opts InitOpts) error {
	mosDir := filepath.Join(root, MosDir)
	if _, err := os.Stat(mosDir); err == nil {
		return fmt.Errorf(".mos/ directory already exists at %s", mosDir)
	}

	if opts.Model == "" {
		opts.Model = "bdfl"
	}
	if opts.Scope == "" {
		opts.Scope = "cabinet"
	}
	if opts.Name == "" {
		opts.Name = detectProjectName(root)
	}
	if opts.Name == "" {
		return fmt.Errorf("could not detect project name; use --name")
	}

	dirs := []string{
		mosDir,
		filepath.Join(mosDir, "lexicon"),
		filepath.Join(mosDir, "resolution"),
		filepath.Join(mosDir, "rules", "mechanical"),
		filepath.Join(mosDir, "rules", "interpretive"),
		filepath.Join(mosDir, "contracts", ActiveDir),
		filepath.Join(mosDir, "contracts", ArchiveDir),
		filepath.Join(mosDir, "specifications", ActiveDir),
		filepath.Join(mosDir, "specifications", ArchiveDir),
		filepath.Join(mosDir, "binders", "active"),
		filepath.Join(mosDir, "binders", "archive"),
		filepath.Join(mosDir, "needs", "active"),
		filepath.Join(mosDir, "needs", "archive"),
		filepath.Join(mosDir, "architectures", "active"),
		filepath.Join(mosDir, "architectures", "archive"),
		filepath.Join(mosDir, "docs", "active"),
		filepath.Join(mosDir, "docs", "archive"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, DirPerm); err != nil {
			return fmt.Errorf("creating directory %s: %w", d, err)
		}
	}

	if err := writeConfig(mosDir, opts); err != nil {
		return fmt.Errorf("writing config.mos: %w", err)
	}
	if err := writeLexicon(mosDir); err != nil {
		return fmt.Errorf("writing lexicon: %w", err)
	}
	if err := writeLayers(mosDir); err != nil {
		return fmt.Errorf("writing layers: %w", err)
	}
	if err := writeDeclaration(mosDir, opts); err != nil {
		return fmt.Errorf("writing declaration: %w", err)
	}

	return nil
}

func writeConfig(mosDir string, opts InitOpts) error {
	items := []dsl.Node{
		&dsl.Block{
			Name: "mos",
			Items: []dsl.Node{
				&dsl.Field{Key: "version", Value: &dsl.IntegerVal{Raw: "1", Val: 1}},
			},
		},
		&dsl.Block{
			Name: "backend",
			Items: []dsl.Node{
				&dsl.Field{Key: "type", Value: &dsl.StringVal{Text: "git"}},
			},
		},
		&dsl.Block{
			Name: "governance",
			Items: []dsl.Node{
				&dsl.Field{Key: "model", Value: &dsl.StringVal{Text: opts.Model}},
				&dsl.Field{Key: "scope", Value: &dsl.StringVal{Text: opts.Scope}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "contracts",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "CON"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
				&dsl.Field{Key: "default", Value: &dsl.BoolVal{Val: true}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "bugs",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "BUG"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "specifications",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "SPEC"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "binders",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "BND"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "needs",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "NEED"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "architectures",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "ARCH"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
		&dsl.Block{
			Name:  "project",
			Title: "docs",
			Items: []dsl.Node{
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "DOC"}},
				&dsl.Field{Key: "sequence", Value: &dsl.IntegerVal{Raw: "0", Val: 0}},
			},
		},
	}

	items = append(items, defaultArtifactTypeBlocks()...)

	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  "config",
			Items: items,
		},
	}
	return writeArtifact(filepath.Join(mosDir, ConfigFile), file)
}

// defaultArtifactTypeBlocks returns the shipped CAD (Custom Artifact Definition)
// blocks for all core artifact types.
func defaultArtifactTypeBlocks() []dsl.Node {
	return []dsl.Node{
		&dsl.Block{
			Name:  "artifact_type",
			Title: "contract",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "contracts"}},
				&dsl.Field{Key: "ledger", Value: &dsl.BoolVal{Val: true}},
			&dsl.Block{
				Name: "fields",
				Items: []dsl.Node{
					&dsl.Block{Name: "title", Items: []dsl.Node{
						&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
					}},
					&dsl.Block{Name: "status", Items: []dsl.Node{
						&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
					}},
					&dsl.Block{Name: "justifies", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "specification"}},
					}},
					&dsl.Block{Name: "implements", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
					}},
					&dsl.Block{Name: "documents", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
					}},
					&dsl.Block{Name: "sprint", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "sprint"}},
					}},
					&dsl.Block{Name: "batch", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "batch"}},
					}},
					&dsl.Block{Name: "parent", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "contract"}},
					}},
					&dsl.Block{Name: "depends_on", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "contract"}},
					}},
				},
			},
			&dsl.Block{
				Name: "scenario_fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "status", Items: []dsl.Node{
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "pending"},
								&dsl.StringVal{Text: "implemented"},
								&dsl.StringVal{Text: "verified"},
							}}},
							&dsl.Field{Key: "default", Value: &dsl.StringVal{Text: "pending"}},
							&dsl.Field{Key: "ordered", Value: &dsl.BoolVal{Val: true}},
							&dsl.Block{Name: "transitions", Items: []dsl.Node{
								&dsl.Block{Name: "implemented", Items: []dsl.Node{
									&dsl.Field{Key: "to", Value: &dsl.StringVal{Text: "verified"}},
									&dsl.Field{Key: "verified_by", Value: &dsl.StringVal{Text: "harness"}},
								}},
							}},
						}},
					},
				},
				&dsl.Block{
					Name: "lifecycle",
					Items: []dsl.Node{
						&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "draft"},
							&dsl.StringVal{Text: "active"},
						}}},
						&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "complete"},
							&dsl.StringVal{Text: "abandoned"},
						}}},
						&dsl.Block{Name: "hooks", Items: []dsl.Node{
							&dsl.Block{Name: "on_any", Items: []dsl.Node{
								&dsl.Field{Key: "watch_field", Value: &dsl.StringVal{Text: "status"}},
								&dsl.Field{Key: "threshold", Value: &dsl.StringVal{Text: "implemented"}},
								&dsl.Field{Key: "set_field", Value: &dsl.StringVal{Text: "status"}},
								&dsl.Field{Key: "set_value", Value: &dsl.StringVal{Text: "active"}},
							}},
							&dsl.Block{Name: "on_all", Items: []dsl.Node{
								&dsl.Field{Key: "watch_field", Value: &dsl.StringVal{Text: "status"}},
								&dsl.Field{Key: "threshold", Value: &dsl.StringVal{Text: "verified"}},
								&dsl.Field{Key: "set_field", Value: &dsl.StringVal{Text: "status"}},
								&dsl.Field{Key: "set_value", Value: &dsl.StringVal{Text: "complete"}},
							}},
						}},
					},
				},
			},
		},
		&dsl.Block{
			Name:  "artifact_type",
			Title: "rule",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "rules"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "name", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "type", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "scope", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "enforcement", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
					},
				},
			},
		},
		&dsl.Block{
			Name:  "artifact_type",
			Title: "specification",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "specifications"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "enforcement", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "disabled"},
								&dsl.StringVal{Text: "warn"},
								&dsl.StringVal{Text: "enforced"},
							}}},
						}},
					&dsl.Block{Name: "non_goals"},
					&dsl.Block{Name: "satisfies", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "need"}},
					}},
					&dsl.Block{Name: "addresses", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "need"}},
					}},
				&dsl.Block{Name: "group"},
					&dsl.Block{Name: "verification_method", Items: []dsl.Node{
						&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "analysis"},
							&dsl.StringVal{Text: "demonstration"},
							&dsl.StringVal{Text: "inspection"},
							&dsl.StringVal{Text: "test"},
						}}},
					}},
			},
		},
		&dsl.Block{
			Name: "lifecycle",
			Items: []dsl.Node{
				&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
					&dsl.StringVal{Text: "candidate"},
					&dsl.StringVal{Text: "active"},
				}}},
				&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
					&dsl.StringVal{Text: "retired"},
					}}},
					&dsl.Block{Name: "expects_downstream", Items: []dsl.Node{
						&dsl.Field{Key: "via", Value: &dsl.StringVal{Text: "implements"}},
						&dsl.Field{Key: "after", Value: &dsl.StringVal{Text: "active"}},
						&dsl.Field{Key: "severity", Value: &dsl.StringVal{Text: "warn"}},
					}},
				},
			},
		},
	},
	&dsl.Block{
		Name:  "artifact_type",
		Title: "binder",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "binders"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
					},
				},
				&dsl.Block{
					Name: "lifecycle",
					Items: []dsl.Node{
						&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "active"},
						}}},
						&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "archived"},
						}}},
					},
				},
			},
		},
		&dsl.Block{
			Name:  "artifact_type",
			Title: "need",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "needs"}},
				&dsl.Field{Key: "ledger", Value: &dsl.BoolVal{Val: true}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "sensation", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "urgency", Items: []dsl.Node{
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "critical"},
								&dsl.StringVal{Text: "high"},
								&dsl.StringVal{Text: "medium"},
								&dsl.StringVal{Text: "low"},
							}}},
						}},
					&dsl.Block{Name: "stakeholders"},
					&dsl.Block{Name: "status", Items: []dsl.Node{
						&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "identified"},
							&dsl.StringVal{Text: "validated"},
							&dsl.StringVal{Text: "addressed"},
							&dsl.StringVal{Text: "retired"},
						}}},
					}},
				&dsl.Block{Name: "acceptance"},
				&dsl.Block{Name: "originating"},
				&dsl.Block{Name: "derives_from", Items: []dsl.Node{
					&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
					&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "need"}},
				}},
			},
		},
		&dsl.Block{
			Name: "lifecycle",
				Items: []dsl.Node{
					&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "identified"},
						&dsl.StringVal{Text: "validated"},
						&dsl.StringVal{Text: "addressed"},
					}}},
					&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "retired"},
					}}},
					&dsl.Block{Name: "expects_downstream", Items: []dsl.Node{
						&dsl.Field{Key: "via", Value: &dsl.StringVal{Text: "satisfies"}},
						&dsl.Field{Key: "after", Value: &dsl.StringVal{Text: "validated"}},
						&dsl.Field{Key: "severity", Value: &dsl.StringVal{Text: "warn"}},
					}},
					&dsl.Block{Name: "transition", Title: "validated", Items: []dsl.Node{
						&dsl.Field{Key: "to", Value: &dsl.StringVal{Text: "addressed"}},
						&dsl.Field{Key: "gate", Value: &dsl.StringVal{Text: "criteria_coverage"}},
					}},
					&dsl.Block{Name: "urgency_propagation", Items: []dsl.Node{
						&dsl.Field{Key: "critical", Value: &dsl.StringVal{Text: "error"}},
						&dsl.Field{Key: "high", Value: &dsl.StringVal{Text: "warn"}},
						&dsl.Field{Key: "medium", Value: &dsl.StringVal{Text: "info"}},
						&dsl.Field{Key: "low", Value: &dsl.StringVal{Text: "ignore"}},
					}},
				},
			},
		},
	},
	&dsl.Block{
		Name:  "artifact_type",
		Title: "architecture",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "architectures"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "resolution", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "service"},
								&dsl.StringVal{Text: "component"},
							}}},
						}},
						&dsl.Block{Name: "status", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
					&dsl.Block{Name: "implements", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
					}},
				},
			},
			&dsl.Block{
				Name: "lifecycle",
				Items: []dsl.Node{
					&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "draft"},
						&dsl.StringVal{Text: "active"},
					}}},
					&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "superseded"},
					}}},
					&dsl.Block{Name: "expects_downstream", Items: []dsl.Node{
						&dsl.Field{Key: "via", Value: &dsl.StringVal{Text: "implements"}},
						&dsl.Field{Key: "after", Value: &dsl.StringVal{Text: "active"}},
						&dsl.Field{Key: "severity", Value: &dsl.StringVal{Text: "warn"}},
					}},
				},
			},
		},
	},
	&dsl.Block{
		Name:  "artifact_type",
		Title: "doc",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "docs"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "kind", Items: []dsl.Node{
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "api-reference"},
								&dsl.StringVal{Text: "runbook"},
								&dsl.StringVal{Text: "adr"},
								&dsl.StringVal{Text: "onboarding"},
								&dsl.StringVal{Text: "generated"},
							}}},
						}},
						&dsl.Block{Name: "status", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
					&dsl.Block{Name: "documents", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
					}},
					&dsl.Block{Name: "source"},
						&dsl.Block{Name: "generated_by"},
					},
				},
				&dsl.Block{
					Name: "lifecycle",
					Items: []dsl.Node{
						&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "draft"},
							&dsl.StringVal{Text: "published"},
							&dsl.StringVal{Text: "stale"},
						}}},
						&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
							&dsl.StringVal{Text: "retired"},
						}}},
					},
				},
			},
		},
		&dsl.Block{
			Name:  "artifact_type",
			Title: "sprint",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "sprints"}},
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "SPR"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "status", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "planned"},
								&dsl.StringVal{Text: "active"},
								&dsl.StringVal{Text: "complete"},
								&dsl.StringVal{Text: "deferred"},
							}}},
						}},
						&dsl.Block{Name: "goal"},
						&dsl.Block{Name: "slug"},
					&dsl.Block{Name: "contracts", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "contract"}},
					}},
				},
			},
			&dsl.Block{
				Name: "lifecycle",
				Items: []dsl.Node{
					&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "planned"},
						&dsl.StringVal{Text: "active"},
						&dsl.StringVal{Text: "deferred"},
					}}},
					&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "complete"},
					}}},
				},
			},
		},
	},
	&dsl.Block{
		Name:  "artifact_type",
		Title: "batch",
			Items: []dsl.Node{
				&dsl.Field{Key: "directory", Value: &dsl.StringVal{Text: "batches"}},
				&dsl.Field{Key: "prefix", Value: &dsl.StringVal{Text: "BAT"}},
				&dsl.Block{
					Name: "fields",
					Items: []dsl.Node{
						&dsl.Block{Name: "title", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
						}},
						&dsl.Block{Name: "status", Items: []dsl.Node{
							&dsl.Field{Key: "required", Value: &dsl.BoolVal{Val: true}},
							&dsl.Field{Key: "enum", Value: &dsl.ListVal{Items: []dsl.Value{
								&dsl.StringVal{Text: "draft"},
								&dsl.StringVal{Text: "active"},
								&dsl.StringVal{Text: "complete"},
							}}},
						}},
					&dsl.Block{Name: "contracts", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "contract"}},
					}},
					&dsl.Block{Name: "depends_on", Items: []dsl.Node{
						&dsl.Field{Key: "link", Value: &dsl.BoolVal{Val: true}},
						&dsl.Field{Key: "ref_kind", Value: &dsl.StringVal{Text: "batch"}},
					}},
				},
			},
			&dsl.Block{
				Name: "lifecycle",
				Items: []dsl.Node{
					&dsl.Field{Key: "active_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "draft"},
						&dsl.StringVal{Text: "active"},
					}}},
					&dsl.Field{Key: "archive_states", Value: &dsl.ListVal{Items: []dsl.Value{
						&dsl.StringVal{Text: "complete"},
					}}},
				},
			},
		},
	},
	}
}

func writeLexicon(mosDir string) error {
	content := `lexicon {

  terms {
    # Common ALM terms -- uncomment and customize for your project.
    # These are open fields, not schema keywords. The lexicon
    # controls what the linter considers valid.
    # pillar = "A test category (acceptance, regression, performance, ...)"
    # component = "A product component for scoping rules and contracts"
    # priority = "Urgency level (critical, high, medium, low)"
    # automation_status = "Test automation state (automated, manual, planned)"
  }
}
`
	return os.WriteFile(filepath.Join(mosDir, "lexicon", "default.mos"), []byte(content), FilePerm)
}

func writeLayers(mosDir string) error {
	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind: "layers",
			Items: []dsl.Node{
				&dsl.Block{
					Name:  "layer",
					Title: "project",
					Items: []dsl.Node{
						&dsl.Field{Key: "level", Value: &dsl.IntegerVal{Raw: "1", Val: 1}},
					},
				},
			},
		},
	}
	return writeArtifact(filepath.Join(mosDir, "resolution", "layers.mos"), file)
}

func writeDeclaration(mosDir string, opts InitOpts) error {
	items := []dsl.Node{
		&dsl.Field{Key: "name", Value: &dsl.StringVal{Text: opts.Name}},
		&dsl.Field{Key: "created", Value: &dsl.DateTimeVal{Raw: time.Now().UTC().Format(time.RFC3339)}},
	}
	if opts.Purpose != "" {
		items = append(items, &dsl.Field{Key: "description", Value: &dsl.StringVal{Text: opts.Purpose}})
	}

	file := &dsl.File{
		Artifact: &dsl.ArtifactBlock{
			Kind:  "declaration",
			Items: items,
		},
	}
	return writeArtifact(filepath.Join(mosDir, "declaration.mos"), file)
}

func writeArtifact(path string, file *dsl.File) error {
	content := dsl.Format(file, nil)
	return atomicWriteFile(path, []byte(content), FilePerm)
}

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".mos-tmp-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}

func detectProjectName(root string) string {
	if name := detectProjectNameFromGoMod(root); name != "" {
		return name
	}
	if name := detectProjectNameFromCargoToml(root); name != "" {
		return name
	}
	return filepath.Base(root)
}

func detectProjectNameFromGoMod(root string) string {
	goMod := filepath.Join(root, "go.mod")
	f, err := os.Open(goMod)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "module ") {
			mod := strings.TrimPrefix(line, "module ")
			mod = strings.TrimSpace(mod)
			parts := strings.Split(mod, "/")
			return parts[len(parts)-1]
		}
	}
	return ""
}

func detectProjectNameFromCargoToml(root string) string {
	data, err := os.ReadFile(filepath.Join(root, "Cargo.toml"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				name := strings.TrimSpace(parts[1])
				name = strings.Trim(name, "\"'")
				if name != "" {
					return name
				}
			}
		}
	}
	return ""
}
