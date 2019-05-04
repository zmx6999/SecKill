package router

import (
	"github.com/astaxie/beego"
	"190504/SecProxy/controller"
)

func init()  {
	beego.Router("/seckill", &controller.SecKillController{}, "POST:SecKill")
}
