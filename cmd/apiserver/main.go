package main

import (
	"log"
	"flag"

	"github.com/BurntSushi/toml"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/apiserver"
)

var (
	configPath string
)

func init() {
	flag.StringVar(
		&configPath, 
		"config-path", 
		"configs/apiserver.toml", 
		"path to config file",
	)
}

func main() {
	config := apiserver.NewConfig()
	_, err := toml.DecodeFile(configPath, config)
	if err != nil {
		log.Fatal(err)
	}
	if err := apiserver.Start(config); err != nil {
		log.Fatal(err)
	}
}
