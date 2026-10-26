---
type: Design
title: Agent open questions
description: Undecided design questions for agents and their executions, linked to the sections they concern.
tags: [afakt, agents, open-questions]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent open questions

These undecided design questions for agents and their executions link to the sections they concern. Each question keeps its ID when answered or removed, and afakt never reuses an ID.

## Agents

### `OQ-AGT-1`

Which harness process state does suspend and resume restore, and which channels or helpers must restart? Compatible VM restoration is the leading option for preserving live process state, subject to qualification for each backend and harness. Paired workspace and native-session recovery is a candidate cold alternative that loses live state. Adopting it requires an explicit change to the rule requiring VM-snapshot checkpoints and qualification of complete native-history restoration. Migration remains a separate question. See [Suspend, resume, and restart](agent.md#suspend-resume-and-restart).

### `OQ-AGT-2`

How should afakt handle a suspend request during an active turn before capturing a checkpoint? Waiting for an idle turn boundary is the leading option for capturing consistent workspace and session state, at the cost of delaying suspension while the turn continues. The wait bound and handling of a turn that doesn't become idle remain open. See [Suspend, resume, and restart](agent.md#suspend-resume-and-restart).

### `OQ-AGT-3`

Which process, device, and session details must a checkpoint preserve alongside its VM snapshot and workspace delta? A verified checkpoint set captured after writers are quiescent is the leading option, at the cost of pausing work and validating restored state before accepting turns. The set must bind workspace, native-session, memory, and external device state. Its exact capture scope remains open. See [Checkpoints](agent.md#checkpoints).

### `OQ-AGT-4`

How should a conversation branch created by rewind represent external effects? The leading option for preserving known external actions retains the original branch with a separate effect ledger, at the cost of retaining audit data alongside the new branch. The ledger must record incomplete coverage because tool and permission records don't cover every effect. Rewind can't undo those actions, and their presentation remains open. See [Checkpoints](agent.md#checkpoints).

### `OQ-AGT-5`

May `afakt agent mcp add` attach a server the deployment hasn't configured? Profiles select configured servers and the harness uses one gateway endpoint. How do local server attachments fit those restrictions? Registration in the deployment's server catalog before attachment is the leading option for keeping gateway grants tied to known server identities. It adds an approval or configuration step for a new server. One gateway URL constrains configured access but can't itself prevent local tool processes from using another protocol. The permitted attachment model remains open. See [Owned resources](agent.md#owned-resources).

### `OQ-AGT-6`

What are agentmail's delivery guarantees, ordering, acknowledgement, scheduling, and undo semantics? A durable outbox with retries and recipient deduplication is the leading delivery option, at the cost of retained replay records. Cancellation ends when dispatch may have begun, and external effects require their own idempotency guarantees. Ordering, scheduling, and acknowledgement policy remain open. See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

### `OQ-AGT-7`

What identity signs agentmail, how does historical verification prevent replay, and how do signatures relate to signed session messages? Signing from authenticated sender context with retained verification keys and replay records is the leading option. It supports historical verification but requires retention and revocation rules. By themselves, signatures don't prove app acknowledgement or external completion. The signing contract and its relationship to session signatures remain open. See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

### `OQ-AGT-8`

Where does memory live, what does a query match, and how do copying, compaction, retention, and access control affect it? Immutable memory versions with authorized lexical queries and atomic content, index, and reference updates are the leading option for exact retrieval. Copying an authorized version gives the destination independent content, while compaction preserves pinned versions. This keeps access and retention explicit but doesn't supply semantic ranking or secure erasure. The storage service and query language remain open. See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

### `OQ-AGT-9`

How do one-run executions map to logical execution IDs and generations? Deliberate attempts share an ID, supersession changes the generation, and reconciliation retries reuse it. Which run boundaries follow an agent restart, resume, or migration? Preserving the durable agent and session while pinning a separate run binding is the leading identity structure, at the cost of another ownership fence. Both a new execution and a later generation can preserve that structure, so restart, resume, and migration still need explicit boundary rules. See [Lifecycle](agent.md#lifecycle).

### `OQ-AGT-10`

How's agent ownership represented in the execution specification, AgentExecution resource, provenance, and execution token? Execution IDs and generation identity remain distinct from agent identity. An immutable agent and session binding on each accepted run is the leading option, with mediator requests checked against that binding and the active ownership fence. It preserves historical provenance when ownership changes but adds coordination between run acceptance, transfer, and request admission. Public field names and token claims remain open. See [Owned resources](agent.md#owned-resources).

### `OQ-AGT-11`

How does durable ownership protect the current workspace and session transcript when generation evidence expires? Which references keep workspace images, revisions, snapshots, and shared artifacts from collection? Independent durable roots for the agent's workspace, transcripts, checkpoints, and shared content are the leading option. Reference creation and collection share a fence, and unresolved releases keep their roots. This protects content beyond generation expiry but can keep extra data and delay collection. Retention periods and the authoritative catalog remain open. See [Owned resources](agent.md#owned-resources).

### `OQ-AGT-12`

How does migration fit the rule that every execution generation and its workspace sources stay on the first agent host? A verified cold transfer after observed source termination is the leading candidate across incompatible backends. It must preserve the complete workspace and native conversation state, but migration with real harnesses and VM backends remains unqualified. Processes and channels restart, and an absent response doesn't establish termination. The execution boundary for a new placement remains open. See [Export, import, and migrate](agent.md#export-import-and-migrate).

### `OQ-AGT-13`

When do `start`, `run`, `restart`, and `restore` address an existing agent versus create another? Reusing an agent preserves its durable identity, while creating another separates ownership and history. Command defaults remain open under either option. See [Lifecycle](agent.md#lifecycle).

### `OQ-AGT-14`

What ownership and status apply to an uncertain launch or stop? Retaining ownership until afakt observes termination is the leading rule. It prevents overlapping replacements but can leave the agent unavailable during reconciliation. Intermediate status names remain open. See [Lifecycle](agent.md#lifecycle).

### `OQ-AGT-15`

How should deletion handle created or suspended agents, and which states count as unused for prune? Restricting removal to agents with resolved execution ownership avoids discarding an uncertain run, but can delay cleanup. A broader rule needs explicit treatment of retained workspace and session state. Deletion behavior and prune eligibility remain open. See [Lifecycle](agent.md#lifecycle).

### `OQ-AGT-16`

For forks, snapshots, archives, Open Container Initiative (OCI) exports, and imports, which resources does afakt copy, reference, or exclude? A content-verified envelope with explicit resource references and destination reauthorization is the leading candidate. Keeping historical audit data separate from active identity and credentials avoids transferring authority with content, but portability requires a supported metadata contract, including extended attributes, and qualified native-store restoration. Identity mapping and the complete resource set remain open. See [Export, import, and migrate](agent.md#export-import-and-migrate).

### `OQ-AGT-17`

Should unarchive always return an agent to stopped state? Returning to stopped state is the leading option because it avoids starting work during restoration, at the cost of a separate start operation. Whether any case permits automatic start remains open. See [Snapshots and the snapshot registry](agent.md#snapshots-and-the-snapshot-registry).

### `OQ-AGT-18`

How do per-agent credential attachments map to references on the profile and execution? How do authorization roles and organization and project scopes affect that mapping? Intersecting an execution's attached references with those the pinned profile permits is the leading candidate check, alongside target-resource authorization. It limits attachment choices but requires a rule for changes during a run. It doesn't establish the profile's own authority or the credential owner's permitted destinations and headers. Repository profile self-grant and credential domain bounds, including the bound on execution references, remain on hold. Authorization roles remain distinct from reporting labels. Inheritance and live attachment policy remain open. See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

### `OQ-AGT-19`

What rules govern ordering, retries, retention, and inactive recipients for group broadcasts and subscription deliveries? A durable outbox with acceptance-time membership and event deduplication is the leading delivery option. It makes retries stable but retains accepted deliveries after membership changes. Queueing for suspended or stopped recipients avoids implicit resource use. Handling archived recipients remains a separate open question. Waking requires separately authorized, retry-safe lifecycle requests. Their completion semantics remain open, as do ordering and retention bounds. See [Agentmail, groups, and subscriptions](agent.md#agentmail-groups-and-subscriptions).

### `OQ-AGT-20`

What access, content isolation, retention, and indexing rules apply to artifact publication? Authorized, expiring publication of immutable artifacts on separate origins is the leading candidate. Download-only delivery reduces active-content exposure but excludes inline applications. Browser isolation and atomic content, index, and reference updates across the selected storage services still need qualification. The storage and query contracts remain open, and revocation can't recall admitted downloads or client-held copies. See [Memory, skills, and artifacts](agent.md#memory-skills-and-artifacts).

### `OQ-AGT-21`

How do standalone sandbox identities, keychain authorizations, SSH and GPG key stores, known hosts, and certificate trust relate to agent and execution authority? What controls key use, rotation, and revocation? Purpose-scoped delegated key use with private material outside the sandbox is the leading option. Keeping durable revocation state and rebuilding trust contexts prevent old authority from returning after restart, but rotation also requires closing affected connections. Retaining public keys supports historical verification without permitting new signing. Compromise semantics and native key-store integration remain open. See [Secrets, credentials, and trust material](agent.md#secrets-credentials-and-trust-material).

### `OQ-AGT-22`

How do agent budgets account for usage across runs and external services? What happens at exhaustion, and which operations remain available? Durable reservations before dispatch, followed by reconciliation with actual usage, are the leading accounting option. Keeping unknown bills reserved prevents premature reuse but can withhold budget indefinitely. A hard spending bound also needs an enforceable upstream usage ceiling. Observation alone can only record an overage. Exhaustion behavior and permitted operations remain open. See [Budgets and usage](agent.md#budgets-and-usage).

### `OQ-AGT-23`

Which network paths does quarantine block, including brokered API calls, ACP control traffic, telemetry, MCP calls, and forwarded ports? What permissions restore connectivity? Explicit blocked-path sets with checks at dispatch and revocation of active flows are the leading mechanism. Denying new connections alone leaves established TCP and streaming traffic usable. Closing those flows enforces quarantine but interrupts in-flight work. Whether ACP, telemetry, and brokered control remain available, and who may restore them, remains open. See [Kill, signal, and quarantine](agent.md#kill-signal-and-quarantine).

### `OQ-AGT-24`

How do steer and queue-message apply to workflow-controlled or headless runs without widening their writer rules? What boundary does `wait` observe when the transcript persists after the harness exits? See [Attach and messages](agent.md#attach-and-messages).

### `OQ-AGT-25`

Which workspace revision does restart use, and how do agent workspace commits and resets relate to generation copies and workspace image publication? Pinning the selected revision before restart is the leading option for stable retries, at the cost of retaining that revision through preparation. Both the latest retained workspace and an explicitly selected older revision fit this mechanism. Independent writable copies preserve history, while workspace image publication needs separate authorization. The default revision and reset semantics remain open. See [Workspace revisions and the workspace registry](agent.md#workspace-revisions-and-the-workspace-registry).

### `OQ-AGT-26`

What authority permits `agent exec` and `agent shell` commands outside the sandbox? The provider confines the harness and its helpers, so running a command outside that confinement, and choosing the command's sandbox mode, needs authority of its own. A separate operator grant bound to the target sandbox and requested execution mode is the leading option. It keeps an agent's run authority from becoming authority to execute outside confinement, but adds a distinct authorization path. Which operators receive that grant remains open, as do the permitted modes. See [Sandboxes](agent.md#sandboxes).

## Agent host

### `OQ-HST-1`

After a whole-machine restart, how does the host reconcile executions whose VMs and mediators stopped, and what observations determine their outcomes and slot release? Worker restart keeps those processes alive, while lease expiry alone doesn't change executions or reservations. See [Restart and recovery](agent-host.md#restart-and-recovery).

## Agent execution

### `OQ-EXE-1`

Where should execution records persist? Each record stays on its agent host across a worker restart while the agents keep running. A central store that any worker can read would replace that, and so would ending every agent when its worker restarts. See [Reconciliation after restart or retry](agent-execution.md#reconciliation-after-restart-or-retry).

### `OQ-EXE-2`

What happens when an agent host never comes back? Its execution then stays `unknown-stopped` or keeps returning `unavailable`, and a higher generation waits in `pending` behind it indefinitely. Options include an `abandoned` outcome that an operator declares and a deadline after which the execution ends by itself. See [Generations and supersession](agent-execution.md#generations-and-supersession).

### `OQ-EXE-3`

How does a higher generation treat an active previous generation? Here the provider cancels the previous generation and holds the new one in `pending` until the previous one ends. Rejecting the launch while the previous generation is still active would be the alternative. See [Generations and supersession](agent-execution.md#generations-and-supersession).

### `OQ-EXE-4`

Should the term "execution provider" stay beside "model provider" in the design? Both terms appear under these rules. One of them could shrink to "provider," or the execution provider could take another name, such as "runner backend." See [Parts](agent-execution.md#parts).

### `OQ-EXE-5`

Which afakt command starts a worker that runs the agent runner or the agent executor? The options are a command of its own in the afakt CLI, a separate worker program, or a mode of the server command. See [Parts](agent-execution.md#parts).

### `OQ-EXE-6`

Should a restarted worker resume an interrupted step instead of failing the execution? An execution ends `failed` when a restart interrupts a step before the agent starts. The alternative runs the interrupted step again, which needs each step, `before-agent` hooks included, to be safe to repeat. See [Reconciliation after restart or retry](agent-execution.md#reconciliation-after-restart-or-retry).

### `OQ-EXE-7`

How does a specification select the deterministic test provider? The options are a reserved harness name, a placement condition that leads to agent hosts with the test provider, or a deployment setting. See [Parts](agent-execution.md#parts).

## Agent execution model

### `OQ-MOD-1`

Which part of afakt assigns the execution ID? afakt assigns it when it creates the logical execution, and the parent workflow records it before the first launch. The parent workflow could assign it, and so could the agent runner workflow or the API server when a request arrives. See [Execution ID format](agent-execution-model.md#execution-id-format).

### `OQ-MOD-2`

Should a provider reject a UUIDv7 whose timestamp lies in the future or is implausibly old? Here the provider checks only the text form, the version digit, and the variant digit. The alternative also rejects a timestamp outside a window around the provider's clock, such as one later than the clock or older than the longest retention. See [Execution ID format](agent-execution-model.md#execution-id-format).

### `OQ-MOD-3`

Should labels count toward the specification digest? They do under these rules, so a repeated launch that changes only its labels fails with `specification-conflict`. Leaving labels out of the digest would let such a launch return the existing execution. See [Execution identity](agent-execution-model.md#execution-identity).

### `OQ-MOD-4`

How should the provider settle conflicting observations? It keeps the first value and marks the record as conflicting, with the later report in the evidence. Keeping the latest value, or every value, would also work. See [Provenance record](agent-execution-model.md#provenance-record).

### `OQ-MOD-5`

Where does evidence live, and how does it leave the agent host? Beyond the published evidence references, and callers that never read the agent host's disk, that remains open. One option uploads each item to a central store at publication, and another fetches it from the agent host through the afakt API. See [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup).

### `OQ-MOD-6`

Which digest algorithm applies? Only an example, `sha256:`, appears in the design. Every digest could use Secure Hash Algorithm (SHA) 256, or a fixed set of named algorithms could apply. See [Execution generation](agent-execution-model.md#execution-generation).

### `OQ-MOD-7`

What format do agent host identities and afakt worker versions take? Neither has one here. A host identity could be a name from the agent host's configuration or a generated ID. The worker version could be the afakt release version or the build's commit ID. See [Provenance record](agent-execution-model.md#provenance-record).

### `OQ-MOD-8`

What format do evidence reference IDs take? The specification gives none. Candidates are the item's digest, a UUIDv7, or an opaque text value from the provider. See [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup).

## Agent execution contract

### `OQ-CON-1`

Should cleanup become a fifth provider operation? Cleanup runs as an activity on the host task queue, outside launch, status, cancel, and result. See [Provider operations](agent-execution-contract.md#provider-operations).

### `OQ-CON-2`

How should an execution with invalid output end? Here it ends `failed`. It could instead end `succeeded` without a payload. See [Execution result](agent-execution-contract.md#execution-result).

### `OQ-CON-3`

What's the default retention period? It comes from a deployment default that has no fixed value. The choice lies between a value written in the contract and a value that each deployment sets for itself in its own configuration. See [Execution specification](agent-execution-contract.md#execution-specification).

### `OQ-CON-4`

Should cancel on a terminal execution return an acknowledgement? Cancel acknowledges with the terminal state and doesn't change anything. Otherwise cancel would return an error for a terminal execution. See [Cancel](agent-execution-contract.md#cancel).

### `OQ-CON-5`

Should status answer for older generations? Status reports any generation. Otherwise it would cover the current generation only. See [Status](agent-execution-contract.md#status).

### `OQ-CON-6`

What format do label values take? Task class, variant, and role have none here, and each deployment defines their values. The profile name rule or a list in the deployment's configuration that the provider checks would each work. See [Execution specification](agent-execution-contract.md#execution-specification).

### `OQ-CON-7`

Where do the prompt and attachments live? The specification holds references with digests. The options are the evidence store, a content store behind the afakt API, or the prompt inline in the specification with attachments by reference. See [Execution specification](agent-execution-contract.md#execution-specification).

### `OQ-CON-8`

Which resource limits exist, and what are their names and units? The specification gives a token budget, CPU, memory, and disk as examples. The options are a fixed set in the contract or a set that each provider publishes. See [Execution specification](agent-execution-contract.md#execution-specification).

## Agent placement

### `OQ-PLC-1`

How long do a [registration lease](agent-placement.md#agent-host-registry) and an unclaimed reservation last? Both lifetimes have bounds but lack fixed values. A short lease takes a lost agent host out of placement quickly, at the cost of more renewals.

### `OQ-PLC-2`

Which matching agent host does the [reservation API](agent-placement.md#reservation-api) pick? Picking the host with the most free slots spreads load, while filling one host first leaves whole hosts free for larger work.

### `OQ-PLC-3`

Where does the [registry](agent-placement.md#agent-host-registry) keep its records? Should the API server's database store them? Temporal's worker list could report which workers are polling and their slot usage, but it doesn't include afakt's host metadata.

### `OQ-PLC-4`

What measurement interval and downward step size should the [capacity controller](agent-placement.md#slot-capacity) use? Increases take one slot at a time. Decreases can remove more than one slot at once within the configured bounds.

### `OQ-PLC-5`

Should rate limits from the harness or model provider feed the [capacity controller](agent-placement.md#slot-capacity) alongside host resource use?

### `OQ-PLC-6`

Should [capacity](agent-placement.md#slot-capacity) account for each execution's resource requests? Agent profiles declare resource limits but don't declare resource requests.

## Agent execution resource

### `OQ-RES-1`

Which group should `apiVersion` use? It uses `afakt/v1alpha1`. Kubernetes custom resources use a domain-style group instead. See [Envelope](agent-execution-resource.md#envelope).

### `OQ-RES-2`

Should keys share one naming style? Envelope keys follow Kubernetes in camel case, and afakt's own keys use snake case. Sandbox keys keep the camel case of the sandbox runtime. Camel case for every key would remove the mix. See [Specification keys](agent-execution-resource.md#specification-keys).

### `OQ-RES-3`

Does one resource need every generation? One resource shows one generation under these rules. A logical execution resource could list its generations under `status`. See [Envelope](agent-execution-resource.md#envelope).

### `OQ-RES-4`

What belongs in `status.result`? Here the execution result draws its other fields from `status`. A `status.result` that repeats them would make the result complete without the rest of `status`. See [Status keys](agent-execution-resource.md#status-keys).

## Agent workspace

### `OQ-WSP-1`

How do changes leave the workspace? As drafted, the agent pushes through the `branches.push` capability, and the workspace changes also go into the evidence. The parent workflow could instead publish the changes from the evidence, in place of a push from the agent. See [Workspace capture](agent-workspace.md#workspace-capture).

### `OQ-WSP-2`

Can a profile restrict which repositories a specification lists as sources? Here a specification may list any repository URL. A profile field of allowed repositories or URL patterns would add that restriction. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-3`

What format does a commit ID take? Only the words full commit ID appear on this page. Secure Hash Algorithm (SHA) 1 alone is one option, and SHA-1 together with SHA-256 is another, in lowercase hexadecimal either way. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-4`

How are repository URLs normalized? `${repository:<target path>}` drops the scheme, but which schemes a source may use and how the scheme-free form comes about are still open. The options are HTTPS alone, with the host and path in lowercase and without a trailing `.git`, or every scheme that Git accepts for a remote, mapped to one scheme-free form. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-5`

Should new branch names have rules beyond those of the version control system? This draft adds none. The alternative is a prefix that each deployment sets, such as one for agent branches, or a pattern in the profile. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-6`

Is `AFAKT_` the right prefix for the execution environment? Every variable that the provider sets uses it. The alternative is a longer prefix, such as `AFAKT_EXECUTION_`, that leaves `AFAKT_` to the settings of the afakt CLI. See [Execution environment](agent-workspace.md#execution-environment).

### `OQ-WSP-7`

How should the environment describe the workspace sources? The provider numbers them, as in `AFAKT_SOURCE_<n>_PATH`. Each name could encode the target path instead, or a JSON file could list the sources. See [Execution environment](agent-workspace.md#execution-environment).

### `OQ-WSP-8`

Should the execution environment expose the labels? Task class, variant, and role stay out of it as drafted. The alternative adds `AFAKT_LABEL_TASK_CLASS`, `AFAKT_LABEL_VARIANT`, and `AFAKT_LABEL_ROLE`. See [Execution environment](agent-workspace.md#execution-environment).

### `OQ-WSP-9`

Should every execution environment variable have a sandbox variable? `AFAKT_PROFILE_SOURCE`, `AFAKT_PROFILE_DIGEST`, `AFAKT_SOURCE_COUNT`, `AFAKT_SOURCE_<n>_TARGET`, `AFAKT_SOURCE_<n>_KIND`, `AFAKT_SOURCE_<n>_COMMIT`, `AFAKT_TIME_LIMIT`, `AFAKT_DEADLINE`, and `AFAKT_PHASE` have none. The alternative adds a sandbox variable for each one that keeps its launch value, which leaves out `AFAKT_PHASE`. See [Execution environment](agent-workspace.md#execution-environment).

### `OQ-WSP-10`

Should locking stay a field of each profile entry? A profile marks single entries with `locked` in this draft. One setting on the profile's environment could instead make the whole list fixed or open to additions. See [Combining environment entries](agent-workspace.md#combining-environment-entries).

### `OQ-WSP-11`

May a specification replace profile entries without `locked`? It may as drafted, or it could add new names only. See [Combining environment entries](agent-workspace.md#combining-environment-entries).

### `OQ-WSP-12`

What should these pages call the service that hosts source repositories? Repository host is the draft's term, as in `${repository-host:<target path>}`. Source host and code host are the alternatives. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-13`

How does the provider choose a fetch credential? Here the agent host's configuration gives one for each repository host. The profile could instead name a fetch credential reference for each source, which placement would then require. See [Workspace sources](agent-workspace.md#workspace-sources).

### `OQ-WSP-14`

How do guest paths for home, scratch, and named shared caches fit the agent user's workspace-only writes? Workspace-backed writable paths are the leading option, but must preserve scratch outside the workspace root and the existing cache-sharing contract without exposing host paths. The provider-installed hook protection decision remains [open](agent-open-questions.md#agent-hooks). See [Reserved names](agent-workspace.md#reserved-names).

## Agent workspace image

### `OQ-IMG-1`

Does afakt need the OCI referrers API for image signatures, software bills of materials (SBOMs), or other attached artifacts? Which verification policy would consume them? For the Open Container Initiative (OCI) registry, see [Registry](agent-workspace-image.md#registry).

### `OQ-IMG-2`

What registry topology fits a multi-host fleet: one registry or one per site? What metadata storage can registry replicas share while preserving tag updates and online collection? One central registry with authorized, verified agent host caches is the leading option. Site copies cut repeated transfers. Their cost is stale tags and duplicated storage, plus publication ownership and retention reconciliation to settle, while shared storage alone doesn't coordinate writers or collection, which leaves replica fencing and the metadata backend open. See [Registry](agent-workspace-image.md#registry).

### `OQ-IMG-3`

How should peer-to-peer layer sharing between agent hosts work at the fleet stage, and how would a peer check permission for each repository pull without treating knowledge of a digest as access? Independent repository authorization on every cache read, followed by digest verification, is the leading access rule for any peer source. It prevents a warm cache from extending a revoked grant but requires an available authorization authority. This narrows the permission check. Peer discovery, transfer, and trust remain open. See [Agent host caches and startup](agent-workspace-image.md#agent-host-caches-and-startup).

### `OQ-IMG-4`

How should workspace access density inform the policy for caching and preheating? Fetching complete images in advance preserves the no-lazy-pull rule but transfers content a run may never read. Selecting which images to preheat reduces that transfer at the cost of cold starts. Selective preheating that yields to disk pressure is the leading option. Reusing verified complete content avoids repeated transfer and extraction, but sparse access can leave most bytes fetched in advance unused. Per-digest admission and space reservations bound concurrent preparation. Workload-based thresholds remain open. See [Placement and preheating](agent-workspace-image.md#placement-and-preheating).

### `OQ-IMG-5`

Can afakt save a finished execution's workspace as a workspace image, and what provenance and publication permissions would that need? Capturing a verified stopped workspace under an explicit publication grant is the leading option. Pinning its source revision, profile, and provenance before publishing a complete image means retries use the same inputs, but requires separate publication and retention records. An agent snapshot also contains conversation and audit state, so converting its workspace doesn't define the scope of those other resources. See [Checkpoints and portable revisions](agent-workspace-image.md#checkpoints-and-portable-revisions).

### `OQ-IMG-6`

How do secrets that a build step needs stay out of every image layer? Brokered network access is the leading option where supported because it keeps credential values outside captured content, but it can't supply every local secret input. Every inherited and intermediate layer remains part of the disclosure boundary. Deleting a secret from the final tree doesn't remove it from earlier layers, and final-tree checks and known-secret scans can't prove secret absence. A separate delivery and capture contract is necessary for local secret inputs. See the [image build rules](agent-workspace-image.md#image-builds).

### `OQ-IMG-7`

Should the source rules allow a `repository` source immediately after an `image` source at the same target path? That would require an exception to the duplicate-target rule and a check that the image contains the same repository. After fetching and checking out the newer commit, rules would also need to cover dirty tracked files and stale dependencies plus outputs. Nested targets would still need their own decision. See [Image sources and digest pinning](agent-workspace-image.md#image-sources-and-digest-pinning).

### `OQ-IMG-8`

Must the worker fetch the image layers before VM boot, or may the guest pull them through its mediator? Worker pulls are the leading option because they reuse verified disk bases and resolve preparation failures before session launch. Guest pulls add a mediated preparation phase and guest resource pressure. Whether they provide the same caching, authorization, and verification behavior remains unqualified. Both require complete verified content before hooks or the harness start. See [Agent host caches and startup](agent-workspace-image.md#agent-host-caches-and-startup).

## Agent hooks

### `OQ-HOOK-1`

Should hooks run when the profile lacks `run-commands`? The provider runs them anyway, because harness permissions don't limit hooks and the sandbox confines them. The alternative requires `run-commands` for any profile or specification with hooks. See [Running hooks](agent-hooks.md#running-hooks).

### `OQ-HOOK-2`

Should some `after-agent` hooks also run after a limit or a cancellation? The provider skips them then and runs only the `session-end` hooks, within the session-end budget. The alternative runs the `exit` hooks too, under a time budget of their own outside the time limit. See [Hook events](agent-hooks.md#hook-events).

### `OQ-HOOK-3`

Should an `after-agent` failure fail the execution? An execution ends `failed` with hook failure as its reason. The alternative records the failure as evidence and keeps the agent's outcome. See [Hook failures](agent-hooks.md#hook-failures).

### `OQ-HOOK-4`

Should hooks read the instruction inputs? Hooks can't read instruction inputs. A variable such as `${input:<key>}` could expose them. See [Hook environment](agent-hooks.md#hook-environment).

### `OQ-HOOK-5`

Should a hook return structured output, such as test results, that joins the payload? Hook output stays in evidence. A field of hook results beside the payload or a hook output file that the payload's schema can include would expose it in the result. See [Hook evidence](agent-hooks.md#hook-evidence).

### `OQ-HOOK-6`

Should `after-agent` hooks also receive the validated payload as a file? Through `AFAKT_FINAL_ANSWER_FILE`, hooks receive the raw final answer. A second file could hold the validated payload, unset when validation fails. See [Hook environment](agent-hooks.md#hook-environment).

### `OQ-HOOK-7`

May a hook's own environment entries replace agent environment entries without `locked`? Hooks may replace those values. The alternative lets a hook add new names only. See [Hook environment](agent-hooks.md#hook-environment).

### `OQ-HOOK-8`

Should the provider record each run of a workspace repository hook? The provider records only the entries it finds in the project configuration, since those hooks run outside the afakt CLI. Recording each run would need the provider to wrap or watch the harness's own hook runner. See [Workspace repository hooks](agent-hooks.md#workspace-repository-hooks).

### `OQ-HOOK-9`

Is `session-end` right as a provider event? Claude Code and Codex each have their own `SessionEnd` hook, and Antigravity has none. The harness runs that hook from inside its own process. When the provider stops the harness by force after a cancellation or a limit, that hook never runs, and both harnesses also cut it off after a few seconds. Provider hooks run after a headless harness exit or ACP session closure. Both normal completion and forced stops trigger them. Mapping `session-end` to the harness's native event instead would miss forced stops. See [Hook events](agent-hooks.md#hook-events).

### `OQ-HOOK-10`

Is 10 seconds right for the session-end, stop, and teardown budgets? Each value is a guess. A profile or a deployment could set its own instead. See [Running hooks](agent-hooks.md#running-hooks).

### `OQ-HOOK-11`

Which decision applies when hooks at `model-request-after` return both `continue` and `stop` for one occurrence? The design leaves open whether `stop` always takes precedence or the first hook's decision applies. Another option gives the profile's hooks precedence over the specification's. See [Decisions](agent-hooks.md#decisions).

### `OQ-HOOK-12`

Should `rewrite` be available where the harness ties it to an approval? Claude Code and Codex apply new tool arguments only with `allow` or `ask`, and for that reason the matrix marks `rewrite` unsupported there. Another option returns `allow` with the new arguments only when the profile's harness permissions already allow the call. See [Harness capability matrix](agent-hooks.md#harness-capability-matrix).

### `OQ-HOOK-13`

Should the `before-teardown` hooks run before the terminal state? The provider runs them before the result, as it does the `session-end` hooks, so they still use brokered credentials and the execution authority that the mediator supplies. Running them after the terminal state would return the result sooner, but without the token. See [Running hooks](agent-hooks.md#running-hooks).

### `OQ-HOOK-14`

Should `stop-requested` also run at a resource limit? The provider stops the agent right away there, because the agent would keep using the resource while the hooks run. See [Running hooks](agent-hooks.md#running-hooks).

### `OQ-HOOK-15`

Should a `sandbox-violation` hook also run for violations by hook processes? The event covers only the agent's processes, so that a hook can't set off its own event. See [Hook events](agent-hooks.md#hook-events).

### `OQ-HOOK-16`

What protects provider-installed hooks from changes by the agent, including hooks that share its sandbox and identity? Protection of hook configuration and dispatch remains open. See [Hook dispatch](agent-hooks.md#hook-dispatch).

### `OQ-HOOK-17`

Which repository project settings may the harness load beyond hooks, and how do they interact with provider settings? The scope of repository project configuration remains open. See [Workspace repository hooks](agent-hooks.md#workspace-repository-hooks).

### `OQ-HOOK-18`

How does the session-end wait interact with the execution token's lifetime? The wait, the session-end budget, and token renewal bounds need a shared rule. See [Running hooks](agent-hooks.md#running-hooks).

## Agent observability

### `OQ-OBS-1`

Should an execution's spans join the parent workflow's trace? They join it, with one `afakt.execution` span for each generation and links between generations. Each execution could instead have a trace of its own, linked to the parent's span. See [Trace topology](agent-observability.md#trace-topology).

### `OQ-OBS-2`

Should a restart start linked recovery spans? It does here, and it sets `afakt.recovered` on each one. An alternative has the afakt worker export each open span at intervals, and a restart then loses only the end of each span. See [Links](agent-observability.md#links).

### `OQ-OBS-3`

Is the telemetry content field right? Its values are `omit` and `include`, and `include` covers prompts, responses, instructions, and instruction inputs. Its name, its values, and the place of instruction inputs under it are each open. The alternatives are a separate value for instruction inputs, such as `include-inputs`, or one boolean field for each kind of content. See [What telemetry never records](agent-observability.md#what-telemetry-never-records).

### `OQ-OBS-4`

Where does the agent host's telemetry endpoint run? It could belong to the afakt worker or run as a separate component. See [Trace context propagation](agent-observability.md#trace-context-propagation).

### `OQ-OBS-5`

Should hook names be a metric attribute only for profile hooks? `afakt.hook.name` appears for profile hooks and not for hooks from a specification, whose names a caller chooses freely. Hook names could instead appear for every hook, or for none. See [Cardinality](agent-observability.md#cardinality).

### `OQ-OBS-6`

Is the agent host an acceptable metric attribute? The active, unresolved, slot, and cleanup metrics include `afakt.agent_host`, with as many values as a deployment has agent hosts. Leaving it to traces and logs would avoid that growth. See [Cardinality](agent-observability.md#cardinality).

## Agent profile

### `OQ-PRF-1`

Where does the profile catalog live, and who may publish to it? The catalog could live inside afakt behind the afakt API or in repository files. Another store would also work. See [Profile sources](agent-profile.md#profile-sources).

### `OQ-PRF-2`

How are catalog versions numbered? Positive integers identify catalog versions, with the profile digest as a check. Semantic versions and pinning by digest alone are alternatives. See [Profile identity](agent-profile.md#profile-identity).

### `OQ-PRF-3`

Should catalog profiles use the same YAML format as repository profile files? Repository profile files and catalog publish requests use the same resource form. Publishing catalog versions through the afakt API as JSON with the same fields is the alternative. See [Profile resource](agent-profile.md#profile-resource).

### `OQ-PRF-4`

Should catalog and repository profile names stay apart? Here neither source shadows the other, so a catalog `review` and a repository `review` are different profiles. One shared set of names would instead let a repository profile shadow a catalog profile, or make a clash fail the launch. See [Profile sources](agent-profile.md#profile-sources).

### `OQ-PRF-5`

What bounds a repository profile's self-grants and credential inject domains? Repository changes can grant `branches.push` or wide network access to the repository's own profile. A deployment-wide ceiling or a catalog profile could bound these grants and credential inject domains. See [Repository profile files](agent-profile.md#repository-profile-files).

### `OQ-PRF-6`

Which commit should a parent pin for a repository profile? A parent chooses the commit, such as the commit of the workspace source. Always using the workspace source commit or the head of the repository's default branch at resolution would remove that choice. See [Pinning a profile version](agent-profile.md#pinning-a-profile-version).

### `OQ-PRF-7`

Should templates have functions beyond Go's built-in ones? Rendering uses the built-in functions alone, and the alternative adds a function library such as Sprig. See [Instruction templates](agent-profile.md#instruction-templates).

### `OQ-PRF-8`

Should tools stay a separate field? Tools have their own field. Harness permissions and capabilities have separate fields. Deriving the tools from harness permissions would remove that field. See [Profile fields](agent-profile.md#profile-fields).

### `OQ-PRF-9`

Who defines the set of capabilities? Names such as `pull-requests.read` are illustrative, and the API server publishes the full set. Who writes that list and its names is still open. One option is the API server's own code, with one capability for each operation. A deployment configuration that groups operations into capabilities is another. See [Capabilities](agent-profile.md#capabilities).

### `OQ-PRF-10`

Should launch validation stop at the first failure? Validation checks in order and returns the first failure, apart from `limit-exceeds-profile` and `instruction-inputs-invalid`, which list every item. The alternative reports every failure at once. See [Launch validation](agent-profile.md#launch-validation).

### `OQ-PRF-11`

Should placement also filter on what the agent host can enforce? Placement offers an execution only to agent hosts that can resolve its credential references, while `harness-permission-unsupported` and `sandbox-unsupported` remain launch errors. The alternative also filters on harness permissions and sandbox settings. See [Launch validation](agent-profile.md#launch-validation).

### `OQ-PRF-12`

When may the catalog delete a retired version? Deletion waits until cleanup has removed every execution that pinned the version, which long retention can delay. A fixed period after retirement or no deletion would be alternatives. See [Profile changes](agent-profile.md#profile-changes).

### `OQ-PRF-13`

Should a specification have to state the harness, model provider, model, and model effort? Omitted values come from an entry in the profile's models that matches the rest. Requiring all four in every specification is the alternative. See [Launch validation](agent-profile.md#launch-validation).

### `OQ-PRF-14`

What are the canonical tool names? In the example, `read`, `search`, and `shell` each map to the harness's own name. A fixed list in the contract could define them, or each harness adapter could publish its names with its mapping. See [Profile fields](agent-profile.md#profile-fields).

### `OQ-PRF-15`

What are the effort levels? The example uses `medium` and `high`, and the provider maps each one to the harness's own form. A fixed scale in the contract could define them, or each harness's own values could pass through unchanged. See [Repository profile files](agent-profile.md#repository-profile-files).

### `OQ-PRF-16`

What's an entry's capacity, and how does afakt measure it? Selection prefers the lowest `priority` number with capacity for the launch, but capacity has no definition. The model provider's rate limits and quota are one source. A limit on concurrent executions per entry is another. Can two entries share a priority? If they can, afakt needs a rule to order them. See [Profile fields](agent-profile.md#profile-fields).

### `OQ-PRF-17`

What role does `default: true` play? A launch that names none of the four fields matches every entry. The default entry could decide that launch, or priority and capacity could decide it as they decide any other. See [Profile fields](agent-profile.md#profile-fields).

## Agent provider requirements

### `OQ-PRQ-1`

Which error fits a target path that breaks the path rules? `VAL-5` and `VAL-6` use `invalid-specification`, and a dedicated error such as `workspace-source-invalid` would let callers tell the cases apart. See [6. Launch validation](agent-provider-requirements.md#6-launch-validation).

### `OQ-PRQ-2`

Does the provider choose among matching model entries, or does the agent executor? `VAL-40` leaves it to the provider as a SHOULD. Model capacity has no definition. See [6.1. Filling omitted fields](agent-provider-requirements.md#61-filling-omitted-fields).

### `OQ-PRQ-3`

Do the credential broker, the telemetry endpoint, and hook dispatch in the afakt CLI need requirements pages of their own? They interact with the execution provider. See [Provider requirements](agent-provider-requirements.md).

### `OQ-PRQ-4`

Should `OBS-4`, `OPS-12`, `PRV-9`, and the other items that point to a table split into one item for each row? Pointing to the table keeps both pages equal, at the cost of one item covering many checks. See [Provider requirements](agent-provider-requirements.md).

### `OQ-PRQ-5`

May a provider emit spans and metric attributes beyond the ones [Agent observability](agent-observability.md) lists? `EVD-3` allows extra evidence. See [21. Telemetry](agent-provider-requirements.md#21-telemetry).

### `OQ-PRQ-6`

Which limits apply to ACP notification size and rate? Bounded defaults set by deployment policy are the leading option, but must permit legitimate large updates without allowing a guest to exhaust the session hub. See [9. Harness transport](agent-provider-requirements.md#9-harness-transport).

## Agent session

### `OQ-SES-1`

Does a suspend and resume cycle start a new execution generation, or a new unit inside one generation? A fresh authenticated run binding with an old-binding fence is the leading option for preventing stale channels from controlling a resumed agent, at the cost of reconnecting clients and helpers. It leaves the durable agent and conversation intact, but doesn't distinguish between a new execution and a new generation. The mapping remains open. See [Suspend and resume](agent-session.md#suspend-and-resume).

### `OQ-SES-2`

Should the hooks get `suspend` and `resume` events of their own, beyond `before-agent` and `before-teardown`? See [Suspend and resume](agent-session.md#suspend-and-resume).

### `OQ-SES-3`

How long do the idle timeout and the lifetime limit last by default? See [Limits](agent-session.md#limits).

### `OQ-SES-4`

Which transport carries the agent host's authenticated outbound connection to the API server? Mutually authenticated HTTP/3 is the leading carrier where User Datagram Protocol (UDP) traffic is reachable, with mutually authenticated HTTP/2 over TCP as the required fallback. HTTP/3 reduces interference between streams under packet loss, while HTTP/2 needs bounded scheduling and both require durable ownership and replay. Those requirements add recovery coordination. Fallback timing, deployed routing, and recovery still need qualification. See [ACP gate](agent-session.md#acp-gate).

### `OQ-SES-5`

Which carrier joins the guest relay to the mediator? For separating session traffic from the data network, vsock is the leading option, at the cost of integration specific to each backend. Protecting it requires enforced process access and authenticated control endpoints for helpers and restored channels as well. Device ownership alone is insufficient, and those boundaries still need qualification for each backend. See [ACP gate](agent-session.md#acp-gate).

### `OQ-SES-6`

How does token-free guest control preserve the sender claims used by [message signing](agent-session-attachment.md#message-signatures)? Deriving the envelope from mediator-authenticated context is the leading option, but the API server must bind the sending user, agent, session, and execution without trusting guest-supplied claims. Token custody stays outside the guest. A channel-bound sender context plus durable message IDs is the leading binding rule. It rejects identity substitutions and replay without guest signing keys, but needs shared replay state and a separate rule for verifying retained signatures after signing authority changes. See [ACP gate](agent-session.md#acp-gate).

## Agent session attachment

### `OQ-ATT-1`

Should a session allow [more than one writer](agent-session-attachment.md#writer-and-readers), with prompts from each one queued under its author?

### `OQ-ATT-2`

How do API server replicas route a client to the replica that holds a session's [mediated connection](agent-session-attachment.md#the-protocol-endpoint)?

### `OQ-ATT-3`

What's the default limit on the hop count for [agent messages](agent-session-attachment.md#agent-messages)?

### `OQ-ATT-4`

How does the [signed sender envelope](agent-session-attachment.md#message-signatures) bind mediator-authenticated identity to the sending user when the guest doesn't have a token? API-side derivation is the leading option, but needs an authenticated claim binding. See the [session signing question](agent-open-questions.md#agent-session).

## Agent sandbox

### `OQ-SBX-1`

Should a specification be able to widen the profile's allow lists? A specification loosens them only within the profile's own allow entries, though it could instead add allow entries beyond the profile's. See [Combining configurations](agent-sandbox.md#combining-configurations).

### `OQ-SBX-2`

Must an execution's credential references stay within those the profile permits? Both the profile and execution carry references. Restricting execution references to those the profile permits is the leading option for limiting credential selection, at the cost of requiring a profile change to select another credential. This bound remains on hold as part of repository profile self-grant and credential domain bounds. See [Credential brokering](agent-sandbox.md#credential-brokering).

### `OQ-SBX-3`

Is the `${name}` form right for sandbox variables? Sandbox rules, capability scopes, inject domains, hooks, and environment values all use it. Go template syntax, like the instructions, would be the alternative. See [Sandbox variables](agent-sandbox.md#sandbox-variables).

### `OQ-SBX-4`

Should the API server issue the execution token at launch? It does under these rules, and the broker renews it. The provider could issue the token itself, or the broker could request a short token for each call. See [Execution token](agent-sandbox.md#execution-token).

### `OQ-SBX-5`

Is "execution token" the right name for the token that the afakt API accepts from an execution? The alternatives are agent token and execution credential. See [Execution token](agent-sandbox.md#execution-token).

### `OQ-SBX-6`

What format do credential references and cache names take? Neither the `service` value nor the `<name>` in `${cache:<name>}` has one here. Each could follow the profile name rule or take free text from the agent host's configuration. See [Sandbox variables](agent-sandbox.md#sandbox-variables).

### `OQ-SBX-7`

Which fields does an `oauth` entry contain, and how does it map a user to a stored login credential? afakt's credential store is the only refresh owner, independent of the operator's own CLI sign-in, and doesn't share a refresh lock with that CLI. A reference to a stored login credential is the leading option for keeping refresh state central, at the cost of a separate authorization lookup. Whether inject rules and sign-in scopes belong in the entry or in the stored credential, and how authenticated users map to credentials, remain open. See [Credential brokering](agent-sandbox.md#credential-brokering).

### `OQ-SBX-8`

Does a deployment require sandbox VMs everywhere? If it allows host sandboxes, where? Sandbox VMs provide a separate guest kernel. On hosts without virtualization, host sandboxes provide weaker isolation. This operator decision remains on hold. See [Enforcement](agent-sandbox.md#enforcement).

### `OQ-SBX-9`

Should macOS use Apple's Containerization framework or drive Virtualization.framework directly? Direct Virtualization.framework is the leading owner choice because afakt controls guest initialization and lifecycle, at the cost of maintaining its own guest initialization. Containerization supplies a maintained init but still requires afakt to own checkpoint coordination. Runtime save failures prevent memory checkpoints despite successful support validation. Memory checkpoints require successful runtime save and restore. Cold recovery needs a separately qualified checkpoint contract. See [Enforcement](agent-sandbox.md#enforcement) and the [checkpoint questions](agent-open-questions.md#agents).

### `OQ-SBX-10`

Does privileged setup count against a rootless runtime? Separating privileged setup from rootless runtime is the leading definition, with setup rights specified for each supported platform. It keeps runtime authority narrow but still requires administrative setup where the platform demands it. See [Enforcement](agent-sandbox.md#enforcement).

### `OQ-SBX-11`

How should the user service keep running after logout and start at boot? User lingering is a candidate for keeping service ownership with the user, but continued operation after logout and boot activation remain unqualified. A system supervisor adds administrative ownership. The persistence policy remains open. See [Enforcement](agent-sandbox.md#enforcement).

### `OQ-SBX-12`

Should the mediator run as a confined process or in its own VM? A confined process is the leading option for density and throughput if the deployment permits that isolation boundary. Running the mediator in a VM adds kernel separation at a cost of tens to about a hundred mebibytes per agent and lower throughput. The host isolation baseline remains on hold. Whether a deployment requires sandbox VMs everywhere or permits host sandboxes remains undecided, as does where it permits them. See [Per-agent mediator](agent-sandbox.md#per-agent-mediator).

### `OQ-SBX-13`

How should afakt protect host-relayed mediator connections? Authenticated Transport Layer Security (TLS) streams are the leading option for peer authentication and traffic integrity, at the cost of cryptographic processing and key management. Using plain frames remains an option only under an explicitly accepted trust model. Authenticating the link retains value even with host access to guest memory. Protecting that link leaves the on-hold host isolation baseline unsettled. See [Per-agent mediator](agent-sandbox.md#per-agent-mediator).

### `OQ-SBX-14`

How should the mediator handle overload without starving control? Bounded admission is a candidate for preserving control responsiveness, at the cost of rejecting useful streams. Complete streaming service, durable decision logging, and hard resource bounds remain unresolved. Admission limits and separate control capacity don't establish those guarantees. See [Per-agent mediator](agent-sandbox.md#per-agent-mediator).

### `OQ-SBX-15`

How does mediator-stamped identity fit [message signatures](agent-session-attachment.md#message-signatures)? Keeping the token at the mediator and deriving the signed envelope at the API server is the leading option, but the binding of user, agent, session, and execution claims needs a shared rule. The guest can't supply a token as proof of authorship. See the [session question](agent-open-questions.md#agent-session). Binding the relay channel to the agent, execution, generation, and session before accepting messages narrows the sender rule. Signing at the API server can then use authenticated user context and durable message IDs, at the cost of shared signing and replay state across API replicas. See [Control requests](agent-sandbox.md#control-requests).

### `OQ-SBX-16`

How do host sandboxes keep agents' loopback listeners apart? A private network namespace with an enforceable loopback bind rule is the leading option on Linux, but namespace separation alone doesn't forbid wildcard listeners. Seatbelt on macOS has no private network namespace, and assigned port ranges don't create one. Rejecting a launch that requires an unavailable private network view is the leading response, at the cost of excluding that configuration. This narrows the enforcement choices without deciding where to permit host sandboxes. See [Enforcement](agent-sandbox.md#enforcement).

## Agent MCP gateway

### `OQ-MCP-1`

Which upstream servers use per-user OAuth identity or a shared identity, and how does a session's user map to that identity? An explicit mapping from authenticated users to upstream authorizations is the leading option, at the cost of authorization records and refresh coordination. Shared hosting also requires a qualified process boundary that prevents unauthorized workspace reads and direct egress, and either a trusted server or isolation between authorization identities. Response correlation alone can't prevent content mixing. The permitted sharing policy remains open. See [Upstream credentials](agent-mcp-gateway.md#upstream-credentials).

### `OQ-MCP-2`

Does the gateway pass through [MCP resources and prompts](agent-mcp-gateway.md#server-and-tool-access), or only tools?

### `OQ-MCP-3`

Do results from [untrusted servers](agent-mcp-gateway.md#server-and-tool-access) need content filtering against prompt injection before reaching the agent?

### `OQ-MCP-4`

What limits and rate limiting apply per execution and per upstream server? Independent per-execution and per-upstream admission limits with bounded pending calls and response waits are the leading option for bounding load. They can reject useful work. Preventing monopolization also requires an explicit allocation or fairness rule. Limits, fairness, and coordination across gateway replicas remain open. See [Server hosting and lifecycle](agent-mcp-gateway.md#server-hosting-and-lifecycle).

### `OQ-MCP-5`

Should harness-native tools, such as file editing and shell commands, ever [route through the gateway](agent-mcp-gateway.md#endpoint-and-harness-integration) for audit?

## Agent ports and previews

### `OQ-PORT-1`

Should port discovery and exposure support User Datagram Protocol (UDP)? Keeping publication TCP-only preserves the stream contract but excludes UDP services. Supporting UDP requires a separately authorized datagram contract. The scope remains open. See [Port discovery and records](agent-ports.md#port-discovery-and-records).

### `OQ-PORT-2`

Can share links admit public recipients without afakt sign-in at all? See [Access and share links](agent-ports.md#access-and-share-links).

### `OQ-PORT-3`

What per-profile limits apply to how many ports and previews an execution may expose? See [Port discovery and records](agent-ports.md#port-discovery-and-records).

### `OQ-PORT-4`

How does afakt certify preview host names beneath a wildcard domain behind a public tunnel? Wildcard coverage is the leading candidate where the provider supports the required host names and certificates, while certificates for each host add issuance and renewal work. Funnel's assigned names require a different arrangement with enrollment and quota costs, potentially using separate node identities. Public issuance, renewal, and the embedded Funnel path remain unqualified. See [Public tunnel providers](agent-ports.md#public-tunnel-providers).

### `OQ-PORT-5`

Do tunnel providers beyond the two built-ins become plug-ins? See [Public tunnel providers](agent-ports.md#public-tunnel-providers).

### `OQ-PORT-6`

Should a preview request wake a suspended agent or fail? A failed request avoids implicit resource use, while waking offers convenience but requires authorization, a bounded restore wait, and verified binding to the intended listener. Rechecking the listener alone doesn't establish that binding. The wake policy and exact listener-binding mechanism remain open. See [Sessions and clients](agent-ports.md#sessions-and-clients).

### `OQ-PORT-7`

Should bulk transfers use a connection separate from interactive preview streams? Separate connections are a candidate for reducing bulk-transfer interference, at the cost of extra ownership and scheduling. Reserved control capacity and bounded data admission reduce queue contention but don't establish a control-latency guarantee or remove shared TCP loss stalls and processor contention. Connection policy and behavior under saturation remain open. See [HTTP previews](agent-ports.md#http-previews).

## Command-line reference

### `OQ-CLI-1`

Which content sources can an index register and where does indexed content live? What query interface does an agent use and how does indexing relate to memory queries and artifacts? See [Index](cli.md#index).

### `OQ-CLI-2`

What scope and output format does `report` select, and how does it combine work and usage across organizations and projects? See [Deployment and client commands](cli.md#deployment-and-client-commands).

### `OQ-CLI-3`

How do configuration scopes and precedence interact with deployment policy, agent profile bindings, and per-agent settings? See [Configuration and indexing](cli.md#configuration-and-indexing).
