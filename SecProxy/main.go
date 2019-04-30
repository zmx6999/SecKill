package main

import (
	_ "190430/SecProxy/service"
	"github.com/astaxie/beego"
	_ "190430/SecProxy/router"
)

func main()  {
	beego.Run()
}
