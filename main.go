package main

import (
	httpInterface "github.com/yoed/yoed-http-interface"
	"github.com/yoed/yoed-handler-transilien/handler"
)

func main() {
	handler := handler.New()
	client := httpInterface.New(handler, &handler.Config.Config.Config)
	client.Run()
}