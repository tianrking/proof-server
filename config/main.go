package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
)

type Config struct {
	DB       DBConfig       `json:"db"`
	Platform PlatformConfig `json:"platform"`
}

type DBConfig struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	TZ       string `json:"tz"`
}

type PlatformConfig struct {
	Twitter  TwitterPlatformConfig  `json:"twitter"`
	Ethereum EthereumPlatformConfig `json:"ethereum"`
}

type TwitterPlatformConfig struct {
	AccessToken       string `json:"access_token"`
	AccessTokenSecret string `json:"access_token_secret"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
}

type EthereumPlatformConfig struct {
	RPCServer string `json:"rpc_server"`
}

var (
	C *Config = new(Config)
)

func Init(configPath string) {
	if C.DB.Host != "" { // Initialized
		return
	}
	configContent, err := ioutil.ReadFile(configPath)
	if err != nil {
		logrus.Fatalf("Error during opening config file! %v", err)
	}

	err = json.Unmarshal(configContent, C)
	if err != nil {
		logrus.Fatalf("Error duriong unmarshaling config file: %v", err)
	}
}

func GetDatabaseDSN() string {
	template := "host=%s port=%d user=%s password=%s dbname=%s TimeZone=%s sslmode=disable"
	return fmt.Sprintf(template,
		C.DB.Host,
		C.DB.Port,
		C.DB.User,
		C.DB.Password,
		C.DB.DBName,
		C.DB.TZ,
	)
}
