package TCPServer

import (
	"github.com/weizhe0422/BackEndServer/server"
)

type channelData struct {
	session     *server.Session
	sessionID   string
	sessionInfo server.SessionInfo
	content     []byte
}

type ServerStatus struct {
	ConnCount    int
	SessInfoSumm map[string]SessionReqInfo
	ConnHist     map[string][]server.SessionInfo
}
type SessionReqInfo struct {
	RequestCount int
	RequestRate  float64
	TimePerReq   float64
}
