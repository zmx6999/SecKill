package service

import (
		"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
	)

type EtcdConf struct {
	EtcdAddr string
	Timeout int
	ProductKey string
}

type AccessLimitConf struct {
	UserSecAccessLimit int
	UserMinAccessLimit int
	IPSecAccessLimit int
	IPMinAccessLimit int
}

type SecProxyConf struct {
	Proxy2LayerConf RedisConf
	Layer2ProxyConf RedisConf
	EtcdConf

	AccessLimitConf

	ReadGoroutineNum int
	WriteGoroutineNum int

	RequestTimeout int
	RequestChanTimeout int
	ResponseChanTimeout int

	RequestChanSize int
	ResponseChanSize int
}

type SecProxyContext struct {
	SecProxyConf

	Proxy2LayerPool *redis.Pool
	Layer2ProxyPool *redis.Pool
	EtcdClient *clientv3.Client

	*ProductMgr
	*AccessMgr

	RequestChan chan *Request
	*ResponseMgr
}
