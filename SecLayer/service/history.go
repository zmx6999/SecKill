package service

import "sync"

type HistoryMgr struct {
	UserHistoryMap map[string]*UserHistory
	UserHistoryLock sync.RWMutex
}

func NewHistoryMgr() *HistoryMgr {
	return &HistoryMgr{
		UserHistoryMap: map[string]*UserHistory{},
	}
}

type UserHistory struct {
	buyProductMap map[string]int
}

func NewUserHistory() *UserHistory {
	return &UserHistory{
		map[string]int{},
	}
}

func (uh *UserHistory) Check(productId string) int {
	return uh.buyProductMap[productId]
}

func (uh *UserHistory) Add(productId string, num int)  {
	uh.buyProductMap[productId] += num
}
