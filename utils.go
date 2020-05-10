package discordgoext

import (
	"strconv"
	"time"
)

// GetCreationTime is a helper function to get the time of creation of any ID.
// ID: ID to get the time from
func GetCreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}
