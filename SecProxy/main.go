package main

import (
	"github.com/astaxie/beego"
	_ "190420/SecProxy/service"
	_ "190420/SecProxy/router"
)

func main()  {
	beego.Run()
}
