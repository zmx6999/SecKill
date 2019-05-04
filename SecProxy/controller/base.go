package controller

import (
	"github.com/astaxie/beego"
	"encoding/json"
)

type BaseController struct {
	beego.Controller
}

func (this *BaseController) handleResponse(code int, msg string, data interface{})  {
	this.Data["json"] = map[string]interface{}{"code": code, "msg": msg, "data": data}
	this.ServeJSON()
}

func (this *BaseController) success(data interface{}) {
	this.handleResponse(200, "ok", data)
}

func (this *BaseController) error(code int, msg string) {
	this.handleResponse(code, msg, nil)
}

func (this *BaseController) getPost() (data map[string]interface{}) {
	data = map[string]interface{}{}
	json.Unmarshal(this.Ctx.Input.RequestBody, &data)
	return
}
