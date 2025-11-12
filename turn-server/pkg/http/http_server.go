package http

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"time"
	"webrtc/turn-server/pkg/config"
	"webrtc/turn-server/pkg/turn"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type HttpServer struct {
	*gin.Engine
	cfg    *config.Config
	ts     *turn.TurnServer
	ttlMap *TTLMap
}

func NewHttpServer(ts *turn.TurnServer, cfg *config.Config) *HttpServer {
	if cfg.Http.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// 配置 CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"*", // 允许这个来源
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false, // 如果需要携带 cookie
	}))

	s := &HttpServer{
		Engine: r,
		cfg:    cfg,
		ts:     ts,
		ttlMap: NewTTLMap(),
	}

	ts.AuthHandler = s.AuthHandler

	return s
}

func (s *HttpServer) makePwd(data string, key string) string {
	hash := hmac.New(sha1.New, []byte(key))
	hash.Write([]byte(data))
	return base64.RawStdEncoding.EncodeToString(hash.Sum(nil))
}

func (s *HttpServer) AuthHandler(username string, realm string, srcAddr net.Addr) (key string, ok bool) {
	if value, ok := s.ttlMap.Get(username); ok {
		cred := value.(turn.TurnCreds)
		return cred.Password, true
	}
	return "", false
}

func (s *HttpServer) HandleTurnCreds(c *gin.Context) {
	service := c.Query("service")
	if len(service) == 0 {
		Error(c, "service不能为空", nil)
		return
	}
	username := c.Query("username")
	if len(username) == 0 {
		Error(c, "username不能为空", nil)
		return
	}

	//生成用户名
	timestamp := time.Now().Unix()
	turnUserName := fmt.Sprintf("%d:%s", timestamp, username)
	//生成密码
	turnPassword := s.makePwd(turnUserName, s.cfg.Http.TurnKey)

	ttl := 86400

	cred := turn.TurnCreds{
		Username: turnUserName,
		Password: turnPassword,
		Ttl:      ttl,
		Uris: []string{
			"turn:" + fmt.Sprintf("%s:%d", s.cfg.Turn.PublicIp, s.cfg.Turn.Port) + "?transport=udp",
		},
	}

	s.ttlMap.Set(turnUserName, cred, time.Duration(ttl)*time.Second)

	Success(c, cred)
}

func (s *HttpServer) Run() error {
	s.GET(s.cfg.Http.TurnApiPath, s.HandleTurnCreds)

	err := s.RunTLS(fmt.Sprintf("%s:%d", s.cfg.Http.Ip, s.cfg.Http.Port), s.cfg.Http.Cert, s.cfg.Http.Key)
	if err != nil {
		return err
	}
	return nil
}
