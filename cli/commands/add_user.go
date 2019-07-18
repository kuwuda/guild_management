package commands

import (
	"fmt"
	"strconv"

	"git.sr.ht/~sircmpwn/getopt"
	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

// AddUser adds a user to the db
func AddUser(conn *grpc.ClientConn, in []string) error {
	opts, optind, err := getopt.Getopts(in, "h")
	if err != nil {
		return err
	}

	keys, err := client.GetKeys(conn)
	if err != nil {
		return err
	}

	var parsed []string
	// get specified arguments
	if len(in) > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'h':
				s, err := getHelp("adduser", keys)
				if err != nil {
					return err
				}
				fmt.Println(s)
				return nil
			}
		}

		// get remaining arguments
		for _, arg := range in[optind:] {
			parsed = append(parsed, arg)
		}
	} else {
		s, err := getHelp("adduser", keys)
		if err != nil {
			return err
		}
		fmt.Println(s)
		return nil
	}
	activityItem := &pb.ActivityItem{Name: parsed[0]}
	activityItem.Activities = make(map[string]uint32)

	for len(parsed)-1 < len(keys) {
		parsed = append(parsed, "0")
	}

	for i, v := range keys {
		parsedAsInt, err := strconv.Atoi(parsed[i+1])
		if err != nil {
			return err
		}
		activityItem.Activities[v] = uint32(parsedAsInt)
	}

	var activities []*pb.ActivityItem
	activities = append(activities, activityItem)

	resp, err := client.WriteMembers(conn, activities)
	fmt.Println(resp)
	return err
}
