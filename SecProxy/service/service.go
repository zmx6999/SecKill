package service

import (
	"sync"
	"fmt"
	"time"
)

func NewSecRequest() *SecRequest {
	return &SecRequest{
		ResponseChan: make(chan *SecResponse, secProxyContext.ResponseChanSize),
	}
}

type ResponseChanMgr struct {
	ResponseChanMap map[string]chan *SecResponse
	Lock sync.RWMutex
}

func newResponseChanMgr() ResponseChanMgr {
	return ResponseChanMgr{
		ResponseChanMap: map[string]chan *SecResponse{},
	}
}

func SecKill(req *SecRequest) (data map[string]interface{}, code int, err error) {
	code, err = antiSpam(req)
	if err != nil {
		return
	}

	code, err = getProduct(req.ProductId)
	if err != nil {
		return
	}

	secProxyContext.ResponseChanMgr.Lock.Lock()

	key := fmt.Sprintf("%s_%s", req.UserId, req.ProductId)
	_, ok := secProxyContext.ResponseChanMap[key]
	if ok {
		secProxyContext.ResponseChanMgr.Lock.Unlock()
		code = NetworkBusyErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}
	secProxyContext.ResponseChanMap[key] = req.ResponseChan

	secProxyContext.ResponseChanMgr.Lock.Unlock()

	defer func() {
		secProxyContext.ResponseChanMgr.Lock.Lock()
		delete(secProxyContext.ResponseChanMap, key)
		secProxyContext.ResponseChanMgr.Lock.Unlock()
	}()

	timer := time.NewTicker(time.Second*time.Duration(secProxyContext.RequestChanTimeout))
	select {
	case secProxyContext.RequestChan <- req:
		timer.Stop()
	case <- timer.C:
		timer.Stop()
		code = TimeoutErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	timer = time.NewTicker(time.Second*time.Duration(secProxyContext.RequestTimeout))
	select {
	case <- timer.C:
		timer.Stop()
		code = TimeoutErr
		err = fmt.Errorf(getErrMsg(code))
		return
	case <- req.CloseNotify:
		timer.Stop()
		code = CloseRequestErr
		err = fmt.Errorf(getErrMsg(code))
		return
	case res := <-req.ResponseChan:
		timer.Stop()
		code = res.Code
		data = map[string]interface{}{}
		data["product_id"] = res.ProductId
		data["user_id"] = res.UserId
		data["token"] = res.Token
		data["token_time"] = res.TokenTime.Format("2006-01-02 15:04:05")
		data["nonce"] = res.Nonce
		return
	}
}
