package utils

import "strings"

func JoinQuoted(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = `"` + a + `"`
	}
	return strings.Join(quoted, ", ")
}
