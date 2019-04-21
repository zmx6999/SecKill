package service

import (
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	"time"
	"sync"
)

type EtcdConf struct {
	EtcdAddr string
	EtcdTimeout int
	EtcdProductKey string
}

type SecLayerConf struct {
	Proxy2LayerRedisConf RedisConf
	Layer2ProxyRedisConf RedisConf
	EtcdConf

	ReadGoroutineNum int
	WriteGoroutineNum int
	HandleUserGoroutineNum int

	RequestTimeout int
	Read2HandleChanTimeout int
	Handle2WriteChanTimeout int

	Read2HandleChanSize int
	Handle2WriteChanSize int

	ProductSecSoldLimit int
	ProductOnePersonBuyLimit int
	ProductSoldRate float64

	LayerSecret string
}

type SecRequest struct {
	UserId string
	ProductId string
	AccessTime time.Time
	Nonce string
}

type SecResponse struct {
	UserId    string
	ProductId string
	Code int
	Token string
	TokenTime time.Time
	Nonce string
}

type SecLayerContext struct {
	SecLayerConf

	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	EtcdClient *clientv3.Client

	WaitGroup sync.WaitGroup
	Read2HandleChan chan *SecRequest
	Handle2WriteChan chan *SecResponse

	ProductMgr
	UserHistoryMgr
}
