package service

import (
	"time"
	"sync"
	"fmt"
	)

type LimitMgr struct {
	UserLimitMap map[string]*Limit
	IPLimitMap map[string]*Limit
	LimitMgrLock sync.RWMutex
}

func antiSpam(secProxyContext *SecProxyContext, req *SecRequest) (code int, err error) {
	/*
	secProxyContext.BlockLock.RLock()

	_, ok := secProxyContext.UserBlockMap[req.UserId]
	if ok {
		secProxyContext.BlockLock.RUnlock()
		return fmt.Errorf(ErrMsg[UserBlockErr])
	}

	_, ok = secProxyContext.IPBlockMap[req.RemoteAddr]
	if ok {
		secProxyContext.BlockLock.RUnlock()
		return fmt.Errorf(ErrMsg[IPBlockErr])
	}

	secProxyContext.BlockLock.RUnlock()
	*/

	secProxyContext.LimitMgr.LimitMgrLock.Lock()

	userLimit, ok := secProxyContext.LimitMgr.UserLimitMap[req.UserId]
	if !ok {
		userLimit = NewLimit()
		secProxyContext.LimitMgr.UserLimitMap[req.UserId] = userLimit
	}

	ipLimit, ok := secProxyContext.LimitMgr.IPLimitMap[req.RemoteAddr]
	if !ok {
		ipLimit = NewLimit()
		secProxyContext.LimitMgr.IPLimitMap[req.RemoteAddr] = ipLimit
	}

	secProxyContext.LimitMgr.LimitMgrLock.Unlock()

	userLimit.LimitLock.Lock()
	defer userLimit.LimitLock.Unlock()

	now := time.Now()
	userSecCount := userLimit.SecLimit.Check(now)
	if userSecCount >= secProxyContext.AccessLimit.UserSecAccessLimit {
		code =  NetworkBusyErr
		err = fmt.Errorf(ErrMsg[code])
		return
	}

	userMinCount := userLimit.MinLimit.Check(now)
	if userMinCount >= secProxyContext.AccessLimit.UserMinAccessLimit {
		code =  NetworkBusyErr
		err = fmt.Errorf(ErrMsg[code])
		return
	}

	ipLimit.LimitLock.Lock()
	defer ipLimit.LimitLock.Unlock()

	ipSecCount := ipLimit.SecLimit.Check(now)
	if ipSecCount >= secProxyContext.AccessLimit.IPSecAccessLimit {
		code =  NetworkBusyErr
		err = fmt.Errorf(ErrMsg[code])
		return
	}

	ipMinCount := ipLimit.MinLimit.Check(now)
	if ipMinCount >= secProxyContext.AccessLimit.IPMinAccessLimit {
		code =  NetworkBusyErr
		err = fmt.Errorf(ErrMsg[code])
		return
	}

	secProxyContext.LimitMgr.UserLimitMap[req.UserId].SecLimit.Count(now)
	secProxyContext.LimitMgr.UserLimitMap[req.UserId].MinLimit.Count(now)
	secProxyContext.LimitMgr.IPLimitMap[req.RemoteAddr].SecLimit.Count(now)
	secProxyContext.LimitMgr.IPLimitMap[req.RemoteAddr].MinLimit.Count(now)
	return
}

type Limit struct {
	SecLimit TimeLimit
	MinLimit TimeLimit
	LimitLock sync.RWMutex
}

func NewLimit() *Limit {
	return &Limit{
		SecLimit: &SecLimit{},
		MinLimit: &MinLimit{},
	}
}

type TimeLimit interface {
	Check(now time.Time) int
	Count(now time.Time)
}

type SecLimit struct {
	count int
	curTime time.Time
}

func (sl *SecLimit) Check(now time.Time) int {
	if int(now.Sub(sl.curTime).Seconds()) > 1 {
		return 0
	}

	return sl.count
}

func (sl *SecLimit) Count(now time.Time)  {
	if int(now.Sub(sl.curTime).Seconds()) > 1 {
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

func (ml *MinLimit) Check(now time.Time) int {
	if int(now.Sub(ml.curTime).Seconds()) > 60 {
		return 0
	}

	return ml.count
}

func (ml *MinLimit) Count(now time.Time)  {
	if int(now.Sub(ml.curTime).Seconds()) > 60 {
		ml.count = 1
		ml.curTime = now
		return
	}
	ml.count++
}
