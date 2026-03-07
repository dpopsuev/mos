# Mos DSL -- Desired Format Draft

This directory contains the draft specification for the Mos DSL (v3).
Each `.mos` file demonstrates a specific artifact type or composition pattern.
The `grammar.ebnf` file defines the formal grammar.

**Status:** The parser in `moslib/dsl/` implements the v3 grammar plus
vocabulary-driven keyword localization (CON-2026-007). These examples serve as
the canonical reference for DSL syntax and artifact patterns.

## Design Principles

- **Primitives, not prescriptions.** The grammar provides composable building
  blocks. What users build with them is emergent.
- **One format policy.** One canonical representation. `mos fmt` decides.
- **Inertness.** Values are data, never programs. No expressions, no variables,
  no conditionals.
- **Freedom of action.** Users define their own vocabulary, labels, and layers.
  The DSL provides structure, not content.

## Grammar Summary

~16 keywords: `rule`, `contract`, `config`, `declaration`, `vocabulary`,
`layers`, `layer`, `spec`, `include`, `feature`, `background`, `scenario`,
`group`, `given`, `when`, `then`.

See [grammar.ebnf](grammar.ebnf) for the full formal grammar.

### Key syntax features

- **Open artifact types:** any identifier is a valid artifact kind — `rule`,
  `guideline`, `settings`, etc. Named: `type "id" { }`, unnamed: `type { }`
- **All keywords lowercase**, all blocks brace-delimited
- **Native datetime literals:** `2024-06-15T00:00:00Z` (bare, not quoted)
- **Triple-quoted strings:** `"""..."""` for multi-line text
- **Inline tables:** `{ key = "val", key2 = "val2" }`
- **Trailing commas** allowed in lists and inline tables
- **Comments:** `# line comment`

### Upstream federation (recursive, not tier-locked)

Moss form a recursive upstream chain, mirroring how
[real-world legal systems](https://en.wikipedia.org/wiki/Law_of_the_United_States)
cascade from federal to state to local -- but **without hardcoding the
number of tiers**. A mos with no `upstream` block is the **domain
root** (the ultimate authority). Every other repo declares one or more
`upstream` blocks, forming a DAG of arbitrary depth:

```
0  CEO/President (no upstream)                     ← domain root
  └─ 1  CTO (upstream → CEO)                      ← scope: "c-suite"
       └─ 2  VP Engineering (upstream → CTO)       ← scope: "vp"
            └─ 3  Division Director (upstream → VP) ← scope: "division"
                 └─ 4  QE Director (upstream → Div) ← scope: "department"
                      └─ 5  QE Manager (upstream → Dir)  ← scope: "collective"
                           └─ 6  Engineer (upstream → Mgr) ← scope: "cabinet"
```

- **Domain root**: `config { ... }` with no `upstream` block. The source of
  all authority. In the example above: CEO/President.
- **Downstream repos**: declare `upstream "name" { url, ref, scope }`. The
  `scope` field is a **user-defined label** (free string) for human
  readability -- it is NOT a structural tier. Depth is derived from the
  chain, not from the label.
- **Cabinet (individual)**: the leaf node. An individual contributor's
  personal governance context. Inherits all ancestor rules.
- **Collective (group)**: a team's shared governance. Manager curates rules
  inherited from above.
- **Rule inheritance**: rules carry `jurisdiction` (free label) and
  optionally `extends = "R-XXX"` to specialize an ancestor rule. The
  linter resolves the upstream chain to validate that downstream rules
  don't contradict ancestors.
- **Multiple upstreams**: cross-org federation via DAG (not just tree).
- **Bootstrap**: `mos init --upstream <url>` creates a new repo inheriting
  rules, vocabulary, layers, and governance from a reference mos.
- See [config.mos](config.mos) for the full 7-tier Red Hat example
  (depth 0 through depth 6) and multi-upstream examples.

### Contract as ticket

A contract is the **repository-native representation of an issue tracker
ticket**. The contract is always the source of truth; external trackers (Jira,
GitHub Issues, Linear, etc.) are mirrors synced bidirectionally via adapter
tooling.

- `tracker "github" { repo = "org/repo", issue = 42 }` -- one named block per
  external system. Multiple trackers per contract are allowed.
- The adapter name (block title) is a free identifier; fields are
  adapter-specific. The grammar imposes no constraints.
- Hierarchy mapping to trackers:

  | Contract structure     | Tracker concept                    |
  |------------------------|------------------------------------|
  | Umbrella contract      | Epic / Initiative                  |
  | Nested sub-contract    | Story / Task                       |
  | Leaf sub-contract      | Subtask                            |
  | `depends_on`           | "blocks" / "is-blocked-by" links   |
  | `status`               | Ticket workflow state              |
  | `feature` / `scenario` | Acceptance criteria                |

### Contract nesting and dependency ordering

- Contracts can nest recursively (contract-of-contracts). A parent contract
  decomposes a large scope into sub-contracts, each with its own acceptance
  criteria, status, and lifecycle.
- Dependencies between sibling sub-contracts are declared via `depends_on`
  (a list of sibling contract IDs). Parse order is not significant.
- **Linear ordering:** A -> B -> C (sequential phases). Each step depends on
  the previous.
- **Non-linear ordering:** A and B independent; C depends on both (diamond
  join). Arbitrary DAGs are supported within a parent scope.
- The optional `ordering` field (`"linear"` or `"non-linear"`) is advisory
  metadata. The dependency DAG declared by `depends_on` is canonical.
- See [contract.mos](contract.mos) for examples of flat, nested non-linear
  (diamond), and linear chain contracts with tracker adapter blocks.

### Specification blocks

- `feature "name" { }` can appear directly inside an artifact (inline) or
  inside an optional `spec { }` grouping block (multi-file composition)
- `spec { include "path" }` scopes `include` directives; `include` can only
  appear inside `spec {}`
- `scenario "name" { key = value, given { }, when { }, then { } }` accepts
  any `key = value` fields (open schema). `sut`, `test`, `case` are conventions,
  not grammar requirements
- `group "name" { }` groups related scenarios (replaces Gherkin's `Rule:`)
- `background { given { } }` defines shared preconditions
- Step blocks contain free-text lines; multiple lines = implicit AND

## Example Files

| File | Artifact Type | Demonstrates |
|------|---------------|--------------|
| [config.mos](config.mos) | `config` | Recursive upstream federation: 7-tier org chain (CEO→Engineer), multi-upstream DAG |
| [declaration.mos](declaration.mos) | `declaration` | Project identity declaration |
| [vocabulary.mos](vocabulary.mos) | `vocabulary` | User-defined terms and labels |
| [layers.mos](layers.mos) | `layers` | Resolution layer hierarchy |
| [simple-rule.mos](simple-rule.mos) | `rule` | Minimal rule with inline feature |
| [kubernetes-rule.mos](kubernetes-rule.mos) | `rule` | Kubernetes-scale rule (16 scenarios, 7 groups, sut/test) |
| [contract.mos](contract.mos) | `contract` | Flat, nested (DAG), and linear chain contracts |
| [multi-file-rule.mos](multi-file-rule.mos) | `rule` | Multi-file composition via `spec { include }` |
| [standalone-feature.mos](standalone-feature.mos) | `feature` | Standalone feature file (included by other artifacts) |

## Decision History

The DSL was chosen over TOML+Gherkin (CON-2026-005) and HCL based on a
10-criterion evaluation (DSL 5, TOML 1, Tie 4). Key advantages: single parser,
canonical formatter, unified AST, 3-4x faster parse at scale.

The grammar evolved from v1 (capitalized Gherkin keywords, `spec {}` mode
switch) to v3 (all lowercase, fully brace-delimited, `spec {}` as optional
grouping). See CON-2026-006 for the migration contract. CON-2026-007 added
vocabulary-driven keyword localization, open artifact types, and open scenario
fields.
