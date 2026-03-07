# Mos -- Current Goal

> Invest in the workshop before building the ship. Then see before you govern.

---

## Objective

Two-part goal:

1. **Build the test infrastructure** that makes agentic, multi-user, Git-compatible development testable from day one. Without this, governance features cannot be meaningfully verified.
2. **Build the minimum Mos** that can see its own code and display it in a TUI. No governance, no rules, no hash chains yet -- pure observability.

The target language is Go `.go` files because Mos is written in Go. The dog-food principle: Mos's first scan target is Mos itself.

---

## Why This Order

The previous roadmap jumped from project skeleton to feature work. But Mos operates in a distributed, multi-user, Git-based environment. Without test infrastructure:

- `mos commit` can only be unit-tested (no real Git interaction)
- `mos push` / `mos clone` require manual setup against real GitHub
- Multi-user governance (two identities signing a Bill) is untestable
- Sync tick / event propagation has no simulated network
- Git compatibility is asserted by faith, not by proof

**The principle: you cannot build an agentic workflow on untestable foundations.**

Once the test universe exists, the observability layer inverts the governance-first ordering:

1. **Test the plumbing** -- prove Git compatibility, multi-user sync, and event propagation in a simulated world.
2. **See the code** -- parse Go source, extract structure, build a module tree and dependency graph.
3. **Show the code** -- render the module tree and imports in a TUI using bubbletea.
4. **Dog-food** -- point the tool at itself. If the TUI cannot render Mos's own codebase, the scanner is broken.
5. **Then govern** -- design rules, scoping, and enforcement informed by the module structure the tool already understands. Rules can reference actual packages and domains because the tool knows they exist.

---

## Concrete Steps

### Step 0: Test Infrastructure (`testkit/`) -- COMPLETE

> **Status:** Done. 28 passing tests across 6 packages. Includes `moslib/primitive/` (added during execution as the artifact primitive that governance commands will be vocabulary sugar on top of). Containerized Gitea forge deferred to a later batch.

Build the simulated actor system that all subsequent development depends on. Mos's distributed architecture follows the Actor Model: each participant (user, forge, sync tick) is an actor with private state, communicating exclusively through asynchronous message passing. Go's goroutines + channels provide the native actor primitives (goroutine = actor, channel = mailbox, select = message dispatch, context = supervision). The testkit simulates this actor system so every feature is testable in a distributed, multi-user environment from day one.

Mos must also masquerade as a Git client -- any Git remote (GitHub, GitLab, self-hosted) accepts `mos push` and serves `mos fetch` without knowing the difference. Adoption of Mos is zero-friction. This must be proven, not assumed.

**`testkit/gitcompat/`** -- Git 1:1 compatibility verification. Runs identical operations with both `git` (real binary) and `mos`, asserts identical results:

- Object format: blobs, trees, commits, tags produced by `mos` are byte-for-byte valid to `git fsck`
- Ref operations: `mos` reads/writes/deletes refs that `git` accepts; reflog entries are valid
- Pack protocol: `mos push` produces pack data that `git receive-pack` accepts; `mos fetch` consumes packs that `git upload-pack` produces
- Config compatibility: `extensions.mos = true` triggers Git's extension guard (standard `git` refuses; `mos` proceeds)
- Round-trip: `mos commit` + `git log` shows the commit; `git commit` + `mos log` shows the commit

This suite is the proof that Mos can masquerade as Git. It runs in CI on every commit.

**`testkit/forge/`** -- Dual-layer fake code forge. Both layers implement the same `Forge` interface so tests run against either:

- In-process (`memory.go`): Go HTTP server using `go-git`'s server-side plumbing. Starts in ~10ms, runs inside `go test`, no external dependencies. Supports clone/push/fetch over HTTP smart protocol, ref advertisement, `refs/mos/*` namespace, and webhook dispatch.
- Containerized (`gitea.go`): Real Gitea instance via `testcontainers-go`. Starts in ~5s. Provides real SSH and HTTP Git access, repository creation via API, pull request API, and webhook delivery.

**`testkit/user/`** -- Simulated users with cryptographic identity. Each test user has:

- Ephemeral ED25519 keypair (generated per test, no disk)
- Own local clone with `.mos/` initialized
- Own Cabinet (`~/.mos/cabinet/`) with the user's identity
- Own `mos` session (can run any `mos` command as that user)
- Signing capability (can sign governance actions with their key)

**`testkit/network/`** -- Actor message transport simulation:

- In-process channel-based message bus (each actor has a mailbox; events are delivered to all mailboxes)
- Message recorder for assertions
- Partition injection (alice's messages reach the forge actor, bob's are dropped)
- Latency injection (delayed message delivery on specific actor-to-actor paths)
- Deterministic clock for sync tick (self-addressed delayed message) testing (no real `time.Sleep`)

**`testkit/world/`** -- Test scenario orchestrator. Composes forge + users + network into a test world:

```go
func TestMultiUserAmendment(t *testing.T) {
    w := world.New(t).
        WithForge(forge.InProcess()).
        WithUsers("alice", "bob").
        WithNetwork(network.Default()).
        Build()
    defer w.Close()

    w.User("alice").Run("init")
    w.User("alice").Run("rule", "create", "no-unused-imports")
    w.User("alice").Push()

    w.User("bob").Clone(w.Forge().URL("project"))
    w.User("bob").Run("bill", "create", "--amend", "no-unused-imports")
    w.User("bob").Push()

    w.User("alice").Pull()
    w.User("alice").Run("bill", "ratify", "BILL-2026-001")
    w.User("alice").Push()

    w.User("bob").Pull()
    w.AssertRuleVersion("no-unused-imports", 2)
    w.AssertChainLength(4)
}
```

Step 0 is complete when: Git compatibility suite passes, in-process and containerized forge work, a 3-user scenario (create, amend, ratify) passes end-to-end, and event propagation assertions work with partition and delay injection.

### Step 1: Project Skeleton -- COMPLETE

> **Status:** Done. Binary and package renamed from `con` to `mos`. `cmd/` directories are `cst/` and `mos/`. The experimental TUI (formerly `curia`) has been relocated to `testkit/curia/`. Binary output name is `mos` via `go build -o`.

```
mos/
├── go.mod
├── go.sum
├── cmd/
│   ├── mos/          # CLI binary
│   │   └── main.go
│   └── (curia relocated to testkit/curia/)

├── moslib/
│   ├── survey/       # Source code scanning
│   │   ├── scanner.go      # Interface + Go scanner implementation
│   │   └── scanner_test.go
│   ├── model/        # Data structures
│   │   ├── module.go       # Module, Package, Symbol, File types
│   │   ├── imports.go      # Import graph types
│   │   └── model_test.go
│   ├── linter/       # .mos/ format linter
│   │   ├── linter.go       # Validate .mos/ artifacts
│   │   ├── rules.go        # Lint rules (schema, Gherkin, @include, etc.)
│   │   └── linter_test.go
│   └── lsp/          # Mos LSP server
│       ├── server.go       # LSP protocol handler
│       ├── toml.go         # TOML language features
│       ├── gherkin.go      # Embedded Gherkin language features
│       └── diagnostics.go  # Real-time lint diagnostics
├── testkit/
│   ├── gitcompat/    # Git 1:1 compatibility verification
│   │   ├── compat.go       # Assert identical output between git and mos
│   │   ├── objects.go      # Verify blob/tree/commit/tag format
│   │   ├── refs.go         # Verify ref read/write/delete/reflog
│   │   ├── protocol.go     # Verify pack protocol (fetch/push wire format)
│   │   └── pack.go         # Verify pack file format and delta compression
│   ├── forge/        # Fake code forge (dual-layer)
│   │   ├── forge.go        # Forge interface
│   │   ├── memory.go       # In-process go-git forge
│   │   ├── gitea.go        # Containerized Gitea forge
│   │   └── hooks.go        # Webhook simulation
│   ├── user/         # Simulated users
│   │   ├── user.go         # User with SSH identity, clone, Cabinet
│   │   ├── keygen.go       # Ephemeral ED25519 key generation
│   │   └── allowed.go      # Generate allowed_signers from test users
│   ├── network/      # Event simulation
│   │   ├── bus.go          # In-process pub-sub event bus
│   │   ├── partition.go    # Network partition injection
│   │   ├── delay.go        # Latency simulation
│   │   └── recorder.go     # Event recorder for assertions
│   └── world/        # Test scenario orchestration
│       ├── world.go        # World builder
│       ├── scenario.go     # Pre-built multi-user scenarios
│       └── assertions.go   # High-level governance assertions
├── ARCHITECTURE.md
├── CURRENT-GOAL.md
└── README.md
```

`go mod init` with the package skeleton. `moslib/` is the library (no I/O). `cmd/mos/` is the CLI binary (renamed from `cmd/con/`). The experimental TUI binary formerly at `cmd/curia/` has been relocated to `testkit/curia/`. `testkit/` is the test infrastructure (used by `_test.go` files throughout the project). `moslib/primitive/` was added during execution as the artifact primitive package. `moslib/linter/` and `moslib/lsp/` are part of Phase 1 because they are the harness applied to Mos's own format -- dog-fooding verification from day one.

### Step 2: Survey Scanner (`moslib/survey/`) -- COMPLETE

> **Status:** Done. `GoScanner` implemented using `go/parser` + `go/ast`. Extracts packages, symbols (func, type, interface, const, var), and import graph with internal/external classification. 4 unit tests + 1 dog-food integration test. Delivered by CON-2026-002.

Build a Go source scanner using `go/parser` and `go/ast`:

- Walk a directory tree, find `*.go` files.
- Parse each file into an AST.
- Extract: package name, file path, declared symbols (functions, types, interfaces, constants, variables), exported vs. unexported, import paths.
- Return a `model.Module` representing the full module tree.

The scanner accepts a filesystem abstraction (not raw `os` calls) so it is testable without touching disk.

### Step 3: Model Types (`moslib/model/`) -- COMPLETE

> **Status:** Done. Constructors (`NewModule`, `NewPackage`, `NewFile`), builder methods (`AddPackage`, `AddFile`, `AddSymbol`), `ImportGraph` with dedup and `EdgesFrom`, `SymbolKind.String()` + JSON marshal/unmarshal, JSON tags on all fields. 8 unit tests. Delivered by CON-2026-002.

Define the core data structures:

- **Module** -- the top-level Go module (from `go.mod`). Contains packages.
- **Package** -- a Go package directory. Contains files and symbols. Has an import path.
- **File** -- a single `.go` file. Belongs to a package.
- **Symbol** -- a declared name: function, type, interface, constant, variable. Has visibility (exported/unexported).
- **ImportGraph** -- directed graph of package-to-package imports. Distinguishes internal (within module) and external (third-party) imports.

The model is language-agnostic at the type level. The first scanner is Go-specific, but `Module`, `Package`, `Symbol`, `File` are generic enough for future language adapters.

### Step 4: CLI Output (`cmd/mos/`) -- COMPLETE

> **Status:** Done. `mos survey <path>` wired with human-readable tree (default) and JSON (`--format json`) output modes. Dog-food verified: Mos sees all 12 packages and 67 import edges. Delivered by CON-2026-002.

`mos survey` prints the module structure and imports to stdout. Two output modes:

- **Human** (default): indented tree with package names, symbol counts, import relationships.
- **JSON** (`--format json`): structured output for machine consumption and TUI ingestion.

### Step 5: Linter (`moslib/linter/`) -- COMPLETE

> **Status:** Done. Full context-aware linter with two layers (universal schema + project-specific context). `ProjectContext` loader, constrained Gherkin parser, schema validators for all artifact types (config, declaration, rule, contract), `@include` resolver, cross-artifact/vocabulary/layer/template validators. `mos lint <path>` with human and JSON output. 24 tests. Delivered by CON-2026-003.

The Mos linter validates `.mos/` artifacts against the format specification. Even though Phase 1 doesn't create `.mos/` files yet, the linter is built now so it's ready the moment Phase 2 produces the first Declaration. The linter is the first harness Mos applies to itself.

- **Schema compliance**: required TOML fields present, types correct, enum values valid.
- **Embedded Gherkin validity**: `[spec].feature` fields parse as valid Gherkin.
- **`@include` resolution**: referenced `.feature` files exist and contain valid Gherkin.
- **Cross-artifact consistency**: rules referenced in contracts exist, vocabulary terms are defined.

Runs as `mos lint`. Returns structured diagnostics (file, line, severity, message) consumable by the LSP.

### Step 6: LSP (`moslib/lsp/`) -- COMPLETE

> **Status:** Done. Full LSP server: JSON-RPC 2.0 transport over stdio, document sync, linter-driven diagnostics, TOML/Gherkin completion, hover docs, go-to-definition for rule IDs and @include paths. Stress fixture generators at Linux Kernel and Kubernetes scale. 16 LSP tests + 3 stressgen tests. Delivered by CON-2026-004.

The Mos LSP server treats `.mos/*.toml` files as compound documents -- TOML at the outer level, Gherkin within `[spec].feature` fields. Ships as `mos lsp`.

- **TOML intelligence**: schema-aware completion, diagnostics for missing fields, hover docs.
- **Embedded Gherkin intelligence**: syntax highlighting within multi-line strings, Given/When/Then step completion.
- **Lint integration**: real-time diagnostics from `moslib/linter/` surfaced as LSP warnings/errors.
- **Cross-reference**: navigate from rule IDs to their definitions, from steps to step definitions.

The LSP is the harness applied to the format itself. It dog-foods Mos's verification principle by continuously validating `.mos/` artifacts as they are edited.

### Step 6.5: DSL Evaluation -- COMPLETE (GO decision)

> **Status:** Done. DSL wins over TOML+Gherkin (5-1-4 on 10 criteria). DSL parse 3-4x faster at scale, eliminates compound document complexity, enables canonical formatting. Produced prototype parser/formatter in `moslib/dsl/` and v3 grammar design. Delivered by CON-2026-005. v1-to-v3 grammar migration delivered by CON-2026-006 (complete). Linter and LSP migrated from TOML to DSL parser in parallel.

The TOML+Gherkin "single format" was actually two formats: TOML at the outer level, Gherkin embedded as opaque multi-line strings. A purpose-built DSL genuinely unifies both into a single parseable grammar where Gherkin keywords are native syntax, not embedded strings. The evaluation spike proved the concept; the linter and LSP will migrate from TOML to DSL after the v3 grammar is implemented.

### Step 6.9: Dependency Modernization & BDD Harness -- COMPLETE

> **Status:** Done. stdlib `maps`/`slices`/`cmp` adopted (zero new deps). Ginkgo v2.28.1 + Gomega v1.39.1 added; `testkit/network/` fully migrated to BDD (6 specs). `dsl.StringValue` extracted. Delivered by CON-2026-007.

Reduce bespoke code surface by adopting stdlib and community packages before adding new feature layers. Three workstreams:

1. **stdlib `maps`/`slices`/`cmp` adoption.** Go 1.25.7 includes `maps`, `slices`, `cmp`, and `iter` -- all unused today. Replace hand-rolled collect-sort-iterate patterns, custom `containsNewline` helpers, and `sort.Strings`/`sort.Slice` calls with their stdlib equivalents. Net: fewer custom helpers, idiomatic modern Go, zero new external dependencies.

2. **Ginkgo + Gomega BDD testing framework.** Mos defines BDD specs (`feature`/`scenario`/`given`/`when`/`then`); its own tests should be written in the same paradigm. Phased rollout:
   - **Phase A:** Add Gomega matchers via `NewWithT(t)` -- zero structural disruption, immediately better assertions.
   - **Phase B:** Write new integration/stress test suites in Ginkgo (`Describe`/`Context`/`It`).
   - **Phase C:** Migrate existing test files to Ginkgo as they are touched.
   - **Harness bridge (future):** Mechanical rule `harness` blocks shell out to `ginkgo run`; JUnit output feeds back into contract `evidence` blocks. Closes the dogfooding loop.

3. **Internal DRY extraction.** Deduplicate repeated patterns: `storeBlob`/`gitExec` pairs, `filepath.Walk` + filter loops, artifact ID generation, `dsl.Value` → string extraction.

This step is sequenced before TUI because a cleaner, well-tested foundation reduces the cost of every subsequent feature.

### Step 7: TUI (`testkit/curia/`, formerly `cmd/curia/`)

A bubbletea application with two panels:

- **Module Tree** (left): navigable tree of packages → files → symbols. Expand/collapse with enter. Shows exported symbols in one style, unexported in another.
- **Import Graph** (right): for the selected package, show what it imports and what imports it. Internal imports distinguished from external.

Charm stack: bubbletea (application framework), lipgloss (styling), bubbles (standard components).

### Step 8: Dog-Food Gate

Run `mos survey`, `mos lint`, `mos lsp`, and `macro` against the `mos/` directory itself.

- The TUI displays Mos's own module structure and imports: `moslib/survey/`, `moslib/model/`, `moslib/linter/`, `moslib/lsp/`, `cmd/mos/`.
- The import graph shows the real dependency chain: `cmd/mos` → `moslib/survey` → `moslib/model` → `go/ast`, etc.
- The linter is ready to validate `.mos/` artifacts the moment Phase 2 creates them.
- The LSP provides intelligence when editing `.mos/` files in any LSP-compatible editor.

If the tool can see itself and is ready to verify itself, Step 8 passes. This is the **zero milestone** from the Roadmap.

---

## What This Enables

Once the test infrastructure, Module Structure, Import views, linter, and LSP exist:

- **Every feature is testable from birth.** The testkit means `mos commit`, `mos push`, multi-user governance, and sync events have E2E test coverage from the first line of feature code. No feature ships without proof.
- **Git compatibility is proven, not assumed.** The gitcompat suite runs in CI. If `mos` ever produces output that `git` rejects, the build breaks. Adoption of Mos is zero-friction by construction.
- **Multi-user scenarios are first-class.** Two users amending the same rule, three users ratifying a Bill, a partitioned user rejoining -- these are all `go test` cases, not manual QA.
- **Rules get scoping for free.** A rule scoped to `moslib/survey/` means something concrete because the tool already knows that package exists and what it contains.
- **Drift detection has a map.** The Module Structure view becomes the Resolution Map's first zoom level. Governance state (rules, contracts, drift indicators) overlays directly onto the structure the tool already renders.
- **The TUI is already working.** Adding governance views (Declaration, rules, contracts, harness results) is incremental -- the TUI shell, navigation model, and rendering pipeline already exist from Phase 1.
- **The harness is already working.** The linter validates `.mos/` files from the moment Phase 2 creates them. The LSP provides real-time feedback. No gap between "we have a format" and "we can verify the format." The harness is ready before the artifacts arrive.
- **The DSL is validated early.** The linter and LSP prove the purpose-built DSL works in practice before hundreds of rules and contracts depend on it. If the grammar needs adjustment, the cost of change is near zero in Phase 1.
- **Dog-food loop tightens.** Every new `moslib/` package, every new `cmd/` binary, every new import immediately appears in the tool's own output. Every `.mos/` artifact is linted and LSP-supported. Development is self-documenting and self-verifying.

---

## Success Criteria

The goal is met when:

1. `go test ./testkit/gitcompat/...` proves object, ref, and protocol compatibility between `mos` and `git`.
2. `go test ./testkit/forge/...` runs against both in-process and containerized Gitea.
3. A 3-user scenario (create repo, push rule, amend via Bill, ratify) passes end-to-end.
4. Event bus correctly propagates governance events; partition test confirms isolation.
5. `mos survey .` produces a correct module tree and import graph for a Go codebase.
6. `mos lint` validates `.mos/` DSL artifacts (schema, Gherkin scenarios, `@include` resolution).
7. `mos lsp` provides real-time intelligence for `.mos` files in an LSP-compatible editor.
8. `macro` renders the module tree and import graph interactively.
9. All of the above work when pointed at the `mos/` repository itself.
10. Tests pass: `go test ./...` covers testkit, scanner, model, linter, and LSP packages.

---

## Phase 2: Self-Hosted Governance

> The dog eats its own steak.

Phase 2 transitions Mos from "tools that could validate governance artifacts" to "tools that do validate their own governance artifacts." Design principle: **primitives, not prescriptions** -- new capabilities compose from existing building blocks (blocks, fields, feature/scenario/given/when/then) with vocabulary controlling validity. No grammar changes. ALM-specific concepts (Polarion pillars, Jira epics) are vocabulary terms, never schema keywords.

### CON-2026-011: Governance CLI + Self-Hosted Bootstrap -- COMPLETE

Built `mos init`, `mos rule create`, `mos contract create`, and `mos fmt` CLI commands in `moslib/governance/` and `cmd/mos/`. Bootstrapped Mos's own `.mos/` directory with config, vocabulary, layers, declaration, and 3 rules. `mos lint .` passes with zero errors. Core constraint: the agent used CLI commands exclusively.

### ~~CON-2026-012: Contract Migration (via CLI) -- ABANDONED~~

Abandoned: `mos contract create` only supports `title`, `status`, `goal`, `depends_on`. Migrating into this minimal schema would produce hollow artifacts missing acceptance criteria, coverage matrices, and environment requirements. Replaced by CON-2026-015.

### CON-2026-015: Enhanced Contract Model + Migration

First enhance the building blocks: `--spec-file` for composing feature/scenario blocks into contracts, `--coverage-file` for structured coverage matrices, `--harness-requires` for environment prerequisites on rules, and vocabulary template expansion for ALM absorption. All enhancements compose from existing parser/AST primitives -- no grammar changes. Then migrate all 14 markdown contracts via the enhanced CLI. Delete originals. Single source of truth.

### CON-2026-013: Harness Bridge

Build `moslib/harness/` and `mos harness run .` to discover and execute harness blocks from `.mos/rules/`. Dog-food against Mos's own rules (`go build ./...`, `go test ./...`). The judiciary gets teeth.

### CON-2026-014: Multi-Repo Split-Brain Governance Research -- COMPLETE

Evaluated 5 approaches to governing a product spanning multiple repositories (PTP Operator: 4 repos, 3 orgs) against 10 dimensions. Analyzed existing art (K8s OWNERS, Helm umbrella charts, go.work, Bazel WORKSPACE, Nix flakes). Stress-tested at Dunix scale (100+ repos). Recommended **Hybrid C+E** (mono-mos in primary repo + cross-repo spec includes). One new contract proposed: CON-2026-016 (Cross-Repo Include Resolver). No injections needed into CON-2026-013 or CON-2026-015. Full document: `docs/multi-repo-governance-research.md`.

### CON-2026-016: Cross-Repo Include Resolver

Implements CON-2026-014's recommendation: multi-repo testkit (`Participant` with `Repos map`), governance scaffolding API (`InitWithUpstream`, `AddSpecInclude`, `WriteProductManifest`, `AddCrossRepoBlock`), and multi-repo linter (`Workspace` type, graph crawler, upstream/include/product-graph resolvers, `LintWorkspace`). Crawl-first architecture: an agent given any single repo can discover the entire organizational graph by following upstream blocks, cross-repo includes, and product manifests. Explicit workspace map serves as offline fallback for air-gapped CI. DAG-only (cycles rejected). No grammar changes.

### CON-2026-017: Governance Model Benchmark

Validates CON-2026-014's qualitative evaluation matrix with quantitative metrics. Scaffolds all 8 governance models (5 pure + 3 hybrid) across PTP 4-repo and Dunix 100-repo scenarios. Collects 8 measurable scorecard dimensions. Produces comparison table. Confirms or revises the Hybrid C+E recommendation.

### CON-2026-018: Magis IDE Architecture Research

Case studies of Void Editor (fork-vs-build), Ollama (BYOB integration), and Aider (terminal AI coding existence proof). Architecture analysis of micro-editor (Go terminal editor library, `internal/` problem), Neovim (headless + RPC), and Helix (batteries-included + tree-sitter). Editor embedding evaluation matrix (4 options). Origami integration design (Macro as consumer, IDE-domain Components, ProviderRouter wiring to Ollama). Produces `docs/macro-ide-architecture.md` with recommended architecture and phased roadmap.

### CON-2026-019: Macro PoC

Implementation contract spawned by CON-2026-018. Builds the Macro agentic IDE prototype: Macro (forked from Micro editor) as the system code editor, connecting to a Macro Server (headless daemon) for sessions, buffers, and circuit execution. Combines code editing, code observability (module tree + import graph), Origami circuit wiring (Macro Component, ProviderRouter, IDE circuits), Sumi circuit visualization (Kami EventBridge), and agent integration (Ollama BYOB, MuxDispatcher, Papercup workers, agent chat). Dog-food gate: edit Mos's own code in Macro, run governance circuits, visualize execution, chat with an agent.

### Sequencing

```
CON-2026-011 (DONE) → CON-2026-015 (Enhanced Building Blocks + Migration)
                    → CON-2026-013 (Harness Bridge)
CON-2026-014 (DONE) → CON-2026-016 (Cross-Repo Include Resolver)
                    → CON-2026-017 (Governance Model Benchmark, depends on 016)
                    → CON-2026-018 (Macro IDE Architecture Research)
                        → CON-2026-019 (Macro PoC, depends on 018)
```

### Phase 2 Success Criteria

1. `mos init .` produces a valid `.mos/` scaffold that passes `mos lint .` -- DONE (CON-2026-011)
2. Mos's own `.mos/` directory was created entirely via CLI commands -- DONE (CON-2026-011)
3. `mos contract create --spec-file` composes feature blocks into contracts (CON-2026-015)
4. All 14 contracts exist in `.mos/contracts/` in valid DSL format with acceptance criteria and coverage (CON-2026-015)
5. Original `contracts/*.md` files are deleted (CON-2026-015)
6. `mos harness run .` executes all harness blocks and passes (CON-2026-013)
7. `mos lint .` passes with zero diagnostics
8. Multi-repo governance research document produced with ranked recommendation -- DONE (CON-2026-014)

---

## Phase 3: Bills and Contracts

> The legislative system. After this phase, work is trackable through the full Bill lifecycle.

**Deliverables:**

- **Bill introduction.** `mos bill create` -- a public proposal visible to all participants. Declares intent, scope, and operating rules.
- **Bill lifecycle.** First reading, committee stage (laboratory), readings and debate, ratification or abandonment. Each stage is tracked in the `.mos`.
- **Contract execution.** Ratified Bills become binding Contracts. Status tracking: draft, active, complete, abandoned.
- **Contract rules.** Temporary amendments scoped to a single Contract's lifetime. Override or extend standing law for the duration of the work. Expire on completion.
- **Negative space preservation.** Rejected Bills, abandoned Contracts, and dissenting opinions remain in the historical record. Nothing is deleted.

**Dependency:** Phase 2.

---

## Phase 4: Harness (Verification Runtime)

> The judiciary. After this phase, rules have teeth.

**Deliverables:**

- **Gherkin scenario execution (Mechanical + Blue Team).** The Harness executes Given/When/Then scenarios from `.mos` feature blocks against the Concrete State. Each scenario is a mechanical gate: pass/fail is binary. The DSL's native Gherkin syntax maps steps to Go step definitions (shipped defaults + project-specific extensions).
- **ROGBY as default methodology.** Red-Orange-Green-Yellow-Blue shipped as the default mechanical verification cycle. Overridable.
- **Rule resolution.** For any given Contract, compute the applicable rule set: standing law + contract amendments - suspended rules. Deterministic -- no ambiguity about what rules apply.
- **Drift as scenario coverage.** Integrity Index metrics derived from scenario pass rates: how many scenarios in the Desired State pass against the Concrete State. Drift = failing scenario count.
- **Blue/Red axis acknowledgment.** The PoC implements **Mechanical + Blue** (scenarios, benevolent stimulus). Mechanical + Red (adversarial probing), Verbal + Blue (interpretive judgment), and Verbal + Red (adversarial review) are deferred to post-PoC.

**Dependency:** Phase 2. (Phases 3 and 4 can run in parallel.)

---

## Phase 5: Macro and Agent Interface

> The full interface layer. After this phase, humans and agents can interact with the complete `.mos` system.

The Console First, Agent-Accelerated principle applies: there is no separate agent API. The CLI is the API. Agents use the same `mos` commands humans do.

**Deliverables:**

- **`mos` CLI completion.** The remaining commands beyond `mos init` (Phase 2) and `mos survey` (Phase 1): `mos verify` (hash chain integrity), `mos sync` (Heartbeat on demand), `mos identity` (allowed_signers management), `mos onboard` (newcomer overview). Human-friendly by default, machine-parseable via `--format json`.
- **Macro maturation.** The editor gains the full governance view: active Contracts, Bill lifecycle, Harness results, drift dashboard. E2E assertions use Gherkin: `Given the Map is visible / When I zoom to domain level / Then the drift indicator shows 0`.
- **`mos survey` expansion.** Beyond Go: additional language adapters for the languages the governed codebase uses. The `moslib/survey/` interface defined in Phase 1 accepts new scanner implementations.
- **Mos Helper Agent.** The gate deliverable. Ingests a `.cursor/` directory or kombucha.mdc seed and bootstraps a `.mos/` from it. Interactive -- asks clarifying questions. Produces a valid, enforceable mos.

**Dependency:** Phases 2, 3, and 4.

---

## Deferred (Post-PoC)

Part of the full vision but not required to retire kombucha or prove Mos works:

- **Sophia (Context Cache) / Monad (Archive)** -- Gravitational memory layer (local and collective). The PoC does not need persistent agentic memory.
- **Origami (Router)** -- Signal routing and the Signal Chain. The PoC uses direct agent invocation, not graph-based routing. (Origami is independently developed as a Go library at `github.com/dpopsuev/origami`.)
- **Embassy (External Adapters)** -- Bidirectional adapters to external systems (issue trackers, CI, infrastructure). The PoC operates within the repository boundary.
- **GUI client** -- A graphical interface is post-PoC. Macro (terminal) is the primary client.
- **Multi-user sessions** -- The PoC targets single-developer workflows.
- **Peer-to-peer sync** -- CRDT-based real-time sync between developers, peer discovery via `refs/mos/participants/`, direct sync via libp2p. The PoC uses Git's hub-and-spoke model with a slow-tick Heartbeat.
- **Federation, Koinonia, Ecumenical Councils** -- Multi-project and cross-fork governance. The PoC targets a single project-state.
- **Verbal adjudication** -- Governance-mediated Warnings require the full judiciary model. The PoC implements Mechanical + Blue only; Verbal rules are stored but not adjudicated.
- **Red Team adversarial probing** -- Adversarial verification requires mature Harness infrastructure and domain-specific attack generation. Deferred until the Blue Team gate is proven.

---

## Dogfood Milestones

**Zero milestone (Phase 1 gate).** Mos's testkit proves Git compatibility (`git fsck` accepts `mos` output), a 3-user governance scenario passes end-to-end, and Macro displays Mos's own module structure and imports. The tool proves its compatibility, tests its own governance, and sees itself.

**First milestone.** Express the founding documents as `.mos` artifacts, rendered and manipulated through Macro. If `.mos` cannot express the very document that defines it, it cannot express anything.

**Gate milestone.** The Mos Helper Agent converts `asterisk/.cursor/` (the hardest case) into a valid `.mos/`. Kombucha is retired. The uncodified mos becomes codified.

**Final milestone.** Macro self-compiles. The borrowed shell (Cursor) is no longer needed. The embryo outgrows its host.
