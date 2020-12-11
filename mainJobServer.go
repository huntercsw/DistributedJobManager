package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
)

func init() {
	ip, err := externalIP()
	if err != nil {
		Logger.Error(fmt.Sprintf("get host ip error: %v", err))
		log.Println(fmt.Sprintf("get host ip error: %v", err))
		os.Exit(1)
	} else {
		NodeIpAddr = ip.String()
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> logger initialize <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	InitLogger("ITRD_jobServer")

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> configuration initialize <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := ConfigInit(); err != nil {
		Logger.Error(fmt.Sprintf("config init error: %v \n", err))
		log.Println(err.Error())
		os.Exit(1)
	}

	if err := Cli3Init(); err != nil {
		Logger.Error(fmt.Sprintf("cli3 init error: %v", err))
		log.Println(fmt.Sprintf("cli3 init error: %v", err))
		os.Exit(1)
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> node system type initialize <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := NodeSysTypeInit(); err != nil {
		os.Exit(1)
	}

	for {
		if err := MasterStatusCheckByHttpRequest(); err != nil {
			log.Println("check to master... ...")
			time.Sleep(2 * time.Second)
		} else {
			log.Println("connected to master")
			break
		}
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> sync configuration from master <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := PullFile(Conf.MasterInfo.Ip, MasterHostUser, MasterBaseDir+ConfigFileName, "./"+ConfigFileName); err != nil {
		Logger.Error(fmt.Sprintf("pull configuration[%s] from master[%s] error: %v", MasterBaseDir+ConfigFileName, Conf.MasterInfo.Ip, err))
		log.Println(fmt.Sprintf("pull configuration[%s] from master[%s] error: %v", MasterBaseDir+ConfigFileName, Conf.MasterInfo.Ip, err))
		os.Exit(1)
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> configuration initialize <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := ConfigInit(); err != nil {
		Logger.Error(fmt.Sprintf("config init error: %v \n", err))
		log.Println(err.Error())
		os.Exit(1)
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> sync py files from master <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := PullPyFiles(); err != nil {
		Logger.Error(fmt.Sprintf("sync py files error: %v", err))
		log.Println(fmt.Sprintf("sync py files error: %v", err))
		os.Exit(1)
	}

	log.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>> sync png files from master <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<")
	if err := PullPngFiles(); err != nil {
		Logger.Error(fmt.Sprintf("sync png files error: %v", err))
		log.Println(fmt.Sprintf("sync png files error: %v", err))
		os.Exit(1)
	}

	if err := ReporterInit(); err != nil {
		os.Exit(1)
	}
}

func main() {
	//if err := Cli3Init(); err != nil {
	//	fmt.Println(err)
	//	return
	//}
	defer Cli3.Close()

	itrdJS := ITRDJobServer{}
	nodeConsole := NodeConsole{}
	registryCtx, registryCancel := context.WithCancel(context.TODO())
	defer func() {
		if _, err := Cli3.Delete(context.TODO(), MyJobServerStatusKey); err != nil {
			Logger.Error(fmt.Sprintf("delete job server status to etcd error: %v", err))
		}
	}()

	go itrdJS.register(registryCtx)

	go itrdJS.run()

	go nodeConsole.run()

	registerMonitorCtx, registerMonitorCancel := context.WithCancel(context.TODO())
	defer registerMonitorCancel()

	go itrdJS.registerMonitor(registerMonitorCtx)
	go PyFileMonitor(registerMonitorCtx)
	go NodeConfMonitor(registerMonitorCtx)
	go SendCheckAlways(registerMonitorCtx)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)
	for {
		select {
		case <- RestartRegisterChannel:
			Logger.Error(fmt.Sprintf("recive signal to restart register"))
			log.Println(fmt.Sprintf("recive signal to restart register"))
			registryCancel()
			time.Sleep(NodeRegisterRestartInterval)
			// TODO: stop running job
			registryCtx, registryCancel = context.WithCancel(context.TODO())
			go itrdJS.register(registryCtx)
		case s := <-c:
			Logger.Error("Job Server exit: ", s)
			return
		}
	}
}