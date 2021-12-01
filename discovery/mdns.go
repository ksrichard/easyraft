package discovery

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"log"
	"math/rand"
	"time"
)

const (
	mdnsServiceName = "_easyraft._tcp"
)

type MDNSDiscovery struct {
	delayTime     time.Duration
	nodeID        string
	nodePort      int
	mdnsServer    *zeroconf.Server
	discoveryChan chan string
	stopChan      chan bool
}

func NewMDNSDiscovery() DiscoveryMethod {
	rand.Seed(time.Now().UnixNano())
	delayTime := time.Duration(rand.Intn(30)+5) * time.Second
	return &MDNSDiscovery{
		delayTime:     delayTime,
		discoveryChan: make(chan string),
		stopChan:      make(chan bool),
	}
}

func (d *MDNSDiscovery) Start(nodeID string, nodePort int) (chan string, error) {
	d.nodeID, d.nodePort = nodeID, nodePort
	if d.discoveryChan == nil {
		d.discoveryChan = make(chan string)
	}
	go d.discovery()
	return d.discoveryChan, nil
}

func (d *MDNSDiscovery) discovery() {
	// expose mdns server
	mdnsServer, err := d.exposeMDNS()
	if err != nil {
		log.Fatal(err)
	}
	d.mdnsServer = mdnsServer

	// fetch mDNS enabled raft nodes
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize mDNS resolver:", err.Error())
	}
	entries := make(chan *zeroconf.ServiceEntry)
	go func() {
		for {
			select {
			case <-d.stopChan:
				break
			case entry := <-entries:
				d.discoveryChan <- fmt.Sprintf("%s:%d", entry.AddrIPv4[0], entry.Port)
			}
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	for {
		select {
		case <-d.stopChan:
			cancel()
			break
		default:
			err = resolver.Browse(ctx, mdnsServiceName, "local.", entries)
			if err != nil {
				log.Printf("Error during mDNS lookup: %v\n", err)
			}
			time.Sleep(d.delayTime)
		}
	}
}

func (d *MDNSDiscovery) exposeMDNS() (*zeroconf.Server, error) {
	return zeroconf.Register(d.nodeID, mdnsServiceName, "local.", d.nodePort, []string{"txtv=0", "lo=1", "la=2"}, nil)
}

func (d *MDNSDiscovery) SupportsNodeAutoRemoval() bool {
	return true
}

func (d *MDNSDiscovery) Stop() {
	d.stopChan <- true
	d.mdnsServer.Shutdown()
	close(d.discoveryChan)
}
