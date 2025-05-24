package config

import (
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
)

type SubscribersConfig struct {
	BufferSize int
}

type MessagesConfig struct {
	BufferSize int
	TimeToLive time.Duration
}

type Config struct {
	Subscriber SubscribersConfig
	Messages   MessagesConfig
}

func LoadConfig() *Config {
	cfg := new(Config)
	_, err := toml.DecodeFile("RedicaMq.toml", &cfg)
	checkError(err)

	fmt.Println("COnfigs: ", *cfg)

	return cfg
}

func checkError(err error) {
	if err != nil {
		log.Fatalln("Error: ", err)
	}
}
