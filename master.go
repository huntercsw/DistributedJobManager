package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	ClientMap             sync.Map		// key: 192.168.1.191:18199
	ClientStatusMap       map[string]string
	ClientConnMap         = make(map[string]*grpc.ClientConn)
	MasterRouterChannel   = make(chan string, 255)
	PyFileMap             = new(FileMap)
	PngFileMap            = new(FileMap)
	VMBlackListAccountMap map[string]struct{}
	PhysicalMachineMap    = make(map[string]struct{})
)

type FileMap map[string]string // [fileName]: fileSha1

func newKey(s string) (string, error) {
	_s := strings.Split(s, "/")
	if len(_s) != 5 {
		msg := fmt.Sprintf("key error: [%s]", s)
		Logger.Error(msg)
		log.Println(msg)
		return "", errors.New(msg)
	} else {
		return _s[4], nil
	}
}

func dial2JobServer(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		errMsg := fmt.Sprintf("dial to JobServer[%s] error: %v", addr, err)
		Logger.Error(errMsg)
		log.Println(errMsg)
	}
	return conn, err
}

func syncConfig(conn *grpc.ClientConn) error {
	c := NewJobManagerClient(conn)
	if rsp, err := c.PullConfig(context.TODO(), &PullConfRequest{
		RemoteIp:       Conf.MasterInfo.Ip,
		UserName:       "itrd",
		Password:       "123456",
		SourceConfPath: "./itrdConf.xml",
	}); err != nil {
		return err
	} else {
		if rsp.Status != PullConfigStatusOK {
			return errors.New(fmt.Sprintf("sync config error: %d", rsp.Status))
		}
		return nil
	}
}

func MasterInit() (err error) {
	if err = clientMapInit(); err != nil {
		Logger.Error(fmt.Sprintf("clientMapInit error: %v", err))
		log.Println(fmt.Sprintf("clientMapInit error: %v", err))
		return err
	}
	if err = jobServerStatusUpdate(); err != nil {
		return err
	}

	ClientMap.Range(func(k, v interface{}) bool {
		for sk, sv := range ClientStatusMap {
			if sk == k.(string) && sv == JobServerStatusFree {
				MasterRouterChannel <- k.(string)
			}
		}
		return true
	})
	fmt.Println(len(MasterRouterChannel))

	return UnlockAllJob()
}

func clientMapInit() error {
	rsp, err := Cli3.Get(context.TODO(), JobServerMetaDataPreKey, clientv3.WithPrefix())
	if err != nil {
		Logger.Error(fmt.Sprintf("get all job server meta data error: %v", err))
		log.Println(fmt.Sprintf("get all job server meta data error: %v", err))
		return err
	}
	for _, i := range rsp.Kvs {
		v := new(JobServerMeta)
		_ = json.Unmarshal(i.Value, v)
		if _k, err1 := newKey(string(i.Key)); err1 != nil {
			Logger.Error(fmt.Sprintf("format key[%s] error: %v", string(i.Key), err1))
			log.Println(fmt.Sprintf("format key[%s] error: %v", string(i.Key), err1))
			continue
		} else {
			conn, err1 := dial2JobServer(_k)
			if err1 != nil {
				Logger.Error(fmt.Sprintf("connect to %s error: %v", _k, err1))
				log.Println(fmt.Sprintf("connect to %s error: %v", _k, err1))
				continue
			}
			ClientMap.Store(_k, *v)
			ClientConnMap[_k] = conn
		}
	}
	return nil
}

func jobServerStatusUpdate() error {
	rsp, err := Cli3.Get(context.TODO(), JobServerStatusPreKey, clientv3.WithPrefix())
	if err != nil {
		Logger.Error(fmt.Sprintf("get all job server status error: %v", err))
		log.Println(fmt.Sprintf("get all job server status error: %v", err))
		return err
	}
	ClientStatusMap = make(map[string]string)
	for _, i := range rsp.Kvs {
		_k, _ := newKey(string(i.Key))
		ClientStatusMap[_k] = string(i.Value)
	}
	return nil
}

func NodeMonitor(ctx context.Context) {
	nodeMetaChannel := Cli3.Watch(ctx, JobServerMetaDataPreKey, clientv3.WithPrefix())
	for {
		select {
		case watchRsp := <-nodeMetaChannel:
			for _, event := range watchRsp.Events {
				switch {
				case event.IsCreate():
					addNodeHandler(event.Kv.Key, event.Kv.Value)
				case event.IsModify():
					removeNodeHandler(string(event.Kv.Key))
					addNodeHandler(event.Kv.Key, event.Kv.Value)
				default:
					removeNodeHandler(string(event.Kv.Key))
				}

			}
		case <-ctx.Done():
			return
		}
	}
}

func NodeStatusMonitor(ctx context.Context) {
	nodeStatusChannel := Cli3.Watch(ctx, JobServerStatusPreKey, clientv3.WithPrefix())
	for {
		select {
		case watchRsp := <-nodeStatusChannel:
			for _, event := range watchRsp.Events {
				switch {
				case event.IsCreate():
					addNodeStatusHandler(event.Kv.Key, event.Kv.Value)
				case event.IsModify():
					modifyNodeStatusHandler(event.Kv.Key, event.Kv.Value)
				default:
					removeNodeStatusHandler(event.Kv.Key)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func addNodeHandler(k, v []byte) {
	value := &JobServerMeta{}
	if err := json.Unmarshal(v, value); err != nil {
		log.Println("add node handler json unmarshal error:", err)
		return
	}

	if _k, err := newKey(string(k)); err != nil {
		return
	} else {
		conn, err1 := dial2JobServer(_k)
		if err1 != nil {
			return
		}
		ClientMap.Store(_k, *value)
		ClientConnMap[_k] = conn
		addNodeStatusHandler([]byte(JobServerStatusPreKey+_k), []byte(JobServerStatusFree))
	}

	//log.Println("node add", string(k))
	//printClientMap()
}

func removeNodeHandler(k string) {
	ClientMap.Delete(k)
	if _, exist := PhysicalMachineMap[k]; exist {
		delete(PhysicalMachineMap, k)
	}
	if conn, exist := ClientConnMap[k]; exist {
		delete(ClientConnMap, k)
		_ = conn.Close()
	}

	if _, err := Cli3.Delete(context.TODO(), JobServerStatusPreKey+strings.Split(k, "/")[4]); err != nil {
		Logger.Error(fmt.Sprintf("delete status %s error: %v", JobServerStatusPreKey+strings.Split(k, "/")[4], err))
	}

	JobUnlockWhenNodeDead(strings.Split(k, "/")[4])

	//log.Println("node delete", k)
	//printClientMap()
}

func addNodeStatusHandler(k, v []byte) {
	_k, _ := newKey(string(k))
	if _, exist := ClientMap.Load(_k); exist {
		ClientStatusMap[_k] = string(v)

		if string(v) == JobServerStatusFree {
			MasterRouterChannel <- _k
		}
	}

	//log.Println("status add", string(k))
	//printClientStatusMap()
}

func removeNodeStatusHandler(k []byte) {
	_k, _ := newKey(string(k))
	delete(ClientStatusMap, _k)
	//log.Println("status delete", string(k))
	//printClientStatusMap()
}

func modifyNodeStatusHandler(k, v []byte) {
	value := string(v)
	key, _ := newKey(string(k))
	if _, exist := ClientMap.Load(key); exist {
		ClientStatusMap[key] = value
		if value == JobServerStatusFree {
			MasterRouterChannel <- key
		}
	}

	//log.Println("status modify", string(k), string(v))
	//printClientStatusMap()
}

func printClientStatusMap() {
	for k, v := range ClientStatusMap {
		fmt.Println("clientStatusMap: ", k, v)
	}
}

func printClientMap() {
	ClientMap.Range(func(k, v interface{}) bool {
		fmt.Println("ClientMap: ", k, v)
		return true
	})
}

func RunMasterRouter(ctx context.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.Println(fmt.Sprintf("Master Router panic: %v", err))
			Logger.Error(fmt.Sprintf("Master Router panic: %v", err))
		}
	}()

	for {
		var accountName string
		select {
		case vmDisableJob := <- VMDisableJobChannel:
			accountName = vmDisableJob
			Logger.Debug(fmt.Sprintf("get job from VMDisableJobChannel -> %s", accountName))
		case job := <-JobChan:
			accountName = job
			Logger.Debug(fmt.Sprintf("get job from JobChannel -> %s", accountName))
		case <-ctx.Done():
			Logger.Info("Master Router was canceled")
			return
		}

		findFreeClient := false
		var client string
		for {
			client = <-MasterRouterChannel
			if isInClientMap(client) && isInClientStatusMap(client) {
				findFreeClient = true
				break
			}
			// if client not in ClientMap or not in ClientStatusMap, continue to get another client
		}
		if findFreeClient {
			Logger.Debug("length: MasterRouterChannel: ", len(MasterRouterChannel))
			var err error
			if _, exist := PhysicalMachineMap[client]; exist {
				if _, exist := VMBlackListAccountMap[accountName]; exist {
					err = assignJob(client, accountName)
				} else {
					if len(VMDisableJobChannel) > 0 {
						_job := <- VMDisableJobChannel
						JobChan <- accountName
						Logger.Debug(fmt.Sprintf("%s is physical machine, %s is not VmDisable job, get job %s from VmDisableChannel", client, accountName, _job))
						err = assignJob(client, _job)
					} else {
						err = assignJob(client, accountName)
					}
				}
			} else {
				if _, exist := VMBlackListAccountMap[accountName]; exist {
					VMDisableJobChannel <- accountName
					if len(JobChan) > 0 {
						_job := <- JobChan
						Logger.Debug(fmt.Sprintf("%s is virtual machine, %s is VmDisable job, get job %s from JobChannel", client, accountName, _job))
						err = assignJob(client, _job)
					} else {
						MasterRouterChannel <- client
						continue
					}
				} else {
					err = assignJob(client, accountName)
				}
			}
			//err := assignJob(client, accountName)
			if err != nil {
				Logger.Error(fmt.Sprintf("assign job %s to %s error: %v", accountName, client, err))
			} else {
				Logger.Debug(fmt.Sprintf("get node[%s] from MasterRouterChannel to do job[%s]", client, accountName))
			}
		}
	}
}

func isInClientMap(k string) bool {
	if _, exist := ClientMap.Load(k); exist {
		return true
	} else {
		return false
	}
}

func isInClientStatusMap(k string) bool {
	if _, exist := ClientStatusMap[k]; exist {
		return true
	} else {
		return false
	}
}

func assignJob(node, accountName string) error {
	conn, exist := ClientConnMap[node]
	if !exist {
		return errors.New(fmt.Sprintf("%s not in ClientConnMap", node))
	}
	c := NewJobManagerClient(conn)
	rsp, err := c.RunJob(context.TODO(), &JobInfo{
		AccountId:   "",
		AccountName: accountName,
	})
	if err != nil {
		Logger.Error(fmt.Sprintf("run job %s on %s error: %v", accountName, node, err))
		log.Println(fmt.Sprintf("run job %s on %s error: %v", accountName, node, err))
		RouteJob2Channel(accountName)
		MasterRouterChannel <- node
		// if error occur, it will be network error ro rpc error
		time.Sleep(time.Second * 5)
	} else {
		switch rsp.ErrCode {
		case ErrorCodeDoJobOK:
			break
		case ErrorCodeNodeIsBusy:
			Logger.Debug(fmt.Sprintf("send job %s to %s , node is busy", accountName, node))
			RouteJob2Channel(accountName)
		case ErrorCodeNodeIsVirtualMachine:
			Logger.Debug(fmt.Sprintf("%s is virtual machine, can not do job %s", node, accountName))
			RouteJob2Channel(accountName)
			MasterRouterChannel <- node
			// TODO: job will sand to this node always
		case ErrorCodeJobRunningOnOtherNode:
			Logger.Debug(fmt.Sprintf("run job on %s failed: %s", node, rsp.ErrMsg))
			MasterRouterChannel <- node
			rspMsg, _ := json.Marshal(&JobResultRequest{
				AccountName: accountName,
				Position:    "err",
				Order:       "err",
				Trade:       "err",
				ErrMsg:      rsp.ErrMsg,
			})
			WsHub.broadcast <- rspMsg
		case ErrorCodeUpdateConfigError:
			Logger.Error(fmt.Sprintf("%s sync config from master error", node))
			if _, err := Cli3.Delete(context.TODO(), JobServerMetaDataPreKey+node); err != nil {
				Logger.Error(fmt.Sprintf("delete node %s error: %v", node, err))
			}
			RouteJob2Channel(accountName)
		case ErrorCodeNodeStatusFree2RunningError:
			Logger.Error(fmt.Sprintf("node %s update node status to running error", node))
			RouteJob2Channel(accountName)
			if _, exist := ClientStatusMap[node]; exist {
				ClientStatusMap[node] = JobServerStatusFree
			}
			MasterRouterChannel <- node
		default:
			Logger.Error(fmt.Sprintf("assign job[%s] to %s error: %s", rsp.AccountName, node, rsp.ErrMsg))
			log.Println(fmt.Sprintf("assign job[%s] to %s error: %s", rsp.AccountName, node, rsp.ErrMsg))
		}
	}
	fmt.Println(fmt.Sprintf("rsp from %s: %v", node, rsp))
	return err
}

func JobUnlockWhenNodeDead(node string) {
	nodeIp := strings.Split(node, ":")[0]
	rsp, err := Cli3.Get(context.TODO(), JobIsRunningOnPreKey, clientv3.WithPrefix())
	if err != nil {
		Logger.Error(fmt.Sprintf("unlock job for node %s error: %v", nodeIp, err))
		return
	}
	for _, kv := range rsp.Kvs {
		if string(kv.Value) == nodeIp {
			if _, err := Cli3.Delete(context.TODO(), string(kv.Key)); err != nil {
				Logger.Error(fmt.Sprintf("unlock job for node %s, delete key %s error: %v", nodeIp, string(kv.Key), err))
			}
		}
	}
}

func UnlockAllJob() error {
	_, err := Cli3.Delete(context.TODO(), JobIsRunningOnPreKey, clientv3.WithPrefix())
	if err != nil {
		Logger.Error(fmt.Sprintf("unlock all job error: %v", err))
	}
	return err
}

func StopAllJobs() error {
	hasErr := ""
	for node, conn := range ClientConnMap {
		c := NewJobManagerClient(conn)
		if stopRes, err := c.StopJob(context.TODO(), &JobInfo{}); err != nil {
			Logger.Error(fmt.Sprintf("send stop signal to %s error: %v", node, err))
			if _, err = Cli3.Delete(context.TODO(), JobServerMetaDataPreKey+node); err != nil {
				Logger.Error(fmt.Sprintf("delete server [%s] from etcd error: %v", node, err))
				hasErr += fmt.Sprintf("stop job which assigned to %s error", node)
			}
		} else {
			if stopRes.ErrMsg != "" {
				Logger.Error(fmt.Sprintf("%s stop job error: %s", node, stopRes.ErrMsg))
				if _, err = Cli3.Delete(context.TODO(), JobServerMetaDataPreKey+node); err != nil {
					Logger.Error(fmt.Sprintf("delete server [%s] from etcd error: %v", node, err))
					hasErr += fmt.Sprintf("stop job which assigned to %s error", node)
				}
			}
		}
	}

	if hasErr != "" {
		return errors.New(hasErr)
	} else {
		return nil
	}
}

func (fm *FileMap) PyFileMapInit() error {
	*fm = make(map[string]string)
	pyFiles, err := ioutil.ReadDir(MasterBaseDir + "pys")
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", MasterBaseDir+"pys", err))
		return err
	}

	for _, file := range pyFiles {
		if pyFileSha1, err1 := fileSha1(filepath.Join(MasterBaseDir, "pys", file.Name())); err1 != nil {
			Logger.Error(fmt.Sprintf("Sha1 of file %s error: %v", file.Name(), err1))
			return err1
		} else {
			(*fm)[file.Name()] = pyFileSha1
		}
	}

	var pyFileList = make([]string, 0)
	for k, _ := range *fm {
		pyFileList = append(pyFileList, k)
	}
	pyFIleListJson, _ := json.Marshal(&pyFileList)
	if _, err = Cli3.Put(context.TODO(), PyFileNameListKey, string(pyFIleListJson)); err != nil {
		Logger.Error(fmt.Sprintf("put py file list to etcd error: %v", err))
		return err
	}

	return nil
}

func (fm *FileMap) PngFileMapInit() error {
	*fm = make(map[string]string)
	pyFiles, err := ioutil.ReadDir(MasterBaseDir + "pngs")
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", MasterBaseDir+"pngs", err))
		return err
	}

	var pngFileSha1 string
	for _, file := range pyFiles {
		if pngFileSha1, err = fileSha1(filepath.Join(MasterBaseDir, "pngs", file.Name())); err != nil {
			Logger.Error(fmt.Sprintf("Sha1 of file %s error: %v", file.Name(), err))
			return err
		} else {
			(*fm)[file.Name()] = pngFileSha1
		}
	}

	var pngFileList = make([]string, 0)
	for k, _ := range *fm {
		pngFileList = append(pngFileList, k)
	}
	pngFIleListJson, _ := json.Marshal(&pngFileList)
	if _, err = Cli3.Put(context.TODO(), PngFileNameListKey, string(pngFIleListJson)); err != nil {
		Logger.Error(fmt.Sprintf("put png file list to etcd error: %v", err))
		return err
	}

	return nil
}

func RouteJob2Channel(accountName string) {
	if _, exist := VMBlackListAccountMap[accountName]; exist {
		VMDisableJobChannel <- accountName
	} else {
		JobChan <- accountName
	}
}