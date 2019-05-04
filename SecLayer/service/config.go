package service

import (
	"time"
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	"sync"
)

type EtcdConf struct {
	EtcdAddr string
	Timeout int
	ProductKey string
}

type SecLayerConf struct {
	Proxy2LayerConf RedisConf
	Layer2ProxyConf RedisConf
	EtcdConf

	ReadGoroutineNum int
	HandleGoroutineNum int
	WriteGoroutineNum int

	RequestTimeout int
	RequestChanTimeout int
	ResponseChanTimeout int

	RequestChanSize int
	ResponseChanSize int

	Secret string
}

type Request struct {
	ProductId string
	ProductNum int
	UserId string
	AccessTime time.Time
	Nonce string
}

type Response struct {
	Code int
	ProductId  string
	ProductNum int
	UserId     string
	Nonce      string
	TokenTime time.Time
	Token string
}

type SecLayerContext struct {
	SecLayerConf

	Proxy2LayerPool *redis.Pool
	Layer2ProxyPool *redis.Pool
	EtcdClient *clientv3.Client

	*ProductMgr
	*HistoryMgr

	WaitGroup sync.WaitGroup
	RequestChan chan *Request
	ResponseChan chan *Response
}
