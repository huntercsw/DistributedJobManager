package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/simplifiedchinese"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func fileSha1(filePath string) (string, error) {
	if _, err := os.Stat(filePath); err != nil {
		return "", errors.New(fmt.Sprintf("file %s error: %v", filePath, err))
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", errors.New(fmt.Sprintf("open file %s error %v", filePath, err))
	}
	defer file.Close()

	hashSha1 := sha1.New()

	if f, err := ioutil.ReadFile(filePath); err != nil {
		return "", errors.New(fmt.Sprintf("read file %s error: %v", filePath, err))
	} else {
		hashSha1.Write(f)
		return hex.EncodeToString(hashSha1.Sum([]byte(""))), nil
	}
}

func MasterStatusCheckByHttpRequest () error {
	sysTypeCode := SystemTypeCodePhysicalMachine
	if IsVirtualMachine {
		sysTypeCode = SystemTypeCodeVirtualMachine
	}
	checkUrl := fmt.Sprintf("http://%s:8090/api/webServerStatus/?sysTypeCode=%d&addr=%s", Conf.MasterInfo.Ip, sysTypeCode, NodeIpAddr)
	requestUrl, err := url.Parse(checkUrl)
	if err != nil {
		Logger.Error(fmt.Sprintf("Check Master by Http error, URL %s error: %v", checkUrl, err))
		log.Println(fmt.Sprintf("Check Master by Http error, URL %s error: %v", checkUrl, err))
		return err
	}

	request := requestUrl.Query()
	requestUrl.RawQuery = request.Encode()

	var res *http.Response
	res, err = http.Get(requestUrl.String())
	if err != nil {
		Logger.Error(fmt.Sprintf("http GET to master Check interface error: %v", err))
		log.Println(fmt.Sprintf("http GET to master Check interface error: %v", err))
		return err
	}

	resCode := res.StatusCode
	err = res.Body.Close()
	if err != nil {
		Logger.Error(fmt.Sprintf("http response of master check interface close error: %v", err))
		log.Println(fmt.Sprintf("http response of master check interface close error: %v", err))
		return err
	}

	switch resCode {
	case http.StatusOK:
		err = nil
	case http.StatusBadRequest:
		Logger.Error("request parameters error: addr or sysTypeCode error")
		log.Println("request parameters error: addr or sysTypeCode error")
		os.Exit(1)
	default:
		Logger.Error(fmt.Sprintf("response code of master check interface is not 200, status code is %d", resCode))
		err = errors.New("response code is not 200")
	}

	return err
}

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")	// chinese
)

func ConvertByte2String(byte []byte, charset Charset) string {
	var str string
	switch charset {
	case GB18030:
		var decodeBytes,_=simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str= string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}

	return str
}