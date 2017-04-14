package pcf

import "time"

type realTimeHelper struct {}

func (r *realTimeHelper) currentTimeInMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}