package healthcheck

// HealthStatus stores health statuses of monitored nodes.
type HealthStatus struct {
	Groups []Group
}

// Group gathers monitored nodes.
type Group struct {
	Name   string
	Anchor string
	Nodes  []Node
}

// Node reports health status of a monitored node.
type Node struct {
	Name         string
	Type         string
	Endpoint     string
	HTTPEndpoint bool
	Status       string
	Message      string
}
