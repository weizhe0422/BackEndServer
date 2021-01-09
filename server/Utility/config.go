package Utility

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Config struct {
	ConnectMethod      string `json:"connectMethod"`
	ServerAddress      string `json:"serverAddress"`
	SocketPort         int    `json:"socketPort"`
	HttpPort           int    `json:"httpPort"`
	ServerStatusPath   string `json:"serverStatusPath"`
	ReceiveBuffer      int    `json:"receiveBuffer"`
	ApiSvrReadTimeOut  int    `json:"apiSvrReadTimeOut"`
	ApiSvrWriteTimeOut int    `json:"apiSvrWriteTimeOut"`
	RateLimitPerMinute int    `json:"rateLimitPerMinute"`
	RateLimitBuffer    int    `json:"rateLimitBuffer"`
	WebRoot            string `json:"webRoot"`
}

var G_Config *Config

func InitConfig(fileName string) (err error){
	var(
		content []byte
		conf    Config
	)

	if content, err = ioutil.ReadFile(fileName); err != nil {
		return errors.New(fmt.Sprintf("failed to read configuration file: %s: %s", fileName, err))
	}

	if err = json.Unmarshal(content, &conf); err != nil {
		return errors.New(fmt.Sprintf("failed to unmarshall configuration file: %s", err))
	}

	G_Config = &conf
	G_Logger.Info(fmt.Sprintf("Config Info: %v", G_Config))
	return
}