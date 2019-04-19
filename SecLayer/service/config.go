package service

import (
	"time"
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	"sync"
)

type RedisConf struct {
	RedisAddr string
	RedisPassword string
	RedisMaxIdle int
	RedisMaxActive int
	RedisIdleTimeout int
	RedisQueueName string
}

type EtcdConf struct {
	EtcdAddr string
	EtcdTimeout int
	EtcdProductKey string
}

type SecLayerConf struct {
	Proxy2LayerRedis RedisConf
	Layer2ProxyRedis RedisConf
	EtcdConf

	WriteGoroutineNum int
	ReadGoroutineNum int
	HandleUserGoroutineNum int
	Read2HandleChanSize int
	Handle2WriteChanSize int
	MaxRequestWaitTimeout int

	Send2WriteChanTimeout int
	Send2HandleChanTimeout int

	LimitPeriod int
	TokenPassword string

	ProductOnePersonBuyLimit int
	ProductSecSoldMaxLimit int
	ProductBuyRate float64
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

type SecRequest struct {
	ProductId string
	Source string
	AuthCode string
	SecTime string
	Nonce string
	UserId string
	UserAuthSign string
	AccessTime time.Time
	ClientAddr string
	// ClientReference string
}

type SecResponse struct {
	ProductId string
	UserId string
	Token string
	TokenTime time.Time
	Code int
}
