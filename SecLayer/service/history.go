package service

import "sync"

type UserHistoryMgr struct {
	UserHistoryMap map[string]*UserHistory
	UserHistoryLock sync.RWMutex
}

func NewUserHistoryMgr() *UserHistoryMgr {
	return &UserHistoryMgr{
		UserHistoryMap: map[string]*UserHistory{},
	}
}

type UserHistory struct {
	historyMap map[string]int
}

func NewUserHistory() *UserHistory {
	return &UserHistory{
		map[string]int{},
	}
}

func (uh *UserHistory) Check(productId string) int {
	return uh.historyMap[productId]
}

func (uh *UserHistory) Add(productId string, num int)  {
	uh.historyMap[productId] += num
}
