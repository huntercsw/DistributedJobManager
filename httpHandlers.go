package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var (
	JobChan             = make(chan string, 1024)
	VMDisableJobChannel = make(chan string, 1024)
)

type Account struct {
	BrokerName string
	Name       string
	Status     [3]bool // status[0]--positionFileStatus; status[1]--orderFileStatus; status[2]--tradeFileStatus
}

type Accounts struct {
	AccountList []string `json:"accountList"`
}

func AccountList(c *gin.Context) {
	if err := Reconfiguration(); err != nil {
		Logger.Error(fmt.Sprintf("reload configuration error: %v", err))
		os.Exit(1)
	}

	if err := createDataDir(); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": "create data dir error"})
		return
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

// prevent different user send same request to the cluster, and the same job do many times
func BrokerClientProcessIsRunning(c *gin.Context) {
	if len(JobChan) > 0 {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": ErrorCodeNodeIsBusy, "Data": "download process is running"})
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

	if _, exist := VMBlackListAccountMap[accounts.AccountList[0]]; exist {
		VMDisableJobChannel <- accounts.AccountList[0]
	} else {
		JobChan <- accounts.AccountList[0]
	}

	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": "waiting for download..."})
}

func Stop(c *gin.Context) {
	jobCount, vmDisableJobCount := len(JobChan), len(VMDisableJobChannel)
	for i := 0; i < jobCount; i++ {
		select {
		case <-JobChan:
		default:
		}
	}

	for i := 0; i < vmDisableJobCount; i++ {
		select {
		case <-VMDisableJobChannel:
		default:
		}
	}

	err := StopAllJobs()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": ErrorCodeStopJobError, "Data": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": ""})
	}
}

func WebServerStatus(c *gin.Context) {
	nodeAddr, nodeTypeCode := c.Query("addr"), c.Query("sysTypeCode")
	if nodeAddr == "" || nodeTypeCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{})
	} else {
		code, _ := strconv.Atoi(nodeTypeCode)
		if code == SystemTypeCodePhysicalMachine {
			PhysicalMachineMap[nodeAddr+":"+JobServerPort] = struct{}{}
		}
		c.JSON(http.StatusOK, gin.H{})
	}
}

func UpdatePyFiles(c *gin.Context) {
	if len(JobChan) > 0 {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "There are jobs in queue, stop all job first"})
		return
	}

	err := StopAllJobs()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": ErrorCodeStopJobError, "Data": err.Error()})
		return
	}

	if err := updatePyFiles(); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": err.Error()})
		return
	}

	if err := updatePngFiles(); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": err.Error()})
		return
	}

	time.Sleep(time.Second)
	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": ""})
}

func responseData() (accounts []Account, err error) {
	var files []os.FileInfo
	outputDir := path.Join(MasterBaseDir, "tmp", time.Now().Format("20060102"))

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

func createDataDir() (err error) {
	dataDir := filepath.Join(MasterBaseDir, "tmp", time.Now().Format("20060102"))
	if _, err = os.Stat(dataDir); os.IsNotExist(err) {
		if err1 := os.Mkdir(dataDir, 0755); err1 != nil {
			Logger.Error(fmt.Sprintf("mkdir %s error: %v", dataDir, err1))
			return err1
		}
		if err1 := os.Chmod(dataDir, 0755); err1 != nil {
			Logger.Error(fmt.Sprintf("chmod dir %s error: %v", dataDir, err1))
			return err1
		}
	} else {
		if err != nil {
			Logger.Error(fmt.Sprintf("create data dir %s, get dir status error: %v", dataDir, err))
			return err
		}
	}
	return nil
}

func updatePyFiles() error {
	newPyFileList, err := ioutil.ReadDir(MasterBaseDir + "pys")
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", MasterBaseDir+"pys", err))
		return err
	}

	if len(newPyFileList) != len(*PyFileMap) {
		if err = PyFileMap.PyFileMapInit(); err != nil {
			return err
		} else {
			return nil
		}
	}

	pyFileMapModified := false
	for _, pyName := range newPyFileList {
		if sha1, exist := (*PyFileMap)[pyName.Name()]; exist {
			newSha1, _ := fileSha1(filepath.Join(MasterBaseDir, "pys", pyName.Name()))
			if sha1 != newSha1 {
				pyFileMapModified = true
				break
			}
		} else {
			pyFileMapModified = true
			break
		}
	}

	if pyFileMapModified {
		if err = PyFileMap.PyFileMapInit(); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func updatePngFiles() error {
	newPngFileList, err := ioutil.ReadDir(MasterBaseDir + "pngs")
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", MasterBaseDir+"pngs", err))
		return err
	}

	if len(newPngFileList) != len(*PngFileMap) {
		if err = PngFileMap.PngFileMapInit(); err != nil {
			return err
		} else {
			return nil
		}
	}

	pngFileMapModified := false
	for _, pngName := range newPngFileList {
		if sha1, exist := (*PngFileMap)[pngName.Name()]; exist {
			newSha1, _ := fileSha1(filepath.Join(MasterBaseDir, "pngs", pngName.Name()))
			if sha1 != newSha1 {
				pngFileMapModified = true
				break
			}
		} else {
			pngFileMapModified = true
			break
		}
	}

	if pngFileMapModified {
		if err = PngFileMap.PngFileMapInit(); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func NodeList(c *gin.Context) {
	type NodeInfo struct {
		Ip         string
		StatusCode string
		IsVM       bool
	}

	nl := make([]NodeInfo, 0)

	for k, clientStatusCode := range ClientStatusMap {
		vm := true
		if _, exist := PhysicalMachineMap[k]; exist {
			vm = false
		}
		n := NodeInfo{Ip: strings.Split(k, ":")[0], StatusCode: clientStatusCode, IsVM: vm}
		nl = append(nl, n)
	}

	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": nl})
}

func MasterRouterChannelLength(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": len(MasterRouterChannel)})
}

func ChangeClientStatus2OffLine(c *gin.Context) {
	nodeAddr := c.Query("addr")
	if nodeAddr == "" {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "node addr is required"})
		return
	}
	conn, exist := ClientConnMap[nodeAddr + ":" + JobServerPort]
	if !exist {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": fmt.Sprintf("%s dose not in ClientConnMap", nodeAddr)})
		return
	}

	if err := UpdateJobServerStatus(JobServerStatusPreKey + nodeAddr + ":" + JobServerPort, JobServerStatusOffLine); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": fmt.Sprintf("update %s to OffLine error: %v", nodeAddr, err)})
		return
	}

	client := NewJobManagerClient(conn)
	if stopRes, err := client.StopJob(context.TODO(), &JobInfo{}); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": fmt.Sprintf("send stop signal to %s error: %v", nodeAddr, err)})
		return
	} else {
		if stopRes.ErrMsg != "" {
			c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": fmt.Sprintf("%s stop job error: %v", nodeAddr, stopRes.ErrMsg)})
			return
		}
	}

	time.Sleep(time.Second)
	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": ""})
}

func ChangeClientStatus2Free(c *gin.Context) {
	nodeAddr := c.Query("addr")
	if nodeAddr == "" {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusBadRequest, "Data": "node addr is required"})
		return
	}

	if err := UpdateJobServerStatus(JobServerStatusPreKey + nodeAddr + ":" + JobServerPort, JobServerStatusFree); err != nil {
		c.JSON(http.StatusOK, gin.H{"ErrorCode": http.StatusInternalServerError, "Data": fmt.Sprintf("update %s to Free error: %v", nodeAddr, err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ErrorCode": 0, "Data": ""})
}
