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
	RedisMaxIdle int
	RedisMaxActive int
	RedisIdleTimeout int
	RedisQueueName string
}

func initRedisPool(redisConf RedisConf) *redis.Pool {
	return &redis.Pool{
		MaxIdle: redisConf.RedisMaxIdle,
		MaxActive: redisConf.RedisMaxActive,
		IdleTimeout: time.Second*time.Duration(redisConf.RedisIdleTimeout),
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisConf.RedisAddr, redis.DialPassword(redisConf.RedisPassword))
		},
	}
}

func initRedis()  {
	secProxyContext.Proxy2LayerRedisPool = initRedisPool(secProxyContext.Proxy2LayerRedisConf)
	secProxyContext.Layer2ProxyRedisPool = initRedisPool(secProxyContext.Layer2ProxyRedisConf)
}

func run()  {
	for i := 0; i < secProxyContext.ReadGoroutineNum; i++ {
		go handleRead()
	}
	for i := 0; i < secProxyContext.WriteGoroutineNum; i++ {
		go handleWrite()
	}
}

func handleRead()  {
	for req := range secProxyContext.RequestChan{
		data, err := json.Marshal(req)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secProxyContext.Proxy2LayerRedisPool.Get()
		_, err = cnn.Do("rpush", secProxyContext.Proxy2LayerRedisConf.RedisQueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}

func handleWrite()  {
	for  {
		cnn := secProxyContext.Layer2ProxyRedisPool.Get()
		r, err := cnn.Do("blpop", secProxyContext.Layer2ProxyRedisConf.RedisQueueName, 0)
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

		var res SecResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			beego.Error(err)
			continue
		}

		secProxyContext.ResponseChanMgr.Lock.RLock()

		key := fmt.Sprintf("%s_%s", res.UserId, res.ProductId)
		responseChan, ok := secProxyContext.ResponseChanMap[key]
		if !ok {
			secProxyContext.ResponseChanMgr.Lock.RUnlock()
			continue
		}

		secProxyContext.ResponseChanMgr.Lock.RUnlock()

		timer := time.NewTicker(time.Second*time.Duration(secProxyContext.ResponseChanTimeout))
		select {
		case responseChan <- &res:
			timer.Stop()
		case <- timer.C:
			timer.Stop()
		}
	}
}
