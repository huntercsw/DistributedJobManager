package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/shirou/gopsutil/process"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"os"
	"os/exec"
	"strconv"
	//"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net"
	"time"
)

const (
	JobServerPort                     = "18199"
	JobServerMetaDataPreKey           = "/ITRD/JobServer/MetaData/"
	JobServerStatusPreKey             = "/ITRD/JobServer/Status/"
	JobIsRunningOnPreKey              = "/ITRD/JobIsRunningOn/" // distribute lock, /ITRD/JobIsRunningOn/XXXX: XXX.XXX.XXX.XXX
	JobServerStatusRunning            = "0"
	JobServerStatusFree               = "1"
	JobServerStatusPullConfigError    = "500"
	JobServerStatusRestartRegister    = "501"
	PullConfigStatusOK                = 0
	PullConfigStatusFailed            = 1
	PushFileStatusOK                  = 0
	PushFileStatusFailed              = 1
	ErrorMessageDoJobOK               = "200"
	ErrorMessageDoJobError            = "400"
	ErrorMessageUpdateNodeStatusError = "401"
	ErrorMessageStopJobError          = "403"
	ErrorMessageStopJobOK             = "404"
	ErrorMessageNodeIsBusy            = "405"
	ErrorMessageJobRunningOnOtherNode = "406"
	UpdateStatusErrorThreshold        = 3
)

var (
	Cli3                        *clientv3.Client
	MyJobServerStatusKey        string
	UpdateConfig                = false
	RestartRegisterChannel      = make(chan struct{})
	Reporter                    *grpc.ClientConn
	UpdateStatusCount           = 0
	NodeBusyBrokerClientRunning = false
	NodeIpAddr                  string
	jobDoneChannel              = make(chan JobResult)
)

type ITRDJobServer struct{}

type JobServerMeta struct {
	Ip   string
	Port string
}

type JobService struct{}

type JobResult struct {
	AccountName string
	Error       error
}

func ReporterInit() error {
	ip, err := externalIP()
	if err != nil {
		Logger.Error(fmt.Sprintf("get host ip error: %v", err))
		log.Println(fmt.Sprintf("get host ip error: %v", err))
		return err
	} else {
		NodeIpAddr = ip.String()
	}

	Reporter, err = grpc.Dial(Conf.MasterInfo.Ip+":"+Conf.MasterInfo.Port, grpc.WithInsecure())
	if err != nil {
		Logger.Error(fmt.Sprintf("dial to master error: %v", err))
		log.Println(fmt.Sprintf("dial to master error: %v", err))
	}
	return err
}

func (js *JobService) RunJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	if NodeBusyBrokerClientRunning {
		return &JobStatus{AccountName: req.AccountName, ErrMsg: ErrorMessageNodeIsBusy}, nil
	}

	defer func() {
		NodeBusyBrokerClientRunning = false
	}()

	if UpdateConfig {
		for i := 0; i < 20; i++ {
			if !UpdateConfig {
				break
			} else {
				log.Println("***update config***")
				time.Sleep(time.Millisecond * 500)
			}
		}
		if UpdateConfig {
			if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusPullConfigError); err != nil {
				errMsg := fmt.Sprintf("run job[%s], modify server status to SYNC_CONF_ERROR error: %v", req.AccountName, err)
				Logger.Error(errMsg)
				log.Println(errMsg)
				RestartRegisterChannel <- struct{}{}
				return &JobStatus{AccountName: req.AccountName, ErrMsg: ErrorMessageUpdateNodeStatusError}, nil
			}
		}
	}

	running, err := jobIsRunning(req.AccountName)
	if err != nil {
		if running {
			return &JobStatus{AccountName: req.AccountName, ErrMsg: ErrorMessageJobRunningOnOtherNode}, nil
		}
	}

	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusRunning); err != nil {
		errMsg := fmt.Sprintf("run job[%s], modify server status to RUNNING error: %v", req.AccountName, err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		return &JobStatus{AccountName: req.AccountName, ErrMsg: ErrorMessageUpdateNodeStatusError}, nil
	}

	NodeBusyBrokerClientRunning = true
	go func() {
		if err := do(req.AccountName); err != nil {
			Logger.Error(fmt.Sprintf("run job[%s] error: %v", req.AccountName, err))
			log.Println(fmt.Sprintf("run job[%s] error: %v", req.AccountName, err))
		}
	}()

	return &JobStatus{AccountName: req.AccountName, ErrMsg: ErrorMessageDoJobOK}, nil
}

func (js *JobService) StopJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	err := closeBrokerClient(req.AccountName)
	if err != nil {
		return &JobStatus{
			AccountId:   req.AccountId,
			AccountName: req.AccountName,
			ErrMsg:      ErrorMessageStopJobError,
		}, err
	} else {
		return &JobStatus{
			AccountId:   req.AccountId,
			AccountName: req.AccountName,
			ErrMsg:      ErrorMessageStopJobOK,
		}, err
	}
}

func (js *JobService) PullConfig(ctx context.Context, req *PullConfRequest) (*PullConfResponse, error) {
	UpdateConfig = true
	log.Println(fmt.Sprintf("---- receive signal to pull config from: %v", *req))
	time.Sleep(time.Second * 2)

	//if err := PullFile(req.RemoteIp, req.UserName, req.Password, req.SourceConfPath, "./"+ConfFileName); err != nil {
	//	Logger.Error(fmt.Sprintf("pull config from %s error: %v", req.RemoteIp, err))
	//	return &PullConfResponse{Status: PullConfigStatusFailed}, err
	//}
	UpdateConfig = false
	return &PullConfResponse{Status: PullConfigStatusOK}, nil
}

func (js *JobService) PushFile(ctx context.Context, req *PushFileRequest) (*PushFileResponse, error) {
	files, err := ioutil.ReadDir(BaseDir + time.Now().Format("20060102"))
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", BaseDir+time.Now().Format("20060102"), err))
		return &PushFileResponse{Status: PushFileStatusFailed}, err
	}

	ok := ""
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := PushFile(req.RemoteIp, req.UserName, req.Password, file.Name(), req.Dir+""); err != nil {
			Logger.Error(fmt.Sprintf("push [%s] to %s error: %v", file.Name(), req.RemoteIp+":"+req.Dir, err))
			ok += file.Name() + " ; "
		}
	}

	if ok == "" {
		return &PushFileResponse{Status: PushFileStatusOK}, nil
	} else {
		return &PushFileResponse{Status: PushFileStatusFailed}, errors.New(fmt.Sprintf("push failed files: %s", ok))
	}
}

func (js *JobService) AliveCheck(ctx context.Context, req *AliveCheckRequest) (*AliveCheckResponse, error) {
	return &AliveCheckResponse{Status: 200}, nil
}

func (js *ITRDJobServer) run() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(fmt.Sprintf("Job server panic: %v", err))
		}
	}()
	network := "tcp"
	addr := "0.0.0.0" + ":" + JobServerPort
	listener, err := net.Listen(network, addr)
	if err != nil {
		Logger.Error(fmt.Sprintf("Job server failed to listen on %s: %v", addr, err))
		return
	}
	s := grpc.NewServer()
	RegisterJobManagerServer(s, &JobService{})

	reflection.Register(s)

	Logger.Debug(fmt.Sprintf("Job server: Listening and serving %s on %s", network, addr))
	err = s.Serve(listener)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR--Job server failed to serve: %v", err))
		return
	}
}

func UpdateJobServerStatus(key, status string) error {
	_, err := Cli3.Put(context.TODO(), key, status)
	if err != nil {
		Logger.Error(fmt.Sprintf("update job server status to etcd error: %v", err))
		log.Println(fmt.Sprintf("update job server status to etcd error: %v", err))
		UpdateStatusCount += 1
		if UpdateStatusCount > UpdateStatusErrorThreshold {
			RestartRegisterChannel <- struct{}{}
		}
	}
	return err
}

func Cli3Init() (err error) {
	var cert tls.Certificate
	cert, err = tls.LoadX509KeyPair("./certificate/client.pem", "./certificate/client-key.pem")
	if err != nil {
		return
	}

	var caData []byte
	caData, err = ioutil.ReadFile("./certificate/ca.pem")
	if err != nil {
		return
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caData)

	if Cli3, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{"192.168.1.60:2379"},
		DialTimeout: 5 * time.Second,
		TLS:         &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: pool},
	}); err != nil {
		fmt.Println("connect etcd error:", err)
	}
	return
}

func (js *ITRDJobServer) register(ctx context.Context) {
	UpdateStatusCount = 0
	closeAllBrokerClient()

	jsm, err := newJobServerMeta()
	if err != nil {
		Logger.Error(fmt.Sprintf("new jobServerMeta error: %v", err))
		log.Println(fmt.Sprintf("new jobServerMeta error: %v", err))
	}

	lease := clientv3.NewLease(Cli3)
	var (
		leaseTime     = int64(3)
		leaseGrantRsp *clientv3.LeaseGrantResponse
	)
	if leaseGrantRsp, err = lease.Grant(ctx, leaseTime); err != nil {
		Logger.Error(fmt.Sprintf("etcd lease grant error: %v", err))
		log.Println(fmt.Sprintf("etcd lease grant error: %v", err))
	}
	if _, err := lease.KeepAlive(ctx, leaseGrantRsp.ID); err != nil {
		Logger.Error(fmt.Sprintf("lease keep alive auto error: %v", err))
		log.Println(fmt.Sprintf("lease keep alive auto error: %v", err))
	}

	key := JobServerMetaDataPreKey + jsm.Ip + ":" + jsm.Port
	value, err1 := json.Marshal(jsm)
	if err1 != nil {
		Logger.Error(fmt.Sprintf("jobServer meta json marshal error: %v", err1))
		log.Println(fmt.Sprintf("jobServer meta json marshal error: %v", err1))
	}
	if _, err = Cli3.Put(context.TODO(), key, string(value), clientv3.WithLease(leaseGrantRsp.ID)); err != nil {
		Logger.Error(fmt.Sprintf("put jobServerMeta to etcd error: %v", err))
		log.Println(fmt.Sprintf("put jobServerMeta to etcd error: %v", err))
	}

	MyJobServerStatusKey = JobServerStatusPreKey + jsm.Ip + ":" + jsm.Port
	if err = UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusFree); err != nil {
		Logger.Error(fmt.Sprintf("put job server status to etcd error: %v", err))
		log.Println(fmt.Sprintf("put job server status to etcd error: %v", err))
	}

	select {
	case <-ctx.Done():
		Logger.Debug("register was canceled")
	}
}

func (js *ITRDJobServer) registerMonitor(ctx context.Context) {
	for {
		time.Sleep(time.Second * 3)
		select {
		case <-ctx.Done():
			return
		default:
			if rsp, err := Cli3.Get(context.TODO(), MyJobServerStatusKey); err != nil {
				Logger.Error(fmt.Sprintf("get %s error: %v", MyJobServerStatusKey, err))
			} else {
				if len(rsp.Kvs) != 1 {
					RestartRegisterChannel <- struct{}{}
				} else {
					if string(rsp.Kvs[0].Value) != JobServerStatusFree && string(rsp.Kvs[0].Value) != JobServerStatusRunning {
						RestartRegisterChannel <- struct{}{}
					}
				}
			}
		}
	}
}

func newJobServerMeta() (*JobServerMeta, error) {
	return &JobServerMeta{Ip: NodeIpAddr, Port: JobServerPort}, nil
}

func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("network disconnected")
}

func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}

func do(accountName string) error {
	ctx, cancel := context.WithCancel(context.TODO())
	jobResChannel := make(chan string)
	defer close(jobResChannel)

	err := jobLock(accountName)
	if err != nil {
		// TODO: lock failed
		return err
	}

	go func(ctx context.Context) {
		log.Println(fmt.Sprintf("do ---------- %s", accountName))
		time.Sleep(time.Second * 30)
		jobResChannel <- fmt.Sprintf("job[%s] done", accountName)
		select {
		case <-ctx.Done():
			return
		}
	}(ctx)

	select {
	case res := <-jobResChannel:
		if res == "" {
			jobDoneChannel <- JobResult{AccountName: accountName, Error: nil}
		} else {
			jobDoneChannel <- JobResult{AccountName: accountName, Error: errors.New(res)}
		}

	case <-time.After(OneAccountTimeOut):
		fmt.Println(fmt.Sprintf("%s timeout", accountName))
		cancel()
	}

	if err = jobUnlock(accountName); err != nil {
		// TODO: unlock failed
		return err
	}

	//c := NewITRDConsoleClient(Reporter)
	//_, err = c.BrokerClient(context.TODO(), &BrokerClientRequest{
	//	AccountId:   "",
	//	AccountName: accountName,
	//	Position:    "test",
	//	Order:       "test",
	//	Trade:       "test",
	//	ErrMsg:      "err-test",
	//})
	//if err != nil {
	//	Logger.Error(fmt.Sprintf("%s rpc to master error: %v", accountName, err))
	//	log.Println(fmt.Sprintf("%s rpc to master error: %v", accountName, err))
	//}
	return nil
}

func closeAllBrokerClient() {
	//TODO
	fmt.Println("close all broker clients")
}

func closeBrokerClient(clientName string) error {
	//TODO
	fmt.Println("close ---- ", clientName)
	return nil
}

func killWindowsProcessByName(processName string) error {
	pid, err := getPidByName(processName)
	if err != nil {
		fmt.Println(fmt.Sprintf("get pid of %s error: %v", processName, err))
		return err
	} else {
		cmd := exec.Command("taskkill", "-f", "-pid", strconv.Itoa(int(pid)))
		if _, err := cmd.CombinedOutput(); err != nil {
			fmt.Println("kill error:", err)
			return err
		}
	}
	return nil
}

func getPidByName(processName string) (int32, error) {
	var p *process.Process
	var processPid = int32(-1)
	pids, err := process.Pids()
	if err != nil {
		log.Println(fmt.Sprintf("get all pids error: %v", err))
		return processPid, err
	}
	for _, pid := range pids {
		p, err = process.NewProcess(pid)
		if err != nil {
			log.Println(fmt.Sprintf("create process struct error: %v", err))
			continue
		}
		if name, err1 := p.Name(); err1 != nil {
			log.Println(fmt.Sprintf("get process name by pid error: %v", err1))
			continue
		} else {
			if name == processName {
				processPid = pid
				break
			}
		}
	}
	if processPid < 0 {
		return processPid, errors.New(fmt.Sprintf("%s dose not exist", processName))
	} else {
		return processPid, nil
	}
}

func jobLock(accountName string) error {
	_, err := Cli3.Put(context.TODO(), JobIsRunningOnPreKey+accountName, NodeIpAddr)
	if err != nil {
		Logger.Error(fmt.Sprintf("add job[%s] lock error: %v", accountName, err))
		log.Println(fmt.Sprintf("add job[%s] lock error: %v", accountName, err))
	}
	return err
}

func jobUnlock(accountName string) error {
	_, err := Cli3.Delete(context.TODO(), JobIsRunningOnPreKey+accountName)
	if err != nil {
		Logger.Error(fmt.Sprintf("remove job[%s] lock error: %v", accountName, err))
		log.Println(fmt.Sprintf("remove job[%s] lock error: %v", accountName, err))
	}
	return err
}

func jobIsRunning(accountName string) (bool, error) {
	rsp, err := Cli3.Get(context.TODO(), JobIsRunningOnPreKey+accountName)
	if err != nil {
		Logger.Error(fmt.Sprintf("get job[%s] lock error: %v", accountName, err))
		log.Println(fmt.Sprintf("get job[%s] lock error: %v", accountName, err))
		return false, err
	}
	if rsp.Count == 0 {
		return false, nil
	} else {
		msg := ""
		for _, kv := range rsp.Kvs {
			msg = string(kv.Value)
		}
		return true, errors.New(fmt.Sprintf("%s is running on %s", accountName, msg))
	}
}

func JobMonitor() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(fmt.Sprintf("JobMonitor Panic: %v", err))
			log.Println(fmt.Sprintf("JobMonitor Panic: %v", err))
			os.Exit(1)
		}
	}()

	for {
		select {
		case res := <-jobDoneChannel:
			if res.Error != nil {
				// TODO
				fmt.Println(res.AccountName, res.Error.Error())
			} else {
				// TODO
				fmt.Println(res.AccountName, "succeed")
			}

			if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusFree); err != nil {
				errMsg := fmt.Sprintf("run job[%s], modify node status to FREE error: %v", res.AccountName, err)
				Logger.Error(errMsg)
				log.Println(errMsg)
			}
		}

	}
}
