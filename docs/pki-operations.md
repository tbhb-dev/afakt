---
type: Design
title: Public key infrastructure operations
description: Identity issuance, renewal, recovery, administrative operations, and threats.
tags: [afakt, pki, security]
status: draft
generated: { by: "codex:gpt-6-astra", at: "2026-10-27T00:00:00Z" }
---

# Public key infrastructure operations

Public key infrastructure (PKI) operations keep authentication, authorization, and retained verification evidence separate. The [authority hierarchy](pki.md#Authority%20hierarchy) defines the trust boundaries. Ownership of PKI administration remains [undecided](agent-open-questions.md#%60OQ-PKI-10%60). Commands, audit durability, recovery permissions, and alert thresholds also remain undecided.

## Bootstrap and issuance

The workload identity issuer needs a first trust anchor independent of workload credentials that it hasn't issued. First-run root setup, authenticated bundle installation, bootstrap-token delivery, and issuer data-store access remain [undecided](agent-open-questions.md#%60OQ-SBX-32%60). A downloaded public key set or a reachable enrollment endpoint doesn't create that first trust.

Host enrollment checks the admitted machine or key claims. The issuer then needs an authorized mapping from host-side identity requests to their workloads. The Workload Application Programming Interface (API) serves host-side processes and excludes the guest. Attestation policy, registration ownership, and renewal authorization remain [undecided](agent-open-questions.md#%60OQ-SBX-30%60). Consumer device enrollment has a separate [identity question](agent-open-questions.md#%60OQ-ATT-23%60).

Outside the guest, the agent proxy receives its launch's X.509 Secure Production Identity Framework for Everyone (SPIFFE) Verifiable Identity Document (SVID). Over mutual Transport Layer Security (TLS), a receiving service authenticates the identity and applies operation and target authorization. Credential exchange also checks the allowed scope and audience. Exchange may use a JavaScript Object Notation (JSON) Web Token (JWT) SVID. See [Workload identity](agent-host.md#Workload%20identity), [Launch authentication](agent-proxy.md#Agent%20launch%20authentication), and [Exchange admission](agent-open-questions.md#%60OQ-SBX-17%60).

## Renewal and authority rotation

The workload starting lifetimes are one hour for X.509-SVIDs, five minutes for JWT-SVIDs, and 24 hours for an issuing certificate authority (CA), with renewal at half-life. See [Certificate policies](pki.md#Certificate%20policies). These values remain subject to the held [identity lifetime decision](agent-open-questions.md#%60OQ-HOOK-18%60). They don't select overlap duration, retry bounds, validity limits, or a grace period after expiry. For a CA transition, consumers need replacement public trust before accepting identities from the replacement authority. During overlap, both authorities can authenticate identities. Authorization must keep the same scope. Removing an old root from a file alone doesn't retire initialized trust contexts or established connections. Bundle distribution, bundle age limits, consumer acknowledgement, overlap bounds, and final removal remain [undecided](agent-open-questions.md#%60OQ-SBX-41%60).

Key delivery and reload differ across identity consumers.

A supervised child that can't reload needs a [trust-update procedure](agent-open-questions.md#%60OQ-SRV-2%60), whose interruption bounds remain undecided. Inspection trust also needs [guest and runtime reload](pki-transport.md#Guest%20trust%20installation). Registry signing-key updates follow the control plane's public key set and reject unknown keys under [Registry authentication](agent-workspace-image.md#Registry%20authentication).

Retired message verification keys remain for as long as signed messages remain, but historical verification doesn't permit new signing with a retired private key. [Message signatures](agent-session-attachment.md#Message%20signatures) define signing-time verification. Compromise handling remains [undecided](agent-open-questions.md#%60OQ-AGT-21%60).

## Expiry and renewal failure

Credential expiry bounds authentication only when the consumer checks time, and an already authenticated stream can last beyond its certificate's expiry. Terminal launch checks prevent new authorized API work, while active stream retirement and exchanged-credential cancellation remain [undecided](agent-open-questions.md#%60OQ-SBX-40%60). An expired host registry lease stops new placement. It doesn't prove that a launch stopped.

| Failure | Remaining behavior and duration |
| --- | --- |
| Workload issuer unavailable | Already issued credentials keep their encoded validity. Cache acceptance, continued renewal, and how long an outage may last remain [undecided](agent-open-questions.md#%60OQ-SBX-32%60). Cached trust doesn't imply a new identity. |
| Credential exchange unavailable | Already exchanged credentials follow their consumer's checks. Expiry and launch-bound retirement remain [undecided](agent-open-questions.md#%60OQ-SBX-48%60). |
| Public certificate renewal fails | Existing certificates keep their encoded validity. Alert thresholds, failover, and behavior at expiry remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60). |
| Inspection trust update fails | Existing client contexts or connections can keep old trust. Readiness, restart, and retirement bounds remain [undecided](agent-open-questions.md#%60OQ-PKI-6%60). Required inspection doesn't become passthrough. |
| Overlay coordination unavailable | Nodes have a static map of stable addresses. Credential expiry, policy propagation, and how long an outage may last remain [undecided](agent-open-questions.md#%60OQ-HST-10%60). |
| Signing-key publication unavailable | Registry authentication rejects unknown keys. Acceptance of cached known keys and their age limit remain [undecided](agent-open-questions.md#%60OQ-PKI-3%60). |
| Host enrollment expires during sleep | Fresh enrollment and recovery authority remain [undecided](agent-open-questions.md#%60OQ-HST-4%60), and copied software state alone doesn't prove the same machine returned. |

The starting certificate lifetimes don't imply a numeric outage allowance. The validity left at failure depends on issuance time, verifier clock rules, and any independent revocation or launch-state check.

## Emergency revocation and disaster recovery

Emergency response distinguishes a stolen leaf key, a compromised host, a signing-key compromise, and loss of a root authority. Stopping issuance and rejecting new use affect distinct authority, as do withdrawing trust and retiring active connections. Revoking exchanged credentials affects distinct authority too. Their scope and propagation bounds remain [undecided](agent-open-questions.md#%60OQ-HST-6%60), and replacing a leaf certificate alone doesn't invalidate another unexpired certificate under a trusted issuer.

Root rollover needs an authenticated replacement anchor. It also needs a way to reject the retired anchor. If all cached identities expire, recovery can't require only those same identities. The independent recovery authority and restoration of issuer data-store access remain [undecided](agent-open-questions.md#%60OQ-SBX-32%60). Whether an emergency transition permits any overlap depends on [trust retirement](agent-open-questions.md#%60OQ-SBX-41%60).

CA backup and restore cover private signing state, registration permissions, public bundles, revocation state, and audit history. Encryption, custody, restore ordering, retained generations, and rollback rejection remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). A successful file restore doesn't prove revoked authority remains revoked. Secret-store recovery also depends on its [master-key contract](pki-signing.md#Secret%20storage).

## Issuance audit and monitoring

Each issuance needs an audit record tied to its authority, authenticated requester, target identity, credential class, public key identifier or certificate serial, requested and granted scope, validity interval, time, and outcome. Private keys, bearer tokens, bootstrap tokens, and exchanged secret values never enter that record. Audit durability, denied-request coverage, retention, and reconciliation when signing succeeds but recording fails remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). [Telemetry exclusions](agent-observability.md#What%20telemetry%20never%20records) also apply to certificate operations.

| Operational signal | Metric meaning and unresolved policy |
| --- | --- |
| Issuance and renewal outcomes | Counts and duration by bounded credential class and outcome. Metric names and retry alerts are [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Expiry margin | Remaining validity of leaf credentials, issuing authorities, and roots. Warning and critical thresholds are [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Trust distribution | Bundle age and consumer adoption of replacements or removals. Allowed age and acknowledgement deadline remain [undecided](agent-open-questions.md#%60OQ-SBX-41%60). |
| Public certificate renewal | Issuance errors, validation failures, and replica certificate state. Alerts and issuer failover remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60). |
| Revocation and audit | Propagation delay, rejected stale authority, and audit-write failures. Bounds and response policy remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |

Identity strings, certificate serials, public key identifiers, and credential references belong in access-controlled audit records rather than unbounded metric attributes. Metric schema and aggregation remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60).

## Administrative operations

The command-line interface (CLI) exposes certificate and key inspection through the [certificate commands](cli.md#Certificate), [SSH key commands](cli.md#SSH%20key), and [GPG key commands](cli.md#GPG%20key). Secure Shell (SSH) and GNU (GNU's Not Unix) Privacy Guard (GPG) keys have separate purposes, and existing trusted-certificate commands don't define CA administration.

| Operation | Command surface and authority |
| --- | --- |
| Inspect trust and expiry | `afakt certificate info` and `afakt certificate list` inspect trusted certificates. Inspection of issuer state, workload credentials, and public key sets is [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Rotate a certificate or signing key | Command names, selected authority, overlap preview, permissions, and completion evidence remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Export the trust bundle | Public-only export format, command name, authenticated distribution, and permissions are [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Issue a bootstrap token | Command name, issuing authority, enrollment scope, expiry, consumption, and secret delivery remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| Revoke authority and recover | Commands, incident scope, approval authority, audit, and observable completion remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |

## Failure procedures

### Workload renewal failure

Find the failing issuer or consumer, validity left, launch state, and accepted trust before assessing the affected work. Recovery permissions and retry bounds are [undecided](agent-open-questions.md#%60OQ-SBX-32%60).

### Public certificate renewal failure

Find the affected name, validation authority, active certificate, and certificate state on each external ingress replica. Issuer failover and repair authority remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60).

### Inspection trust mismatch

Distinguish a missing guest root, stale runtime trust, a pinned destination, and an already established connection. Guest reload and restart procedures are [undecided](agent-open-questions.md#%60OQ-PKI-6%60).

### Host or leaf-key compromise

Find affected registrations, launch identities, exchanged credentials, and active connections. Revocation completion and required independent evidence remain [undecided](agent-open-questions.md#%60OQ-SBX-37%60).

### Issuing authority or root compromise

Find the affected trust boundary and the authority for distributing a replacement anchor. Emergency overlap, removal acknowledgement, and recovery remain [undecided](agent-open-questions.md#%60OQ-SBX-41%60).

### Signing-key compromise

Separate rejection of new signatures from interpretation of retained signatures. Publication withdrawal, historical validity, and release recovery remain [undecided](agent-open-questions.md#%60OQ-PKI-8%60).

### Restore after authority-state loss

Check the recovered key state, registration permissions, revocations, and audit history before admitting renewed authority. Recovery permissions and rollback protection remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60).

## Threat model

The PKI threat model extends the [agent threat model](agent-threat-model.md), with private issuance keys, enrollment authority, trusted public bundles, secret-store keys, and verification history as its assets. Attackers include a hostile guest, a compromised worker or host, a network attacker, and a compromised signing or recovery authority. STRIDE classifies spoofing, tampering, repudiation, information disclosure, denial of service, and elevation of privilege.

| Threat ID | Category | Threat | Mitigation and remaining gap |
| --- | --- | --- | --- |
| `TM-PKI-1` | Spoofing | An attacker substitutes the first trust anchor. | Bootstrap needs an independent trust boundary whose authority and installation remain [undecided](agent-open-questions.md#%60OQ-SBX-32%60). |
| `TM-PKI-2` | Elevation of privilege | A compromised host requests identities for other hosts or launches. | Guest exclusion from the Workload API limits guest requests. Host registration and broker scope remain [undecided](agent-open-questions.md#%60OQ-SBX-31%60). |
| `TM-PKI-3` | Information disclosure | A copied enrollment or workload key carries authority outside its intended machine. | Private launch material stays outside the guest. Hardware custody and non-exportability requirements remain [undecided](agent-open-questions.md#%60OQ-SBX-42%60). |
| `TM-PKI-4` | Tampering | A stale bundle accepts a retired authority after rollover. | Consumers need trust removal and context reload. Age limits and removal acknowledgement remain [undecided](agent-open-questions.md#%60OQ-SBX-41%60). |
| `TM-PKI-5` | Elevation of privilege | A verifier accepts a valid signature for the wrong identity, audience, or certificate purpose. | Receiving services check launch authority. Registry tokens have audience and scope checks. Complete certificate-profile validation remains [undecided](agent-open-questions.md#%60OQ-SBX-46%60). |
| `TM-PKI-6` | Spoofing | An attacker replays a stolen token for repeated credential exchange. | Audience and validity checks narrow use. Replay resistance and issuance limits remain [undecided](agent-open-questions.md#%60OQ-SBX-50%60). |
| `TM-PKI-7` | Elevation of privilege | Expired or terminal launch authority survives in an active stream or exchanged credential. | API requests reject terminal launches. Downstream retirement and expiry coupling remain [undecided](agent-open-questions.md#%60OQ-SBX-40%60). |
| `TM-PKI-8` | Denial of service | Issuance failure or renewal load causes widespread expiry. | Starting renewal at half-life leaves time for recovery. Retry, rate, and outage bounds remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| `TM-PKI-9` | Repudiation | A signing authority issues credentials without durable audit evidence. | Issuance always needs an audit record without secrets. Atomic recording, retention, and independent integrity remain [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| `TM-PKI-10` | Tampering | Restoring an old backup restores revoked authority. | Recovery must reconcile key and revocation state. Independent rollback detection remains [undecided](agent-open-questions.md#%60OQ-PKI-10%60). |
| `TM-PKI-11` | Information disclosure | Inspection or public certificate private keys leak through guests, backups, or replica distribution. | Inspection keys remain outside the guest. Public certificate distribution and backup custody remain [undecided](agent-open-questions.md#%60OQ-PKI-5%60). |
| `TM-PKI-12` | Tampering | Compromised release authority signs malicious executable or captured content. | Content digests check byte identity but don't establish trusted publication. Signing authority, verification, and withdrawal remain [undecided](agent-open-questions.md#%60OQ-PKI-8%60). |
| `TM-PKI-13` | Information disclosure | A secret-store backup exposes master keys and encrypted secret values together. | Secret references keep values out of launch records. Backup separation and encryption authority remain [undecided](agent-open-questions.md#%60OQ-PKI-9%60). |
| `TM-PKI-14` | Spoofing | An attacker replays a valid webhook signature or a consumer accepts it for a different destination. | Webhooks need provider-specific signing and replay checks. The binding and rotation contract remains [undecided](agent-open-questions.md#%60OQ-PKI-11%60). |

Compromise of a root or trusted host crosses the corresponding authentication or enforcement boundary. Short certificate lifetimes reduce the validity interval of issued credentials. They don't prove immediate revocation, host integrity, or recovery from a compromised issuer.
