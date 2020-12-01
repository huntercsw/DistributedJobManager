package main
//
//import (
//	"context"
//	"fmt"
//	"log"
//	"os"
//	"os/signal"
//)
//
//func init() {
//	InitLogger("ITRD_jobServer")
//
//	if err := ConfigInit(); err != nil {
//		Logger.Error(fmt.Sprintf("config init error: %v \n", err))
//		log.Println(err.Error())
//		os.Exit(1)
//	}
//
//	log.Println("******************sync config**********************")
//	// TODO: sync config
//
//	if err := ReporterInit(); err != nil {
//		os.Exit(1)
//	}
//}
//
//func main() {
//	if err := Cli3Init(); err != nil {
//		fmt.Println(err)
//		return
//	}
//	defer Cli3.Close()
//
//	itrdJS := ITRDJobServer{}
//	registryCtx, registryCancel := context.WithCancel(context.TODO())
//	defer func() {
//		if _, err := Cli3.Delete(context.TODO(), MyJobServerStatusKey); err != nil {
//			Logger.Error(fmt.Sprintf("delete job server status to etcd error: %v", err))
//		}
//	}()
//
//	go itrdJS.register(registryCtx)
//
//	go itrdJS.run()
//
//	registerMonitorCtx, registerMonitorCancel := context.WithCancel(context.TODO())
//	defer registerMonitorCancel()
//
//	go itrdJS.registerMonitor(registerMonitorCtx)
//
//	go JobMonitor()
//
//	c := make(chan os.Signal)
//	signal.Notify(c, os.Interrupt, os.Kill)
//	for {
//		select {
//		case <- RestartRegisterChannel:
//			Logger.Error(fmt.Sprintf("recive signal to restart register"))
//			log.Println(fmt.Sprintf("recive signal to restart register"))
//			registryCancel()
//			// TODO: stop running job
//			registryCtx, registryCancel = context.WithCancel(context.TODO())
//			go itrdJS.register(registryCtx)
//		case s := <-c:
//			Logger.Error("Job Server exit: ", s)
//			return
//		}
//	}
//}