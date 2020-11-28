package main

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

const (
	PSM_ASSUME_A_SINGLE_UNIFORM_BLOCK_OF_TEXT = "6"
	PSM_TREAT_THE_IMAGE_AS_A_SINGLE_LINE      = "7"
	OEM_DEFAULT                               = "3"
	DIGIT_ONLY                                = "01234567890"
	TESSERACT_OUTPUT_MODE_STDOUT              = "stdout"
)

type ValidateCodeRecognizer struct {
	imgPath                    string
	PageSegmentationMode       string
	OCREngineMode              string
	TesseractCommandOutputMode string
	TesseditCharWhiteList      string
}

func NewValidateCodeRecognizer(imagePath, psm, oem, tesseditCharWhiteList string) *ValidateCodeRecognizer {
	return &ValidateCodeRecognizer{
		imgPath:                    imagePath,
		PageSegmentationMode:       psm,
		OCREngineMode:              oem,
		TesseractCommandOutputMode: TESSERACT_OUTPUT_MODE_STDOUT,
		TesseditCharWhiteList:      tesseditCharWhiteList,
	}
}

func (recognizer *ValidateCodeRecognizer) ValidateCodeTOString() (validateCode string, err error){
	cmd := exec.Command(
		"tesseract",
		recognizer.imgPath,
		recognizer.TesseractCommandOutputMode,
		"--psm",
		recognizer.PageSegmentationMode,
		"--oem",
		recognizer.OCREngineMode,
		"-c",
		fmt.Sprintf("tessedit_char_whitelist=%s", recognizer.TesseditCharWhiteList),
	)
	stdErrMsg := bytes.NewBuffer(nil)
	cmd.Stderr = stdErrMsg
	if res, err1 := cmd.Output(); err1 != nil {
		return "", errors.New(fmt.Sprintf("tesseract Error: %v \n %s", err1, stdErrMsg.String()))
	} else {
		return filterDigitFromString(string(res)), nil
	}
}

func filterDigitFromString(s string) string {
	reg := regexp.MustCompile(`[0-9]`)
	codes := reg.FindAllString(s, -1)
	return strings.Join(codes, "")
}


