package commands

import (
	"fmt"

	"git.sr.ht/~sircmpwn/getopt"
	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

// DelUser deletes a user from the DB
func DelUser(conn *grpc.ClientConn, in []string) error {
	opts, _, err := getopt.Getopts(in, "h")
	if err != nil {
		return err
	}

	// get specified arguments
	if len(in) > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'h':
				s, err := getHelp("deluser", nil)
				if err != nil {
					return err
				}
				fmt.Println(s)
				return nil
			}
		}
	} else {
		s, err := getHelp("adduser", nil)
		if err != nil {
			return err
		}
		fmt.Println(s)
		return nil
	}
	if len(in) < 2 {
		s, err := getHelp("adduser", nil)
		if err != nil {
			return err
		}
		fmt.Println(s)
		return nil
	}
	delItem := &pb.DeleteRequest{Name: in[1]}

	var delReq []*pb.DeleteRequest
	delReq = append(delReq, delItem)

	resp, err := client.DeleteMembers(conn, delReq)
	fmt.Println(resp)
	return err
}
