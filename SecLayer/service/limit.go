package service

import "time"

type SecLimit struct {
	count int
	curTime time.Time
}

func (sl *SecLimit) Check() int {
	now := time.Now()
	if now.Sub(sl.curTime).Seconds() > 1 {
		return 0
	}
	return sl.count
}

func (sl *SecLimit) Add()  {
	now := time.Now()
	if now.Sub(sl.curTime).Seconds() > 1 {
		sl.count = 1
		sl.curTime = now
	} else {
		sl.count++
	}
}
