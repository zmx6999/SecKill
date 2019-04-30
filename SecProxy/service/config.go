package service

import (
	"github.com/garyburd/redigo/redis"
	"go.etcd.io/etcd/clientv3"
		)

type EtcdConf struct {
	EtcdAddr string
	DialTimeout int
	ProductKey string
}

type AccessLimitConf struct {
	UserSecLimit int
	UserMinLimit int
	IPSecLimit int
	IPMinLimit int
}

type SecProxyConf struct {
	ProxyToLayerConf RedisConf
	LayerToProxyConf RedisConf
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

	ProxyToLayerPool *redis.Pool
	LayerToProxyPool *redis.Pool
	EtcdClient *clientv3.Client

	RequestChan chan *Request
	*ResponseMgr

	*ProductMgr
	*LimitMgr
}
