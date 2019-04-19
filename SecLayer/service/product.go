package service

import (
	"context"
	"encoding/json"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"time"
	"sync"
)

type ProductMgr struct {
	ProductMap map[string]*Product
	ProductMapLock sync.RWMutex
}

type Product struct {
	ProductId string
	StartTime time.Time
	EndTime time.Time
	Status int
	Total int
	Sold int

	Limit *Limit
}

func loadProduct(secLayerContext *SecLayerContext) error {
	secLayerConf := secLayerContext.SecLayerConf
	r, err := secLayerContext.EtcdClient.Get(context.Background(), secLayerConf.EtcdProductKey)
	if err != nil {
		return err
	}

	var productList []Product
	for _, v := range r.Kvs{
		json.Unmarshal(v.Value, &productList)
	}

	updateProductMap(secLayerContext, productList)

	go watchProduct(secLayerContext)

	return nil
}

func updateProductMap(secLayerContext *SecLayerContext, productList []Product)  {
	productMap := make(map[string]*Product)
	for _, v := range productList{
		product := v
		product.Limit = &Limit{}
		productMap[product.ProductId] = &product
	}

	secLayerContext.ProductMapLock.Lock()
	secLayerContext.ProductMap = productMap
	secLayerContext.ProductMapLock.Unlock()
}

func watchProduct(secLayerContext *SecLayerContext) {
	key := secLayerContext.SecLayerConf.EtcdProductKey
	for  {
		r := secLayerContext.EtcdClient.Watch(context.Background(), key)
		var productList []Product
		s := true
		for v := range r{
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == key {
					err := json.Unmarshal(ev.Kv.Value, &productList)
					if err != nil {
						s = true
					}
				}
			}

			if s {
				updateProductMap(secLayerContext, productList)
			}
		}
	}
}
