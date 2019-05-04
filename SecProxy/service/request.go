package service

import "time"

type Request struct {
	ProductId string
	ProductNum int
	IP string
	UserId string
	AccessTime time.Time
	Nonce string

	ResponseChan chan *Response `json:"-"`
	CloseNotify <-chan bool `json:"-"`
}

func NewRequest() *Request {
	return &Request{
		ResponseChan: make(chan *Response, secProxyContext.ResponseChanSize),
	}
}
