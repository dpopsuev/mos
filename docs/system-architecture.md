# System Architecture

> Independent modules connected by standardized interfaces, producing a coordinated result greater than any module alone.

---

## 1. Design Philosophy

The system is multi-component and highly decoupled. The structural analogy is a modular synthesizer: independent modules -- each with a defined input, a defined output, and a standardized interface that lets any module connect to any other. The operational analogy is Kubernetes: controllers that watch state, detect drift, and reconcile. Both describe the same architecture from different angles: the synthesizer describes signal flow, Kubernetes describes the control loop.

The core is a small set of modules that together produce the full signal chain -- from raw agent output through filtering, shaping, routing, and verification to final externalized result. The core supports the five lifecycle phases (Sensation, Contextualization, Abstraction, Materialization, Externalization). Everything beyond the core is optional, community-extensible, and replaceable -- consistent with the Ship of Theseus principle.

---

## 2. Core Modules

The **min-max principle**: the smallest set of modules that produces the full signal chain. Each module exists because a specific problem demands it.

### Mos (State Store)

The state store for the entire system. Reads and writes the `.mos` format -- the persistent configuration that encodes the project's governance, rules, contracts, and resolution layers. Responsible for:

- Parsing and rendering resolution layers (Organization, Group, Domain, Individual)
- Tracking Desired State vs. Concrete State and computing drift
- Executing the onboarding wizard (first-run setup)
- Managing the hybrid layer when `.git` and `.mos` coexist
- Exposing the mosal state to other modules through a defined API

The primary interface is a CLI. Agents interact with Mos through the same CLI commands that humans use -- there is no privileged agent channel. An agent calling `mos contract create` goes through the same code path as a human typing the same command. The CLI is designed to be both human-friendly (readable output, contextual help) and machine-parseable (structured output modes, exit codes that encode status).

Implementation language: Go. The package architecture follows Go's standard layout: `moslib/` (library package, no I/O), `cmd/mos/` (CLI binary). The `.mos` format uses a purpose-built DSL where Gherkin keywords (Given/When/Then) are native syntax, not embedded strings.

### Macro (Editor)

The operator's primary interface. Forked from the Micro terminal editor and extended for agentic workflows. Macro connects to a **Macro Server** (headless daemon) that manages sessions, buffers, agents, and circuit execution. If no Macro Server exists on the path, Macro spawns one.

Responsibilities:

- Rendering the resolution map at the zoom level the operator selects
- Visualizing agent activity, locked zones, Contract progress, and drift indicators
- Providing controls for Bill introduction, Contract ratification, and governance operations
- Supporting Legacy Mode (traditional code editing) behind a lens toggle
- Multi-user session awareness (who is working where, what is locked)

### Origami (Router)

The signal routing layer for agent execution. Origami is a Go library (`github.com/dpopsuev/origami`) providing graph-based agentic circuit orchestration:

- **Circuit definition** via a YAML DSL that compiles into a typed graph (Nodes, Edges, Graphs)
- **Walker-based execution** where Walkers traverse the graph, carrying context and triggering Extractors at each node
- **Observation** via signal trace and monitoring for real-time visibility
- **Concurrent execution** for multi-agent orchestration where multiple Walkers propagate simultaneously

Origami is a zero-domain-import framework: all domain logic lives in the consumer, not in Origami itself. Consumers define the circuit topologies; Origami routes the signals.

### Sophia (Context Cache)

The gravitational graph database for agentic memory. Sophia provides the pre-execution context pitstop -- local short-term memory with natural decay:

- Stores knowledge nodes (rules, code summaries, domain concepts, decision records) with mass and semantic edges
- Organizes nodes into proximity layers (Now, Near, Close, Far) based on recency and relevance
- Promotes relevant nodes toward Now when queried; compresses dormant nodes toward Far
- Integrates with Origami as a mandatory pre-execution step

Sophia operates alongside **Monad** (the Archive), a remote, collective-scale gravitational graph database that serves as the long-term knowledge store. Sophia is the L1 cache (local, fast, per-developer); Monad is the L2/L3 (remote, shared, collective). Sophia evicts cold nodes to Monad; Sophia queries Monad on cache miss. Both are built on **Universalis** -- a shared Rust library that implements the gravitational graph engine. Sophia and Monad are Rust binaries embedding Universalis; they expose gRPC APIs consumed by the Go components.

### Harness (Verification Runtime)

Verifies and shapes agent output against the specification. The Harness is the judiciary in executable form -- subtractive by nature: start with rich agent output, remove what does not conform:

- **Gates (Errors)**: compilation, tests, CI results, type checks, linting. Hard cutoff -- binary pass/fail. Non-negotiable.
- **Advisories (Warnings)**: domain constraints, goal statements, specifications, architectural principles. Surfaced for human judgment.
- **Continuous evaluation**: the Harness runs during Contract execution, not just at the end. Drift, rule violations, and boundary crossings are flagged as they occur.
- **Rule resolution**: for any given Contract, the Harness computes the applicable rule set -- standing law plus temporary amendments minus suspended rules -- and evaluates against it.

### Embassy (External Adapters)

The adapter layer for bidirectional integration with external systems -- the I/O boundary:

- **Ideal Plane adapters**: issue trackers, project management tools, roadmaps, OKRs, spec documents. Inbound: materialize external planning artifacts as Bills or context. Outbound: propagate mosal state changes back to external systems.
- **Concrete Plane adapters**: CI/CD pipelines, infrastructure platforms, monitoring systems, observability stacks. Inbound: reflect runtime state into the Map. Outbound: trigger deployments, report pipeline outcomes.
- **Embassy protocol**: a defined interface that each adapter implements. Adapters are replaceable modules -- swap one issue tracker Embassy for another without touching the core.

---

## 3. The Signal Chain

The power of a modular system comes not from individual modules but from how they are connected. The signal chain -- the path from agent output through verification, routing, and enrichment to final result -- determines the character of the coordination.

### Module Specialization

Like Kubernetes controllers each specialized for a specific resource type, agent types are specialized for different concerns:

- **Sensory modules** detect change. They watch the filesystem, monitor Embassy feeds, scan for drift between Desired and Concrete state, and fire signals when something deviates.
- **Librarian modules** maintain context integrity. They organize knowledge in Sophia, update code summaries, keep the Map current, and ensure domain context is accurate.
- **Controllers** (invariant enforcers) continuously verify that the system satisfies its declared invariants and resists violations of them. QA enforces functional invariants; security enforces trust invariants. Each controller develops stickiness in its domain through its domain Sophia. Every controller operates in both Blue Team and Red Team modes.
- **Executor modules** carry out ratified Contracts. They write code, generate documentation, run transformations, and produce the artifacts that materialize the Desired State.
- **Reporter modules** surface outcomes. They compile execution reports, generate drift assessments, and prepare evidence for Bills.

### Origami as Router

Origami's graph model IS the routing layer:

- **Nodes** are the processing units -- each typed, each specialized, each operating within a defined domain.
- **Edges** are the signal pathways -- typed connections that carry artifacts, events, and control signals between nodes.
- **Circuits** are functional groupings -- nodes that process signals together within a defined topology.
- **Walkers** are the messages themselves -- they traverse the graph from node to node, carrying context and triggering execution at each stop.

The pipeline is not a linear sequence. It is a graph -- signals can fan out, converge, loop, and branch based on the topology of the problem.

### The Controller Bank

The graph has a specific shape: concentric loops with radial spokes. In synthesizer terms, an oscillator bank. In Kubernetes terms, a set of specialized controllers each running their own reconciliation loop.

**The Reconciler (Central Loop).** The innermost ring. It continuously compares Desired State against Concrete State, detects drift, and triggers corrective signals. This is the Kubernetes reconciliation loop applied to the entire project.

**Controller Loops.** Each controller type runs its own continuous loop:

- **Security controller** -- monitors trust boundaries, validates compliance, scans for violations.
- **Architecture controller** -- monitors structural boundaries, dependency graphs, boundary crossings.
- **Harness controller** -- evaluates mechanical rules (ROGBY, compilation, tests).
- **Drift controller** -- reconciles documentation vs code, specification vs implementation.

Each loop runs semi-independently. The Reconciler coordinates but does not micromanage. A controller can flag a violation and act on it without waiting for the Reconciler.

**Why multiple controllers, not one.** Stickiness. A single omnibus controller has no domain specialization -- it starts cold on every concern. Multiple controllers, each dedicated to one enforcement domain, develop the equivalent of muscle memory. The security controller knows the trust boundaries. The architecture controller knows the structural invariants. They carry context between cycles via their domain Sophia.

**Topology:**

- **Inner ring:** the Reconciler -- central loop (Desired State vs Concrete State).
- **Middle rings:** controller loops (security, architecture, harness, drift).
- **Radial spokes:** connections outward to Embassies, tooling, external systems.

### Blue Team / Red Team Verification Axis

Controller loops are the primary organizational axis -- each loop develops stickiness in its domain. Orthogonal to the domain axis is the verification mode: every controller operates in both **Blue Team** (benevolent, positive space) and **Red Team** (adversarial, negative space) modes.

Each controller applies both modes to its own invariants:

- **Security controller** -- Blue: "TLS is configured, tokens expire, RBAC rules hold." Red: "Can I bypass authentication through a race condition?"
- **Architecture controller** -- Blue: "Dependency graph conforms to declared boundaries." Red: "Can I smuggle an import across a boundary through a transitive dependency?"
- **Harness controller** -- Blue: "Tests pass, types check, linting clean." Red: "Can code pass all mechanical checks but violate the behavioral specification?"
- **Drift controller** -- Blue: "Documentation matches code." Red: "Can the documentation look correct while the implementation has semantically diverged?"

---

## 4. Signal Flow Dynamics

When do modules fire? The system needs the same property as a well-designed synthesizer: modules should not require explicit dispatch, nor should they fire randomly and waste tokens. They should activate proportionally to demand and fall silent during equilibrium.

Five mechanisms govern the signal flow:

**The Heartbeat.** A minimal, always-running probe. It fires on a fixed interval regardless of system state. Its only job: check if drift has accumulated, if pending signals have gone unhandled, if housekeeping is due. Near-zero cost.

**The Governor.** The system-wide resource scaler. It has two modes: **tonic** -- low, steady baseline allocation that maintains readiness -- and **phasic** -- burst allocation in response to salient input that globally increases processing capacity. The Governor does not decide *what* to process; it decides *how much processing power to allocate*.

**The Guard.** The default state of every controller is **guarded** -- loaded, warm, context connected via its domain Sophia, but not executing. A signal in the relevant domain opens the guard. No dispatcher is needed. The signal itself qualifies.

**The Run Loop.** Controller loops are autonomous cyclic execution that handles enforcement once a direction is set. The operator does not command each enforcement step. They set intent and the run loops handle the rhythm autonomously.

**Housekeeping.** Idle is not zero activity. When the system reaches equilibrium, the Governor switches to tonic and the system enters housekeeping. Librarian modules perform Sophia maintenance: strengthening frequently-accessed edges, evicting cooling nodes toward Monad, reorganizing neighborhoods.

### Operational States

- **Idle.** Heartbeat ticks. Governor tonic. All controllers guarded. Librarians in housekeeping. Near-zero token cost.
- **Reactive.** Heartbeat catches accumulated drift. Governor fires burst. Relevant guard opens. Drift handled. Returns to idle.
- **Active.** Signal arrives (contract filed). Governor shifts to phasic. Relevant guards open. Run loops execute autonomously. Moderate token cost.
- **Peak.** Signal load high. Governor sustained phasic. Multiple guards open simultaneously. High concurrency -- token cost proportional to demand, never to idle time.

---

## 5. Ecosystem

Everything beyond the core is ecosystem. Like Eurorack modules from third-party manufacturers or Kubernetes operators from the community, ecosystem components extend the system without being required for basic operation:

- **Alternative displays** -- graph views, 3D visualizations, dashboard layouts for the Map
- **Additional Embassies** -- adapters for specific external systems (Jira, Linear, ArgoCD, Datadog)
- **Alternative clients** -- GUI clients, web clients, mobile clients that read `.mos` through the same protocol
- **Specialized harness modules** -- domain-specific verification (HIPAA compliance, OWASP scanning, performance budgets)
- **Domain-specific Monad instances** -- remote gravitational graph databases serving as collective knowledge archives

The ecosystem is unbounded by design. The core provides the interfaces; the community builds the modules.

---

## 6. Relationship to Mos

Macro is a client. Mos is the state store.

The architecture of the client does not constrain the `.mos` format. The format specification is defined by the Mos project, not by Macro. Any tool can implement its own client. The protocol is open; the implementation is replaceable.

If Macro is replaced tomorrow by a better client, the mosal state remains intact. The `.mos` files, the resolution layers, the governance model, the historical record, the Bills and Contracts and Declarations: all persist. Mos does not depend on Macro. Macro depends on Mos.

State stores outlive the clients built on top of them. Git has outlived every IDE generation since 2005. Mos is designed with the same permanence in mind.
