package room

import (
	"webrtc/p2p-server/pkg/ws"
)

type UserInfo struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type User struct {
	info UserInfo
	conn *ws.WsConn
}

type Session struct {
	Id   string
	from User
	to   User
}
