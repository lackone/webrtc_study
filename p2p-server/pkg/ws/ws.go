package ws

import (
	"errors"
	"net"
	"sync/atomic"
	"time"
	"webrtc/p2p-server/pkg/config"
	"webrtc/p2p-server/pkg/logger"
	"webrtc/p2p-server/pkg/msg"
	"webrtc/p2p-server/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type HandleFunc func(ws *WsConn, c *gin.Context)

type WsConn struct {
	*Emitter[[]byte]
	conn     *websocket.Conn
	cfg      *config.Config
	closed   chan struct{}
	isClosed atomic.Bool
	msg      chan []byte
}

func NewWsConn(conn *websocket.Conn, cfg *config.Config) *WsConn {
	wc := &WsConn{
		Emitter: NewEmitter[[]byte](),
		conn:    conn,
		cfg:     cfg,
		closed:  make(chan struct{}),
		msg:     make(chan []byte),
	}

	wc.conn.SetCloseHandler(func(code int, text string) error {
		logger.Log.Warnf("%s %d", text, code)

		wc.Emit("close", []byte(utils.Marshal(msg.Close{
			Code: code,
			Text: text,
		})))

		wc.Close()

		return nil
	})

	return wc
}

func (wc *WsConn) Loop() {
	ticker := time.NewTicker(time.Duration(wc.cfg.Ws.HeartbeatTime) * time.Second)

	go func() {
		for {
			_, data, err := wc.conn.ReadMessage()
			if err != nil {
				logger.Log.Warnf("读取消息错误 %v", err)

				if c, ok := err.(*websocket.CloseError); ok {
					wc.Emit("close", []byte(utils.Marshal(msg.Close{
						Code: c.Code,
						Text: c.Text,
					})))
				} else {
					if c, ok := err.(*net.OpError); ok {
						wc.Emit("close", []byte(utils.Marshal(msg.Close{
							Code: 1008,
							Text: c.Error(),
						})))
					}
				}

				wc.Close()
				break
			}

			//消息放入通道
			wc.msg <- data
		}
	}()

	for {
		select {
		case <-wc.closed:
			return
		case data := <-wc.msg:
			wc.Emit("message", data)
		case <-ticker.C:
			if err := wc.Send(utils.Marshal(msg.Heartbeat{
				Type: "heartbeat",
				Data: "",
			})); err != nil {
				logger.Log.Errorf("发送心跳包错误 %v", err)
				ticker.Stop()
			}
		}
	}
}

func (wc *WsConn) Send(msg string) error {
	select {
	case <-wc.closed:
		return errors.New("closed")
	default:
		return wc.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	}
}

func (wc *WsConn) Close() {
	if !wc.isClosed.Swap(true) {
		wc.conn.Close()

		close(wc.closed)
	}
}
