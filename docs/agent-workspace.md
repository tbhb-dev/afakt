---
type: Design
title: Agent workspace
description: The workspace sources an agent sandbox provider builds an agent workspace from, and the environment variables that the agent and its hooks see.
tags: [afakt, agents, agent-launch, workspace]
status: draft
generated: { by: "codex:gpt-6", at: "2026-10-28T17:30:00Z" }
---

# Agent workspace

[Public key infrastructure](pki.md).

An agent workspace is the filesystem mounted into an agent sandbox. It belongs to a durable [agent](agent.md#Workspace%20revisions%20and%20the%20workspace%20registry). Each agent launch has an independent workspace materialization built from its agent launch spec before the agent harness starts. Guest paths describe the filesystem the harness sees. The agent launch environment identifies the launch, the agent environment supplies configured variables, and trace context joins telemetry to the launch trace.

How agent launch copies map to the agent's revision history and survive agent launch cleanup remains an [open question](agent-open-questions.md#Agents).

## Workspace sources

| Kind | Fields | What the provider places at the target path |
| --- | --- | --- |
| `repository` | Repository uniform resource locator (URL), full commit ID, and an optional new branch name | A checkout of the repository at that commit. With a new branch name, the provider creates that branch at the checked-out commit and switches to it. Without one, the checkout has no branch. |
| `image` | Image reference with a repository and tag or manifest digest | The directory tree from a prebuilt workspace image. See [Agent workspace image](agent-workspace-image.md). |
| `agent-launch` | Agent launch ID | A copy of the agent workspace that agent launch left when it ended. |
| `empty` | None | An empty directory. |

Each source has a target path relative to the workspace root. A target path uses `/` between segments, and none of its segments is empty, `.`, or `..`. The target path `.` means the workspace root itself. No two target paths are the same, and none lies inside another, so a source at `.` is the only source. One workspace can contain more than one repository, such as one at `app` and another at `deps/library`.

- **Fetching.** The provider fetches each repository outside the agent's agent sandbox. A repository that needs credentials to fetch uses a fetch credential, which the agent host's configuration gives for each repository host. Fetch credentials are separate from the profile's [brokered credentials](agent-sandbox.md#Credential%20brokering), and they never enter the sandbox.
- **Images.** [Image sources and digest pinning](agent-workspace-lifecycle.md#Image%20sources%20and%20digest%20pinning) and [Agent host caches and startup](agent-workspace-lifecycle.md#Agent%20host%20caches%20and%20startup). Worker pulls before boot are the leading option, while guest pulls remain [open](agent-open-questions.md#%60OQ-IMG-8%60).
- **Copies and agent launches.** [Creation and pinned inputs](agent-workspace-lifecycle.md#Creation%20and%20pinned%20inputs).
- **Access.** Workspace access applies to the whole workspace once the provider has built it.

[States and transitions](agent-workspace-lifecycle.md#States%20and%20transitions) covers workspace preparation failures.

## Agent workspace capture

[Workspace disk deltas and portable revisions](agent-workspace-lifecycle.md#Workspace%20disk%20deltas%20and%20portable%20revisions).

## Agent launch environment

For the agent and for each hook, the provider sets environment variables with information about the agent launch. Each name starts with `AFAKT_`. Values are plain text without a trailing newline, and none of them contains a secret.

At launch the provider resolves the [agent sandbox variables](agent-sandbox.md#Sandbox%20variables), and the environment variables take their values from them. An environment variable and its sandbox variable always have identical text. Some environment variables, such as the commit of a repository source, have no sandbox variable.

| Variable                           | Sandbox variable                   | Value                                                                                                                                                                         |
| ---------------------------------- | ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AFAKT_AGENT_LAUNCH_ID`               | `${agent-launch-id}`                  | Agent launch ID, as a universally unique identifier (UUID) in canonical UUIDv7 form, such as `019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b`.                                                                                          |
| `AFAKT_GENERATION`                 | `${generation}`                    | The agent spec generation recorded at finalization, as a decimal integer.                                                                                                                                             |
| `AFAKT_PROFILE_SOURCE`             | None                               | `catalog` or `repository`.                                                                                                                                                    |
| `AFAKT_PROFILE_NAME`               | `${profile-name}`                  | Name of the pinned profile.                                                                                                                                                   |
| `AFAKT_PROFILE_VERSION`            | `${profile-version}`               | For a catalog profile, its version number as a decimal integer. For a repository profile, the repository URL without its scheme, then `@`, then the full commit ID.           |
| `AFAKT_PROFILE_DIGEST`             | None                               | Profile digest, as the algorithm name, `:`, and the digest in lowercase hexadecimal, such as `sha256:3f5a…`.                                                                  |
| `AFAKT_WORKSPACE`                  | `${workspace}`                     | Absolute path of the workspace root.                                                                                                                                          |
| `AFAKT_SOURCE_COUNT`               | None                               | Number of workspace sources, as a decimal integer.                                                                                                                            |
| `AFAKT_SOURCE_<n>_TARGET`          | None                               | Target path of source `<n>`. Sources count from 1, in the order the specification lists them.                                                                                 |
| `AFAKT_SOURCE_<n>_KIND`            | None                               | `repository`, `image`, `agent-launch`, or `empty`.                                                                                                                                        |
| `AFAKT_SOURCE_<n>_PATH`            | `${workspace:<target path>}`       | Absolute path of source `<n>`.                                                                                                                                                |
| `AFAKT_SOURCE_<n>_REPOSITORY`      | `${repository:<target path>}`      | For a `repository` source, its repository URL without the scheme.                                                                                                             |
| `AFAKT_SOURCE_<n>_REPOSITORY_HOST` | `${repository-host:<target path>}` | For a `repository` source, the host name in its repository URL.                                                                                                               |
| `AFAKT_SOURCE_<n>_COMMIT`          | None                               | For a `repository` source, the full commit ID of its checkout.                                                                                                                |
| `AFAKT_SOURCE_<n>_BRANCH`          | `${branch:<target path>}`          | For a `repository` source with a new branch, the branch name. Unset for a source without one.                                                                                 |
| `AFAKT_SCRATCH`                    | `${scratch}`                       | Absolute path of the agent launch's scratch directory.                                                                                                                           |
| `AFAKT_API_ADDRESS`                | `${afakt-api}`                     | Host name and port of the afakt Application Programming Interface (API), as `<host>:<port>`. The afakt command-line interface (CLI) reads it.                                                                                              |
| `AFAKT_TIME_LIMIT`                 | None                               | Time limit, as a whole number of seconds.                                                                                                                                     |
| `AFAKT_DEADLINE`                   | None                               | When the time limit runs out, as a Coordinated Universal Time (UTC) timestamp in the Request for Comments (RFC) 3339 form with whole seconds, such as `2026-10-25T14:30:00Z`. |
| `AFAKT_PHASE`                      | None                               | `before-agent`, `agent`, `after-agent`, or `before-teardown`.                                                                                                                                    |

`HOME` gets the value of `${home}`. Hooks see a few more variables, which [Hook environment](agent-hooks.md#Hook%20environment) lists.

The CLI sends control requests through the guest relay. The agent mediator stamps the agent's identity. The agent proxy authenticates with the launch's X.509 SPIFFE Verifiable Identity Document (SVID) over mutual Transport Layer Security (TLS) on the overlay. TLS ends at the API server. See [Per-launch services](agent-workload-proxy.md#Launch%20boundary). Neither an SVID nor its private key enters the guest. Only the provider sets an `AFAKT_` variable, and no profile, specification, or hook entry can add or replace one.

## Agent environment

An agent profile and the specification for an agent launch add environment variables, such as `GOFLAGS` or `NODE_ENV`, to the agent environment. Both the profile and the specification have an environment field, a list of entries.

| Entry field | Meaning |
| --- | --- |
| Name | Starts with an uppercase letter or `_`, followed by uppercase letters, digits, and `_`, up to 128 characters. It isn't a [reserved name](#Reserved%20names). |
| Value | The variable's text, up to 32,768 bytes. It accepts [agent sandbox variables](agent-sandbox.md#Sandbox%20variables), which the provider inserts literally, and `$$` stands for a literal `$`. |
| Locked | For a profile entry only. Neither the specification nor a hook may replace a locked entry. |

A locked entry guarantees only a start value. The agent and every hook start with the profile's value, but a running process can still change the variable for the processes it starts, or ignore it.

Environment values never contain secrets, because the configuration digest and the provenance record show the values, and so does evidence. A profile that needs a secret lists it under [credential brokering](agent-sandbox.md#Credential%20brokering) instead.

### Combining environment entries

At agent launch, the agent sandbox provider combines the two lists into the agent environment:

- Each name appears at most once in each list.
- A profile entry applies unless the specification lists the same name.
- A specification entry with the name of a profile entry without `locked` replaces that entry's value.
- A specification entry with the name of a locked profile entry fails the launch with `environment-locked`.
- A specification entry with a new name adds that variable.

The agent sees the agent environment together with the agent launch environment and the trace context, and so does every hook. A hook's own entries apply on top for that hook alone, as [Hook environment](agent-hooks.md#Hook%20environment) describes.

After variable replacement, the agent environment has one canonical encoding, sorted by name, which the configuration digest includes. The provenance record lists each name and value with its source in the profile or the specification. The provider also publishes the agent environment as evidence. The agent launch environment and the trace context aren't part of it, because the provider derives them for each attempt.

### Reserved names

The provider sets these names itself or keeps them as the agent host provides them, so no environment entry may use them. An entry with a reserved name fails the launch with `environment-name-reserved`, and the error gives the entry and the name.

| Reserved name | Why |
| --- | --- |
| Any name that starts with `AFAKT_` | The agent launch environment and the hook variables. |
| `HOME` | The provider sets it to `${home}`. |
| `PATH` | It decides which programs the agent, the hooks, and the afakt CLI run. |
| Any name that starts with `LD_` or `DYLD_`, and `NODE_OPTIONS` | They load code into each process that starts. |
| `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, and `NO_PROXY` | The provider reserves these routing settings. Network enforcement still routes every guest flow through the agent proxy under mediator decisions even if a tool ignores them. |
| `SSL_CERT_FILE`, `SSL_CERT_DIR`, `NODE_EXTRA_CA_CERTS`, `REQUESTS_CA_BUNDLE`, `CURL_CA_BUNDLE`, and `GIT_SSL_CAINFO` | They decide which certificate authorities the agent sandbox trusts, including the [inspection authority](pki-transport.md#Inspection%20certificates). |
| Any name that starts with `CLAUDE_`, `ANTHROPIC_`, `CODEX_`, `OPENAI_`, `GEMINI_`, or `GOOGLE_` | The harnesses read them as settings, and the provider writes the agent harness settings itself. An `apiKey.name` may still use one of these names. |
| `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL`, and `GIT_CONFIG_SYSTEM` | They move the configuration that the harness and its tools read. |
| Each `apiKey.name` in the profile's credential brokering | The variable contains the credential broker's placeholder. |
| `TRACEPARENT`, `TRACESTATE`, and any name that starts with `OTEL_` | The trace context. |

### Trace context

The provider sets the trace context variables for the agent and for each hook, and the agent harness's own telemetry joins the [trace of the agent launch](agent-observability.md) through them.

| Variable | Value |
| --- | --- |
| `TRACEPARENT` | The W3C Trace Context `traceparent` of the span the process runs under: the agent harness span for the agent, and the hook's own span for a hook. |
| `TRACESTATE` | The W3C Trace Context `tracestate` that goes with it, when it isn't empty. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The address of `${telemetry}`, as an `http://` URL. |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf`. |
| `OTEL_RESOURCE_ATTRIBUTES` | The agent launch's resource attributes from [Agent observability](agent-observability.md#Resource%20attributes). |

The provider also turns on the selected harness's telemetry export in the harness's own settings, in the same step as the harness permissions. Those settings keep prompt and response content out of the harness's telemetry unless the profile opts in, as [What telemetry never records](agent-observability.md#What%20telemetry%20never%20records) describes.
