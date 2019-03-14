package boxwebsock

import (
	"bytes"
	"encoding/json"

	"github.com/okoshiyoshinori/votebox/util"

	"github.com/okoshiyoshinori/votebox/config"
	"github.com/okoshiyoshinori/votebox/logger"

	"github.com/go-redis/redis"
)

var (
	RedisClient = newConn()
)

func newConn() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.APConfig.RedisIp,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	return client
}

func GetConn() *redis.Client {
	return RedisClient
}

func GetSubscribe() *redis.PubSub {
	pubsub := RedisClient.Subscribe("result", "delete", "regist")
	return pubsub
}

func Receive() {
	pubsub := GetSubscribe()
	defer func() {
		pubsub.Close()
		RedisClient.Close()
	}()
	for {
		mes, err := pubsub.ReceiveMessage()
		if err != nil {
			logger.Error.Println(err)
		}
		switch mes.Channel {
		case "result":
			var result interface{}
			buf := bytes.NewReader([]byte(mes.Payload))
			deco := json.NewDecoder(buf)
			if err := deco.Decode(&result); err != nil {
				logger.Error.Println(err)
			}
			n, _ := result.(map[string]interface{})["Payload"].([]interface{})[0].(map[string]interface{})["bid"].(float64)
			key := util.Encode(int(n))
			s := make(map[string][]byte)
			s[key] = []byte(mes.Payload)
			RootHub.BroadCast <- s
		case "delete":
			RootHub.HubUnRegister <- mes.Payload
		case "regist":
			hub := NewHub()
			m := make(map[string]*Hub)
			m[mes.Payload] = hub
			RootHub.HubRegister <- m

		}
	}
}
