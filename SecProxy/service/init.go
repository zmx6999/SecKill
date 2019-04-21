package service

import (
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
	"time"
	"go.etcd.io/etcd/clientv3"
)

var (
	secProxyContext *SecProxyContext
)

func initEtcd(etcdConf EtcdConf) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdConf.EtcdAddr},
		DialTimeout: time.Second*time.Duration(etcdConf.EtcdTimeout),
	})
	if err != nil {
		return err
	}
	secProxyContext.EtcdClient = cli

	return nil
}

func initSecProxyConf() error {
	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		return err
	}

	secProxyContext.Proxy2LayerRedisConf.RedisAddr = conf.String("proxy2layer_redis_addr")
	secProxyContext.Proxy2LayerRedisConf.RedisPassword = conf.String("proxy2layer_redis_password")
	secProxyContext.Proxy2LayerRedisConf.RedisQueueName = conf.String("proxy2layer_redis_queue_name")
	secProxyContext.Proxy2LayerRedisConf.RedisMaxIdle, err = conf.Int("proxy2layer_redis_max_idle")
	if err != nil {
		return err
	}
	secProxyContext.Proxy2LayerRedisConf.RedisMaxActive, err = conf.Int("proxy2layer_redis_max_active")
	if err != nil {
		return err
	}
	secProxyContext.Proxy2LayerRedisConf.RedisIdleTimeout, err = conf.Int("proxy2layer_redis_idle_timeout")
	if err != nil {
		return err
	}

	secProxyContext.Layer2ProxyRedisConf.RedisAddr = conf.String("layer2proxy_redis_addr")
	secProxyContext.Layer2ProxyRedisConf.RedisPassword = conf.String("layer2proxy_redis_password")
	secProxyContext.Layer2ProxyRedisConf.RedisQueueName = conf.String("layer2proxy_redis_queue_name")
	secProxyContext.Layer2ProxyRedisConf.RedisMaxIdle, err = conf.Int("layer2proxy_redis_max_idle")
	if err != nil {
		return err
	}
	secProxyContext.Layer2ProxyRedisConf.RedisMaxActive, err = conf.Int("layer2proxy_redis_max_active")
	if err != nil {
		return err
	}
	secProxyContext.Layer2ProxyRedisConf.RedisIdleTimeout, err = conf.Int("layer2proxy_redis_idle_timeout")
	if err != nil {
		return err
	}

	secProxyContext.EtcdAddr = conf.String("etcd_addr")
	secProxyContext.EtcdProductKey = conf.String("etcd_product_key")
	secProxyContext.EtcdTimeout, err = conf.Int("etcd_timeout")
	if err != nil {
		return err
	}

	secProxyContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	if err != nil {
		return err
	}
	secProxyContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")
	if err != nil {
		return err
	}

	secProxyContext.RequestTimeout, err = conf.Int("request_timeout")
	if err != nil {
		return err
	}
	secProxyContext.RequestChanTimeout, err = conf.Int("request_chan_timeout")
	if err != nil {
		return err
	}
	secProxyContext.ResponseChanTimeout, err = conf.Int("response_chan_timeout")
	if err != nil {
		return err
	}

	secProxyContext.RequestChanSize, err = conf.Int("request_chan_size")
	if err != nil {
		return err
	}
	secProxyContext.ResponseChanSize, err = conf.Int("response_chan_size")
	if err != nil {
		return err
	}

	secProxyContext.UserMinLimit, err = conf.Int("user_min_limit")
	if err != nil {
		return err
	}
	secProxyContext.UserSecLimit, err = conf.Int("user_sec_limit")
	if err != nil {
		return err
	}
	secProxyContext.IPMinLimit, err = conf.Int("ip_min_limit")
	if err != nil {
		return err
	}
	secProxyContext.IPSecLimit, err = conf.Int("ip_sec_limit")
	if err != nil {
		return err
	}

	return nil
}

func init() {
	secProxyContext = &SecProxyContext{}

	err := initSecProxyConf()
	if err != nil {
		beego.Error(err)
		return
	}

	initRedis()

	err = initEtcd(secProxyContext.EtcdConf)
	if err != nil {
		beego.Error(err)
		return
	}

	err = loadProduct()
	if err != nil {
		beego.Error(err)
		return
	}

	secProxyContext.RequestChan = make(chan *SecRequest, secProxyContext.RequestChanSize)
	secProxyContext.ResponseChanMgr = newResponseChanMgr()

	// secProxyContext.ProductMgr = newProductMgr()
	secProxyContext.UserLimitMgr = newUserLimitMgr()

	run()
}
