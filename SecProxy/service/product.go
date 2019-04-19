package service

import (
	"time"
	"github.com/astaxie/beego"
	"encoding/json"
	"sync"
	"context"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"fmt"
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
}

func initProduct(secProxyContext *SecProxyContext) error {
	r, err := secProxyContext.EtcdClient.Get(context.Background(), secProxyContext.EtcdConf.EtcdProductKey)
	if err != nil {
		return err
	}

	var productList []Product
	for _, v := range r.Kvs{
		json.Unmarshal(v.Value, &productList)
	}

	updateProduct(secProxyContext, productList)

	go watchProduct(secProxyContext)

	return nil
}

func watchProduct(secProxyContext *SecProxyContext) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secProxyContext.EtcdConf.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secProxyContext.EtcdConf.EtcdTimeout),
	})
	if err != nil {
		beego.Error(err)
		return
	}
	defer cli.Close()

	for  {
		ch := cli.Watch(context.Background(), secProxyContext.EtcdConf.EtcdProductKey)
		for v := range ch{
			var productList []Product
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secProxyContext.EtcdConf.EtcdProductKey {
					json.Unmarshal(ev.Kv.Value, &productList)
				}
			}
			updateProduct(secProxyContext, productList)
		}
	}
}

func updateProduct(secProxyContext *SecProxyContext, productList []Product)  {
	productMap := make(map[string]*Product)
	for _, v := range productList{
		product := v
		productMap[product.ProductId] = &product
	}
	secProxyContext.ProductMapLock.Lock()
	secProxyContext.ProductMap = productMap
	secProxyContext.ProductMapLock.Unlock()
}

type ProductInfo struct {
	ProductId string
	Process int
	Msg string
}

func GetProduct(productId string) (data ProductInfo, code int, err error) {
	secProxyContext.ProductMapLock.RLock()
	defer secProxyContext.ProductMapLock.RUnlock()

	product, ok := secProxyContext.ProductMap[productId]
	if !ok {
		code = ProductNotFoundErr
		err = fmt.Errorf(ErrMsg[code])
		return
	}

	t := time.Now()
	if t.Sub(product.StartTime).Seconds() < 0 {
		code = ProductNotStartErr
		data.Process = ProductProcessNotStart
		data.Msg = ErrMsg[code]
	} else if t.Sub(product.EndTime).Seconds() < 0 {
		code = OK
		data.Process = ProductProcessStart
		data.Msg = ErrMsg[code]
	} else {
		code = ProductAlreadyEndErr
		data.Process = ProductProcessEnd
		data.Msg = ErrMsg[code]
	}

	if product.Status == ProductStatusSoldOut || product.Status == ProductStatusForceSoldOut {
		code = ProductSoldOutErr
		data.Process = ProductProcessEnd
		data.Msg = ErrMsg[code]
	}

	data.ProductId = productId

	return
}

func GetProductList() (data []ProductInfo) {
	secProxyContext.ProductMapLock.RLock()
	defer secProxyContext.ProductMapLock.RUnlock()

	for _, product := range secProxyContext.ProductMap{
		item, _, err := GetProduct(product.ProductId)
		if err != nil {
			continue
		}

		data = append(data, item)
	}

	return
}
