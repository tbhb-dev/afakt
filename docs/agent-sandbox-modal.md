---
type: Design
title: "Agent sandbox provider for Modal Sandboxes"
description: Remote isolation, mediation, storage, and lifecycle for Modal Sandboxes.
tags: [afakt, agents, sandbox]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Agent sandbox provider for Modal Sandboxes

[Public key infrastructure](pki.md).

## Platforms and privileges

The Modal Sandboxes provider uses remote Linux execution managed through the Modal Application Programming Interface (API). Modal offers a gVisor runtime and a virtual-machine (VM) runtime. Omitting the runtime lets Modal select it. Graphics processing unit (GPU) sandboxes require gVisor. Modal permits resource requests and upper limits, with execution able to exceed a request unless an upper limit caps it. afakt's selected runtime, guest privileges, and compatibility with process confinement remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). [Sandboxes](https://modal.com/docs/guide/sandboxes), [Sandbox resources](https://modal.com/docs/guide/sandbox-resources).

The agent host worker retains provider management authority outside the sandbox under [Runtime privileges and platforms](agent-host.md#Runtime%20privileges%20and%20platforms). afakt rejects settings it can't enforce under [Enforcement](agent-sandbox.md#Enforcement). The host isolation baseline remains on hold under [`OQ-SBX-18`](agent-open-questions.md#%60OQ-SBX-18%60).

## Isolation boundary

Modal's gVisor runtime handles Linux system calls in userspace, while its VM runtime supplies a private Linux kernel behind Kernel-based Virtual Machine (KVM) hardware virtualization. Sandboxes lack authority over other Modal workspace resources by default. These statements define runtime boundaries, but don't describe containment after exploiting gVisor or the virtual-machine host. That gap and the runtime choice remain under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60). Protection from vendor access remains under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). [Networking and security](https://modal.com/docs/guide/sandbox-networking).

## Process placement and startup

Modal Sandboxes runs an entrypoint and other commands. It supplies sandbox execution rather than a required managed agent loop. afakt's [Host process sandbox](agent-sandbox.md#Host%20process%20sandbox) places afakt-agent, afakt-sandbox, and the harness inside that execution environment, while the agent host worker retains remote lifecycle control. The documentation doesn't establish compatibility of afakt-sandbox's confinement mechanisms or the guest relay's carrier with either runtime. Qualification remains under [`OQ-SBX-65`](agent-open-questions.md#%60OQ-SBX-65%60), [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Sandboxes](https://modal.com/docs/guide/sandboxes), [Running commands](https://modal.com/docs/guide/sandbox-spawn).

Commands requested through `afakt agent exec` require standalone afakt-sandbox, with placement, inherited environment, launch binding, and descendant lifetime unresolved under [`OQ-AGT-40`](agent-open-questions.md#%60OQ-AGT-40%60). Provider-installed hook protection and repository project configuration remain on hold under [`OQ-HOOK-16`](agent-open-questions.md#%60OQ-HOOK-16%60), [`OQ-HOOK-17`](agent-open-questions.md#%60OQ-HOOK-17%60).

Modal can place sidecar containers beside the main sandbox, with the isolation boundary of the selected runtime and separate images. They share resource allocation and a bridge network. Placement of the agent mediator or agent proxy in sidecars remains undecided. That remains under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Sandbox Sidecars](https://modal.com/docs/guide/sandbox-sidecars).

## Mediator connection and name resolution

Each launch requires [agent mediation](agent-mediator.md#Launch%20boundary) for control, session, preview, and network traffic.

Remote launches require the [agent proxy](agent-proxy.md#Launch%20boundary) and [global egress path](components.md#Agent%20egress%20gateway), with their placement and carrier subject to the provider mapping below.

Remote placement, mandatory transport, Domain Name System (DNS) mediation, and prevention of direct connections remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). Provider assertions for launch identity remain under [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60).

Modal allows outbound public addresses by default. Modal can block all outbound traffic. It can instead allow address ranges using Classless Inter-Domain Routing (CIDR) or domain names from Transport Layer Security (TLS) Server Name Indication (SNI) on port 443. Address and domain allowances are additive. [Networking and security](https://modal.com/docs/guide/sandbox-networking).

Domain Name System queries remain unrestricted unless Modal blocks all outbound traffic. Without decryption of Hypertext Transfer Protocol (HTTP) headers, domain filtering permits domain fronting to another tenant behind an allowed TLS endpoint. Establishing afakt's exclusive mediator path remains unresolved under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Networking and security](https://modal.com/docs/guide/sandbox-networking).

Modal's optional sidecar relay redirects main-container traffic on port 443 and refuses it while the sidecar with that name is absent. That traffic follows the sidecar's policy, which defaults to open. Other ports keep their separate controls. Traffic between sidecars and the main sandbox bypasses outbound allowlists, and sidecars keep access to Modal's resolver. Mandatory mediation of every protocol and name query remains unresolved. The provider mapping remains under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Sandbox Sidecars](https://modal.com/docs/guide/sandbox-sidecars).

Modal's Outbound Policy injects secret headers outside the sandbox by intercepting Hypertext Transfer Protocol Secure (HTTPS) on port 443. Unmatched traffic passes through, and sidecars aren't covered. This feature is incompatible with network blocking, address or domain allowlists, the sidecar relay, or memory snapshots. Ordinary Modal Secrets instead enter the sandbox environment. Optional OpenID Connect (OIDC) identity tokens also enter the sandbox and aren't afakt launch credentials. Broker integration and enrollment remain under [`OQ-SBX-23`](agent-open-questions.md#%60OQ-SBX-23%60), [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60), [`OQ-SBX-52`](agent-open-questions.md#%60OQ-SBX-52%60). Repository profile self-grants and credential domain bounds remain on hold under [`OQ-SBX-2`](agent-open-questions.md#%60OQ-SBX-2%60). [Sandbox Secret Injection](https://modal.com/docs/guide/sandbox-secret-injection), [Sandboxes](https://modal.com/docs/guide/sandboxes).

Modal supports authenticated Sandbox Connect Tokens and public port tunnels, but tunnels without encryption omit enforcement of the inbound address allowlist. A sidecar can connect to the token-protected port directly and supply its identity header. These paths don't replace afakt's [Port discovery and records](agent-ports.md#Port%20discovery%20and%20records). Preview and session transport remain under [`OQ-SBX-36`](agent-open-questions.md#%60OQ-SBX-36%60). [Networking and security](https://modal.com/docs/guide/sandbox-networking), [Sandbox Sidecars](https://modal.com/docs/guide/sandbox-sidecars).

## Workspace, scratch, and caches

The agent workspace needs an independent writable tree, separate scratch, and a verified image cache outside launch-writable content. Provider backing, permissions, capture access, and cleanup remain under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). Shared named tool cache access and poisoning remain under [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). Hostile repository metadata during capture remains under [`OQ-WSP-18`](agent-open-questions.md#%60OQ-WSP-18%60). See [Sandbox variables](agent-sandbox.md#Sandbox%20variables) and [Agent workspace capture](agent-workspace.md#Agent%20workspace%20capture).

Modal exposes filesystem reads and writes and can mount Volumes or CloudBucketMounts across sandboxes. Volume changes persist through background and final commits. Readers need explicit reloads to see later changes. Modal also mounts a volume subdirectory on its own. How afakt maps per-launch ownership, scratch cleanup and named-cache sharing onto Modal volumes is undecided under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60), [`OQ-WSP-19`](agent-open-questions.md#%60OQ-WSP-19%60). [Filesystem access](https://modal.com/docs/guide/sandbox-files).

## Workspace image materialization

[Agent workspace lifecycle](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage) connects materialization to capture and cleanup.

Agent workspace images contain directory trees in Open Container Initiative (OCI) form. [Agent host caches and startup](agent-workspace-image.md#Agent%20host%20caches%20and%20startup) requires content verification, confined extraction, and independent writable materialization before hooks or the harness start. The remote storage and transfer mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). A vendor runtime image or saved environment configuration doesn't establish this workspace contract.

Modal starts sandboxes from Images and can include files in an Image. Filesystem snapshots also become reusable Images, without establishing verification or ownership of a separate afakt workspace image. The transfer and materialization mapping remains under [`OQ-WSP-20`](agent-open-questions.md#%60OQ-WSP-20%60). [Sandboxes](https://modal.com/docs/guide/sandboxes), [Filesystem access](https://modal.com/docs/guide/sandbox-files), [Snapshots](https://modal.com/docs/guide/sandbox-snapshots).

## Agent snapshots and agent checkpoints

An agent snapshot captures virtual-machine state, while an agent checkpoint captures native harness session state. Using both requires pairing at an idle turn boundary under [agent checkpoints](agent.md#Agent%20checkpoints) and [suspend and resume](agent-sandbox-lifecycle.md#Suspend%20and%20resume). Provider capture scope and complete session restoration remain under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). Exclusion of active authority and capture authorization remain under [`OQ-AGT-16`](agent-open-questions.md#%60OQ-AGT-16%60), [`OQ-AGT-31`](agent-open-questions.md#%60OQ-AGT-31%60). By itself, vendor persistence doesn't establish an afakt agent snapshot or agent checkpoint.

Modal filesystem snapshots capture the root filesystem but exclude mounted Volumes. Directory snapshots capture one directory. Memory snapshots capture memory and filesystem state, stop the source, close network connections, and can't run during an active command. Background processes launched by another command aren't properly restored. Memory capture requires the same instance type on restore. It disallows region pinning and expires after seven days, while filesystem and directory snapshots default to thirty days with configurable retention. These limits leave afakt's complete paired capture under [`OQ-AGT-1`](agent-open-questions.md#%60OQ-AGT-1%60), [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60), and residency under [`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60). [Snapshots](https://modal.com/docs/guide/sandbox-snapshots).

Modal exit snapshots capture the main filesystem after graceful or ungraceful exits, but omit memory, mounts, and sidecars. After hardware failure, capture can fail without preserving state. They don't establish an agent snapshot or a complete checkpoint pair under [`OQ-AGT-3`](agent-open-questions.md#%60OQ-AGT-3%60). [Sandbox Exit Snapshots](https://modal.com/docs/guide/sandbox-exit-snapshots).

## Cancellation and teardown

The agent sandbox provider retains launch ownership until it observes termination of every process started for the launch. That observation must precede terminal status or slot release. [Cancellation and limits](agent-sandbox-lifecycle.md#Cancellation%20and%20limits) and [Reconciliation after restart or retry](agent-launch.md#Reconciliation%20after%20restart%20or%20retry) distinguish a stop request, an unavailable provider, and observed exit. Remote termination evidence, retries, detached work, and replacement safety remain under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). Credential retirement remains under [`OQ-SBX-40`](agent-open-questions.md#%60OQ-SBX-40%60). Session-end wait and identity lifetime remain on hold under [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Workload identities use the provisional lifetimes and half-life renewal in [Certificate policies](pki.md#Certificate%20policies), subject to [`OQ-HOOK-18`](agent-open-questions.md#%60OQ-HOOK-18%60).

Modal Sandboxes have a five-minute default lifetime and a twenty-four-hour upper limit. The lifecycle distinguishes creation, scheduling, startup, optional readiness, and finished state. Entrypoint exit, explicit termination, timeout, or memory exhaustion can finish a sandbox. Readiness failure doesn't stop it, and disconnecting the client doesn't stop detached work. Stopping the sandbox stops its sidecars. Mapping these observations to independent afakt termination evidence remains under [`OQ-SBX-66`](agent-open-questions.md#%60OQ-SBX-66%60). [Sandboxes](https://modal.com/docs/guide/sandboxes), [Sandbox Sidecars](https://modal.com/docs/guide/sandbox-sidecars).

## Host observations

The agent host worker needs launch-bound status, resource use, termination evidence, and capture results under [Provenance record](agent-launch-model.md#Provenance%20record). Vendor observations aren't independent proof against vendor compromise. [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) excludes secret values and workspace file contents from afakt telemetry. It doesn't constrain vendor custody. Vendor personnel access, internal copies, deletion completion, and independently protected evidence remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60).

Modal exposes command output, filesystem access, resource metrics, and exit reasons. Its secret-injection proxy decrypts matching HTTPS requests to inject headers outside the sandbox. These interfaces place workspace, output, and any supplied proxy secrets within vendor processing. The cited pages leave internal operator access limits, complete log retention, and erasure timing unstated, along with protection from vendor-side disclosure. Those gaps remain under [`OQ-SBX-64`](agent-open-questions.md#%60OQ-SBX-64%60). [Filesystem access](https://modal.com/docs/guide/sandbox-files), [Sandbox resources](https://modal.com/docs/guide/sandbox-resources), [Sandbox Secret Injection](https://modal.com/docs/guide/sandbox-secret-injection).

Modal keeps command output for up to ten minutes after a command completes, then discards it. [Running commands](https://modal.com/docs/guide/sandbox-spawn).

## Platform dependencies

The Modal Sandboxes provider depends on the selected runtime, management service, image storage, and any attached volumes, sidecars, or proxy. Modal documents resource upper limits and a fixed root disk for its VM runtime. A complete compute-region inventory remains absent from the cited sandbox guides, as do location guarantees for volumes, snapshots, and logs. Memory snapshot restoration runs wherever its instance type is available. Region, retention, and account eligibility remain under [`OQ-PLC-7`](agent-open-questions.md#%60OQ-PLC-7%60), while provider verification remains under [`OQ-PLC-9`](agent-open-questions.md#%60OQ-PLC-9%60). [Sandbox resources](https://modal.com/docs/guide/sandbox-resources), [Snapshots](https://modal.com/docs/guide/sandbox-snapshots).
