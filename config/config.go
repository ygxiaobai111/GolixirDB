package config

import (
	"flag"
	"github.com/spf13/viper"
	"log"
)

// ServerProperties 配置信息
type ServerProperties struct {
	Bind           string
	Port           int
	AppendOnly     bool
	AppendFilename string
	MaxClients     int
	RequirePass    string
	Databases      int
	ClusterMode    bool
	Peers          []string
	Self           string
}

// Properties holds global config properties
var Properties *ServerProperties

func init() {

	confFile := flag.String("cf", "config.yaml", "local configFile")

	flag.Parse()
	// 默认 config
	Properties = &ServerProperties{
		Bind:       "127.0.0.1",
		Port:       14332,
		AppendOnly: false,
	}
	confF := *confFile

	// 设置配置文件的名称和路径
	viper.SetConfigName(confF[:len(confF)-5])
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	// 读取配置文件
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic(err)
	}
	Properties.Port = viper.GetInt("server.port")
	Properties.Bind = viper.GetString("server.bind")
	Properties.Databases = viper.GetInt("server.databases")
	Properties.AppendOnly = viper.GetBool("server.appendOnly")
	Properties.AppendFilename = viper.GetString("server.appendFilename")
	Properties.ClusterMode = viper.GetBool("server.clusterMode")
	Properties.Self = viper.GetString("server.self")
	Properties.Peers = viper.GetStringSlice("server.peers")
}
