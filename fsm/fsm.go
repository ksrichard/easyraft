package fsm

import (
	"github.com/hashicorp/raft"
	"github.com/ksrichard/easyraft/serializer"
)

type FSM interface {
	raft.FSM
	Init(ser serializer.Serializer)
}

type FSMService interface {
	Name() string
	NewLog(requestType interface{}, request map[string]interface{}) interface{}
	GetReqDataTypes() []interface{}
	ApplySnapshot(input interface{}) error
}
