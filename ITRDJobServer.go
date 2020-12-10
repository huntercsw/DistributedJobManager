package main

import (
	"bytes"
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
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var (
	Cli3                      *clientv3.Client
	MyJobServerStatusKey      string
	UpdateConfig              = false
	RestartRegisterChannel    = make(chan struct{})
	Reporter                  *grpc.ClientConn
	UpdateStatusCount         = 0
	LockErrorCount            = 0
	NodeIpAddr                string
	CurrentAccountName        string
	CurrentBrokerClientCancel context.CancelFunc
	NodeConfigurationVersion = 0
)

type ITRDJobServer struct{}

type JobService struct{}

type JobServerMeta struct {
	Ip   string
	Port string
}

type JobResult struct {
	AccountName string
	ErrCode     int
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
		return err
	}

	if _, err = NewITRDConsoleClient(Reporter).Check(context.TODO(), &CheckRequest{}); err != nil {
		Logger.Error(fmt.Sprintf("%s check to master error: %v", NodeIpAddr, err))
		log.Println(fmt.Sprintf("%s check to master error: %v", NodeIpAddr, err))
		return err
	}

	return nil
}

func (js *JobService) RunJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	if CurrentAccountName != "" {
		return &JobStatus{AccountName: req.AccountName, ErrCode: ErrorCodeNodeIsBusy}, nil
	}

	if !BeAbleToDoJob(req.AccountName) {
		return &JobStatus{AccountName: req.AccountName, ErrCode: ErrorCodeNodeIsVirtualMachine}, nil
	}

	CurrentAccountName = req.AccountName

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
				CurrentAccountName = ""
				return &JobStatus{ErrCode: ErrorCodeUpdateConfigError}, nil
			}
		}
	}

	running, err := jobIsRunning(req.AccountName)
	if err != nil {
		if running {
			CurrentAccountName = ""
			return &JobStatus{
				AccountName: req.AccountName,
				ErrCode:     ErrorCodeJobRunningOnOtherNode,
				ErrMsg:      fmt.Sprintf("%s is running on %s", req.AccountName, err.Error()),
			}, nil
		}
	}

	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusRunning); err != nil {
		errMsg := fmt.Sprintf("run job[%s], modify server status to RUNNING error: %v", req.AccountName, err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		CurrentAccountName = ""
		return &JobStatus{AccountName: req.AccountName, ErrCode: ErrorCodeNodeStatusFree2RunningError}, nil
	}

	go do(req.AccountName)

	return &JobStatus{AccountName: req.AccountName, ErrCode: ErrorCodeDoJobOK}, nil
}

func (js *JobService) StopJob(ctx context.Context, req *JobInfo) (*JobStatus, error) {
	err := closeBrokerClient(CurrentAccountName)
	if err != nil {
		return &JobStatus{
			AccountName: CurrentAccountName,
			ErrCode:     ErrorCodeStopJobError,
			ErrMsg:      fmt.Sprintf("%s stop job %s error: %v", NodeIpAddr, CurrentAccountName, err),
		}, nil
	} else {
		return &JobStatus{
			AccountName: CurrentAccountName,
			ErrCode:     ErrorCodeStopJobOK,
		}, nil
	}
}

func (js *JobService) PullConfig(ctx context.Context, req *PullConfRequest) (*PullConfResponse, error) {
	UpdateConfig = true
	defer func() {UpdateConfig = false}()

	if err := PullFile(Conf.MasterInfo.Ip, MasterHostUser, MasterBaseDir+ConfigFileName, "./"+ConfigFileName); err != nil {
		Logger.Error(fmt.Sprintf("pull configuration[%s] from master[%s] error: %v", MasterBaseDir+ConfigFileName, Conf.MasterInfo.Ip, err))
		return &PullConfResponse{Status: PullConfigStatusFailed}, err
	}

	return &PullConfResponse{Status: PullConfigStatusOK}, nil
}

func (js *JobService) PushFile(ctx context.Context, req *PushFileRequest) (*PushFileResponse, error) {
	files, err := ioutil.ReadDir(TmpDir + time.Now().Format("20060102"))
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", TmpDir+time.Now().Format("20060102"), err))
		return &PushFileResponse{Status: PushFileStatusFailed}, err
	}

	ok := ""
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if err := PushFile(Conf.MasterInfo.Ip, MasterHostUser, file.Name(), req.Dir+""); err != nil {
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
		log.Println(fmt.Sprintf("Job server failed to listen on %s: %v", addr, err))
		os.Exit(1)
	}
	s := grpc.NewServer()
	RegisterJobManagerServer(s, &JobService{})

	reflection.Register(s)

	Logger.Debug(fmt.Sprintf("Job server: Listening and serving %s on %s", network, addr))
	err = s.Serve(listener)
	if err != nil {
		Logger.Error(fmt.Sprintf("ERROR--Job server failed to serve: %v", err))
		log.Println(fmt.Sprintf("ERROR--Job server failed to serve: %v", err))
		os.Exit(1)
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
	cert, err = tls.X509KeyPair([]byte(RegisterCert), []byte(RegisterKey))
	if err != nil {
		return
	}
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM([]byte(RegisterCaCert))

	if Cli3, err = clientv3.New(clientv3.Config{
		Endpoints:   []string{RegisterAddr1},
		DialTimeout: 5 * time.Second,
		TLS:         &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: pool},
	}); err != nil {
		fmt.Println("connect etcd error:", err)
	}
	return
}

func (js *ITRDJobServer) register(ctx context.Context) {
	UpdateStatusCount = 0
	LockErrorCount = 0
	if CurrentAccountName != "" {
		if err := closeBrokerClient(CurrentAccountName); err != nil {
			os.Exit(1)
		}
	}

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
		Logger.Error(fmt.Sprintf("etcd lease grant error 1: %v", err))
		log.Println(fmt.Sprintf("etcd lease grant error 1: %v", err))
		time.Sleep(time.Millisecond * 500)
		if leaseGrantRsp, err = lease.Grant(ctx, leaseTime); err != nil {
			Logger.Error(fmt.Sprintf("etcd lease grant error 2: %v", err))
			log.Println(fmt.Sprintf("etcd lease grant error 2: %v", err))
			os.Exit(1)
		}
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
					Logger.Debug(fmt.Sprintf("RegisterMonitor: node status key %s not found", MyJobServerStatusKey))
					RestartRegisterChannel <- struct{}{}
				} else {
					statusCode := string(rsp.Kvs[0].Value)
					if statusCode != JobServerStatusFree && statusCode != JobServerStatusRunning && statusCode != JobServerStatusSyncPyFiles{
						Logger.Debug(fmt.Sprintf("RegisterMoniter: nodd status value of %s error, value is %s", MyJobServerStatusKey, statusCode))
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

func do(accountName string) {
	var ctx context.Context
	ctx, CurrentBrokerClientCancel = context.WithCancel(context.TODO())
	jobResultChannel := make(chan JobResult, 1)
	defer close(jobResultChannel)

	pyResultChannelLength := len(PyResultChannel)
	for i := 0; i < pyResultChannelLength; i++ {
		select {
		case <-PyResultChannel:
		default:
		}
	}

	err := jobLock(accountName)
	if err != nil {
		jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeJobLockError, Error: err}
	} else {
		CurrentAccountName = accountName
		cmd := exec.CommandContext(ctx, "python", Conf.PyProcessPath, accountName)
		stdOutMsg, stdErrMsg := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
		cmd.Stdout, cmd.Stderr = stdOutMsg, stdErrMsg
		go func() {
			defer func() {
				if err := recover(); err != nil {
					Logger.Error(fmt.Sprintf("do job %s panic: %v", accountName, err))
				}
			}()

			if err := cmd.Start(); err != nil {
				errMsg := fmt.Sprintf("start job %s error: %v", accountName, err)
				Logger.Error(errMsg)
				log.Println(errMsg)
				jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeDoJobError, Error: errors.New(errMsg)}
			} else {
				Logger.Debug(fmt.Sprintf("start to do job %s", accountName))
				log.Println(fmt.Sprintf("start to do job %s", accountName))
			}

			if err := cmd.Wait(); err != nil {
				errMsg := fmt.Sprintf("Error: deal with job %s error: %v", accountName, err)
				Logger.Error(errMsg)
				log.Println(errMsg)
				select {
				case _, ok := <-jobResultChannel:
					if ok {
						jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeDoJobError, Error: errors.New(errMsg)}
					}
				default:
					jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeDoJobError, Error: errors.New(errMsg)}
				}

			} else {
				if stdErrMsg.String() != "" {
					errMsg := fmt.Sprintf("do %s error: %v", accountName, err)
					Logger.Error(errMsg)
					log.Println(errMsg)
					select {
					case _, ok := <-jobResultChannel:
						if ok {
							jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeDoJobError, Error: errors.New(errMsg)}
						}
					default:
						jobResultChannel <- JobResult{AccountName: accountName, ErrCode: ErrorCodeDoJobError, Error: errors.New(errMsg)}
					}
				}
			}
		}()
	}

	report := &ReportRequest{AccountName: accountName, Host: NodeIpAddr}
	pyDone := false
	select {
	case res := <-jobResultChannel:
		switch res.ErrCode {
		case ErrorCodeDoJobOK:
			// impossible
		case ErrorCodeDoJobError:
			if CurrentAccountName == "" {		// job canceled by master
				report.Code = ErrorCodeJobCancelByMaster
			} else {
				report.Code, report.Msg = ErrorCodeDoJobError, res.Error.Error()
			}
		case ErrorCodeJobLockError:
			report.Code = ErrorCodeJobLockError
		}
	case rsp := <-PyResultChannel:
		Logger.Debug(fmt.Sprintf("py response for job %s: %v", accountName, rsp))
		switch {
		case rsp.Position != "ing" && rsp.ErrMsg == "":
			pyDone = true
			if _, err := NewITRDConsoleClient(Reporter).ReportJobResult(context.TODO(), newJobResultRequest(rsp)); err != nil {
				Logger.Error(fmt.Sprintf("[%s] report to master.ReportJobResult error: %v", accountName, err))
				log.Println(fmt.Sprintf("[%s] report to master.ReportJobResult error: %v", accountName, err))
			}

			if err := PushFileByAccount(accountName); err != nil {
				report.Msg = fmt.Sprintf("files of %s upload to master incomplete", accountName)
			}
		case rsp.Position != "ing" && rsp.ErrMsg != "":
			report.Code, report.Msg = ErrorCodeDoJobError, rsp.ErrMsg
		default:
			Logger.Error(fmt.Sprintf("request from py can not be recognized: %v", *rsp))
			report.Code = ErrorCodePyRequestError
		}
	case <-time.After(OneAccountTimeOut):
		_ = closeBrokerClient(accountName)
		report.Code = ErrorCodeJobTimeOut
		Logger.Error(fmt.Sprintf("%s timeout", accountName))
		log.Println(fmt.Sprintf("%s timeout", accountName))
	}

	CurrentAccountName = ""

	if err := UpdateJobServerStatus(MyJobServerStatusKey, JobServerStatusFree); err != nil {
		errMsg := fmt.Sprintf("run job[%s], modify node status to FREE error: %v", accountName, err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		if _, err := NewITRDConsoleClient(Reporter).Report2Master(context.TODO(), &ReportRequest{
			AccountName: accountName,
			Host:        NodeIpAddr,
			Code:        ErrorCodeNodeStatusRunning2FreeError,
		}); err != nil {
			Logger.Error(fmt.Sprintf("[%s] send update status free error message to master error: %v", accountName, err))
			log.Println(fmt.Sprintf("[%s] send update status free error message to master error: %v", accountName, err))
		}
		if UpdateStatusCount > UpdateStatusErrorThreshold {
			RestartRegisterChannel <- struct{}{}
		} else {
			UpdateStatusCount += 1
		}
	}

	if err := jobUnlock(accountName); err != nil {
		errMsg := fmt.Sprintf("job[%s] unlock error: %v", accountName, err)
		Logger.Error(errMsg)
		log.Println(errMsg)
		if _, err := NewITRDConsoleClient(Reporter).Report2Master(context.TODO(), &ReportRequest{
			AccountName: accountName,
			Host:        NodeIpAddr,
			Code:        ErrorCodeJobUnLockError,
		}); err != nil {
			Logger.Error(fmt.Sprintf("send unlock error message to master error: %v", err))
			log.Println(fmt.Sprintf("send unlock error message to master error: %v", err))
		}
		if LockErrorCount > LockErrorThreshold {
			RestartRegisterChannel <- struct{}{}
		} else {
			LockErrorCount += 1
		}
	}

	// report job status to master, should behind SetNodeStatusToFree and Unlock job
	if !pyDone {
		if _, err := NewITRDConsoleClient(Reporter).Report2Master(context.TODO(), report); err != nil {
			Logger.Error(fmt.Sprintf("[%s] report to master error: %v", accountName, err))
			log.Println(fmt.Sprintf("[%s] report to master error: %v", accountName, err))
		}
	}
}

func closeBrokerClient(accountName string) error {
	if accountName == "" {
		return nil
	}

	CurrentBrokerClientCancel()
	defer func() {
		CurrentAccountName = ""
	}()

	brokerClient := Conf.BrokerClientName(accountName)
	if brokerClient == "" {
		Logger.Error(fmt.Sprintf("account name error, no account named %s", accountName))
		log.Println(fmt.Sprintf("account name error, no account named %s", accountName))
		return errors.New(fmt.Sprintf("account name error, no account named %s", accountName))
	}

	if err := killWindowsProcessByName(brokerClient); err != nil {
		Logger.Error(fmt.Sprintf("%s kill broker client error: %v", accountName, err))
		log.Println(fmt.Sprintf("%s kill broker client error: %v", accountName, err))
		return err
	}

	return nil
}

func killWindowsProcessByName(processName string) error {
	pid, err := getPidByName(processName)
	if err != nil {
		Logger.Error(fmt.Sprintf("get pid of %s error: %v", processName, err))
		log.Println(fmt.Sprintf("get pid of %s error: %v", processName, err))
		return err
	} else {
		cmd := exec.Command("taskkill", "-f", "-pid", strconv.Itoa(int(pid)))
		if _, err := cmd.CombinedOutput(); err != nil {
			Logger.Error(fmt.Sprintf("kill process %s error: %v", processName, err))
			log.Println(fmt.Sprintf("kill process %s error: %v", processName, err))
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
		Logger.Error(fmt.Sprintf("get all pids error: %v", err))
		log.Println(fmt.Sprintf("get all pids error: %v", err))
		return processPid, err
	}
	for _, pid := range pids {
		p, err = process.NewProcess(pid)
		if err != nil {
			Logger.Error(fmt.Sprintf("create process struct error: %v", err))
			log.Println(fmt.Sprintf("create process struct error: %v", err))
			continue
		}
		if name, err1 := p.Name(); err1 != nil {
			Logger.Error(fmt.Sprintf("get process name by pid error: %v", err1))
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
		if LockErrorCount > LockErrorThreshold {
			RestartRegisterChannel <- struct{}{}
		} else {
			LockErrorCount += 1
		}
	}
	return err
}

func jobUnlock(accountName string) error {
	_, err := Cli3.Delete(context.TODO(), JobIsRunningOnPreKey+accountName)
	if err != nil {
		Logger.Error(fmt.Sprintf("remove job[%s] lock error: %v", accountName, err))
		log.Println(fmt.Sprintf("remove job[%s] lock error: %v", accountName, err))
		if LockErrorCount > LockErrorThreshold {
			RestartRegisterChannel <- struct{}{}
		} else {
			LockErrorCount += 1
		}
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

func BeAbleToDoJob(accountName string) bool {
	var broker = ""
	for _, brokerObj := range Conf.Brokers {
		if brokerObj.AccountName == accountName {
			broker = brokerObj.Name
			break
		}
	}

	if broker == "" {
		Logger.Error(fmt.Sprintf("account %s not found in config", accountName))
		return false
	}

	for _, b := range Conf.VMBlackList.Brokers {
		if broker == b && IsVirtualMachine {
			return false
		}
	}

	return true
}
