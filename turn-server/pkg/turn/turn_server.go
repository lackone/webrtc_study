package turn

import (
	"fmt"
	"net"
	"webrtc/turn-server/pkg/config"

	"github.com/pion/turn/v4"
)

type TurnServer struct {
	ts          *turn.Server
	cfg         *config.Config
	AuthHandler func(username, realm string, srcAddr net.Addr) (key string, ok bool)
	packet      net.PacketConn
}

func NewTurnServer(cfg *config.Config) *TurnServer {
	if len(cfg.Turn.PublicIp) == 0 {
		panic("turn public ip is empty")
	}

	s := &TurnServer{
		cfg:         cfg,
		AuthHandler: nil,
	}

	packet, err := net.ListenPacket("udp4", fmt.Sprintf("%s:%d", "0.0.0.0", cfg.Turn.Port))
	if err != nil {
		panic(err)
	}

	ts, err := turn.NewServer(turn.ServerConfig{
		Realm:       cfg.Turn.Realm,
		AuthHandler: s.HandleAuth,
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: packet,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(cfg.Turn.PublicIp),
					Address:      "0.0.0.0",
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	s.ts = ts
	s.packet = packet

	return s
}

func (s *TurnServer) HandleAuth(username, realm string, srcAddr net.Addr) (key []byte, ok bool) {
	if s.AuthHandler != nil {
		if pwd, ok := s.AuthHandler(username, realm, srcAddr); ok {
			return turn.GenerateAuthKey(username, realm, pwd), true
		}
	}
	return nil, false
}

func (s *TurnServer) Close() error {
	return s.ts.Close()
}
