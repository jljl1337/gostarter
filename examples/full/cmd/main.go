package main

import (
	"github.com/jljl1337/gostarter/examples/full/internal/server"
)

func main() {
	server := server.MustNewServer()
	server.StartWithGracefulShutdown()
}
