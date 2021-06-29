package firebase

import "time"

func canChange(t time.Time) bool {
	return time.Duration(time.Now().UnixNano()-t.UnixNano()) > time.Second
}
