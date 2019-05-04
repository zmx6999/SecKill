package service

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
)

var (
	secProxyContext *SecProxyContext
)

func initEtcd() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.EtcdConf.Timeout),
	})
	if err != nil {
		return
	}
	secProxyContext.EtcdClient = cli
	return
}

func initSecProxyConf() (err error) {
	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		return
	}

	secProxyContext.Proxy2LayerConf.RedisAddr = conf.String("proxy2layer_redis_addr")
	secProxyContext.Proxy2LayerConf.Password = conf.String("proxy2layer_redis_password")
	secProxyContext.Proxy2LayerConf.MaxActive, err = conf.Int("proxy2layer_redis_max_active")
	if err != nil {
		return
	}
	secProxyContext.Proxy2LayerConf.MaxIdle, err = conf.Int("proxy2layer_redis_max_idle")
	if err != nil {
		return
	}
	secProxyContext.Proxy2LayerConf.IdleTimeout, err = conf.Int("proxy2layer_redis_idle_timeout")
	if err != nil {
		return
	}
	secProxyContext.Proxy2LayerConf.QueueName = conf.String("proxy2layer_redis_queue_name")

	secProxyContext.Layer2ProxyConf.RedisAddr = conf.String("layer2proxy_redis_addr")
	secProxyContext.Layer2ProxyConf.Password = conf.String("layer2proxy_redis_password")
	secProxyContext.Layer2ProxyConf.MaxActive, err = conf.Int("layer2proxy_redis_max_active")
	if err != nil {
		return
	}
	secProxyContext.Layer2ProxyConf.MaxIdle, err = conf.Int("layer2proxy_redis_max_idle")
	if err != nil {
		return
	}
	secProxyContext.Layer2ProxyConf.IdleTimeout, err = conf.Int("layer2proxy_redis_idle_timeout")
	if err != nil {
		return
	}
	secProxyContext.Layer2ProxyConf.QueueName = conf.String("layer2proxy_redis_queue_name")

	secProxyContext.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secProxyContext.EtcdConf.Timeout, err = conf.Int("etcd_timeout")
	if err != nil {
		return
	}
	secProxyContext.EtcdConf.ProductKey = conf.String("etcd_product_key")

	secProxyContext.AccessLimitConf.UserSecAccessLimit, err = conf.Int("user_sec_access_limit")
	if err != nil {
		return
	}
	secProxyContext.AccessLimitConf.UserMinAccessLimit, err = conf.Int("user_min_access_limit")
	if err != nil {
		return
	}
	secProxyContext.AccessLimitConf.IPSecAccessLimit, err = conf.Int("ip_sec_access_limit")
	if err != nil {
		return
	}
	secProxyContext.AccessLimitConf.IPMinAccessLimit, err = conf.Int("ip_min_access_limit")
	if err != nil {
		return
	}

	secProxyContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	if err != nil {
		return
	}
	secProxyContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")
	if err != nil {
		return
	}

	secProxyContext.RequestTimeout, err = conf.Int("request_timeout")
	if err != nil {
		return
	}
	secProxyContext.RequestChanTimeout, err = conf.Int("request_chan_timeout")
	if err != nil {
		return
	}
	secProxyContext.ResponseChanTimeout, err = conf.Int("response_chan_timeout")
	if err != nil {
		return
	}

	secProxyContext.RequestChanSize, err = conf.Int("request_chan_size")
	if err != nil {
		return
	}
	secProxyContext.ResponseChanSize, err = conf.Int("response_chan_size")
	if err != nil {
		return
	}

	return
}

func initSecProxyContext() (err error) {
	err = initSecProxyConf()
	if err != nil {
		return
	}

	initRedis()

	err = initEtcd()
	if err != nil {
		return
	}

	secProxyContext.AccessMgr = NewAccessMgr()
	secProxyContext.ProductMgr = NewProductMgr()

	secProxyContext.RequestChan = make(chan *Request, secProxyContext.RequestChanSize)
	secProxyContext.ResponseMgr = NewResponseMgr()

	return
}

func init()  {
	secProxyContext = &SecProxyContext{}

	err := initSecProxyContext()
	if err != nil {
		beego.Error(err)
		return
	}

	err = loadProduct()
	if err != nil {
		beego.Error(err)
		return
	}

	run()
}
