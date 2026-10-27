workspace "afakt" "Architecture of the afakt modular monolith." {
    properties {
        // Documentation and decisions stay in the repository's docs bundle.
        structurizr.inspection.model.softwaresystem.documentation ignore
        structurizr.inspection.model.softwaresystem.decisions ignore
    }

    model {
        developer = person "Developer or tester" "Runs afakt and inspects its workflows."
        github = softwareSystem "GitHub" "Delivers webhooks and exposes the GitHub API."
        native = softwareSystem "Native apps" "Clients of afakt's REST API."
        agents = softwareSystem "AI coding agents" "Perform coding work started by Agent runner."
        temporal = softwareSystem "External Temporal" "Provides durable execution when workers use an external server."

        afakt = softwareSystem "afakt" "Coordinates coding work through its API and workflows." {
            runtime = container "afakt process" "Runs the selected runtime roles together or separately from the same binary." "Go" {
                cli = component "Command-line interface" "Runs server roles and communicates with the API." "Go"
                api = component "API server" "Serves the REST API and starts workflows." "Go"
                webhook = component "GitHub webhook receiver" "Accepts webhook deliveries and starts coordination work." "Go"
                dashboard = component "Web dashboard" "Provides the web UI from the API's origin." "React"
                orchestrator = component "Embedded workflow orchestrator" "Provides durable execution inside the process for development and tests." "Temporal"
                workflows = component "Coordination workflows" "Coordinate work on workers inside the binary and start other workflows." "Temporal"
                runner = component "Agent runner" "Runs AI coding agents as a workflow." "Temporal"
                temporalUI = component "Temporal Web UI" "Shows the embedded server's namespaces, workflows, and histories." "Temporal"
            }
        }

        developer -> cli "Runs commands" "CLI"
        developer -> dashboard "Uses the dashboard" "HTTP"
        developer -> temporalUI "Inspects workflows" "HTTP"
        native -> api "Calls the REST API" "HTTP"
        cli -> api "Calls the REST API" "HTTP"
        dashboard -> api "Calls the REST API" "HTTP"
        github -> webhook "Delivers webhooks" "HTTP"
        webhook -> api "Calls the REST API" "HTTP"
        webhook -> orchestrator "Starts workflows in embedded mode" "Temporal"
        webhook -> temporal "Starts workflows in external mode" "Temporal"
        api -> orchestrator "Starts workflows in embedded mode" "Temporal"
        api -> temporal "Starts workflows in external mode" "Temporal"
        api -> github "Calls the GitHub API" "HTTP"
        workflows -> api "Calls the REST API" "HTTP"
        workflows -> github "Calls the GitHub API" "HTTP"
        workflows -> orchestrator "Uses durable execution in embedded mode" "Temporal"
        workflows -> temporal "Uses durable execution in external mode" "Temporal"
        workflows -> runner "Starts agent work" "Temporal"
        runner -> agents "Runs coding agents" "Process"
        temporalUI -> orchestrator "Reads namespaces and workflow histories" "Temporal"
    }

    views {
        properties {
            plantuml.skinparams "nodesep=80,ranksep=160"
        }
        systemContext afakt "afakt-context" "afakt and the people and systems it interacts with." {
            include *
            autoLayout lr
        }
        container afakt "afakt-containers" "The afakt process and its external dependencies." {
            include *
            autoLayout lr
        }
        component runtime "afakt-components" "The roles inside the combined afakt process." {
            include *
            autoLayout lr
        }
    }

    configuration {
        scope softwaresystem
    }
}
