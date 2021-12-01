package fsm

import (
	"github.com/hashicorp/raft"
	"github.com/ksrichard/easyraft/serializer"
)

// FSM is the interface for using Finite State Machine in Raft Cluster
type FSM interface {
	raft.FSM

	// Init is used to pass the original serializer from EasyRaft Node to be able to deserialize messages
	// coming from other nodes
	Init(ser serializer.Serializer)
}

// FSMService interface makes it easier to build State Machines
type FSMService interface {
	// Name returns the unique ID/Name which will identify the FSM Service when it comes to routing incoming messages
	Name() string

	// NewLog is called when a new raft log message is committed in the cluster and matched with any of the GetReqDataTypes returned types
	// in this method we can handle what should happen when we got a new raft log regarding our FSM service
	NewLog(requestType interface{}, request map[string]interface{}) interface{}

	// GetReqDataTypes returns all the request structs which are used by this FSMService
	GetReqDataTypes() []interface{}

	// ApplySnapshot is used to decode and apply a snapshot to the FSMService
	ApplySnapshot(input interface{}) error
}
