---
type: Design
title: Transport certificates and keys
description: Overlay credentials, public and inspection certificates, and storage transport trust.
tags: [afakt, pki, security]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Transport certificates and keys

Transport Layer Security (TLS) authenticates service connections separately from app authorization. Workload identities use Secure Production Identity Framework for Everyone (SPIFFE) Verifiable Identity Documents (SVIDs). Their keys belong to the [workload inventory](pki-workload.md). The credentials below serve other transport purposes or endpoints whose workload integration remains undecided.

## Overlay credentials

The cluster overlay uses isolated coordination, separate launch and service nodes, and relays. In the control plane, the coordination interface uses a presented SVID to register nodes and issue single-use tagged pre-authentication keys. It also sets tags and access rules and deletes nodes. See [Coordination interface](workload-overlay.md#Coordination%20interface). Designated Encrypted Relay for Packets (DERP) provides the relay service. Overlay high availability (HA), expiry behavior, and propagation remain [undecided](agent-open-questions.md#%60OQ-HST-10%60).

| Material and purpose | Private custody | Issuer, verifier, and trust anchor | Other undecided fields |
| --- | --- | --- | --- |
| `KT-01` Coordinator private keys. Authenticate overlay coordination. | Coordinator. | Coordinator generates keys that nodes verify. | Key kinds and initial coordinator anchor. |
| `KT-02` Launch node keys. Authenticate one launch overlay node. | Workload proxy outside the guest. | Keys originate at the node. Coordinator and overlay peers verify under node registration. | Node registration. |
| `KT-03` Host node keys. Authenticate host-level overlay membership. | Host-level node outside agent confinement. | Node generates. Coordinator and peers verify under host registration. | Host registration. |
| `KT-04` Control plane node keys. Authenticate each service instance. | Each service node. | Node generates. Service registration governs verification by the coordinator and peers. | Service registration. |
| `KT-05` Single-use pre-authentication keys. Enroll a node with authorized tags. | Coordination interface and enrolling node. | Coordination interface issues against an SVID. Coordinator verifies. | Bootstrap anchor, expiry, and invalidation. |
| `KT-06` DERP relay private key. Identify the relay protocol endpoint. | Relay. | Undecided. | Relay key generation, peer verification, and anchor. |
| `KT-07` DERP TLS certificate and key. Authenticate the relay transport endpoint. | Relay TLS terminator. | Undecided. | Certificate issuer, verifier profile, and transport anchor. |
| `KT-08` User-tailnet attachment node keys. Authenticate the external ingress or optional host attachment. | Attachment node. | Tailnet enrollment authority and peers. | Attachment trust. |

Storage, algorithm, key size, lifetime, and rotation remain [undecided](agent-open-questions.md#%60OQ-PKI-4%60) for every entry in this table, along with the fields in the last column. Revocation uses node removal or key retirement for all entries except `KT-05`, with [undecided propagation](agent-open-questions.md#%60OQ-HST-10%60). Each `KT-05` key permits one use.

The node-key class includes the transport and machine key material required by the selected overlay protocol. Its exact inventory and independent rotation remain [undecided](agent-open-questions.md#%60OQ-PKI-4%60). During coordination outages, the static map of stable nodes supplies addresses. It doesn't renew credentials or bypass service authentication. User-tailnet membership never grants afakt user authority or joins a launch to a user tailnet. See [Overlay transport](workload-overlay.md#Overlay%20transport).

## External ingress gateway

The external ingress gateway uses Traefik with afakt authentication middleware under [External ingress gateway](workload-overlay.md#External%20ingress%20gateway). Its middleware integration remains undecided in [`OQ-SES-13`](agent-open-questions.md#%60OQ-SES-13%60).

Traefik provides public TLS and Automated Certificate Management Environment (ACME) integration for the external ingress gateway. Its public certificates serve the app and preview names. Public certificate trust remains distinct from internal workload trust. ACME wildcard issuance requires Domain Name System (DNS) validation. A wildcard covers one label beneath its domain. It doesn't cover the domain apex or deeper descendants. The [preview naming scheme](workload-previews.md#Preview%20routing%20and%20names) uses one label beneath the preview domain.

| Material and purpose | Private custody | Issuer, verifier, and trust anchor | Other undecided fields |
| --- | --- | --- | --- |
| `KT-09` ACME account key. Allow issuance for the external ingress gateway. | Traefik ACME issuer or delegated certificate controller. | Account authority verifies account-key proof. | Ownership, storage, account authority, and bootstrap trust. |
| `KT-10` External ingress certificate and key. Authenticate app host names. | Traefik external ingress gateway. | Public issuing authority. External clients verify their configured public roots. | Storage and replica distribution. |
| `KT-11` Preview wildcard certificate and key. Authenticate names beneath the preview domain. | Same gateway as `KT-10`. | Public issuing authority after DNS validation. Clients verify public roots and host name. | Same fields as `KT-10`. |
| `KT-12` DNS challenge credential. Allow domain-validation record changes. | Public certificate issuer. | DNS provider issues and verifies. Provider account is the authority. | Storage. |
| `KT-13` User-tailnet public certificate and key. Authenticate the external ingress attachment name. | External ingress attachment. | Tailnet certificate facility and public clients. | Storage, issuer, and anchor. |
| `KT-14` Public tunnel connector credential. Authenticate the outbound publication connector. | Control plane outside workload hosts and guests. | Tunnel provider issues the credential and verifies it under the authority over the workload provider's account. | Protected storage. |
| `KT-15` Tunnel edge certificate and key. Authenticate provider-hosted public preview names. | Tunnel provider. | Provider certificate authority issues. Public roots support verification by external clients. | Custody contract. |

Algorithm, key size, lifetime, rotation, revocation or substitute, and authority selection remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60) for every entry in this table. The same question covers the other fields in the last column.

Traefik's local ACME store retains account and certificate private material. A shared writable store doesn't coordinate concurrent issuers across replicas. Issuer ownership, protected storage, renewal failover, certificate distribution, and backup remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60). Certificate consumers and the component authorized to issue certificates need separate permissions. The public tunnel's edge certificate and the external ingress certificate can apply at separate TLS termination points. Internal connections require workload identity regardless of these certificates.

## Inspection certificates

The workload proxy is the TLS endpoint for inspected connections after the workload proxy mediator admits the destination. Guest clients trust the per-workload certificate authority (CA). Signing material remains outside the sandbox, with custody defined by the workload host configuration. Per-Server Name Indication (SNI) certificates authenticate inspected destinations. Their creation from a destination inventory or on demand remains [undecided](agent-open-questions.md#%60OQ-SBX-38%60). See [Network policy](workload-proxy-mediator.md#Network%20policy).

| Material and purpose | Private custody | Issuer, verifier, and trust anchor | Other undecided fields |
| --- | --- | --- | --- |
| `KT-16` Inspection CA certificate and key. Establish guest trust for inspected destinations. | Workload host outside the guest. | Guest clients trust the installed public inspection authority. | Exact process, storage, parent or self-issued anchor, and revocation or substitute. |
| `KT-17` Per-SNI inspection leaf and key. Authenticate an admitted inspected destination. | Workload proxy outside the guest. | Inspection CA issues. Guest clients verify the chain and destination name against inspection trust. | Process and storage. |
| `KT-18` Guest inspection trust bundle. Distribute public inspection trust to guest clients. | Public material only. No private key. Workload provider controls installation. | Provider supplies public CA material. Guest TLS clients verify destination certificates. | Runtime stores, authentic installation, and propagation bound. |

Algorithm, key size, lifetime, and rotation remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60) for these entries, as do the fields in the last column. For `KT-18`, algorithm and size refer to the underlying public certificates. Revocation for `KT-18` removes retired roots and reloads client contexts. Destination removal and active connection retirement for `KT-17` remain [undecided](agent-open-questions.md#%60OQ-SBX-38%60).

### Guest trust installation

The workload provider adds the inspection CA to the sandbox's trusted certificates under [Combining configurations](workload-sandbox.md#Combining%20configurations). Only public trust enters the guest. Runtime-private trust stores can differ from operating-system trust, and replacing a trust file doesn't close an established TLS connection or guarantee that a process reloads its verification context.

| Guest or confinement environment | Trust installation contract |
| --- | --- |
| Linux guest, including a Linux guest on macOS or Windows | Add inspection trust to the guest's public trust bundle and cover runtime-private stores. Exact installation, startup ordering, and reload remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60). |
| macOS host mode | Trust for confined processes, keychain scope, and separation from unrelated host clients remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60). A macOS host doesn't imply a macOS guest. |
| Linux or Windows host mode | Per-launch trust exposure and runtime coverage remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60). Launch admission doesn't imply machine-wide trust. |
| Containers and remote sandboxes | Container image trust, runtime-private bundles, provider injection, and removal remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60). Remote mediation also depends on [provider integration](agent-open-questions.md#%60OQ-SBX-36%60). |

Pinned certificates and clients that can't use required inspection trust don't gain a policy bypass. Destination classification remains [undecided](agent-open-questions.md#%60OQ-SBX-24%60). Trust reload, connection retirement, and snapshot restoration need separate handling. A retained guest trust bundle doesn't restore launch authority.

## Storage and workflow transport

The registry checks scoped bearer tokens, while the object store has explicitly configured transport security and credentials. Both contracts leave their certificate profiles undecided. See [Registry authentication](agent-workspace-image.md#Registry%20authentication) and [Object store](agent-workspace-image.md#Object%20store). Database role mapping remains [undecided](agent-open-questions.md#%60OQ-SBX-45%60). Each trust profile below includes the issuer, verifier, anchor, names, and reuse of workload identity.

| Material and purpose | Prospective private custody |
| --- | --- |
| `KT-19` Registry TLS certificate and key. Authenticate registry transport. | Registry or its authenticated serving boundary. |
| `KT-20` Object-store TLS certificate and key. Authenticate the configured object endpoint. | Embedded or external object-store terminator. |
| `KT-21` Operational data-store TLS certificate and key. Authenticate database transport. | Data-store service. |
| `KT-22` Database client certificate and key. Authenticate a service to its database role. | Authorized control plane client. |
| `KT-23` Workflow service TLS certificate and key. Authenticate an external workflow service connection. | Workflow service. |
| `KT-24` Workflow client certificate and key. Authenticate a worker to the workflow service. | Worker. |
| `KT-25` Published service TLS certificate and key. Authenticate a published encrypted service. | Termination component. |

Private custody and storage remain undecided for every entry in this table, as do algorithm, key size, lifetime, rotation, trust profile, and revocation or substitute. These decisions belong to [the storage transport question](agent-open-questions.md#%60OQ-PKI-7%60) for `KT-19` through `KT-24` and [the published service question](agent-open-questions.md#%60OQ-HUB-10%60) for `KT-25`.

A local embedded service doesn't imply a separate TLS connection. The [transport question](agent-open-questions.md#%60OQ-PKI-7%60) distinguishes local access, external endpoints, and service workload identities. Supervised children that can't reload trust require an explicit [trust-update contract](agent-open-questions.md#%60OQ-SRV-2%60).
