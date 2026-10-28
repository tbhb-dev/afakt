---
type: Design
title: Workload sandbox
description: The sandbox configuration, sandbox variables, credential brokering, and authentication of an agent launch.
tags: [afakt, agents, agent-launch, sandbox, credentials]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Workload sandbox

[Public key infrastructure](pki.md).

A workload sandbox is an `afakt-sandbox` process and the isolation boundary it applies. For an agent, it confines the harness as the workload process. The agent spec contains the desired sandbox configuration of a durable [agent](agent.md). afakt enforces one common sandbox capability across every supported harness. Sandbox settings and enforcement rules apply equally across supported harnesses and [harness transports](agent-profile.md#Harness%20transport). Both the [agent profile](agent-profile.md) and the [agent launch spec](agent-launch-contract.md#Agent%20launch%20spec) contain a sandbox configuration. At launch, the workload provider builds one sandbox configuration from the profile and launch spec. A provider that can't enforce a setting rejects the launch instead of ignoring it.

The VM provider is the workload provider used by default. It provisions, restores, and releases agent sandboxes. Each agent runs in its own workload guest, a virtual machine (VM), which is the isolation boundary for the whole harness process tree, including the Agent Client Protocol (ACP) adapter when present, helpers, and hooks. On ACP launches, the adapter launch turns off the harness's built-in sandbox and command approval prompts. On headless launches, the harness integration does the same. afakt supplies confinement. ACP permission requests replace native command approval prompts on ACP agent launches, while headless runs use the profile's fixed harness permissions without interactive prompts.

| Harness | Launch behavior |
| --- | --- |
| Claude Code | The adapter launch disables the built-in sandbox and native command approval prompts. afakt confines the process tree and uses ACP permissions. |
| Codex | afakt provides process confinement and ACP permissions. The adapter disables Codex's sandbox and native command approval prompts when it starts. |
| Other supported harnesses | Each integration applies the same replacement, through its adapter launch on ACP or its headless launch settings. |

Each harness integration defines its launch settings. Profile tool grants and harness permissions still apply. For ACP agent launches, the relay answers permission requests from those grants. For interactive sessions, the writer answers them under the [session attachment rules](agent-acp.md#Requests%20from%20the%20agent).

[Attachment and provider storage](agent-workspace-lifecycle.md#Attachment%20and%20provider%20storage).

## Enforcement

### Sandbox modes

afakt permits all the [sandbox modes](workload-host.md#Runtime%20privileges%20and%20platforms). The local workload provider kinds are container manager providers (`docker-desktop` and `docker-engine`), VM providers (`macos-vm`, `windows-vm`, and `linux-vm`), and host providers (`macos-host`, `windows-host`, and `linux-host`). Docker Desktop uses the Docker VM on macOS and Windows, with WSL2 or Hyper-V on Windows. Docker Engine supports amd64 Linux bare metal and VMs, and arm64 Linux VMs.

Hosted agent providers include [Claude Managed Agents](workload-provider-claude-managed-agents.md) and [OpenAI Agents API](workload-provider-openai-agents-api.md). Remote sandbox providers include [Modal Sandboxes](workload-provider-modal.md) and [Cloudflare Sandbox SDK](workload-provider-cloudflare.md).

The host modes are the weaker isolation tier.

Each workload guest has its own guest kernel and disks. Linux uses a Kernel-based Virtual Machine (KVM) microVM driven by `Cloud Hypervisor`. Windows uses the same Linux microVM inside Windows Subsystem for Linux (WSL2) by default. A Windows-native utility VM serves an agent that needs a Windows guest. macOS uses Virtualization.framework. Whether macOS uses Apple's Containerization framework and its `vminitd` guest init or afakt drives Virtualization.framework directly remains [open](agent-open-questions.md#Agent%20sandbox).

### Runtime privileges

The workload host worker uses a dedicated user. It never runs as root. Host-mode agents use a dedicated user separate from the worker, under [Workload host access](workload-overlay.md#Workload%20host%20access). Runtime privileges follow the selected [provider boundary](workload-provider-linux-host.md#Platforms%20and%20privileges). Privileged setup and platform capture services remain provider-specific. Container manager deployments use rootless Docker. afakt uses a systemd user unit on Linux and a launchd agent on macOS. Setup privileges and service survival after logout remain [open](agent-open-questions.md#Agent%20sandbox). Running each agent launch under its own User Identifier (UID), separate from the worker user, remains [open](agent-open-questions.md#%60OQ-SBX-18%60).

### Host process sandbox

afakt-agent is afakt's cross-harness runner. For locally controlled harness execution, it starts afakt-sandbox, afakt's cross-platform process sandbox, and runs the agent harness inside it. In host modes, afakt-agent runs directly on the workload host. afakt-sandbox also runs on its own, without afakt-agent, to confine a command that `afakt agent exec` runs in an agent's workspace. Enforcement follows the [Anthropic sandbox runtime](https://github.com/anthropics/sandbox-runtime), using Seatbelt profiles through `sandbox-exec` on macOS, and bubblewrap namespaces plus seccomp on Linux. The [Linux host boundary](workload-provider-linux-host.md#Isolation%20boundary) uses bubblewrap, Landlock, seccomp, and the workload provider's network isolation together. The provider never drops a setting when selecting an enforcement mechanism. If it can't enforce the configuration, it rejects the launch with `sandbox-unsupported`.

### Guest layout and lifecycle control

In a Linux guest, afakt's init runs as process ID (`PID`) 1, then the guest relay runs under its own user. Access to the relay carrier follows the selected [provider boundary](workload-provider-linux-vm.md#Mediator%20connection%20and%20name%20resolution). The general carrier choice remains [`OQ-SES-5`](agent-open-questions.md#%60OQ-SES-5%60). A separate agent user runs the ACP adapter and agent harness. Its writable paths are the permitted workspace, scratch, and cache paths under [Combining configurations](workload-sandbox.md#Combining%20configurations). No long-lived secrets enter guest storage. On macOS, the init integration remains [open](agent-open-questions.md#Agent%20sandbox).

The guest relay handles the ACP session channel, the command-line interface (CLI) and MCP control channel from inside the agent, port discovery, and preview connections to guest loopback listeners. It advertises no ACP file system or terminal capabilities. [Agent session](agent-acp.md#Relay%20behavior) defines permission answers, buffering, and reconnect behavior.

[Startup and readiness](workload-lifecycle.md#Startup%20and%20readiness) use the independent lifecycle path. Guest traffic follows the [workload proxy mediator](workload-proxy-mediator.md) policy.

## Sandbox configuration

An agent sandbox configuration uses the filesystem, credential masking, and process restrictions of the [Anthropic sandbox runtime](https://github.com/anthropics/sandbox-runtime). Ingress and egress restrictions belong to the [workload firewall](workload.md#Firewall). Its [resource format](agent-open-questions.md#%60OQ-SBX-67%60) remains open, including in host modes where capture sends traffic through the same enforcement point at the workload proxy.

| Group                 | Settings                                                                                                                                                                                                                                                                                |
| --------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Local sockets and binding | `allowUnixSockets` and `allowLocalBinding` |
| `filesystem`          | `denyRead`, `allowRead`, `allowWrite`, and `denyWrite`, including literal entries                                                                                                                                                                                                       |
| `ignoreViolations`    | Command patterns mapped to the paths whose violations the sandbox doesn't report                                                                                                                                                                                                        |
| `credentials`         | `envVars` and `files` entries that the sandbox withholds from the agent or masks                                                                                                                                                                                                        |
| `git`                 | `safeDirectories`                                                                                                                                                                                                                                                                       |
| Isolation relaxations | `enableWeakerNestedSandbox`, `allowNestedUserNamespaces`, `enableWeakerNetworkIsolation`, `allowAppleEvents`, `allowPty`, and `allowAllUnixSockets`                                                                                                                                     |

The precedence rules match the sandbox runtime:

- Firewall egress defaults to a cluster-managed deny-all policy.
- Reads start allowed. `allowRead` wins over `denyRead`, but a more specific `denyRead` entry still applies.
- Writes start denied. `denyWrite` wins over `allowWrite`.

Settings about the workload host rather than the agent launch belong to the workload host's own configuration. Examples are the paths to sandbox helper programs and the custody of the workload proxy's certificate authority. A sandbox configuration that sets them is invalid.

Relative paths in filesystem rules resolve against the workspace root, because the agent starts there.

## Sandbox variables

An agent sandbox rule may contain sandbox variables in the form `${name}`. The workload provider replaces each variable with its value before it combines the configurations. A variable outside this table fails the launch with `sandbox-variable-unknown`. `$$` stands for a literal `$`.

| Variable | Value |
| --- | --- |
| `${workspace}` | The absolute path of the workspace root. |
| `${workspace:<target path>}` | The absolute path of the workspace source at that target path. |
| `${repository:<target path>}` | The repository uniform resource locator (URL) of the `repository` source at that target path, without its scheme, such as `git.example.com/team/app`. |
| `${repository-host:<target path>}` | The host name in the repository URL of the `repository` source at that target path. |
| `${branch:<target path>}` | The new branch name of the `repository` source at that target path. |
| `${scratch}` | The absolute path of the agent launch's scratch directory, outside the agent workspace. Cleanup removes it with the workspace. |
| `${home}` | The absolute path of the agent's home directory inside the sandbox. |
| `${cache:<name>}` | The path inside the guest for the workload host cache with that name, such as `go-build`. Every agent launch on the workload host that uses this variable with that name shares the cache. |
| `${afakt-api}` | The host name and port of the afakt Application Programming Interface (API) that the agent reaches through the guest relay and workload proxy under mediator decisions. |
| `${telemetry}` | The host name and port of the workload host's telemetry endpoint, which receives the agent harness's OpenTelemetry data through the workload proxy under mediator decisions. See [Agent observability](agent-observability.md#Trace%20context%20propagation). |
| `${agent-launch-id}` and `${generation}` | The canonical agent launch ID and the agent spec generation recorded at finalization, as a decimal integer. See [Agent launch ID format](agent-launch-model.md#Agent%20launch%20ID%20format). |
| `${profile-name}` and `${profile-version}` | The pinned profile's name and version. |

The provider inserts each value literally. A value that contains glob characters matches only itself, never as a pattern. Capability scopes, credential brokering domains, [hooks](agent-hooks.md), and [workload environment](agent-workspace.md#Workload%20environment) values accept the same variables.

## Combining configurations

Each profile's agent sandbox configuration is the base and the ceiling. A specification may tighten it, and it may loosen it only within the profile's own allow rules.

| Setting kind | Combined value |
| --- | --- |
| Deny lists: `denyRead`, `denyWrite`, and `credentials` entries | Both lists joined. |
| Allow lists: `allowUnixSockets`, `allowRead`, and `allowWrite` | Without a specification list, the profile's list. Otherwise the specification's list, where each entry must fall within an entry of the profile's list. |
| Isolation relaxations and `allowLocalBinding` | True when the profile sets it true and the specification doesn't set it false. |
| `ignoreViolations`, `git.safeDirectories` | Only the profile's value. A specification that sets one is invalid. |

An entry falls within another entry when each path or domain it matches also matches the other entry. An entry that the workload provider can't confirm as within another counts as outside it.

After it combines the two configurations, the provider adds the rules that the agent launch itself needs:

- With `read-write` workspace access, `allowWrite` gains `${workspace}`. With `read-only` access, `denyWrite` gains `${workspace}`.
- `allowWrite` gains `${scratch}`.
- The workload proxy's per-workload certificate authority joins the sandbox's trusted certificates. Mediator connections to control plane services use the workload proxy under its fixed [control plane rules](workload-proxy-mediator.md#Control%20plane%20connections).

After variable replacement, the provider publishes the sandbox configuration built from the profile and launch spec as evidence with its digest. Both the configuration digest and the provenance record include that digest.

## Credential brokering

The credential broker is a host-side component outside the guest. Its placement relative to the workload proxy and workload proxy mediator remains [open](agent-open-questions.md#%60OQ-SBX-23%60). It takes credential references from both the profile and the agent launch. Brokering entries in the profile describe the requests that receive each credential. Whether agent launch references must stay within those the profile permits remains [on hold](agent-open-questions.md#Agent%20sandbox). Each credential type defines which material the control plane keeps and which material the host-side credential broker receives.

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
| `service` | The credential reference, a stable name in the profile or agent launch that resolves through afakt's credential store. |
| `apiKey.name` | The environment variable that the agent sees. It holds the placeholder `afakt-brokered`, never the secret. |
| `apiKey.inject[].domain` | A domain whose requests receive the credential, using domain patterns. |
| `apiKey.inject[].header` | The request header that the broker writes. |
| `apiKey.inject[].format` | The header value, with `%s` standing for the secret. |
| `oauth` | In place of `apiKey`, a login credential held in afakt's credential store, the only refresh owner. The broker refreshes through that store and never passes the token to the agent. Entry fields and user mapping remain [open](agent-open-questions.md#Agent%20sandbox). |

The broker follows these rules:

- It writes the header only on requests to the listed domains, and it replaces any value the agent sent in that header.
- It writes headers only, never request bodies.
- The workload firewall must also permit each inject domain. Only a request that the broker sends on over Transport Layer Security (TLS) receives a credential, never a request without TLS. Credential injection requires the inspected path under [Network policy](workload-proxy-mediator.md#Network%20policy).
- It resolves each credential reference through afakt's credential store at launch. Refresh ownership stays in that store. The broker sends service credentials only inside requests to the listed domains. Placement offers an agent launch only to workload hosts that can resolve every one of its credential references.
- It stops writing credentials when the agent launch enters a terminal state.

For every request, including each request after TLS termination, the broker checks the request against its agent launch identity and inject domain. It keeps secrets and signing material outside the agent sandbox. Brokering supports per-request scoping and an audit record for each injected use, with the agent launch ID, credential reference, matched rule, time, and outcome. Authentication material and request and response content never appear in the audit. These controls enforce the configured inject rules and never grant a request that the network rules deny. Repository profile self-grant and credential domain bounds remain [on hold](agent-open-questions.md#Agent%20sandbox).

Credential brokering uses references on both the profile and the agent launch. The operator adds credentials through the [CLI authentication flows](agent.md#Secrets,%20credentials,%20and%20trust%20material), isolated from their own agent harness sign-ins. Refresh ownership isn't shared with the operator's CLI. Profiles, agent launch specs, sandbox configurations, provenance records, and evidence never contain a secret value.

### Credential types and custody

Each credential type defines custody and exchange. Exchangeable credentials stay in the control plane, and the host-side credential broker receives short-lived credentials issued in exchange. Examples include an Open Authorization (OAuth) refresh token such as a Codex login, a GitHub App key, and cloud role credentials. For one agent launch, a static secret that the control plane can't exchange stays in the credential broker's memory outside the guest. The broker doesn't deliver the secret directly to the guest. Preventing a remote service from reflecting credentials in response content remains [`OQ-SBX-26`](agent-open-questions.md#%60OQ-SBX-26%60). See [Credential types](agent.md#Secrets,%20credentials,%20and%20trust%20material).

[Agent launch authentication](workload-proxy.md#Agent%20launch%20authentication).
