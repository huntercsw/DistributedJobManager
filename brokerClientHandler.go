package main

import "C"
import (
	"errors"
	"fmt"
	"github.com/go-vgo/robotgo"
	"os/exec"
	"path"
)

type BrokerClient struct {
	Br                 *Broker
	ClientExeName      string
	Pid                int32
	ClientWindowHandle int
}

func (b *Broker) newBrokerClient() *BrokerClient {
	return &BrokerClient{
		Br:                 b,
		ClientExeName:      path.Base(b.ExePath),
		Pid:                0,
		ClientWindowHandle: 0,
	}
}

func (b *Broker) OpenClient() (bc *BrokerClient, err error) {
	if pids, err1 := clientHasRun(path.Base(b.ExePath)); err1 != nil {
		return nil, err1
	} else {
		if len(pids) > 0 {
			for i := 0; i < len(pids); i++ {
				robotgo.Kill(pids[i])
			}
		}
	}

	cmd := exec.Command("cmd.exe", "/c", "start "+b.ExePath)
	err = cmd.Run()
	if err != nil {
		return nil, err
	} else {
		bc = b.newBrokerClient()
		var pid int32
		if pid, err = getClientPidByName(bc.ClientExeName); err != nil {
			return nil, errors.New(fmt.Sprintf("get pid error: %v", err))
		}
		bc.Pid = pid
		bc.ClientWindowHandle = robotgo.GetHandle()
		return
	}
}

func clientHasRun(clientExe string) (pids []int32, err error) {
	return robotgo.FindIds(clientExe)
}

func getClientPidByName(clientExe string) (pid int32, err error) {
	var pids []int32
	pids, err = robotgo.FindIds(clientExe)
	if err != nil {
		return 0, err
	} else {
		return pids[0], nil
	}
}

func SelectTradeOnlyButton(x, y int) {

}
