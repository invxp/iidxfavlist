package main

import (
	"log"

	"github.com/invxp/iidxfavlist/iidxfavlist"
)

const (
	version = "0.0.1-alpha"
)

func main() {
	srv := iidxfavlist.New()

	log.Println("start program version", version)

	srv.Run()
}
