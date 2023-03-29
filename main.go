package main

import (
	"fmt"
	"ggslayer/model"
	"ggslayer/routers"
	"ggslayer/tasks"
	"ggslayer/utils/config"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"time"
)

func init() {
	model.Setup()
	//启动定时任务
	tasks.Setup()
}

func main() {
	gin.SetMode(config.RunMode)
	r := routers.Setup()
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)
	server := &http.Server{
		Addr:           config.AppAddr,
		Handler:        r,
		ReadTimeout:    time.Duration(10) * time.Second,
		WriteTimeout:   time.Duration(10) * time.Second,
		MaxHeaderBytes: 2 << 20,
	}

	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}

}
