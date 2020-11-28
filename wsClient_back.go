package main
//
//import (
//	"fmt"
//	"github.com/gorilla/websocket"
//	"log"
//	"runtime/debug"
//	"strconv"
//	"sync"
//	"time"
//)
//
//type ClientMap struct {
//	cMap map[string]*Client
//	mutex sync.Mutex
//}
//
//func (cm *ClientMap) add(client *Client) {
//	cm.mutex.Lock()
//	defer cm.mutex.Unlock()
//	if client, ok := cm.cMap[client.ID]; ok {
//		if err := client.close(); err != nil {
//			log.Println(fmt.Sprintf("close client[%s] error: %v", client.ID, err))
//		}
//	}
//	cm.cMap[client.ID] = client
//
//	for k, _ := range cm.cMap {
//		fmt.Println(k)
//	}
//}
//
//func (cm *ClientMap) delete(id string) {
//	cm.mutex.Lock()
//	defer cm.mutex.Unlock()
//	if client, ok := cm.cMap[id]; ok {
//		if err := client.close(); err != nil {
//			log.Println(fmt.Sprintf("close client[%s] error: %v", client.ID, err))
//		}
//		delete(cm.cMap, id)
//	}
//
//	for k, _ := range cm.cMap {
//		fmt.Println(k)
//	}
//}
//
//func (cm *ClientMap) BroadCast(msg []byte) {
//	for _, client := range cm.cMap {
//		client.SendMsg(msg)
//	}
//}
//
//var CM = ClientMap{cMap:make(map[string] *Client), mutex:sync.Mutex{}}
//
//type Client struct {
//	ID                string // remote ip + timestamp
//	Socket            *websocket.Conn
//	MsgSendChannel    chan []byte
//	LastHeartbeatTime uint64
//}
//
//func (c *Client) close() error {
//	if _, closed := <-c.MsgSendChannel; !closed {
//		close(c.MsgSendChannel)
//	}
//	return c.Socket.Close()
//}
//
//func NewClient(remoteAddr string, socket *websocket.Conn) *Client {
//	return &Client{
//		ID:                remoteAddr + "-" + strconv.FormatInt(time.Now().Unix(), 10),
//		Socket:            socket,
//		MsgSendChannel:    make(chan []byte, 500),
//		LastHeartbeatTime: 0,
//	}
//}
//
//func (c *Client) read() {
//	defer func() {
//		if r := recover(); r != nil {
//			log.Println(fmt.Sprintf("ws[%s] read goroutine panic", c.ID), string(debug.Stack()), r)
//			Logger.Error(fmt.Sprintf("ws goroutine panic: %v", r))
//			CM.delete(c.ID)
//		}
//	}()
//
//	for {
//		msgType, message, err := c.Socket.ReadMessage()
//		if err != nil {
//			CM.delete(c.ID)
//			log.Println("socket.read error:", c.ID, msgType, err)
//			return
//		}
//
//		fmt.Println(string(message))
//
//		time.Sleep(time.Millisecond * 100)
//	}
//}
//
//func (c *Client) write() {
//	defer func() {
//		if r := recover(); r != nil {
//			log.Println(fmt.Sprintf("ws[%s] write goroutine panic", c.ID), string(debug.Stack()), r)
//			Logger.Error(fmt.Sprintf("ws goroutine panic: %v", r))
//			CM.delete(c.ID)
//		}
//	}()
//
//	for {
//		select {
//		case message, ok := <-c.MsgSendChannel:
//			if !ok {
//				return
//			}
//			if err := c.Socket.WriteMessage(websocket.TextMessage, message); err != nil {
//				log.Println(fmt.Sprintf("ws[%s] write message error: %v", c.ID, err))
//			}
//		}
//		time.Sleep(time.Millisecond * 100)
//	}
//}
//
//func (c *Client) SendMsg(msg []byte) {
//	c.MsgSendChannel <- msg
//}
//
//
