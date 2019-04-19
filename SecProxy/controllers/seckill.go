package controllers

import (
	"190414/SecProxy/service"
	"strconv"
	"time"
	"strings"
)

type SecKill struct {
	BaseController
}

func (this *SecKill) GetProduct()  {
	productId := this.GetString("product_id")
	if productId == "" {
		data := service.GetProductList()
		this.success(data)
	} else {
		data, code, err := service.GetProduct(productId)
		if err != nil {
			this.error(code, err.Error())
		} else {
			this.success(data)
		}
	}
}

/*
func (this *SecKill) UserLogin()  {
	data := this.getPostParamMap()
	userId, err := strconv.Atoi(data["user_id"])
	if err != nil {
		code := service.InvalidRequestParamErr
		this.error(code, service.ErrMsg[code])
		return
	}

	authStr := service.Login(userId)
	this.Ctx.SetCookie("user_id", strconv.Itoa(userId))
	this.Ctx.SetCookie("user_auth_sign", authStr)

	this.success(nil)
}
*/

func (this *SecKill) SecKill()  {
	data := this.getPostParamMap()
	productId := data["product_id"]
	if productId == "" {
		code := service.InvalidRequestParamErr
		this.error(code, service.ErrMsg[code])
		return
	}
	userId := data["user_id"]
	if userId == "" {
		code := service.InvalidRequestParamErr
		this.error(code, service.ErrMsg[code])
		return
	}

	authCode := data["auth_code"]
	nonce := data["nonce"]
	secTime, _ := strconv.Atoi(data["sec_time"])
	source := data["source"]
	userAuthSign := data["user_auth_sign"]

	request := service.NewSecRequest()
	request.AuthCode = authCode
	request.Nonce = nonce
	request.SecTime = time.Unix(int64(secTime), 0)
	request.Source = source
	request.UserId = userId
	request.UserAuthSign = userAuthSign
	request.ProductId = productId
	request.AccessTime = time.Now()
	request.CloseNotify = this.Ctx.ResponseWriter.CloseNotify()

	remoteAddr := this.Ctx.Request.RemoteAddr
	if strings.Contains(remoteAddr, ":") {
		request.RemoteAddr = strings.Split(remoteAddr, ":")[0]
	} else {
		request.RemoteAddr = remoteAddr
	}

	/*
	valid := service.ValidateUser(request)
	if !valid {
		code := service.UserValidationErr
		this.error(code, service.ErrMsg[code])
		return
	}
	*/

	r, code, err := service.SecKill(request)
	if code == 1004 {
		service.UpdateProductStatus(productId, service.ProductStatusSoldOut)
	}

	if err != nil {
		this.error(code, err.Error())
		return
	}

	if r == nil || r["token"] == "" {
		this.error(code, "seckill failed")
		return
	}

	this.success(r)
}
