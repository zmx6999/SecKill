package service

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/astaxie/beego"
	"encoding/json"
	"math/rand"
	"crypto/sha256"
	"fmt"
	"encoding/hex"
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
	secLayerContext.Proxy2LayerRedisPool = initRedisPool(secLayerContext.Proxy2LayerRedisConf)
	secLayerContext.Layer2ProxyRedisPool = initRedisPool(secLayerContext.Layer2ProxyRedisConf)
}

func run()  {
	for i := 0; i < secLayerContext.ReadGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleRead()
	}
	for i := 0; i < secLayerContext.WriteGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleWrite()
	}
	for i := 0; i < secLayerContext.HandleUserGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleUser()
	}
	secLayerContext.WaitGroup.Wait()
}

func handleRead()  {
	for  {
		cnn := secLayerContext.Proxy2LayerRedisPool.Get()
		r, err := cnn.Do("blpop", secLayerContext.Proxy2LayerRedisConf.RedisQueueName, 0)
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

		var req SecRequest
		err = json.Unmarshal(data, &req)
		if err != nil {
			beego.Error(err)
			continue
		}

		now := time.Now()
		if now.Sub(req.AccessTime).Seconds() > float64(secLayerContext.RequestTimeout) {
			continue
		}

		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.Read2HandleChanTimeout))
		select {
		case secLayerContext.Read2HandleChan <- &req:
			timer.Stop()
		case <- timer.C:
			timer.Stop()
		}
	}
}

func handleWrite()  {
	for rsp := range secLayerContext.Handle2WriteChan{
		data, err := json.Marshal(rsp)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secLayerContext.Layer2ProxyRedisPool.Get()
		_, err = cnn.Do("rpush", secLayerContext.Layer2ProxyRedisConf.RedisQueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}

func handleUser()  {
	for req := range secLayerContext.Read2HandleChan{
		rsp := secKill(req)

		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.Handle2WriteChanTimeout))
		select {
		case secLayerContext.Handle2WriteChan <- rsp:
			timer.Stop()
		case <- timer.C:
			timer.Stop()
		}
	}
}

func secKill(req *SecRequest) (rsp *SecResponse) {
	rsp = &SecResponse{}
	rsp.UserId = req.UserId
	rsp.ProductId = req.ProductId
	rsp.Nonce = req.Nonce

	secLayerContext.ProductLock.Lock()
	defer secLayerContext.ProductLock.Unlock()

	rate := rand.Float64()
	if rate < secLayerContext.ProductSoldRate {
		rsp.Code = NetworkBusyErr
		return
	}

	now := time.Now()
	if now.Sub(req.AccessTime).Seconds() > float64(secLayerContext.RequestTimeout) {
		rsp.Code = TimeoutErr
		return
	}

	product, ok := secLayerContext.ProductMap[req.ProductId]
	if !ok {
		rsp.Code = ProductNotFoundErr
		return
	}

	if product.Status == ProductStatusSoldOut {
		rsp.Code = ProductSoldOutErr
		return
	}

	if product.Sold >= product.Total {
		product.Status = ProductStatusSoldOut
		rsp.Code = ProductSoldOutErr
		return
	}

	secSoldNum := product.SecLimit.Check()
	if secSoldNum >= secLayerContext.ProductSecSoldLimit {
		rsp.Code = NetworkBusyErr
		return
	}

	secLayerContext.UserHistoryLock.Lock()
	defer secLayerContext.UserHistoryLock.Unlock()

	userHistory, ok := secLayerContext.UserHistoryMap[req.UserId]
	if !ok {
		userHistory = newUserHistory()
		secLayerContext.UserHistoryMap[req.UserId] = userHistory
	}

	userBuyNum := userHistory.Check(req.ProductId)
	if userBuyNum >= secLayerContext.ProductOnePersonBuyLimit {
		rsp.Code = AlreadyBuyErr
		return
	}

	product.Sold++
	product.SecLimit.Add()
	userHistory.Add(req.ProductId)

	rsp.Code = SecKillSuccess
	rsp.TokenTime = now

	str := fmt.Sprintf("user_id=%s&product_id=%s&timestamp=%d&nonce=%s&secret=%s", rsp.UserId, rsp.ProductId, rsp.TokenTime, rsp.Nonce, secLayerContext.LayerSecret)
	hbyte := sha256.Sum256([]byte(str))
	token := hex.EncodeToString(hbyte[:])
	rsp.Token = token

	return
}
