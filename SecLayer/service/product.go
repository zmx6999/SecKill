package service

import (
	"time"
	"sync"
	"encoding/json"
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type Product struct {
	ProductId string
	StartTime time.Time
	EndTime time.Time
	Status int
	Sold int
	Total int

	*SecLimit
}

type ProductMgr struct {
	ProductMap map[string]*Product
	ProductLock sync.RWMutex
}

/*
func newProductMgr() ProductMgr {
	return ProductMgr{
		ProductMap: map[string]*Product{},
	}
}
*/

func loadProduct() error {
	r, err := secLayerContext.EtcdClient.Get(context.Background(), secLayerContext.EtcdProductKey)
	if err != nil {
		return err
	}

	var productList []Product
	for _, v := range r.Kvs{
		err = json.Unmarshal(v.Value, &productList)
		if err != nil {
			continue
		}
	}

	updateProduct(productList)

	go watchProduct()

	return nil
}

func updateProduct(productList []Product)  {
	productMap := map[string]*Product{}
	for _, v := range productList{
		product := v
		product.SecLimit = &SecLimit{}
		productMap[product.ProductId] = &product
	}

	secLayerContext.ProductLock.Lock()
	secLayerContext.ProductMap = productMap
	secLayerContext.ProductLock.Unlock()
}

func watchProduct()  {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secLayerContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secLayerContext.EtcdTimeout),
	})
	if err != nil {
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secLayerContext.EtcdProductKey)
		var productList []Product
		for v := range w{
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secLayerContext.EtcdProductKey {
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
