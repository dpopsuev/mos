# Glossary

> Four lenses on the same architecture. The SE/IT column is the primary vocabulary for documents and code. The Synth and Neuroscience columns are secondary analogies that illuminate the *why* behind each concept.

## Core Components

| SE/IT Name | Synth Analogy | Neuro/Gov Origin | Concept |
|---|---|---|---|
| **Macro** | Synthesizer | Ideas | The Agentic Age IDE (forked from Micro editor) |
| **Origami** (Router) | Patch Matrix | Pipeline Engine | Signal routing -- any output to any input |
| **Harness** | Filter Bank | Harness Runtime | Verify and shape agent output against specification |
| **Gate** (Errors) | Low-Pass Filter | Mechanical Harness | Hard cutoff -- binary pass/fail, non-negotiable |
| **Advisory** (Warnings) | Parametric EQ | Interpretive Harness | Shaped enforcement requiring human judgment |
| **Governor** | VCA | Locus Coeruleus | Scale resource allocation proportional to demand |
| **Heartbeat** | Master Clock | SCN Tick | Fixed-interval probe, independent of all work |
| **Controller** | Oscillator | Enforcer Loop | Continuous domain-tuned verification (K8s controller pattern) |
| **Reconciler** | Mixer | Central Reconciliation Loop | Cross-domain convergence and state comparison |
| **Mos** | Patch Memory | Mos (.mos/) | State store -- versionable, recallable configuration |
| **Profile** | Player Preset | Cabinet | Portable personal configuration across projects |
| **Macro** | Control Surface | TUI / CLI | The operator's primary interface |
| **MCP** | MIDI | MCP | Machine-to-machine wire protocol |
| **Sophia** (Context Cache) | Delay/Reverb | Context Engine | Agentic context funnel -- wide intake, gravitational narrowing, precise output |
| **Monad** (Archive) | Sample Library | Collective Memory | Long-term collective knowledge store |
| **Embassy** | Audio Interface | Embassy | Adapter bridge to external systems |
| **Map** | Spectrum Analyzer | Resolution Map | Multi-zoom project state visualization |
| **Integrity Index** | Signal-to-Noise Ratio | Trust Thread | Quantified invariant coverage across all resolutions |
| **Universalis** | -- | -- | Shared Rust library: gravitational graph engine for Sophia and Monad |
| **ROGBY** | ROGBY | ROGBY | Red-Orange-Green-Yellow-Blue verification methodology |

## Signal Flow and Dynamics

| SE/IT Name | Synth Analogy | Neuro/Gov Origin | Concept |
|---|---|---|---|
| **Signal** | CV Signal | Walker | Context-carrying message traversing the router graph |
| **Actor Model** | Modular Rack | Actor Model | Computational foundation -- isolated actors, async message passing, no shared state |
| **Event Bus** | CV Bus | SignalBus | Broadcast event propagation without direct coupling |
| **Namespace** | Rack Section | Zone | Functional grouping of related modules |
| **Acceptance Criteria** | Frequency Cutoff Points | Gherkin Scenarios | Specification of what passes and what is rejected |
| **Guard** | Noise Gate | Tonic Inhibition | Activation threshold -- default is idle, signal must qualify |
| **Burst** | Envelope Attack | Phasic Burst | Rapid resource onset when demand arrives |
| **Run Loop** | Arpeggiator | Central Pattern Generator | Autonomous cyclic execution once direction is set |
| **Housekeeping** | Sustain/Release | Consolidation Mode | Background maintenance during idle (GC, index, cache) |
| **Drift** | Detuning | Drift | Desired State vs Concrete State misalignment |
| **Blue Team** | Fundamental | Blue Team | Benevolent verification -- proving invariants hold |
| **Red Team** | Overtone | Red Team | Adversarial probing -- testing invariant resistance |
| **High Concurrency** | Full Polyphony | Multiple Contracts | Many controllers active simultaneously under load |
| **Operational States** | ADSR Envelope | Four System States | Idle, Reactive, Active, Peak |

## Unchanged Terms

These terms are shared across all lenses:

- **Mos** -- The `.mos` format and governance system
- **Desired State** -- The intent: what the system should be
- **Concrete State** -- The reality: what the system actually is
- **Declaration** -- The founding document of a project
- **Bill** -- A public proposal for change
- **Contract** -- A ratified, binding unit of work
- **Gherkin** -- The Given/When/Then specification language (native to the `.mos` DSL)
- **Resolution Layers** -- Organization, Group, Domain, Individual
- **Lifecycle Phases** -- Sensation, Contextualization, Abstraction, Materialization, Externalization
- **Ship of Theseus** -- The replaceability principle
- **SSH Key Pair** -- The cryptographic identity primitive
- **Signed Hash Chain** -- The tamper-evident history model
- **Operator** -- The human controlling the system

## SE/IT Rationale

Why each SE/IT term was chosen:

- **Controller** -- The Kubernetes controller pattern: watch state, detect drift, reconcile.
- **Reconciler** -- The Kubernetes reconciliation loop: compare desired vs actual, trigger correction.
- **Embassy** -- Bidirectional adapter to external systems, analogous to API gateways and payment gateways.
- **Governor** -- Resource governor (SQL Server, Linux cgroups). Controls rate/capacity allocation based on demand.
- **Heartbeat** -- Distributed systems liveness signal. Fixed-interval, independent, near-zero cost.
- **Guard** -- Route guards, auth guards, type guards. Default-deny with threshold qualification.
- **Harness** -- Test harness (Mitchell Hashimoto). Tooling that agents call to verify their own output.
- **Gate** -- CI/CD quality gate. Binary checkpoint that must pass before proceeding.
- **Advisory** -- Advisory lock, advisory review. Visible but not blocking; requires human judgment.
