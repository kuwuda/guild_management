package commands

import (
	"git.sr.ht/~sircmpwn/getopt"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

// AddCol adds a column to the db
func AddCol(conn *grpc.ClientConn, in []string) (string, error) {
	opts, _, err := getopt.Getopts(in, "h")
	if err != nil {
		return "", err
	}

	// get specified arguments
	if len(in) > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'h':
				s, err := getHelp("addcol", nil)
				if err != nil {
					return "", err
				}
				return s, nil
			}
		}
	} else {
		s, err := getHelp("addcol", nil)
		if err != nil {
			return "", err
		}
		return s, nil
	}
	key := in[1]
	var keys []string
	keys = append(keys, key)

	_, err = client.AddColumns(conn, keys)
	return "", err
}
