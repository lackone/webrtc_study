package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type Config struct {
	Http HttpConfig
	Log  LogConfig
	Ws   WsConfig
}

type HttpConfig struct {
	Ip       string `mapstructure:"ip"`
	Port     int    `mapstructure:"port"`
	Mode     string `mapstructure:"mode"`
	Key      string `mapstructure:"key"`
	Cert     string `mapstructure:"cert"`
	HtmlRoot string `mapstructure:"html_root"`
	WsPath   string `mapstructure:"ws_path"`
}

type LogConfig struct {
	Path string `mapstructure:"path"`
}

type WsConfig struct {
	HeartbeatTime int `mapstructure:"heartbeat_time"`
}

var conf Config

func GetConfig() *Config {
	return &conf
}

func InitConfig(cfgFile string) *Config {
	viper.SetConfigType("yaml")

	if len(cfgFile) == 0 {
		viper.SetConfigFile("./config/config.yaml")
	} else {
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	if err := viper.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	go func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(in fsnotify.Event) {
			if err := viper.Unmarshal(&conf); err != nil {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
		})
	}()

	return &conf
}
