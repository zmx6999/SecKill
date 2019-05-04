package main

import (
	_ "190504/SecProxy/service"
	"github.com/astaxie/beego"
	_ "190504/SecProxy/router"
)

func main()  {
	beego.Run()
}
