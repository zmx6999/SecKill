package controllers

import (
	"190420/SecProxy/service"
	"time"
	"strings"
	)

type SecKillController struct {
	BaseController
}

func (this *SecKillController) SecKill()  {
	data := this.getPostParams()
	productId, ok := data["product_id"].(string)
	if !ok || productId == "" {
		this.error(1011, "Invalid request param")
		return
	}

	productNum, ok := data["product_num"].(float64)
	if !ok  {
		this.error(1011, "Invalid request param")
		return
	}
	if productNum < 1 {
		this.error(1011, "Invalid request param")
		return
	}

	userId, ok := data["user_id"].(string)
	if !ok || userId == "" {
		this.error(1011, "Invalid request param")
		return
	}

	nonce, ok := data["nonce"].(string)
	if !ok || nonce == "" {
		this.error(1011, "Invalid request param")
		return
	}

	req := service.NewSecRequest()
	req.ProductId = productId
	req.ProductNum = int(productNum)
	req.UserId = userId
	req.Nonce = nonce
	req.AccessTime = time.Now()

	addr := this.Ctx.Request.RemoteAddr
	if strings.Contains(addr, ":") {
		addr = strings.Split(addr, ":")[0]
	}
	req.IP = addr

	req.CloseNotify = this.Ctx.ResponseWriter.CloseNotify()

	data, code, err := service.SecKill(req)
	if code == service.ProductSoldOutErr {
		service.UpdateProductStatus(productId, service.ProductStatusSoldOut)
	}

	if err != nil {
		this.error(code, err.Error())
		return
	}

	if code != service.SecKillSuccess {
		this.error(code, service.GetErrMsg(code))
		return
	}

	this.success(data)
}
