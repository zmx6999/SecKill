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
		DialTimeout: time.Second*time.Duration(secLayerContext.DialTimeout),
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

	secLayerContext.ProxyToLayerConf.RedisAddr = conf.String("proxy_layer_redis_addr")
	secLayerContext.ProxyToLayerConf.RedisPassword = conf.String("proxy_layer_redis_password")
	secLayerContext.ProxyToLayerConf.MaxIdle, err = conf.Int("proxy_layer_redis_max_idle")
	secLayerContext.ProxyToLayerConf.MaxActive, err = conf.Int("proxy_layer_redis_max_active")
	secLayerContext.ProxyToLayerConf.IdleTimeout, err = conf.Int("proxy_layer_redis_idle_timeout")
	secLayerContext.ProxyToLayerConf.QueueName = conf.String("proxy_layer_redis_queue_name")

	secLayerContext.LayerToProxyConf.RedisAddr = conf.String("layer_proxy_redis_addr")
	secLayerContext.LayerToProxyConf.RedisPassword = conf.String("layer_proxy_redis_password")
	secLayerContext.LayerToProxyConf.MaxIdle, err = conf.Int("layer_proxy_redis_max_idle")
	secLayerContext.LayerToProxyConf.MaxActive, err = conf.Int("layer_proxy_redis_max_active")
	secLayerContext.LayerToProxyConf.IdleTimeout, err = conf.Int("layer_proxy_redis_idle_timeout")
	secLayerContext.LayerToProxyConf.QueueName = conf.String("layer_proxy_redis_queue_name")

	secLayerContext.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secLayerContext.EtcdConf.DialTimeout, err = conf.Int("etcd_dial_timeout")
	secLayerContext.EtcdConf.ProductKey = conf.String("etcd_product_key")

	secLayerContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	secLayerContext.HandleUserGoroutineNum, err = conf.Int("handle_user_goroutine_num")
	secLayerContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")

	secLayerContext.RequestTimeout, err = conf.Int("request_timeout")
	secLayerContext.ReadToHandleChanTimeout, err = conf.Int("read_handle_chan_timeout")
	secLayerContext.HandleToWriteChanTimeout, err = conf.Int("handle_write_chan_timeout")

	secLayerContext.ReadToHandleChanSize, err = conf.Int("read_handle_chan_size")
	secLayerContext.HandleToWriteChanSize, err = conf.Int("handle_write_chan_size")

	secLayerContext.LayerSecret = conf.String("layer_secret")

	return
}

func init()  {
	secLayerContext = &SecLayerContext{}

	err := initSecLayerConf()
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

	secLayerContext.ProductMgr = NewProductMgr()
	err = loadProduct()
	if err != nil {
		beego.Error(err)
		return
	}

	secLayerContext.ReadToHandleChan = make(chan *Request, secLayerContext.ReadToHandleChanSize)
	secLayerContext.HandleToWriteChan = make(chan *Response, secLayerContext.HandleToWriteChanSize)

	secLayerContext.UserHistoryMgr = NewUserHistoryMgr()

	run()
}
