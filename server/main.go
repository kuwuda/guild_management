package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sort"
	"strings"
	"time"

	pb "github.com/kuwuda/guild_management/api"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	mongoClient *mongo.Client
}

func checkNameExists(col *mongo.Collection, name string) (bool, error) {
	// collation option here makes search case-insensitive
	// not sure about any performance drawbacks or anything
	collation := options.Collation{Locale: "en", Strength: 2}
	options := options.Find().SetCollation(&collation)
	cur, err := col.Find(context.Background(), bson.M{"name": name}, options)
	if err != nil {
		return false, err
	}
	if cur.Next(context.Background()) {
		return true, nil
	}
	return false, nil
}

// checks if a slice of strings is equal regardless of order
// requires copying both slices and sorts them in-place
// significant overhead because of this;
// there are more optimal ways but this is more readable. might change later.
func checkEqual(a []string, b []string) bool {
	// if they aren't the same length, exit early
	if len(a) != len(b) {
		return false
	}

	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// parseKeys gets keys from an ActivityMember struct
func parseKeys(data pb.ActivityMember) (keys []string) {
	for i := range data.Activities {
		keys = append(keys, i)
	}
	return
}

// getKeys gets the keys for a given collection
func getKeys(col *mongo.Collection) ([]string, error) {
	var data pb.ActivityMember

	err := col.FindOne(context.Background(), bson.D{{}}).Decode(&data)
	if err != nil {
		return nil, err
	}

	return parseKeys(data), nil
}

func (s *server) GetKeys(ctx context.Context, req *pb.KeyRequest) (ret *pb.ActivityKeys, err error) {
	collection := s.mongoClient.Database("test").Collection("activity")
	keys, err := getKeys(collection)
	if err != nil {
		return
	}

	var keyHolder pb.ActivityKeys
	keyHolder.Keys = make([]string, len(keys))
	copy(keyHolder.Keys, keys)

	ret = &keyHolder
	return
}

func (s *server) DeleteMembers(stream pb.ActivityService_DeleteMembersServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		name, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		filter := make(bson.M)
		filter["name"] = name.Name
		collation := options.Collation{Locale: "en", Strength: 2}
		options := options.Delete().SetCollation(&collation)

		// Not sure what I should be doing with the DB's responses
		// Is discarding them like this fine?
		_, err = collection.DeleteOne(context.Background(), filter, options)
		if err != nil {
			return err
		}

		entries++
	}
}

func (s *server) UpdateMembers(stream pb.ActivityService_UpdateMembersServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		member, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		var actMember pb.ActivityMember
		actMember.Name = member.Name
		actMember.Activities = member.Activities

		// Check if "schema" of database matches that of data input
		colKeys, err := getKeys(collection)
		if err != nil {
			return err
		}

		if !checkEqual(parseKeys(actMember), colKeys) {
			return errors.New("does not match current schema")
		}

		filter := make(bson.M)
		filter["name"] = actMember.Name

		collation := options.Collation{Locale: "en", Strength: 2}
		options := options.Update().SetCollation(&collation)

		insert := bson.D{primitive.E{Key: "$set", Value: actMember}}

		// Not sure what I should be doing with the DB's responses
		// Is discarding them like this fine?
		_, err = collection.UpdateOne(context.Background(), filter, insert, options)
		if err != nil {
			return err
		}

		entries++
	}
}

func (s *server) WriteMembers(stream pb.ActivityService_WriteMembersServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		member, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		exists, err := checkNameExists(collection, member.Name)
		if err != nil {
			return err
		}

		// Not totally sure if breaking like this is optimal when an entry already exists
		// Might change later
		if exists {
			return errors.New("entry already exists")
		}

		var actMember pb.ActivityMember
		actMember.Name = member.Name
		actMember.Activities = member.Activities

		// Check if "schema" of database matches that of data input
		colKeys, err := getKeys(collection)
		if err != nil {
			return err
		}

		if !checkEqual(parseKeys(actMember), colKeys) {
			return errors.New("does not match current schema")
		}

		_, err = collection.InsertOne(context.Background(), actMember)
		if err != nil {
			return err
		}
		entries++
	}
}

// It's interesting that I'm essentially writing functions to enforce a schema on a schema-less database
// I'm not really sure if there's a better way to do this
// Perhaps mongodb wasn't really the right choice for this
// Or maybe there's a better way to do this.
func (s *server) DeleteColumns(stream pb.ActivityService_DeleteColumnsServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		// This is disgusting. Not sure if this is like vulnerable to something similar to SQL injection either.
		// Will have to test
		res, err := collection.UpdateMany(context.Background(), bson.D{}, bson.D{
			primitive.E{Key: "$unset", Value: bson.M{"activities." + entry.Key: 0}}})
		if err != nil {
			return err
		}

		entries += uint32(res.ModifiedCount)
	}
}

func (s *server) AddColumns(stream pb.ActivityService_AddColumnsServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		keys, err := getKeys(collection)
		if err != nil {
			return err
		}

		// Not sure if exiting the function like this is the proper way to handle this error
		for _, v := range keys {
			if strings.EqualFold(v, entry.Key) {
				return errors.New("entry already exists")
			}
		}

		// This is disgusting. Not sure if this is like vulnerable to something similar to SQL injection either.
		// Will have to test
		res, err := collection.UpdateMany(context.Background(), bson.D{}, bson.D{
			primitive.E{Key: "$set", Value: bson.M{"activities." + entry.Key: 0}}})
		if err != nil {
			return err
		}

		entries += uint32(res.ModifiedCount)
	}
}

func (s *server) IncrementActivities(stream pb.ActivityService_IncrementActivitiesServer) error {
	var entries uint32
	startTime := time.Now()

	collection := s.mongoClient.Database("test").Collection("activity")

	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			endTime := time.Now()
			return stream.SendAndClose(&pb.ActivityResponse{
				Entries:     entries,
				ElapsedTime: int32(endTime.Sub(startTime).Seconds()),
			})
		}
		if err != nil {
			return err
		}

		keys, err := getKeys(collection)
		if err != nil {
			return err
		}

		var exists bool
		for _, v := range keys {
			if entry.Key == v {
				exists = true
			}
		}

		if !exists {
			return errors.New("key " + entry.Key + " does not exist")
		}

		var names bson.A
		for _, v := range entry.Names {
			name := bson.D{primitive.E{Key: "name", Value: v}}
			names = append(names, name)
		}

		collation := options.Collation{Locale: "en", Strength: 2}
		options := options.Update().SetCollation(&collation)

		filter := bson.D{primitive.E{Key: "$or", Value: names}}
		update := bson.D{primitive.E{Key: "$inc", Value: bson.M{"activities." + entry.Key: entry.Amount}}}
		res, err := collection.UpdateMany(context.Background(), filter, update, options)
		if err != nil {
			return err
		}

		entries += uint32(res.ModifiedCount)
	}
}

func (s *server) GetActivities(args *pb.ActivityRequest, stream pb.ActivityService_GetActivitiesServer) error {
	collection := s.mongoClient.Database("test").Collection("activity")

	filter := make(bson.M)
	collation := options.Collation{Locale: "en", Strength: 2}
	options := options.Find().SetCollation(&collation)
	if args.User != "" {
		filter["name"] = args.User
	}
	if args.Amount != 0 {
		options.SetLimit(int64(args.Amount))
	}

	// collation option here makes search case-insensitive
	// not sure about any performance drawbacks or anything
	cur, err := collection.Find(context.Background(), filter, options)
	if err != nil {
		return err
	}
	defer cur.Close(context.Background())

	var activityHolder []*pb.ActivityItem
	for cur.Next(context.Background()) {
		var ret pb.ActivityItem

		err := cur.Decode(&ret)
		if err != nil {
			return err
		}

		activityHolder = append(activityHolder, &ret)
	}

	for _, v := range activityHolder {
		if err := stream.Send(v); err != nil {
			return err
		}
	}

	return nil
}

func main() {
	// start a mongodb connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
	}

	// create a listener on TCP port 7777
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create a gRPC server object
	s := grpc.NewServer()

	pb.RegisterActivityServiceServer(s, &server{mongoClient: mongoClient})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	err = mongoClient.Disconnect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connection to MongoDB closed.")
}
