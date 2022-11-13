package main

import "connor/go/web"

func main() {
	server := web.NewHttpServer()

	server.Start(":8080")
}
