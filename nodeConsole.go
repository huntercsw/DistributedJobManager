package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
)

var (
	PyResultChannel = make(chan *PyRequest, 10)
)

type NodeConsole struct {}

func (node *NodeConsole) Py2JobServerConsole(ctx context.Context, req *PyRequest) (*JobServerConsoleResponse, error) {
	if req.Position == "ing" {
		req.ErrMsg = fmt.Sprintf("%s is running on %s", req.AccountName, NodeIpAddr)
		if _, err := NewITRDConsoleClient(Reporter).ReportJobResult(context.TODO(), newJobResultRequest(req)); err != nil {
			Logger.Error(fmt.Sprintf("send py result of job %s to master error: %v", req.AccountName, err))
		}
	} else {
		PyResultChannel <- req
	}
	return &JobServerConsoleResponse{}, nil
}


func (node *NodeConsole) run() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error(fmt.Sprintf("node console panic: %v", err))
		}
	}()
	network := "tcp"
	addr := "127.0.0.1:" + JobServerConsolePort
	listener, err := net.Listen(network, addr)
	if err != nil {
		Logger.Error(fmt.Sprintf("node console failed to listen on %s: %v", addr, err))
		log.Println(fmt.Sprintf("node console failed to listen on %s: %v", addr, err))
		os.Exit(1)
	}
	s := grpc.NewServer()
	RegisterITRDNodeConsoleServer(s, &NodeConsole{})

	reflection.Register(s)

	Logger.Debug(fmt.Sprintf("Node Console: Listening and serving on %s %s", network, addr))
	err = s.Serve(listener)
	if err != nil {
		Logger.Error(fmt.Sprintf("ERROR--Node Console failed to serve: %v", err))
		log.Println(fmt.Sprintf("ERROR--Node Console failed to serve: %v", err))
		os.Exit(1)
	}
}

func newJobResultRequest(p *PyRequest) *JobResultRequest {
	return &JobResultRequest{
		AccountId:            p.AccountId,
		AccountName:          p.AccountName,
		Position:             p.Position,
		Order:                p.Order,
		Trade:                p.Trade,
		ErrMsg:               p.ErrMsg,
	}
}