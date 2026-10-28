---
type: Design
title: Agent open questions
description: Undecided design questions for agents, agent launches, and control plane server supervision, linked to the sections they concern.
tags: [afakt, agents, open-questions]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent open questions

These questions cover undecided aspects of agents, agent launches, and supervision of control plane servers. Each question links to the relevant sections and keeps its ID when answered or removed. afakt never reuses an ID.

## Agents

### `OQ-AGT-1`

Which process state does restoring an agent snapshot preserve, and which channels or helpers must restart? How does the agent harness restore its session state from the agent checkpoint taken together with that snapshot? Compatible VM restoration is the leading option for preserving live process state, subject to qualification for each backend and agent harness. Recovering the agent workspace and the agent checkpoint without an agent snapshot is a candidate cold alternative that loses live state. Adopting it requires qualification of complete native-history restoration. Migration remains a separate question. See [Suspend, resume, and restart](agent-lifecycle.md#Suspend,%20resume,%20and%20restart).

### `OQ-AGT-2`

How should afakt handle a suspend request during an active turn before taking an agent snapshot and an agent checkpoint together? Waiting for an idle turn boundary is the leading option for capturing consistent agent workspace and session state, at the cost of delaying suspension while the turn continues. The wait bound and handling of a turn that doesn't become idle remain open. See [Suspend, resume, and restart](agent-lifecycle.md#Suspend,%20resume,%20and%20restart).

### `OQ-AGT-3`

Which process and device state must an agent snapshot preserve, and which session state must the agent harness preserve in its agent checkpoint? How does afakt bind an agent snapshot and an agent checkpoint taken together at the same idle turn boundary to the frozen workspace disk delta? A verified set captured after writers are quiescent is the leading option, at the cost of pausing work and validating restored state before accepting turns. The set must bind workspace, native-session, memory, and external device state. Its exact capture scope remains open. What independently enforced boundary proves that every agent workspace writer in host mode is quiescent, and what happens when that proof is unavailable? See [Agent checkpoints](agent.md#Agent%20checkpoints).

### `OQ-AGT-4`

How should a conversation branch created by rewind represent external effects? The leading option for preserving known external actions retains the original branch with a separate effect ledger, at the cost of retaining audit data alongside the new branch. The ledger must record incomplete coverage because tool and permission records don't cover every effect. Rewind can't undo those actions, and their presentation remains open. See [Agent checkpoints](agent.md#Agent%20checkpoints).

### `OQ-AGT-5`

May `afakt agent mcp add` attach a server the deployment hasn't configured? Profiles select configured servers and the agent harness uses one gateway endpoint. How do local server attachments fit those restrictions? Registration in the deployment's server catalog before attachment is the leading option for keeping gateway grants tied to known server identities. It adds an approval or configuration step for a new server. One gateway uniform resource locator (URL) constrains configured access but can't itself prevent local tool processes from using another protocol. The permitted attachment model remains open. See [Owned resources](agent.md#Owned%20resources).

### `OQ-AGT-6`

What are agentmail's delivery guarantees, ordering, acknowledgement, scheduling, and undo semantics? A durable outbox with retries and recipient deduplication is the leading delivery option, at the cost of retained replay records. Cancellation ends when dispatch may have begun, and external effects require their own idempotency guarantees. Ordering, scheduling, and acknowledgement policy remain open. See [Agentmail, groups, and subscriptions](agent.md#Agentmail,%20groups,%20and%20subscriptions).

### `OQ-AGT-7`

What identity signs agentmail, how does historical verification prevent replay, and how do signatures relate to signed session messages? Signing from authenticated sender context with retained verification keys and replay records is the leading option. It supports historical verification but requires retention and revocation rules. By themselves, signatures don't prove app acknowledgement or external completion. The signing contract and its relationship to session signatures remain open. See [Agentmail, groups, and subscriptions](agent.md#Agentmail,%20groups,%20and%20subscriptions).

### `OQ-AGT-8`

Where does memory live, what does a query match, and how do copying, compaction, retention, and access control affect it? Immutable memory versions with authorized lexical queries and atomic content, index, and reference updates are the leading option for exact retrieval. Copying an authorized version gives the destination independent content, while compaction preserves pinned versions. This keeps access and retention explicit but doesn't supply semantic ranking or secure erasure. The storage service and query language remain open. See [Memory, skills, and artifacts](agent.md#Memory,%20skills,%20and%20artifacts).

### `OQ-AGT-10`

How's agent ownership represented in the agent launch spec, AgentLaunch resource, and provenance? Agent launch IDs remain distinct from agent identity. An immutable agent and session binding on each accepted agent launch is the leading option, with agent mediator requests checked against that binding and the active ownership fence. It preserves historical provenance when ownership changes but adds coordination between agent launch acceptance, transfer, and request admission. Public field names remain open. See [Owned resources](agent.md#Owned%20resources).

### `OQ-AGT-11`

How does durable ownership protect the current agent workspace and session transcript when agent launch evidence expires? Which references keep workspace images, revisions, agent snapshots, and shared artifacts from collection? Independent durable roots for the agent's workspace, transcripts, agent checkpoints, and shared content are the leading option. Reference creation and collection share a fence, and unresolved releases keep their roots. This protects content beyond agent launch expiry but can keep extra data and delay collection. Retention periods and the authoritative catalog remain open. What should the API return when a retention hold, pin, or copy reference arrives after collection commits its deletion intent? See [Owned resources](agent.md#Owned%20resources).

### `OQ-AGT-12`

How does migration transfer agent workspace and conversation state to another agent host for a new agent launch linked to its predecessor? A verified cold transfer after observed source termination is the leading candidate across incompatible backends. It must preserve the complete workspace and native conversation state, but migration with real agent harnesses and VM backends remains unqualified. Processes and channels restart, and an absent response doesn't establish termination. See [Export, import, and migrate](agent.md#Export,%20import,%20and%20migrate).

### `OQ-AGT-13`

When do `start`, `run`, `restart`, and `restore` address an existing agent versus create another? Reusing an agent preserves its durable identity, while creating another separates ownership and history. Command defaults remain open under either option. See [Lifecycle](agent-lifecycle.md#Agent%20states).

### `OQ-AGT-14`

What ownership and status apply to an uncertain launch or stop? Retaining ownership until afakt observes termination is the leading rule. It prevents overlapping replacements but can leave the agent unavailable during reconciliation. Intermediate status names remain open. See [Lifecycle](agent-lifecycle.md#Agent%20states).

### `OQ-AGT-15`

How should deletion handle created or suspended agents, and which states count as unused for prune? Restricting removal to agents with resolved ownership of their agent launches avoids discarding an uncertain agent launch, but can delay cleanup. A broader rule needs explicit treatment of retained agent workspace and session state. Deletion behavior and prune eligibility remain open. See [Lifecycle](agent-lifecycle.md#Agent%20states).

### `OQ-AGT-16`

For forks, agent snapshots, archives, Open Container Initiative (OCI) exports, and imports, which resources does afakt copy, reference, or exclude? A content-verified envelope with explicit resource references and destination reauthorization is the leading candidate. Keeping historical audit data separate from active identity and credentials avoids transferring authority with content, but portability requires a supported metadata contract, including extended attributes, and qualified native-store restoration. Identity mapping and the complete resource set remain open. See [Export, import, and migrate](agent.md#Export,%20import,%20and%20migrate).

### `OQ-AGT-17`

Should unarchive always return an agent to stopped state? Returning to stopped state is the leading option because it avoids starting work during restoration, at the cost of a separate start operation. Whether any case permits automatic start remains open. See [Agent snapshots and the snapshot registry](agent.md#Agent%20snapshots%20and%20the%20snapshot%20registry).

### `OQ-AGT-18`

How do per-agent credential attachments map to references on the profile and agent launch? How do authorization roles and organization and project scopes affect that mapping? Intersecting an agent launch's attached references with those the pinned profile permits is the leading candidate check, alongside target-resource authorization. It limits attachment choices. An attachment change changes the agent spec and takes effect through a new agent launch at a turn boundary. It doesn't establish the profile's own authority or the credential owner's permitted destinations and headers. Repository profile self-grant and credential domain bounds, including the bound on agent launch references, remain on hold. Authorization roles remain distinct from reporting labels. Inheritance remains open. Each credential type defines custody and exchange. The control plane can check attachments when it supplies a credential. How changes affect credentials already supplied remains open. See [Secrets, credentials, and trust material](agent.md#Secrets,%20credentials,%20and%20trust%20material).

### `OQ-AGT-19`

What rules govern ordering, retries, retention, and inactive recipients for group broadcasts and subscription deliveries? A durable outbox with acceptance-time membership and event deduplication is the leading delivery option. It makes retries stable but retains accepted deliveries after membership changes. Queueing for suspended or stopped recipients avoids implicit resource use. Handling archived recipients remains a separate open question. Waking requires separately authorized, retry-safe lifecycle requests. Their completion semantics remain open, as do ordering and retention bounds. See [Agentmail, groups, and subscriptions](agent.md#Agentmail,%20groups,%20and%20subscriptions).

### `OQ-AGT-20`

What access, content isolation, retention, and indexing rules apply to artifact publication? Authorized, expiring publication of immutable artifacts on separate origins is the leading candidate. Download-only delivery reduces active-content exposure but excludes inline applications. Browser isolation and atomic content, index, and reference updates across the selected storage services still need qualification. The storage and query contracts remain open, and revocation can't recall admitted downloads or client-held copies. See [Memory, skills, and artifacts](agent.md#Memory,%20skills,%20and%20artifacts).

### `OQ-AGT-21`

How do standalone sandbox identities, keychain authorizations, Secure Shell (SSH) and GNU (GNU's Not Unix) Privacy Guard (GPG) key stores, known hosts, and certificate trust relate to agent and agent launch authority? What controls key use, rotation, and revocation? Purpose-scoped delegated key use with private material outside the agent sandbox is the leading option. A control plane Application Programming Interface (API) operation that exchanges a request authenticated with a SPIFFE Verifiable Identity Document (SVID) for short-lived credentials, such as SSH certificates from an internal certificate authority, is the leading way to deliver that use. Each credential type defines custody: exchangeable credentials such as Open Authorization (OAuth) refresh tokens stay central and yield short-lived credentials for the host-side credential broker, while a static secret that the control plane can't exchange stays in the credential broker's memory outside the guest for one agent launch. Credential delivery requires the control plane, but each later use needn't request a new exchange. Keeping durable revocation state and rebuilding trust contexts prevent old authority from returning after restart, but rotation also requires closing affected connections. Retaining public keys supports historical verification without permitting new signing. Compromise semantics and native key-store integration remain open. See [Secrets, credentials, and trust material](agent.md#Secrets,%20credentials,%20and%20trust%20material).

### `OQ-AGT-22`

How do agent budgets account for usage across agent launches and external services? What happens at exhaustion, and which operations remain available? Durable reservations before dispatch, followed by reconciliation with actual usage, are the leading accounting option. Keeping unknown bills reserved prevents premature reuse but can withhold budget indefinitely. A hard spending bound also needs an enforceable upstream usage ceiling. Observation alone can only record an overage. Exhaustion behavior and permitted operations remain open. When a subscription runs out, may that permit spending on another service, or must the agent profile grant that fallback separately within the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60)? See [Budgets and usage](agent.md#Budgets%20and%20usage).

### `OQ-AGT-23`

Which network paths does quarantine block, including brokered API calls, Agent Client Protocol (ACP) control traffic, telemetry, Model Context Protocol (MCP) calls, and forwarded ports? What permissions restore connectivity? Explicit blocked-path sets with checks at dispatch and revocation of active flows are the leading mechanism. Denying new connections alone leaves established Transmission Control Protocol (TCP) and streaming traffic usable. Closing those flows enforces quarantine but interrupts in-flight work. Whether ACP, telemetry, and brokered control remain available, and who may restore them, remains open. See [Kill, signal, and quarantine](agent-lifecycle.md#Kill,%20signal,%20and%20quarantine).

### `OQ-AGT-24`

How do steer and queue-message apply to workflow-controlled or headless agent launches without widening their writer rules? What boundary does `wait` observe when the transcript persists after the agent harness exits? When afakt replaces an agent launch, should queued input, waits, and transcript subscriptions stay with that agent launch or follow its successor? See [Attach and messages](agent.md#Attach%20and%20messages).

### `OQ-AGT-25`

Which workspace revision does restart use, and how do agent workspace commits and resets relate to agent launch copies and workspace image publication? Pinning the selected revision before restart is the leading option for stable retries, at the cost of retaining that revision through preparation. Both the latest retained agent workspace and an explicitly selected older revision fit this mechanism. Independent writable copies preserve history, while workspace image publication needs separate authorization. The default revision and reset semantics remain open. See [Workspace revisions and the workspace registry](agent.md#Workspace%20revisions%20and%20the%20workspace%20registry).

### `OQ-AGT-26`

What authority permits `agent exec` and `agent shell` commands outside the agent sandbox? The provider confines the agent harness and its helpers, so running a command outside that confinement, and choosing the command's sandbox mode, needs authority of its own. A separate operator grant bound to the target agent sandbox and requested sandbox mode is the leading option. It keeps an agent's agent launch authority from becoming authority to execute outside confinement, but adds a distinct authorization path. Which operators receive that grant remains open, as do the permitted modes. See [Sandboxes](agent.md#Sandboxes).

### `OQ-AGT-27`

What happens to work in flight and the session when a change to the agent spec replaces its agent launch at a turn boundary? Finishing the active turn before replacement is the leading option for preserving its work, at the cost of delaying the change. The wait bound, treatment of an unfinished turn, and session continuation mechanism remain open. See [Agent spec](agent.md#Agent%20spec).

### `OQ-AGT-28`

What do rewind and resume restore in host modes and container managers, which have no agent snapshot, and how does that restored state relate to the agent harness's agent checkpoint and the agent workspace? See [Agent checkpoints](agent.md#Agent%20checkpoints) and [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume).

### `OQ-AGT-29`

Who may transfer an agent to another owner, and must the recipient accept before ownership changes? Explicit transfer and receive grants are the leading option, preserving authority on both sides at the cost of an acceptance step. How should transfer affect access to existing history and attached resources? See [Owned resources](agent.md#Owned%20resources).

### `OQ-AGT-30`

Which component controls the combined outcome, retry policy, and cancellation of an agent launch request across its successor agent launches? An explicit link from the request to each agent launch could preserve these semantics without treating an agent spec generation as an attempt counter. Which resource controls that link, and when's the request complete? See [Agent launch request](agent.md#Agent%20launch%20request).

### `OQ-AGT-31`

Which grants permit agent snapshot and agent checkpoint upload, download, restoration, and deletion? How do the snapshot registry and object store bind captured content to its owner and agent launch as well as its compatible workspace? How do they protect its confidentiality and integrity during storage and transfer? See [Agent snapshots and the snapshot registry](agent.md#Agent%20snapshots%20and%20the%20snapshot%20registry).

### `OQ-AGT-32`

Which authorization semantics should afakt use when comparing a policy proposal with the active policy? Different treatment of evaluation errors can make a change appear to preserve authority while admitting new requests. Comparison could use the runtime decision procedure directly or a restricted policy language whose error behavior matches it. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-33`

Which principals, actions, resources, and attributes must a policy comparison cover before afakt can approve the proposed change? Omitting a reachable request can hide added authority. The comparison could cover a bounded authorization domain enforced at runtime or every request admitted by the policy schema. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-34`

How should approval of a policy proposal bind the entity records and credential associations that determine its authority? Those records can change permissions without changing policy text. Approval could pin their revisions or include their transitions in the comparison or require a new review when they change. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-35`

How should afakt classify a policy proposal that both grants and removes authority? Treating only its net effect as a reduction could hide a newly permitted action. Review could classify it as a distinct mixed change or require the same approval as any grant of added authority. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-36`

Which policy proposals may receive delegated approval after comparison with the active policy? A reduction in authority can interrupt work even when it doesn't add permissions. Delegation could cover only policies with the same authorization behavior or specified reductions or changes within an operator-defined scope. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-37`

What should afakt do when analysis of a policy proposal times out or can't cover part of the policy? An inconclusive result supplies no basis for automatic approval. afakt could reject the proposal or keep the active policy while requesting an authorized review. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-38`

Beyond authorization rules, what enforcement assumptions must a policy proposal's approval include? Policy comparison alone can't establish an allowed action's effects on budgets, agent workspaces, or agent snapshots. Approval could require separately reviewed enforcement invariants. Or it could require an analysis that includes those effects. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Policies, roles, and organizational scope](agent.md#Policies,%20roles,%20and%20organizational%20scope).

### `OQ-AGT-39`

What stream and signal behavior must `agent exec` and `agent shell` provide across sandbox modes? Differences in terminal handling or end-of-input behavior can change command results and prevent graceful shutdown. The commands could expose byte streams with explicit signal operations or a terminal interface or both with declared provider support. See [Sandboxes](agent.md#Sandboxes).

### `OQ-AGT-40`

Where does `afakt agent exec` run for each agent sandbox provider, and how does it bind the caller, agent launch, workspace, inherited environment, and egress policy? What prevents concurrent commands from invalidating workspace capture, and which authority stops every command descendant during cancellation, suspension, or teardown? Which command events establish attribution without recording private content? See [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox).

## Agent host

### `OQ-HST-1`

After a whole-machine restart, how does the host reconcile its agent launches whose VMs and agent mediators stopped, and what observations determine their outcomes and slot release? Worker restart keeps those processes alive, while lease expiry alone doesn't change agent launches or reservations. See [Restart and recovery](agent-host.md#Restart%20and%20recovery).

### `OQ-HST-2`

What bounds on control latency, failure rate, memory use, and free disk space must an agent host meet during sustained work? How should agent density and workspace revision capture affect those bounds alongside restore latency? How should those bounds divide processor time and open-file capacity between shared host services and individual agent launches? See [Fleet membership and capacity](agent-host.md#Fleet%20membership%20and%20capacity).

### `OQ-HST-3`

What evidence may permit enrollment of an agent host? Proof of a key, a network identity, or device integrity establishes different facts about the machine requesting admission. Enrollment could require operator approval or a channel-bound network assertion or device attestation with explicit key, ownership, and software claims. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-HST-4`

How should an agent host regain authority after sleep or credential expiry or loss of its enrollment key? Recovery must distinguish the original host from a replacement using copied state. afakt could require fresh operator approval or independently provisioned credentials or renewed platform attestation. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-HST-5`

How should agent host enrollment behave when platform attestation is unsupported or unavailable? A fallback can admit a device without the evidence normally required of it. afakt could reject enrollment or permit only a restricted role or require independent operator approval for a declared device class. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-HST-6`

What must revocation of an agent host stop, and within what propagation bound? Preventing fresh enrollment alone can leave issued identities and active connections usable. Revocation could stop active connections immediately or allow a bounded drain while denying new use and renewal. See [Trust boundary](agent-host.md#Trust%20boundary).

### `OQ-HST-7`

How should an agent host worker using a remote provider account restrict management operations and authenticate callbacks for each agent launch? Account-level access can otherwise let one launch affect another launch's lifecycle. The worker could enforce launch ownership on every operation or partition accounts by trust scope or use delegated credentials limited to one launch. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Trust boundary](agent-host.md#Trust%20boundary).

### `OQ-HST-8`

Which user runs each per-launch agent mediator and agent proxy, and what separates their authority from the dedicated worker user and dedicated host-mode agent user? How does that choice preserve independent launch data paths without settling the on-hold host isolation baseline? See [Agent host](network-topology.md#Agent%20host).

### `OQ-HST-9`

How does an agent host provide host-level name resolution and authorized connectivity to agent ports, previews, services, and cluster services while excluding unauthorized host-mode agent processes? Who installs the name routing, and how do host clients establish certificate trust? A host-level network service with one loopback reverse proxy and process or user identity checks is one option. See [Agent host access](agent-overlay.md#Agent%20host%20access).

### `OQ-HST-10`

Which operational properties can afakt rely on when Tailscale's userspace client library, `tsnet`, uses an isolated Headscale behind afakt's coordination interface? What behavior applies to high availability, node churn, access-control propagation, memory per node, Internet Control Message Protocol (ICMP), `ListenTLS` certificates, and `WhoIs`? How do gateway attachments to user tailnets work? Can one node join two tailnets? What happens when coordination is unavailable and nodes use their static map of stable nodes? See [Cluster](network-topology.md#Cluster) and [Phone through a user tailnet to the dashboard](network-topology.md#Phone%20through%20a%20user%20tailnet%20to%20the%20dashboard).

## Agent launch

### `OQ-PRQ-1`

Which error fits a target path that breaks the path rules? Should callers receive `invalid-specification` or a dedicated error such as `workspace-source-invalid`? See [Launch validation](agent-profile.md#Launch%20validation).

### `OQ-EXE-2`

What happens when an agent host never comes back? Its agent launch then stays `unknown-stopped` or keeps returning `unavailable`, and a successor agent launch waits in `pending` behind it indefinitely. Options include an `abandoned` outcome that an operator declares and a deadline after which the agent launch ends by itself. See [Agent launches and supersession](agent-lifecycle.md#Agent%20launches%20and%20supersession).

### `OQ-EXE-3`

How should a request for a successor agent launch treat an active predecessor? Here the provider cancels the predecessor agent launch and holds the successor in `pending` until the previous one ends. Rejecting the request while the predecessor agent launch is still active would be the alternative. Agent spec changes take effect at a turn boundary. Work in flight remains [open](#%60OQ-AGT-27%60). See [Agent launches and supersession](agent-lifecycle.md#Agent%20launches%20and%20supersession).

### `OQ-EXE-5`

Which afakt command starts a worker that runs the agent dispatcher or the agent launcher? The options are a command of its own in the afakt command-line interface (CLI), a separate worker program, or a mode of the server command. See [Parts](agent-launch.md#Parts).

### `OQ-EXE-7`

Which identifier selects the deterministic test provider through the agent sandbox provider field in the [agent launch spec](agent-launch-contract.md#Agent%20launch%20spec), and how does placement find hosts that supply it?

### `OQ-EXE-8`

How are competing requests for successor agent launches admitted, ordered, and rejected as stale? Checking the predecessor link against the agent's active launch during admission is the leading option for preventing overlapping successors, at the cost of coordinated admission state. An agent spec generation only identifies a configuration revision and can't order retries that share it. The stale-request error and admission mechanism remain open. See [Agent launch identity](agent-launch-model.md#Agent%20launch%20identity).

### `OQ-EXE-9`

Who schedules retention cleanup durably after cancellation of the parent workflow or after the agent launcher workflow ends? An independent durable scheduler is the leading option, at the cost of persistent scheduling and reconciliation alongside the agent launch record. See [Evidence, retention, and cleanup](agent-launch-model.md#Evidence,%20retention,%20and%20cleanup).

### `OQ-EXE-10`

May a workflow continue-as-new or reset operation leave the same agent harness running, and how does afakt correlate workflow history with agent launch history? What happens to outstanding activities and their authority when workflow history changes without replacing the agent launch? See [Reconciliation after restart or retry](agent-launch.md#Reconciliation%20after%20restart%20or%20retry).

### `OQ-EXE-11`

Which authority may submit, read, cancel, or replay tasks on each workflow task queue, and how do workers verify that authority? What protects workflow history and task payloads from an unauthorized worker or network peer? See [Agent launch sequence](agent-launch-sequence.md#Agent%20launch%20sequence).

## Agent launch model

### `OQ-MOD-1`

Which part of afakt assigns the agent launch ID? afakt assigns it when it creates the agent launch, and the parent workflow records it before provision. The parent workflow could assign it, and so could the agent dispatcher workflow or the API server when a request arrives. See [Agent launch ID format](agent-launch-model.md#Agent%20launch%20ID%20format).

### `OQ-MOD-2`

Should a provider reject a universally unique identifier (UUID) in UUIDv7 form whose timestamp lies in the future or is implausibly old? Here the provider checks only the text form, the version digit, and the variant digit. The alternative also rejects a timestamp outside a window around the provider's clock, such as one later than the clock or older than the longest retention. See [Agent launch ID format](agent-launch-model.md#Agent%20launch%20ID%20format).

### `OQ-MOD-5`

Where does evidence live, and how does it leave the agent host? Beyond the published evidence references, and callers that never read the agent host's disk, that remains open. One option uploads each item to a central store at publication, and another fetches it from the agent host through the afakt API. See [Evidence, retention, and cleanup](agent-launch-model.md#Evidence,%20retention,%20and%20cleanup).

### `OQ-MOD-6`

Which digest algorithm applies? Only an example, `sha256:`, appears in the design. Every digest could use Secure Hash Algorithm (SHA) 256, or a fixed set of named algorithms could apply. See [Agent launch record](agent-launch-model.md#Agent%20launch%20record).

### `OQ-MOD-7`

What format do agent host identities and versions of the agent host worker take? A host identity is a SPIFFE ID in the control plane's trust domain, with a path from the scheme in [`OQ-MOD-9`](#%60OQ-MOD-9%60). That path could contain a name from the agent host's configuration or a generated ID. The worker version could be the afakt release version or the build's commit ID. Should an agent host keep one registry identity across re-enrollment when its authentication identity changes? See [Provenance record](agent-launch-model.md#Provenance%20record). What stable host label appears in an agent launch's internal name?

### `OQ-MOD-8`

What format do evidence reference IDs take? The specification gives none. Candidates are the item's digest, a UUIDv7, or an opaque text value from the provider. See [Evidence, retention, and cleanup](agent-launch-model.md#Evidence,%20retention,%20and%20cleanup).

### `OQ-MOD-9`

What path scheme do SPIFFE IDs use for agent hosts, agents, agent launches, and control plane components? Workload identities take the form `spiffe://<trust-domain>/...`, with one trust domain for each control plane and no federation. Does an agent's ID identify the durable agent or the agent launch, or include both? For a component in the control plane, should its identity name the service or one replica or both? See [Provenance record](agent-launch-model.md#Provenance%20record).

### `OQ-MOD-10`

What canonical representation defines the digest of an agent launch spec for numbers, Unicode, omitted fields, explicit empty values, and defaults? One versioned encoding is the leading option for consistent digests across implementations, at the cost of shared normalization and compatibility rules. See [Agent launch identity](agent-launch-model.md#Agent%20launch%20identity).

## Agent launch contract

### `OQ-CON-1`

Should cleanup become a separate provider operation? Cleanup runs as an activity on the host task queue, outside provision, restore, release, status, cancel, and result. See [Provider operations](agent-launch-contract.md#Provider%20operations).

### `OQ-CON-3`

What's the default retention period? It comes from a deployment default that has no fixed value. The choice lies between a value written in the contract and a value that each deployment sets for itself in its own configuration. See [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec).

### `OQ-CON-4`

If afakt acknowledged cancel before the agent launch ended, should a repeated cancel return the original acknowledgement or its terminal state? See [Cancel](agent-launch-contract.md#Cancel).

### `OQ-CON-6`

What format do label values take? Task class, variant, and role have none here, and each deployment defines their values. The profile name rule or a list in the deployment's configuration that the provider checks would each work. See [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec).

### `OQ-CON-7`

Where do the prompt and attachments live? The specification holds references with digests. The options are the evidence store, a content store behind the afakt API, or the prompt inline in the specification with attachments by reference. See [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec).

### `OQ-CON-8`

Which resource limits exist, and what are their names and units? The specification gives a token budget, central processing unit (CPU), memory, and disk as examples. The options are a fixed set in the contract or a set that each provider publishes. See [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec).

### `OQ-CON-9`

Does provision return when the agent sandbox is ready or after the agent harness is ready? Returning at sandbox readiness is the leading option for separating sandbox acquisition from harness startup, at the cost of a separate readiness observation. The return boundary and error ownership remain open. See [Provision](agent-launch-contract.md#Provision).

### `OQ-CON-10`

Should the abstract operation for restoring an agent snapshot be restore or resume? Restore is the leading contract verb for saved VM state, at the cost of differing from session resume terminology. The final verb remains open. Either choice creates a new agent launch linked to its predecessor. See [Restore](agent-launch-contract.md#Restore).

### `OQ-CON-11`

Which single meaning should the abstract verb start have? Reserving it for the agent harness reaching its running stage is the leading option for distinguishing it from agent sandbox readiness, at the cost of mapping provider-native start operations separately. The abstract meaning remains open. See [Provider operations](agent-launch-contract.md#Provider%20operations).

### `OQ-CON-12`

What scope and retention period apply to an agent launch request's idempotency key? Does the key only prevent duplicate submissions, or may a caller use it to follow later agent launches? Retaining the accepted request and rejecting changed content under the same key could prevent duplicate work, at the cost of retained request records. See [Agent launch request](agent.md#Agent%20launch%20request).

### `OQ-CON-13`

Which JSON Schema dialect and keywords may an agent launch's output schema use? What error rejects an invalid schema, and what wire representation does plain text output use? A declared dialect with validation before launch is the leading option, at the cost of rejecting schemas outside the supported subset. See [Agent launch spec](agent-launch-contract.md#Agent%20launch%20spec).

### `OQ-CON-14`

How should afakt expose an agent sandbox provider's native disk recovery operation that preserves files but restarts processes? Should recovered content become a workspace source for provision or use a separate operation, and which native operation metadata should clients receive? See [Restore](agent-launch-contract.md#Restore).

## Agent placement

### `OQ-PLC-1`

How long do a [registration lease](agent-placement.md#Agent%20host%20registry) and an unclaimed reservation last? Both lifetimes have bounds but lack fixed values. A short lease takes a lost agent host out of placement quickly, at the cost of more renewals.

### `OQ-PLC-2`

Which matching agent host does the [reservation API](agent-placement.md#Reservation%20API) pick? Picking the host with the most free slots spreads load, while filling one host first leaves whole hosts free for larger work. How should selection distribute work across provider accounts as well as agent hosts?

### `OQ-PLC-3`

Where does the [registry](agent-placement.md#Agent%20host%20registry) keep its records? Should the API server's database store them? Temporal's worker list could report which workers are polling and their slot usage, but it doesn't include afakt's host metadata.

### `OQ-PLC-4`

What measurement interval and downward step size should the [capacity controller](agent-placement.md#Slot%20capacity) use? Increases take one slot at a time. Decreases can remove more than one slot at once within the configured bounds.

### `OQ-PLC-5`

Should rate limits from the agent harness or model provider feed the [capacity controller](agent-placement.md#Slot%20capacity) alongside host resource use?

### `OQ-PLC-6`

Should [capacity](agent-placement.md#Slot%20capacity) account for each agent launch's resource requests? Agent profiles declare resource limits but don't declare resource requests.

### `OQ-PLC-7`

What residency, retention, account capacity, cost, and keepalive constraints make a remote agent host eligible for placement? How should placement distinguish compute lifetime from conversation and artifact retention, and should task queues belong to a provider or an account? See [Agent host registry](agent-placement.md#Agent%20host%20registry).

### `OQ-PLC-8`

How should placement decide whether an agent sandbox meets the isolation an agent profile requires? A provider name alone doesn't establish the same confinement. Placement could require an exact mode or use a reviewed compatibility table or check explicit security properties. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Agent host registry](agent-placement.md#Agent%20host%20registry).

### `OQ-PLC-9`

Who verifies the isolation and capability claims that an agent host advertises for placement? Trusting an unsupported claim can place work on a host unable to enforce its requirements. The registry could accept operator attestations or independently observed checks or signed claims with a bounded lifetime. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Agent host registry](agent-placement.md#Agent%20host%20registry).

### `OQ-PLC-10`

How should placement bind an authorized agent launch request to the reservation it creates when the caller's grants can change? A stale authorization can reserve or launch work beyond the caller's current bounds. The reservation could share an admission transaction with authorization or carry an authorization revision that the provider checks before claiming it. This choice depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Slot reservations](agent-placement.md#Slot%20reservations).

### `OQ-PLC-11`

When does the provider record a pending agent launch relative to claiming its slot reservation, including when an expired reservation leaves it waiting for capacity? See [Slot reservations](agent-placement.md#Slot%20reservations) and [Agent launch lifecycle](agent-sandbox-lifecycle.md#Agent%20launch%20lifecycle).

## Agent launch resource

### `OQ-RES-3`

How should an agent expose the history of its agent launches? Each AgentLaunch resource shows one agent launch. The durable agent resource could list its agent launches under `status`. See [Envelope](agent-launch-resource.md#Envelope).

## Agent workspace

### `OQ-WSP-1`

How do changes leave the agent workspace? The agent pushes through the `branches.push` capability, and the workspace changes also go into the evidence. The parent workflow could instead publish the changes from the evidence, in place of a push from the agent. See [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

### `OQ-WSP-2`

Can a profile restrict which repositories a specification lists as sources? Here a specification may list any repository URL. A profile field of allowed repositories or URL patterns would add that restriction. See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-3`

What format does a commit ID take? Workspace sources require full commit IDs. Secure Hash Algorithm (SHA) 1 alone is one option, and SHA-1 together with SHA-256 is another, in lowercase hexadecimal either way. See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-4`

How are repository URLs normalized? `${repository:<target path>}` drops the scheme, but which schemes a source may use and how the scheme-free form comes about are unresolved. The options are Hypertext Transfer Protocol Secure (HTTPS) alone, with the host and path in lowercase and without a trailing `.git`, or every scheme that Git accepts for a remote, mapped to one scheme-free form. See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-5`

Should new branch names have rules beyond those of the version control system? No further rules apply. The alternative is a prefix that each deployment sets, such as one for agent branches, or a pattern in the profile. See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-14`

How do guest paths for home, scratch, and named shared caches map to the permitted writable paths? Workspace-backed writable paths are the leading option, but must preserve scratch outside the workspace root and the existing cache-sharing contract without exposing host paths. The provider-installed hook protection decision remains [open](agent-open-questions.md#Agent%20hooks). See [Reserved names](agent-workspace.md#Reserved%20names).

### `OQ-WSP-15`

Which component creates and captures an agent workspace in host mode, and may the agent host worker run Git on a repository owned by the agent launch's user? What read and capture rights survive changes to output permissions? Capture material owned by the worker is a leading candidate, at the cost of a separate ownership boundary. Its access rules depend on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

### `OQ-WSP-16`

Must an agent workspace use the same canonical path across agent launches in host mode, or may afakt move path-dependent native history and project state? Stable paths are the leading option for preserving session lookup, at the cost of constraining workspace placement. When an agent moves between sandbox modes, must its native session state keep those paths, or may afakt translate them? See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-17`

What admission, quota, and cancellation rules apply to agent workspace revision capture and publication? At what durable boundary does cancellation concern an already committed revision? Manifest publication is the leading option for that boundary, at the cost of reconciling an uncertain publication before returning a final cancellation outcome. See [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

### `OQ-WSP-18`

Which Git metadata may agent workspace imports and image publication preserve, including worktree pointers, alternates, submodules, large-file objects, and executable configuration? A verified private repository with explicit handling of external references is the leading option, at the cost of limiting portability for repositories with external dependencies. Configuration handling depends on the on-hold [repository project configuration scope](agent-open-questions.md#%60OQ-HOOK-17%60). See [Workspace sources](agent-workspace.md#Workspace%20sources).

### `OQ-WSP-19`

What authorization and isolation rules apply to named tool caches shared across agent launches? How can one agent launch use a cache without reading another agent launch's private content or executing content that another launch changed, and what bounds cache consumption? See [Sandbox variables](agent-sandbox.md#Sandbox%20variables).

### `OQ-WSP-20`

For each agent sandbox provider, where do workspace files, scratch, shared named tool caches, and verified image caches live, and which identities can access their backing storage? How do mounts, ownership, quotas, cleanup, and reuse preserve separation across launches? What workspace disk and capture layout applies to the Windows-native utility virtual machine (VM), separately from the Linux microVM? See [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) and [Sandbox variables](agent-sandbox.md#Sandbox%20variables).

### `OQ-WSP-21`

How should archives and uploaded files become pinned workspace inputs, and how do their capture, authorization, metadata, and retention rules relate to prompt attachments and the existing workspace source kinds? During a launch, which workspace writes require continuous durable storage, and what failure boundary does that storage guarantee? See [Creation and pinned inputs](agent-workspace-lifecycle.md#Creation%20and%20pinned%20inputs) and [During a launch](agent-workspace-lifecycle.md#During%20a%20launch).

### `OQ-WSP-22`

What transaction binds a workspace disk, native harness state, memory, and device state into a recoverable checkpoint set? At each capture stage, how should crashes and cancellation behave? These stages include the idle wait and writer stop, filesystem freeze and member capture, object upload and manifest publication, then thaw and reopening admission. What proves commit visibility after a lost acknowledgement or host power loss, and what bounds capture waits, freeze duration, and recovery? See [Crash consistency and cancellation](agent-workspace-lifecycle.md#Crash%20consistency%20and%20cancellation).

### `OQ-WSP-23`

How should detachment and deletion follow the capture and hook order in [Hook events](agent-hooks.md#Hook%20events)? Which provider observations prove that no launch-owned process, mount, disk, directory, remote resource, or active credential authority remains? How should cleanup report missing retained content or partial removal, and which workspace and snapshot API operations and events expose these outcomes? See [Deletion and teardown](agent-workspace-lifecycle.md#Deletion%20and%20teardown) and [Operations and telemetry](agent-workspace-lifecycle.md#Operations%20and%20telemetry).

## Agent workspace image

### `OQ-IMG-1`

Does afakt need the OCI referrers API for image signatures, software bills of materials (SBOMs), or other attached artifacts? Which verification policy would consume them? For the Open Container Initiative (OCI) registry, see [Registry](agent-workspace-image.md#Registry).

### `OQ-IMG-2`

What registry topology fits a multi-host fleet: one registry or one per site? What metadata storage can registry replicas share while preserving tag updates and online collection? One central registry with authorized, verified agent host caches is the leading option. Site copies cut repeated transfers. Their cost is stale tags and duplicated storage, plus publication ownership and retention reconciliation to settle, while shared storage alone doesn't coordinate writers or collection, which leaves replica fencing and the metadata backend open. See [Registry](agent-workspace-image.md#Registry).

### `OQ-IMG-3`

How should peer-to-peer layer sharing between agent hosts work at the fleet stage, and how would a peer check permission for each repository pull without treating knowledge of a digest as access? Independent repository authorization on every cache read, followed by digest verification, is the leading access rule for any peer source. It prevents a warm cache from extending a revoked grant but requires an available authorization authority. This narrows the permission check. Peer discovery, transfer, and trust remain open. See [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup).

### `OQ-IMG-4`

How should workspace access density inform the policy for caching and preheating? Fetching complete images in advance preserves the no-lazy-pull rule but transfers content an agent launch may never read. Selecting which images to preheat reduces that transfer at the cost of cold starts. Selective preheating that yields to disk pressure is the leading option. Reusing verified complete content avoids repeated transfer and extraction, but sparse access can leave most bytes fetched in advance unused. Per-digest admission and space reservations bound concurrent preparation. Workload-based thresholds remain open. See [Placement and preheating](agent-workspace-image.md#Placement%20and%20preheating).

### `OQ-IMG-5`

Can afakt save a finished agent launch's agent workspace as a workspace image, and what provenance and publication permissions would that need? Capturing a verified stopped workspace under an explicit publication grant is the leading option. Pinning its source revision, profile, and provenance before publishing a complete image means retries use the same inputs, but requires separate publication and retention records. An agent snapshot captures VM state, while an agent checkpoint captures harness session state. Converting the agent workspace into an image doesn't define the transfer scope of either capture or of retained audit data. See [Workspace disk deltas and portable revisions](agent-workspace-image.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

### `OQ-IMG-6`

How do secrets that a build step needs stay out of every image layer? Brokered network access is the leading option where supported because it keeps credential values outside captured content, but it can't supply every local secret input. Every inherited and intermediate layer remains part of the disclosure boundary. Deleting a secret from the final tree doesn't remove it from earlier layers, and final-tree checks and known-secret scans can't prove secret absence. A separate delivery and capture contract is necessary for local secret inputs. See the [image build rules](agent-workspace-image.md#Image%20builds).

### `OQ-IMG-7`

Should the source rules allow a `repository` source immediately after an `image` source at the same target path? That would require an exception to the duplicate-target rule and a check that the image contains the same repository. After fetching and checking out the newer commit, rules would also need to cover dirty tracked files and stale dependencies plus outputs. Nested targets would still need their own decision. See [Image sources and digest pinning](agent-workspace-image.md#Image%20sources%20and%20digest%20pinning).

### `OQ-IMG-8`

Must the worker fetch the image layers before VM boot, or may the guest pull them through its agent mediator? Worker pulls are the leading option because they reuse verified disk bases and resolve preparation failures before session launch. Guest pulls add a mediated preparation phase and guest resource pressure. Whether they provide the same caching, authorization, and verification behavior remains unqualified. Both require complete verified content before hooks or the agent harness start. See [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup).

### `OQ-IMG-9`

What format and compatibility rules govern local disk deltas for agent workspaces, including exact parent identity, disk geometry, transitions to zeroed blocks, backing chain depth, and flattening? Authenticated immutable parents and full reconstruction verification are the leading option, at the cost of retaining parent disks and verifying complete reconstructed content. What limits and collection rules bound those chains? See [Workspace disk deltas and portable revisions](agent-workspace-image.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

### `OQ-IMG-10`

When overlapping workspace image builds finish out of order, should completion order or the newest accepted build determine the output tag? A publisher with exclusive publication authority and reconciliation by build identity is the leading option for avoiding duplicate publication after a lost acknowledgement. It requires durable coordination and leaves the winning build and notification atomicity undecided. See [workspace image builds](agent-workspace-image.md#Image%20builds).

### `OQ-IMG-11`

Which grants permit registry administration, and how do revocation and grant expiry affect transfers already in progress? Which upstream authentication services and redirect destinations may receive registry pull-through credentials? Explicit repository grants and independently bounded upstream destinations are the leading option, at the cost of separate authorization for each destination. Credential destinations depend on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Registry authentication](agent-workspace-image.md#Registry%20authentication).

### `OQ-IMG-12`

While uploads continue, what bounds apply to registry collection lag and abandoned uploads? During those uploads, what bounds apply to physical space reclamation? Which reconciliation deadlines and capacity guarantees apply when uncertain references must preserve their content? See [Retention and garbage collection](agent-workspace-image.md#Retention%20and%20garbage%20collection).

### `OQ-IMG-13`

Which principals may access the object store directly, and how do their grants restrict object reads, writes, and deletion? What protects stored content and credentials from disclosure or substitution? How does that protection cover backups and access outside the registry interface? See [Object store](agent-workspace-image.md#Object%20store).

### `OQ-IMG-14`

How does workspace image materialization preserve or reject filesystem metadata across Linux, macOS, and Windows, including name collisions, case handling, links, and permissions? Which compatibility checks keep native extraction and capture within the selected workspace while preserving verified content identity and executable dependency compatibility? See [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup).

### `OQ-IMG-15`

How should afakt produce guest images and workspace images through its own build pipeline, and which inputs, tests, signatures, publication permissions, and update rules would it require? What rules should a profile use to select a workspace image reference? Which guest Linux flavors and container builders should afakt support? For an optional container manager inside a VM sandbox, what changes would apply to the image and workspace mounts? What would change for network policy and cleanup? These choices are undecided. See the [workspace image build lifecycle](agent-workspace-lifecycle.md#Image%20builds).

### `OQ-IMG-16`

Which registry software should afakt use to store workspace images and guest images alongside agent snapshots and artifacts, with storage for build caches as well? How should its admission, publication, retention, collection, and multi-node behavior meet the shared registry interface? Selection is undecided. See [Registry admission](agent-workspace-lifecycle.md#Registry%20admission).

## Agent hooks

### `OQ-HOOK-10`

Is 10 seconds right for the session-end, stop, and teardown budgets? Each value is a guess. A profile or a deployment could set its own instead. See [Running hooks](agent-hooks.md#Running%20hooks).

### `OQ-HOOK-11`

Which decision applies when hooks at `model-request-after` return both `continue` and `stop` for one occurrence? The design leaves open whether `stop` always takes precedence or the first hook's decision applies. Another option gives the profile's hooks precedence over the specification's. See [Decisions](agent-hooks.md#Decisions).

### `OQ-HOOK-16`

What protects provider-installed hooks from changes by the agent, including hooks that share its agent sandbox and identity? Protection of hook configuration and dispatch remains open. See [Hook dispatch](agent-hooks.md#Hook%20dispatch).

### `OQ-HOOK-17`

Which repository project settings may the agent harness load beyond hooks, and how do they interact with provider settings? The scope of repository project configuration remains open. Which runtime trust-bundle settings must remain under the agent sandbox provider's exclusive control, and what precedence applies when repository configuration also sets them? See [Workspace repository hooks](agent-hooks.md#Workspace%20repository%20hooks).

### `OQ-HOOK-18`

How does the session-end wait interact with SVID lifetime? The wait, the session-end budget, and SVID renewal bounds need a shared rule. This operator decision remains on hold. See [Running hooks](agent-hooks.md#Running%20hooks).

## Agent observability

### `OQ-PRQ-5`

May a provider emit spans and metric attributes beyond the ones [Agent observability](agent-observability.md) lists? See [Conventions](agent-observability.md#Conventions).

### `OQ-OBS-4`

Where does the agent host's telemetry endpoint run? It could belong to the agent host worker or run as a separate component. See [Trace context propagation](agent-observability.md#Trace%20context%20propagation).

### `OQ-OBS-7`

Should a protected action wait or fail when afakt can't durably log its authorization decision, or may afakt proceed under an accepted loss policy? Durable log admission before the action is the leading option when afakt requires a complete decision history. It delays or rejects work during logging outages. How should afakt reserve capacity for control operations within bounded log storage? How should afakt reconcile credentials issued before an audit failure so retrying the request can't create authority absent from the audit records? See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-OBS-8`

Who may read the agent's decision log, and how long should afakt store it? How do retention holds and export cursors affect collection? A decision log policy independent of trace retention is the leading option, at the cost of separate access and retention controls. See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-OBS-9`

What bounds on issuance and rotation latency, propagation delay, memory use, and registration churn must the workload identity service meet? Which audit records must remain complete at those rates? See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-OBS-10`

What protects decision log and audit records from alteration, deletion, replay, or false attribution after a writer or storage service suffers compromise? Which authority may append, export, or remove those records, and how can a reader detect missing records? Where should audit checkpoints and their verification keys stay so compromise of the log writer can't conceal removal of the log's tail? See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-OBS-11`

What admission, size, rate, and storage limits apply to guest telemetry and evidence publication? How does afakt preserve control service when a guest floods permitted attributes, log records, or evidence? See [Trace context propagation](agent-observability.md#Trace%20context%20propagation).

## Agent profile

### `OQ-PRF-1`

Where does the profile catalog live, and who may publish to it? The catalog could live inside afakt behind the afakt API or in repository files. Another store would also work. See [Profile sources](agent-profile.md#Profile%20sources).

### `OQ-PRF-5`

What bounds a repository profile's self-grants and credential inject domains? Repository changes can grant `branches.push` or wide network access to the repository's own profile. A deployment-wide ceiling or a catalog profile could bound these grants and credential inject domains. An agent never edits its own agent spec and may submit an [agent change request](agent.md#Agent%20change%20request) instead. Repository profile self-grant and credential domain bounds remain on hold. See [Repository profile files](agent-profile.md#Repository%20profile%20files).

### `OQ-PRF-9`

Who defines the set of capabilities? Names such as `pull-requests.read` are illustrative, and the API server publishes the full set. Who writes that list and its names is unresolved. One option is the API server's own code, with one capability for each operation. A deployment configuration that groups operations into capabilities is another. See [Capabilities](agent-profile.md#Capabilities).

### `OQ-PRF-10`

What precedence applies among structural, identity, profile, source, and unsupported-setting errors when a forbidden profile override is also an unknown resource field? See [Launch validation](agent-profile.md#Launch%20validation).

### `OQ-PRF-11`

Should placement also filter on what the agent host can enforce? Placement offers an agent launch only to agent hosts that can resolve its credential references, while `harness-permission-unsupported` and `sandbox-unsupported` remain launch errors. The alternative also filters on agent harness permissions and sandbox settings. See [Launch validation](agent-profile.md#Launch%20validation).

### `OQ-PRF-14`

What are the canonical tool names? In the example, `read`, `search`, and `shell` each map to the agent harness's own name. A fixed list in the contract could define them, or each agent harness adapter could publish its names with its mapping. See [Profile fields](agent-profile.md#Profile%20fields).

### `OQ-PRF-15`

What are the effort levels? The example uses `medium` and `high`, and the provider maps each one to the agent harness's own form. A fixed scale in the contract could define them, or each agent harness's own values could pass through unchanged. See [Repository profile files](agent-profile.md#Repository%20profile%20files).

### `OQ-PRF-16`

What's an entry's capacity, and how does afakt measure it? Selection prefers the lowest `priority` number with capacity for the launch, but capacity has no definition. The model provider's rate limits and quota are one source. A limit on concurrent agent launches per entry is another. Can two entries share a priority? If they can, afakt needs a rule to order them. Which provider account observations may establish capacity, and when must missing or stale observations make the agent dispatcher wait rather than use an estimate? See [Profile fields](agent-profile.md#Profile%20fields).

### `OQ-PRF-17`

What role does `default: true` play? An agent launch request that names none of the four fields matches every entry. The default entry could decide that request, or priority and capacity could decide it as they decide any other. See [Profile fields](agent-profile.md#Profile%20fields).

### `OQ-PRF-18`

Which requirements for an agent sandbox provider should an agent profile state explicitly, and which details should the agent dispatcher resolve into the agent launch spec? The division between profiles and the dispatcher determines how portable profiles remain without hiding placement constraints. Profiles could name provider-specific settings or select named presets or declare capabilities for the dispatcher to resolve. See [Profile fields](agent-profile.md#Profile%20fields).

### `OQ-PRF-19`

May a new successor launch use the retired profile version pinned by its unchanged agent spec, or must the agent spec change first? See [Pinning a profile](agent-profile.md#Pinning%20a%20profile%20version) and [Profile changes](agent-profile.md#Profile%20changes).

## Agent overlay and names

### `OQ-HUB-1`

How many coordination service instances may a deployment run, and how do they divide ownership and recover it after failure? See [Overlay transport](agent-overlay.md#Overlay%20transport).

### `OQ-HUB-2`

What defines a namespace for agent, port, service, and preview names, and how does authorization govern name allocation and changes? See [Names](agent-overlay.md#Names).

### `OQ-HUB-3`

What does name resolution return for an agent or published port with no active launch, and how do clients distinguish absence from an unavailable launch? See [Names](agent-overlay.md#Names).

### `OQ-HUB-7`

Which resources and operations publish a port and grant caller access, and how does another launch map that port, and what are their ownership and lifetime rules? See [Published ports and services](agent-services.md#Published%20ports%20and%20services).

### `OQ-HUB-8`

How does a host-mode launch map a published port without colliding with other host listeners or granting another launch access to the mapping? See [Published ports and services](agent-services.md#Published%20ports%20and%20services).

### `OQ-HUB-10`

Where does Transport Layer Security (TLS) end for a service that maps a guest port to an encrypted service port, and which component manages its certificate and validates its name? See [Published ports and services](agent-services.md#Published%20ports%20and%20services).

### `OQ-HUB-11`

Can a service select more than one agent or move to a different agent? During transfer to a successor launch, what preserves the port mapping and handles overlapping ownership or in-flight connections? See [Published ports and services](agent-services.md#Published%20ports%20and%20services).

### `OQ-HUB-12`

Who may reserve a vanity preview name? How does allocation prevent collisions with default preview names, and when may another owner reuse a released name? See [Previews](agent-previews.md#Preview%20routing%20and%20names).

### `OQ-HUB-13`

Does viewer authorization belong to the published port or the vanity preview name? May a vanity preview move to another launch, and what happens to its authorization? See [Previews](agent-previews.md#Preview%20routing%20and%20names).

### `OQ-HUB-15`

How do container manager hosts and host modes supply one agent mediator and one agent proxy per workload while restricting the mediator to connections through its proxy? See [Launch boundary](agent-workload-proxy.md#Launch%20boundary).

### `OQ-HUB-16`

Is the agent proxy a separate program or an adapted proxy, and which component manages its lifecycle and configuration? See [Per-launch services](agent-workload-proxy.md#Launch%20boundary).

### `OQ-HUB-17`

May user tailnet identity contribute to afakt authorization, and how's that signal bound to the authenticated afakt user without replacing sign-in? See [User tailnet attachments](agent-overlay.md#User%20tailnet%20attachments).

### `OQ-HUB-18`

Is an agent service a preview whose configuration persists between agent launches, so one model and one lifecycle serve both with two Domain Name System (DNS) names and endpoints, or are services and previews separate models? See [Agent previews](agent-previews.md) and [Agent services](agent-services.md).

## Agent session

### `OQ-PRQ-6`

Which limits apply to ACP notification size and rate? Bounded defaults set by deployment policy are the leading option, but must permit legitimate large updates without allowing a guest to exhaust session services. See [Limits](agent-acp.md#ACP%20message%20policy).

### `OQ-SES-2`

Should the hooks get `suspend` and `resume` events of their own, beyond `before-agent` and `before-teardown`? See [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume).

### `OQ-SES-3`

How long do the idle timeout and the lifetime limit last by default? See [Limits](agent-lifecycle.md#Session%20limits).

### `OQ-SES-5`

Which carrier joins the guest relay to the agent proxy for session traffic under mediator message checks? Protecting it requires enforced process access and authenticated control endpoints for helpers and restored channels as well. Device ownership alone is insufficient, and those boundaries still need qualification for each backend. See [ACP message policy](agent-acp.md#ACP%20message%20policy).

### `OQ-SES-6`

How does guest control through the mediator preserve the sender claims used by [message signing](agent-session-attachment.md#Message%20signatures)? Deriving the envelope from context authenticated by the agent mediator is the leading option, but the API server must bind the sending user, agent, session, and agent launch without trusting guest-supplied claims. An SVID that the agent proxy obtains for the agent, acting as a SPIFFE broker, is the authenticated agent and agent launch claim. The agent proxy keeps the SVID outside the guest. User and session claims still need their own rule. A channel-bound sender context plus durable message IDs is the leading binding rule. It rejects identity substitutions and replay without guest signing keys, but needs shared replay state and a separate rule for verifying retained signatures after signing authority changes. See [ACP message policy](agent-acp.md#ACP%20message%20policy).

API-side derivation is the leading option, but needs an authenticated claim binding.

Deriving the signed envelope at the API server is the leading option, but the binding of user, agent, session, and agent launch claims needs a shared rule. Binding the relay channel to the agent and agent launch before accepting messages narrows the sender rule. That binding also includes the session. Signing at the API server can then use authenticated user context and durable message IDs, at the cost of shared signing and replay state across API replicas. See [Control requests](agent-workload-mediator.md#Control%20requests).

### `OQ-SES-7`

Should the [external ingress gateway](agent-overlay.md#External%20ingress%20gateway) share the API server process or run as an independently supervised process with separate restart and configuration reload behavior?

### `OQ-SES-8`

Which native events, permissions, and cancellation semantics of an agent sandbox provider may translate into an ACP session, and how do clients distinguish that subset from native ACP support? Explicit translation that rejects unknown events is the leading option. It adds a state machine and can't supply hooks, approval control, or visibility into actions the provider doesn't expose. See [Agent launches over ACP](agent-acp.md#Agent%20launches%20over%20ACP).

### `OQ-SES-9`

How should the agent mediator map ACP permission requests that contain commands or more than one target to policy resources? Explicit normalization that rejects unsupported forms is the leading option, at the cost of rejecting useful requests until definitions exist for their resource and effect mappings. See [ACP message policy](agent-acp.md#ACP%20message%20policy).

### `OQ-SES-10`

How does user authentication bind a client connection to its session and resource grants across ingress gateway and API server replicas? What protects forwarded user context and rejects cross-site requests, and what revokes active access and bounds authentication and attachment load? See [The protocol endpoint](agent-acp.md#The%20protocol%20endpoint).

### `OQ-SES-11`

How do the API server's per-agent ACP and remote-control routes relate to the shared `/acp` endpoint? Where does client transport end, and how does it map onto the harness transport? Endpoint ownership remains [`OQ-SES-17`](#%60OQ-SES-17%60). See [Remote control and sessions](agent-acp.md#Remote%20control%20and%20sessions).

### `OQ-SES-12`

What names should identify the component inside the guest's `afakt-agent` command that runs and talks to the harness program and its two modes? The component supports interaction through the Agent Client Protocol (ACP) and running the harness headless to print its output. Candidates are “harness adapter” for the component and “interactive session” and “headless run” for the modes. The names remain undecided. See [Running the harness](agent-acp.md#Running%20the%20harness).

### `OQ-SES-13`

How does afakt authentication middleware integrate with Traefik at the [external ingress gateway](agent-overlay.md#External%20ingress%20gateway)?

### `OQ-SES-14`

Which component controls the durable prompt queue, and how do the agent session workflow's control state and session delivery state agree across reconnects and restarts? See [Session workflow](agent-session.md#Session%20workflow) and [Turns](agent-acp.md#Turns).

### `OQ-SES-15`

In what order are harness settings installed and the guest relay started, and which component starts the ACP adapter? Both must be ready before applicable hooks and harness activity. See [Agent launch lifecycle](agent-sandbox-lifecycle.md#Agent%20launch%20lifecycle) and [Session connection](agent-acp.md#Session%20connection).

### `OQ-SES-16`

Which headless transport defaults and integration rules apply to workflow-controlled and interactive launches, including output delivery, approvals, cancellation, attachment, hooks, results, and evidence? See [Running the harness without interaction](agent-acp.md#Running%20the%20harness%20without%20interaction).

### `OQ-SES-17`

Which component accepts client Agent Client Protocol (ACP) and remote-control connections and reaches the agent's workload? How does it bind authenticated callers to the target launch? See [Remote control and sessions](agent-acp.md#Remote%20control%20and%20sessions).

### `OQ-SES-18`

Which component retains the session message log and enforces writer and reader roles, and how does replay work across reconnects and service failures? See [Writer and readers](agent-acp.md#Writer%20and%20readers) and [Session message log](agent-acp.md#Session%20message%20log).

## Agent session attachment

### `OQ-ATT-2`

How do ingress gateway and API server replicas route an attached client to the session when its workload changes or reconnects? Routing depends on [endpoint ownership](#%60OQ-SES-17%60) and [session state ownership](#%60OQ-SES-18%60). See [The protocol endpoint](agent-acp.md#The%20protocol%20endpoint).

### `OQ-ATT-3`

What's the default limit on the hop count for [agent messages](agent-session-attachment.md#Agent%20messages)?

### `OQ-ATT-5`

How does a client of an agent session recover when its replay cursor predates retained history or its accepted update kinds change? Bounded pagination or a verified copy of session history could supply recovery, while an explicit gap could identify missing history. What cursor lifetime, filtering watermarks, and recovery behavior should session replay provide? See [History and replay](agent-session-attachment.md#History%20and%20replay). What does a connected client observe while the launch connection is unavailable?

### `OQ-ATT-6`

How should an agent session client present an uncertain capture or restore result, and which retry is safe before reconciliation? Explicit uncertainty with no automatic repeat is the leading option for avoiding duplicate work, at the cost of delaying recovery until the client can reconcile the operation. Which progress and recovery controls should the client offer? See [Suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume).

### `OQ-ATT-7`

What transcript content may a client for an agent session select and copy as one operation? The selection boundary determines whether users can preserve context across messages. Clients could copy individual messages or arbitrary spans of the transcript or a structured export. See [History and replay](agent-session-attachment.md#History%20and%20replay).

### `OQ-ATT-8`

What latency and memory bounds should a client for an agent session meet when displaying retained transcripts? Loading all retained history can make a long conversation unusable on a constrained device. Clients could keep a bounded visible history or page older messages on demand or load the full transcript within device-specific limits. See [History and replay](agent-session-attachment.md#History%20and%20replay).

### `OQ-ATT-9`

How should an agent session client handle its live connection while the app is in the background? Losing local stream state can leave the displayed transcript incomplete when the user returns. The client could disconnect and replay from a durable cursor or refresh from retained history or maintain an authorized background connection. See [History and replay](agent-session-attachment.md#History%20and%20replay).

### `OQ-ATT-10`

Which permission requests may the writer of an agent session answer on a device that can't display the full action context? Limited presentation can hide the target or consequences of an approval. A client could require full context for every answer or permit a defined subset of actions or offer only viewing and denial on constrained devices. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-11`

How should afakt notify the owner of an agent session that the agent needs a permission answer? Notification delivery determines whether the user notices blocked work and which details appear outside the app. Clients could offer alerts with user permission or quiet notifications or app-only notices. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-12`

Which window sizes and display arrangements must native clients for agent sessions support? Required layouts determine whether a user can read action context and operate the session in a constrained window. Clients could require one full window or support windows whose size users can change or support more than one window across displays. See [Parts](agent-acp.md#Parts).

### `OQ-ATT-13`

Which agent session selection and unsent prompt state belongs to each client window? Shared drafts can overwrite input, while independent drafts can present conflicting writer state. Windows could keep independent drafts or share one draft or coordinate one active editor under the session's writer rules. See [Writer and readers](agent-acp.md#Writer%20and%20readers).

### `OQ-ATT-14`

Which keyboard and pointer operations must a client for an agent session support? Input support determines whether a user can move through the transcript and answer requests without touchscreen input. Clients could provide text entry and send alone or full keyboard navigation and shortcuts or those operations with pointer selection and drag and drop. See [Parts](agent-acp.md#Parts).

### `OQ-ATT-15`

Must a wearable client for agent sessions connect independently, or may it require a companion device? Requiring a companion can make permission answers unavailable when the devices are apart. The client could connect directly or relay through the companion or support direct access only when that connection is available. See [Parts](agent-acp.md#Parts).

### `OQ-ATT-16`

What should an agent session client do with a permission answer entered while it's offline? Delayed delivery can arrive after the request ends or the writer changes. The client could fail visibly or queue the answer for revalidation before expiry or hand it to an online companion without transferring writer authority. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-17`

How should clients for agent sessions present a permission request after another client resolves it? A stale pending display can invite an answer to a request that has already ended. Clients could remove the request or show its outcome briefly or keep it as read-only history. See [Tool approvals](agent-mcp-gateway.md#Tool%20approvals).

### `OQ-ATT-18`

What expiry and delivery bounds should apply to interactive permission requests and their outcomes in an agent session? An old request can present obsolete action context or keep work waiting indefinitely. afakt could use shared deployment bounds or limits specific to the action, with expiry checked when afakt accepts an answer. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-19`

What may a compact notification for an agent session show when it can't refresh its pending-request state? A stale count can imply that a resolved action awaits attention. It could show a launch shortcut or a pending count with its age or no session state until refreshed. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-20`

How should an agent session client present an answer rejected because its permission request expired? The user needs to distinguish an expired request from an accepted decision without reviving the action. The client could show an explicit expiry error or replace the request with its final outcome or keep a read-only entry. See [Tool approvals](agent-mcp-gateway.md#Tool%20approvals).

### `OQ-ATT-21`

How should an agent session client reconcile a permission answer with a lost response? A missing response doesn't establish whether afakt accepted the decision. The client could query the canonical outcome or offer an explicit retry tied to the same request or retry automatically within a bound that prevents duplicate dispatch. See [Tool approvals](agent-mcp-gateway.md#Tool%20approvals).

### `OQ-ATT-22`

In an agent session, which pending permission requests and accepted answers must persist through a service or client restart? Losing their state can leave a client uncertain whether an action may run. afakt could keep server-side decisions through service recovery or also persist pending client answers until expiry or invalidate undelivered answers and require fresh confirmation. See [Requests from the agent](agent-acp.md#Requests%20from%20the%20agent).

### `OQ-ATT-23`

Should a native client for agent sessions have a device identity as well as its authenticated user, and how should it prove that identity? Device credentials can bind access to an installation but can't replace the user's session grants. Clients could use a private key for mutual authentication or a bearer credential restricted to its audience or an authenticated companion acting for the device. See [The protocol endpoint](agent-acp.md#The%20protocol%20endpoint).

## Agent sandbox

### `OQ-PRQ-7`

For each permitted mode of an agent sandbox provider, which operating system editions, architectures, and releases must it support? How should that support account for native dependencies and agent snapshot compatibility? Which signing and distribution requirements apply to privileged sandbox components on those platforms? See [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms).

### `OQ-SBX-2`

Must an agent launch's credential references stay within those the profile permits? Credential references appear on both the profile and the agent launch spec. Restricting agent launch references to those the profile permits is the leading option for limiting credential selection, at the cost of requiring a profile change to select another credential. This bound remains on hold as part of repository profile self-grant and credential domain bounds. The central store can enforce the bound when it supplies a credential under that credential type's custody rules. How a changed bound affects credentials already supplied remains open. See [Credential brokering](agent-sandbox.md#Credential%20brokering).

### `OQ-SBX-6`

What format do credential references and cache names take? Neither the `service` value nor the `<name>` in `${cache:<name>}` has one here. Each could follow the profile name rule or take free text from the agent host's configuration. See [Sandbox variables](agent-sandbox.md#Sandbox%20variables).

### `OQ-SBX-7`

Which fields does an `oauth` entry contain, and how does it map a user to a stored login credential? afakt's credential store is the only refresh owner, independent of the operator's own CLI sign-in, and doesn't share a refresh lock with that CLI. A reference to a stored login credential is the leading option for keeping refresh state central, at the cost of a separate authorization lookup. Whether inject rules and sign-in scopes belong in the entry or in the stored credential, and how authenticated users map to credentials, remain open. See [Credential brokering](agent-sandbox.md#Credential%20brokering).

### `OQ-SBX-9`

Should macOS use Apple's Containerization framework or drive Virtualization.framework directly? Direct Virtualization.framework is the leading owner choice because afakt controls guest initialization and lifecycle, at the cost of maintaining its own guest initialization. Containerization supplies a maintained init but still requires afakt to coordinate agent snapshots with agent checkpoints. Runtime save failures prevent agent snapshots with memory state despite successful support validation. Preserving memory state requires successful runtime save and restore. Cold recovery needs a separately qualified restoration contract. Should afakt reuse a guest initialization and process protocol across virtual machine backends while retaining lifecycle, cancellation, and channel ownership? See [Enforcement](agent-sandbox.md#Enforcement) and the [agent snapshot and agent checkpoint questions](agent-open-questions.md#Agents).

### `OQ-SBX-10`

Does privileged setup count against a rootless runtime? Separating privileged setup from rootless runtime is the leading definition, with setup rights specified for each supported platform. It keeps runtime authority narrow but still requires administrative setup where the platform demands it. See [Enforcement](agent-sandbox.md#Enforcement).

### `OQ-SBX-11`

How should the user service keep running after logout and start at boot? User lingering is a candidate for keeping service ownership with the user, but continued operation after logout and boot activation remain unqualified. A system supervisor adds administrative ownership. The persistence policy remains open. See [Enforcement](agent-sandbox.md#Enforcement).

### `OQ-SBX-12`

Should the agent mediator run as a confined process or in its own VM? A confined process is the leading option for density and throughput. Running the mediator in a VM adds kernel separation at the cost of more memory and lower throughput. afakt permits all [agent sandbox provider modes](agent-host.md#Runtime%20privileges%20and%20platforms), including the weaker `macos-host`, `windows-host`, and `linux-host` modes. Which traffic parsers and privileged callbacks must share the mediator's confinement boundary under the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60)? See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-13`

How does each platform bind local proxy and mediator decision channels to the correct workload and launch, rejecting stale identities after reuse? The guest-to-proxy link is private to the workload and doesn't require encryption. Its [conditions each platform must meet](#%60OQ-SBX-70%60) remain draft. Host-mode isolation depends on the held [`OQ-SBX-18`](#%60OQ-SBX-18%60). See [Agent workload guest](agent-workload-guest.md).

### `OQ-SBX-14`

How should the agent mediator handle overload without starving control? Bounded admission is a candidate for preserving control responsiveness, at the cost of rejecting useful streams. Complete streaming service, durable decision logging, and hard resource bounds remain unresolved. Admission limits and separate control capacity don't establish those guarantees. See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-16`

How do the permitted `macos-host`, `windows-host`, and `linux-host` sandboxes keep agent launches' loopback listeners apart? A private network namespace with an enforceable loopback bind rule is the leading option on Linux, but namespace separation alone doesn't forbid wildcard listeners. Seatbelt on macOS has no private network namespace, and assigned port ranges don't create one. Rejecting a launch that requires an unavailable private network view is the leading response, at the cost of excluding that configuration. See [Enforcement](agent-sandbox.md#Enforcement).

### `OQ-SBX-17`

Which credential exchange calls use X.509-SVIDs and which use JavaScript Object Notation (JSON) Web Token (JWT) SVIDs? Connections from an agent host to the control plane, including agent proxy connections to the API server and permitted services, use mutual TLS with X.509-SVIDs. Calls that exchange an SVID for other credentials use the X.509-SVID over mutual TLS or a JWT-SVID whose audience is the exchange service. The guest never receives an SVID. See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-18`

Should each agent launch run under its own User Identifier (UID), separate from the dedicated non-root worker user? A small privileged launch helper that only starts a confined command as that UID is the leading option. It separates agent launches from the worker user at the cost of a privileged component on every host and a separate mechanism on each operating system. The worker never runs as root. The helper remains open alongside [rootless setup privileges](#%60OQ-SBX-10%60). See [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms).

### `OQ-SBX-19`

How does each launch's own agent mediator fit the process and user boundaries of its provider? Each launch has one mediator and one agent proxy, with separate network state. One process per agent launch is the leading option if [per-launch user identifiers](#%60OQ-SBX-18%60) work, because each mediator can run under its launch's UID. It costs more memory and startup time per agent launch. See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-20`

How does agent egress gateway placement relate to [proxy TLS termination](#%60OQ-SBX-72%60)? The mediator decides connections without transmitting traffic. What disruption and recovery bounds must the agent egress gateway meet when its process restarts or drains? See [Agent egress gateway](components.md#Agent%20egress%20gateway).

### `OQ-SBX-21`

Where does the agent mediator get upstream Domain Name System (DNS) answers? Options include DNS over HTTPS or DNS over TLS to deployment-configured upstreams. Full recursion and forward zones for split-horizon names are other options. See [Agent DNS](agent-workload-mediator.md#Agent%20DNS).

### `OQ-SBX-22`

Should the guest receive synthetic addresses for every name while real addresses stay inside the agent mediator? Synthetic addresses are the leading option. See [Agent DNS](agent-workload-mediator.md#Agent%20DNS).

### `OQ-SBX-23`

Where does the host-side credential broker sit relative to the agent egress gateway and agent mediator? It remains outside the guest. Its placement may depend on the proxy software chosen for the agent egress gateway. See [Credential brokering](agent-sandbox.md#Credential%20brokering).

### `OQ-SBX-24`

How should egress classification by the agent mediator handle encrypted or absent server names and literal network addresses? How should it handle pinned certificates, non-HTTP TLS, or encrypted datagram transports? Denying an unclassifiable destination is the leading option, at the cost of excluding clients whose destinations the mediator can't verify. Which visibility rules permit passthrough without bypassing required inspection? Which datagram and address-family capabilities must an agent sandbox preserve while enforcing these visibility rules? See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-SBX-25`

How should afakt authenticate and atomically install deployment policy revisions without changing an immutable agent launch spec? Rule changes end existing connections the new rules deny. How does policy expiry or failure to load affect admission, and how does [persistent egress and User Datagram Protocol (UDP) revocation](#%60OQ-SBX-69%60) work? See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-SBX-26`

What response content may reach an agent after a credentialed request? May it include a secret that an authorized upstream reflects or transforms? Trusted destinations or constrained response handling could bound disclosure, at the cost of limiting useful streaming responses. The permitted response boundary depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). See [Credential brokering](agent-sandbox.md#Credential%20brokering).

### `OQ-SBX-27`

Must the privileged launch helper authenticate one instance of the agent host worker, or may any process under the worker's user identity invoke its allowed client? Binding admission to a verified instance could narrow caller authority at the cost of restart and renewal coordination. The caller boundary depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms).

### `OQ-SBX-28`

Which public files, Unix sockets, and Mach services may an agent launch in host mode read or reach beyond its private workspace and home? An explicit read and service policy could bound that access at the cost of restricting tools that depend on shared host services. Permitted exposure depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox).

### `OQ-SBX-29`

Who reconciles surviving agent launch processes after an agent host worker or privileged launch helper fails, and what proves that a launch user identifier is safe to reuse? Holding the identity out of the pool until afakt verifies process and access retirement is the leading option, at the cost of unavailable capacity during uncertainty. Identity reuse depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Restart and recovery](agent-host.md#Restart%20and%20recovery).

### `OQ-SBX-30`

Before workload identity issuance, which authority verifies a guest endpoint's mapping to its agent and agent launch as well as its agent host? May a trusted agent proxy supply that mapping, or must an independent attestor verify the agent launch record? What authorizes renewal after consumption of bootstrap authority? Which authority may change the runtime metadata used to establish that mapping under the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60)? See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-31`

What limits must apply to a host-side workload identity broker's reference types, target agent launches, audiences, and returned private keys? May compromise of one broker expose every active target identity on its agent host? The broker's authority depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-32`

What independent trust bootstraps the workload identity issuer and its data store, and who can recover them after all cached identities expire? How long may consumers operate during an issuer outage? Separate bootstrap trust or local data store access could break a startup dependency on the issuer, at the cost of a separate recovery authority. What separate expiry and recovery bounds apply to workload credentials and agent host credentials as well as cached trust material alongside the on-hold [session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60)? See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-33`

Should afakt integrate a separately operated workload identity issuer or maintain its own issuer? Which component owns issuance, attestation, registration, and trust distribution under either arrangement? See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-34`

Should an upstream credential exchange map each agent launch to its own provider principal or share a principal namespace? How should provider-side authority stay within the agent launch's allowed scope? Principal sharing depends on the on-hold [repository profile self-grant and credential domain bounds](agent-open-questions.md#%60OQ-PRF-5%60). Which principal and scope restrictions must the credential issuer or downstream service enforce independently of admission by the API server? See [Credential types and custody](agent-sandbox.md#Credential%20types%20and%20custody).

### `OQ-SBX-35`

Which name queries may leave the agent mediator for upstream name services, and what prevents a guest from sending data through query names to an attacker-controlled name service? How do query authorization, response validation, and query limits relate to connection policy? See [Agent DNS](agent-workload-mediator.md#Agent%20DNS).

### `OQ-SBX-36`

How does a remote agent sandbox integrate its per-launch mediator and agent proxy with the global egress gateway while preventing direct connections that bypass policy? How do its control requests, sessions, previews, and MCP calls use the launch-bound local channel and authenticated direct overlay connections? See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-37`

What contains the authority of a compromised agent host or site gateway, and which independent observations can detect false policy or lifecycle claims? How does afakt revoke that component and protect other agent launches while reconciling its active connections and credentials? Which container management operations may the worker on the agent host invoke under the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60), and what prevents it from controlling unrelated sandboxes? See [Trust boundary](agent-host.md#Trust%20boundary).

### `OQ-SBX-38`

How should the agent egress gateway get destination certificates and replace them when permitted destinations change? The choice affects connection startup and removal of obsolete signing authority. It could receive certificates for an explicit destination inventory or request them as authorized connections arrive. See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-SBX-39`

May policy evaluations reuse a process or sandbox, and what state must afakt clear between evaluations? Reuse can reduce startup cost but keep data or authority from an earlier request. afakt could create a fresh evaluator for each request or reuse evaluators after a defined reset or keep one evaluator per agent launch. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-SBX-40`

When workload credentials expire or an agent launch ends, how quickly must afakt close already authorized connections and revoke exchanged credentials? Refusing new requests doesn't retire authority already admitted by a downstream service. afakt could cancel active use immediately or permit a bounded drain or require downstream checks before each further operation. This choice depends on the on-hold [session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60). See [Agent launch authentication](agent-workload-proxy.md#Agent%20launch%20authentication).

### `OQ-SBX-41`

How should workload identity consumers remove a retired certificate authority and reject stale trust bundles? Retaining old trust can extend compromised authority, while early removal can disconnect valid workloads. Consumers could require a bounded bundle age or acknowledge coordinated removal or require independent enrollment after trust expires. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-42`

Which enrollment and workload private keys must resist export by the process that uses them? Software custody and hardware-backed signing expose different authority after process compromise. afakt could require protected hardware keys for enrollment or require them for workload keys too or permit software keys within a separately confined broker. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-43`

How should each control plane service grant access to the workload identities that connect to it? Successful mutual authentication establishes identity but doesn't grant access to every service. A service could use a configured caller allowlist or check current grants through the control plane before admitting operations. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-44`

When's an agent mediator ready to decide guest connections during concurrent agent launch startup? Early admission can expose traffic before identity and policy are available, while waiting for unrelated launches delays ready agents. Readiness could be independent for each mediator or depend on a shared service becoming ready for all launches it serves. See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-45`

How should a data store in the control plane map an authenticated workload to its database role? A certificate common name and its workload identity can identify different principals without enforcement of their relationship. afakt could constrain certificate common names at issuance or use an identity-aware intermediary or require native workload identity validation by the data store. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-46`

Where must afakt check the complete workload identity certificate profile before accepting an identity? A valid signature alone can leave ambiguous identity fields that different consumers interpret differently. afakt could require validation at each consumer or use a shared verifier at the service's authentication boundary. See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-47`

How should host-side components receive and rotate workload identity documents and private keys? Delivery through files, a local service, or an embedded client gives different processes access to identity material. afakt could use protected files with reload notification or direct identity-service clients or a local delivery adapter with per-component authorization. This choice depends on the on-hold [host isolation baseline](agent-open-questions.md#%60OQ-SBX-18%60). See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-48`

How should an exchanged credential's lifetime relate to the requesting workload identity document and the agent launch? An independently valid credential can outlast the authority used to get it. Issuance could cap expiry at the requesting identity's expiry or use a launch-bound deadline checked by consumers or use a separate lifetime with explicit revocation. This choice depends on the on-hold [session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60). See [Credential types and custody](agent-sandbox.md#Credential%20types%20and%20custody).

### `OQ-SBX-49`

What clock skew and certificate backdating may workload identity and credential exchange consumers accept? Different tolerances can extend authority beyond the intended expiry or reject valid requests. Consumers could enforce synchronized clocks with strict deadlines or accept an explicit bounded tolerance shared by issuers and verifiers. This choice depends on the on-hold [session-end wait and SVID lifetime](agent-open-questions.md#%60OQ-HOOK-18%60). See [Workload identity](agent-host.md#Workload%20identity).

### `OQ-SBX-50`

What replay resistance must an upstream credential exchange provide for bearer identity documents? Possession during validity can permit repeated exchanges even when the original request has completed. The exchange could rely on bounded expiry or keep one-use request records or require proof of possession where the upstream supports it. See [Credential types and custody](agent-sandbox.md#Credential%20types%20and%20custody).

### `OQ-SBX-51`

What issuance limits and retry semantics apply when one workload identity requests upstream credentials repeatedly? Per-credential lifetime limits don't bound the total authority or load produced by repeated issuance. afakt could limit request rates and counts or bind issuance to a specific operation or keep idempotency records that return the original outcome. See [Credential types and custody](agent-sandbox.md#Credential%20types%20and%20custody).

### `OQ-SBX-52`

Which provider assertions may permit enrollment on behalf of a remote agent launch? A provider account claim alone may not bind the requested identity to the intended launch. afakt could require an authenticated provisioning channel or a verified runtime assertion bound to the launch or a trusted external representative with proof of the key it uses. See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-53`

What confines host-side virtual device processing and virtual machine monitors after hostile guest input exploits them? Which privileges, descriptors, file mappings, and host interfaces remain reachable for macos-vm, linux-vm, both windows-vm guests, and the outer Docker Desktop virtual machine? After a desktop guest compromise, what prevents access to physical host integrations? See [Enforcement](agent-sandbox.md#Enforcement).

### `OQ-SBX-54`

How does each host mode and container provider force all harness and workspace-command traffic through its own proxy under mediator decisions? That traffic includes name queries and loopback publication streams. How do both windows-vm guests connect to the proxy across their distinct device and host boundaries? Which interfaces, inherited descriptors, daemon paths, and host services could bypass that routing, and how does each provider enforce the required network view? See [Agent mediator](agent-workload-mediator.md#Launch%20boundary) and [Loopback listeners](agent-ports.md#Loopback%20listeners).

### `OQ-SBX-55`

Which Windows-native mechanism does afakt-sandbox use for process, filesystem, handle, and network confinement in windows-host and the Windows utility VM guest? Which sandbox settings can it enforce, and which settings require rejection? How does it confine the complete descendant process tree? This question doesn't settle the on-hold host isolation baseline. See [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox).

### `OQ-SBX-56`

Which namespace, mount, capability, daemon socket, and host integration restrictions apply to rootless docker-engine and Docker Desktop launches? Which Windows desktop backend does afakt support, and how does each desktop backend separate containers, shared Linux state, and physical host file access? How are management peers authenticated, actions attributed, and shared backend resources bounded? See [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms).

### `OQ-SBX-57`

For each host mode and container provider, which independent authority stops every launch descendant and proves termination when the harness, command process, or runtime refuses cancellation? How do worker failure, helper failure, runtime restart, and surviving children affect slot release and identity reuse? See [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) and [Restart and recovery](agent-host.md#Restart%20and%20recovery).

### `OQ-SBX-58`

Which host operating systems and architectures does the `podman` provider support? How does Podman run without root privileges, and which setup privileges and services does it need? Is a Podman virtual machine part of the macOS or Windows deployment, and if so, where do the container kernel, runtime management, and physical host boundaries lie? How do the container and host integration restrictions in [`OQ-SBX-56`](#%60OQ-SBX-56%60) apply to Podman, and does its VM require the containment considered in [`OQ-SBX-53`](#%60OQ-SBX-53%60)? Setup and service survival remain under [`OQ-SBX-10`](#%60OQ-SBX-10%60), proxy routing under [`OQ-SBX-54`](#%60OQ-SBX-54%60), and storage placement under [`OQ-WSP-20`](#%60OQ-WSP-20%60). See [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms).

### `OQ-SBX-59`

For each host mode, what authenticates registration of a launch with its traffic capture component and binds its tunnel or flows to the current proxy? The choice needs to preserve original destinations and reject stale registrations after identity reuse. Installation, updates, and removal also need an authority whose privileged components remain confined after hostile input. How do component failure, restart, and removal preserve isolation for existing flows as well as new connections? See [Agent mediator](agent-workload-mediator.md#Launch%20boundary).

### `OQ-SBX-60`

Windows needs management adapters for startup, supervision, capture, and stopping both the Linux microVM inside WSL2 and the Windows-native utility VM. Which adapters and authenticated management calls let a non-administrator worker delegate only the required operations? A successful management response may precede process exit. What independent observations prove that all guest processes stopped after channel failure, and how do outer-environment restart and shared resource accounting affect that proof? See [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits).

### `OQ-SBX-61`

Which policy engine evaluates connection, message, and launch-to-launch grant rules at the agent mediator? How does the choice preserve the same authorization meaning at both ends of a granted connection? See [Agent launch](network-topology.md#Agent%20launch) and [Launch to launch under a grant](network-topology.md#Launch%20to%20launch%20under%20a%20grant).

### `OQ-SBX-62`

How does each operating system restrict a host-mode mediator's connections to its proxy while preserving the fixed control plane rules and its local decision channel? Which interfaces, inherited descriptors, resolver paths, and indirect host services must the confinement deny, and how does enforcement behave when a capture or confinement component fails? See [Agent launch](network-topology.md#Agent%20launch).

### `OQ-SBX-63`

How does a provider with a vendor-managed agent loop map its harness, session, tool execution, and model requests to an afakt agent launch? Which afakt-agent, afakt-sandbox, guest relay, Agent Client Protocol (ACP), and hook responsibilities can run in the vendor sandbox, and which require an external adapter or rejection of the launch? How are server-side tools and app-run tools bound to the same launch when they execute outside that sandbox? See [Process placement for Claude Managed Agents](agent-sandbox-claude-managed-agents.md#Process%20placement%20and%20startup) and [Process placement for OpenAI Agents API](agent-sandbox-openai-agents-api.md#Process%20placement%20and%20startup).

### `OQ-SBX-64`

Which vendor services, personnel, and account roles may read or change remote workspace content, session history, outputs, snapshots, credentials, and runtime observations? What bounds their access, retained copies, and deletion completion, and what evidence remains trustworthy after vendor compromise? How does afakt enforce its content exclusions where a vendor retains data independently? Region and retention eligibility remain under [`OQ-PLC-7`](#%60OQ-PLC-7%60). See [Remote vendor custody](agent-threat-model-claude-managed-agents.md#Vendor%20custody%20and%20administration) and [Host observations](agent-sandbox-cloudflare.md#Host%20observations).

### `OQ-SBX-65`

Which remote runtime, execution-user privileges, kernel boundary, and host privileges can each provider guarantee? How can afakt-sandbox enforce the pinned configuration there, and what contains hostile runtime or file-processing input after it crosses the sandbox boundary into vendor services? Which independent evidence qualifies those claims without trusting the compromised component? This question doesn't settle the on-hold host isolation baseline. See [Modal isolation](agent-sandbox-modal.md#Isolation%20boundary) and [Cloudflare isolation](agent-sandbox-cloudflare.md#Isolation%20boundary).

### `OQ-SBX-66`

For each remote provider, which observation proves that every process and detached operation started for an agent launch has stopped before afakt releases its slot or starts a successor? How do session cancellation, shared environment lifetime, asynchronous deletion, provider unavailability, and late requests that recreate execution affect that proof? Which authority prevents further work under retired vendor handles? See [OpenAI termination](agent-sandbox-openai-agents-api.md#Cancellation%20and%20teardown) and [Cloudflare termination](agent-sandbox-cloudflare.md#Cancellation%20and%20teardown).

### `OQ-SBX-67`

What resource shape represents a workload's firewall and its ingress and egress rules? Egress defaults to a cluster-managed deny-all policy, including in host modes. See [Firewall](agent-workload.md#Firewall).

### `OQ-SBX-68`

How does a rule specifying a domain apply to QUIC traffic through the agent proxy, including domains with HTTP-level rules or credential injection? See [Network policy](agent-workload-mediator.md#Network%20policy).

### `OQ-SBX-69`

How does a rule change revoke persistent egress connections that the guest starts and UDP flows? A rule change ends existing connections the new rules deny. See [Rule changes](agent-workload-mediator.md#Rule%20changes).

### `OQ-SBX-70`

Which conditions must the private guest-to-proxy link meet on every platform? The working draft covers original destinations that the guest can't falsify to bypass policy, TCP, UDP and QUIC decisions, flow revocation, confinement of routing components, metadata trust, and a common mediator contract. Its preferred mechanisms, license constraints, and control interfaces remain draft conditions. Host modes apply containment against the agent at its sandbox privileges and depend on the held [`OQ-SBX-18`](#%60OQ-SBX-18%60). See [Agent workload guest](agent-workload-guest.md).

### `OQ-SBX-71`

Which component manages the workload lifecycle, and what are its states? A workload can exist without a launch or workspace. It runs exactly one agent launch, and afakt destroys it afterwards. A follow-up or relaunch uses a new workload. See [Lifecycle](agent-workload.md#Lifecycle).

### `OQ-SBX-72`

Does TLS termination in the agent proxy take over the agent egress gateway's decryption role? How does that choice affect inspection and credential brokering at a standalone site gateway? See [Agent egress gateway](agent-workload-mediator.md#Agent%20egress%20gateway).

## Agent MCP gateway

### `OQ-MCP-1`

Which upstream servers use per-user OAuth identity or a shared identity, and how does a session's user map to that identity? An explicit mapping from authenticated users to upstream authorizations is the leading option, at the cost of authorization records and refresh coordination. Shared hosting also requires a qualified process boundary that prevents unauthorized workspace reads and direct egress, and either a trusted server or isolation between authorization identities. Response correlation alone can't prevent content mixing. The permitted sharing policy remains open. See [Upstream credentials](agent-mcp-gateway.md#Upstream%20credentials).

### `OQ-MCP-2`

Does the gateway pass through [MCP resources and prompts](agent-mcp-gateway.md#Server%20and%20tool%20access), or only tools?

### `OQ-MCP-3`

Do results from [untrusted servers](agent-mcp-gateway.md#Server%20and%20tool%20access) need content filtering against prompt injection before reaching the agent?

### `OQ-MCP-4`

What limits and rate limiting apply per agent launch and per upstream server? Independent admission limits per agent launch and per upstream server with bounded pending calls and response waits are the leading option for bounding load. They can reject useful work. Preventing monopolization also requires an explicit allocation or fairness rule. Limits, fairness, and coordination across gateway replicas remain open. See [Server hosting and lifecycle](agent-mcp-gateway.md#Server%20hosting%20and%20lifecycle).

### `OQ-MCP-5`

Should agent harness-native tools, such as file editing and shell commands, ever [route through the gateway](agent-mcp-gateway.md#Endpoint%20and%20harness%20integration) for audit?

## Agent ports and previews

### `OQ-PORT-1`

Should port discovery and exposure support User Datagram Protocol (UDP)? Keeping publication TCP-only preserves the stream contract but excludes UDP services. Supporting UDP requires a separately authorized datagram contract. The scope remains open. See [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records).

### `OQ-PORT-2`

Can share links admit public recipients without afakt sign-in at all? See [Access and share links](agent-previews.md#Access%20and%20share%20links).

### `OQ-PORT-3`

What per-profile limits apply to how many ports and previews an agent launch may expose? See [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records).

### `OQ-PORT-4`

How does afakt certify preview host names beneath a wildcard domain behind a public tunnel? Wildcard coverage is the leading candidate where the provider supports the required host names and certificates, while certificates for each host add issuance and renewal work. Funnel's assigned names require a different arrangement with enrollment and quota costs, potentially using separate node identities. Public issuance, renewal, and the embedded Funnel path remain unqualified. See [Public tunnel providers](agent-previews.md#Public%20tunnel%20providers).

### `OQ-PORT-5`

Do tunnel providers beyond the two built-ins become plug-ins? See [Public tunnel providers](agent-previews.md#Public%20tunnel%20providers).

### `OQ-PORT-6`

Should a preview request wake a suspended agent or fail? A failed request avoids implicit resource use, while waking offers convenience but requires authorization, a bounded restore wait, and verified binding to the intended listener. Rechecking the listener alone doesn't establish that binding. The wake policy and exact listener-binding mechanism remain open. See [Sessions and clients](agent-previews.md#Sessions%20and%20clients).

### `OQ-PORT-7`

Should bulk transfers use a connection separate from interactive preview streams? Separate connections are a candidate for reducing bulk-transfer interference, at the cost of extra ownership and scheduling. Reserved control capacity and bounded data admission reduce queue contention but don't establish a control-latency guarantee or remove shared TCP loss stalls and processor contention. Connection policy and behavior under saturation remain open. See [HTTP previews](agent-previews.md#HTTP%20previews).

### `OQ-PORT-9`

What browser origin and site boundary isolates an untrusted preview from sibling preview cookies while preserving app compatibility? Separate site boundaries or restrictions on supported applications could prevent scripts from writing shared parent-domain cookies, at the cost of extra domains or reduced compatibility. See [HTTP previews](agent-previews.md#HTTP%20previews).

### `OQ-PORT-10`

What should happen to an active preview connection when its route or publication changes? Preserving the stream can preserve access to a listener after the publication no longer permits new connections. afakt could close affected streams immediately or drain them within a bound or permit them until their existing authorization expires. See [HTTP previews](agent-previews.md#HTTP%20previews).

### `OQ-PORT-11`

How are public tunnel route changes authenticated and audited, and how are tunnel credentials protected from a compromised connector embedded in the API server process? What confines hostile connector input and bounds enrollment, reconnect, and forwarding resource use? During connector failure, how do the API listener and unrelated publications behave? See [Public tunnel providers](agent-previews.md#Public%20tunnel%20providers).

### `OQ-PORT-13`

Which preview request checks and transformations belong to the external ingress gateway and which belong to the preview server? Where are viewer authorization, afakt credential removal, and host-header rewriting enforced for both public and host-node access? How does the preview service bind its published-port request to the permitted caller and target launch? See [Operator browser to a preview](network-topology.md#Operator%20browser%20to%20a%20preview).

### `OQ-PORT-14`

How do the preview server and other callers reach a published port through the workload's proxy? Connections require the direct overlay path with mutual authentication and mediator checks. The binding from a publication to a guest listener and the caller context remain open. See [HTTP previews](agent-previews.md#HTTP%20previews) and [TCP client tunnels](agent-ports.md#TCP%20client%20tunnels).

## Control plane server supervision, data upgrades, and rollback

### `OQ-SRV-1`

What restart, readiness withdrawal, and drain policy applies to control plane server child processes, and which data upgrades permit rollback? Bounded retries, explicit dependency health, and one global drain deadline are the leading option, at the cost of stopping recovery attempts when those bounds expire. Which component owns migration compatibility and recovery? Must control plane upgrades preserve service through overlapping compatible processes, or may they require a bounded maintenance outage? See [Registry](agent-workspace-image.md#Registry).

### `OQ-SRV-2`

How should a child process in the control plane receive updated certificate trust when it can't reload that trust while running? Restarting it can interrupt active service, while retaining its old trust delays removal of retired authority. afakt could coordinate a bounded restart or place authentication in a separately supervised component that supports trust updates. See [Registry](components.md#Registry).

## Command-line reference

### `OQ-CLI-1`

Which content sources can an index register and where does indexed content live? What query interface does an agent use and how does indexing relate to memory queries and artifacts? See [Index](cli.md#Index).

### `OQ-CLI-2`

What scope and output format does `report` select, and how does it combine work and usage across organizations and projects? See [Deployment and client commands](cli.md#Deployment%20and%20client%20commands).

### `OQ-CLI-3`

How do configuration scopes and precedence interact with deployment policy, agent profile bindings, and per-agent settings? See [Configuration and indexing](cli.md#Configuration%20and%20indexing).

## Public key infrastructure

### `OQ-PKI-1`

What operational contract must an issuer of workload identities provide for attestation, registration, issuance, key custody, and trust distribution under either issuer arrangement in [issuer ownership](agent-open-questions.md#%60OQ-SBX-33%60)? Which component provides each capability, and how are its administrative permissions separated from workload requests? See [Issuer capability](pki.md#Issuer%20capability).

### `OQ-PKI-2`

Which algorithms, key sizes, certificate profiles, and lifetime limits apply to each authority and credential class? What root and intermediate hierarchy, online or offline custody, name constraints, path lengths, extended key usages, and issuance rate limits apply? How do these requirements preserve the held [identity lifetime decision](agent-open-questions.md#%60OQ-HOOK-18%60)? See [Certificate policies](pki.md#Certificate%20policies).

### `OQ-PKI-3`

How are login tokens, session cookies, personal access tokens, preview grants, and registry tokens generated, stored, verified, rotated, and revoked? Which signing keys and public key sets are separate, what algorithms and validity limits apply, and how do replicas handle key distribution, cache expiry, and recovery? See [User authentication](pki-signing.md#User%20authentication).

### `OQ-PKI-4`

Which coordinator, node, enrollment, and relay keys does the overlay keep, and who generates and verifies each one? What algorithms, sizes, storage protections, expiry, rotation, and compromise recovery apply independently of workload identity? See [Overlay credentials](pki-transport.md#Overlay%20credentials).

### `OQ-PKI-5`

Who owns public certificate issuance, automated renewal, account recovery, and certificate distribution for the external ingress gateway? Which public authorities, account and domain-validation credentials, key algorithms, lifetimes, storage protections, and replica coordination apply to ordinary and wildcard names? See [External ingress gateway](pki-transport.md#External%20ingress%20gateway).

### `OQ-PKI-6`

What hierarchy, algorithms, key sizes, validity limits, custody, and rotation apply to inspection authorities and destination certificates? How does each guest operating system and runtime install, reload, and remove public inspection trust without exposing private keys or changing unrelated host trust? See [Inspection certificates](pki-transport.md#Inspection%20certificates).

### `OQ-PKI-7`

Which certificate authorities, identities, names, algorithms, key sizes, lifetimes, storage protections, and renewal mechanisms secure the registry, object store, operational data store, and workflow service? How do external endpoints and supervised child processes receive and retire trust? See [Storage and workflow transport](pki-transport.md#Storage%20and%20workflow%20transport).

### `OQ-PKI-8`

Which authorities sign workspace images, guest images, snapshots, portable exports, and binary releases, and who verifies them before use? What key custody, algorithms, sizes, validity, rotation, revocation, time stamping, notarization, and platform entitlements apply to release binaries and privileged components? See [Artifact and platform signing](pki-signing.md#Artifact%20and%20platform%20signing).

### `OQ-PKI-9`

How does the secret store encrypt secrets at rest and protect its master keys, data encryption keys, and backups? Which algorithms, key sizes, custody tiers, authority to open the encrypted store, rotation periods, recovery permissions, and destruction rules apply? See [Secret storage](pki-signing.md#Secret%20storage).

### `OQ-PKI-10`

What administrative commands, permissions, audit guarantees, metrics, alert thresholds, backup rules, and recovery procedures govern public key infrastructure? How are issuance and its audit record reconciled after a partial failure, and what proves that a restored authority hasn't restored revoked access? See [Public key infrastructure operations](pki-operations.md).

### `OQ-PKI-11`

How are inbound and outbound webhook signing secrets provisioned, stored, rotated, and revoked? Which keyed hash algorithm, secret size, signed fields, delivery identity, replay window, and overlap rule apply to each provider and destination? See [Webhook authentication](pki-signing.md#Webhook%20authentication).
