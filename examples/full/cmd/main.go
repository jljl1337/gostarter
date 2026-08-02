package main

import (
	"flag"

	"github.com/jljl1337/gostarter/examples/full/internal/server"
)

func main() {
	envFile := flag.String("env", ".env", "Path to the .env file")
	flag.Parse()

	server := server.MustNewServer(*envFile)
	server.StartWithGracefulShutdown()
}
