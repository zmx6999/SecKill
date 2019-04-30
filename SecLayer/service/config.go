package service

import (
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	"time"
	"sync"
)

type EtcdConf struct {
	EtcdAddr string
	DialTimeout int
	ProductKey string
}

type SecLayerConf struct {
	ProxyToLayerConf RedisConf
	LayerToProxyConf RedisConf
	EtcdConf

	ReadGoroutineNum int
	HandleUserGoroutineNum int
	WriteGoroutineNum int

	RequestTimeout int
	ReadToHandleChanTimeout int
	HandleToWriteChanTimeout int

	ReadToHandleChanSize int
	HandleToWriteChanSize int

	LayerSecret string
}

type Request struct {
	ProductId string
	ProductNum int
	UserId string
	Nonce string
	AccessTime time.Time
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

type SecLayerContext struct {
	SecLayerConf

	ProxyToLayerPool *redis.Pool
	LayerToProxyPool *redis.Pool
	EtcdClient *clientv3.Client

	WaitGroup sync.WaitGroup
	ReadToHandleChan chan *Request
	HandleToWriteChan chan *Response

	*ProductMgr
	*UserHistoryMgr
}
