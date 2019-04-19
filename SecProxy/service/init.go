package service

import (
	"time"
	"go.etcd.io/etcd/clientv3"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
)

var secProxyContext *SecProxyContext

func initEtcd(secProxyContext *SecProxyContext) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdConf.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.EtcdConf.EtcdTimeout),
	})
	if err != nil {
		return err
	}
	secProxyContext.EtcdClient = cli

	return nil
}

func initSecProxyConf(secProxyContext *SecProxyContext) error {
	conf, err := config.NewConfig("ini", "./conf/app.conf")
	if err != nil {
		return err
	}

	secProxyConf := SecProxyConf{}
	secProxyConf.RedisProxy2LayerConf.RedisAddr = conf.String("proxy2layer_redis_addr")
	secProxyConf.RedisProxy2LayerConf.RedisMaxIdle, _ = conf.Int("proxy2layer_redis_max_idle")
	secProxyConf.RedisProxy2LayerConf.RedisMaxActive, _ = conf.Int("proxy2layer_redis_max_active")
	secProxyConf.RedisProxy2LayerConf.RedisIdleTimeout, _ = conf.Int("proxy2layer_redis_idle_timeout")
	secProxyConf.RedisProxy2LayerConf.RedisQueueName = conf.String("proxy2layer_redis_queue_name")
	secProxyConf.RedisProxy2LayerConf.RedisPassword = conf.String("proxy2layer_redis_password")

	secProxyConf.RedisLayer2ProxyConf.RedisAddr = conf.String("layer2proxy_redis_addr")
	secProxyConf.RedisLayer2ProxyConf.RedisMaxIdle, _ = conf.Int("layer2proxy_redis_max_idle")
	secProxyConf.RedisLayer2ProxyConf.RedisMaxActive, _ = conf.Int("layer2proxy_redis_max_active")
	secProxyConf.RedisLayer2ProxyConf.RedisIdleTimeout, _ = conf.Int("layer2proxy_redis_idle_timeout")
	secProxyConf.RedisLayer2ProxyConf.RedisQueueName = conf.String("layer2proxy_redis_queue_name")
	secProxyConf.RedisLayer2ProxyConf.RedisPassword = conf.String("layer2proxy_redis_password")

	secProxyConf.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secProxyConf.EtcdConf.EtcdTimeout, _ = conf.Int("etcd_timeout")
	secProxyConf.EtcdConf.EtcdProductKey = conf.String("etcd_product_key")

	secProxyConf.AccessLimit.IPMinAccessLimit, _ = conf.Int("ip_min_access_limit")
	secProxyConf.AccessLimit.IPSecAccessLimit, _ = conf.Int("ip_sec_access_limit")
	secProxyConf.AccessLimit.UserMinAccessLimit, _ = conf.Int("user_min_access_limit")
	secProxyConf.AccessLimit.UserSecAccessLimit, _ = conf.Int("user_sec_access_limit")

	secProxyConf.ReadGoroutineNum, _ = conf.Int("read_goroutine_num")
	secProxyConf.WriteGoroutineNum, _ = conf.Int("write_goroutine_num")
	secProxyConf.RequestChanSize, _ = conf.Int("request_chan_size")

	secProxyContext.SecProxyConf = secProxyConf

	return nil
}

func initSecLayerContext(secProxyContext *SecProxyContext) error {
	err := initSecProxyConf(secProxyContext)
	if err != nil {
		return err
	}

	initRedis(secProxyContext)

	err = initEtcd(secProxyContext)
	if err != nil {
		return err
	}

	err = initProduct(secProxyContext)
	if err != nil {
		return err
	}

	secProxyContext.LimitMgr.UserLimitMap = map[string]*Limit{}
	secProxyContext.LimitMgr.IPLimitMap = map[string]*Limit{}

	secProxyContext.RequestChan = make(chan *SecRequest, secProxyContext.RequestChanSize)

	secProxyContext.ResponseChanMgr.ResponseChanMap = map[string]*ResponseChan{}

	return nil
}

func InitService()  {
	secProxyContext = &SecProxyContext{}
	err := initSecLayerContext(secProxyContext)
	if err != nil {
		beego.Error(err)
		return
	}
	run(secProxyContext)
}
