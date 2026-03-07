# Multi-Repo Split-Brain Governance Research

**Contract:** CON-2026-014
**Date:** 2026-03-02
**Status:** Complete

## Abstract

This document evaluates five approaches to governing products that span multiple
repositories, with particular attention to the "split brain" scenario where E2E
tests live in a separate repository from the product code. The PTP Operator on
OpenShift (4 repos, 3 GitHub orgs) is the concrete reference case. Each option
is scored against 10 evaluation dimensions, stress-tested at Dunix scale (100+
repos), and assessed for hybrid combinations. The recommended approach is
**C+E (Mono-mos with cross-repo spec includes)**.

---

## 1. Reference Scenario: PTP Operator on OCP

### Repositories

| Repo | Org | Role |
|------|-----|------|
| `openshift/ptp-operator` | openshift | Kubernetes operator, CRDs, controllers |
| `openshift/linuxptp-daemon` | openshift | DaemonSet managing ptp4l/phc2sys per node |
| `redhat-cne/cloud-event-proxy` | redhat-cne | O-RAN cloud-native event sidecar |
| `rh-ecosystem-edge/eco-gotests` | rh-ecosystem-edge | Multi-product E2E Ginkgo suite |

### Key tensions

1. **Governance fragmentation:** A product-level rule ("all PTP APIs must be
   backward-compatible") spans three component repos. Where does it live?
2. **Test-product split brain:** Acceptance criteria belong to the product but
   the Ginkgo specs validating them live in `eco-gotests` (different org,
   different CI pipeline).
3. **Cross-org ownership:** Three GitHub orgs with different OWNERS, merge
   policies, and CI systems.
4. **N:M shared test repo:** `eco-gotests` tests PTP, SR-IOV, ZTP, KMM, and
   more. PTP-specific governance must not constrain SR-IOV test code.
5. **Versioning drift:** Component repos release on different cadences. A
   contract referencing a spec in `eco-gotests` can drift silently.

---

## 2. Options Enumerated

### Option A: Product-Level Umbrella Repo

A dedicated repository (`openshift/ptp-mos`) owns product-level
`.mos/`. Component repos and the test repo each have minimal `.mos/` with
`upstream` blocks pointing at the umbrella.

```
┌──────────────────────────────────────┐
│  ptp-mos (umbrella)         │
│  .mos/                               │
│  ├── config.mos       (domain root)  │
│  ├── vocabulary/default.mos          │
│  ├── resolution/layers.mos           │
│  ├── rules/                          │
│  │   ├── mechanical/ptp-api-compat.mos
│  │   └── mechanical/ptp-event-schema.mos
│  └── contracts/                      │
│      └── active/CON-PTP-E2E/         │
│          └── contract.mos            │
└────────────┬─────────────────────────┘
             │ upstream
   ┌─────────┼─────────┬──────────────┐
   ▼         ▼         ▼              ▼
ptp-op   daemon   event-proxy   eco-gotests
.mos/    .mos/    .mos/         .mos/
config   config   config        config
(↑umb)   (↑umb)   (↑umb)        (↑umb)
```

**Pros:** Clean separation. Product rules in one place. Existing `upstream`
model works without changes. Component repos inherit and specialize.

**Cons:** Extra repo to maintain with no code (pure governance overhead).
Hard to justify for small products. At 100+ repos, many codeless umbrella
repos add noise.

#### Concrete `.mos/` layouts

**ptp-mos (new repo):**
```
.mos/
├── config.mos
├── vocabulary/default.mos
├── resolution/layers.mos
├── rules/
│   ├── mechanical/ptp-api-compat.mos
│   └── mechanical/ptp-event-schema.mos
└── contracts/
    └── active/CON-PTP-E2E/contract.mos
```

`config.mos` — domain root:
```
config {
  mos { version = 1 }
  backend { type = "git" }
  governance {
    model = "committee"
    scope = "product"
    jurisdiction = "ptp-operator"
  }
}
```

**ptp-operator, linuxptp-daemon, cloud-event-proxy** — each identical structure:
```
.mos/
└── config.mos
```

`config.mos`:
```
config {
  mos { version = 1 }
  backend { type = "git" }
  upstream "ptp-product" {
    url   = "https://github.com/openshift/ptp-mos"
    ref   = "main"
    scope = "product"
  }
  governance {
    model = "bdfl"
    scope = "component"
  }
}
```

**eco-gotests:**
```
.mos/
└── config.mos
```

`config.mos`:
```
config {
  mos { version = 1 }
  backend { type = "git" }
  upstream "ptp-product" {
    url   = "https://github.com/openshift/ptp-mos"
    ref   = "main"
    scope = "product"
  }
  upstream "eco-qe-org" {
    url   = "https://github.com/rh-ecosystem-edge/qe-mos"
    ref   = "main"
    scope = "organization"
  }
  governance {
    model = "consensus"
    scope = "test-suite"
  }
}
```

---

### Option B: Peer Federation with Cross-References

No umbrella. Each repo has `.mos/`. A new `cross_repo` block enables lateral
references between peers.

```
┌─────────────────┐   cross_repo   ┌──────────────────┐
│  ptp-operator    │◄─────────────►│  linuxptp-daemon  │
│  .mos/           │               │  .mos/            │
└────────┬────────┘               └───────────────────┘
         │ cross_repo
         ▼
┌──────────────────┐               ┌──────────────────┐
│ cloud-event-proxy│               │   eco-gotests     │
│  .mos/           │               │   .mos/           │
└──────────────────┘               └──────────────────┘
```

**DSL example (new primitive):**
```
rule "ptp-api-compat" {
  scope = "product"
  cross_repo {
    repos = ["openshift/linuxptp-daemon", "redhat-cne/cloud-event-proxy"]
    sync  = "bidirectional"
  }
}
```

**Pros:** No extra repos. Peers coordinate directly.

**Cons:** New `cross_repo` DSL primitive. Circular reference risk. Linter must
clone peers (expensive). No clear authority on conflict. Harder to reason about
than hierarchy.

#### Concrete `.mos/` layouts

Each of the 4 repos has a full `.mos/`:
```
.mos/
├── config.mos
├── vocabulary/default.mos
├── rules/
│   └── mechanical/ptp-api-compat.mos   # with cross_repo block
└── contracts/
    └── active/CON-PTP-E2E/contract.mos # duplicated or split
```

`config.mos` (no upstream, cross-repo peers declared in rules):
```
config {
  mos { version = 1 }
  backend { type = "git" }
  governance {
    model = "consensus"
    scope = "peer"
  }
}
```

The `ptp-api-compat` rule in each repo lists sibling repos via `cross_repo`.
Rules and contracts must stay synchronized across all 4 repos manually or via
a sync agent.

---

### Option C: Mono-Mos in the Operator Repo

`ptp-operator` is the "primary" and owns product-level `.mos/`. Daemon and
event-proxy point `upstream`. The test repo connects via `tracker` adapters.

```
┌────────────────────────────────────────┐
│  ptp-operator (primary)                │
│  .mos/                                  │
│  ├── config.mos           (domain root) │
│  ├── vocabulary/default.mos             │
│  ├── rules/                             │
│  │   ├── mechanical/ptp-api-compat.mos  │
│  │   └── mechanical/ptp-event-schema.mos│
│  └── contracts/                         │
│      └── active/CON-PTP-E2E/            │
│          └── contract.mos               │
│              tracker "eco-gotests" { .. }│
└───────────┬────────────────────────────┘
            │ upstream
   ┌────────┴────────┐
   ▼                 ▼
daemon          event-proxy       eco-gotests
.mos/           .mos/             (no .mos/ for PTP)
config(↑op)     config(↑op)       connected via tracker
```

**Pros:** No extra repo. Uses existing `upstream` model. Clear authority.
`tracker` adapter connects test repo loosely.

**Cons:** Operator repo carries product-level governance spanning peers.
Event-proxy is a different org — making it downstream of `openshift/` may not
match organizational reality. Test repo link is loose.

#### Concrete `.mos/` layouts

**ptp-operator (primary):**
```
.mos/
├── config.mos
├── vocabulary/default.mos
├── resolution/layers.mos
├── rules/
│   ├── mechanical/ptp-api-compat.mos
│   └── mechanical/ptp-event-schema.mos
└── contracts/
    └── active/CON-PTP-E2E/contract.mos
```

`config.mos`:
```
config {
  mos { version = 1 }
  backend { type = "git" }
  governance {
    model = "committee"
    scope = "product"
    jurisdiction = "ptp-operator"
  }
}
```

`CON-PTP-E2E/contract.mos`:
```
contract "CON-PTP-E2E" {
  title  = "PTP E2E Validation"
  status = "active"
  goal   = "All PTP acceptance criteria are validated by E2E tests."

  tracker "eco-gotests" {
    repo = "rh-ecosystem-edge/eco-gotests"
    path = "tests/cnf/core/network/ptp/"
    sync = "unidirectional"
  }

  feature "Clock synchronization" {
    scenario "PTP clock converges within 2s" {
      given { PTP operator is deployed on OCP cluster with supported NIC }
      when  { ptp4l process starts on worker nodes }
      then  { clock offset stabilizes below 100ns within 2 seconds }
    }
  }
}
```

**linuxptp-daemon:**
```
.mos/
└── config.mos
```

```
config {
  mos { version = 1 }
  backend { type = "git" }
  upstream "ptp-product" {
    url   = "https://github.com/openshift/ptp-operator"
    ref   = "main"
    scope = "product"
  }
  governance { model = "bdfl"  scope = "component" }
}
```

**cloud-event-proxy:** Same structure as `linuxptp-daemon`.

**eco-gotests:** No `.mos/` directory for PTP governance. Connected only via
the `tracker` block in `ptp-operator`'s contract.

---

### Option D: Virtual Product Manifest

A `product` block (new artifact type) in any repo lists all component
repositories. The linter uses it to resolve cross-repo references.

```
┌──────────────────────────────────────┐
│  ptp-operator                        │
│  .mos/                               │
│  ├── config.mos                      │
│  └── product.mos   (new artifact)    │
│      components: [ptp-op, daemon, ep]│
│      test_suites: [eco-gotests/ptp/] │
└──────────────────────────────────────┘
```

**DSL example (new artifact type):**
```
product "ptp-operator" {
  components = [
    { repo = "openshift/ptp-operator",       role = "primary"   },
    { repo = "openshift/linuxptp-daemon",    role = "component" },
    { repo = "redhat-cne/cloud-event-proxy", role = "component" },
  ]
  test_suites = [
    { repo = "rh-ecosystem-edge/eco-gotests", path = "tests/cnf/core/network/ptp/" },
  ]
  shared_rules = ["ptp-api-compat", "ptp-event-schema"]
}
```

**Pros:** Explicit product graph. No extra repo. Linter knows the full picture.
Analogous to `go.work`.

**Cons:** New artifact type requiring grammar and linter changes. Cross-repo
resolution is expensive (clone N repos to lint). "Any repo can host" creates
ambiguity.

#### Concrete `.mos/` layouts

**ptp-operator (hosts manifest):**
```
.mos/
├── config.mos
├── product.mos
├── vocabulary/default.mos
├── rules/
│   ├── mechanical/ptp-api-compat.mos
│   └── mechanical/ptp-event-schema.mos
└── contracts/
    └── active/CON-PTP-E2E/contract.mos
```

**Other repos:** Minimal `.mos/` with `upstream` pointing at `ptp-operator`,
or no `.mos/` at all (the manifest declares their membership).

---

### Option E: Test-Repo as Governance Consumer Only

Product repos own `.mos/`. The test repo has no `.mos/` for PTP governance.
Product contracts reference test specs via cross-repo `include` paths.

```
┌─────────────────────────────────────────┐
│  ptp-operator                            │
│  .mos/contracts/active/CON-PTP-E2E/      │
│    contract.mos                          │
│      spec {                              │
│        include "git://rh-ecosystem-edge/ │
│          eco-gotests@main:tests/.../     │
│          ptp_suite_test.go"              │
│      }                                   │
└─────────────────────────────────────────┘

eco-gotests: NO .mos/ for PTP governance
```

**DSL example (extends existing `include`):**
```
contract "CON-PTP-E2E" {
  title  = "PTP E2E Validation"
  status = "active"
  spec {
    include "git://rh-ecosystem-edge/eco-gotests@main:tests/cnf/core/network/ptp/ptp_suite_test.go"
  }
}
```

**Pros:** Clean ownership split: product repos own governance, test repo owns
implementation. No governance artifacts in shared test repo. Extends existing
`spec { include }` primitive.

**Cons:** Cross-repo include resolution requires network at lint time. Version
pinning (which commit?) adds coordination overhead. Silent breakage when test
repo changes. N:M problem remains (multiple products include from the same
test repo).

#### Concrete `.mos/` layouts

**ptp-operator:**
```
.mos/
├── config.mos
├── vocabulary/default.mos
├── rules/
│   ├── mechanical/ptp-api-compat.mos
│   └── mechanical/ptp-event-schema.mos
└── contracts/
    └── active/CON-PTP-E2E/contract.mos  # spec { include "git://..." }
```

**linuxptp-daemon, cloud-event-proxy:** Minimal `.mos/config.mos` with
`upstream` pointing at `ptp-operator`.

**eco-gotests:** No `.mos/` directory for PTP governance.

---

## 3. Existing Art Analysis

### 3.1 Kubernetes OWNERS

Per-directory OWNERS files with `approvers` and `reviewers` lists. OWNERS
cascade hierarchically within a single repo (parent directory applies to
children). An `aliases` file at repo root maps group names to individuals.

**Cross-repo model:** None. OWNERS is purely per-repo. Cross-repo coordination
happens via organizational structures (SIGs, Working Groups) that exist outside
the OWNERS file format. Prow enforces OWNERS for PR approval but cannot
reference another repo's OWNERS.

**Relevance to Mos:** OWNERS proves that per-repo hierarchical
governance works at Kubernetes scale (2000+ OWNERS files across 70+ repos).
But it explicitly does NOT solve cross-repo governance — that gap is filled by
human process (SIG charters, KEP process). Mos's `upstream` federation
is strictly more expressive than OWNERS.

### 3.2 Helm Umbrella Charts

A parent `Chart.yaml` declares `dependencies` listing sub-charts from other
repos (by repository URL + version). Values flow top-down from umbrella to
children. Sub-charts are versioned; `helm dependency update` resolves them.

**Cross-repo model:** Versioned dependency resolution. The umbrella pins
specific versions of sub-charts. Drift is prevented by explicit version bumps,
but this creates coordination overhead. The umbrella repo is a release artifact,
not a governance artifact.

**Relevance to Mos:** Helm umbrella charts are the closest analog to
Option A (umbrella repo). The key difference is that Helm's umbrella exists
for *deployment* composition, not *governance* composition. Mos's
umbrella would serve a different purpose (rule and contract authoring) but the
structural pattern — codeless parent repo with versioned references to children
— is identical.

### 3.3 Go Workspaces (`go.work`)

A workspace file listing multiple Go modules for unified building. `use`
directives point at local filesystem paths. `replace` directives override
module resolution for development.

**Cross-repo model:** Local development only. `go.work` is not checked in to
repositories (by convention). It solves "build these modules together" not
"govern these modules together." There is no remote resolution — all paths
must be local.

**Relevance to Mos:** `go.work` is the closest analog to Option D
(product manifest). The concept of a workspace file listing related modules
maps directly to a product manifest listing component repos. The key insight
is that `go.work` is an *offline* tool — it does not fetch anything, it just
redirects resolution to local paths. Mos's cross-repo linting could
adopt the same pattern: a local workspace override for offline/CI use, with
remote resolution as a convenience for development.

### 3.4 Bazel WORKSPACE / MODULE.bazel

`WORKSPACE` file declares external repositories via `http_archive`,
`git_repository` rules. Cross-repo references use `@repo//package:target`
syntax. The newer `MODULE.bazel` (bzlmod) adds versioned dependency resolution
with a central registry.

**Cross-repo model:** The most explicit of all systems. Every external
dependency is declared with a URL and hash. Cross-repo addressing
(`@repo//path:target`) is unambiguous. But it requires a centralized registry
for version resolution (Bazel Central Registry) and the `WORKSPACE` file
becomes enormous for large projects.

**Relevance to Mos:** Bazel's `@repo//path:target` syntax is directly
analogous to `git://org/repo@ref:path` for cross-repo includes (Option E).
The key lesson is that unambiguous cross-repo addressing works well but needs
an offline fallback (Bazel uses a local repository cache). Mos should
adopt a similar pattern: `git://` URLs for remote resolution, local path
overrides for CI.

### 3.5 Nix Flakes

`flake.nix` declares `inputs` pointing at other flakes (URLs, GitHub refs, or
local paths). Inputs are hash-locked in `flake.lock`. The `follows` keyword
deduplicates transitive inputs ("input A should use the same nixpkgs as
input B").

**Cross-repo model:** Explicit, hash-locked inputs with deduplication via
`follows`. Every dependency is reproducible. The lock file is verbose but
ensures bit-for-bit consistency. Updates are explicit (`nix flake update`).

**Relevance to Mos:** Nix's `follows` concept is relevant when
multiple repos in a product should use the same upstream mos. Without
`follows`, each repo independently resolves its upstream, potentially at
different commits. With `follows`, a product manifest could declare "all
component repos follow the same upstream at the same ref." This is not needed
for the initial implementation but is a useful future extension.

### 3.6 Blockchain Consensus (Bitcoin, Ethereum, PBFT)

Peer-to-peer networks where nodes reach agreement on shared state without a
central authority. Conflict resolution is algorithmic: longest chain (PoW),
highest stake (PoS), or 2/3+ quorum (PBFT). Every node validates every
transaction. Authority emerges from consensus rules, not from hierarchy.

**Cross-repo model:** True peer federation with decentralized authority. Nodes
are equal. Conflicts are resolved by protocol-defined consensus (not human
decision). State is fully replicated — every node holds the complete ledger.

**Relevance to Mos:** Blockchain proves that peer coordination *can*
work at scale without hierarchy. However, the lessons cut both ways:

1. **O(N) validation cost.** Every node must validate every transaction. In
   Mos terms: every repo's linter would need to clone and validate
   every peer's governance state. This is exactly the scalability problem
   identified in Option B's "linter must clone N peers" concern.
2. **State replication, not state divergence.** Blockchain works because all
   nodes converge on the *same* state (one ledger). Mos repos have
   *intentionally divergent* states — different repos have different rules for
   different code. There is no single truth to converge on.
3. **Consensus cost.** Bitcoin processes ~7 TPS globally. Ethereum ~30 TPS.
   The cost of decentralized consensus is throughput. For Mos, where
   linting must be fast and local, this cost model is prohibitive.
4. **Conflict resolution is mechanical.** Blockchain resolves forks by longest
   chain or most stake — deterministic, no human judgment. Governance rule
   conflicts in Mos require human judgment ("should this API be
   deprecated?"). Algorithmic consensus doesn't apply.

The existing art is real, but it reinforces rather than resolves Option B's
scalability concerns: achieving peer consensus requires either full state
replication (expensive) or accepting eventual consistency (governance cannot
be "eventually consistent" — a rule either applies or it doesn't).

### 3.7 BitTorrent / DHT (Distributed Hash Tables)

Peer-to-peer content distribution via swarm coordination. Trackerless
BitTorrent uses Distributed Hash Tables (Kademlia) for peer discovery.
Content is addressed by infohash (SHA-1 of the torrent metadata). Nodes
join and leave freely; the swarm is self-healing.

**Cross-repo model:** Decentralized peer discovery and content distribution.
No authority over content — any node can seed any data. Coordination is
about availability and routing, not about validity or governance.

**Relevance to Mos:** DHT solves a different problem than governance.
It could inform *peer discovery* (how does a repo find its siblings in a
product?) but not *peer governance* (who decides what rules apply?). Specific
lessons:

1. **Content-addressing works.** Identifying artifacts by content hash rather
   than location is robust. Mos already uses git's content-addressed
   object store, so this lesson is already absorbed.
2. **Discovery ≠ authority.** Finding peers is easy; deciding who gets to make
   rules is hard. BitTorrent is deliberately authority-free — any seeder is
   equal. Mos requires *directed* authority (some rules override
   others, some repos inherit from others). Peer federation without directed
   authority is how you get "who wins on conflict?" — the unsolved problem in
   Option B.
3. **Swarm resilience is irrelevant.** BitTorrent's self-healing swarm handles
   node churn gracefully. Mos repos don't churn — they are stable Git
   repositories. The problem is not "what if a repo disappears?" but "who
   decides the rules?"

### Summary Table

| System | Cross-repo model | Offline? | Authority model | Scale |
|--------|-----------------|----------|-----------------|-------|
| K8s OWNERS | None (per-repo only) | Yes | Per-directory cascade | 2000+ files |
| Helm umbrella | Versioned deps | Yes (after fetch) | Umbrella is parent | Medium |
| go.work | Local paths only | Yes | None (dev tool) | Small |
| Bazel | @repo//path:target | Yes (after fetch) | WORKSPACE declares all | Large |
| Nix flakes | Hash-locked inputs | Yes (after fetch) | flake.nix declares all | Large |
| Blockchain | Full state replication + consensus | No (online consensus) | Algorithmic (PoW/PoS/PBFT) | Large (but O(N) per tx) |
| BitTorrent/DHT | Content-addressed swarm | Yes (after discovery) | None (authority-free) | Very large |

---

## 4. Evaluation Matrix (5 × 10)

Scale: **Strong** (++), **Good** (+), **Neutral** (○), **Weak** (−), **Poor** (−−)

| # | Dimension | A (Umbrella) | B (Peer Fed) | C (Mono) | D (Manifest) | E (Consumer) |
|---|-----------|:---:|:---:|:---:|:---:|:---:|
| 1 | **Repo count** | −− (5: adds 1 codeless repo) | ○ (4) | + (4, only 3 need .mos/) | ○ (4, 1 hosts manifest) | + (3 with .mos/) |
| 2 | **New DSL primitives** | ++ (none) | −− (`cross_repo` block) | ++ (none) | − (`product` artifact) | ○ (cross-repo `include` path) |
| 3 | **Linter complexity** | + (each repo lints locally, upstream via git) | −− (must clone N peers) | + (upstream via git, 1 fetch) | − (must clone all component repos) | ○ (must resolve cross-repo includes) |
| 4 | **Agent bootstrapping** | − (5 repos, sequencing: umbrella first) | ○ (4 repos, no sequencing) | + (init primary, then 2 downstream) | ○ (init primary with manifest, coordinate) | + (init 3 product repos only) |
| 5 | **Authority model** | + (umbrella is authority) | −− (no single authority) | ++ (operator repo is authority) | + (manifest declares roles explicitly) | + (product repos own governance) |
| 6 | **Dunix scale** | − (~20 umbrella repos for 100 repos) | −− (N×M peer refs, combinatorial) | ++ (each product has 1 primary) | + (manifests scale linearly) | + (test repos excluded) |
| 7 | **Existing art** | + (Helm umbrella charts) | ○ (blockchain consensus, BitTorrent/DHT — precedent exists but lessons reinforce scalability concerns) | + (K8s OWNERS per-repo primary) | + (go.work, Bazel WORKSPACE) | + (Bazel @repo// syntax) |
| 8 | **Drift detection** | + (upstream ref pinning) | −− (bidirectional sync complexity) | + (upstream ref pinning) | ○ (manifest lists versions) | − (cross-repo includes can drift) |
| 9 | **Organizational fit** | + (separate repo respects org boundaries) | − (cross-org peers = shared authority) | ○ (one org "owns" others) | + (manifest declares org roles) | + (test repo stays independent) |
| 10 | **ALM metadata fit** | + (vocabulary in umbrella, inherited) | ○ (each repo defines own vocabulary) | ++ (vocabulary in primary, inherited) | ++ (manifest maps ALM tags to repos) | ○ (limited to product repos) |

### Score summary

| Option | ++ | + | ○ | − | −− | Net |
|--------|:--:|:-:|:-:|:-:|:--:|:---:|
| **A (Umbrella)** | 1 | 5 | 0 | 3 | 1 | +2 |
| **B (Peer Fed)** | 0 | 0 | 4 | 1 | 5 | −11 |
| **C (Mono)** | 4 | 3 | 1 | 0 | 0 | +13* |
| **D (Manifest)** | 1 | 3 | 3 | 2 | 0 | +4 |
| **E (Consumer)** | 0 | 4 | 4 | 1 | 0 | +3 |

\* Weighted: ++ = +2, + = +1, ○ = 0, − = −1, −− = −2.

**Option C (Mono-mos) is the clear winner** on the evaluation matrix,
with the highest score across all dimensions. Option B (Peer Federation) is
the lowest-scoring option. While peer federation has real existing art
(blockchain consensus, BitTorrent/DHT), those systems solve state replication
and content distribution — not directed governance authority over divergent
codebases. The precedent exists but the lessons reinforce the scalability
concerns rather than resolve them.

---

## 5. Dunix Scale Test (Thought Experiment)

**Scenario:** 100+ repos across Terra and Mars orgs. 2 human-protocol
languages (English, Martian via vocabulary keywords). ~20 products with an
average of 5 repos each. Some products span both orgs. Test suites may live
in dedicated repos or inline.

### Option A at Dunix Scale

20 products → 20 umbrella repos → 120 total repos. Each umbrella maintains
product-level rules and vocabulary (2 language variants). The umbrella repos
have no code — pure governance. In a 100+ repo GitHub org, 20 codeless
governance repos add noise to repo listings, CI dashboards, and developer
mental models.

**Verdict:** Functional but noisy. Governance overhead scales linearly with
products.

### Option B at Dunix Scale

100 repos × ~4 peers each = ~400 `cross_repo` references. The linter must
clone 4 repos per lint run. Cross-language peers (English/Martian vocabulary)
add conflict: which vocabulary wins when two peers define different terms?
Bidirectional sync across orgs requires write access that cross-org repos
may not grant.

**Verdict: Breaks at scale.** Combinatorial peer references, vocabulary
conflicts, and cross-org write access make this unworkable. Blockchain solves
peer consensus at scale but only by replicating the *same* state to all nodes
— Mos repos have intentionally *different* states (different rules for
different code), so the blockchain model does not transfer.

### Option C at Dunix Scale

20 products, each with 1 primary and ~4 downstream repos. Downstream repos
have minimal `.mos/` (just `config.mos` with `upstream` pointer). Linter
resolves upstream via git (one fetch per lint). Vocabulary inheritance flows
top-down — Martian teams use Martian keywords in their primary repo, English
teams use English keywords. No cross-repo vocabulary conflicts.

The 7-tier Red Hat hierarchy in `config.mos` already demonstrates this model
at depth 7. Extending to 100+ repos adds width, not depth — the model handles
both dimensions.

**Verdict: Works at scale.** Linear growth, no combinatorial explosion, clean
vocabulary inheritance.

### Option D at Dunix Scale

20 product manifests across 100 repos. Each manifest lists ~5 component repos.
Linter must clone all repos listed in a manifest to validate cross-repo
references. At 5 repos per product: 100 clones for full validation across all
products. Each product-level lint: 5 clones.

Without an offline fallback (local workspace paths), CI is slow and fragile
(network-dependent). With a `go.work`-style local override, it works but
requires developers to maintain workspace files.

**Verdict: Marginal.** Works with offline fallback but linting cost grows
linearly with product size.

### Option E at Dunix Scale

80 product repos with `.mos/`, 20 test repos without. Each product has 1–3
cross-repo includes pointing at test specs. Total: ~80–240 cross-repo include
paths. Linter resolves each include via git clone or local path.

**Verdict: Marginal.** Similar to D in linting cost. Works with caching or
local overrides.

### Scale Summary

| Option | 100+ repos? | Failure mode |
|--------|:-----------:|-------------|
| A | ○ (functional but noisy) | Governance repo proliferation |
| B | ✗ (breaks) | Combinatorial peer refs, vocabulary conflicts |
| C | ✓ (works) | None identified |
| D | ○ (marginal) | Linting cost without offline fallback |
| E | ○ (marginal) | Linting cost without offline fallback |

---

## 6. Hybrid Options

### 6.1 Hybrid C+E: Mono-Mos with Cross-Repo Spec Includes

Combines Option C's clear authority model with Option E's clean test-repo
separation.

**How it works:**

1. `ptp-operator` owns the product-level `.mos/` (Option C).
2. `linuxptp-daemon` and `cloud-event-proxy` point `upstream` at
   `ptp-operator` (existing model, zero new primitives).
3. `eco-gotests` has **no `.mos/`** for PTP governance.
4. Product contracts in `ptp-operator` reference test specs via cross-repo
   include paths in `spec { include "git://..." }` (Option E).

**Layout:**
```
ptp-operator/.mos/
├── config.mos                    # domain root
├── vocabulary/default.mos        # product vocabulary (ALM terms go here)
├── resolution/layers.mos
├── rules/
│   ├── mechanical/ptp-api-compat.mos
│   └── mechanical/ptp-event-schema.mos
└── contracts/
    └── active/CON-PTP-E2E/
        └── contract.mos
            # spec {
            #   include "git://rh-ecosystem-edge/eco-gotests@main:tests/cnf/core/network/ptp/..."
            # }

linuxptp-daemon/.mos/
└── config.mos                    # upstream -> ptp-operator

cloud-event-proxy/.mos/
└── config.mos                    # upstream -> ptp-operator

eco-gotests/
  (no .mos/ for PTP governance)
```

**DSL content — CON-PTP-E2E contract:**
```
contract "CON-PTP-E2E" {
  title  = "PTP E2E Validation"
  status = "active"
  goal   = "All PTP acceptance criteria pass in E2E."

  scope {
    rules = ["ptp-api-compat", "ptp-event-schema"]
  }

  tracker "eco-gotests" {
    repo = "rh-ecosystem-edge/eco-gotests"
    path = "tests/cnf/core/network/ptp/"
    sync = "unidirectional"
  }

  spec {
    include "git://rh-ecosystem-edge/eco-gotests@v4.18:tests/cnf/core/network/ptp/ptp_suite_test.go"
  }

  feature "Clock synchronization" {
    scenario "PTP clock converges within SLA" {
      sut  = "openshift/ptp-operator:pkg/controller/"
      test = "git://rh-ecosystem-edge/eco-gotests@v4.18:tests/cnf/core/network/ptp/tests/ptp_test.go"
      case = "TestPTP/Clock synchronization converges within 2s"
      given { PTP operator is deployed on OCP cluster with supported NIC }
      when  { ptp4l starts on worker nodes }
      then  { clock offset stabilizes below 100ns within 2 seconds }
    }
  }
}
```

**Evaluation against the 10 dimensions:**

| Dimension | Score | Rationale |
|-----------|:-----:|-----------|
| Repo count | + | 4 repos, 3 with .mos/ |
| New DSL primitives | + | Only `git://` URL resolver for existing `include` |
| Linter complexity | + | Upstream via git (cheap), cross-repo includes on demand |
| Agent bootstrapping | ++ | Init primary, 2 downstream `mos init`, no test repo setup |
| Authority model | ++ | Single authority (operator repo) |
| Dunix scale | ++ | Primary per product, upstream chains, test repos excluded |
| Existing art | + | K8s OWNERS (primary repo) + Bazel @repo// (cross-repo ref) |
| Drift detection | + | Upstream ref pinning + include version pinning (`@v4.18`) |
| Organizational fit | + | Test repo stays independent, cross-org via upstream |
| ALM metadata fit | ++ | Vocabulary in primary inherited by all components |

**Net score: +15** (vs C's +13, the next best).

### 6.2 Hybrid A+D: Umbrella with Product Manifest

The umbrella repo hosts a `product` manifest declaring all component repos.
This merges two options but adds no value over pure A — the umbrella IS the
manifest. The extra `product.mos` artifact in the umbrella just formalizes
what the repo's existence already implies.

**Verdict:** Redundant. No benefit over pure Option A.

### 6.3 Hybrid C+D: Mono-Mos with Product Manifest

The operator repo hosts a `product` block declaring component repos and test
suites. Downstream repos point upstream. The manifest makes the product graph
explicit and discoverable in one place.

**How it differs from pure C:** Without the manifest, discovering the product
graph requires traversing all repos to find who points upstream at the
operator. With the manifest, the graph is declared in one file.

**How it differs from C+E:** C+D uses the manifest for graph discovery; C+E
uses cross-repo includes for spec linkage. C+E is more targeted — it solves
the split-brain problem (test specs) without requiring a full product manifest
(which is a broader, more complex feature).

**Verdict:** Valid but higher cost than C+E. The manifest requires a new
artifact type, grammar changes, and linter support for product graph
validation. C+E achieves the primary goal (split-brain resolution) with just
a URL resolver for an existing primitive.

---

## 7. Recommendation

### Ranked Options

| Rank | Option | Net Score | Verdict |
|:----:|--------|:---------:|---------|
| 1 | **C+E (Mono + Consumer)** | +15 | **Recommended** |
| 2 | C (Mono-mos) | +13 | Strong fallback (no cross-repo includes needed) |
| 3 | D (Product manifest) | +4 | Future enhancement, not first iteration |
| 4 | A (Umbrella repo) | +2 | Viable for large orgs, excessive for small products |
| 5 | E (Test consumer only) | +3 | Incomplete without C's authority model |
| 6 | B (Peer federation) | −11 | Rejected: no directed authority, breaks at scale despite blockchain/DHT precedent |

### Recommended Approach: Hybrid C+E

**Mono-mos in the primary repo (C) with cross-repo spec includes to
test repos (E).**

This approach:

1. **Uses zero new grammar elements.** The `upstream` block, `spec { include }`,
   and `tracker` adapter already exist.
2. **Adds one new capability:** A `git://` URL resolver for the `include`
   directive, extending an existing primitive rather than creating a new one.
3. **Respects primitives-not-prescriptions.** No new artifact types, no new
   keywords. ALM metadata goes in `vocabulary/default.mos`. Product structure
   emerges from `upstream` chains, not from a declared manifest.
4. **Scales.** Each product has one primary repo. Downstream repos have minimal
   `.mos/`. Test repos have no `.mos/`. The graph grows linearly.

### Required DSL/Linter Capabilities

| Capability | Type | Description |
|-----------|------|-------------|
| **Cross-repo include resolver** | Linter enhancement | Resolve `git://org/repo@ref:path` URLs in `spec { include }` and `scenario { test }` fields. Fetch via git clone (cached). |
| **Offline workspace override** | Configuration | A `workspace.mos` or CLI flag that maps `git://` URLs to local filesystem paths for CI/offline use. Analogous to `go.work` replace directives. |
| **Include version pinning** | Convention | Encourage `@tag` or `@sha` in cross-repo include URLs. The linter warns on `@branch` (mutable ref) and passes on `@tag` or `@sha` (immutable). |

### What Is NOT Needed

- No `cross_repo` primitive (Option B).
- No umbrella repos (Option A).
- No `product` artifact type (Option D) — the product graph is implicit in
  the `upstream` chain. If explicitness is needed later, a `product` block
  can be added as a vocabulary-driven regular block without grammar changes.

---

## 8. Impact Assessment

### CON-2026-015: Enhanced Contract Model + Migration

The cross-repo include resolver is **not needed** for migrating Mos's
own contracts (all local). The recommended approach does not inject new
requirements into CON-2026-015's Phase 1 (building block enhancements) or
Phase 2 (contract migration).

**Optional addition:** The `product` block concept from Option D could be
mentioned in the vocabulary template expansion task as a future vocabulary
term, not a structural primitive. This is informational only.

**Verdict: No injection needed.**

### CON-2026-013: Harness Bridge

Multi-repo governance affects harness execution when a harness command
references code in another repo (e.g., running Ginkgo tests from eco-gotests
as part of a ptp-operator contract). For Mos's own dog-fooding
(CON-2026-013's scope), all harness commands are local. Cross-repo harness
execution is a future concern for adopters, not for CON-2026-013.

**Verdict: No injection needed.**

### New Contracts Needed

One new implementation contract is recommended:

**CON-2026-016: Cross-Repo Include Resolver**

| Field | Value |
|-------|-------|
| Goal | `spec { include }` and `scenario { test }` support `git://org/repo@ref:path` URLs |
| Depends on | CON-2026-015 (contract model enhancements must land first) |
| Scope | Linter URL resolver, offline workspace override, ref pinning lint rule |
| Primitives | `git://` URL scheme for `include` paths, `workspace.mos` for offline overrides |
| Not in scope | `product` artifact type, `cross_repo` block, umbrella repos |

This contract directly implements the recommended capability from this
research. It should be created after CON-2026-015 is complete, as the enhanced
contract model (with `spec { include }` fully wired) is a prerequisite.

A second contract for the `product` manifest (Option D) is **deferred
indefinitely.** The product graph is discoverable via `upstream` chains, and
explicitness can be added later via vocabulary without grammar changes.

---

## 9. Design Principles Applied

Throughout this analysis, the following Mos design principles were
applied as evaluation filters:

### Primitives, not prescriptions

Every capability must compose from existing building blocks. The recommended
approach adds exactly one new primitive (URL resolver for `include`) and
composes everything else from existing `upstream`, `tracker`, `vocabulary`,
and `spec { include }` primitives.

Options rejected on this principle:
- **B (Peer Federation):** Requires a new `cross_repo` block that duplicates
  what `upstream` already provides hierarchically. Blockchain and BitTorrent
  prove peer coordination works, but for *replicated* state (one ledger) or
  *content-neutral* distribution — not for directed governance over divergent
  codebases where rules intentionally differ per repo.
- **D (Product Manifest):** Requires a new `product` artifact type that could
  be expressed as a vocabulary-driven block instead.

### ALM-as-vocabulary

ALM metadata (product, component, subcomponent, test pillar) enters
Mos via `vocabulary/default.mos` terms and open fields on scenarios,
not via schema-enforced keywords. The recommended approach inherits vocabulary
from the primary repo to all downstream components, ensuring consistent
ALM terminology across the product graph without hardcoding any ALM vendor's
taxonomy.

### Upstream federation is recursive, not tier-locked

The existing `upstream` model already supports multi-repo governance for the
component-to-product relationship. It handles the 7-tier Red Hat hierarchy.
It handles cross-org via multiple `upstream` blocks (DAG). The only gap is
the test-repo split brain, which cross-repo `include` fills.

---

## Appendix A: Glossary

| Term | Definition |
|------|-----------|
| **Split brain** | When governance authority for a single product is fragmented across repos with no clear primary. Specifically: when acceptance criteria (product concern) and test implementations (test concern) cannot be traced to a single source of truth. |
| **Peer federation** | A governance model where repos have lateral (non-hierarchical) relationships. Contrasted with upstream federation (hierarchical). |
| **Product manifest** | A declarative artifact listing all repositories that constitute a product, their roles, and their test suites. Analogous to `go.work` or Helm umbrella `Chart.yaml`. |
| **Mono-mos** | A governance model where one repo in a product group owns the authoritative `.mos/` and other repos inherit from it via `upstream`. |
| **Cross-repo include** | An `include` directive in `spec {}` that references a file in another repository via a `git://` URL. |
| **Offline workspace override** | A local configuration that maps `git://` URLs to filesystem paths, enabling linting without network access. |

## Appendix B: Decision Record

| Date | Decision | Rationale |
|------|----------|-----------|
| 2026-03-02 | Recommend C+E over pure C | Pure C leaves the split-brain unsolved; cross-repo includes (E) close the gap |
| 2026-03-02 | Reject B (peer federation) | No directed authority model, breaks at scale. Blockchain/DHT precedent exists but solves state replication, not directed governance over divergent codebases |
| 2026-03-02 | Defer D (product manifest) | Higher implementation cost, product graph is implicit in upstream chains |
| 2026-03-02 | Propose CON-2026-016 | Cross-repo include resolver is the only new capability needed |
| 2026-03-02 | No injection into CON-2026-013/015 | Dog-fooding scope is local; cross-repo is future adopter concern |
