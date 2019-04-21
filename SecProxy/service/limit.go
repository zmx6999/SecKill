package service

import (
	"time"
	"sync"
	"fmt"
	"github.com/astaxie/beego"
)

type UserLimitMgr struct {
	UserLimitMap map[string]*UserLimit
	IPLimitMap map[string]*UserLimit
	Lock sync.RWMutex
}

func newUserLimitMgr() UserLimitMgr {
	return UserLimitMgr{
		UserLimitMap: map[string]*UserLimit{},
		IPLimitMap: map[string]*UserLimit{},
	}
}

func antiSpam(req *SecRequest) (code int, err error) {
	secProxyContext.UserLimitMgr.Lock.Lock()

	userLimit, ok := secProxyContext.UserLimitMap[req.UserId]
	if !ok {
		userLimit = NewUserLimit()
		secProxyContext.UserLimitMap[req.UserId] = userLimit
	}

	ipLimit, ok := secProxyContext.IPLimitMap[req.IP]
	if !ok {
		ipLimit = NewUserLimit()
		secProxyContext.IPLimitMap[req.IP] = ipLimit
	}

	secProxyContext.UserLimitMgr.Lock.Unlock()

	userLimit.Lock.Lock()
	defer userLimit.Lock.Unlock()

	userSecNum := userLimit.SecLimit.Check()
	if userSecNum >= secProxyContext.UserSecLimit {
		beego.Info("UserSecLimit")
		code = NetworkBusyErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	userMinNum := userLimit.MinLimit.Check()
	if userMinNum >= secProxyContext.UserMinLimit {
		beego.Info("UserMinLimit")
		code = NetworkBusyErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	ipLimit.Lock.Lock()
	defer ipLimit.Lock.Unlock()

	ipSecNum := ipLimit.SecLimit.Check()
	if ipSecNum >= secProxyContext.IPSecLimit {
		beego.Info("IPSecLimit")
		code = NetworkBusyErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	ipMinNum := ipLimit.MinLimit.Check()
	if ipMinNum >= secProxyContext.IPMinLimit {
		beego.Info("IPMinLimit")
		code = NetworkBusyErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	userLimit.SecLimit.Add()
	userLimit.MinLimit.Add()
	ipLimit.SecLimit.Add()
	ipLimit.MinLimit.Add()

	return
}

type UserLimit struct {
	SecLimit TimeLimit
	MinLimit TimeLimit
	Lock sync.RWMutex
}

func NewUserLimit() *UserLimit {
	return &UserLimit{
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

func (sl *SecLimit) Add()  {
	now := time.Now()
	if now.Sub(sl.curTime).Seconds() > 1 {
		sl.count = 1
		sl.curTime = now
	} else {
		sl.count++
	}
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

func (ml *MinLimit) Add()  {
	now := time.Now()
	if now.Sub(ml.curTime).Seconds() > 60 {
		ml.count = 1
		ml.curTime = now
	} else {
		ml.count++
	}
}
