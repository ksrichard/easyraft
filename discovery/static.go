package discovery

type StaticDiscovery struct {
	Peers         []string
	discoveryChan chan string
	stopChan      chan bool
}

func NewStaticDiscovery(peers []string) DiscoveryMethod {
	return &StaticDiscovery{
		Peers:         peers,
		discoveryChan: make(chan string),
		stopChan:      make(chan bool),
	}
}

func (d *StaticDiscovery) SupportsNodeAutoRemoval() bool {
	return false
}

func (d *StaticDiscovery) Start(_ string, _ int) (chan string, error) {
	go func() {
		for _, peer := range d.Peers {
			d.discoveryChan <- peer
		}
	}()
	return d.discoveryChan, nil
}

func (d *StaticDiscovery) Stop() {
	close(d.discoveryChan)
	d.stopChan <- true
}
