package service

import (
	"fmt"
	"time"
)

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
		err = fmt.Errorf(GetMsg(code))
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
	select {
	case r := <-req.ResponseChan:
		code = r.Code
		data = map[string]interface{}{}
		data["product_id"] = r.ProductId
		data["product_num"] = r.ProductNum
		data["user_id"] = r.UserId
		data["nonce"] = r.Nonce
		data["token_time"] = r.TokenTime
		data["token"] = r.Token
	case <-timer.C:
		code = Timeout
		err = fmt.Errorf(GetMsg(code))
	case <-req.CloseNotify:
		code = NotifyClosed
		err = fmt.Errorf(GetMsg(code))
	}
	timer.Stop()

	secProxyContext.ResponseLock.Lock()
	delete(secProxyContext.ResponseMap, key)
	secProxyContext.ResponseLock.Unlock()

	return
}
