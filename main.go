package main

import (
	"fmt"
	"github.com/ygxiaobai111/GolixirDB/config"
	util "github.com/ygxiaobai111/GolixirDB/lib/logger"
	"github.com/ygxiaobai111/GolixirDB/resp/handler"
	"github.com/ygxiaobai111/GolixirDB/tcp"
)

func main() {
	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d",
				config.Properties.Bind,
				config.Properties.Port),
		},
		handler.MakeHandler())
	if err != nil {
		util.LogrusObj.Error(err)
	}
}
