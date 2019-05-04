package service

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/astaxie/beego"
	"encoding/json"
	"fmt"
)

type RedisConf struct {
	RedisAddr string
	Password string
	MaxActive int
	MaxIdle int
	IdleTimeout int
	QueueName string
}

func initRedisPool(redisConf RedisConf) *redis.Pool {
	return &redis.Pool{
		MaxActive: redisConf.MaxActive,
		MaxIdle: redisConf.MaxIdle,
		IdleTimeout: time.Second*time.Duration(redisConf.IdleTimeout),
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConf.RedisAddr, redis.DialPassword(redisConf.Password))
		},
	}
}

func initRedis()  {
	secProxyContext.Proxy2LayerPool = initRedisPool(secProxyContext.Proxy2LayerConf)
	secProxyContext.Layer2ProxyPool = initRedisPool(secProxyContext.Layer2ProxyConf)
}

func read()  {
	for  {
		cnn := secProxyContext.Layer2ProxyPool.Get()
		r, err := cnn.Do("blpop", secProxyContext.Layer2ProxyConf.QueueName, 0)
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

		cnn := secProxyContext.Proxy2LayerPool.Get()
		_, err = cnn.Do("rpush", secProxyContext.Proxy2LayerConf.QueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}

func run()  {
	for i := 1; i <= secProxyContext.ReadGoroutineNum; i++ {
		go read()
	}
	for i := 1; i <= secProxyContext.WriteGoroutineNum; i++ {
		go write()
	}
}
