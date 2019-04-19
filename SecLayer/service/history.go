package service

import "sync"

type UserHistoryMgr struct {
	UserHistoryMap map[string]*UserHistory
	UserHistoryLock sync.RWMutex
}

type UserHistory struct {
	history map[string]int
}

func NewUserHistory() (*UserHistory) {
	return &UserHistory{history: map[string]int{}}
}

func (uh *UserHistory) Get(productId string) int {
	return uh.history[productId]
}

func (uh *UserHistory) Add(productId string, count int) {
	uh.history[productId] += count
}
