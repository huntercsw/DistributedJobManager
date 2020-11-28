package main

import (
	"fmt"
	sftp2 "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
)

const (
	ConfFileName = "itrdConf.xml"
	BaseDir = "./tmp/"
)

func PullFile(remoteIp, user, password, source, dest string) error {
	ssh, err := ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {return nil},
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
	})
	if err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
		return err
	}
	defer ssh.Close()

	sftp, err1 := sftp2.NewClient(ssh)
	if err1 != nil {
		Logger.Error(fmt.Sprintf("new sftp client error: %v", err1))
		return err1
	}
	defer sftp.Close()

	sourceFile, err2 := sftp.Open(source)
	if err2 != nil {
		Logger.Error(fmt.Sprintf("open source config file[%s] error: %v", source, err2))
		return err2
	}
	defer sourceFile.Close()

	destFile, err3 := os.Create(dest)
	if err3 != nil {
		Logger.Error(fmt.Sprintf("create dest config file[%s] error: %v", dest, err3))
		return err3
	}
	defer destFile.Close()

	if _, err = sourceFile.WriteTo(destFile); err != nil {
		Logger.Error(fmt.Sprintf("write source config to dest file error: %v", err))
		return err
	}

	return nil
}

func PushFile(remoteIp, user, password, source, dest string) error {
	ssh, err := ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {return nil},
		Auth:            []ssh.AuthMethod{ssh.Password(password)},
	})
	if err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
		return err
	}
	defer ssh.Close()

	sftp, err1 := sftp2.NewClient(ssh)
	if err1 != nil {
		Logger.Error(fmt.Sprintf("new sftp client error: %v", err1))
		return err1
	}
	defer sftp.Close()

	sourceFileBuf, err2 := ioutil.ReadFile(source)
	if err2 != nil {
		Logger.Error(fmt.Sprintf("open local file[%s] error: %v", source, err2))
		return err2
	}

	destFile, err3 := sftp.Create(dest)
	if err3 != nil {
		Logger.Error(fmt.Sprintf("sftp create dest file[%s] error: %v", dest, err3))
		return err3
	}

	if _, err = destFile.Write(sourceFileBuf); err != nil {
		Logger.Error(fmt.Sprintf("write local file to dest host error: %v", err))
		return err
	}
	return nil
}





