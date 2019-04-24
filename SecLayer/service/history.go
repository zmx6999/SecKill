package service

import "sync"

type UserHistoryMgr struct {
	UserHistoryMap map[string]*UserHistory
	UserHistoryLock sync.RWMutex
}

func newUserHistoryMgr() UserHistoryMgr {
	return UserHistoryMgr{
		UserHistoryMap: map[string]*UserHistory{},
	}
}

type UserHistory struct {
	historyMap map[string]int
}

func newUserHistory() *UserHistory {
	return &UserHistory{
		map[string]int{},
	}
}

func (uh *UserHistory) Check(productId string) int {
	return uh.historyMap[productId]
}

func (uh *UserHistory) Add(productId string, count int)  {
	uh.historyMap[productId] += count
}
