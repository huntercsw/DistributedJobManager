package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"log"
	"strings"
	"sync"
	"time"
)

var (
	ClientMap           sync.Map
	ClientStatusMap     map[string]string
	ClientConnMap       = make(map[string]*grpc.ClientConn)
	MasterRouterChannel = make(chan string, 255)
)

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
	return
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

	log.Println("node add", string(k))
	printClientMap()
}

func removeNodeHandler(k string) {
	ClientMap.Delete(k)
	if conn, exist := ClientConnMap[k]; exist {
		delete(ClientConnMap, k)
		conn.Close()
	}

	if _, err := Cli3.Delete(context.TODO(), JobServerStatusPreKey+strings.Split(k, "/")[4]); err != nil {
		Logger.Error(fmt.Sprintf("delete status %s error: %v", JobServerStatusPreKey+strings.Split(k, "/")[4], err))
	}

	log.Println("node delete", k)
	printClientMap()
}

func addNodeStatusHandler(k, v []byte) {
	_k, _ := newKey(string(k))
	if _, exist := ClientMap.Load(_k); exist {
		ClientStatusMap[_k] = string(v)

		if string(v) == JobServerStatusFree {
			MasterRouterChannel <- _k
		}
	}

	log.Println("status add", string(k))
	printClientStatusMap()
}

func removeNodeStatusHandler(k []byte) {
	_k, _ := newKey(string(k))
	delete(ClientStatusMap, _k)
	log.Println("status delete", string(k))
	printClientStatusMap()
}

func modifyNodeStatusHandler(k, v []byte) {
	value := string(v)
	key, _ := newKey(string(k))
	if _, exist := ClientMap.Load(key); exist {
		ClientStatusMap[key] = value
		if value == JobServerStatusFree {
			MasterRouterChannel <- key
		}
		fmt.Println("length: MasterRouterChannel: ", len(MasterRouterChannel))
	}

	log.Println("status modify", string(k), string(v))
	printClientStatusMap()
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
		select {
		case accountName := <-JobChan:
			fmt.Println("JobChannel -> ", accountName)
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
				fmt.Println(accountName, client)
				err := assignJob(client, accountName)
				fmt.Println("result from -->>", client, accountName, err)
			}
		case <-ctx.Done():
			Logger.Info("Master Router was canceled")
			return
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
		JobChan <- accountName
		MasterRouterChannel <- node
		// if error occur, it will be network error ro rpc error
		time.Sleep(time.Second * 5)
	} else {
		switch rsp.ErrMsg {
		case ErrorMessageDoJobOK:
			break
		case ErrorMessageNodeIsBusy:
			JobChan <- accountName
			MasterRouterChannel <- node
		case ErrorMessageJobRunningOnOtherNode:
			MasterRouterChannel <- node
			// TODO: broadcast

		case ErrorMessageUpdateNodeStatusError:
		// TODO: if rsp.ErrMsg == ErrorMessageUpdateNodeStatusError, restart(register cancel and register again) node
		default:
			Logger.Error(fmt.Sprintf("assign job to %s error: %v", node, err))
			log.Println(fmt.Sprintf("assign job to %s error: %v", node, err.Error()))
		}
	}
	fmt.Println(fmt.Sprintf("rsp from %s: %v", node, rsp))
	return err
}
