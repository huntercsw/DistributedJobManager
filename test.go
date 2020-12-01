package main

//import (
//	"context"
//	"crypto/tls"
//	"crypto/x509"
//	"fmt"
//	"github.com/coreos/etcd/clientv3"
//	"io/ioutil"
//	"time"
//)
//
//func main() {
//	var cert tls.Certificate
//	var cli *clientv3.Client
//	cert, err := tls.LoadX509KeyPair("./certificate/client.pem", "./certificate/client-key.pem")
//	if err != nil {
//		fmt.Println(err)
//	}
//
//	var caData []byte
//	caData, err = ioutil.ReadFile("./certificate/ca.pem")
//	if err != nil {
//		fmt.Println(err)
//	}
//	pool := x509.NewCertPool()
//	pool.AppendCertsFromPEM(caData)
//
//	if cli, err = clientv3.New(clientv3.Config{
//		Endpoints:   []string{"192.168.1.60:2379"},
//		DialTimeout: 5 * time.Second,
//		TLS:         &tls.Config{Certificates: []tls.Certificate{cert}, RootCAs: pool},
//	}); err != nil {
//		fmt.Println("connect etcd error:", err)
//	}
//
//	rsp, err1 := cli.Get(context.TODO(), "/ITRD/JobServer/Status/192.168.1.151:1819900")
//	if err1 != nil {
//		fmt.Println(err1)
//	}
//	fmt.Println(rsp.Count)
//	fmt.Println(rsp.Kvs)
//}
