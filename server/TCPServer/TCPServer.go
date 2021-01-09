package TCPServer

import (
	"errors"
	"fmt"
	"github.com/weizhe0422/BackEndServer/server"
	"github.com/weizhe0422/BackEndServer/server/Utility"
	"net"
	"strconv"
	"sync"
)

type TCPServer struct {
	Method       string
	Address      string
	Port         int
	Sessions     *sync.Map
	Listener     net.Listener
	Connects     map[string]int
	SvrStatus    *ServerStatus
	SessInfoSumm map[string]SessionReqInfo
}

var G_TCPServer *TCPServer

func InitTCPServer() {
	G_TCPServer = &TCPServer{
		Method: Utility.G_Config.ConnectMethod,
		Address: Utility.G_Config.ServerAddress,
		Port: Utility.G_Config.SocketPort,
		Sessions: &sync.Map{},
		Connects: make(map[string]int, 0 ),
		SvrStatus: &ServerStatus{
			ConnCount: 0,
			ConnHist: make(map[string][]server.SessionInfo, 0),
		},
		SessInfoSumm: make(map[string]SessionReqInfo, 0),
	}
}

func (t *TCPServer) CreateListener() (err error){
	var(
		listener net.Listener
	)

	if listener, err = net.Listen(t.Method, t.Address+":"+strconv.Itoa(t.Port)); err!=nil{
		return errors.New(fmt.Sprintf("fail to create listener: %v", err))
	}

	t.Listener = listener
	return nil
}

func (t *TCPServer) ListenAndAction(Action func(conn net.Conn)) (err error) {
	var(
		conn net.Conn
	)

	if conn, err = t.Listener.Accept(); err != nil {
		return errors.New(fmt.Sprintf("failed to accept request: %v", err))
	}

	go Action(conn)
	return
}

func (t *TCPServer) StartToService() (err error) {

	if err = t.CreateListener(); err != nil {
		Utility.G_Logger.Error(fmt.Sprintf("[TCP Svr] %v", err))
	}
	Utility.G_Logger.Info(fmt.Sprintf("[TCP Svr] ok to create listener"))

	Utility.G_Logger.Info("[TCP Svr] start to accept and handling requests...")

	for {
		err = t.ListenAndAction(doReceiveMsg)
	}
}

func (t *TCPServer) SetConnHist(sessionID string, data server.SessionInfo) {
	t.SvrStatus.ConnHist[sessionID] = append(t.SvrStatus.ConnHist[sessionID], data)
}

func (t *TCPServer) GetConnHistBySessID (sessionID string) []server.SessionInfo {
	var(
		connHist []server.SessionInfo
		ok bool
	)
	if connHist, ok = t.SvrStatus.ConnHist[sessionID]; ok {
		return connHist
	}
	return nil
}

func (t *TCPServer) GetProcTimeSum(sessionId string) float64 {
	hdlTimeSum := float64(0)
	for _, infoItem := range t.GetConnHistBySessID(sessionId) {
		hdlTimeSum += infoItem.Duration
	}
	return hdlTimeSum
}

func (t *TCPServer) UpdateServerSummary(sessionId string, reqCnt int) {
	sessProcTimeSum := t.GetProcTimeSum(sessionId)
	reqRate := float64(0)
	if sessProcTimeSum != 0 {
		reqRate = float64(reqCnt) / t.GetProcTimeSum(sessionId)
	}

	timePerReq := float64(0)
	if reqCnt != 0 {
		timePerReq = t.GetProcTimeSum(sessionId) / float64(reqCnt)
	}
	t.SessInfoSumm[sessionId] = SessionReqInfo{
		RequestCount: reqCnt,
		RequestRate: reqRate,
		TimePerReq: timePerReq,
	}
}

func (t *TCPServer) GetConnsCount() int {
	var count int
	t.Sessions.Range(func(k, v interface{}) bool {
		count++
		return true
	})
	return count
}

func (t *TCPServer) GetConnHistALL() (connHist map[string][]server.SessionInfo) {
	return t.SvrStatus.ConnHist
}

func (t *TCPServer) GetServerSummary() map[string]SessionReqInfo {
	return t.SessInfoSumm
}