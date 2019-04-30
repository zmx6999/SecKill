package service

import (
	"time"
	"sync"
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"github.com/astaxie/beego"
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

type Product struct {
	ProductId string
	ProductName string
	Start time.Time
	End time.Time
	Status int
	Total int
}

func loadProduct() error {
	r, err := secProxyContext.EtcdClient.Get(context.Background(), secProxyContext.ProductKey)
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
		productMap[product.ProductId] = &product
	}

	secProxyContext.ProductLock.Lock()
	defer secProxyContext.ProductLock.Unlock()
	secProxyContext.ProductMap = productMap
}

func watchProduct()  {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.DialTimeout),
	})
	if err != nil {
		beego.Error(err)
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secProxyContext.ProductKey)
		for v := range w{
			var productList []Product
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secProxyContext.ProductKey {
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

func getProduct(productId string) (code int, err error) {
	secProxyContext.ProductLock.RLock()
	defer secProxyContext.ProductLock.RUnlock()

	product, ok := secProxyContext.ProductMap[productId]
	if !ok {
		code = ProductNotFound
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	now := time.Now()
	if now.Sub(product.Start).Seconds() < 0 {
		code = SecKillNotStart
		err = fmt.Errorf(GetErrMsg(code))
		return
	}
	if now.Sub(product.End).Seconds() > 0 {
		code = SecKillEnd
		err = fmt.Errorf(GetErrMsg(code))
		return
	}
	if product.Status == ProductStatusSoldOut {
		code = ForceSoldOut
		err = fmt.Errorf(GetErrMsg(code))
		return
	}

	return
}

func UpdateProductStatus(productId string, status int)  {
	secProxyContext.ProductLock.Lock()
	defer secProxyContext.ProductLock.Unlock()
	secProxyContext.ProductMap[productId].Status = status
}
