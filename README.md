<p align="center">
<img src="https://github.com/ksrichard/easyraft/raw/main/logo.png" width="50%">
</p>

[![Go Report Card](https://goreportcard.com/badge/github.com/ksrichard/easyraft)](https://goreportcard.com/report/github.com/ksrichard/easyraft)
[![Go Reference](https://pkg.go.dev/badge/github.com/ksrichard/easyraft.svg)](https://pkg.go.dev/github.com/ksrichard/easyraft)
[![GitHub go.mod Go version of a Go module](https://img.shields.io/github/go-mod/go-version/ksrichard/easyraft.svg)](https://github.com/ksrichard/easyraft)
[![GitHub release](https://img.shields.io/github/release/ksrichard/easyraft.svg)](https://github.com/ksrichard/easyraft/releases/latest/)

An easy to use customizable library to make your Go application
Distributed, Highly available, Fault Tolerant etc...
using Hashicorp's [Raft](https://github.com/hashicorp/raft) library which implements the
[Raft Consensus Algorithm](https://raft.github.io/).

Features
---
- **Configure and start** a fully functional Raft node by writing ~10 lines of code
- **Automatic Node discovery** (nodes are discovering each other using Discovery method)
  1. **Built-in discovery methods**:
     1. **Static Discovery** (having a fixed list of nodes addresses)
     2. **mDNS Discovery** for local network node discovery
     3. **Kubernetes discovery**
- **Automatic forward to leader** - you can contact any node to perform operations, everything will be forwarded to the actual leader node
- **Node monitoring/removal** - the nodes are monitoring each other and if there are some failures then the 
offline nodes get removed automatically from cluster
- **Simplified state machine** - there is an already implemented generic state machine 
which handles the basic operations and routes requests to State Machine Services (see **Examples**)
- **All layers are customizable** - you can select or implement your own **State Machine Service, Message Serializer** and **Discovery Method**
- **gRPC transport layer** - the internal communications are done through gRPC based communication, if needed you can add your own services

**Note:** snapshots are not supported at the moment, will be handled at later point
**Note:** at the moment the communication between nodes are insecure, I recommend to not expose that port

Get Started
---
You can create a simple EasyRaft Node with local mDNS discovery, an in-memory Map service
and MsgPack as serializer(this is the only one built-in at the moment)
```go
node, err := easyraft.NewNode(
		raftPort,
		discoveryPort,
		dataDir,
		[]fsm.FSMService{fsm.NewInMemoryMapService()},
		serializer.NewMsgPackSerializer(),
		discovery.NewMDNSDiscovery(),
		false,
	)

	if err != nil {
		panic(err)
	}
	stoppedCh, err := node.Start()
	if err != nil {
		panic(err)
	}
	defer node.Stop()
```

TODO
---
- [ ] Add more examples
- [ ] Test coverage
- [ ] Secure communication between nodes (SSL/TLS)
- [ ] Backup/Restore backup handling
- [ ] Allow configuration option to pass any custom raft.FSM



