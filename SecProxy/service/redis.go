package service

import (
	"time"
	"github.com/garyburd/redigo/redis"
	"encoding/json"
	"github.com/astaxie/beego"
	"fmt"
)

type RedisConf struct {
	RedisAddr string
	RedisPassword string
	MaxIdle int
	MaxActive int
	IdleTimeout int
	QueueName string
}

func initRedisPool(redisConf RedisConf) *redis.Pool {
	return &redis.Pool{
		MaxIdle: redisConf.MaxIdle,
		MaxActive: redisConf.MaxActive,
		IdleTimeout: time.Second*time.Duration(redisConf.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConf.RedisAddr, redis.DialPassword(redisConf.RedisPassword))
		},
	}
}

func initRedis()  {
	secProxyContext.ProxyToLayerPool = initRedisPool(secProxyContext.ProxyToLayerConf)
	secProxyContext.LayerToProxyPool = initRedisPool(secProxyContext.LayerToProxyConf)
}

func run()  {
	for i := 1; i <= secProxyContext.ReadGoroutineNum; i++ {
		go read()
	}
	for i := 1; i <= secProxyContext.WriteGoroutineNum; i++ {
		go write()
	}
}

func read()  {
	for  {
		cnn := secProxyContext.LayerToProxyPool.Get()
		r, err := cnn.Do("blpop", secProxyContext.LayerToProxyConf.QueueName, 0)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}

		t, ok := r.([]interface{})
		if !ok || len(t) < 2 {
			continue
		}

		data, ok := t[1].([]byte)
		if !ok {
			continue
		}

		var res Response
		err = json.Unmarshal(data, &res)
		if err != nil {
			beego.Error(err)
			continue
		}

		key := fmt.Sprintf("%s_%s", res.ProductId, res.UserId)
		secProxyContext.ResponseLock.RLock()
		responseChan, ok := secProxyContext.ResponseMap[key]
		secProxyContext.ResponseLock.RUnlock()
		if !ok {
			continue
		}

		timer := time.NewTicker(time.Second*time.Duration(secProxyContext.ResponseChanTimeout))
		select {
		case <-timer.C:
		case responseChan<-&res:
		}
		timer.Stop()
	}
}

func write()  {
	for req := range secProxyContext.RequestChan{
		data, err := json.Marshal(req)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secProxyContext.ProxyToLayerPool.Get()
		_, err = cnn.Do("rpush", secProxyContext.ProxyToLayerConf.QueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}
