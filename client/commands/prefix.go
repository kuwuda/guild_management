package commands

import (
	"git.sr.ht/~sircmpwn/getopt"
	"google.golang.org/grpc"
)

// Prefix sets the prefix for the bot
func Prefix(conn *grpc.ClientConn, in []string) (string, error) {
	opts, _, err := getopt.Getopts(in, "h")
	if err != nil {
		return "", err
	}

	// get specified arguments
	if len(in) > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'h':
				s, err := getHelp("prefix", nil)
				if err != nil {
					return "", err
				}
				return s, nil
			}
		}
	} else {
		s, err := getHelp("prefix", nil)
		if err != nil {
			return "", err
		}
		return s, nil
	}

	// put function that edits config here
	return "", err
}
