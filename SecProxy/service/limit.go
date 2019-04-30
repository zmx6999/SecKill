package service

import (
	"time"
	"sync"
)

type LimitMgr struct {
	UserLimitMap map[string]*Limit
	IPLimitMap map[string]*Limit
	Lock sync.RWMutex
}

func NewLimitMgr() *LimitMgr {
	return &LimitMgr{
		UserLimitMap: map[string]*Limit{},
		IPLimitMap: map[string]*Limit{},
	}
}

type Limit struct {
	SecLimit TimeLimit
	MinLimit TimeLimit
	Lock sync.RWMutex
}

func NewLimit() *Limit {
	return &Limit{
		SecLimit: &SecLimit{},
		MinLimit: &MinLimit{},
	}
}

type TimeLimit interface {
	Check() int
	Add()
}

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

func (sl *SecLimit) Add() {
	now := time.Now()
	if now.Sub(sl.curTime).Seconds() > 1 {
		sl.count = 1
		sl.curTime = now
		return
	}
	sl.count++
}

type MinLimit struct {
	count int
	curTime time.Time
}

func (ml *MinLimit) Check() int {
	now := time.Now()
	if now.Sub(ml.curTime).Seconds() > 60 {
		return 0
	}
	return ml.count
}

func (ml *MinLimit) Add() {
	now := time.Now()
	if now.Sub(ml.curTime).Seconds() > 60 {
		ml.count = 1
		ml.curTime = now
		return
	}
	ml.count++
}
