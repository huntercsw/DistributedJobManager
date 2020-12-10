package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	sftp2 "github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//func PullFile(remoteIp, user, password, source, dest string) error {
//	sshClient, err := ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
//		User:            user,
//		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {return nil},
//		Auth:            []ssh.AuthMethod{ssh.Password(password)},
//	})
//	if err != nil {
//		Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
//		return err
//	}
//	defer sshClient.Close()
//
//	sftp, err1 := sftp2.NewClient(sshClient)
//	if err1 != nil {
//		Logger.Error(fmt.Sprintf("new sftp client error: %v", err1))
//		return err1
//	}
//	defer sftp.Close()
//
//	sourceFile, err2 := sftp.Open(source)
//	if err2 != nil {
//		Logger.Error(fmt.Sprintf("open source config file[%s] error: %v", source, err2))
//		return err2
//	}
//	defer sourceFile.Close()
//
//	destFile, err3 := os.Create(dest)
//	if err3 != nil {
//		Logger.Error(fmt.Sprintf("create dest config file[%s] error: %v", dest, err3))
//		return err3
//	}
//	defer destFile.Close()
//
//	if _, err = sourceFile.WriteTo(destFile); err != nil {
//		Logger.Error(fmt.Sprintf("write source config to dest file error: %v", err))
//		return err
//	}
//
//	return nil
//}

func PullFile(remoteIp, user, sourceFilePath, destFilePath string) error {
	key, err := ssh.ParsePrivateKey([]byte(PrivateKey))
	if err != nil {
		Logger.Error(fmt.Sprintf("grant private key error: %v", err))
		return err
	}

	var sshClient *ssh.Client
	if sshClient, err = ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}); err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
		return err
	}
	defer sshClient.Close()

	var sftp *sftp2.Client
	if sftp, err = sftp2.NewClient(sshClient); err != nil {
		Logger.Error(fmt.Sprintf("new sftp client error: %v", err))
		return err
	}
	defer sftp.Close()

	//var sourceFile *sftp2.File
	//if sourceFile, err = sftp.Open(sourceFilePath); err != nil {
	//	Logger.Error(fmt.Sprintf("open source config file[%s] error: %v", sourceFilePath, err))
	//	return err
	//}
	//defer sourceFile.Close()
	//
	//var destFile *os.File
	//if destFile, err = os.Create(destFilePath); err != nil {
	//	Logger.Error(fmt.Sprintf("create dest config file[%s] error: %v", destFilePath, err))
	//	return err
	//}
	//defer destFile.Close()
	//
	//if _, err = sourceFile.WriteTo(destFile); err != nil {
	//	Logger.Error(fmt.Sprintf("write source config to dest file error: %v", err))
	//	return err
	//}
	//
	//return nil
	return pullFile(sftp, sourceFilePath, destFilePath)
}

func pullFile (sftpClient *sftp2.Client, sourceFilePath, destFilePath string) (err error) {
	var sourceFile *sftp2.File
	if sourceFile, err = sftpClient.Open(sourceFilePath); err != nil {
		Logger.Error(fmt.Sprintf("open source config file[%s] error: %v", sourceFilePath, err))
		return err
	}
	defer sourceFile.Close()

	var destFile *os.File
	if destFile, err = os.Create(destFilePath); err != nil {
		Logger.Error(fmt.Sprintf("create dest config file[%s] error: %v", destFilePath, err))
		return err
	}
	defer destFile.Close()

	if _, err = sourceFile.WriteTo(destFile); err != nil {
		Logger.Error(fmt.Sprintf("write source config to dest file error: %v", err))
		return err
	}

	return nil
}

func pullFilesByName(files []string) error {
	key, err := ssh.ParsePrivateKey([]byte(PrivateKey))
	if err != nil {
		Logger.Error(fmt.Sprintf("grant private key error: %v", err))
		return err
	}

	var sshClient *ssh.Client
	if sshClient, err = ssh.Dial("tcp", Conf.MasterInfo.Ip+":22", &ssh.ClientConfig{
		User:            MasterHostUser,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}); err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", Conf.MasterInfo.Ip, err))
		return err
	}
	defer sshClient.Close()

	var sftp *sftp2.Client
	if sftp, err = sftp2.NewClient(sshClient); err != nil {
		Logger.Error(fmt.Sprintf("new sftp client error: %v", err))
		return err
	}
	defer sftp.Close()

	for _, fileName := range files {
		var sourceFilePath, destFilePath string
		if strings.HasSuffix(fileName, ".py") {
			sourceFilePath = MasterBaseDir + "pys/" + fileName
			destFilePath = "./" + fileName
		} else {
			sourceFilePath = MasterBaseDir + "pngs/" + fileName
			destFilePath = "./pngs/" + fileName
		}

		if err = pullFile(sftp, sourceFilePath, destFilePath); err != nil {
			Logger.Error(fmt.Sprintf("pull py file %s from master error: %v", sourceFilePath, err))
			return err
		}
	}

	return nil
}

func PullPyFiles() error {
	rsp, err := Cli3.Get(context.TODO(), PyFileNameListKey)
	if err != nil {
		Logger.Error(fmt.Sprintf("get %s error: %v", PyFileNameListKey, err))
		return err
	}

	if rsp.Count != 1 {
		Logger.Error(fmt.Sprintf("number of py file list is not 1: number is %d", rsp.Count))
		return errors.New("number of py file list error")
	}

	var pyFileList = make([]string, rsp.Count)
	if err = json.Unmarshal(rsp.Kvs[0].Value, &pyFileList); err != nil {
		Logger.Error(fmt.Sprintf("josn unmarshal py file list from []byte to slice error: %v", err))
		return err
	}

	return pullFilesByName(pyFileList)
}

func PullPngFiles() error {
	rsp, err := Cli3.Get(context.TODO(), PngFileNameListKey)
	if err != nil {
		Logger.Error(fmt.Sprintf("get %s error: %v", PngFileNameListKey, err))
		return err
	}

	if rsp.Count != 1 {
		Logger.Error(fmt.Sprintf("number of png file list is not 1: number is %d", rsp.Count))
		return errors.New("number of png file list error")
	}

	var pngFileList = make([]string, rsp.Count)
	if err = json.Unmarshal(rsp.Kvs[0].Value, &pngFileList); err != nil {
		Logger.Error(fmt.Sprintf("josn unmarshal png file list from []byte to slice error: %v", err))
		return err
	}

	return pullFilesByName(pngFileList)
}

func PushFile(remoteIp, user, source, dest string) error {
	key, err := ssh.ParsePrivateKey([]byte(PrivateKey))
	if err != nil {
		Logger.Error(fmt.Sprintf("grant private key error: %v", err))
		return err
	}

	var sshClient *ssh.Client
	if sshClient, err = ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
		User:            user,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}); err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
		return err
	}
	defer sshClient.Close()

	//sshClient, err := ssh.Dial("tcp", remoteIp+":22", &ssh.ClientConfig{
	//	User:            user,
	//	HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {return nil},
	//	Auth:            []ssh.AuthMethod{ssh.Password(password)},
	//})
	//if err != nil {
	//	Logger.Error(fmt.Sprintf("ssh to %s error: %v", remoteIp, err))
	//	return err
	//}
	//defer sshClient.Close()

	sftp, err1 := sftp2.NewClient(sshClient)
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

func PushFileByAccount(accountName string) error {
	key, err := ssh.ParsePrivateKey([]byte(PrivateKey))
	if err != nil {
		Logger.Error(fmt.Sprintf("grant private key error: %v", err))
		return err
	}

	var sshClient *ssh.Client
	if sshClient, err = ssh.Dial("tcp", Conf.MasterInfo.Ip+":22", &ssh.ClientConfig{
		User:            MasterHostUser,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil },
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(key)},
	}); err != nil {
		Logger.Error(fmt.Sprintf("ssh to %s error: %v", Conf.MasterInfo.Ip, err))
		return err
	}
	defer sshClient.Close()

	var sftp *sftp2.Client
	sftp, err = sftp2.NewClient(sshClient)
	if err != nil {
		Logger.Error(fmt.Sprintf("new sftp client error: %v", err))
		return err
	}
	defer sftp.Close()

	var files []string
	if files, err = fileListForAccount(accountName); err != nil {
		return err
	}

	var errMsg = ""
	for _, fileName := range files {
		var sourceFileBuf []byte
		if sourceFileBuf, err = ioutil.ReadFile(fileName); err != nil {
			Logger.Error(fmt.Sprintf("open local file[%s] error: %v", fileName, err))
			errMsg += fmt.Sprintf("open local file[%s] error: %v", fileName, err) + "\n"
			continue
		}

		var destFile *sftp2.File
		destFileName := MasterBaseDir + "tmp/" + time.Now().Format("20060102") + "/" + filepath.Base(fileName)
		if destFile, err = sftp.Create(destFileName); err != nil {
			Logger.Error(fmt.Sprintf("sftp create dest file[%s] error: %v", destFileName, err))
			errMsg += fmt.Sprintf("sftp create dest file[%s] error: %v", destFileName, err) + "\n"
			continue
		}

		if _, err = destFile.Write(sourceFileBuf); err != nil {
			Logger.Error(fmt.Sprintf("copy file %s from %s to master error: %v", fileName, NodeIpAddr, err))
			errMsg += fmt.Sprintf("copy file %s from %s to master error: %v", fileName, NodeIpAddr, err) + "\n"
		}
		destFile.Close()
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}
	return nil
}

func fileListForAccount(accountName string) ([]string, error) {
	var res = make([]string, 0)
	DataDir := TmpDir + time.Now().Format("20060102")
	files, err := ioutil.ReadDir(DataDir)
	if err != nil {
		Logger.Error(fmt.Sprintf("open dir[%s] error: %v", DataDir, err))
		return res, err
	}

	filePreList := []string{
		fmt.Sprintf("%s_%s", accountName, time.Now().Format("20060102")),
		fmt.Sprintf("%s_%s", accountName, "order"),
		fmt.Sprintf("%s_%s", accountName, "trade"),
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		for _, filePre := range filePreList {
			if strings.HasPrefix(file.Name(), filePre) {
				res = append(res, filepath.Join(DataDir, file.Name()))
			}
		}
	}
	return res, nil
}
