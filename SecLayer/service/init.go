package service

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego"
)

func initEtcd(secLayerContext *SecLayerContext)  error {
	secLayerConf := secLayerContext.SecLayerConf
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secLayerConf.EtcdConf.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secLayerConf.EtcdConf.EtcdTimeout),
	})
	if err != nil {
		return err
	}

	secLayerContext.EtcdClient = cli
	return nil
}

func initSecLayerConf(secLayerContext *SecLayerContext) error {
	conf, err := config.NewConfig("ini", "./conf/app.conf")
	if err != nil {
		return err
	}

	secLayerConf := SecLayerConf{}
	secLayerConf.Proxy2LayerRedis.RedisAddr = conf.String("proxy2layer_redis_addr")
	secLayerConf.Proxy2LayerRedis.RedisMaxIdle, _ = conf.Int("proxy2layer_redis_max_idle")
	secLayerConf.Proxy2LayerRedis.RedisMaxActive, _ = conf.Int("proxy2layer_redis_max_active")
	secLayerConf.Proxy2LayerRedis.RedisIdleTimeout, _ = conf.Int("proxy2layer_redis_idle_timeout")
	secLayerConf.Proxy2LayerRedis.RedisQueueName = conf.String("proxy2layer_redis_queue_name")
	secLayerConf.Proxy2LayerRedis.RedisPassword = conf.String("proxy2layer_redis_password")

	secLayerConf.Layer2ProxyRedis.RedisAddr = conf.String("layer2proxy_redis_addr")
	secLayerConf.Layer2ProxyRedis.RedisMaxIdle, _ = conf.Int("layer2proxy_redis_max_idle")
	secLayerConf.Layer2ProxyRedis.RedisMaxActive, _ = conf.Int("layer2proxy_redis_max_active")
	secLayerConf.Layer2ProxyRedis.RedisIdleTimeout, _ = conf.Int("layer2proxy_redis_idle_timeout")
	secLayerConf.Layer2ProxyRedis.RedisQueueName = conf.String("layer2proxy_redis_queue_name")
	secLayerConf.Layer2ProxyRedis.RedisPassword = conf.String("layer2proxy_redis_password")

	secLayerConf.EtcdConf.EtcdAddr = conf.String("etcd_addr")
	secLayerConf.EtcdConf.EtcdTimeout, _ = conf.Int("etcd_timeout")
	secLayerConf.EtcdConf.EtcdProductKey = conf.String("etcd_product_key")

	secLayerConf.ReadGoroutineNum, _ = conf.Int("read_goroutine_num")
	secLayerConf.WriteGoroutineNum, _ = conf.Int("write_goroutine_num")
	secLayerConf.HandleUserGoroutineNum, _ = conf.Int("handle_user_goroutine_num")
	secLayerConf.Read2HandleChanSize, _ = conf.Int("read2handle_channel_size")
	secLayerConf.Handle2WriteChanSize, _ = conf.Int("handle2write_channel_size")
	secLayerConf.MaxRequestWaitTimeout, _ = conf.Int("max_request_wait_timeout")

	secLayerConf.Send2HandleChanTimeout, _ = conf.Int("send2handle_chan_timeout")
	secLayerConf.Send2WriteChanTimeout, _ = conf.Int("send2write_chan_timeout")

	secLayerConf.LimitPeriod, _ = conf.Int("limit_period")
	secLayerConf.TokenPassword = conf.String("token_password")

	secLayerConf.ProductOnePersonBuyLimit, _ = conf.Int("product_one_person_buy_limit")
	secLayerConf.ProductSecSoldMaxLimit, _ = conf.Int("product_sec_sold_max_limit")
	secLayerConf.ProductBuyRate, _ = conf.Float("product_buy_rate")

	secLayerContext.SecLayerConf = secLayerConf

	return nil
}

func initSecLayerContext(secLayerContext *SecLayerContext) error {
	err := initSecLayerConf(secLayerContext)
	if err != nil {
		return err
	}

	initRedis(secLayerContext)

	err = initEtcd(secLayerContext)
	if err != nil {
		return err
	}

	err = loadProduct(secLayerContext)
	if err != nil {
		return err
	}

	secLayerContext.Read2HandleChan = make(chan *SecRequest, secLayerContext.SecLayerConf.Read2HandleChanSize)
	secLayerContext.Handle2WriteChan = make(chan *SecResponse, secLayerContext.SecLayerConf.Handle2WriteChanSize)
	secLayerContext.UserHistoryMap = map[string]*UserHistory{}

	return nil
}

func init()  {
	secLayerContext := &SecLayerContext{}
	err := initSecLayerContext(secLayerContext)
	if err != nil {
		beego.Error(err)
		return
	}

	run(secLayerContext)
}
