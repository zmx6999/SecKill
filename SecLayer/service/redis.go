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
	secLayerContext.Proxy2LayerPool = initRedisPool(secLayerContext.Proxy2LayerConf)
	secLayerContext.Layer2ProxyPool = initRedisPool(secLayerContext.Layer2ProxyConf)
}

func read()  {
	defer secLayerContext.WaitGroup.Done()
	for {
		cnn := secLayerContext.Proxy2LayerPool.Get()
		r, err := cnn.Do("blpop", secLayerContext.Proxy2LayerConf.QueueName, 0)
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

		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.RequestChanTimeout))
		select {
		case <-timer.C:
		case secLayerContext.RequestChan<-&req:
		}
		timer.Stop()
	}
}

func handle()  {
	defer secLayerContext.WaitGroup.Done()
	for req := range secLayerContext.RequestChan{
		res := secKill(req)
		timer := time.NewTicker(time.Second*time.Duration(secLayerContext.ResponseChanTimeout))
		select {
		case <-timer.C:
		case secLayerContext.ResponseChan<-res:
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

	if rand.Float64() < product.SoldRate {
		beego.Error("rate")
		res.Code = NetworkBusy
		return
	}

	productSold, ok := secLayerContext.ProductSoldMap[req.ProductId]
	if !ok {
		productSold = &ProductSold{}
		secLayerContext.ProductSoldMap[req.ProductId] = productSold
	}

	if productSold.Status == ProductStatusSoldOut {
		res.Code = ProductSoldOut
		return
	}

	if productSold.Sold >= product.Total {
		productSold.Status = ProductStatusSoldOut
		res.Code = ProductSoldOut
		return
	}

	if productSold.Sold+req.ProductNum > product.Total {
		res.Code = ProductStockInsufficient
		return
	}

	secSold := product.ProductSecSold
	if secSold.Check() >= product.SecSoldLimit {
		beego.Error("limit")
		res.Code = NetworkBusy
		return
	}

	secLayerContext.UserHistoryLock.Lock()
	defer secLayerContext.UserHistoryLock.Unlock()

	userHistory, ok := secLayerContext.UserHistoryMap[req.UserId]
	if !ok {
		userHistory = NewUserHistory()
		secLayerContext.UserHistoryMap[req.UserId] = userHistory
	}

	if userHistory.Check(req.ProductId)+req.ProductNum > product.OnePersonBuyLimit {
		res.Code = ProductOnePersonBuyExceed
		return
	}

	productSold.Sold += req.ProductNum
	secSold.Add()
	userHistory.Add(req.ProductId, req.ProductNum)

	res.Code = SecKillSuccess
	res.TokenTime = now
	str := fmt.Sprintf("product_id=%s&product_num=%d&user_id=%s&nonce=%s&token_time=%d", res.ProductId, res.ProductNum, res.UserId, res.Nonce, res.TokenTime.Unix())
	h := sha256.Sum256([]byte(str))
	res.Token = hex.EncodeToString(h[:])

	return
}

func write()  {
	defer secLayerContext.WaitGroup.Done()
	for res := range secLayerContext.ResponseChan{
		data, err := json.Marshal(res)
		if err != nil {
			beego.Error(err)
			continue
		}

		cnn := secLayerContext.Layer2ProxyPool.Get()
		_, err = cnn.Do("rpush", secLayerContext.Layer2ProxyConf.QueueName, data)
		cnn.Close()
		if err != nil {
			beego.Error(err)
			continue
		}
	}
}

func run()  {
	for i := 1; i <= secLayerContext.ReadGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go read()
	}
	for i := 1; i <= secLayerContext.HandleGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go handle()
	}
	for i := 1; i <= secLayerContext.WriteGoroutineNum; i++ {
		secLayerContext.WaitGroup.Add(1)
		go write()
	}
	secLayerContext.WaitGroup.Wait()
}
