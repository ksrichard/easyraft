package easyraft

import (
	"context"
	rgrpc "github.com/ksrichard/easyraft/grpc"
)

func NewClientGrpcService(node *Node) *ClientGrpcServices {
	return &ClientGrpcServices{
		Node: node,
	}
}

type ClientGrpcServices struct {
	Node *Node
	rgrpc.UnimplementedRaftServer
}

func (s *ClientGrpcServices) ApplyLog(ctx context.Context, request *rgrpc.ApplyRequest) (*rgrpc.ApplyResponse, error) {
	result := s.Node.Raft.Apply(request.GetRequest(), 0)
	if result.Error() != nil {
		return nil, result.Error()
	}
	respPayload, err := s.Node.Serializer.Serialize(result.Response())
	if err != nil {
		return nil, err
	}
	return &rgrpc.ApplyResponse{Response: respPayload}, nil
}

func (s *ClientGrpcServices) GetDetails(context.Context, *rgrpc.GetDetailsRequest) (*rgrpc.GetDetailsResponse, error) {
	return &rgrpc.GetDetailsResponse{
		ServerId:      s.Node.ID,
		DiscoveryPort: int32(s.Node.DiscoveryPort),
	}, nil
}
