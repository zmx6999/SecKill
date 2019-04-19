package service

import (
	"time"
	)

type Limit struct {
	count int
	curTime time.Time
}

func (sl *Limit) Count(now time.Time, period int) {
	if int(now.Sub(sl.curTime).Seconds()) > period {
		sl.count = 1
		sl.curTime = now
	} else {
		sl.count++
	}
}

func (sl *Limit) Check(now time.Time, period int) int {
	if int(now.Sub(sl.curTime).Seconds()) > period {
		return 0
	}
	return sl.count
}
