package common

import (
	"strconv"
	"strings"

	"github.com/idirall22/twee/pb"
)

// GetFolloweeString return a string contains all follow ids ex: {1,2,3,4}
func GetFolloweeString(in string, followList []*pb.Follow) string {
	out := "{" + in + ","
	if len(followList) == 0 {
		out += "}"
		return out
	}

	for _, f := range followList {
		out += strconv.FormatInt(f.Followee, 10) + ","
	}
	out = strings.TrimRight(out, ",")
	out += "}"
	return out
}
