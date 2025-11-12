package server

import (
	"fmt"
	"net/http"
	"webrtc/p2p-server/pkg/config"
	"webrtc/p2p-server/pkg/logger"
	"webrtc/p2p-server/pkg/ws"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Server struct {
	*gin.Engine
	cfg       *config.Config
	upgrader  websocket.Upgrader
	handleMsg ws.HandleFunc
}

func NewServer(handleMsg ws.HandleFunc, cfg *config.Config) *Server {
	if cfg.Http.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	s := &Server{
		Engine: r,
		cfg:    cfg,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		handleMsg: handleMsg,
	}

	return s
}

func (s *Server) handlerUpgrade(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Log.Infof("upgrade err: %v", err)
	}

	wsConn := ws.NewWsConn(conn, s.cfg)

	s.handleMsg(wsConn, c)

	wsConn.Loop()
}

func (s *Server) Run() error {
	s.GET(s.cfg.Http.WsPath, s.handlerUpgrade)

	s.StaticFS("/html", http.Dir(s.cfg.Http.HtmlRoot))

	err := s.RunTLS(fmt.Sprintf("%s:%d", s.cfg.Http.Ip, s.cfg.Http.Port), s.cfg.Http.Cert, s.cfg.Http.Key)
	if err != nil {
		return err
	}
	return nil
}
