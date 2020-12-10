package main

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/disk"
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const CONF_NAME = "./itrdConf.xml"

var (
	Conf      *ITRDConf
	ConfMutex sync.Mutex
)

type ITRDConf struct {
	Brokers       []Broker `xml:"broker"`
	OutputDir     string   `xml:"outputDir"`
	Logger        LogConf  `xml:"log"`
	PyProcessPath string   `xml:"pyProcessPath"`
	configSha1    string
	MasterInfo    MasterInfo              `xml:"master"`
	VMBlackList   VirtualMachineBlackList `xml:"virtualMachineBlackList"`
}

type Broker struct {
	Name        string `xml:"name"`
	AppVersion  string `xml:"appVersion"`
	AccountId   string `xml:"accountId"`
	AccountName string `xml:"accountName"`
	Credit      string `xml:"credit"`
	Password    string `xml:"password"`
	CommCode    string `xml:"commCode"`
	ExePath     string `xml:"exePath"`
	ProcessName string
}

type LogConf struct {
	Level string `xml:"level"`
	Path  string `xml:"path"`
}

type MasterInfo struct {
	Ip   string `xml:"ip"`
	Port string `xml:"port"`
}

type VirtualMachineBlackList struct {
	Brokers []string `xml:"broker"`
}

func (conf *ITRDConf) newConf() (err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("configuration initialize panic: ", err)
		}
	}()

	var (
		confBates []byte
	)
	if confBates, err = ioutil.ReadFile(CONF_NAME); err != nil {
		fmt.Println("read config file error:", err)
		return
	}

	if err = xml.Unmarshal(confBates, conf); err != nil {
		fmt.Println("conf unmarshal to struct error:", err)
		return
	}

	device := baseDevice()

	VMBlackListAccountMap = make(map[string]struct{})

	for i := 0; i < len(conf.Brokers); i++ {
		conf.Brokers[i].ProcessName = filepath.Base(conf.Brokers[i].ExePath)
		configMatchingDevice(&device, &(conf.Brokers[i].ExePath))

		for _, b := range conf.VMBlackList.Brokers {
			if conf.Brokers[i].Name == b {
				VMBlackListAccountMap[b] = struct{}{}
				break
			}
		}
	}

	configMatchingDevice(&device, &(conf.OutputDir))
	configMatchingDevice(&device, &(conf.PyProcessPath))
	configMatchingDevice(&device, &(conf.Logger.Path))

	conf.configSha1, err = fileSha1(CONF_NAME)
	if err != nil {
		return err
	}

	return nil
}

func Reconfiguration() error {
	currentSha1, err := fileSha1(CONF_NAME)
	if err != nil {
		return err
	}
	if currentSha1 == Conf.configSha1 {
		return nil
	}

	ConfMutex.Lock()
	err = ConfigInit()
	ConfMutex.Unlock()

	if err = updateConfigVersion(); err != nil {
		return err
	}

	return err
}

func ConfigInit() (err error) {
	Conf = new(ITRDConf)
	if err = Conf.newConf(); err != nil {
		return
	}

	return nil
}

func (conf *ITRDConf) BrokerClientName(accountName string) string {
	for _, b := range conf.Brokers {
		if b.AccountName == accountName {
			return b.ProcessName
		}
	}
	return ""
}

func hasDevice(device string) bool {
	infos, _ := disk.Partitions(false)
	for _, info := range infos {
		if info.Device == device && info.Fstype != "UDF" {
			return true
		}
	}
	return false
}

func baseDevice() string {
	var device string
	if hasDevice(DeviceE) {
		device = DeviceE
	} else if hasDevice(DeviceD) {
		device = DeviceD
	} else {
		device = DeviceC
	}
	return device
}

func configMatchingDevice(device, path *string) {
	tmp := strings.Split(*path, "/")
	tmp[0] = *device
	*path = strings.Join(tmp, "/")
}

func updateConfigVersion() error {
	rsp, err := Cli3.Get(context.TODO(), ConfigFileVersionKey)
	if err != nil {
		Logger.Error(fmt.Sprintf("get config version error: %v", err))
		return err
	}

	if rsp.Count != 1 {
		Logger.Error("number of config version key error")
		return errors.New("number of config version key error")
	}

	var version int
	version, err = strconv.Atoi(string(rsp.Kvs[0].Value))
	if err != nil {
		Logger.Error(fmt.Sprintf("config version not a number: %v", err))
		return err
	}

	if _, err = Cli3.Put(context.TODO(), ConfigFileVersionKey, strconv.Itoa(version+1)); err != nil {
		Logger.Error(fmt.Sprintf("config version puls error: %v", err))
		return err
	}

	return nil
}
