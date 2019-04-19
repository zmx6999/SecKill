package main

import (
	"github.com/astaxie/beego"
	"190414/SecProxy/service"
	_ "190414/SecProxy/router"
	)

func main()  {
	service.InitService()
	beego.Run()
}
