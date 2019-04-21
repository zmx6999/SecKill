package service

import (
	"time"
	"sync"
	"encoding/json"
	"fmt"
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type Product struct {
	ProductId string
	StartTime time.Time
	EndTime time.Time
	Status int
	// Sold int
	Total int
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
	r, err := secProxyContext.EtcdClient.Get(context.Background(), secProxyContext.EtcdProductKey)
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
		productMap[product.ProductId] = &product
	}

	secProxyContext.ProductLock.Lock()
	secProxyContext.ProductMap = productMap
	secProxyContext.ProductLock.Unlock()
}

func watchProduct()  {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.EtcdTimeout),
	})
	if err != nil {
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secProxyContext.EtcdProductKey)
		var productList []Product
		for v := range w{
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secProxyContext.EtcdProductKey {
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

func UpdateProductStatus(productId string, status int)  {
	secProxyContext.ProductLock.Lock()
	secProxyContext.ProductMap[productId].Status = status
	secProxyContext.ProductLock.Unlock()
}

func getProduct(productId string) (code int, err error) {
	secProxyContext.ProductLock.RLock()
	defer secProxyContext.ProductLock.RUnlock()

	product, ok := secProxyContext.ProductMap[productId]
	if !ok {
		code = ProductNotFoundErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	if product.Status == ProductStatusSoldOut {
		code = ProductForceSoldOutErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	now := time.Now()
	if now.Sub(product.StartTime).Seconds() < 0 {
		code = SecKillNotStartErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}
	if now.Sub(product.EndTime).Seconds() > 0 {
		code = SecKillAlreadyEndErr
		err = fmt.Errorf(getErrMsg(code))
		return
	}

	return
}
