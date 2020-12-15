package main

import (
	"fmt"
	"project/server/config"
	"project/server/controller"
)

func main() {

	router := controller.InitRouter()
	_ = router.Run(":" + config.Cfg.Section("common").Key("port").String())
	fmt.Println("=========程序到此==============")
}

