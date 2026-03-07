# Mos -- The Political Model

> Every project is a polity. The question is not whether it has a mos, but whether that mos is written down.

Version control systems outlive the editors built on top of them. Git has survived every generation of IDE since 2005. The editors change; the history does not. Mos is designed with the same permanence in mind. It is the everlasting layer -- the format, the protocol, the historical record -- that persists regardless of which client renders it. Macro is the first client; it will not be the last. Clients come and go. The mosal state endures.

---

## 1. The Lifecycle

All change follows the same arc. Five phases, each a narrowing of scope from sensation to artifact.

**Sensation.** A raw signal that the current state is sub-optimal -- the first twitch before any specification is written.

**Contextualization.** Drawing a boundary around the sensation: what matters, what does not, where the domain begins and ends.

**Abstraction.** The Platonic model -- a specification that exists in the realm of logic, independent of implementation.

**Materialization.** Theory meets friction. Implementation: where the specification collides with the digital world.

**Externalization.** Knowledge extraction -- from the creator's head into a persistent medium so the system can function without any single person.

The first three phases define the **Desired State** -- intent. The last two constitute the **Concrete State** -- reality. The gap between them is the fundamental tension the entire system exists to manage. Mos tracks the gap, governance determines who may close it, the Harness verifies it.

### Recursion

The lifecycle is not a single pass. The first application creates the outermost scope -- the project's Declaration (0 to 1). Within that scope, domains emerge, each running its own lifecycle instance (1 to many). Within each domain, components emerge, each running their own. Drift tracking operates at every level: a project can be aligned at the domain level while a specific component has drifted badly. Drift is not a global binary but a per-scope measurement at every level of the hierarchy.

### Sub-Resolution

Each phase decomposes into sub-steps. The phases do not change; the granularity increases. The cardinal rule: sub-steps through Abstraction produce decisions, not files. Code starts at Materialization.

- **1.1 Problem framing.** Context, pain points, stakeholders.
- **1.2 Goal statement.** Measurable success criteria and explicit non-goals.
- **2.1 Constraints and assumptions.** Security, scale, dependencies, budget, timeline.
- **2.2 Personas and domain research.** Who is affected, current workflows, friction points.
- **3.1 Behavior specification.** Given/When/Then acceptance criteria.
- **3.2 Solution framing.** Options, trade-offs, rationale. Rejected alternatives preserved as negative space.
- **3.3 Architecture.** Boundaries, data flows, trust zones.
- **3.4 Components and interfaces.** Responsibilities, contracts, failure modes.
- **4.1 Technology selection.** Languages, frameworks, infrastructure evaluated against constraints.
- **4.2 Implementation.** Code under Harness supervision -- gates verify compliance continuously.
- **5.1 Documentation.** API references, runbooks, architecture diagrams.
- **5.2 Knowledge capture.** Decision records, lessons learned, post-mortems.
- **5.3 Drift reconciliation.** Update the Desired State or flag intentional divergence as acknowledged debt.
- **5.4 Observability.** Metrics, traces, dashboards -- the system externalizes its own state.
- **5.5 Feedback loop.** What was learned feeds the next Sensation. The lifecycle restarts from accumulated state.

---

## 2. The Political Analogy

A software project governs itself whether it intends to or not. Somebody decides what gets built. Somebody decides what gets rejected. Somebody decides the rules agents follow. The act of creating a repository is the act of founding a state.

The analogy is not metaphorical -- it is structural:

- A **project** is a sovereign state. It has territory (codebase), citizens (contributors and agents), laws (rules, conventions, contracts), and a history of decisions that shaped its present form.
- A **mos** is the fundamental law of that state. It defines the vocabulary, the boundaries, the resolution layers, and the governance model that all participants -- human and agent -- operate within.
- An **organization** is a federation. Multiple project-states under the same umbrella share a common DNA -- inherited defaults, shared invariants, a body of "international law" -- while retaining sovereignty over their own internal affairs.

Today, every project already has a mos. The problem is that it is uncodified.

The distinction matters. The United Kingdom operates under an uncodified mos -- a patchwork of statutes, conventions, judicial precedents, and unwritten norms accumulated over centuries. It works because human actors share cultural context and institutional memory. The United States, by contrast, adopted a codified mos -- a single authoritative document that enumerates powers, defines structure, and provides an amendment process. It works because the rules are explicit, discoverable, and enforceable.

Software projects today follow the British model by accident. Rules live in README files, wiki pages, scattered `.mdc` files, Slack channels, and tribal knowledge. There is no single source of law. There is no enforcement mechanism. There is no amendment process. The mos exists, but nobody can point to it.

**Mos [Version Control System]** (`.mos`) moves projects from the British model to the American one -- not by imposing a specific government, but by providing the mechanism to codify whatever government a project chooses.

---

## 3. The Vision

Macro is an Agentic Age IDE built on these pillars:

- **Natural Language Interface.** For new projects: concrete discussion about the problem domain. For existing projects: an onboarding enabler. The system adapts to its user -- newcomer or veteran.
- **Architecture-First Visualization.** The system communicates health, not just syntax. Public and private interfaces, failure points, performance bottlenecks, and anti-patterns are visible -- the architecture, not just the text.
- **Signal Chain.** Agents activate in response to signals, route through Origami, and scale proportional to demand. During equilibrium the system idles at near-zero cost. Under load it scales to high concurrency. BYOB -- Bring Your Own Bot(s).
- **Legacy Mode.** A complete traditional code editing experience, available on demand behind a lens toggle.
- **Context Libraries.** Persistent, shareable context that outlives individual sessions. Portable between agents, projects, and environments. Today's `.cursor/` directories are the first evidence that project-bound context wants to be a first-class primitive. Mos liberates them.
- **Warm Start.** No developer, no agent, and no project begins from zero context. A new team member inherits the project's mosal state. A new agent inherits the domain's Sophia. A new project inherits its organization's DNA. Starting cold is a system failure.
- **Console First, Agent-Accelerated.** Every module exposes a CLI that humans can operate directly. There is no separate agent API -- the CLI is the API. The system is designed for agents but never dependent on them.
- **Free and Open Source.** Permissive license. Self-editable. Nothing hidden.
- **Multi-User by Default.** Concurrent sessions with visibility into who is working where and on what.

---

## 4. The Uncodified Problem (Case Study)

The evidence that projects already behave as self-governing states is sitting in plain sight. Three artifacts illustrate the pattern and its limits.

### kombucha.mdc -- The Mosal Convention

`kombucha.mdc` is a pre-prompt seed that generates an entire File-System Context (FSC) for a project: meta rules, developer flow, test matrices, work contracts, security analysis, skills, and save triggers. It is, in political terms, a mosal convention -- a one-time event that drafts the initial body of law for a new state.

The convention produces a scaffold. But the scaffold has no enforcement. The generated rules are advisory. An agent may read them, or it may not. There is no judiciary.

### asterisk/.cursor -- A Mature Uncodified Mos

The `.cursor/` directory in the asterisk project contains approximately 200 files organized across rules, contracts, documentation, notes, prompts, skills, glossary, strategy, tactics, taxonomy, security cases, goals, and configuration. It has:

- **Always-apply rules** (`project-standards.mdc`, `agent-operations.mdc`, `knowledge-store.mdc`, `rule-router.mdc`) that function as mosal amendments -- intended to be in force at all times.
- **Glob-scoped rules** that apply to specific file patterns, functioning as domain-specific statutes.
- **Agent-requestable rules** that agents can invoke on demand, functioning as case law -- precedent available when relevant.
- **Contracts** (`current-goal.mdc`) that define active work, functioning as legislation in progress.
- **Skills** (`asterisk-analyze`, `asterisk-calibrate`) that encode operational procedures, functioning as executive orders.
- **A glossary and taxonomy** that define the project's vocabulary.

This is a fully developed body of law. It is also completely uncodified in the mosal sense: there is no single format, no versioning of the rules themselves, no inheritance mechanism, and no enforcement. Rules can be added, modified, or ignored without ceremony.

### origami/.cursor -- A Leaner Variant

The origami project's `.cursor/` directory contains approximately 150 files following a similar pattern but with fewer always-apply rules (only `project-standards.mdc`) and different domain rules (`agent-bus.mdc`, `dsl-design-principles.mdc`). Both projects share many "universal rules" -- copied manually between repositories. This manual copying is proto-international-law: the same norms applied across multiple states, maintained through convention rather than a formal treaty.

### The Enforcement Gap

The critical weakness across all three cases is identical: rules can be skipped by agents. An always-apply rule is a suggestion backed by convention, not a constraint backed by machinery. The gap between intent (the rule exists) and reality (the agent ignored it) is the same gap that Mos is designed to close.

---

## 5. Resolution Layers as Map Data

Mos organizes the project-state at multiple resolutions, each carrying its own rules, vocabulary, and contracts:

**Organization** -- The federation level. Shared DNA across all projects under the same umbrella. Red Hat's upstream-first philosophy, safety invariants, multi-operator defaults -- these are the "international treaties" that propagate downward. An organization-level `.mos` defines what all member projects inherit by default.

**Group** -- The team or squad level. Shared conventions within a working group that may span multiple repositories. A platform team's API style guide, a frontend guild's component patterns -- norms that apply across a group's territory but not necessarily beyond it.

**Domain** -- The bounded context level. A service boundary, a module, a self-contained area of the codebase. Domain-level rules define the local law: naming conventions, dependency constraints, testing requirements, architectural invariants specific to that territory.

**Individual** -- The file or function level. The most granular resolution. A specific file can carry metadata about its purpose, its owner, its constraints, its relationship to other files. This is the atom -- the smallest unit Mos tracks.

Each layer is a data layer on the same map. The terrain (code) does not change. What changes is the resolution at which you view it and the rules that apply at that resolution. An organization-level rule cascades down unless overridden. A domain-level rule applies within its boundary. An individual-level annotation is hyperlocal.

The layers are not hierarchical in a strict inheritance sense -- they are composable. A domain can reference organization-level vocabulary while defining its own contracts. A group can inherit from the organization and add constraints that do not propagate to sibling groups.

### The Integrity Index

Security and testing are not phases in the lifecycle or lenses in the resolution model. They are continuous threads -- cross-cutting data layers that run through every resolution level, every lifecycle phase, and every module.

The deeper insight: security is not a separate discipline from quality assurance. Both are invariant enforcement. QA enforces functional invariants (the system does what it should). Security enforces trust invariants (the system resists what it shouldn't). The industry separates them artificially; the methodology is identical.

Trust boundaries and coverage are quantifiable -- they form the system's **Integrity Index**. The Index can be measured and visualized as an overlay on the Map at any zoom level. Areas with high coverage and well-defined trust boundaries are illuminated. Areas with untested paths, implicit trust, and unverified assumptions are fogged. Integrity is not binary. It is a number that trends over time: improving or degrading, consolidating or fragmenting.

---

## 6. Bills and Contracts

Two legal instruments exist in the real world and both have analogues in Mos. They are distinct and the distinction matters.

A **Bill** is a proposal for new law. In parliament, a Bill binds no one until it is enacted; it exists to be debated, amended, tested, and either ratified or rejected. A Bill is public -- visible to every participant in the project-state. Its purpose is to change the standing law: rules, vocabulary, architectural invariants, governance structures. A Bill that passes becomes part of the mos itself. A Bill that fails remains in the record as negative space.

A **Contract** is a binding agreement between specific parties about specific work. In law, a contract requires offer, acceptance, consideration, and mutual intent to be bound. It is private in the sense that it binds the signatories, not the entire jurisdiction. Its purpose is execution: who does what, under what rules, with what resources, toward what outcome.

In Mos, these two instruments fuse into a single entity with a dual nature. Every unit of work begins life as a Bill -- a public proposal -- and, upon ratification, becomes a binding Contract -- an agreement to execute. The standing rules it introduces during its active phase behave as temporary legislation. When the Contract completes, its code effects merge into the Concrete State and become permanent, while its temporary amendments expire.

The lifecycle maps to the stages a Bill passes through in a parliamentary system:

**First Reading** -- The Bill is introduced. A contributor or agent identifies that something needs to change and files a proposal. The proposal declares its intent, its scope, and the rules it will operate under. It is visible to all participants but binds no one. This is the sensation formalized: the raw signal that the current state is sub-optimal, placed on the record.

**Committee Stage** -- The Bill enters the Laboratory. A mosal fork is created: the proposed changes are run in a sandboxed environment where the harness evaluates their effects against the standing mos in real time. Rule compliance, drift from desired state, architectural boundary crossings, resolution-level impact -- all are observable. Evidence is gathered. The proposal can be amended based on what the lab reveals. This is pre-legislative scrutiny with teeth: not opinions about what might happen, but measured evidence of what does happen.

**Readings and Debate** -- Stakeholders -- human and agent -- deliberate on the proposal in light of the laboratory evidence. Amendments are proposed, contested, and resolved. The governance model of the project determines who participates and how decisions are reached: a BDFL decides alone, a committee votes, a consensus process seeks broad agreement.

**Ratification** -- The Bill becomes a binding Contract. The elements of a legal contract are now present: offer (the scope of work defined in the Bill), acceptance (ratification by the governance authority), consideration (the resources allocated -- agent time, compute, human attention), and mutual intent (all parties agree to the rules for the duration of the work). From this moment, the Contract binds its participants. Agents operating under the Contract are subject to its rules. The harness enforces them.

**Enactment** -- Active work under the Contract's terms. Agents execute. The operator monitors. The harness evaluates compliance continuously, not just at the end. The Contract's temporary amendments are in force for its scope and duration, overlaying the project's standing law.

**Completion** -- The work is done, verified against its harness, and merged into the Concrete State. The Contract moves to the historical record. Its code effects become permanent law. Its temporary amendments expire. This is royal assent: the moment a Bill's effects become part of the living mos.

**Abandonment** -- The Bill is rejected at any stage, or the Contract is voided before completion. The proposal, its laboratory evidence, its deliberation history, and the reasons for rejection all remain in the record. Abandoned Bills are not deleted. They are the mosal equivalent of case law -- precedent that shapes future proposals even though the original did not pass.

**Contract Rules** are temporary amendments specific to a single Bill. A Contract can introduce rules that apply only within its scope and duration -- overriding or extending the project's standing law for the lifetime of that piece of work. They function like riders or provisions attached to a parliamentary Bill: binding while the legislation is in force, expiring when it is repealed or fulfilled. This mechanism allows targeted deviation from standing law without permanently altering the mos.

The dual nature of this instrument -- Bill in its public, deliberative phase; Contract in its private, executive phase -- solves the problem of planning and execution being severed. It is not a ticket in an external system. It is a first-class entity within Mos that links sensation to proposal to evidence to deliberation to agreement to execution to verification to historical record.

---

## 7. Governance Models (User-Defined)

Mos provides the mechanism for self-governance. It does not prescribe the model. The choice of how a project governs itself belongs to the project's founders and contributors.

**Benevolent Dictator for Life (BDFL)** -- A single individual holds final authority over all decisions. The Linux kernel under Linus Torvalds, Python under Guido van Rossum (historically). Mos supports this by allowing a single identity to hold ratification authority over all Contracts.

**Committee / Steering Group** -- A small group shares decision-making authority. The Rust project's governance model, the Kubernetes steering committee. Mos supports this by allowing ratification to require approval from a defined set of identities.

**Consensus / Apache-Style Voting** -- Decisions require broad agreement, with formal voting mechanisms for contentious issues. The Apache Software Foundation model. Mos supports this by allowing ratification rules to specify quorum and voting thresholds.

**Emergent / Informal** -- No explicit governance model. Decisions happen through convention and momentum. Most small projects operate this way. Mos supports this by not requiring a governance model to be specified -- the defaults work without ceremony.

The governance model is a replaceable part, consistent with the Ship of Theseus principle. A project can start as a BDFL dictatorship, evolve into a committee structure as it grows, and formalize into a voting democracy when the community demands it. Mos tracks the transition as part of the project's mosal history -- amendments to the fundamental law.

---

## 8. Federation

Projects do not exist in isolation. They belong to organizations, ecosystems, communities. The federation model allows shared governance without surrendering sovereignty.

**Parent Moss** -- An organization-level `.mos` defines the inherited defaults for all member projects. These are the "international treaties" -- shared vocabulary, shared invariants, shared quality bars. Red Hat's DNA (upstream-first, open source, multi-operator) propagates to every project under the Red Hat umbrella as a set of defaults.

**Inherited Defaults vs Local Overrides** -- Inheritance is opt-out, not opt-in. A project inherits its parent's mos by default. It can override any inherited rule, extend the vocabulary, add its own contracts. The override is explicit and versioned -- it is a mosal amendment, not a silent deviation.

**Exit Right** -- A project can fork its governance entirely. It can sever its relationship to the parent mos and become fully sovereign. The exit is clean because Mos tracks what was inherited and what was local. Forking governance is a first-class operation, not a messy divorce.

**Cross-Pollination** -- The "universal rules" pattern observed in the asterisk and origami `.cursor/` directories -- manually copying shared norms between repositories -- is the problem federation solves. Instead of copying files, projects inherit from a shared parent. Changes to the parent propagate automatically. Local overrides are preserved.

---

## 9. The Judiciary (Harness Enforcement)

Laws without enforcement are suggestions. This is the central lesson of the uncodified mos problem.

Today's `.cursor/` rules are advisory. They exist as files on disk. They are loaded into agent context when the system decides to load them. An agent can acknowledge a rule, partially follow it, or ignore it entirely. There is no mechanism to verify compliance, no consequence for violation, no way to distinguish between "the agent followed the rule" and "the agent happened to produce output that looks like it followed the rule."

The gap is structural, not accidental. `.cursor/` was designed for human-readable guidance, not machine-enforceable law. It is a library, not a court.

Mos must provide the judiciary: the branch of governance that evaluates whether the law was followed. This is Harness Engineering applied to the mosal model. A harness is a test that an agent can invoke to verify its output against the specification. A mosal harness is a test that evaluates whether the rules defined in the `.mos` were respected during the execution of a Contract.

The difference between `.cursor/` (advisory) and `.mos` (enforced) is the difference between a code of conduct and a legal system. Both express norms. Only one has teeth.

Deterministic rule evaluation means that for any given Contract, the set of applicable rules is computable: the standing law of the project, plus the amendments introduced by the Contract, minus any rules explicitly suspended for that scope. The agent knows -- before it begins work -- exactly what rules apply. The harness knows -- after the work is done -- exactly what rules to check. There is no ambiguity, no hope, no vibes. Only proof.

### The Dual Nature of Enforcement

Not all rules are enforced the same way. The harness has two natures, corresponding to two classes of law:

**Mechanical / Deterministic (Errors).** Compilation failures, test failures, CI pipeline results, type mismatches, linting violations. These are binary outcomes -- pass or fail. Non-negotiable. They are strict liability in legal terms: violation is violation regardless of intent, regardless of context, regardless of how reasonable the deviation seemed. An agent that produces code that does not compile has failed the mechanical harness. The judiciary does not deliberate; it rules automatically.

**Verbal / Interpretive (Warnings).** Problem domain constraints, goal statements, specifications, architectural principles, style conventions that admit judgment. These are open to interpretation. An agent surfaces them; a human -- or the governance authority defined by the project's mos -- decides whether the spirit of the rule was met even if the letter was not. These are negligence standards: context matters, and reasonable deviation may be acceptable. A specification that says "the API should be RESTful" requires judgment about what "RESTful" means in a specific context. The harness flags the question; the judiciary (human governance) adjudicates.

The distinction maps to legal tradition: Errors are the criminal code (bright-line rules, automatic enforcement, no discretion). Warnings are the civil code (standards of care, adjudicated case by case, context-dependent). Both are law. Both are tracked. But only Errors halt execution automatically. Warnings are surfaced, recorded, and deferred to governance authority for resolution.

### Default Harness Methodologies

Mos ships with opinionated defaults for how mechanical harnesses operate during Contract execution. These are starting points, not mandates -- overridable like any mosal vocabulary, consistent with the Ship of Theseus principle.

**ROGBY (Red-Orange-Green-Yellow-Blue).** The default development cycle for mechanical harness compliance. ROGBY extends TDD's Red-Green-Blue cycle with two observability steps that ensure both failure and success modes are instrumented from the start, not retrofitted after the fact.

**Red -- Write a failing test.** Capture the intended behavior or reproduce the bug. The test must fail. If it passes, it does not cover the change. Each test maps to a Given/When/Then acceptance criterion from the behavior specification (Lifecycle sub-step 3.1).

**Orange -- Instrument error signals.** Before implementing the fix, add logging and instrumentation that surfaces failures, errors, and anomalies. Orange output answers: "What went wrong? Where? Why?" Log at error paths, failed assertions, rejected inputs, timeout triggers. Include machine-readable fields in structured logs. The first run after a change is otherwise blind. Orange ensures failures are visible from the start -- for humans reading terminal output and for agents parsing logs.

**Green -- Make the test pass.** Implement the minimal production code to pass the Red test. Run the full affected test suite. Everything must be green. Do not over-implement -- the test defines the scope.

**Yellow -- Instrument success signals.** With the code working, add logging that surfaces healthy operation and key decisions. Yellow output answers: "What happened? What did we choose? How long did it take?" Log successful path completions, selected branches, chosen classifications, per-operation timing, and throughput metrics. Yellow confirms the system is operating correctly and provides the baseline for regression detection and performance monitoring.

**Blue -- Refactor.** With tests green and full observability (Orange + Yellow) in place: remove duplication, improve naming, extract helpers. Review log levels -- Orange stays at Warn/Error, Yellow demotes from Info to Debug where appropriate. Verify the coverage matrix still holds.

The key insight: Orange and Yellow are not optional polish. They are the observability pair that makes the mechanical harness self-documenting. A Contract executed under ROGBY produces not just passing tests but a complete signal record -- what failed, what succeeded, how fast, and why. The harness evaluates the tests; the signals evaluate the system.

### The Verification Axis

The Mechanical/Verbal distinction describes *how* invariants are enforced. A second, orthogonal axis describes *from what direction* they are verified.

**Blue Team** (benevolent, positive space) -- confirming the system satisfies its declared invariants. A test passes. A specification is met. A boundary holds.

**Red Team** (adversarial, negative space) -- probing whether the system can be made to violate its invariants. A fuzz test finds a crash. An adversarial review discovers the spirit of the spec was violated despite the letter being met.

The two axes are independent, producing four quadrants:

- **Mechanical + Blue.** A test passes -- hard cutoff, benevolent stimulus.
- **Mechanical + Red.** An adversarial fuzz test finds a crash -- hard cutoff, adversarial stimulus.
- **Verbal + Blue.** A specification appears satisfied -- judgment call, benevolent framing.
- **Verbal + Red.** An adversarial review finds a spirit violation -- judgment call, adversarial framing.

A complete Harness evaluates all four quadrants. Blue gives you a gate -- the contract can proceed. Red gives you a posture -- the system is always being probed.

---

## 10. Freedom of Action

Every governance system risks becoming an obstacle to the thing it was designed to protect. A mos that makes change impossible is not a mos -- it is a cage. Process for process's sake is the death of momentum. The judiciary without an escape valve becomes tyranny.

Mos must enshrine freedoms alongside constraints. A project's governance is incomplete if it only defines what is forbidden and never defines what is permitted. The balance is not freedom *or* structure; it is freedom *within* structure, with explicit mechanisms to renegotiate the boundary when the structure no longer serves.

**Right to experiment.** Any contributor -- human or agent -- can fork without permission. The fork inherits the standing mos but is free to amend it locally. Experimentation is not rebellion; it is the mechanism by which new ideas are tested before they are proposed as law. A fork is a laboratory, not a secession.

**Right to propose.** Anyone can introduce a Bill. The barrier to proposing change must be near zero. The cost of proposing is visibility -- the proposal enters the public record -- but the cost of proposing is never permission. Only enacting change carries ceremony; suggesting change does not. A system where proposals require approval before they can even be heard has confused deliberation with gatekeeping.

**Graduated governance.** Not all changes need the same level of process. A typo fix is not a Bill. An architectural restructuring is. Mos must distinguish between executive action -- small, immediate, low-risk changes that can be applied without deliberation -- and legislation -- large, high-impact changes that require the full Bill lifecycle. The threshold between the two is configurable, another replaceable part consistent with the Ship of Theseus principle. A small team might set it high, trusting contributors to act freely. A large organization might set it low, requiring deliberation for more categories of change. Neither is wrong; both are valid governance choices.

**Right to dissent.** Negative space is preserved as a mosal primitive. Abandoned Bills, rejected proposals, minority opinions, dissenting arguments -- all remain in the historical record. Dissent is not erased; it is archived. A future contributor who arrives at the same idea can see that it was tried, why it failed, and what has changed since. The record of what was rejected is as valuable as the record of what was enacted.

**The escape valve.** If the governance model itself becomes an obstacle, the governance model itself can be amended. This is the meta-mosal right: the process is subject to the process. The Ship of Theseus applies to governance, not just to code. A project can begin with no formal governance, adopt a committee structure when it grows, and replace that committee with a voting democracy when the community demands it. Mos tracks every transition as an amendment to the fundamental law. No governance model is permanent unless the project chooses to make it so.

---

## 11. The Laboratory (Live Forks)

Today, forks are blind. You clone a repository, make changes, and the only feedback is whether tests pass. You cannot see the mosal effects of your changes: drift from desired state, rule violations, architectural boundary crossings, impact at resolutions above or below the file level. The fork exists in a vacuum, disconnected from the project's map. The decision to merge is based on code review and gut instinct, not on measured evidence of mosal impact.

The Laboratory is a mosal fork that runs in a sandboxed environment with full harness evaluation in real time.

**Fork the `.mos` alongside the code.** When a contributor forks a repository, the fork carries not just the source code but the entire mosal state: the standing rules, the vocabulary, the resolution layers, the desired state, the governance model. The fork is a complete project-state, not a partial copy.

**Evaluate against the standing mos.** The lab environment runs the harness against the forked state continuously. As the contributor makes changes, the harness evaluates: what rules are followed, what rules are violated, what drift is introduced between the fork's concrete state and the project's desired state. The contributor sees a diff not just of code but of mosal state -- a resolution-aware impact assessment that shows effects at every layer, from individual files to architectural boundaries to domain invariants.

**The committee stage, realized.** In parliament, the committee stage is where evidence is gathered. Experts testify. Impact assessments are produced. The Bill is scrutinized not in the abstract but against the concrete reality of its effects. The Laboratory makes this concrete for software. Before a Bill proceeds to deliberation and ratification, it has already been run. The evidence of its effects is attached to the proposal. Ratification is an informed decision, not a leap of faith.

**Real-time feedback loop.** The Laboratory is not a batch process. As the fork evolves, the harness evaluates continuously. The operator watches drift, compliance, and impact update in real time -- like monitoring a staging environment, but for the mosal state, not just runtime behavior. A change that violates a standing rule is flagged the moment it is made, not after a review cycle. A change that closes drift between desired and concrete state is visible immediately as progress.

**From lab to law.** When the evidence supports it, the fork's changes are proposed as a Bill. The Bill enters the First Reading with its laboratory evidence already attached: what rules were followed, what rules were violated, what drift was introduced or closed, what the resolution-level impact looks like. The deliberation that follows is grounded in measured reality. The committee stage has already happened in the lab; the readings and debate happen with the results in hand.

This mechanism resolves the tension between freedom and governance. The right to experiment is unconditional -- anyone can fork and run a lab. The right to enact is conditional -- only proposals with evidence proceed to ratification. Freedom lives in the fork; accountability lives in the merge. The Laboratory is the bridge between the two.

---

## 12. Embassies -- Bidirectional Integration

Mos does not exist in a vacuum. Every project already lives across two planes of reality that have never been connected.

The **Ideal Plane** holds the Desired State in the forms organizations already use: issue trackers (Jira, Linear, GitHub Issues), project management tools, roadmaps, OKRs, ADRs, specification documents. This is the Platonic realm -- where intent is expressed, priorities are set, and plans are drawn. The Ideal Plane answers the question "what should be true?"

The **Concrete Plane** holds the Concrete State in the systems that run production: Kubernetes clusters, CI/CD pipelines, observability stacks, cloud infrastructure, monitoring dashboards. This is the realm of friction -- where theory meets metal, containers crash, latency spikes, and dependencies conflict. The Concrete Plane answers the question "what is actually true?"

Today these planes are disconnected. A Jira ticket says "deploy feature X." The Kubernetes cluster already runs a conflicting version of X. No system surfaces the contradiction. A roadmap says "service Y is deprecated." The pipeline still builds and deploys Y nightly. Drift between intent and reality accumulates silently until it erupts as an incident, a broken release, or a six-month project that discovers on day one of integration that its assumptions were wrong.

Mos bridges the planes through **Embassies** -- bidirectional adapters that synchronize the `.mos` map with external systems without replacing them.

**Ideal Plane Embassies.** An Embassy to Jira materializes a Jira ticket as a Bill within Mos: the ticket's description becomes the Bill's intent, its priority maps to governance urgency, its assignee maps to the Contract's executor. When the Bill's status changes within Mos -- ratified, enacted, abandoned -- the change propagates back to Jira. The external system remains the organization's interface for planning; Mos becomes the mechanism that connects plans to execution and tracks whether they converge.

**Concrete Plane Embassies.** An Embassy to Kubernetes reads cluster state and reflects it into the `.mos` resolution map. What the deployment descriptor says, what the cluster actually runs, and what the code repository contains are three data points that Mos can compare. Drift between any pair is visible at the appropriate resolution layer. An Embassy to a CI/CD system maps pipeline runs to Contract execution: a build triggered by a ratified Contract reports its outcomes back into the `.mos`, closing the loop between "what was agreed" and "what was delivered."

**The Embassy protocol.** Each Embassy is an adapter, not a monolith. Mos defines the interface -- what data it expects, what events it emits, what synchronization guarantees it provides. The community builds the adapters: a Jira Embassy, a Linear Embassy, a Kubernetes Embassy, an ArgoCD Embassy. Each adapter is a replaceable module, consistent with the Ship of Theseus principle. If an organization uses Jira today and Linear tomorrow, it swaps one Embassy for another. The `.mos` does not care which external system provides the data; it cares about the data.

**The key principle.** Mos never becomes a "source of truth" that replaces Jira or Kubernetes. It is the map that reads from and writes to both planes. The external systems remain authoritative for their domain: Jira owns the organization's planning workflow, Kubernetes owns the cluster state, the CI/CD system owns the build pipeline. Mos provides the unified view -- the single place where a human or agent can see the Desired State, the Concrete State, and the drift between them at any resolution. The map is not the territory, but without a map, you navigate blind.

---

## 13. The Operator Experience

The system should feel like a real-time strategy game combined with a studio mixing console. The terrain is code. The units are AI agents. The operator commands from above.

**The Map is the resolution control.** Zoom out: the domain map with colored territories, module icons between zones, drift indicators pulsing where Desired and Concrete states have diverged. Zoom in: file-level detail -- test results, individual module actions, line-by-line changes under an active Contract. The zoom level determines what information is visible.

**Fog of war.** Areas with low Integrity Index -- undocumented boundaries, unverified invariants, untested paths -- are fogged. The fog lifts as modules verify, document, and test. The goal is not zero fog but informed fog: knowing where the unknowns are.

**Locked zones.** Files under active work are locked from concurrent writes. Locked files communicate ownership, progress, and the Contract they belong to.

### Operator Modes

Three modes, each a different level of engagement:

**Observation.** The operator watches the Map, sees drift indicators update, watches controllers flag and resolve violations. No direct intervention. This is the default during Contract execution.

**Direction.** The operator introduces Bills, ratifies Contracts, sets priorities, defines scope. They specify the *what* and *why*; modules handle the *how*.

**Intervention.** The operator writes code directly, overrides module decisions, resolves conflicts that modules cannot adjudicate. This is Legacy Mode activated at a specific scope -- the emergency brake, always available.

---

## 14. The Parliamentary Assistant -- Agent Role Model

In parliamentary systems, elected officials -- Members of Parliament, Senators, Representatives -- hold decision authority. They debate, amend, and vote. They are accountable to their constituents. They set the laws of the state.

Parliamentary assistants serve them. They research policy, brief legislators on complex topics, draft Bills and amendments, coordinate between offices, prepare evidence for committee hearings, execute administrative tasks, and report outcomes. They do not vote. They do not set strategic direction. They are essential to the functioning of the legislature, but the authority belongs to the elected officials.

Mos adopts this model exactly. Humans are legislators. AI agents are parliamentary assistants.

**Agent powers.** Within the mosal framework, an agent is empowered to:

- **Research.** Explore the codebase, query the resolution map, gather evidence from the Ideal and Concrete planes through Embassies, analyze drift, surface risks, and compile briefings for the human operator.
- **Draft.** Propose Bills: write code, specifications, rule amendments, vocabulary changes. A draft is not law; it is a proposal awaiting deliberation.
- **Prepare evidence.** Run proposed changes in the Laboratory. Attach harness results, drift assessments, and resolution-level impact reports to the Bill. The agent builds the evidence base that informs the human's decision.
- **Execute.** Carry out ratified Contracts under harness supervision. Once a human ratifies a Bill into a Contract, the agent performs the work within the Contract's rules. The harness verifies compliance continuously, not just at the boundary.
- **Report.** Surface outcomes: what was done, what drift was introduced or closed, what rules were followed or violated, what friction was encountered. Transparency is not optional; it is the agent's mosal obligation.

**Agent limits.** Within the mosal framework, an agent is prohibited from:

- Ratifying Bills. Only the governance authority -- human, committee, or consensus body as defined by the project's governance model -- can ratify.
- Amending standing law. An agent can propose amendments via a Bill, but cannot enact them unilaterally.
- Overriding governance authority. If the governance model requires human approval for a category of change, the agent cannot bypass it regardless of confidence level.
- Concealing friction. An agent that encounters a contradiction between the Desired State and the Concrete State must surface it. Silently working around a problem -- producing output that appears correct while deviating from specification -- is a mosal violation.

**The friction principle.** The Desired State is Platonic -- a specification in the realm of logic where everything is consistent and complete. The Concrete State has friction -- real systems break, dependencies conflict, edge cases multiply, production behaves in ways the specification did not anticipate. Theory is ideal; reality has friction.

The agent's role is to navigate friction on behalf of the human, not to pretend it does not exist. An agent that reports "this cannot be done as specified because of X, here are three alternatives with trade-offs" is more valuable than one that silently deviates and produces output that technically compiles but violates the intent. The parliamentary assistant who tells the legislator "your proposed law conflicts with an existing statute" is doing their job. The one who rewrites the law quietly to avoid the conflict is overstepping.

**Graduated agency.** The governance model determines how much autonomy agents receive, and this is deliberately configurable -- another replaceable part.

- A **BDFL project** might grant agents broad executive action authority: any change below a configurable threshold can be applied without deliberation. The BDFL trusts the assistant to handle routine work and escalate only when the threshold is crossed.
- A **committee-governed project** might require human approval for any change above a low threshold. The agent drafts and proposes; the committee decides. Agency is narrow but well-defined.
- A **consensus-driven project** might allow agents to execute only after broad stakeholder acknowledgment. The agent's role leans heavily toward research and evidence-gathering, with execution reserved for uncontested work.

The threshold between executive action and legislation is the dial. Mos does not decide where the dial is set. The project-state's governance model does. The dial can be adjusted as the project evolves -- turned up when trust is high, turned down when risk is high. The adjustment itself is a mosal amendment, tracked in the historical record.

## 15. The Declaration -- Mos at Birth

Every system described so far assumes a project already exists -- code in repositories, rules in `.mos`, governance models chosen, vocabulary defined. But a project's most critical moment is its founding: the point where there is no state, only a sense. A business need. A pain. A conviction that something must exist that does not yet exist.

Today this founding moment produces the most scattered, most uncodified, and most consequential artifacts of any project's lifetime. Markdown files in a folder. Slack threads that scroll away. Pitch decks emailed between stakeholders. Wiki pages that nobody links to correctly. Napkin sketches photographed and forgotten. Every decision made in this embryonic phase shapes everything that follows -- the architecture, the team structure, the governance model, the vocabulary -- yet none of it is captured in a system that can track, version, render, or enforce it.

This is the phase Mos must support first, not last.

**The Declaration as the first `.mos` artifact.** Before rules, before vocabulary, before resolution layers, before Embassies, before Bills and Contracts -- there is the Declaration. Every project has one, whether or not it is written down. It is the project's raison d'etre fused with its declaration of independence: the founding document that articulates WHY this project must break from the status quo and WHAT it aspires to become.

The Declaration captures:

- **The Sensation.** The raw business need -- the pain, the opportunity, the gap that cannot be ignored. This is not a specification; it is the felt experience that starts the project. "We need this because..."
- **The Contextualization.** The domain boundary -- what is in scope, what is out of scope, what adjacent systems exist, what constraints are inherited. This is the first act of mapping: drawing a border around the problem before solving it.
- **The founding principles.** The non-negotiable convictions that will guide every subsequent decision. These are not rules yet -- they are philosophical commitments. "We believe that..."
- **The initial vocabulary.** The first words the project uses to describe itself. These will evolve, but they must be captured at birth because they shape how the project thinks about its own domain.

The Declaration is not code and not a specification. It is the seed from which both grow. It sits at the root of the `.mos` -- the founding document from which the resolution map, the governance model, the vocabulary, and eventually the codebase all descend. Every subsequent artifact can trace its lineage back to the Declaration. If a Bill contradicts the Declaration, that contradiction is visible. If the project drifts from its founding principles, the drift is trackable.

**The political parallel.** In mosal law, the declaration of independence precedes the mos. The American Declaration of Independence (1776) preceded the Mos (1787) by eleven years. The declaration articulates the grievances and the aspirations; the mos provides the governance structure to realize them. The same sequence applies here: the Declaration comes first, establishing the WHY, and the governance model follows, establishing the HOW.

Organizations that operate as federations -- multiple project-states under a shared umbrella -- may have an overarching Declaration at the federation level (the organization's mission, values, and strategic direction) and individual Declarations at the project-state level. A project under Red Hat inherits organizational DNA but declares its own specific raison d'etre. The federation's Declaration constrains what project-states can declare -- a form of mosal supremacy -- but does not replace the project-level founding act.

**The bootstrap problem.** Building an IDE from scratch requires an IDE to build it in. Every self-hosting system faces this chicken-and-egg constraint. The project begins within a borrowed shell -- an existing IDE pushed to its limits while the actual system takes shape inside it. The final milestone that resolves the paradox: Macro self-compiles. The borrowed shell is no longer needed. The embryo outgrows its host.

**The dogfood test.** This design document -- `DESIGN.md` -- is the Declaration of the Macro and Mos project. They articulate the Sensation (existing IDEs fail the Agentic Age), the Contextualization (the domain of human-agent collaboration, version control, and governance), the founding principles (multi-component architecture, Ship of Theseus, desired vs. concrete state), and the initial vocabulary (Bills, Contracts, Embassies, Declarations, resolution layers).

They are also the proof that the Declaration phase is currently unsupported. They are markdown files maintained by hand, scattered across a folder, with no resolution control, no enforcement, no Embassy integration, no drift tracking, no governance model governing their own evolution. They are links to links to links -- the very problem they describe.

The **first PoC milestone** is not "Macro self-compiles." That is the final milestone. The first milestone is: **express these founding documents as `.mos` artifacts, rendered and manipulated through Macro.** If Mos cannot manage its own Declaration -- cannot version it, cannot track drift between what it says and what the codebase does, cannot render it at multiple resolutions, cannot enforce its founding principles against subsequent work -- then it cannot manage anything. The manifesto is the dogfood. When the markdown files are replaced by `.mos` artifacts and the Macro Control Surface can render, navigate, and govern them, the Declaration phase is proven and Mos has demonstrated it can support a project from conception, not just from first commit.

---

## 16. Koinonia -- Communion Between Forks

Forks are the mechanism of divergence. In traditional version control, a fork is a one-way door: you clone, you diverge, and the only way back is a merge request reviewed by a human who may or may not remember what the fork was for. The relationship between a fork and its parent is implicit -- a shared commit history and nothing more. There is no ongoing awareness, no graduated notification, no formal model for the degree of alignment between the two.

Mos introduces **Koinonia** -- communion between forks. The term is borrowed from the Greek concept of fellowship and shared participation. In mosal terms, Koinonia is the formal relationship between a fork and its parent (or between sibling forks) that determines how aware each is of the other's changes.

Communion exists on a spectrum:

**Full Communion.** Both forks share governance, vocabulary, and intent. They are aligned on the fundamental law and diverge only on implementation details. Changes in one fork propagate to the other as informational updates -- no barriers, no approval required, no friction. A fork in full communion with its parent is a branch in all but name. The two are working toward the same mosal goals and trust each other's governance.

**Partial Communion.** The forks share some governance but have diverged on specific rules, vocabulary, or architectural decisions. They are aligned in principle but differ in practice. Changes from one fork appear as **warnings** in the other -- visible, tracked, but not blocking. The warning signals: "something has changed in a related fork that you may want to know about." The recipient decides whether to act. Partial communion is the natural state of forks that are exploring different approaches to the same problem -- the laboratory model extended across fork boundaries.

**No Communion.** The forks have fully diverged. They no longer share governance intent. Changes from one fork appear as **errors** in the other, or are permission-denied entirely. This is diplomatic severance -- the mosal equivalent of a state withdrawing from a treaty. The forks are now sovereign and independent. They may share historical ancestry, but they no longer share a living relationship.

**Communion as metadata.** Mos tracks the degree of communion as metadata on the fork relationship. The metadata includes: which rules are shared, which vocabulary is common, what governance model each fork uses, and what notification level applies. Communion can be renegotiated at any time -- a fork that was in full communion can reduce to partial if its direction diverges, and a fork with no communion can re-establish partial communion if the two sides find common ground.

**The incentive structure.** Communion is not enforced; it is incentivized. A fork in full communion benefits from upstream improvements automatically. A fork with no communion must maintain everything independently. The cost of divergence is real but chosen. Mos makes the cost visible -- it does not hide it, and it does not prevent it. The choice to diverge is a sovereign right; the consequence of divergence is a mosal fact.

---

## 17. Ecumenical Councils -- Coordinated Pivots

The older a project, the harder it is to pivot. This is not a bug. Mosal inertia -- the accumulated weight of standing law, vocabulary, architectural decisions, and historical precedent -- protects stability. A project with ten years of history and a hundred contributors cannot change direction as easily as a solo project on day one. This is by design: inertia ensures that changes are deliberate, not accidental.

But sometimes a pivot is necessary. The domain shifts. The technology landscape changes. A fundamental assumption proves wrong. The governance model no longer serves. When the scale of change exceeds what a normal Bill can address -- when the change affects not just one project-state but multiple forks, multiple teams, or an entire federation -- the normal legislative process is insufficient.

An **Ecumenical Council** is a governance event where representatives from multiple forks (or multiple project-states in a federation) convene to discuss a major mosal change that affects all parties.

The analogy is historical. In the Christian tradition, ecumenical councils were extraordinary gatherings convened to resolve doctrinal disputes that could not be settled by individual churches alone. The Council of Nicaea (325), the Council of Chalcedon (451), the Second Vatican Council (1962-1965) -- each addressed questions that transcended individual jurisdictions and required collective deliberation to resolve. The outcomes were binding across the participating churches, and the decisions shaped doctrine for centuries.

In Mos, Ecumenical Councils serve the same function:

**Trigger conditions.** A Council is convened when a proposed change meets criteria that exceed normal Bill processes: it affects multiple project-states in a federation, it requires vocabulary changes that break backward compatibility, it proposes a governance model transition that affects fork relationships, or it addresses a cross-cutting architectural decision that no single project-state can make unilaterally.

**Representation.** Each affected project-state (or fork) sends a representative -- human, not agent. Agents may prepare evidence, draft proposals, and compile briefings, but the Council is a governance body and governance authority belongs to humans. The governance model of each participating project-state determines who its representative is: the BDFL, a committee delegate, a consensus-chosen spokesperson.

**Deliberation.** The Council deliberates on the proposed amendments. Laboratory evidence is presented. Impact assessments -- mosal drift, vocabulary incompatibilities, governance model conflicts -- are reviewed. Amendments are proposed, contested, and resolved through the collective governance process.

**Ratification.** Council decisions are ratified according to the rules agreed upon at the Council's convening. A federation-level Council might require unanimous consent; a cross-fork Council might require majority approval. The ratification rules are themselves subject to deliberation -- the Council decides how the Council decides, within the bounds of its charter.

**Propagation.** Ratified amendments propagate across communion lines. Forks in full communion receive them automatically. Forks in partial communion receive them as warnings. Forks with no communion are unaffected -- they have already severed the relationship.

**Historical record.** The Council's proceedings, evidence, deliberation, dissenting opinions, and ratified amendments are all recorded in the `.mos` as a mosal event. Future contributors can trace any cross-cutting decision back to its Council, understand the evidence that informed it, and see the dissenting views that were considered and overruled.

Councils are extraordinary events, not routine governance. They are the mosal equivalent of a mosal convention -- convened only when the normal legislative process cannot address the scale of change required. Their existence ensures that even the most deeply entrenched mosal decisions can be revisited, but only with the gravity and deliberation that such revision demands.

---

## 18. The Cabinet -- Personal Governance Overlay

Everything described so far is project-level: the mos, the Bills, the Contracts, the governance model, the vocabulary -- all belong to the project-state. But a developer is not bound to one project. They move between repositories, organizations, and technology stacks. They carry personal preferences, accumulated heuristics, a working philosophy, and opinions about how agents should behave. Today this personal layer is scattered across `~/.gitconfig`, shell aliases, editor settings, and muscle memory. It has no formal relationship to the projects it operates within.

The **Cabinet** is a per-developer `.mos` fragment that formalizes this personal layer. The name comes from two sources: the minister's working cabinet -- the private advisory group and personal papers a minister carries between portfolios -- and the physical cabinet (the piece of furniture) where one keeps personal documents separate from institutional ones. The technical analogy is `~/.gitconfig` or `.bashrc`: global user configuration that shapes behavior in every project without the project needing to know the internals.

### What the Cabinet Carries

**Personal rules.** Working style, naming preferences, code organization habits, testing philosophy, documentation discipline. These are the developer's own mosal principles -- the things they care about regardless of which project they are working on. In kombucha terms, this is the "Soul" formalized: the developer's philosophy encoded as mosal rules rather than a freeform text block.

**Personal vocabulary.** Preferred terminology, domain shorthands, mental model labels. A developer who thinks in terms of "adapters" and "ports" carries that vocabulary into every project, even when the project's vocabulary uses different words for the same patterns.

**Accumulated heuristics.** Lessons carried from previous projects. "Never trust library X for concurrency." "Always verify Y before deploying to Z." "When the build takes longer than N minutes, check W first." These are personal case law -- precedents that shape the developer's judgment even when no project has codified them as standing rules.

**Agent tuning.** How the developer wants their parliamentary assistants to behave: verbose or terse reporting, cautious or aggressive execution, how much autonomy to grant before escalating, preferred explanation depth, risk tolerance for experimental changes. The Cabinet provides the agent with a personality calibration for its human operator.

### Load and Unload

When a developer opens a project, their Cabinet is loaded as a personal overlay on top of the project's mos. When they leave or switch projects, the Cabinet detaches cleanly -- no trace remains in the project's `.mos`.

The interaction follows the CSS cascade model. The project's mos is the base stylesheet -- it defines the authoritative rules. The Cabinet is the user stylesheet -- it provides personal preferences that fill gaps and adjust margins. When the two conflict, the project wins. If the project mandates four-space indentation and the Cabinet prefers tabs, the project's rule governs. If the project is silent on commit message style and the Cabinet has a preference, the Cabinet's preference applies. The precedence is unambiguous: project mos overrides Cabinet, always.

This is the same inheritance model as `~/.gitconfig`: git's configuration cascades from system (`/etc/gitconfig`) to global (`~/.gitconfig`) to local (`.git/config`). The local (project) level overrides the global (Cabinet) level. Mos applies the same principle to governance, not just settings.

### Save Back

A developer can save personal learnings from a project back into their Cabinet. A heuristic discovered during a difficult debugging session, a rule of thumb about a technology stack, a refined preference about agent behavior -- these can be explicitly captured in the Cabinet for future use. The save is one-directional: personal takeaways flow from the project into the Cabinet, but the project's `.mos` is unaffected. Fire-and-forget for the project; retained personally.

### Privacy

The Cabinet is opaque to the project and to other developers. The project's mos sees the developer's *actions* -- their commits, their Bills, their Contract execution, their harness compliance -- but not their personal configuration. Other developers cannot read each other's Cabinets. The harness evaluates against project law; Cabinet preferences operate only within the margins the mos allows.

This is a deliberate boundary. A project-state has no right to inspect or govern a developer's personal preferences, just as a government has no right to inspect a minister's private papers. The developer's mosal compliance is public (and must be -- the harness enforces it). Their personal working style is private.

### Portability

The Cabinet is a `.mos` fragment that lives outside any project's `.mos`. It resides in the developer's personal space -- the mosal equivalent of `~/.mos/cabinet/` or similar. It is portable across machines through whatever mechanism the developer uses to sync dotfiles: a git repository, a sync service, a USB drive, or a personal Monad partition. The hosting model is an implementation detail; the semantic contract is that the Cabinet belongs to the developer and travels with them.

### Ship of Theseus

The Cabinet is a replaceable part, consistent with the principle applied everywhere else. A developer can start fresh with an empty Cabinet. They can import someone else's Cabinet as a template -- adopting a senior colleague's heuristics as a starting point, then overriding and evolving them over time. They can maintain the same Cabinet for a decade, watching it accumulate the mosal history of their own career. The Cabinet's amendments -- additions, overrides, discarded preferences, evolved philosophy -- are tracked like any `.mos` artifact. The developer's personal mosal journey is preserved: how their working style changed, what they learned, what they abandoned.
