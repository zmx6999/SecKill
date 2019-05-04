package service

import (
	"time"
	"sync"
	"fmt"
	"github.com/astaxie/beego"
)

type AccessMgr struct {
	UserAccessMap map[string]*Access
	IPAccessMap map[string]*Access
	AccessMgrLock sync.RWMutex
}

func NewAccessMgr() *AccessMgr {
	return &AccessMgr{
		UserAccessMap: map[string]*Access{},
		IPAccessMap: map[string]*Access{},
	}
}

func antiSpam(req *Request) (code int, err error) {
	secProxyContext.AccessMgrLock.Lock()

	userAccess, ok := secProxyContext.UserAccessMap[req.UserId]
	if !ok {
		userAccess = NewAccess()
		secProxyContext.UserAccessMap[req.UserId] = userAccess
	}

	ipAccess, ok := secProxyContext.IPAccessMap[req.IP]
	if !ok {
		ipAccess = NewAccess()
		secProxyContext.IPAccessMap[req.IP] = ipAccess
	}

	secProxyContext.AccessMgrLock.Unlock()

	userAccess.AccessLock.Lock()
	defer userAccess.AccessLock.Unlock()

	if userAccess.SecAccess.Check() >= secProxyContext.UserSecAccessLimit {
		beego.Error("user sec")
		code = NetworkBusy
		err = fmt.Errorf(GetMsg(code))
		return
	}
	if userAccess.MinAccess.Check() >= secProxyContext.UserMinAccessLimit {
		beego.Error("user min")
		code = NetworkBusy
		err = fmt.Errorf(GetMsg(code))
		return
	}

	ipAccess.AccessLock.Lock()
	defer ipAccess.AccessLock.Unlock()

	if ipAccess.SecAccess.Check() >= secProxyContext.IPSecAccessLimit {
		beego.Error("ip sec")
		code = NetworkBusy
		err = fmt.Errorf(GetMsg(code))
		return
	}
	if ipAccess.MinAccess.Check() >= secProxyContext.IPMinAccessLimit {
		beego.Error("ip min")
		code = NetworkBusy
		err = fmt.Errorf(GetMsg(code))
		return
	}

	userAccess.SecAccess.Add()
	userAccess.MinAccess.Add()
	ipAccess.SecAccess.Add()
	ipAccess.MinAccess.Add()

	return
}

type Access struct {
	SecAccess *SecAccess
	MinAccess *MinAccess
	AccessLock sync.RWMutex
}

func NewAccess() *Access {
	return &Access{
		SecAccess: &SecAccess{},
		MinAccess: &MinAccess{},
	}
}

type TimeAccess struct {
	num int
	curTime time.Time
}

func (ta *TimeAccess) check(period float64) int {
	now := time.Now()
	if now.Sub(ta.curTime).Seconds() > period {
		return 0
	}
	return ta.num
}

func (ta *TimeAccess) add(period float64)  {
	now := time.Now()
	if now.Sub(ta.curTime).Seconds() > period {
		ta.num = 1
		ta.curTime = now
	} else {
		ta.num++
	}
}

type SecAccess struct {
	TimeAccess
}

func (sa *SecAccess) Check() int {
	return sa.check(1)
}

func (sa *SecAccess) Add()  {
	sa.add(1)
}

type MinAccess struct {
	TimeAccess
}

func (ma *MinAccess) Check() int {
	return ma.check(60)
}

func (ma *MinAccess) Add()  {
	ma.add(60)
}
