package room

import (
	"strings"
	"webrtc/p2p-server/pkg/config"
	"webrtc/p2p-server/pkg/logger"
	"webrtc/p2p-server/pkg/msg"
	"webrtc/p2p-server/pkg/utils"
	"webrtc/p2p-server/pkg/ws"

	"github.com/gin-gonic/gin"
)

const (
	JoinRoom       = "joinRoom"       //加入房间
	Offer          = "offer"          //Offer消息
	Answer         = "answer"         //Answer消息
	Candidate      = "candidate"      //Candidate消息
	HangUp         = "hangUp"         //挂断
	LeaveRoom      = "leaveRoom"      //离开房间
	UpdateUserList = "updateUserList" //更新房间用户列表
)

type Room struct {
	//房间ID
	Id string
	//所有用户
	users map[string]*User
	//所有会话
	sessions map[string]*Session
}

func (r *Room) GetUser(id string) *User {
	return r.users[id]
}

func (r *Room) AddUser(u *User) {
	r.users[u.info.Id] = u
}

func (r *Room) RemoveUser(id string) {
	delete(r.users, id)
}

func (r *Room) Exists(id string) bool {
	_, ok := r.users[id]
	return ok
}

func NewRoom(id string) *Room {
	return &Room{
		Id:       id,
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
	}
}

type RoomManager struct {
	rooms map[string]*Room
	cfg   *config.Config
}

func NewRoomManager(cfg *config.Config) *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
		cfg:   cfg,
	}
}

func (rm *RoomManager) AddRoom(id string) *Room {
	rm.rooms[id] = NewRoom(id)
	return rm.rooms[id]
}

func (rm *RoomManager) RemoveRoom(id string) {
	delete(rm.rooms, id)
}

func (rm *RoomManager) GetRoom(id string) *Room {
	return rm.rooms[id]
}

func (rm *RoomManager) Exists(id string) bool {
	_, ok := rm.rooms[id]
	return ok
}

func (rm *RoomManager) HandleMsg(conn *ws.WsConn, c *gin.Context) {
	conn.On("message", func(message []byte) {
		req := utils.Unmarshal(string(message))

		d, ok := req["data"]
		if !ok {
			logger.Log.Errorf("没有data")
			return
		}
		t, ok := req["type"]
		if !ok {
			logger.Log.Errorf("没有type")
			return
		}

		logger.Log.Infof("收到的请求 %v", req)

		dd := d.(map[string]any)
		tt := t.(string)

		switch tt {
		case JoinRoom:
			rm.onJoinRoom(conn, dd)
		case Offer:
			fallthrough
		case Answer:
			fallthrough
		case Candidate:
			rm.onCandidate(conn, dd, req)
		case HangUp:
			rm.onHangUp(conn, dd)
		default:
			logger.Log.Errorf("未知的请求 %v", req)
		}
	})

	conn.On("close", func(message []byte) {
		rm.onClose(conn)
	})
}

func (rm *RoomManager) onJoinRoom(conn *ws.WsConn, data map[string]any) {
	var room *Room
	var user *User

	userId := data["id"].(string)
	userName := data["name"].(string)
	roomId := data["room_id"].(string)

	if !rm.Exists(roomId) {
		room = rm.AddRoom(roomId)
	} else {
		room = rm.GetRoom(roomId)
	}

	if !room.Exists(userId) {
		user = &User{
			info: UserInfo{
				Id:   userId,
				Name: userName,
			},
			conn: conn,
		}
	} else {
		user = room.GetUser(userId)
	}

	//添加用到房间
	room.AddUser(user)

	rm.notifyUsersUpdate(conn, room.users)
}

// 通知所有的用户更新
func (rm *RoomManager) notifyUsersUpdate(conn *ws.WsConn, users map[string]*User) {
	//更新信息
	var infos []UserInfo
	for _, user := range users {
		infos = append(infos, user.info)
	}

	//创建发送消息数据结构
	data := utils.Marshal(msg.Msg{
		Type: UpdateUserList,
		Data: infos,
	})

	//迭代所有的User
	for _, user := range users {
		user.conn.Send(data)
	}
}

func (rm *RoomManager) onCandidate(conn *ws.WsConn, data map[string]any, req map[string]any) {
	to := data["to"].(string)
	roomId := data["room_id"].(string)
	room := rm.GetRoom(roomId)

	if user, ok := room.users[to]; !ok {
		logger.Log.Errorf("用户不存在")
		return
	} else {
		user.conn.Send(utils.Marshal(req))
	}
}

func (rm *RoomManager) onHangUp(conn *ws.WsConn, data map[string]any) {
	sessionID := data["session_id"].(string)
	ids := strings.Split(sessionID, "-")

	roomId := data["room_id"].(string)
	room := rm.GetRoom(roomId)

	//根据Id查找User
	if user, ok := room.users[ids[0]]; !ok {
		logger.Log.Warnf("用户 [" + ids[0] + "] 没有找到")
		return
	} else {
		//发送信息给目标User,即自己[0]
		user.conn.Send(utils.Marshal(msg.Msg{
			Type: HangUp,
			Data: map[string]any{
				//0表示自己 1表示对方
				"to": ids[0],
				//会话Id
				"session_id": sessionID,
			},
		}))
	}

	//根据Id查找User
	if user, ok := room.users[ids[1]]; !ok {
		logger.Log.Warnf("用户 [" + ids[1] + "] 没有找到")
		return
	} else {
		//发送信息给目标User,即对方[1]
		user.conn.Send(utils.Marshal(msg.Msg{
			Type: HangUp,
			Data: map[string]any{
				//0表示自己  1表示对方
				"to": ids[1],
				//会话Id
				"session_id": sessionID,
			},
		}))
	}
}

func (rm *RoomManager) onClose(conn *ws.WsConn) {
	var userId string
	var roomId string

	for _, room := range rm.rooms {
		for _, user := range room.users {
			if user.conn == conn {
				userId = user.info.Id
				roomId = room.Id
				break
			}
		}
	}

	if len(roomId) == 0 {
		logger.Log.Errorf("没有查找到退出的房间")
		return
	}

	room := rm.GetRoom(roomId)

	for _, user := range room.users {
		if user.conn != conn {
			user.conn.Send(utils.Marshal(msg.Msg{
				Type: LeaveRoom,
				Data: userId,
			}))
		}
	}

	room.RemoveUser(userId)

	rm.notifyUsersUpdate(conn, room.users)
}
