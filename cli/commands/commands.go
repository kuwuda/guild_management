package commands

import (
	"strings"
)

// I really am not sure what a good way to do all this stuff is
func getHelp(command string, keys []string) (string, error) {
	var help strings.Builder

	switch command {
	case "gettable":
		help.WriteString("usage: " + command + " [options]\n")
		help.WriteString("Options:\n")
		help.WriteString("-n Specifies the maximum amount of results to return\n" +
			"-u Specifies the specific user to get\n" +
			"-h Prints this dialog\n" +
			"If no options are input, will default to returning every user.")
		return help.String(), nil
	case "adduser":
		help.WriteString("usage: newActivityMember [options] name ")
		for _, v := range keys {
			help.WriteRune('"')
			help.WriteString(v)
			help.WriteRune('"')
			help.WriteRune(' ')
		}
		help.WriteString("\nWhere name etc. is replaced with the value for the input\n" +
			"EX:\n" +
			"newActivityMember Apples 4 8\n")
		help.WriteString("Any non-specified values will default to 0\n Extraneous values are ignored")
		help.WriteString("Options:\n " +
			"-h Prints this message")
		return help.String(), nil
	case "addcol":
		help.WriteString("usage: newColumn [options] key")
		help.WriteString("\nWhere key is the name of the column being created\n" +
			"EX:\n" +
			"newColumn Cats\n")
		help.WriteString("Options:\n " +
			"-h Prints this message")
		return help.String(), nil
	case "incact":
		help.WriteString("usage: incActivity [options] activity members")
		help.WriteString("\nWhere activity is the name of the activity being incremented, and" +
			"members is the total list of users who are to be incremented\n" +
			"EX:\n" +
			"incActivity -n 4 kraken kusu apples tyny\n" +
			"Would increase the activity \"kraken\" for all users by 4")
		help.WriteString("Options:\n " +
			"-h Prints this message\n" +
			"-n Specify the amount to increment by (defaults to 1)")
		return help.String(), nil
	}
	return "", nil
}
