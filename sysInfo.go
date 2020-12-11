package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

var (
	IsVirtualMachine = false
)

func NodeSysTypeInit() error {
	var manufacturer = ""
	cmd := exec.Command("systeminfo")
	if stdOut, err := cmd.CombinedOutput(); err != nil {
		Logger.Error(fmt.Sprintf("get system info error: %v", err))
		return err
	} else {
		sysInfo := strings.Split(ConvertByte2String(stdOut, GB18030), "\n")
		for _, row := range sysInfo {
			kv := strings.Split(row, ":")
			if len(kv) != 2 {
				continue
			} else {
				if kv[0] == "系统制造商" || kv[0] == "System Manufacturer"{
					manufacturer = kv[1]
				}
			}
		}
	}

	if manufacturer == "" {
		Logger.Error("node system type init error: system manufacturer not found")
		log.Println("node system type init error: system manufacturer not found")
		return errors.New("node system type init error: system manufacturer not found")
	}
	if strings.Contains(manufacturer, WindowsVirtualMachinePrimaryKey) {
		IsVirtualMachine = true
	}
	return nil
}