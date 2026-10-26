---
type: Design
title: Agent execution resource
description: The AgentExecution resource, a Kubernetes-style form of one execution generation with its specification in spec and the provider's record in status, and a full example.
tags: [afakt, agents, execution, resources]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent execution resource

Each execution is a harness run from start to end and belongs to a durable [agent](agent.md). Its provider record has one resource form per generation, `AgentExecution`, in the style of a Kubernetes resource. `metadata` identifies the generation, `spec` contains its [execution specification](agent-execution-contract.md#execution-specification), and `status` contains what the execution provider records about it. Launch takes the resource without `status`. Status and result return it with `status`. The [profile resource](agent-profile.md#profile-resource) uses the same `apiVersion`.

The field that links this resource to the agent it belongs to remains an [open question](agent-open-questions.md#agents), along with the mapping between generations and harness runs.

## Envelope

| Key | Meaning | Who sets it |
| --- | --- | --- |
| `apiVersion` | The schema version of the resource, `afakt/v1alpha1`. The provider rejects any other with `invalid-specification`. | The parent workflow |
| `kind` | `AgentExecution`. | The parent workflow |
| `metadata.name` | The [execution ID](agent-execution-model.md#execution-id-format). | afakt, when it creates the execution |
| `metadata.generation` | The generation. | The parent workflow |
| `metadata.creationTimestamp` | When afakt created the logical execution. | afakt, when it creates the execution |
| `metadata.labels` | The specification's labels: `task-class`, `variant`, and `role`. | The parent workflow |
| `spec` | The other specification fields, under the keys in [Specification keys](#specification-keys). | The parent workflow |
| `status` | The provider's record of the generation, under the keys in [Status keys](#status-keys). | The provider |

One resource shows one generation. Status for an older generation returns that generation's resource, and its `status.current_generation` gives the newest. The specification digest covers `metadata.name`, `metadata.labels`, and `spec`. It never covers `metadata.generation`, `metadata.creationTimestamp`, or `status`.

## Specification keys

| Specification field | Key |
| --- | --- |
| Agent profile | `profile`, with `source`, `name`, `version`, and `digest`. `version` takes the form of `metadata.version` in the profile resource. |
| Credential references | `credentials`, with `service` naming each credential in afakt's credential store. See [Credential brokering](agent-sandbox.md#credential-brokering). |
| Workspace access | `workspace_access` |
| Harness | `harness`, with `name` and an optional `version` |
| Model provider | `model_provider` |
| Model | `model` |
| Model effort | `effort` |
| Inputs | `inputs`, with `prompt` and a list of `attachments`. Each reference has a `ref` and a `digest`, and an attachment also has a `media_type`. |
| Input digest | `input_digest` |
| Instruction inputs | `instruction_inputs` |
| Workspace sources | `workspace_sources`. Each entry has a `target` and a `kind`. A `repository` source adds `url`, `commit`, and an optional `branch`. An `image` source adds `reference` for an image repository and tag or manifest digest. An `execution` source adds `execution_id` and `generation`. |
| Environment | `environment`, with the entry keys of the profile's `environment` |
| Hooks | `hooks`, with the entry keys of the profile's `hooks`, plus `working_directory`, `environment`, and `when` |
| Time limit | `time_limit` |
| Stop grace period | `stop_grace_period` |
| Resource limits | `resource_limits`, a map of limit names to values |
| Output schema | `output_schema` |
| Capabilities | `capabilities`, in the form of the profile's `capabilities` |
| Model Context Protocol (MCP) access | `mcp.servers` and `mcp.tools`, with the profile’s allow and deny list keys. Approval keys are profile-only. See [MCP access](agent-profile.md#mcp-access). |
| Sandbox | `sandbox`, in the form of the profile's `sandbox` |
| Placement requirements | `placement`, a map of names to values |
| Retention | `retention` |

Credential references appear in `spec.credentials` and in the pinned profile. Their bounds remain [on hold](agent-open-questions.md#agent-sandbox). Neither the specification nor the provider's record contains secret values.

## Status keys

| Key | Field |
| --- | --- |
| `current_generation` | The execution's current generation. |
| `specification_digest` and `configuration_digest` | The digests that identify the specification and the configuration. See [Execution generation](agent-execution-model.md#execution-generation). |
| `placement` | The `agent_host` and `task_queue` the generation runs on. |
| `workspace` | The workspace `path` inside the guest. |
| `configuration` | The combined configuration: `sandbox`, `environment` with the source of each entry, `hooks` as the combined list of names in run order, `transport` as `acp` or `headless`, and `harness_settings_digest`. |
| `state`, `phase`, and `last_observed` | Where the generation stands. `phase` is absent before the first part runs and once the generation is terminal. |
| `state_history` | Each transition, with its `state`, the time `at`, and the `observation` behind it. |
| `cancellation` | The cancellation acknowledgement, with `accepted_at`, `state`, and `reason`. Absent without one. |
| `superseded_by` | The higher generation that superseded this one. Absent unless one did. |
| `observed_times` | `launch_accepted`, `agent_start`, `agent_end`, and `execution_end`. |
| `evidence` | The [evidence references](agent-execution-model.md#evidence-retention-and-cleanup), each with `id`, `name`, `media_type`, and `digest`. |
| `provenance` | The [provenance record](agent-execution-model.md#provenance-record), with `requested` and `observed` under each field. Each `image` workspace source has its requested `reference` and observed manifest `digest` at its `target`. |
| `result` | Once terminal, the result's `reason`, `payload`, `payload_digest`, and `observed_usage`, with `exit_status` for headless runs or `stop_reason` for ACP runs. |
| `retention` | `ends_at` and the list of `holds`. |
| `cleanup` | `state`, one of `none`, `complete`, or `incomplete`, and the `remaining` items. |

The [execution result](agent-execution-contract.md#execution-result) consists of `status.result` together with `metadata.name`, `metadata.generation`, `status.specification_digest`, `status.state`, `status.evidence`, `status.provenance`, and `status.observed_times` at the moment the generation became terminal.

## Example

This generation 2 of a `build` execution uses headless transport. Its cancellation and supersession fields appear as comments. It runs the catalog profile `build` against two repositories, a copy of generation 1's workspace, and an empty directory. Its hooks add to the profile's `lint` and `check-command` hooks.

```yaml
apiVersion: afakt/v1alpha1
kind: AgentExecution
metadata:
  name: 019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b
  generation: 2
  creationTimestamp: "2026-10-25T14:00:00Z"
  labels:
    task-class: build
    variant: b
    role: builder
spec:
  profile:
    source: catalog
    name: build
    version: 7
    # A repository profile gives the repository and commit instead:
    # version: git.example.com/team/app@9f8e7d6c5b4a39281706f5e4d3c2b1a098765432
    digest: sha256:3f5a…
  credentials:
    - service: model-provider-api
  workspace_access: read-write
  harness: { name: example-harness, version: example-release }
  model_provider: example-provider
  model: example-model-large-2026-09
  effort: high
  inputs:
    prompt: { ref: content/prompts/7c1d, digest: sha256:7c1d… }
    attachments:
      - { ref: content/attachments/a41e, media_type: text/markdown, digest: sha256:a41e… }
  input_digest: sha256:e02b…
  instruction_inputs:
    issue: 37
    focus: migrations
  workspace_sources:
    - target: app
      kind: repository
      url: https://git.example.com/team/app
      commit: 9f8e7d6c5b4a39281706f5e4d3c2b1a098765432
      branch: agent/issue-37
    - target: deps/library
      kind: repository
      url: https://git.example.com/team/library
      commit: 1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d
    - target: previous
      kind: execution
      execution_id: 019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b
      generation: 1
    - target: notes
      kind: empty
  environment:
    - { name: NODE_ENV, value: test }
    - { name: GOCACHE, value: "${cache:go-build-arm64}" }
  hooks:
    - name: install-deps
      event: before-agent
      command: [make, deps]
      working_directory: app
      environment:
        - { name: DEPS_LOG, value: "${scratch}/deps.txt" }
      timeout: 10m
    - name: add-context
      event: session-start
      command: [app/.afakt/hooks/context]
      decisions: [context]
    - name: check-push
      event: tool-call-before
      matcher: [shell]
      command: [app/.afakt/hooks/check-push]
      timeout: 10s
      decisions: [deny]
      on_failure: block
    - name: keep-going
      event: agent-stop
      command: [app/.afakt/hooks/require-tests]
      decisions: [continue]
    - name: collect-logs
      event: session-end
      command: [cp, -R, "${home}/.example-harness/logs", "${scratch}/harness-logs"]
    - name: run-tests
      event: after-agent
      when: success
      command: [make, test]
      working_directory: app
      timeout: 20m
  time_limit: 2h
  stop_grace_period: 2m
  resource_limits: { tokens: 4000000, cpu: 4, memory: 8Gi, disk: 20Gi }
  output_schema:
    type: object
    required: [summary, branch]
    properties:
      summary: { type: string }
      branch: { type: string }
      tests_passed: { type: boolean }
  capabilities:
    pull-requests.read: { repositories: ["${repository:app}"] }
    branches.push: { repositories: ["${repository:app}"], branches: ["${branch:app}"] }
  sandbox:
    network:
      allowedDomains: ["${repository-host:app}", api.example-provider.com, proxy.golang.org]
      deniedDomains: [uploads.example.com]
      refuseOpaqueTunnels: true
      ambiguousRequestTarget: refuse
    filesystem:
      denyRead: ["${home}/.config/gcloud"]
      allowWrite: ["${cache:go-build-arm64}"]
    credentials:
      envVars: [CI_TOKEN]
      files: ["${home}/.netrc"]
    allowPty: false
  placement:
    operating_system: linux
    architecture: arm64
    harness: example-harness
  retention: 14d
status:
  current_generation: 2
  specification_digest: sha256:5d2c…
  configuration_digest: sha256:b7e9…
  placement:
    agent_host: example-agent-host
    task_queue: agent-example-agent-host
  workspace:
    path: /var/lib/afakt/workspaces/019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b/2
  configuration:
    sandbox:
      network:
        allowedDomains: [git.example.com, api.example-provider.com, proxy.golang.org, "afakt-api.internal:8443", "127.0.0.1:4318"]
        deniedDomains: [uploads.example.com]
        refuseOpaqueTunnels: true
        ambiguousRequestTarget: refuse
      filesystem:
        denyRead: [/home/agent/.aws, /home/agent/.docker, /home/agent/.config/gcloud]
        allowWrite:
          - /var/cache/afakt/go-build-arm64
          - /var/lib/afakt/workspaces/019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b/2
          - /var/lib/afakt/scratch/019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b/2
      credentials:
        envVars: [CI_TOKEN]
        files: [/home/agent/.netrc]
    environment:
      - { name: GOCACHE, value: /var/cache/afakt/go-build-arm64, source: specification }
      - { name: GOFLAGS, value: -mod=vendor, source: profile }
      - { name: NODE_ENV, value: test, source: specification }
    hooks: [lint, install-deps, add-context, check-command, check-push, keep-going, collect-logs, run-tests]
    transport: headless
    harness_settings_digest: sha256:0c4f…
  state: succeeded
  last_observed: "2026-10-25T15:01:28Z"
  state_history:
    - { state: pending, at: "2026-10-25T14:05:02Z", observation: launch recorded }
    - { state: running, at: "2026-10-25T14:05:09Z", observation: lint hook started }
    - { state: succeeded, at: "2026-10-25T15:01:30Z", observation: agent exited 0 with a valid payload, and run-tests succeeded }
  # Only after an acknowledged cancellation:
  # cancellation: { accepted_at: "2026-10-25T14:40:00Z", state: running, reason: operator request }
  # Only on a generation that a higher one superseded:
  # superseded_by: 3
  observed_times:
    launch_accepted: "2026-10-25T14:05:02Z"
    agent_start: "2026-10-25T14:08:41Z"
    agent_end: "2026-10-25T14:52:10Z"
    execution_end: "2026-10-25T15:01:30Z"
  evidence:
    - { id: ev-01, name: transcript, media_type: application/x-ndjson, digest: sha256:91aa… }
    - { id: ev-02, name: standard-error, media_type: text/plain, digest: sha256:2b70… }
    - { id: ev-03, name: rendered-instructions, media_type: text/markdown, digest: sha256:c3d1… }
    - { id: ev-04, name: sandbox-configuration, media_type: application/json, digest: sha256:8e12… }
    - { id: ev-05, name: sandbox-violations, media_type: application/x-ndjson, digest: sha256:4f6b… }
    - { id: ev-06, name: agent-environment, media_type: application/json, digest: sha256:a9c0… }
    - { id: ev-07, name: hooks, media_type: application/json, digest: sha256:17de… }
    - { id: ev-08, name: hook-runs, media_type: application/x-ndjson, digest: sha256:66f3… }
    - { id: ev-09, name: project-hook-configuration, media_type: application/json, digest: sha256:d24a… }
    - { id: ev-10, name: final-answer, media_type: application/json, digest: sha256:5a8e… }
    - { id: ev-11, name: workspace-changes, media_type: application/x-git-bundle, digest: sha256:e7b5… }
  provenance:
    execution_id: 019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b
    generation: 2
    specification_digest: sha256:5d2c…
    input_digest: sha256:e02b…
    configuration_digest: sha256:b7e9…
    trace_id: 4bf92f3577b34da6a3ce929d0e0e4736
    span_id: 00f067aa0ba902b7
    harness:
      requested: { name: example-harness, version: example-release }
      observed: { name: example-harness, version: example-release }
    transport:
      requested: headless
      observed: headless
    model_provider:
      requested: example-provider
      observed: example-provider
    model:
      requested: example-model-large-2026-09
      observed: example-model-large-2026-09
    effort:
      requested: high
      observed: high
    profile:
      requested: { source: catalog, name: build, version: 7, digest: sha256:3f5a…, instructions_digest: sha256:c3d1… }
    workspace:
      requested:
        - { target: app, kind: repository, url: https://git.example.com/team/app, commit: 9f8e7d6c5b4a39281706f5e4d3c2b1a098765432, branch: agent/issue-37 }
        - { target: deps/library, kind: repository, url: https://git.example.com/team/library, commit: 1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d }
        - { target: previous, kind: execution, execution_id: 019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b, generation: 1 }
        - { target: notes, kind: empty }
      observed:
        - { target: app, commit: 9f8e7d6c5b4a39281706f5e4d3c2b1a098765432, branch: agent/issue-37 }
        - { target: deps/library, commit: 1a2b3c4d5e6f708192a3b4c5d6e7f8091a2b3c4d }
        - { target: previous, digest: sha256:b1f8… }
    sandbox:
      requested:
        configuration_digest: sha256:8e12…
        capabilities:
          pull-requests.read: { repositories: [git.example.com/team/app] }
          branches.push: { repositories: [git.example.com/team/app], branches: [agent/issue-37] }
      observed:
        applied: [network, filesystem, credentials]
    environment:
      requested:
        - { name: GOCACHE, value: /var/cache/afakt/go-build-arm64, source: specification }
        - { name: GOFLAGS, value: -mod=vendor, source: profile }
        - { name: NODE_ENV, value: test, source: specification }
    hooks:
      requested:
        - { name: lint, event: before-agent, source: profile, command: [make, lint, "LINT_REPORT=/var/lib/afakt/scratch/019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b/2/lint.txt"] }
        - { name: install-deps, event: before-agent, source: specification, command: [make, deps] }
        - { name: add-context, event: session-start, source: specification, decisions: [context], command: [app/.afakt/hooks/context] }
        - { name: check-command, event: tool-call-before, source: profile, matcher: [shell], decisions: [deny], command: [.afakt/hooks/check-command] }
        - { name: check-push, event: tool-call-before, source: specification, matcher: [shell], decisions: [deny], command: [app/.afakt/hooks/check-push] }
        - { name: keep-going, event: agent-stop, source: specification, decisions: [continue], command: [app/.afakt/hooks/require-tests] }
        - { name: collect-logs, event: session-end, source: specification, command: [cp, -R, /home/agent/.example-harness/logs, /var/lib/afakt/scratch/019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b/2/harness-logs] }
        - { name: run-tests, event: after-agent, source: specification, command: [make, test] }
      observed:
        runs:
          - { name: lint, runs: 1, failures: 0 }
          - { name: install-deps, runs: 1, failures: 0 }
          - { name: add-context, runs: 1, failures: 0 }
          - { name: check-command, runs: 42, failures: 0 }
          - { name: check-push, runs: 42, failures: 1 }
          - { name: keep-going, runs: 2, failures: 0 }
          - { name: collect-logs, runs: 1, failures: 0 }
          - { name: run-tests, runs: 1, failures: 0 }
        support_warnings:
          - { hook: keep-going, event: agent-stop, support: partial }
        repository_hooks:
          - { file: app/.example-harness/settings.json, event: PostToolUse, command: [./scripts/format-changed] }
    agent_host:
      requested: { operating_system: linux, architecture: arm64, harness: example-harness }
      observed: { agent_host: example-agent-host, worker_version: example-release }
    conflicting: false
  result:
    reason: agent-exit
    exit_status: 0
    # ACP runs report stop_reason, such as end_turn, instead of exit_status.
    payload:
      summary: Added the migrations runner and its tests.
      branch: agent/issue-37
      tests_passed: true
    payload_digest: sha256:5a8e…
    observed_usage: { input_tokens: 1843210, output_tokens: 96402, cost: "11.82 USD" }
  retention:
    ends_at: "2026-11-08T15:01:30Z"
    holds: []
  cleanup:
    state: none
    remaining: []
```
