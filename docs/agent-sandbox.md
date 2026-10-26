---
type: Design
title: Agent sandbox
description: The sandbox configuration, sandbox variables, and credential brokering that confine an agent execution, and the execution token that grants its capabilities.
tags: [afakt, agents, execution, sandbox, credentials]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent sandbox

A durable [agent](agent.md) has its own sandbox configuration for its runs. afakt enforces one common sandbox capability across every supported harness. Sandbox settings and enforcement rules apply equally across supported harnesses and [harness transports](agent-profile.md#harness-transport). Both the [agent profile](agent-profile.md) and the [execution specification](agent-execution-contract.md#execution-specification) contain a sandbox configuration. At launch, the execution provider combines them into one combined sandbox configuration. A provider that can't enforce a setting rejects the launch instead of ignoring it.

The host agent provider is the default execution provider. Each agent runs in its own sandbox virtual machine (VM), which is the isolation boundary. The provider confines the whole harness process tree there, including the ACP adapter when present. Helpers and hooks run inside it too. On ACP launches, the adapter launch turns off the harness's built-in sandbox and command approval prompts. On headless launches, the harness integration does the same. afakt's sandbox replaces the built-in sandbox, and ACP permission requests replace native command approval prompts on ACP runs. Headless runs use the profile's fixed harness permissions without interactive prompts.

| Harness | Launch behavior |
| --- | --- |
| Claude Code | The adapter launch disables the built-in sandbox and native command approval prompts. afakt confines the process tree and uses ACP permissions. |
| Codex | afakt provides process confinement and ACP permissions. The adapter disables Codex's sandbox and native command approval prompts when it starts. |
| Other supported harnesses | Each integration applies the same replacement, through its adapter launch on ACP or its headless launch settings. |

Each harness integration defines its launch settings. Profile tool grants and harness permissions still apply. For ACP executions, the relay answers permission requests from those grants. For interactive sessions, the writer answers them under the [session attachment rules](agent-session-attachment.md#requests-from-the-agent).

## Enforcement

Each sandbox VM has its own guest kernel and disks. Linux uses a Kernel-based Virtual Machine (KVM) microVM driven by `Cloud Hypervisor`. Windows uses the same Linux microVM inside Windows Subsystem for Linux (WSL2) by default. A Windows-native utility VM serves an agent that needs a Windows guest. macOS uses Virtualization.framework. Whether macOS uses Apple's Containerization framework and its `vminitd` guest init or afakt drives Virtualization.framework directly remains [open](agent-open-questions.md#agent-sandbox).

afakt runs rootless at runtime: a systemd user unit on Linux and a launchd agent on macOS. Setup privileges and service survival after logout remain [open](agent-open-questions.md#agent-sandbox).

The host sandbox is a weaker mode for hosts without virtualization. Enforcement follows the [Anthropic sandbox runtime](https://github.com/anthropics/sandbox-runtime), using Seatbelt profiles through `sandbox-exec` on macOS, and bubblewrap namespaces plus seccomp on Linux. Where bubblewrap isn't available, Linux uses Landlock with the network controls needed to enforce the same configuration. The provider never drops a setting when selecting an enforcement mechanism. If it can't enforce the configuration, it rejects the launch with `sandbox-unsupported`. Which deployments permit this mode remains an on-hold [operator decision](agent-open-questions.md#agent-sandbox).

### Guest layout and lifecycle control

Inside the guest, afakt's init runs as process ID (`PID`) 1, then the guest relay runs under its own user. The VM's vsock device belongs only to that user. A separate agent user runs the ACP adapter and harness and can write only to the workspace. The guest doesn't store long-lived secrets. On macOS, the init integration remains [open](agent-open-questions.md#agent-sandbox).

The guest relay handles the ACP session channel, the CLI and MCP control channel from inside the agent, port discovery, and preview connections to guest loopback listeners. It advertises no ACP file system or terminal capabilities. [Agent session](agent-session.md#relay-behavior) defines permission answers, buffering, and reconnect behavior.

Start, suspend, resume, stop, `kill`, and snapshot operations go through the agent host's supervisor and hypervisor (`VMM`). They never go through ACP or the guest. No guest or session policy can block lifecycle control. Filesystem freezing assists consistent checkpoint capture without granting the guest authority over host lifecycle operations.

## Per-agent mediator

Each agent's mediator, a user-space TCP/IP stack outside the guest, serves that sandbox VM's only network device.

The guest has no other route out.

Enforcement doesn't depend on tools honoring proxy environment variables. A confined process on the agent host is the leading option for the mediator. Whether it needs its own VM remains [open](agent-open-questions.md#agent-sandbox), with the host isolation baseline on hold. It runs separately from the supervisor and `VMM`. Its network stack belongs to that agent alone, so every guest can use the same addresses and loopback ports. The mediator listens on no agent host port: its outbound connections use ordinary ephemeral ports, and under the confined-process option it reaches the supervisor and guest relay through Unix sockets or inherited descriptors.

### Network policy

For the guest, the mediator resolves Domain Name System (DNS) queries and records which names led to which addresses. Cedar evaluates each connection against its destination address, port, and resolved name, and evaluates each HTTP request. Network access defaults to deny.

Unless policy permits it, the mediator denies User Datagram Protocol (UDP) and Internet Control Message Protocol (ICMP) traffic. Combined domain and resolved-address rules remain bounds on these decisions.

For inspected domains, Transport Layer Security (TLS) connections end at the mediator, which uses a per-agent certificate authority that the guest trusts. Its credential broker injects credentials into permitted requests there, keeping credential values outside the guest. [Credential brokering](#credential-brokering) defines the injection rules. Network, HTTP, ACP, and control decisions enter one decision log for the agent. This log excludes secrets and request and response content.

### Control requests

On CLI, MCP, and port exposure control requests arriving from the guest relay, the mediator stamps the agent's identity, execution ID, and generation. Inside the guest, there is no execution token. The guest can't choose another agent's identity. In both directions, the mediator's [ACP gate](agent-session.md#acp-gate) also checks every session message.

With `allowLocalBinding`, a process can listen only on loopback in the sandbox's own network view, never on the agent host's external interfaces. Local app traffic stays in that view. Traffic leaving it still crosses the mediator. Authorized previews connect to those listeners through the guest relay over the mediator's outbound preview connection. See [Agent ports and previews](agent-ports.md).

### MCP access

Inside the VM, the harness connects to the [Model Context Protocol (MCP) gateway](agent-mcp-gateway.md#endpoint-and-harness-integration) through the guest relay's control channel and mediator. The mediator supplies authenticated agent and execution identity, using the execution token outside the guest. The gateway enforces the allowlists and keeps upstream credentials outside the sandbox.

## Sandbox configuration

A sandbox configuration offers the settings of the [Anthropic sandbox runtime](https://github.com/anthropics/sandbox-runtime) under the same names and with the same meaning.

| Group                 | Settings                                                                                                                                                                                                                                                                                |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `network`             | `allowedDomains`, `deniedDomains`, `deniedDomainReasons`, `allowUnixSockets`, `allowLocalBinding`, `deniedResolvedAddresses`, `allowPlaintextHeaderSet`, `stripResponseHeaders`, `refuseOpaqueTunnels`, `requireHostMatch`, `ambiguousRequestTarget`, and `tlsTerminate.excludeDomains` |
| `filesystem`          | `denyRead`, `allowRead`, `allowWrite`, and `denyWrite`, including literal entries                                                                                                                                                                                                       |
| `ignoreViolations`    | Command patterns mapped to the paths whose violations the sandbox doesn't report                                                                                                                                                                                                        |
| `credentials`         | `envVars` and `files` entries that the sandbox withholds from the agent or masks                                                                                                                                                                                                        |
| `git`                 | `safeDirectories`                                                                                                                                                                                                                                                                       |
| Isolation relaxations | `enableWeakerNestedSandbox`, `allowNestedUserNamespaces`, `enableWeakerNetworkIsolation`, `allowAppleEvents`, `allowPty`, and `allowAllUnixSockets`                                                                                                                                     |

The precedence rules match the sandbox runtime:

- Network access starts denied. `deniedDomains` wins over `allowedDomains`.
- Reads start allowed. `allowRead` wins over `denyRead`, but a more specific `denyRead` entry still applies.
- Writes start denied. `denyWrite` wins over `allowWrite`.

Settings about the agent host rather than the execution belong to the agent host's own configuration. Examples are the paths to sandbox helper programs and the custody of the mediator's certificate authority. A sandbox configuration that sets them is invalid.

Relative paths in filesystem rules resolve against the workspace root, because the agent starts there.

## Sandbox variables

A sandbox rule may contain sandbox variables in the form `${name}`. The provider replaces each variable with its value before it combines the configurations. A variable outside this table fails the launch with `sandbox-variable-unknown`. `$$` stands for a literal `$`.

| Variable | Value |
| --- | --- |
| `${workspace}` | The absolute path of the workspace root. |
| `${workspace:<target path>}` | The absolute path of the workspace source at that target path. |
| `${repository:<target path>}` | The repository URL of the `repository` source at that target path, without its scheme, such as `git.example.com/team/app`. |
| `${repository-host:<target path>}` | The host name in the repository URL of the `repository` source at that target path. |
| `${branch:<target path>}` | The new branch name of the `repository` source at that target path. |
| `${scratch}` | The absolute path of the execution's scratch directory, outside the workspace. Cleanup removes it with the workspace. |
| `${home}` | The absolute path of the agent's home directory inside the sandbox. |
| `${cache:<name>}` | The path inside the guest for the agent host cache with that name, such as `go-build`. Every execution on the agent host that uses this variable with that name shares the cache. |
| `${afakt-api}` | The host name and port of the afakt API that the agent reaches through the credential broker. |
| `${telemetry}` | The host name and port of the agent host's telemetry endpoint, which receives the harness's OpenTelemetry data through the credential broker. See [Agent observability](agent-observability.md#trace-context-propagation). |
| `${execution-id}` and `${generation}` | The execution ID as a canonical UUIDv7, such as `019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b`, and the generation as a decimal integer. See [Execution ID format](agent-execution-model.md#execution-id-format). |
| `${profile-name}` and `${profile-version}` | The pinned profile's name and version. |

The provider inserts each value literally. A value that contains glob characters matches only itself, never as a pattern. Capability scopes, credential brokering domains, [hooks](agent-hooks.md), and [agent environment](agent-workspace.md#agent-environment) values accept the same variables.

## Combining configurations

Each profile's sandbox configuration is the base and the ceiling. A specification may tighten it, and it may loosen it only within the profile's own allow rules.

| Setting kind | Combined value |
| --- | --- |
| Deny lists: `deniedDomains`, `deniedResolvedAddresses`, `stripResponseHeaders`, `denyRead`, `denyWrite`, and `credentials` entries | Both lists joined. |
| Allow lists: `allowedDomains`, `allowUnixSockets`, `allowRead`, and `allowWrite` | Without a specification list, the profile's list. Otherwise the specification's list, where each entry must fall within an entry of the profile's list. |
| Isolation relaxations and `allowLocalBinding`, `allowPlaintextHeaderSet` | True when the profile sets it true and the specification doesn't set it false. |
| `refuseOpaqueTunnels` and `requireHostMatch` | True when either configuration sets it true. |
| `ambiguousRequestTarget` | Whichever value is stricter, where `refuse` is stricter than `normalize`, which is stricter than `asSpelled`. |
| `ignoreViolations`, `deniedDomainReasons`, `git.safeDirectories`, and `tlsTerminate.excludeDomains` | Only the profile's value. A specification that sets one is invalid. |

An entry falls within another entry when each path or domain it matches also matches the other entry. An entry that the provider can't confirm as within another counts as outside it.

After it combines the two configurations, the provider adds the rules that the execution itself needs:

- With `read-write` workspace access, `allowWrite` gains `${workspace}`. With `read-only` access, `denyWrite` gains `${workspace}`.
- `allowWrite` gains `${scratch}`.
- `allowedDomains` gains `${afakt-api}` and `${telemetry}`, and the mediator's per-agent certificate authority joins the sandbox's trusted certificates.

After variable replacement, the provider publishes the combined sandbox configuration as evidence with its digest. Both the configuration digest and the provenance record include that digest.

## Credential brokering

Inside the mediator, credential brokering takes credential references from both the profile and the execution. The profile's brokering entries describe the requests that receive each credential. Whether execution references must stay within those the profile permits remains [on hold](agent-open-questions.md#agent-sandbox). Its format starts from the credential configuration of Docker Sandboxes.

```yaml
credentials:
  - service: model-provider-api
    apiKey:
      name: MODEL_PROVIDER_API_KEY
      inject:
        - domain: api.example-provider.com
          header: Authorization
          format: "Bearer %s"
  - service: package-registry
    apiKey:
      name: REGISTRY_TOKEN
      inject:
        - domain: "*.registry.example.com"
          header: Authorization
          format: "Bearer %s"
```

| Field | Meaning |
| --- | --- |
| `service` | The credential reference, a stable name in the profile or execution that resolves through afakt's credential store. |
| `apiKey.name` | The environment variable that the agent sees. It holds the placeholder `afakt-brokered`, never the secret. |
| `apiKey.inject[].domain` | A domain whose requests receive the credential, with the same wildcards as `allowedDomains`. |
| `apiKey.inject[].header` | The request header that the broker writes. |
| `apiKey.inject[].format` | The header value, with `%s` standing for the secret. |
| `oauth` | In place of `apiKey`, a login credential held in afakt's credential store, the only refresh owner. The broker refreshes through that store and never passes the token to the agent. Entry fields and user mapping remain [open](agent-open-questions.md#agent-sandbox). |

The broker follows these rules:

- It writes the header only on requests to the listed domains, and it replaces any value the agent sent in that header.
- It writes headers only, never request bodies.
- Each inject domain must also match the combined `allowedDomains`. The broker writes a credential only on a request that it sends on over Transport Layer Security (TLS), never on a request without TLS. To write headers on encrypted requests, the broker decrypts the traffic for those domains, so `tlsTerminate.excludeDomains` can't list them.
- It resolves each credential reference through afakt's credential store at launch. Refresh ownership stays in that store. The broker sends service credentials only inside requests to the listed domains. Placement offers an execution only to agent hosts that can resolve every one of its credential references.
- It stops writing credentials when the execution enters a terminal state.

The broker evaluates every request against its execution identity and inject domain, including each request after TLS termination. It keeps secrets and signing material outside the sandbox. Brokering supports per-request scoping and an audit record for each injected use, with the execution ID, generation, credential reference, matched rule, time, and outcome. The audit contains no credential value, execution token, or request or response content. These controls enforce the configured inject rules and never grant a request that the network rules deny. Repository profile self-grant and credential domain bounds remain [on hold](agent-open-questions.md#agent-sandbox).

Credential brokering uses references on both the profile and the execution. The operator adds credentials through the [CLI authentication flows](agent.md#secrets-credentials-and-trust-material), isolated from their own harness sign-ins. Refresh ownership isn't shared with the operator's CLI. Profiles, execution specifications, sandbox configurations, provenance records, and evidence never contain a secret value.

## Execution token

For each execution, the afakt API is a brokered service. Its token belongs to the mediator or agent host, never the guest. When the provider launches an execution, the API server issues an execution token for it, a signed JSON Web Token (JWT) with the claims in this table.

| Claim | Value |
| --- | --- |
| Subject | Execution ID as a canonical UUIDv7 and the generation, which identify the run to the afakt API. |
| Agent profile | Source, name, version, and digest of the pinned profile. |
| Capabilities | Each of the execution's capabilities, with its scopes after variable replacement. See [Capabilities](agent-profile.md#capabilities). |
| Audience | Only the afakt API. |
| Issued at and expires at | Its lifetime, which ends no later than the execution's time limit plus its stop grace period. |

After stamping its identity, the mediator adds the token as a bearer token to each permitted request to `${afakt-api}`. Inside the sandbox, the afakt CLI uses the guest relay without holding a token. Before the token expires, the mediator renews it, up to the moment the execution enters a terminal state. Session-end wait and renewal bounds remain [open](agent-open-questions.md#agent-hooks).

For each request, the API server checks the token's capabilities and scopes, and it records the token's subject on each action it takes for the agent. Once the execution has entered a terminal state, the API server rejects its token, even before the token expires. Operations on external services, such as pull request comments and reviews, go through the afakt API under these capabilities. The gateway’s [built-in afakt MCP server](agent-mcp-gateway.md#built-in-afakt-mcp-server) applies the same execution authority to its API calls. Agent-owned credential references permit brokered use rather than exposing service credentials to the harness. How attachment during operation changes the profile's and execution's credential references remains [open](agent-open-questions.md#agents).
