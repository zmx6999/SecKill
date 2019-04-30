package service

import (
	"time"
	"sync"
	"github.com/astaxie/beego"
	"fmt"
)

type Request struct {
	ProductId string
	ProductNum int
	UserId string
	IP string
	Nonce string
	AccessTime time.Time

	ResponseChan chan *Response `json:"-"`
	CloseNotify <-chan bool `json:"-"`
}

func NewRequest() *Request {
	return &Request{
		ResponseChan: make(chan *Response, secProxyContext.ResponseChanSize),
	}
}

type ResponseMgr struct {
	ResponseMap map[string]chan *Response
	ResponseLock sync.RWMutex
}

func NewResponseMgr() *ResponseMgr {
	return &ResponseMgr{
		ResponseMap: map[string]chan *Response{},
	}
}

type Response struct {
	ProductId string
	ProductNum int
	UserId string
	Code int
	Nonce string
	TokenTime time.Time
	Token string
}

func antiSpam(req *Request) (code int, err error) {
	secProxyContext.LimitMgr.Lock.Lock()

	userLimit, ok := secProxyContext.UserLimitMap[req.UserId]
	if !ok {
		userLimit = NewLimit()
		secProxyContext.UserLimitMap[req.UserId] = userLimit
	}

	ipLimit, ok := secProxyContext.IPLimitMap[req.IP]
	if !ok {
		ipLimit = NewLimit()
		secProxyContext.IPLimitMap[req.IP] = ipLimit
	}

	secProxyContext.LimitMgr.Lock.Unlock()

	userLimit.Lock.Lock()
	defer userLimit.Lock.Unlock()

	if userLimit.SecLimit.Check() >= secProxyContext.UserSecLimit {
		beego.Error("user sec")
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	if userLimit.MinLimit.Check() >= secProxyContext.UserMinLimit {
		beego.Error("user min")
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	ipLimit.Lock.Lock()
	defer ipLimit.Lock.Unlock()

	if ipLimit.SecLimit.Check() >= secProxyContext.IPSecLimit {
		beego.Error("ip sec")
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	if ipLimit.MinLimit.Check() >= secProxyContext.IPMinLimit {
		beego.Error("ip min")
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	userLimit.SecLimit.Add()
	userLimit.MinLimit.Add()
	ipLimit.SecLimit.Add()
	ipLimit.MinLimit.Add()

	return
}

func SecKill(req *Request) (data map[string]interface{}, code int, err error) {
	code, err = antiSpam(req)
	if err != nil {
		return
	}

	code, err = getProduct(req.ProductId)
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s_%s", req.ProductId, req.UserId)

	secProxyContext.ResponseLock.Lock()
	_, ok := secProxyContext.ResponseMap[key]
	if ok {
		secProxyContext.ResponseLock.Unlock()
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	}
	secProxyContext.ResponseMap[key] = req.ResponseChan
	secProxyContext.ResponseLock.Unlock()

	timer := time.NewTicker(time.Second*time.Duration(secProxyContext.RequestChanTimeout))
	select {
	case <-timer.C:
	case secProxyContext.RequestChan<-req:
	}
	timer.Stop()

	timer = time.NewTicker(time.Second*time.Duration(secProxyContext.RequestTimeout))
	defer func() {
		timer.Stop()
		secProxyContext.ResponseLock.Lock()
		delete(secProxyContext.ResponseMap, key)
		secProxyContext.ResponseLock.Unlock()
	}()

	select {
	case <-timer.C:
		code = NetworkBusy
		err = fmt.Errorf(GetErrMsg(code))
		return
	case res := <-req.ResponseChan:
		code = res.Code
		data = map[string]interface{}{}
		data["product_id"] = res.ProductId
		data["product_num"] = res.ProductNum
		data["user_id"] = res.UserId
		data["nonce"] = res.Nonce
		data["token_time"] = res.TokenTime
		data["token"] = res.Token
		return
	case <-req.CloseNotify:
		code = RequestClose
		err = fmt.Errorf(GetErrMsg(code))
		return
	}
}
