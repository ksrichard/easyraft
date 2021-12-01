package fsm

import (
	"github.com/hashicorp/raft"
	"sync"
)

type BaseFSMSnapshot struct {
	sync.Mutex
	fsm *RoutingFSM
}

func NewBaseFSMSnapshot(fsm *RoutingFSM) raft.FSMSnapshot {
	return &BaseFSMSnapshot{fsm: fsm}
}

func (i *BaseFSMSnapshot) Persist(sink raft.SnapshotSink) error {
	i.Lock()
	snapshotData, err := i.fsm.ser.Serialize(i.fsm.services)
	if err != nil {
		return err
	}
	_, err = sink.Write(snapshotData)
	return err
}

func (i *BaseFSMSnapshot) Release() {
	i.Unlock()
}
