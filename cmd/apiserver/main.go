package main

import (
	"log"

	"github.com/Vadimkatr/twitterlikewebapp/internal/app/apiserver"
)

func main() {
	if err := apiserver.Start(); err != nil {
		log.Fatal(err)
	}
}
