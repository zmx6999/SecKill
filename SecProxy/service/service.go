package service

import (
	"time"
	"fmt"
	"sync"
	)

/*
func ValidateUser(request SecRequest) bool {
	secret := getUserSecret(request.UserId)
	str := "user_id="+strconv.Itoa(request.UserId)+"&secret="+secret
	authSign := GetSha256(str)
	return request.UserAuthSign == authSign
}
*/

type SecRequest struct {
	ProductId string
	UserId string
	RemoteAddr string
	UserAuthSign string
	AccessTime time.Time
	Source string
	AuthCode string
	Nonce string
	SecTime time.Time
	// ClientReference string
	CloseNotify <-chan bool `json:"-"`

	*ResponseChan `json:"-"`
}

func NewSecRequest() *SecRequest {
	return &SecRequest{
		ResponseChan: newResponseChan(),
	}
}

type SecResponse struct {
	ProductId string
	UserId string
	Token string
	Code int
}

type ResponseChanMgr struct {
	ResponseChanMap  map[string]*ResponseChan
	ResponseChanMapLock sync.RWMutex
}

type ResponseChan struct {
	Chan chan *SecResponse
	ChanLock sync.RWMutex
}

func newResponseChan() *ResponseChan {
	return &ResponseChan{
		Chan: make(chan *SecResponse, 1),
	}
}

func SecKill(request *SecRequest) (data map[string]interface{}, code int, err error) {
	code, err = antiSpam(secProxyContext, request)
	if err != nil {
		return
	}

	_, code, err = GetProduct(request.ProductId)
	if err != nil || code != OK {
		if err == nil {
			err = fmt.Errorf(ErrMsg[code])
		}
		return
	}

	secProxyContext.ResponseChanMapLock.Lock()

	key := request.UserId+"_"+request.ProductId
	if _, ok := secProxyContext.ResponseChanMap[key]; ok {
		code = NetworkBusyErr
		err = fmt.Errorf(ErrMsg[code])
		secProxyContext.ResponseChanMapLock.Unlock()
		return
	}
	secProxyContext.ResponseChanMap[key] = request.ResponseChan

	secProxyContext.ResponseChanMapLock.Unlock()

	secProxyContext.RequestChan <- request

	timer := time.NewTicker(time.Second * 60)

	defer func() {
		timer.Stop()
		secProxyContext.ResponseChanMapLock.Lock()
		delete(secProxyContext.ResponseChanMap, key)
		secProxyContext.ResponseChanMapLock.Unlock()
	}()

	select {
	case <-timer.C:
		code = TimeoutErr
		err = fmt.Errorf(ErrMsg[code])
		return
	case <-request.CloseNotify:
		code = ClientClosedErr
		err = fmt.Errorf(ErrMsg[code])
		return
	case res := <-request.ResponseChan.Chan:
		code = res.Code
		data = map[string]interface{}{}
		data["product_id"] = res.ProductId
		data["user_id"] = res.UserId
		data["token"] = res.Token
		return
	}

	return
}
