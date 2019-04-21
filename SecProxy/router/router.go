package router

import (
	"github.com/astaxie/beego"
	"190420/SecProxy/controllers"
)

func init()  {
	beego.Router("/seckill/seckill", &controllers.SecKillController{}, "post:SecKill")
}
