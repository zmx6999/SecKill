package service

import (
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/astaxie/beego"
)

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

func initRedis(secProxyContext *SecProxyContext)  {
	secProxyContext.Proxy2LayerRedisPool = initRedisPool(secProxyContext.SecProxyConf.RedisProxy2LayerConf)
	secProxyContext.Layer2ProxyRedisPool = initRedisPool(secProxyContext.SecProxyConf.RedisLayer2ProxyConf)
}

func run(secProxyContext *SecProxyContext)  {
	for i := 0; i < secProxyContext.ReadGoroutineNum; i++ {
		go readHandle(secProxyContext)
	}
	for i := 0; i < secProxyContext.WriteGoroutineNum; i++ {
		go writeHandle(secProxyContext)
	}
}

func writeHandle(secProxyContext *SecProxyContext)  {
	for  {
		req := <-secProxyContext.RequestChan
		beego.Info(req)
		data, err := json.Marshal(req)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secProxyContext.Proxy2LayerRedisPool.Get()
		_, err = cnn.Do("rpush", secProxyContext.SecProxyConf.RedisProxy2LayerConf.RedisQueueName, string(data))
		if err != nil {
			beego.Error(err)
			cnn.Close()
			continue
		}

		/*
		r, err := redis.Strings(cnn.Do("lrange", secProxyContext.SecProxyConf.RedisProxy2LayerConf.RedisQueueName, 0, -1))
		if err != nil {
			fmt.Print(err)
			continue
		}
		beego.Info(r)
		*/

		cnn.Close()
	}
}

func readHandle(secProxyContext *SecProxyContext)  {
	for  {
		cnn := secProxyContext.Layer2ProxyRedisPool.Get()
		r, err := cnn.Do("blpop", secProxyContext.SecProxyConf.RedisLayer2ProxyConf.RedisQueueName, 0)
		if err != nil {
			beego.Error(time.Now().String(), "redis.go:63 ", err)
			cnn.Close()
			continue
		}
		cnn.Close()

		beego.Info(secProxyContext.SecProxyConf.RedisLayer2ProxyConf.RedisQueueName)

		t, ok := r.([]interface{})
		if !ok || len(t) < 2 {
			continue
		}

		beego.Info(t)

		data, ok := t[1].([]byte)
		if !ok {
			continue
		}

		beego.Info(string(data))

		var res SecResponse
		err = json.Unmarshal(data, &res)
		if err != nil {
			beego.Error(time.Now().String(), "redis.go:72 ", err)
			continue
		}

		beego.Info(time.Now().String(), "redis.go:76 ", res)

		secProxyContext.ResponseChanMapLock.RLock()
		key := res.UserId+"_"+res.ProductId
		responseChan, ok := secProxyContext.ResponseChanMap[key]
		secProxyContext.ResponseChanMapLock.RUnlock()
		if !ok {
			continue
		}
		responseChan.Chan <- &res
	}
}
