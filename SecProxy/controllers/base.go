package controllers

import (
	"github.com/astaxie/beego"
	"encoding/json"
)

type ResponseJSON struct {
	Code int
	Msg string
	Data interface{}
}

type BaseController struct {
	beego.Controller
}

func (this *BaseController) handleResponse(code int, msg string, data interface{})  {
	this.Data["json"] = &ResponseJSON{Code: code, Msg: msg, Data: data}
	this.ServeJSON()
}

func (this *BaseController) success(data interface{}) {
	this.handleResponse(200, "ok", data)
}

func (this *BaseController) error(code int, msg string)  {
	this.handleResponse(code, msg, nil)
}

func (this *BaseController) getPostParamMap() map[string]string {
	body := this.Ctx.Input.RequestBody
	data := make(map[string]string)
	json.Unmarshal(body, &data)
	return data
}
