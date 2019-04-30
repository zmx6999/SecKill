package router

import (
	"github.com/astaxie/beego"
	"190430/SecProxy/controller"
)

func init()  {
	beego.Router("/seckill", &controller.SecKillController{}, "POST:SecKill")
}
