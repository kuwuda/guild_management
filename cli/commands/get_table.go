package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"git.sr.ht/~sircmpwn/getopt"
	pb "github.com/kuwuda/guild_management/api"
	"github.com/kuwuda/guild_management/client"
	"google.golang.org/grpc"
)

func digits(n int) (digits int) {
	for n != 0 {
		n /= 10
		digits++
	}
	return
}

// Returns the greater int of the pair
// necessary since Golang's standard library has no function for ints, only floats
func max(x int, y int) int {
	if x < y {
		return y
	}
	return x
}

// GetTable returns elements in the db as a table
func GetTable(conn *grpc.ClientConn, in []string) error {
	opts, optind, err := getopt.Getopts(in, "n:u:h")
	if err != nil {
		return err
	}

	// get specified arguments
	// there has to be a more eloquent way to do this
	var request pb.ActivityRequest
	if optind > 1 {
		for _, opt := range opts {
			switch opt.Option {
			case 'n':
				// amount
				t, err := strconv.Atoi(opt.Value)
				if err != nil {
					return err
				}
				request.Amount = uint32(t)
			case 'u':
				//user
				request.User = opt.Value
			case 'h':
				s, err := getHelp("gettable", nil)
				if err != nil {
					return err
				}
				fmt.Println(s)
				return nil
			}
		}
	}

	activities, err := client.GetActivities(conn, &request)
	if err != nil {
		return err
	}

	if len(activities) <= 0 {
		return errors.New("no activities found")
	}

	// get the keys of db (doesn't use client.GetKeys since that would incur an unnecessary  request to server)
	var keys []string
	keys = append(keys, "Name")
	for k := range activities[0].Activities {
		keys = append(keys, k)
	}

	// since the way ranges iterate on maps is non-deterministic, store each row in a slice of string slices instead
	var rows [][]string
	// appends keys to rows, for convenience purposes
	rows = append(rows, keys)
	for _, v := range activities {
		var row []string
		row = append(row, v.Name)
		for _, k := range keys {
			// ints are converted to strings since this is just for display purposes anyway
			if k != "Name" {
				row = append(row, strconv.Itoa(int(v.Activities[k])))
			}
		}
		rows = append(rows, row)
	}

	// get the maximum length of every entry for the purpose of printing
	var maxlens []int
	for _, r := range rows {
		for n, v := range r {
			if n > len(maxlens)-1 {
				maxlens = append(maxlens, len(v))
			} else {
				maxlens[n] = max(maxlens[n], len(v))
			}
		}
	}

	var sb strings.Builder
	for _, r := range rows {
		for n, v := range r {
			sb.WriteString(v)
			i := maxlens[n]
			for i > len(v) {
				sb.WriteRune(' ')
				i--
			}
			sb.WriteRune(' ')
		}
		fmt.Println(sb.String())
		sb.Reset()
	}
	return nil
}
