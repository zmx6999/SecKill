package controller

import (
	"github.com/astaxie/beego"
	"encoding/json"
)

type BaseController struct {
	beego.Controller
}

type ResponseJSON struct {
	Code int
	Msg string
	Data interface{}
}

func (this *BaseController) handleResponse(code int, msg string, data interface{})  {
	this.Data["json"] = &ResponseJSON{code, msg, data}
	this.ServeJSON()
}

func (this *BaseController) success(data interface{})  {
	this.handleResponse(200, "ok", data)
}

func (this *BaseController) error(code int, msg string)  {
	this.handleResponse(code, msg, nil)
}

func (this *BaseController) getPost() (data map[string]interface{}, err error) {
	data = map[string]interface{}{}
	err = json.Unmarshal(this.Ctx.Input.RequestBody, &data)
	return
}
