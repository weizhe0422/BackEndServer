package APIServer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/weizhe0422/BackEndServer/server/TCPServer"
	"github.com/weizhe0422/BackEndServer/server/Utility"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

type errHandler func(http.ResponseWriter, *http.Request) error

type userError string

func (e userError) Error() string {
	return e.Message()
}

func (e userError) Message() string {
	return string(e)
}

type ApiServer struct {
	httpSvr  *http.Server
	Method   string
	Address  string
	Port     int
	Listener net.Listener
	StopCh   chan error
}

var G_APIServer *ApiServer

func InitApiServer() {
	G_APIServer = &ApiServer{
		httpSvr: &http.Server{
			ReadTimeout:  time.Duration(Utility.G_Config.ApiSvrReadTimeOut) * time.Millisecond,
			WriteTimeout: time.Duration(Utility.G_Config.ApiSvrWriteTimeOut) * time.Millisecond,
		},
		Method:  Utility.G_Config.ConnectMethod,
		Address: Utility.G_Config.ServerAddress,
		Port:    Utility.G_Config.SocketPort,
		StopCh:  make(chan error),
	}
}

func (a *ApiServer) StartToService() (err error) {
	var (
		listener      net.Listener
		mux           *http.ServeMux
		staticDir     http.Dir
		staticHandler http.Handler
	)
	if listener, err = a.CreateListener(); err != nil {
		return
	}
	a.Listener = listener
	Utility.G_Logger.Info("[API Svr] create API server listener success")

	mux = createHandleFunc(Utility.G_Config.ServerStatusPath, chkServerStatus)
	mux.HandleFunc("/mock", mockExternAPI)

	staticDir = http.Dir(Utility.G_Config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/", http.StripPrefix("/", staticHandler))
	Utility.G_Logger.Info("[API Svr] staticDir", staticDir)

	a.httpSvr.Handler = mux
	Utility.G_Logger.Info("[API Svr] create API server HandleFunc success")

	Utility.G_Logger.Info("[API Svr] start to API server service")
	a.httpSvr.Serve(a.Listener)

	return
}

func (a *ApiServer) CreateListener() (listener net.Listener, err error) {

	if listener, err = net.Listen(Utility.G_Config.ConnectMethod, Utility.G_Config.ServerAddress+":"+strconv.Itoa(Utility.G_Config.HttpPort)); err != nil {
		Utility.G_Logger.Fatal("[API Svr] failed to create a listener:", err.Error())
	}
	Utility.G_Logger.Info("[API Svr]  start " + Utility.G_Config.ConnectMethod + " at " + Utility.G_Config.ServerAddress + ":" + strconv.Itoa(Utility.G_Config.HttpPort))
	return
}

func (a *ApiServer) Stop(reason string) {
	a.StopCh <- errors.New(reason)
}

func createHandleFunc(routerPath string, handlFunc errHandler) (mux *http.ServeMux) {

	mux = http.NewServeMux()
	mux.HandleFunc(routerPath, errWrapper(handlFunc))

	return
}

func chkServerStatus(resp http.ResponseWriter, req *http.Request) (err error) {
	var (
		respSvrStatus *TCPServer.ServerStatus
		respJson      []byte
	)
	Utility.G_Logger.Info("[API Svr] chkServerStatus start")
	respSvrStatus = &TCPServer.ServerStatus{
		ConnCount:    TCPServer.G_TCPServer.GetConnsCount(),
		SessInfoSumm: TCPServer.G_TCPServer.GetServerSummary(),
		ConnHist:     TCPServer.G_TCPServer.GetConnHistALL(),
	}

	resp.Header().Set("Content-Type", "application/json;charset=UTF-8")

	Utility.G_Logger.Info("[API Svr][chkServerStatus] ",respSvrStatus.SessInfoSumm)

	if respSvrStatus == nil {
		return userError("There is no connection history")
	}

	if respJson, err = json.Marshal(respSvrStatus); err != nil {
		return userError("Failed to convert connection history as JSON format")
	}

	Utility.G_Logger.Info("[API Svr][chkServerStatus] send respJson",string(respJson))
	resp.WriteHeader(http.StatusOK)
	resp.Write(respJson)
	return nil
}

func mockExternAPI(resp http.ResponseWriter, req *http.Request) {
	var (
		respMsg string
		reqKeys []string
		ok      bool
	)
	if reqKeys, ok = req.URL.Query()["ReceiveMSG"]; !ok || len(reqKeys) < 1 {
		resp.Write([]byte("can not get valud of ReceiveMSG"))
		return
	}
	respMsg = req.RemoteAddr + ":" + reqKeys[0]

	resp.Write([]byte(respMsg))
}

func errWrapper(handler errHandler) func(http.ResponseWriter, *http.Request) {
	var (
		code int
	)
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			r := recover()
			if r != nil {
				Utility.G_Logger.Info(fmt.Sprintf("[API Svr] %v", r))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()
		err := handler(w, r)
		if userErr, ok := err.(userError); ok {
			http.Error(w, userErr.Message(), http.StatusBadRequest)
			log.Println("user error")
		}
		log.Println(err)
		if err != nil {
			code = http.StatusInternalServerError
			http.Error(w, http.StatusText(code), code)
		}
	}
}
