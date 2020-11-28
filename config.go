package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
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
	MasterInfo    MasterInfo `xml:"master"`
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

	//conf.OutputDir = path.Join(conf.OutputDir, )

	for i := 0; i < len(conf.Brokers); i++ {
		conf.Brokers[i].ProcessName = filepath.Base(conf.Brokers[i].ExePath)
	}

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
	return err
}

func ConfigInit() (err error) {
	Conf = new(ITRDConf)
	if err = Conf.newConf(); err != nil {
		return
	}
	return nil
}
