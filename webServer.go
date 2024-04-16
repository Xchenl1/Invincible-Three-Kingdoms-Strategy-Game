package main

import (
	"fmt"
	"log"
	"net/http"
	"sgserver/config"
	"sgserver/server/web"
	"time"
)
import "github.com/gin-gonic/gin"

func main() {
	host := config.File.MustValue("web_server", "host", "127.0.0.1")
	port := config.File.MustValue("web_server", "port", "8088")

	router := gin.Default()
	//路由
	web.Init(router)
	s := &http.Server{
		Addr:           host + ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	fmt.Println("项目执行在", s.Addr)
	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
		return
	}
}
