package main

import (
	"fmt"
	"pic-server/router"
	"pic-server/utils"
)

const Port = "8888"

func main() {
	utils.InitConfig()

	fmt.Println("hello world")
	r := router.Router()
	r.Run(":" + Port)
}
