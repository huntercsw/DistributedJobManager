package main

import "time"

const (
	WebPort                              = "8090"
	WsPort                               = "8099"
	ConsolePort                          = "8199"
	JobServerConsolePort                 = "8120"
	JobServerPort                        = "18199"
	DeviceD                              = "D:"
	DeviceE                              = "E:"
	DeviceC                              = "C:"
	MasterHostUser                       = "xy"
	MasterBaseDir                        = "/home/xy/ITRD/"
	JobServerMetaDataPreKey              = "/ITRD/JobServer/MetaData/"
	JobServerStatusPreKey                = "/ITRD/JobServer/Status/"
	JobIsRunningOnPreKey                 = "/ITRD/JobIsRunningOn/" // distribute lock, /ITRD/JobIsRunningOn/XXXX: XXX.XXX.XXX.XXX
	MasterComeOnlineKey                   = "/ITRD/Master/Online"
	FileListPreKey                       = "/ITRD/FileList"
	PyFileNameListKey                    = "/ITRD/FileList/py"
	PngFileNameListKey                   = "/ITRD/FileList/png"
	ConfigFileVersionKey                 = "/ITRD/ConfigFile/version"
	ConfigFileName                       = "itrdConf.xml"
	TmpDir                               = "./tmp/"
	JobServerStatusRunning               = "0"
	JobServerStatusFree                  = "1"
	JobServerStatusSyncPyFiles           = "3"
	JobServerStatusPullConfigError       = "500"
	JobServerStatusRestartRegister       = "501"
	PullConfigStatusOK                   = 100
	PullConfigStatusFailed               = 101
	PushFileStatusOK                     = 102
	PushFileStatusFailed                 = 103
	ErrorCodeDoJobOK                     = 200
	ErrorCodeDoJobIng                    = 201
	ErrorCodePyRequestError              = 202
	ErrorCodeDoJobError                  = 400
	ErrorCodeUpdateConfigError           = 401
	ErrorCodeNodeStatusFree2RunningError = 411
	ErrorCodeNodeStatusRunning2FreeError = 412
	ErrorCodeStopJobError                = 403
	ErrorCodeStopJobOK                   = 404
	ErrorCodeNodeIsBusy                  = 405
	ErrorCodeJobRunningOnOtherNode       = 406
	ErrorCodeJobLockError                = 407
	ErrorCodeJobUnLockError              = 408
	ErrorCodeJobTimeOut                  = 409
	ErrorCodeJobCancelByMaster           = 410
	ErrorCodeNodeIsVirtualMachine        = 420
	UpdateStatusErrorThreshold           = 3
	LockErrorThreshold                   = 5
	NodeRegisterRestartInterval          = 3 * time.Second
	JobRetryThreshold                    = 3
	OneAccountTimeOut                    = time.Second * 450
	WindowsVirtualMachinePrimaryKey      = "VMware"
	SystemTypeCodeVirtualMachine         = 1
	SystemTypeCodePhysicalMachine        = 2
)
