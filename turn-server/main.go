package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"webrtc/turn-server/pkg/config"
	"webrtc/turn-server/pkg/http"
	"webrtc/turn-server/pkg/turn"
)

var CfgFile string

func main() {
	flag.StringVar(&CfgFile, "c", "./config/config.yaml", "config file")

	cfg := config.InitConfig(CfgFile)
	fmt.Println(cfg)

	ts := turn.NewTurnServer(cfg)

	server := http.NewHttpServer(ts, cfg)

	go func() {
		server.Run()
	}()

	// 等待中断信号
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	ts.Close()
}
