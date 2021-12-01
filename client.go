package easyraft

import (
	"context"
	"github.com/ksrichard/easyraft/grpc"
	ggrpc "google.golang.org/grpc"
)

func ApplyOnLeader(node *Node, payload []byte) (interface{}, error) {
	var opt ggrpc.DialOption = ggrpc.EmptyDialOption{}
	conn, err := ggrpc.Dial(string(node.Raft.Leader()), ggrpc.WithInsecure(), ggrpc.WithBlock(), opt)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := grpc.NewRaftClient(conn)

	response, err := client.ApplyLog(context.Background(), &grpc.ApplyRequest{Request: payload})
	if err != nil {
		return nil, err
	}

	result, err := node.Serializer.Deserialize(response.Response)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func GetPeerDetails(address string) (*grpc.GetDetailsResponse, error) {
	var opt ggrpc.DialOption = ggrpc.EmptyDialOption{}
	conn, err := ggrpc.Dial(address, ggrpc.WithInsecure(), ggrpc.WithBlock(), opt)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := grpc.NewRaftClient(conn)

	response, err := client.GetDetails(context.Background(), &grpc.GetDetailsRequest{})
	if err != nil {
		return nil, err
	}

	return response, nil
}
