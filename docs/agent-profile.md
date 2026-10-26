---
type: Design
title: Agent profile
description: The versioned definition of an agent's behavior, which bounds what an execution specification may request and which a launch validates against.
tags: [afakt, agents, execution, profiles]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent profile

An agent profile defines how a durable [agent](agent.md) behaves and what it may do. The profile binding belongs to the agent, and each harness run pins its own exact profile content. Its fields cover the agent's instruction template and workspace access, its tools, capabilities, sandbox, credentials, environment, and hooks, and the harnesses and models allowed to run it. Each [execution specification](agent-execution-contract.md#execution-specification) pins one exact profile version, and the execution provider validates the launch against that version before it starts anything.

## Profile sources

A profile comes from one of two sources.

| Source | Where it lives | Version |
| --- | --- | --- |
| `catalog` | The profile catalog, the central store of published profile versions. | A positive integer. A new version has a higher number than the version before it. |
| `repository` | A YAML file at `.afakt/agent-profiles/<name>.yaml` in a source repository. See [Repository profile files](#repository-profile-files). | The repository and the exact commit that contains the file. |

A profile name is unique within its source. A repository profile can't shadow a catalog profile with the same name, or the reverse. A catalog profile named `review` and a repository profile named `review` are different profiles. An execution specification states which source it means.

## Profile identity

| Field | Meaning |
| --- | --- |
| Source | `catalog` or `repository`. |
| Name | A stable label such as `build` or `review`: 1 to 64 lowercase letters, digits, and `-`. The first character is a letter. For a repository profile, the name is the filename without `.yaml`. |
| Version | The catalog version number, or the repository and commit. |
| Profile digest | The digest of the version's canonical encoding, including the content of each file it references. One source, name, and version always map to one digest. A different digest for them means corruption or tampering. |
| Status | `active` or `retired`, for catalog profiles. See [Profile changes](#profile-changes). |
| Description | A sentence about the agent's purpose, for people choosing a profile. |
| Labels | Key and value pairs for finding and grouping profiles, such as `team: platform`. They don't change how the agent runs. |

A published catalog version never changes, and neither do its labels. A repository profile version never changes either, because the content of a commit can't change.

## Profile resource

A profile has one resource form, in the style of a Kubernetes resource. Repository profile files use it. Catalog publish requests use it too, and afakt returns it whenever it reads a profile back.

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | The schema version of the resource, `afakt/v1alpha1`. afakt reads each schema version it supports and rejects any other. | The author |
| `kind` | `AgentProfile`. | The author |
| `metadata.name` | The profile name. In a repository profile file, it equals the filename without `.yaml`. | The author |
| `metadata.labels` | The profile's labels, as a map of strings. Optional. | The author |
| `metadata.source` | `catalog` or `repository`. | afakt |
| `metadata.version` | For a catalog profile, its version number. For a repository profile, the repository URL without its scheme, then `@`, then the full commit ID. | afakt |
| `metadata.digest` | The profile digest. | afakt |
| `spec` | The description and the [profile fields](#profile-fields), under the keys that [Repository profile files](#repository-profile-files) shows. | The author |
| `status.state` | `active` or `retired`, for a catalog profile. | afakt |

A file or a publish request that sets a key that afakt sets is invalid. The profile digest covers `apiVersion`, `kind`, `metadata.name`, `metadata.labels`, and `spec`, with the content of each file that `spec` references. It never covers `status` or the keys that afakt sets.

```yaml
apiVersion: afakt/v1alpha1
kind: AgentProfile
metadata:
  name: build
  labels: { team: platform }
  source: catalog
  version: 7
  digest: sha256:3f5a…
spec:
  description: Implement an issue on a new branch and push it.
  # The profile fields follow, as in a repository profile file.
status:
  state: active
```

## Profile fields

Each profile version fixes these fields.

| Field                | Meaning                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Instructions         | A Go template that renders to the text the harness gives the agent before the execution's prompt. See [Instruction templates](#instruction-templates).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Input schema         | A JSON Schema for the instruction inputs the template uses. Without one, the profile doesn't accept instruction inputs.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| Workspace access     | `read-only` or `read-write`. A read-only agent can read the workspace and run permitted commands, but it can't change files in the workspace or in its source history.                                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| Models               | The entries a launch may run with. Each entry sets one `harness`, one `transport`, one `provider`, one `model`, and one `effort`, together with a boolean `default` and an integer `priority`. The `harness` value has a required `name` and an optional `version` constraint, such as `">=2.0"`. An entry without a `version` accepts any version of that harness. Exactly one entry has `default: true`. A profile that allows a model with another harness or effort level adds another entry for it. When afakt chooses among entries that match a request, it prefers the entry with the lowest `priority` number that has capacity for the launch. |
| Harness permissions  | What the harness lets the agent do in the workspace. See [Harness permissions](#harness-permissions).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Tools                | The harness-native tools offered to the agent, by tool name. Tools absent from the list are unavailable.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| MCP servers | The allowed upstream servers, selected from deployment configuration, with deny entries taking precedence. See [MCP access](#mcp-access). |
| MCP tool allowlists | Per-server tool allow and deny lists enforced by the gateway. See [MCP access](#mcp-access). |
| MCP approvals | Per-server tools requiring approval and the unattended allow-or-deny rule. See [MCP access](#mcp-access). |
| Capabilities         | The afakt API operations the agent may call through the afakt CLI or built-in MCP server, with their scopes. See [Capabilities](#capabilities).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| Sandbox              | The sandbox configuration that confines every execution of the profile. See [Agent sandbox](agent-sandbox.md).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          |
| Credential brokering | The profile's credential references, stored in `credentials` entries under `service`, and the requests that receive each credential. Execution specifications also contain credential references. The bounds between them remain [on hold](agent-open-questions.md#agent-sandbox). See [Credential brokering](agent-sandbox.md#credential-brokering).                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| Environment          | Environment variables for the agent and its hooks. The profile may lock any entry against replacement. See [Agent environment](agent-workspace.md#agent-environment).                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| Hooks                | Commands that run on hook events before, during, and after the agent's work, in every execution of the profile. See [Agent hooks](agent-hooks.md).                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| Telemetry content    | `omit` or `include`. With `include`, telemetry may record prompt and response content. Without it, `omit` applies. See [What telemetry never records](agent-observability.md#what-telemetry-never-records).                                                                                                                                                                                                                                                                                                                                                                                                                             |
| Output schema        | A JSON Schema for the structured payload, or a statement that the payload is plain text.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| Output schema policy | `fixed` or `per-execution`. A fixed schema applies to each execution of the version. With `per-execution`, the specification supplies the schema.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |
| Default limits       | The time limit, stop grace period, and resource limits a launch receives when it doesn't request its own.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| Limit ceilings       | The highest time limit, stop grace period, and resource limits a launch may request.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| Retention ceiling    | The longest retention period a launch may request.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |

### Harness transport

Each entry in the profile's `models` selects a harness transport, `acp` or `headless`. Without `transport`, an entry uses `acp` when the harness has an Agent Client Protocol (ACP) adapter and `headless` otherwise.

| Harness | ACP adapter | Default transport |
| --- | --- | --- |
| Claude Code | `claude-agent-acp` | `acp` |
| Codex | `codex-acp` | `acp` |
| Other harnesses | None | `headless` |

An entry may set `transport: headless` even when its harness has an adapter, such as for a feature available only in the headless form. A launch that selects `transport: acp` for a harness without an adapter fails with `harness-transport-unsupported`. The selected entry fixes the transport. The execution specification can't override it. Both transports support the same output schema validation, hooks, results, and evidence. See [Harness transport](agent-execution.md#harness-transport).

### Instruction templates

A profile's instructions are a Go template. At launch, the execution provider renders the template with the execution's instruction inputs as its data, and the rendered text becomes the agent's instructions.

- The template can read the instruction inputs alone. It can't read environment variables, credentials, files, or the clock.
- Rendering uses the built-in Go template functions alone. The same template with the same instruction inputs always renders the same text.
- A reference to an input that the execution doesn't supply fails the render rather than producing empty text.
- The input schema describes each value the template uses. The provider validates the instruction inputs against the input schema before it renders.
- The provider publishes the rendered instructions as evidence, with their digest. The provenance record and the configuration digest both include that digest.

Instruction inputs appear in workflow history and in evidence, so they never contain a secret value. A profile that needs a secret lists it under credential brokering instead.

### Harness permissions

A harness permission lets the harness offer the agent one kind of action inside the sandbox. For each one, the execution provider sets the matching permission mode and tool grants in the harness's own settings when it launches the agent. A harness that can't enforce one of a profile's harness permissions doesn't belong in any entry of that profile's models.

| Harness permission | Meaning                                                                                                     |
| ------------------ | ----------------------------------------------------------------------------------------------------------- |
| `run-commands`     | The agent may run commands. The sandbox still confines those commands.                                      |
| `edit-files`       | The agent may create, change, and delete files in the workspace. It requires `read-write` workspace access. |
| `commit`           | The agent may record changes in the source history of the workspace. It requires `edit-files`.              |

The provider grants each listed harness permission. It denies the rest. Before it starts the agent, it checks that the harness settings it produced grant exactly the profile's harness permissions.

### MCP access

The [Model Context Protocol (MCP) gateway](agent-mcp-gateway.md) gives each harness one endpoint for profile-selected servers. Harness-native tools use the separate tools field. The harness integration turns off other MCP server configuration on both ACP and headless launches.

| Profile key | Meaning |
| --- | --- |
| `mcp.servers.allow` | Configured server identities the execution may use, including `afakt` for the built-in server. |
| `mcp.servers.deny` | Server identities denied even when allowed. |
| `mcp.tools.<server>.allow` | Upstream tool names allowed for that server, with exposed names namespaced by the gateway. |
| `mcp.tools.<server>.deny` | Tools denied even when allowed. |
| `mcp.approvals.<server>` | Tools requiring approval before dispatch. Approval doesn't widen any allowlist or capability. |
| `mcp.unattended_approval` | `deny` or `allow` for approval-required calls in executions without an attached writer, including headless runs. Without it, `deny` applies. |

Omitted allow and deny lists are empty in a profile. With an empty allowlist, the agent can't use any server or tool at that level. The execution specification may narrow server and tool allowlists, and its deny lists add to the profile's. It can't remove required approvals or change the unattended rule. Interactive session approvals use the session's writer. Unattended executions use the pinned profile's rule. See [Tool approvals](agent-mcp-gateway.md#tool-approvals).

Each server's deployment configuration fixes remote HTTP or local standard input and output transport and, for a local server, agent-host or central hosting. It also supplies credential references. Profiles select these configured identities. Specifications can't replace their endpoints, commands, hosting, or credentials. Credential reference format and OAuth entry fields remain [open questions](agent-open-questions.md#agent-sandbox).

### Capabilities

A capability allows the agent one afakt API operation, which it calls through the afakt CLI or the [built-in afakt MCP server](agent-mcp-gateway.md#built-in-afakt-mcp-server). Reading a pull request or posting a review are afakt API operations too. A capability controls which external operations the agent may perform and where. Its [execution token](agent-sandbox.md#execution-token) lists the execution's capabilities, and the API server checks them on each request.

Each capability has the form `<resource>.<action>`, such as `pull-requests.read` or `pull-requests.review`. The API server publishes the full set. A profile lists each capability with optional scopes, such as repositories and branches, that limit where it applies. Scopes accept [sandbox variables](agent-sandbox.md#sandbox-variables). With them, a profile can limit a capability to the repository in the workspace.

```yaml
capabilities:
  pull-requests.read: { repositories: ["${repository:app}"] }
  pull-requests.review: { repositories: ["${repository:app}"] }
  branches.push: { repositories: ["${repository:app}"], branches: ["${branch:app}"] }
```

The agent has the listed capabilities and no others. An execution specification may narrow them by omitting capabilities or narrowing scopes, but it never adds a capability or widens a scope.

## Repository profile files

A repository defines its own profiles as YAML files in `.afakt/agent-profiles/`, one profile per file. In this example, `.afakt/agent-profiles/review.yaml` defines the repository profile `review`. The file is a [profile resource](#profile-resource) without the keys that afakt sets. Its `spec` keys match the profile fields, and its sandbox and credentials keys use the formats in [Agent sandbox](agent-sandbox.md). Instructions, input schema, and output schema appear inline or as paths to other files in the same repository at the same commit.

```yaml
apiVersion: afakt/v1alpha1
kind: AgentProfile
metadata:
  name: review
  labels: { team: platform }
spec:
  description: Review a pull request and return a verdict.
  instructions: .afakt/agent-profiles/review.md.tmpl
  input_schema:
    type: object
    required: [pull_request]
    properties:
      pull_request: { type: integer, minimum: 1 }
      focus: { type: string }
  workspace_access: read-only
  models:
    - harness: { name: example-harness, version: ">=2.0" }
      transport: headless
      provider: example-provider
      model: example-model-large
      effort: high
      default: true
      priority: 1
    - harness: { name: example-harness, version: ">=2.0" }
      provider: example-provider
      model: example-model-large
      effort: medium
      default: false
      priority: 2
    - harness: { name: example-harness }
      provider: example-provider
      model: example-model-small
      effort: high
      default: false
      priority: 3
  harness_permissions: [run-commands]
  tools: [read, search, shell]
  mcp:
    servers: { allow: [issue-tracker], deny: [] }
    tools:
      issue-tracker: { allow: [read-issue, comment-issue], deny: [] }
    approvals:
      issue-tracker: [comment-issue]
    unattended_approval: deny
  capabilities:
    pull-requests.read: { repositories: ["${repository:.}"] }
    pull-requests.comment: { repositories: ["${repository:.}"] }
    pull-requests.review: { repositories: ["${repository:.}"] }
  sandbox:
    network:
      allowedDomains: ["${repository-host:.}", api.example-provider.com]
    filesystem:
      denyRead: ["${home}/.aws", "${home}/.docker"]
  credentials:
    - service: model-provider-api
      apiKey:
        name: MODEL_PROVIDER_API_KEY
        inject:
          - domain: api.example-provider.com
            header: Authorization
            format: "Bearer %s"
  environment:
    - { name: GOFLAGS, value: -mod=vendor, locked: true }
    - { name: GOCACHE, value: "${cache:go-build}" }
  hooks:
    - name: lint
      event: before-agent
      command: [make, lint, "LINT_REPORT=${scratch}/lint.txt"]
      timeout: 5m
    - name: check-command
      event: tool-call-before
      matcher: [shell]
      command: [.afakt/hooks/check-command]
      timeout: 10s
      decisions: [deny]
      on_failure: block
  telemetry_content: omit
  output_schema: .afakt/agent-profiles/review-verdict.schema.json
  output_schema_policy: fixed
  default_limits: { time_limit: 30m, stop_grace_period: 1m }
  limit_ceilings: { time_limit: 2h, stop_grace_period: 5m }
  retention_ceiling: 30d
```

A repository profile file must parse, have `apiVersion` `afakt/v1alpha1` and `kind` `AgentProfile`, have a `metadata.name` that equals its filename without `.yaml`, set none of the keys that afakt sets, reference files that exist at its commit, contain a valid template and valid schemas, and have exactly one model entry with `default: true`. At launch, the provider checks this when it loads the file. Publishing a catalog version checks the same rules.

## Overrides

An execution specification may override some profile fields, within the bounds the profile sets. It can't override any other field.

| Specification field | Override bound |
| --- | --- |
| Harness, model provider, model, and model effort | Together they must match one entry in the profile's models, with a harness version that meets the entry's `version` constraint when it has one. Values the specification omits come from an entry that matches the values it states. When more than one entry matches, priority and capacity choose among them. |
| Workspace access | `read-only` for a `read-write` profile. A specification can narrow access but never widen it. |
| Output schema | Allowed when the profile's output schema policy is `per-execution`. |
| Time limit, stop grace period, and resource limits | Up to the profile's limit ceilings. |
| Retention | Up to the profile's retention ceiling. |
| Capabilities | A subset of the profile's capabilities, each with scopes that fall within the profile's scopes. |
| MCP servers and tools | Omitted allowlists inherit the profile. Supplied allowlists must be subsets. Deny lists join, and deny wins. Approval rules remain fixed by the profile. See [MCP access](#mcp-access). |
| Credential references | Present on both the profile and execution. Whether execution references must stay within those the profile permits remains [on hold](agent-open-questions.md#agent-sandbox). |
| Sandbox | Combined with the profile's sandbox configuration under the rules in [Combining configurations](agent-sandbox.md#combining-configurations). |
| Environment | New names, and new values for profile entries without `locked`, under the rules in [Combining environment entries](agent-workspace.md#combining-environment-entries). |
| Hooks | Always allowed. The specification's hooks add to the profile's under the rules in [Combining hooks](agent-hooks.md#combining-hooks). |
| Labels | Free, because labels don't change how the agent runs. |

The specification also supplies the instruction inputs, which the input schema bounds. Instructions, the input schema, harness transport, harness permissions, harness-native tools, MCP approval rules, and telemetry content come from the profile alone. A specification that sets any of them is invalid.

## Pinning a profile version

An execution specification identifies its profile by source, name, exact version, and profile digest. The execution provider never resolves a name to the newest version or a branch to its latest commit.

Resolving the version is the parent workflow's job, before it builds the specification. It records the result in workflow history. For a catalog profile, the parent looks up the newest active version, and for a repository profile, it reads the file at a commit it chooses, such as the commit of the workspace source. A replay or retry then pins the same version. A higher generation of an execution may pin a different version, as a deliberate new attempt.

At launch, the provider stores a copy of the pinned profile content, with its referenced files, in the execution record. Status, cancel, result, and reconciliation then work from that copy, even if the commit later becomes unreachable.

The specification digest covers the pinned version's digest together with the overrides, and so does the configuration digest that the provider computes at launch. Launches that pin different profile versions have different specification digests, and the provider handles them under the [execution identity](agent-execution-model.md#execution-identity) rules.

## Launch validation

Before it records an execution, the execution provider validates the specification against the pinned profile version. It checks the items below in order and stops at the first failure, returning that failure's error. A failed validation leaves the execution unrecorded, so a corrected launch with the same execution ID and generation can follow.

| Error | Meaning |
| --- | --- |
| `profile-not-found` | The catalog has no profile with that name, or the repository commit has no file for it. |
| `profile-version-not-found` | The profile exists, but the catalog has no version with that number, or the repository has no such commit. |
| `profile-digest-mismatch` | The version exists, but its digest differs from the digest in the specification. |
| `profile-invalid` | The repository profile file breaks one of the rules in [Repository profile files](#repository-profile-files). The error gives the file and the first problem. |
| `profile-retired` | The catalog version has status `retired`, and the launch would start a new execution. |
| `model-not-allowed` | No entry in the profile's models matches the specification's harness name and version, model provider, model, and model effort. |
| `harness-transport-unsupported` | The selected model entry sets `acp` for a harness without an ACP adapter. |
| `workspace-access-not-allowed` | The specification requests wider workspace access than the profile grants. |
| `override-not-allowed` | The specification sets a field that only the profile may set. Examples include instructions and MCP approval rules. A per-execution schema under a fixed policy or MCP access outside the profile's allowlists also returns this error. |
| `output-schema-missing` | The output schema policy is `per-execution`, and the specification omits a schema. |
| `limit-exceeds-profile` | A requested limit or the requested retention exceeds the profile's ceiling. The error lists each limit over its ceiling. |
| `instruction-inputs-invalid` | The instruction inputs fail validation against the input schema. The error lists each failing input path. |
| `instructions-render-failed` | The inputs pass the input schema, but rendering the template fails, such as on a reference to an input the execution omits. |
| `capability-unknown` | A capability isn't among those the API server publishes, or a scope names something the capability doesn't take. |
| `capability-not-allowed` | The specification lists a capability the profile lacks, or a scope outside the profile's scope. |
| `sandbox-invalid` | A sandbox configuration sets an unknown setting, a setting of the agent host, or a setting that only the profile may set. |
| `environment-invalid` | An environment entry has a name that breaks the name rules, a name that appears twice in one list, a value over the size limit, or `locked` in a specification. The error gives the entry. |
| `environment-name-reserved` | An environment entry, in the profile, the specification, or a hook, uses a [reserved name](agent-workspace.md#reserved-names). The error identifies the entry. |
| `environment-locked` | A specification entry or a hook's entry sets the name of a locked profile entry. The error gives the entry and the name. |
| `hook-invalid` | A hook breaks a rule from [Hook fields](agent-hooks.md#hook-fields), such as a duplicate name in the combined list, an empty command, a working directory without a matching workspace source, or a timeout no shorter than the time limit. Other examples are a field that its event doesn't take, a tool name outside the profile's tools, or a decision that its event doesn't allow. The error gives the hook and the rule. |
| `sandbox-variable-unknown` | A sandbox rule, capability scope, inject domain, environment value, or hook uses a variable that [Sandbox variables](agent-sandbox.md#sandbox-variables) doesn't define, or a target path without a matching workspace source. |
| `sandbox-not-allowed` | A specification allow entry falls outside the profile's allow list, or the specification turns on an isolation relaxation that the profile leaves off. |
| `credential-domain-not-allowed` | A credential brokering inject domain doesn't match the combined `allowedDomains`, or `tlsTerminate.excludeDomains` lists it. |
| `harness-permission-unsupported` | The harness on this agent host can't enforce one of the profile's harness permissions. |
| `sandbox-unsupported` | The provider can't enforce a setting of the combined sandbox configuration on this agent host. The error gives the setting. |
| `credential-unavailable` | The agent host can't resolve one of the profile's or execution's credential references. The error gives the reference, never a value. |

A hook that the selected harness doesn't support never fails the launch. The provider skips it with a warning, as [Harness support](agent-hooks.md#harness-support) describes. These errors join the [provider operation errors](agent-execution-contract.md#provider-operations). A repeated launch of an execution that already exists returns that execution under the identity rules, without validating again, even when its profile version has since retired.

## Profile changes

A profile changes by gaining a new version. Existing versions stay as published, so an execution that pinned a version runs, reconciles, and reports against that exact content.

- A catalog change publishes a new version with a higher number, which becomes the newest active version. Executions that pinned an older version continue unchanged.
- Retiring a catalog version stops new executions from pinning it. Existing executions of that version continue, and status, cancel, result, and repeated launches of them still work.
- The catalog deletes a retired version only after cleanup has removed every execution that pinned it.
- A repository change is a new commit that edits, adds, or deletes a profile file. Executions pinned to an earlier commit continue against their stored copy.
- The provenance record of each execution gives the source, name, version, and digest of the profile it ran.

## Example profiles

| Profile | Workspace access | Harness permissions | Capabilities | Sandbox network | Output schema policy |
| --- | --- | --- | --- | --- | --- |
| `build` | `read-write` | `run-commands`, `edit-files`, `commit` | `pull-requests.read`, and `branches.push` limited to the new branch | The repository host, package registries, and the model provider | `per-execution` |
| `review` | `read-only` | `run-commands` | `pull-requests.read`, `pull-requests.comment`, and `pull-requests.review`, limited to the workspace repository | The repository host and the model provider | `fixed`, a review verdict |

A `review` execution starts from a fresh copy of the source at the pinned revision, so it never needs write access to the workspace to update it.
