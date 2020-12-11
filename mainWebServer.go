package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"os"
	"os/signal"
)

func init() {
	InitLogger("ITRD_webServer")

	if err := ConfigInit(); err != nil {
		Logger.Error(fmt.Sprintf("config init error: %v \n", err))
		log.Fatal(err.Error())
	}

	if err := Cli3Init(); err != nil {
		Logger.Error(fmt.Sprintf("cli3 init error: %v", err))
		log.Println(fmt.Sprintf("cli3 init error: %v", err))
		os.Exit(1)
	}

	if _, err := Cli3.Put(context.TODO(), ConfigFileVersionKey, "0"); err != nil {
		Logger.Error(fmt.Sprintf("put initial config version to etcd error: %v", err))
		os.Exit(1)
	}

	if err := PyFileMap.PyFileMapInit(); err != nil {
		Logger.Error(fmt.Sprintf("py file map map init error: %v", err))
		log.Println(fmt.Sprintf("py file map map init error: %v", err))
		os.Exit(1)
	}

	if err := PngFileMap.PngFileMapInit(); err != nil {
		Logger.Error(fmt.Sprintf("png file map init error: %v", err))
		log.Println(fmt.Sprintf("png file map init error: %v", err))
		os.Exit(1)
	}
}

func main() {
	fmt.Println(VMBlackListAccountMap)
	defer Cli3.Close()
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(fmt.Sprintf("panic: %v \n", err))
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error(fmt.Sprintf("web server panic: %v", err))
			}
		}()
		r := gin.Default()
		RouterSetUp(r)
		if err := r.Run("0.0.0.0:" + WebPort); err != nil {
			Logger.Error("web server start error:", err)
			os.Exit(1)
		}
	}()

	go WsHub.run()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error(fmt.Sprintf("ws server panic: %v", err))
			}
		}()
		r := gin.Default()
		WsRouterSetUp(r)
		if err := r.Run("0.0.0.0:" + WsPort); err != nil {
			log.Println("ws server start error:", err)
			os.Exit(1)
		}
	}()

	go RunConsoleServer()

	err := MasterInit()

	if err != nil {
		log.Println(err)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go NodeMonitor(ctx)
	go NodeStatusMonitor(ctx)
	go RunMasterRouter(ctx)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	Logger.Error(s)
}
