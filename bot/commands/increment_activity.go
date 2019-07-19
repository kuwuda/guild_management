package commands

import (
	"strconv"

	"git.sr.ht/~sircmpwn/getopt"
	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

// IncAct increments a given activity in the db
func IncAct(conn *grpc.ClientConn, in []string) (string, error) {
	// incActivity kraken kusu apples cat
	opts, optind, err := getopt.Getopts(in, "hn:")
	if err != nil {
		return "", err
	}

	var incReq pb.IncRequest
	incReq.Amount = 1
	// get specified arguments
	if len(in) > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'h':
				s, err := getHelp("incact", nil)
				if err != nil {
					return "", err
				}
				return s, nil
			case 'n':
				n, err := strconv.Atoi(opt.Value)
				if err != nil {
					return "", err
				}
				incReq.Amount = int32(n)
			}
		}
	} else {
		s, err := getHelp("incact", nil)
		if err != nil {
			return "", err
		}
		return s, nil
	}
	args := in[optind:]
	incReq.Key = args[0]
	incReq.Names = args[1:]

	var activities []*pb.IncRequest
	activities = append(activities, &incReq)

	_, err = client.IncrementActivities(conn, activities)
	return "", err
}
