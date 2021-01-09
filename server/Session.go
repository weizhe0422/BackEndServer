package server

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/weizhe0422/BackEndServer/server/Utility"
	"net"
	"time"
)

type SessionInfo struct {
	ReqTime time.Time
	RespTime time.Time
	RemoteAddress string
	Data string
	Duration float64
}

type Session struct {
	sID string
	uID string
	Conn *net.Conn
	settings map[string][]SessionInfo
	MaxReqLimit map[string]string
}

func NewSession(conn *net.Conn) *Session{
	var(
		id uuid.UUID
	)

	id = uuid.New()

	session := &Session{
		sID: id.String(),
		uID: "",
		Conn: conn,
		settings: make(map[string][]SessionInfo, 0),
	}

	return session
}

func (s *Session) SetSessionSetting(key string, value SessionInfo){
	s.settings[key] = append(s.settings[key], value)
}

func (s *Session) GetSessionSetting(key string) interface{}{
	var(
		setting []SessionInfo
		ok bool
	)

	if setting, ok = s.settings[key]; ok {
		return setting
	}

	Utility.G_Logger.Error(fmt.Sprintf("failed to get session(%s) %s info", s.sID, key))
	return nil
}

func (s *Session) GetSessionID() string{
	return s.sID
}

func (s *Session) BindUserID(uid string) {
	s.uID = uid
}

func (s *Session) GetUserID() string{
	return s.uID
}

func (s *Session) SetConnect(conn *net.Conn) {
	s.Conn = conn
}

func (s *Session) GetConnect() *net.Conn{
	return s.Conn
}

