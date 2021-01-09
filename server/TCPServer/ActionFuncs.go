package TCPServer

import (
	"fmt"
	"github.com/weizhe0422/BackEndServer/server"
	"github.com/weizhe0422/BackEndServer/server/Utility"
	"io"
	"log"
	"net"
	"time"
)

func doReceiveMsg(conn net.Conn) {
	clientIPAddr := conn.RemoteAddr().String()
	G_TCPServer.Connects[clientIPAddr] = 0
	session := server.NewSession(&conn)
	sessionID := session.GetSessionID()

	G_TCPServer.Sessions.Store(sessionID, session)
	Utility.G_Logger.Info(fmt.Sprintf("[TCP Svr] Address(%s): Dial in! Session ID: %s", clientIPAddr, sessionID))

	defer func(){
		conn.Close()
		G_TCPServer.Sessions.Delete(sessionID)
	}()

	readMsgChan := make(chan []byte, 1024)
	writeMsgChan := make(chan *channelData, 1024)
	sessionInfo := &server.SessionInfo{}

	go DoReadMsg(conn, readMsgChan, sessionID, sessionInfo)
	go DoWriteMsg(conn, writeMsgChan)

	for {
		select {
		case reqData := <-readMsgChan:
			if string(reqData) == "bye"{
				return
			}
			passData:=&channelData{
				session: session,
				sessionID:sessionID,
				sessionInfo: *sessionInfo,
				content:reqData,
			}
			writeMsgChan <- passData
		}
	}
}

func DoReadMsg(conn net.Conn, readMsgChan chan []byte, sessionID string, sessionInfo *server.SessionInfo) {
	var(
		msgLength int
		msgBuffer []byte
		clientIPAddr string
		receiveData []byte
		err error
	)

	limiter := make(chan time.Time, Utility.G_Config.RateLimitBuffer)
	for i:=0; i<Utility.G_Config.RateLimitBuffer; i++{
		limiter <- time.Now()
	}

	go func() {
		for t := range time.Tick(time.Minute * time.Duration(Utility.G_Config.RateLimitPerMinute)) {
			limiter <- t
		}
	}()

	for {
		select{
		case _, ok := <-limiter:
			if !ok {
				log.Println("block....")
				continue
			}
		}

		clientIPAddr = conn.RemoteAddr().String()
		msgBuffer = make([]byte, Utility.G_Config.ReceiveBuffer)
		if msgLength, err = conn.Read(msgBuffer); err != nil {
			if err == io.EOF {
				Utility.G_Logger.Info(fmt.Sprintf("[TCP Svr] Address(%s): Close this connection! Sesstion ID: %s", clientIPAddr, sessionID))
				return
			}
			Utility.G_Logger.Error("[TCP Svr] failed to read message: " + err.Error())
			continue
		}

		G_TCPServer.Connects[clientIPAddr]++
		sessionInfo.RemoteAddress = clientIPAddr
		sessionInfo.ReqTime = time.Now()
		receiveData = msgBuffer[:msgLength]
		Utility.G_Logger.Info(fmt.Sprintf("[TCP Svr] Received Msg from %s: %v", clientIPAddr, string(receiveData)))
		sessionInfo.Data = string(receiveData)
		readMsgChan <- receiveData
	}
}

func DoWriteMsg(conn net.Conn, writeMsgChan chan *channelData){
	for {
		select {
		case reqData := <- writeMsgChan:
			reqData.sessionInfo.RespTime = time.Now()
			reqData.sessionInfo.Duration = reqData.sessionInfo.ReqTime.Sub(reqData.sessionInfo.ReqTime).Seconds()
			reqData.session.SetSessionSetting(reqData.sessionID, reqData.sessionInfo)
			clientIPAddr := conn.RemoteAddr().String()
			G_TCPServer.SetConnHist(reqData.sessionID, reqData.sessionInfo)
			G_TCPServer.UpdateServerSummary(reqData.sessionID, G_TCPServer.Connects[clientIPAddr])
		}
	}
}
