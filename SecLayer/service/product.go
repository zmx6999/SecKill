package service

import (
	"time"
	"sync"
	"context"
	"encoding/json"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

type ProductMgr struct {
	ProductMap map[string]*Product
	ProductSoldMap map[string]*ProductSold
	ProductLock sync.RWMutex
}

func NewProductMgr() *ProductMgr {
	return &ProductMgr{
		ProductMap: map[string]*Product{},
		ProductSoldMap: map[string]*ProductSold{},
	}
}

func loadProduct() (err error) {
	r, err := secLayerContext.EtcdClient.Get(context.Background(), secLayerContext.EtcdConf.ProductKey)
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
		product.ProductSecSold = &ProductSecSold{}
		productMap[product.ProductId] = &product
	}

	secLayerContext.ProductLock.Lock()
	secLayerContext.ProductMap = productMap
	secLayerContext.ProductLock.Unlock()
}

func watchProduct() (err error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: []string{secLayerContext.EtcdAddr},
		DialTimeout: time.Second*time.Duration(secLayerContext.EtcdConf.Timeout),
	})
	if err != nil {
		return
	}
	defer cli.Close()

	for  {
		w := cli.Watch(context.Background(), secLayerContext.EtcdConf.ProductKey)
		var productList []Product
		for v := range w{
			success := true
			for _, ev := range v.Events{
				if ev.Type == mvccpb.PUT && string(ev.Kv.Key) == secLayerContext.EtcdConf.ProductKey {
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

	SecSoldLimit int
	OnePersonBuyLimit int
	SoldRate float64

	*ProductSecSold
}

type ProductSold struct {
	Sold int
	Status int
}

type ProductSecSold struct {
	num int
	curTime time.Time
}

func (ps *ProductSecSold) Check() int {
	now := time.Now()
	if now.Sub(ps.curTime).Seconds() > 1 {
		return 0
	}
	return ps.num
}

func (ps *ProductSecSold) Add()  {
	now := time.Now()
	if now.Sub(ps.curTime).Seconds() > 1 {
		ps.num = 1
		ps.curTime = now
	} else {
		ps.num++
	}
}
