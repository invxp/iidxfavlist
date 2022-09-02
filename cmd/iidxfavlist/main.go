package main

import (
	"github.com/invxp/iidxfavlist/iidxfavlist"
	"log"
)

const (
	version = "0.0.1-alpha"
)

func main() {
	srv, err := iidxfavlist.New()

	if err != nil {
		log.Panic(err)
	}

	log.Println("start program version", version)

	srv.Run()
}
