package controller

import (
	"190430/SecProxy/service"
	"time"
	"strings"
)

type SecKillController struct {
	BaseController
}

func (this *SecKillController) SecKill()  {
	data, err := this.getPost()
	if err != nil {
		this.error(1020, "Invalid param")
		return
	}

	productId, ok := data["product_id"].(string)
	if !ok {
		this.error(1020, "Invalid param")
		return
	}

	userId, ok := data["user_id"].(string)
	if !ok {
		this.error(1020, "Invalid param")
		return
	}

	_productNum, ok := data["product_num"].(float64)
	if !ok {
		this.error(1020, "Invalid param")
		return
	}
	productNum := int(_productNum)
	if productNum < 1 {
		this.error(1020, "Invalid param")
		return
	}

	nonce, ok := data["nonce"].(string)
	if !ok {
		this.error(1020, "Invalid param")
		return
	}

	request := service.NewRequest()
	ip := this.Ctx.Request.RemoteAddr
	if strings.Contains(ip, ":") {
		request.IP = strings.Split(ip, ":")[0]
	}
	request.ProductId = productId
	request.ProductNum = productNum
	request.UserId = userId
	request.Nonce = nonce
	request.AccessTime = time.Now()
	request.CloseNotify = this.Ctx.ResponseWriter.CloseNotify()

	data, code, err := service.SecKill(request)
	if code == service.SoldOut {
		service.UpdateProductStatus(productId, service.ProductStatusSoldOut)
	}
	if err != nil {
		this.error(code, err.Error())
		return
	}
	if data["token"] == "" {
		this.error(code, service.GetErrMsg(code))
		return
	}

	this.success(data)
}
