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
            runtime = container "afakt process" "Runs the selected runtime services together or separately from the same binary." "Go" {
                cli = component "Command-line interface" "Runs server services and communicates with the API." "Go"
                api = component "API server" "Serves the REST API and starts workflows." "Go"
                webhook = component "GitHub webhook receiver" "Accepts webhook deliveries and starts coordination work." "Go"
                dashboard = component "Web dashboard" "Provides the web UI from the API's origin." "React"
                orchestrator = component "Embedded workflow orchestrator" "Provides durable execution inside the process for development and tests." "Temporal"
                workflows = component "Coordination workflows" "Coordinate work on workers inside the binary and start other workflows." "Temporal"
                runner = component "Agent runner" "Runs AI coding agents as a workflow." "Temporal"
                temporalUI = component "Temporal Web UI" "Shows the embedded server's namespaces, workflows, and histories." "Temporal"
            }

            group "Control plane services" {
                controlApi = container "API server and web dashboard" "Authorizes operations and serves the dashboard from its own origin." "Application service" "Topology"
                previewServer = container "Preview server" "Dials workload proxies over the overlay; published-port listener binding remains open." "Application service" "Topology"
                mcpGateway = container "MCP gateway" "Checks launch tool grants and keeps upstream credentials outside guests." "Application service" "Topology"
                staticOrigin = container "Static origin" "Serves static site and single-page app deploys from local copies of their files; keeps no cache." "Caddy" "Topology"
                coordination = container "Overlay coordination" "An afakt interface fronts isolated Headscale coordination; cluster nodes retain a static map of stable nodes." "Headscale" "Topology"
                overlayRelay = container "Overlay relay" "Carries encrypted overlay traffic when a relay path is needed." "Overlay transport" "Topology"
            }
            group "External ingress boundary" {
                edge = container "External ingress gateway" "Accepts public and attached-tailnet application requests; never routes packets between tailnets and the cluster." "Application proxy" "Topology"
            }
            group "Host management boundary" {
                hostWorker = container "Workload host worker" "Controls launch lifecycle as a dedicated user; stays outside every launch data path." "Host service" "Topology"
                helper = container "Privileged helper" "Performs narrowly authorized setup and launch operations; caller binding is undecided." "Host privilege boundary" "Topology"
                hostResolver = container "Host name resolver" "Resolves names under *.afakt.internal for host clients." "Domain Name System" "Topology"
            }
            group "Operator access on the workload host" {
                loopbackProxy = container "Host access boundary" "Host-level authorization for services, previews and ports; local caller and connection mechanisms undecided." "Access mechanism undecided" "Topology"
                hostNetwork = container "Host-level connectivity" "Host authorization for agent and cluster services; connection mechanism undecided; optional user tailnet attachment." "Host connectivity service" "Topology"
            }
            group "Launch guest boundary" {
                guest = container "Guest context" "Untrusted harness, relay and loopback listeners; no overlay membership or workload keys." "Workload provider confinement" "Topology"
            }
            group "Launch mediation boundary" {
                mediator = container "Workload proxy mediator" "Decides every inbound and outbound connection and stamps identity; never carries traffic." "Local policy enforcement" "Topology"
            }
            group "Launch workload proxy boundary" {
                workload-proxy = container "Workload proxy" "Carries workload traffic in both directions; own overlay node and launch X.509-SVID; asks the mediator to decide each connection." "Network proxy" "Topology"
            }
            group "Destination launch guest boundary" {
                destinationGuest = container "Destination guest" "Untrusted harness and published loopback listener; no overlay membership or workload keys." "Workload provider confinement" "Topology"
            }
            group "Destination launch mediation boundary" {
                destinationMediator = container "Destination mediator" "Checks the source grant and current destination launch before listener access; no independent network." "Local policy enforcement" "Topology"
            }
            group "Destination launch network boundary" {
                destination-workload-proxy = container "Destination workload proxy" "Carries destination workload traffic under mediator decisions; own overlay node and launch X.509-SVID." "Network proxy" "Topology"
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

        // Existing implied relationships precede this point and remain unchanged.
        !impliedRelationships false
        group "Operator devices" {
            browser = softwareSystem "Operator browser" "A browser under the operator's local user identity." "Topology"
            phone = softwareSystem "Phone and native clients" "Operator devices using their attached user tailnet and afakt sign-in." "Topology"
        }
        group "User network boundary" {
            userTailnet = softwareSystem "Attached user tailnet" "User-controlled private connectivity; separate from the cluster overlay." "Topology"
        }
        group "External service boundary" {
            internet = softwareSystem "Internet services" "Permitted upstream destinations outside the cluster trust domain." "Topology"
        }

        browser -> hostResolver "Resolves *.afakt.internal" "Local name query" "TopologyConnection,OverviewHidden"
        browser -> loopbackProxy "Requests service, preview or port" "Host access mechanism undecided" "TopologyConnection,OverviewHidden"
        loopbackProxy -> hostNetwork "Uses host-level authorization" "Connection mechanism undecided" "TopologyConnection,OverviewHidden"
        hostNetwork -> previewServer "Requests preview as host identity" "Workload mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        hostWorker -> helper "Requests authorized setup and launch operations" "Authenticated local control" "TopologyConnection"
        hostWorker -> hostResolver "Requires host name resolution" "Local control" "TopologyConnection,OverviewHidden"
        hostWorker -> mediator "Starts and manages launch; no data forwarding" "Local lifecycle control" "LifecycleDetail,TopologyConnection,OverviewHidden"
        hostWorker -> destinationMediator "Starts and manages destination launch; no data forwarding" "Local lifecycle control" "LifecycleDetail,TopologyConnection,OverviewHidden"
        hostWorker -> hostNetwork "Sends host management requests" "Local channel" "TopologyConnection"
        hostNetwork -> controlApi "Registers host and reports capacity as host identity" "Workload mutual TLS over overlay" "TopologyConnection"
        phone -> userTailnet "Connects with device identity" "User tailnet" "TopologyConnection,OverviewHidden"
        userTailnet -> edge "Delivers application requests" "Transport Layer Security" "ClusterAttachment,OverviewHidden"
        userTailnet -> hostNetwork "Optional host attachment; application access only" "User tailnet" "ClusterAttachment,OverviewHidden"
        browser -> edge "Sends public application requests as signed-in user" "Transport Layer Security" "TopologyConnection,OverviewHidden"
        edge -> controlApi "Forwards authenticated dashboard and API requests" "Application protocol over overlay" "TopologyConnection,OverviewHidden"
        edge -> previewServer "Forwards preview requests with user context" "Application protocol over overlay" "TopologyConnection,OverviewHidden"
        edge -> staticOrigin "Forwards static site requests by hostname" "Application protocol over overlay" "TopologyConnection,OverviewHidden"
        edge -> workload-proxy "Dials permitted workload ingress directly" "SPIFFE mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        previewServer -> workload-proxy "Dials workload proxy; published-port binding open" "SPIFFE mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        mcpGateway -> controlApi "Calls permitted operations with launch context" "Authenticated service connection" "TopologyConnection"
        guest -> workload-proxy "Sends workload traffic including DNS, TCP, UDP and QUIC" "Private workload link" "TopologyConnection,OverviewHidden"
        mediator -> workload-proxy "Returns connection decision and stamped launch identity" "Local channel" "TopologyConnection,OverviewHidden"
        workload-proxy -> controlApi "Calls authorized operations with launch X.509-SVID" "Workload mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        workload-proxy -> mcpGateway "Calls allowed tools with launch X.509-SVID" "Workload mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        workload-proxy -> internet "Opens permitted egress; inspection where required" "Destination protocol" "TopologyConnection,OverviewHidden"
        workload-proxy -> mediator "Asks for inbound and outbound connection decisions" "Local channel" "TopologyConnection,ReturnDetail,OverviewHidden"
        workload-proxy -> guest "Carries permitted ingress; published-port binding open" "Private workload link" "TopologyConnection,ReturnDetail,OverviewHidden"
        destinationGuest -> destination-workload-proxy "Sends destination workload traffic" "Private workload link" "TopologyConnection,OverviewHidden"
        destinationMediator -> destination-workload-proxy "Returns destination connection decision" "Local channel" "TopologyConnection,OverviewHidden"
        destination-workload-proxy -> destinationMediator "Requests decision with source identity and target grant" "Local channel" "TopologyConnection,ReturnDetail,OverviewHidden"
        destination-workload-proxy -> destinationGuest "Carries permitted ingress; published-port binding open" "Private workload link" "TopologyConnection,ReturnDetail,OverviewHidden"
        workload-proxy -> destination-workload-proxy "Dials directly; both mediators check the connection" "SPIFFE mutual TLS over overlay" "TopologyConnection,OverviewHidden"
        workload-proxy -> coordination "Obtains overlay coordination through afakt interface" "Overlay control" "TopologyConnection,OverviewHidden"
        hostNetwork -> coordination "Obtains overlay coordination through afakt interface" "Overlay control" "TopologyConnection,OverviewHidden"
        workload-proxy -> overlayRelay "Uses encrypted relay transport when needed" "Overlay transport" "TopologyConnection,OverviewHidden"

        deploymentEnvironment "Topology context" {
            context_operatorDeviceBoundaryNode = deploymentNode "Operator device boundary" "Operator-controlled clients." "Client devices" {
                contextClients = infrastructureNode "Browser, phone, tablet and native apps" "Operator and device identities." "Application clients"
            }
            context_userTailnetBoundaryNode = deploymentNode "User tailnet boundary" "User-owned network; no cluster routing." "User network" {
                contextTailnet = infrastructureNode "Attached user tailnets" "Private application access to attached gateways." "Tailscale"
            }
            context_clusterTrustDomainNode = deploymentNode "Cluster trust domain" "Control plane and workload hosts share authenticated overlay reachability." "Private cluster" {
                context_externalIngressBoundaryNode = deploymentNode "External ingress boundary" "Application entry only." "Gateway node" {
                    contextIngress = infrastructureNode "External ingress gateway" "Public and user-tailnet application traffic; no packet routing." "Application proxy"
                }
                contextCluster = infrastructureNode "Control plane and workload hosts" "Per-service, per-host and per-launch nodes; guests have no overlay membership." "Isolated Headscale overlay"
            }
            context_internetBoundaryNode = deploymentNode "Internet boundary" "Upstream destinations outside cluster trust." "External network" {
                contextInternet = infrastructureNode "Internet services" "Only permitted egress leaves a workload host." "Destination services"
            }
            contextClients -> contextTailnet "Device identity" "User tailnet"
            contextClients -> contextIngress "Signed-in user" "Public application request"
            contextTailnet -> contextIngress "User context; no packet routing" "Private application request"
            contextIngress -> contextCluster "Authenticated application request" "Cluster overlay"
            contextCluster -> contextInternet "Permitted destination request" "Workload proxy egress"
        }
        deploymentEnvironment "Topology cluster" {
            cluster_userTailnetBoundaryNode = deploymentNode "User tailnet boundary" "Separate user network." "User tailnet" {
                clusterTailnet = softwareSystemInstance userTailnet
            }
            cluster_externalIngressBoundaryNode = deploymentNode "External ingress boundary" "User-tailnet attachments do not route cluster packets." "Gateway node" {
                clusterEdge = containerInstance edge
            }
            cluster_clusterOverlayBoundaryNode = deploymentNode "Cluster overlay boundary" "Tailscale tsnet nodes use isolated Headscale; every node retains a static stable-node map." "Private overlay" {
                clusterOverlay = infrastructureNode "Private overlay reachability" "Each cluster node retains a static map of stable nodes. Ingress rules admit only the gateway and named peers." "Tailscale tsnet"
                cluster_controlPlaneBoundaryNode = deploymentNode "Control plane boundary" "One overlay node for each service." "Control plane" {

                    cluster_apiServiceNodeNode = deploymentNode "API service node" "Independent service identity; dashboard shares its origin." "tsnet node" {
                        cluster_controlApi = containerInstance controlApi
                    }
                    cluster_previewServiceNodeNode = deploymentNode "Preview service node" "Independent service identity." "tsnet node" {
                        cluster_previewServer = containerInstance previewServer
                    }
                    cluster_mcpServiceNodeNode = deploymentNode "MCP service node" "Independent service identity." "tsnet node" {
                        cluster_mcpGateway = containerInstance mcpGateway
                    }
                    cluster_staticOriginNodeNode = deploymentNode "Static origin node" "Independent service identity." "tsnet node" {
                        cluster_staticOrigin = containerInstance staticOrigin
                    }
                    cluster_coordinationNodeNode = deploymentNode "Coordination node" "Isolated Headscale behind an afakt interface; instance count undecided." "Coordination service" {
                        cluster_coordination = containerInstance coordination
                    }
                    cluster_relayNodeNode = deploymentNode "Relay node" "Encrypted overlay transport." "Relay service" {
                        cluster_overlayRelay = containerInstance overlayRelay
                    }
                }
                cluster_agentHostBoundaryNode = deploymentNode "Workload host boundary" "Host management is separate from launch traffic." "Workload host" {
                    cluster_hostNodeNode = deploymentNode "Host node" "Host identity; optional user-tailnet attachment." "tsnet node" {
                        cluster_hostNetwork = containerInstance hostNetwork
                    }
                    cluster_agentLaunchBoundaryNode = deploymentNode "Agent launch boundary" "Independent mediator, workload proxy and guest; no shared launch data path." "Launch isolation" {
                        cluster_workload-proxy = containerInstance workload-proxy
                    }
                }
                cluster_anotherAgentHostBoundaryNode = deploymentNode "Another workload host boundary" "Independent host and launch identities." "Workload host" {
                    cluster_anotherHostNodeNode = deploymentNode "Another host node" "Independent host identity; optional attached user tailnet." "tsnet node" {
                        cluster_otherHost = containerInstance hostNetwork
                    }
                    cluster_anotherLaunchBoundaryNode = deploymentNode "Another launch boundary" "Independent launch connection." "Launch isolation" {
                        cluster_destination-workload-proxy = containerInstance destination-workload-proxy
                    }
                }
            }
            clusterEdge -> clusterOverlay "Authenticated app requests" "Overlay access"
            cluster_controlApi -> clusterOverlay "Service identity" "Overlay membership"
            cluster_previewServer -> clusterOverlay "Service identity" "Overlay membership"
            cluster_mcpGateway -> clusterOverlay "Service identity" "Overlay membership"
            cluster_staticOrigin -> clusterOverlay "Service identity" "Overlay membership"
            cluster_coordination -> clusterOverlay "Node coordination" "afakt interface"
            cluster_overlayRelay -> clusterOverlay "Encrypted transport" "Relay path"
            cluster_hostNetwork -> clusterOverlay "Host identity" "Overlay membership"
            cluster_otherHost -> clusterOverlay "Host identity" "Overlay membership"
            cluster_workload-proxy -> clusterOverlay "Launch identity" "Coarse service access"
            cluster_destination-workload-proxy -> clusterOverlay "Destination launch identity" "Coarse service access"
        }
        deploymentEnvironment "Topology host" {
            host_agentHostTrustBoundaryNode = deploymentNode "Workload host trust boundary" "Privileged host control remains outside every guest." "Workload host" {
                host_dedicatedWorkerUserNode = deploymentNode "Dedicated worker user" "Separate from the host-mode agent user." "User boundary" {
                    host_hostWorker = containerInstance hostWorker
                }
                host_privilegedSetupBoundaryNode = deploymentNode "Privileged setup boundary" "Helper authority and caller binding remain undecided." "Privilege boundary" {
                    host_helper = containerInstance helper
                }
                host_hostResolver = containerInstance hostResolver
                host_operatorAccessBoundaryNode = deploymentNode "Operator access boundary" "Host access requires authorization; name resolution grants none." "Host service" {
                    host_loopbackProxy = containerInstance loopbackProxy
                    host_hostNetwork = containerInstance hostNetwork
                }
                host_agentLaunchANode = deploymentNode "Agent launch A" "Independent data path; mediation and network users undecided." "Launch boundary" {
                    host_guestANode = deploymentNode "Guest A" "Host mode uses the dedicated agent user, separate from the worker." "Dedicated agent user in host mode" {
                        host_guest = containerInstance guest
                    }
                    host_mediatorANode = deploymentNode "Mediator A" "Policy boundary; no independent network." "Local confinement; user undecided" {
                        host_mediator = containerInstance mediator
                    }
                    host_agentProxyANode = deploymentNode "Workload proxy A" "Own overlay node and launch X.509-SVID; carries workload traffic." "tsnet node; user undecided" {
                        host_workload-proxy = containerInstance workload-proxy
                    }
                }
                host_agentLaunchBNode = deploymentNode "Agent launch B" "Independent data path; same-host traffic uses the overlay." "Launch boundary" {
                    host_guestBNode = deploymentNode "Guest B" "Host mode uses the dedicated agent user, separate from the worker." "Dedicated agent user in host mode" {
                        host_destinationGuest = containerInstance destinationGuest
                    }
                    host_mediatorBNode = deploymentNode "Mediator B" "Policy boundary; no independent network." "Local confinement; user undecided" {
                        host_destinationMediator = containerInstance destinationMediator
                    }
                    host_agentProxyBNode = deploymentNode "Workload proxy B" "Own overlay node and launch X.509-SVID; carries workload traffic." "tsnet node; user undecided" {
                        host_destination-workload-proxy = containerInstance destination-workload-proxy
                    }
                }
            }
        }
        deploymentEnvironment "Topology launch" {
            launch_agentHostTrustBoundaryNode = deploymentNode "Workload host trust boundary" "Worker stays outside this data path; service users undecided." "Workload host" {
                launch_agentLaunchBoundaryNode = deploymentNode "Agent launch boundary" "Launch-bound local carriers; private keys stay outside the guest." "Launch isolation" {
                    launch_guestTrustBoundaryNode = deploymentNode "Guest trust boundary" "Guest has no overlay node or workload key." "Workload provider confinement" {
                        launch_guest = containerInstance guest
                    }
                    launch_mediationTrustBoundaryNode = deploymentNode "Mediation trust boundary" "Enforces connection and message policy; no independent network." "Local confinement; user undecided" {
                        launch_mediator = containerInstance mediator
                    }
                    launch_agentProxyTrustBoundaryNode = deploymentNode "Workload proxy trust boundary" "Only launch component opening off-host connections." "tsnet node; user undecided" {
                        launch_workload-proxy = containerInstance workload-proxy
                    }
                }
            }
            launch_controlPlaneTrustBoundaryNode = deploymentNode "Control plane trust boundary" "Service identities authenticate workload connections." "Cluster overlay" {

                launch_apiServiceNodeNode = deploymentNode "API service node" "Checks live launch capabilities and scopes." "tsnet node" {
                    launch_controlApi = containerInstance controlApi
                }
            }
            launch_internetBoundaryNode = deploymentNode "Internet boundary" "Destination authentication differs from workload identity." "External services" {
                launch_internet = softwareSystemInstance internet
            }
        }
    }

    views {
        properties {
            plantuml.skinparams "nodesep=80,ranksep=160"
        }
        systemContext afakt "afakt-context" "afakt and the people and systems it interacts with." {
            include *
            exclude "element.tag==Topology"
            autoLayout lr
        }
        container afakt "afakt-containers" "The afakt process and its external dependencies." {
            include *
            exclude "element.tag==Topology"
            autoLayout lr
        }
        component runtime "afakt-components" "The components inside the combined afakt process." {
            include *
            exclude "element.tag==Topology"
            autoLayout lr
        }

        deployment * "Topology context" "network-context" "Operator devices, attached tailnets, cluster and internet trust boundaries." {
            include *
            autoLayout lr
        }
        deployment * "Topology cluster" "network-cluster" "Service nodes and independent launch nodes on the private overlay." {
            include *
            exclude "relationship.tag==TopologyConnection"
            autoLayout lr
        }
        deployment * "Topology host" "network-host" "Host management, operator access and independent launch trust boundaries." {
            include *
            exclude "relationship.tag==LifecycleDetail"
            exclude "relationship.tag==ReturnDetail"
            exclude "relationship.tag==Undecided"
            autoLayout tb
        }
        deployment * "Topology launch" "network-launch" "One launch crosses guest, mediation, network and service boundaries." {
            include *
            exclude "relationship.tag==ReturnDetail"
            autoLayout lr
        }
        dynamic afakt "network-flow-internet" "Guest egress through its proxy under mediator decisions." {
            guest -> workload-proxy "Traffic on private workload link; no guest workload key"
            workload-proxy -> mediator "Request destination decision"
            mediator -> workload-proxy "Permit destination; instruct proxy TLS termination when required"
            workload-proxy -> internet "Permitted request; destination credentials when authorized"
            autoLayout lr
        }
        dynamic afakt "network-flow-control" "Guest request to the control plane API service." {
            guest -> workload-proxy "Control request bound to current launch"
            workload-proxy -> mediator "Request policy decision and identity binding"
            mediator -> workload-proxy "Policy permit and stamped launch identity"
            workload-proxy -> controlApi "Launch X.509-SVID; service checks capabilities and scopes"
            autoLayout lr
        }
        dynamic afakt "network-flow-preview" "Operator browser reaches a preview through the host node." {
            browser -> hostResolver "Resolve internal preview name"
            browser -> loopbackProxy "Host access request; mechanism undecided"
            loopbackProxy -> hostNetwork "Request under host-level authorization"
            hostNetwork -> previewServer "Host workload identity; served app may require its own sign-in"
            previewServer -> workload-proxy "Direct overlay connection; mutual SPIFFE authentication"
            workload-proxy -> mediator "Request decision for caller and publication"
            mediator -> workload-proxy "Return decision; publication binding remains open"
            workload-proxy -> guest "Permitted ingress; listener binding open (OQ-PORT-14)"
            autoLayout tb
        }
        dynamic afakt "network-flow-phone" "Phone reaches the dashboard through an attached user tailnet." {
            phone -> userTailnet "Device identity in user-owned network"
            userTailnet -> edge "Private application request; afakt sign-in"
            edge -> controlApi "Authenticated user context to web dashboard; no packet routing"
            autoLayout lr
        }
        dynamic afakt "network-flow-static" "A browser requests a static site through the external ingress gateway." {
            browser -> edge "Public request for a static site hostname"
            edge -> staticOrigin "Hostname mapped to the active deploy; gateway compresses and caches"
            autoLayout lr
        }
        dynamic afakt "network-flow-grant" "Direct overlay connection checked by both mediators." {
            guest -> workload-proxy "Source launch requests published destination"
            workload-proxy -> mediator "Request egress decision"
            mediator -> workload-proxy "Source policy permits destination name and port"
            workload-proxy -> destination-workload-proxy "Direct overlay connection with mutual SPIFFE authentication"
            destination-workload-proxy -> destinationMediator "Authenticated source launch and destination grant"
            destinationMediator -> destination-workload-proxy "Return ingress decision"
            destination-workload-proxy -> destinationGuest "Permitted ingress; listener binding open (OQ-PORT-14)"
            autoLayout tb
        }
        container afakt "network-service-boundaries" "Services supporting the deployment topology; connections are control or transport dependencies." {
            include hostWorker helper controlApi mcpGateway coordination overlayRelay
            exclude "relationship.tag==OverviewHidden"
            autoLayout tb
        }
        styles {
            element "Topology" {
                background #e8f1fa
                color #122b44
                stroke #315b7d
            }
            element "Deployment Node" {
                background #f5f7fa
                color #182c3e
                stroke #536b7e
            }
            element "Group" {
                color #182c3e
                stroke #536b7e
            }
            relationship "Undecided" {
                color #8a5a00
                style dashed
            }
        }
    }

    configuration {
        scope softwaresystem
    }
}
