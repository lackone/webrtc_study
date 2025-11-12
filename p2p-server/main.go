package main

import (
	"flag"
	"fmt"
	"webrtc/p2p-server/pkg/config"
	"webrtc/p2p-server/pkg/logger"
	"webrtc/p2p-server/pkg/room"
	"webrtc/p2p-server/pkg/server"
)

var CfgFile string

func main() {
	flag.StringVar(&CfgFile, "c", "./config/config.yaml", "config file")

	cfg := config.InitConfig(CfgFile)
	fmt.Println(cfg)

	logger.InitLogger(cfg)

	rm := room.NewRoomManager(cfg)

	server := server.NewServer(rm.HandleMsg, cfg)

	server.Run()
}
