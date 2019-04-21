package service

import (
	"time"
		"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
)

type EtcdConf struct {
	EtcdAddr string
	EtcdTimeout int
	EtcdProductKey string
}

type SecProxyConf struct {
	Proxy2LayerRedisConf RedisConf
	Layer2ProxyRedisConf RedisConf
	EtcdConf

	ReadGoroutineNum int
	WriteGoroutineNum int

	RequestTimeout int
	RequestChanTimeout int
	ResponseChanTimeout int

	RequestChanSize int
	ResponseChanSize int

	UserMinLimit int
	UserSecLimit int
	IPMinLimit int
	IPSecLimit int
}

type SecRequest struct {
	UserId string
	ProductId string
	AccessTime time.Time
	Nonce string
	IP string

	ResponseChan chan *SecResponse `json:"-"`
	CloseNotify <- chan bool `json:"-"`
}

type SecResponse struct {
	UserId    string
	ProductId string
	Code int
	Token string
	TokenTime time.Time
	Nonce string
}

type SecProxyContext struct {
	SecProxyConf

	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	EtcdClient *clientv3.Client

	RequestChan chan *SecRequest
	ResponseChanMgr

	ProductMgr
	UserLimitMgr
}
