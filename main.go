package main

import (
	"boudle/client"
	"boudle/deploy"
	"boudle/until"
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	client.Init()
	if err := until.Init("2021-04-16", 1); err != nil {
		fmt.Printf("init failed, err:%v\n", err)
		return
	}
	r := gin.Default()
	r.MaxMultipartMemory = 50 << 20 // 50 MB 设置表单上传文件的最大为50 MB
	r.POST("/api/v1/deploy", deploy.Create)
	r.Run(":7777")
}
