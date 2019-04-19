package router

import (
	"github.com/astaxie/beego"
	"190414/SecProxy/controllers"
)

func init()  {
	beego.Router("/seckill/product", &controllers.SecKill{}, "Get:GetProduct")
	// beego.Router("/seckill/login", &controllers.SecKill{}, "Post:UserLogin")
	beego.Router("/seckill/seckill", &controllers.SecKill{}, "Post:SecKill")
}
