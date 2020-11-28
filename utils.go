package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
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


func FileHasUpdate(path string) (bool, error) {
	return true, nil
}