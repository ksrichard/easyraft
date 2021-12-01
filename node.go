package easyraft

import (
	"errors"
	"fmt"
	"github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/ksrichard/easyraft/discovery"
	"github.com/ksrichard/easyraft/fsm"
	"github.com/ksrichard/easyraft/grpc"
	"github.com/ksrichard/easyraft/serializer"
	"github.com/ksrichard/easyraft/util"
	"github.com/zemirco/uid"
	ggrpc "google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type Node struct {
	ID               string
	RaftPort         int
	DiscoveryPort    int
	address          string
	dataDir          string
	Raft             *raft.Raft
	GrpcServer       *ggrpc.Server
	DiscoveryMethod  discovery.DiscoveryMethod
	TransportManager *transport.Manager
	Serializer       serializer.Serializer
	mList            *memberlist.Memberlist
	discoveryConfig  *memberlist.Config
	stopped          *uint32
	logger           *log.Logger
	stoppedCh        chan interface{}
	snapshotEnabled  bool
}

func NewNode(raftPort, discoveryPort int, dataDir string, services []fsm.FSMService, serializer serializer.Serializer, discoveryMethod discovery.DiscoveryMethod, snapshotEnabled bool) (*Node, error) {
	// default raft config
	addr := fmt.Sprintf("%s:%d", "0.0.0.0", raftPort)
	nodeId := uid.New(50)
	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(nodeId)
	raftLogCacheSize := 512
	raftConf.LogLevel = "Info"

	// stable/log/snapshot store config
	if !util.IsDir(dataDir) {
		err := util.RemoveCreateDir(dataDir)
		if err != nil {
			return nil, err
		}
	}
	stableStoreFile := filepath.Join(dataDir, "store.boltdb")
	if util.FileExists(stableStoreFile) {
		err := os.Remove(stableStoreFile)
		if err != nil {
			return nil, err
		}
	}
	stableStore, err := raftboltdb.NewBoltStore(stableStoreFile)
	if err != nil {
		return nil, err
	}

	logStore, err := raft.NewLogCache(raftLogCacheSize, stableStore)
	if err != nil {
		return nil, err
	}

	var snapshotStore raft.SnapshotStore
	if !snapshotEnabled {
		snapshotStore = raft.NewDiscardSnapshotStore()
	} else {
		// TODO: implement: snapshotStore = NewLogsOnlySnapshotStore(serializer)
		return nil, errors.New("snapshots are not supported at the moment")
	}

	// grpc transport
	grpcTransport := transport.New(raft.ServerAddress(addr), []ggrpc.DialOption{ggrpc.WithInsecure()})

	// init FSM
	sm := fsm.NewRoutingFSM(services)
	sm.Init(serializer)

	// memberlist config
	mlConfig := memberlist.DefaultWANConfig()
	mlConfig.BindPort = discoveryPort
	mlConfig.Name = fmt.Sprintf("%s:%d", nodeId, raftPort)

	// raft server
	raftServer, err := raft.NewRaft(raftConf, sm, logStore, stableStore, snapshotStore, grpcTransport.Transport())
	if err != nil {
		return nil, err
	}

	// logging
	logger := log.Default()
	logger.SetPrefix("[EasyRaft] ")

	// initial stopped flag
	var stopped uint32

	return &Node{
		ID:               nodeId,
		RaftPort:         raftPort,
		address:          addr,
		dataDir:          dataDir,
		Raft:             raftServer,
		TransportManager: grpcTransport,
		Serializer:       serializer,
		DiscoveryPort:    discoveryPort,
		DiscoveryMethod:  discoveryMethod,
		discoveryConfig:  mlConfig,
		logger:           logger,
		stopped:          &stopped,
		snapshotEnabled:  snapshotEnabled,
	}, nil
}

func (n *Node) Start() (chan interface{}, error) {
	n.logger.Println("Starting Node...")
	// set stopped as false
	if atomic.LoadUint32(n.stopped) == 1 {
		atomic.StoreUint32(n.stopped, 0)
	}

	// raft server
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(n.ID),
				Address: n.TransportManager.Transport().LocalAddr(),
			},
		},
	}
	f := n.Raft.BootstrapCluster(configuration)
	err := f.Error()
	if err != nil {
		return nil, err
	}

	// memberlist discovery
	n.discoveryConfig.Events = n
	list, err := memberlist.Create(n.discoveryConfig)
	if err != nil {
		return nil, err
	}
	n.mList = list

	// grpc server
	grpcListen, err := net.Listen("tcp", n.address)
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := ggrpc.NewServer()
	n.GrpcServer = grpcServer

	// register management services
	n.TransportManager.Register(grpcServer)

	// register client services
	clientGrpcServer := NewClientGrpcService(n)
	grpc.RegisterRaftServer(grpcServer, clientGrpcServer)

	// discovery method
	discoveryChan, err := n.DiscoveryMethod.Start(n.ID, n.RaftPort)
	if err != nil {
		return nil, err
	}
	go n.handleDiscoveredNodes(discoveryChan)

	// serve grpc
	go func() {
		if err := grpcServer.Serve(grpcListen); err != nil {
			n.logger.Fatal(err)
		}
	}()

	// handle interruption
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGABRT, syscall.SIGKILL)
	go func() {
		_ = <-sigs
		n.Stop()
	}()

	n.logger.Printf("Node started on port %d and discovery port %d\n", n.RaftPort, n.DiscoveryPort)
	n.stoppedCh = make(chan interface{})

	return n.stoppedCh, nil
}

func (n *Node) Stop() {
	if atomic.LoadUint32(n.stopped) == 0 {
		atomic.StoreUint32(n.stopped, 1)
		if n.snapshotEnabled {
			n.logger.Println("Creating snapshot...")
			err := n.Raft.Snapshot().Error()
			if err != nil {
				n.logger.Println("Failed to create snapshot!")
			}
		}
		n.logger.Println("Stopping Node...")
		n.DiscoveryMethod.Stop()
		err := n.mList.Leave(10 * time.Second)
		if err != nil {
			n.logger.Printf("Failed to leave from discovery: %q\n", err.Error())
		}
		err = n.mList.Shutdown()
		if err != nil {
			n.logger.Printf("Failed to shutdown discovery: %q\n", err.Error())
		}
		n.logger.Println("Discovery stopped")
		err = n.Raft.Shutdown().Error()
		if err != nil {
			n.logger.Printf("Failed to shutdown Raft: %q\n", err.Error())
		}
		n.logger.Println("Raft stopped")
		n.GrpcServer.GracefulStop()
		n.logger.Println("Raft Server stopped")
		n.logger.Println("Node Stopped!")
		n.stoppedCh <- true
	}
}

func (n *Node) handleDiscoveredNodes(discoveryChan chan string) {
	for peer := range discoveryChan {
		detailsResp, err := GetPeerDetails(peer)
		if err == nil {
			serverId := detailsResp.ServerId
			needToAddNode := true
			for _, server := range n.Raft.GetConfiguration().Configuration().Servers {
				if server.ID == raft.ServerID(serverId) || string(server.Address) == peer {
					needToAddNode = false
					break
				}
			}
			if needToAddNode {
				peerHost := strings.Split(peer, ":")[0]
				peerDiscoveryAddr := fmt.Sprintf("%s:%d", peerHost, detailsResp.DiscoveryPort)
				_, err = n.mList.Join([]string{peerDiscoveryAddr})
				if err != nil {
					log.Printf("failed to join to cluster using discovery address: %s\n", peerDiscoveryAddr)
				}
			}
		}
	}
}

func (n *Node) NotifyJoin(node *memberlist.Node) {
	nameParts := strings.Split(node.Name, ":")
	nodeId, nodePort := nameParts[0], nameParts[1]
	nodeAddr := fmt.Sprintf("%s:%s", node.Addr, nodePort)
	if err := n.Raft.VerifyLeader().Error(); err == nil {
		result := n.Raft.AddVoter(raft.ServerID(nodeId), raft.ServerAddress(nodeAddr), 0, 0)
		if result.Error() != nil {
			log.Println(result.Error().Error())
		}
	}
}

func (n *Node) NotifyLeave(node *memberlist.Node) {
	if n.DiscoveryMethod.SupportsNodeAutoRemoval() {
		nodeId := strings.Split(node.Name, ":")[0]
		if err := n.Raft.VerifyLeader().Error(); err == nil {
			result := n.Raft.RemoveServer(raft.ServerID(nodeId), 0, 0)
			if result.Error() != nil {
				log.Println(result.Error().Error())
			}
		}
	}
}

func (n *Node) NotifyUpdate(_ *memberlist.Node) {
}

func (n *Node) RaftApply(request interface{}, timeout time.Duration) (interface{}, error) {
	payload, err := n.Serializer.Serialize(request)
	if err != nil {
		return nil, err
	}

	if err := n.Raft.VerifyLeader().Error(); err == nil {
		result := n.Raft.Apply(payload, timeout)
		if result.Error() != nil {
			return nil, result.Error()
		}
		switch result.Response().(type) {
		case error:
			return nil, result.Response().(error)
		default:
			return result.Response(), nil
		}
	}

	response, err := ApplyOnLeader(n, payload)
	if err != nil {
		return nil, err
	}
	return response, nil
}
