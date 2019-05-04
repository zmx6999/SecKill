package service

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
)

var (
	secLayerContext *SecLayerContext
)

func initEtcd() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secLayerContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secLayerContext.EtcdConf.Timeout),
	})
	if err != nil {
		return
	}
	secLayerContext.EtcdClient = cli
	return
}

func initSecLayerConf() (err error) {
	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		return
	}

	secLayerContext.Proxy2LayerConf.RedisAddr = conf.String("proxy2layer_redis_addr")
	secLayerContext.Proxy2LayerConf.Password = conf.String("proxy2layer_redis_password")
	secLayerContext.Proxy2LayerConf.MaxActive, err = conf.Int("proxy2layer_redis_max_active")
	if err != nil {
		return
	}
	secLayerContext.Proxy2LayerConf.MaxIdle, err = conf.Int("proxy2layer_redis_max_idle")
	if err != nil {
		return
	}
	secLayerContext.Proxy2LayerConf.IdleTimeout, err = conf.Int("proxy2layer_redis_idle_timeout")
	if err != nil {
		return
	}
	secLayerContext.Proxy2LayerConf.QueueName = conf.String("proxy2layer_redis_queue_name")

	secLayerContext.Layer2ProxyConf.RedisAddr = conf.String("layer2proxy_redis_addr")
	secLayerContext.Layer2ProxyConf.Password = conf.String("layer2proxy_redis_password")
	secLayerContext.Layer2ProxyConf.MaxActive, err = conf.Int("layer2proxy_redis_max_active")
	if err != nil {
		return
	}
	secLayerContext.Layer2ProxyConf.MaxIdle, err = conf.Int("layer2proxy_redis_max_idle")
	if err != nil {
		return
	}
	secLayerContext.Layer2ProxyConf.IdleTimeout, err = conf.Int("layer2proxy_redis_idle_timeout")
	if err != nil {
		return
	}
	secLayerContext.Layer2ProxyConf.QueueName = conf.String("layer2proxy_redis_queue_name")

	secLayerContext.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secLayerContext.EtcdConf.Timeout, err = conf.Int("etcd_timeout")
	if err != nil {
		return
	}
	secLayerContext.EtcdConf.ProductKey = conf.String("etcd_product_key")

	secLayerContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	if err != nil {
		return
	}
	secLayerContext.HandleGoroutineNum, err = conf.Int("handle_goroutine_num")
	if err != nil {
		return
	}
	secLayerContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")
	if err != nil {
		return
	}

	secLayerContext.RequestTimeout, err = conf.Int("request_timeout")
	if err != nil {
		return
	}
	secLayerContext.RequestChanTimeout, err = conf.Int("request_chan_timeout")
	if err != nil {
		return
	}
	secLayerContext.ResponseChanTimeout, err = conf.Int("response_chan_timeout")
	if err != nil {
		return
	}

	secLayerContext.RequestChanSize, err = conf.Int("request_chan_size")
	if err != nil {
		return
	}
	secLayerContext.ResponseChanSize, err = conf.Int("response_chan_size")
	if err != nil {
		return
	}

	secLayerContext.Secret = conf.String("secret")

	return
}

func initSecLayerContext() (err error) {
	err = initSecLayerConf()
	if err != nil {
		return
	}

	initRedis()

	err = initEtcd()
	if err != nil {
		return
	}

	secLayerContext.ProductMgr = NewProductMgr()
	secLayerContext.HistoryMgr = NewHistoryMgr()

	secLayerContext.RequestChan = make(chan *Request, secLayerContext.RequestChanSize)
	secLayerContext.ResponseChan = make(chan *Response, secLayerContext.ResponseChanSize)

	return
}

func init()  {
	secLayerContext = &SecLayerContext{}

	err := initSecLayerContext()
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
