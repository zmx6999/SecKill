package service

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"encoding/json"
	"math/rand"
			"fmt"
	"crypto/sha256"
	"github.com/astaxie/beego/logs"
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

func initRedis(secLayerContext *SecLayerContext)  {
	secLayerContext.Proxy2LayerRedisPool = initRedisPool(secLayerContext.SecLayerConf.Proxy2LayerRedis)
	secLayerContext.Layer2ProxyRedisPool = initRedisPool(secLayerContext.SecLayerConf.Layer2ProxyRedis)
}

func run(secLayerContext *SecLayerContext)  {
	for i := 0; i < secLayerContext.SecLayerConf.ReadGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleReader(secLayerContext)
	}
	for i := 0; i < secLayerContext.SecLayerConf.WriteGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleWriter(secLayerContext)
	}
	for i := 0; i < secLayerContext.SecLayerConf.HandleUserGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleUser(secLayerContext)
	}
	secLayerContext.WaitGroup.Wait()
}

func handleReader(secLayerContext *SecLayerContext)  {
	for  {
		cnn := secLayerContext.Proxy2LayerRedisPool.Get()
		r, err := cnn.Do("blpop", secLayerContext.SecLayerConf.Proxy2LayerRedis.RedisQueueName, 0)
		if err != nil {
			fmt.Println(time.Now().String(), "redis.go:49 failed to connect to redis ", err)
			cnn.Close()
			continue
		}
		cnn.Close()

		t, ok := r.([]interface{})
		if !ok || len(t) < 2 {
			continue
		}

		data, ok := t[1].([]byte)
		if !ok {
			continue
		}

		var req SecRequest
		err = json.Unmarshal(data, &req)
		if err != nil {
			continue
		}

		fmt.Println(time.Now().String(), "redis.go:71 ", req)

		now := time.Now()
		if int(now.Sub(req.AccessTime).Seconds()) >= secLayerContext.SecLayerConf.MaxRequestWaitTimeout {
			continue
		}

		timer := time.NewTicker(time.Millisecond*time.Duration(secLayerContext.SecLayerConf.Send2HandleChanTimeout))
		select {
		case secLayerContext.Read2HandleChan <- &req:
		case <-timer.C:
			fmt.Println(time.Now().String(), "redis.go:82 timeout")
			continue
		}
	}
}

func handleWriter(secLayerContext *SecLayerContext)  {
	for res := range secLayerContext.Handle2WriteChan{
		err := writeToRedis(secLayerContext, res)
		if err != nil {
			continue
		}
	}
}

func writeToRedis(secLayerContext *SecLayerContext, res *SecResponse) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	fmt.Println(time.Now().String(), "redis.go:103 ", res)

	cnn := secLayerContext.Layer2ProxyRedisPool.Get()
	defer cnn.Close()
	_, err = cnn.Do("rpush", secLayerContext.SecLayerConf.Layer2ProxyRedis.RedisQueueName, data)
	if err != nil {
		fmt.Println(time.Now().String(), "redis.go:107 failed to connect to redis ", err)
		return err
	}

	return nil
}

func handleUser(secLayerContext *SecLayerContext)  {
	for req := range secLayerContext.Read2HandleChan{
		res := handleSecKill(secLayerContext, req)
		fmt.Println(time.Now().String(), "redis.go:119 ", res)

		timer := time.NewTicker(time.Millisecond*time.Duration(secLayerContext.SecLayerConf.Send2WriteChanTimeout))
		select {
		case secLayerContext.Handle2WriteChan <- res:
		case <-timer.C:
			fmt.Println(time.Now().String(), "redis.go:125 timeout")
			continue
		}
	}
}

func handleSecKill(secLayerContext *SecLayerContext, req *SecRequest) (res *SecResponse) {
	res = &SecResponse{}
	res.ProductId = req.ProductId
	res.UserId = req.UserId

	secLayerContext.ProductMapLock.Lock()
	defer secLayerContext.ProductMapLock.Unlock()

	now := time.Now()
	if int(now.Sub(req.AccessTime).Seconds()) >= secLayerContext.SecLayerConf.MaxRequestWaitTimeout {
		res.Code = ErrRetry
		return
	}

	product, ok := secLayerContext.ProductMap[req.ProductId]
	if !ok {
		res.Code = ErrNotFoundProduct
		return
	}
	if product.Status == ProductStatusSoldOut {
		res.Code = ErrSoldOut
		return
	}

	secLayerContext.UserHistoryLock.Lock()
	defer secLayerContext.UserHistoryLock.Unlock()

	userHistory, ok := secLayerContext.UserHistoryMgr.UserHistoryMap[req.UserId]
	if !ok {
		userHistory = NewUserHistory()
		secLayerContext.UserHistoryMgr.UserHistoryMap[req.UserId] = userHistory
	}

	rate := rand.Float64()
	if rate < secLayerContext.ProductBuyRate {
		res.Code = ErrRetry
		return
	}

	alreadySoldCount := product.Limit.Check(now, secLayerContext.LimitPeriod)
	if alreadySoldCount >= secLayerContext.ProductSecSoldMaxLimit {
		res.Code = ErrServiceBusy
		return
	}

	if product.Sold >= product.Total {
		product.Status = ProductStatusSoldOut
		res.Code = ErrSoldOut
		return
	}

	historyCount := userHistory.Get(req.ProductId)

	if historyCount >= secLayerContext.ProductOnePersonBuyLimit {
		res.Code = ErrAlreadyBuy
		return
	}

	logs.Info(product)

	product.Limit.Count(now, secLayerContext.LimitPeriod)
	product.Sold++
	userHistory.Add(req.ProductId, 1)

	res.Code = ErrSecKillSuccess
	res.Token = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("user_id=%s&product_id=%s&timestamp=%d&security=%s", res.UserId, res.ProductId, now.Unix(), secLayerContext.SecLayerConf.TokenPassword))))
	res.TokenTime = now

	return
}
