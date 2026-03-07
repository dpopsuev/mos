# Dunix: A Solar System History

> Canonical timeline for the Dunix Production Simulation.
> 150 years, 5 eras, 14 sites (10 civilian + 4 military), 52+ repos,
> 7 locales, 8 governance models, 2 wars.

---

## The Physical Constraint: Light-Speed Delay

Every governance decision in Dunix is shaped by one constant: *c*.

| Route | One-way delay | Governance implication |
|---|---|---|
| Earth -- Luna | 1.3 s | Real-time co-development possible |
| Earth -- Mars | 4-24 min | Async-first culture mandatory |
| Earth -- Ceres (Belt) | 14-28 min | Politically distinct from Mars |
| Earth -- Jupiter (Ganymede) | 33-54 min | Local autonomy mandatory |
| Earth -- Saturn (Titan) | 68-84 min | Reviews are day-long affairs |
| Earth -- Uranus (Miranda) | 2.5-2.8 hr | Fully decoupled development cycles |
| Earth -- Neptune (Triton) | 4-4.2 hr | Batch sync only |

### Simulation Mapping

1 tick = 1 day. Latency is collapsed to tick-scale integers for simulation:

| Route | Ticks (one-way) |
|---|---|
| Earth -- Luna | 1 |
| Earth -- Mars | 14 |
| Earth -- Ceres | 21 |
| Earth -- Ganymede | 43 |
| Earth -- Titan | 76 |
| Earth -- Miranda | 165 |
| Earth -- Triton | 280 |

---

## Geopolitical Factions

Six factions shape the 150-year history. Their interests, governance
philosophies, and conflicts drive the narrative:

| Faction | Type | Governance style | Interests |
|---|---|---|---|
| **ISA** (Interplanetary Space Agency) | Civilian, neutral | Arbiter; controls Luna relay | Communication infrastructure, neutrality |
| **USSN** (USA Space Navy) | Military | Compartmentalized hierarchy + async | Belt resource extraction, outer patrol, comms dominance |
| **SRJSN** (Sino-Russo Joint Space Navy) | Military | Chain of command, sealed orders | Strategic parity, independent supply chains, Jupiter control |
| **BFA** (Belt Free Association) | Civilian, independent | Oral + mandate (no `need` artifacts) | Resource sovereignty, autonomy from Inner System |
| **OWC** (Outer Worlds Council) | Civilian, sub-federation | Covenant + reflection, compact async | Self-sufficiency, philosophical counterweight to militarism |
| **ISA-Civilian** (general) | Civilian | Varies by site | Open-source OS development, scientific cooperation |

### Tensions

- **USSN vs SRJSN**: Cold war in space. Competing military doctrines, intelligence
  operations, proxy conflicts through Belt politics.
- **Belt vs Inner System**: Resource extraction politics. The Belt produces raw
  materials; Earth and Mars consume them. BFA demands sovereignty.
- **Titan vs everyone**: Governance philosophy. Titan insists on reflection and
  conscience. Others consider it overhead -- until Enceladus proves them wrong.
- **Military vs civilian**: Compartmentalization vs transparency. Military forks
  operate as opaque nodes in the federation graph. Civilians resent the blind
  spots; militaries resent the exposure.

---

## Era I: Genesis (Years 0-5) -- "One Repo, One Planet"

**Ticks 0 -- 1,825**

- **Year 0 (tick 0)**: The Interplanetary Space Agency (ISA) commissions Dunix --
  Distributed Unix -- for astronaut-ground communication. Four repositories are
  created on Terra: `dunix-kernel`, `dunix-userspace`, `dunix-net`, `dunix-hab`.
  TerranEnglish is the sole locale. 3 teams, ~15 engineers.

- **Year 1 (tick 365)**: The first Mars colony boots. Six engineers clone all
  four repos. The first cross-planet commit review takes 48 minutes round-trip.
  Mars adopts Terra's governance wholesale.

- **Year 3 (tick 1,095)**: Two events. Luna Relay Station comes online -- Luna
  becomes the CI/CD hub and neutral arbiter, running the linter on every
  cross-site push but never writing code. Simultaneously, the **USA Space Navy
  (USSN)** adopts Dunix for fleet communications. USSN forks `dunix-kernel` and
  `dunix-net` into compartmented repos: `ussn-kernel` and `ussn-fleetcom`. The
  first `compartment` block appears -- artifacts tagged with clearance levels
  (`unclassified`, `confidential`, `secret`, `top-secret`) and codeword access
  lists. Outsiders see artifact IDs and kind but not titles or content. The era
  of **opaque references** begins.

- **Year 5 (tick 1,825)**: Mars has 20 engineers. The **MartianEnglish** locale
  emerges: `capability` replaces `rule`, `mandate` replaces `contract`. The
  cultural rift between Earth's bureaucratic governance and Mars's pragmatic
  approach begins.

### Era I Key Numbers

| Metric | Value |
|---|---|
| Sites | 3 civilian (Terra, Mars, Luna) + 1 military (USSN-Terra) |
| Repos | 4 civilian + 2 compartmented |
| Engineers | ~41 civilian, ~50 military |
| Locales | 2 (TerranEnglish, MartianEnglish) |
| Governance models | 1 civilian (sync) + 1 military (compartmentalized hierarchy) |

### Era I Demographics

| Location | Total population | Engineers | Notes |
|---|---|---|---|
| Terra | 10.2 billion | 15 (civilian) | ISA headquarters |
| Luna | 50 | 0 | Relay/arbiter only |
| Mars | 200 | 6 | First colony |
| USSN (Earth orbit) | 500 | 50 | Fleet stations |
| **Total off-Earth** | **~750** | **56** | |

---

## Era II: The Divergence (Years 5-20) -- "Two Planets, Two Cultures"

**Ticks 1,825 -- 7,300**

- **Year 7 (tick 2,555)**: The **Sino-Russo Joint Space Navy (SRJSN)** independently
  adopts Dunix. Their fork uses a strict hierarchical clearance model with 6
  levels. SRJSN governance is command-driven: no voting, the commanding officer
  ratifies by seal. They add a `chain_of_command` block to their config. SRJSN
  operates from **Tiangong-Ares** station at Mars L1.

- **Year 8 (tick 2,920)**: Mars proposes the **Async Governance Amendment**: 72-hour
  no-objection ratification replaces synchronous voting. Earth resists -- sync
  ratification has worked for 8 years. The proposal stalls.

- **Year 10 (tick 3,650)**: The **Dual Protocol Compromise**. Contracts declare
  `protocol = "sync"` or `"async"`. Both are valid. This is the first expression
  of the Human Protocol / Machine Protocol Duality: the same artifact is
  human-readable governance *and* machine-enforceable schema.

- **Year 12 (tick 4,380)**: Mars forks `dunix-net` into **`mars-net`**. The design
  disagreement: Earth favors store-and-forward networking; Mars insists on
  persistent mesh. Neither side will yield.

- **Year 15 (tick 5,475)**: **Ceres Belt Station** comes online. 8 engineers. The
  Belt adopts Mars as upstream (`topology.Promote`). The **Belt Protocol** is
  radical: it skips `need` artifacts entirely. Oral agreement is sufficient for
  the Belt's tight-knit crew. The Belt Free Association (BFA) is founded --
  partly ideological. "Needs" are what Earth imposes; the Belt has "signals."

- **Year 18 (tick 6,570)**: The Belt develops its stripped-down governance chain:
  oral agreement → mandate → code. The **BeltCreole** locale emerges: `mandate`
  for contract, `edict` for rule, `signal` for need.

- **Year 20 (tick 7,300)**: Mars reaches **code parity** with Earth. Commit volume
  is 50/50. Three distinct civilian governance models coexist: Terra sync, Mars
  async, Belt oral. The military forks operate in parallel, invisible to
  civilians except as opaque nodes in the federation graph.

### Era II Key Numbers

| Metric | Value |
|---|---|
| Sites | 4 civilian (Terra, Luna, Mars, Ceres) + 2 military (USSN-Terra, Tiangong-Ares) |
| Repos | 6 civilian + 4 compartmented |
| Engineers | ~93 civilian, ~200 military |
| Locales | 3 (TerranEnglish, MartianEnglish, BeltCreole) |
| Governance models | 3 civilian (sync, async, oral) + 2 military (hierarchy, chain-of-command) |

### Era II Demographics

| Location | Total population | Engineers | Notes |
|---|---|---|---|
| Terra | 10.5 billion | 50 | ISA + civilian teams growing |
| Luna | 200 | 0 | Relay expansion |
| Mars | 5,000 | 43 | Colony growing fast |
| Belt (Ceres) | 400 | 8 | Mining outpost |
| USSN (stations) | 2,000 | 120 | Fleet expansion |
| SRJSN (Tiangong-Ares) | 1,500 | 80 | Mars L1 station |
| **Total off-Earth** | **~9,100** | **~301** | |

---

## Era III: The Outer Expansion (Years 20-50) -- "The Schism and the Relay"

**Ticks 7,300 -- 18,250**

- **Year 22 (tick 8,030)**: **The Kernel Schism**. Mars proposes capability-based IPC
  inspired by seL4. Earth refuses: ABI stability is sacred. Mars invokes
  `topology.Secede`. The `mars-kernel` fork begins. This is the first true
  governance schism -- not a design disagreement but a philosophical split.

- **Year 25 (tick 9,125)**: The Belt sides with Mars on the kernel. `topology.Promote`
  from `mars-kernel`. Two kernel lines, two worldviews. In the same year,
  **The Ceres Leak**: a Belt civilian engineer's linter crashes when resolving a
  cross-reference to a USSN compartmented artifact. The linter attempted to read
  content it had no clearance for. The crash exposes the fragility of
  compartmentalization at federation scale. This forces the **Opaque Reference
  Protocol**: cross-references to compartmented artifacts return
  `exists: true, clearance: denied` instead of failing. The linter learns to
  skip content validation on opaque refs while still tracking the reference edge
  in the governance graph.

- **Year 28 (tick 10,220)**: **Luna Arbitration**. After 6 years of divergence, Luna
  brokers a truce. The `dunix-compat` shim is created -- a compatibility layer
  between the two kernel ABIs. Neither side loves it. Both use it.

- **Year 30 (tick 10,950)**: **Ganymede Station** (Jupiter system). 50 engineers.
  The **Jovian** locale emerges: `charter` for contract, `signal` for need. Ganymede
  operates with 33-54 minute delay to Earth -- local autonomy is not a preference
  but a physical necessity. SRJSN establishes **Zvezdny** station in Ganymede
  orbit, 15 engineers -- watching Jupiter's resources.

- **Year 33 (tick 12,045)**: Jupiter creates **`jovian-net`** -- the third networking
  fork. Erasure coding with merkle-tree diff synchronization, designed for
  unreliable deep-space links.

- **Year 35 (tick 12,775)**: **The Grand Schism**. Three incompatible kernels, three
  networking stacks, four locales. Cross-site `justifies` links break because
  there is no canonical ID scheme. A need on Mars cannot reference a specification
  on Jupiter. The governance graph fragments.

- **Year 38 (tick 13,870)**: **The Accord of Ceres**. All active sites convene
  (asynchronously, over 6 weeks) and agree to a federation schema: canonical
  `site:ID` notation, Luna as permanent arbiter, and a shared cross-site linter
  protocol. The Accord does not unify governance models -- it creates the
  infrastructure to let them coexist.

- **Year 40 (tick 14,600)**: Post-Accord stability. 12 repos across 4 sites.
  Cross-site linter takes 3 minutes.

- **Year 42 (tick 15,330)**: **The Phobos War**. A brief kinetic conflict between
  USSN and SRJSN forces near Mars. Phobos Relay Station is destroyed. Three
  USSN compartmented repos hosted exclusively on Phobos are lost -- 2 years of
  classified kernel patches gone. No off-site backup existed because
  compartmentalization rules prohibited replication to non-cleared sites. The
  loss is total and unrecoverable.

  The Phobos War triggers the **Disaster Recovery Amendment** to the Accord of
  Ceres: every site must maintain N+1 redundant copies. Compartmented repos must
  have a **dead man's key** -- a recovery key held by Luna arbiter that can
  decrypt repo metadata (artifact IDs, governance graph edges) but not content,
  enabling graph reconstruction without clearance breach.

  > The Phobos War proves that compartmentalization without disaster recovery is
  > a single point of failure. "Phobos Rules" becomes shorthand for backup
  > policy in the Dunix community.

- **Year 45 (tick 16,425)**: **Mars Kernel Reunification**. `topology.Union` after
  23 years of divergence. The capability model wins: Earth adopts it with a POSIX
  shim that becomes first-class. 4,000+ artifacts are reconciled. 340 orphan
  needs are retired. The reunification proves that `topology.Union` works at
  scale -- and that the cost is enormous.

### Era III Key Numbers

| Metric | Value |
|---|---|
| Sites | 5 civilian + 3 military (USSN-Terra, Tiangong-Ares, Zvezdny) |
| Repos | 14 civilian + 6 compartmented |
| Engineers | ~193 civilian, ~350 military |
| Locales | 4 (TerranEnglish, MartianEnglish, BeltCreole, Jovian) |
| Governance models | 4 civilian + 2 military |

### Era III Demographics

| Location | Total population | Engineers | Notes |
|---|---|---|---|
| Terra | 11 billion | 100 | |
| Luna | 500 | 0 | Relay + arbiter |
| Mars | 50,000 | 93 | Rapidly growing |
| Belt (Ceres) | 3,000 | 20 | BFA established |
| Ganymede | 2,000 | 50 | Jupiter system hub |
| USSN (Olympus + Phobos) | 8,000 | 200 | Phobos lost Year 42 |
| SRJSN (Tiangong + Zvezdny) | 5,000 | 150 | Mars L1 + Ganymede orbit |
| **Total off-Earth** | **~68,500** | **~613** | |

---

## Era IV: The Deep Reach (Years 50-100) -- "Titan, the Philosopher"

**Ticks 18,250 -- 36,500**

- **Year 50 (tick 18,250)**: **Titan Station** (Saturn system). 30 engineers. The
  **Titanese** locale is unlike any before it: `covenant` for contract, `intuition`
  for need, `thesis` for specification. Every covenant carries a mandatory
  `reflection` block -- a prose section explaining not just *what* the artifact
  does but *why it matters* and *what it might break*.

- **Year 55 (tick 20,075)**: Titan proposes **`dunix-conscience`** -- a subsystem
  that evaluates the ethical implications of system calls. Earth and Mars dismiss
  it as philosophical overhead. USSN, however, quietly forks it into the
  classified **`ussn-conscience`** -- they recognize its value for weapons-system
  safety analysis.

- **Year 60 (tick 21,900)**: **Titan Philosophical Secession**. The **Titanos** fork.
  ABI compatible with the reunified kernel, but governance-independent. Titan's
  position: governance without reflection is governance without conscience.

- **Year 65 (tick 23,725)**: **The Enceladus Incident**. A life-support system crash
  on Enceladus Station -- an integer overflow in the atmosphere controller. The
  `dunix-conscience` subsystem had flagged the relevant code path 6 months prior
  as a critical-urgency need with acceptance criteria that were never addressed.
  The warning was ignored. After the incident, all sites adopt `reflection` blocks
  on critical-urgency needs. Urgency propagation goes federation-wide.

  In the aftermath, USSN **declassifies `ussn-conscience`** -- their fork that had
  been running classified for 10 years. The declassification process itself
  becomes a governance primitive: `compartment.declassify` transitions an artifact
  from opaque to public while preserving the sealed audit chain. The civilian
  community is stunned to discover the military had a working conscience module
  years before anyone else.

  > The Enceladus Incident becomes the canonical example of why governance
  > matters. The "Enceladus Index" -- the ratio of flagged-but-ignored warnings
  > to actual incidents -- becomes a standard metric.

- **Year 70 (tick 25,550)**: **The Outer Worlds Compact**. Jupiter and Saturn form a
  sub-federation. Shared `deep-net` networking stack, independent from the Inner
  System. Sync cycle: once per Jovian day (~10 hours). The Outer Worlds Council
  (OWC) is established as a diplomatic counterweight to military pragmatism.

- **Year 75 (tick 27,375)**: **Miranda Station** (Uranus system). 8 engineers. Ice
  mining operations. A single review cycle through 4 relay hops takes 12+ hours.
  USSN establishes **Vesta** station in the Belt (25 engineers) -- resource
  security for outer system operations.

- **Year 80 (tick 29,200)**: Miranda invents **Speculative Merge**: push with
  `speculative = true`, deploy locally, rollback if upstream rejects. The protocol
  is born of necessity -- waiting 12 hours for a review on a critical ice-mining
  patch is not viable.

- **Year 85 (tick 31,025)**: **The Quiet War**. A prolonged cyber campaign targets
  SRJSN's governance graphs. 40% of cross-reference links are corrupted --
  `justifies` and `implements` links point to nonexistent or wrong targets.
  Artifacts still exist but the traceability chains are broken.

  SRJSN's sealed governance chains save them: every artifact's seal includes a
  cryptographic hash of the governance subgraph at seal time. By walking seal
  chains backwards, the uncorrupted graph can be reconstructed from metadata
  alone -- even when the content is corrupted.

  The Quiet War triggers federation-wide adoption of **Sealed Governance Chains**:
  every seal includes a merkle root of the artifact's upstream references.

  > The Quiet War proves that the governance graph is as critical as the code it
  > governs. Destroy the graph, and the code becomes an undifferentiated mass of
  > files with no traceability, no lifecycle, and no accountability.

### Era IV Key Numbers

| Metric | Value |
|---|---|
| Sites | 8 civilian (+ Europa, Enceladus) + 4 military (Olympus, Vesta, Tiangong-Ares, Zvezdny) |
| Repos | 30+ civilian + 10 compartmented |
| Engineers | ~400 civilian, ~540 military |
| Locales | 6 (+ Titanese, Compact) |
| Governance models | 5 civilian + 2 military + sealed chains (post-Quiet War) |

### Era IV Demographics

| Location | Total population | Engineers | Notes |
|---|---|---|---|
| Terra | 11.5 billion | 180 | |
| Luna | 1,000 | 12 | Arbiter infrastructure growing |
| Mars | 500,000 | 130 | Major colony |
| Belt (Ceres) | 20,000 | 45 | BFA politically mature |
| Ganymede | 15,000 | 80 | Jupiter hub |
| Europa | 3,000 | 15 | Science station |
| Titan | 5,000 | 50 | Philosopher colony |
| Enceladus | 1,000 | 8 | Post-incident rebuild |
| Miranda | 200 | 8 | Ice mining |
| USSN (Olympus + Vesta) | 30,000 | 350 | Belt + Mars orbit |
| SRJSN (Tiangong + Zvezdny) | 20,000 | 190 | Post-Quiet War recovery |
| **Total off-Earth** | **~595,200** | **~1,068** | |

---

## Era V: The Edge (Years 100-150) -- "Neptune and the Void"

**Ticks 36,500 -- 54,750**

- **Year 100 (tick 36,500)**: **Triton Station** (Neptune system). 4 engineers. 4+
  hour delay to Earth. The **Hermetic Protocol**: unanimous 4-person decisions,
  synced upstream in daily batches. The **Tritonic** locale has 12 governance
  keywords -- the most specific of any locale, because when 4 people must agree
  unanimously, precision of language is survival.

- **Year 110 (tick 40,150)**: Triton's isolation produces innovations impossible
  in the Inner System: a single-digit-user scheduler (no contention by design),
  a decades-long-uptime filesystem (no one is coming to fix it), and memory
  management optimized for $10k/gram hardware.

- **Year 120 (tick 43,800)**: **Solar Census**.

  **Civilian sites:**

  | Site | Repos | Engineers | Locale | Governance |
  |---|---|---|---|---|
  | Terra | 8 | 200 | TerranEnglish | Sync ratification |
  | Luna | 0 | 12 | TerranEnglish | Arbiter (no code) |
  | Mars | 12 | 180 | MartianEnglish | 72h async |
  | Ceres | 6 | 45 | BeltCreole | Oral + mandate |
  | Ganymede | 10 | 80 | Jovian | Compact async |
  | Europa | 3 | 15 | Jovian | Compact async |
  | Titan | 8 | 50 | Titanese | Covenant + reflection |
  | Enceladus | 2 | 8 | Titanese | Covenant |
  | Miranda | 2 | 8 | Compact | Speculative merge |
  | Triton | 1 | 4 | Tritonic | Hermetic (unanimous) |
  | **Civilian total** | **52** | **602** | **7** | **6** |

  **Military sites (opaque nodes in federation graph):**

  | Site | Repos | Engineers | Faction | Governance |
  |---|---|---|---|---|
  | Olympus (Mars orbit) | 4 | 60 | USSN | Compartmentalized hierarchy |
  | Vesta (Belt) | 2 | 25 | USSN | Compartmentalized hierarchy |
  | Tiangong-Ares (Mars L1) | 3 | 40 | SRJSN | Chain of command |
  | Zvezdny (Ganymede orbit) | 2 | 15 | SRJSN | Chain of command |
  | **Military total** | **11** | **140** | **2** | **2** |

  | **Grand total** | **63** | **742** | **7 locales** | **8 governance models** |
  |---|---|---|---|---|

- **Year 130 (tick 47,450)**: **The Grand Union**. The Solar Compact is ratified.
  Every site keeps its locale and governance model. What is unified: federation
  schema, canonical IDs, Luna arbiter role, reflection blocks on critical needs,
  speculative merge protocol, hermetic protocol recognition.

  The Grand Union includes a **Compartmentalization Annex**: the civilian
  federation formally acknowledges opaque nodes. The linter skips their content.
  Canonical IDs work across the clearance boundary. Declassification has a
  formal protocol with ledger-preserved audit chains. Dead man's keys are
  mandatory for all compartmented repos.

- **Year 150 (tick 54,750)**: **Steady state**. 100+ repos, 1,000+ engineers.
  The longest traceability chain spans 5 sites and 3 locales. Federation-wide
  linter runtime: 45 minutes. Military sites contribute 15+ compartmented repos
  as opaque nodes. The Enceladus Index is computed federation-wide; the Quiet
  War's sealed governance chains are standard on every artifact.

### Era V Key Numbers

| Metric | Value |
|---|---|
| Sites | 10 civilian + 4 military |
| Repos | 52 civilian + 11 military → 100+ total |
| Engineers | 602 civilian + 140 military → 1,000+ |
| Locales | 7 |
| Governance models | 8 (6 civilian + 2 military) |
| Kernel lines | 3 (unified ABI, divergent governance) |

### Era V Demographics

| Location | Total population | Engineers | Notes |
|---|---|---|---|
| Terra | 12 billion | 250 | |
| Luna | 3,000 | 15 | Arbiter + dead-man's-key vault |
| Mars | 2,000,000 | 250 | Self-sustaining colony |
| Belt (Ceres) | 80,000 | 60 | BFA sovereign |
| Ganymede | 60,000 | 100 | OWC co-capital |
| Europa | 10,000 | 20 | Ocean research |
| Titan | 25,000 | 60 | OWC co-capital, philosophical center |
| Enceladus | 3,000 | 10 | Rebuilt, memorial station |
| Miranda | 1,000 | 10 | Ice export hub |
| Triton | 50 | 4 | Deep edge |
| USSN (Olympus + Vesta) | 80,000 | 400 | Full fleet presence |
| SRJSN (Tiangong + Zvezdny) | 50,000 | 250 | Post-Quiet War rebuilt |
| **Total off-Earth** | **~2,312,050** | **~1,429** | |

Engineer-to-population ratio: ~1:300 civilian, ~1:10 military (military is
engineer-heavy by operational necessity).

---

## Milestone Event Index

Quick-reference for simulation scripting. All ticks assume 1 tick = 1 day.

| Tick | Year | Event | Type |
|---|---|---|---|
| 0 | 0 | Genesis: 4 repos, Terra bootstrap | Genesis |
| 365 | 1 | Mars colony joins | NewSiteJoins |
| 1,095 | 3 | Luna Relay Station | NewSiteJoins |
| 1,095 | 3 | USSN adopts Dunix, first compartmented repos | MilitaryAdoption |
| 1,825 | 5 | MartianEnglish locale | LocaleCreation |
| 2,555 | 7 | SRJSN adopts Dunix, chain-of-command governance | MilitaryAdoption |
| 4,380 | 12 | mars-net fork | RepoFork |
| 5,475 | 15 | Ceres Belt station, BFA founded | NewSiteJoins |
| 6,570 | 18 | BeltCreole locale | LocaleCreation |
| 8,030 | 22 | Kernel Schism (Mars secedes) | Schism |
| 9,125 | 25 | Ceres Leak -- Opaque Reference Protocol | Incident |
| 10,950 | 30 | Ganymede station | NewSiteJoins |
| 10,950 | 30 | Jovian locale | LocaleCreation |
| 10,950 | 30 | SRJSN establishes Zvezdny (Ganymede orbit) | MilitaryAdoption |
| 12,045 | 33 | jovian-net fork | RepoFork |
| 12,775 | 35 | Grand Schism | Schism |
| 13,870 | 38 | Accord of Ceres | Accord |
| 15,330 | 42 | Phobos War -- catastrophic data loss, DR Amendment | War |
| 16,425 | 45 | Mars kernel reunification | Union |
| 18,250 | 50 | Titan station | NewSiteJoins |
| 18,250 | 50 | Titanese locale | LocaleCreation |
| 21,900 | 60 | Titan philosophical secession (Titanos) | Schism |
| 23,725 | 65 | Enceladus Incident | Incident |
| 23,725 | 65 | USSN declassifies ussn-conscience | Declassification |
| 25,550 | 70 | Outer Worlds Compact, OWC established | Accord |
| 27,375 | 75 | Miranda station | NewSiteJoins |
| 27,375 | 75 | USSN establishes Vesta (Belt) | MilitaryAdoption |
| 29,200 | 80 | Speculative Merge invented | ProtocolInnovation |
| 31,025 | 85 | The Quiet War -- governance graph corruption | War |
| 36,500 | 100 | Triton station | NewSiteJoins |
| 36,500 | 100 | Tritonic locale | LocaleCreation |
| 43,800 | 120 | Solar Census (with military sites) | Census |
| 47,450 | 130 | Grand Union + Compartmentalization Annex | Accord |
| 54,750 | 150 | Steady state | EndState |

---

## Governance Models

| Model | Origin | Mechanism | Sites |
|---|---|---|---|
| Sync ratification | Terra (Year 0) | Quorum vote, real-time | Terra |
| 72h async | Mars (Year 8) | No-objection window | Mars |
| Oral + mandate | Ceres (Year 15) | Verbal agreement → written mandate | Ceres |
| Compact async | Ganymede (Year 30) | Federation-scoped async vote | Ganymede, Europa |
| Covenant + reflection | Titan (Year 50) | Mandatory reflection block | Titan, Enceladus |
| Speculative merge | Miranda (Year 80) | Push-deploy-rollback | Miranda |
| Hermetic | Triton (Year 100) | Unanimous 4-person decision | Triton |
| Compartmentalized hierarchy | USSN (Year 3) | Clearance levels + codeword access + async | Olympus, Vesta |
| Chain of command | SRJSN (Year 7) | CO ratifies by seal, sealed orders | Tiangong-Ares, Zvezdny |

The simulation tracks 8 distinct models (6 civilian + 2 military).

---

## Compartmentalization

Information compartmentalization in Dunix follows the [need-to-know principle](https://en.wikipedia.org/wiki/Compartmentalization_(information_security)):
access to artifact content is restricted by clearance level and codeword,
independent of authority level.

### Clearance Levels

| Level | USSN | SRJSN equivalent | Visibility |
|---|---|---|---|
| 0 | Unclassified | Open | Full content visible |
| 1 | Confidential | Restricted | Content visible to cleared operators |
| 2 | Secret | Secret | Content visible, title redacted for outsiders |
| 3 | Top Secret | Supreme | Opaque: ID + kind only, no title, no content |

### Key Primitives

- **Opaque reference**: Cross-reference to a compartmented artifact returns
  `exists: true, clearance: denied`. The governance graph edge is tracked but
  content validation is skipped.
- **Dead man's key**: Encrypted metadata export (IDs, graph edges, no content)
  held by Luna arbiter for disaster recovery.
- **Declassification**: `compartment.declassify` transitions an artifact from
  opaque to public, preserving the sealed audit chain.
- **Sealed governance chain**: Every seal includes a merkle root of the
  artifact's upstream references (adopted federation-wide after the Quiet War).

### Timeline

| Year | Event | Primitive introduced |
|---|---|---|
| 3 | USSN adoption | `compartment` block, clearance levels, opaque artifacts |
| 7 | SRJSN adoption | `chain_of_command`, sealed orders |
| 25 | Ceres Leak | Opaque Reference Protocol |
| 42 | Phobos War | Dead man's key, Disaster Recovery Amendment |
| 65 | USSN declassification | `compartment.declassify` |
| 85 | Quiet War | Sealed governance chains (merkle roots in seals) |
| 130 | Grand Union Annex | Compartmentalization formally recognized in federation |

---

## Locales

| Locale | Origin | contract | rule | specification | need | architecture | docs |
|---|---|---|---|---|---|---|---|
| TerranEnglish | Terra | contract | rule | specification | need | architecture | docs |
| MartianEnglish | Mars | mandate | capability | specification | need | structure | docs |
| BeltCreole | Ceres | mandate | edict | spec | signal | blueprint | log |
| Jovian | Ganymede | charter | signal | specification | need | architecture | record |
| Titanese | Titan | covenant | intuition | thesis | need | form | reflection |
| Compact | Miranda | pact | rule | spec | need | plan | note |
| Tritonic | Triton | protocol | axiom | thesis | signal | schema | entry |

---

## Wars and Catastrophic Data Loss

### The Phobos War (Year 42, tick 15,330)

**Cause**: USSN-SRJSN territorial dispute over Mars approach corridors.

**Event**: Brief kinetic conflict. Phobos Relay Station destroyed by SRJSN
kinetic strike. ISA declares neutrality. Luna continues arbitration.

**Data loss**: 3 USSN compartmented repos hosted exclusively on Phobos are
permanently lost. 2 years of classified kernel patches, 847 artifacts, and the
complete governance graph for fleet communication protocols -- gone. No off-site
backup existed because compartmentalization rules prohibited replication to
non-cleared sites.

**Aftermath**: Disaster Recovery Amendment to the Accord of Ceres. "Phobos Rules"
become standard: N+1 redundant copies for all repos, dead man's keys for all
compartmented repos, Luna vault for encrypted metadata.

**Simulation tests**: Repo destruction event, governance graph repair from partial
metadata, dead-man's-key decryption protocol.

### The Quiet War (Year 85, tick 31,025)

**Cause**: Unknown attacker (suspected USSN, never confirmed) conducts a prolonged
cyber campaign against SRJSN governance infrastructure.

**Event**: Over 60 days, 40% of cross-reference links in SRJSN repos are
corrupted. `justifies` links point to wrong targets. `implements` links reference
nonexistent specifications. The governance graph is intact structurally but
semantically poisoned -- the traceability chain tells lies.

**Data loss**: No artifacts are deleted. Instead, the governance graph's integrity
is destroyed. SRJSN cannot determine which specifications fulfill which needs,
which contracts implement which specs, or which docs describe which components.
The code works. The governance is meaningless.

**Recovery**: SRJSN's sealed governance chains contain cryptographic hashes of
the governance subgraph at seal time. By walking the seal chain backwards from
the most recent seals, the pre-corruption graph state can be reconstructed from
metadata alone. Recovery takes 4 months but is complete.

**Aftermath**: Federation-wide adoption of sealed governance chains. Every seal
now includes a merkle root of upstream references. The Quiet War proves that
the governance graph is a target of strategic value equal to the code itself.

**Simulation tests**: Governance graph corruption event (randomize 40% of
cross-refs), seal-chain-based reconstruction algorithm, integrity verification,
recovery time measurement.
