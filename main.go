package main

import (
	"flag"
	"log"

	"github.com/kepnok/bedis/config"
	"github.com/kepnok/bedis/server"
)

func setupFlags() {

	flag.StringVar(&config.Host, "host", "0.0.0.0", "setup host for bedis")
	flag.IntVar(&config.Port, "port", 7379, "setup port for the bedis")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Print("starting the server :-)\n")
	server.RunServer()
	
}