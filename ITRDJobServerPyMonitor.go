package main

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"log"
	"os"
)

func PyFileMonitor(ctx context.Context) {
	eventChannel := Cli3.Watch(ctx, FileListPreKey, clientv3.WithPrefix())
	for {
		select {
		case watchRsp := <- eventChannel:
			for _, event := range watchRsp.Events {
				pyFileModifiedHandler(event.Kv.Key)
			}
		case <-ctx.Done():
			return
		}
	}
}

func pyFileModifiedHandler(k []byte) {
	CurrentAccountName = "SyncPy"

	var err error
	switch string(k) {
	case PyFileNameListKey:
		err = pyFileReload()
	case PngFileNameListKey:
		err = pngFileReload()
	}

	if err == nil {
		CurrentAccountName = ""
	} else {
		os.Exit(1)
	}
}

func pyFileReload() error {
	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusSyncPyFiles); err != nil {
		errMsg := fmt.Sprintf("reload py files, modify node status to sync error: %v", err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		return err
	}

	if err := PullPyFiles(); err != nil {
		Logger.Error(fmt.Sprintf("sync py files error: %v", err))
		log.Println(fmt.Sprintf("sync py files error: %v", err))
		os.Exit(1)
	}

	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusFree); err != nil {
		errMsg := fmt.Sprintf("reload py files, modify node status to free error: %v", err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		return err
	}

	return nil
}

func pngFileReload() error {
	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusSyncPyFiles); err != nil {
		errMsg := fmt.Sprintf("reload png files, modify node status to sync error: %v", err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		return err
	}

	if err := PullPngFiles(); err != nil {
		Logger.Error(fmt.Sprintf("sync png files error: %v", err))
		log.Println(fmt.Sprintf("sync png files error: %v", err))
		os.Exit(1)
	}

	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusFree); err != nil {
		errMsg := fmt.Sprintf("reload png files, modify node status to free error: %v", err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		return err
	}

	return nil
}