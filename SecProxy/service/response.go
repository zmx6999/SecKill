package service

import (
	"time"
	"sync"
)

type Response struct {
	Code int
	ProductId  string
	ProductNum int
	UserId     string
	Nonce      string
	TokenTime time.Time
	Token string
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
