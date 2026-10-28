---
type: Design
title: Agent observability
description: The OpenTelemetry trace of an agent launch, the spans inside it, the trace context the harness receives, and the metrics afakt publishes about agent launches.
tags: [afakt, agents, agent-launch, observability, opentelemetry]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent observability

Each agent launch of a durable [agent](agent.md) produces agent launch-level OpenTelemetry traces and metrics. The history across these agent launches belongs to the agent. The trace shows each step of the [lifecycle](workload-lifecycle.md#Agent%20launch%20lifecycle) as a span, from placement to the result, with the agent harness's own spans nested under the agent harness span when the harness accepts a trace context. The metrics count agent launches and measure their durations, hooks, usage, and workload host capacity without one series per agent launch.

[Operations and telemetry](agent-workspace-lifecycle.md#Operations%20and%20telemetry).

## Conventions

- Spans and metrics follow the [OpenTelemetry semantic conventions](https://github.com/open-telemetry/semantic-conventions), such as `error.type` and `process.exit.code`.
- The agent harness span and the usage metrics follow the [OpenTelemetry GenAI semantic conventions](https://github.com/open-telemetry/semantic-conventions-genai) for agent spans, at commit `9ffc7e0` of that repository. Those conventions have Development status, so afakt fixes the attribute names in the contract and changes them only in a new afakt release.
- Every attribute and metric that afakt defines itself has a name under `afakt.`. The usage metrics are afakt's own totals for each agent launch, not the GenAI client metrics.
- Each enumerated value that afakt defines is lowercase with `-` between words, such as `hook-failure`. Values from other conventions, such as `gen_ai.operation.name` and `error.type`, follow those conventions. A reason from the [agent launch result](agent-launch-contract.md#Agent%20launch%20result) takes the form `agent-exit`, `agent-refusal`, `launch-failure`, `workspace-failure`, `hook-failure`, `invalid-output`, `time-limit`, `resource-limit`, `sandbox-violation`, `cancellation`, `supersession`, or `worker-restart`.

## Trace topology

Headless transport integration and the observations available through it remain [`OQ-SES-16`](agent-open-questions.md#%60OQ-SES-16%60).

An agent launch's spans join its parent workflow's trace. Trace context passes from the workflow orchestrator through the agent dispatcher workflow to the agent launcher workflow, then from each workflow to its activities. Workload provider operations run inside their calling activities on the workload host worker, with a child span under the activity span in the same trace.

- **Agent launch.** Each agent dispatcher run handles one agent launch and has one `afakt.agent_launch` span. Without a trace context from the parent, that span starts a new trace.
- **Successor agent launches.** A successor agent launch's `afakt.agent_launch` span links to the predecessor agent launch's `afakt.agent_launch` span, with the link attribute `afakt.link.relation` set to `supersedes` when the predecessor agent launch was still active and `follows` otherwise. Spans of different agent launches share a trace only when the parent starts both from one trace.
- **Retries.** A retried activity is a new activity span under the same parent span, and a repeated provision request that returns the existing agent launch doesn't start a new `afakt.agent_launch.lifecycle` span. Replay of workflow history never emits a span again.
- **Restarts.** For recovery, the workload provider records the trace ID and span ID of each open span in the agent launch record. A restart of the workload host worker loses each span that was open, because the worker never exports a span before it ends. For the rest of each such step, the provider starts a recovery span with the same name and parent. Each recovery span links to the recorded span and has `afakt.recovered` set to true. Reconciliation runs in an `afakt.agent_launch.reconcile` span that links to each recovery span it starts.
- **Cleanup.** Cleanup runs after retention, long after the trace ends, so each `afakt.cleanup` span starts a new trace and links to the `afakt.agent_launch.lifecycle` span of the agent launch it removes.
- **Sampling.** Every afakt span follows its parent's sampling decision. The trace flags in `TRACEPARENT` pass that decision to the agent harness.

The provenance record gives the trace ID and the span ID of the agent launch's `afakt.agent_launch.lifecycle` span, so a reader can go from an agent launch to its trace.

## Spans

| Span name | Kind | Parent | Covers |
| --- | --- | --- | --- |
| `afakt.agent_launch` | `INTERNAL` | The parent workflow's span that starts the agent dispatcher | One agent dispatcher run, from placement until it returns the result. |
| `afakt.agent_launch.placement` | `INTERNAL` | `afakt.agent_launch` | The placement activity, from the request until the reservation Application Programming Interface (API) reserves a slot. |
| `afakt.agent_launcher` | `INTERNAL` | `afakt.agent_launch` | One agent launcher run on the workload host, from its start until it returns the result. |
| `afakt.provider.provision`, `afakt.provider.status`, `afakt.provider.cancel`, and `afakt.provider.result` | `INTERNAL` | The calling activity's span, under `afakt.agent_launcher` | One workload provider operation on the workload host. |
| `afakt.agent_launch.lifecycle` | `INTERNAL` | The `afakt.provider.provision` span that accepted the launch | The workload provider's work on one agent launch, from the accepted launch until evidence publication and removal of the agent sandbox finish. It can end after its parent ends. |
| `afakt.workspace.build` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Building the agent workspace. |
| `afakt.workspace.source` | `INTERNAL` | `afakt.workspace.build` | Placing one workspace source. This covers repository fetches and image materialization as well as copies of earlier workspaces. |
| `afakt.workspace.image.build` | `INTERNAL` | The activity span that runs the image build | One image build, from source preparation through sandboxed steps and snapshotting to publication. |
| `afakt.sandbox.configure` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Combining the sandbox configurations and the workload environment and rendering the instructions. |
| `afakt.harness.configure` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Writing the agent harness settings, with the harness permissions, the tools, the hooks for harness events, and the harness's telemetry export. |
| `afakt.hook` | `INTERNAL` | `afakt.stop` for `stop-requested`, `afakt.agent_launch.lifecycle` for every other provider event, and `afakt.hook.dispatch` for a harness event | One run of one hook. |
| `invoke_agent {gen_ai.agent.name}` | `INTERNAL` | `afakt.agent_launch.lifecycle` | The agent harness activity, from harness or Agent Client Protocol (ACP) adapter startup until the provider observes that the harness, adapter, and their helpers have exited. Control services remain available for later hooks under [Running hooks](agent-hooks.md#Running%20hooks). |
| `afakt.hook.dispatch` | `INTERNAL` | The span in the `TRACEPARENT` that the afakt command-line interface (CLI) receives from the harness, which is the agent harness span or a harness span under it | One occurrence of a harness event, from the moment the harness starts the afakt CLI until the CLI returns the decisions. |
| `afakt.mcp.tool.call` | `SERVER` | The authenticated agent launch’s agent harness span | One gateway tool call, including any approval wait and denied or built-in calls. |
| `afakt.output.validate` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Validating the final answer against the output schema. |
| `afakt.result.build` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Recording the workspace changes and building the agent launch result. |
| `afakt.evidence.publish` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Publishing one evidence item. |
| `afakt.stop` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Stopping a hook or the agent after a cancellation, a supersession, or a reached limit, until the provider observes that every process it started for the stopped hook or agent has exited. |
| `afakt.sandbox.teardown` | `INTERNAL` | `afakt.agent_launch.lifecycle` | Ending the credential broker's use of credentials and the agent launch's authority, then removing the sandbox. |
| `afakt.agent_launch.reconcile` | `INTERNAL` | The span of the workload host worker's startup | Loading the agent launch records after a restart and observing each agent again. |
| `afakt.preview.request` | `SERVER` | The incoming request context, or a new trace | One preview request or streaming connection, linked to its agent launch. |
| `afakt.cleanup` | `INTERNAL` | None, a new trace | Cleanup of one agent launch. |

The `afakt.agent_launch.lifecycle` span ends after the last `afakt.evidence.publish` and `afakt.sandbox.teardown` spans under it. The agent harness span's name uses the profile name, such as `invoke_agent review`.

### Attributes

Each span from `afakt.agent_launch` down includes the agent launch attributes. Other spans add their own.

| Attribute | Type | Spans | Value |
| --- | --- | --- | --- |
| `afakt.agent_launch.id` | string | All agent launch spans | The agent launch ID, as a canonical UUIDv7. |
| `afakt.agent_spec.generation` | int | All agent launch spans | The agent spec generation recorded at finalization. |
| `afakt.profile.source` | string | All agent launch spans | `catalog` or `repository`. |
| `afakt.profile.name` | string | All agent launch spans | The profile name. |
| `afakt.profile.version` | string | All agent launch spans | The profile version, in the form of `AFAKT_PROFILE_VERSION`. |
| `afakt.agent_host` | string | All spans from placement on | The workload host's name. |
| `afakt.label.task_class` and `afakt.label.variant` | string | `afakt.agent_launch` and `afakt.agent_launch.lifecycle` | The agent launch's labels, when set. |
| `afakt.agent_launch.role.name` | string | `afakt.agent_launch` and `afakt.agent_launch.lifecycle` | The name of the launch's own Identity and Access Management (IAM) role. |
| `afakt.specification.digest` | string | `afakt.agent_launch.lifecycle` | The specification digest. |
| `afakt.agent_launch.state` | string | `afakt.agent_launch` and `afakt.agent_launch.lifecycle` | The terminal state, set when the span ends. |
| `afakt.agent_launch.reason` | string | `afakt.agent_launch` and `afakt.agent_launch.lifecycle` | The reason from the result, set when the span ends. |
| `afakt.workspace.source.index` | int | `afakt.workspace.source` | The source's position, counted from 1. |
| `afakt.workspace.source.kind` | string | `afakt.workspace.source` | `repository`, `image`, `agent-launch`, or `empty`. |
| `afakt.workspace.source.target` | string | `afakt.workspace.source` | The target path. |
| `afakt.workspace.source.commit` | string | `afakt.workspace.source` | For a `repository` source, the full commit ID of the checkout. |
| `afakt.workspace.image.digest` | string | `afakt.workspace.source` for an image, and `afakt.workspace.image.build` on successful publication | The resolved or published manifest digest. |
| `afakt.workspace.image.build.trigger` | string | `afakt.workspace.image.build` | `schedule`, `push`, or `manual`. |
| `afakt.workspace.image.build.result` | string | `afakt.workspace.image.build` | `success`, `failure`, or `canceled`. |
| `afakt.hook.event` | string | `afakt.hook` and `afakt.hook.dispatch` | The hook event. |
| `afakt.hook.name` | string | `afakt.hook` | The hook's name. |
| `afakt.hook.source` | string | `afakt.hook` | `profile` or `specification`. |
| `afakt.hook.outcome` | string | `afakt.hook` | `success`, `failure`, `timeout`, or `skipped`. A hook that the failure rules skip has a span of zero duration. |
| `process.exit.code` | int | `afakt.hook` and the agent harness span | The hook exit status, or the headless agent harness's exit status, when one exists. |
| `afakt.hook.tool` | string | `afakt.hook` and `afakt.hook.dispatch` | For a tool event, the tool name from the profile's tools. |
| `afakt.hook.decision` | string | `afakt.hook` and `afakt.hook.dispatch` | For a harness event or `sandbox-violation`, the hook's decision, or on the dispatch span the decision that ended the event, or `none`. |
| `gen_ai.operation.name` | string | Agent harness span | `invoke_agent`. |
| `gen_ai.agent.name` | string | Agent harness span | The profile name. |
| `gen_ai.agent.version` | string | Agent harness span | The profile version. |
| `gen_ai.provider.name` | string | Agent harness span | The model provider, as a well-known value of the GenAI conventions when one matches and as afakt's own name otherwise. |
| `gen_ai.request.model` | string | Agent harness span | The model from the specification. |
| `gen_ai.response.model` | string | Agent harness span | The model that served the requests, when the harness or model provider reports it. |
| `gen_ai.usage.input_tokens` and `gen_ai.usage.output_tokens` | int | Agent harness span | Observed usage, when reported. |
| `afakt.harness.name` and `afakt.harness.version` | string | Agent harness span and `afakt.harness.configure` | The harness and the version it reports. |
| `afakt.harness.transport` | string | Agent harness span | `acp` or `headless`, from the selected model entry. |
| `afakt.agent.stop_reason` | string | Agent harness span | The ACP stop reason, when reported. Absent on headless agent launches. |
| `afakt.effort` | string | Agent harness span | The reasoning effort level that the specification's model effort sets. |
| `afakt.evidence.media_type` | string | `afakt.evidence.publish` | The evidence item's media type. |
| `afakt.stop.cause` | string | `afakt.stop` | `cancellation`, `supersession`, `time-limit`, `resource-limit`, `sandbox-violation`, or `timeout` for a hook. |
| `afakt.recovered` | boolean | Any recovery span | True. |
| `afakt.mcp.server` | string | `afakt.mcp.tool.call` | Configured server identity, including `afakt` for the built-in server. |
| `afakt.mcp.tool` | string | `afakt.mcp.tool.call` | The namespaced tool name. |
| `afakt.mcp.outcome` | string | `afakt.mcp.tool.call` | `success`, `denied`, `error`, or `canceled`. |
| `afakt.mcp.approval.outcome` | string | `afakt.mcp.tool.call` for an approval-required call | `allowed`, `denied`, or `canceled`. |
| `afakt.mcp.approval.source` | string | `afakt.mcp.tool.call` when an answer exists | `writer` or `profile`. |
| `error.type` | string | Any span with error status | The reason or the error name, such as `hook-failure` or `workspace-source-unavailable`. |

### Span status

- A span that ends in a failure has status `ERROR` with `error.type` set. For `afakt.agent_launch` and `afakt.agent_launch.lifecycle`, that means the terminal states `failed` and `terminated`.
- A `canceled` agent launch leaves the status unset, because the caller asked for the stop. Its `afakt.agent_launch.reason` is `cancellation` or `supersession`.
- A span for a workload provider operation has status `ERROR` for each error from a workload provider operation except `not-found` during the first status call of reconciliation and `not-terminal`, which the agent launcher expects as answers.
- An `afakt.hook` span has status `ERROR` when the hook fails, even when its on-failure rule lets the event go on.
- Every other span leaves the status unset on success.

### Events

| Event | Span | Attributes |
| --- | --- | --- |
| `afakt.state_change` | `afakt.agent_launch.lifecycle` | `afakt.agent_launch.state.from`, `afakt.agent_launch.state.to`, and the observation behind it as `afakt.observation` |
| `afakt.cancel_acknowledged` | `afakt.agent_launch.lifecycle` | `afakt.agent_launch.reason` |
| `afakt.stop.signal` | `afakt.stop` | `afakt.stop.signal_name`, such as `SIGTERM` |
| `afakt.stop.force` | `afakt.stop` | None |
| `afakt.termination.verified` | `afakt.stop`, the agent harness span, and each `afakt.hook` span for a workload provider event | `afakt.process.count`, the number of processes the workload provider saw exit |
| `afakt.hook.warning` | `afakt.harness.configure` | `afakt.hook.name`, `afakt.hook.event`, `afakt.hook.decision` for an unsupported decision, `afakt.harness.name`, and `afakt.hook.warning.kind`, `hook-partial` or `hook-unsupported` |
| `afakt.sandbox.violation` | Agent harness span or hook span | `afakt.sandbox.violation.kind`, `network` or `filesystem` |
| `afakt.observation.conflict` | `afakt.agent_launch.lifecycle` | The provenance field whose observations disagree, as `afakt.provenance.field` |

### Links

- A successor agent launch's `afakt.agent_launch` span links to the predecessor agent launch's, as [Trace topology](#Trace%20topology) describes.
- An `afakt.workspace.source` span for an `agent-launch` source links to the `afakt.agent_launch.lifecycle` span of the agent launch it copies.
- An `afakt.workspace.source` span for an `image` source links to the `afakt.workspace.image.build` span that produced its digest, when that context is available. The image build belongs to its triggering workflow's trace. It needn't share the consuming agent launch's trace.
- A recovery span links to the span that was open at the restart, and `afakt.agent_launch.reconcile` links to each recovery span it starts.
- `afakt.preview.request` links to its agent launch's `afakt.agent_launch.lifecycle` span, including when the request starts a separate trace.
- `afakt.cleanup` links to the `afakt.agent_launch.lifecycle` span of its agent launch.

## Trace context propagation

The workload provider passes the trace context into the agent sandbox through the [trace context variables](agent-workspace.md#Trace%20context). The agent receives the agent harness span's context, and each hook receives its own span's context, so a hook's child processes can add spans under it.

- Through the workload proxy mediator and workload proxy, the agent harness sends telemetry to the workload host's endpoint at `${telemetry}`. The sandbox's network rules block every other telemetry destination.
- The telemetry endpoint identifies the launch from the authenticated workload-proxy connection and the context stamped by its mediator. See [Per-launch services](workload-proxy.md#Launch%20boundary). It adds `afakt.agent_launch.id` and `afakt.agent_spec.generation` to each span and log record but never to a metric. Then it forwards the data to the deployment's collector.
- For each harness, the telemetry endpoint keeps only the attributes on its list of permitted names and drops the rest. It also drops any log record that [What telemetry never records](#What%20telemetry%20never%20records) excludes, even when the harness's own settings fail to leave it out.
- A harness that accepts an inbound `TRACEPARENT` nests its spans under the agent harness span, while any other harness starts its own traces, which join the agent launch through the attributes that the telemetry endpoint adds.
- When the harness doesn't export telemetry, the trace still contains every afakt span, including the hook dispatch spans, because the afakt CLI emits those.

### Resource attributes

The workload host worker, the API server, and the afakt CLI set `service.name` to `afakt` and `service.version` to the afakt version. The workload provider passes `afakt.profile.source`, `afakt.profile.name`, and `afakt.harness.name` to the agent harness through `OTEL_RESOURCE_ATTRIBUTES`, and the harness sets its own `service.name`. The agent launch ID and recorded agent spec generation never become resource attributes, because the harness attaches its resource attributes to its metrics as well.

## Metrics

afakt publishes these metrics from the workload host worker and, for gateway, preview, and tunnel traffic, the control plane, with the units of the OpenTelemetry conventions.

| Metric | Instrument | Unit | Description | Attributes |
| --- | --- | --- | --- | --- |
| `afakt.agent_launches` | Counter | `{agent_launch}` | Agent launches that reached a terminal state. | Agent launch metric attributes |
| `afakt.agent_launch.duration` | Histogram | `s` | Time from the accepted launch to the terminal state. | Agent launch metric attributes |
| `afakt.agent_launch.agent.duration` | Histogram | `s` | Time from harness or ACP adapter startup until the harness, adapter, and their helpers have exited, matching the agent harness span. | Agent launch metric attributes |
| `afakt.agent_launch.active` | UpDownCounter | `{agent_launch}` | Agent launches in a non-terminal state. | `afakt.agent_launch.state`, `afakt.profile.name`, and `afakt.agent_host` |
| `afakt.agent_launch.unresolved.duration` | Histogram | `s` | Time each agent launch spent in `launch-unknown` or `unknown-stopped`. | `afakt.agent_launch.state` and `afakt.agent_host` |
| `afakt.placement.wait.duration` | Histogram | `s` | Time from the placement request until the reservation API reserves a slot. | `afakt.profile.name` |
| `afakt.agent_host.slot.limit` | UpDownCounter | `{slot}` | The host's slot capacity from the capacity controller. | `afakt.agent_host` |
| `afakt.agent_host.slot.reserved` | UpDownCounter | `{slot}` | Slots reserved and not yet claimed, as recorded by the reservation API. | `afakt.agent_host` |
| `afakt.agent_host.slot.usage` | UpDownCounter | `{slot}` | Claimed slots in use, as recorded by the reservation API, including agent launches that haven't started their agents. | `afakt.agent_host` |
| `afakt.agent_host.slot.capacity.changes` | Counter | `{change}` | Each change in slot capacity. | `afakt.agent_host` and `afakt.capacity.reason`, `cpu-use`, `memory-use`, `free-disk-space`, or `resource-headroom` |
| `afakt.workspace.source.duration` | Histogram | `s` | Time to place one workspace source. | `afakt.workspace.source.kind`, `afakt.profile.name`, and `error.type` on failure |
| `afakt.workspace.image.build.duration` | Histogram | `s` | Duration from image source preparation until publication, failure, or cancellation. | Image build metric attributes |
| `afakt.workspace.image.builds` | Counter | `{build}` | Image builds that ended. | Image build metric attributes |
| `afakt.workspace.layer.cache.requests` | Counter | `{request}` | Layer cache checks. One hit or miss per tier checked. | `afakt.agent_host`, `afakt.workspace.cache.tier`, `blob` or `unpacked`, and `afakt.workspace.cache.result`, `hit` or `miss` |
| `afakt.workspace.layer.fetched` | Counter | `By` | Layer bytes received from the registry, including bytes fetched again after a failed transfer. | `afakt.agent_host` |
| `afakt.workspace.build.duration` | Histogram | `s` | Elapsed workspace construction time spent in each step category. | `afakt.workspace.build.step`, `fetch`, `unpack`, or `repository`, `afakt.profile.name`, and `error.type` on failure |
| `afakt.hook.duration` | Histogram | `s` | Duration of one hook run. | Hook metric attributes |
| `afakt.hook.decisions` | Counter | `{decision}` | Decisions returned by hooks for harness events and `sandbox-violation`. | `afakt.hook.event`, `afakt.hook.decision`, `afakt.hook.source`, and `afakt.profile.name` |
| `afakt.hook.warnings` | Counter | `{warning}` | Hooks that the workload provider installed with a `hook-partial` warning or skipped with a `hook-unsupported` warning. | `afakt.hook.event`, `afakt.hook.warning.kind`, `afakt.harness.name`, and `afakt.profile.name` |
| `afakt.agent_launch.stops` | Counter | `{stop}` | Stops of a hook or the agent. | `afakt.stop.cause`, `afakt.profile.name`, and `afakt.stop.forced`, true for a forced stop |
| `afakt.sandbox.violations` | Counter | `{violation}` | Violations that the agent sandbox reported. | `afakt.sandbox.violation.kind` and `afakt.profile.name` |
| `afakt.agent_launch.usage.input_tokens` | Counter | `{token}` | Observed input tokens per agent launch, when reported. | Usage metric attributes |
| `afakt.agent_launch.usage.output_tokens` | Counter | `{token}` | Observed output tokens per agent launch, when reported. | Usage metric attributes |
| `afakt.cleanups` | Counter | `{cleanup}` | Each cleanup of one agent launch. | `afakt.cleanup.outcome`, `complete` or `incomplete`, and `afakt.agent_host` |

- **Agent launch metric attributes.** `afakt.profile.source`, `afakt.profile.name`, `afakt.harness.name`, `afakt.harness.transport`, `gen_ai.provider.name`, `gen_ai.request.model`, `afakt.effort`, `afakt.agent_launch.state`, `afakt.agent_launch.reason`, `afakt.label.task_class`, `afakt.label.variant`, and `afakt.agent_launch.role.name`.
- **Hook metric attributes.** `afakt.hook.event`, `afakt.hook.source`, `afakt.hook.outcome`, and `afakt.profile.name`. `afakt.hook.name` appears only for a hook from the profile.
- **Image build metric attributes.** `afakt.profile.source`, `afakt.profile.name`, `afakt.workspace.image.build.trigger`, and `afakt.workspace.image.build.result`. These use the same bounded catalog profile names and `other` rule as agent launch metrics.
- **Usage metric attributes.** `gen_ai.operation.name` as `invoke_agent`, `gen_ai.provider.name`, `gen_ai.request.model`, `afakt.profile.name`, and `afakt.harness.name`.

Harness metrics, such as per-request token counts, pass through the telemetry endpoint under their own names.

### Workspace images

Under the activity that runs the build, an image build span follows the triggering workflow's trace context. Failure sets span status `ERROR` and `error.type`. On a requested cancellation, status stays unset and `canceled` becomes the build result. This span covers builds without a model or agent launch. For each successful build, the worker records its published manifest digest and span context for links from consuming agent launches.

The build duration histogram and result counter record each ended build once. Cache requests distinguish the blob and unpacked tiers, with `hit` only for a complete verified entry. Preheat cache checks and transfers use these same metrics. Their cost remains visible without an agent launch ID attribute.

Workspace build timing splits into `fetch` for image manifests, config blobs, and missing layer transfers, `unpack` for extraction and clone or copy materialization, and `repository` for repository fetches and checkouts. An `agent-launch` source copy or an `empty` directory uses `unpack`. Each step category measures elapsed time over the intervals when that work runs, counting overlapping tasks within the category once. Categories can overlap, so their sum can exceed the parent workspace build span's duration. The existing workspace source duration metric still measures each source as a whole.

### Slot accounting

The worker publishes the counts of reserved and claimed slots from the reservation API separately. Their sum gives the slots that block new reservations. After a decrease, it can exceed `afakt.agent_host.slot.limit`. The counter for capacity changes increments once for each change, with the reason that [Slot capacity](workload-placement.md#Slot%20capacity) defines.

### Ports and previews

The proxy mediator publishes forwarded port counts. Preview, share-link, and tunnel metrics come from the control plane. Each request emits an `afakt.preview.request` server span that follows the incoming request context or starts a trace and links to its agent launch's `afakt.agent_launch.lifecycle` span. Its attributes include the agent launch attributes and `afakt.port.protocol` as `http`, plus `afakt.preview.outcome` as `allowed`, `denied`, `unavailable`, or `error`. A streaming span lasts until the stream closes. Transport failures set status `ERROR` and a bounded `error.type`. Normal closure leaves status unset. A denied request links to an agent launch only after its preview route identifies the owner.

| Metric | Instrument | Unit | Description | Attributes |
| --- | --- | --- | --- | --- |
| `afakt.ports.active` | UpDownCounter | `{port}` | Open forwarded port records. | `afakt.port.protocol`, `http` or `tcp`, `afakt.profile.name`, and `afakt.agent_host` |
| `afakt.preview.requests` | Counter | `{request}` | Preview requests. Counts include denied requests and streaming upgrades. | Preview metric attributes |
| `afakt.preview.transferred` | Counter | `By` | Application bytes carried by previews, including streams. | Preview metric attributes and `afakt.transfer.direction`, `request` or `response` |
| `afakt.preview.request.duration` | Histogram | `s` | Request latency from arrival until response or stream closure. | Preview metric attributes |
| `afakt.preview.share_links.created` | Counter | `{link}` | Share grants created. | `afakt.profile.name` |
| `afakt.preview.share_links.used` | Counter | `{use}` | Share grant authorization attempts. | `afakt.profile.name` and `afakt.share_link.outcome`, `allowed`, `expired`, `revoked`, or `invalid` |
| `afakt.tunnel.connections` | UpDownCounter | `{connection}` | Connector connections in each state, moving between states without counting one twice. | `afakt.tunnel.provider`, `cloudflare` or `tailscale`, and `afakt.tunnel.state`, `connecting`, `connected`, `reconnecting`, or `failed` |

Preview metric attributes are `afakt.profile.name`, `afakt.port.protocol` as `http`, and `afakt.preview.outcome`. Profile names follow the catalog bounds and repository `other` rule. Unattributed requests use `other` for profile name. Port numbers, port names, preview host names and uniform resource locators (URLs), session IDs, share-link IDs, and embedded node identities never become metric attributes. Preview telemetry never records request or response content, authorization headers, cookies, or share-link secrets. Credential injection audit records follow the same secret and content exclusions.

### MCP gateway

The [Model Context Protocol (MCP) gateway](agent-mcp-gateway.md#Audit%20and%20telemetry) emits an `afakt.mcp.tool.call` span for every call, including denied attempts and built-in afakt tools. The span is a child of the authenticated agent launch's agent harness span. The gateway obtains that parent from afakt's agent launch context, including for an agent harness without telemetry export. It records the agent launch attributes and MCP attributes from [Attributes](#Attributes). Its duration runs from call receipt through completion or denial, including approval wait. When an upstream call fails, the gateway sets the span's status to `ERROR` with a bounded `error.type`. Policy denial and requested cancellation leave status unset.

| Metric | Instrument | Unit | Description | Attributes |
| --- | --- | --- | --- | --- |
| `afakt.mcp.tool.calls` | Counter | `{call}` | Completed tool calls, including denied and built-in calls. | MCP call metric attributes |
| `afakt.mcp.tool.call.duration` | Histogram | `s` | Elapsed call time, including approval wait. | MCP call metric attributes |
| `afakt.mcp.approvals.requested` | Counter | `{approval}` | Approval requests raised, including requests answered by profile policy. | `afakt.mcp.server`, `afakt.mcp.tool`, and `afakt.profile.name` |
| `afakt.mcp.approvals.resolved` | Counter | `{approval}` | Approval requests answered or canceled. | `afakt.mcp.server`, `afakt.mcp.tool`, `afakt.profile.name`, `afakt.mcp.approval.outcome`, and `afakt.mcp.approval.source` when answered |
| `afakt.mcp.server.health` | Gauge | `1` | Upstream ready state: 1 when ready to serve calls, 0 when unavailable. | `afakt.mcp.server` and `afakt.mcp.hosting`, `remote`, `agent-host`, or `control-plane` |

MCP call metric attributes are `afakt.mcp.server`, `afakt.mcp.tool`, `afakt.mcp.outcome`, and `afakt.profile.name`. Approval outcomes are `allowed`, `denied`, or `canceled`. Answer sources are `writer` or `profile`. Every completed call and every resolved approval counts once. Central services report health from supervision and remote services from connection readiness. For a workload-host server, health is 1 only when all active instances of that configured server are ready. A failed instance makes the health value 0. With no active instance, the gateway doesn't emit a workload-host health observation.

Deployment configuration bounds server identities and namespaced tool names for metrics. Values outside those sets become `other`. Profile names follow the catalog bounds and repository `other` rule. Agent launch and session IDs, call and approval IDs, upstream URLs, credential references, arguments, and results never become metric attributes. Audit and telemetry exclude secrets and authentication material. By default, telemetry omits arguments and results under [What telemetry never records](#What%20telemetry%20never%20records).

### Cardinality

A metric attribute takes values from a set that a deployment's configuration bounds. These values never appear as metric attributes:

- The agent launch ID, the recorded agent spec generation, the specification digest, and the profile digest
- Trace and span IDs
- Port numbers and names, preview host names and URLs, session IDs, share-link IDs, and embedded node identities
- Image references, tags, manifest and layer digests, and image build IDs
- Workspace target paths, repository URLs, commits, and branch names
- Hook names from a specification, commands, and working directories
- Environment names and values
- Error messages and any free text

afakt records an attribute value outside its enumerated set as `other`. Labels count as bounded because each deployment defines its task classes and variants. Launch IAM role names count as bounded for the same reason. The names of catalog profiles and their hooks count as bounded too, but a metric records the name of a repository profile, and each hook name in it, as `other`.

## What telemetry never records

No span, event, metric, or log record from afakt or from the telemetry endpoint contains any of these:

- A secret, a credential value, or authentication material, including SPIFFE Verifiable Identity Documents (SVIDs) and their [private keys](pki.md)
- An environment value, from the workload environment or from a hook's own entries
- Hook input, standard output, or standard error, and file contents from the agent workspace

With the profile's telemetry content set to `omit`, which is the default, telemetry also never records the prompt, the rendered instructions, the instruction inputs, model responses, the final answer, or tool arguments and results. The workload provider turns the agent harness's content logging off, and the telemetry endpoint drops any such content that arrives.

With telemetry content set to `include`, the agent harness span may include the opt-in GenAI attributes `gen_ai.input.messages`, `gen_ai.output.messages`, and `gen_ai.system_instructions`, and the harness may log prompt and response content. The preceding list still applies in every case. Evidence holds the full transcript and the final answer either way, under the agent launch's retention.
