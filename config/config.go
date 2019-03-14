package config

import "github.com/BurntSushi/toml"

type Config struct {
	DBserver DBserverConfig
	APserver ApserverConfig
}

type DBserverConfig struct {
	Dbms      string `toml:"dbms"`
	User      string `toml:"user"`
	Pass      string `toml:"pass"`
	Protocol  string `toml:"protocol"`
	DBname    string `toml:"dbname"`
	Charset   string `toml:"charset"`
	Parsetime string `toml:"parsetime"`
}

type ApserverConfig struct {
	Name                     string `toml:"name"`
	Port                     string `toml:"port"`
	PerNum                   int    `toml:"pernum"`
	Salt                     string `toml:"salt"`
	Avatar                   string `toml:"avatar"`
	Item                     string `toml:"item"`
	Ogb                      string `toml:"ogb"`
	TwitterConsumerKey       string `toml:"twitterConsumerKey"`
	TwitterConsumerSecret    string `toml:"twitterConsumerSecret"`
	Log                      string `toml:"log"`
	ErrorLog                 string `toml:"errorlog"`
	RedisIp                  string `toml:"redisip"`
	Origin                   string `toml:"origin"`
	KeyPath                  string `toml:"keypath"`
	TwitterAccessToken       string `toml:"twitterAccessToken"`
	TwitterAccessTokenSecret string `toml:"twitterAccessTokenSecret"`
	CallbackUrl              string `toml:"callbackUrl"`
	TweetBaseUrl             string `toml:"tweetBaseUrl"`
}

var (
	DBConfig *DBserverConfig
	APConfig *ApserverConfig
)

func init() {
	DBConfig = GetDBServerConfig()
	APConfig = GetApServerConfig()
}

func GetDBServerConfig() *DBserverConfig {
	config := Config{}
	_, err := toml.DecodeFile("Config.toml", &config)
	if err != nil {
		panic(err.Error())
	}
	return &config.DBserver
}

func GetApServerConfig() *ApserverConfig {
	config := Config{}
	_, err := toml.DecodeFile("Config.toml", &config)
	if err != nil {
		panic(err.Error())
	}
	return &config.APserver
}
