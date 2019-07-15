// Package client consists of tools which are helpful for making any client
package client

import (
	"context"
	"io"
	"time"

	pb "github.com/kuwuda/guild_management/api"
	"google.golang.org/grpc"
)

// GetActivities sends a ActivityRequest to the server and returns a slice of ActivityItems
// Primarily used to query the database for activities
func GetActivities(conn *grpc.ClientConn, opts *pb.ActivityRequest) (ret []*pb.ActivityItem, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)

	stream, err := client.GetActivities(ctx, opts)
	if err != nil {
		return nil, err
	}
	for {
		member, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ret = append(ret, member)
	}
	return ret, nil
}

// GetKeys gets the keys for activities from the server
func GetKeys(conn *grpc.ClientConn) (ret []string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)
	keys, err := client.GetKeys(ctx, &pb.KeyRequest{})
	if err != nil {
		return
	}
	ret = keys.Keys
	return
}

// WriteMembers sends a given array of members to the server to write them in the DB
func WriteMembers(conn *grpc.ClientConn, members []*pb.ActivityItem) (ret *pb.ActivityResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)

	stream, err := client.WriteMembers(ctx)
	if err != nil {
		return
	}

	for _, v := range members {
		if err = stream.Send(v); err != nil {
			return
		}
	}

	ret, err = stream.CloseAndRecv()
	if err != nil {
		return
	}

	return
}

// UpdateMembers sends a given array of members to the server in order to update the DB
// It's notable that using ActivityItem for the datastructure which gets sent to the server here
// prevents name-updates from happening.
// A potential solution could be to include an objectid in the query and change it via that
// or include an optional "namechange" field in the struct.
// Not sure what the best solution is.
func UpdateMembers(conn *grpc.ClientConn, members []*pb.ActivityItem) (ret *pb.ActivityResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)

	stream, err := client.UpdateMembers(ctx)
	if err != nil {
		return
	}

	for _, v := range members {
		if err = stream.Send(v); err != nil {
			return
		}
	}
	ret, err = stream.CloseAndRecv()
	if err != nil {
		return
	}
	return
}

// DeleteMembers asks the server to delete a given slice of members
func DeleteMembers(conn *grpc.ClientConn, members []*pb.DeleteRequest) (ret *pb.ActivityResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)

	stream, err := client.DeleteMembers(ctx)
	if err != nil {
		return
	}

	for _, v := range members {
		if err = stream.Send(v); err != nil {
			return
		}
	}
	ret, err = stream.CloseAndRecv()
	if err != nil {
		return
	}
	return
}

// AddColumns asks the server to add the given keys to all documents in the collection
func AddColumns(conn *grpc.ClientConn, keys []string) (ret *pb.ActivityResponse, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client := pb.NewActivityServiceClient(conn)

	stream, err := client.AddColumns(ctx)
	if err != nil {
		return
	}

	for _, v := range keys {

		if err = stream.Send(&pb.ColRequest{Key: v}); err != nil {
			return
		}
	}
	ret, err = stream.CloseAndRecv()
	if err != nil {
		return
	}
	return
}
