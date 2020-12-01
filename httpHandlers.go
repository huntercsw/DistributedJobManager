package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	OneAccountTimeOut = time.Second * 400
)

var (
	BrokerClientIsRunning = false
	JobChan = make(chan string, 1000)
	currentProcessCancel context.CancelFunc
	currentBrokerClient string
)

type Account struct {
	BrokerName string
	Name       string
	Status     [3]bool // status[0]--positionFileStatus; status[1]--orderFileStatus; status[2]--tradeFileStatus
}

func AccountList(c *gin.Context) {
	if err := Reconfiguration(); err != nil {
		Logger.Error(fmt.Sprintf("reload configuration error: %v", err))
		os.Exit(1)
	}

	accounts, err := responseData()
	if err != nil {
		c.JSON(
			http.StatusOK,
			gin.H{
				"ErrorCode": http.StatusInternalServerError,
				"Data":      err.Error(),
			},
		)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": accounts})
}

func responseData() (accounts []Account, err error) {
	var files []os.FileInfo
	outputDir := path.Join(Conf.OutputDir, time.Now().Format("20060102"))

	_, err1 := os.Stat(outputDir)
	if err1 == nil {
		files, err = ioutil.ReadDir(outputDir)
		if err != nil {
			return
		}
	} else {
		if os.IsExist(err1) {
			files, err = ioutil.ReadDir(outputDir)
			if err != nil {
				return
			}
		}
	}

	accounts = make([]Account, 0)
	for _, broker := range Conf.Brokers {
		account := Account{BrokerName: broker.Name, Name: broker.AccountName}
		for _, file := range files {
			if strings.HasPrefix(file.Name(), broker.AccountName) {
				switch strings.Split(file.Name(), "_")[1] {
				case time.Now().Format("20060102"):
					account.Status[0] = true
				case "order":
					account.Status[1] = true
				case "trade":
					account.Status[2] = true
				}
			}
		}
		accounts = append(accounts, account)
	}

	return
}

func BrokerClientProcessIsRunning(c *gin.Context) {
	if BrokerClientIsRunning {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "download process is running"})
	} else {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": ""})
	}
}

func DownLoad(c *gin.Context) {
	log.Println("current WS connection number:", len(WsHub.clients))
	req, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "request body error: " + err.Error()})
	}
	accounts := new(Accounts)
	if err := json.Unmarshal(req, accounts); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "accounts unmarshal error: " + err.Error()})
		return
	}

	if len(accounts.AccountList) != 1 {
		c.JSON(
			http.StatusOK,
			gin.H{
				"ErrorCode": http.StatusBadRequest,
				"Data":      fmt.Sprintf("number of account require 1 but %d give", len(accounts.AccountList)),
			},
		)
		return
	}
	//if rsp, err := downLoadAccount(accounts.AccountList[0]); err != nil {
	//	c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": fmt.Sprintf("start python process error: %v", err)})
	//} else {
	//	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": rsp})
	//}

	JobChan <- accounts.AccountList[0]
	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": "waiting for download..."})
}

func Stop(c *gin.Context) {
	defer func() {
		currentBrokerClient = ""
	}()
	jobCount := len(JobChan)
	for i := 0; i < jobCount; i++ {
		select {
		case <- JobChan:
		default:
		}
	}
	currentProcessCancel()

	rspData := ""
	if currentBrokerClient != "" {
		if err := killWindowsProcessByName(currentBrokerClient); err != nil {
			rspData = fmt.Sprintf("kill client[%s] error: %v", currentBrokerClient, err)
			log.Println(rspData)
			Logger.Error(rspData)
		}
	}
	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": rspData})
}

func downLoadAccount(account string) (*BrokerClientRequest, error) {
	defer func() {
		BrokerClientIsRunning = false
		if err := recover(); err != nil {
			Logger.Error(err)
		}
	}()

	for i := 0; i < len(BrokerClientResponseChannel); i++ {
		<-BrokerClientResponseChannel
	}

	BrokerClientIsRunning = true
	log.Println(account)
	ctx, cancel := context.WithCancel(context.TODO())
	currentProcessCancel = cancel

	var rsp *BrokerClientRequest
	processErrorChannel := make(chan error)
	defer close(processErrorChannel)

	for _, b := range Conf.Brokers {
		if b.AccountName == account {
			currentBrokerClient = b.ProcessName
		}
	}

	cmd := exec.CommandContext(ctx, "python", Conf.PyProcessPath, account)
	stdOutMsg, stdErrMsg := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stdout, cmd.Stderr = stdOutMsg, stdErrMsg

	go func() {
		defer func() {
			if err := recover(); err != nil {
				Logger.Error(fmt.Sprintf("download %s panic: %v \n", account, err))
			}
		}()

		if err := cmd.Start(); err != nil {
			Logger.Error(fmt.Sprintf("start %s error: %v", account, err))
			log.Println(fmt.Sprintf("start %s error: %v", account, err))
			processErrorChannel <- err
		}

		if err := cmd.Wait(); err != nil {
			Logger.Error(fmt.Sprintf("Error: deal with %s error %v", account, err))
			log.Println(fmt.Sprintf("Error: deal with %s error %v", account, err))
			select {
			case _, ok := <- processErrorChannel:
				if ok {
					processErrorChannel <- err
				}
			default:
				processErrorChannel <- err
			}

		} else {
			if stdErrMsg.String() != "" {
				Logger.Error(fmt.Sprintf("do %s error: %v", account, err))
				log.Println(fmt.Sprintf("do %s error: %v", account, err))
				select {
				case _, ok := <- processErrorChannel:
					if ok {
						processErrorChannel <- errors.New(stdErrMsg.String())
					}
				default:
					processErrorChannel <- errors.New(stdErrMsg.String())
				}
			}
		}
	}()

	select {
	case err := <-processErrorChannel:
		Logger.Error(fmt.Sprintf("exec py [%s] error: %v", account, err))
		log.Println(fmt.Sprintf("exec py [%s] error: %v", account, err))
		return rsp, err
	case rsp = <-BrokerClientResponseChannel:
		Logger.Debug(fmt.Sprintf("py response[%s]: %v \n", account, rsp))
		log.Println(fmt.Sprintf("py response[%s]: %v \n", account, rsp))
		return rsp, nil
	case <-time.After(OneAccountTimeOut):
		Logger.Debug(fmt.Sprintf("py [%s] time out \n", account))
		log.Println(fmt.Sprintf("py [%s] time out \n", account))
		cancel()
		if err := killWindowsProcessByName(currentBrokerClient); err != nil {
			Logger.Error(fmt.Sprintf("timeout cancel kill client[%s] error: %v", currentBrokerClient, err))
			log.Println(fmt.Sprintf("timeout cancel kill client[%s] error: %v", currentBrokerClient, err))
		}
		currentBrokerClient = ""
		return rsp, errors.New(fmt.Sprintf("%s timeout", account))
	}
}

type Accounts struct {
	AccountList []string `json:"accountList"`
}

func WebServerStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"Status": "OK"})
}

func JobServer() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(fmt.Sprintf("Job Server panic: %v", err))
			Logger.Error(fmt.Sprintf("Job Server panic: %v", err))
		}
	}()

	for {
		select {
		case accountName := <- JobChan:
			_, err := downLoadAccount(accountName)
			if err != nil {
				resData := fmt.Sprintf("download %s error: %v", accountName, err)
				log.Println(resData)
				Logger.Error(resData)
				msg, _ := json.Marshal(BrokerClientRequest{
					AccountId:            "",
					AccountName:          accountName,
					Position:             "err",
					Order:                "err",
					Trade:                "err",
					ErrMsg:               resData,
				})
				WsHub.broadcast <- msg
			}
		}
	}
}
