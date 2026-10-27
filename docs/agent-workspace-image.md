---
type: Design
title: Agent workspace image
description: Prebuilt workspace directory trees in Open Container Initiative (OCI) images, their sandboxed builds, registry and object storage, and agent host caches for fast workspace start.
tags: [afakt, agents, agent-launch, workspace, images, storage]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent workspace image

[Public key infrastructure](pki.md).

A workspace image contains a prepared directory tree with repository checkouts and installed dependencies plus build outputs. To start, an agent launch places that tree in its agent workspace. It can then reach its agent without repeating the clone and dependency installation or the build. A warm agent host reuses local content, while a cold host fetches and unpacks the missing layers before the agent starts.

## Image contents

A workspace image is an Open Container Initiative (OCI) image with a manifest, a config blob, and layers. Those layers contain a directory tree and unpack beneath a workspace source's target path. Inside the [agent sandbox VM](agent-sandbox.md#Enforcement), the agent uses a writable workspace disk built from these layers. A separate read-only root disk supplies the guest operating system. Workspace image process settings don't determine how afakt starts the agent harness.

The config records the inputs and steps that built the tree:

| Field | Content |
| --- | --- |
| Sources | Each source repository, its full commit ID, and its target path within the tree. The checkout includes the repository metadata needed for agent work. |
| Build steps | The ordered commands and working directories that produced the dependencies and outputs. |
| Agent profile | The source, name, version, and digest of the profile whose sandbox confined the steps. |
| Base image | The reference and resolved manifest digest of the optional base workspace image, or no base. |
| Built at | The build completion time as a Coordinated Universal Time (UTC) timestamp in Request for Comments (RFC) 3339 form. |
| Compatibility | The operating system and architecture used for the build, with any toolchain constraints its outputs require. |

A base image supplies an earlier workspace directory tree, which the builder unpacks before fetching the sources and running the steps. It doesn't supply the agent host's operating system or install tools outside the agent workspace, and its layers remain part of the derived image's layer chain. Later layers take precedence. This includes OCI whiteouts that remove earlier entries and opaque directory markers that hide earlier directory contents.

The agent launch's placement requirements identify agent hosts compatible with the image's outputs, since reusing a tree doesn't make native dependencies portable across operating systems, architectures, or incompatible toolchains.

## Image sources and digest pinning

[Image sources and digest pinning](agent-workspace-lifecycle.md#Image%20sources%20and%20digest%20pinning).

## Workspace image definitions

A workspace image definition names the source repositories and their target paths, an optional base image, the ordered build steps, the agent profile whose agent sandbox runs them, and the destination image repository and tag. It also sets the number of successful builds to keep and its build triggers.

| Trigger | Build input |
| --- | --- |
| Schedule | Source branch heads resolved when the scheduled build starts. Schedules use Coordinated Universal Time (UTC). |
| Push to a source repository's branch | The pushed commit for that source, with the other source revisions resolved when the build starts. |
| Manual request | The requested source commits, with omitted revisions resolved from the definition's branches. |

Each build records a fixed copy of its definition, full source commit IDs, the resolved base image digest, and the pinned profile version and digest before running steps. Retries use those recorded inputs. A new trigger creates a separate image build, even if its source commits match an earlier build.

## Image builds

[Workspace image build lifecycle](agent-workspace-lifecycle.md#Image%20builds).

## Registry

afakt needs an OCI distribution registry. It must cache base images on pull. Registry selection and the choice between an embedded service and a supervised child process remain [`OQ-IMG-16`](agent-open-questions.md#%60OQ-IMG-16%60). A supervised child needs startup readiness, failure observation, restart with backoff, and a bounded shutdown period.

An afakt-owned registry interface covers reference resolution, manifest and blob transfer, tag publication, retention references, and collection. Workers and the image builder depend on that interface rather than vendor-specific calls or storage metadata. Replacing the registry preserves those operations and their authorization and retention rules.

### Registry authentication

The registry accepts bearer tokens only. After authorizing the worker's request, the control plane issues a short-lived registry token with an issuer, a subject identifying the worker or build job, the registry as audience, a validity interval, and scopes naming repositories and actions.

On every request, the registry checks the signature against the control plane's published JavaScript Object Notation (JSON) Web Key Set (JWKS), the trusted issuer and registry audience, the validity interval, and the repository and action scopes. It rejects missing or invalid tokens, unknown signing keys, and requests outside those scopes. Knowing a digest alone never authorizes a read. Blob reads, ranged reads, cross-repository blob mounts, tag operations, and deletion all receive the applicable scope check. Push scopes don't grant deletion or registry administration.

For worker-owned pulls, agent hosts use tokens issued to their worker for the required image repositories. Pull ownership remains [`OQ-IMG-8`](agent-open-questions.md#%60OQ-IMG-8%60). A build job pushes through its worker with a token scoped to its destination image repository. Reading a base needs a separate pull grant. Tokens and store credentials stay outside the agent sandbox, images, evidence, and telemetry. Workers renew registry tokens through the control plane when a transfer outlasts their validity. The registry pulls through and caches base images from configured upstream registries. Before fetching upstream content, it authenticates and authorizes a downstream request. It preserves upstream content digests and keeps upstream credentials outside build sandboxes. Its signing key configuration follows the control plane's JWKS rotation without accepting unknown keys.

## Object store

afakt includes an embedded S3 Application Programming Interface (API) compatible object store component using RustFS as a supervised child process for single-node and small installs. Under supervision, afakt starts and stops it and waits for its readiness before starting the registry. Deployments can instead configure any external S3 API compatible endpoint that meets the storage contract. The registry and every other afakt component that stores blobs use that endpoint. Worker layer caches remain local disposable data.

The storage contract includes object reads and writes, metadata lookup, paginated listing, deletion, multipart uploads and aborting incomplete uploads, byte-range reads, conditional writes with `If-None-Match` and `If-Match`, and lifecycle expiry. Conditional writes reject failed preconditions rather than overwriting an object. afakt explicitly configures path-style addressing, the endpoint, bucket, region, transport security, and credentials.

Lifecycle expiry applies to temporary objects and abandoned upload data. Registry manifests and reachable blobs remain under registry retention and collection. An object-store expiry rule can't remove content that an agent launch protects. Using an external store changes the endpoint and credentials while preserving image semantics and agent launch retention.

## Agent host caches and startup

[Agent host caches and startup](agent-workspace-lifecycle.md#Agent%20host%20caches%20and%20startup).

## Workspace disk deltas and portable revisions

[Workspace disk deltas and portable revisions](agent-workspace-lifecycle.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

## Placement and preheating

[Placement and preheating](agent-workspace-lifecycle.md#Placement%20and%20preheating).

## Retention and garbage collection

[Retention and garbage collection](agent-workspace-lifecycle.md#Retention%20and%20garbage%20collection).

## Telemetry

[Operations and telemetry](agent-workspace-lifecycle.md#Operations%20and%20telemetry).
