package main

import "github.com/gin-gonic/gin"


func RouterSetUp(r *gin.Engine) {
	r.GET("/api/accounts", AccountList)
	r.POST("/api/download", DownLoad)
	r.GET("/api/brokerClientRunningCheck", BrokerClientProcessIsRunning)
	r.GET("/api/stop", Stop)
	r.GET("/api/webServerStatus", WebServerStatus)
	r.POST("/api/py/update/files", UpdatePyFiles)
	r.GET("/api/node/list", NodeList)
	r.GET("/api/masterRouterChannelLength", MasterRouterChannelLength)
	r.PUT("/api/node/statusCode/update/offline", ChangeClientStatus2OffLine)
	r.PUT("/api/node/statusCode/update/free", ChangeClientStatus2Free)
}

func WsRouterSetUp(r *gin.Engine) {
	r.GET("/ws/accountStatusReporter", RunWsServer)
}
