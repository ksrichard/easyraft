package discovery

type DiscoveryMethod interface {
	Start(nodeID string, nodePort int) (chan string, error)
	SupportsNodeAutoRemoval() bool
	Stop()
}
