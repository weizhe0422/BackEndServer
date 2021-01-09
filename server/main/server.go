package main

import (
	"flag"
	"fmt"
	"github.com/weizhe0422/BackEndServer/server/TCPServer"
	"github.com/weizhe0422/BackEndServer/server/APIServer"
	"github.com/weizhe0422/BackEndServer/server/Utility"
	"os"
	"os/signal"
)

var (
	cfgFilePath string
)

func initArgs(){
	flag.StringVar(&cfgFilePath, "config", "./server.json", "configuration file path")
	flag.Parse()
}

func main()  {
	Utility.InitLogger()
	initArgs()
	if err := Utility.InitConfig(cfgFilePath); err != nil {
		Utility.G_Logger.Error("[INIT STEP]" + err.Error())
		return
	}
	Utility.G_Logger.Info("[INIT STEP] ok to load configuration setting." )

	TCPServer.InitTCPServer()
	Utility.G_Logger.Info("[TCP Svr] ok to initialize." )

	APIServer.InitApiServer()
	Utility.G_Logger.Info("[API Svr] ok to initialize." )

	go WaitAndWaitTerminal()

	go APIServer.G_APIServer.StartToService()
	TCPServer.G_TCPServer.StartToService()


}

func WaitAndWaitTerminal(){
	Utility.G_Logger.Info("[] WaitAndWaitTerminal...")
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	<- c

	Utility.G_Logger.Info("Start to stop TCP Server...")
	if err := TCPServer.G_TCPServer.Listener.Close(); err != nil {
		Utility.G_Logger.Error(fmt.Sprintf("[TCP Svr] failed to stop TCP Server: %v", err))
	}
	Utility.G_Logger.Info("[TCP Svr] ok to stop")

	Utility.G_Logger.Info("Start to stop API Server...")
	if err := APIServer.G_APIServer.Listener.Close(); err != nil {
		Utility.G_Logger.Error(fmt.Sprintf("[API Svr] failed to stop TCP Server: %v", err))
	}
	Utility.G_Logger.Info("[API Svr] ok to stop")

	os.Exit(-1)
}