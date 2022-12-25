package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
)

var AppPath string

type Log struct {
	MaxSize    int `mapstructure:"max_size"`
	MaxAge     int `mapstructure:"max_age"`
	MaxBackups int `mapstructure:"max_backups"`
}

type System struct {
	RuntimePath string `mapstructure:"runtime_path"`
}

type Boot struct {
	Token string `token:"token"`
}

var SystemC System

var LogC Log
var BootC Boot

func InitConfig() {
	path, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	AppPath = path
	fmt.Println("APP PATH:" + AppPath)
	viper.SetConfigFile(path + "/config.toml")
	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal("load config file err:", err)
	}
	err = viper.UnmarshalKey("log", &LogC)
	if err != nil {
		log.Fatal("load config log err:", err)
	}
	err = viper.UnmarshalKey("boot", &BootC)
	if err != nil {
		log.Fatal("load config log err:", err)
	}
}
