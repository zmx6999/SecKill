package service

import (
	"time"
	"sync"
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"fmt"
)

type ProductMgr struct {
	ProductMap map[string]*Product
	ProductLock sync.RWMutex
}

func NewProductMgr() *ProductMgr {
	return &ProductMgr{
		ProductMap: map[string]*Product{},
	}
}

func loadProduct() (err error) {
	r, err := secProxyContext.EtcdClient.Get(context.Background(), secProxyContext.EtcdConf.ProductKey)
	if err != nil {
		return
	}

	var productList []Product
	for _, v := range r.Kvs{
		json.Unmarshal(v.Value, &productList)
	}
	updateProduct(productList)

	go watchProduct()

	return
}

func updateProduct(productList []Product)  {
	productMap := map[string]*Product{}
	for _, v := range productList {
		product := v
		productMap[product.ProductId] = &product
	}

	secProxyContext.ProductLock.Lock()
	secProxyContext.ProductMap = productMap
	secProxyContext.ProductLock.Unlock()
}

func watchProduct() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.EtcdConf.Timeout),
	})
	if err != nil {
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secProxyContext.EtcdConf.ProductKey)
		var productList []Product
		for v := range w{
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secProxyContext.EtcdConf.ProductKey {
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

type Product struct {
	ProductId string
	Total int
	Start time.Time
	End time.Time
	Status int
}

func getProduct(productId string) (code int, err error) {
	secProxyContext.ProductLock.RLock()
	defer secProxyContext.ProductLock.RUnlock()

	product, ok := secProxyContext.ProductMap[productId]
	if !ok {
		code = ProductNotFound
		err = fmt.Errorf(GetMsg(code))
		return
	}

	if product.Status == ProductStatusSoldOut {
		code = ProductForceSoldOut
		err = fmt.Errorf(GetMsg(code))
		return
	}

	now := time.Now()
	if now.Sub(product.Start).Seconds() < 0 {
		code = SecKillNotStart
		err = fmt.Errorf(GetMsg(code))
		return
	}
	if now.Sub(product.End).Seconds() > 0 {
		code = SecKillEnd
		err = fmt.Errorf(GetMsg(code))
		return
	}

	return
}

func UpdateProductStatus(productId string, status int)  {
	secProxyContext.ProductLock.Lock()
	secProxyContext.ProductMap[productId].Status = status
	secProxyContext.ProductLock.Unlock()
}
