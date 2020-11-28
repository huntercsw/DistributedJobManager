package main

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"

	"fmt"
	"net"
)

var BrokerClientResponseChannel = make(chan *BrokerClientRequest, 255)

type console struct {}

func (console *console) BrokerClient(ctx context.Context, req *BrokerClientRequest) (*ConsoleResponse, error)  {
	var response string
	msg, err := json.Marshal(req)
	if err != nil {
		response = fmt.Sprintf("Type error, message from client[%s] to json error", req.AccountName)
		log.Println(response)
	}

	WsHub.broadcast <- msg

	if req.Position != "ing" {
		log.Println(fmt.Sprintf("%s response: %s", req.AccountName, string(msg)))
		BrokerClientResponseChannel <- req
	}
	return &ConsoleResponse{}, nil
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
