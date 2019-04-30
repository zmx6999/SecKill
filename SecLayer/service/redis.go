package service

import (
	"github.com/garyburd/redigo/redis"
	"time"
	"github.com/astaxie/beego"
	"encoding/json"
	"math/rand"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
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
	secLayerContext.ProxyToLayerPool = initRedisPool(secLayerContext.ProxyToLayerConf)
	secLayerContext.LayerToProxyPool = initRedisPool(secLayerContext.LayerToProxyConf)
}

func run()  {
	for i := 1; i <= secLayerContext.ReadGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go read()
	}
	for i := 1; i <= secLayerContext.HandleUserGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handleUser()
	}
	for i := 1; i <= secLayerContext.WriteGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go write()
	}
	secLayerContext.WaitGroup.Wait()
}

func read()  {
	for  {
		cnn := secLayerContext.ProxyToLayerPool.Get()
		r, err := cnn.Do("blpop", secLayerContext.ProxyToLayerConf.QueueName, 0)
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

		var req Request
		err = json.Unmarshal(data, &req)
		if err != nil {
			beego.Error(err)
			continue
		}

		now := time.Now()
		if now.Sub(req.AccessTime).Seconds() > float64(secLayerContext.RequestTimeout) {
			continue
		}

		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.ReadToHandleChanTimeout))
		select {
		case <-timer.C:
		case secLayerContext.ReadToHandleChan<-&req:
		}
		timer.Stop()
	}
}

func handleUser()  {
	for req := range secLayerContext.ReadToHandleChan{
		res := secKill(req)

		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.HandleToWriteChanTimeout))
		select {
		case <-timer.C:
		case secLayerContext.HandleToWriteChan<-res:
		}
		timer.Stop()
	}
}

func secKill(req *Request) (res *Response) {
	res = &Response{}
	res.ProductId = req.ProductId
	res.ProductNum = req.ProductNum
	res.UserId = req.UserId
	res.Nonce = req.Nonce

	secLayerContext.ProductLock.Lock()
	defer secLayerContext.ProductLock.Unlock()

	now := time.Now()
	if now.Sub(req.AccessTime).Seconds() > float64(secLayerContext.RequestTimeout) {
		res.Code = Timeout
		return
	}

	product, ok := secLayerContext.ProductMap[req.ProductId]
	if !ok {
		res.Code = ProductNotFound
		return
	}

	if rand.Float64() < product.BuyRate {
		beego.Info("buy rate")
		res.Code = NetworkBusy
		return
	}

	if product.ProductLimit.Check() >= product.SecSoldLimit {
		beego.Info("sold limit")
		res.Code = NetworkBusy
		return
	}

	productExt, ok := secLayerContext.ProductExtMap[req.ProductId]
	if !ok {
		productExt = &ProductExt{}
		secLayerContext.ProductExtMap[req.ProductId] = productExt
	}

	if productExt.Status == ProductStatusSoldOut {
		res.Code = SoldOut
		return
	}

	if productExt.Sold >= product.Total {
		productExt.Status = ProductStatusSoldOut
		res.Code = SoldOut
		return
	}

	if productExt.Sold+req.ProductNum > product.Total {
		res.Code = ProductNotEnough
		return
	}

	history, ok := secLayerContext.UserHistoryMap[req.UserId]
	if !ok {
		history = NewUserHistory()
		secLayerContext.UserHistoryMap[req.UserId] = history
	}

	if history.Check(req.ProductId)+req.ProductNum > product.OnePersonBuyLimit {
		res.Code = PurchaseExceed
		return
	}

	product.ProductLimit.Add()
	productExt.Sold += req.ProductNum
	history.Add(req.ProductId, req.ProductNum)

	res.Code = SecKillSuccess
	res.TokenTime = now
	str := fmt.Sprintf("product_id=%s&user_id=%s&nonce=%s&timestamp=%d&secret=%s", res.ProductId, res.UserId, res.Nonce, res.TokenTime.Unix(), secLayerContext.LayerSecret)
	h := sha256.Sum256([]byte(str))
	res.Token = hex.EncodeToString(h[:])

	return
}

func write()  {
	for res := range secLayerContext.HandleToWriteChan{
		data, err := json.Marshal(res)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secLayerContext.LayerToProxyPool.Get()
		_, err = cnn.Do("rpush", secLayerContext.LayerToProxyConf.QueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}
