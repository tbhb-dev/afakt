---
type: Design
title: Agent execution sequence
description: One execution generation that succeeds, from the start of the agent runner workflow to the result, across the workflow orchestrator, the agent host, the API server, and the model provider.
tags: [afakt, agents, execution, workflows]
status: draft
generated: { by: "codex", at: "2026-10-26T00:00:00Z" }
---

# Agent execution sequence

The execution belongs to a durable [agent](agent.md), which retains its conversation after the harness run. This diagram follows one execution generation that succeeds, from the moment the parent workflow starts the agent runner workflow until the parent receives the result. It shows each message between the parts of an execution, where the [lifecycle](agent-execution.md#lifecycle) gives only the steps.

```mermaid
sequenceDiagram
    autonumber
    participant Parent as Parent workflow
    participant Orch as Workflow orchestrator
    participant Runner as Agent runner workflow
    participant Placer as afakt worker on the runner's task queue
    box Agent host
        participant Executor as Agent executor workflow
        participant Worker as afakt worker and provider
        participant Broker as Mediator and credential broker
        participant Relay as Guest relay inside sandbox VM
        participant Harness as Harness and adapter
        participant CLI as afakt CLI
        participant Telemetry as Telemetry endpoint
    end
    participant API as API server
    participant Hub as Session hub in the API server
    participant Reader as Attached reader
    participant Model as Model provider
    participant Repo as Repository host

    Worker->>API: Register the agent host, its host task queue, and its metadata
    API-->>Worker: Lease, which the worker renews

    Parent->>Parent: Record the pinned profile version and the execution ID
    Parent->>Orch: Start the agent runner as a child workflow, with the specification
    Orch->>Runner: Workflow task, on any afakt worker

    Runner->>Orch: Schedule the placement activity
    Orch->>Placer: Placement activity
    Placer->>API: Reserve a slot, with the placement requirements
    API->>API: Match the registry and record the reservation
    API-->>Placer: Agent host and host task queue
    Placer-->>Orch: Agent host and host task queue
    Orch-->>Runner: Placement, recorded in workflow history

    Runner->>Orch: Start the agent executor as a child workflow on the host task queue
    Orch->>Executor: Workflow task, on the placed host's worker

    Executor->>Orch: Schedule status
    Orch->>Worker: Status activity
    Worker-->>Orch: not-found
    Orch-->>Executor: not-found
    Executor->>Orch: Schedule launch
    Orch->>Worker: Launch activity
    Worker->>Worker: Validate against the pinned profile and store its copy
    Worker->>API: Claim the reservation
    API-->>Worker: Slot
    Worker->>Worker: Record the execution as pending
    Worker-->>Orch: pending
    Orch-->>Executor: pending

    Executor->>Orch: Schedule the watch activity
    Orch->>Worker: Watch activity
    Worker->>Repo: Fetch each repository source, outside the sandbox
    Repo-->>Worker: Commits
    Worker->>Worker: Build the workspace
    Worker->>API: Request the execution token
    API-->>Worker: Execution token
    Worker->>Broker: Start per-agent network stack and keep secrets and token outside guest
    Worker->>Worker: Combine the configurations and render the instructions
    Worker->>Worker: Write the harness settings and the hook dispatch entries
    Worker->>Relay: Start inside sandbox VM under relay user on either transport
    Worker->>Worker: Run the before-agent hooks in the sandbox
    Note over Worker: The time limit starts, and the state becomes running
    alt ACP transport
        Relay->>Harness: Start adapter as agent user with TRACEPARENT
        Relay->>Harness: Initialize and open the ACP session
        Relay->>Broker: Register session
        Broker->>Hub: Register over agent host authenticated outbound connection
        Executor->>Hub: Send instructions and prompt together as the only writer
        Hub->>Broker: One prompt
        Broker->>Relay: ACP gate permits sending user prompt
        Relay->>Harness: session/prompt
        Reader->>Hub: Attach with read access
        Hub-->>Reader: Session updates, with no writer rights
    else Headless transport
        Worker->>Harness: Start with the instructions, the prompt, and TRACEPARENT
    end

    par The agent works
        loop Each model request
            Harness->>Broker: Model request with the placeholder
            Broker->>Model: Request with the brokered credential
            Model-->>Broker: Response
            Broker-->>Harness: Response
        end
    and Hooks for harness events
        loop Each harness event with hooks
            Harness->>CLI: Native hook input
            CLI->>CLI: Run the matching hooks in combined order
            CLI-->>Harness: Combined decision
        end
    and Capabilities
        loop Each afakt API operation
            Harness->>CLI: Run an afakt CLI command
            CLI->>Relay: API control request
            Relay->>Broker: Control request
            Broker->>Broker: Stamp agent identity and check policy
            Broker->>API: Request with the execution token
            API->>API: Check the capabilities and scopes
            API-->>Broker: Response
            Broker-->>Relay: Response
            Relay-->>CLI: Response
            CLI-->>Harness: Output
        end
    and Telemetry
        Harness->>Broker: OpenTelemetry data
        Broker->>Telemetry: Data, with the connection's execution
        Telemetry->>Telemetry: Add the execution ID and drop unlisted attributes
    and ACP agent requests
        opt ACP transport
            Harness->>Relay: session/request_permission
            Relay->>Broker: Gate permission request and profile answer
            Broker-->>Relay: Policy decision
            Relay-->>Harness: Allow only profile grants permitted by gate, reject the rest
            Harness->>Relay: elicitation/create
            Relay->>Broker: Gate elicitation and cancellation
            Broker-->>Relay: Policy decision
            Relay-->>Harness: Cancel elicitation
        end
    and Heartbeats
        loop While the agent runs
            Worker->>Orch: Watch activity heartbeat
        end
    end

    alt ACP transport
        Harness-->>Relay: Last agent message and end_turn
        Relay->>Broker: Turn ended with final answer and stop reason
        Broker->>Hub: ACP gate permits agent update
        Hub-->>Executor: Turn ended
        Relay->>Harness: Close the ACP session
    else Headless transport
        Harness-->>Worker: Exit with success
    end
    Worker->>Worker: Run session-end hooks after exit or session closure
    Worker->>Worker: Verify termination of every process, including relay and adapter on ACP
    Worker->>Worker: Validate the final answer against the output schema
    Worker->>Worker: Run the after-agent hooks
    Worker->>Worker: Record the workspace changes and publish the evidence
    Worker->>Worker: Run the before-teardown hooks
    Worker->>Worker: Record the result, and the state becomes succeeded
    Worker->>Broker: Stop writing credentials for the execution
    Worker->>API: Free the slot
    Worker->>Worker: Record the execution-ended event
    Worker-->>Orch: Watch activity ends with succeeded
    Orch-->>Executor: Watch activity result

    Executor->>Orch: Schedule result
    Orch->>Worker: Result activity
    Worker-->>Orch: Execution result
    Orch-->>Executor: Execution result
    Executor-->>Orch: Workflow ends with the execution result
    Orch-->>Runner: Child workflow result
    Runner-->>Orch: Workflow ends with the execution result
    Orch-->>Parent: Child workflow result
```

- **Activities** go to the agent host only through the workflow orchestrator. The afakt worker on the host polls the orchestrator for them, and no other part connects to the host.
- **Workflows.** Any afakt worker can run the agent runner workflow. The agent executor workflow runs on the placed host's worker, and each activity it schedules goes to that host's task queue.
- **Placement** runs on the agent runner workflow's task queue and asks the API server's reservation API for a slot. The agent executor workflow then starts on the host task queue of the agent host that holds the reservation. See [Agent placement](agent-placement.md).
- **Status before launch** is the reconciliation rule, so that a retried activity and a step run again after a worker restart both continue from the existing execution and never start a second agent. See [Reconciliation after restart or retry](agent-execution.md#reconciliation-after-restart-or-retry).
- **The sandbox** confines the harness, the afakt CLI, and every hook. Each network request from the sandbox VM goes through the mediator. The credential broker injects credentials outside the guest. See [Agent sandbox](agent-sandbox.md).
- **Harness transport** comes from the selected profile model entry. ACP uses the guest relay and adapter inside the sandbox VM, with the mediator applying its ACP gate between relay and hub. Headless starts the harness directly. ACP executions allow readers and keep the agent executor workflow as the only writer. See [Harness transport](agent-execution.md#harness-transport).
- **Hook dispatch** runs the hooks for harness events through the harness's own hook system and the afakt CLI on both transports. See [Hook dispatch](agent-hooks.md#hook-dispatch).
- **Telemetry** from the harness joins the execution trace through the telemetry endpoint. See [Trace context propagation](agent-observability.md#trace-context-propagation).
- **After the terminal state,** the API server rejects the execution token. Cleanup runs after the retention period, outside this sequence. See [Evidence, retention, and cleanup](agent-execution-model.md#evidence-retention-and-cleanup).
