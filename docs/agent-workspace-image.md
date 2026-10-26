---
type: Design
title: Agent workspace image
description: Prebuilt workspace directory trees in Open Container Initiative (OCI) images, their sandboxed builds, registry and object storage, and agent host caches for fast workspace start.
tags: [afakt, agents, execution, workspace, images, storage]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent workspace image

A workspace image contains a prepared directory tree with repository checkouts and installed dependencies plus build outputs. To start, an execution places that tree in its workspace. It can then reach its agent without repeating the clone and dependency installation or the build. A warm agent host reuses local content, while a cold host fetches and unpacks the missing layers before the agent starts.

## Image contents

A workspace image is an Open Container Initiative (OCI) image with a manifest, a config blob, and layers. Those layers contain a directory tree and unpack beneath a workspace source's target path. Inside the [sandbox VM](agent-sandbox.md#enforcement), the agent uses a writable workspace disk built from these layers. A separate read-only root disk supplies the guest operating system. Workspace image process settings don't determine how afakt starts the harness.

The config records the inputs and steps that built the tree:

| Field | Content |
| --- | --- |
| Sources | Each source repository, its full commit ID, and its target path within the tree. The checkout includes the repository metadata needed for agent work. |
| Build steps | The ordered commands and working directories that produced the dependencies and outputs. |
| Agent profile | The source, name, version, and digest of the profile whose sandbox confined the steps. |
| Base image | The reference and resolved manifest digest of the optional base workspace image, or no base. |
| Built at | The build completion time as a Coordinated Universal Time (UTC) timestamp in RFC 3339 form. |
| Compatibility | The operating system and architecture used for the build, with any toolchain constraints its outputs require. |

A base image supplies an earlier workspace directory tree, which the builder unpacks before fetching the sources and running the steps. It doesn't supply the agent host's operating system or install tools outside the workspace, and its layers remain part of the derived image's layer chain. Later layers take precedence. This includes OCI whiteouts that remove earlier entries and opaque directory markers that hide earlier directory contents.

The execution's placement requirements identify agent hosts compatible with the image's outputs. Reusing a tree doesn't make native dependencies portable across operating systems, architectures, or incompatible toolchains.

## Image sources and digest pinning

An `image` [workspace source](agent-workspace.md#workspace-sources) supplies an image reference and a target path. This reference specifies an image repository and tag, or an image repository and manifest digest. Relative path and non-overlap rules for the target match those of the other source kinds. Before applying workspace access and running hooks, the provider places the directory tree there.

During the workspace build, the provider resolves a tag to a manifest digest once and records both the requested reference and the resolved digest on the execution generation. It persists the digest before fetching layers or making the workspace available. Resolution also establishes the [retention reference](#retention-and-garbage-collection) that protects the image from collection. Failure to record and protect the digest fails the workspace build. A retried operation uses the recorded digest without resolving the tag again. To reuse the same content, a later generation can request the repository and that digest explicitly. Supplying the tag instead resolves it for that generation's own workspace. Resolution and cache use leave the immutable execution specification and its digest unchanged. The [provenance record](agent-execution-model.md#provenance-record) keeps the requested reference separate from the observed manifest digest.

The provider verifies the manifest, config, and layer digests before using their content. A failed pull, verification, or unpack ends the workspace build with `workspace-failure`, without falling back to another tag or to an unverified cache entry.

An image can contain a checkout at an older commit than an execution needs.

Following an `image` source with a `repository` source at the same target would allow an incremental fetch and checkout while retaining dependencies and build outputs. The source rules reject duplicate targets. That form requires a source rule change and remains an [open question](agent-open-questions.md#agent-workspace-image).

## Workspace image definitions

A workspace image definition names the source repositories and their target paths, an optional base image, the ordered build steps, the agent profile whose sandbox runs them, and the destination image repository and tag. It also sets the number of successful builds to keep and its build triggers.

| Trigger | Build input |
| --- | --- |
| Schedule | Source branch heads resolved when the scheduled build starts. Schedules use Coordinated Universal Time (UTC). |
| Push to a source repository's branch | The pushed commit for that source, with the other source revisions resolved when the build starts. |
| Manual request | The requested source commits, with omitted revisions resolved from the definition's branches. |

Each build records a fixed copy of its definition, full source commit IDs, the resolved base image digest, and the pinned profile version and digest before running steps. Retries use those recorded inputs. A new trigger creates a separate image build, even if its source commits match an earlier build.

## Image builds

The image builder coordinates an execution-like job on an agent host through the workflow orchestrator and the host's task queue. This job reserves a slot under the [slot capacity](agent-placement.md#slot-capacity) rules. It releases the slot when it ends. Commands run inside the same sandbox VM boundary as agents, bounded by the pinned profile's sandbox, workspace access, credentials, and limits, without requiring a container build daemon.

1. Resolve and record the build inputs, then reserve a compatible agent host.
2. Materialize the optional base into a fresh workspace directory and fetch each source repository at its recorded commit. Repository fetches run outside the sandbox, with the same fetch credential handling as `repository` sources. Fetch credentials never enter the sandbox or the checkout's stored configuration.
3. Run the ordered build steps inside the profile's sandbox. Their network traffic uses the credential broker under that profile's rules. Dependency installation and compilation write into the workspace tree that the image captures.
4. After every build process stops, the worker snapshots the tree into layers and records the build config. It verifies the output digests, then pushes the blobs and manifest through the registry interface. The push token stays with the worker outside the sandbox, for use only with the build's image repository.
5. Publish a distinct build tag and move the definition's output tag only after the complete image is available. A failed build or upload leaves the previous output tag unchanged. The builder serializes publication for each definition so overlapping builds can't race a tag update.
6. Release the build's slot and request any configured [preheating](#placement-and-preheating).

The snapshot covers only the workspace tree, excluding the sandbox home, scratch directory, broker state, and worker credentials. Secrets never belong in image config or layers. A step that needs a secret uses the profile's brokered access where it fits. How to supply other build secrets without leaving them in any layer remains an open question. Removing a secret in a later layer doesn't remove it from an earlier one.

## Registry

afakt includes an OCI distribution registry using Zot as a supervised child process. Zot runs separately. afakt doesn't link it into the binary but starts it and waits for readiness. It observes failures and restarts it with backoff, then stops it within a bounded shutdown period. The Zot build includes only extensions afakt uses, including pull-through caching for base images.

An afakt-owned registry interface covers reference resolution, manifest and blob transfer, tag publication, retention references, and collection. Workers and the image builder depend on that interface rather than Zot-specific calls or storage metadata. Replacing the registry preserves those operations and their authorization and retention rules.

### Registry authentication

The registry accepts bearer tokens only. After authorizing the worker's request, the control plane issues a short-lived registry token with an issuer, a subject identifying the worker or build job, the registry as audience, a validity interval, and scopes naming repositories and actions.

This token is separate from the agent's [execution token](agent-sandbox.md#execution-token).

On every request, the registry checks the signature against the control plane's published JSON Web Key Set (JWKS), the trusted issuer and registry audience, the validity interval, and the repository and action scopes. It rejects missing or invalid tokens, unknown signing keys, and requests outside those scopes. Knowing a digest alone never authorizes a read. Blob reads, ranged reads, cross-repository blob mounts, tag operations, and deletion all receive the applicable scope check. Push scopes don't grant deletion or registry administration.

Agent hosts pull using tokens issued to their worker for the required image repositories. A build job pushes through its worker with a token scoped to its destination image repository. Reading a base needs a separate pull grant. Tokens and store credentials stay outside the sandbox, images, evidence, and telemetry. Workers renew registry tokens through the control plane when a transfer outlasts their validity. The registry pulls through and caches base images from configured upstream registries. Before fetching upstream content, it authenticates and authorizes a downstream request. It preserves upstream content digests and keeps upstream credentials outside build sandboxes. Its signing key configuration follows the control plane's JWKS rotation without accepting unknown keys.

## Object store

afakt includes an embedded S3 API compatible object store component using RustFS as a supervised child process for single-node and small installs. Under supervision, afakt starts and stops it and waits for its readiness before starting the registry. Deployments can instead configure any external S3 API compatible endpoint that meets the storage contract. The registry and every other afakt component that stores blobs use that endpoint. Worker layer caches remain local disposable data.

The storage contract includes object reads and writes, metadata lookup, paginated listing, deletion, multipart uploads and aborting incomplete uploads, byte-range reads, conditional writes with `If-None-Match` and `If-Match`, and lifecycle expiry. Conditional writes reject failed preconditions rather than overwriting an object. afakt explicitly configures path-style addressing, the endpoint, bucket, region, transport security, and credentials.

Lifecycle expiry applies to temporary objects and abandoned upload data. Registry manifests and reachable blobs remain under registry retention and collection. An object-store expiry rule can't remove content that an execution protects. Using an external store changes the endpoint and credentials while preserving image semantics and execution retention.

## Agent host caches and startup

Each worker keeps a local layer cache keyed by content digest and a cache of unpacked layers keyed by the verified layer digest. Cache entries are immutable. Only complete, verified entries become reusable, and the caches stay outside the agent's writable workspace, distinct from the profile's named tool caches.

The worker fetches each image source's manifest and config. It finds the missing layers and fetches them with parallel ranged reads, verifying each assembled blob against its descriptor before admitting it to the layer cache. The worker discards a corrupt entry and fetches it again.

Unpacked layers become reusable only after extraction and validation finish.

The worker extracts independent layers in parallel into private staging directories, then builds a writable ext4 workspace disk from their entries. A separate read-only root disk supplies the guest operating system. Writes to disjoint final paths can run in parallel. Overlapping paths, deletions, and opaque directories follow manifest order. The completed tree has exactly the result of applying the layers in that order. Even archive paths or links never cause extraction to write outside the source's target. The worker caches the unpacked base and disk by digest so agents from the same image reuse that base. Copy-on-write clones of the disk, using reflinks, APFS clones, or qcow2 overlays, give each agent its own writable disk. Where cloning is unavailable, a plain copy preserves independence. It never shares writable hard links with a cache entry or another workspace. Changing or deleting files in an execution leaves the cached layers and other executions unchanged.

All required content is present before hooks or the agent start. Whether the worker must pull it before VM boot or the guest may pull it through its mediator remains [open](agent-open-questions.md#agent-workspace-image).

Workspace start doesn't use Filesystem in User Space (FUSE) lazy pull or a snapshotter. A cache hit avoids a transfer or extraction. The worker still obtains control plane authorization for each execution's image repository even when every layer is local.

## Checkpoints and portable revisions

At an idle turn boundary, a local workspace checkpoint captures a block delta of the workspace disk with the guest filesystem frozen. An [agent checkpoint](agent.md#checkpoints) pairs it with the VM snapshot and harness session state from that same boundary.

A portable workspace revision is an OCI diff layer plus a content-hashed file inventory, pushed to the registry. Content hashes determine correctness. Modification times and sizes never determine it. Portable revisions preserve files across disk layouts. Local block deltas require their matching base disk and don't replace the portable revision format.

## Placement and preheating

Workers report cached layer digests as advisory host metadata. Among registered agent hosts that meet all placement requirements and have a free slot, the reservation API may prefer one that already caches more of the image's layers. Locality never overrides the binding of higher generations to their first agent host, the host required by an `execution` source, lease eligibility, or slot accounting. A cache report can become stale. A miss triggers an ordinary pull rather than a placement failure. See [Image locality](agent-placement.md#image-locality).

After a successful build, the control plane can enqueue a preheat request on chosen agent hosts' task queues. Each worker pulls and verifies the published digest with a pull token and populates both caches without starting an agent or creating a workspace. Preheating respects the host's free-disk target and doesn't reserve a future execution slot. The worker polls its task queue for preheat requests without the control plane connecting to the agent host. A preheat failure leaves the published image usable through an ordinary pull.

## Retention and garbage collection

Each definition retains tags for its last N successful builds, where N is its configured positive build count. Its output tag points to a complete successful build, and dropping an older build tag doesn't remove any execution's protection of its digest.

The control plane keeps retention references for resolved image digests used by non-terminal executions and by terminal executions whose retention hasn't ended or whose retention hold remains in force. Resolution protects the digest before making it available to the workspace build.

The registry interface uses protected tags or collection roots with the same effect for those references, separate from the last N build tags. Releasing the final execution reference makes an image eligible for ordinary collection only when no other root retains it.

The registry collects untagged, unprotected manifests and unreferenced blobs online while pulls and pushes continue. Its collection roots include build tags, output tags, execution references, and active build inputs and uploads. Protection covers the manifest, config, and every referenced layer, including layers inherited from a base. Collection and reference creation coordinate so an image can't disappear between resolution and protection. If the control plane can't confirm that a reference is safe to release, collection retains it.

Worker caches have separate retention. The worker evicts least recently used layer blobs and unpacked entries to meet the free-disk [resource target](agent-placement.md#slot-capacity), skipping entries in use by a transfer or workspace materialization and bases backing active disks or retained checkpoint deltas. Preheating yields to that target. Eviction doesn't remove workspace files or registry content, and a later miss fetches the pinned digest again. If eviction can't recover enough space, the capacity controller still applies its slot rules.

## Telemetry

Each image build emits an `afakt.workspace.image.build` span covering source preparation, sandboxed steps, snapshotting, and publication. It joins its triggering workflow's trace through the build activity. An execution's `afakt.workspace.source` span records the resolved image digest and links to that build span when its context is available, connecting the execution trace to the prebuild without repeating the build at launch.

[Agent observability](agent-observability.md#workspace-images) defines the build duration and result metrics, layer cache hits and misses, fetched bytes, and workspace build duration split into fetch, unpack, and repository steps. Image references, digests, source URLs, and commands never become metric attributes. Build credentials and file contents remain outside telemetry under the same content rules as agent executions.
