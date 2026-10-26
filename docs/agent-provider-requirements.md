---
type: Specification
title: Agent provider requirements
description: The requirements for an execution provider, in BCP 14 key words, from the provider operations and execution identity to evidence, cleanup, and telemetry.
tags: [afakt, agents, execution, requirements]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent provider requirements

## 1. Introduction

Each harness run belongs to one durable [agent](agent.md). After the run ends, the agent retains its conversation transcript. The representation of ownership and its mapping to execution generations remain [open](agent-open-questions.md#agents).

An execution provider launches, observes, stops, and reports agent executions on an agent host for the agent executor. These requirements define the provider contract. They apply to every provider, the deterministic test provider included. Transport requirements also bind the ACP relay and the agent executor workflow where named. Capacity requirements bind the afakt worker and the reservation API where named. Image requirements also bind the control plane, registry, and worker where named. Port and preview requirements bind the guest relay, mediator, preview broker, control plane, clients, and tunnel providers where named. Gateway requirements bind the provider, harness integration, Model Context Protocol (MCP) gateway, session hub, and API server where named. The API server, mediator, credential broker, telemetry endpoint, and afakt CLI are the parties the provider deals with.

## 2. Conventions and terminology

The key words `MUST`, `MUST NOT`, `REQUIRED`, `SHALL`, `SHALL NOT`, `SHOULD`, `SHOULD NOT`, `RECOMMENDED`, `NOT RECOMMENDED`, `MAY`, and `OPTIONAL` have the meanings that BCP 14 ([RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) and [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174)) gives them when, and only when, they appear in all capitals, in this specification.

Each numbered item has an ID made of its section's prefix and a number, such as `OPS-1`, and tests and reviews cite items by these IDs.

"Provider" means the execution provider. "Caller" means the party that sends a provider operation, which is the agent executor's activity. The [glossary](glossary.md) defines the other terms, such as execution, generation, specification, profile, workspace, sandbox, evidence, and provenance record.

## 3. Provider obligations

A provider conforms when it meets each MUST and MUST NOT in this specification. Departing from a SHOULD takes a reason, recorded in the provider's own documentation. Items about an OPTIONAL feature bind only the providers that have it.

## 4. Provider operations

See [Provider operations](agent-execution-contract.md#provider-operations).

- **`OPS-1`** The provider MUST offer the launch, status, cancel, and result operations.
- **`OPS-2`** The provider MUST identify the execution in each operation by its execution ID and generation.
- **`OPS-3`** When a status request omits the generation, the provider MUST report the current generation.
- **`OPS-4`** When a status request specifies a generation lower than the current generation, the provider MUST report that generation's state.
- **`OPS-5`** When a caller repeats an operation with the same inputs, the provider MUST apply the operation's effect at most once.
- **`OPS-6`** The provider MUST reject a request that gives an execution ID in any form other than the canonical text form of [Execution ID format](agent-execution-model.md#execution-id-format) with `invalid-specification`.
- **`OPS-7`** The provider MUST reject a malformed request with `invalid-specification`.
- **`OPS-8`** The provider MUST return `not-found` only for an execution ID, or a generation of one, that it has no record of.
- **`OPS-9`** When the provider can't read the execution record of a requested execution, it MUST return `unavailable`, and it MUST NOT return `not-found`.
- **`OPS-10`** The provider MUST run each operation as the body of an activity on the placed agent host's task queue.
- **`OPS-11`** The provider MUST return each error under its own name from [Provider operations](agent-execution-contract.md#provider-operations) or [Launch validation](agent-profile.md#launch-validation).

### 4.1. Status

- **`OPS-12`** For a recorded execution, a status response MUST hold the generation, the current generation, the state, the phase, the last observed time, any cancellation acknowledgement, the evidence references published so far, and the provenance record.

### 4.2. Cancel

- **`OPS-13`** For a non-terminal execution, the provider MUST return a cancellation acknowledgement with the time it accepted the request and the state at that time.
- **`OPS-14`** For a non-terminal execution, the provider MUST record the cancellation acknowledgement with the request's reason.
- **`OPS-15`** For an execution that already has a cancellation acknowledgement, the provider MUST return the recorded acknowledgement.
- **`OPS-16`** For a terminal execution, the provider MUST return an acknowledgement with the terminal state, and it MUST NOT change the execution.

### 4.3. Result

- **`OPS-17`** For a terminal execution whose result remains, the provider MUST return its execution result.
- **`OPS-18`** For a non-terminal execution, the provider MUST answer with `not-terminal`.
- **`OPS-19`** For an execution whose result and evidence cleanup removed, the provider MUST answer with `expired`.

## 5. Execution identity

See [Execution identity](agent-execution-model.md#execution-identity).

- **`ID-1`** The provider MUST compute the specification digest over the canonical encoding of every specification field except the generation.
- **`ID-2`** When a valid launch specifies an unrecorded execution ID with generation 1, the provider MUST record the execution with generation 1 as its current generation.
- **`ID-3`** The provider MUST reject a launch that specifies an unrecorded execution ID with a generation greater than 1 with `invalid-specification`.
- **`ID-4`** When a launch repeats a recorded execution ID and generation with the same specification digest, the provider MUST return the existing execution in its current state, and it MUST NOT check the launch again.
- **`ID-5`** The provider MUST reject a launch that repeats a recorded execution ID and generation with a different specification digest with `specification-conflict`.
- **`ID-6`** The provider MUST reject a launch that specifies a recorded execution ID with a generation lower than its current generation with `stale-generation`.
- **`ID-7`** When a valid launch specifies a recorded execution ID with a generation higher than its current generation, the provider MUST make that generation the current generation.
- **`ID-8`** The provider MUST NOT run more than one agent for an execution ID at any time.
- **`ID-9`** The provider MUST NOT change a recorded specification or any value it derived from the specification at launch.
- **`ID-10`** The provider MUST keep each accepted generation in the execution record until cleanup.

## 6. Launch validation

See [Launch validation](agent-profile.md#launch-validation) and [Overrides](agent-profile.md#overrides).

- **`VAL-1`** The provider MUST check a new launch's whole specification against its pinned profile version before it records the execution.
- **`VAL-2`** The provider MUST check the profile validation errors in the order that [Launch validation](agent-profile.md#launch-validation) lists them, and it MUST return the error of the first failed check.
- **`VAL-3`** The provider MUST NOT record an execution whose launch fails validation.
- **`VAL-4`** The provider MUST reject a time limit of zero with `invalid-specification`.
- **`VAL-5`** The provider MUST reject a target path of a workspace source with an empty, `.`, or `..` segment, other than the target path `.` alone, with `invalid-specification`.
- **`VAL-6`** The provider MUST reject two workspace sources with the same target path, or with one target path inside another, with `invalid-specification`.
- **`VAL-7`** The provider MUST reject a launch that pins a profile name its source lacks with `profile-not-found`.
- **`VAL-8`** The provider MUST reject a launch that pins a catalog version number or a repository commit that doesn't exist with `profile-version-not-found`.
- **`VAL-9`** The provider MUST reject a launch whose pinned profile version's digest differs from the specification's profile digest with `profile-digest-mismatch`.
- **`VAL-10`** The provider MUST reject a pinned repository profile file that breaks a rule of [Repository profile files](agent-profile.md#repository-profile-files) with a `profile-invalid` error that identifies the file and its first problem.
- **`VAL-11`** The provider MUST reject a new execution that pins a catalog version with status `retired` with `profile-retired`.
- **`VAL-12`** The provider MUST reject a launch whose harness name and version, model provider, model, and model effort match no entry in the profile's models with `model-not-allowed`.
- **`VAL-13`** The provider MUST reject a launch that requests wider workspace access than the profile grants with `workspace-access-not-allowed`.
- **`VAL-14`** The provider MUST reject a launch that sets the instructions, the input schema, the harness permissions, the tools, the harness transport, or the telemetry content with `override-not-allowed`.
- **`VAL-15`** The provider MUST reject a launch that sets an output schema for a profile with the output schema policy `fixed` with `override-not-allowed`.
- **`VAL-16`** The provider MUST reject a launch that omits the output schema for a profile with the output schema policy `per-execution` with `output-schema-missing`.
- **`VAL-17`** The provider MUST reject a launch that requests a limit or a retention over the profile's ceiling with `limit-exceeds-profile`, listing each value over its ceiling.
- **`VAL-18`** The provider MUST reject instruction inputs that fail validation against the profile's input schema with `instruction-inputs-invalid`, listing each failing input path.
- **`VAL-19`** The provider MUST reject a launch whose instruction template fails to render with valid instruction inputs with `instructions-render-failed`.
- **`VAL-20`** The provider MUST reject a capability that the API server doesn't publish, or a scope that the capability doesn't take, with `capability-unknown`.
- **`VAL-21`** The provider MUST reject a capability that the profile lacks, or a scope outside the profile's scope, with `capability-not-allowed`.
- **`VAL-22`** The provider MUST reject a sandbox configuration that sets an unknown setting, a setting of the agent host, or a setting that only the profile sets with `sandbox-invalid`.
- **`VAL-23`** The provider MUST reject an environment entry with `environment-invalid`, naming the entry, when its name breaks the name rules or repeats within one list, its value goes over the size limit, or a specification sets `locked` on it.
- **`VAL-24`** The provider MUST reject an environment entry in the profile, the specification, or a hook that uses a [reserved name](agent-workspace.md#reserved-names) with `environment-name-reserved`, naming the entry.
- **`VAL-25`** The provider MUST reject a specification entry or a hook's entry that sets the name of a locked profile entry with an `environment-locked` error that identifies the entry and the locked name.
- **`VAL-26`** The provider MUST reject a hook that breaks a rule of [Hook fields](agent-hooks.md#hook-fields) with a `hook-invalid` error that identifies the hook and the broken rule.
- **`VAL-27`** The provider MUST reject a sandbox rule, capability scope, inject domain, environment value, or hook that uses a variable [Sandbox variables](agent-sandbox.md#sandbox-variables) doesn't define, or a target path without a matching workspace source, with `sandbox-variable-unknown`.
- **`VAL-28`** The provider MUST reject a specification allow entry outside the profile's allow list with `sandbox-not-allowed`.
- **`VAL-29`** The provider MUST reject a specification that turns on an isolation relaxation the profile leaves off with `sandbox-not-allowed`.
- **`VAL-30`** The provider MUST reject a credential brokering inject domain that falls outside the combined `allowedDomains` or appears in `tlsTerminate.excludeDomains` with `credential-domain-not-allowed`.
- **`VAL-31`** The provider MUST reject a launch whose profile has a harness permission that the harness on the agent host can't enforce with `harness-permission-unsupported`.
- **`VAL-32`** The provider MUST reject a launch whose combined sandbox configuration has a setting it can't enforce on the agent host with `sandbox-unsupported`, naming the setting.
- **`VAL-33`** The provider MUST reject a launch whose profile or execution specification has a credential reference the agent host can't resolve with `credential-unavailable`, naming the reference.
- **`VAL-34`** The provider MUST reject an `execution` source that references an execution it doesn't keep, one that hasn't ended, or one whose workspace cleanup removed with `workspace-source-unavailable`.
- **`VAL-35`** For validation, the provider MUST treat an earlier generation that the launch supersedes as ended when an `execution` source references it.
- **`VAL-36`** The provider MUST reject a harness, model provider, model, or model effort that it can't run with `unsupported`.
- **`VAL-37`** The provider MUST reject a time limit or resource limit that it can't enforce with `unsupported`.
- **`VAL-38`** An error about a credential reference MUST include the reference alone, and it MUST NOT include the secret value.

### 6.1. Filling omitted fields

- **`VAL-39`** When a launch omits any of the harness, model provider, model, or model effort, the provider MUST take each omitted value from one matching entry in the profile's models.
- **`VAL-40`** Among the matching models entries, the provider SHOULD choose the entry with the lowest `priority` number that has capacity for the launch.
- **`VAL-41`** When a launch omits a time limit, stop grace period, or resource limit, the provider MUST apply the profile's default for it.
- **`VAL-42`** When a launch omits the workspace access, the provider MUST apply the profile's workspace access.
- **`VAL-43`** When a launch omits the capabilities, the provider MUST apply the profile's capabilities.
- **`VAL-44`** When a launch omits the retention, the provider MUST apply the deployment's default retention.

### 6.2. Resource form

See [Agent execution resource](agent-execution-resource.md).

- **`VAL-45`** The provider MUST reject a launch resource whose `apiVersion` isn't `afakt/v1alpha1` or whose `kind` isn't `AgentExecution` with `invalid-specification`.
- **`VAL-46`** The provider MUST reject a launch resource that sets `status`, or a key that [Agent execution resource](agent-execution-resource.md) doesn't define, with `invalid-specification`.
- **`VAL-47`** Status and result MUST return the generation in the resource form, with `status` filled in.

## 7. Launch

See [Lifecycle](agent-execution.md#lifecycle) and [Pinning a profile version](agent-profile.md#pinning-a-profile-version).

- **`LCH-1`** The provider MUST run each generation's lifecycle steps in the order that [Lifecycle](agent-execution.md#lifecycle) gives.
- **`LCH-2`** The provider MUST record a validated launch as `pending` before it starts any process for it.
- **`LCH-3`** The provider MUST store a copy of the pinned profile version's content, with its referenced files, in the execution record.
- **`LCH-4`** The provider MUST serve status, cancel, result, and reconciliation for a generation from its stored profile copy.
- **`LCH-5`** The provider MUST claim the slot reservation for an execution ID and generation through the [reservation API](agent-placement.md#reservation-api) before it records the execution.
- **`LCH-6`** When the reservation API doesn't have a reservation for a launch's execution ID and generation, the provider MUST request a slot on its own agent host.
- **`LCH-7`** The provider MUST keep a recorded execution in `pending` until it has a slot.
- **`LCH-8`** The provider MUST free an execution's slot through the reservation API when the execution enters a terminal state.
- **`LCH-9`** For an accepted launch, the provider MUST return the execution ID, the generation, the specification digest, and the state after launch.
- **`LCH-10`** When an attempt to start the agent returns no answer the provider can trust, the provider MUST return `launch-unknown` from launch.
- **`LCH-11`** The provider MUST compute the configuration digest over the inputs that [Execution generation](agent-execution-model.md#execution-generation) lists, and it MUST record the digest in the generation and in the provenance record.
- **`LCH-12`** The provider MUST get the execution's [execution token](agent-sandbox.md#execution-token) from the API server when it configures the sandbox and MUST keep it at the mediator or agent host outside the guest.
- **`LCH-13`** When the `before-agent` hooks end without a failure, the provider MUST start the harness through the selected transport with the rendered instructions and the prompt.
- **`LCH-14`** The provider MUST register its agent host with the [agent host registry](agent-placement.md#agent-host-registry) before it accepts work, and it MUST renew the lease with its current host metadata, slot capacity, most recent capacity change reason, and the execution IDs and generations that hold its slots.

## 8. Slot capacity

See [Slot capacity](agent-placement.md#slot-capacity) and [Slot reservations](agent-placement.md#slot-reservations).

- **`CAP-1`** The afakt worker MUST read the lower and upper bounds on its execution slot count and its resource targets from the agent host's configuration.
- **`CAP-2`** The worker MUST keep slot capacity between the configured lower and upper bounds, and it MUST keep capacity equal to the upper bound when the configuration doesn't set resource targets.
- **`CAP-3`** With resource targets set, the worker's capacity controller MUST measure host CPU use, memory use, and free disk space on the volume that holds workspaces at intervals and compare each measurement with its configured target.
- **`CAP-4`** The controller MUST raise capacity only when every configured target permits more work, and it MUST NOT raise capacity by more than one slot at a time.
- **`CAP-5`** When CPU or memory use exceeds its target, or free workspace disk space is less than its target, the controller MUST lower capacity at once within the configured bounds, unless capacity is already at the lower bound.
- **`CAP-6`** The worker MUST report its current slot capacity and the reason for its most recent change in every lease renewal, using the reasons that [Slot capacity](agent-placement.md#slot-capacity) defines.
- **`CAP-7`** The worker MUST register the host's lower and upper bounds and current capacity, and the reservation API MUST keep those values and the most recent capacity change reason in the agent host registry.
- **`CAP-8`** The reservation API MUST count reserved and claimed slots together and MUST NOT reserve a new slot that would make their total exceed the most recently reported capacity, including when concurrent requests arrive.
- **`CAP-9`** While reserved and claimed slots total at least the reported capacity, the reservation API MUST NOT reserve a new slot on that host.
- **`CAP-10`** A capacity decrease MUST NOT cause the worker to stop or preempt a running execution, and it MUST NOT cause the reservation API to cancel an existing reservation or reject its claim.
- **`CAP-11`** Placement through the reservation API MUST match against free slots only, under the slot accounting of `CAP-8` and `CAP-9`.
- **`CAP-12`** The worker MUST publish the slot capacity, reserved slots, claimed slots in use, and capacity change metrics with the instruments, units, and attributes that [Metrics](agent-observability.md#metrics) lists.

## 9. Harness transport

See [Harness transport](agent-profile.md#harness-transport), [Executions over ACP](agent-session.md#executions-over-acp), and [Execution attachment](agent-session-attachment.md#execution-attachment).

- **`TRN-1`** The provider MUST use the transport of the selected entry in the profile's models, and it MUST NOT accept a transport override from the execution specification.
- **`TRN-2`** For an entry without `transport`, the provider MUST use `acp` when the harness has an ACP adapter and `headless` otherwise.
- **`TRN-3`** The provider MUST use `claude-agent-acp` for Claude Code over ACP and `codex-acp` for Codex over ACP.
- **`TRN-4`** The provider MUST support an entry that explicitly selects `headless` for a harness with an ACP adapter.
- **`TRN-5`** The provider MUST reject a launch that selects `acp` for a harness without an ACP adapter with `harness-transport-unsupported` before recording the execution.
- **`TRN-6`** In sandbox VM mode, the provider MUST start the guest relay inside the VM under its own user, and the relay MUST start the adapter under the separate agent user over standard input and output.
- **`TRN-7`** The relay MUST initialize the adapter and open an ACP session whose only writer is the agent executor workflow.
- **`TRN-8`** The agent executor workflow MUST send the rendered instructions and the prompt together as one prompt through the relay and wait for the turn to end.
- **`TRN-9`** The relay MUST answer each `session/request_permission` by allowing only what the profile's harness permissions grant and rejecting the rest.
- **`TRN-10`** The relay MUST cancel every `elicitation/create`, and it MUST NOT pass permission or elicitation requests to a person for an answer in an execution.
- **`TRN-11`** After the turn ends, the relay MUST close the ACP session, and the provider MUST verify termination of every process it started, including the relay, adapter, harness, and helpers, before treating the agent as stopped.
- **`TRN-12`** A running ACP execution MUST allow people with read access to attach as readers, and it MUST NOT allow an attached client to become its writer or change the session.
- **`TRN-13`** A headless execution MUST NOT offer attachment.
- **`TRN-14`** The provider MUST apply the same output schema validation, hooks, results, and evidence requirements to both transports.
- **`TRN-15`** The relay MUST advertise no ACP file system or terminal capabilities and MUST keep permission answers from profile grants, reconnect with backoff, and buffer adapter messages during disconnection.
- **`TRN-16`** The relay's ACP channel MUST end at the per-agent mediator. Its ACP gate MUST apply Cedar policy to every `JSON-RPC` message in both directions with default deny.
- **`TRN-17`** The ACP gate MUST use the sending user as principal for messages toward the agent, including prompts and cancel, and MUST use the agent as principal for requests from it, including file reads and writes, terminal creation, and permission requests.
- **`TRN-18`** The ACP gate MUST bound update notification size and rate. For a denied message, it MUST return a `JSON-RPC` error and record the decision in the agent's decision log.
- **`TRN-19`** The mediator MUST forward permitted session messages to the session hub over the agent host's authenticated outbound connection to the API server. ACP MUST NOT bypass mediator policy or network rules.
- **`TRN-20`** On reconnect, buffered ACP messages MUST pass the gate before delivery. The relay MUST submit permission and elicitation requests and local answers to the gate even when it handles them locally, and MUST NOT override a gate denial.

## 10. Workspace

See [Workspace sources](agent-workspace.md#workspace-sources).

- **`WS-1`** The provider MUST build a generation's workspace from the specification's workspace sources.
- **`WS-2`** The provider MUST NOT build a generation's workspace while an earlier generation of the same execution ID is non-terminal.
- **`WS-3`** The provider MUST place each workspace source at its target path relative to the workspace root.
- **`WS-4`** The provider MUST check out a `repository` source at the source's full commit ID.
- **`WS-5`** When a `repository` source has a new branch name, the provider MUST create that branch at the checked-out commit and switch to it.
- **`WS-6`** When a `repository` source has no new branch name, the provider MUST leave the checkout at the commit without a branch.
- **`WS-7`** The provider MUST fetch each repository outside the agent's sandbox.
- **`WS-8`** When a repository fetch needs credentials, the provider MUST use the fetch credential that the agent host's configuration gives for that repository host.
- **`WS-9`** The provider MUST place an independent copy of the workspace that an `execution` source's execution left when it ended.
- **`WS-10`** The provider MUST place an empty directory for an `empty` source.
- **`WS-11`** When the workspace build fails, the provider MUST end the execution `failed` with `workspace-failure` as its reason.
- **`WS-12`** When the workspace build ends, the provider MUST apply the execution's workspace access to the whole workspace.
- **`WS-13`** The provider MUST create each execution's scratch directory outside its workspace.

### 10.1. Image sources

See [Image sources and digest pinning](agent-workspace-image.md#image-sources-and-digest-pinning).

- **`WS-14`** The provider MUST accept an `image` source with an image repository and tag or manifest digest, and MUST apply the target path rules of `VAL-5` and `VAL-6` to it.
- **`WS-15`** The provider MUST materialize an image's tree at its source target on the guest workspace disk before hooks or the agent start, and MUST NOT use the workspace image as the agent host's root filesystem.
- **`WS-16`** During workspace construction, the provider MUST resolve an image tag to a manifest digest once and persist the digest with the execution generation before fetching layers.
- **`WS-17`** The provider MUST keep the requested image reference separate from its observed manifest digest in provenance, and MUST NOT change the specification to record resolution.
- **`WS-18`** A retried operation MUST reuse the generation's recorded image digest and MUST NOT resolve its tag again.
- **`WS-19`** For an image source with a digest, the provider MUST use that digest without substituting another image, including when a higher generation requests a digest recorded by an earlier generation.
- **`WS-20`** The provider MUST verify manifest, config, and layer content against their digests before using it, and MUST end a failed pull, verification, or unpack under `WS-11`.
- **`WS-21`** Before workspace construction uses a resolved image, the control plane MUST establish a retention reference that protects its manifest, config, and layers from registry collection.
- **`WS-22`** The control plane and registry MUST preserve that reference while any referencing execution is non-terminal, retained, or held, and MUST NOT allow collection to race its creation.
- **`WS-23`** In sandbox VM mode, the worker MUST supply a separate read-only root disk for the guest operating system and a writable ext4 workspace disk built from the workspace layers.
- **`WS-24`** A local workspace checkpoint MUST capture a block delta with the guest filesystem frozen at an idle turn boundary. An agent checkpoint MUST pair the VM snapshot, workspace disk delta, and harness session state from that same boundary.
- **`WS-25`** A portable workspace revision MUST use an Open Container Initiative (OCI) diff layer plus a content-hashed file inventory pushed to the registry. Correctness MUST depend on content hashes rather than modification times or sizes.
- **`WS-26`** Rewind MUST start a new conversation branch from a checkpoint and keep the original for audit. Suspend MUST save a checkpoint and stop the VM. Resume MUST restore the same guest paths and reacquire credentials outside the guest.

### 10.2. Registry tokens

See [Registry authentication](agent-workspace-image.md#registry-authentication).

- **`REG-1`** The registry MUST accept bearer tokens only, and MUST reject a request without a valid token.
- **`REG-2`** The registry MUST verify each token's signature using the control plane's published JSON Web Key Set (JWKS) and MUST reject an unknown signing key or invalid signature.
- **`REG-3`** The registry MUST check the trusted issuer, registry audience, and validity interval on every request, and MUST reject a token that fails any check.
- **`REG-4`** The registry MUST check the requested repository and action against token scopes for every operation, including ranged reads, cross-repository blob mounts, tag operations, and deletion, and MUST reject an operation outside those scopes.
- **`REG-5`** The registry MUST NOT treat knowledge of a digest as pull authorization, or a push grant as a deletion or administration grant.
- **`REG-6`** The control plane MUST check the worker's permission before issuing a token and MUST issue short-lived tokens scoped to the repositories and actions the worker needs. A build's push grant MUST cover only its destination image repository.
- **`REG-7`** The worker MUST keep registry tokens outside the sandbox and MUST NOT write them into images, evidence, or telemetry.
- **`REG-8`** The worker MUST get control plane authorization for an execution's image repository even when it has already cached all required content locally.

### 10.3. Layer caches

See [Agent host caches and startup](agent-workspace-image.md#agent-host-caches-and-startup) and [Retention and garbage collection](agent-workspace-image.md#retention-and-garbage-collection).

- **`CACHE-1`** The worker MUST cache verified layer blobs by digest and complete unpacked layers by their verified layer digest, and reusable entries MUST be immutable and outside the sandbox's writable workspace.
- **`CACHE-2`** The worker MUST fetch missing layers with parallel ranged reads, verify the complete blob before caching it, and discard a corrupt entry before fetching it again.
- **`CACHE-3`** The worker MUST extract independent layers in parallel while preserving the result of applying layers in manifest order for overlapping paths, whiteouts, and opaque directories.
- **`CACHE-4`** Layer extraction MUST NOT write outside the source target. This also applies to archive paths and links.
- **`CACHE-5`** The worker MUST give each agent an independent writable disk using copy-on-write clones where supported, including reflinks, APFS clones, or qcow2 overlays, and fall back to a plain copy otherwise. It MUST NOT share writable hard links with caches or other workspaces.
- **`CACHE-6`** The worker MUST ensure all required image content is present before hooks or the agent start, and MUST NOT use Filesystem in User Space (FUSE), lazy pull, or a snapshotter for workspace start. This rule MUST NOT choose whether pulling occurs before VM boot or through the guest mediator.
- **`CACHE-7`** The worker MUST evict least recently used layer blobs and unpacked entries under the free-disk resource target, MUST skip entries in active transfers or materialization and bases backing active disks or retained checkpoint deltas, and MUST NOT remove execution workspace files through cache eviction.
- **`CACHE-8`** Cache locality MUST NOT override placement requirements, host bindings, lease eligibility, or slot accounting, and preheating MUST yield to the free-disk resource target.
- **`CACHE-9`** The worker MUST cache the unpacked workspace base and disk by verified digest so agents from the same image reuse that base without sharing writes.

## 11. Environment

See [Execution environment](agent-workspace.md#execution-environment), [Agent environment](agent-workspace.md#agent-environment), and [Hook environment](agent-hooks.md#hook-environment).

- **`ENV-1`** The provider MUST set each variable of the [execution environment](agent-workspace.md#execution-environment) table for the agent and for each hook.
- **`ENV-2`** An execution environment variable that has a sandbox variable MUST hold the identical text of that sandbox variable.
- **`ENV-3`** The provider MUST write each execution environment value as plain text without a trailing newline.
- **`ENV-4`** The provider MUST set `HOME` to the value of `${home}`.
- **`ENV-5`** The provider MUST apply each profile entry whose name the specification doesn't list.
- **`ENV-6`** When a specification entry has the name of a profile entry without `locked`, the provider MUST replace that entry's value with the specification's value.
- **`ENV-7`** When a specification entry has a name that the profile lacks, the provider MUST add that variable to the agent environment.
- **`ENV-8`** The provider MUST insert each sandbox variable in an environment value literally, and it MUST read `$$` as a literal `$`.
- **`ENV-9`** The provider MUST give the agent and every hook the agent environment together with the execution environment and the trace context.
- **`ENV-10`** The provider MUST include the agent environment's canonical encoding, sorted by name, in the configuration digest.
- **`ENV-11`** The provider MUST apply a hook's own environment entries on top of the agent environment for that hook alone.
- **`ENV-12`** The provider MUST set `AFAKT_PHASE` to the hook's event for a `before-agent`, `after-agent`, or `before-teardown` hook, and to `agent` for every other hook.
- **`ENV-13`** The provider MUST set each hook variable that [Hook environment](agent-hooks.md#hook-environment) lists for the hooks it names.
- **`ENV-14`** The provider MUST NOT set `AFAKT_AGENT_EXIT_STATUS` on an ACP run or unless a headless harness reports an exit status.
- **`ENV-15`** The provider MUST give each `after-agent` hook the raw final answer in a read-only file at `AFAKT_FINAL_ANSWER_FILE`.
- **`ENV-22`** The provider MUST set `AFAKT_AGENT_STOP_REASON` to the reported ACP stop reason for `session-end`, `after-agent`, and `before-teardown` hooks, and it MUST NOT set it on headless runs or before the adapter reports a stop reason.

### 11.1. Trace context

- **`ENV-16`** The provider MUST set `TRACEPARENT` for the agent to the agent run span's trace context, and for each hook to that hook's own span's trace context.
- **`ENV-17`** When a `tracestate` accompanies the trace context, the provider MUST set `TRACESTATE` to it.
- **`ENV-18`** The provider MUST set `OTEL_EXPORTER_OTLP_ENDPOINT` to the address of `${telemetry}` as an `http://` URL.
- **`ENV-19`** The provider MUST set `OTEL_EXPORTER_OTLP_PROTOCOL` to `http/protobuf`.
- **`ENV-20`** The provider MUST set `OTEL_RESOURCE_ATTRIBUTES` to `afakt.profile.source`, `afakt.profile.name`, and `afakt.harness.name`.
- **`ENV-21`** The resource attributes the provider passes to the harness MUST NOT hold the execution ID or the generation.

## 12. Sandbox

See [Agent sandbox](agent-sandbox.md).

- **`SBX-1`** The provider MUST run the agent and every hook inside the execution's sandbox under the combined sandbox configuration.
- **`SBX-2`** The provider MUST replace each sandbox variable with its value before it combines the configurations.
- **`SBX-3`** The provider MUST insert each sandbox variable's value literally, so that a value with glob characters matches only itself.
- **`SBX-4`** The provider MUST combine the profile's and the specification's sandbox configurations under the rules of [Combining configurations](agent-sandbox.md#combining-configurations).
- **`SBX-5`** The provider MUST count an entry that it can't confirm as within another entry as outside it.
- **`SBX-6`** The provider MUST apply the precedence rules of the sandbox runtime to the combined sandbox configuration.
- **`SBX-7`** The provider MUST resolve each relative path in a filesystem rule against the workspace root.
- **`SBX-8`** The provider MUST add `${workspace}` to `allowWrite` for `read-write` workspace access, and to `denyWrite` for `read-only` workspace access.
- **`SBX-9`** The provider MUST add `${scratch}` to `allowWrite`.
- **`SBX-10`** The provider MUST add `${afakt-api}` and `${telemetry}` to `allowedDomains`.
- **`SBX-11`** The provider MUST add the mediator's per-agent certificate authority to the guest's trusted certificates for inspected domains.
- **`SBX-12`** The provider MUST give each sandbox VM its own mediator outside the guest and MUST attach its only network device to that mediator's user-space TCP/IP stack. The guest MUST NOT have another route out.
- **`SBX-13`** The provider MUST publish the combined sandbox configuration as evidence with its digest.
- **`SBX-14`** The provider MUST publish each violation that the sandbox reports as evidence.
- **`SBX-15`** The provider MUST remove an execution's sandbox when the execution enters a terminal state.
- **`SBX-16`** The provider MUST apply afakt's common sandbox capability to the entire harness process tree with the same configuration semantics across supported harnesses. Its ACP adapter when present and every hook MUST run inside that sandbox too.
- **`SBX-17`** In host sandbox mode, the provider MUST use Seatbelt profiles through `sandbox-exec` on macOS and bubblewrap namespaces plus seccomp on Linux, or Landlock with network controls that have the same effect where bubblewrap isn't available. It MUST reject a launch with `sandbox-unsupported` if it can't enforce even one setting. This mode MUST NOT settle the on-hold deployment isolation baseline.
- **`SBX-19`** In sandbox VM mode, the provider MUST give each agent its own VM as the isolation boundary. Linux MUST use a Kernel-based Virtual Machine (KVM) microVM driven by `Cloud Hypervisor`. Windows MUST use the same Linux microVM inside WSL2 by default and MUST use a Windows-native utility VM only for an agent needing a Windows guest. macOS MUST use Virtualization.framework without choosing its open init integration.
- **`SBX-20`** afakt MUST run rootless at runtime, using a systemd user unit on Linux and a launchd agent on macOS. This rule MUST NOT choose setup privileges or the mechanism for service persistence after logout.
- **`SBX-21`** The guest MUST run afakt's init with process ID 1, subject to the open macOS Containerization integration. The guest relay MUST start before hooks on ACP and headless runs and MUST run under its own user with exclusive access to the vsock device. The adapter and harness MUST run as a separate agent user with writes confined to the workspace.
- **`SBX-22`** Start, suspend, resume, stop, `kill`, and snapshot control MUST use the agent host supervisor and hypervisor (`VMM`) independently of ACP and the guest. No guest or session policy MAY block those operations.
- **`SBX-23`** The mediator MUST resolve guest DNS queries and record name-to-address bindings. Cedar MUST decide each connection using destination address, port, and resolved name, and MUST decide each HTTP request within the combined network bounds.
- **`SBX-24`** For inspected domains, the mediator MUST act as the Transport Layer Security (TLS) endpoint with its per-agent certificate authority and inject brokered credentials into permitted requests outside the guest. It MUST deny User Datagram Protocol (UDP) and Internet Control Message Protocol (ICMP) unless policy permits them.
- **`SBX-25`** The mediator MUST run as a confined process on the agent host. Whether to place it in a separate VM or encrypt its guest link MUST remain open.

## 13. Harness configuration

See [Instruction templates](agent-profile.md#instruction-templates), [Harness permissions](agent-profile.md#harness-permissions), and [Trace context](agent-workspace.md#trace-context).

- **`HRN-1`** The provider MUST check the instruction inputs against the input schema before it renders the instructions.
- **`HRN-2`** The provider MUST render the instruction template with the instruction inputs as its only data and the built-in Go template functions as its only functions.
- **`HRN-3`** The provider MUST publish the rendered instructions as evidence with their digest.
- **`HRN-4`** The provider MUST set the permission mode and tool grants for each of the profile's harness permissions in the harness's own settings, and it MUST deny each harness permission that the profile doesn't list.
- **`HRN-5`** Before it starts the agent, the provider MUST verify that the harness settings grant exactly the profile's harness permissions.
- **`HRN-6`** The provider MUST offer the agent exactly the harness-native tools in the profile's tools, mapping each tool name to the harness's own name.
- **`HRN-7`** The provider MUST map the model effort to the harness's own form.
- **`HRN-8`** The provider MUST turn on the selected harness's telemetry export in the same step as the harness permissions.
- **`HRN-9`** When the profile's telemetry content is `omit`, the provider MUST turn off the harness's prompt and response content logging.
- **`HRN-10`** For every supported harness, the integration MUST turn off the harness's built-in sandbox and native command approval prompts at adapter launch on ACP or harness launch on headless runs, and MUST keep the profile's tool grants and harness permissions.
- **`HRN-11`** On ACP runs, the integration MUST use ACP permission requests in place of native command approval prompts, with execution answers under `TRN-9` and interactive session answers under [Requests from the agent](agent-session-attachment.md#requests-from-the-agent).

## 14. Hooks

See [Agent hooks](agent-hooks.md).

### 14.1. Combining and running hooks

- **`HK-1`** The provider MUST run every hook of the profile in each execution of it, and it MUST add the specification's hooks to the profile's hooks.
- **`HK-2`** At `before-agent` and `sandbox-violation`, the provider MUST run the profile's hooks before the specification's hooks.
- **`HK-3`** At `stop-requested`, `session-end`, `after-agent`, and `before-teardown`, the provider MUST run the specification's hooks before the profile's hooks.
- **`HK-4`** Within one source and one occurrence of a provider event, the provider MUST run the hooks one at a time in the order they appear.
- **`HK-5`** The provider MUST run each hook's program directly, and it MUST NOT run it through a shell.
- **`HK-6`** The provider MUST start a hook without a working directory in the workspace root.
- **`HK-7`** The provider MUST apply a timeout of 60 seconds to a hook without a timeout.
- **`HK-8`** The provider MUST run each hook with the agent's workspace access and mediated execution authority, and MUST NOT give the hook an execution token inside the guest.
- **`HK-9`** For a provider event, the provider MUST write the hook input with the `event`, `hook`, and `harness` fields to the hook's standard input.
- **`HK-10`** The provider MUST count a hook run as failed on a nonzero exit status, a timeout, or a failed start.

### 14.2. Provider events

- **`HK-11`** The provider MUST run the `before-agent` hooks after the workspace build ends and before it starts the agent.
- **`HK-12`** The provider MUST start the time limit when the first `before-agent` hook starts, or when the agent starts if no `before-agent` hook runs.
- **`HK-13`** When a `before-agent` hook fails, the provider MUST skip the remaining `before-agent` hooks and end the execution `failed` with `hook-failure` as its reason.
- **`HK-14`** When the provider observes a headless harness exit or ACP session closure, for any reason, it MUST run the `session-end` hooks.
- **`HK-15`** The provider MUST run the `session-end` hooks within a session-end budget of 10 seconds in total, outside the time limit, starting when it observes a headless harness exit or ACP session closure after any stop grace period.
- **`HK-16`** When the session-end budget runs out, the provider MUST skip the `session-end` hooks still to come.
- **`HK-17`** When a `session-end` hook fails, the provider MUST record the failure, and it MUST NOT change the execution's outcome for it.
- **`HK-18`** When the agent completes on its own and the provider observes verified termination, the provider MUST check the final answer before it runs the `after-agent` hooks.
- **`HK-19`** When the agent completes on its own with success and a valid final answer, the provider MUST run each `after-agent` hook.
- **`HK-20`** When the agent completes on its own with a failure or invalid output, the provider MUST run only the `after-agent` hooks with `when` set to `exit`.
- **`HK-21`** When an `after-agent` hook fails, the provider MUST skip the remaining `after-agent` hooks with `when` set to `success`.
- **`HK-22`** When an `after-agent` hook fails after the agent completed with success and a valid final answer, the provider MUST end the execution `failed` with `hook-failure` as its reason and no payload.
- **`HK-23`** When the agent completed with a failure or invalid output, the provider MUST end the execution `failed` with the agent's own reason once the `after-agent` hooks end.
- **`HK-24`** When a time limit, resource limit, cancellation, or supersession stops a hook or the agent, the provider MUST skip the remaining `before-agent` and `after-agent` hooks.
- **`HK-25`** When a hook runs out its timeout, the provider MUST signal the hook to stop, and it MUST stop the hook by force if the hook outlasts the stop grace period.
- **`HK-26`** After it stops a hook for a provider event, the provider MUST NOT start the next step before it observes verified termination of the hook's processes.
- **`HK-27`** A hook's timeout MUST end no later than the time limit, and a `stop-requested`, `session-end`, or `before-teardown` hook's timeout MUST end no later than its event's budget.

### 14.3. Hook dispatch

- **`HK-28`** For each harness event that has hooks, the provider MUST write one entry that runs the afakt CLI into the harness's global hook configuration under `${home}`.
- **`HK-29`** The provider MUST keep the harness's global hook configuration outside the workspace, where the sandbox denies the agent writes.
- **`HK-30`** Each hook entry's timeout MUST be at least the sum of the timeouts of the hooks it dispatches.
- **`HK-31`** Each hook entry's timeout SHOULD add a margin for the time the afakt CLI takes to start and combine decisions.
- **`HK-32`** Before it starts the agent, the provider MUST verify that the written hook configuration has an entry for each harness event with installed hooks.

### 14.4. Harness support

- **`HK-33`** At launch on both transports, the provider MUST assess harness support for the events and decisions that the combined hook list uses, from the [harness capability matrix](agent-hooks.md#harness-capability-matrix) for the selected harness and version.
- **`HK-34`** The provider MUST install a hook whose event and decisions the selected harness supports.
- **`HK-35`** The provider MUST install a hook whose event or decisions the selected harness partly supports, with a `hook-partial` warning.
- **`HK-36`** The provider MUST skip a hook whose event, or one of whose decisions, the selected harness doesn't support, with a `hook-unsupported` warning.
- **`HK-37`** A support warning MUST identify the hook, the event or decision, and the harness, and the provider MUST list it in the provenance record and the evidence.
- **`HK-38`** The provider MUST NOT reject a launch for partial or missing harness support of a hook.
- **`HK-39`** A provider that runs Codex MUST complete what Codex requires for an execution's hooks to run, including its trust review.

### 14.5. Workspace repository hooks

- **`HK-40`** Before it starts the agent, the provider MUST read the project hook configuration that the selected harness reads from the workspace.
- **`HK-41`** For each hook entry in a workspace repository, the provider MUST record the entry's file, native event, and command in the provenance record.
- **`HK-42`** The provider MUST publish each hook configuration file it finds in the project as evidence.
- **`HK-43`** The provider MUST NOT change, move, or remove a hook configuration file that comes from a workspace repository.

### 14.6. Hook evidence

- **`HK-44`** The provider MUST publish a record of each hook as evidence with the hook's name, event, source, command, own environment entries after variable replacement, and working directory.
- **`HK-45`** When a provider event's hook run ends, the provider MUST add the run's decision, exit status, failure, start time, and end time to the hook record.
- **`HK-46`** When a provider event's hook run ends, the provider MUST publish the run's standard output and standard error logs as evidence.

### 14.7. Stop, violation, and teardown events

- **`HK-47`** When a cancellation, a supersession, or the time limit ends the agent's run, the provider MUST run the matching `stop-requested` hooks before it signals the agent to stop.
- **`HK-48`** The provider MUST run the `stop-requested` hooks within a stop budget of 10 seconds in total, outside the time limit, and it MUST signal the agent once they end or the budget runs out.
- **`HK-49`** The provider MUST NOT run the `stop-requested` hooks for a resource limit, or for a stop that arrives while no agent runs.
- **`HK-50`** Each time the sandbox reports a violation by the agent's processes, the provider MUST run the `sandbox-violation` hooks that match its kind, with the kind and the denied host or path in the hook input.
- **`HK-51`** When a `sandbox-violation` hook returns `terminate`, the provider MUST skip the remaining hooks for that occurrence, stop the agent as after a cancellation without the `stop-requested` hooks, and end the execution `failed` with `sandbox-violation` as its reason.
- **`HK-52`** At every end of a generation whose sandbox it opened, the provider MUST run the `before-teardown` hooks after the other hooks and after it records the workspace changes, and before it builds the result.
- **`HK-53`** The provider MUST run the `before-teardown` hooks within a teardown budget of 10 seconds in total, outside the time limit, and it MUST skip the hooks still to come once the budget runs out.
- **`HK-54`** When a `stop-requested`, `sandbox-violation`, or `before-teardown` hook fails, the provider MUST record the failure, and it MUST NOT change the execution's outcome for it.
- **`HK-55`** The provider MUST set `AFAKT_END_STATE` for each `before-teardown` hook to the terminal state the execution enters next.
- **`HK-56`** On ACP runs, the provider MUST install harness event hooks through the harness's own hook system inside the sandbox, and it MUST NOT substitute ACP messages for those events.
- **`HK-57`** The provider MUST apply `session-start` to the start of the headless harness session or the opening of the ACP session, before the first model request.

## 15. Execution states

See [Execution states](agent-execution.md#execution-states).

- **`ST-1`** The provider MUST change an execution's state only along a transition that [Execution states](agent-execution.md#execution-states) allows.
- **`ST-2`** The provider MUST NOT change a terminal state.
- **`ST-3`** The provider MUST move an execution into a terminal state only on its own observation of the outcome.
- **`ST-4`** The provider MUST NOT count a stop as observed before verified termination of every process it started for the execution, helper processes included.
- **`ST-5`** The provider MUST NOT treat a timeout as proof that the agent stopped.
- **`ST-6`** When the provider observes the first `before-agent` hook or the agent start of a `pending` execution, it MUST move the execution to `running`.
- **`ST-7`** When an attempt to start the agent returns no answer the provider can trust, the provider MUST move the execution to `launch-unknown`.
- **`ST-8`** From `launch-unknown`, the provider MUST move the execution to `running` when it observes the agent running, to `succeeded` or `failed` by the observed outcome when it observes that the agent already ended, and to `failed` when it observes that the execution has no agent.
- **`ST-9`** When a cancellation arrives for a `pending` execution whose agent hasn't started, the provider MUST move the execution to `canceled`.
- **`ST-10`** When the provider acknowledges a cancellation of a `launch-unknown` or `running` execution, it MUST move the execution to `cancel-requested`.
- **`ST-11`** When the agent of a `cancel-requested` execution completes on its own before the cancellation takes effect, the provider MUST move the execution to `succeeded` or `failed` by the observed outcome.
- **`ST-12`** When a time limit, resource limit, or stop grace period runs out before the provider observes the agent stop, the provider MUST move the execution to `unknown-stopped` with that stop cause.
- **`ST-13`** When the provider observes the agent stop for an `unknown-stopped` execution, it MUST move the execution to `canceled` for a cancellation or supersession stop cause, and to `terminated` for a limit stop cause.
- **`ST-14`** The provider MUST answer `launch-unknown` and `unknown-stopped` as non-terminal states.
- **`ST-15`** The provider MUST NOT change an execution's state for an `unavailable` answer.
- **`ST-16`** When an execution changes state, the provider MUST append the transition, its time, and the observation behind it to the state history.
- **`ST-17`** The provider MUST keep the phase at `before-agent`, `agent`, `after-agent`, or `before-teardown` while that part of the generation runs, and at none otherwise.
- **`ST-18`** The provider MUST update the last observed time each time it observes the hooks or the agent.

## 16. Cancellation, limits, and supersession

See [Cancellation and limits](agent-execution.md#cancellation-and-limits) and [Generations and supersession](agent-execution.md#generations-and-supersession).

- **`STP-1`** When the provider acknowledges a cancellation, it MUST signal the running hook or agent to stop.
- **`STP-2`** The provider MUST stop by force a signaled hook or agent that outlasts the stop grace period.
- **`STP-3`** When the time limit runs out, the provider MUST signal the running hook or agent to stop.
- **`STP-4`** When a hook or the agent exceeds a resource limit, the provider MUST signal it to stop.
- **`STP-5`** The provider MUST enforce the time limit and every resource limit over the hooks and the agent together, apart from the `stop-requested`, `session-end`, and `before-teardown` hooks.
- **`STP-6`** When the provider accepts a higher generation while an earlier generation is non-terminal, it MUST cancel the earlier generation with `supersession` as its reason and record the superseding generation in its superseded-by field.
- **`STP-7`** The provider MUST keep a higher generation in `pending` while an earlier generation of the same execution ID is non-terminal.

## 17. Output and result

See [Execution result](agent-execution-contract.md#execution-result).

- **`RES-1`** The provider MUST take the final answer from the headless harness's final-answer channel or the ACP turn's last agent message and check it against the execution's output schema.
- **`RES-2`** When the agent completes with success but its final answer doesn't match the output schema, the provider MUST end the execution `failed` with `invalid-output` as its reason and no payload.
- **`RES-3`** When an execution enters a terminal state, the provider MUST record the workspace changes and build the execution result with the fields that [Execution result](agent-execution-contract.md#execution-result) lists.
- **`RES-4`** The provider MUST NOT change an execution result once it exists.
- **`RES-5`** A result with a payload MUST include the payload's digest.
- **`RES-6`** On a headless run, the result MUST include the observed exit status when the harness reports one, and it MUST NOT include an ACP stop reason.
- **`RES-7`** When the harness or the model provider reports the usage of an execution, the result MUST include the observed usage.
- **`RES-8`** The provider MUST record the acceptance time of the launch, the observed start and end times of the agent, and the end time of the execution, each once.
- **`RES-9`** The provider MUST end an execution `succeeded` only when the agent completed on its own with success and produced a valid payload, and each `after-agent` hook succeeded.
- **`RES-10`** When an execution ends, the provider MUST record exactly one execution-ended event with the execution ID, generation, terminal state, reason, duration, and labels, even when it observes the end more than once.
- **`RES-11`** An ACP result MUST include the reported ACP stop reason, and it MUST NOT include a harness exit status.
- **`RES-12`** The provider MUST process ACP `end_turn` through the success path, including output validation and `after-agent` hooks, and use `agent-exit` as the reason when that path succeeds.
- **`RES-13`** The provider MUST process ACP `cancelled` through the cancellation path and end the execution `canceled` after verified termination, with `cancellation` or `supersession` as its reason. A provider stop for a time limit or resource limit MUST keep its limit cause and follow the limit path.
- **`RES-14`** The provider MUST end an ACP execution that returns `refusal` as `failed` with `agent-refusal` as its reason once the applicable hooks and verified termination finish.
- **`RES-15`** The provider MUST end an ACP execution that returns `max_tokens` or `max_turn_requests` as `failed` with `resource-limit` as its reason once the applicable hooks and verified termination finish.

## 18. Provenance

See [Provenance record](agent-execution-model.md#provenance-record).

- **`PRV-1`** The provider MUST keep one provenance record for each execution generation.
- **`PRV-2`** The provider MUST keep each requested value and each observed value in separate fields.
- **`PRV-3`** The provider MUST fill each requested field from the specification and the pinned profile.
- **`PRV-4`** The provider MUST fill each observed field only from a value that the harness, the model provider, or the agent host reported, and it MUST leave the field empty until a source reports one.
- **`PRV-5`** The provider MUST NOT change a recorded observed value for its generation.
- **`PRV-6`** The provider MUST update each hook's run and failure counts while the generation is non-terminal.
- **`PRV-7`** When a later report disagrees with a recorded observed value, the provider MUST publish the later report as evidence and mark the record's observations of that field as conflicting.
- **`PRV-8`** The provenance record MUST hold the canonical execution ID, the generation, the specification digest, the input digest, the configuration digest, and the trace ID and span ID of the generation's `afakt.generation` span.
- **`PRV-9`** The provider MUST fill the requested and observed fields of each row that [Provenance record](agent-execution-model.md#provenance-record) lists.
- **`PRV-10`** The result MUST include the provenance record as it stood when the execution entered its terminal state.

## 19. Evidence, retention, and cleanup

See [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup).

- **`EVD-1`** The provider MUST publish each evidence item as an evidence reference with the item's media type and digest.
- **`EVD-2`** The provider MUST publish the evidence items that [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup) lists.
- **`EVD-3`** The provider MAY publish evidence items beyond that list.
- **`EVD-4`** The provider MUST assign each evidence item to exactly one generation.
- **`EVD-5`** The provider MUST publish an execution's evidence and fill its observed provenance fields before the execution enters a terminal state.
- **`EVD-6`** When an execution enters a terminal state, the provider MUST set the end of its retention from the specification's retention.
- **`EVD-7`** The provider MUST keep an execution's workspace and evidence while its retention hasn't ended or a retention hold applies to it.
- **`EVD-8`** The provider MAY run cleanup for an execution at any time after its retention ends.
- **`EVD-9`** The provider MUST run cleanup on the placed agent host.
- **`EVD-10`** Cleanup MUST skip an execution that's non-terminal, held, still has running processes, or has a workspace that a `pending` execution on the same agent host lists as a source.
- **`EVD-11`** Cleanup MUST check the skip conditions again right before it removes anything.
- **`EVD-12`** Cleanup MUST remove only the items that belong to the execution, and it MUST remove the scratch directory with the workspace.
- **`EVD-13`** When cleanup can't remove an item, the provider MUST record the item and mark the cleanup `incomplete`.
- **`EVD-14`** The provider MUST report the terminal state of an execution that cleanup removed through status.

## 20. Restart and reconciliation

See [Reconciliation after restart or retry](agent-execution.md#reconciliation-after-restart-or-retry).

- **`RST-1`** The provider MUST persist execution records on the agent host across a restart of its afakt worker.
- **`RST-2`** A restart of the afakt worker MUST NOT stop its running agents. The agent host supervisor MUST keep their `VMM` processes and mediators running independently of that worker.
- **`RST-3`** When the afakt worker starts again, the provider MUST load the execution records.
- **`RST-4`** Until a restarted worker observes an execution's agent again, the provider MUST answer status for that execution with `unavailable`.
- **`RST-5`** When a restart interrupted a step before the agent started, the provider MUST wait for verified termination of that step's processes and then end the execution `failed` with `worker-restart` as its reason.
- **`RST-6`** While the agent of a restarted execution runs, the provider MUST continue the execution from its recorded state.
- **`RST-7`** The provider MUST count a `stop-requested`, `session-end`, `after-agent`, or `before-teardown` hook that a restart interrupted as a failed run of that hook.
- **`RST-8`** When a restart interrupted output validation or the result build, the provider MUST run that step again from its beginning.
- **`RST-9`** After restart or checkpoint restoration, the provider MUST reconnect session channels through the mediator and MUST check port publications and generation identity before resuming their traffic. Reconciliation MUST preserve the existing execution identity and host binding rules without choosing the open resume or migration generation semantics.

## 21. Telemetry

See [Agent observability](agent-observability.md).

### 21.1. Spans

- **`OBS-1`** The provider MUST emit each provider operation's span as a child of the calling activity's span.
- **`OBS-2`** The provider MUST start exactly one `afakt.generation` span for each accepted generation, as a child of the `afakt.provider.launch` span that accepted it.
- **`OBS-3`** The provider MUST end the `afakt.generation` span after its last `afakt.evidence.publish` and `afakt.sandbox.teardown` spans.
- **`OBS-4`** The provider MUST emit each span that [Spans](agent-observability.md#spans) lists for a step it performs, with that page's parent, attributes, status, events, and links.
- **`OBS-5`** The provider MUST follow each span's parent sampling decision.
- **`OBS-6`** The provider MUST record the trace ID and span ID of each open span in the execution record.
- **`OBS-7`** When a restarted worker resumes a step whose span was open, the provider MUST start a recovery span with the same name and parent, linked to the recorded span, with `afakt.recovered` set to true.
- **`OBS-8`** The provider MUST run reconciliation after a restart in an `afakt.execution.reconcile` span that links to each recovery span it starts.
- **`OBS-9`** The provider MUST start each `afakt.cleanup` span in a new trace that links to the `afakt.generation` span of its execution.
- **`OBS-15`** The provider MUST start the agent run span when the headless harness starts or when the relay starts the ACP adapter, and it MUST end the span only after verified termination of every process.
- **`OBS-16`** The provider MUST set `afakt.harness.transport` to `acp` or `headless` on the agent run span and in the execution metric attributes.

### 21.2. Metrics

- **`OBS-10`** The provider MUST publish the metrics that [Metrics](agent-observability.md#metrics) lists with their instruments, units, and attributes.
- **`OBS-11`** The provider MUST record a metric attribute value outside its enumerated set as `other`.
- **`OBS-12`** The provider MUST record the name of a repository profile, and each hook name in it, as `other` in every metric.
- **`OBS-13`** A metric attribute MUST NOT hold a value that [Cardinality](agent-observability.md#cardinality) excludes.

### 21.3. Content

- **`OBS-14`** When the profile's telemetry content is `omit`, the provider MUST NOT put the prompt, the rendered instructions, the instruction inputs, model responses, the final answer, or tool arguments and results in any span, event, metric, or log record.

## 22. Ports and previews

See [Agent ports and previews](agent-ports.md).

- **`PORT-1`** With `allowLocalBinding` true, the provider MUST permit listeners only on loopback ports of the sandbox's own network view, and MUST NOT permit binding the agent host's external interfaces. With it false, the provider MUST NOT permit listeners.
- **`PORT-2`** The guest relay MUST attribute discovered listening TCP sockets to the agent's process tree and execution generation. Discovery MUST NOT grant permission to publish or forward an unrelated listener.
- **`PORT-3`** The CLI MUST accept `afakt port expose <port>` inside the VM, and the matching MCP tool MUST support the same exposure with optional name and protocol. The guest relay MUST forward the request to the mediator, which MUST stamp the agent's identity and get a Cedar permit for that agent and port before creating a publication. The relay MUST verify listener ownership and guest loopback scope before connecting.
- **`PORT-4`** The mediator MUST maintain one publication per exposed listener on its execution or session with the random preview label, agent, port, generation, agent host, and the other fields in [Port discovery and records](agent-ports.md#port-discovery-and-records).
- **`PORT-5`** The mediator MUST close forwarding and mark the publication closed when its listener disappears outside suspension or its execution ends. It MUST NOT transfer closed routes or share grants to a new listener or generation. Resume MUST check publications before routing without choosing whether suspended previews wake the agent or fail.
- **`PORT-6`** The mediator MUST dial out to a preview broker beside the preview proxy and maintain that connection for preview and client tunnel streams. Preview transport MUST NOT require inbound listeners on the agent host or in the guest.
- **`PORT-7`** The preview proxy MUST assign each published HTTP port a URL with a random label under the preview domain. Routing MUST distinguish generations and publications and preserve site-root paths, WebSockets, and server-sent events.
- **`PORT-8`** Before HTTP forwarding or starting a WebSocket upgrade or an event stream, the preview proxy MUST apply Cedar policy to each viewer request and check afakt sign-in and execution or session access. A valid share grant MUST grant access only to its one port for a signed-in recipient.
- **`PORT-9`** The control plane MUST allow a user with execution or session access to create an expiring, revocable share link scoped to one port, MUST check expiry, revocation, and port state on use, and MUST close shared streams on expiry or revocation.
- **`PORT-10`** The preview proxy MUST run inside the API server process with a listener separate from the API and dashboard listener. It MUST isolate their origins and cookies as [HTTP previews](agent-ports.md#http-previews) defines and MUST strip afakt cookies, Authorization headers, and share-link secrets before forwarding to the app.
- **`PORT-11`** Raw TCP ports MUST be accessible only through an authenticated afakt client tunnel after execution or session authorization, and MUST NOT have public listeners or public share links.
- **`PORT-12`** The Cloudflare Tunnel and Tailscale Funnel providers MUST run their embedded connectors inside the API server process. They MUST keep provider credentials and identity state there and MUST NOT send them to agent hosts or sandboxes.
- **`PORT-13`** Tunnel publication MUST preserve preview authorization, origin separation, streaming, and the prohibition on public raw TCP ports.
- **`PORT-14`** The web client and chat bridges MUST show authorized session port records and preview links as they change, and a preview link MUST open a browser tab beside the session.
- **`PORT-15`** The mediator and control plane MUST publish the [ports and previews telemetry](agent-observability.md#ports-and-previews) with its instruments, units, bounded attributes, and execution trace links.
- **`PORT-16`** For each preview or client tunnel stream, the mediator MUST check that the publication is current and its generation matches, then connect through the guest relay to the guest's loopback listener.
- **`PORT-17`** The preview proxy MUST rewrite the Host header or configure the development server's allowed hosts so the server accepts preview host names.

## 23. MCP gateway

See [Agent MCP gateway](agent-mcp-gateway.md).

- **`MCP-1`** The gateway MUST serve afakt executions and sessions only and MUST NOT accept outside MCP clients.
- **`MCP-2`** The harness integration MUST configure exactly one MCP endpoint using Streamable HTTP through the guest relay and mediator and MUST turn off the harness's other MCP server configuration on ACP and headless launches.
- **`MCP-3`** The mediator MUST stamp agent identity on MCP control requests from the guest relay and supply the execution token outside the guest. The gateway MUST check that token and bind each request to its execution and generation, and MUST NOT forward the token upstream.
- **`MCP-4`** The gateway MUST keep upstream credentials outside the execution sandbox and MUST NOT expose them to the agent through configuration, tools, errors, evidence, or telemetry.
- **`MCP-5`** For remote servers using MCP OAuth authorization, afakt MUST complete authorization for the operator or required user identity, store and refresh tokens, and reuse valid authorization across executions. Static API keys MUST use the sandbox credential references.
- **`MCP-6`** The gateway MUST filter tool discovery and enforce server and tool allowlists on every call, and MUST apply deny entries before allow entries. Before recording a launch, the provider MUST reject an execution specification that requests wider MCP access than the profile permits with `override-not-allowed`.
- **`MCP-7`** The gateway MUST namespace tool names by configured server identity so distinct server and tool pairs never collide, and MUST reserve the `afakt` server identity for the built-in server.
- **`MCP-8`** The gateway MUST connect to remote servers over HTTP. In sandbox VM mode, the provider MUST start workspace stdio servers inside the VM under the agent user, stop them with it, and verify their termination.
- **`MCP-9`** afakt MUST supervise shared central stdio servers in a control plane sandbox, and the gateway MUST scope each request and response to its execution and upstream authorization identity.
- **`MCP-10`** Before dispatching a tool marked for approval, the gateway MUST raise an ACP permission request on the session or execution making the call and MUST wait for an allow answer. The gateway MUST NOT dispatch denied or canceled calls.
- **`MCP-11`** The session hub MUST present gateway approval requests and outcomes to every attached client. This includes the web client and chat bridges. It MUST accept interactive answers only from the attached writer.
- **`MCP-12`** For unattended executions without an attached writer, including headless runs and ACP executions with only a workflow writer, the gateway MUST apply the pinned profile's `mcp.unattended_approval` rule, MUST default to `deny` when absent, and MUST record the policy answer.
- **`MCP-13`** The gateway MUST expose the built-in afakt server's tools subject to the profile's server and tool allowlists. These tools MUST support starting sub-executions or sessions and sending messages to other sessions. They MUST also support reading execution status and results and declaring forwarded ports.
- **`MCP-14`** The built-in afakt server and API server MUST check every call against the execution's own capabilities and scopes, MUST bound child authority by the caller's authority, and MUST NOT grant more authority through an upstream credential, service identity, approval, or child profile.
- **`MCP-15`** Every gateway tool call, including denied and built-in calls, MUST produce an execution trace span and an audit record with server, tool, outcome, and duration as [Audit and telemetry](agent-mcp-gateway.md#audit-and-telemetry) defines. Audit records MUST NOT contain arguments, results, credentials, or execution tokens, and telemetry MUST omit arguments and results by default.
- **`MCP-16`** The gateway MUST publish tool call counts and durations, approval requests and outcomes, and upstream health with the instruments, units, and bounded attributes in [MCP gateway telemetry](agent-observability.md#mcp-gateway).
- **`MCP-17`** Guest MCP control traffic and gateway approval messages MUST remain subject to mediator policy, including the ACP gate where applicable. A tool approval MUST NOT bypass that policy or network rules.

## 24. Security considerations

A provider runs agent code it can't trust beside real credentials. The sandbox requirements in section 12 contain the agent, and the requirements below bound what leaves the agent host.

- **`SEC-1`** The provider MUST keep every secret value out of the execution environment.
- **`SEC-2`** The provider MUST keep every fetch credential outside the sandbox.
- **`SEC-3`** Evidence MUST NOT hold a secret value.
- **`SEC-4`** The provider MUST NOT put a secret, a credential value, or the execution token in any span, event, metric, or log record.
- **`SEC-5`** The provider MUST NOT put an environment value, hook input, hook output, or workspace file contents in any span, event, metric, or log record.
- **`SEC-6`** The provider MUST NOT give the agent or a hook an egress path that bypasses the per-agent mediator. Enforcing this restriction MUST NOT depend on proxy environment variables.
- **`SEC-7`** The guest MUST NOT contain long-lived secrets, brokered credential values, or execution tokens. Credentials MUST remain outside the guest during checkpoint capture and restoration.
- **`SEC-8`** The mediator MUST stamp the agent's identity on CLI, MCP, and port exposure control requests from the guest relay. It MUST NOT accept guest-supplied identity as authority to act for another agent.
- **`SEC-9`** Every mediator network, HTTP, ACP, and control decision MUST enter one decision log for the agent. That log MUST exclude credential values, execution tokens, and request and response content.
- **`SEC-10`** ACP policy MUST NOT prevent supervisor and `VMM` lifecycle control. A policy permit MUST NOT bypass network policy, broaden a capability, or expose credentials inside the guest.

## 25. References

- [RFC 2119](https://www.rfc-editor.org/rfc/rfc2119) and [RFC 8174](https://www.rfc-editor.org/rfc/rfc8174), which together form BCP 14, define the key words.
- [Agent execution](agent-execution.md), [Agent execution contract](agent-execution-contract.md), [Agent execution model](agent-execution-model.md), [Agent profile](agent-profile.md), [Agent workspace](agent-workspace.md), [Agent workspace image](agent-workspace-image.md), [Agent sandbox](agent-sandbox.md), [Agent ports and previews](agent-ports.md), [Agent MCP gateway](agent-mcp-gateway.md), [Agent hooks](agent-hooks.md), and [Agent observability](agent-observability.md) hold the design behind each section.
