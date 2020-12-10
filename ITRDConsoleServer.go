package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

var (
	JobRetryTimesMap = make(map[string]int32)
)

type console struct{}

func (console *console) ReportJobResult(ctx context.Context, req *JobResultRequest) (*JobResultResponse, error) {
	msg, err := json.Marshal(req)
	if err != nil {
		Logger.Error(fmt.Sprintf("Type error, result of job %s to json error", req.AccountName))
		WsHub.broadcast <- []byte(fmt.Sprintf("Type error, result of job %s to json error", req.AccountName))
	}

	if req.Position != "ing" {
		log.Println(fmt.Sprintf(">>>>>>>>>>>>>>>> %s done: %v", req.AccountName, *req))
		_ = jobRetryTimesReset(req.AccountName)
	}

	WsHub.broadcast <- msg

	return &JobResultResponse{}, nil
}

func (console *console) Report2Master(ctx context.Context, req *ReportRequest) (*ReportResponse, error) {
	fmt.Println("Report2Master-->>", *req)
	switch req.Code {
	case ErrorCodePyRequestError:
		log.Println(fmt.Sprintf("!!!! %s", req.Msg))
		_ = jobRetryTimesReset(req.AccountName)
		broadcastMsg, _ := json.Marshal(req)
		WsHub.broadcast <- broadcastMsg
	case ErrorCodeDoJobOK:
		log.Println(fmt.Sprintf(">>>>>>>>>>>>>>>> %s done by %s: %s", req.AccountName, req.Host, req.Msg))
		_ = jobRetryTimesReset(req.AccountName)
		//broadcastMsg, _ := json.Marshal(req)
		//WsHub.broadcast <- broadcastMsg
	case ErrorCodeJobCancelByMaster:
		// TODO
	case ErrorCodeDoJobError:
		Logger.Error(fmt.Sprintf("job %s on %s error: %s", req.AccountName, req.Host, req.Msg))
		err := jobRetryTimesPlus(req.AccountName)
		if err != nil {
			Logger.Error(fmt.Sprintf("job %s retry over %d times, latest try on %s", req.AccountName, JobRetryThreshold, req.Host))
			_ = jobRetryTimesReset(req.AccountName)

			broadcastMsg, _ := json.Marshal(&JobResultRequest{
				AccountName: req.AccountName,
				ErrMsg:      fmt.Sprintf("%s and retry over %d times, latest try on %s, error: %s", req.AccountName, JobRetryThreshold, req.Host, req.Msg),
			})
			WsHub.broadcast <- broadcastMsg
		} else {
			errMsg, _ := json.Marshal(&JobResultRequest{
				AccountName: req.AccountName,
				Position: "retry",
				ErrMsg: fmt.Sprintf("do %s error on %s, %s is waiting for retry... ...", req.AccountName, req.Host, req.AccountName),
			})
			WsHub.broadcast <- errMsg
			RouteJob2Channel(req.AccountName)
		}
	case ErrorCodeJobLockError:
		Logger.Error(fmt.Sprintf("job %s on %s add lock error: %s", req.AccountName, req.Host, req.Msg))
		// error on lock, means job has not run
		RouteJob2Channel(req.AccountName)
	case ErrorCodeJobUnLockError:
		if err := jobUnlock(req.AccountName); err != nil {
			Logger.Error(fmt.Sprintf("master unlock %s error: %v", req.AccountName, err))
		}
		// report of unlock error and report of job status are in different report
		if _, err := Cli3.Delete(context.TODO(), JobIsRunningOnPreKey+req.AccountName); err != nil {
			Logger.Error(fmt.Sprintf("master unlock job %s error: %v", req.AccountName, err))
		}
	case ErrorCodeJobTimeOut:
		Logger.Error(fmt.Sprintf("job %s on %s timeout: %s", req.AccountName, req.Host, req.Msg))
		err := jobRetryTimesPlus(req.AccountName)
		if err != nil {
			Logger.Error(fmt.Sprintf("job %s retry over %d times, latest try on %s", req.AccountName, JobRetryThreshold, req.Host))
			_ = jobRetryTimesReset(req.AccountName)

			broadcastMsg, _ := json.Marshal(&JobResultRequest{
				AccountName: req.AccountName,
				ErrMsg:      fmt.Sprintf("job %s retry over %d times, latest try on %s, error: %s", req.AccountName, JobRetryThreshold, req.Host, "timeout"),
			})
			WsHub.broadcast <- broadcastMsg
		} else {
			errMsg, _ := json.Marshal(&JobResultRequest{
				AccountName: req.AccountName,
				Position: "retry",
				ErrMsg: fmt.Sprintf("do %s timeout on %s, %s is waiting for retry... ...", req.AccountName, req.Host, req.AccountName),
			})
			WsHub.broadcast <- errMsg
			RouteJob2Channel(req.AccountName)
		}
	case ErrorCodeStopJobOK:
		break
	case ErrorCodeStopJobError:
		// TODO
	default:
		// TODO
	}
	return &ReportResponse{}, nil
}

func (console *console) Check(ctx context.Context, req *CheckRequest) (*CheckResponse, error) {
	return &CheckResponse{}, nil
}

func RunConsoleServer() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(fmt.Sprintf("rpc server panic: %v", err))
		}
	}()
	network := "tcp"
	addr := "0.0.0.0:" + ConsolePort
	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Println(fmt.Sprintf("console server failed to listen: %v", err))
		return
	}
	s := grpc.NewServer()
	RegisterITRDConsoleServer(s, &console{})

	reflection.Register(s)

	log.Println(fmt.Sprintf("Listening and serving %s on %s", network, addr))
	err = s.Serve(listener)
	if err != nil {
		log.Println(fmt.Sprintf("ERROR--console server failed to serve: %v", err))
		return
	}
}

func jobRetryOverThreshold(accountName string) bool {
	retryTimes, exist := JobRetryTimesMap[accountName]
	if !exist {
		JobRetryTimesMap[accountName] = 0
		return false
	}
	if retryTimes >= JobRetryThreshold {
		return true
	}
	return false
}

func jobRetryTimesPlus(accountName string) error {
	retryTimes, exist := JobRetryTimesMap[accountName]
	if exist {
		if retryTimes+1 > JobRetryThreshold {
			return errors.New(fmt.Sprintf("%s out of retry threshold", accountName))
		}
		JobRetryTimesMap[accountName] = retryTimes + 1
	} else {
		JobRetryTimesMap[accountName] = 1
	}
	return nil
}

func jobRetryTimesReset(accountName string) error {
	_, exist := JobRetryTimesMap[accountName]
	if !exist {
		return errors.New(fmt.Sprintf("%s is not in job retry times map", accountName))
	} else {
		JobRetryTimesMap[accountName] = 0
		return nil
	}
}
