package service

import (
	"time"
	"sync"
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"github.com/astaxie/beego"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type ProductMgr struct {
	ProductMap map[string]*Product
	ProductExtMap map[string]*ProductExt
	ProductLock sync.RWMutex
}

func NewProductMgr() *ProductMgr {
	return &ProductMgr{
		ProductMap: map[string]*Product{},
		ProductExtMap: map[string]*ProductExt{},
	}
}

type Product struct {
	ProductId string
	Total int

	SecSoldLimit int
	OnePersonBuyLimit int
	BuyRate float64
	*ProductLimit
}

func loadProduct() error {
	r, err := secLayerContext.EtcdClient.Get(context.Background(), secLayerContext.ProductKey)
	if err != nil {
		return err
	}

	var productList []Product
	for _, v := range r.Kvs{
		json.Unmarshal(v.Value, &productList)
	}

	updateProduct(productList)

	go watchProduct()

	return nil
}

func updateProduct(productList []Product)  {
	productMap := map[string]*Product{}
	for _, v := range productList{
		product := v
		product.ProductLimit = &ProductLimit{}
		productMap[product.ProductId] = &product
	}

	secLayerContext.ProductLock.Lock()
	defer secLayerContext.ProductLock.Unlock()
	secLayerContext.ProductMap = productMap
}

func watchProduct()  {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secLayerContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secLayerContext.DialTimeout),
	})
	if err != nil {
		beego.Error(err)
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secLayerContext.ProductKey)
		for v := range w{
			var productList []Product
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secLayerContext.ProductKey {
					err = json.Unmarshal(ev.Kv.Value, &productList)
					if err != nil {
						success = false
					}
				}
			}

			if success {
				updateProduct(productList)
			}
		}
	}
}

type ProductExt struct {
	Sold int
	Status int
}

type ProductLimit struct {
	count int
	curTime time.Time
}

func (pl *ProductLimit) Check() int {
	now := time.Now()
	if now.Sub(pl.curTime).Seconds() > 1 {
		return 0
	}
	return pl.count
}

func (pl *ProductLimit) Add() {
	now := time.Now()
	if now.Sub(pl.curTime).Seconds() > 1 {
		pl.count = 1
		pl.curTime = now
		return
	}
	pl.count++
}
