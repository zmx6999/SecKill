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

func initEtcd(etcdConf EtcdConf) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdConf.EtcdAddr},
		DialTimeout: time.Second*time.Duration(etcdConf.EtcdTimeout),
	})
	if err != nil {
		return err
	}
	secLayerContext.EtcdClient = cli

	return nil
}

func initSecLayerConf() error {
	conf, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		return err
	}

	secLayerContext.Proxy2LayerRedisConf.RedisAddr = conf.String("proxy2layer_redis_addr")
	secLayerContext.Proxy2LayerRedisConf.RedisPassword = conf.String("proxy2layer_redis_password")
	secLayerContext.Proxy2LayerRedisConf.RedisQueueName = conf.String("proxy2layer_redis_queue_name")
	secLayerContext.Proxy2LayerRedisConf.RedisMaxIdle, err = conf.Int("proxy2layer_redis_max_idle")
	if err != nil {
		return err
	}
	secLayerContext.Proxy2LayerRedisConf.RedisMaxActive, err = conf.Int("proxy2layer_redis_max_active")
	if err != nil {
		return err
	}
	secLayerContext.Proxy2LayerRedisConf.RedisIdleTimeout, err = conf.Int("proxy2layer_redis_idle_timeout")
	if err != nil {
		return err
	}

	secLayerContext.Layer2ProxyRedisConf.RedisAddr = conf.String("layer2proxy_redis_addr")
	secLayerContext.Layer2ProxyRedisConf.RedisPassword = conf.String("layer2proxy_redis_password")
	secLayerContext.Layer2ProxyRedisConf.RedisQueueName = conf.String("layer2proxy_redis_queue_name")
	secLayerContext.Layer2ProxyRedisConf.RedisMaxIdle, err = conf.Int("layer2proxy_redis_max_idle")
	if err != nil {
		return err
	}
	secLayerContext.Layer2ProxyRedisConf.RedisMaxActive, err = conf.Int("layer2proxy_redis_max_active")
	if err != nil {
		return err
	}
	secLayerContext.Layer2ProxyRedisConf.RedisIdleTimeout, err = conf.Int("layer2proxy_redis_idle_timeout")
	if err != nil {
		return err
	}

	secLayerContext.EtcdAddr = conf.String("etcd_addr")
	secLayerContext.EtcdProductKey = conf.String("etcd_product_key")
	secLayerContext.EtcdTimeout, err = conf.Int("etcd_timeout")
	if err != nil {
		return err
	}

	secLayerContext.ReadGoroutineNum, err = conf.Int("read_goroutine_num")
	if err != nil {
		return err
	}
	secLayerContext.WriteGoroutineNum, err = conf.Int("write_goroutine_num")
	if err != nil {
		return err
	}
	secLayerContext.HandleUserGoroutineNum, err = conf.Int("handle_user_goroutine_num")
	if err != nil {
		return err
	}

	secLayerContext.RequestTimeout, err = conf.Int("request_timeout")
	if err != nil {
		return err
	}
	secLayerContext.Read2HandleChanTimeout, err = conf.Int("read2handle_chan_timeout")
	if err != nil {
		return err
	}
	secLayerContext.Handle2WriteChanTimeout, err = conf.Int("handle2write_chan_timeout")
	if err != nil {
		return err
	}

	secLayerContext.Read2HandleChanSize, err = conf.Int("read2handle_chan_size")
	if err != nil {
		return err
	}
	secLayerContext.Handle2WriteChanSize, err = conf.Int("handle2write_chan_size")
	if err != nil {
		return err
	}

	secLayerContext.ProductSecSoldLimit, err = conf.Int("product_sec_sold_limit")
	if err != nil {
		return err
	}
	secLayerContext.ProductOnePersonBuyLimit, err = conf.Int("product_one_person_buy_limit")
	if err != nil {
		return err
	}
	secLayerContext.ProductSoldRate, err = conf.Float("product_sold_rate")
	if err != nil {
		return err
	}

	secLayerContext.LayerSecret = conf.String("layer_secret")

	return nil
}

func init() {
	secLayerContext = &SecLayerContext{}

	err := initSecLayerConf()
	if err != nil {
		beego.Error(err)
		return
	}

	initRedis()

	err = initEtcd(secLayerContext.EtcdConf)
	if err != nil {
		beego.Error(err)
		return
	}

	err = loadProduct()
	if err != nil {
		beego.Error(err)
		return
	}

	secLayerContext.Read2HandleChan = make(chan *SecRequest, secLayerContext.Read2HandleChanSize)
	secLayerContext.Handle2WriteChan = make(chan *SecResponse, secLayerContext.Handle2WriteChanSize)

	// secLayerContext.ProductMgr = newProductMgr()
	secLayerContext.UserHistoryMgr = newUserHistoryMgr()

	run()
}
