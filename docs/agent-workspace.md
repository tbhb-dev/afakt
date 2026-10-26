---
type: Design
title: Agent workspace
description: The workspace sources an execution provider builds an agent's workspace from, and the environment variables that the agent and its hooks see.
tags: [afakt, agents, execution, workspace]
status: draft
generated: { by: "codex", at: "2026-10-27T00:00:00Z" }
---

# Agent workspace

A workspace is the directory tree owned by a durable [agent](agent.md#workspace-revisions-and-the-workspace-registry). Each execution generation has its own workspace materialization, which the provider builds from the specification's workspace sources into the sandbox VM's writable workspace disk before it starts the agent. Guest paths describe the workspace the harness sees, independently of host disk paths. Inside the sandbox, the agent and its hooks also see three sets of environment variables. Information about the execution comes from the execution environment. The agent environment contains the variables that the profile and the specification add, and the trace context ties the harness's telemetry to the execution trace.

How generation copies map to the agent's revision history and survive execution cleanup remains an [open question](agent-open-questions.md#agents).

## Workspace sources

| Kind | Fields | What the provider places at the target path |
| --- | --- | --- |
| `repository` | Repository URL, full commit ID, and an optional new branch name | A checkout of the repository at that commit. With a new branch name, the provider creates that branch at the checked-out commit and switches to it. Without one, the checkout has no branch. |
| `image` | Image reference with a repository and tag or manifest digest | The directory tree from a prebuilt workspace image. See [Agent workspace image](agent-workspace-image.md). |
| `execution` | Execution ID and generation | A copy of the workspace that execution left when it ended. |
| `empty` | None | An empty directory. |

Each source has a target path relative to the workspace root. A target path uses `/` between segments, and none of its segments is empty, `.`, or `..`. The target path `.` means the workspace root itself. No two target paths are the same, and none lies inside another, so a source at `.` is the only source. One workspace can contain more than one repository, such as one at `app` and another at `deps/library`.

- **Fetching.** The provider fetches each repository outside the agent's sandbox. A repository that needs credentials to fetch uses a fetch credential, which the agent host's configuration gives for each repository host. Fetch credentials are separate from the profile's [brokered credentials](agent-sandbox.md#credential-brokering), and they never enter the sandbox.
- **Images.** During the workspace build, the provider resolves a tag once and records its manifest digest on the execution. Retries use that digest. A later generation can request the same digest explicitly. Pulls use registry tokens held by the worker outside the sandbox, and cached content still requires authorization. Layers build an independent ext4 workspace disk containing the tree at the target path, alongside a separate read-only root disk for the guest operating system. Digest caches and copy-on-write disk clones reuse the base without sharing writes. Worker pulls before boot are the leading option, while guest pulls through the mediator remain [open](agent-open-questions.md#agent-workspace-image). The same non-overlap rules apply to images. An image followed by a repository at the same target remains an [open question](agent-open-questions.md#agent-workspace-image).
- **Copies.** An `execution` source names an execution that has ended and whose workspace cleanup hasn't removed. A generation that the new one supersedes qualifies too, because the provider builds the new workspace only after that generation becomes terminal. The copy is independent, so changes in the new workspace leave the earlier one unchanged.
- **Generations.** A higher generation starts from its own sources and never inherits the previous generation's workspace unless it asks. To reuse that workspace, its specification lists an `execution` source with its own execution ID and an earlier generation.
- **Access.** Workspace access applies to the whole workspace once the provider has built it.

If the provider can't build the workspace, such as when a fetch fails or the new branch name already exists in the checkout, the execution ends `failed` with workspace failure as its reason before its agent starts.

## Workspace capture

A local workspace checkpoint is a block delta of the workspace disk taken with the guest filesystem frozen at an idle turn boundary. The [agent checkpoint](agent.md#checkpoints) captures the VM snapshot and harness session state at the same boundary. A portable workspace revision uses an Open Container Initiative (OCI) diff layer and content-hashed file inventory pushed to the registry. Correctness depends on content hashes rather than modification times or sizes. See [Checkpoints and portable revisions](agent-workspace-image.md#checkpoints-and-portable-revisions).

## Execution environment

For the agent and for each hook, the provider sets environment variables with information about the execution. Each name starts with `AFAKT_`. Values are plain text without a trailing newline, and none of them contains a secret.

At launch the provider resolves the [sandbox variables](agent-sandbox.md#sandbox-variables), and the environment variables take their values from them. An environment variable and its sandbox variable always have identical text. Some environment variables, such as the commit of a repository source, have no sandbox variable.

| Variable                           | Sandbox variable                   | Value                                                                                                                                                                         |
| ---------------------------------- | ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `AFAKT_EXECUTION_ID`               | `${execution-id}`                  | Execution ID, as a canonical UUIDv7, such as `019a1a2b-3c4d-7e5f-8a6b-7c8d9e0f1a2b`.                                                                                          |
| `AFAKT_GENERATION`                 | `${generation}`                    | Generation, as a decimal integer.                                                                                                                                             |
| `AFAKT_PROFILE_SOURCE`             | None                               | `catalog` or `repository`.                                                                                                                                                    |
| `AFAKT_PROFILE_NAME`               | `${profile-name}`                  | Name of the pinned profile.                                                                                                                                                   |
| `AFAKT_PROFILE_VERSION`            | `${profile-version}`               | For a catalog profile, its version number as a decimal integer. For a repository profile, the repository URL without its scheme, then `@`, then the full commit ID.           |
| `AFAKT_PROFILE_DIGEST`             | None                               | Profile digest, as the algorithm name, `:`, and the digest in lowercase hexadecimal, such as `sha256:3f5a…`.                                                                  |
| `AFAKT_WORKSPACE`                  | `${workspace}`                     | Absolute path of the workspace root.                                                                                                                                          |
| `AFAKT_SOURCE_COUNT`               | None                               | Number of workspace sources, as a decimal integer.                                                                                                                            |
| `AFAKT_SOURCE_<n>_TARGET`          | None                               | Target path of source `<n>`. Sources count from 1, in the order the specification lists them.                                                                                 |
| `AFAKT_SOURCE_<n>_KIND`            | None                               | `repository`, `image`, `execution`, or `empty`.                                                                                                                                        |
| `AFAKT_SOURCE_<n>_PATH`            | `${workspace:<target path>}`       | Absolute path of source `<n>`.                                                                                                                                                |
| `AFAKT_SOURCE_<n>_REPOSITORY`      | `${repository:<target path>}`      | For a `repository` source, its repository URL without the scheme.                                                                                                             |
| `AFAKT_SOURCE_<n>_REPOSITORY_HOST` | `${repository-host:<target path>}` | For a `repository` source, the host name in its repository URL.                                                                                                               |
| `AFAKT_SOURCE_<n>_COMMIT`          | None                               | For a `repository` source, the full commit ID of its checkout.                                                                                                                |
| `AFAKT_SOURCE_<n>_BRANCH`          | `${branch:<target path>}`          | For a `repository` source with a new branch, the branch name. Unset for a source without one.                                                                                 |
| `AFAKT_SCRATCH`                    | `${scratch}`                       | Absolute path of the execution's scratch directory.                                                                                                                           |
| `AFAKT_API_ADDRESS`                | `${afakt-api}`                     | Host name and port of the afakt API, as `<host>:<port>`. The afakt CLI reads it.                                                                                              |
| `AFAKT_TIME_LIMIT`                 | None                               | Time limit, as a whole number of seconds.                                                                                                                                     |
| `AFAKT_DEADLINE`                   | None                               | When the time limit runs out, as a Coordinated Universal Time (UTC) timestamp in the Request for Comments (RFC) 3339 form with whole seconds, such as `2026-10-25T14:30:00Z`. |
| `AFAKT_PHASE`                      | None                               | `before-agent`, `agent`, `after-agent`, or `before-teardown`.                                                                                                                                    |

`HOME` gets the value of `${home}`. Hooks see a few more variables, which [Hook environment](agent-hooks.md#hook-environment) lists.

The CLI sends control requests through the guest relay. The mediator stamps the agent's identity and adds the execution token outside the guest, so a guest process can't read it or send it anywhere else. Only the provider sets an `AFAKT_` variable, and no profile, specification, or hook entry can add or replace one.

## Agent environment

The agent environment consists of the environment variables that a profile and a specification add for the agent, such as `GOFLAGS` or `NODE_ENV`. Both the profile and the specification have an environment field, a list of entries.

| Entry field | Meaning |
| --- | --- |
| Name | Starts with an uppercase letter or `_`, followed by uppercase letters, digits, and `_`, up to 128 characters. It isn't a [reserved name](#reserved-names). |
| Value | The variable's text, up to 32,768 bytes. It accepts [sandbox variables](agent-sandbox.md#sandbox-variables), which the provider inserts literally, and `$$` stands for a literal `$`. |
| Locked | For a profile entry only. Neither the specification nor a hook may replace a locked entry. |

A locked entry guarantees only a start value. The agent and every hook start with the profile's value, but a running process can still change the variable for the processes it starts, or ignore it.

Environment values never contain secrets, because the configuration digest and the provenance record show the values, and so does evidence. A profile that needs a secret lists it under [credential brokering](agent-sandbox.md#credential-brokering) instead.

### Combining environment entries

At launch, the provider combines the two lists into the agent environment:

- Each name appears at most once in each list.
- A profile entry applies unless the specification lists the same name.
- A specification entry with the name of a profile entry without `locked` replaces that entry's value.
- A specification entry with the name of a locked profile entry fails the launch with `environment-locked`.
- A specification entry with a new name adds that variable.

The agent sees the agent environment together with the execution environment and the trace context, and so does every hook. A hook's own entries apply on top for that hook alone, as [Hook environment](agent-hooks.md#hook-environment) describes.

After variable replacement, the agent environment has one canonical encoding, sorted by name, which the configuration digest includes. The provenance record lists each name and value with its source in the profile or the specification. The provider also publishes the agent environment as evidence. The execution environment and the trace context aren't part of it, because the provider derives them for each attempt.

### Reserved names

The provider sets these names itself or keeps them as the agent host provides them, so no environment entry may use them. An entry with a reserved name fails the launch with `environment-name-reserved`, and the error gives the entry and the name.

| Reserved name | Why |
| --- | --- |
| Any name that starts with `AFAKT_` | The execution environment and the hook variables. |
| `HOME` | The provider sets it to `${home}`. |
| `PATH` | It decides which programs the agent, the hooks, and the afakt CLI run. |
| Any name that starts with `LD_` or `DYLD_`, and `NODE_OPTIONS` | They load code into each process that starts. |
| `HTTP_PROXY`, `HTTPS_PROXY`, `ALL_PROXY`, and `NO_PROXY` | The provider reserves these routing settings. Network enforcement still routes every guest flow through the mediator even if a tool ignores them. |
| `SSL_CERT_FILE`, `SSL_CERT_DIR`, `NODE_EXTRA_CA_CERTS`, `REQUESTS_CA_BUNDLE`, `CURL_CA_BUNDLE`, and `GIT_SSL_CAINFO` | They decide which certificate authorities the sandbox trusts, the broker's among them. |
| Any name that starts with `CLAUDE_`, `ANTHROPIC_`, `CODEX_`, `OPENAI_`, `GEMINI_`, or `GOOGLE_` | The harnesses read them as settings, and the provider writes the harness settings itself. An `apiKey.name` may still use one of these names. |
| `XDG_CONFIG_HOME`, `GIT_CONFIG_GLOBAL`, and `GIT_CONFIG_SYSTEM` | They move the configuration that the harness and its tools read. |
| Each `apiKey.name` in the profile's credential brokering | The variable contains the credential broker's placeholder. |
| `TRACEPARENT`, `TRACESTATE`, and any name that starts with `OTEL_` | The trace context. |

### Trace context

The provider sets the trace context variables for the agent and for each hook, and the harness's own telemetry joins the [execution trace](agent-observability.md) through them.

| Variable | Value |
| --- | --- |
| `TRACEPARENT` | The W3C Trace Context `traceparent` of the span the process runs under: the agent run span for the agent, and the hook's own span for a hook. |
| `TRACESTATE` | The W3C Trace Context `tracestate` that goes with it, when it isn't empty. |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | The address of `${telemetry}`, as an `http://` URL. |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | `http/protobuf`. |
| `OTEL_RESOURCE_ATTRIBUTES` | The execution's resource attributes from [Agent observability](agent-observability.md#resource-attributes). |

The provider also turns on the selected harness's telemetry export in the harness's own settings, in the same step as the harness permissions. Those settings keep prompt and response content out of the harness's telemetry unless the profile opts in, as [What telemetry never records](agent-observability.md#what-telemetry-never-records) describes.
