package main

import (
	"github.com/CzarSimon/httputil/logger"
	"go.uber.org/zap"
)

var log = logger.GetDefaultLogger("turn-server/main")

func main() {
	e := setupEnv()
	defer e.close()
	err := e.register()
	if err != nil {
		log.Fatal("registration failed", zap.Error(err))
	}

	log.Info("Hello World")
}
