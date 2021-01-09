package Utility

import (
	"github.com/sirupsen/logrus"
)

var G_Logger *logrus.Logger

func InitLogger(){
	G_Logger = logrus.New()
	return
}

