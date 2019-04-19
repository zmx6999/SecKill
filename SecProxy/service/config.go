package service

import (
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
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

type AccessLimit struct {
	IPSecAccessLimit int
	UserSecAccessLimit int
	IPMinAccessLimit int
	UserMinAccessLimit int
}

type SecProxyConf struct {
	// RedisBlockConf RedisConf
	RedisProxy2LayerConf RedisConf
	RedisLayer2ProxyConf RedisConf

	EtcdConf

	AccessLimit

	ReadGoroutineNum int
	WriteGoroutineNum int
	RequestChanSize int

	// CookieSecretKey string
	// ReferWhiteList []string
}

type SecProxyContext struct {
	SecProxyConf

	// BlockRedisPool *redis.Pool
	Proxy2LayerRedisPool *redis.Pool
	Layer2ProxyRedisPool *redis.Pool
	EtcdClient *clientv3.Client

	/*
	UserBlockMap map[string]bool
	IPBlockMap map[string]bool
	BlockLock sync.RWMutex
	*/

	ProductMgr
	LimitMgr

	RequestChan chan *SecRequest
	ResponseChanMgr
}
