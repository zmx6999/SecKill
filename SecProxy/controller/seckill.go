package controller

import (
	"190504/SecProxy/service"
	"time"
	"strings"
)

type SecKillController struct {
	BaseController
}

func (this *SecKillController) SecKill()  {
	data := this.getPost()
	productId, ok := data["product_id"].(string)
	if !ok {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}

	_productNum, ok := data["product_num"].(float64)
	if !ok {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}
	productNum := int(_productNum)
	if productNum < 1 {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}

	userId, ok := data["user_id"].(string)
	if !ok {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}

	nonce, ok := data["nonce"].(string)
	if !ok {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}

	addr := this.Ctx.Request.RemoteAddr
	if !strings.Contains(addr, ":") {
		code := service.InvalidRequest
		this.error(code, service.GetMsg(code))
		return
	}

	request := service.NewRequest()
	request.ProductId = productId
	request.ProductNum = productNum
	request.IP = strings.Split(addr, ":")[0]
	request.UserId = userId
	request.Nonce = nonce
	request.AccessTime = time.Now()
	request.CloseNotify = this.Ctx.ResponseWriter.CloseNotify()

	r, code, err := service.SecKill(request)
	if code == service.ProductSoldOut {
		service.UpdateProductStatus(request.ProductId, service.ProductStatusSoldOut)
	}
	if err != nil {
		this.error(code, err.Error())
		return
	}
	if r["token"] == "" {
		this.error(code, service.GetMsg(code))
		return
	}

	this.success(r)
}
