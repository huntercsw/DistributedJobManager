package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
)

func NodeConfMonitor(ctx context.Context) {
	eventChannel := Cli3.Watch(ctx, ConfigFileVersionKey)
	for {
		select {
		case watchRsp := <- eventChannel:
			for _, event := range watchRsp.Events {
				Logger.Debug(fmt.Sprintf("master configuration version modified, sync Conf on node... ..."))
				reConfig(event.Kv.Value)
			}
		case <-ctx.Done():
			return
		}
	}
}

func reConfig(v []byte) {
	newConfigVersion, err := strconv.Atoi(string(v))
	if err != nil {
		Logger.Error(fmt.Sprintf("configuration version not a number: %v", err))
		os.Exit(1)
	}

	if newConfigVersion <= NodeConfigurationVersion {
		Logger.Debug(fmt.Sprintf("master config version[%d] smaller then node config version[%d]", newConfigVersion, NodeConfigurationVersion))
		return
	}

	err = closeBrokerClient(CurrentAccountName)
	if err != nil {
		Logger.Error(fmt.Sprintf("node reConfig, stop running job[%s] error: %v", CurrentAccountName, err))
		return
	}

	CurrentAccountName = "ReConfig"
	defer func() {
		CurrentAccountName = ""
	}()

	if err = PullFile(Conf.MasterInfo.Ip, MasterHostUser, MasterBaseDir+ConfigFileName, "./"+ConfigFileName); err != nil {
		Logger.Error(fmt.Sprintf("pull configuration[%s] from master[%s] error: %v", MasterBaseDir+ConfigFileName, Conf.MasterInfo.Ip, err))
		os.Exit(1)
	}

	if err = ConfigInit(); err != nil {
		Logger.Error(fmt.Sprintf("config reInitialize error: %v \n", err))
		os.Exit(1)
	}

	NodeConfigurationVersion = newConfigVersion
}