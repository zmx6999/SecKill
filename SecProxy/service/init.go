package service

import (
	"time"
	"go.etcd.io/etcd/clientv3"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
)

var (
	secProxyContext *SecProxyContext
)

func initEtcd() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.DialTimeout),
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

	secProxyContext.ProxyToLayerConf.RedisAddr = conf.String("proxy_layer_redis_addr")
	secProxyContext.ProxyToLayerConf.RedisPassword = conf.String("proxy_layer_redis_password")
	secProxyContext.ProxyToLayerConf.MaxIdle, err = conf.Int("proxy_layer_redis_max_idle")
	secProxyContext.ProxyToLayerConf.MaxActive, err = conf.Int("proxy_layer_redis_max_active")
	secProxyContext.ProxyToLayerConf.IdleTimeout, err = conf.Int("proxy_layer_redis_idle_timeout")
	secProxyContext.ProxyToLayerConf.QueueName = conf.String("proxy_layer_redis_queue_name")

	secProxyContext.LayerToProxyConf.RedisAddr = conf.String("layer_proxy_redis_addr")
	secProxyContext.LayerToProxyConf.RedisPassword = conf.String("layer_proxy_redis_password")
	secProxyContext.LayerToProxyConf.MaxIdle, err = conf.Int("layer_proxy_redis_max_idle")
	secProxyContext.LayerToProxyConf.MaxActive, err = conf.Int("layer_proxy_redis_max_active")
	secProxyContext.LayerToProxyConf.IdleTimeout, err = conf.Int("layer_proxy_redis_idle_timeout")
	secProxyContext.LayerToProxyConf.QueueName = conf.String("layer_proxy_redis_queue_name")

	secProxyContext.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secProxyContext.EtcdConf.DialTimeout, err = conf.Int("etcd_dial_timeout")
	secProxyContext.EtcdConf.ProductKey = conf.String("etcd_product_key")

	secProxyContext.UserSecLimit, err = conf.Int("user_sec_limit")
	secProxyContext.UserMinLimit, err = conf.Int("user_min_limit")
	secProxyContext.IPSecLimit, err = conf.Int("ip_sec_limit")
	secProxyContext.IPMinLimit, err = conf.Int("ip_min_limit")

	secProxyContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	secProxyContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")

	secProxyContext.RequestTimeout, err = conf.Int("request_timeout")
	secProxyContext.RequestChanTimeout, err = conf.Int("request_chan_timeout")
	secProxyContext.ResponseChanTimeout, err = conf.Int("response_chan_timeout")

	secProxyContext.RequestChanSize, err = conf.Int("request_chan_size")
	secProxyContext.ResponseChanSize, err = conf.Int("response_chan_size")

	return
}

func init()  {
	secProxyContext = &SecProxyContext{}

	err := initSecProxyConf()
	if err != nil {
		beego.Error(err)
		return
	}

	initRedis()

	err = initEtcd()
	if err != nil {
		beego.Error(err)
		return
	}

	secProxyContext.ProductMgr = NewProductMgr()
	err = loadProduct()
	if err != nil {
		beego.Error(err)
		return
	}

	secProxyContext.LimitMgr = NewLimitMgr()

	secProxyContext.RequestChan = make(chan *Request, secProxyContext.RequestChanSize)
	secProxyContext.ResponseMgr = NewResponseMgr()

	run()
}
